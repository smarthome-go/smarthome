package drivers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type RichDriver struct {
	Driver        database.DeviceDriver `json:"driver"`
	ExtractedInfo homescript.DriverInfo `json:"info"`
	// TODO: implement something like this for device as well
	// Saves the persistent value(s) of the setting-fields of the `Driver` singleton.
	// If this field is `nil`, the user has not configured their driver yet.
	Configuration    interface{}             `json:"configuration"`
	IsValid          bool                    `json:"isValid"`
	ValidationErrors []diagnostic.Diagnostic `json:"validationErrors"`
}

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func extractInfoFromDriver(
	vendorID string,
	modelID string,
	homescriptCode string,
) (homescript.DriverInfo, []diagnostic.Diagnostic, error) {
	driverInfo, diagnostics, err := homescript.ExtractDriverInfoTotal(
		vendorID,
		modelID,
		homescriptCode,
	)
	if err != nil {
		return homescript.DriverInfo{}, nil, err
	}

	// Filter step: only include actual errors, not warnings and infos.
	filtered := make([]diagnostic.Diagnostic, 0)
	for _, diag := range diagnostics {
		if diag.Level == diagnostic.DiagnosticLevelError {
			filtered = append(filtered, diag)
		}
	}

	if len(filtered) > 0 {
		log.Tracef("Driver `%s:%s` is not working: `%s`", vendorID, modelID, filtered[0].Message)
		return homescript.DriverInfo{}, filtered, nil
	}

	return driverInfo, make([]diagnostic.Diagnostic, 0), nil
}

func ListWithoutStoredValues() ([]RichDriver, error) {
	defaultDrivers, err := database.ListDeviceDrivers()
	if err != nil {
		return nil, err
	}

	richDrivers := make([]RichDriver, len(defaultDrivers))
	for idx, driver := range defaultDrivers {
		richDriver := RichDriver{
			Driver: driver,
			//nolint:exhaustruct
			ExtractedInfo:    homescript.DriverInfo{},
			Configuration:    nil,
			IsValid:          true,
			ValidationErrors: make([]diagnostic.Diagnostic, 0),
		}

		driverInfo, validationErrors, err := extractInfoFromDriver(driver.VendorId, driver.ModelId, driver.HomescriptCode)
		if err != nil {
			return nil, err
		}

		if len(validationErrors) > 0 {
			richDriver.IsValid = false
			richDriver.ValidationErrors = validationErrors
		} else {
			richDriver.ExtractedInfo = driverInfo
		}

		richDrivers[idx] = richDriver
	}

	return richDrivers, nil
}

func ListWithStoredConfig() ([]RichDriver, error) {
	drivers, err := ListWithoutStoredValues()
	if err != nil {
		return nil, err
	}

	for idx, driver := range drivers {
		if !driver.IsValid {
			log.Tracef("Skipping driver `%s:%s` in list with stored values: driver is not valid", driver.Driver.VendorId, driver.Driver.ModelId)
			continue
		}

		val, found := retrieveValueOfSingleton(
			DriverTuple{
				VendorID: driver.Driver.VendorId,
				ModelID:  driver.Driver.ModelId,
			},
			SingletonKindDriver,
		)

		// This should not happen: a zero value for every driver-spec is created automatically.
		if !found {
			panic(fmt.Sprintf("Configuration entry not found for driver `%s:%s`", driver.Driver.VendorId, driver.Driver.ModelId))
		}

		// TODO: deal with non-settings fields.

		configuration, _ := value.MarshalValue(
			filterObjFieldsWithoutSetting(val, driver.ExtractedInfo.DriverConfig.HmsType),
			false,
		)

		drivers[idx].Configuration = configuration
	}

	return drivers, nil
}

func Create(vendorID, modelID, name, version, hmsCode string) (hmsErr error, dbErr error) {
	driverData := database.DeviceDriver{
		VendorId:       vendorID,
		ModelId:        modelID,
		Name:           name,
		Version:        version,
		HomescriptCode: hmsCode,
		SingletonJSON:  nil,
	}

	// Try to create default JSON from schema.
	// This can fail if the Homescript code is invalid.
	configInfo, hmsErrs, err := extractInfoFromDriver(vendorID, modelID, hmsCode)
	if err != nil {
		return nil, err
	}

	if hmsErrs != nil {
		DriverStore[DriverTuple{
			VendorID: vendorID,
			ModelID:  modelID,
		}] = value.ObjectZeroValue(configInfo.DriverConfig.HmsType)
	} else {
		DriverStore[DriverTuple{
			VendorID: vendorID,
			ModelID:  modelID,
		}] = value.ValueObject{
			FieldsInternal: make(map[string]*value.Value),
		}
	}

	if err := database.CreateNewDeviceDriver(driverData); err != nil {
		return nil, err
	}

	return nil, nil
}

