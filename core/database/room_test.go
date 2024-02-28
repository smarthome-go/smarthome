package database

import (
	"fmt"
	"testing"
	"time"
)

func TestCreateRoomTable(t *testing.T) {
	if err := createRoomTable(); err != nil {
		t.Error(err.Error())
		return

	}
}

func TestCreateRooms(t *testing.T) {
	table := []struct {
		Room     RoomData
		Listable bool // If the room will be in user rooms
	}{
		{
			Room: RoomData{
				ID:          "test_1",
				Name:        "test_1",
				Description: "test_1",
			},
			Listable: true,
		},
		{
			Room: RoomData{
				ID:          "test_2",
				Name:        "test_2",
				Description: "test_2",
			},
			Listable: true,
		},
		{
			Room: RoomData{
				ID:          "test_3",
				Name:        "test_3",
				Description: "test_3",
			},
			Listable: false,
		},
	}

	for _, test := range table {
		if err := CreateRoom(test.Room); err != nil {
			t.Error(err.Error())
		}
		rooms, err := ListRooms()
		if err != nil {
			t.Error(err.Error())
		}
		valid := false
		for _, item := range rooms {
			if item.ID == test.Room.ID {
				valid = true
			}
		}
		if !valid {
			t.Errorf("Room %s was not found after creation", test.Room.ID)
		}
	}
}

// TODO: reimplement this
// func TestListRooms(t *testing.T) {
// 	TestCreateRooms(t)
// 	// Create test switches
// 	switches := []Device{
// 		{
// 			Id:     "test1",
// 			RoomId: "test_1",
// 		},
// 		{
// 			Id:     "test2",
// 			RoomId: "test_2",
// 		},
// 	}
// 	for _, switchItem := range switches {
// 		if err := CreateDevice(
// 			switchItem.Id,
// 			switchItem.Name,
// 			switchItem.RoomId,
// 			switchItem.Watts,
// 			nil,
// 		); err != nil {
// 			t.Error(err.Error())
// 		}
// 		if _, err := AddUserDevicePermission("admin", switchItem.Id); err != nil {
// 			t.Error(err.Error())
// 		}
// 	}
//
// 	table := []struct {
// 		Room     RoomData
// 		Listable bool // If the room will be in user rooms
// 	}{
// 		{
// 			Room: RoomData{
// 				Id:          "test_1",
// 				Name:        "test_1",
// 				Description: "test_1",
// 			},
// 			Listable: true,
// 		},
// 		{
// 			Room: RoomData{
// 				Id:          "test_2",
// 				Name:        "test_2",
// 				Description: "test_2",
// 			},
// 			Listable: true,
// 		},
// 		{
// 			Room: RoomData{
// 				Id:          "test_3",
// 				Name:        "test_3",
// 				Description: "test_3",
// 			},
// 			Listable: false,
// 		},
// 	}
//
// 	for _, test := range table {
// 		rooms, err := ListPersonalRoomData("admin")
// 		if err != nil {
// 			t.Error(err.Error())
// 		}
// 		valid := false
// 		for _, item := range rooms {
// 			if item.Id == test.Room.Id {
// 				valid = true
// 			}
// 		}
// 		// Check if the room was listable despite being marked as not listable
// 		if valid != test.Listable {
// 			t.Errorf("Room %s did not follow `listable` spec: want: %t got: %t", test.Room.Id, test.Listable, valid)
// 		}
// 		newRooms, err := ListPersonalRoomsWithData("admin")
// 		if err != nil {
// 			t.Error(err.Error())
// 		}
// 		valid = false
// 		for _, room := range newRooms {
// 			if room.Data.Id == test.Room.Id {
// 				// Verify by retrieving room by id
// 				roomTemp, found, err := GetRoomDataById(test.Room.Id)
// 				if err != nil {
// 					t.Error(err.Error())
// 				}
// 				if !found {
// 					t.Errorf("`GetRoomDataById` indicates that it was not found want: %t got: %t", valid, found)
// 				}
// 				assert.Equal(t, test.Room, roomTemp, "room from id has invalid metadata")
// 				assert.Equal(t, test.Room, room.Data, "room from listings has invalid metadata")
// 				valid = true
// 			}
// 		}
// 		if valid != test.Listable {
// 			t.Errorf("Room %s did not follow `listable` spec: want: %t got: %t", test.Room.Id, test.Listable, valid)
// 		}
// 	}
// }

func TestDeleteRoom(t *testing.T) {
	table := []struct {
		Room     RoomData
		Listable bool // If the room will be in user rooms
	}{
		{
			Room: RoomData{
				ID:          "test_1",
				Name:        "test_1",
				Description: "test_1",
			},
			Listable: true,
		},
		{
			Room: RoomData{
				ID:          "test_2",
				Name:        "test_2",
				Description: "test_2",
			},
			Listable: true,
		},
		{
			Room: RoomData{
				ID:          "test_3",
				Name:        "test_3",
				Description: "test_3",
			},
			Listable: false,
		},
	}

	// Create Test Data
	for _, test := range table {
		if err := CreateRoom(test.Room); err != nil {
			t.Error(err.Error())
		}
	}
	// Work with test data
	rooms, err := ListRooms()
	if err != nil {
		t.Error(err)
	}
	for _, room := range rooms {
		// Validate creation in order to avoid false positives
		_, found, err := GetRoomDataById(room.ID)
		if err != nil {
			t.Error(err.Error())
		}
		if !found {
			t.Errorf("Room %s was not found after creation", room.ID)
		}
		// Perform deletion
		if err := DeleteRoom(room.ID); err != nil {
			t.Error(err.Error())
		}
		// Validate Deletion
		_, found, err = GetRoomDataById(room.ID)
		if err != nil {
			t.Error(err.Error())
		}
		if found {
			t.Errorf("Room %s found after deletion", room.ID)
		}
	}
}

func TestModifyRoom(t *testing.T) {
	table := []struct {
		Room     RoomData
		Listable bool // If the room will be in user rooms
	}{
		{
			Room: RoomData{
				ID:          "test_1",
				Name:        "test_1",
				Description: "test_1",
			},
			Listable: true,
		},
		{
			Room: RoomData{
				ID:          "test_3",
				Name:        "test_3",
				Description: "test_3",
			},
			Listable: false,
		},
	}

	// Create Test Data
	for _, test := range table {
		if err := CreateRoom(test.Room); err != nil {
			t.Error(err.Error())
		}
		// Used as a random data source in order to guarantee unique labels
		currentTime := time.Now().UnixMilli()
		if err := ModifyRoomData(test.Room.ID, fmt.Sprintf("name:%d", currentTime), fmt.Sprintf("description:%d", currentTime)); err != nil {
			t.Error(err.Error())
		}
		room, found, err := GetRoomDataById(test.Room.ID)
		if err != nil {
			t.Error(err.Error())
		}
		if !found {
			t.Errorf("Room %s not found after modifycation", test.Room.ID)
		}
		if room.Name != fmt.Sprintf("name:%d", currentTime) || room.Description != fmt.Sprintf("description:%d", currentTime) {
			t.Errorf("Invalid values after modification: want: (name:%d, description:%d), got(%s, %s)", currentTime, currentTime, room.Name, room.Description)
		}
	}
}
