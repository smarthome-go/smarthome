package homescript

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/interpreter/value"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
	"github.com/smarthome-go/smarthome/core/homescript/automation"
	"github.com/smarthome-go/smarthome/services/weather"
)

type interpreterExecutor struct {
	username          string
	ioWriter          io.Writer
	args              map[string]string
	automationContext *AutomationContext
	cancelCtxFunc     context.CancelFunc
}

func (self interpreterExecutor) GetUser() string {
	return self.username
}

func newInterpreterExecutor(
	username string,
	writer io.Writer,
	args map[string]string,
	automationContext *AutomationContext,
	cancelCtxFunc context.CancelFunc,
) interpreterExecutor {
	return interpreterExecutor{
		username:          username,
		ioWriter:          writer,
		args:              args,
		automationContext: automationContext,
		cancelCtxFunc:     cancelCtxFunc,
	}
}

func parseDate(year, month, day int) (time.Time, bool) {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	y, m, d := t.Date()
	return t, y == year && int(m) == month && d == day
}

// if it exists, returns a value which is part of the host builtin modules
func (self interpreterExecutor) GetBuiltinImport(moduleName string, toImport string) (val value.Value, found bool) {
	switch moduleName {
	case "hms":
		switch toImport {
		case "exec":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				hmsId := args[0].(value.ValueString).Inner
				argOpt := args[1].(value.ValueOption)

				arguments := make(map[string]string)

				if argOpt.IsSome() {
					argFields := (*argOpt.Inner).(value.ValueAnyObject).FieldsInternal
					for key, value := range argFields {
						disp, i := (*value).Display()
						if i != nil {
							return nil, i
						}
						arguments[key] = disp
					}
				}

				res, err := HmsManager.RunById(
					hmsId,
					self.username,
					InitiatorExec,
					*cancelCtx,
					self.cancelCtxFunc,
					nil,
					arguments,
					self.ioWriter,
					nil,
				)

				if err != nil {
					return nil, value.NewRuntimeErr(
						err.Error(),
						value.HostErrorKind,
						span,
					)
				}

				if !res.Success {
					message := ""

					for _, err := range res.Errors {
						if err.SyntaxError != nil {
							message = err.SyntaxError.Message
							break
						}
						if err.DiagnosticError != nil && err.DiagnosticError.Level == diagnostic.DiagnosticLevelError {
							message = err.DiagnosticError.Message
							break
						}
						if err.RuntimeInterrupt != nil {
							message = err.RuntimeInterrupt.Message
							break
						}
					}

					return nil, value.NewThrowInterrupt(
						span,
						message,
					)
				}

				return value.NewValueNull(), nil
			}), true
		}
		return nil, false
	case "location":
		switch toImport {
		case "sun_times":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {

				serverConfig, found, err := database.GetServerConfiguration()
				if err != nil || !found {
					return nil, value.NewRuntimeErr(
						"Could not retrieve system configuration",
						value.HostErrorKind,
						span,
					)
				}

				rise, set := automation.CalculateSunRiseSet(serverConfig.Latitude, serverConfig.Longitude)

				return value.NewValueObject(map[string]*value.Value{
					"sunrise": value.NewValueObject(map[string]*value.Value{
						"hour":   value.NewValueInt(int64(rise.Hour)),
						"minute": value.NewValueInt(int64(rise.Minute)),
					}),
					"sunset": value.NewValueObject(map[string]*value.Value{
						"hour":   value.NewValueInt(int64(set.Hour)),
						"minute": value.NewValueInt(int64(set.Minute)),
					}),
				}), nil
			}), true
		case "weather":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				weather, err := weather.GetCurrentWeather()
				if err != nil {
					return nil, value.NewRuntimeErr(
						fmt.Sprintf("Could not load weather: %s", err.Error()),
						value.HostErrorKind,
						span,
					)
				}

				return value.NewValueObject(map[string]*value.Value{
					"title":       value.NewValueString(weather.WeatherTitle),
					"description": value.NewValueString(weather.WeatherDescription),
					"temperature": value.NewValueFloat(float64(weather.Temperature)),
					"feels_like":  value.NewValueFloat(float64(weather.FeelsLike)),
					"humidity":    value.NewValueInt(int64(weather.Humidity)),
				}), nil
			}), true
		}
	case "switch":
		switch toImport {
		case "get_switch":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				sw, found, err := database.GetSwitchById(args[0].(value.ValueString).Inner)
				if err != nil {
					return nil, value.NewRuntimeErr(
						err.Error(),
						value.HostErrorKind,
						span,
					)
				}

				if !found {
					return value.NewNoneOption(), nil
				}

				targetNode := value.NewNoneOption()

				if sw.TargetNode != nil {
					targetNode = value.NewValueOption(value.NewValueString(*sw.TargetNode))
				}

				return value.NewValueOption(value.NewValueObject(map[string]*value.Value{
					"name":        value.NewValueString(sw.Name),
					"room_id":     value.NewValueString(sw.RoomId),
					"power":       value.NewValueBool(sw.PowerOn),
					"watts":       value.NewValueInt(int64(sw.Watts)),
					"target_node": targetNode,
				})), nil
			}), true
		case "power":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				switchId := args[0].(value.ValueString).Inner
				powerOn := args[1].(value.ValueBool).Inner

				err := hardware.SetSwitchPowerAll(switchId, powerOn, self.username)
				if err != nil {
					return nil, value.NewRuntimeErr(err.Error(), value.HostErrorKind, span)
				}

				return value.NewValueNull(), nil
			}), true
		default:
			return nil, true
		}
	case "widget":
		switch toImport {
		case "on_click_js":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				targetCode := strings.ReplaceAll(args[0].(value.ValueString).Inner, "\"", "\\\"")
				inner := args[1].(value.ValueString).Inner
				wrapper := fmt.Sprintf("<span onclick=\"%s\">%s</span>", targetCode, inner)
				return value.NewValueString(wrapper), nil
			}), true
		case "on_click_hms":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				targetCode := strings.ReplaceAll(args[0].(value.ValueString).Inner, "'", "\\'")
				targetCode = strings.ReplaceAll(targetCode, "\"", "\\\"")
				inner := args[1].(value.ValueString).Inner
				callBackCode := fmt.Sprintf("fetch('/api/homescript/run/live', {method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code: `%s`, args: [] }) })", targetCode)
				wrapper := fmt.Sprintf("<span onclick=\"%s\">%s</span>", callBackCode, inner)
				return value.NewValueString(wrapper), nil
			}), true
		default:
			return nil, false
		}
	case "testing":
		switch toImport {
		case "assert_eq":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				lhsDisp, i := args[0].Display()
				if i != nil {
					return nil, i
				}
				rhsDisp, i := args[1].Display()
				if i != nil {
					return nil, i
				}

				if args[0].Kind() != args[1].Kind() {
					return nil, value.NewThrowInterrupt(span, fmt.Sprintf("`%s` is not equal to `%s`", lhsDisp, rhsDisp))
				}

				isEqual, i := args[0].IsEqual(args[1])
				if i != nil {
					return nil, i
				}

				if !isEqual {
					return nil, value.NewThrowInterrupt(span, fmt.Sprintf("`%s` is not equal to `%s`", lhsDisp, rhsDisp))
				}

				return value.NewValueNull(), nil
			}), true
		}
	case "storage":
		switch toImport {
		case "set_storage":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				key := args[0].(value.ValueString).Inner
				disp, i := args[1].Display()
				if i != nil {
					return nil, i
				}

				if err := database.InsertHmsStorageEntry(executor.GetUser(), key, disp); err != nil {
					return nil, value.NewRuntimeErr(
						fmt.Sprintf("Could not set storage: %s", err.Error()),
						value.HostErrorKind,
						span,
					)
				}

				return value.NewValueNull(), nil
			}), true
		case "get_storage":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				key := args[0].(value.ValueString).Inner

				val, err := database.GetHmsStorageEntry(executor.GetUser(), key)
				if err != nil {
					return nil, value.NewRuntimeErr(
						fmt.Sprintf("Could not set storage: %s", err.Error()),
						value.HostErrorKind,
						span,
					)
				}

				if val != nil {
					return value.NewValueOption(value.NewValueString(*val)), nil
				}

				return value.NewNoneOption(), nil
			}), true
		}
	case "reminder":
		switch toImport {
		case "remind":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				fields := args[0].(value.ValueObject).FieldsInternal

				title := (*fields["title"]).(value.ValueString).Inner
				description := (*fields["title"]).(value.ValueString).Inner
				priority := (*fields["priority"]).(value.ValueInt).Inner
				dueDateDay := (*fields["due_date_day"]).(value.ValueInt).Inner
				dueDateMonth := (*fields["due_date_month"]).(value.ValueInt).Inner
				dueDateYear := (*fields["due_date_year"]).(value.ValueInt).Inner

				if priority < 1.0 || priority > 5.0 {
					return nil, value.NewRuntimeErr(
						fmt.Sprintf("Reminder urgency has to be 0 < and < 6, got %d", int(priority)),
						value.ValueErrorKind,
						span,
					)
				}

				dueDate, valid := parseDate(int(dueDateYear), int(dueDateMonth), int(dueDateDay))
				if !valid {
					return nil, value.NewRuntimeErr(
						"Invalid due date provided",
						value.ValueErrorKind,
						span,
					)
				}

				newId, err := database.CreateNewReminder(
					title,
					description,
					dueDate,
					executor.GetUser(),
					database.ReminderPriority(priority),
				)
				if err != nil {
					return nil, value.NewRuntimeErr(
						fmt.Sprintf("Could not create reminder: %s", err.Error()),
						value.HostErrorKind,
						span,
					)
				}

				return value.NewValueInt(int64(newId)), nil
			}), true
		}
	case "net":
		switch toImport {
		case "ping":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				ip := args[0].(value.ValueString).Inner
				timeout := args[1].(value.ValueFloat).Inner

				pinger, err := ping.NewPinger(ip)
				if err != nil {
					return nil, value.NewThrowInterrupt(
						span,
						err.Error(),
					)
				}

				// perform the ping
				pinger.Count = 1
				pinger.Timeout = time.Millisecond * time.Duration(timeout*1000)
				err = pinger.Run() // blocks until the ping is finished or timed-out
				if err != nil {
					return nil, value.NewThrowInterrupt(
						span,
						err.Error(),
					)
				}
				stats := pinger.Statistics()
				return value.NewValueBool(stats.PacketsRecv > 0), nil // If at least 1 packet was received back, the host is considered online
			}), true
		case "http":
			return *value.NewValueObject(map[string]*value.Value{
				"get": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
					hasPermission, err := database.UserHasPermission(self.username, database.PermissionHomescriptNetwork)
					if err != nil {
						return nil, value.NewRuntimeErr(
							fmt.Sprintf("Could not send GET request: failed to validate user's permissions: %s", err.Error()),
							value.HostErrorKind,
							span,
						)
					}
					if !hasPermission {
						return nil, value.NewRuntimeErr(
							fmt.Sprintf("will not send GET request: you lack permission to access the network via homescript. If this is unintentional, contact your administrator"),
							value.HostErrorKind,
							span,
						)
					}

					url := args[0].(value.ValueString).Inner

					// Create a new request
					req, err := http.NewRequest(
						http.MethodGet,
						url,
						nil,
					)
					if err != nil {
						return nil, value.NewRuntimeErr(err.Error(), value.HostErrorKind, span)
					}
					// Set the user agent to the Smarthome HMS client
					req.Header.Set("User-Agent", "Smarthome-Homescript")

					// Create a new context for cancellatioon
					req = req.WithContext(*cancelCtx)

					// Perform the request
					// Create a client for the request
					client := http.Client{
						// Set the client's timeout to 60 seconds
						Timeout: 60 * time.Second,
					}

					res, err := client.Do(req)

					// Evaluate the request's outcome
					if err != nil {
						return nil, value.NewRuntimeErr(err.Error(), value.HostErrorKind, span)
					}

					// Read request response body
					defer res.Body.Close()
					resBody, err := io.ReadAll(res.Body)
					if err != nil {
						return nil, value.NewRuntimeErr(err.Error(), value.HostErrorKind, span)
					}

					outCookies := make(map[string]*value.Value)
					for _, cookie := range res.Cookies() {
						outCookies[cookie.Name] = value.NewValueString(cookie.Value)
					}

					return value.NewValueObject(map[string]*value.Value{
						"status":      value.NewValueString(res.Status),
						"status_code": value.NewValueInt(int64(res.StatusCode)),
						"body":        value.NewValueString(string(resBody)),
						"cookies":     value.NewValueAnyObject(outCookies),
					}), nil
				}),
				"generic": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
					// Check permissions and request building beforehand
					hasPermission, err := database.UserHasPermission(executor.GetUser(), database.PermissionHomescriptNetwork)
					if err != nil {
						return nil, value.NewRuntimeErr(
							fmt.Sprintf("Could not perform request: failed to validate your permissions: %s", err.Error()),
							value.HostErrorKind,
							span,
						)
					}
					if !hasPermission {
						return nil, value.NewRuntimeErr(
							fmt.Sprintf("Will not perform request: lacking permission to access the network via Homescript. If this is unintentional, contact your administrator"),
							value.HostErrorKind,
							span,
						)
					}

					url := args[0].(value.ValueString).Inner
					method := args[1].(value.ValueString).Inner
					body := args[2].(value.ValueOption)
					headers := args[3].(value.ValueAnyObject).FieldsInternal
					cookies := args[4].(value.ValueAnyObject).FieldsInternal

					var bodyStr string
					if body.IsSome() {
						bodyStr = (*body.Inner).(value.ValueString).Inner
					}

					// Create a new request
					req, err := http.NewRequest(
						method,
						url,
						strings.NewReader(bodyStr),
					)
					if err != nil {
						return nil, value.NewThrowInterrupt(
							span,
							err.Error(),
						)
					}
					// Set the user agent to the Smarthome HMS client
					req.Header.Set("User-Agent", "Smarthome-homescript")

					// Set the headers included via the function call
					for headerKey, headerValue := range headers {
						disp, i := (*headerValue).Display()
						if i != nil {
							return nil, i
						}

						req.Header.Set(headerKey, disp)
					}

					// Set the cookies
					for cookieKey, cookieValue := range cookies {
						disp, i := (*cookieValue).Display()
						if i != nil {
							return nil, i
						}

						c := http.Cookie{
							Name:  cookieKey,
							Value: disp,
						}
						req.AddCookie(&c)
					}

					req = req.WithContext(*cancelCtx)

					// Perform the request
					// Create a client for the request
					client := http.Client{
						// Set the client's timeout to 60 seconds
						Timeout: 60 * time.Second,
					}
					res, err := client.Do(req)
					// Evaluate the request's outcome
					if err != nil {
						return nil, value.NewThrowInterrupt(
							span,
							err.Error(),
						)
					}

					// Read request response body
					defer res.Body.Close()
					resBody, err := io.ReadAll(res.Body)
					if err != nil {
						return nil, value.NewThrowInterrupt(
							span,
							err.Error(),
						)
					}

					outCookies := make(map[string]*value.Value)
					for _, cookie := range res.Cookies() {
						outCookies[cookie.Name] = value.NewValueString(cookie.Value)
					}

					return value.NewValueObject(map[string]*value.Value{
						"status":      value.NewValueString(res.Status),
						"status_code": value.NewValueInt(int64(res.StatusCode)),
						"body":        value.NewValueString(string(resBody)),
						"cookies":     value.NewValueAnyObject(outCookies),
					}), nil
				}),
			}), true
		default:
			return nil, false
		}
	case "log":
		switch toImport {
		case "logger":
			testPermissions := func(username string, span errors.Span) *value.Interrupt {
				hasPermission, err := database.UserHasPermission(self.GetUser(), database.PermissionLogging)
				if err != nil {
					return value.NewRuntimeErr(err.Error(), value.HostErrorKind, span)
				}
				if !hasPermission {
					return value.NewRuntimeErr(fmt.Sprintf("Failed to add log event: lacking permission to add records to the internal logging system."), value.HostErrorKind, span)
				}
				return nil
			}

			return *value.NewValueObject(map[string]*value.Value{
				"trace": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(executor.GetUser(), span); i != nil {
						return nil, i
					}
					event.Trace(title, description)
					return value.NewValueNull(), nil
				}),
				"debug": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(executor.GetUser(), span); i != nil {
						return nil, i
					}
					event.Debug(title, description)
					return value.NewValueNull(), nil
				}),
				"info": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(executor.GetUser(), span); i != nil {
						return nil, i
					}
					event.Info(title, description)
					return value.NewValueNull(), nil
				}),
				"warn": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(executor.GetUser(), span); i != nil {
						return nil, i
					}
					event.Warn(title, description)
					return value.NewValueNull(), nil
				}),
				"error": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(executor.GetUser(), span); i != nil {
						return nil, i
					}
					event.Error(title, description)
					return value.NewValueNull(), nil
				}),
				"fatal": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(executor.GetUser(), span); i != nil {
						return nil, i
					}
					event.Fatal(title, description)
					return value.NewValueNull(), nil
				}),
			}), true
		}
	case "context":
		switch toImport {
		case "args":
			return *value.NewValueAnyObject(self.getArgs()), true
		case "notification":
			if self.automationContext == nil || self.automationContext.NotificationContext == nil {
				return nil, false
			}

			return *value.NewValueObject(map[string]*value.Value{
				"id":          value.NewValueInt(int64(self.automationContext.NotificationContext.Id)),
				"title":       value.NewValueString(self.automationContext.NotificationContext.Title),
				"description": value.NewValueString(self.automationContext.NotificationContext.Description),
				"level":       value.NewValueInt(int64(self.automationContext.NotificationContext.Level)),
			}), true
		}
	case "scheduler":
		switch toImport {
		case "create_schedule":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				data := args[0].(value.ValueObject).FieldsInternal

				hour := (*data["hour"]).(value.ValueInt).Inner
				minute := (*data["minute"]).(value.ValueInt).Inner

				if hour < 0 || minute < 0 {
					return nil, value.NewThrowInterrupt(span, "Fields `hour` and `minute` have to be >= 0")
				}

				newId, err := CreateNewSchedule(database.ScheduleData{
					Name:           (*data["name"]).(value.ValueString).Inner,
					Hour:           uint(hour),
					Minute:         uint(minute),
					TargetMode:     database.ScheduleTargetModeCode,
					HomescriptCode: (*data["code"]).(value.ValueString).Inner,
				}, executor.GetUser())

				if err != nil {
					return nil, value.NewRuntimeErr(fmt.Sprintf("Backend error: %s", err.Error()), value.HostErrorKind, span)
				}

				return value.NewValueInt(int64(newId)), nil
			}), true
		case "delete_schedule":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				id := args[0].(value.ValueInt).Inner

				if id < 0 {
					return nil, value.NewThrowInterrupt(span, fmt.Sprintf("IDs must be > 0, got %d", id))
				}

				_, found, err := GetUserScheduleById(executor.GetUser(), uint(id))
				if err != nil {
					return nil, value.NewRuntimeErr(
						fmt.Sprintf("Could not delete schedule: %s", err.Error()),
						value.HostErrorKind,
						span,
					)
				}

				if !found {
					return nil, value.NewThrowInterrupt(span, fmt.Sprintf("No schedule with ID %d exists", id))
				}

				if err := RemoveScheduleById(uint(id)); err != nil {
					return nil, value.NewRuntimeErr(
						fmt.Sprintf("Could not delete schedule: %s", err.Error()),
						value.HostErrorKind,
						span,
					)
				}

				return nil, nil
			}), true
		case "list_schedules":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				schedules, err := database.GetUserSchedules(executor.GetUser())
				if err != nil {
					return nil, value.NewRuntimeErr(
						fmt.Sprintf("Could not list schedules: %s", err.Error()),
						value.HostErrorKind,
						span,
					)
				}

				list := make([]*value.Value, 0)

				for _, sched := range schedules {

					hmsId := value.NewNoneOption()
					switches := value.NewNoneOption()

					switch sched.Data.TargetMode {
					case database.ScheduleTargetModeSwitches:
						innerValues := make([]*value.Value, 0)

						for _, job := range sched.Data.SwitchJobs {
							innerValues = append(innerValues, value.NewValueObject(map[string]*value.Value{
								"switch": value.NewValueString(job.SwitchId),
								"power":  value.NewValueBool(job.PowerOn),
							}))
						}

						switches = value.NewValueOption(value.NewValueList(innerValues))
					case database.ScheduleTargetModeHMS:
						hmsId = value.NewValueOption(value.NewValueString(sched.Data.HomescriptTargetId))
					}

					schedule := value.NewValueObject(map[string]*value.Value{
						"id":          value.NewValueInt(int64(sched.Id)),
						"name":        value.NewValueString(sched.Data.Name),
						"hour":        value.NewValueInt(int64(sched.Data.Hour)),
						"minute":      value.NewValueInt(int64(sched.Data.Minute)),
						"target_mode": value.NewValueString(string(sched.Data.TargetMode)),
						"hms_id":      hmsId,
						"switches":    switches,
					})
					list = append(list, schedule)
				}

				return value.NewValueList(list), nil
			}), true

			// ast.NewObjectTypeField(pAst.NewSpannedIdent("id", span), ast.NewIntType(span), span),
			// ast.NewObjectTypeField(pAst.NewSpannedIdent("name", span), ast.NewStringType(span), span),
			// ast.NewObjectTypeField(pAst.NewSpannedIdent("hour", span), ast.NewIntType(span), span),
			// ast.NewObjectTypeField(pAst.NewSpannedIdent("minute", span), ast.NewIntType(span), span),
			// ast.NewObjectTypeField(pAst.NewSpannedIdent("target_mode", span), ast.NewStringType(span), span),
		}
	case "notification":
		switch toImport {
		case "notify":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				title := args[0].(value.ValueString).Inner
				description := args[1].(value.ValueString).Inner
				level := args[2].(value.ValueInt).Inner

				hmsExecutor := executor.(interpreterExecutor)

				// only run notification hooks if this homescript was NOT triggered due to a notification
				// this avoids unconditional recursion and thus prevents a crash
				runHooks := hmsExecutor.automationContext == nil || hmsExecutor.automationContext.NotificationContext == nil

				newId, err := Notify(
					executor.GetUser(),
					title,
					description,
					NotificationLevel(level),
					runHooks,
				)

				if err != nil {
					return nil, value.NewRuntimeErr(
						fmt.Sprintf("Could not add notification: %s", err.Error()),
						value.HostErrorKind,
						span,
					)
				}

				return value.NewValueInt(int64(newId)), nil
			}), true
		}
	}
	return nil, false
}

