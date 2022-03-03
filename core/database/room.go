package database

import "fmt"

func createRoomTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS room(
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

func CreateRoom(RoomId string, Name string, Description string) error {
	query, err := db.Prepare(`
	INSERT INTO room(Id, Name, Description) VALUES(?,?,?) ON DUPLICATE KEY UPDATE Name=VALUES(Name), Description=VALUES(Description)
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
	return nil
}
