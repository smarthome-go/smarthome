package homescript

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/user"
)

func TestMain(m *testing.M) {
	log := logrus.New()
	log.Level = logrus.TraceLevel
	InitLogger(log)
	event.InitLogger(log)
	user.InitLogger(log)
	InitManager()
	if err := initDB(true); err != nil {
		panic(err.Error())
	}

	// Homescript driver value cache initialization
	if err := driver.Manager.PopulateValueCache(); err != nil {
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
