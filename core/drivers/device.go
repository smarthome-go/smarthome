package drivers

import (
	"encoding/json"
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type Device struct {
	DeviceType     database.DEVICE_TYPE               `json:"type"`
	ID             string                             `json:"id"`
	Name           string                             `json:"name"`
	RoomID         string                             `json:"roomId"`
	DriverVendorID string                             `json:"vendorId"`
	DriverModelID  string                             `json:"modelId"`
	SingletonJSON  any                                `json:"singletonJson"`
	HmsErrors      []homescript.HmsError              `json:"hmsErrors"`
	Config         homescript.ConfigInfoWrapperDevice `json:"config"`

	// Device-specific information.
	PowerInformation DevicePowerInformation `json:"powerInformation"`
}

type DevicePowerInformation struct {
	State          bool `json:"state"`
	PowerDrawWatts uint `json:"powerDrawWatts"`
}

func ListAllDevices() ([]Device, error) {
	raw, err := database.ListAllDevices()
	if err != nil {
		return nil, err
	}

	return EnrichDevicesList(raw)
}

func ListPersonalDevices(username string) ([]Device, error) {
	raw, err := database.ListUserDevices(username)
	if err != nil {
		return nil, err
	}

	return EnrichDevicesList(raw)
}

func EnrichDevicesList(input []database.Device) ([]Device, error) {
	drivers, err := ListDriversWithoutStoredValues()
	if err != nil {
		return nil, err
	}

	output := make([]Device, len(input))
	for index, device := range input {
		// Find correct driver.
		var fittingDriver RichDriver

		for _, driver := range drivers {
			if driver.Driver.VendorId == device.VendorId && driver.Driver.ModelId == device.ModelId {
				fittingDriver = driver
				break
			}
		}

		hmsErrors := homescript.HmsErrorsFromDiagnostics(fittingDriver.ValidationErrors)
		storedDeviceValue := DeviceStore[device.Id]

		savedConfig, _ := value.MarshalValue(
			filterObjFieldsWithoutSetting(storedDeviceValue, fittingDriver.ExtractedInfo.DeviceConfig.Info.HmsType),
			false,
		)

		// Extract additional information by invoking driver function code.
		// TODO: a hot / ready / precompiled VM instance would lead to additional performance gains here.

		// TODO: fuse these
		powerDraw, hmsErrs, err := InvokeDriverReportPowerDraw(
			DriverInvocationIDs{
				deviceID: device.Id, vendorID: device.VendorId, modelID: device.ModelId,
			},
		)
		if err != nil {
			return nil, err
		}
		if hmsErrs != nil {
			hmsErrors = append(hmsErrors, hmsErrs...)
		}

		powerState, hmsErrs, err := InvokeDriverReportPowerState(
			DriverInvocationIDs{
				deviceID: device.Id,
				vendorID: device.VendorId,
				modelID:  device.ModelId,
			},
		)
		if err != nil {
			return nil, err
		}
		if hmsErrs != nil {
			hmsErrors = append(hmsErrors, hmsErrs...)
		}

		output[index] = Device{
			DeviceType:     device.DeviceType,
			ID:             device.Id,
			Name:           device.Name,
			RoomID:         device.RoomId,
			DriverVendorID: device.VendorId,
			DriverModelID:  device.ModelId,
			SingletonJSON:  savedConfig,
			HmsErrors:      hmsErrs,
			Config:         fittingDriver.ExtractedInfo.DeviceConfig,
			PowerInformation: DevicePowerInformation{
				State:          powerState.State,
				PowerDrawWatts: powerDraw.Watts,
			},
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
	defaultDevice := value.ObjectZeroValue(driverInfo.DeviceConfig.Info.HmsType)
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
