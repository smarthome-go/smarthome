package database

import (
	"testing"
)

func TestCreateHasSwitchPermissionTable(t *testing.T) {
	if err := createHasDevicePermissionTable(); err != nil {
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

	// testDriver := DeviceDriver{
	// 	VendorId:       "golang",
	// 	ModelId:        "test-1",
	// 	Name:           "",
	// 	Version:        "0.0.1",
	// 	HomescriptCode: DefaultDriverHomescriptCode,
	// 	SingletonJSON:  nil,
	// }

	for _, test := range table {
		// TODO: reimplement this

		// if err := CreateDevice(
		// 	DEVICE_TYPE_OUTPUT,
		// 	test.Switch,
		// 	"",
		// 	"test_permissions",
		// 	testDriver.VendorId,
		// 	testDriver.ModelId,
		// ); err != nil {
		// 	t.Error(err.Error())
		// 	return
		// }
		removed, err := RemoveUserDevicePermission("permissions_switch", test.Switch)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if removed {
			t.Errorf("Switch permission %s was removed but was never added", test.Switch)
			return
		}
		if test.Add {
			added, err := AddUserDevicePermission("permissions_switch", test.Switch)
			if err != nil {
				t.Error(err.Error())
				return
			}
			if !added {
				t.Errorf("Permission for switch %s was not created", test.Switch)
				return
			}
		}
		hasPermission, err := UserHasDevicePermission("permissions_switch", test.Switch)
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
		added, err := AddUserDevicePermission("permissions_switch", test.Switch)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if added == test.Add {
			t.Errorf("Added response for permission for switch %s does not match expected value: want: %t got: %t", test.Switch, !test.Add, added)
			return
		}
		// Remove switch permission
		removed, err := RemoveUserDevicePermission("permissions_switch", test.Switch)
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
