package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/user"
)

var setupPath = "./data/config/setup.json"

// TODO: add some sort of web import / export later
// Returns the setup struct, a bool that indicates that a setup file has been read and an error
func readSetupFile() (SetupStruct, bool, error) {
	log.Trace(fmt.Sprintf("Looking for setup file at `%s`", setupPath))
	// Read file from `setupPath` on disk
	content, err := os.ReadFile(setupPath)
	if err != nil {
		return SetupStruct{}, false, nil
	}
	// Move the file after a successful read
	if err := os.WriteFile(
		fmt.Sprintf("%s.old", setupPath),
		content,
		0755,
	); err != nil {
		return SetupStruct{}, false, err
	}
	if err := os.Remove(setupPath); err != nil {
		return SetupStruct{}, false, err
	}
	// Parse setup file to struct `Setup`
	var setupTemp SetupStruct
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&setupTemp); err != nil {
		log.Error(fmt.Sprintf("Failed to parse setup file at `%s` into setup struct: %s", configPath, err.Error()))
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

	log.Debug(fmt.Sprintf("Setup file was detected. Moved setup file to `%s.old`.", setupPath))
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
		config.Database,
		"rescue",
	); err != nil {
		return err
	}
	return addRescueUser()
}

func FactoryReset() error {
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

func RunSetupStruct(setup SetupStruct) error {
	if err := FactoryReset(); err != nil {
		return err
	}
	// Run the actual setup
	if err := runSetupStruct(setup); err != nil {
		if err2 := abortSetup(); err2 != nil {
			log.Fatal(fmt.Sprintf("Aborting setup failed: could not add rescue user: %s", err2.Error()))
		}
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
	if err := createHardwareNodesInDatabase(setup.HardwareNodes); err != nil {
		log.Error("Aborting setup: could not create hardware nodes in database: ", err.Error())
		return err
	}
	if err := createRoomsInDatabase(setup.Rooms); err != nil {
		log.Error("Aborting setup: could not create rooms in database: ", err.Error())
		return err
	}
	if err := createUsersInDatabase(setup.Users); err != nil {
		log.Error("Aborting setup: could not create users in database: ", err.Error())
		return err
	}
	log.Info(fmt.Sprintf("Successfully ran setup using `%s`", setupPath))
	return nil
}

// Takes the `users` slice and createas according database entries
func createUsersInDatabase(users []setupUser) error {
	for _, usr := range users {
		_, alreadyExists, err := database.GetUserByUsername(usr.User.Username)
		if err != nil {
			return err
		}
		// It is likely that an `admin` user already exists
		// In this case, the admin user is re-created using the new metadata
		if alreadyExists && usr.User.Username != "admin" {
			return fmt.Errorf("cannot create user: user `%s` already exists", usr.User.Username)
		}
		// Create the user itself
		if err := database.InsertUser(database.FullUser{
			Username:          usr.User.Username,
			Forename:          usr.User.Forename,
			Surname:           usr.User.Surname,
			PrimaryColorDark:  usr.User.PrimaryColorDark,
			PrimaryColorLight: usr.User.PrimaryColorLight,
			Password:          usr.User.Password,
		}); err != nil {
			return err
		}
		// Setup the user's authentication tokens
		for _, token := range usr.Tokens {
			if err := database.InsertUserToken(token.Token, usr.User.Username, token.Label); err != nil {
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
				return fmt.Errorf("cannot grant invalid permission: `%s` to user `%s`", permission, usr.User.Username)
			}
			if _, err := user.AddPermission(
				usr.User.Username,
				database.PermissionType(permission),
			); err != nil {
				return err
			}
		}
		// Setup the user's switch permissions
		for _, swPermission := range usr.SwitchPermissions {
			_, found, err := database.GetSwitchById(swPermission)
			if err != nil {
				return err
			}
			if !found {
				return fmt.Errorf("cannot grant invalid switch permission `%s` to user `%s`", swPermission, usr.User.Username)
			}
			if _, err := database.AddUserSwitchPermission(
				usr.User.Username,
				swPermission,
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
				return fmt.Errorf("cannot grant invalid camera permission `%s` to user `%s`", camPermission, usr.User.Username)
			}
			if _, err := database.AddUserCameraPermission(
				usr.User.Username,
				camPermission,
			); err != nil {
				return err
			}
		}
		// Setup the user's Homescripts
		// Current arguments are being used for checking preexistence of arguments
		argsDB, err := database.ListAllHomescriptArgsOfUser(usr.User.Username)
		if err != nil {
			return err
		}
		for _, homescript := range usr.Homescripts {
			_, found, err := database.GetUserHomescriptById(
				homescript.Data.Id,
				usr.User.Username,
			)
			if err != nil {
				return err
			}
			if found {
				return fmt.Errorf("cannot create Homescript: id `%s` is already taken", homescript.Data.Id)
			}
			if err := database.CreateNewHomescript(database.Homescript{
				Owner: usr.User.Username,
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
					Owner: usr.User.Username,
					Data: database.AutomationData{
						Name:           autom.Name,
						Description:    autom.Description,
						CronExpression: autom.CronExpression,
						HomescriptId:   homescript.Data.Id,
						Enabled:        autom.Enabled,
						TimingMode:     autom.TimingMode,
					},
				}); err != nil {
					return err
				}
			}
		}

		// Setup Homescript storage
		for _, storageItem := range usr.HomescriptStorage {
			if err := database.InsertHmsStorageEntry(usr.User.Username, storageItem.Key, storageItem.Value); err != nil {
				return err
			}
		}

		// Setup the user's reminders
		for _, rem := range usr.Reminders {
			if _, err := database.CreateNewReminder(
				rem.Name,
				rem.Description,
				rem.DueDate,
				usr.User.Username,
				rem.Priority,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

// Takes the specified `rooms` and creates according database entries
func createRoomsInDatabase(rooms []setupRoom) error {
	for _, room := range rooms {
		if err := database.CreateRoom(room.Data); err != nil {
			log.Error("Could not create rooms from setup file: ", err.Error())
			return err
		}
		for _, switchItem := range room.Switches {
			if err := database.CreateSwitch(
				switchItem.Id,
				switchItem.Name,
				room.Data.Id,
				switchItem.Watts,
				switchItem.TargetNode,
			); err != nil {
				log.Error("Could not create switches from setup file: ", err.Error())
				return err
			}
		}
		for _, camera := range room.Cameras {
			// Override the (possible) empty room-id to match the current room
			if err := database.CreateCamera(database.Camera{
				Id:     camera.Id,
				Name:   camera.Name,
				Url:    camera.Url,
				RoomId: room.Data.Id,
			}); err != nil {
				log.Error("Could not create cameras from setup file: ", err.Error())
				return err
			}
		}
	}
	return nil
}

// Takes the specified `hardwareNodes` and creates according database entries
func createHardwareNodesInDatabase(nodes []setupHardwareNode) error {
	for _, node := range nodes {
		if err := database.CreateHardwareNode(
			database.HardwareNode{
				Name:    node.Name,
				Url:     node.Url,
				Token:   node.Token,
				Enabled: node.Enabled,
			},
		); err != nil {
			log.Error("Could not create hardware nodes from setup file: ", err.Error())
			return err
		}
	}
	return nil
}

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
