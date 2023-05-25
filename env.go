package main

import (
	"math"
	"os"
	"strconv"

	"github.com/smarthome-go/smarthome/core/config"
)

// Scans environment variables
// `SMARTHOME_ADMIN_PASSWORD`: (String) If specified, the admin user that is created on first launch will receive this password instead of `admin`
// `SMARTHOME_ENV_PRODUCTION`: (Bool  ) Whether the server should use production presets
// `SMARTHOME_SESSION_KEY`   : (String) (Only during production) Specifies a manual key for session encryption (used for larger instances): random key generation is skipped
// `SMARTHOME_DB_DATABASE`   : (String) Sets the database name
// `SMARTHOME_DB_HOSTNAME`   : (String) Sets the database hostname
// `SMARTHOME_DB_PORT`       : (Int   ) Sets the database port
// `SMARTHOME_DB_PASSWORD`   : (String) Sets the database user's password
// `SMARTHOME_DB_USER`       : (String) Sets the database user
func scanEnv(configStruct *config.Config) string {
	// Admin passord
	newAdminPassword := "admin"
	if adminPassword, adminPasswordOk := os.LookupEnv("SMARTHOME_ADMIN_PASSWORD"); adminPasswordOk {
		newAdminPassword = adminPassword
	}

	// Operational mode
	if productionEnvStr, productionEnvStrOk := os.LookupEnv("SMARTHOME_ENV_PRODUCTION"); productionEnvStrOk {
		switch productionEnvStr {
		case "TRUE":
			configStruct.Server.Production = true
			log.Debug("Detected `SMARTHOME_ENV_PRODUCTION` (TRUE), server will start using production presets")
		case "FALSE":
			configStruct.Server.Production = false
			log.Debug("Detected `SMARTHOME_ENV_PRODUCTION` (FALSE), server will start in development mode")
		default:
			log.Warn("Could not use `SMARTHOME_ENV_PRODUCTION` as boolean value, using development mode\nValid modes are `TRUE` and `FALSE`")
		}
	}
	// Web server session-key
	if sessionKey, sessionKeyOk := os.LookupEnv("SMARTHOME_SESSION_KEY"); sessionKeyOk {
		if !configStruct.Server.Production {
			log.Warn("Using manually specified session encryption key during development mode. This will have no effect unless using production")
		} else {
			if configStruct.Server.SessionKey != "" {
				log.Debug("Selected SMARTHOME_SESSION_KEY over value from config file")
			} else {
				log.Info("Using manually specified session encryption key from SMARTHOME_SESSION_KEY")
			}
			configStruct.Server.SessionKey = sessionKey
		}
	}
	// DB variables
	if dbUsername, dbUsernameOk := os.LookupEnv("SMARTHOME_DB_USER"); dbUsernameOk {
		log.Debug("Selected SMARTHOME_DB_USER over value from config file")
		configStruct.Database.Username = dbUsername
	}
	if dbPassword, dbPasswordOk := os.LookupEnv("SMARTHOME_DB_PASSWORD"); dbPasswordOk {
		log.Debug("Selected SMARTHOME_DB_PASSWORD over value from config file")
		configStruct.Database.Password = dbPassword
	}
	if dbDatabase, dbDatabaseOk := os.LookupEnv("SMARTHOME_DB_DATABASE"); dbDatabaseOk {
		log.Debug("Selected SMARTHOME_DB_DATABASE over value from config file")
		configStruct.Database.Database = dbDatabase
	}
	if dbHostname, dbHostnameOk := os.LookupEnv("SMARTHOME_DB_HOSTNAME"); dbHostnameOk {
		log.Debug("Selected SMARTHOME_DB_HOSTNAME over value from config file")
		configStruct.Database.Hostname = dbHostname
	}
	if dbPort, dbPortOk := os.LookupEnv("SMARTHOME_DB_PORT"); dbPortOk {
		portInt, err := strconv.Atoi(dbPort)
		if err != nil || portInt > math.MaxUint16 || portInt < 0 {
			log.Warn("Could not parse `SMARTHOME_DB_PORT` to uint16, using value from config.json")
		} else {
			log.Debug("Selected SMARTHOME_DB_PORT over value from config file")
			configStruct.Database.Port = uint16(portInt)
		}
	}
	// Port of the webserver
	if webPort, webPortOk := os.LookupEnv("SMARTHOME_PORT"); webPortOk {
		webPortInt, err := strconv.Atoi(webPort)
		if err != nil || webPortInt > math.MaxUint16 || webPortInt < 0 {
			log.Warn("Could not parse `SMARTHOME_PORT` to uint16, using 8082")
		} else {
			log.Debug("Selected `SMARTHOME_PORT` over default")
			port = uint16(webPortInt)
		}
	}

	return newAdminPassword
}
