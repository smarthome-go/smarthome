package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const DEFAULT_ADMIN_USERNAME = "admin"

func Init(databaseConfig DatabaseConfig, adminPassword string) error {
	log.Trace("Initializing database connection...")
	config = databaseConfig
	if err := createDatabase(); err != nil {
		return err
	}
	dbTemp, err := connection()
	if err != nil {
		return err
	}
	log.Trace("Initializing database schema...")
	db = dbTemp
	if err := createConfigTable(); err != nil {
		return err
	}
	if err := createUserTable(); err != nil {
		return err
	}
	if err := createUserTokenTable(); err != nil {
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
	if err := createPowerUsageTable(); err != nil {
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
	if err := createHasCameraPermissionsTable(); err != nil {
		return err
	}
	if err := createHardwareNodeTable(); err != nil {
		return err
	}
	if err := createHomescriptTable(); err != nil {
		return err
	}
	if err := createHomescriptArgTable(); err != nil {
		return err
	}
	// if err := createHomescriptOwnerTable(); err != nil {
	// 	return err
	// }
	if err := createDeviceDriverTable(); err != nil {
		return err
	}
	if err := createDeviceTable(); err != nil {
		return err
	}
	if err := createHasDevicePermissionTable(); err != nil {
		return err
	}
	if err := createAutomationTable(); err != nil {
		return err
	}
	if err := createScheduleTable(); err != nil {
		return err
	}
	if err := createSchedulerSwitchesTable(); err != nil {
		return err
	}
	if err := createReminderTable(); err != nil {
		return err
	}
	if err := createHomescriptStorageTable(); err != nil {
		return err
	}
	if err := createWeatherTable(); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Successfully initialized database `%s`", databaseConfig.Database))
	return nil
}

// Is used on the very first start to create the `smarthome` database
// If the `smarthome` database already exists, the existent version is used
func createDatabase() error {
	dbTemp, err := sql.Open(
		"mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/",
			config.Username,
			config.Password,
			config.Hostname,
			config.Port,
		))
	if err != nil {
		log.Error("Could not connect to intermediate database: ", err.Error())
		return err
	}
	defer dbTemp.Close()
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := dbTemp.ExecContext(
		ctx,
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.Database),
	)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create database `%s`: executing query failed: %s", config.Database, err.Error()))
		return err
	}
	log.Trace("Successfully connected to database using intermediate connection")
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error(fmt.Sprintf("Failed to evaluate outcome of database creation: reading rows affected failed: %s", err.Error()))
		return err
	}
	if rowsAffected == 1 {
		log.Info(fmt.Sprintf("Successfully created new database `%s`", config.Database))
	} else {
		log.Debug(fmt.Sprintf("Skipped database creation: using existing database `%s`", config.Database))
	}
	return nil
}

// Is used to initialize an `admin` user when initializing the database
func initAdminUser(password string) error {
	// The admin user is only created if no user with user-management privileges is present
	userAdminExists := false
	users, err := ListUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		hasPermission, err := UserHasPermission(
			user.Username,
			PermissionManageUsers,
		)
		if err != nil {
			return err
		}
		if hasPermission {
			userAdminExists = true
		}
	}
	if userAdminExists {
		return nil
	}

	if err := AddUser(FullUser{
		Username:          DEFAULT_ADMIN_USERNAME,
		Forename:          "Admin",
		Surname:           "User",
		PrimaryColorDark:  "#88FF70",
		PrimaryColorLight: "#2E7D32",
		Password:          password,
	}); err != nil {
		if err.Error() != "could not add user: user already exists" {
			return err
		}
	}
	if err := AddUserPermission("admin", "*"); err != nil {
		log.Error("Failed to create admin user: permission setup failed: ", err.Error())
		return err
	}
	return nil
}
