package database

import (
	"testing"
)

func TestCreateScheduleTable(t *testing.T) {
	if err := createScheduleTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestSchedule(t *testing.T) {
	table := []Schedule{
		{
			Name:           "test1",
			Owner:          "admin",
			Hour:           0,
			Minute:         0,
			HomescriptCode: "",
		},
		{
			Name:           "test2",
			Owner:          "admin",
			Hour:           1,
			Minute:         1,
			HomescriptCode: "print('')",
		},
	}
	for _, item := range table {
		newId, err := CreateNewSchedule(Schedule{
			Name:           item.Name,
			Owner:          item.Owner,
			Hour:           item.Hour,
			Minute:         item.Minute,
			HomescriptCode: item.HomescriptCode,
		})
		if err != nil {
			t.Error(err.Error())
			return
		}
		schedule, found, err := GetScheduleById(newId)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !found {
			t.Errorf("Schedule %d not found after creation", newId)
			return
		}
		if schedule.HomescriptCode != item.HomescriptCode &&
			schedule.Hour != item.Hour &&
			schedule.Minute != item.Minute &&
			schedule.Name != item.Name &&
			schedule.Owner != item.Owner {
			t.Errorf("Created schedule %d has invalid metadata", schedule.Id)
			return
		}
	}
	userSchedules, err := GetUserSchedules("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, v := range userSchedules {
		found := false
		for _, tableItem := range table {
			if tableItem.Name == v.Name &&
				tableItem.HomescriptCode == v.HomescriptCode &&
				tableItem.Hour == v.Hour &&
				tableItem.Minute == v.Minute &&
				tableItem.Owner == v.Owner {
				found = true
			}
		}
		if !found {
			t.Errorf("Schedule with name: %s not found in user schedules", v.Name)
			return
		}
	}

	allSchedules, err := GetSchedules()
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, v := range allSchedules {
		found := false
		for _, tableItem := range table {
			if tableItem.Name == v.Name &&
				tableItem.HomescriptCode == v.HomescriptCode &&
				tableItem.Hour == v.Hour &&
				tableItem.Minute == v.Minute &&
				tableItem.Owner == v.Owner {
				found = true
			}
		}
		if !found {
			t.Errorf("Schedule with name: %s not found in schedules", v.Name)
			return
		}
	}
}

func TestGetExistentScheduleById(t *testing.T) {
	schedule := Schedule{
		Name:           "test1",
		Owner:          "admin",
		Hour:           1,
		Minute:         1,
		HomescriptCode: "print('a')",
	}
	newId, err := CreateNewSchedule(schedule)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fromDb, found, err := GetScheduleById(newId)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("Schedule %d not found after creation: want: %v, got: %v", newId, schedule, fromDb)
		return
	}
}

func TestGetNonExistentSchedule(t *testing.T) {
	// Test for non-existent schedule
	_, found, err := GetScheduleById(9999999)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if err != nil {
		t.Error(err.Error())
		return
	}
	if found {
		t.Errorf("Schedule 9999999 found but does not exist want: not-found, got: found")
		return
	}
}

func TestModifyDeleteSchedule(t *testing.T) {
	if err := AddUser(FullUser{Username: "modify_schedule_user"}); err != nil {
		t.Error(err.Error())
		return
	}
	table := []struct {
		Before Schedule
		After  Schedule
	}{
		{
			Before: Schedule{
				Name:           "before",
				Owner:          "admin",
				Hour:           1,
				Minute:         2,
				HomescriptCode: "print('before')",
			},
			After: Schedule{
				Name:           "after",
				Owner:          "admin",
				Hour:           3,
				Minute:         4,
				HomescriptCode: "print('after')",
			},
		},
		{
			Before: Schedule{
				Name:           "before2",
				Owner:          "modify_schedule_user",
				Hour:           5,
				Minute:         6,
				HomescriptCode: "print('before2')",
			},
			After: Schedule{
				Name:           "after2",
				Owner:          "modify_schedule_user",
				Hour:           7,
				Minute:         8,
				HomescriptCode: "print('after2')",
			},
		},
	}
	for _, test := range table {
		newId, err := CreateNewSchedule(test.Before)
		if err != nil {
			t.Error(err.Error())
			return
		}
		// Modify the schedule
		if err := ModifySchedule(newId, ScheduleWithoudIdAndUsername{
			Name:           test.After.Name,
			Hour:           test.After.Hour,
			Minute:         test.After.Minute,
			HomescriptCode: test.After.HomescriptCode,
		}); err != nil {
			t.Error(err.Error())
			return
		}
		// Get the metadata from the modified schedule
		schedule, found, err := GetScheduleById(newId)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !found {
			t.Errorf("Schedule %d not found after modification", newId)
			return
		}
		if schedule.Id != newId ||
			schedule.Name != test.After.Name ||
			schedule.HomescriptCode != test.After.HomescriptCode ||
			schedule.Hour != test.After.Hour ||
			schedule.Minute != test.After.Minute ||
			schedule.Owner != test.After.Owner {
			t.Errorf("Metadate did not change completely after modification: want: %v got: %v", test.After, schedule)
		}
		// Test deletion
		if err := DeleteScheduleById(newId); err != nil {
			t.Error(err.Error())
			return
		}
		_, found, err = GetScheduleById(newId)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if found {
			t.Errorf("Schedule %d still present after deletion", newId)
			return
		}
	}
}
