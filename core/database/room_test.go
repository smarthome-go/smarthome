package database

import "testing"

func TestCreateRoomTable(t *testing.T) {
	if err := createRoomTable(); err != nil {
		t.Error(err.Error())
		return

	}
}

func TestCreateRoom(t *testing.T) {
	table := []Room{
		{
			Id:          "test_1",
			Name:        "test_1",
			Description: "test_1",
		},
		{
			Id:          "test_2",
			Name:        "test_2",
			Description: "test_2",
		},
		{
			Id:          "test_3",
			Name:        "test_3",
			Description: "test_3",
		},
	}
	for _, room := range table {
		if err := CreateRoom(
			room.Id,
			room.Name,
			room.Description,
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
			if item.Id == room.Id {
				valid = true
			}
		}
		if !valid {
			t.Errorf("Room %s was not found after creation", room.Id)
			return
		}
		rooms, err = listPersonalRoomsWithoutMetadata("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}
		valid = false
		for _, item := range rooms {
			if item.Id == room.Id {
				valid = true
			}
		}
		if valid {
			t.Errorf("Room %s was found in personal rooms despite no switches are set up", room.Id)
			return
		}
	}
}
