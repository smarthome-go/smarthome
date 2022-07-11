package homescript

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

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
	InitManager()
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
		Owner: "admin",
		Data: database.HomescriptData{
			Id:   "test",
			Code: "print('exec works', getArg('key'))",
		},
	}); err != nil {
		t.Error(err.Error())
		return
	}
	if err := database.CreateNewHomescript(database.Homescript{
		Owner: "test",
		Data: database.HomescriptData{
			Id:   "test2",
			Code: "print('exec should not work')",
		},
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
				FirstError: "Could not get power state of switch 'does_not_exist': switch does not exists",
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
			Code: "print(exec('test'))",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "",
				Code:       1,
				FirstError: "Homescript terminated with exit code 1: Failed to retrieve argument 'key': not provided to the Homescript runtime",
			},
		},
		{
			Code: "print(exec('test', pair('key', 'value')))",
			Result: struct {
				Output     string
				Code       int
				FirstError string
			}{
				Output:     "exec worksvalue",
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
		output, code, errors := HmsManager.Run(
			"admin",
			"testing",
			test.Code,
			false,
			make(map[string]string, 0),
			make([]string, 0),
			InitiatorInternal,
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

// Is used in order to test the recursion detector and call stack implementation
func TestRecursion(t *testing.T) {
	/* Recursive code */
	t.Run("test_prevent_recursion", func(t *testing.T) {
		// A script which calls another script which then calls the start again
		if err := database.CreateNewHomescript(database.Homescript{
			Owner: "admin",
			Data: database.HomescriptData{
				Id:   "recursive-start",
				Code: "exec('recursive-end')",
			},
		}); err != nil {
			t.Error(err.Error())
		}
		if err := database.CreateNewHomescript(database.Homescript{
			Owner: "admin",
			Data: database.HomescriptData{
				Id:   "recursive-end",
				Code: "exec('recursive-start')",
			},
		}); err != nil {
			t.Error(err.Error())
		}

		// Run the actual test
		output, exitCode, err := HmsManager.RunById(
			"recursive-start",
			"admin",
			make([]string, 0),
			false,
			make(map[string]string),
			InitiatorInternal,
		)
		assert.EqualError(t, err, "Homescript terminated with exit code 1: Homescript terminated with exit code 1: Exec violation: executing 'recursive-start' could cause infinite recursion.\n=== Call Stack ===\n   0: recursive-start (INITIAL)\n   1: recursive-end\n   2: recursive-start (PREVENTED)\n")
		assert.Equal(t, 1, exitCode)
		assert.Equal(t, "execution error", output)
	})

	/* Non-recursive code */
	t.Run("test_no_false_positives", func(t *testing.T) {
		// A normal script which calls another one multiple times
		// Useful for checking if the recursion detector is too agressive and prevents executing scripts twice
		// However, the current implementation never detects false positives because of the way of how the call stack is pushed to
		if err := database.CreateNewHomescript(database.Homescript{
			Owner: "admin",
			Data: database.HomescriptData{
				Id:   "normal1",
				Code: "print(exec('normal2')); exec('normal2'); print(exec('normal3')); exec('normal3')",
			},
		}); err != nil {
			t.Error(err.Error())
		}
		if err := database.CreateNewHomescript(database.Homescript{
			Owner: "admin",
			Data: database.HomescriptData{
				Id:   "normal2",
				Code: "print(2)",
			},
		}); err != nil {
			t.Error(err.Error())
		}
		if err := database.CreateNewHomescript(database.Homescript{
			Owner: "admin",
			Data: database.HomescriptData{
				Id:   "normal3",
				Code: "print(3)",
			},
		}); err != nil {
			t.Error(err.Error())
		}

		// Run the actual test
		output2, exitCode, err := HmsManager.RunById(
			"normal1",
			"admin",
			make([]string, 0),
			false,
			make(map[string]string),
			InitiatorInternal,
		)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "23", output2)
	})
}
