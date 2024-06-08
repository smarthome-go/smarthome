package core

import (
	"errors"
	"fmt"
	"sync"

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

var dispatcherInitialized = struct {
	lock  sync.Mutex
	value bool
}{
	lock:  sync.Mutex{},
	value: false,
}

func OnMqttRetryHook() error {
	dispatcherInitialized.lock.Lock()
	initialized := dispatcherInitialized.value
	dispatcherInitialized.lock.Unlock()

	if !initialized {
		return nil
	}

	// TODO: weird mutex errors, use a channel, it would be better!

	return dispatcher.Instance.RegisterPending()
}

func InitDevices() error {
	// Compile every driver's source code (register any triggers if existent)
	// TODO: implement this in a better way
	// devices, err := driver.Manager.ListAllDevicesRich()
	// if err != nil {
	// 	return err
	// }

	// for idx, device := range devices {
	// 	fmt.Printf("=== %02d | (%s) %s\n", idx, device.Shallow.DeviceType, device.Shallow.Name)
	// 	fmt.Printf("\t -> errors=%v\n", device.Extractions.HmsErrors)
	// }

	return dispatcher.Instance.RegisterDriverAnnotations()
}

func Init(config database.ServerConfig) error {
	// Homescript Manager initialization
	hmsManager := homescript.InitManager()

	dispatcher.InitModule()

	// Mqtt manager initialization
	mqttManager, err := dispatcher.NewMqttManager(config.Mqtt, OnMqttRetryHook)
	if err != nil {
		log.Errorf("MQTT initialization failed: %s", err.Error())
	}

	// Homescript dispatcher initialization
	if err := dispatcher.InitInstance(hmsManager, mqttManager); err != nil {
		log.Errorf("Failed to initialize HMS dispatcher: %s", err.Error())
	}

	dispatcherInitialized.lock.Lock()
	dispatcherInitialized.value = true
	dispatcherInitialized.lock.Unlock()

	// Homescript driver initialization
	driver.InitManager(hmsManager)
	if err := driver.Manager.PopulateValueCache(); err != nil {
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

	//
	// Devices.
	//

	if err := InitDevices(); err != nil {
		log.Errorf("Failed to initialize all devices, using best effort attempt: %s", err.Error())
	}

	//
	// END Devices.
	//

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
