package database

import (
	"database/sql"
	"fmt"
)

func databaseConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.Username, config.Password, config.Hostname, config.Port, config.Database)
}

func connection() (*sql.DB, error) {
	dbTemp, err := sql.Open("mysql", databaseConnectionString())
	if err != nil {
		log.Error("Could not connect to MYSQL: ", err.Error())
		return nil, err
	}
	return dbTemp, nil
}
