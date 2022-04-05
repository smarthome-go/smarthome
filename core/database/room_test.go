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
	if err := createMockSwitches(); err != nil {
		t.Error(err.Error())
		return
	}
	table := []struct {
		Room     Room
		Listable bool // If the room will be in user rooms
	}{
		{
			Room: Room{
				Id:          "test_1",
				Name:        "test_1",
				Description: "test_1",
			},
			Listable: true,
		},
		{
			Room: Room{
				Id:          "test_2",
				Name:        "test_2",
				Description: "test_2",
			},
			Listable: true,
		},
		{
			Room: Room{
				Id:          "test_3",
				Name:        "test_3",
				Description: "test_3",
			},
			Listable: false,
		},
	}
	for _, room := range table {
		if err := CreateRoom(
			room.Room.Id,
			room.Room.Name,
			room.Room.Description,
		); err != nil {
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
			if item.Id == room.Room.Id {
				valid = true
			}
		}
		if !valid {
			t.Errorf("Room %s was not found after creation", room.Room.Id)
			return
		}
		rooms, err = listPersonalRoomsWithoutMetadata("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}
		valid = false
		for _, item := range rooms {
			if item.Id == room.Room.Id &&
				item.Name == room.Room.Name &&
				item.Description == room.Room.Description {
				valid = true
			}
		}
		if valid != room.Listable { // Check if the room was listable despite being marked as not listable
			t.Errorf("Room %s did not follow `lisable` spec", room.Room.Id)
			return
		}
		rooms, err = ListPersonalRoomsAll("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}
		valid = false
		for _, item := range rooms {
			if item.Id == room.Room.Id {
				valid = true
			}
		}
		if valid != room.Listable {
			t.Errorf("Room %s did not follow `lisable` spec", room.Room.Id)
			return
		}
	}
}
