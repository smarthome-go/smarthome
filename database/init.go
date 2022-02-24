package database

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func Init(databaseConfig DatabaseConfig) error {
	config = databaseConfig
	dbTemp, err := connection()
	if err != nil {
		return err
	}
	db = dbTemp
	err = createDatabase()
	if err != nil {
		return err
	}
	err = createUserTable()
	if err != nil {
		return err
	}
	err = createPermissionsTable()
	if err != nil {
		return err
	}
	return nil
}

func createDatabase() error {
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + config.Database)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create database `%s`: Query execution error: %s", config.Database, err.Error()))
		return err
	}
	log.Info(fmt.Sprintf("Successfully initialized database `%s`.", config.Database))
	return nil
}
