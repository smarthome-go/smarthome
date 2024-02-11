package database

import (
	"testing"
)

const scheduleOwner string = "admin"

func TestCreateScheduleTable(t *testing.T) {
	if err := createScheduleTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestSchedule(t *testing.T) {
	table := []ScheduleData{
		{
			Name:               "test1",
			Hour:               0,
			Minute:             0,
			TargetMode:         ScheduleTargetModeHMS,
			HomescriptCode:     "",
			HomescriptTargetId: "test",
		},
		{
			Name:           "test2",
			Hour:           1,
			Minute:         1,
			TargetMode:     ScheduleTargetModeCode,
			HomescriptCode: "print('Hello World!')",
		},
	}
	for _, item := range table {
		newId, err := CreateNewSchedule(
			scheduleOwner,
			item,
		)
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
		if schedule.Data.HomescriptCode != item.HomescriptCode &&
			schedule.Data.Hour != item.Hour &&
			schedule.Data.Minute != item.Minute &&
			schedule.Data.Name != item.Name &&
			schedule.Owner != scheduleOwner {
			t.Errorf("Created schedule %d has invalid metadata", schedule.Id)
			return
		}
	}
	userSchedules, err := GetUserSchedules("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, userSchedule := range userSchedules {
		found := false
		for _, tableItem := range table {
			if tableItem.Name == userSchedule.Data.Name &&
				tableItem.HomescriptCode == userSchedule.Data.HomescriptCode &&
				tableItem.Hour == userSchedule.Data.Hour &&
				tableItem.Minute == userSchedule.Data.Minute &&
				scheduleOwner == userSchedule.Owner {
				found = true
			}
		}
		if !found {
			t.Errorf("Schedule with name: %s not found in user schedules", userSchedule.Data.Name)
			return
		}
	}

	allSchedules, err := GetSchedules()
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, allScheduleIter := range allSchedules {
		found := false
		for _, tableItem := range table {
			if tableItem.Name == allScheduleIter.Data.Name &&
				tableItem.HomescriptCode == allScheduleIter.Data.HomescriptCode &&
				tableItem.Hour == allScheduleIter.Data.Hour &&
				tableItem.Minute == allScheduleIter.Data.Minute &&
				scheduleOwner == allScheduleIter.Owner {
				found = true
			}
		}
		if !found {
			t.Errorf("Schedule with name: %s not found in schedules", allScheduleIter.Data.Name)
			return
		}
	}
}

func TestGetExistentScheduleById(t *testing.T) {
	schedule := ScheduleData{
		Name:           "test1",
		Hour:           1,
		Minute:         1,
		TargetMode:     ScheduleTargetModeCode,
		HomescriptCode: "print('a')",
	}
	newId, err := CreateNewSchedule(
		scheduleOwner,
		schedule,
	)
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
		Before ScheduleData
		After  ScheduleData
	}{
		{
			Before: ScheduleData{
				Name:           "before",
				Hour:           1,
				Minute:         2,
				TargetMode:     ScheduleTargetModeCode,
				HomescriptCode: "print('before')",
			},
			After: ScheduleData{
				Name:               "after",
				Hour:               3,
				Minute:             4,
				TargetMode:         ScheduleTargetModeHMS,
				HomescriptCode:     "print('after')",
				HomescriptTargetId: "test",
			},
		},
		{
			Before: ScheduleData{
				Name:       "before2",
				Hour:       5,
				Minute:     6,
				TargetMode: ScheduleTargetModeDevices,
			},
			After: ScheduleData{
				Name:           "after2",
				Hour:           7,
				Minute:         8,
				TargetMode:     ScheduleTargetModeCode,
				HomescriptCode: "print('after2')",
			},
		},
	}
	for _, test := range table {
		newId, err := CreateNewSchedule(
			scheduleOwner,
			test.Before,
		)
		if err != nil {
			t.Error(err.Error())
			return
		}
		// Modify the schedule
		if err := ModifySchedule(newId, test.After); err != nil {
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
			schedule.Data.Name != test.After.Name ||
			schedule.Data.HomescriptCode != test.After.HomescriptCode ||
			schedule.Data.Hour != test.After.Hour ||
			schedule.Data.Minute != test.After.Minute ||
			schedule.Data.TargetMode != test.After.TargetMode ||
			schedule.Data.HomescriptTargetId != test.After.HomescriptTargetId ||
			schedule.Owner != scheduleOwner {
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
