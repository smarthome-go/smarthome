package homescript

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/go-ping/ping"
	"github.com/smarthome-go/homescript/homescript"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/user"
	"github.com/smarthome-go/smarthome/services/weather"
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

	// Is set to true as soon as a sigTerm is received
	// Is required and read by the manager's run functions and by HMS `exec` calls
	// Used in order to determine whether a script has been terminated using a sigTerm or if it exited conventionally
	WasTerminated bool
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

		// Set the `WasTerminated` boolean to true
		// Is required so that the manager's run functions are informed about this script's death cause
		self.WasTerminated = true

		return true
	default:
		return false
	}
}

// Resolves a Homescript module
func (self *Executor) ResolveModule(id string) (string, bool, error) {
	script, found, err := database.GetUserHomescriptById(id, self.Username)
	if !found || err != nil {
		return "", found, err
	}
	return script.Data.Code, true, nil
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
	self.Output += strings.Join(args, " ")
}

// Emulates printing to the console
// Instead, appends the provided message to the output of the executor
// Exists in order to return the script's output to the user
// Just like `Print` but appends a newline to the end
func (self *Executor) Println(args ...string) {
	if self.DryRun {
		return
	}
	self.Output += strings.Join(args, " ") + "\n"
}

// Returns an object with contains data about the requested switch
// Returns an error if the provided switch does not exist
func (self *Executor) GetSwitch(switchId string) (homescript.SwitchResponse, error) {
	switchData, found, err := database.GetSwitchById(switchId)
	if !found {
		return homescript.SwitchResponse{}, fmt.Errorf("switch '%s' was not found", switchId)
	}
	if err != nil {
		log.Debug(fmt.Sprintf("[Homescript] ERROR: script: '%s' user: '%s': failed to read power state: %s", self.ScriptName, self.Username, err.Error()))
		return homescript.SwitchResponse{}, err
	}
	return homescript.SwitchResponse{
		Name:  switchData.Name,
		Power: switchData.PowerOn,
		Watts: uint(switchData.Watts),
	}, nil
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

// Performs a ICMP ping and returns a boolean which states whether the target host is online or offline
func (self *Executor) Ping(ip string, timeoutSecs float64) (bool, error) {
	// Is still executed because the function tries to resolve the specified IP (allows some checking)
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return false, err
	}
	// If DryRun is being used, stop here
	if self.DryRun {
		return false, nil
	}
	// Perform the ping
	pinger.Count = 1
	pinger.Timeout = time.Millisecond * time.Duration(timeoutSecs*1000)
	err = pinger.Run() // Blocks until the ping is finished or timed-out
	if err != nil {
		return false, err
	}
	stats := pinger.Statistics()
	return stats.PacketsRecv > 0, nil // If at least 1 packet was received back, the host is considered online
}

