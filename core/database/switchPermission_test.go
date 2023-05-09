package database

import "testing"

func TestCreateHasSwitchPermissionTable(t *testing.T) {
	if err := createHasSwitchPermissionTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

// tested in a better context in `switch_test.go`
func TestAddUserSwitchPermission(t *testing.T) {
	if err := AddUser(FullUser{
		Username: "permissions_switch",
	}); err != nil {
		t.Error(err.Error())
		return
	}
	if err := CreateRoom(
		RoomData{Id: "test_permissions"}); err != nil {
		t.Error(err.Error())
		return
	}
	table := []struct {
		Switch string
		Add    bool
	}{
		{
			Switch: "test_permission_1",
			Add:    true,
		},
		{
			Switch: "test_permission_2",
			Add:    false,
		},
	}
	for _, test := range table {
		if err := CreateSwitch(test.Switch, "", "test_permissions", 0, nil); err != nil {
			t.Error(err.Error())
			return
		}
		removed, err := RemoveUserSwitchPermission("permissions_switch", test.Switch)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if removed {
			t.Errorf("Switch permission %s was removed but was never added", test.Switch)
			return
		}
		if test.Add {
			added, err := AddUserSwitchPermission("permissions_switch", test.Switch)
			if err != nil {
				t.Error(err.Error())
				return
			}
			if !added {
				t.Errorf("Permission for switch %s was not created", test.Switch)
				return
			}
		}
		hasPermission, err := UserHasSwitchPermission("permissions_switch", test.Switch)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if hasPermission != test.Add {
			t.Errorf("Switch %s does not match added value: want: %t got: %t", test.Switch, test.Add, hasPermission)
			return
		}
	}
	// Add permissions again, this time for all switches
	for _, test := range table {
		added, err := AddUserSwitchPermission("permissions_switch", test.Switch)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if added == test.Add {
			t.Errorf("Added response for permission for switch %s does not match expected value: want: %t got: %t", test.Switch, !test.Add, added)
			return
		}
		// Remove switch permission
		removed, err := RemoveUserSwitchPermission("permissions_switch", test.Switch)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if !removed {
			t.Errorf("Switch permission %s was not be removed", test.Switch)
			return
		}
	}
}
