package homescript

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/exp/utf8string"

	"github.com/smarthome-go/homescript/homescript/interpreter"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/user"
)

type Executor struct {
	ScriptName string
	Username   string
	Output     string
}

// Emulates printing to the console
// Instead, appends the provided message to the output of the executor
// Exists in order to return the script's output to the user
func (self *Executor) Print(args ...string) {
	var output string
	for _, arg := range args {
		self.Output += arg
		output += arg
	}
}

// Returns a boolean if the requested switch is on or off
// Returns an error if the provided switch does not exist
func (self *Executor) SwitchOn(switchId string) (bool, error) {
	powerState, err := hardware.GetPowerState(switchId)
	if err != nil {
		log.Debug(fmt.Sprintf("[Homescript] ERROR: script: '%s' user: '%s': failed to read power state: %s", self.ScriptName, self.Username, err.Error()))
	}
	return powerState, err
}

// Changes the power state of an arbitrary switch
// Checks if the switch exists, if the user is allowed to interact with switches and if the user has the matching switch-permission
// If a check fails, an error is returned
func (self *Executor) Switch(switchId string, powerOn bool) error {
	err := hardware.SetSwitchPowerAll(switchId, powerOn, self.Username)
	if err != nil {
		log.Debug(fmt.Sprintf("[Homescript] ERROR: script: '%s' user: '%s': failed to set power: %s", self.ScriptName, self.Username, err.Error()))
		return err
	}
	onOffText := "on"
	if !powerOn {
		onOffText = "off"
	}
	log.Debug(fmt.Sprintf("[Homescript] script: '%s' user: '%s': turning switch %s %s", self.ScriptName, self.Username, switchId, onOffText))
	return nil
}

// Sends a mode request to a given radiGo server
// TODO: implement this feature
func (self *Executor) Play(server string, mode string) error {
	return errors.New("The feature 'radiGo' is not yet implemented")
}

// Makes a GET request to an arbitrary URL and returns the result
func (self *Executor) Get(url string) (string, error) {
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionHomescriptNetwork)
	if err != nil {
		return "", fmt.Errorf("Could not send GET request: failed to validate your permissions: %s", err.Error())
	}
	if !hasPermission {
		return "", fmt.Errorf("Will not send GET request: you lack permission to access the network via homescript. If this is unintentional, contact your administrator")
	}
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Makes a request to an arbitrary URL using a custom method and body in order to return the result
func (self *Executor) Http(url string, method string, body string) (string, error) {
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionHomescriptNetwork)
	if err != nil {
		return "", fmt.Errorf("Could not send %s request: failed to validate your permissions: %s", method, err.Error())
	}
	if !hasPermission {
		return "", fmt.Errorf("Will not send %s request: you lack permission to access the network via homescript. If this is unintentional, contact your administrator", method)
	}
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(resBody), nil
}

// Sends a notification to the user who issues this command
func (self *Executor) Notify(
	title string,
	description string,
	level interpreter.LogLevel,
) error {
	err := user.Notify(
		self.Username,
		title,
		description,
		user.NotificationLevel(level-1),
	)
	if err != nil {
		log.Error(fmt.Sprintf("[Homescript] ERROR: script: '%s' user: '%s': failed to notify user: %s", self.ScriptName, self.Username, err.Error()))
	}
	return nil
}

// Adds a new user to the system
// If the user already exists, an error is returned
func (self *Executor) AddUser(username string, password string, forename string, surname string) error {
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionManageUsers)
	if err != nil {
		return fmt.Errorf("Failed to add user: could not validate your permissions: %s", err.Error())
	}
	if !hasPermission {
		return errors.New("Failed to add user: You lack permission to manage users: if this is unintentional, contact your administrator.")
	}
	if len(username) == 0 || strings.Contains(username, " ") || !utf8string.NewString(username).IsASCII() {
		return errors.New("Failed to add user: username should only include ASCII characters and must not contain whitespaces or be blank")
	}
	if len(password) == 0 {
		return errors.New("Blank passwords are not allowed")
	}
	if len(username) > 20 || len(forename) > 20 || len(surname) > 20 {
		return errors.New("length of username, forename, and surname must not exceed 20 characters")
	}
	if err := database.AddUser(database.FullUser{
		Username:          username,
		Password:          password,
		Forename:          forename,
		Surname:           surname,
		PrimaryColorDark:  "#88FF70",
		PrimaryColorLight: "#2E7D32",
	}); err != nil {
		return err
	}
	return nil
}

