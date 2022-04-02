package database

import (
	"strings"
	"testing"
)

func TestCreateAutomationTable(t *testing.T) {
	// Create automation table
	if err := createAutomationTable(); err != nil {
		t.Error(err.Error())
		return
	}
	// Query automation table for data
	if _, err := GetAutomations(); err != nil {
		t.Error(err.Error())
		return
	}
}

/*Tests:
- Creation of automations
- Error handline
- Foreign keys
- Listing automations
- Metadata integrity
*/
func TestCreateNewAutomation(t *testing.T) {
	table := []struct {
		Automation Automation
		Error      string
	}{
		{
			Automation: Automation{
				Name:           "test1",
				Description:    "test1",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingNormal,
			},
			Error: "",
		},
		{
			Automation: Automation{
				Name:           "test1",
				Description:    "test1",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingSunrise,
			},
			Error: "",
		},
		{
			Automation: Automation{
				Name:           "test1",
				Description:    "test1",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingSunset,
			},
			Error: "",
		},
		{
			Automation: Automation{
				Name:           "test2",
				Description:    "test2",
				CronExpression: "* * * * * *",
				HomescriptId:   "test_invalid", // Test for invalid homescript
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingNormal,
			},
			Error: "Error 1452: Cannot add or update a child row: a foreign key constraint fails",
		},
		{
			Automation: Automation{
				Name:           "test2",
				Description:    "test2",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin_invalid", // Test for invalid user
				Enabled:        false,
				TimingMode:     TimingNormal,
			},
			Error: "Error 1452: Cannot add or update a child row: a foreign key constraint fails",
		},
	}
	// Create and evaluate the automations
	for _, automation := range table {
		newId, err := CreateNewAutomation(automation.Automation)
		// Check for error validity
		if err != nil {
			if !strings.Contains(err.Error(), automation.Error) || automation.Error == "" {
				t.Errorf("Unexpected error at name: %s : want: `%s` got: `%s`", automation.Automation.Name, automation.Error, err.Error())
				return
			}
		} else if automation.Error != "" {
			t.Errorf("Expected error: want: `%s` got: ``", automation.Error)
			return
		}
		// Check for metadata validity
		automationsFromDb, err := GetAutomations()
		if err != nil {
			t.Error(err.Error())
			return
		}
		valid := false
		for _, item := range automationsFromDb {
			if item.Id == newId &&
				item.Name == automation.Automation.Name &&
				item.Description == automation.Automation.Description &&
				item.CronExpression == automation.Automation.CronExpression &&
				item.HomescriptId == automation.Automation.HomescriptId &&
				item.Owner == automation.Automation.Owner &&
				item.Enabled == automation.Automation.Enabled &&
				item.TimingMode == automation.Automation.TimingMode {
				valid = true
			}
		}
		// Only trow a comparison error if the query did not return an error
		if !valid && automation.Error == "" {
			t.Errorf("Metadata comparison failed: want: %v", automation.Automation)
			return
		}
	}
}

func TestGetAutomationById(t *testing.T) {
	table := []struct {
		Automation    Automation
		Error         string
		UseFakeSearch bool // Specifies if a wrong id should be queried
	}{
		{
			Automation: Automation{
				Name:           "test1",
				Description:    "test1",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingNormal,
			},
			Error:         "",
			UseFakeSearch: false,
		},
		{
			Automation: Automation{
				Name:           "test2",
				Description:    "test2",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingSunrise,
			},
			Error:         "",
			UseFakeSearch: false,
		},
		{
			Automation: Automation{
				Name:           "test3",
				Description:    "test3",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingSunset,
			},
			Error:         "",
			UseFakeSearch: true,
		},
	}
	// Create and evaluate the automations
	for _, automation := range table {
		newId, err := CreateNewAutomation(automation.Automation)
		// Check for error validity
		if err != nil {
			if err.Error() != automation.Error {
				t.Errorf("Unexpected error: want: `%s` got: `%s`", automation.Error, err.Error())
				return
			}
		} else if automation.Error != "" {
			t.Errorf("Expected error: want: `%s` got: ``", automation.Error)
			return
		}
		searchId := newId
		if automation.UseFakeSearch {
			searchId = 99999999
		}
		// Check for metadata validity
		res, found, err := GetAutomationById(searchId)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !found && !automation.UseFakeSearch {
			t.Errorf("Id %d could no be found in created dataset", newId)
			return
		}
		if res.Id == newId &&
			res.Name == automation.Automation.Name &&
			res.Description == automation.Automation.Description &&
			res.CronExpression == automation.Automation.CronExpression &&
			res.HomescriptId == automation.Automation.HomescriptId &&
			res.Owner == automation.Automation.Owner &&
			res.Enabled == automation.Automation.Enabled &&
			res.TimingMode == automation.Automation.TimingMode {
		} else if !automation.UseFakeSearch {
			// Only throw an error if the fake search is not used
			t.Errorf("Metadata comparison failed: want: %v", automation.Automation)
			return
		}
	}
}

