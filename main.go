package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/MikMuellerDev/smarthome/core/config"
	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/hardware"
	"github.com/MikMuellerDev/smarthome/core/user"
	"github.com/MikMuellerDev/smarthome/core/utils"
	"github.com/MikMuellerDev/smarthome/server/middleware"
	"github.com/MikMuellerDev/smarthome/server/routes"
	"github.com/MikMuellerDev/smarthome/server/templates"
	"github.com/sirupsen/logrus"
)

const version = "0.0.3"

var port = 8082

func main() {
	// Create new logger
	log, err := utils.NewLogger(logrus.TraceLevel)
	if err != nil {
		panic(err.Error())
	}

	// Initialize <module> loggers
	utils.InitLogger(log)
	config.InitLogger(log)
	database.InitLogger(log)
	middleware.InitLogger(log)
	routes.InitLogger(log)
	templates.InitLogger(log)
	user.InitLogger(log)
	hardware.InitLogger(log)
	event.InitLogger(log)

	// Read config file
	if err := config.ReadConfigFile(); err != nil {
		log.Fatal("Failed to read config file: startup halted.")
	}
	config := config.GetConfig()
	hardware.InitConfig(config.Hardware)
	log.Debug("Loaded and successfully initialized config")

	// check for environment variables
	if dbUsername, dbUsernameOk := os.LookupEnv("SMARTHOME_DB_USER"); dbUsernameOk {
		log.Debug("Selected SMARTHOME_DB_USER over value from config file")
		config.Database.Username = dbUsername
	}
	if dbPassword, dbPasswordOk := os.LookupEnv("SMARTHOME_DB_PASSWORD"); dbPasswordOk {
		log.Debug("Selected SMARTHOME_DB_PASSWORD over value from config file")
		config.Database.Password = dbPassword
	}
	if dbDatabase, dbDatabaseOk := os.LookupEnv("SMARTHOME_DB_DATABASE"); dbDatabaseOk {
		log.Debug("Selected SMARTHOME_DB_DATABASE over value from config file")
		config.Database.Database = dbDatabase
	}
	if dbHostname, dbHostnameOk := os.LookupEnv("SMARTHOME_DB_HOSTNAME"); dbHostnameOk {
		log.Debug("Selected SMARTHOME_DB_HOSTNAME over value from config file")
		config.Database.Hostname = dbHostname
	}
	if webPort, webPortOk := os.LookupEnv("SMARTHOME_PORT"); webPortOk {
		webPortInt, err := strconv.Atoi(webPort)
		if err != nil {
			log.Warn("Could not parse `SMARTHOME_PORT` to int, using 8082")
		} else {
			log.Debug("Selected `SMARTHOME_PORT` over default")
			port = webPortInt
		}
	}
	if dbPort, dbPortOk := os.LookupEnv("SMARTHOME_DB_PORT"); dbPortOk {
		portInt, err := strconv.Atoi(dbPort)
		if err != nil {
			log.Warn("Could not parse `SMARTHOME_DB_PORT` to int, using value from config.json")
		} else {
			log.Debug("Selected SMARTHOME_DB_PORT over value from config file")
			config.Database.Port = portInt
		}
	}

	// Initialize database
	var dbErr error = nil
	for i := 0; i <= 5; i++ {
		dbErr = database.Init(config.Database, config.Rooms)
		if dbErr == nil {
			break
		} else {
			log.Warn("Failed to connect to database, retrying in 2 seconds")
			time.Sleep(time.Second * 5)
		}
	}
	if dbErr != nil {
		log.Error("Failed to connect to database after 10 retries, exiting now")
		panic(dbErr.Error())
	}

	// TODO: Move this to for example the makefile (via curl and API): only used during development
	if userAlreadyExists, _ := database.DoesUserExist("mik"); !userAlreadyExists {
		// if err := database.AddUser(database.User{Username: "mik", Password: "test"}); err != nil {
		if err := database.AddUser(database.User{Username: "mik", Password: "test"}); err != nil {
			log.Error("Could not create a new user in the database: ", err.Error())
			return
		}
	}
	if _, err := database.AddUserPermission("mik", "getUserSwitches"); err != nil {
		log.Fatal(err.Error())
	}
	if _, err := database.AddUserPermission("mik", "setPower"); err != nil {
		log.Fatal(err.Error())
	}
	if _, err := database.AddUserSwitchPermission("mik", "s1"); err != nil {
		log.Error("Could not add switch to switchPermissions of the user")
		panic(err.Error())
	}
	if _, err := database.AddUserSwitchPermission("mik", "s2"); err != nil {
		log.Error("Could not add switch to switchPermissions of the user")
		panic(err.Error())
	}

	fmt.Println(database.GetUserSwitchPermissions("mik"))
	a, err := database.UserHasSwitchPermission("mik", "s2")
	if err != nil {
		panic(err.Error())
	}

	if !config.Server.Production {
		// If the server is in development mode, all logs should be flushed
		database.FlushAllLogs()
	}

	log.Info("Flushing logs older than 30 days")
	database.FlushOldLogs()

	fmt.Printf("mik has permission `s2`: %t\n", a)
	success, err := database.SetPowerState("s22", true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("success: %t\n", success)

	database.AddUserPermission("mik", "deleteOldLogs")
	database.AddUserPermission("mik", "deleteAllLogs")
	database.AddUserPermission("mik", "listLogs")
	database.AddUserPermission("mik", "uploadAvatar")
	database.AddUserPermission("mik", "deleteAvatar")
	database.AddUserPermission("mik", "addUserPermission")
	database.AddUserPermission("mik", "removeUserPermission")
	database.AddUserPermission("mik", "removeSwitchPermission")
	database.AddUserPermission("mik", "addSwitchPermission")

	r := routes.NewRouter()
	middleware.Init(config.Server.Production)
	templates.LoadTemplates("./web/html/**/*.html")
	http.Handle("/", r)
	log.Info(fmt.Sprintf("Smarthome v%s is running.", version))
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}
