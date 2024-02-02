package drivers

import (
	"encoding/json"
	"fmt"

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

// Maps a device-ID to the corresponding saved value.
var DeviceStore map[string]value.ValueObject = make(map[string]value.ValueObject)
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
	fromJSON any,
) error {
	// TODO: need to patch the original value by only applying the changed fields.

	marshaled, err := json.Marshal(fromJSON)
	if err != nil {
		panic(fmt.Sprintf("Impossible marshal error: %s", err.Error()))
	}
	marshaledStr := string(marshaled)

	if err = database.ModifyDeviceDriverSingletonJSON(
		vendorID,
		modelID,
		&marshaledStr,
	); err != nil {
		return err
	}

	val, i := value.UnmarshalValue(errors.Span{}, fromJSON)
	if i != nil {
		panic(fmt.Sprintf("Parsing / validation error: %s", (*i).Message()))
	}

	DriverStore[DriverTuple{
		VendorID: vendorID,
		ModelID:  modelID,
	}] = (*val).(value.ValueObject)

	return nil
}

func StoreDeviceSingleton(
	deviceID string,
	fromJSON any,
) error {
	// TODO: need to patch the original value by only applying the changed fields.

	marshaled, err := json.Marshal(fromJSON)
	if err != nil {
		panic(fmt.Sprintf("Impossible marshal error: %s", err.Error()))
	}
	marshaledStr := string(marshaled)

	if err = database.ModifyDeviceSingletonJSON(
		deviceID,
		marshaledStr,
	); err != nil {
		return err
	}

	val, i := value.UnmarshalValue(errors.Span{}, fromJSON)
	if i != nil {
		panic(fmt.Sprintf("Parsing / validation error: %s", (*i).Message()))
	}

	DeviceStore[deviceID] = (*val).(value.ValueObject)

	return nil
}

//
// Generic, for both.
//

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

// func addZeroObjFieldsWithoutSettings(input value.ValueObject, singletonType ast.ObjectType) value.ValueObject {
// 	patched := value.ValueObject{
// 		FieldsInternal: input.FieldsInternal,
// 	}
//
// 	for _, field := range singletonType.ObjFields {
// 		// If this field does not exist on the value, fill it.
// 		fieldExists := false
// 		for fieldName := range input.FieldsInternal {
// 			if fieldName == field.FieldName.Ident() {
// 				fieldExists = true
// 				break
// 			}
// 		}
//
// 		if fieldExists {
// 			continue
// 		}
//
// 		patched.FieldsInternal[field.FieldName.Ident()] = value.ZeroValue(field.Type)
// 	}
//
// 	return patched
// }

func PopulateValueCache() error {
	// Retrieve devices list.
	devices, err := database.ListAllDevices()
	if err != nil {
		return err
	}

	// Populate driver singleton cache.
	drivers, err := database.ListDeviceDrivers()
	if err != nil {
		return err
	}

	for _, driver := range drivers {
		if driver.SingletonJSON != nil {
			var unmarshaledJSON any
			if err := json.Unmarshal([]byte(*driver.SingletonJSON), &unmarshaledJSON); err != nil {
				return fmt.Errorf("Could not parse driver JSON: %s", err.Error())
			}

			unmarshaledValue, i := value.UnmarshalValue(errors.Span{}, unmarshaledJSON)
			if i != nil {
				return fmt.Errorf("Could not parse driver JSON to HMS value: %s", (*i).Message())
			}

			DriverStore[DriverTuple{
				VendorID: driver.VendorId,
				ModelID:  driver.ModelId,
			}] = (*unmarshaledValue).(value.ValueObject)
		} else {
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

			DriverStore[DriverTuple{
				VendorID: driver.VendorId,
				ModelID:  driver.ModelId,
			}] = value.ObjectZeroValue(information.DriverConfig.HmsType)
		}

		// Populate each device which uses this driver.
		for _, device := range devices {
			if device.VendorId != driver.VendorId || device.ModelId != driver.ModelId {
				continue
			}

			val, found, err := RetrieveDeviceSingletonFromDB(device.Id)
			if err != nil {
				return err
			}

			if !found {
				panic(fmt.Sprintf("Device not found in database: `%s`", device.Id))
			}

			DeviceStore[device.Id] = val.(value.ValueObject)
		}
	}

	return nil
}

func RetrieveDeviceSingletonFromDB(deviceId string) (v value.Value, found bool, err error) {
	device, found, err := database.GetDeviceById(deviceId)
	if err != nil {
		return nil, false, err
	}

	if !found {
		return nil, false, nil
	}

	var unmarshaled any
	if err := json.Unmarshal([]byte(device.SingletonJSON), &unmarshaled); err != nil {
		return nil, false, err
	}

	unmarshaledValue, i := value.UnmarshalValue(errors.Span{}, unmarshaled)
	if i != nil {
		return nil, false, fmt.Errorf("%s", (*i).Message())
	}

	return *unmarshaledValue, true, nil
}
