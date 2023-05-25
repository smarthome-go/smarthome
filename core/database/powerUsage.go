package database

import (
	"database/sql"
	"fmt"
	"time"
)

// For an API-friendly version of this struct, visit the hardware module
type PowerDataPoint struct {
	Id   uint
	Time time.Time
	On   PowerDrawData
	Off  PowerDrawData
}

type PowerDrawData struct {
	SwitchCount uint    `json:"switchCount"` // How many switches are involved in this state (how many are active / disabled)
	Watts       uint    `json:"watts"`       // The power-draw sum of these switches
	Percent     float64 `json:"percent"`     // How much percent of the total power draw this is equal to
}

func createPowerUsageTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	powerUsage(
		Id					INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		Time				DATETIME DEFAULT CURRENT_TIMESTAMP,

		OnSwitchCount		INT UNSIGNED,
		OnWatts				INT UNSIGNED,
		OnPercent			FLOAT(24) UNSIGNED,

		OffSwitchCount	INT UNSIGNED,
		OffWatts			INT UNSIGNED,
		OffPercent			FLOAT(24) UNSIGNED
	)
	`); err != nil {
		log.Error("Failed to create power usage table: executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Inserts a new data point into the power usage time record table
func AddPowerUsagePoint(onData PowerDrawData, offData PowerDrawData, entryTime time.Time) (uint, error) {
	query, err := db.Prepare(`
	INSERT INTO
	powerUsage(
		Id,
		Time,
		OnSwitchCount,
		OnWatts,
		OnPercent,
		OffSwitchCount,
		OffWatts,
		OffPercent
	)
	VALUES(
		DEFAULT, ?, ?, ?, ?, ?, ?, ?
	)
	`)
	if err != nil {
		log.Error("Failed to add power usage point: preparing query failed: ", err.Error())
		return 0, err
	}
	defer query.Close()
	res, err := query.Exec(
		entryTime,
		// On data
		onData.SwitchCount,
		onData.Watts,
		onData.Percent,
		// Off data
		offData.SwitchCount,
		offData.Watts,
		offData.Percent,
	)
	if err != nil {
		log.Error("Failed to add power usage point: executing query failed: ", err.Error())
		return 0, err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		log.Error("Failed to add power usage point: obtaining id failed: ", err.Error())
		return 0, err
	}
	return uint(newId), nil
}

// Returns records from the power usage records
// Only returns records which are younger than x hours
// If the max-age is set to < 0, all records are returned
func GetPowerUsageRecords(maxAgeHours int) ([]PowerDataPoint, error) {
	rawQuery := `
	SELECT
		Id,
		Time,

		OnSwitchCount,
		OnWatts,
		OnPercent,

		OffSwitchCount,
		OffWatts,
		OffPercent
	FROM powerUsage
	`

	if maxAgeHours >= 0 {
		rawQuery += `
		WHERE
			Time > NOW() - INTERVAL ? HOUR
		`
	}

	query, err := db.Prepare(rawQuery)
	if err != nil {
		log.Error("Failed to get power usage records: preparing query failed: ", err.Error())
		return nil, err
	}
	defer query.Close()

	// Decide if the argument (max-age) should be passed
	var res *sql.Rows
	if maxAgeHours >= 0 {
		res, err = query.Query(maxAgeHours)
	} else {
		res, err = query.Query()
	}

	if err != nil {
		log.Error("Failed to get power usage records: executing query failed: ", err.Error())
		return nil, err
	}
	// Append the results to the output slice
	records := make([]PowerDataPoint, 0)
	for res.Next() {
		var row PowerDataPoint
		var rowTime sql.NullTime
		// Scan the results into the current row
		if err := res.Scan(
			// Time + id data
			&row.Id,
			&rowTime,
			// On data
			&row.On.SwitchCount,
			&row.On.Watts,
			&row.On.Percent,
			// Off data
			&row.Off.SwitchCount,
			&row.Off.Watts,
			&row.Off.Percent,
		); err != nil {
			log.Error("Failed to get power usage records: scanning query results failed: ", err.Error())
			return nil, err
		}
		// Validate that the scanned time is valid
		if !rowTime.Valid {
			log.Error("Failed to get power usage records: time value is invalid")
			return nil, fmt.Errorf("Failed to get power usage records: time value is invalid")
		}
		// If the time is valid, set it in the actual row
		row.Time = rowTime.Time
		// Append the row to the output
		records = append(records, row)
	}
	return records, err
}

// Deletes power statistics which are older than x hours
// Also returns the amount of records which have been deleted by this query
func FlushPowerUsageRecords(olderThanHours uint) (uint, error) {
	query, err := db.Prepare(`
	DELETE FROM powerUsage
	WHERE Time < NOW() - INTERVAL ? HOUR
	`)
	if err != nil {
		log.Error("Failed to flush old power usage records: preparing query failed: ", err.Error())
		return 0, err
	}
	defer query.Close()
	res, err := query.Exec(olderThanHours)
	if err != nil {
		log.Error("Failed to flush old power usage records: executing query failed: ", err.Error())
		return 0, err
	}
	deletedRecords, err := res.RowsAffected()
	if err != nil {
		log.Error("Failed to flush old power usage records: obtaining affected rows failed: ", err.Error())
		return 0, err
	}
	return uint(deletedRecords), nil
}

// Deletes a power usage data point given its id
func DeletePowerUsagePointById(id uint) error {
	query, err := db.Prepare(`
	DELETE FROM powerUsage
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to delete power usage record by id: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err := query.Exec(id); err != nil {
		log.Error("Failed to delete power usage record by id: executing query failed: ", err.Error())
		return err
	}
	return nil
}
