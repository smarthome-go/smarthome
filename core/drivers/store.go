package drivers

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type DriverTuple struct {
	VendorID string `json:"vendorId"`
	ModelID  string `json:"modelId"`
}

type DriverSingletonKind uint8

const (
	SingletonKindDriver DriverSingletonKind = iota
	SingletonKindDevice
)

// TODO: add deep clone for retrieval?

var DeviceStore map[DriverTuple]value.ValueObject = make(map[DriverTuple]value.ValueObject)
var DriverStore map[DriverTuple]value.ValueObject = make(map[DriverTuple]value.ValueObject)

// This package contains the storage backend implementation for per-driver / per-device configuration data.
//
//	1. The user sends a JSON configuration string
//	2. The HTTP layer parses this request and passes it into this module
//	3. This module performs sanity-checks on the JSON and tries to parse it into a Homescript value
//		- Here, the schema / type of the according singleton must be used to validate that the JSON has a valid schema.
//	4. If the parsing succeeded, store the data in: TODO any kind of database

func StoreValueInSingleton(
	file DriverTuple,
	targetSingleton DriverSingletonKind,
	fromJson any,
) (found bool, softErr error, dbErr error) {
	driver, found, err := database.GetDeviceDriver(file.VendorID, file.ModelID)
	if err != nil {
		return false, nil, err
	}

	if !found {
		return false, nil, nil
	}

	_, hmsErrs, err := extractInfoFromDriver(driver)
	if err != nil {
		return false, nil, err
	}

	if len(hmsErrs) > 0 {
		return false,
			fmt.Errorf(
				"Could not extract driver information for sanity check: %s",
				hmsErrs[0].Display(driver.HomescriptCode),
			),
			nil
	}

	fmt.Printf("storing: %v in target singleton %d in file %s:%s...\n", fromJson, targetSingleton, file.VendorID, file.ModelID)

	val, i := value.UnmarshalValue(errors.Span{}, fromJson)
	if i != nil {
		panic(fmt.Sprintf("Parsing / validation error: %s", (*i).Message()))
	}

	switch targetSingleton {
	case SingletonKindDevice:
		DeviceStore[file] = (*val).(value.ValueObject)
	case SingletonKindDriver:
		DriverStore[file] = (*val).(value.ValueObject)
	default:
		panic(fmt.Sprintf("A new target singleton kind (%d) was added without updating this code", targetSingleton))
	}

	spew.Dump(DriverStore)
	spew.Dump(DeviceStore)

	return true, nil, nil
}

func retrieveValueOfSingleton(file DriverTuple, targetSingleton DriverSingletonKind) (res value.ValueObject, found bool) {
	switch targetSingleton {
	case SingletonKindDevice:
		res, found = DeviceStore[file]
	case SingletonKindDriver:
		res, found = DriverStore[file]
	default:
		panic(fmt.Sprintf("A new target singleton kind (%d) was added without updating this code", targetSingleton))
	}
	return res, found
}

func filterObjFieldsWithoutSetting(input value.ValueObject, singletonType ast.ObjectType) value.ValueObject {
	outputFields := make(map[string]*value.Value)

	for fieldName, field := range input.FieldsInternal {
		// Check whether this field has the required annotation.
		isSetting := false

		for _, typeField := range singletonType.ObjFields {
			if typeField.FieldName.Ident() == fieldName &&
				typeField.Annotation != nil &&
				typeField.Annotation.Ident() == homescript.DriverFieldRequiredAnnotation {
				isSetting = true
				break
			}
		}

		if !isSetting {
			continue
		}

		outputFields[fieldName] = field
	}

	return value.ValueObject{
		FieldsInternal: outputFields,
	}
}

func PopulateValueCache() error {
	drivers, err := database.ListDeviceDrivers()
	if err != nil {
		return err
	}

	for _, driver := range drivers {
		if false {
			// TODO: load persistent state from database.
			continue
		}

		// Load type information for this driver.
		information, hmsErrs, err := extractInfoFromDriver(driver)
		if err != nil {
			return err
		}

		// Just skip this driver, its value will never be required anyways.
		if len(hmsErrs) > 0 {
			log.Tracef("Skipping default value instantiation of driver `%s:%s`", driver.VendorId, driver.ModelId)
			continue
		}

		DriverStore[DriverTuple{
			VendorID: driver.VendorId,
			ModelID:  driver.ModelId,
		}] = value.ObjectZeroValue(information.DriverConfig.HmsType)

		log.Tracef("Populated driver store line `%s:%s` with default value", driver.VendorId, driver.ModelId)
	}

	return nil
}

// TODO: remove this.
// func createDefaultConfigurationFromSpec(spec homescript.ConfigFieldDescriptor) value.Value {
// 	switch spec.Kind() {
// 	case homescript.CONFIG_FIELD_TYPE_INT:
// 		return *value.NewValueInt(0)
// 	case homescript.CONFIG_FIELD_TYPE_FLOAT:
// 		return *value.NewValueFloat(0.0)
// 	case homescript.CONFIG_FIELD_TYPE_BOOL:
// 		return *value.NewValueBool(false)
// 	case homescript.CONFIG_FIELD_TYPE_STRING:
// 		return *value.NewValueString("")
// 	case homescript.CONFIG_FIELD_TYPE_LIST:
// 		return *value.NewValueList(make([]*value.Value, 0))
// 	case homescript.CONFIG_FIELD_TYPE_STRUCT:
// 		// nolint:forcetypeassert
// 		structSpec := spec.(homescript.ConfigFieldDescriptorStruct)
//
// 		fields := make(map[string]*value.Value)
//
// 		for _, field := range structSpec.Fields {
// 			v := createDefaultConfigurationFromSpec(field.Type)
// 			fields[field.Name] = &v
// 		}
//
// 		return *value.NewValueObject(fields)
// 	case homescript.CONFIG_FIELD_TYPE_OPTION:
// 		return *value.NewValueNull()
// 	}
//
// 	panic(fmt.Sprintf("A new config spec was added without updating this code: %s", spec.Kind()))
// }
