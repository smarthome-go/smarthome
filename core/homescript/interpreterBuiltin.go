package homescript

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/interpreter/value"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/hardware"
)

type interpreterExecutor struct {
	username string
	ioWriter io.Writer
}

func (self interpreterExecutor) GetUser() string {
	return self.username
}

func newInterpreterExecutor(username string, writer io.Writer) interpreterExecutor {
	return interpreterExecutor{
		username: username,
		ioWriter: writer,
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
	case "switch":
		switch toImport {
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

				valTemp := ""
				if val != nil {
					valTemp = *val
				}

				fields := map[string]*value.Value{
					"value": value.NewValueString(valTemp),
					"found": value.NewValueBool(val != nil),
				}

				return value.NewValueObject(fields), nil
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
	}
	return nil, false
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
