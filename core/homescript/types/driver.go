package types

import (
	"fmt"
	"strings"

	"github.com/smarthome-go/smarthome/core/database"
)

type AnalyzerDriverMetadata struct {
	VendorID string
	ModelID  string
}

const DRIVER_ID_PREFIX = "@driver"

func ParseHmsToDriver(id string) (driver database.DriverTuple, validationErr error) {
	delimiter := ":"
	split := strings.Split(id, delimiter)

	if len(split) != 3 {
		return database.DriverTuple{}, fmt.Errorf("Expected 3 segments split by `%s`, found %d", delimiter, len(split))
	}

	if split[0] != DRIVER_ID_PREFIX {
		return database.DriverTuple{}, fmt.Errorf("Expected `%s`, found `%s`", DRIVER_ID_PREFIX, split[0])
	}

	vendorId := split[1]
	modelId := split[2]

	return database.DriverTuple{
		VendorID: vendorId,
		ModelID:  modelId,
	}, nil
}

func DriverFromHmsId(id string) (driver database.DeviceDriver, validationErr error, databaseErr error) {
	tuple, err := ParseHmsToDriver(id)
	if err != nil {
		return driver, err, nil
	}

	driver, found, err := database.GetDeviceDriver(tuple.VendorID, tuple.ModelID)
	if err != nil {
		return database.DeviceDriver{}, nil, err
	}

	if !found {
		return database.DeviceDriver{},
			fmt.Errorf(
				"Could not determine driver from HMS ID `%s`, driver `%s:%s` not found",
				id,
				tuple.VendorID,
				tuple.ModelID,
			), nil
	}

	return driver, nil, nil
}

func CreateDriverHmsId(driver database.DriverTuple) string {
	return fmt.Sprintf("%s:%s:%s", DRIVER_ID_PREFIX, driver.VendorID, driver.ModelID)
}
