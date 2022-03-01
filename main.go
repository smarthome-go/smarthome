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

const version = "0.0.1"
const port = "8082"

func main() {
	// Create new logger
	log, err := utils.NewLogger(logrus.TraceLevel)
	if err != nil {
		panic(err.Error())
	}
	// Initialize <module> loggers
	utils.InitLogger(log)
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
	log.Trace("Loaded config file")

	hardware.InitConfig(config.Hardware)

	// Initialize database
	if err := database.Init(config.Database); err != nil {
		panic(err.Error())
	}
	if err := database.AddUser(database.User{Username: "admin", Password: "admin"}); err != nil {
		if err.Error() != "could not add user: user already exists" {
			panic(err.Error())
		}
	}
	r := routes.NewRouter()
	middleware.Init(config.Server.Production)
	// TODO: replace with config variable for random seed
	templates.LoadTemplates("./web/html/*.html")
	http.Handle("/", r)

	// TODO: Remove
	success := hardware.ExecuteJob("s2", true)
	fmt.Printf("Success: %t\n", success)
	success = hardware.ExecuteJob("s2", false)
	fmt.Printf("Success: %t\n", success)

	log.Info(fmt.Sprintf("Smarthome v%s is running.", version))
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
}
