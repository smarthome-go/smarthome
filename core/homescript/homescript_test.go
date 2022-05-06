package homescript

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
)

func TestMain(m *testing.M) {
	log := logrus.New()
	log.Level = logrus.FatalLevel
	InitLogger(log)
	event.InitLogger(log)
	hardware.InitLogger(log)
	hardware.Init()
	if err := initDB(true); err != nil {
		panic(err.Error())
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

func TestRun(t *testing.T) {
	// Create a mock switch and room
	if err := database.CreateRoom(database.RoomData{Id: "test"}); err != nil {
		t.Error(err.Error())
		return
	}
	if err := database.CreateSwitch("test", "", "test", 0); err != nil {
		t.Error(err.Error())
		return
	}
	if err := database.AddUser(database.FullUser{
		Username: "test",
	}); err != nil {
		t.Error(err.Error())
		return
	}
	// Add homescript which will be later used for exec
	if err := database.CreateNewHomescript(database.Homescript{
		Id:    "test",
		Owner: "admin",
		Code:  "print('exec works')",
	}); err != nil {
		t.Error(err.Error())
		return
	}
	if err := database.CreateNewHomescript(database.Homescript{
		Id:    "test2",
		Owner: "test",
		Code:  "print('exec should not work')",
	}); err != nil {
		t.Error(err.Error())
		return
	}
	table := []struct {
		Code   string
		Result struct {
			Output     string
			Code       int
			FirstError string
		}
	}{
		{
			Code: "print(user)",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "admin",
				Code:       0,
				FirstError: "",
			},
		},
		{
			Code: "print('hello world')",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "hello world",
				Code:       0,
				FirstError: "",
			},
		},
		{
			Code: "print('hello world'",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "",
				Code:       1,
				FirstError: "Expected ')', found 'EOF'",
			},
		},
		{
			Code: "switch('test', on); print(switchOn('test'))",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "true",
				Code:       0,
				FirstError: "",
			},
		},
		{
			Code: "switch('test', off); print(switchOn('test'))",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "false",
				Code:       0,
				FirstError: "",
			},
		},
		{
			Code: "switch('does_not_exist', on)",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "",
				Code:       1,
				FirstError: "Failed to set power: switch 'does_not_exist' does not exist",
			},
		},
		{
			Code: "print(switchOn('does_not_exist'))",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "",
				Code:       1,
				FirstError: "can not get power state of switch 'does_not_exist': switch does not exists",
			},
		},
		{
			Code: "notify('', '', 1)",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "",
				Code:       0,
				FirstError: "",
			},
		},
		{
			Code: "notify('', '', 2)",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "",
				Code:       0,
				FirstError: "",
			},
		},
		{
			Code: "notify('', '', 3)",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "",
				Code:       0,
				FirstError: "",
			},
		},
		{
			Code: "notify('', '', 4)",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "",
				Code:       1,
				FirstError: "Notification level has to be one of 1, 2, or 3, got 4",
			},
		},
		{
			Code: "play('', '')",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "",
				Code:       1,
				FirstError: "The feature 'radiGo' is not yet implemented",
			},
		},
		{
			Code: "print(exec('test'))",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "exec works",
				Code:       0,
				FirstError: "",
			},
		},
		{
			Code: "exec('test2')",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "",
				Code:       1,
				FirstError: "Invalid Homescript id: no data associated with id",
			},
		},
	}
	for _, test := range table {
		output, code, errors := Run(
			"admin", "testing", test.Code,
		)
		if len(errors) > 0 {
			if errors[0].Message != test.Result.FirstError {
				t.Errorf("Unmatched error. want: %s got: %s", test.Result.FirstError, errors[0].Message)
				return
			}
		} else if test.Result.FirstError != "" {
			t.Errorf("Expected abundant error: expected: %s, got none", test.Result.FirstError)
			return
		}
		if code != test.Result.Code {
			t.Errorf("Unexpected exit code. want: `%d` got: `%d`", test.Result.Code, code)
			return
		}
		if output != test.Result.Output {
			t.Errorf("Unexpected output: want: `%s` got: `%s`", test.Result.Output, output)
			return
		}
	}
}
