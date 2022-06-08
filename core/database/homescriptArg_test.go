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

func TestHomescriptArgs(t *testing.T) {
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
				Data: HomescriptArgData{
					HomescriptId: "error",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    String,
					Display:      TypeDefault,
				},
			},
			Error: "Error 1452: Cannot add or update a child row: a foreign key constraint fails (`smarthome`.`homescriptArg`, CONSTRAINT `homescriptArg_ibfk_1` FOREIGN KEY (`HomescriptId`) REFERENCES `homescript` (`Id`))",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    "invalid",
					Display:      TypeDefault,
				},
			},
			Error: "Error 1265: Data truncated for column 'InputType' at row 1",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    String,
					Display:      "invalid",
				},
			},
			Error: "Error 1265: Data truncated for column 'Display' at row 1",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    String,
					Display:      TypeDefault,
				},
			},
			Error: "",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    String,
					Display:      TypeDefault,
				},
			},
			Error: "",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    String,
					Display:      StringSwitches,
				},
			},
			Error: "",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    Boolean,
					Display:      TypeDefault,
				},
			},
			Error: "",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    Boolean,
					Display:      BooleanOnOff,
				},
			},
			Error: "",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    Boolean,
					Display:      BooleanYesNo,
				},
			},
			Error: "",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    Number,
					Display:      TypeDefault,
				},
			},
			Error: "",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    Number,
					Display:      NumberHour,
				},
			},
			Error: "",
		},
		{
			Data: HomescriptArg{
				Data: HomescriptArgData{
					HomescriptId: "arg_test",
					Prompt:       "enter something",
					MDIcon:       "code",
					InputType:    Number,
					Display:      NumberMinute,
				},
			},
			Error: "",
		},
	}
	for testIndex, test := range table {
		t.Run(fmt.Sprintf("homescript_args/add/iter-%d", testIndex), func(t *testing.T) {
			// Add the argument to the database
			newId, err := AddHomescriptArg(test.Data.Data)
			if err != nil {
				if test.Error == "" {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.Error, err.Error())
				return
			} else if newId == 0 {
				t.Errorf("Newly created ID should not be 0\n")
			}
			table[testIndex].Data.Id = newId
			assert.Empty(t, test.Error)
		})
	}
	t.Run("homescript_args/mirror", func(t *testing.T) {
		for _, item := range table {
			data, found, err := GetUserHomescriptArgById(item.Data.Id, "admin")
			assert.NoError(t, err)
			if item.Error == "" {
				assert.True(t, found)
				assert.Equal(t, item.Data, data)
			} else {
				assert.False(t, found)
				assert.Empty(t, data)
			}
		}
	})
	t.Run("homescript_args/list_by_id", func(t *testing.T) {
		tableDataTemp := make([]HomescriptArg, 0)
		for _, item := range table {
			if item.Error == "" {
				tableDataTemp = append(tableDataTemp, item.Data)
			}
		}
		fromDb, err := ListArgsOfHomescript("arg_test")
		assert.NoError(t, err)
		assert.Equal(t, tableDataTemp, fromDb)
	})
	t.Run("homescript_args/delete_all_sequential", func(t *testing.T) {
		fromDb, err := ListArgsOfHomescript("arg_test")
		assert.NoError(t, err)
		for _, item := range fromDb {
			err := DeleteHomescriptArg(item.Id)
			assert.NoError(t, err)
		}
		fromDbEmpty, err := ListArgsOfHomescript("arg_test")
		assert.NoError(t, err)
		assert.Empty(t, fromDbEmpty)
	})
	t.Run("homescript_args/delete_all_together", func(t *testing.T) {
		// Add the test data first
		for _, item := range table {
			_, err := AddHomescriptArg(item.Data.Data)
			if item.Error == "" {
				assert.NoError(t, err)
			}
		}
		// Validate creation
		fromDbFull, err := ListArgsOfHomescript("arg_test")
		assert.NoError(t, err)
		assert.NotEmpty(t, fromDbFull)

		// Delete all arguments at once
		assert.NoError(t, DeleteAllHomescriptArgsFromScript("arg_test"))

		// Validate deletion
		fromDbEmpty, err := ListArgsOfHomescript("arg_test")
		assert.NoError(t, err)
		assert.Empty(t, fromDbEmpty)

		// Query each argument individually in order to check if the `GetById` function works
		for _, item := range table {
			data, found, err := GetUserHomescriptArgById(item.Data.Id, "admin")
			assert.NoError(t, err)
			assert.False(t, found)
			assert.Empty(t, data)
		}
	})
}
