package hardware

import (
	"strings"
	"sync"
	"testing"

	"github.com/smarthome-go/smarthome/core/database"
)

func TestSetPower(t *testing.T) {
	if err := initDB(true); err != nil {
		t.Error(err.Error())
		return
	}
	if err := database.CreateHardwareNode(database.HardwareNode{
		Name:    "test",
		Online:  true,
		Enabled: true,
		Url:     "http://localhost",
		Token:   "",
	}); err != nil {
		t.Error(err.Error())
		return
	}
	table := []struct {
		Switch string
		Power  bool
		Error  string
	}{
		{
			Switch: "test",
			Power:  true,
			// Only the first request will throw an error due to node being marked as offline
			Error: `Post "http://localhost/power?token=": dial tcp`, // Different on other machines
		},
		{
			Switch: "test",
			Power:  false,
			Error:  ``,
		},
		{
			Switch: "test2",
			Power:  true,
			Error:  ``,
		},
		{
			Switch: "test2",
			Power:  false,
			Error:  ``,
		},
		{
			Switch: "test3",
			Power:  true,
			Error:  ``,
		},
		{
			Switch: "test3",
			Power:  false,
			Error:  ``,
		},
		{
			Switch: "test4",
			Power:  true,
			Error:  ``,
		},
		{
			Switch: "test4",
			Power:  false,
			Error:  ``,
		},
	}
	// Create a test room
	if err := database.CreateRoom(database.RoomData{Id: "test", Name: "test", Description: "test"}); err != nil {
		t.Error("Failed to create room:", err.Error())
		return
	}
	for _, req := range table {
		if err := database.CreateSwitch(req.Switch, req.Switch, "test", 0, nil); err != nil {
			t.Error(err.Error())
			return
		}

		switchItem, found, err := database.GetSwitchById(req.Switch)
		if err != nil {
			t.Error(err.Error())
			return
		}

		if !found {
			t.Errorf("Switch `%s` was just created but could not be found", req.Switch)
			return
		}

		if err := SetPower(switchItem, req.Power); err != nil {
			if !strings.Contains(err.Error(), req.Error) || req.Error == "" {
				t.Errorf("Unexpected error: want: `%s` got: `%s`", req.Error, err.Error())
				return
			}
		} else if req.Error != "" {
			t.Errorf("Expected error: want: `%s` got: `%s`", req.Error, "")
			return
		}
		if len(GetResults()) > 0 {
			t.Errorf("Some results have not been consumed. want: 0 got: %d", len(GetResults()))
			return
		}
		powerState, err := GetPowerState(req.Switch)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if powerState != req.Power {
			t.Errorf("Power state unaffected: want: `%t` got: `%t`", req.Power, powerState)
			return
		}
	}
	// Errors of last running daemon should be 0 due to many daemons being used above
	if GetJobsWithErrorInHandler() > 0 {
		t.Errorf("Invalid jobs with error count. want: %d got: %d", 0, GetJobsWithErrorInHandler())
		return
	}
	// When no node is registered, the request should not fail
	if err := database.DeleteHardwareNode("http://localhost"); err != nil {
		t.Error(err.Error())
		return
	}

	switchItem, found, err := database.GetSwitchById("test")
	if err != nil {
		t.Error(err.Error())
		return
	}

	if !found {
		t.Errorf("Switch `%s` was just created but could not be found", "test")
		return
	}

	if err := SetPower(switchItem, false); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestSetPowerAsync(t *testing.T) {
	if err := initDB(true); err != nil {
		t.Error(err.Error())
		return
	}
	if err := database.CreateHardwareNode(database.HardwareNode{
		Name:    "test",
		Online:  true,
		Enabled: true,
		Url:     "http://localhost",
		Token:   "",
	}); err != nil {
		t.Error(err.Error())
		return
	}
	table := []struct {
		Switch string
		Power  bool
		Error  string
	}{
		{
			Switch: "test_1",
			Power:  true,
			// Only the first request will throw an error due to node being marked as offline
			Error: `Post "http://localhost/power?token=": dial tcp`, // Different on other machines
		},
		{
			Switch: "test_1",
			Power:  false,
			Error:  ``,
		},
		{
			Switch: "test_2",
			Power:  true,
			Error:  ``,
		},
		{
			Switch: "test_2",
			Power:  false,
			Error:  ``,
		},
		{
			Switch: "test_3",
			Power:  true,
			Error:  ``,
		},
		{
			Switch: "test_3",
			Power:  false,
			Error:  ``,
		},
		{
			Switch: "test_4",
			Power:  true,
			Error:  ``,
		},
		{
			Switch: "test_4",
			Power:  false,
			Error:  ``,
		},
	}
	// Create a test room
	if err := database.CreateRoom(database.RoomData{Id: "test", Name: "test", Description: "test"}); err != nil {
		t.Error("Failed to create room:", err.Error())
		return
	}
	var wg sync.WaitGroup
	for _, req := range table {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := database.CreateSwitch(req.Switch, req.Switch, "test", 0, nil); err != nil {
				t.Error(err.Error())
				return
			}

			switchItem, found, err := database.GetSwitchById(req.Switch)
			if err != nil {
				t.Error(err.Error())
				return
			}

			if !found {
				t.Errorf("Switch `%s` was just created but could not be found", req.Switch)
				return
			}

			if err := SetPower(switchItem, req.Power); err != nil {
				if !strings.Contains(err.Error(), req.Error) || req.Error == "" {
					t.Errorf("Unexpected error: want: `%s` got: `%s`", req.Error, err.Error())
					return
				}
			} else if req.Error != "" {
				t.Errorf("Expected error: want: `%s` got: `%s`", req.Error, "")
				return
			}
			if len(GetPendingJobs()) != GetPendingJobCount() {
				t.Errorf("Pending job count (%d) does not match count of current jobs (%d)", GetPendingJobCount(), len(GetPendingJobs()))
				return
			}
		}()
		wg.Wait()
		// When no node is registered, the request should not fail
		if err := database.DeleteHardwareNode("http://localhost"); err != nil {
			t.Error(err.Error())
			return
		}

		switchItem, found, err := database.GetSwitchById("test")
		if err != nil {
			t.Error(err.Error())
			return
		}

		if !found {
			t.Errorf("Switch `%s` was just created but could not be found", req.Switch)
			return
		}

		if err := SetPower(switchItem, false); err != nil {
			t.Error(err.Error())
			return
		}
		if GetPendingJobCount() != 0 {
			t.Errorf("Current job count invalid. want: 0 got: %d", GetPendingJobCount())
			return
		}
	}
}
