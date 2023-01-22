package homescript

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"unicode/utf8"

	"github.com/go-ping/ping"
	"github.com/smarthome-go/homescript/v2/homescript"
	"github.com/smarthome-go/smarthome/core/database"
)

type AnalyzerExecutor struct {
	Username string
}

// Resolves a Homescript module
func (self *AnalyzerExecutor) ResolveModule(id string) (string, bool, bool, error) {
	script, found, err := database.GetUserHomescriptById(id, self.Username)
	if !found || err != nil {
		return "", found, true, err
	}
	return script.Data.Code, true, true, nil
}

func (self *AnalyzerExecutor) Sleep(seconds float64) {
}

func (self *AnalyzerExecutor) Print(args ...string) error {
	return nil
}

func (self *AnalyzerExecutor) Println(args ...string) error {
	return nil
}

func (self *AnalyzerExecutor) GetSwitch(switchId string) (homescript.SwitchResponse, error) {
	switchData, found, err := database.GetSwitchById(switchId)
	if !found {
		return homescript.SwitchResponse{}, fmt.Errorf("switch '%s' was not found", switchId)
	}
	if err != nil {
		return homescript.SwitchResponse{}, err
	}
	return homescript.SwitchResponse{
		Name:  switchData.Name,
		Power: switchData.PowerOn,
		Watts: uint(switchData.Watts),
	}, nil
}

func (self *AnalyzerExecutor) Switch(switchId string, powerOn bool) error {
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

func (self *AnalyzerExecutor) Ping(ip string, timeoutSecs float64) (bool, error) {
	_, err := ping.NewPinger(ip)
	if err != nil {
		return false, err
	}
	return false, nil
}

func (self *AnalyzerExecutor) Get(requestUrl string) (homescript.HttpResponse, error) {
	// The permissions can be validated beforehand
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionHomescriptNetwork)
	if err != nil {
		return homescript.HttpResponse{}, fmt.Errorf("could not send GET request: failed to validate your permissions: %s", err.Error())
	}
	if !hasPermission {
		return homescript.HttpResponse{}, fmt.Errorf("will not send GET request: you lack permission to access the network via homescript. If this is unintentional, contact your administrator")
	}
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

func (self *AnalyzerExecutor) Http(requestUrl string, method string, body string, headers map[string]string) (homescript.HttpResponse, error) {
	// Check permissions and request building beforehand
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionHomescriptNetwork)
	if err != nil {
		return homescript.HttpResponse{}, fmt.Errorf("could not perform %s request: failed to validate your permissions: %s", method, err.Error())
	}
	if !hasPermission {
		return homescript.HttpResponse{}, fmt.Errorf("will not perform %s request: you lack permission to access the network via Homescript. If this is unintentional, contact your administrator", method)
	}
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

func (self *AnalyzerExecutor) Notify(
	title string,
	description string,
	level homescript.NotificationLevel,
) error {
	return nil
}

func (self *AnalyzerExecutor) Remind(
	title string,
	description string,
	urgency homescript.ReminderUrgency,
	dueDate time.Time,
) (uint, error) {
	return 0, nil
}

func (self *AnalyzerExecutor) Log(
	title string,
	description string,
	level homescript.LogLevel,
) error {
	hasPermission, err := database.UserHasPermission(self.Username, database.PermissionLogging)
	if err != nil {
		return err
	}
	if !hasPermission {
		return fmt.Errorf("failed to add log event: you lack permission to add records to the internal logging system.")
	}
	if level > 5 {
		return fmt.Errorf("failed to add log event: invalid logging level <%d>: valid logging levels are 1, 2, 3, 4, or 5", level)
	}
	return nil
}

// Executes another Homescript based on its Id
func (self AnalyzerExecutor) Exec(homescriptId string, args map[string]string) (homescript.ExecResponse, error) {
	_, found, err := database.GetUserHomescriptById(homescriptId, self.Username)
	if err != nil {
		return homescript.ExecResponse{}, err
	}
	if !found {
		return homescript.ExecResponse{}, fmt.Errorf("invalid script: homescript '%s' was not found", homescriptId)
	}
	return homescript.ExecResponse{ReturnValue: homescript.ValueNull{}}, nil
}

// Returns the name of the user who is currently running the script
func (self *AnalyzerExecutor) GetUser() string {
	return self.Username
}

func (self *AnalyzerExecutor) GetWeather() (homescript.Weather, error) {
	return homescript.Weather{}, nil
}

func (self *AnalyzerExecutor) GetStorage(key string) (*string, error) {
	if utf8.RuneCountInString(key) > 50 {
		return nil, errors.New("key is larger than 50 characters")
	}
	return nil, nil
}

func (self *AnalyzerExecutor) SetStorage(key string, value string) error {
	if utf8.RuneCountInString(key) > 50 {
		return errors.New("key is larger than 50 characters")
	}
	return nil
}
