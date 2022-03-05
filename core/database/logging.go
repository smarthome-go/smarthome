package database

import "fmt"

// Creates (if not exists) the table containing internal (mostly non-error) loggin events
// For example a user logging in or altering power states
func createLoggingEventTable() error {
	query := `
	 CREATE TABLE IF NOT EXISTS logs(
		 Id INT AUTO_INCREMENT,
		 Name VARCHAR(100),
		 Description TEXT,
		 Level INT,
		 Date DATETIME DEFAULT CURRENT_TIMESTAMP,
		 PRIMARY KEY (Id)
	 )
	 `
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Could not create logging table: Executing query failed: ", err.Error())
		return err
	}
	return nil
}

func AddLogEvent(name string, description string, level int) error {
	query, err := db.Prepare(`
	INSERT INTO logs(
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
	_, err = query.Exec(name, description, level)
	if err != nil {
		log.Error("Failed to add log event: preparing query failed: ", err.Error())
		return err
	}
	return nil
}

// Deletes log events older than 30 days in order to save storage space
// This function will later be used by a scheduler for daily jobs
func FlushOldLogs() error {
	query := `
	DELETE FROM logs
	WHERE Date < NOW() - INTERVAL 30 DAY
	`
	res, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to flush old log events: failed to execute query: ", err.Error())
		return err
	}
	deletedMessages, err := res.RowsAffected()
	if err != nil {
		log.Error("Could not evaluate outcome of `FlushOldLogs`: ", err.Error())
		return err
	}
	log.Debug(fmt.Sprintf("Successfully flushed old log messages: deleted %d messages", deletedMessages))
	return nil
}

func FlushAllLogs() error {
	query := `DELETE FROM logs`
	res, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to flush all log events: failed to execute query: ", err.Error())
		return err
	}
	deletedMessages, err := res.RowsAffected()
	if err != nil {
		log.Error("Could not evaluate outcome of `FlushAllLogs`: ", err.Error())
		return err
	}
	log.Debug(fmt.Sprintf("Successfully flushed all log messages: deleted %d messages", deletedMessages))
	return nil
}
