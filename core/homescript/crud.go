package homescript

import (
	"fmt"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

// Returns a Homescript given its id
// Returns Homescript, has been found, error
// This also includes drivers and other types of Homescript.
func (m *Manager) GetPersonalScriptById(homescriptID string, username string) (database.Homescript, bool, error) {
	homescripts, err := ListPersonal(username)
	if err != nil {
		logger.Error("Failed to get Homescript by id: ", err.Error())
		return database.Homescript{}, false, err
	}
	for _, homescriptItem := range homescripts {
		if homescriptItem.Data.Id == homescriptID {
			return homescriptItem, true, nil
		}
	}
	return database.Homescript{}, false, nil
}

func GetSources(username string, ids []string) (sources map[string]string, found bool, err error) {
	sources = make(map[string]string)

	rawSources, found, err := database.GetHmsSources(username, ids)
	if err != nil {
		return nil, false, err
	}

	// If the user has rights to modify and view drivers, also include drivers
	hasDriverPermission, err := database.UserHasPermission(username, database.PermissionSystemConfig)
	if err != nil {
		return nil, false, err
	}

	if !hasDriverPermission {
		return sources, found, nil
	}

	for key, val := range rawSources {
		sources[key] = val
	}

	// Try to parse the driver ids
	remaining := make([]database.DriverTuple, 0)

	for _, id := range ids {
		// If this was already loaded, this is not a driver
		if _, alreadyLoaded := sources[id]; alreadyLoaded {
			panic(remaining)
		}

		tuple, err := types.ParseHmsToDriver(id)
		if err != nil {
			return nil, false, err
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
		sources[types.CreateDriverHmsId(driver)] = driverCode
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
					Id: types.CreateDriverHmsId(database.DriverTuple{
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
