package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/MikMuellerDev/smarthome/core/database"
)

type Setup struct {
	HardwareNodes []database.HardwareNode `json:"hardwareNodes"`
	Rooms         []database.Room         `json:"rooms"`
}

const setupPath = "./data/config/setup.json"

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
	if err := createRoomsInDatabase(setup.Rooms); err != nil {
		log.Error("Aboring setup: could not create room entries in database: ", err.Error())
		return err
	}
	if err := createHardwareNodesInDatabase(setup.HardwareNodes); err != nil {
		log.Error("Aboring setup: could not create hardware node entries in database: ", err.Error())
		return err
	}
	log.Info("Successfully ran setup")
	return nil
}

// Takes the specified `rooms` and creates according database entries
func createRoomsInDatabase(rooms []database.Room) error {
	for _, room := range rooms {
		if err := database.CreateRoom(room.Id, room.Name, room.Description); err != nil {
			log.Error("Could not create rooms from setup file: ", err.Error())
			return err
		}
		for _, switchItem := range room.Switches {
			if err := database.CreateSwitch(switchItem.Id, switchItem.Name, room.Id); err != nil {
				log.Error("Could not create switches from setup file: ", err.Error())
				return err
			}
			if _, err := database.AddUserSwitchPermission("admin", switchItem.Id); err != nil {
				log.Error("Could not add switch to switchPermissions of the admin user: ", err.Error())
				return err
			}
		}
	}
	return nil
}

// Tokes the specified `hardwareNodes` and creates according database entries
func createHardwareNodesInDatabase(nodes []database.HardwareNode) error {
	for _, node := range nodes {
		if err := database.CreateHardwareNode(
			node.Name,
			node.Url,
			node.Token,
		); err != nil {
			log.Error("Could not create hardware nodes from setup file: ", err.Error())
			return err
		}
	}
	return nil
}
