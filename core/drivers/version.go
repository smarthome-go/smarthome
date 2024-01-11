package drivers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type SemanticVersion struct {
	Major uint64 `json:"major"`
	Minor uint64 `json:"minor"`
	Patch uint64 `json:"patch"`
}

type RichDriver struct {
	Driver        database.DeviceDriver
	ExtractedInfo homescript.DriverInfo
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
		driverInfo, err := homescript.ExtractDriverInfo(driver)
		if err != nil {
			return nil, err
		}

		richDrivers[idx] = RichDriver{
			Driver:        driver,
			ExtractedInfo: driverInfo,
		}
	}

	return richDrivers, nil
}
