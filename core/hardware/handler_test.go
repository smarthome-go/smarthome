package hardware

import (
	"strings"
	"sync"
	"testing"

	"github.com/smarthome-go/smarthome/core/database"
	"github.com/stretchr/testify/assert"
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
			Error:  `There are no nodes or all nodes are disabled, no action performed`,
		},
		{
			Switch: "test4",
			Power:  true,
			Error:  `There are no nodes or all nodes are disabled, no action performed`,
		},
	}
	// Create a test room
	if err := database.CreateRoom(database.RoomData{Id: "test", Name: "test", Description: "test"}); err != nil {
		t.Error("Failed to create room:", err.Error())
		return
	}
	for _, req := range table {
		if err := database.CreateDevice(req.Switch, req.Switch, "test", 0, nil); err != nil {
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
		if powerState == req.Power {
			t.Errorf("Power state affected: want: `%t` got: `%t`", req.Power, powerState)
			return
		}
	}
	// The last daemon had one job which failed
	if GetJobsWithErrorInHandler() > 1 {
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

	err = SetPower(switchItem, false)
	assert.Error(t, err)
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
			Switch: "async_test_1",
			Power:  true,
			Error:  `There are no nodes or all nodes are disabled, no action performed`,
		},
		{
			Switch: "async_test_1",
			Power:  true,
			Error:  `There are no nodes or all nodes are disabled, no action performed`,
		},
		{
			Switch: "async_test_2",
			Power:  false,
			Error:  `There are no nodes or all nodes are disabled, no action performed`,
		},
		{
			Switch: "async_test_3",
			Power:  true,
			Error:  `There are no nodes or all nodes are disabled, no action performed`,
		},
		{
			Switch: "async_test_3",
			Power:  false,
			Error:  `There are no nodes or all nodes are disabled, no action performed`,
		},
		{
			Switch: "async_test_4",
			Power:  true,
			Error:  `There are no nodes or all nodes are disabled, no action performed`,
		},
		{
			Switch: "async_test_4",
			Power:  false,
			Error:  `There are no nodes or all nodes are disabled, no action performed`,
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
			if err := database.CreateDevice(req.Switch, req.Switch, "test", 0, nil); err != nil {
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

		switchItem, found, err := database.GetSwitchById(req.Switch)
		if err != nil {
			t.Error(err.Error())
			return
		}

		if !found {
			t.Errorf("Switch `%s` was just created but could not be found", req.Switch)
			return
		}

		if err := SetPower(switchItem, false); err != nil {
			log.Error(err.Error())
		}

		if GetPendingJobCount() != 0 {
			t.Errorf("Current job count invalid. want: 0 got: %d", GetPendingJobCount())
			return
		}
	}
}
