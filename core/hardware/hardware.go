package hardware

import (
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/smarthome-go/smarthome/core/database"
)

// TODO: this must be implemented cleaner:
// this is the data that is handed over to the driver
type DeviceJobData any

type DeviceOutputJob struct {
	Id     int64           `json:"id"`
	Device database.Device `json:"device"`
	Data   DeviceJobData   `json:"data"`
}

type JobResult struct {
	Id    int64 `json:"id"`
	Error error `json:"error"`
}

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

var ErrorLockDownMode = errors.New("cannot set power: lockdown mode is enabled")

// // As setPower, just with additional logs and account for taking a snapshot of the power states
// func SetPower(switchItem database.Switch, powerOn bool) error {
// 	// Check if lockdown mode is enabled
// 	config, _, err := database.GetServerConfiguration()
// 	if err != nil {
// 		return err
// 	}
// 	if config.LockDownMode {
// 		log.Warn("Cannot set power: lockdown mode is enabled")
// 		return ErrorLockDownMode
// 	}
// 	if err := setPower(switchItem, powerOn); err != nil {
// 		go event.Warn(
// 			"Hardware Error",
// 			fmt.Sprintf("The hardware failed while a user tried to interact with switch '%s': Error: %s",
// 				switchItem.Id,
// 				err.Error(),
// 			),
// 		)
// 		return err
// 	}
// 	// Take a snapshot which includes the power states after the switch has been modified
// 	go SaveCurrentPowerUsageWithLogs()
// 	// Add event logs that inform about the switch power change
// 	if powerOn {
// 		go event.Info("Switch Activated", fmt.Sprintf("Switch '%s' was activated", switchItem.Id))
// 	} else {
// 		go event.Info("Switch Deactivated", fmt.Sprintf("Switch '%s' was deactivated", switchItem.Id))
// 	}
// 	return nil
// }

// Sets the power-state of a specific switch
// Checks if the switch exists
// Checks if the user has all required permissions
// Sends a power request to all available nodes
// func SetSwitchPowerAll(switchId string, powerOn bool, username string) error {
// 	switchItem, switchExists, err := database.GetDeviceById(switchId)
// 	if err != nil {
// 		return err
// 	}
// 	if !switchExists {
// 		return fmt.Errorf("Failed to set power: switch '%s' does not exist", switchId)
// 	}
// 	userHasPowerPermission, err := database.UserHasPermission(username, database.PermissionPower)
// 	if err != nil {
// 		return fmt.Errorf("Failed to set power: could not check if user is allowed to interact with switches: %s", err.Error())
// 	}
// 	if !userHasPowerPermission {
// 		return errors.New("Failed to set power: user is not allowed to interact with switches")
// 	}
// 	userHasSwitchPermission, err := database.UserHasSwitchPermission(username, switchId)
// 	if err != nil {
// 		return fmt.Errorf("Failed to set power: could not check if user is allowed to interact with this switch: %s", err.Error())
// 	}
// 	if !userHasSwitchPermission {
// 		return fmt.Errorf("Failed to set power: user is not allowed to interact with switch '%s'", switchId)
// 	}
// 	if err := SetPower(switchItem, powerOn); err != nil {
// 		return fmt.Errorf("Failed to set power: hardware error: %s", err.Error())
// 	}
// 	return nil
// }
