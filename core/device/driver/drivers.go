package driver

import (
	_ "embed"
	"fmt"
	"reflect"
	"strings"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
)

type RichDriver struct {
	Driver        database.DeviceDriver `json:"driver"`
	ExtractedInfo DriverInfo            `json:"info"`
	// TODO: implement something like this for device as well
	// Saves the persistent value(s) of the setting-fields of the `Driver` singleton.
	// If this field is `nil`, the user has not configured their driver yet.
	Configuration    interface{}             `json:"configuration"`
	IsValid          bool                    `json:"isValid"`
	ValidationErrors []diagnostic.Diagnostic `json:"validationErrors"`
}

func (self RichDriver) DeviceSupports(check DeviceCapability) bool {
	return self.ExtractedInfo.DeviceConfig.Capabilities.Has(check)
}

func (d DriverManager) extractInfoFromDriver(
	vendorID string,
	modelID string,
	homescriptCode string,
) (DriverInfo, []diagnostic.Diagnostic, error) {
	driverInfo, diagnostics, err := d.ExtractDriverInfoTotal(
		vendorID,
		modelID,
		homescriptCode,
	)
	if err != nil {
		return DriverInfo{}, nil, err
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
		// nolint:exhaustruct
		return DriverInfo{}, filtered, nil
	}

	return driverInfo, make([]diagnostic.Diagnostic, 0), nil
}

func (d DriverManager) GetDriverWithInfos(vendorID, modelID string) (RichDriver, bool, error) {
	rawDriver, found, err := database.GetDeviceDriver(vendorID, modelID)
	if err != nil {
		return RichDriver{}, false, err
	}

	if !found {
		return RichDriver{}, false, nil
	}

	driverInfo, diagnostics, err := d.extractInfoFromDriver(vendorID, modelID, rawDriver.HomescriptCode)
	if err != nil {
		return RichDriver{}, false, err
	}

	configuration := DriverStore[database.DriverTuple{
		VendorID: vendorID,
		ModelID:  modelID,
	}]

	marshaled, _ := value.MarshalValue(configuration, false)

	return RichDriver{
		Driver:           rawDriver,
		ExtractedInfo:    driverInfo,
		Configuration:    marshaled,
		IsValid:          len(diagnostics) == 0,
		ValidationErrors: diagnostics,
	}, true, nil
}

