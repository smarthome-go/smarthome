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

/*
Tests:
- Creation of automations
- Error handline
- Foreign keys
- Listing automations
- Metadata integrity
*/
func TestCreateNewAutomation(t *testing.T) {
	cronExpression1 := "* * * * * *"

	table := []struct {
		Automation Automation
		Error      string
	}{
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "test1",
					Description:           "test1",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerCron,
				},
			},
			Error: "",
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:         "test1",
					Description:  "test1",
					HomescriptId: "test",
					Enabled:      false,
					Trigger:      TriggerSunrise,
				},
			},
			Error: "",
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:         "test1",
					Description:  "test1",
					HomescriptId: "test",
					Enabled:      false,
					Trigger:      TriggerSunset,
				},
			},
			Error: "",
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "test2",
					Description:           "test2",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test_invalid", // Test for invalid homescript
					Enabled:               false,
					Trigger:               TriggerCron,
				},
			},
			Error: "Error 1452 (23000): Cannot add or update a child row: a foreign key constraint fails",
		},
		{
			Automation: Automation{
				Owner: "admin_invalid", // Test for invalid user
				Data: AutomationData{

					Name:                  "test2",
					Description:           "test2",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerCron,
				},
			},
			Error: "Error 1452 (23000): Cannot add or update a child row: a foreign key constraint fails",
		},
	}
	// Create and evaluate the automations
	for _, automation := range table {
		newId, err := CreateNewAutomation(automation.Automation)
		// Check for error validity
		if err != nil {
			if !strings.Contains(err.Error(), automation.Error) || automation.Error == "" {
				t.Errorf("Unexpected error at name: %s : want: `%s` got: `%s`", automation.Automation.Data.Name, automation.Error, err.Error())
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
				item.Data.Name == automation.Automation.Data.Name &&
				item.Data.Description == automation.Automation.Data.Description &&
				(item.Data.TriggerCronExpression == automation.Automation.Data.TriggerCronExpression ||
					*item.Data.TriggerCronExpression == *automation.Automation.Data.TriggerCronExpression) &&
				item.Data.HomescriptId == automation.Automation.Data.HomescriptId &&
				item.Owner == automation.Automation.Owner &&
				item.Data.Enabled == automation.Automation.Data.Enabled &&
				item.Data.Trigger == automation.Automation.Data.Trigger {
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
	cronExpression1 := "* * * * * *"

	table := []struct {
		Automation    Automation
		Error         string
		UseFakeSearch bool // Specifies if a wrong id should be queried
	}{
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "test1",
					Description:           "test1",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerCron,
				},
			},
			Error:         "",
			UseFakeSearch: false,
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "test2",
					Description:           "test2",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerCron,
				},
			},
			Error:         "",
			UseFakeSearch: false,
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "test3",
					Description:           "test3",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerCron,
				},
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
			res.Data.Name == automation.Automation.Data.Name &&
			res.Data.Description == automation.Automation.Data.Description &&
			(res.Data.TriggerCronExpression == automation.Automation.Data.TriggerCronExpression || *res.Data.TriggerCronExpression == *automation.Automation.Data.TriggerCronExpression) &&
			res.Data.HomescriptId == automation.Automation.Data.HomescriptId &&
			res.Owner == automation.Automation.Owner &&
			res.Data.Enabled == automation.Automation.Data.Enabled &&
			res.Data.Trigger == automation.Automation.Data.Trigger {
		} else if !automation.UseFakeSearch {
			// Only throw an error if the fake search is not used
			t.Errorf("Metadata comparison failed: want: %v", automation.Automation)
			return
		}
	}
}

