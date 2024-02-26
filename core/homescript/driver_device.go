package homescript

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
)

type Device struct {
	DeviceType     database.DEVICE_TYPE    `json:"type"`
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	RoomID         string                  `json:"roomId"`
	DriverVendorID string                  `json:"vendorId"`
	DriverModelID  string                  `json:"modelId"`
	SingletonJSON  any                     `json:"singletonJson"`
	HmsErrors      []HmsError              `json:"hmsErrors"`
	Config         ConfigInfoWrapperDevice `json:"config"`

	// Device-specific information.
	PowerInformation    DevicePowerInformation                   `json:"powerInformation"`
	DimmableInformation []DriverActionReportDimOutput            `json:"dimmables"`
	SensorReadings      []DriverActionReportSensorReadingsOutput `json:"sensors"`
}

type DevicePowerInformation struct {
	State          bool `json:"state"`
	PowerDrawWatts uint `json:"powerDrawWatts"`
}

// type DeviceDimmableInformation struct {
// 	Percent uint8 `json:"percent"`
// }

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

		hmsErrors := HmsErrorsFromDiagnostics(fittingDriver.ValidationErrors)
		storedDeviceValue := DeviceStore[device.Id]

		savedConfig, _ := value.MarshalValue(
			filterObjFieldsWithoutSetting(storedDeviceValue, fittingDriver.ExtractedInfo.DeviceConfig.Info.HmsType),
			false,
		)

		// Extract additional information by invoking driver function code.
		// TODO: a hot / ready / precompiled VM instance would lead to additional performance gains here.
		// TODO: fuse these
		var powerStateInfo DriverActionGetPowerStateOutput
		var powerDrawInfo DriverActionGetPowerDrawOutput
		if fittingDriver.DeviceSupports(DeviceCapabilityPower) {
			//
			// Power state
			//
			powerStateTemp, hmsErrs, err := InvokeDriverReportPowerState(
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

			powerStateInfo = powerStateTemp

			//
			// Power draw
			//
			powerDrawTemp, hmsErrs, err := InvokeDriverReportPowerDraw(
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

			powerDrawInfo = powerDrawTemp
		}

		var dimmableInformation []DriverActionReportDimOutput
		if fittingDriver.DeviceSupports(DeviceCapabilityDimmable) {
			dimmableInformationTemp, hmsErrs, err := InvokeDriverReportDimmable(
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

			dimmableInformation = dimmableInformationTemp
		}

		var sensorReadings []DriverActionReportSensorReadingsOutput
		if fittingDriver.DeviceSupports(DeviceCapabilitySensor) {
			readingsTemp, hmsErrs, err := InvokeDriverReportSensors(
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

			sensorReadings = readingsTemp
		}

		output[index] = Device{
			DeviceType:     device.DeviceType,
			ID:             device.Id,
			Name:           device.Name,
			RoomID:         device.RoomId,
			DriverVendorID: device.VendorId,
			DriverModelID:  device.ModelId,
			SingletonJSON:  savedConfig,
			HmsErrors:      hmsErrors,
			Config:         fittingDriver.ExtractedInfo.DeviceConfig,
			PowerInformation: DevicePowerInformation{
				State:          powerStateInfo.State,
				PowerDrawWatts: powerDrawInfo.Watts,
			},
			DimmableInformation: dimmableInformation,
			SensorReadings:      sensorReadings,
		}
	}

	// TODO: maybe remove this?
	// Sort output by number of capabilities.
	slices.SortFunc[[]Device](output, func(a Device, b Device) int {
		aLen := len(a.Config.Capabilities)
		bLen := len(b.Config.Capabilities)

		if aLen < bLen {
			return 1
		}

		if aLen > bLen {
			return -1
		}

		return 0
	})

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

func SetDevicePower(deviceId string, power bool) (output DriverActionPowerOutput, deviceFound bool, hmsErr *HmsError, err error) {
	switchData, found, err := database.GetDeviceById(deviceId)
	if err != nil {
		return DriverActionPowerOutput{}, false, nil, err
	}

	if !found {
		return DriverActionPowerOutput{}, false, nil, nil
	}

	output, hmsErrs, err := InvokeDriverSetPower(
		deviceId,
		switchData.VendorId,
		switchData.ModelId,
		DriverActionPower{State: power},
	)

	if err != nil {
		return DriverActionPowerOutput{}, false, nil, err
	}

	if hmsErrs != nil {
		return DriverActionPowerOutput{}, false, &hmsErrs[0], nil
	}

	return output, true, nil, nil
}

func SetDeviceDim(deviceId string, function string, value int64) (output DriverActionDimOutput, deviceFound bool, hmsErr *HmsError, err error) {
	switchData, found, err := database.GetDeviceById(deviceId)
	if err != nil {
		return DriverActionDimOutput{}, false, nil, err
	}

	if !found {
		return DriverActionDimOutput{}, false, nil, nil
	}

	output, hmsErrs, err := InvokeDriverDim(
		deviceId,
		switchData.VendorId,
		switchData.ModelId,
		DriverActionDim{
			Value: value,
			Label: function,
		},
	)

	if err != nil {
		return DriverActionDimOutput{}, false, nil, err
	}

	if hmsErrs != nil {
		return DriverActionDimOutput{}, false, &hmsErrs[0], nil
	}

	return output, true, nil, nil
}
