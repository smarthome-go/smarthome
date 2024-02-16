package homescript

import (
	"fmt"
	"strings"

	"github.com/smarthome-go/smarthome/core/database"
)

const DRIVER_ID_PREFIX = "@driver"

func CreateDriverHmsId(driver database.DriverTuple) string {
	return fmt.Sprintf("%s:%s:%s", DRIVER_ID_PREFIX, driver.VendorID, driver.ModelID)
}

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

func GetSources(username string, ids []string) (sources map[string]string, found bool, err error) {
	normalSources, found, err := database.GetHmsSources(username, ids)
	if err != nil {
		return nil, false, err
	}

	if !found {
		return nil, false, nil
	}

	// If the user has rights to modify and view drivers, also include drivers
	hasDriverPermission, err := database.UserHasPermission(username, database.PermissionSystemConfig)
	if err != nil {
		return nil, false, err
	}

	if !hasDriverPermission {
		return normalSources, true, nil
	}

	// Try to parse the driver ids
	remaining := make([]database.DriverTuple, 0)

	for _, id := range normalSources {
		// If this was already loaded, this is not a driver
		if _, alreadyLoaded := normalSources[id]; alreadyLoaded {
			continue
		}

		tuple, err := ParseHmsToDriver(id)
		if err != nil {
			return nil, false, nil
		}

		remaining = append(remaining, tuple)
	}

	drivers, allFound, err := database.GetDriverSources(remaining)
	if err != nil {
		return nil, false, err
	}

	if !allFound {
		return nil, false, nil
	}

	for driver, driverCode := range drivers {
		sources[CreateDriverHmsId(driver)] = driverCode
	}

	return sources, true, nil
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
					Id: CreateDriverHmsId(database.DriverTuple{
						VendorID: driver.VendorId,
						ModelID:  driver.ModelId,
					}),
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
