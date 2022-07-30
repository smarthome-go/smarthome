package homescript

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/exp/utf8string"

	"github.com/smarthome-go/homescript/homescript/interpreter"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/user"
	"github.com/smarthome-go/smarthome/core/utils"
)

type Executor struct {
	// Is required for error handling and recursion prevention
	// Allows pretty-print for potential errors
	ScriptName string

	// Specifies the time a script was started
	// Can be used to keep track of the script's runtime
	StartTime time.Time

	// The Username is required for functions which rely on permissions-check
	// or need to access the username for other reasons, e.g. `notify`
	Username string

	// Will be appended to when the print function is used
	// Is required in order to return the complete output of a Homescript
	Output string

	// If set to true, a script will only check its correctness
	// Does not actually modify or wait for any data
	DryRun bool

	// Holds the script's arguments as a map
	// Is filled by a helper function like `Run`
	// Is used by the `CheckArg` and `GetArg` methods for providing args
	Args map[string]string

	// The CallStack saves the history of which script called another
	// Additionally, it specifies which Homescripts have to be excluded from the `exec` function
	// Is required in order to prevent recursion until the system's database runs out of resources
	// Acts like a blacklist which holds the blacklisted Homescript ids
	// The last item in the CallStack is the script which was called the most recently (from script exit)
	CallStack []string

	// Pointer to the interpreter's sigTerm channel
	// Is used here in order to allow the abortion during expensive operations, e.g. the sleep function
	sigTermInternalPtr *chan int

	// Sigterm receiver which is visible to the outside
	// Any signal will be forwarded to the internal `sigTermPtr`
	SigTerm chan int
}

// Is used to allow the abort of a running script at any point in time
// => Checks if a sigTerm has been received
// Is used to break out of expensive operations, for example sleep calls
// Only a bool static that a code has been received is returned
// => The real sigTerm handling is done in the AST execution of the interpreter
func (self *Executor) checkSigTerm() bool {
	select {
	case code := <-self.SigTerm:
		// Forwards the signal to the interpreter
		go func() {
			// This goroutine is required because otherwise,
			// The sending of the signal would block forever
			// This is due to the interpreter only handling sigTerms on every AST-node
			// However, the interpreter will only handle the next node if this function's caller quits
			// Because of this, not using a goroutine would invoke a deadlock
			*self.sigTermInternalPtr <- code
		}()
		return true
	default:
		return false
	}
}

// Validates that a given argument has been passed to the Homescript runtime
// Returns a boolean indicating whether the argument has been found in the `Args` map
func (self *Executor) CheckArg(toCheck string) bool {
	if self.DryRun {
		return true
	}
	_, ok := self.Args[toCheck]
	return ok
}

// Returns the value of an expected argument from the `Args` map
// If the value could not be found in the map, it was not provided to the Homescript runtime.
// This situation will cause the function to return an error, so `CheckArg` should be used beforehand
func (self *Executor) GetArg(toGet string) (string, error) {
	if self.DryRun {
		return "", nil
	}
	value, ok := self.Args[toGet]
	if !ok {
		return "", fmt.Errorf("Failed to retrieve argument '%s': not provided to the Homescript runtime", toGet)
	}
	return value, nil
}

