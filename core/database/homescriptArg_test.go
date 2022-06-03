package database

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateHomescriptArgTable(t *testing.T) {
	if err := createHomescriptArgTable(); err != nil {
		t.Error(err.Error())
	}
}

func TestHomescritArgs(t *testing.T) {
	if err := initDB(true); err != nil {
		t.Error(err.Error())
	}
	if err := CreateNewHomescript(Homescript{Data: HomescriptData{Id: "arg_test"}, Owner: "admin"}); err != nil {
		t.Error(err.Error())
	}
	table := []struct {
		Data  HomescriptArg
		Error string
	}{
		{
			Data: HomescriptArg{
				HomescriptId: "error",
				Data: HomescriptArgData{
					Prompt:    "enter something",
					InputType: String,
					Display:   TypeDefault,
				},
			},
			Error: "Error 1452: Cannot add or update a child row: a foreign key constraint fails (`smarthome`.`homescriptArg`, CONSTRAINT `homescriptArg_ibfk_1` FOREIGN KEY (`HomescriptId`) REFERENCES `homescript` (`Id`))",
		},
		{
			Data: HomescriptArg{
				HomescriptId: "arg_test",
				Data: HomescriptArgData{
					Prompt:    "enter something",
					InputType: String,
					Display:   TypeDefault,
				},
			},
			Error: "",
		},
	}
	for testIndex, test := range table {
		t.Run(fmt.Sprintf("homescript_args/test/%d", testIndex), func(t *testing.T) {
			// Add the argument to the database
			newId, err := AddHomescriptArg(test.Data)
			if err != nil {
				if test.Error == "" {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.Error, err.Error())
				return
			}
			assert.Empty(t, test.Error)
			test.Data.Id = newId
		})
	}
}
