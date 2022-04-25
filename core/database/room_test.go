package database

import "testing"

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

func TestCreateRoom(t *testing.T) {
	table := []struct {
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
			t.Errorf("Room %s did not follow `listable` spec", test.Room.Id)
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
				valid = true
			}
		}
		if valid != test.Listable {
			t.Errorf("Room %s did not follow `listable` spec", test.Room.Id)
			return
		}
	}
}
