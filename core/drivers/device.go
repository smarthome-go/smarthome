package drivers

import (
	"encoding/json"
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
)

type Device struct {
	DeviceType     database.DEVICE_TYPE `json:"type"`
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	RoomID         string               `json:"roomId"`
	DriverVendorID string               `json:"vendorId"`
	DriverModelID  string               `json:"modelId"`
	SingletonJSON  any                  `json:"singletonJson"`
}

func ListAllDevices() ([]Device, error) {
	raw, err := database.ListAllDevices()
	if err != nil {
		return nil, err
	}

	return ParseSingletonListJSON(raw)
}

func ListPersonalDevices(username string) ([]Device, error) {
	raw, err := database.ListUserDevices(username)
	if err != nil {
		return nil, err
	}

	return ParseSingletonListJSON(raw)
}

func ParseSingletonListJSON(input []database.Device) ([]Device, error) {
	output := make([]Device, len(input))
	for i, elem := range input {
		var parsed any

		if err := json.Unmarshal([]byte(elem.SingletonJSON), &parsed); err != nil {
			log.Errorf("Could not parse singleton of device `%s` JSON: %s", elem.Id, err.Error())
			return nil, fmt.Errorf("Could not parse singleton of device `%s` JSON: %s", elem.Id, err.Error())
		}

		output[i] = Device{
			DeviceType:     elem.DeviceType,
			ID:             elem.Id,
			Name:           elem.Name,
			RoomID:         elem.RoomId,
			DriverVendorID: elem.VendorId,
			DriverModelID:  elem.ModelId,
			SingletonJSON:  parsed,
		}
	}
	return output, nil
}

// 1. Fetch the corresponding driver from the DB
// 2. Generate a default JSON from the driver device singleton
// 3. Create the new device.
func CreateDevice(
	type_ database.DEVICE_TYPE,
	id string,
	name string,
	roomID string,
	driverVendorID string,
	driverModelID string,
) (driverFound bool, hmsErr, dbErr error) {
	// Retrieve corresponding driver.
	driver, found, err := database.GetDeviceDriver(driverVendorID, driverModelID)
	if err != nil {
		return false, nil, err
	}

	if !found {
		return false, nil, nil
	}

	driverInfo, validationErrors, err := extractInfoFromDriver(driver.VendorId, driver.ModelId, driver.HomescriptCode)
	if err != nil {
		return false, nil, err
	}

	if len(validationErrors) > 0 {
		return false, fmt.Errorf("Could not extract driver schema: %s", validationErrors[0].Message), nil
	}

	// Generate default JSON from driver info.
	defaultDevice := value.ObjectZeroValue(driverInfo.DeviceConfig.HmsType)
	defaultDeviceInterface, _ := value.MarshalValue(defaultDevice, false)

	marshaled, err := json.Marshal(defaultDeviceInterface)
	if err != nil {
		return false, fmt.Errorf("Failed to create JSON from configuration: %s", err.Error()), nil
	}

	// Create an entry in the store.
	DeviceStore[id] = defaultDevice

	// Create device in database.
	if err := database.CreateDevice(database.Device{
		DeviceType:    type_,
		Id:            id,
		Name:          name,
		RoomId:        roomID,
		VendorId:      driverVendorID,
		ModelId:       driverModelID,
		SingletonJSON: string(marshaled),
	}); err != nil {
		return false, nil, err
	}

	return true, nil, nil
}
