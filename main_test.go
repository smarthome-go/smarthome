package main

import (
	"testing"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/sirupsen/logrus"
)

func TestMain(t *testing.T) {
	log := logrus.New()
	database.InitLogger(log)
	if err := database.Init(database.DatabaseConfig{
		Username: "smarthome",
		Password: "testing",
		Hostname: "localhost",
		Database: "smarthome",
		Port:     3330,
	}, "admin",
	); err != nil {
		t.Error(err.Error())
		return
	}
}
