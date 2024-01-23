package drivers

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type RichDriver struct {
	Driver        database.DeviceDriver `json:"driver"`
	ExtractedInfo homescript.DriverInfo `json:"info"`
	// TODO: implement something like this for device as well
	// Saves the persistent value(s) of the setting-fields of the `Driver` singleton.
	// If this field is `nil`, the user has not configured their driver yet.
	Configuration    interface{}             `json:"configuration"`
	IsValid          bool                    `json:"isValid"`
	ValidationErrors []diagnostic.Diagnostic `json:"validationErrors"`
}

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

func extractInfoFromDriver(driver database.DeviceDriver) (homescript.DriverInfo, []diagnostic.Diagnostic, error) {
	driverInfo, diagnostics, err := homescript.ExtractDriverInfoTotal(driver)
	if err != nil {
		return homescript.DriverInfo{}, nil, err
	}

	// Filter step: only include actual errors, not warnings and infos.
	filtered := make([]diagnostic.Diagnostic, 0)
	for _, diag := range diagnostics {
		if diag.Level == diagnostic.DiagnosticLevelError {
			filtered = append(filtered, diag)
		}
	}

	if len(filtered) > 0 {
		log.Tracef("Driver `%s:%s` is not working: `%s`", driver.VendorId, driver.ModelId, filtered[0].Message)
		return homescript.DriverInfo{}, filtered, nil
	}

	return driverInfo, make([]diagnostic.Diagnostic, 0), nil
}

func ListWithoutStoredValues() ([]RichDriver, error) {
	defaultDrivers, err := database.ListDeviceDrivers()
	if err != nil {
		return nil, err
	}

	richDrivers := make([]RichDriver, len(defaultDrivers))
	for idx, driver := range defaultDrivers {
		richDriver := RichDriver{
			Driver: driver,
			//nolint:exhaustruct
			ExtractedInfo:    homescript.DriverInfo{},
			Configuration:    nil,
			IsValid:          true,
			ValidationErrors: make([]diagnostic.Diagnostic, 0),
		}

		driverInfo, validationErrors, err := extractInfoFromDriver(driver)
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

func ListWithStoredConfig() ([]RichDriver, error) {
	drivers, err := ListWithoutStoredValues()
	if err != nil {
		return nil, err
	}

	for idx, driver := range drivers {
		if !driver.IsValid {
			log.Tracef("Skipping driver `%s:%s` in list with stored values: driver is not valid", driver.Driver.VendorId, driver.Driver.ModelId)
			continue
		}

		val, found := retrieveValueOfSingleton(
			DriverTuple{
				VendorID: driver.Driver.VendorId,
				ModelID:  driver.Driver.ModelId,
			},
			SingletonKindDriver,
		)

		// This should not happen: a zero value for every driver-spec is created automatically.
		if !found {
			panic(fmt.Sprintf("Configuration entry not found for driver `%s:%s`", driver.Driver.VendorId, driver.Driver.ModelId))
		}

		// TODO: deal with non-settings fields.

		configuration, _ := value.MarshalValue(
			filterObjFieldsWithoutSetting(val, driver.ExtractedInfo.DriverConfig.HmsType),
			errors.Span{},
			false,
		)

		drivers[idx].Configuration = configuration
	}

	return drivers, nil
}
