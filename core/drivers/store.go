package drivers

import (
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript"
)

type DriverTuple struct {
	VendorId string `json:"vendorId"`
	ModelId  string `json:"modelId"`
}

type DRIVER_SINGLETON_KIND uint8

const (
	DRIVER_SINGLETON_KIND_DRIVER DRIVER_SINGLETON_KIND = iota
	DRIVER_SINGLETON_KIND_DEVICE
)

// This package contains the storage backend implementation for per-driver / per-devicec configuration data.
//
//	1. The user sends a JSON configuration string
//	2. The HTTP layer parses this request and passes it into this module
//	3. This module performs sanity-checks on the JSON and tries to parse it into a Homescript value
//		- Here, the schema / type of the according singleton must be used to validate that the JSON has a valid schema.
//	4. If the parsing succeeded, store the data in: TODO any kind of database

func StoreValueInSingleton(file DriverTuple, targetSingleton DRIVER_SINGLETON_KIND, fromJson any) (found bool, softErr error, dbErr error) {
	driver, found, err := database.GetDeviceDriver(file.VendorId, file.ModelId)
	if err != nil {
		return false, nil, err
	}

	if !found {
		return false, nil, nil
	}

	_, hmsErrs, err := homescript.ExtractDriverInfoTotal(driver)
	if err != nil {
		return false, nil, err
	}

	if len(hmsErrs) > 0 {
		return false, fmt.Errorf("Could not extract driver information for sanity check: %s", hmsErrs[0].Display(driver.HomescriptCode)), nil
	}

	fmt.Printf("storing: %v in targt singleton %d in file %s:%s...\n", fromJson, targetSingleton, file.VendorId, file.ModelId)

	// info.DriverConfig

	return true, nil, nil
}
