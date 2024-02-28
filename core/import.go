package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/user"
)

var SetupPath = "./data/config/setup.json"

// Returns the setup struct, a bool that indicates that a setup file has been read and an error
func readSetupFile() (SetupStruct, bool, error) {
	log.Trace(fmt.Sprintf("Looking for setup file at `%s`", SetupPath))
	// Read file from `setupPath` on disk
	content, err := os.ReadFile(SetupPath)
	if err != nil {
		return SetupStruct{}, false, nil
	}
	// Move the file after a successful read
	if err := os.WriteFile(
		fmt.Sprintf("%s.old", SetupPath),
		content,
		0755,
	); err != nil {
		return SetupStruct{}, false, err
	}
	if err := os.Remove(SetupPath); err != nil {
		return SetupStruct{}, false, err
	}
	// Parse setup file to struct `Setup`
	var setupTemp SetupStruct
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&setupTemp); err != nil {
		log.Error(fmt.Sprintf("Failed to parse setup file at `%s` into setup struct: %s", SetupPath, err.Error()))
		return SetupStruct{}, false, err
	}
	return setupTemp, true, nil
}

// Used for setting up a Smarthome server quickly
// Reads a setup file at startup and starts functions that initialize those values in the database
func RunSetup() error {
	setup, fileDetected, err := readSetupFile()
	if err != nil {
		log.Error("Failed to run setup: ", err.Error())
		return err
	}
	if !fileDetected {
		log.Debug("No setup file detected: skipping setup")
		return nil
	}

	log.Debug(fmt.Sprintf("Setup file was detected. Moved setup file to `%s.old`.", SetupPath))
	return runSetupStruct(setup)
}

// In case something went wrong with the setup, a rescue user is created
// The rescue user has all permissions
func addRescueUser() error {
	log.Info("Creating rescue user: (username: `rescue`, password: `rescue`)")
	if err := database.AddUser(database.FullUser{
		Username: "rescue",
		Password: "rescue",
	}); err != nil {
		return err
	}
	return database.AddUserPermission("rescue", database.PermissionWildCard)
}

// Is executed if the setup runner fails
func abortSetup() error {
	// Delete database (remove junk from failed setup)
	if err := database.DeleteTables(); err != nil {
		return err
	}
	// Initialize database (fresh setup)
	if err := database.Init(
		GetConfig().Database,
		"rescue",
	); err != nil {
		return err
	}
	return addRescueUser()
}

func FactoryReset() error {
	// Shutdown the core
	serverConfig, err := getServerConfiguration()
	if err != nil {
		return err
	}

	// Shutdown the core
	if err := Shutdown(serverConfig); err != nil {
		log.Error("Setup failed, could not shutdown core: ", err.Error())
		return err
	}

	// Delete database first
	if err := database.DeleteTables(); err != nil {
		return err
	}
	// Initialize database (fresh setup)
	if err := database.Init(
		GetConfig().Database,
		"admin",
	); err != nil {
		return err
	}
	return nil
}

func getServerConfiguration() (database.ServerConfig, error) {
	config, found, err := database.GetServerConfiguration()
	if err != nil {
		return database.ServerConfig{}, err
	}

	if !found {
		return database.ServerConfig{}, errors.New("Server configuration was not found")
	}
	return config, nil
}

func RunSetupStruct(setup SetupStruct) error {
	if err := FactoryReset(); err != nil {
		return err
	}
	// Run the actual setup
	if err := runSetupStruct(setup); err != nil {
		if err2 := abortSetup(); err2 != nil {
			log.Error(fmt.Sprintf("Aborting setup failed: could not add rescue user: %s", err2.Error()))
		}
		return err
	}

	config, err := getServerConfiguration()
	if err != nil {
		return err
	}

	// Start the core again
	if err := Init(config); err != nil {
		log.Error("Setup failed, could not restart core: ", err.Error())
		return err
	}

	// Remove redundant power data points
	if err := hardware.SaveCurrentPowerUsage(); err != nil {
		return err
	}
	return nil
}

func runSetupStruct(setup SetupStruct) error {
	log.Info("Running configuration setup...")
	if err := createSystemConfigInDatabase(setup.ServerConfiguration); err != nil {
		log.Error("Aborting setup: could not update system configuration in database: ", err.Error())
		return err
	}
	// if err := createHardwareNodesInDatabase(setup.HardwareNodes); err != nil {
	// 	log.Error("Aborting setup: could not create hardware nodes in database: ", err.Error())
	// 	return err
	// }
	if err := createRoomsInDatabase(setup.Rooms); err != nil {
		log.Error("Aborting setup: could not create rooms in database: ", err.Error())
		return err
	}
	if err := createUsersInDatabase(setup.Users); err != nil {
		log.Error("Aborting setup: could not create users in database: ", err.Error())
		return err
	}
	if err := createCacheDataInDatabase(setup.CacheData); err != nil {
		log.Error("Aborting setup: could not create cache data in database: ", err.Error())
		return err
	}
	log.Info("Successfully ran setup`")
	return nil
}

