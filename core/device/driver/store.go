package driver

import (
	"encoding/json"
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
)

type DriverSingletonKind uint8

const (
	SingletonKindDriver DriverSingletonKind = iota
	SingletonKindDevice
)

// TODO: add deep clone for retrieval?

///
/// Device driver singleton store and cache.
/// This is a write-through cache as writing does not happen often (once at the end of each driver invocation)
/// However, the written data is quite important and should be saved in the database.
///

// Maps a device-ID to the corresponding saved value.
var DeviceStore map[string]value.ValueObject = make(map[string]value.ValueObject)
var DriverStore map[database.DriverTuple]value.ValueObject = make(map[database.DriverTuple]value.ValueObject)

// This package contains the storage backend implementation for per-driver / per-device configuration data.
//
//	1. The user sends a JSON configuration string
//	2. The HTTP layer parses this request and passes it into this module
//	3. This module performs sanity-checks on the JSON and tries to parse it into a Homescript value
//		- Here, the schema / type of the according singleton must be used to validate that the JSON has a valid schema.
//	4. If the parsing succeeded, store the data in: TODO any kind of database

// This function exists alongside its backend because invoking this function BEFORE new HMS code is saved in the DB
// would revert the schema changes that it is intended to write, as it pulls data from the DB
// which at this point resides in an outdated state.
func (d DriverManager) StoreDriverSingletonConfigUpdate(
	vendorID string,
	modelID string,
	fromJSON any,
) error {
	// TODO: need to patch the original value by only applying the changed fields.
	driver, found, err := d.GetDriverWithInfos(vendorID, modelID)
	if err != nil {
		return err
	}

	if !found {
		panic(fmt.Sprintf("Driver `%s:%s` to be stored not found", vendorID, modelID))
	}

	oldValue := DriverStore[database.DriverTuple{
		VendorID: vendorID,
		ModelID:  modelID,
	}]

	fromJSONhms := value.TypeAwareUnmarshalValue(fromJSON, driver.ExtractedInfo.DriverConfig.Info.HmsType)
	withOldValues := ApplyTransactionOnStored(
		oldValue,
		(*fromJSONhms).(value.ValueObject),
		driver.ExtractedInfo.DriverConfig.Info.HmsType,
	)

	if err := StoreDriverSingletonBackend(vendorID, modelID, withOldValues); err != nil {
		return err
	}

	devices, err := database.ListAllDevices()
	if err != nil {
		return err
	}

	for _, dev := range devices {
		fmt.Printf("DEVICE: %s\n", dev.ID)
		// TODO: trigger a re-registration of any triggers.
		panic("TODO: not implemented")
	}

	return nil
}

// This function just stores a value in the store backend without applying transformations on it.
func StoreDriverSingletonBackend(vendorID, modelID string, val value.ValueObject) error {
	marshaledInterface, _ := value.MarshalValue(val, false)

	marshaledBytes, err := json.Marshal(marshaledInterface)
	if err != nil {
		panic(fmt.Sprintf("Impossible marshal error: %s", err.Error()))
	}

	marshaledStr := string(marshaledBytes)

	if err = database.ModifyDeviceDriverSingletonJSON(
		vendorID,
		modelID,
		&marshaledStr,
	); err != nil {
		return err
	}

	DriverStore[database.DriverTuple{
		VendorID: vendorID,
		ModelID:  modelID,
	}] = val

	return nil
}

// This function exists alongside its backend because invoking this function BEFORE new HMS code is saved in the DB
// would revert the schema changes that it is intended  to write, as it pulls data from the DB
// which at this point resides in an outdated state.
func (d DriverManager) StoreDeviceSingletonConfigUpdate(
	deviceID string,
	fromJSON any,
) error {
	device, found, err := database.GetDeviceById(deviceID)
	if err != nil {
		return err
	}

	if !found {
		panic(fmt.Sprintf("Device `%s` to be stored not found", deviceID))
	}

	driver, found, err := d.GetDriverWithInfos(device.VendorID, device.ModelID)
	if err != nil {
		return err
	}

	if !found {
		panic(fmt.Sprintf("Driver `%s:%s` to be stored not found", device.VendorID, device.ModelID))
	}

	oldValue := DeviceStore[deviceID]

	fromJSONhms := value.TypeAwareUnmarshalValue(fromJSON, driver.ExtractedInfo.DeviceConfig.Info.HmsType)

	withOldValues := ApplyTransactionOnStored(
		oldValue,
		(*fromJSONhms).(value.ValueObject),
		driver.ExtractedInfo.DeviceConfig.Info.HmsType,
	)

	// TODO: trigger re-registration of any triggers.
	panic("TODO: not implemented")

	return StoreDeviceSingletonBackend(deviceID, withOldValues)
}

