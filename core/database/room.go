package database

import (
	"database/sql"
	"fmt"
)

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
		log.Debug(fmt.Sprintf("Added room `%s` (Id: `%s`)", room.Name, room.Id))
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
		Id,
		Name,
		Description
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

// Returns a list containing room data of rooms which contain devices that the user is allowed to use
func ListPersonalRoomData(username string) ([]RoomData, error) {
	query, err := db.Prepare(`
	SELECT
		room.Id,
		room.Name,
		room.Description
	FROM room
		WHERE (
			SELECT COUNT(*)
			FROM device
			JOIN hasDevicePermission
				ON device.Id = hasDevicePermission.Device
			WHERE hasDevicePermission.Username=? AND device.RoomId=room.Id
		) > 0
		OR (
			SELECT COUNT(*)
			FROM camera
			JOIN hasCameraPermission
				ON camera.Id = hasCameraPermission.Camera
			WHERE hasCameraPermission.Username=? AND camera.RoomId=room.Id
		) > 0
	ORDER BY room.Name ASC;
	`)
	if err != nil {
		log.Error("Failed to list personal room data: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query(username, username)
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

func DeleteRoom(id string) error {
	if err := DeleteRoomDevices(id); err != nil {
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
