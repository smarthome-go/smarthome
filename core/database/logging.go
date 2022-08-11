package database

import (
	"database/sql"
	"fmt"
	"time"
)

type LogEvent struct {
	Id          uint
	Name        string
	Description string
	Level       LogLevel
	Time        time.Time
}

type LogLevel uint

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// Creates (unless it exists) the table containing internal logging events
// For example a user logging in or altering a power states
func createLoggingEventTable() error {
	if _, err := db.Exec(`
	CREATE TABLE
	IF NOT EXISTS
	logs(
		Id INT AUTO_INCREMENT,
		Name VARCHAR(100),
		Description TEXT,
		Level INT,
		Date DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (Id)
	)`); err != nil {
		log.Error("Could not create logging table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

// Add a logged internal event based on `name`, `description`, and `level`
func AddLogEvent(name string, description string, level LogLevel) error {
	query, err := db.Prepare(`
	INSERT INTO
	logs(
		Id,
		Name,
		Description,
		Level,
		Date
	)
	VALUES (DEFAULT, ?, ?, ?, DEFAULT)`)
	if err != nil {
		log.Error("Failed to add log event: preparing query failed: ", err.Error())
		return err
	}
	defer query.Close()
	if _, err = query.Exec(name, description, level); err != nil {
		log.Error("Failed to add log event: preparing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes log events older than 30 days in order to free storage space
// This function will later be used by a scheduler for daily jobs
func FlushOldLogs() (uint, error) {
	res, err := db.Exec(`
	DELETE FROM logs
	WHERE
	Date < NOW() - INTERVAL 30 DAY
	`)
	if err != nil {
		log.Error("Failed to flush old log events: failed to execute query: ", err.Error())
		return 0, err
	}
	deletedMessages, err := res.RowsAffected()
	if err != nil {
		log.Error("Could not evaluate outcome of `FlushOldLogs`: ", err.Error())
		return 0, err
	}
	return uint(deletedMessages), nil
}

// Deletes all logs which are currently stored in the database
func FlushAllLogs() (uint, error) {
	res, err := db.Exec(`
	DELETE FROM logs
	`)
	if err != nil {
		log.Error("Failed to flush all log events: failed to execute query: ", err.Error())
		return 0, err
	}
	deletedMessages, err := res.RowsAffected()
	if err != nil {
		log.Error("Could not evaluate outcome of `FlushAllLogs`: ", err.Error())
		return 0, err
	}
	return uint(deletedMessages), nil
}

// Deletes a log record matching the provided id
// Also returns a boolean indicating whether an entry has been deleted or not
func DeleteLogById(id uint) (bool, error) {
	query, err := db.Prepare(`
	DELETE FROM logs
	WHERE Id=?
	`)
	if err != nil {
		log.Error("Failed to delete log record: failed to prepare query: ", err.Error())
		return false, err
	}
	defer query.Close()
	res, err := query.Exec(id)
	if err != nil {
		log.Error("Failed to delete log record: failed to prepare query: ", err.Error())
		return false, err
	}
	rowAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("Failed to delete log record: failed to get affected rows: ", err.Error())
		return false, err
	}
	return rowAffected > 0, nil
}

// Returns all logs currently in the database
func GetLogs() ([]LogEvent, error) {
	res, err := db.Query(`
	SELECT
		Id,
		Name,
		Description,
		Level,
		Date
	FROM logs`)
	if err != nil {
		log.Error("Could not get all logs: failed to execute query: ", err.Error())
		return nil, err
	}
	defer res.Close()
	logs := make([]LogEvent, 0)
	for res.Next() {
		var logItem LogEvent
		var logTime sql.NullTime
		err := res.Scan(
			&logItem.Id,
			&logItem.Name,
			&logItem.Description,
			&logItem.Level,
			&logTime,
		)
		if err != nil {
			log.Error("Could not list all logs: Failed to scan results ", err.Error())
			return nil, err
		}
		if !logTime.Valid {
			log.Error("Invalid time column when scanning logs")
			return nil, fmt.Errorf("invalid time column when scanning logs")
		} else {
			logItem.Time = logTime.Time
			logs = append(logs, logItem)
		}
	}
	return logs, err
}
