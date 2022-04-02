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
				Id:                  "test1",
				Owner:               "admin",
				Name:                "test",
				Description:         "test",
				QuickActionsEnabled: false,
				SchedulerEnabled:    false,
				Code:                "print('a')",
			},
			Error: "",
			AfterModification: Homescript{
				Id:                  "test1",
				Owner:               "admin",
				Name:                "test2",
				Description:         "test2",
				QuickActionsEnabled: true,
				SchedulerEnabled:    false,
				Code:                "print('b')",
			},
			ErrorModification: "",
		},
		{
			Homescript: Homescript{
				Id:                  "test2",
				Owner:               "admin",
				Name:                "test",
				Description:         "test",
				QuickActionsEnabled: true,
				SchedulerEnabled:    false,
			},
			Error: "",
			AfterModification: Homescript{
				Id:                  "test2",
				Owner:               "admin",
				Name:                "test",
				Description:         "test",
				QuickActionsEnabled: true,
				SchedulerEnabled:    true,
			},
			ErrorModification: "",
		},
		{
			Homescript: Homescript{
				Id:                  "test4",
				Owner:               "invalid",
				Name:                "test",
				Description:         "test",
				QuickActionsEnabled: false,
				SchedulerEnabled:    false,
			},
			Error:             "Error 1452: Cannot add or update a child row: a foreign key constraint fails (`smarthome`.`homescript`, CONSTRAINT `HomescriptOwner` FOREIGN KEY (`Owner`) REFERENCES `user` (`Username`))",
			AfterModification: Homescript{},
			ErrorModification: "", // Modification will not take place when an error is expected
		},
	}
	for _, item := range table {
		if err := CreateNewHomescript(item.Homescript); err != nil {
			if item.Error != err.Error() {
				t.Errorf("Unexpected error at script %s: want: %s got: %s", item.Homescript.Id, item.Error, err.Error())
				return
			}
		} else if item.Error != "" {
			t.Errorf("Expected abundant error: want: %s got: ", item.Error)
			return
		}
		homescript, exists, err := GetUserHomescriptById(item.Homescript.Id, item.Homescript.Owner)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !exists {
			if item.Error != "" {
				continue
			}
			t.Errorf("Expected existence of Homescript %s but it does not exist", item.Homescript.Id)
			return
		}
		// Check metadata
		if homescript.Code != item.Homescript.Code ||
			homescript.Description != item.Homescript.Description ||
			homescript.Id != item.Homescript.Id ||
			homescript.Name != item.Homescript.Name ||
			homescript.Owner != item.Homescript.Owner ||
			homescript.QuickActionsEnabled != item.Homescript.QuickActionsEnabled ||
			homescript.SchedulerEnabled != item.Homescript.SchedulerEnabled {
			t.Errorf("Metadata of newly created homescript does not match: want: %v got: %v", item.Homescript, homescript)
			return
		}
		// Modify Homescript
		if item.Error == "" {
			if err := ModifyHomescriptById(
				item.Homescript.Id,
				HomescriptFrontend{
					Name:                item.AfterModification.Name,
					Description:         item.AfterModification.Description,
					QuickActionsEnabled: item.AfterModification.QuickActionsEnabled,
					SchedulerEnabled:    item.AfterModification.SchedulerEnabled,
					Code:                item.AfterModification.Code,
				},
			); err != nil {
				if err.Error() != item.ErrorModification {
					t.Errorf("Unexpected error during modification of %s: want: %s got: %s", item.Homescript.Id, item.ErrorModification, err.Error())
					return
				}
				continue
			} else if item.ErrorModification != "" {
				t.Errorf("Expected abundant error during modification of %s: want: %s got: %s", item.Homescript.Id, item.ErrorModification, "")
				return
			}
			homescript, exists, err := GetUserHomescriptById(item.Homescript.Id, item.Homescript.Owner)
			if err != nil {
				t.Error(err.Error())
				return
			}
			if !exists {
				t.Errorf("Homescript %s does not exists after modification", item.Homescript.Id)
				return
			}
			if homescript.Id != item.AfterModification.Id ||
				homescript.Name != item.AfterModification.Name ||
				homescript.Description != item.AfterModification.Description ||
				homescript.Owner != item.AfterModification.Owner ||
				homescript.QuickActionsEnabled != item.AfterModification.QuickActionsEnabled ||
				homescript.SchedulerEnabled != item.AfterModification.SchedulerEnabled ||
				homescript.Code != item.AfterModification.Code {
				t.Errorf("Metadata of %s did not change completely after modification: want: %v got: %v", item.Homescript.Id, item.AfterModification, homescript)
				return
			}
		}
		// Delete Homescript
		if err := DeleteHomescriptById(item.Homescript.Id); err != nil {
			t.Error(err.Error())
			return
		}
		exists, err = DoesHomescriptExist(item.Homescript.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if exists {
			t.Errorf("Homescript %s still exists after deletion", homescript.Id)
			return
		}
	}
}

func TestListHomeScript(t *testing.T) {
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
			Id:                  "hms_testing",
			Owner:               "hms_testing",
			Name:                "test",
			Description:         "test",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "",
		},
		{
			Id:                  "admin",
			Owner:               "admin",
			Name:                "test",
			Description:         "test",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "",
		},
		{
			Id:                  "hms_testing2",
			Owner:               "hms_testing",
			Name:                "test",
			Description:         "test",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "",
		},
		{
			Id:                  "admin2",
			Owner:               "admin",
			Name:                "test",
			Description:         "test",
			QuickActionsEnabled: false,
			SchedulerEnabled:    false,
			Code:                "",
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
			Id:    "admin",
			Owner: "admin",
		},
		{
			Id:    "hms_testing2",
			Owner: "hms_testing2",
		},
	}
	for _, item := range table {
		if err := CreateNewHomescript(item); err != nil {
			t.Error(err.Error())
			return
		}
	}
	_, exists, err := GetUserHomescriptById("admin", "hms_testing2")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if exists {
		t.Errorf("Homescript `admin` should not be accessible by user `hms_testing2`")
		return
	}
	_, exists, err = GetUserHomescriptById("hms_testing2", "hms_testing2")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Errorf("Homescript `hms_testing2` should be accessible by user `hms_testing2`")
		return
	}
}
