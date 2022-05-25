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
	Port     int    `json:"port"`
}

func InitLogger(logger *logrus.Logger) {
	log = logger
}

type DBStatus struct {
	OpenConnections int `json:"openConnections"`
	InUse           int `json:""`
	Idle            int `json:""`
}

func GetDatabaseStats() DBStatus {
	return DBStatus{
		OpenConnections: db.Stats().OpenConnections,
		InUse:           db.Stats().InUse,
		Idle:            db.Stats().Idle,
	}
}

// Closes the database
func Shutdown() error {
	return db.Close()
}

func DeleteTables() error {
	tables := []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"DROP TABLE IF EXISTS hasCameraPermission",
		"DROP TABLE IF EXISTS camera",
		"DROP TABLE IF EXISTS rooms",
		"DROP TABLE IF EXISTS hasSwitchPermission",
		"DROP TABLE IF EXISTS switch",
		"DROP TABLE IF EXISTS schedule",
		"DROP TABLE IF EXISTS automation",
		"DROP TABLE IF EXISTS homescript",
		"DROP TABLE IF EXISTS notifications",
		"DROP TABLE IF EXISTS hasPermission",
		"DROP TABLE IF EXISTS permission",
		"DROP TABLE IF EXISTS user",
		"DROP TABLE IF EXISTS hardware",
		"DROP TABLE IF EXISTS logs",
		"SET FOREIGN_KEY_CHECKS = 1",
	}
	for _, query := range tables {
		_, err := db.Exec(query)
		if err != nil {
			log.Error(fmt.Sprintf("Failed to drop table %s : executing query failed: %s", query, err.Error()))
			return err
		}
	}
	log.Warn("Database has been deleted")
	return nil
}
