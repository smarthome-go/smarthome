package database

import (
	"time"

	"github.com/sirupsen/logrus"
)

func initDB(args ...bool) error {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	if err := Init(DatabaseConfig{
		Username: "smarthome",
		Password: "testing",
		Hostname: "localhost",
		Database: "smarthome",
		Port:     3330,
	}, "admin",
	); err != nil {
		return err
	}
	if len(args) > 0 {
		if err := DeleteTables(); err != nil {
			return err
		}
		time.Sleep(time.Second)
		initDB()
	}
	return nil
}
