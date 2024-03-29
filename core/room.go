package core

import (
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
)

type Room struct {
	Data    database.RoomData        `json:"data"`
	Devices []database.ShallowDevice `json:"devices"`
	Cameras []database.Camera        `json:"cameras"`
}

// Returns a complete list of rooms, includes its metadata like devices and cameras
func ListAllRoomsWithData(redactCameraUrl bool) ([]Room, error) {
	rooms, err := database.ListRooms()
	if err != nil {
		return nil, err
	}

	// Get all devices.
	devices, err := driver.Manager.ListAllDevicesShallow()
	if err != nil {
		return nil, err
	}

	// Get all cameras.
	cameras, err := database.ListCameras()
	if err != nil {
		return nil, err
	}

	outputRooms := make([]Room, 0)
	for _, room := range rooms {
		devicesTemp := make([]database.ShallowDevice, 0)
		camerasTemp := make([]database.Camera, 0)

		// Add all devices of the current room
		for _, device := range devices {
			if device.RoomID == room.ID {
				devicesTemp = append(devicesTemp, device)
			}
		}

		// Add all cameras of the current room
		for _, camera := range cameras {
			if redactCameraUrl {
				camera.Url = "redacted"
			}
			if camera.RoomID == room.ID {
				camerasTemp = append(camerasTemp, camera)
			}
		}

		outputRooms = append(outputRooms, Room{
			Data:    room,
			Devices: devicesTemp,
			Cameras: camerasTemp,
		})
	}

	return outputRooms, nil
}

// Returns a complete list of rooms to which a user has access to, includes its metadata like devices and cameras.
// NOTE: do to extreme (> 1s) response times in the web UI, this function no longer includes rich devices.
// Instead, only the shallow device list of each room is returned.
func ListPersonalRoomsWithData(username string) ([]Room, error) {
	rooms, err := database.ListPersonalRoomData(username)
	if err != nil {
		return nil, err
	}

	// Get the user's devices.
	devices, err := driver.Manager.ListPersonalDevicesShallow(username)
	if err != nil {
		return nil, err
	}

	// Get the user's cameras.
	cameras, err := database.ListUserCameras(username)
	if err != nil {
		return nil, err
	}

	outputRooms := make([]Room, 0)
	for _, room := range rooms {
		devicesTemp := make([]database.ShallowDevice, 0)
		camerasTemp := make([]database.Camera, 0)

		// Add every device which is in the current room.
		for _, device := range devices {
			if device.RoomID == room.ID {
				devicesTemp = append(devicesTemp, device)
			}
		}

		// Add every camera which is in the current room.
		for _, camera := range cameras {
			if camera.RoomID == room.ID {
				camerasTemp = append(camerasTemp, camera)
			}
		}

		// Append to the output rooms
		outputRooms = append(outputRooms, Room{
			Data:    room,
			Devices: devicesTemp,
			Cameras: camerasTemp,
		})
	}

	return outputRooms, nil
}
