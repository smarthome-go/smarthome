package config

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

func TestRunSetup(t *testing.T) {
	setup := Setup{
		HardwareNodes: []database.HardwareNode{
			{
				Name:    "test",
				Online:  false,
				Enabled: false,
				Url:     "",
				Token:   "",
			},
		},
		Rooms: []database.Room{
			{
				Id:          "test",
				Name:        "test",
				Description: "test",
				Switches: []database.Switch{
					{
						Id:      "test",
						Name:    "test",
						RoomId:  "test",
						PowerOn: false,
						Watts:   0,
					},
					{
						Id:      "test2",
						Name:    "test2",
						RoomId:  "test2",
						PowerOn: false,
						Watts:   0,
					},
				},
				Cameras: []database.Camera{
					{
						Id:     1,
						RoomId: "test",
						Url:    "",
						Name:   "test",
					},
				},
			},
		},
	}
	if err := RunSetup(&setup); err != nil {
		t.Error(err.Error())
		return
	}
	for _, switchItem := range setup.Rooms[0].Switches {
		exists, err := database.DoesSwitchExist(switchItem.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !exists {
			t.Errorf("Switch %s does not exist after setup", switchItem.Id)
			return
		}
	}
	nodes, err := database.GetHardwareNodes()
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, setupNode := range setup.HardwareNodes {
		nodesvalid := false
		for _, node := range nodes {
			if node.Url == setupNode.Url && node.Name == setupNode.Name && node.Token == setupNode.Token {
				nodesvalid = true
			}
		}
		if !nodesvalid {
			t.Errorf("Node %s does not exists after creation", setupNode.Url)
			return
		}
	}
	rooms, err := database.ListRooms()
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, setupRoom := range setup.Rooms {
		roomValid := false
		for _, room := range rooms {
			if room.Id == setupRoom.Id && room.Description == setupRoom.Description {
				roomValid = true
			}
		}
		if !roomValid {
			t.Errorf("Room %s does not exist after creation", setupRoom.Id)
			return
		}
	}
}
