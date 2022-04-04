package database

import "testing"

func createUserMockData() error {
	// Create user
	if err := AddUser(FullUser{
		Username:         "delete_me",
		Firstname:        "forename",
		Surname:          "surname",
		PrimaryColor:     "#121212",
		Password:         "test",
		AvatarPath:       "/invalid",
		SchedulerEnabled: true,
	}); err != nil {
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

func TestGetUserByUsername(t *testing.T) {
	if err := createUserMockData(); err != nil {
		t.Error(err.Error())
		return
	}
	fromDb, exists, err := GetUserByUsername("delete_me")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Errorf("User `delete_me` does not exist after creation")
		return
	}
	if fromDb.Username != "delete_me" ||
		fromDb.Firstname != "forename" ||
		fromDb.Surname != "surname" ||
		fromDb.PrimaryColor != "#121212" ||
		!fromDb.SchedulerEnabled {
		t.Errorf("User `delete_me` has invalid metadata: got: %v", fromDb)
		return
	}
	if err := DeleteUser("delete_me"); err != nil {
		t.Error(err.Error())
		return
	}
	_, exists, err = GetUserByUsername("delete_me")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if exists {
		t.Errorf("User `delete_me` does still exist after creation")
		return
	}
	// Cleanup
	if err := DeleteUser("delete_me"); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestUserPasswordHash(t *testing.T) {
	if err := createUserMockData(); err != nil {
		t.Error(err.Error())
		return
	}
	hash, err := GetUserPasswordHash("delete_me")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if hash == `` {
		t.Errorf("unexpected password hash length: got %s", hash)
		return
	}
	// Cleanup
	if err := DeleteUser("delete_me"); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestUserAvatarPath(t *testing.T) {
	if err := createUserMockData(); err != nil {
		t.Error(err.Error())
		return
	}
	avatarpath, err := GetAvatarPathByUsername("delete_me")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if avatarpath != "./web/assets/avatar/default.png" {
		t.Errorf("Unexpected avatar path: want: ./web/assets/avatar/default.png got: %s", avatarpath)
		return
	}
	// Cleanup
	if err := DeleteUser("delete_me"); err != nil {
		t.Error(err.Error())
		return
	}
}