// This function just stores a value in the store backend without applying transformations on it.
func StoreDeviceSingletonBackend(deviceID string, val value.ValueObject) error {
	marshaledInterface, _ := value.MarshalValue(val, false)

	marshaledBytes, err := json.Marshal(marshaledInterface)
	if err != nil {
		panic(fmt.Sprintf("Impossible marshal error: %s", err.Error()))
	}

	marshaledStr := string(marshaledBytes)

	if err = database.ModifyDeviceSingletonJSON(
		deviceID,
		marshaledStr,
	); err != nil {
		return err
	}

	DeviceStore[deviceID] = val
	return nil
}

//
// Generic, for both.
//

// This function patches the `old` value using the `new` fields whilst respecting the annotated `@setting` fields.
// It is required as otherwise, just the `new` value would be saved in the DB,
// causing a VM runtime crash due to possibly missing fields.
func ApplyTransactionOnStored(
	oldVal value.ValueObject,
	newVal value.ValueObject,
	singletonType ast.ObjectType,
) value.ValueObject {
	// Use the old value as a starting point.
	transformed := oldVal

	for _, field := range singletonType.ObjFields {
		if field.Annotation == nil || field.Annotation.Ident() != DriverFieldRequiredAnnotation {
			// If this field is not a `@setting`, it can never be changed from the outside.
			continue
		}

		transformed.FieldsInternal[field.FieldName.Ident()] = newVal.FieldsInternal[field.FieldName.Ident()]
	}

	return transformed
}

func filterObjFieldsWithoutSetting(input value.ValueObject, singletonType ast.ObjectType) value.ValueObject {
	outputFields := make(map[string]*value.Value)

	for fieldName, field := range input.FieldsInternal {
		// Check whether this field has the required annotation.
		isSetting := false

		for _, typeField := range singletonType.ObjFields {
			if typeField.FieldName.Ident() == fieldName &&
				typeField.Annotation != nil &&
				typeField.Annotation.Ident() == DriverFieldRequiredAnnotation {
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

func (d DriverManager) PopulateValueCache() error {
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
		// Load type information for this driver.
		information, hmsErrs, err := d.extractInfoFromDriver(driver.VendorId, driver.ModelId, driver.HomescriptCode)
		if err != nil {
			return err
		}

		// Just skip this driver, its value will never be required anyways.
		if len(hmsErrs) > 0 {
			log.Tracef("Skipping default value instantiation of driver `%s:%s`", driver.VendorId, driver.ModelId)
			continue
		}

		if driver.SingletonJSON != nil {
			var unmarshaledJSON any
			if err := json.Unmarshal([]byte(*driver.SingletonJSON), &unmarshaledJSON); err != nil {
				return fmt.Errorf("Could not parse driver JSON: %s", err.Error())
			}

			unmarshaledValue := value.TypeAwareUnmarshalValue(unmarshaledJSON, information.DriverConfig.Info.HmsType)

			DriverStore[database.DriverTuple{
				VendorID: driver.VendorId,
				ModelID:  driver.ModelId,
			}] = (*unmarshaledValue).(value.ValueObject)
		} else {
			DriverStore[database.DriverTuple{
				VendorID: driver.VendorId,
				ModelID:  driver.ModelId,
			}] = value.ObjectZeroValue(information.DriverConfig.Info.HmsType)
		}

		// Populate each device which uses this driver.
		for _, device := range devices {
			if device.VendorID != driver.VendorId || device.ModelID != driver.ModelId {
				continue
			}

			val, found, err := RetrieveDeviceSingletonFromDB(device.ID, information.DeviceConfig.Info.HmsType)
			if err != nil {
				return err
			}

			if !found {
				panic(fmt.Sprintf("Device not found in database: `%s`", device.ID))
			}

			DeviceStore[device.ID] = val.(value.ValueObject)
		}
	}

	return nil
}

func RetrieveDeviceSingletonFromDB(deviceID string, hmsType ast.Type) (v value.Value, found bool, err error) {
	device, found, err := database.GetDeviceById(deviceID)
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

	return *value.TypeAwareUnmarshalValue(unmarshaled, hmsType), true, nil
}
