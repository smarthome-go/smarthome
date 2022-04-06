package database

import (
	"fmt"
	"testing"
	"time"
)

// Creates the reminder table in order to check if the sql query works
func TestCreateReminderTable(t *testing.T) {
	if err := createReminderTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestCreateNewReminder(t *testing.T) {
	table := []Reminder{
		{
			Name:        "reminder 1",
			Description: "description 1",
			Priority:    Low,
			DueDate:     time.Now(),
			Owner:       "admin",
		},
		{
			Name:        "reminder 2",
			Description: "description 2",
			Priority:    Normal,
			DueDate:     time.Now(),
			Owner:       "admin",
		},
		{
			Name:        "reminder 3",
			Description: "description 3",
			Priority:    Medium,
			DueDate:     time.Now(),
			Owner:       "admin",
		},
		{
			Name:        "reminder 4",
			Description: "description 4",
			Priority:    High,
			DueDate:     time.Now(),
			Owner:       "admin",
		},
		{
			Name:        "reminder 4",
			Description: "description 4",
			Priority:    Urgent,
			DueDate:     time.Now(),
			Owner:       "admin",
		},
	}
	for _, i := range table {
		id, err := CreateNewReminder(
			i.Name,
			i.Description,
			i.DueDate,
			i.Owner,
			i.Priority,
		)
		if err != nil {
			t.Error(err.Error())
			return
		}
		fmt.Println(id)
	}
	reminders, found, err := GetUserReminders(table[0].Owner)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(reminders, found)
}
