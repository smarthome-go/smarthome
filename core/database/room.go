package database

import "fmt"

// Camera struct, used in `config.rooms.cameras``
type Camera struct {
	Id     int    `json:"id"`
	RoomId string `json:"roomId"`
	Url    string `json:"url"`
	Name   string `json:"name"`
}

// Identified by a unique Id, has a Name and Description
// When used in config file, the Switches slice is also populated
type Room struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Switches    []Switch `json:"switches"`
	Cameras     []Camera `json:"cameras"`
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

// Creates a new room
func CreateRoom(RoomId string, Name string, Description string) error {
	query, err := db.Prepare(`
	INSERT INTO
	room(Id, Name, Description)
	VALUES(?,?,?)
	ON DUPLICATE KEY
	UPDATE Name=VALUES(Name),
	Description=VALUES(Description)
	`)
	if err != nil {
		log.Error("Could not create room: preparing query failed: ", err.Error())
		return err
	}
	res, err := query.Exec(RoomId, Name, Description)
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
		log.Debug(fmt.Sprintf("Added room `%s` with name `%s`", RoomId, Name))
	}
	defer query.Close()
	return nil
}

// Returns a list of rooms, excludes metadata like switches and cameras
func ListRooms() ([]Room, error) {
	query := `
	SELECT
	Id, Name, Description
	FROM room
	`
	res, err := db.Query(query)
	if err != nil {
		log.Error("Failed to list rooms: executing query failed: ", err.Error())
		return nil, err
	}
	rooms := make([]Room, 0)
	for res.Next() {
		var roomTemp Room
		if err := res.Scan(&roomTemp.Id, &roomTemp.Name, &roomTemp.Description); err != nil {
			log.Error("Failed to list rooms: failed to scan results: ", err.Error())
			return nil, err
		}
		rooms = append(rooms, roomTemp)
	}
	return rooms, nil
}

func listPersonalRoomsWithoutMetadata(username string) ([]Room, error) {
	query := `
	SELECT DISTINCT
	room.Id, room.Name, room.Description
	FROM room
	JOIN switch
	ON switch.RoomId=room.Id
	JOIN hasSwitchPermission
	ON switch.Id=hasSwitchPermission.Switch
	`
	res, err := db.Query(query)
	if err != nil {
		log.Error("Failed to list personal rooms: executing query failed: ", err.Error())
		return nil, err
	}
	rooms := make([]Room, 0)
	for res.Next() {
		roomTemp := Room{
			Cameras:  make([]Camera, 0),
			Switches: make([]Switch, 0),
		}
		if err := res.Scan(&roomTemp.Id, &roomTemp.Name, &roomTemp.Description); err != nil {
			log.Error("Failed to list personal rooms: failed to scan results: ", err.Error())
			return nil, err
		}
		rooms = append(rooms, roomTemp)
	}
	return rooms, nil
}

// Returns a complete list of rooms, includes metadata like switches
func ListPersonalRoomsAll(username string) ([]Room, error) {
	rooms, err := listPersonalRoomsWithoutMetadata(username)
	if err != nil {
		return nil, err
	}
	switches, err := ListUserSwitches(username)
	if err != nil {
		return nil, err
	}
	for index, room := range rooms {
		rooms[index].Switches = make([]Switch, 0)
		for _, switchItem := range switches {
			if switchItem.RoomId == room.Id {
				rooms[index].Switches = append(rooms[index].Switches, switchItem)
			}
		}
	}
	return rooms, nil
}