// Pauses the execution of the current script for the amount of the specified seconds
// Implements special checks to cancel the sleep function during its execution
func (self *Executor) Sleep(seconds float64) {
	if self.DryRun {
		return
	}
	for i := 0; i < int(seconds*1000); i += 10 {
		if self.checkSigTerm() {
			// Sleep function is terminated
			// Additional wait time is used to dispatch the signal to the interpreter
			time.Sleep(time.Millisecond * 10)
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
}

// Emulates printing to the console
// Instead, appends the provided message to the output of the executor
// Exists in order to return the script's output to the user
func (self *Executor) Print(args ...string) {
	if self.DryRun {
		return
	}
	for _, arg := range args {
		self.Output += arg + "\n"
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
	// If running in DryRun, only check the values
	if self.DryRun {
		return self.testSwitch(switchId, powerOn)
	}
	// Actual function implementation
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

// Used for DryRun, only checks the existence of the specified switch and the user's permissions
func (self *Executor) testSwitch(switchId string, powerOn bool) error {
	_, switchExists, err := database.GetSwitchById(switchId)
	if err != nil {
		return err
	}
	if !switchExists {
		return fmt.Errorf("Failed to set power: switch '%s' does not exist", switchId)
	}
	userHasPowerPermission, err := database.UserHasPermission(self.Username, database.PermissionPower)
	if err != nil {
		return fmt.Errorf("Failed to set power: could not check if user is allowed to interact with switches: %s", err.Error())
	}
	if !userHasPowerPermission {
		return errors.New("Failed to set power: user is not allowed to interact with switches")
	}
	userHasSwitchPermission, err := database.UserHasSwitchPermission(self.Username, switchId)
	if err != nil {
		return fmt.Errorf("Failed to set power: could not check if user is allowed to interact with this switch: %s", err.Error())
	}
	if !userHasSwitchPermission {
		return fmt.Errorf("Failed to set power: user is not allowed to interact with switch '%s'", switchId)
	}
	return nil
}

// Makes a GET request to an arbitrary URL and returns the result
func (self *Executor) Get(requestUrl string) (string, error) {
	// The permissions can be validated beforehand
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionHomescriptNetwork)
	if err != nil {
		return "", fmt.Errorf("Could not send GET request: failed to validate your permissions: %s", err.Error())
	}
	if !hasPermission {
		return "", fmt.Errorf("Will not send GET request: you lack permission to access the network via homescript. If this is unintentional, contact your administrator")
	}
	// DryRun only checks the URL's validity
	if self.DryRun {
		_, err := url.ParseRequestURI(requestUrl)
		if err != nil {
			return "", fmt.Errorf("Invalid URL provided: could not parse URL: %s", err.Error())
		}
		return "", nil
	}
	res, err := http.Get(requestUrl)
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
func (self *Executor) Http(requestUrl string, method string, body string, headers map[string]string) (string, error) {
	// Check permissions and request building beforehand
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionHomescriptNetwork)
	if err != nil {
		return "", fmt.Errorf("Could not perform %s request: failed to validate your permissions: %s", method, err.Error())
	}
	if !hasPermission {
		return "", fmt.Errorf("Will not perform %s request: you lack permission to access the network via Homescript. If this is unintentional, contact your administrator", method)
	}
	req, err := http.NewRequest(method, requestUrl, strings.NewReader(body))
	if err != nil {
		return "", err
	}

	// Set the user agent to the Smarthome HMS client
	req.Header.Set("User-Agent", fmt.Sprintf("Smarthome-homescript/%s", utils.Version))

	// Set the headers included via the function call
	for headerKey, headerValue := range headers {
		req.Header.Set(headerKey, headerValue)
	}

	client := http.Client{
		Timeout: 60 * time.Second,
	}
	// If using DryRun, stop here
	if self.DryRun {
		_, err := url.ParseRequestURI(requestUrl)
		if err != nil {
			return "", fmt.Errorf("Invalid URL provided: could not parse URL: %s", err.Error())
		}
		return "", nil
	}
	res, err := client.Do(req)
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
	// If using DryRun, stop here
	if self.DryRun {
		return nil
	}
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
		return errors.New("Length of username, forename, and surname must not exceed 20 characters")
	}
	_, found, err := database.GetUserByUsername(username)
	if err != nil {
		return fmt.Errorf("Failed to check existence of user: %s", err.Error())
	}
	if found {
		return fmt.Errorf("Will not add user: user '%s' already exists", username)
	}
	// If using DryRun, stop here
	if self.DryRun {
		return nil
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
	// If using DryRun, stop here
	if self.DryRun {
		return nil
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
	// If using DryRun, stop here
	if self.DryRun {
		return nil
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
	// If using DryRun, stop here
	if self.DryRun {
		return nil
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
	// If using DryRun, stop here
	if self.DryRun {
		if level > 5 {
			return fmt.Errorf("Failed to add log event: invalid logging level <%d>: valid logging levels are 1, 2, 3, 4, or 5", level)
		}
		return nil
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
func (self Executor) Exec(homescriptId string, args map[string]string) (string, error) {
	// The dryRun value is passed to the executed script
	// Before the target script can begin execution, the call stack is analyzed in order to prevent recursion

	// If the CallStack is empty, the script was initially called by string
	// In this case, the own id must be appended to the CallStack fist
	if len(self.CallStack) == 0 {
		self.CallStack = append(self.CallStack, self.ScriptName)
	}

	// Analyze call stack
	for _, call := range self.CallStack {
		if homescriptId == call {
			// Would call a script which is already located in the CallStack (upstream)
			// In order to show the problem to the user, the stack is unwinded and transformed into a pretty display
			callStackVisual := "=== Call Stack ===\n"
			for callIndex, callVis := range self.CallStack {
				if callIndex == 0 {
					callStackVisual += fmt.Sprintf("  %2d: %-10s (INITIAL)\n", 0, self.CallStack[0])
				} else {
					callStackVisual += fmt.Sprintf("  %2d: %-10s\n", callIndex, callVis)
				}
			}
			callStackVisual += fmt.Sprintf("  %2d: %-10s (PREVENTED)\n", len(self.CallStack), homescriptId)
			return self.Output, fmt.Errorf("Exec violation: executing '%s' could cause infinite recursion.\n%s", homescriptId, callStackVisual)
		}
	}

	// Execute the target script after the checks
	output, exitCode, err := HmsManager.RunById(
		homescriptId,
		self.Username,
		// The previous CallStack is passed in order to preserve the history
		// Because the RunById function implicitly appends its target id the provided call stack, it doesn't need to be added here
		self.CallStack,
		self.DryRun,
		args,
		InitiatorExec,
	)
	if err != nil {
		self.Print(fmt.Sprintf("Exec failed: called Homescript failed with exit code %d", exitCode))
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
	if self.DryRun {
		return "", nil
	}
	return "rainy", nil
}

// TODO: Will later be implemented, should return the temperature in Celsius
func (self *Executor) GetTemperature() (int, error) {
	if self.DryRun {
		return 0, nil
	}
	return 42, nil
}

// Returns the current time variables
func (self *Executor) GetDate() (int, int, int, int, int, int) {
	now := time.Now()
	return now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second()
}
