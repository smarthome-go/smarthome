package event

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/MikMuellerDev/smarthome/core/database"
)

func TestMain(m *testing.M) {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	if err := initDB(true); err != nil {
		panic(err.Error())
	}
	code := m.Run()
	os.Exit(code)
}

func initDB(args ...bool) error {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	database.InitLogger(log)
	if err := database.Init(database.DatabaseConfig{
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
		if err := database.DeleteTables(); err != nil {
			return err
		}
		time.Sleep(time.Second)
		return initDB()
	}
	return nil
}

func TestLogEvent(t *testing.T) {
	table := []struct {
		Name        string
		Description string
		Level       int
	}{
		{Name: "test1", Description: "test1", Level: 0},
		{Name: "test1", Description: "test1", Level: 1},
		{Name: "test2", Description: "test2", Level: 2},
		{Name: "test3", Description: "test3", Level: 3},
		{Name: "test4", Description: "test4", Level: 4},
		{Name: "test5", Description: "test5", Level: 5},
	}
	logs, err := database.GetLogs()
	if err != nil {
		t.Error(err.Error())
		return
	}
	lenPrev := len(logs)
	for _, tableItem := range table {
		if err := logEvent(tableItem.Name, tableItem.Description, tableItem.Level); err != nil {
			t.Error(err.Error())
			return
		}
	}
	logs, err = database.GetLogs()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if lenPrev+len(table) != len(logs) {
		t.Errorf("Log test failed: want %d logs in database, got %d", lenPrev+len(table), len(logs))
	}
}

func TestDeleteLogs(t *testing.T) {
	if err := database.FlushAllLogs(); err != nil {
		t.Error(err.Error())
		return
	}
	logs, err := database.GetLogs()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(logs) != 0 {
		t.Errorf("Log deletion test failed: want 0 logs, got %d", len(logs))
	}
	if err := database.FlushOldLogs(); err != nil {
		t.Error(err.Error())
		return
	}
}
