package database

import "testing"

func createUserMockData() error {
	if err := initDB(true); err != nil {
		return err
	}
	// Create user
	if err := AddUser(FullUser{
		Username:          "delete_me",
		Firstname:         "forename",
		Surname:           "surname",
		PrimaryColorDark:  "#121212",
		PrimaryColorLight: "#121212",
		Password:          "test",
		AvatarPath:        "/invalid",
		SchedulerEnabled:  true,
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
		fromDb.PrimaryColorDark != "#121212" ||
		fromDb.PrimaryColorLight != "#121212" ||
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
	if avatarpath != "./resources/avatar/default.png" {
		t.Errorf("Unexpected avatar path: want: ./resources/avatar/default.png got: %s", avatarpath)
		return
	}
	if err := SetUserAvatarPath("delete_me", "invalid_path"); err != nil {
		t.Error(err.Error())
		return
	}
	avatarpath, err = GetAvatarPathByUsername("delete_me")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if avatarpath != "invalid_path" {
		t.Errorf("Unexpected avatar path: want: invalid_path got: %s", avatarpath)
		return
	}
	// Cleanup
	if err := DeleteUser("delete_me"); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestSetScheduleEnabled(t *testing.T) {
	if err := createUserMockData(); err != nil {
		t.Error(err.Error())
		return
	}
	before, exists, err := GetUserByUsername("delete_me")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Error("User `delete_me` does not exist after creation")
		return
	}
	if err := SetUserSchedulerEnabled("delete_me", !before.SchedulerEnabled); err != nil {
		t.Error(err.Error())
		return
	}
	after, exists, err := GetUserByUsername("delete_me")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Error("User `delete_me` does not exist after modification")
		return
	}
	if after.SchedulerEnabled == before.SchedulerEnabled {
		t.Errorf("ScheduleEnabled not toggled: want: %t got: %t", !before.SchedulerEnabled, after.SchedulerEnabled)
		return
	}
	// Cleanup
	if err := DeleteUser("delete_me"); err != nil {
		t.Error(err.Error())
		return
	}
}

func assertDarkTheme(username string, want bool, t *testing.T) {
	user, found, err := GetUserByUsername("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("user with username `admin` not found in database")
		return
	}
	if user.DarkTheme {
		t.Errorf("Dark theme does not match: want: %t got: %t", want, user.DarkTheme)
		return
	}
}

func TestSetUserDarkTheme(t *testing.T) {
	if err := SetUserDarkThemeEnabled("admin", false); err != nil {
		t.Error(err.Error())
		return
	}
	assertDarkTheme("admin", false, t)
	if err := SetUserDarkThemeEnabled("admin", false); err != nil {
		t.Error(err.Error())
		return
	}
	assertDarkTheme("admin", false, t)
}

func assertPrimaryColors(username string, wantDark string, wantLight string, t *testing.T) {
	user, found, err := GetUserByUsername("admin")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !found {
		t.Errorf("user with username `admin` not found in database")
		return
	}
	if user.PrimaryColorDark != wantDark {
		t.Errorf("Primary color for dark theme does not match: want: %s got: %s", wantDark, user.PrimaryColorDark)
		return
	}
	if user.PrimaryColorLight != wantLight {
		t.Errorf("Primary color for light theme does not match: want: %s got: %s", wantLight, user.PrimaryColorLight)
		return
	}
}

func TestSetUserPrimaryColors(t *testing.T) {
	if err := SetUserPrimaryColors("admin", "#001122", "#334455"); err != nil {
		t.Error(err.Error())
		return
	}
	assertPrimaryColors("admin", "#001122", "#334455", t)
	if err := SetUserPrimaryColors("admin", "#667788", "#112233"); err != nil {
		t.Error(err.Error())
		return
	}
	assertPrimaryColors("admin", "#667788", "#112233", t)
}
