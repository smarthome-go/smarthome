package database

import (
	"database/sql"
	"fmt"
)

const DEVICE_DRIVER_MODVEN_ID_LEN = 50
const DEVICE_DRIVER_VERSION_LEN = 50
const DEVICE_DRIVER_DEFAULT_ICON = "developer_board"

type DriverTuple struct {
	VendorID string `json:"vendorId"`
	ModelID  string `json:"modelId"`
}

// TODO: change this so that there is no user owning the device driver script

type DeviceDriver struct {
	VendorID       string  `json:"vendorId"`
	ModelID        string  `json:"modelId"`
	Name           string  `json:"name"`
	Version        string  `json:"version"`
	HomescriptCode string  `json:"homescriptCode"`
	SingletonJSON  *string `json:"singletonJson"`
	// Useful for tracking if a driver needs to be restarted.
	Dirty bool `json:"dirty"`
}

// Creates the table containing device driver code and metadata
// If the database fails, this function returns an error
func createDeviceDriverTable() error {
	query, err := db.Prepare(fmt.Sprintf(`
	CREATE TABLE
	IF NOT EXISTS
	deviceDriver(
		VendorId			VARCHAR(%d) NULL,
		ModelId				VARCHAR(%d) NULL,
		Name				TEXT NOT NULL,
		Version				VARCHAR(%d) NOT NULL,
		HomescriptCode		LONGTEXT NOT NULL,
		SingletonJson		JSON,
		Dirty				BOOL,
		PRIMARY KEY (VendorId, ModelId)
	)
	`,
		DEVICE_DRIVER_MODVEN_ID_LEN,
		DEVICE_DRIVER_MODVEN_ID_LEN,
		DEVICE_DRIVER_VERSION_LEN,
	))
	if err != nil {
		return err
	}
	defer query.Close()
	if _, err := query.Exec(); err != nil {
		log.Error("Failed to create device driver table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

func generateHmsIdForDriver(vendorId string, modelId string) string {
	return fmt.Sprintf("@%s:%s", vendorId, modelId)
}

// Creates a new device driver entry
// Returns the ID of the newly internally used Homescript
func CreateNewDeviceDriver(driverData DeviceDriver) error {
	query, err := db.Prepare(`
	INSERT INTO
	deviceDriver(
		VendorId,
		ModelId,
		Name,
		Version,
		HomescriptCode,
		SingletonJson,
		Dirty
	)
	VALUES(?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new device driver: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()

	// Create the device driver
	if _, err = query.Exec(
		driverData.VendorID,
		driverData.ModelID,
		driverData.Name,
		driverData.Version,
		driverData.HomescriptCode,
		driverData.SingletonJSON,
		driverData.Dirty,
	); err != nil {
		log.Error("Failed to create new device driver: executing query failed: ", err.Error())
		return err
	}

	return nil
}

// Modifies the metadata of a given device driver
func ModifyDeviceDriver(newData DeviceDriver) error {
	query, err := db.Prepare(`
	UPDATE deviceDriver
	SET
		Name=?,
		Version=?,
		HomescriptCode=?,
		SingletonJson=?,
		Dirty=?
	WHERE VendorId=? AND ModelId=?
	`)
	if err != nil {
		log.Error("Failed to update device driver: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()

	if _, err := query.Exec(
		newData.Name,
		newData.Version,
		newData.HomescriptCode,
		newData.SingletonJSON,
		newData.Dirty,
		newData.VendorID,
		newData.ModelID,
	); err != nil {
		log.Error("Failed to update device driver: executing query failed: ", err.Error())
		return err
	}

	return nil
}

// Modifies only the JSON column, returns if the driver was found.
// TODO: remove `found` parameter
func ModifyDeviceDriverSingletonJSON(vendorId string, modelId string, newJson *string) error {
	query, err := db.Prepare(`
	UPDATE deviceDriver
	SET
		SingletonJson=?
	WHERE VendorId=? AND ModelId=?
	`)
	if err != nil {
		log.Error("Failed to update device driver JSON: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()

	_, err = query.Exec(
		newJson,
		vendorId,
		modelId,
	)
	if err != nil {
		log.Error("Failed to update device driver JSON: executing query failed: ", err.Error())
		return err
	}

	return nil
}

// Modifies the code of a given device driver.
// Returns `true` if the driver was found and modified.
func ModifyDeviceDriverCode(vendorId string, modelId string, newCode string) error {
	query, err := db.Prepare(`
	UPDATE deviceDriver
	SET
		HomescriptCode=?
	WHERE VendorId=? AND ModelId=?
	`)
	if err != nil {
		log.Error("Failed to update device driver code: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()

	if _, err := query.Exec(
		newCode,
		vendorId,
		modelId,
	); err != nil {
		log.Error("Failed to update device driver code: executing query failed: ", err.Error())
		return err
	}

	return nil
}

func ModifyDeviceDriverDirty(vendorId string, modelId string, dirty bool) error {
	query, err := db.Prepare(`
	UPDATE deviceDriver
	SET Dirty=? WHERE VendorId=? AND ModelId=?
	`)
	if err != nil {
		log.Error("Failed to update device driver code: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()

	if _, err := query.Exec(
		dirty,
		vendorId,
		modelId,
	); err != nil {
		log.Error("Failed to update device driver dirty: executing query failed: ", err.Error())
		return err
	}

	return nil
}

func GetDriverSources(ids []DriverTuple) (drivers map[DriverTuple]string, allFound bool, err error) {
	drivers = make(map[DriverTuple]string)

	query, err := db.Prepare(`
	SELECT deviceDriver.HomescriptCode
	FROM deviceDriver WHERE
		deviceDriver.VendorId = ?
		AND deviceDriver.ModelId = ?
	`)

	if err != nil {
		log.Errorf("Could not list driver sources: preparing query failed: %s\n", err.Error())
		return nil, false, err
	}

	for _, id := range ids {
		row := query.QueryRow(id.VendorID, id.ModelID)
		if err != nil {
			log.Errorf("Could not list driver sources: query row failed: %s\n", err.Error())
			return nil, false, err
		}

		var code string

		if err := row.Scan(&code); err != nil {
			if err == sql.ErrNoRows {
				return nil, false, nil
			}

			log.Errorf("Could not list driver sources: scanning failed: %s\n", err.Error())
			return nil, false, err
		}

		drivers[id] = code
	}

	return drivers, true, nil
}

// Returns a list of homescripts owned by a given user
func ListDeviceDrivers() ([]DeviceDriver, error) {
	query, err := db.Prepare(`
	SELECT
		deviceDriver.VendorId,
		deviceDriver.ModelId,
		deviceDriver.Name,
		deviceDriver.Version,
		deviceDriver.HomescriptCode,
		deviceDriver.SingletonJSON,
		deviceDriver.Dirty
	FROM deviceDriver
	`)
	if err != nil {
		log.Error("Failed to list device drivers: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()
	res, err := query.Query()
	if err != nil {
		log.Error("Failed to list device drivers: executing query failed: ", err.Error())
		return nil, err
	}
	defer res.Close()
	var drivers []DeviceDriver = make([]DeviceDriver, 0)
	for res.Next() {
		var driver DeviceDriver
		err := res.Scan(
			&driver.VendorID,
			&driver.ModelID,
			&driver.Name,
			&driver.Version,
			&driver.HomescriptCode,
			&driver.SingletonJSON,
			&driver.Dirty,
		)
		if err != nil {
			log.Error("Failed to list Homescript of user: scanning results failed: ", err.Error())
			return nil, err
		}
		drivers = append(drivers, driver)
	}

	return drivers, nil
}

func GetDeviceDriver(vendorId string, modelId string) (DeviceDriver, bool, error) {
	query, err := db.Prepare(`
	SELECT
		deviceDriver.Name,
		deviceDriver.Version,
		deviceDriver.HomescriptCode,
		deviceDriver.SingletonJSON,
		deviceDriver.Dirty
	FROM deviceDriver
	WHERE deviceDriver.VendorId=?
	AND deviceDriver.ModelId=?
	`)
	if err != nil {
		log.Error("Failed to get device driver: preparing query failed: ", err.Error())
		return DeviceDriver{}, false, err
	}
	defer query.Close()

	var driver DeviceDriver
	driver.VendorID = vendorId
	driver.ModelID = modelId

	if err := query.QueryRow(
		driver.VendorID,
		driver.ModelID,
	).Scan(
		&driver.Name,
		&driver.Version,
		&driver.HomescriptCode,
		&driver.SingletonJSON,
		&driver.Dirty,
	); err != nil {
		if err == sql.ErrNoRows {
			return DeviceDriver{}, false, nil
		}
		log.Error("Failed to get device driver: executing query failed: ", err.Error())
		return DeviceDriver{}, false, err
	}

	return driver, true, nil
}

// Deletes a homescript by its Id, does not check if the user has access to the homescript
func DeleteDeviceDriver(vendorId string, modelId string) error {
	query, err := db.Prepare(`
	DELETE FROM deviceDriver
	WHERE deviceDriver.VendorId=?
	AND deviceDriver.ModelId=?
	`)
	if err != nil {
		log.Error("Failed to delete device driver: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()

	_, err = query.Exec(vendorId, modelId)
	if err != nil {
		log.Error("Failed to delete device driver: executing query failed: ", err.Error())
		return err
	}

	return nil
}

// Used when deleting a device driver
func DriverHasDependentDevices(vendorId string, modelId string) (bool, error) {
	devices, err := ListAllDevices()
	if err != nil {
		return false, err
	}

	for _, dev := range devices {
		if dev.VendorID == vendorId && dev.ModelID == modelId {
			return true, nil
		}
	}

	return false, nil
}
