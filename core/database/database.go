package database

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger
var db *sql.DB
var config DatabaseConfig

type DatabaseConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hostname string `json:"hostname"`
	Database string `json:"database"`
	Port     uint16 `json:"port"`
}

func InitLogger(logger *logrus.Logger) {
	log = logger
}

type DBStatus struct {
	OpenConnections int `json:"openConnections"`
	InUse           int `json:""`
	Idle            int `json:""`
}

// Returns common database statistics
// Is used in the debug function
func GetDatabaseStats() DBStatus {
	return DBStatus{
		OpenConnections: db.Stats().OpenConnections,
		InUse:           db.Stats().InUse,
		Idle:            db.Stats().Idle,
	}
}

// Closes the database connection
func Shutdown() error {
	return db.Close()
}

// Deletes all tables in the active `smarthome` database
// This function is used in testing and could be used in the future to allow for a system reset
// Todo: tests this function better
func DeleteTables() error {
	// The queries are executed after another and represent raw SQL queries
	queries := []string{
		// Required in order to dismiss foreign key constraint errors
		"SET FOREIGN_KEY_CHECKS = 0",
		"DROP TABLE IF EXISTS automation",
		"DROP TABLE IF EXISTS camera",
		"DROP TABLE IF EXISTS configuration",
		"DROP TABLE IF EXISTS hardware",
		"DROP TABLE IF EXISTS hasCameraPermission",
		"DROP TABLE IF EXISTS hasPermission",
		"DROP TABLE IF EXISTS hasSwitchPermission",
		"DROP TABLE IF EXISTS homescript",
		"DROP TABLE IF EXISTS homescriptArg",
		"DROP TABLE IF EXISTS homescriptUrlCache",
		"DROP TABLE IF EXISTS logs",
		"DROP TABLE IF EXISTS notifications",
		"DROP TABLE IF EXISTS permission",
		"DROP TABLE IF EXISTS reminder",
		"DROP TABLE IF EXISTS room",
		"DROP TABLE IF EXISTS schedule",
		"DROP TABLE IF EXISTS scheduleSwitches",
		"DROP TABLE IF EXISTS switch",
		"DROP TABLE IF EXISTS user",
		"DROP TABLE IF EXISTS userToken",
		"DROP TABLE IF EXISTS weather",
		"SET FOREIGN_KEY_CHECKS = 1",
	}
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to execute query `%s`: %s", query, err.Error()))
			return err
		}
	}
	log.Warn("Database tables have been deleted")
	return nil
}