// Takes the `users` slice and createas according database entries
func createUsersInDatabase(users []SetupUser) error {
	for _, usr := range users {
		_, alreadyExists, err := database.GetUserByUsername(usr.Data.Username)
		if err != nil {
			return err
		}
		// It is likely that an `admin` user already exists
		// In this case, the admin user is re-created using the new metadata
		if alreadyExists && usr.Data.Username != "admin" {
			return fmt.Errorf("cannot create user: user `%s` already exists", usr.Data.Username)
		}
		// Create the user itself
		if err := database.InsertUser(database.FullUser{
			Username:          usr.Data.Username,
			Forename:          usr.Data.Forename,
			Surname:           usr.Data.Surname,
			PrimaryColorDark:  usr.Data.PrimaryColorDark,
			PrimaryColorLight: usr.Data.PrimaryColorLight,
			Password:          usr.Data.Password,
		}); err != nil {
			return err
		}

		// Create the user's profile picture (if exported)
		if usr.ProfilePicture != nil {
			imgBytes, err := base64.StdEncoding.DecodeString(usr.ProfilePicture.B64Data)
			if err != nil {
				return err
			}
			if err := user.UploadAvatar(usr.Data.Username, fmt.Sprintf("imported.%s", usr.ProfilePicture.FileExtension), imgBytes); err != nil {
				return err
			}
		}

		// Setup the user's authentication tokens
		for _, token := range usr.Tokens {
			if err := database.InsertUserToken(token.Token, usr.Data.Username, token.Label); err != nil {
				return err
			}
		}

		// Setup the user's permissions
		for _, permission := range usr.Permissions {
			valid := false
			for _, perm := range database.Permissions {
				if string(perm.Permission) == permission {
					valid = true
				}
			}
			if !valid {
				return fmt.Errorf("cannot grant invalid permission: `%s` to user `%s`", permission, usr.Data.Username)
			}
			if _, err := user.AddPermission(
				usr.Data.Username,
				database.PermissionType(permission),
			); err != nil {
				return err
			}
		}

		// Setup the user's switch permissions
		for _, devPermission := range usr.DevicePermissions {
			_, found, err := database.GetDeviceById(devPermission)
			if err != nil {
				return err
			}
			if !found {
				return fmt.Errorf("cannot grant invalid device permission `%s` to user `%s`", devPermission, usr.Data.Username)
			}
			if _, err := database.AddUserDevicePermission(
				usr.Data.Username,
				devPermission,
			); err != nil {
				return err
			}
		}

		// Setup the user's camera permissions
		for _, camPermission := range usr.CameraPermissions {
			_, found, err := database.GetCameraById(camPermission)
			if err != nil {
				return err
			}
			if !found {
				return fmt.Errorf("cannot grant invalid camera permission `%s` to user `%s`", camPermission, usr.Data.Username)
			}
			if _, err := database.AddUserCameraPermission(
				usr.Data.Username,
				camPermission,
			); err != nil {
				return err
			}
		}

		// Setup the user's Homescripts
		// Current arguments are being used for checking preexistence of arguments
		argsDB, err := database.ListAllHomescriptArgsOfUser(usr.Data.Username)
		if err != nil {
			return err
		}

		for _, homescript := range usr.Homescripts {
			// This function is used as drivers and other types of script should not be included
			_, found, err := database.GetPersonalHomescriptById(
				homescript.Data.Id,
				usr.Data.Username,
			)
			if err != nil {
				return err
			}
			if found {
				return fmt.Errorf("cannot create Homescript: id `%s` is already taken", homescript.Data.Id)
			}
			if err := database.CreateNewHomescript(database.Homescript{
				Owner: usr.Data.Username,
				Data:  homescript.Data,
			}); err != nil {
				return err
			}

			// Create arguments of Homecript
			for _, arg := range homescript.Arguments {
				// Check if the argument to be inserted already exists
				argAlreadyExists := false
				for _, argCheck := range argsDB {
					if argCheck.Data.ArgKey == arg.ArgKey && argCheck.Data.HomescriptId == homescript.Data.Id {
						argAlreadyExists = true
					}
				}

				if argAlreadyExists {
					return fmt.Errorf("cannot create HMS arg: key `%s` is already taken for script `%s`", arg.ArgKey, homescript.Data.Id)
				}

				if _, err := database.AddHomescriptArg(database.HomescriptArgData{
					ArgKey:       arg.ArgKey,
					HomescriptId: homescript.Data.Id,
					Prompt:       arg.Prompt,
					MDIcon:       arg.MDIcon,
					InputType:    arg.InputType,
					Display:      arg.Display,
				}); err != nil {
					return err
				}
			}

			// Create automations using this Homescript
			for _, autom := range homescript.Automations {
				if _, err := database.CreateNewAutomation(database.Automation{
					Owner: usr.Data.Username,
					Data: database.AutomationData{
						Name:                   autom.Name,
						Description:            autom.Description,
						HomescriptId:           homescript.Data.Id,
						Enabled:                autom.Enabled,
						Trigger:                autom.Trigger,
						TriggerCronExpression:  autom.TriggerCronExpression,
						TriggerIntervalSeconds: autom.TriggerIntervalSeconds,
					},
				}); err != nil {
					return err
				}
			}
		}

		// Setup Homescript storage
		for _, storageItem := range usr.HomescriptStorage {
			if err := database.InsertHmsStorageEntry(usr.Data.Username, storageItem.Key, storageItem.Value); err != nil {
				return err
			}
		}

		// Setup the user's reminders
		for _, rem := range usr.Reminders {
			if _, err := database.CreateNewReminder(
				rem.Name,
				rem.Description,
				rem.DueDate,
				usr.Data.Username,
				rem.Priority,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// Takes the specified `rooms` and creates according database entries
func createRoomsInDatabase(rooms []SetupRoom) error {
	for _, room := range rooms {
		if err := database.CreateRoom(room.Data); err != nil {
			log.Error("Could not create rooms from setup file: ", err.Error())
			return err
		}

		for _, device := range room.Devices {
			if err := database.CreateDevice(database.ShallowDevice{
				DeviceType:    device.DeviceType,
				ID:            device.Id,
				Name:          device.Name,
				RoomID:        room.Data.ID,
				VendorID:      device.VendorId,
				ModelID:       device.ModelId,
				SingletonJSON: device.SingletonJSON,
			}); err != nil {
				log.Error("Could not create devices from setup file: ", err.Error())
				return err
			}
		}

		for _, camera := range room.Cameras {
			// Override the (possible) empty room-id to match the current room
			if err := database.CreateCamera(database.Camera{
				ID:     camera.Id,
				Name:   camera.Name,
				Url:    camera.Url,
				RoomID: room.Data.ID,
			}); err != nil {
				log.Error("Could not create cameras from setup file: ", err.Error())
				return err
			}
		}
	}

	return nil
}

// Takes the specified `hardwareNodes` and creates according database entries
// func createHardwareNodesInDatabase(nodes []config.SetupHardwareNode) error {
// 	for _, node := range nodes {
// 		if err := database.CreateHardwareNode(
// 			database.HardwareNode{
// 				Name:    node.Name,
// 				Url:     node.Url,
// 				Token:   node.Token,
// 				Enabled: node.Enabled,
// 			},
// 		); err != nil {
// 			log.Error("Could not create hardware nodes from setup file: ", err.Error())
// 			return err
// 		}
// 	}
// 	return nil
// }

// Takes the specified `systemConfig` and modifies an according database entry
func createSystemConfigInDatabase(systemConfig database.ServerConfig) error {
	if systemConfig.Latitude < -90 || systemConfig.Latitude > 90 {
		return fmt.Errorf("invalid latitude: must be (> -90 and < 90)")
	}
	if systemConfig.Longitude < -180 || systemConfig.Longitude > 180 {
		return fmt.Errorf("invalid longitude: must be (> -180 and < 180)")
	}
	if err := database.SetServerConfiguration(systemConfig); err != nil {
		log.Error("Could not create system configuration from setup file: ", err.Error())
		return err
	}
	return nil
}

func createCacheDataInDatabase(cacheData SetupCacheData) error {
	var wg sync.WaitGroup
	error := struct {
		err  error
		lock sync.Mutex
	}{
		err:  nil,
		lock: sync.Mutex{},
	}

	// Insert weather data
	if len(cacheData.WeatherHistory) > 0 {
		wg.Add(1)
		go func() {
			log.Trace("Importing weather data...")
			defer func() { wg.Done() }()
			for _, weatherEntry := range cacheData.WeatherHistory {
				timestamp := time.UnixMilli(int64(weatherEntry.Time)).Local()
				if _, err := database.AddWeatherDataRecord(
					weatherEntry.WeatherTitle,
					&timestamp,
					weatherEntry.WeatherDescription,
					weatherEntry.Temperature,
					weatherEntry.FeelsLike,
					weatherEntry.Humidity,
				); err != nil {
					error.lock.Lock()
					defer error.lock.Unlock()
					error.err = err
					return
				}
			}
			log.Debug("Successfully imported weather data")
		}()
	}

	// Insert power usage data
	if len(cacheData.PowerUsageData) > 0 {
		wg.Add(1)
		go func() {
			log.Trace("Importing power usage data...")
			defer func() { wg.Done() }()
			for _, pwr := range cacheData.PowerUsageData {
				if _, err := database.AddPowerUsagePoint(
					database.PowerDrawData{
						SwitchCount: pwr.On.SwitchCount,
						Watts:       pwr.On.Watts,
						Percent:     pwr.On.Percent,
					},
					database.PowerDrawData{
						SwitchCount: pwr.Off.SwitchCount,
						Watts:       pwr.Off.Watts,
						Percent:     pwr.Off.Percent,
					},
					time.UnixMilli(int64(pwr.Time)),
				); err != nil {
					error.lock.Lock()
					defer error.lock.Unlock()
					error.err = err
					return
				}
			}
			log.Debug("Successfully imported power data")
		}()
	}

	wg.Wait()

	return nil
}
