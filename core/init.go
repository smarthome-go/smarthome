package core

import (
	"errors"
	"fmt"

	"github.com/smarthome-go/smarthome/core/automation"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
	hardware "github.com/smarthome-go/smarthome/core/hardware_deprecated"
	"github.com/smarthome-go/smarthome/core/homescript"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher"
	"github.com/smarthome-go/smarthome/core/scheduler"
	"github.com/smarthome-go/smarthome/core/user/notify"
	"github.com/smarthome-go/smarthome/services/reminder"
)

func OnMqttRetryHook() error {
	return dispatcher.Instance.RegisterPending()
}

func Init(config database.ServerConfig) error {
	// Homescript Manager initialization
	hmsManager := homescript.InitManager()

	// Mqtt manager initialization
	mqttManager, err := dispatcher.NewMqttManager(config.Mqtt, OnMqttRetryHook)
	if err != nil {
		log.Errorf("MQTT initialization failed: %s", err.Error())
	}

	// Homescript dispatcher initialization
	dispatcher.InitInstance(hmsManager, mqttManager)

	// Homescript driver initialization
	driver.InitManager(hmsManager)
	if err := driver.Manager.PopulateValueCache(); err != nil {
		return err
	}

	if err := driver.Manager.InitDevices(); err != nil {
		return err
	}

	if err := automation.InitManager(hmsManager, config); err != nil {
		return fmt.Errorf("Failed to activate automation system: %s", err.Error())
	}

	notify.InitManager(hmsManager, automation.Manager)

	if err := scheduler.InitManager(hmsManager); err != nil {
		return fmt.Errorf("Failed to activate scheduler system: %s", err.Error())
	}

	if err := reminder.InitSchedule(); err != nil {
		return fmt.Errorf("Failed to activate reminder scheduler: %s", err.Error())
	}

	// Hardware handler
	hardware.Init()

	if err := hardware.StartPowerUsageSnapshotScheduler(); err != nil {
		return fmt.Errorf("Failed to start periodic power usage snapshot scheduler: %s", err.Error())
	}

	return nil
}

func Reload() error {
	var hasErr error

	config, found, err := database.GetServerConfiguration()
	if err != nil {
		log.Errorf("Could not reload core: could not get server config: %s", err.Error())
		return err
	}

	if !found {
		msg := "Could not reload core: no server config present"
		log.Error(msg)
		return errors.New(msg)
	}

	// Reload dispatcher (and MQTT subsystem)
	if err := dispatcher.Instance.Reload(config.Mqtt); err != nil {
		log.Warnf("Could not fully reload core: dispatcher reload error: %s", err.Error())
		hasErr = err
	}

	return hasErr
}

func UpdateMqttConfig(newConfig database.MqttConfig) (reloadErr, dbErr error) {
	if err := database.UpdateMqttConfig(newConfig); err != nil {
		return nil, err
	}

	return Reload(), nil
}
