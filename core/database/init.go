package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Init(databaseConfig DatabaseConfig, adminPassword string) error {
	config = databaseConfig
	if err := createDatabase(); err != nil {
		return err
	}
	dbTemp, err := connection()
	if err != nil {
		return err
	}
	db = dbTemp
	if err := CreateConfigTable(); err != nil {
		return err
	}
	if err := createUserTable(); err != nil {
		return err
	}
	if err := createPermissionTable(); err != nil {
		return err
	}
	if err := initializePermissions(); err != nil {
		return err
	}
	if err := createHasPermissionTable(); err != nil {
		return err
	}
	if err := createRoomTable(); err != nil {
		return err
	}
	if err := createSwitchTable(); err != nil {
		return err
	}
	if err := createHasSwitchPermissionTable(); err != nil {
		return err
	}
	if err := createNotificationTable(); err != nil {
		return err
	}
	if err := initAdminUser(adminPassword); err != nil {
		return err
	}
	if err := createLoggingEventTable(); err != nil {
		return err
	}
	if err := createCameraTable(); err != nil {
		return err
	}
	if err := createHardwareNodeTable(); err != nil {
		return err
	}
	if err := createHomescriptTable(); err != nil {
		return err
	}
	if err := createAutomationTable(); err != nil {
		return err
	}
	if err := createScheduleTable(); err != nil {
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
		log.Info(fmt.Sprintf("Successfully initialized database `%s`", config.Database))
	} else {
		log.Debug(fmt.Sprintf("Using existing database `%s`", config.Database))
	}
	defer dbTemp.Close()
	return nil
}

func initAdminUser(password string) error {
	if err := AddUser(FullUser{
		Username:     "admin",
		Firstname:    "Admin",
		Surname:      "User",
		PrimaryColor: "#88ff70",
		Password:     password,
	}); err != nil {
		if err.Error() != "could not add user: user already exists" {
			return err
		}
	}
	if _, err := AddUserPermission("admin", "*"); err != nil {
		log.Error("Failed to create admin user: permission setup failed: ", err.Error())
		return err
	}
	return nil
}
