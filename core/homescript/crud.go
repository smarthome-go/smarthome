package homescript

import (
	"fmt"
	"strings"

	"github.com/smarthome-go/smarthome/core/database"
)

const DRIVER_ID_PREFIX = "@driver"

func CreateDriverHmsId(driver database.DeviceDriver) string {
	return fmt.Sprintf("%s:%s:%s", DRIVER_ID_PREFIX, driver.VendorId, driver.ModelId)
}

func DriverFromHmsId(id string) (driver database.DeviceDriver, validationErr error, databaseErr error) {
	delimeter := ":"
	split := strings.Split(id, delimeter)

	if len(split) != 3 {
		return database.DeviceDriver{}, fmt.Errorf("Expected 3 segments split by `%s`, found %d", delimeter, len(split)), nil
	}

	if split[0] != DRIVER_ID_PREFIX {
		return database.DeviceDriver{}, fmt.Errorf("Expected `%s`, found `%s`", DRIVER_ID_PREFIX, split[0]), nil
	}

	vendorId := split[1]
	modelId := split[2]

	driver, found, err := database.GetDeviceDriver(vendorId, modelId)
	if err != nil {
		return database.DeviceDriver{}, nil, err
	}

	if !found {
		return database.DeviceDriver{}, fmt.Errorf("Could not determine driver from HMS ID `%s`, driver `%s:%s` not found", id, vendorId, modelId), nil
	}

	return driver, nil, nil
}

// Returns a Homescript given its id
// Returns Homescript, has been found, error
// This also includes drivers and other types of Homescript.
func GetPersonalScriptById(homescriptId string, username string) (database.Homescript, bool, error) {
	homescripts, err := ListPersonal(username)
	if err != nil {
		log.Error("Failed to get Homescript by id: ", err.Error())
		return database.Homescript{}, false, err
	}
	for _, homescriptItem := range homescripts {
		if homescriptItem.Data.Id == homescriptId {
			return homescriptItem, true, nil
		}
	}
	return database.Homescript{}, false, nil
}

// Includes drivers and other types of Homescript.
func ListPersonal(username string) ([]database.Homescript, error) {
	base, err := database.ListHomescriptOfUser(username)
	if err != nil {
		return nil, err
	}

	// If the user has rights to modify and view drivers, also include drivers
	hasDriverPermission, err := database.UserHasPermission(username, database.PermissionSystemConfig)
	if err != nil {
		return nil, err
	}

	if hasDriverPermission {
		drivers, err := database.ListDeviceDrivers()
		if err != nil {
			return nil, err
		}

		for _, driver := range drivers {
			base = append(base, database.Homescript{
				Owner: "", // TODO: who owns this?
				Data: database.HomescriptData{
					Id:                  CreateDriverHmsId(driver),
					Name:                driver.Name,
					Description:         fmt.Sprintf("Hardware driver '%s'", driver.Name),
					QuickActionsEnabled: false,
					IsWidget:            false,
					SchedulerEnabled:    false,
					Code:                driver.HomescriptCode,
					MDIcon:              "code",     // TODO: what to do here
					Workspace:           "@drivers", // TODO: maybe just name it `Drivers` but disallow this id when changing a workspace
					Type:                database.HOMESCRIPT_TYPE_DRIVER,
				},
			})
		}
	}

	return base, nil
}

// Modifies the code of a given Homescript.
// This function also handles dispatching to the correct storage backend, meaning
// that a driver script updates the driver and a normal script updates in the `homescripts` table.
func ModifyHomescriptCode(id string, owner string, newCode string) (found bool, err error) {
	// Determine whether this is a driver script or a normal script
	script, found, err := GetPersonalScriptById(id, owner)
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
		driver, validationErr, dbErr := DriverFromHmsId(id)
		if dbErr != nil {
			return false, dbErr
		}

		if validationErr != nil {
			return false, validationErr
		}
		return database.ModifyDeviceDriverCode(driver.VendorId, driver.ModelId, newCode)
	default:
		panic(fmt.Sprintf("BUG warning: a new Homescript type was added without updating this code"))
	}
}
