package homescript

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
)

type ShallowDevice struct {
	DeviceType     database.DEVICE_TYPE `json:"type"`
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	RoomID         string               `json:"roomId"`
	DriverVendorID string               `json:"vendorId"`
	DriverModelID  string               `json:"modelId"`
	SingletonJSON  any                  `json:"singletonJson"`
}

type RichDevice struct {
	Shallow     ShallowDevice     `json:"shallow"`
	Extractions DeviceExtractions `json:"extractions"`
}

type DeviceExtractions struct {
	HmsErrors []HmsError              `json:"hmsErrors"`
	Config    ConfigInfoWrapperDevice `json:"config"`

	// Device-specific information.
	PowerInformation    DevicePowerInformation                   `json:"powerInformation"`
	DimmableInformation []DriverActionReportDimOutput            `json:"dimmables"`
	SensorReadings      []DriverActionReportSensorReadingsOutput `json:"sensors"`
}

type DevicePowerInformation struct {
	State          bool `json:"state"`
	PowerDrawWatts uint `json:"powerDrawWatts"`
}

func shallowSorter(input *[]database.ShallowDevice) error {
	needsRebuild := false

	for _, dev := range *input {
		_, found := CachedDriverMeta[database.DriverTuple{
			VendorID: dev.VendorID,
			ModelID:  dev.ModelID,
		}]

		if !found {
			needsRebuild = true
			break
		}
	}

	if needsRebuild {
		logger.Trace("Driver cache outdated, needs rebuild before list.")
		if err := RebuildCache(); err != nil {
			return err
		}
	}

	slices.SortFunc[[]database.ShallowDevice](*input, func(_a database.ShallowDevice, _b database.ShallowDevice) int {
		a := CachedDriverMeta[database.DriverTuple{
			VendorID: _a.VendorID,
			ModelID:  _a.ModelID,
		}]

		b := CachedDriverMeta[database.DriverTuple{
			VendorID: _b.VendorID,
			ModelID:  _b.ModelID,
		}]

		return deviceSorter(a.DeviceConfig, b.DeviceConfig)
	})

	return nil
}

func deviceSorter(a, b ConfigInfoWrapperDevice) int {
	aLen := len(a.Capabilities)
	bLen := len(b.Capabilities)

	if aLen < bLen {
		return 1
	}

	if aLen > bLen {
		return -1
	}

	if a.Capabilities.Has(DeviceCapabilityDimmable) && !b.Capabilities.Has(DeviceCapabilityDimmable) {
		return -1
	} else if !a.Capabilities.Has(DeviceCapabilityDimmable) && b.Capabilities.Has(DeviceCapabilityDimmable) {
		return 1
	}

	return 0
}

func ListAllDevicesShallow() ([]database.ShallowDevice, error) {
	unsorted, err := database.ListAllDevices()
	if err != nil {
		return nil, err
	}

	if err := shallowSorter(&unsorted); err != nil {
		return nil, err
	}

	return unsorted, nil
}

func ListPersonalDevicesShallow(username string) ([]database.ShallowDevice, error) {
	unsorted, err := database.ListUserDevices(username)
	if err != nil {
		return nil, err
	}

	old := ""

	for _, e := range unsorted {
		old += fmt.Sprintf("%s\n", e.ID)
	}

	if err := shallowSorter(&unsorted); err != nil {
		return nil, err
	}

	fmt.Printf("%s\n=====\n", old)

	for _, e := range unsorted {
		fmt.Printf("%s\n", e.ID)
	}

	return unsorted, nil
}

func ListAllDevicesRich() ([]RichDevice, error) {
	raw, err := database.ListAllDevices()
	if err != nil {
		return nil, err
	}

	return EnrichDevicesList(raw)
}

func ListPersonalDevicesRich(username string) ([]RichDevice, error) {
	raw, err := database.ListUserDevices(username)
	if err != nil {
		return nil, err
	}

	return EnrichDevicesList(raw)
}

func EnrichDeviceAll(deviceID string) (RichDevice, bool, error) {
	device, found, err := database.GetDeviceById(deviceID)
	if err != nil || !found {
		return RichDevice{}, found, err
	}

	driver, found, err := GetDriverWithInfos(device.VendorID, device.ModelID)
	if err != nil || !found {
		return RichDevice{}, found, err
	}

	richDevice, err := EnrichDevice(device, driver)
	return richDevice, true, err
}