func (self interpreterExecutor) getArgs() map[string]*value.Value {
	result := make(map[string]*value.Value)

	for key, val := range self.args {
		result[key] = value.NewValueString(val)
	}

	return result
}

// returns the Homescript code of the requested module
func (self interpreterExecutor) ResolveModuleCode(moduleName string) (code string, found bool, err error) {
	return "", false, nil
}

// Writes the given string (produced by a print function for instance) to any arbitrary source
func (self interpreterExecutor) WriteStringTo(input string) error {
	self.ioWriter.Write([]byte(input)) // TODO: does this even work?
	return nil
}

func checkCancelation(ctx *context.Context, span errors.Span) *value.Interrupt {
	select {
	case <-(*ctx).Done():
		return value.NewTerminationInterrupt((*ctx).Err().Error(), span)
	default:
		// do nothing, this should not block the entire interpreter
		return nil
	}
}

func interpreterScopeAdditions() map[string]value.Value {
	// TODO: fill this
	return map[string]value.Value{
		"exit": *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
			return nil, value.NewExitInterrupt(args[0].(value.ValueInt).Inner)
		}),
		"fmt": *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
			displays := make([]any, 0)

			for idx, arg := range args {
				if idx == 0 {
					continue
				}

				var out any

				switch arg.Kind() {
				case value.NullValueKind:
					out = "null"
				case value.IntValueKind:
					out = arg.(value.ValueInt).Inner
				case value.FloatValueKind:
					out = arg.(value.ValueFloat).Inner
				case value.BoolValueKind:
					out = arg.(value.ValueBool).Inner
				case value.StringValueKind:
					out = arg.(value.ValueString).Inner
				default:
					display, i := arg.Display()
					if i != nil {
						return nil, i
					}
					out = display
				}

				displays = append(displays, out)
			}

			out := fmt.Sprintf(args[0].(value.ValueString).Inner, displays...)

			return value.NewValueString(out), nil
		}),
		"println": *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
			output := make([]string, 0)
			for _, arg := range args {
				disp, i := arg.Display()
				if i != nil {
					return nil, i
				}
				output = append(output, disp)
			}

			outStr := strings.Join(output, " ") + "\n"

			if err := executor.WriteStringTo(outStr); err != nil {
				return nil, value.NewRuntimeErr(
					err.Error(),
					value.HostErrorKind,
					span,
				)
			}

			return value.NewValueNull(), nil
		}),
		"time": *value.NewValueObject(map[string]*value.Value{
			"sleep": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				durationSecs := args[0].(value.ValueFloat).Inner

				for i := 0; i < int(durationSecs*1000); i += 10 {
					if i := checkCancelation(cancelCtx, span); i != nil {
						return nil, i
					}
					time.Sleep(time.Millisecond * 10)
				}

				return nil, nil
			}),
			"since": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				milliObj := args[0].(value.ValueObject).FieldsInternal["unix_milli"]
				then := time.UnixMilli((*milliObj).(value.ValueInt).Inner)
				return createDurationObject(time.Since(then)), nil
			}),
			"add_days": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				base := createTimeStructFromObject(args[0])
				days := args[1].(value.ValueInt).Inner
				return createTimeObject(base.Add(time.Hour * 24 * time.Duration(days))), nil
			}),
			"add_hours": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				base := createTimeStructFromObject(args[0])
				hours := args[1].(value.ValueInt).Inner
				return createTimeObject(base.Add(time.Hour * time.Duration(hours))), nil
			}),
			"add_minutes": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				base := createTimeStructFromObject(args[0])
				minutes := args[1].(value.ValueInt).Inner
				return createTimeObject(base.Add(time.Minute * time.Duration(minutes))), nil
			}),
			"now": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
				now := time.Now()

				return createTimeObject(now), nil
			}),
		}),
	}
}