func TestGetUserAutomations(t *testing.T) {
	table := []struct {
		Automation Automation
		Error      string
	}{
		{
			Automation: Automation{
				Name:           "test1",
				Description:    "test1",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "testing",
				Enabled:        false,
				TimingMode:     TimingNormal,
			},
			Error: "",
		},
		{
			Automation: Automation{
				Name:           "test1",
				Description:    "test1",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "testing",
				Enabled:        false,
				TimingMode:     TimingSunrise,
			},
			Error: "",
		},
		{
			Automation: Automation{
				Name:           "test1",
				Description:    "test1",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingSunset,
			},
			Error: "",
		},
		{
			Automation: Automation{
				Name:           "test2",
				Description:    "test2",
				CronExpression: "* * * * * *",
				HomescriptId:   "test_invalid", // Test for invalid homescript
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingNormal,
			},
			Error: "Error 1452: Cannot add or update a child row: a foreign key constraint fails",
		},
		{
			Automation: Automation{
				Name:           "test2",
				Description:    "test2",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin_invalid", // Test for invalid user
				Enabled:        false,
				TimingMode:     TimingNormal,
			},
			Error: "Error 1452: Cannot add or update a child row: a foreign key constraint fails",
		},
	}
	// Create and evaluate the automations
	for _, automation := range table {
		_, err := CreateNewAutomation(automation.Automation)
		// Check for error validity
		if err != nil {
			if !strings.Contains(err.Error(), automation.Error) || automation.Error == "" {
				t.Errorf("Unexpected error: want: `%s` got: `%s`", automation.Error, err.Error())
				return
			}
		} else if automation.Error != "" {
			t.Errorf("Expected error: want: `%s` got: ``", automation.Error)
			return
		}
		// Check for metadata validity
		automationsFromDb, err := GetUserAutomations("testing")
		if err != nil {
			t.Error(err.Error())
			return
		}
		for _, item := range automationsFromDb {
			// Check if there are non-testing scripts in the result
			if item.Owner != "testing" {
				// Only trow a comparison error if the query did not return an error
				if automation.Error == "" {
					t.Errorf("Non-testing automation in personal query %v", automation.Automation)
					return
				}
			}
		}
	}
}

func TestModifyAutomation(t *testing.T) {
	table := []struct {
		Automation Automation
		Error      string
	}{
		{
			Automation: Automation{
				Name:           "1",
				Description:    "1",
				CronExpression: "* * * * * 1",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingNormal,
			},
			Error: "",
		},
		{
			Automation: Automation{
				Name:           "2",
				Description:    "2",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        true,
				TimingMode:     TimingNormal,
			},
			Error: "",
		},
		{
			Automation: Automation{
				Name:           "3",
				Description:    "3",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingSunrise,
			},
			Error: "",
		},
		{
			Automation: Automation{
				Name:           "4",
				Description:    "4",
				CronExpression: "* * * * * *",
				HomescriptId:   "test",
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingSunset,
			},
			Error: "",
		},
		{
			Automation: Automation{
				Name:           "5",
				Description:    "5",
				CronExpression: "* * * * * *",
				HomescriptId:   "test_invalid", // Test for invalid homescript
				Owner:          "admin",
				Enabled:        false,
				TimingMode:     TimingNormal,
			},
			Error: "Error 1452: Cannot add or update a child row: a foreign key constraint fails",
		},
	}
	// Create the initial automation
	newId, err := CreateNewAutomation(Automation{
		Id:             1,
		Name:           "before",
		Description:    "before",
		CronExpression: "before",
		HomescriptId:   "test",
		Owner:          "admin",
		Enabled:        false,
		TimingMode:     TimingNormal,
	})
	if err != nil {
		t.Error(err.Error())
		return
	}
	// Modify the automation to these values and evaluate the outcome
	for _, automation := range table {
		err := ModifyAutomation(newId, AutomationWithoutIdAndUsername{
			Name:           automation.Automation.Name,
			Description:    automation.Automation.Description,
			CronExpression: automation.Automation.CronExpression,
			HomescriptId:   automation.Automation.HomescriptId,
			Enabled:        automation.Automation.Enabled,
			TimingMode:     automation.Automation.TimingMode,
		})
		// Check for error validity
		if err != nil {
			if !strings.Contains(err.Error(), automation.Error) || automation.Error == "" {
				t.Errorf("Unexpected error: want: `%s` got: `%s`", automation.Error, err.Error())
				return
			}
		} else if automation.Error != "" {
			t.Errorf("Automation Name: %s Expected error: want: `%s` got: ``", automation.Automation.Name, automation.Error)
			return
		}
		// Check for metadata validity
		item, found, err := GetAutomationById(newId)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !found {
			t.Errorf("Automation %d could not be found in database", newId)
			return
		}
		if item.Name != automation.Automation.Name ||
			item.Description != automation.Automation.Description ||
			item.CronExpression != automation.Automation.CronExpression ||
			item.Enabled != automation.Automation.Enabled ||
			item.HomescriptId != automation.Automation.HomescriptId ||
			item.Owner != item.Owner ||
			item.TimingMode != automation.Automation.TimingMode {
			if automation.Error == "" {
				t.Errorf("Modification did not succeed: want: %v got: %v", automation.Automation, item)
				return
			}
		}
	}
}
