package database

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

func createBelongsToRoomTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS belongsToRoom(
		RoomId VARCHAR(30),
		SwitchId VARCHAR(2),
		CONSTRAINT RoomId FOREIGN KEY (RoomId)
		REFERENCES room(Id),
		CONSTRAINT SwitchId FOREIGN KEY (SwitchId)
		REFERENCES switch(Id)
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create belongsToRoom table: ", err.Error())
		return err
	}
	return nil
}
