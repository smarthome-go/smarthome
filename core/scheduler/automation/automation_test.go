package automation

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/MikMuellerDev/smarthome/core/database"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/hardware"
	"github.com/MikMuellerDev/smarthome/core/homescript"
)

func TestMain(m *testing.M) {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	if err := initDB(true); err != nil {
		panic(err.Error())
	}
	if err := database.CreateRoom("test_room", "", ""); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch", "", "test_room", 0); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch_modify", "", "test_room", 0); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch_inactive", "", "test_room", 0); err != nil {
		panic(err.Error())
	}
	if err := database.CreateSwitch("test_switch_abort", "", "test_room", 0); err != nil {
		panic(err.Error())
	}
	_, doesExists, err := database.GetUserHomescriptById("test", "admin")
	if err != nil {
		panic(err.Error())
	}
	if !doesExists {
		// Create Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Id:                  "test",
			Owner:               "admin",
			Name:                "Testing",
			Description:         "A Homescript for testing purposes",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "switch('test_switch', on)",
		}); err != nil {
			panic(err.Error())
		}
	}
	_, doesExists, err = database.GetUserHomescriptById("test_modify", "admin")
	if err != nil {
		panic(err.Error())
	}
	if !doesExists {
		// Create another Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Id:                  "test_modify",
			Owner:               "admin",
			Name:                "Testing 2",
			Description:         "Another Homescript for testing purposes",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "switch('test_switch_modify', on)",
		}); err != nil {
			panic(err.Error())
		}
	}
	_, doesExists, err = database.GetUserHomescriptById("test_inactive", "admin")
	if err != nil {
		panic(err.Error())
	}
	if !doesExists {
		// Create another Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Id:                  "test_inactive",
			Owner:               "admin",
			Name:                "Testing 2",
			Description:         "Another Homescript for testing purposes",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "switch('test_switch_inactive', on)",
		}); err != nil {
			panic(err.Error())
		}
	}
	_, doesExists, err = database.GetUserHomescriptById("test_abort", "admin")
	if err != nil {
		panic(err.Error())
	}
	if !doesExists {
		// Create another Homescript
		if err := database.CreateNewHomescript(database.Homescript{
			Id:                  "test_abort",
			Owner:               "admin",
			Name:                "Testing 2",
			Description:         "Another Homescript for testing purposes",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "switch('test_switch_abort', on)",
		}); err != nil {
			panic(err.Error())
		}
	}
	code := m.Run()
	os.Exit(code)
}

func initDB(args ...bool) error {
	database.InitLogger(logrus.New())
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

// Todo: organize function
func TestAutomation(t *testing.T) {
	now := time.Now()
	then := now.Add(time.Minute)
	log := logrus.New()
	log.Level = logrus.FatalLevel
	event.InitLogger(log)
	homescript.InitLogger(log)
	hardware.InitLogger(log)
	// Flush all logs before automation runs
	if err := database.FlushAllLogs(); err != nil {
		t.Error(err.Error())
		return
	}
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
	if _, err := CreateNewAutomation(
		"name",
		"description",
		uint8(then.Hour()),
		uint8(then.Minute()),
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test",
		"admin",
		true,
		database.TimingNormal,
	); err != nil {
		t.Error(err.Error())
		return
	}
	modifyId, err := CreateNewAutomation(
		"name",
		"description",
		uint8(then.Hour()),
		uint8(then.Minute()),
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test",
		"admin",
		true,
		database.TimingNormal,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if _, err := CreateNewAutomation(
		"name",
		"description",
		uint8(then.Hour()),
		uint8(then.Minute()),
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test_inactive",
		"admin",
		false,
		database.TimingNormal,
	); err != nil {
		t.Error(err.Error())
		return
	}
	abortId, err := CreateNewAutomation(
		"name",
		"description",
		uint8(then.Hour()),
		uint8(then.Minute()),
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test_abort",
		"admin",
		true,
		database.TimingNormal,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	// Modify second automation to use the other homescript file
	cronExpression, err := GenerateCronExpression(uint8(then.Hour()), uint8(then.Minute()), []uint8{0, 1, 2, 3, 4, 5, 6})
	if err != nil {
		t.Error(err.Error())
		return
	}
	if err := ModifyAutomationById(abortId,
		database.AutomationWithoutIdAndUsername{
			Name:           "name",
			Description:    "description",
			CronExpression: cronExpression,
			HomescriptId:   "test_abort",
			Enabled:        false,
			TimingMode:     database.TimingNormal,
		}); err != nil {
		t.Error(err.Error())
		return
	}
	if err := ModifyAutomationById(modifyId,
		database.AutomationWithoutIdAndUsername{
			Name:           "name",
			Description:    "description",
			CronExpression: cronExpression,
			HomescriptId:   "test_modify",
			Enabled:        true,
			TimingMode:     database.TimingNormal,
		}); err != nil {
		t.Error(err.Error())
		return
	}
	valid := false
	for i := 0; i < 30; i++ {
		time.Sleep(time.Second * 5)
		power, err := database.GetPowerStateOfSwitch("test_switch")
		if err != nil {
			t.Error(err.Error())
			return
		}
		if power {
			valid = true
			break
		}
	}
	if !valid {
		t.Error("Power of `test_switch` did not change")
		return
	}
	power, err := database.GetPowerStateOfSwitch("test_switch_modify")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !power {
		t.Error("Power of `test_switch_modify` did not change: want: true got: false")
		return
	}
	power, err = database.GetPowerStateOfSwitch("test_switch_inactive")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if power {
		t.Error("Power of `test_switch_inactive` changed: want: false got: true")
		return
	}
	power, err = database.GetPowerStateOfSwitch("test_switch_abort")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if power {
		t.Error("Power of `test_switch_abort` changed: want: false got: true")
		return
	}
}

func TestTimingModes(t *testing.T) {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	event.InitLogger(log)
	homescript.InitLogger(log)
	hardware.InitLogger(log)
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
	sunriseId, err := CreateNewAutomation(
		"name",
		"description",
		23,
		59,
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test",
		"admin",
		true,
		database.TimingSunrise,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	sunSetId, err := CreateNewAutomation(
		"name",
		"description",
		23,
		59,
		[]uint8{0, 1, 2, 3, 4, 5, 6},
		"test",
		"admin",
		true,
		database.TimingSunset,
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
	sunrise, found, err := database.GetAutomationById(sunriseId)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation %d was not found after creation", sunriseId)
		return
	}
	sunSet, found, err := database.GetAutomationById(sunSetId)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Automation %d was not found after creation", sunSetId)
		return
	}
	if sunrise.CronExpression == sunSet.CronExpression {
		t.Errorf("Cron expression of sunrise and sunset is not valid. `%s`|`%s`", sunSet.CronExpression, sunrise.CronExpression)
	}
}

func TestInit(t *testing.T) {
	if err := Init(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestActivate(t *testing.T) {
	TestInit(t)
	if err := DeactivateAutomationSystem(); err != nil {
		t.Error(err.Error())
	}
	if err := ActivateAutomationSystem(); err != nil {
		t.Error(err.Error())
	}
}
