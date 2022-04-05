package database

import (
	"testing"
)

func TestCreateLoggingEventTable(t *testing.T) {
	if err := createLoggingEventTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestLogging(t *testing.T) {
	table := []LogEvent{
		{
			Name:        "name",
			Description: "description",
			Level:       0,
		},
		{
			Name:        "name",
			Description: "description",
			Level:       1,
		},
		{
			Name:        "name",
			Description: "description",
			Level:       2,
		},
		{
			Name:        "name",
			Description: "description",
			Level:       3,
		},
		{
			Name:        "name",
			Description: "description",
			Level:       4,
		},
		{
			Name:        "name",
			Description: "description",
			Level:       5,
		},
		{
			Name:        "name",
			Description: "description",
			Level:       6,
		},
	}
	for _, item := range table {
		if err := AddLogEvent(item.Name, item.Description, item.Level); err != nil {
			t.Error(err.Error())
			return
		}
	}
	logs, err := GetLogs()
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, logItem := range logs {
		found := false
		for _, v := range table {
			if v.Description == logItem.Description &&
				v.Name == logItem.Name &&
				v.Level == logItem.Level {
				found = true
			}
		}
		if !found {
			t.Errorf("Log item %v has not been found in dataset or it's metadata is invalid", logItem)
			return
		}
	}
	if err := FlushOldLogs(); err != nil {
		t.Error(err.Error())
		return
	}
	if err := FlushAllLogs(); err != nil {
		t.Error(err.Error())
		return
	}
	logs, err = GetLogs()
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(logs) > 0 {
		t.Errorf("Amount of logs after deletion is supposed to be 0 but is actually :%d", len(logs))
		return
	}
}
