package core

import (
	"errors"
	"fmt"
	"sync"

	"github.com/smarthome-go/smarthome/core/automation"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
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

	return dispatcher.Instance.RegisterPending()
}

// TODO: this mostly works, but breaks right now

func InitUserScripts() error {
	scripts, err := database.ListAllHomescripts()
	if err != nil {
		return err
	}

	for _, script := range scripts {
		if err := dispatcher.Instance.RegisterUserScript(script.Data.Id, script.Owner); err != nil {
			log.Debugf("Failed to initialize user script (%s): %s", script.Data.Id, err.Error())
		}
	}

	return nil
}

func InitDevices() error {
	// Compile every driver's source code (register any triggers if existent)
	// TODO: implement this in a better way
	devices, err := driver.Manager.ListAllDevicesRich()
	if err != nil {
		return err
	}

	for idx, device := range devices {
		fmt.Printf("=== %02d | (%s) %s\n", idx, device.Shallow.DeviceType, device.Shallow.Name)
		fmt.Printf("\t -> errors=%v\n", device.Extractions.HmsErrors)
	}

	// drivers, err := driver.Manager.List

	// 	driver.Manager.InvokeValidateCheckDriver(types.DriverInvocationIDs{
	// 		DeviceID: new(string),
	// 		VendorID: "",
	// 		ModelID:  "",
	// 	})

	err = dispatcher.Instance.RegisterDriverAnnotations()

	driver.Manager.Initialized.Lock.Lock()
	driver.Manager.Initialized.Value = true
	driver.Manager.Initialized.Lock.Unlock()

	return err
}

func Init(config database.ServerConfig) error {
	log.Debug("Initializing smarthome core...")

	// Homescript Manager initialization
	hmsManager := homescript.InitManager()

	dispatcher.InitModule()

	// Mqtt manager initialization (connection deferred until after registrations are queued)
	mqttManager, err := dispatcher.NewMqttManager(config.Mqtt, OnMqttRetryHook)
	if err != nil {
		log.Errorf("MQTT manager creation failed: %s", err.Error())
	}

	// Homescript dispatcher initialization (MQTT connection is deferred)
	disp, err := dispatcher.InitInstance(hmsManager, mqttManager)
	if err != nil {
		log.Errorf("Failed to initialize HMS dispatcher: %s", err.Error())
	}

	// Homescript driver initialization
	log.Debugf("Initializing driver manager...")
	driver.InitManager(hmsManager, disp.DriverReloadCallBackFn, disp.DeviceReloadCallBackFn)
	if err := driver.Manager.PopulateValueCache(); err != nil {
		return err
	}
	log.Debugf("Value cache initialized.")

	if err := automation.InitManager(hmsManager, config); err != nil {
		return fmt.Errorf("failed to activate automation system: %s", err.Error())
	}

	notify.InitManager(hmsManager, automation.Manager)

	if err := scheduler.InitManager(hmsManager); err != nil {
		return fmt.Errorf("failed to activate scheduler system: %s", err.Error())
	}

	if err := reminder.InitSchedule(); err != nil {
		return fmt.Errorf("failed to activate reminder scheduler: %s", err.Error())
	}

	if err := driver.StartPowerUsageSnapshotScheduler(); err != nil {
		return fmt.Errorf("failed to start periodic power usage snapshot scheduler: %s", err.Error())
	}

	//
	// Devices (registrations are queued as pending since MQTT is not connected yet).
	//

	log.Debug("Initializing devices...")
	if err := InitDevices(); err != nil {
		log.Warnf("Failed to initialize all devices, using best effort attempt: %s", err.Error())
	}

	//
	// Init user scripts
	//

	if err := InitUserScripts(); err != nil {
		log.Errorf("Failed to initialize all user programs, using best effort attempt: %s", err.Error())
	}

	//
	// Now connect MQTT and process all pending registrations in one coordinated pass.
	//

	dispatcherInitialized.lock.Lock()
	dispatcherInitialized.value = true
	dispatcherInitialized.lock.Unlock()

	if err := disp.ConnectMqtt(); err != nil {
		log.Errorf("MQTT connection failed: %s", err.Error())
	} else {
		if err := dispatcher.Instance.RegisterPending(); err != nil {
			log.Warnf("Failed to register some pending MQTT subscriptions: %s", err.Error())
		}
	}

	mqttManager.EndBootPhase()

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
		msg := "could not reload core: no server config present"
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
