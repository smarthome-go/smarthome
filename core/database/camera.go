package database

import "database/sql"

type Camera struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Url    string `json:"url"`
	RoomId string `json:"roomId"`
}

// Creates the table which contains all the cameras
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

// Creates a new camera
// Checks, for example if the camera already exists should be completed beforehand
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
		Name=VALUES(Name),
		Url=VALUES(Url),
		RoomId=VALUES(RoomId)
	`)
	if err != nil {
		log.Error("Failed to create camera: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
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
	defer query.Close()
	if _, err := query.Exec(newName, newUrl, id); err != nil {
		log.Error("Failed to modify camera: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a list containing all cameras
// Used when deleting all cameras in a room
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
	defer res.Close()
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

// Returns the metadata of a given camera, wheter it could be found and a potential error
func GetCameraById(id string) (cam Camera, exists bool, err error) {
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
	defer query.Close()
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

// Deletes a camera, removes dependencies, such as camera-permission first
// Used in deleting all room cameras and deleting one camera
func DeleteCamera(id string) error {
	if err := RemoveCameraFromPermissions(id); err != nil {
		return err
	}
	query, err := db.Prepare(`
	DELETE FROM camera
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to delete camera: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(id); err != nil {
		log.Error("Failed to delete camera: executing query failed: ", err.Error())
		return err

	}
	return nil
}

// Deletes all cameras in an arbitrary room
// Uses `DeleteCamera` in order to remove dependencies beforehand
func DeleteRoomCameras(roomId string) error {
	cameras, err := ListCameras()
	if err != nil {
		return err
	}
	for _, cam := range cameras {
		// Skip any cameras which are not in the given room
		if cam.RoomId != roomId {
			continue
		}
		if err := DeleteCamera(cam.Id); err != nil {
			return err
		}
	}
	return nil
}
