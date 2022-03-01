package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Init(databaseConfig DatabaseConfig) error {
	config = databaseConfig
	err := createDatabase()
	if err != nil {
		return err
	}
	dbTemp, err := connection()
	if err != nil {
		return err
	}
	db = dbTemp
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
	dbTemp, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", config.Username, config.Password, config.Hostname, config.Port))
	if err != nil {
		log.Error("Could not connect to Database: ", err.Error())
	}
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := dbTemp.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+config.Database)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create database `%s`: Query execution error: %s", config.Database, err.Error()))
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error(fmt.Sprintf("Failed to get result of `create database` `%s`: Reading rows affected failed with error: %s", config.Database, err.Error()))
		return err
	}
	if rowsAffected == 1 {
		log.Info(fmt.Sprintf("Successfully initialized database `%s`: %d rows affected", config.Database, rowsAffected))
	} else {
		log.Debug(fmt.Sprintf("Using existing database `%s`", config.Database))
	}
	defer dbTemp.Close()
	return nil
}
