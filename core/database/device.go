package database

import (
	"database/sql"
	"fmt"
)

type DEVICE_TYPE string

const (
	DEVICE_TYPE_INPUT  DEVICE_TYPE = "INPUT"
	DEVICE_TYPE_OUTPUT             = "OUTPUT"
)

var DEVICE_TYPE_MAP = map[string]DEVICE_TYPE{
	"INPUT":  DEVICE_TYPE_INPUT,
	"OUTPUT": DEVICE_TYPE_OUTPUT,
}

func ParseDeviceType(from string) (DEVICE_TYPE, bool) {
	type_, valid := DEVICE_TYPE_MAP[from]
	return type_, valid
}

// Identified by a device ID, has a name and belongs to one room
// Each device can either be an input device or an output device
type Device struct {
	DeviceType    DEVICE_TYPE `json:"type"`
	Id            string      `json:"id"`
	Name          string      `json:"name"`
	RoomId        string      `json:"roomId"`
	VendorId      string      `json:"vendorId"`
	ModelId       string      `json:"modelId"`
	SingletonJSON string      `json:"singletonJson"`
}

// Creates the table containing devices
// If the database fails, this function returns an error
func createDeviceTable() error {
	query := fmt.Sprintf(`
	CREATE TABLE
	IF NOT EXISTS
	device(
		DeviceType ENUM(
			'INPUT',
			'OUTPUT'
		),
		Id VARCHAR(20),
		Name VARCHAR(30),
		RoomId VARCHAR(30),
		DriverVendorId VARCHAR(%d),
		DriverModelId VARCHAR(%d),
		SingletonJson JSON,

		PRIMARY KEY(Id),
		FOREIGN KEY (RoomId)
		REFERENCES room(Id)
	)
	`,
		DEVICE_DRIVER_MODVEN_ID_LEN,
		DEVICE_DRIVER_MODVEN_ID_LEN,
	)
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create device Table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Creates a new device
// Will return an error if the database fails
func CreateDevice(data Device) error {
	query, err := db.Prepare(`
	INSERT INTO
	device(
		DeviceType,
		Id,
		Name,
		RoomId,
		DriverVendorId,
		DriverModelId
	)
	VALUES(?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY
		UPDATE
		DeviceType=VALUES(DeviceType),
		Id=VALUES(Id),
		Name=VALUES(Name),
		RoomId=VALUES(RoomId),
		DriverVendorId=VALUES(DriverVendorId),
		DriverModelId=VALUES(DriverModelId)
	`)
	if err != nil {
		log.Error("Failed to add device: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	res, err := query.Exec(
		data.DeviceType,
		data.Id,
		data.Name,
		data.RoomId,
		data.VendorId,
		data.ModelId,
	)
	if err != nil {
		log.Error("Failed to add device: executing query failed: ", err.Error())
		return err
	}

	// TODO: handle drivers

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("Could not get result of device creation: obtaining rowsAffected failed: ", err.Error())
		return err
	}
	if rowsAffected > 0 {
		log.Debug(fmt.Sprintf("Added device `%s` with name `%s`", data.Id, data.Name))
	}
	return nil
}

// Modifies the name of a given device.
func ModifyDeviceName(id string, name string) error {
	query, err := db.Prepare(`
	UPDATE device
	SET
		Name=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to modify device name: preparing query failed: ", err.Error())
		return err
	}

	defer query.Close()

	if _, err := query.Exec(name, id); err != nil {
		log.Error("Failed to modify device name: executing query failed: ", err.Error())
		return err
	}

	return nil
}

// Modifies the singleton JSON of a given device.
func ModifyDeviceSingletonJSON(id string, newJson string) error {
	query, err := db.Prepare(`
	UPDATE device
	SET
		SingletonJson=?
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to modify device singleton JSON: preparing query failed: ", err.Error())
		return err
	}

	defer query.Close()

	if _, err := query.Exec(newJson, id); err != nil {
		log.Error("Failed to modify device singleton JSON: executing query failed: ", err.Error())
		return err
	}

	return nil
}

// Delete a given device after all data which depends on this device has been deleted
func DeleteDevice(deviceId string) error {
	if err := RemoveDeviceFromPermissions(deviceId); err != nil {
		return err
	}

	query, err := db.Prepare(`
	DELETE FROM
	device
	WHERE Id=?
	`)

	if err != nil {
		log.Error("Failed to remove device: preparing query failed: ", err.Error())
		return err
	}

	defer query.Close()
	if _, err = query.Exec(deviceId); err != nil {
		log.Error("Failed to remove device: executing query failed: ", err.Error())
		return err
	}

	return nil
}

// Deletes all devices from an arbitrary room
func DeleteRoomDevices(roomId string) error {
	devices, err := ListAllDevices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		if device.RoomId != roomId {
			continue
		}

		if err := DeleteDevice(device.Id); err != nil {
			return err
		}
	}

	return nil
}

