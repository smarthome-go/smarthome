package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/hardware"
)

type Setup struct {
	HardwareNodes []hardware.Node `json:"hardwareNodes"`
	Rooms         []database.Room `json:"rooms"`
}

const setupPath = "./data/config/setup.json"

// Used for setting up a smarthome server quickly
// Reads a setup file at startup and starts functions that initialize those values in the database
// Used for quick setup of a smarthome instance
func RunSetup() error {
	setup, shouldProceed, err := readSetupFile()
	if err != nil {
		log.Error("Failed to run setup: ", err.Error())
		return err
	}
	if !shouldProceed {
		log.Debug("No setup file found: starting without setup.")
		return nil
	}
	err = createRoomsInDatabase(setup.Rooms)
	if err != nil {
		log.Error("Failed to run setup: could not create entries in database: ", err.Error())
	}
	log.Info("Successfully ran setup")
	return nil
}

// TODO: add some sort of web import / export later
// Returns the setup struct, a bool that indicates that a setup file has been read and an error
func readSetupFile() (Setup, bool, error) {
	log.Trace("Looking for `setup.json`")
	// Read file from <setupPath> on disk
	content, err := ioutil.ReadFile(setupPath)
	if err != nil {
		return Setup{}, false, nil
	}
	// Parse setup file to struct <Setup>
	var setupTemp Setup
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&setupTemp)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to parse setup file at `%s` into Setup struct: %s", configPath, err.Error()))
		return Setup{}, false, err
	}
	return setupTemp, true, nil
}

// Takes the specified `rooms` and creates according database entries
func createRoomsInDatabase(rooms []database.Room) error {
	for _, room := range rooms {
		if err := database.CreateRoom(room.Id, room.Name, room.Description); err != nil {
			log.Error("Could not create rooms from config file")
			return err
		}
		for _, switchItem := range room.Switches {
			if err := database.CreateSwitch(switchItem.Id, switchItem.Name, room.Id); err != nil {
				log.Error("Could not create switches from config file:")
				return err
			}
			if _, err := database.AddUserSwitchPermission("admin", switchItem.Id); err != nil {
				log.Error("Could not add switch to switchPermissions of the admin user")
				return err
			}
		}
	}
	return nil
}