func ModifyCode(vendorID, modelID, newCode string) (found bool, dbErr error) {
	// Try to create default JSON from schema.
	// This can fail if the Homescript code is invalid.
	configInfo, hmsErrs, err := extractInfoFromDriver(vendorID, modelID, newCode)
	if err != nil {
		return false, err
	}

	if hmsErrs != nil {
		DriverStore[DriverTuple{
			VendorID: vendorID,
			ModelID:  modelID,
		}] = value.ObjectZeroValue(configInfo.DriverConfig.HmsType)
	} else {
		DriverStore[DriverTuple{
			VendorID: vendorID,
			ModelID:  modelID,
		}] = value.ValueObject{
			FieldsInternal: make(map[string]*value.Value),
		}
	}

	// TODO: detect if fields need to be removed, create a `delta` function in order to prevent overwriting EVERYTHING stored.

	found, err = database.ModifyDeviceDriverCode(
		vendorID,
		modelID,
		newCode,
	)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	// Try to extract a schema.
	schema, hmsErrs, err := extractInfoFromDriver(vendorID, modelID, newCode)
	if err != nil {
		return false, err
	}

	if len(hmsErrs) != 0 {
		return true, nil
	}

	// Create default value for extracted schema.
	if err != nil {
		panic(fmt.Sprintf("JSON marshaling failed: %s", err.Error()))
	}

	marshaled, _ := value.MarshalValue(value.ObjectZeroValue(schema.DriverConfig.HmsType), false)
	err = StoreDriverSingleton(vendorID, modelID, marshaled)
	if err != nil {
		return false, err
	}

	return true, nil
}

func ValidateDriverConfigurationChange(vendorID, modelID string, newConfig interface{}) (found bool, validateErr error, dbErr error) {
	driver, found, err := database.GetDeviceDriver(vendorID, modelID)
	if err != nil {
		return false, nil, err
	}

	if !found {
		return false, nil, nil
	}

	oldInfo, validationErrors, err := extractInfoFromDriver(vendorID, modelID, driver.HomescriptCode)
	if err != nil {
		return false, nil, err
	}

	if len(validationErrors) > 0 {
		return false, fmt.Errorf("%s", validationErrors[0].Message), nil
	}

	valid, stack, msg := valueMatchesSpec(newConfig, oldInfo.DriverConfig.Config, make([]FieldAccess, 0))

	stackStr := make([]string, 0)
	for _, elem := range stack {
		elemStr := ""
		switch elem.Type {
		case FieldAccessMember:
			elemStr = fmt.Sprintf(".%s", elem.Member)
		case FieldAccessIndex:
			elemStr = fmt.Sprintf("[%d]", elem.Index)
		}

		stackStr = append(stackStr, elemStr)
	}

	if !valid {
		return false, fmt.Errorf("Invalid new configuration: field `%s`: %s", strings.Join(stackStr, ""), msg), nil
	}

	return true, nil, nil
}

type FieldAccessType uint8

const (
	FieldAccessMember = iota
	FieldAccessIndex
)

type FieldAccess struct {
	Type   FieldAccessType
	Member string
	Index  int
}

func valueMatchesSpec(
	configValue interface{},
	spec homescript.ConfigFieldDescriptor,
	fieldAccessStack []FieldAccess,
) (
	valid bool,
	fieldAccessStackOut []FieldAccess,
	errMsg string,
) {
	switch self := configValue.(type) {
	case string:
		if spec.Kind() != homescript.CONFIG_FIELD_TYPE_STRING {
			return false, fieldAccessStack, fmt.Sprintf("Expected %s, got STRING", spec.Kind())
		}
		return true, fieldAccessStack, ""
	case int, int64:
		if spec.Kind() != homescript.CONFIG_FIELD_TYPE_INT {
			return false, fieldAccessStack, fmt.Sprintf("Expected %s, got INT", spec.Kind())
		}
		return true, nil, ""
	case float64:
		// Check if this is actually an int and an int was expected.
		if float64(int64(self)) == self && spec.Kind() == homescript.CONFIG_FIELD_TYPE_INT {
			return true, nil, ""
		}

		if spec.Kind() != homescript.CONFIG_FIELD_TYPE_FLOAT {
			return false, fieldAccessStack, fmt.Sprintf("Expected %s, got FLOAT", spec.Kind())
		}
		return true, nil, ""
	case bool:
		if spec.Kind() != homescript.CONFIG_FIELD_TYPE_BOOL {
			return false, fieldAccessStack, fmt.Sprintf("Expected %s, got BOOL", spec.Kind())
		}
		return true, nil, ""
	case map[string]interface{}:
		if spec.Kind() != homescript.CONFIG_FIELD_TYPE_STRUCT {
			return false, fieldAccessStack, fmt.Sprintf("Expected %s, got STRUCT", spec.Kind())
		}

		structSpec := spec.(homescript.ConfigFieldDescriptorStruct)

		if len(structSpec.Fields) != len(self) {
			return false, fieldAccessStack, fmt.Sprintf("Expected %d object fields, got %d", len(structSpec.Fields), len(self))
		}

		for _, field := range structSpec.Fields {
			item, found := self[field.Name]
			if !found {
				return false, fieldAccessStack, fmt.Sprintf("Missing object field `%s`", field.Name)
			}

			if valid, stack, msg := valueMatchesSpec(item, field.Type, append(fieldAccessStack, FieldAccess{
				Type:   FieldAccessMember,
				Member: field.Name,
				Index:  0,
			})); !valid {
				return false, stack, msg
			}
		}

		return true, nil, ""
	case []interface{}:
		listSpec := spec.(homescript.ConfigFieldDescriptorWithInner)

		for index, elem := range self {
			if valid, stack, msg := valueMatchesSpec(elem, listSpec.Inner, append(fieldAccessStack, FieldAccess{
				Type:   FieldAccessIndex,
				Index:  index,
				Member: "",
			})); !valid {
				return false, stack, msg
			}
		}

		return true, nil, ""
	default:
		panic(fmt.Sprintf("unhandled case: %v", reflect.TypeOf(configValue)))
	}
}
