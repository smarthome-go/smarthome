package database

import "testing"

func createUserMockData() error {
	// Create user
	if err := AddUser(FullUser{Username: "delete_me"}); err != nil {
		return err
	}
	// Give him permissions
	for _, permission := range Permissions {
		if _, err := AddUserPermission("delete_me", permission.Permission); err != nil {
			return err
		}
	}
	// Create a switch for the room
	if err := CreateRoom("delete_me", "", ""); err != nil {
		return err
	}
	// Create a switch
	if err := CreateSwitch("delete_me", "", "delete_me", 0); err != nil {
		return err
	}
	// Give the user switch permission
	if _, err := AddUserSwitchPermission("delete_me", "delete_me"); err != nil {
		return err
	}
	// Create a homescript
	if err := CreateNewHomescript(Homescript{Id: "delete_me", Owner: "delete_me"}); err != nil {
		return err
	}
	// Create a automation
	if _, err := CreateNewAutomation(Automation{Owner: "delete_me", TimingMode: TimingNormal, HomescriptId: "delete_me"}); err != nil {
		return err
	}
	// Create a schedule
	if _, err := CreateNewSchedule(Schedule{Owner: "delete_me"}); err != nil {
		return err
	}
	// Create notification
	if err := AddNotification("delete_me", "", "", 1); err != nil {
		return err
	}
	return nil
}

// Checks if all 'dependencies' have been deleted
func TestDeleteuser(t *testing.T) {
	// Create mock user with some
	// - permissions
	// - notifications
	// - Homescripts
	// - Schedules
	// - Automations
	// - Switch permissions
	if err := createUserMockData(); err != nil {
		t.Error(err.Error())
		return
	}
	if err := DeleteUser("delete_me"); err != nil {
		t.Error(err.Error())
		return
	}
}
