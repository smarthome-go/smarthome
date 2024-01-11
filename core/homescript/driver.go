package homescript

import (
	"errors"
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/smarthome/core/database"
)

const DRIVER_SINGLETON_IDENT = "@Driver"
const DRIVER_DEVICE_IDENT = "@Device"

// TODO: fill this
type DriverInfo struct {
	fields []ConfigField
}

type CONFIG_FIELD_TYPE uint8

const (
	CONFIG_FIELD_TYPE_INT CONFIG_FIELD_TYPE = iota
	CONFIG_FIELD_TYPE_FLOAT
	CONFIG_FIELD_TYPE_BOOL
	CONFIG_FIELD_TYPE_STRING
	CONFIG_FIELD_TYPE_LIST
	CONFIG_FIELD_TYPE_STRUCT
)

type ConfigField interface {
	Type() CONFIG_FIELD_TYPE
}

type ConfigFieldString struct {
	Value string
}

func (self ConfigFieldString) Kind() CONFIG_FIELD_TYPE {
	return CONFIG_FIELD_TYPE_STRING
}

func ExtractDriverInfo(driver database.DeviceDriver) (DriverInfo, error) {
	filename := fmt.Sprintf("@driver_%s:%s", driver.VendorId, driver.ModelId)

	analyzed, res, err := HmsManager.Analyze(
		"", // TODO: what to do with this field??
		filename,
		driver.HomescriptCode,
		HMS_PROGRAM_KIND_DEVICE_DRIVER,
	)
	if err != nil {
		return DriverInfo{}, err
	}

	if !res.Success || len(res.Errors) != 0 {
		err0 := res.Errors[0]
		return DriverInfo{}, errors.New(err0.String())
	}

	driverSingleton, driverSingletonFound := ast.AnalyzedSingletonTypeDefinition{}, false
	deviceSingleton, deviceSingletonFound := ast.AnalyzedSingletonTypeDefinition{}, false

	// Iterate over singletons, assert that there is a `driver` singleton
	for _, singleton := range analyzed[filename].Singletons {
		if singleton.Ident.Ident() == DRIVER_SINGLETON_IDENT {
			driverSingleton = singleton
			driverSingletonFound = true
			continue
		}

		if singleton.Ident.Ident() == DRIVER_DEVICE_IDENT {
			deviceSingleton = singleton
			deviceSingletonFound = true
			continue
		}
	}

	return DriverInfo{}, nil
}
