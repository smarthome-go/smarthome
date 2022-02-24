package database

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func createUserTable() error {
	query := `
	CREATE TABLE
	IF NOT EXISTS
	user(
		Username VARCHAR(20) PRIMARY KEY,
		Password text
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Error("Failed to create user Table: ", err.Error())
		return err
	}
	return nil
}

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
	return nil
}

func createDatabase() error {
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + config.Database)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create database `%s`: Statement execution error: %s", config.Database, err.Error()))
		return err
	}
	log.Info(fmt.Sprintf("Successfully initialized database `%s`.", config.Database))
	return nil
}