// Returns a list of all available devices with their attributes
func ListAllDevices() ([]Device, error) {
	res, err := db.Query(`
	SELECT
		device.DeviceType,
		device.Id,
		device.Name,
		device.RoomId,
		device.DriverVendorId,
		device.DriverModelId,
		device.SingletonJson
	FROM device
	`)
	if err != nil {
		log.Error("Could not list devices: failed to execute query: ", err.Error())
		return nil, err
	}
	defer res.Close()

	devices := make([]Device, 0)
	for res.Next() {
		var device Device
		if err := res.Scan(
			&device.DeviceType,
			&device.Id,
			&device.Name,
			&device.RoomId,
			&device.VendorId,
			&device.ModelId,
			&device.SingletonJSON,
		); err != nil {
			log.Error("Could not list devices: Failed to scan results: ", err.Error())
			return nil, err
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// Like `list all devices` but takes a username as a filter
// Only returns devices which are contained in the device-permission table with the given user
func ListUserDevicesQuery(username string) ([]Device, error) {
	query, err := db.Prepare(`
	SELECT
		device.DeviceType,
		device.Id,
		device.Name,
		device.RoomId,
		device.DriverVendorId,
		device.DriverModelId,
		device.SingletonJson
	FROM device
	JOIN hasDevicePermission
		ON hasDevicePermission.Device=device.Id
	WHERE hasDevicePermission.Username=?`,
	)
	if err != nil {
		log.Error("Could not list user devices: preparing query failed: ", err.Error())
		return nil, err
	}

	defer query.Close()
	res, err := query.Query(username)
	if err != nil {
		log.Error("Could not list user devices: executing query failed: ", err.Error())
		return nil, err
	}

	devices := make([]Device, 0)
	for res.Next() {
		var device Device

		if err := res.Scan(
			&device.DeviceType,
			&device.Id,
			&device.Name,
			&device.RoomId,
			&device.VendorId,
			&device.ModelId,
			&device.SingletonJSON,
		); err != nil {
			log.Error("Could not list user devices: Failed to scan results: ", err.Error())
			return nil, err
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func ListUserDevices(username string) ([]Device, error) {
	hasPermissionToAllDevices, err := UserHasPermission(username, PermissionModifyRooms)
	if err != nil {
		return nil, err
	}
	if hasPermissionToAllDevices {
		return ListAllDevices()
	}
	return ListUserDevicesQuery(username)
}

// Returns an arbitrary device given its ID
func GetDeviceById(id string) (dev Device, found bool, err error) {
	query, err := db.Prepare(`
	SELECT
		device.DeviceType,
		device.Id,
		device.Name,
		device.RoomId,
		device.DriverVendorId,
		device.DriverModelId,
		device.SingletonJson
	FROM device
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to get device by ID: preparing query failed: ", err.Error())
		return Device{}, false, err
	}
	defer query.Close()

	var device Device
	if err := query.QueryRow(id).Scan(
		&device.DeviceType,
		&device.Id,
		&device.Name,
		&device.RoomId,
		&device.VendorId,
		&device.ModelId,
		&device.SingletonJSON,
	); err != nil {
		if err == sql.ErrNoRows {
			return Device{}, false, nil
		}
		log.Error("Failed to get device by id: scanning results failed: ", err.Error())
		return Device{}, false, err
	}

	return device, true, nil
}
