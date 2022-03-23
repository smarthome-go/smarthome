package database

func createCameraTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	camera(
		Id INT AUTO_INCREMENT,
		RoomId VARCHAR(30),
		Url text,
		Name VARCHAR(50),
		PRIMARY KEY(Id),
		CONSTRAINT CameraRoomId
		FOREIGN KEY (RoomId)
		REFERENCES room(Id)
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create camera table: executing query failed: ", err.Error())
		return err
	}
	return nil
}
