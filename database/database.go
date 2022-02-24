package database

import (
	"database/sql"

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
