package driver

import (
	"encoding/json"
	"fmt"
	"sync"

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
var ValueStoreLock sync.RWMutex

func GetDeviceSingleton(deviceId string) (value.ValueObject, bool) {
	val, found := DeviceStore[deviceId]
	return val, found
}

func GetDriverSingleton(vendor, model string) (value.ValueObject, bool) {
	val, found := DriverStore[database.DriverTuple{
		VendorID: vendor,
		ModelID:  model,
	}]
	return val, found
}

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
	affetedASetting, withOldValues := ApplyTransactionOnStored(
		oldValue,
		(*fromJSONhms).(value.ValueObject),
		driver.ExtractedInfo.DriverConfig.Info.HmsType,
	)

	if err := StoreDriverSingletonBackend(vendorID, modelID, withOldValues); err != nil {
		return err
	}

	// TODO: work on this.
	// devices, err := database.ListAllDevices()
	// if err != nil {
	// 	return err
	// }

	if affetedASetting {
		log.Debugf("Driver `%s:%s` singleton update affected a `settings` field, triggering reload...", driver.Driver.VendorID, driver.Driver.ModelID)
		// TODO: only make the driver dirty if the change originates from the web.
		// TODO: only trigger this if there were changes.
		// TLDR: only mark dirty if settings fields were altered.
		if err := MakeDriverDirty(vendorID, modelID, true); err != nil {
			return err
		}

		d.ReloadDriverCallBackFunc(driver.Driver)
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

	affetedASetting, withOldValues := ApplyTransactionOnStored(
		oldValue,
		(*fromJSONhms).(value.ValueObject),
		driver.ExtractedInfo.DeviceConfig.Info.HmsType,
	)

	if affetedASetting {
		d.ReloadDeviceCallBackFunc(device.ID)
	}

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

	ValueStoreLock.Lock()
	DeviceStore[deviceID] = val
	ValueStoreLock.Unlock()

	return nil
}

func MakeDriverDirty(vendorID, modelID string, dirty bool) error {
	// TODO: only do this if the driver actually uses trigger statements.

	if dirty {
		log.Debugf("Marking device driver `%s:%s` as dirty, it needs reloading...", vendorID, modelID)
	}

	if err := database.ModifyDeviceDriverDirty(vendorID, modelID, dirty); err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
) (changedASettingsField bool, newValR value.ValueObject) {
	// Use the old value as a starting point.
	transformed := oldVal

	for _, field := range singletonType.ObjFields {
		if field.Annotation == nil || field.Annotation.Ident() != DriverFieldRequiredAnnotation {
			// If this field is not a `@setting`, it can never be changed from the outside.

			newF, found := newVal.FieldsInternal[field.FieldName.Ident()]
			if !found {
				continue
			}

			if transformed.FieldsInternal[field.FieldName.Ident()] != newF {
				changedASettingsField = true
				transformed.FieldsInternal[field.FieldName.Ident()] = newF
			}
		}

		transformed.FieldsInternal[field.FieldName.Ident()] = newVal.FieldsInternal[field.FieldName.Ident()]
	}

	return changedASettingsField, transformed
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

func (d DriverManager) PopulateValueCache() error {
	ValueStoreLock.Lock()
	defer ValueStoreLock.Unlock()

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
		information, hmsErrs, err := d.extractInfoFromDriver(driver.VendorID, driver.ModelID, driver.HomescriptCode)
		if err != nil {
			return err
		}

		// Just skip this driver, its value will never be required anyways.
		if len(hmsErrs) > 0 {
			log.Tracef("Default value instantiation of driver `%s:%s` produced errors: %s", driver.VendorID, driver.ModelID, hmsErrs[0].Message)
		}

		if driver.SingletonJSON != nil {
			var unmarshaledJSON any
			if err := json.Unmarshal([]byte(*driver.SingletonJSON), &unmarshaledJSON); err != nil {
				return fmt.Errorf("Could not parse driver JSON: %s", err.Error())
			}

			unmarshaledValue := value.TypeAwareUnmarshalValue(unmarshaledJSON, information.DriverConfig.Info.HmsType)

			DriverStore[database.DriverTuple{
				VendorID: driver.VendorID,
				ModelID:  driver.ModelID,
			}] = (*unmarshaledValue).(value.ValueObject)
		} else {
			DriverStore[database.DriverTuple{
				VendorID: driver.VendorID,
				ModelID:  driver.ModelID,
			}] = value.ObjectZeroValue(information.DriverConfig.Info.HmsType)
		}

		// Populate each device which uses this driver.
		for _, device := range devices {
			if device.VendorID != driver.VendorID || device.ModelID != driver.ModelID {
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
