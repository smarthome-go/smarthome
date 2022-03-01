package main

import (
	"fmt"
	"net/http"

	"github.com/MikMuellerDev/smarthome/core/database"
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
	log.Trace("Logging initialized.")

	// Read config file
	if err := utils.ReadConfigFile(); err != nil {
		log.Fatal("Failed to read config file: startup halted.")
	}
	config := utils.GetConfig()
	log.Trace("Loaded config file")

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
	middleware.Init(true)
	// TODO: replace with config variable for random seed
	templates.LoadTemplates("./web/html/*.html")
	http.Handle("/", r)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
	log.Info(fmt.Sprintf("Smarthome v%s is running.", version))
}
