package database

import (
	"testing"
)

func TestCreateHomescriptTable(t *testing.T) {
	if err := createHomescriptTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

// Tests creation and retrieving Homescript
func TestHomescript(t *testing.T) {
	table := []struct {
		Homescript        Homescript
		AfterModification Homescript
		Error             string
		ErrorModification string
	}{
		{
			Homescript: Homescript{
				Owner: "admin",
				Data: HomescriptData{
					Id:                  "test1",
					Name:                "test",
					Description:         "test",
					QuickActionsEnabled: false,
					SchedulerEnabled:    false,
					Code:                "print('a')",
					MDIcon:              "code",
				},
			},
			Error: "",
			AfterModification: Homescript{
				Owner: "admin",
				Data: HomescriptData{

					Id:                  "test1",
					Name:                "test2",
					Description:         "test2",
					QuickActionsEnabled: true,
					SchedulerEnabled:    false,
					Code:                "print('b')",
					MDIcon:              "code_off",
				},
			},
			ErrorModification: "",
		},
		{
			Homescript: Homescript{
				Owner: "admin",
				Data: HomescriptData{

					Id:                  "test2",
					Name:                "test",
					Description:         "test",
					QuickActionsEnabled: true,
					SchedulerEnabled:    false,
					Code:                "",
					MDIcon:              "check",
				},
			},
			Error: "",
			AfterModification: Homescript{
				Owner: "admin",
				Data: HomescriptData{

					Id:                  "test2",
					Name:                "test",
					Description:         "test",
					QuickActionsEnabled: true,
					SchedulerEnabled:    true,
					Code:                ";",
					MDIcon:              "cancel",
				},
			},
			ErrorModification: "",
		},
		{
			Homescript: Homescript{
				Owner: "invalid",
				Data: HomescriptData{

					Id:                  "test4",
					Name:                "test",
					Description:         "test",
					QuickActionsEnabled: false,
					SchedulerEnabled:    false,
				},
			},
			Error:             "Error 1452 (23000): Cannot add or update a child row: a foreign key constraint fails (`smarthome`.`homescript`, CONSTRAINT `HomescriptOwner` FOREIGN KEY (`Owner`) REFERENCES `user` (`Username`))",
			AfterModification: Homescript{},
			ErrorModification: "", // Modification will not take place when an error is expected
		},
	}
	for _, item := range table {
		if err := CreateNewHomescript(item.Homescript); err != nil {
			if item.Error != err.Error() {
				t.Errorf("Unexpected error at script %s: want: %s got: %s", item.Homescript.Data.Id, item.Error, err.Error())
				return
			}
		} else if item.Error != "" {
			t.Errorf("Expected abundant error: want: %s got: ", item.Error)
			return
		}
		homescript, exists, err := GetUserHomescriptById(item.Homescript.Data.Id, item.Homescript.Owner)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !exists {
			if item.Error != "" {
				continue
			}
			t.Errorf("Expected existence of Homescript %s but it does not exist", item.Homescript.Data.Id)
			return
		}
		// Check metadata
		if homescript.Data.Code != item.Homescript.Data.Code ||
			homescript.Data.Description != item.Homescript.Data.Description ||
			homescript.Data.Id != item.Homescript.Data.Id ||
			homescript.Data.Name != item.Homescript.Data.Name ||
			homescript.Owner != item.Homescript.Owner ||
			homescript.Data.QuickActionsEnabled != item.Homescript.Data.QuickActionsEnabled ||
			homescript.Data.SchedulerEnabled != item.Homescript.Data.SchedulerEnabled {
			t.Errorf("Metadata of newly created homescript does not match: want: %v got: %v", item.Homescript, homescript)
			return
		}
		// Modify Homescript
		if item.Error == "" {
			if err := ModifyHomescriptById(
				item.Homescript.Data.Id,
				"admin",
				item.AfterModification.Data,
			); err != nil {
				if err.Error() != item.ErrorModification {
					t.Errorf("Unexpected error during modification of %s: want: %s got: %s", item.Homescript.Data.Id, item.ErrorModification, err.Error())
					return
				}
				continue
			} else if item.ErrorModification != "" {
				t.Errorf("Expected abundant error during modification of %s: want: %s got: %s", item.Homescript.Data.Id, item.ErrorModification, "")
				return
			}
			homescript, exists, err := GetUserHomescriptById(item.Homescript.Data.Id, item.Homescript.Owner)
			if err != nil {
				t.Error(err.Error())
				return
			}
			if !exists {
				t.Errorf("Homescript %s does not exists after modification", item.Homescript.Data.Id)
				return
			}
			if homescript.Data.Id != item.AfterModification.Data.Id ||
				homescript.Data.Name != item.AfterModification.Data.Name ||
				homescript.Data.Description != item.AfterModification.Data.Description ||
				homescript.Owner != item.AfterModification.Owner ||
				homescript.Data.QuickActionsEnabled != item.AfterModification.Data.QuickActionsEnabled ||
				homescript.Data.SchedulerEnabled != item.AfterModification.Data.SchedulerEnabled ||
				homescript.Data.Code != item.AfterModification.Data.Code {
				t.Errorf("Metadata of %s did not change completely after modification: want: %v got: %v", item.Homescript.Data.Id, item.AfterModification, homescript)
				return
			}
		}
		// Delete Homescript
		if err := DeleteHomescriptById(item.Homescript.Data.Id, "admin"); err != nil {
			t.Error(err.Error())
			return
		}
		exists, err = DoesHomescriptExist(item.Homescript.Data.Id, "admin")
		if err != nil {
			t.Error(err.Error())
			return
		}
		if exists {
			t.Errorf("Homescript %s still exists after deletion", homescript.Data.Id)
			return
		}
	}
}

func TestListHomescript(t *testing.T) {
	// Create test user
	if err := AddUser(FullUser{
		Username: "hms_testing",
	}); err != nil {
		t.Error(err.Error())
		return
	}
	// Add one script for the admin and one for the testuser
	scripts := []Homescript{
		{
			Owner: "hms_testing",
			Data: HomescriptData{
				Id:                  "hms_testing",
				Name:                "test",
				Description:         "test",
				QuickActionsEnabled: false,
				SchedulerEnabled:    false,
				Code:                "",
				MDIcon:              "code",
			},
		},
		{
			Owner: "admin",
			Data: HomescriptData{
				Id:                  "admin",
				Name:                "test",
				Description:         "test",
				QuickActionsEnabled: false,
				SchedulerEnabled:    false,
				Code:                ";",
				MDIcon:              "code_off",
			},
		},
		{
			Owner: "hms_testing",
			Data: HomescriptData{
				Id:                  "hms_testing2",
				Name:                "test",
				Description:         "test",
				QuickActionsEnabled: false,
				SchedulerEnabled:    false,
				Code:                ";;",
				MDIcon:              "check",
			},
		},
		{
			Owner: "admin",
			Data: HomescriptData{
				Id:                  "admin2",
				Name:                "test",
				Description:         "test",
				QuickActionsEnabled: false,
				SchedulerEnabled:    false,
				Code:                ";;;",
				MDIcon:              "cancel",
			},
		},
	}
	for _, script := range scripts {
		if err := CreateNewHomescript(script); err != nil {
			t.Error(err.Error())
			return
		}
		personalScripts, err := ListHomescriptOfUser(script.Owner)
		if err != nil {
			t.Error(err.Error())
			return
		}
		for _, item := range personalScripts {
			if item.Owner != script.Owner {
				t.Errorf("Unexpected homescript-item %v for user %s: this is a security vulnerability", item, script.Owner)
				return
			}
		}
	}
}

func TestGetUserHomescriptById(t *testing.T) {
	if err := AddUser(FullUser{
		Username: "hms_testing2",
	}); err != nil {
		t.Error(err.Error())
		return
	}
	table := []Homescript{
		{
			Owner: "admin",
			Data: HomescriptData{
				Id: "admin_new",
			},
		},
		{
			Owner: "hms_testing2",
			Data: HomescriptData{
				Id: "hms_testing_temp",
			},
		},
	}
	for _, item := range table {
		if err := CreateNewHomescript(item); err != nil {
			t.Error(err.Error())
			return
		}
	}
	_, exists, err := GetUserHomescriptById("admin_new", "hms_testing2")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if exists {
		t.Errorf("Homescript `admin_new` should not be accessible by user `hms_testing2`")
		return
	}
	_, exists, err = GetUserHomescriptById("hms_testing_temp", "hms_testing2")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Errorf("Homescript `hms_testing_temp` should be accessible by user `hms_testing2`")
		return
	}
}
