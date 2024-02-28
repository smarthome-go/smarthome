package database

import (
	"fmt"
	"strings"
	"testing"
)

func createTestRoom() error {
	return CreateRoom(
		RoomData{
			ID:          "test",
			Name:        "test_room",
			Description: "test_description",
		})
}

func createTestUser() error {
	return AddUser(FullUser{Username: "switches_test"})
}

func TestCreateSwitchTable(t *testing.T) {
	if err := createDeviceTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestSwitches(t *testing.T) {
	if err := createTestRoom(); err != nil {
		t.Error(err.Error())
		return
	}
	table := []struct {
		Switch ShallowDevice
		Error  string
	}{
		{
			Switch: ShallowDevice{
				DeviceType:    "",
				ID:            "test_1",
				Name:          "test_1",
				RoomID:        "test",
				VendorID:      "",
				ModelID:       "",
				SingletonJSON: "",
			},
			Error: "",
		},
		{
			Switch: ShallowDevice{
				DeviceType:    "",
				ID:            "test_2",
				Name:          "test_2",
				RoomID:        "test",
				VendorID:      "",
				ModelID:       "",
				SingletonJSON: "",
			},
			Error: "",
		},
		{
			Switch: ShallowDevice{
				DeviceType:    "",
				ID:            "test_3",
				Name:          "test_3",
				RoomID:        "invalid",
				VendorID:      "",
				ModelID:       "",
				SingletonJSON: "",
			},
			Error: "Error 1452 (23000): Cannot add or update a child row: a foreign key constraint fails",
		},
	}
	for _, test := range table {
		t.Run(fmt.Sprintf("create switch/%s", test.Switch.ID), func(t *testing.T) {
			if err := CreateDevice(ShallowDevice{
				DeviceType:    DEVICE_TYPE_OUTPUT,
				ID:            test.Switch.ID,
				Name:          test.Switch.Name,
				RoomID:        test.Switch.RoomID,
				VendorID:      test.Switch.VendorID,
				ModelID:       test.Switch.ModelID,
				SingletonJSON: test.Switch.SingletonJSON,
			}); err != nil {
				if !strings.Contains(err.Error(), test.Error) || test.Error == "" {
					t.Errorf("Unexpected error: want: %s got: %s ", test.Error, err.Error())
					return
				}
			} else if test.Error != "" {
				t.Errorf("Abundant error did not occur. want: %s got: %s", test.Error, "")
				return
			}
		})
		t.Run(fmt.Sprintf("get switch/%s", test.Switch.ID), func(t *testing.T) {
			switches, err := ListAllDevices()
			if err != nil {
				t.Errorf("Could not list switches: %s", err.Error())
				return
			}
			valid := false
			for _, s := range switches {
				if s.ID == test.Switch.ID &&
					s.Name == test.Switch.Name &&
					s.RoomID == test.Switch.RoomID &&
					s.SingletonJSON == test.Switch.SingletonJSON {
					valid = true
				}
			}
			if !valid && test.Error == "" {
				t.Errorf("Switch %s not found or has invalid metadata. want: %v", test.Switch.ID, test.Switch)
				return
			}
		})
		t.Run(fmt.Sprintf("delete switch/%s", test.Switch.ID), func(t *testing.T) {
			if err := DeleteDevice(test.Switch.ID); err != nil {
				t.Error(err.Error())
				return
			}
			switches, err := ListAllDevices()
			if err != nil {
				t.Errorf("Could not list switches: %s", err.Error())
				return
			}
			valid := false
			for _, s := range switches {
				if s.ID == test.Switch.ID {
					valid = true
				}
			}
			if valid {
				t.Errorf("Switch %s was found after deletion", test.Switch.ID)
				return
			}
		})
	}
}

func TestUserSwitches(t *testing.T) {
	if err := createTestRoom(); err != nil {
		t.Error(err.Error())
		return
	}
	if err := createTestUser(); err != nil {
		t.Error(err.Error())
		return
	}
	switches := []ShallowDevice{
		{
			DeviceType:    "",
			ID:            "1",
			Name:          "1",
			RoomID:        "test",
			VendorID:      "",
			ModelID:       "",
			SingletonJSON: "",
		},
		{
			DeviceType:    "",
			ID:            "2",
			Name:          "2",
			RoomID:        "test",
			VendorID:      "",
			ModelID:       "",
			SingletonJSON: "",
		},
		{
			DeviceType:    "",
			ID:            "3",
			Name:          "3",
			RoomID:        "test",
			VendorID:      "",
			ModelID:       "",
			SingletonJSON: "",
		},
		{
			DeviceType:    "",
			ID:            "4",
			Name:          "4",
			RoomID:        "test",
			VendorID:      "",
			ModelID:       "",
			SingletonJSON: "",
		},
	}
	hasSwitchPermissionTable := map[string]struct {
		User  string
		Error string
	}{
		"1": {
			User:  "switches_test",
			Error: "",
		},
		"2": {
			User:  "switches_test",
			Error: "",
		},
		"3": {
			User:  "admin",
			Error: "",
		},
		"4": {
			User:  "invalid",
			Error: "Error 1452 (23000): Cannot add or update a child row: a foreign key constraint fails",
		},
		"invalid": {
			User:  "admin",
			Error: "Error 1452 (23000): Cannot add or update a child row: a foreign key constraint fails",
		},
	}

	t.Run("create switches", func(t *testing.T) {
		for _, switchItem := range switches {
			t.Run(fmt.Sprintf("create switches/%s", switchItem.ID), func(t *testing.T) {
				if err := CreateDevice(
					ShallowDevice{
						DeviceType:    DEVICE_TYPE_OUTPUT,
						ID:            switchItem.ID,
						Name:          switchItem.Name,
						RoomID:        switchItem.RoomID,
						VendorID:      switchItem.VendorID,
						ModelID:       switchItem.ModelID,
						SingletonJSON: switchItem.SingletonJSON,
					},
				); err != nil {
					t.Error(err.Error())
					return
				}
			})
		}

		t.Run("add switch permissions", func(t *testing.T) {
			for switchId, test := range hasSwitchPermissionTable {
				t.Run(fmt.Sprintf("add switch permissions/%s", switchId), func(t *testing.T) {
					if _, err := AddUserDevicePermission(test.User, switchId); err != nil {
						if !strings.Contains(err.Error(), test.Error) || test.Error == "" {
							t.Errorf("Unexpected error for %s:%v: want: %s got: %s", switchId, test, test.Error, err.Error())
							return
						}
					} else if test.Error != "" {
						t.Errorf("Expected abundant error: %s which did not occur", test.Error)
						return
					}
				})
			}
		})

		t.Run("query user switches", func(t *testing.T) {
			for switchId, test := range hasSwitchPermissionTable {
				t.Run(fmt.Sprintf("query user switches/%s", switchId), func(t *testing.T) {
					hasPermission, err := UserHasDevicePermission(test.User, switchId)
					if err != nil {
						t.Error(err.Error())
						return
					}
					if !hasPermission && test.Error == "" {
						t.Errorf("User %s does not have switch permission %s", test.User, switchId)
						return
					}
					userSwitches, err := ListUserDevices(test.User)
					if err != nil {
						t.Error(err.Error())
						return
					}
					valid := false
					for _, s := range userSwitches {
						if s.ID == switchId {
							valid = true
						}
					}
					if !valid && test.Error == "" {
						t.Errorf("Switch %s not found in user switches", switchId)
						return
					}
					hasPermission, err = UserHasDevicePermission("__invalid__", switchId)
					if err != nil {
						t.Error(err.Error())
						return
					}
					if hasPermission {
						t.Errorf("User __invalid__ does has switch permission %s but should not have it", switchId)
						return
					}
				})
			}
		})

		// TODO: implement device tests

		// t.Run("test power states", func(t *testing.T) {
		// 	for _, switchTest := range switches {
		// 		switchPrev, found, err := GetDeviceById(switchTest.Id)
		// 		if err != nil {
		// 			t.Error(err.Error())
		// 		}
		// 		assert.True(t, found, "Switch not found")
		// 		if _, err := SetPowerState(switchTest.Id, !switchPrev.PowerOn); err != nil {
		// 			t.Error(err.Error())
		// 			return
		// 		}
		// 		switchItem, found, err := GetDeviceById(switchTest.Id)
		// 		if err != nil {
		// 			t.Error(err.Error())
		// 			return
		// 		}
		// 		assert.True(t, found, "Switch not found")
		// 		if switchItem.PowerOn == switchPrev.PowerOn {
		// 			t.Errorf("Power state did not change after toggle. want: %t got: %t", !switchPrev.PowerOn, switchItem.PowerOn)
		// 			return
		// 		}
		// 		powerStates, err := GetPowerStates()
		// 		if err != nil {
		// 			t.Error(err.Error())
		// 			return
		// 		}
		// 		valid := false
		// 		for _, s := range powerStates {
		// 			if s.Switch == switchTest.Id && s.PowerOn != switchPrev.PowerOn {
		// 				valid = true
		// 			}
		// 		}
		// 		if !valid {
		// 			t.Errorf("Switch %s with correct power state not matched in power states", switchTest.Id)
		// 			return
		// 		}
		// 	}
		// })
	})
}

// func TestDoesSwitchExist(t *testing.T) {
// 	if err := createTestRoom(); err != nil {
// 		t.Error(err.Error())
// 		return
// 	}
// 	if err := CreateDevice(Device{
// 		DeviceType:    DEVICE_TYPE_OUTPUT,
// 		Id:            "test1",
// 		Name:          "test1",
// 		RoomId:        "test",
// 		VendorId:      "",
// 		ModelId:       "",
// 		SingletonJSON: "",
// 	}); err != nil {
// 		t.Error(err.Error())
// 		return
// 	}
// 	_, switchExists, err := GetDeviceById("test1")
// 	if err != nil {
// 		t.Error(err.Error())
// 		return
// 	}
// 	assert.True(t, switchExists, "Switch 'test1' does not exist after creation")
// 	_, switchExists, err = GetDeviceById("invalid")
// 	if err != nil {
// 		t.Error(err.Error())
// 		return
// 	}
// 	assert.False(t, switchExists, "Switch 'invalid' exists but should not")
// }
//
// func TestModifySwitch(t *testing.T) {
// 	if err := CreateRoom(RoomData{Id: "test"}); err != nil {
// 		t.Error(err.Error())
// 	}
//
// 	foobarNode := HardwareNode{
// 		Url:     "http://foo.bar.com",
// 		Name:    "FooBar",
// 		Enabled: true,
// 		Online:  true,
// 		Token:   "foobar",
// 	}
//
// 	if err := CreateHardwareNode(foobarNode); err != nil {
// 		t.Error(err.Error())
// 		return
// 	}
//
// 	table := []struct {
// 		Origin   Switch
// 		Modified Switch
// 	}{
// 		{
// 			Origin: Switch{
// 				Id:         "test_1",
// 				Name:       "Test 1",
// 				RoomId:     "test",
// 				Watts:      0,
// 				PowerOn:    false, // Power is set to false because the power state is not modified
// 				TargetNode: nil,
// 			},
// 			Modified: Switch{
// 				Id:         "test_1",
// 				Name:       "Test 1-2",
// 				RoomId:     "test",
// 				Watts:      1,
// 				PowerOn:    false,
// 				TargetNode: &foobarNode.Url,
// 			},
// 		},
// 		{
// 			Origin: Switch{
// 				Id:         "test_2",
// 				Name:       "Test 2",
// 				RoomId:     "test",
// 				Watts:      2,
// 				PowerOn:    false,
// 				TargetNode: &foobarNode.Url,
// 			},
// 			Modified: Switch{
// 				Id:         "test_2",
// 				Name:       "Test 2-2",
// 				RoomId:     "test",
// 				Watts:      3,
// 				PowerOn:    false,
// 				TargetNode: nil,
// 			},
// 		},
// 		{
// 			Origin: Switch{
// 				Id:         "test_3",
// 				Name:       "Test 3",
// 				RoomId:     "test",
// 				Watts:      3,
// 				PowerOn:    false,
// 				TargetNode: &foobarNode.Url,
// 			},
// 			Modified: Switch{
// 				Id:         "test_3",
// 				Name:       "Test 3",
// 				RoomId:     "test",
// 				Watts:      3,
// 				PowerOn:    false,
// 				TargetNode: &foobarNode.Url,
// 			},
// 		},
// 	}
// 	for _, test := range table {
// 		// Create Switch
// 		if err := CreateDevice(
// 			test.Origin.Id,
// 			test.Origin.Name,
// 			test.Origin.RoomId,
// 			test.Origin.Watts,
// 			test.Origin.TargetNode,
// 		); err != nil {
// 			t.Error(err.Error())
// 		}
//
// 		// Validate Creation
// 		switchDb, found, err := GetDeviceById(test.Origin.Id)
// 		if err != nil {
// 			t.Error(err.Error())
// 		}
// 		assert.True(t, found)
// 		assert.Equal(t, test.Origin, switchDb, "Created switch does not match origin")
//
// 		// Modify Switch
// 		if err := ModifySwitch(
// 			test.Origin.Id,
// 			test.Modified.Name,
// 			test.Modified.Watts,
// 			test.Modified.TargetNode,
// 		); err != nil {
// 			t.Error(err.Error())
// 		}
//
// 		// Validate Modification
// 		switchDb, found, err = GetDeviceById(test.Origin.Id)
// 		if err != nil {
// 			t.Error(err.Error())
// 		}
// 		assert.True(t, found)
// 		assert.Equal(t, test.Modified, switchDb, "Modified switch does not match")
// 	}
// 	invalidSwitch, found, err := GetDeviceById(fmt.Sprint(time.Now().Unix()))
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// 	assert.False(t, found, "invalid switch found")
// 	assert.Empty(t, invalidSwitch, "invalid switch is not empty")
// }

// TODO: add method which tests user switches with modifyRoom permission and powertStates