func EnrichDevice(device database.ShallowDevice, fittingDriver RichDriver) (RichDevice, error) {
	hmsErrors := HmsErrorsFromDiagnostics(fittingDriver.ValidationErrors)

	storedDeviceValue := DeviceStore[device.ID]
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
				deviceID: device.ID,
				vendorID: device.VendorID,
				modelID:  device.ModelID,
			},
		)
		if err != nil {
			return RichDevice{}, err
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
				deviceID: device.ID, vendorID: device.VendorID, modelID: device.ModelID,
			},
		)
		if err != nil {
			return RichDevice{}, err
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
				deviceID: device.ID,
				vendorID: device.VendorID,
				modelID:  device.ModelID,
			},
		)
		if err != nil {
			return RichDevice{}, err
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
				deviceID: device.ID,
				vendorID: device.VendorID,
				modelID:  device.ModelID,
			},
		)
		if err != nil {
			return RichDevice{}, err
		}
		if hmsErrs != nil {
			hmsErrors = append(hmsErrors, hmsErrs...)
		}

		sensorReadings = readingsTemp
	}

	return RichDevice{
		Shallow: ShallowDevice{
			DeviceType:     device.DeviceType,
			ID:             device.ID,
			Name:           device.Name,
			RoomID:         device.RoomID,
			DriverVendorID: device.VendorID,
			DriverModelID:  device.ModelID,
			SingletonJSON:  savedConfig,
		},
		Extractions: DeviceExtractions{
			HmsErrors: hmsErrors,
			Config:    fittingDriver.ExtractedInfo.DeviceConfig,
			PowerInformation: DevicePowerInformation{
				State:          powerStateInfo.State,
				PowerDrawWatts: powerDrawInfo.Watts,
			},
			DimmableInformation: dimmableInformation,
			SensorReadings:      sensorReadings,
		},
	}, nil
}

