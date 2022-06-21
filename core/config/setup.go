package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/smarthome-go/smarthome/core/database"
)

type Setup struct {
	HardwareNodes []database.HardwareNode `json:"hardwareNodes"`
	Rooms         []database.Room         `json:"rooms"`
}

// Is again a variable for testing
var setupPath = "./data/config/setup.json"

// TODO: add some sort of web import / export later
// Returns the setup struct, a bool that indicates that a setup file has been read and an error
func readSetupFile() (Setup, bool, error) {
	log.Trace(fmt.Sprintf("Looking for setup file at `%s`", setupPath))
	// Read file from `setupPath` on disk
	content, err := ioutil.ReadFile(setupPath)
	if err != nil {
		return Setup{}, false, nil
	}
	// Parse setup file to struct `Setup`
	var setupTemp Setup
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&setupTemp); err != nil {
		log.Error(fmt.Sprintf("Failed to parse setup file at `%s` into setup struct: %s", configPath, err.Error()))
		return Setup{}, false, err
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
	if err := createRoomsInDatabase(setup.Rooms); err != nil {
		log.Error("Aborting setup: could not create rooms in database: ", err.Error())
		return err
	}
	if err := createHardwareNodesInDatabase(setup.HardwareNodes); err != nil {
		log.Error("Aborting setup: could not create hardware nodes in database: ", err.Error())
		return err
	}
	log.Info(fmt.Sprintf("Successfully ran setup using `%s`", setupPath))
	return nil
}

// Takes the specified `rooms` and creates according database entries
func createRoomsInDatabase(rooms []database.Room) error {
	for _, room := range rooms {
		if err := database.CreateRoom(room.Data); err != nil {
			log.Error("Could not create rooms from setup file: ", err.Error())
			return err
		}
		for _, switchItem := range room.Switches {
			if err := database.CreateSwitch(switchItem.Id, switchItem.Name, room.Data.Id, switchItem.Watts); err != nil {
				log.Error("Could not create switches from setup file: ", err.Error())
				return err
			}
		}
		for _, camera := range room.Cameras {
			// Override the (possible) empty room-id to match the current room
			camera.RoomId = room.Data.Id
			if err := database.CreateCamera(camera); err != nil {
				log.Error("Could not create cameras from setup file: ", err.Error())
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
			database.HardwareNode{
				Name:  node.Name,
				Url:   node.Url,
				Token: node.Token,
			},
		); err != nil {
			log.Error("Could not create hardware nodes from setup file: ", err.Error())
			return err
		}
	}
	return nil
}
