package database

import (
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

// Test the creation, deletion and retrieval of reminders
func TestReminders(t *testing.T) {
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
		_, exists, err := GetReminderById(id, i.Owner)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !exists {
			t.Errorf("Reminder '%d' does not exist after creation", id)
			return
		}
		_, exists, err = GetReminderById(id, "invalid_owner")
		if err != nil {
			t.Error(err.Error())
			return
		}
		if exists {
			t.Errorf("Reminder '%d' exists for invalid owner", id)
			return
		}
	}
	reminders, err := GetUserReminders(table[0].Owner)
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, test := range table {
		valid := false
		for _, reminder := range reminders {
			if reminder.Name == test.Name &&
				reminder.Description == test.Description &&
				reminder.Priority == test.Priority &&
				reminder.Owner == test.Owner {
				valid = true
			}
		}
		if !valid {
			t.Errorf("Reminder '%s' was not found in the database", test.Name)
			return
		}
	}
	if err := DeleteAllRemindersFromUser(table[0].Owner); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestDeleteUserReminderById(t *testing.T) {
	// Create a test user in order to test if another user is allowed to delete foreign reminders
	if err := AddUser(FullUser{Username: "reminder"}); err != nil {
		t.Error(err.Error())
		return
	}
	// Create a reminder for the user
	id, err := CreateNewReminder("reminder", "reminder", time.Now(), "reminder", Low)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if err := DeleteUserReminderById("admin", id); err != nil {
		t.Error(err.Error())
		return
	}
	reminders, err := GetUserReminders("reminder")
	if err != nil {
		t.Error(err.Error())
		return
	}
	// If this fails, users may be able to delete foreign reminders
	if len(reminders) == 0 {
		t.Errorf("Length of user reminders after deletion is 0 but should not.")
		return
	}
	// Modify the reminder
	if err := ModifyReminder(
		id,
		"new name",
		"new description",
		time.Date(1999, 6, 0, 0, 0, 0, 0, time.Now().Location()),
		Urgent,
	); err != nil {
		t.Error(err.Error())
		return
	}
	// Check if the modification succeeded
	reminder, _, err := GetReminderById(id, "reminder")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if reminder.Id != id ||
		reminder.Name != "new name" ||
		reminder.Description != "new description" ||
		reminder.Owner != "reminder" ||
		reminder.DueDate.Year() != 1999 ||
		reminder.Priority != Urgent {
		t.Errorf("Reminder has invalid metadata after modification got: %v", reminder)
		return
	}
}