// func EnrichDevicesList(input []database.ShallowDevice) ([]RichDevice, error) {
// 	drivers, err := ListDriversWithoutStoredValues()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	output := make([]RichDevice, len(input))
// 	for index, device := range input {
// 		// Find correct driver.
// 		var fittingDriver RichDriver
//
// 		for _, driver := range drivers {
// 			if driver.Driver.VendorId == device.VendorID && driver.Driver.ModelId == device.ModelID {
// 				fittingDriver = driver
// 				break
// 			}
// 		}
//
// 		hmsErrors := HmsErrorsFromDiagnostics(fittingDriver.ValidationErrors)
//
// 		storedDeviceValue := DeviceStore[device.ID]
// 		savedConfig, _ := value.MarshalValue(
// 			filterObjFieldsWithoutSetting(storedDeviceValue, fittingDriver.ExtractedInfo.DeviceConfig.Info.HmsType),
// 			false,
// 		)
//
// 		// Extract additional information by invoking driver function code.
// 		// TODO: a hot / ready / precompiled VM instance would lead to additional performance gains here.
// 		// TODO: fuse these
// 		var powerStateInfo DriverActionGetPowerStateOutput
// 		var powerDrawInfo DriverActionGetPowerDrawOutput
// 		if fittingDriver.DeviceSupports(DeviceCapabilityPower) {
// 			//
// 			// Power state
// 			//
// 			powerStateTemp, hmsErrs, err := InvokeDriverReportPowerState(
// 				DriverInvocationIDs{
// 					deviceID: device.ID,
// 					vendorID: device.VendorID,
// 					modelID:  device.ModelID,
// 				},
// 			)
// 			if err != nil {
// 				return nil, err
// 			}
// 			if hmsErrs != nil {
// 				hmsErrors = append(hmsErrors, hmsErrs...)
// 			}
//
// 			powerStateInfo = powerStateTemp
//
// 			//
// 			// Power draw
// 			//
// 			powerDrawTemp, hmsErrs, err := InvokeDriverReportPowerDraw(
// 				DriverInvocationIDs{
// 					deviceID: device.ID, vendorID: device.VendorID, modelID: device.ModelID,
// 				},
// 			)
// 			if err != nil {
// 				return nil, err
// 			}
// 			if hmsErrs != nil {
// 				hmsErrors = append(hmsErrors, hmsErrs...)
// 			}
//
// 			powerDrawInfo = powerDrawTemp
// 		}
//
// 		var dimmableInformation []DriverActionReportDimOutput
// 		if fittingDriver.DeviceSupports(DeviceCapabilityDimmable) {
// 			dimmableInformationTemp, hmsErrs, err := InvokeDriverReportDimmable(
// 				DriverInvocationIDs{
// 					deviceID: device.ID,
// 					vendorID: device.VendorID,
// 					modelID:  device.ModelID,
// 				},
// 			)
// 			if err != nil {
// 				return nil, err
// 			}
// 			if hmsErrs != nil {
// 				hmsErrors = append(hmsErrors, hmsErrs...)
// 			}
//
// 			dimmableInformation = dimmableInformationTemp
// 		}
//
// 		var sensorReadings []DriverActionReportSensorReadingsOutput
// 		if fittingDriver.DeviceSupports(DeviceCapabilitySensor) {
// 			readingsTemp, hmsErrs, err := InvokeDriverReportSensors(
// 				DriverInvocationIDs{
// 					deviceID: device.ID,
// 					vendorID: device.VendorID,
// 					modelID:  device.ModelID,
// 				},
// 			)
// 			if err != nil {
// 				return nil, err
// 			}
// 			if hmsErrs != nil {
// 				hmsErrors = append(hmsErrors, hmsErrs...)
// 			}
//
// 			sensorReadings = readingsTemp
// 		}
//
// 		output[index] = RichDevice{
// 			Shallow: ShallowDevice{
// 				DeviceType:     device.DeviceType,
// 				ID:             device.ID,
// 				Name:           device.Name,
// 				RoomID:         device.RoomID,
// 				DriverVendorID: device.VendorID,
// 				DriverModelID:  device.ModelID,
// 				SingletonJSON:  savedConfig,
// 			},
// 			Extractions: DeviceExtractions{
// 				HmsErrors: hmsErrors,
// 				Config:    fittingDriver.ExtractedInfo.DeviceConfig,
// 				PowerInformation: DevicePowerInformation{
// 					State:          powerStateInfo.State,
// 					PowerDrawWatts: powerDrawInfo.Watts,
// 				},
// 				DimmableInformation: dimmableInformation,
// 				SensorReadings:      sensorReadings,
// 			},
// 		}
// 	}
//
// 	// TODO: maybe remove this?
// 	// Sort output by number of capabilities.
// 	slices.SortFunc[[]RichDevice](output, func(a RichDevice, b RichDevice) int {
// 		aLen := len(a.Extractions.Config.Capabilities)
// 		bLen := len(b.Extractions.Config.Capabilities)
//
// 		if aLen < bLen {
// 			return 1
// 		}
//
// 		if aLen > bLen {
// 			return -1
// 		}
//
// 		return 0
// 	})
//
// 	return output, nil
// }

var CachedDriverMeta map[database.DriverTuple]DriverInfo = make(map[database.DriverTuple]DriverInfo)

// TODO: only run this function on drivers which actually changed
func RebuildCache() error {
	logger.Debug("Rebuilding homescript driver metadata cache...")
	drivers, err := ListDriversWithoutStoredValues()
	if err != nil {
		return err
	}

	for _, driver := range drivers {
		CachedDriverMeta[database.DriverTuple{
			VendorID: driver.Driver.VendorId,
			ModelID:  driver.Driver.ModelId,
		}] = driver.ExtractedInfo
	}

	return nil
}

