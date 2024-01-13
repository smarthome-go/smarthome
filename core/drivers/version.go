package drivers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type SemanticVersion struct {
	Major uint64 `json:"major"`
	Minor uint64 `json:"minor"`
	Patch uint64 `json:"patch"`
}

type RichDriver struct {
	Driver           database.DeviceDriver   `json:"driver"`
	ExtractedInfo    homescript.DriverInfo   `json:"info"`
	IsValid          bool                    `json:"isValid"`
	ValidationErrors []diagnostic.Diagnostic `json:"validationErrors"`
}

func ParseDriverVersion(source string) (SemanticVersion, error) {
	delimeter := "."

	split := strings.Split(source, delimeter)
	if len(split) != 3 {
		return SemanticVersion{}, fmt.Errorf("Expected exactly 3 version components, got %d", len(split))
	}

	parsedValues := make([]uint64, 3)
	for idx, element := range split {
		parsed, err := strconv.ParseUint(element, 10, 64)
		if err != nil {
			return SemanticVersion{}, err
		}

		parsedValues[idx] = parsed
	}

	return SemanticVersion{}, nil
}

func List() ([]RichDriver, error) {
	defaultDrivers, err := database.ListDeviceDrivers()
	if err != nil {
		return nil, err
	}

	richDrivers := make([]RichDriver, len(defaultDrivers))
	for idx, driver := range defaultDrivers {
		richDriver := RichDriver{
			Driver:           driver,
			IsValid:          true,
			ValidationErrors: make([]diagnostic.Diagnostic, 0),
		}

		driverInfo, diagnostics, err := homescript.ExtractDriverInfoTotal(driver)
		if err != nil {
			return nil, err
		}

		// Filter step: only include actual errors, not warnings and infos
		filtered := make([]diagnostic.Diagnostic, 0)
		for _, diag := range diagnostics {
			if diag.Level == diagnostic.DiagnosticLevelError {
				filtered = append(filtered, diag)
			}
		}

		if len(filtered) > 0 {
			richDriver.IsValid = false
			richDriver.ValidationErrors = filtered

			log.Tracef("Driver `%s:%s` is not working: `%s`", driver.VendorId, driver.ModelId, filtered[0].Message)
		}

		richDriver.ExtractedInfo = driverInfo
		richDrivers[idx] = richDriver
	}

	return richDrivers, nil
}