func TestGetUserAutomations(t *testing.T) {
	cronExpression1 := "* * * * * *"

	table := []struct {
		Automation Automation
		Error      string
	}{
		{
			Automation: Automation{
				Owner: "testing",
				Data: AutomationData{

					Name:                  "test1",
					Description:           "test1",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerCron,
				},
			},
			Error: "",
		},
		{
			Automation: Automation{
				Owner: "testing",
				Data: AutomationData{

					Name:                  "test1",
					Description:           "test1",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerSunrise,
				},
			},
			Error: "",
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "test1",
					Description:           "test1",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerSunset,
				},
			},
			Error: "",
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "test2",
					Description:           "test2",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test_invalid", // Test for invalid homescript
					Enabled:               false,
					Trigger:               TriggerCron,
				},
			},
			Error: "Error 1452 (23000): Cannot add or update a child row: a foreign key constraint fails",
		},
		{
			Automation: Automation{
				Owner: "admin_invalid", // Test for invalid user
				Data: AutomationData{

					Name:                  "test2",
					Description:           "test2",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerCron,
				},
			},
			Error: "Error 1452 (23000): Cannot add or update a child row: a foreign key constraint fails",
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

func TestModifyDeleteAutomation(t *testing.T) {
	cronExpression1 := "* * * * * 1"
	cronExpression2 := "* * * * * *"

	table := []struct {
		Automation Automation
		Error      string
	}{
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "1",
					Description:           "1",
					TriggerCronExpression: &cronExpression1,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerCron,
				},
			},
			Error: "",
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "2",
					Description:           "2",
					TriggerCronExpression: &cronExpression2,
					HomescriptId:          "test",
					Enabled:               true,
					Trigger:               TriggerCron,
				},
			},
			Error: "",
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "3",
					Description:           "3",
					TriggerCronExpression: &cronExpression2,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerSunrise,
				},
			},
			Error: "",
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "4",
					Description:           "4",
					TriggerCronExpression: &cronExpression2,
					HomescriptId:          "test",
					Enabled:               false,
					Trigger:               TriggerSunset,
				},
			},
			Error: "",
		},
		{
			Automation: Automation{
				Owner: "admin",
				Data: AutomationData{

					Name:                  "5",
					Description:           "5",
					TriggerCronExpression: &cronExpression2,
					HomescriptId:          "test_invalid", // Test for invalid homescript
					Enabled:               false,
					Trigger:               TriggerCron,
				},
			},
			Error: "Error 1452 (23000): Cannot add or update a child row: a foreign key constraint fails",
		},
	}
	// Create the initial automation
	newId, err := CreateNewAutomation(Automation{
		Id:    1,
		Owner: "admin",
		Data: AutomationData{
			Name:                  "before",
			Description:           "before",
			TriggerCronExpression: &cronExpression1,
			HomescriptId:          "test",
			Enabled:               false,
			Trigger:               TriggerCron,
		},
	})
	if err != nil {
		t.Error(err.Error())
		return
	}
	// Modify the automation to these values and evaluate the outcome
	for _, automation := range table {
		err := ModifyAutomation(newId, AutomationData{
			Name:                  automation.Automation.Data.Name,
			Description:           automation.Automation.Data.Description,
			TriggerCronExpression: automation.Automation.Data.TriggerCronExpression,
			HomescriptId:          automation.Automation.Data.HomescriptId,
			Enabled:               automation.Automation.Data.Enabled,
			Trigger:               automation.Automation.Data.Trigger,
		})
		// Check for error validity
		if err != nil {
			if !strings.Contains(err.Error(), automation.Error) || automation.Error == "" {
				t.Errorf("Unexpected error: want: `%s` got: `%s`", automation.Error, err.Error())
				return
			}
		} else if automation.Error != "" {
			t.Errorf("Automation Name: %s Expected error: want: `%s` got: ``", automation.Automation.Data.Name, automation.Error)
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
		if item.Data.Name != automation.Automation.Data.Name ||
			item.Data.Description != automation.Automation.Data.Description ||
			(item.Data.TriggerCronExpression != automation.Automation.Data.TriggerCronExpression && *item.Data.TriggerCronExpression != *automation.Automation.Data.TriggerCronExpression) ||
			item.Data.Enabled != automation.Automation.Data.Enabled ||
			item.Data.HomescriptId != automation.Automation.Data.HomescriptId ||
			item.Owner != automation.Automation.Owner ||
			item.Data.Trigger != automation.Automation.Data.Trigger {
			if automation.Error == "" {
				t.Errorf("Modification did not succeed: want: %v got: %v", automation.Automation, item)
				return
			}
		}
	}
	automations, err := GetAutomations()
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, automation := range automations {
		// Delete automation after it has been successfully modified
		if err := DeleteAutomationById(automation.Id); err != nil {
			t.Error(err.Error())
			return
		}
		// Check if the automation still exists after planned deletion
		_, exists, err := GetAutomationById(automation.Id)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if exists {
			t.Errorf("Automation %d still exists after deletion", automation.Id)
			return
		}
	}
}
