package database

import "database/sql"

type Camera struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Url    string `json:"url"`
	RoomId string `json:"roomId"`
}

func createCameraTable() error {
	_, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	camera(
		Id VARCHAR(50) PRIMARY KEY,
		Name VARCHAR(50),
		Url text,
		RoomId VARCHAR(30),
		FOREIGN KEY (RoomId) REFERENCES room(Id)
	)
	`)
	if err != nil {
		log.Error("Failed to create camera table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

func CreateCamera(data Camera) error {
	query, err := db.Prepare(`
	INSERT INTO
	camera(
		Id,
		Name,
		Url,
		RoomId
	)
	VALUES(?, ?, ?, ?)
	ON DUPLICATE KEY
		UPDATE
		Name=VALUES(Name)
	`)
	if err != nil {
		log.Error("Failed to create camera: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(
		data.Id,
		data.Name,
		data.Url,
		data.RoomId,
	); err != nil {
		log.Error("Failed to create camera: executing query failed: ", err.Error())
		return err
	}
	return nil
}

func ModifyCamera(id string, newName string, newUrl string) error {
	query, err := db.Prepare(`
  UPDATE camera
  SET
    Name=?,
    Url=?
  WHERE ID=?
  `)
	if err != nil {
		log.Error("Failed to modify camera: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(newName, newUrl, id); err != nil {
		log.Error("Failed to modify camera: executing query failed: ", err.Error())
		return err
	}
	return nil
}

func ListCameras() ([]Camera, error) {
	res, err := db.Query(`
	SELECT
		Id,
		Name,
		Url,
		RoomId
	FROM camera
	`)
	if err != nil {
		log.Error("Failed to list cameras: executing query failed: ", err.Error())
		return nil, err
	}

	cameras := make([]Camera, 0)
	for res.Next() {
		var camera Camera
		if err := res.Scan(
			&camera.Id,
			&camera.Name,
			&camera.Url,
			&camera.RoomId,
		); err != nil {
			log.Error("Failed to list cameras: scanning results failed: ", err.Error())
			return nil, err
		}
		cameras = append(cameras, camera)
	}
	return cameras, nil
}

func GetCameraById(id string) (Camera, bool, error) {
	query, err := db.Prepare(`
	SELECT
		Id,
		Name,
		Url,
		RoomId
	FROM camera
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to get camera by id: preparing query failed: ", err.Error())
		return Camera{}, false, err
	}
	var camera Camera
	if err := query.QueryRow(id).Scan(
		&camera.Id,
		&camera.Name,
		&camera.Url,
		&camera.RoomId,
	); err != nil {
		if err == sql.ErrNoRows {
			return Camera{}, false, nil
		}
		log.Error("Failed to get camera by id: executing query failed: ", err.Error())
		return Camera{}, false, err
	}
	return camera, true, nil
}

func DeleteCamera(id string) error {
	query, err := db.Prepare(`
	DELETE FROM camera
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to delete camera: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(id); err != nil {
		log.Error("Failed to delete camera: executing query failed: ", err.Error())
		return err

	}
	return nil
}

// Deletes all cameras in an arbitrary room
func DeleteRoomCameras(roomId string) error {
	query, err := db.Prepare(`
	DELETE FROM camera
	WHERE RoomId=?
	`)
	if err != nil {
		log.Error("Failed to delete room cameras: preparing query failed: ", err.Error())
		return err
	}
	if _, err := query.Exec(roomId); err != nil {
		log.Error("Failed to delete room cameras: executing query failed: ", err.Error())
		return err
	}
	return nil
}