func (d DriverManager) ListDriversWithoutStoredValues() ([]RichDriver, error) {
	defaultDrivers, err := database.ListDeviceDrivers()
	if err != nil {
		return nil, err
	}

	richDrivers := make([]RichDriver, len(defaultDrivers))
	for idx, driver := range defaultDrivers {
		richDriver := RichDriver{
			Driver: driver,
			//nolint:exhaustruct
			ExtractedInfo:    DriverInfo{},
			Configuration:    nil,
			IsValid:          true,
			ValidationErrors: make([]diagnostic.Diagnostic, 0),
		}

		driverInfo, validationErrors, err := d.extractInfoFromDriver(driver.VendorID, driver.ModelID, driver.HomescriptCode)
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

func (d DriverManager) ListDriversWithStoredConfig() ([]RichDriver, error) {
	drivers, err := d.ListDriversWithoutStoredValues()
	if err != nil {
		return nil, err
	}

	for idx, driver := range drivers {
		if !driver.IsValid {
			log.Tracef("Skipping driver `%s:%s` in list with stored values: driver is not valid", driver.Driver.VendorID, driver.Driver.ModelID)
			continue
		}

		val, found := DriverStore[database.DriverTuple{
			VendorID: driver.Driver.VendorID,
			ModelID:  driver.Driver.ModelID,
		}]

		// This should not happen: a zero value for every driver-spec is created automatically.
		if !found {
			panic(fmt.Sprintf("Configuration entry not found for driver `%s:%s`", driver.Driver.VendorID, driver.Driver.ModelID))
		}

		// TODO: deal with non-settings fields.

		configuration, _ := value.MarshalValue(
			filterObjFieldsWithoutSetting(val, driver.ExtractedInfo.DriverConfig.Info.HmsType),
			false,
		)

		drivers[idx].Configuration = configuration
	}

	return drivers, nil
}

func (d DriverManager) CreateDriver(vendorID, modelID, name, version string, hmsCode *string) (hmsErr error, dbErr error) {
	hmsCodeToUse := database.DefaultDriverHomescriptCode

	if hmsCode != nil {
		hmsCodeToUse = *hmsCode
	}

	driverData := database.DeviceDriver{
		VendorID:       vendorID,
		ModelID:        modelID,
		Name:           name,
		Version:        version,
		HomescriptCode: hmsCodeToUse,
		SingletonJSON:  nil,
	}

	// Try to create default JSON from schema.
	// This can fail if the Homescript code is invalid.
	configInfo, hmsErrs, err := d.extractInfoFromDriver(vendorID, modelID, hmsCodeToUse)
	if err != nil {
		return nil, err
	}

	if hmsErrs != nil {
		DriverStore[database.DriverTuple{
			VendorID: vendorID,
			ModelID:  modelID,
		}] = value.ObjectZeroValue(configInfo.DriverConfig.Info.HmsType)
	} else {
		DriverStore[database.DriverTuple{
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

func ApplyNewSchemaOnData(oldData value.Value, newSchema ast.Type) (newData *value.Value) {
	switch newSchema.Kind() {
	case ast.UnknownTypeKind, ast.NeverTypeKind, ast.AnyTypeKind, ast.NullTypeKind,
		ast.IdentTypeKind, ast.FnTypeKind, ast.RangeTypeKind, ast.AnyObjectTypeKind:
		panic(fmt.Sprintf("Unsupported type: `%s`", newSchema.Kind()))
	case ast.IntTypeKind:
		fallthrough
	case ast.FloatTypeKind:
		fallthrough
	case ast.BoolTypeKind:
		fallthrough
	case ast.StringTypeKind:
		// If there is a type mismatch, create a zero value.
		if oldData.Kind().TypeKind() != newSchema.Kind() {
			return value.ZeroValue(newSchema)
		}

		// Otherwise, return the original value.
		return &oldData
	case ast.ListTypeKind:
		valList := oldData.(value.ValueList)
		if len(*valList.Values) == 0 {
			return value.NewValueList(make([]*value.Value, 0))
		}

		// If the first list element differs from the new schema, return an empty list.
		first := *(*valList.Values)[0]
		listType := newSchema.(ast.ListType)
		if first.Kind().TypeKind() != listType.Inner.Kind() {
			return value.NewValueList(make([]*value.Value, 0))
		}

		return value.NewValueList(*valList.Values)
	case ast.ObjectTypeKind:
		return ApplyNewSchemaOnObjData(oldData.(value.ValueObject), newSchema.(ast.ObjectType))
	case ast.OptionTypeKind:
		oldOption := oldData.(value.ValueOption)
		if oldOption.Inner == nil {
			return value.NewNoneOption()
		}

		// If the inner type differs, also return a `none` option.
		optionType := newSchema.(ast.OptionType)
		if (*oldOption.Inner).Kind().TypeKind() != optionType.Inner.Kind() {
			return value.NewNoneOption()
		}

		return value.NewValueOption(oldOption.Inner)
	default:
		panic(fmt.Sprintf("A new data type was added without updating this code: `%s`", newSchema.Kind()))
	}
}

// Tries to transform old data into a new schema without loosing too much information.
// TODO: return a bool that indicates whether a field was removed or that its data was invalidated.
// This will be useful for a warning that informs the user that 'committing' change will likely result in data loss.
func ApplyNewSchemaOnObjData(oldData value.ValueObject, newSchema ast.ObjectType) (newData *value.Value) {
	newFields := make(map[string]*value.Value)

outer:
	// By iterating over the new fields whilst ignoring any old ones, removed fields are deleted automatically.
	for _, field := range newSchema.ObjFields {
		// If this field already exists, apply schema recursively on this field.
		for objFieldName, objField := range oldData.FieldsInternal {
			if objFieldName == field.FieldName.Ident() {
				newFields[field.FieldName.Ident()] = ApplyNewSchemaOnData(*objField, field.Type)
				continue outer
			}
		}

		// Other case: this field does not currently exist: create a zero value.
		newFields[field.FieldName.Ident()] = value.ZeroValue(field.Type)
	}

	return value.NewValueObject(newFields)
}

// Apart from actually modifying the code of the driver in the DB,
// the saved singleton state of this driver and all dependent devices must be rebuilt.
func (d DriverManager) ModifyCode(vendorID, modelID, newCode string) (found bool, dbErr error) {
	// Try to create default JSON from schema. TODO: why default: ???
	// This can fail if the Homescript code is invalid.
	configInfo, hmsErrs, err := d.extractInfoFromDriver(vendorID, modelID, newCode)
	if err != nil {
		return false, err
	}

	// Only apply transactions if there are no errors in the code.
	// Otherwise, the stored data of every singleton would be erased as soon as there is an error.
	if hmsErrs != nil {
		log.Debugf("[singleton] Not updating singleton stores of driver / devices due to errors in new code")
		return true, nil
	}

	objVal := value.ValueObject{FieldsInternal: make(map[string]*value.Value)}

	// TODO: add proper error handling in here:
	// - check if there is an erorr and return early
	// - otherwise (no error) load the current data and perform the patches on it.

	if hmsErrs == nil {
		old := DriverStore[database.DriverTuple{
			VendorID: vendorID,
			ModelID:  modelID,
		}]
		objVal = (*ApplyNewSchemaOnObjData(old, configInfo.DriverConfig.Info.HmsType)).(value.ValueObject)

		if err := StoreDriverSingletonBackend(vendorID, modelID, objVal); err != nil {
			return false, err
		}
	}

	// Also patch every device that uses this driver.
	// TODO: add a separate device-list that only lists devices of a certain driver.

	// TODO: what to do on HMS errors?

	devices, err := database.ListAllDevices()
	if err != nil {
		return false, err
	}
	for _, device := range devices {
		if device.VendorID != vendorID || device.ModelID != modelID {
			continue
		}
		oldDeviceData := DeviceStore[device.ID]
		newDeviceData := (*ApplyNewSchemaOnObjData(oldDeviceData, configInfo.DeviceConfig.Info.HmsType)).(value.ValueObject)
		if err := StoreDeviceSingletonBackend(device.ID, newDeviceData); err != nil {
			return false, err
		}
	}

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

	return true, nil
}

// TODO: a lot of overlapping code!
func (d DriverManager) ValidateDeviceConfigurationChange(deviceId string, newConfig interface{}) (found bool, validateErr error, dbErr error) {
	device, found, err := database.GetDeviceById(deviceId)
	if err != nil {
		return false, nil, err
	}

	if !found {
		return false, nil, nil
	}

	// Retrieve driver in order to perform validation
	driver, found, err := database.GetDeviceDriver(device.VendorID, device.ModelID)
	if err != nil {
		return false, nil, err
	}

	if !found {
		panic(fmt.Sprintf("Driver `%s:%s` was not found in DB", device.VendorID, device.ModelID))
	}

	oldInfo, validationErrors, err := d.extractInfoFromDriver(device.VendorID, device.ModelID, driver.HomescriptCode)
	if err != nil {
		return false, nil, err
	}

	if len(validationErrors) > 0 {
		return false, fmt.Errorf("%s", validationErrors[0].Message), nil
	}

	valid, stack, msg := valueMatchesSpec(newConfig, oldInfo.DeviceConfig.Info.Config, make([]FieldAccess, 0))

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
		stackDisp := ""
		if len(stackStr) > 0 {
			stackDisp = fmt.Sprintf("field `%s`: ", strings.Join(stackStr, ""))
		}
		return false, fmt.Errorf("Invalid new configuration: %s%s", stackDisp, msg), nil
	}

	return true, nil, nil
}

func (d DriverManager) ValidateDriverConfigurationChange(vendorID, modelID string, newConfig interface{}) (found bool, validateErr error, dbErr error) {
	driver, found, err := database.GetDeviceDriver(vendorID, modelID)
	if err != nil {
		return false, nil, err
	}

	if !found {
		return false, nil, nil
	}

	oldInfo, validationErrors, err := d.extractInfoFromDriver(vendorID, modelID, driver.HomescriptCode)
	if err != nil {
		return false, nil, err
	}

	if len(validationErrors) > 0 {
		return false, fmt.Errorf("%s", validationErrors[0].Message), nil
	}

	valid, stack, msg := valueMatchesSpec(newConfig, oldInfo.DriverConfig.Info.Config, make([]FieldAccess, 0))

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
		stackDisp := ""
		if len(stackStr) > 0 {
			stackDisp = fmt.Sprintf("field `%s`: ", strings.Join(stackStr, ""))
		}
		return false, fmt.Errorf("Invalid new configuration: %s%s", stackDisp, msg), nil
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
	spec ConfigFieldDescriptor,
	fieldAccessStack []FieldAccess,
) (
	valid bool,
	fieldAccessStackOut []FieldAccess,
	errMsg string,
) {
	switch self := configValue.(type) {
	case string:
		if spec.Kind() != CONFIG_FIELD_TYPE_STRING {
			return false, fieldAccessStack, fmt.Sprintf("Expected %s, got STRING", spec.Kind())
		}
		return true, fieldAccessStack, ""
	case int, int64:
		if spec.Kind() != CONFIG_FIELD_TYPE_INT {
			return false, fieldAccessStack, fmt.Sprintf("Expected %s, got INT", spec.Kind())
		}
		return true, nil, ""
	case float64:
		// Check if this is actually an int and an int was expected.
		if float64(int64(self)) == self && spec.Kind() == CONFIG_FIELD_TYPE_INT {
			return true, nil, ""
		}

		if spec.Kind() != CONFIG_FIELD_TYPE_FLOAT {
			return false, fieldAccessStack, fmt.Sprintf("Expected %s, got FLOAT", spec.Kind())
		}
		return true, nil, ""
	case bool:
		if spec.Kind() != CONFIG_FIELD_TYPE_BOOL {
			return false, fieldAccessStack, fmt.Sprintf("Expected %s, got BOOL", spec.Kind())
		}
		return true, nil, ""
	case map[string]interface{}:
		if spec.Kind() != CONFIG_FIELD_TYPE_STRUCT {
			return false, fieldAccessStack, fmt.Sprintf("Expected %s, got STRUCT", spec.Kind())
		}

		structSpec := spec.(ConfigFieldDescriptorStruct)

		// Check that all struct fields are satisfied.
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

		// Check that there are no new fields which do not exist on the struct.
	outer:
		for name := range self {
			for _, field := range structSpec.Fields {
				if field.Name == name {
					continue outer
				}
			}

			return false, fieldAccessStack, fmt.Sprintf("Illegal additional field `%s`", name)
		}

		return true, nil, ""
	case []interface{}:
		listSpec := spec.(ConfigFieldDescriptorWithInner)

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
