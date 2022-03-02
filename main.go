package main

import (
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/config"
	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/hardware"
	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/MikMuellerDev/smarthome/core/utils"
	"github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/MikMuellerDev/smarthome/server/routes"
	"github.com/MikMuellerDev/smarthome/server/templates"
	"github.com/sirupsen/logrus"
)

const version = "0.0.2"
const port = "8082"

func main() {
	// Create new logger
	log, err := utils.NewLogger(logrus.TraceLevel)
	if err != nil {
		panic(err.Error())
	}

	// TODO: check if every module has got a corresponding logger
	// Initialize <module> loggers
	utils.InitLogger(log)
	config.InitLogger(log)
	database.InitLogger(log)
	middleware.InitLogger(log)
	routes.InitLogger(log)
	templates.InitLogger(log)
	user.InitLogger(log)
	hardware.InitLogger(log)

	// Read config file
	if err := config.ReadConfigFile(); err != nil {
		log.Fatal("Failed to read config file: startup halted.")
	}
	config := config.GetConfig()
	hardware.InitConfig(config.Hardware)
	log.Trace("Loaded config")

	// Initialize database
	if err := database.Init(config.Database); err != nil {
		panic(err.Error())
	}
	// TODO: move user creation to somewhere else
	if err := database.AddUser(database.User{Username: "admin", Password: "admin"}); err != nil {
		if err.Error() != "could not add user: user already exists" {
			panic(err.Error())
		}
	}
	// TODO: move this somewhere else
	for _, room := range config.Rooms {
		for _, switchItem := range room.Switches {
			if err := database.CreateSwitch(switchItem.Id, switchItem.Name); err != nil {
				log.Error("Could not create switches from config file:")
				panic(err.Error())
			}
		}
	}
	r := routes.NewRouter()
	middleware.Init(config.Server.Production)
	templates.LoadTemplates("./web/html/*.html")
	http.Handle("/", r)
	log.Info(fmt.Sprintf("Smarthome v%s is running.", version))
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
}
