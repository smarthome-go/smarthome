package main

import (
	"fmt"

	"github.com/MikMuellerDev/smarthome/database"
	"github.com/MikMuellerDev/smarthome/utils"
	"github.com/sirupsen/logrus"
)

const version = "0.0.1"

func main() {
	// Create new logger
	log, err := utils.NewLogger(logrus.TraceLevel)
	if err != nil {
		panic(err.Error())
	}
	// Initialize <module> loggers
	utils.InitLogger(log)
	database.InitLogger(log)
	log.Trace("Logging initialized.")

	// Read config file
	err = utils.ReadConfigFile()
	if err != nil {
		log.Fatal("Failed to read config file: startup halted.")
	}
	config := utils.GetConfig()
	log.Trace("Loaded config file")

	// Initialize database
	err = database.Init(config.Database)
	if err != nil {
		panic(err.Error())
	}

	database.AddUser(database.User{Username: "mik", Password: "password"})
	database.AddUserPermission("mik", "foo")
	database.AddUserPermission("mik", "bar")
	database.AddUserPermission("mik", "authentication")
	database.AddUserPermission("mik", "baz")
	fmt.Println(database.GetUserPermissions("mik"))
	// database.DeleteUser("mik")

	users, err := database.ListUsers()
	if err != nil {
		log.Error("Failed to obtain user list: ", err.Error())
	}
	fmt.Println(users)

	log.Info(fmt.Sprintf("Smarthome v%s is running.", version))
}
