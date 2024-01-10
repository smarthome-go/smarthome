package database

import (
	"database/sql"
	"fmt"
)

const DEVICE_DRIVER_MODVEN_ID_LEN = 50
const DEVICE_DRIVER_VERSION_LEN = 50
const DEVICE_DRIVER_DEFAULT_ICON = "developer_board"

// TODO: change this so that there is no user owning the device driver script

type DeviceDriver struct {
	Driver     DeviceDriverData `json:"driver"`
	Homescript Homescript       `json:"homescript"`
}

type DeviceDriverData struct {
	VendorId       string `json:"vendorId"`
	ModelId        string `json:"modelId"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	HomescriptCode string `json:"homescriptCode"`
}

// Creates the table containing device driver code and metadata
// If the database fails, this function returns an error
func createDeviceDriverTable() error {
	query, err := db.Prepare(fmt.Sprintf(`
	CREATE TABLE
	IF NOT EXISTS
	deviceDriver(
		VendorId			VARCHAR(%d),
		ModelId				VARCHAR(%d),
		Name				TEXT,
		Version				VARCHAR(%d),
		HomescriptCode		TEXT,
		PRIMARY KEY (VendorId, ModelId),
		CONSTRAINT HomescriptId
		FOREIGN KEY (HomescriptId)
		REFERENCES homescript(Id)
	)
	`,
		DEVICE_DRIVER_MODVEN_ID_LEN,
		DEVICE_DRIVER_MODVEN_ID_LEN,
		DEVICE_DRIVER_VERSION_LEN,
		// HOMESCRIPT_ID_LEN,
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

// func generateHmsIdForDriver(vendorId string, modelId string) string {
// 	return fmt.Sprintf("@driver_%s:%s", vendorId, modelId)
// }

// Creates a new device driver entry
// Returns the ID of the newly internally used Homescript
func CreateNewDeviceDriver(driverData DeviceDriverData, homescriptCode string) (string, error) {
	// driverHmsCode := generateHmsIdForDriver(driverData.VendorId, driverData.ModelId)

	// Create new Homescript for this driver
	// if err := CreateNewHomescript(Homescript{
	// 	Owner: DEFAULT_ADMIN_USERNAME,
	// 	Data: HomescriptData{
	// 		Id:                  driverHmsCode,
	// 		Name:                fmt.Sprintf("%s Driver", driverData.Name),
	// 		Description:         fmt.Sprintf("%s:%s", driverData.VendorId, driverData.ModelId),
	// 		QuickActionsEnabled: false,
	// 		IsWidget:            false,
	// 		SchedulerEnabled:    false,
	// 		Code:                homescriptCode,
	// 		MDIcon:              DEVICE_DRIVER_DEFAULT_ICON,
	// 	},
	// }); err != nil {
	// 	return "", nil
	// }

	query, err := db.Prepare(`
	INSERT INTO
	deviceDriver(
		VendorId,
		ModelId,
		Name,
		Version,
		HomescriptCode
	)
	VALUES(?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Error("Failed to create new device driver: preparing query failed: ", err.Error())
		return "", err
	}
	defer query.Close()

	// Create the device driver
	if _, err = query.Exec(
		driverData.VendorId,
		driverData.ModelId,
		driverData.Name,
		driverData.Version,
		driverData.HomescriptCode,
	); err != nil {
		log.Error("Failed to create new device driver: executing query failed: ", err.Error())
		return "", err
	}

	return driverData.HomescriptCode, nil
}

// Modifies the metadata of a given homescript
// Does not check the validity of the homescript's id
func ModifyDeviceDriver(newData DeviceDriver) error {
	query, err := db.Prepare(`
	UPDATE deviceDriver
	SET
		Name=?,
		Version=?,
		HomescriptId=?
	WHERE VendorId=? AND ModelId=?
	`)
	if err != nil {
		log.Error("Failed to update device driver: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	_, err = query.Exec(
		newData.Driver.Name,
		newData.Driver.Version,
		newData.Homescript.Data.Id,
		newData.Driver.VendorId,
		newData.Driver.ModelId,
	)
	if err != nil {
		log.Error("Failed to update device driver: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Returns a list of homescripts owned by a given user
func ListDeviceDrivers() ([]DeviceDriverData, error) {
	query, err := db.Prepare(`
	SELECT
		deviceDriver.VendorId,
		deviceDriver.ModelId,
		deviceDriver.Name,
		deviceDriver.Version,
		deviceDriver.HomescriptCode
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
	var drivers []DeviceDriverData = make([]DeviceDriverData, 0)
	for res.Next() {
		var driver DeviceDriverData
		err := res.Scan(
			&driver.VendorId,
			&driver.ModelId,
			&driver.Name,
			&driver.Version,
			&driver.HomescriptCode,
		)
		if err != nil {
			log.Error("Failed to list Homescript of user: scanning results failed: ", err.Error())
			return nil, err
		}
		drivers = append(drivers, driver)
	}
	return drivers, nil
}

func GetDeviceDriver(vendorId string, modelId string) (DeviceDriverData, bool, error) {
	query, err := db.Prepare(`
	SELECT
		deviceDriver.VendorId,
		deviceDriver.ModelId,
		deviceDriver.Name,
		deviceDriver.Version,
		deviceDriver.HomescriptCode
	FROM deviceDriver
	WHERE deviceDriver.VendorId=?
	AND deviceDriver.ModelId=?
	`)
	if err != nil {
		log.Error("Failed to get device driver: preparing query failed: ", err.Error())
		return DeviceDriverData{}, false, err
	}
	defer query.Close()

	var driver DeviceDriverData

	if err := query.QueryRow(
		driver.VendorId,
		driver.ModelId,
	).Scan(
		&driver.VendorId,
		&driver.ModelId,
		&driver.Name,
		&driver.Version,
		&driver.HomescriptCode,
	); err != nil {
		if err == sql.ErrNoRows {
			return DeviceDriverData{}, false, nil
		}
		log.Error("Failed to get device driver: executing query failed: ", err.Error())
		return DeviceDriverData{}, false, err
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

	_, err = query.Exec()
	return err
}
