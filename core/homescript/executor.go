package homescript

import (
	"errors"
	"fmt"
	"time"

	"github.com/MikMuellerDev/homescript/homescript/interpreter"
	"github.com/MikMuellerDev/smarthome/core/event"
	"github.com/MikMuellerDev/smarthome/core/hardware"
	"github.com/MikMuellerDev/smarthome/core/user"
)

type Executor struct {
	ScriptName string
	Username   string
}

func (self Executor) Exit(code int) {
	// TODO: implement an actual quit
}

func (self Executor) Print(args ...string) {
	var output string
	for _, arg := range args {
		output += arg
	}
	log.Info(fmt.Sprintf("[Homescript] script: '%s' user: '%s': %s", self.ScriptName, self.Username, output))
}

func (self Executor) SwitchOn(switchId string) (bool, error) {
	powerState, err := hardware.GetPowerState(switchId)
	if err != nil {
		log.Error(fmt.Sprintf("[Homescript] ERROR: script: '%s' user: '%s': failed to read power state: %s", self.ScriptName, self.Username, err.Error()))
	}
	return powerState, err
}

func (self Executor) Switch(switchId string, powerOn bool) error {
	err := hardware.SetSwitchPowerAll(switchId, powerOn, self.Username)
	if err != nil {
		log.Error(fmt.Sprintf("[Homescript] ERROR: script: '%s' user: '%s': failed to set power: %s", self.ScriptName, self.Username, err.Error()))
		return err
	}
	onOffText := "on"
	if !powerOn {
		onOffText = "off"
	}
	log.Debug(fmt.Sprintf("[Homescript] script: '%s' user: '%s': turning switch %s %s", self.ScriptName, self.Username, switchId, onOffText))
	return nil
}

func (self Executor) Play(server string, mode string) error {
	return errors.New("The feature 'radiGo' is not yet implemented")
}

func (self Executor) Notify(
	title string,
	description string,
	level interpreter.LogLevel,
) error {
	err := user.Notify(
		self.Username,
		title,
		description,
		user.NotificationLevel(level),
	)
	if err != nil {
		log.Error(fmt.Sprintf("[Homescript] ERROR: script: '%s' user: '%s': failed to notify user: %s", self.ScriptName, self.Username, err.Error()))
	}
	return nil
}

func (self Executor) Log(
	title string,
	description string,
	level interpreter.LogLevel,
) error {
	switch level {
	case 0:
		event.Trace(title, description)
	case 1:
		event.Debug(title, description)
	case 2:
		event.Info(title, description)
	case 3:
		event.Warn(title, description)
	case 4:
		event.Error(title, description)
	case 5:
		event.Fatal(title, description)
	default:
		log.Error(fmt.Sprintf("[Homescript] ERROR: script: '%s' user: '%s': failed to log event: invalid level", self.ScriptName, self.Username))
	}
	return nil
}

func (self Executor) GetUser() string {
	return self.Username
}

func (self Executor) GetWeather() (string, error) {
	log.Error(fmt.Sprintf("[Homescript] ERROR: script: '%s' user: '%s': weather is not implemented yet", self.ScriptName, self.Username))
	return "rainy", nil
}

func (self Executor) GetTemperature() (int, error) {
	log.Error(fmt.Sprintf("[Homescript] ERROR: script: '%s' user: '%s': temperature is not implemented yet", self.ScriptName, self.Username))
	return 42, nil
}

func (self Executor) GetDate() (int, int, int, int, int, int) {
	now := time.Now()
	return now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second()
}
