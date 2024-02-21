package homescript

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/user"
)

func TestMain(m *testing.M) {
	log := logrus.New()
	log.Level = logrus.TraceLevel
	InitLogger(log)
	event.InitLogger(log)
	user.InitLogger(log)
	hardware.InitLogger(log)
	hardware.Init()
	InitManager()
	if err := initDB(true); err != nil {
		panic(err.Error())
	}

	// Homescript driver value cache initialization
	if err := PopulateValueCache(); err != nil {
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

// TODO: implement these tests again

// func TestRun(t *testing.T) {
// 	assert.NoError(t, initDB(true))
//
// 	// Create a mock switch and room
// 	if err := database.CreateRoom(database.RoomData{Id: "test"}); err != nil {
// 		t.Error(err.Error())
// 		return
// 	}
// 	if err := database.CreateDevice("test", "", "test", 0, nil); err != nil {
// 		t.Error(err.Error())
// 		return
// 	}
// 	if err := database.AddUser(database.FullUser{
// 		Username: "test",
// 	}); err != nil {
// 		t.Error(err.Error())
// 		return
// 	}
// 	// Add homescript which will be later used for exec
// 	if err := database.CreateNewHomescript(database.Homescript{
// 		Owner: "admin",
// 		Data: database.HomescriptData{
// 			Id:   "test",
// 			Code: "print(ARGS.key);",
// 		},
// 	}); err != nil {
// 		t.Error(err.Error())
// 		return
// 	}
// 	if err := database.CreateNewHomescript(database.Homescript{
// 		Owner: "test",
// 		Data: database.HomescriptData{
// 			Id:   "test2",
// 			Code: "print('exec should not work')",
// 		},
// 	}); err != nil {
// 		t.Error(err.Error())
// 		return
// 	}
// 	table := []struct {
// 		Code   string
// 		Result struct {
// 			Output     string
// 			Code       int
// 			FirstError string
// 		}
// 	}{
// 		{
// 			Code: "println(user);",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "admin\n",
// 				Code:       0,
// 				FirstError: "",
// 			},
// 		},
// 		{
// 			Code: "println('Hello World');",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "Hello World\n",
// 				Code:       0,
// 				FirstError: "",
// 			},
// 		},
// 		{
// 			Code: "println('Hello World';",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       1,
// 				FirstError: "Unclosed function call: Expected r-paren, found semicolon",
// 			},
// 		},
// 		{
// 			Code: "switch('test', on); println(get_switch('test').power);",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       1,
// 				FirstError: "Failed to set power: hardware error: There are no hardware nodes, power state unaffected",
// 			},
// 		},
// 		{
// 			Code: "switch('test', off); println(get_switch('test').power);",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       1,
// 				FirstError: "Failed to set power: hardware error: There are no hardware nodes, power state unaffected",
// 			},
// 		},
// 		{
// 			Code: "switch('does_not_exist', on);",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       1,
// 				FirstError: "Failed to set power: switch 'does_not_exist' does not exist",
// 			},
// 		},
// 		{
// 			Code: "print(get_switch('does_not_exist'));",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       1,
// 				FirstError: "switch 'does_not_exist' was not found",
// 			},
// 		},
// 		{
// 			Code: "notify('', '', 1);",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       0,
// 				FirstError: "",
// 			},
// 		},
// 		{
// 			Code: "notify('', '', 2);",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       0,
// 				FirstError: "",
// 			},
// 		},
// 		{
// 			Code: "notify('', '', 3);",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       0,
// 				FirstError: "",
// 			},
// 		},
// 		{
// 			Code: "notify('', '', 4);",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       1,
// 				FirstError: "notification level has to be one of 1, 2, or 3, got 4",
// 			},
// 		},
// 		{
// 			Code: "print(exec('test'));",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       1,
// 				FirstError: "TypeError: object has no member named key (1:12)",
// 			},
// 		},
// 		{
// 			Code: "print(exec('test', 'key' => 'value').value);",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "valuenull",
// 				Code:       0,
// 				FirstError: "",
// 			},
// 		},
// 		{
// 			Code: "exec('test2');",
// 			Result: struct {
// 				Output     string
// 				Code       int
// 				FirstError string
// 			}{
// 				Output:     "",
// 				Code:       1,
// 				FirstError: "invalid Homescript id: no data associated with id",
// 			},
// 		},
// 	}
// 	for _, test := range table {
// 		var buffer bytes.Buffer
// 		res := HmsManager.Run(
// 			"admin",
// 			"testing",
// 			test.Code,
// 			make(map[string]string, 0),
// 			make([]string, 0),
// 			InitiatorInternal,
// 			make(chan int),
// 			&buffer,
// 			nil,
// 			make(map[string]homescript.Value),
// 		)
// 		if len(res.Errors) > 0 {
// 			if res.Errors[0].Message != test.Result.FirstError {
// 				t.Errorf("Unmatched error: want: %s got: %s", test.Result.FirstError, res.Errors[0].Message)
// 				return
// 			}
// 		} else if test.Result.FirstError != "" {
// 			t.Errorf("Expected abundant error: expected: %s, got none", test.Result.FirstError)
// 			return
// 		}
// 		if res.ExitCode != test.Result.Code {
// 			t.Errorf("Unexpected exit code. want: `%d` got: `%d`", test.Result.Code, res.ExitCode)
// 			return
// 		}
// 		output := buffer.String()
// 		if output != test.Result.Output {
// 			t.Errorf("Unexpected output: want: `%s` got: `%s`", test.Result.Output, output)
// 			return
// 		}
// 	}
// }
//
// // Is used in order to test the recursion detector and call stack implementation
// func TestRecursion(t *testing.T) {
// 	/* Recursive code */
// 	t.Run("test_prevent_recursion", func(t *testing.T) {
// 		// A script which calls another script which then calls the start again
// 		if err := database.CreateNewHomescript(database.Homescript{
// 			Owner: "admin",
// 			Data: database.HomescriptData{
// 				Id:   "recursive-start",
// 				Code: "exec('recursive-end');",
// 			},
// 		}); err != nil {
// 			t.Error(err.Error())
// 		}
// 		if err := database.CreateNewHomescript(database.Homescript{
// 			Owner: "admin",
// 			Data: database.HomescriptData{
// 				Id:   "recursive-end",
// 				Code: "exec('recursive-start');",
// 			},
// 		}); err != nil {
// 			t.Error(err.Error())
// 		}
//
// 		// Run the actual test
// 		res, err := HmsManager.RunById(
// 			"recursive-start",
// 			"admin",
// 			make([]string, 0),
// 			make(map[string]string),
// 			InitiatorInternal,
// 			make(chan int),
// 			nil, nil,
// 			make(map[string]homescript.Value),
// 		)
// 		assert.NoError(t, err)
// 		if len(res.Errors) == 0 {
// 			t.Errorf("Expected error, received none")
// 			return
// 		}
// 		if !strings.Contains(res.Errors[0].Message, "Exec violation") {
// 			t.Errorf("Expected exec violation error, got: %s: %s (%d:%d)", res.Errors[0].Kind, res.Errors[0].Message, res.Errors[0].Span.Start.Line, res.Errors[0].Span.Start.Column)
// 		}
// 		assert.Equal(t, 1, res.ExitCode)
// 	})
//
// 	/* Non-recursive code */
// 	t.Run("test_no_false_positives", func(t *testing.T) {
// 		// A normal script which calls another one multiple times
// 		// Useful for checking if the recursion detector is too aggressive and prevents executing scripts twice
// 		// However, the current implementation never detects false positives because of the way of how the call stack is pushed to
// 		if err := database.CreateNewHomescript(database.Homescript{
// 			Owner: "admin",
// 			Data: database.HomescriptData{
// 				Id:   "normal1",
// 				Code: "println(exec('normal2').value); exec('normal2'); println(exec('normal3').value); exec('normal3');",
// 			},
// 		}); err != nil {
// 			t.Error(err.Error())
// 		}
// 		if err := database.CreateNewHomescript(database.Homescript{
// 			Owner: "admin",
// 			Data: database.HomescriptData{
// 				Id:   "normal2",
// 				Code: "println(2);",
// 			},
// 		}); err != nil {
// 			t.Error(err.Error())
// 		}
// 		if err := database.CreateNewHomescript(database.Homescript{
// 			Owner: "admin",
// 			Data: database.HomescriptData{
// 				Id:   "normal3",
// 				Code: "println(3);",
// 			},
// 		}); err != nil {
// 			t.Error(err.Error())
// 		}
//
// 		var buffer bytes.Buffer
// 		// Run the actual test
// 		res, err := HmsManager.RunById(
// 			"normal1",
// 			"admin",
// 			make([]string, 0),
// 			make(map[string]string),
// 			InitiatorInternal,
// 			make(chan int),
// 			&buffer,
// 			nil,
// 			make(map[string]homescript.Value),
// 		)
// 		assert.NoError(t, err)
// 		if len(res.Errors) != 0 {
// 			fmt.Printf("%s: %s (%d:%d)", res.Errors[0].Kind, res.Errors[0].Message, res.Errors[0].Span.Start.Line, res.Errors[0].Span.Start.Column)
// 		}
// 		assert.Equal(t, 0, res.ExitCode)
// 		assert.Equal(t, "2\nnull\n2\n3\nnull\n3\n", buffer.String())
// 	})
// }