// Deletes a given user: checks whether it is okay to delete this user
func (self *Executor) DelUser(username string) error {
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionManageUsers)
	if err != nil {
		return fmt.Errorf("Failed to remove user: could not validate your permissions: %s", err.Error())
	}
	if !hasPermission {
		return errors.New("Failed to remove user: You lack permission to manage users: if this is unintentional, contact your administrator.")
	}
	isStandaloneUserAdmin, err := user.IsStandaloneUserAdmin(username)
	if err != nil {
		return err
	}
	if isStandaloneUserAdmin {
		return errors.New("Did not delete user: target is the only user-administrator: deleting this user would break the system.")
	}
	if err := user.DeleteUser(username); err != nil {
		return err
	}
	return nil
}

// Adds an arbitrary permission to a given user
func (self *Executor) AddPerm(username string, permission string) error {
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionManageUsers)
	if err != nil {
		return fmt.Errorf("Failed to remove user: could not validate your permissions: %s", err.Error())
	}
	if !hasPermission {
		return errors.New("Failed to remove user: You lack permission to manage users: if this is unintentional, contact your administrator.")
	}
	if !database.DoesPermissionExist(permission) {
		return fmt.Errorf("Failed to add permission: the permission '%s' does not exist. You can view a complete list of valid permissions under user-management > manage user permissions (any user) > permissions", permission)
	}
	edited, err := user.AddPermission(username, database.PermissionType(permission))
	if err != nil {
		return fmt.Errorf("Failed to add permission: database failure: %s", err.Error())
	}
	if !edited {
		return errors.New("Did not add permission: user already has this permission")
	}
	return nil
}

// Removes an arbitrary permission from a given user
func (self *Executor) DelPerm(username string, permission string) error {
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionManageUsers)
	if err != nil {
		return fmt.Errorf("Failed to remove user: could not validate your permissions: %s", err.Error())
	}
	if !hasPermission {
		return errors.New("Failed to remove user: You lack permission to manage users: if this is unintentional, contact your administrator.")
	}
	if !database.DoesPermissionExist(permission) {
		return fmt.Errorf("The permission '%s' does not exist.", permission)
	}
	if permission == string(database.PermissionManageUsers) || permission == string(database.PermissionWildCard) {
		isAlone, err := user.IsStandaloneUserAdmin(username)
		if err != nil {
			return err
		}
		if isAlone {
			return errors.New("Did not remove permission: target user is the only user-administrator: removing this permission would break the system.")
		}
	}
	edited, err := user.RemovePermission(username, database.PermissionType(permission))
	if err != nil {
		return fmt.Errorf("Failed to add permission: database failure: %s", err.Error())
	}
	if !edited {
		return errors.New("User does not have this permission")
	}
	return nil
}

// Adds a log entry to the internal logging system
func (self *Executor) Log(
	title string,
	description string,
	level interpreter.LogLevel,
) error {
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionLogs)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("Failed to add log event: user '%s' is not allowed to use the internal logging system.", self.Username)
	}
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
		return fmt.Errorf("Failed to add log event: invalid logging level <%d>: valid logging levels are 1, 2, 3, 4, or 5", level)
	}
	return nil
}

// Executes another Homescript based on its Id
func (self Executor) Exec(homescriptId string) (string, error) {
	output, exitCode, err := RunById(self.Username, homescriptId)
	if err != nil {
		self.Print(fmt.Sprintf("Exec failed: called homescript failed with exit code %d", exitCode))
		return output, err
	}
	return output, nil
}

// Returns the name of the user who is currently running the script
func (self *Executor) GetUser() string {
	return self.Username
}

// TODO: Will later be implemented, should return the weather as a human-readable string
func (self *Executor) GetWeather() (string, error) {
	return "rainy", nil
}

// TODO: Will later be implemented, should return the temperature in Celsius
func (self *Executor) GetTemperature() (int, error) {
	return 42, nil
}

// Returns the current time variables
func (self *Executor) GetDate() (int, int, int, int, int, int) {
	now := time.Now()
	return now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second()
}
