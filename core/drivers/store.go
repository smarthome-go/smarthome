package drivers

import (
	"encoding/json"
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

func StoreDriverSingleton(
	vendorID string,
	modelID string,
	fromJson any,
) error {
	marshaled, err := json.Marshal(fromJson)
	if err != nil {
		panic(fmt.Sprintf("Impossible marshal error: %s", err.Error()))
	}
	marshaledStr := string(marshaled)

	if _, err = database.ModifyDeviceDriverConfigJSON(
		vendorID,
		modelID,
		&marshaledStr,
	); err != nil {
		return err
	}

	fmt.Printf(
		"storing: %v in target singleton in file %s:%s...\n",
		spew.Sdump(fromJson),
		vendorID,
		modelID,
	)

	val, i := value.UnmarshalValue(errors.Span{}, fromJson)
	if i != nil {
		panic(fmt.Sprintf("Parsing / validation error: %s", (*i).Message()))
	}

	// storeSingletonInternal(SingletonKindDriver, DriverTuple{
	// 	VendorID: vendorID,
	// 	ModelID:  modelID,
	// }, (*val).(value.ValueObject))

	DriverStore[DriverTuple{
		VendorID: vendorID,
		ModelID:  modelID,
	}] = (*val).(value.ValueObject)

	spew.Dump(DriverStore)
	spew.Dump(DeviceStore)

	return nil
}

func storeSingletonInternal(targetSingleton DriverSingletonKind, file DriverTuple, value value.ValueObject) {
	switch targetSingleton {
	case SingletonKindDevice:
		DeviceStore[file] = value
	case SingletonKindDriver:
		DriverStore[file] = value
	default:
		panic(fmt.Sprintf("A new target singleton kind (%d) was added without updating this code", targetSingleton))
	}
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
		information, hmsErrs, err := extractInfoFromDriver(driver.VendorId, driver.ModelId, driver.HomescriptCode)
		if err != nil {
			return err
		}

		// Just skip this driver, its value will never be required anyways.
		if len(hmsErrs) > 0 {
			log.Tracef("Skipping default value instantiation of driver `%s:%s`", driver.VendorId, driver.ModelId)
			continue
		}

		storeSingletonInternal(SingletonKindDriver, DriverTuple{
			VendorID: driver.VendorId,
			ModelID:  driver.ModelId,
		}, value.ObjectZeroValue(information.DriverConfig.HmsType))

		log.Tracef("Populated driver store line `%s:%s` with default value", driver.VendorId, driver.ModelId)
	}

	return nil
}