// Makes a GET request to an arbitrary URL and returns the result
func (self *Executor) Get(requestUrl string) (homescript.HttpResponse, error) {
	// The permissions can be validated beforehand
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionHomescriptNetwork)
	if err != nil {
		return homescript.HttpResponse{}, fmt.Errorf("could not send GET request: failed to validate your permissions: %s", err.Error())
	}
	if !hasPermission {
		return homescript.HttpResponse{}, fmt.Errorf("will not send GET request: you lack permission to access the network via homescript. If this is unintentional, contact your administrator")
	}

	// DryRun only checks the URL's validity
	if self.DryRun {
		// Check if the URL is already cached
		cached, err := database.IsHomescriptUrlCached(requestUrl)
		if err != nil {
			return homescript.HttpResponse{}, fmt.Errorf("internal error: Could not check URL cache: %s", err.Error())
		}
		if cached {
			log.Trace(fmt.Sprintf("Homescript URL `%s` is cached, omitting checks...", requestUrl))
			return homescript.HttpResponse{}, nil
		}
		log.Trace(fmt.Sprintf("Homescript URL `%s` is not cached, running checks...", requestUrl))
		url, err := url.ParseRequestURI(requestUrl)
		if err != nil {
			return homescript.HttpResponse{}, fmt.Errorf("invalid URL provided: could not parse URL: %s", err.Error())
		}
		if url.Scheme != "http" && url.Scheme != "https" {
			return homescript.HttpResponse{}, fmt.Errorf("invalid URL provided: Invalid scheme: `%s`.\n=> Valid schemes are `http` and `https`", url.Scheme)
		}
		if url.Scheme != "http" && url.Scheme != "https" {
			return homescript.HttpResponse{}, fmt.Errorf("invalid URL provided: Invalid scheme: `%s`.\n=> Valid schemes are `http` and `https`", url.Scheme)
		}
		_, err = http.Head(requestUrl)
		if err != nil {
			return homescript.HttpResponse{}, err
		}
		// If all checks were successful, insert the URL into the URL cache
		if err := insertCacheEntry(*url); err != nil {
			return homescript.HttpResponse{}, fmt.Errorf("internal error: Could not update URL cache entry: %s", err.Error())
		}
		log.Trace(fmt.Sprintf("Updated URL cache to include `%s`", requestUrl))
		return homescript.HttpResponse{}, nil
	}

	// Create a new request
	req, err := http.NewRequest(
		http.MethodGet,
		requestUrl,
		nil,
	)
	if err != nil {
		return homescript.HttpResponse{}, err
	}
	// Set the user agent to the Smarthome HMS client
	req.Header.Set("User-Agent", "Smarthome-homescript")

	// Create a new context for cancellatioon
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	req = req.WithContext(ctx)

	// Start the http request monitor Go routine
	requestHasFinished := make(chan struct{})
	go self.httpCancelationMonitor(
		ctx,
		cancel,
		requestHasFinished,
	)

	// Perform the request
	// Create a client for the request
	client := http.Client{
		// Set the client's timeout to 60 seconds
		Timeout: 60 * time.Second,
	}

	res, err := client.Do(req)
	// Evaluate the request's outcome
	if err != nil {
		return homescript.HttpResponse{}, err
	}

	// Stop the request monitor Go routine
	requestHasFinished <- struct{}{}

	// Read request response body
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return homescript.HttpResponse{}, err
	}
	return homescript.HttpResponse{
		Status:     res.Status,
		StatusCode: uint16(res.StatusCode),
		Body:       string(resBody),
	}, nil
}

// Makes a request to an arbitrary URL using a custom method and body in order to return the result
func (self *Executor) Http(requestUrl string, method string, body string, headers map[string]string) (homescript.HttpResponse, error) {
	// Check permissions and request building beforehand
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionHomescriptNetwork)
	if err != nil {
		return homescript.HttpResponse{}, fmt.Errorf("could not perform %s request: failed to validate your permissions: %s", method, err.Error())
	}
	if !hasPermission {
		return homescript.HttpResponse{}, fmt.Errorf("will not perform %s request: you lack permission to access the network via Homescript. If this is unintentional, contact your administrator", method)
	}

	// If using DryRun, stop here and just validate the request URL
	if self.DryRun {
		// Check if the URL is already cached
		cached, err := database.IsHomescriptUrlCached(requestUrl)
		if err != nil {
			return homescript.HttpResponse{}, fmt.Errorf("Internal error: Could not check URL cache: %s", err.Error())
		}
		if cached {
			log.Trace(fmt.Sprintf("Homescript URL `%s` is cached, omitting checks...", requestUrl))
			return homescript.HttpResponse{}, nil
		}
		log.Trace(fmt.Sprintf("Homescript URL `%s` is not cached, running checks...", requestUrl))

		// URL-specific checks
		url, err := url.ParseRequestURI(requestUrl)
		if err != nil {
			return homescript.HttpResponse{}, fmt.Errorf("invalid URL provided: could not parse URL: %s", err.Error())
		}
		if url.Scheme != "http" && url.Scheme != "https" {
			return homescript.HttpResponse{}, fmt.Errorf("invalid URL provided: Invalid scheme: `%s`.\n=> Valid schemes are `http` and `https`", url.Scheme)
		}
		if url.Scheme != "http" && url.Scheme != "https" {
			return homescript.HttpResponse{}, fmt.Errorf("invalid URL provided: Invalid scheme: `%s`.\n=> Valid schemes are `http` and `https`", url.Scheme)
		}
		_, err = http.Head(requestUrl)
		if err != nil {
			return homescript.HttpResponse{}, err
		}
		// If all checks were successful, insert the URL into the URL cache
		if err := insertCacheEntry(*url); err != nil {
			return homescript.HttpResponse{}, fmt.Errorf("internal error: Could not update URL cache entry: %s", err.Error())
		}
		log.Trace(fmt.Sprintf("updated URL cache to include `%s`", requestUrl))
		return homescript.HttpResponse{}, nil
	}

	// Create a new request
	req, err := http.NewRequest(
		method,
		requestUrl,
		strings.NewReader(body),
	)
	if err != nil {
		return homescript.HttpResponse{}, err
	}
	// Set the user agent to the Smarthome HMS client
	req.Header.Set("User-Agent", "Smarthome-homescript")
	// Set the headers included via the function call
	for headerKey, headerValue := range headers {
		req.Header.Set(headerKey, headerValue)
	}

	// Create a new context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	req = req.WithContext(ctx)

	// Start the http request monitor Go routine
	requestHasFinished := make(chan struct{})
	go self.httpCancelationMonitor(
		ctx,
		cancel,
		requestHasFinished,
	)

	// Perform the request
	// Create a client for the request
	client := http.Client{
		// Set the client's timeout to 60 seconds
		Timeout: 60 * time.Second,
	}
	res, err := client.Do(req)
	// Evaluate the request's outcome
	if err != nil {
		return homescript.HttpResponse{}, err
	}

	// Stop the request monitor Go routine
	requestHasFinished <- struct{}{}

	// Read request response body
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return homescript.HttpResponse{}, err
	}
	return homescript.HttpResponse{
		Status:     res.Status,
		StatusCode: uint16(res.StatusCode),
		Body:       string(resBody),
	}, nil
}

