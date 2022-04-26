package database

import (
	"fmt"
	"testing"
	"time"
)

var table = []struct {
	Room     RoomData
	Listable bool // If the room will be in user rooms
}{
	{
		Room: RoomData{
			Id:          "test_1",
			Name:        "test_1",
			Description: "test_1",
		},
		Listable: true,
	},
	{
		Room: RoomData{
			Id:          "test_2",
			Name:        "test_2",
			Description: "test_2",
		},
		Listable: true,
	},
	{
		Room: RoomData{
			Id:          "test_3",
			Name:        "test_3",
			Description: "test_3",
		},
		Listable: false,
	},
}

func TestCreateRoomTable(t *testing.T) {
	if err := createRoomTable(); err != nil {
		t.Error(err.Error())
		return

	}
}

func createMockSwitches() error {
	switches := []Switch{
		{
			Id:     "test1",
			RoomId: "test_1",
		},
		{
			Id:     "test2",
			RoomId: "test_2",
		},
	}
	for _, switchItem := range switches {
		if err := CreateSwitch(
			switchItem.Id,
			switchItem.Name,
			switchItem.RoomId,
			switchItem.Watts,
		); err != nil {
			return err
		}
		if _, err := AddUserSwitchPermission("admin", switchItem.Id); err != nil {
			return err
		}
	}
	return nil
}

func TestRooms(t *testing.T) {
	for _, test := range table {
		if err := CreateRoom(test.Room); err != nil {
			t.Error(err.Error())
			return
		}
		rooms, err := ListRooms()
		if err != nil {
			t.Error(err.Error())
			return
		}
		valid := false
		for _, item := range rooms {
			if item.Id == test.Room.Id {
				valid = true
			}
		}
		if !valid {
			t.Errorf("Room %s was not found after creation", test.Room.Id)
			return
		}
	}
	// After room has been created, create switches
	if err := createMockSwitches(); err != nil {
		t.Error(err.Error())
		return
	}
	for _, test := range table {
		rooms, err := listPersonalRoomData("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}
		valid := false
		for _, item := range rooms {
			if item.Id == test.Room.Id &&
				item.Name == test.Room.Name &&
				item.Description == test.Room.Description {
				valid = true
			}
		}
		if valid != test.Listable { // Check if the room was listable despite being marked as not listable
			t.Errorf("Room %s did not follow `listable` spec: want: %t got: %t", test.Room.Id, test.Listable, valid)
			return
		}
		newRooms, err := ListPersonalRooms("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}
		valid = false
		for _, room := range newRooms {
			if room.Data.Id == test.Room.Id {
				// Verify by retrieving room by id
				roomTemp, found, err := GetRoomDataById(test.Room.Id)
				if err != nil {
					t.Error(err.Error())
					return
				}
				if !found {
					t.Errorf("`GetRoomDataById` indicates that it was not found want: %t got: %t", valid, found)
					return
				}
				if roomTemp.Id != test.Room.Id || roomTemp.Name != test.Room.Name || roomTemp.Description != test.Room.Description {
					t.Errorf("`GetRoomDataById` returned different metadata than intended: want: %v got: %v", test.Room, roomTemp)
					return
				}

				// Compare current values against test table
				if room.Data.Name != test.Room.Name || room.Data.Description != test.Room.Description {
					t.Errorf("Matched room holds different metadata than intended: want: %v got: %v", test.Room, room)
					return
				}
				valid = true
			}
		}
		if valid != test.Listable {
			t.Errorf("Room %s did not follow `listable` spec: want: %t got: %t", test.Room.Id, test.Listable, valid)
			return
		}
	}
}

func TestDeleteRoom(t *testing.T) {
	// Create Test Data
	for _, test := range table {
		if err := CreateRoom(test.Room); err != nil {
			t.Error(err.Error())
			return
		}
	}
	// Work with test data
	rooms, err := ListRooms()
	if err != nil {
		t.Error(err)
		return
	}
	for _, room := range rooms {
		// Validate creation in order to avoid false positives
		_, found, err := GetRoomDataById(room.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !found {
			t.Errorf("Room %s was not found after creation", room.Id)
			return
		}

		// Perform deletion
		if err := DeleteRoom(room.Id); err != nil {
			t.Error(err.Error())
			return
		}

		// Validate Deletion
		_, found, err = GetRoomDataById(room.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if found {
			t.Errorf("Room %s found after deletion", room.Id)
			return
		}
	}
}

func TestModifyRoom(t *testing.T) {
	// Create Test Data
	for _, test := range table {
		if err := CreateRoom(test.Room); err != nil {
			t.Error(err.Error())
			return
		}

		// Used as a random data source in order to guarantee unique labels
		currentTime := time.Now().UnixMilli()

		if err := ModifyRoomData(test.Room.Id, fmt.Sprintf("name:%d", currentTime), fmt.Sprintf("description:%d", currentTime)); err != nil {
			t.Error(err.Error())
			return
		}

		room, found, err := GetRoomDataById(test.Room.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !found {
			t.Errorf("Room %s not found after modifycation", test.Room.Id)
			return
		}

		if room.Name != fmt.Sprintf("name:%d", currentTime) || room.Description != fmt.Sprintf("description:%d", currentTime) {
			t.Errorf("Invalid values after modification: want: (name:%d, description:%d), got(%s, %s)", currentTime, currentTime, room.Name, room.Description)
			return
		}
	}
}