func createDurationObject(t time.Duration) *value.Value {
	return value.NewValueObject(map[string]*value.Value{
		"hours":   value.NewValueFloat(t.Hours()),
		"minutes": value.NewValueFloat(t.Minutes()),
		"seconds": value.NewValueFloat(t.Seconds()),
		"millis":  value.NewValueInt(t.Milliseconds()),
		"display": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.Interrupt) {
			return value.NewValueString(t.String()), nil
		}),
	})
}

func createTimeObject(t time.Time) *value.Value {
	return value.NewValueObject(
		map[string]*value.Value{
			"year":          value.NewValueInt(int64(t.Year())),
			"month":         value.NewValueInt(int64(t.Month())),
			"year_day":      value.NewValueInt(int64(t.YearDay())),
			"hour":          value.NewValueInt(int64(t.Hour())),
			"minute":        value.NewValueInt(int64(t.Minute())),
			"second":        value.NewValueInt(int64(t.Second())),
			"month_day":     value.NewValueInt(int64(t.Day())),
			"week_day":      value.NewValueInt(int64(t.Weekday())),
			"week_day_text": value.NewValueString(t.Weekday().String()),
			"unix_milli":    value.NewValueInt(t.UnixMilli()),
		},
	)
}

func createTimeStructFromObject(t value.Value) time.Time {
	tObj := t.(value.ValueObject)
	fields, i := tObj.Fields()
	if i != nil {
		panic(i)
	}
	millis := (*fields["unix_milli"]).(value.ValueInt).Inner
	return time.UnixMilli(millis)
}