func EnrichDevicesList(input []database.ShallowDevice) ([]RichDevice, error) {
	drivers, err := ListDriversWithoutStoredValues()
	if err != nil {
		return nil, err
	}

	output := make([]RichDevice, len(input))
	for index, device := range input {
		// Find correct driver.
		var fittingDriver RichDriver

		for _, driver := range drivers {
			if driver.Driver.VendorId == device.VendorID && driver.Driver.ModelId == device.ModelID {
				fittingDriver = driver
				break
			}
		}

		enriched, err := EnrichDevice(device, fittingDriver)
		if err != nil {
			return nil, err
		}

		output[index] = enriched

		// hmsErrors := HmsErrorsFromDiagnostics(fittingDriver.ValidationErrors)
		//
		// storedDeviceValue := DeviceStore[device.ID]
		// savedConfig, _ := value.MarshalValue(
		// 	filterObjFieldsWithoutSetting(storedDeviceValue, fittingDriver.ExtractedInfo.DeviceConfig.Info.HmsType),
		// 	false,
		// )
		//
		// // Extract additional information by invoking driver function code.
		// // TODO: a hot / ready / precompiled VM instance would lead to additional performance gains here.
		// // TODO: fuse these
		// var powerStateInfo DriverActionGetPowerStateOutput
		// var powerDrawInfo DriverActionGetPowerDrawOutput
		// if fittingDriver.DeviceSupports(DeviceCapabilityPower) {
		// 	//
		// 	// Power state
		// 	//
		// 	powerStateTemp, hmsErrs, err := InvokeDriverReportPowerState(
		// 		DriverInvocationIDs{
		// 			deviceID: device.ID,
		// 			vendorID: device.VendorID,
		// 			modelID:  device.ModelID,
		// 		},
		// 	)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	if hmsErrs != nil {
		// 		hmsErrors = append(hmsErrors, hmsErrs...)
		// 	}
		//
		// 	powerStateInfo = powerStateTemp
		//
		// 	//
		// 	// Power draw
		// 	//
		// 	powerDrawTemp, hmsErrs, err := InvokeDriverReportPowerDraw(
		// 		DriverInvocationIDs{
		// 			deviceID: device.ID, vendorID: device.VendorID, modelID: device.ModelID,
		// 		},
		// 	)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	if hmsErrs != nil {
		// 		hmsErrors = append(hmsErrors, hmsErrs...)
		// 	}
		//
		// 	powerDrawInfo = powerDrawTemp
		// }
		//
		// var dimmableInformation []DriverActionReportDimOutput
		// if fittingDriver.DeviceSupports(DeviceCapabilityDimmable) {
		// 	dimmableInformationTemp, hmsErrs, err := InvokeDriverReportDimmable(
		// 		DriverInvocationIDs{
		// 			deviceID: device.ID,
		// 			vendorID: device.VendorID,
		// 			modelID:  device.ModelID,
		// 		},
		// 	)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	if hmsErrs != nil {
		// 		hmsErrors = append(hmsErrors, hmsErrs...)
		// 	}
		//
		// 	dimmableInformation = dimmableInformationTemp
		// }
		//
		// var sensorReadings []DriverActionReportSensorReadingsOutput
		// if fittingDriver.DeviceSupports(DeviceCapabilitySensor) {
		// 	readingsTemp, hmsErrs, err := InvokeDriverReportSensors(
		// 		DriverInvocationIDs{
		// 			deviceID: device.ID,
		// 			vendorID: device.VendorID,
		// 			modelID:  device.ModelID,
		// 		},
		// 	)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	if hmsErrs != nil {
		// 		hmsErrors = append(hmsErrors, hmsErrs...)
		// 	}
		//
		// 	sensorReadings = readingsTemp
		// }
		//
		// output[index] = RichDevice{
		// 	Shallow: ShallowDevice{
		// 		DeviceType:     device.DeviceType,
		// 		ID:             device.ID,
		// 		Name:           device.Name,
		// 		RoomID:         device.RoomID,
		// 		DriverVendorID: device.VendorID,
		// 		DriverModelID:  device.ModelID,
		// 		SingletonJSON:  savedConfig,
		// 	},
		// 	Extractions: DeviceExtractions{
		// 		HmsErrors: hmsErrors,
		// 		Config:    fittingDriver.ExtractedInfo.DeviceConfig,
		// 		PowerInformation: DevicePowerInformation{
		// 			State:          powerStateInfo.State,
		// 			PowerDrawWatts: powerDrawInfo.Watts,
		// 		},
		// 		DimmableInformation: dimmableInformation,
		// 		SensorReadings:      sensorReadings,
		// 	},
		// }
	}

	// TODO: maybe remove this?
	// Sort output by number of capabilities.
	slices.SortFunc[[]RichDevice](output, func(a RichDevice, b RichDevice) int {
		return deviceSorter(a.Extractions.Config, b.Extractions.Config)
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
	if err := database.CreateDevice(database.ShallowDevice{
		DeviceType:    type_,
		ID:            id,
		Name:          name,
		RoomID:        roomID,
		VendorID:      driverVendorID,
		ModelID:       driverModelID,
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
		switchData.VendorID,
		switchData.ModelID,
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
		switchData.VendorID,
		switchData.ModelID,
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
