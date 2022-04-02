package database

import "testing"

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
}

// TODO: continue with getSchedules
