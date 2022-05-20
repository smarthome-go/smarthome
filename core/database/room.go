package database

import (
	"database/sql"
	"fmt"
)

type Room struct {
	Data     RoomData `json:"data"`
	Switches []Switch `json:"switches"`
	Cameras  []Camera `json:"cameras"`
}

type RoomData struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Initializes the table containing the rooms
func createRoomTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	room(
		Id VARCHAR(30) PRIMARY KEY,
		Name VARCHAR(50),
		Description text
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create room table: ", err.Error())
		return err
	}
	return nil
}

// Creates a new room given an arbitrary, non-existing id
func CreateRoom(room RoomData) error {
	query, err := db.Prepare(`
	INSERT INTO 
	room(
		Id,
		Name,
		Description
	)
	VALUES(?, ?, ?)
	ON DUPLICATE KEY
		UPDATE
		Name=VALUES(Name),
		Description=VALUES(Description)
	`)
	if err != nil {
		log.Error("Could not create room: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	res, err := query.Exec(room.Id, room.Name, room.Description)
	if err != nil {
		log.Error("Could not create room: executing query failed: ", err.Error())
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("Could not get result of createRoom: obtaining rowsAffected failed: ", err.Error())
		return err
	}
	if rowsAffected > 0 {
		log.Debug(fmt.Sprintf("Added room `%s` with name `%s`", room.Id, room.Name))
	}
	return nil
}

// Updates the room's name and description
func ModifyRoomData(id string, newName string, newDescription string) error {
	query, err := db.Prepare(`
	UPDATE room
	SET
		Name=?,
		Description=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to modify room: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(newName, newDescription, id); err != nil {
		log.Error("Failed to modify room: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a list of room data
func ListRooms() ([]RoomData, error) {
	res, err := db.Query(`
	SELECT
		Id,
		Name,
		Description
	FROM room
	ORDER BY NAME ASC
	`)
	if err != nil {
		log.Error("Failed to list rooms: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	rooms := make([]RoomData, 0)
	for res.Next() {
		var roomTemp RoomData
		if err := res.Scan(&roomTemp.Id, &roomTemp.Name, &roomTemp.Description); err != nil {
			log.Error("Failed to list rooms: failed to scan results: ", err.Error())
			return nil, err
		}
		rooms = append(rooms, roomTemp)
	}
	return rooms, nil
}

// Returns an arbitrary room given its id, whether it exists an a possible error
func GetRoomDataById(id string) (RoomData, bool, error) {
	query, err := db.Prepare(`
	SELECT
		Id, Name, Description
	FROM room
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to get room data by id: preparing query failed: ", err.Error())
		return RoomData{}, false, err
	}
	defer query.Close()
	var room RoomData
	if err := query.QueryRow(id).Scan(&room.Id, &room.Name, &room.Description); err != nil {
		if err == sql.ErrNoRows {
			return RoomData{}, false, nil
		}
		log.Error("Failed to get room data by id: executing query failed: ", err.Error())
		return RoomData{}, false, err
	}
	return room, true, nil
}

// Returns a list containing room data of rooms which contain switches the user is allowed to use
func listPersonalRoomData(username string) ([]RoomData, error) {
	// TODO: restructure query to also account for cameraPermissions
	query, err := db.Prepare(`
	SELECT DISTINCT
		room.Id,
		room.Name,
		room.Description
	FROM room
		JOIN switch ON switch.RoomId=room.Id
		JOIN hasSwitchPermission ON switch.Id=hasSwitchPermission.Switch
	WHERE username=?
	ORDER BY NAME ASC
	`)
	if err != nil {
		log.Error("Failed to list personal room data: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Failed to list personal room data: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	rooms := make([]RoomData, 0)
	for res.Next() {
		roomTemp := RoomData{}
		if err := res.Scan(&roomTemp.Id, &roomTemp.Name, &roomTemp.Description); err != nil {
			log.Error("Failed to list personal room data: failed to scan results: ", err.Error())
			return nil, err
		}
		rooms = append(rooms, roomTemp)
	}
	return rooms, nil
}

// Deletes a room and all entities that depend on the room
func DeleteRoomQuery(id string) error {
	query, err := db.Prepare(`
	DELETE FROM room
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to delete room: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(id); err != nil {
		log.Error("Failed to delete room: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a complete list of rooms to which a user has access to, includes its metadata like switches and cameras
func ListPersonalRoomsWithData(username string) ([]Room, error) {
	rooms, err := listPersonalRoomData(username)
	if err != nil {
		return nil, err
	}
	// Get the user's switches
	switches, err := ListUserSwitches(username)
	if err != nil {
		return nil, err
	}
	// Get the user's cameras
	cameras, err := ListUserCameras(username)
	if err != nil {
		return nil, err
	}

	outputRooms := make([]Room, 0)
	for _, room := range rooms {
		switchesTemp := make([]Switch, 0)
		camerasTemp := make([]Camera, 0)
		// Add every switch which is in the current room
		for _, switchItem := range switches {
			if switchItem.RoomId == room.Id {
				switchesTemp = append(switchesTemp, switchItem)
			}
		}
		// Add every camera which is in the current room
		for _, camera := range cameras {
			if camera.RoomId == room.Id {
				camerasTemp = append(camerasTemp, camera)
			}
		}
		// Append to the output rooms
		outputRooms = append(outputRooms, Room{
			Data:     room,
			Switches: switchesTemp,
			Cameras:  camerasTemp,
		})
	}
	return outputRooms, nil
}

// Returns a complete list of rooms, includes its metadata like switches and cameras
func ListAllRoomsWithData() ([]Room, error) {
	rooms, err := ListRooms()
	if err != nil {
		return nil, err
	}
	// Get all switches
	switches, err := ListSwitches()
	if err != nil {
		return nil, err
	}
	// Get all cameras
	cameras, err := ListCameras()
	if err != nil {
		return nil, err
	}
	outputRooms := make([]Room, 0)
	for _, room := range rooms {
		switchesTemp := make([]Switch, 0)
		camerasTemp := make([]Camera, 0)
		// Add all switches of the current room
		for _, switchItem := range switches {
			if switchItem.RoomId == room.Id {
				switchesTemp = append(switchesTemp, switchItem)
			}
		}
		// Add all cameras of the current room
		for _, camera := range cameras {
			if camera.RoomId == room.Id {
				camerasTemp = append(camerasTemp, camera)
			}
		}
		outputRooms = append(outputRooms, Room{
			Data:     room,
			Switches: switchesTemp,
			Cameras:  camerasTemp,
		})
	}
	return outputRooms, nil
}

func DeleteRoom(id string) error {
	if err := DeleteRoomSwitches(id); err != nil {
		return err
	}
	if err := DeleteRoomCameras(id); err != nil {
		return err
	}
	if err := DeleteRoomQuery(id); err != nil {
		return err
	}
	return nil
}