func (self *Executor) httpCancelationMonitor(
	cnt context.Context,
	cancelRequest context.CancelFunc,
	requestHasFinished chan struct{},
) {
	log.Trace("Started Homescript http request monitoring")
	defer log.Trace("Finished Homescript http request monitoring")
	for {
		select {
		// If the request has finished regularely, stop the monitor and to nothing
		case <-requestHasFinished:
			log.Trace("Homescript http request finished regularely")
			return
		// If the request has not finished, run the check below
		default:
			// If a sigTerm is received whilst waiting for the request to be completed, cancel the request and stop the monitor
			if self.checkSigTerm() {
				cancelRequest()
				log.Debug("Detected sigTerm while waiting for Homescript http request to be completed: canceling request...")
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// Sends a notification to the user who issues this command
func (self *Executor) Notify(
	title string,
	description string,
	level homescript.NotificationLevel,
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

// Adds a log entry to the internal logging system
func (self *Executor) Log(
	title string,
	description string,
	level homescript.LogLevel,
) error {
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionLogging)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("Failed to add log event: you lack permission to add records to the internal logging system.")
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
func (self Executor) Exec(homescriptId string, args map[string]string) (homescript.ExecResponse, error) {
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
			return homescript.ExecResponse{}, fmt.Errorf("Exec violation: executing '%s' could cause infinite recursion.\n%s", homescriptId, callStackVisual)
		}
	}
	// Execute the target script after the checks
	start := time.Now()
	res, err := HmsManager.RunById(
		homescriptId,
		self.Username,
		// The previous CallStack is passed in order to preserve the history
		// Because the RunById function implicitly appends its target id the provided call stack, it doesn't need to be added here
		self.CallStack,
		self.DryRun,
		args,
		InitiatorExec,
		self.SigTerm,
	)
	// Check if the script was killed using a sigTerm
	if res.WasTerminated {
		self.WasTerminated = true
		return homescript.ExecResponse{}, fmt.Errorf("Exec received sigTerm whilst processing Homescript `%s`", homescriptId)
	}
	if err != nil {
		return homescript.ExecResponse{}, err
	}
	return homescript.ExecResponse{
		Output:      res.Output,
		RuntimeSecs: float64(time.Since(start).Seconds()),
		ReturnValue: res.ReturnValue,
		RootScope:   res.RootScope,
	}, nil
}

// Returns the name of the user who is currently running the script
func (self *Executor) GetUser() string {
	return self.Username
}

func (self *Executor) GetWeather() (homescript.Weather, error) {
	if self.DryRun {
		return homescript.Weather{}, nil
	}
	wthr, err := weather.GetCurrentWeather()
	if err != nil {
		return homescript.Weather{}, fmt.Errorf("could not fetch weather: %s", err.Error())
	}
	return homescript.Weather{
		WeatherTitle:       wthr.WeatherTitle,
		WeatherDescription: wthr.WeatherDescription,
		Temperature:        float64(wthr.Temperature),
		FeelsLike:          float64(wthr.FeelsLike),
		Humidity:           wthr.Humidity,
	}, nil
}
