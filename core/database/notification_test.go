package database

import "testing"

func TestCreateNotificationTable(t *testing.T) {
	if err := createNotificationTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestNotifications(t *testing.T) {
	notifications, err := GetUserNotifications("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	notificationCountOld, err := GetUserNotificationCount("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if len(notifications) != int(notificationCountOld) {
		t.Errorf("Notification length and notification count are not the same: len: %d count: %d", len(notifications), notificationCountOld)
		return
	}

	table := []Notification{
		{
			Priority:    1,
			Name:        "name1",
			Description: "description1",
		},
		{
			Priority:    2,
			Name:        "name2",
			Description: "description2",
		},
		{
			Priority:    3,
			Name:        "name3",
			Description: "description3",
		},
	}
	for _, item := range table {
		if err := AddNotification(
			"admin",
			item.Name,
			item.Description,
			item.Priority,
		); err != nil {
			t.Error(err.Error())
			return
		}
		notifications, err := GetUserNotifications("admin")
		if err != nil {
			t.Error(err.Error())
			return
		}
		found := false
		for _, v := range notifications {
			if v.Name == item.Name &&
				v.Description == item.Description &&
				v.Priority == item.Priority {
				found = true
			}
		}
		if !found {
			t.Errorf("Notification %v not found in database or metadata is invalid", item)
			return
		}
	}
	notifications, err = GetUserNotifications("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if err := DeleteNotificationFromUserById(notifications[0].Id, "admin"); err != nil {
		t.Errorf("Notification %d could not be deleted: %s", notifications[0].Id, err.Error())
		return
	}
	notificationCountNew, err := GetUserNotificationCount("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if int(notificationCountNew) != len(notifications)-1 {
		t.Errorf("Did deletion fail? count does not match expectation: want: %d got: %d", len(notifications)-1, notificationCountNew)
		return
	}
	if err := DeleteAllNotificationsFromUser("admin"); err != nil {
		t.Error(err.Error())
		return
	}
	notificationCountDel, err := GetUserNotificationCount("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if notificationCountDel > 0 {
		t.Errorf("Notification count after deleting all does not match expectation: want: 0 got: %d", notificationCountDel)
		return
	}
}
