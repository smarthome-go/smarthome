package core

import (
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

// Modifies the code of a given Homescript.
// This function also handles dispatching to the correct storage backend, meaning
// that a driver script updates the driver and a normal script updates in the `homescripts` table.
func ModifyHomescriptCode(id string, owner string, newCode string) (found bool, err error) {
	// Determine whether this is a driver script or a normal script
	script, found, err := homescript.GetPersonalScriptById(id, owner)
	if err != nil {
		return false, err
	}

	if !found {
		return false, nil
	}

	switch script.Data.Type {
	case database.HOMESCRIPT_TYPE_NORMAL:
		if err := database.ModifyHomescriptCode(id, owner, newCode); err != nil {
			return false, err
		}
		return true, nil
	case database.HOMESCRIPT_TYPE_DRIVER:
		driverData, validationErr, dbErr := types.DriverFromHmsId(id)
		if dbErr != nil {
			return false, dbErr
		}

		if validationErr != nil {
			return false, validationErr
		}

		return driver.Manager.ModifyCode(driverData.VendorId, driverData.ModelId, newCode)
	default:
		panic(fmt.Sprintf("BUG warning: a new Homescript type was added without updating this code"))
	}
}
