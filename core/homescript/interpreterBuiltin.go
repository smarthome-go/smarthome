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
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/automation"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/device/driver"
	"github.com/smarthome-go/smarthome/core/event"
	"github.com/smarthome-go/smarthome/core/homescript/dispatcher"
	dispatcherT "github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
	"github.com/smarthome-go/smarthome/core/homescript/types"
	"github.com/smarthome-go/smarthome/core/scheduler"
	"github.com/smarthome-go/smarthome/core/user/notify"
	"github.com/smarthome-go/smarthome/services/weather"
)

type interpreterExecutor struct {
	// All `attaching` registrations in the dispatcher (these need to be revoked before the VM is deleted).
	registrations *[]dispatcherT.RegistrationID
	// Other.
	jobID     uint64
	programID string

	// username          string
	ioWriter io.Writer
	// args              map[string]string
	// automationContext *types.AutomationContext
	// cancelCtxFunc     context.CancelFunc

	singletons map[string]value.Value

	context types.ExecutionContext

	cancelation types.Cancelation

	// Mangled names of the functions that are to be called when this program is killed.
	onKillCallbackFuncs *[]string
}

func (self interpreterExecutor) Free() error {
	var errRes error = nil

	for _, registration := range *self.registrations {
		// Return the first error that is found
		if err := dispatcher.Instance.Unregister(registration); err != nil && errRes == nil {
			errRes = err
		}
	}

	return errRes
}

func NewInterpreterExecutor(
	jobID uint64,
	programID string,
	// username string,
	writer io.Writer,
	// args map[string]string,
	// automationContext *types.AutomationContext,
	// cancelCtxFunc context.CancelFunc,
	cancelation types.Cancelation,
	singletons map[string]value.Value,
	context types.ExecutionContext,
) interpreterExecutor {
	registrations := make([]dispatcherT.RegistrationID, 0)
	onKillCallbackFuncs := make([]string, 0)

	return interpreterExecutor{
		registrations:       &registrations,
		jobID:               jobID,
		programID:           programID,
		ioWriter:            writer,
		singletons:          singletons,
		context:             context,
		cancelation:         cancelation,
		onKillCallbackFuncs: &onKillCallbackFuncs,
	}
}

func parseDate(year, month, day int) (time.Time, bool) {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	y, m, d := t.Date()
	return t, y == year && int(m) == month && d == day
}

func (self interpreterExecutor) LoadSingleton(singletonIdent, moduleName string) (val value.Value, valid bool, err error) {
	logger.Tracef("Loading singleton `%s` from module `%s`", singletonIdent, moduleName)
	value, available := self.singletons[singletonIdent]

	if !available {
		panic(fmt.Sprintf("Singleton `%s` could not be loaded from: %v", singletonIdent, self.singletons))
	}

	disp, e := value.Display()
	if e != nil {
		panic(e)
	}

	logger.Tracef("Successfully loaded singleton `%s` from module `%s`: %s", singletonIdent, moduleName, disp)

	// TODO: maybe load these on demand?
	return value, available, nil
}

// var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
// 	fmt.Printf("TOPIC: %s\n", msg.Topic())
// 	fmt.Printf("MSG: %s\n", msg.Payload())
// }
//
// func mqttSubscribe() value.Value {
// 	return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
// 		host := args[0].(value.ValueString).Inner
// 		username := args[1].(value.ValueString).Inner
// 		password := args[2].(value.ValueString).Inner
//
// 		topicArgsRaw := args[3].(value.ValueList).Values
// 		topicArgs := make(map[string]byte)
// 		for _, topic := range *topicArgsRaw {
// 			topicStr := (*topic).(value.ValueString).Inner
// 			topicArgs[topicStr] = 0
// 		}
// 		// callBackArg := args[1].(value.ValueVMFunction)
//
// 		spew.Dump(args)
// 		// return value.NewValueNull(), nil
//
// 		mqtt.DEBUG = log.New(os.Stdout, "", 0)
// 		mqtt.ERROR = log.New(os.Stdout, "", 0)
// 		opts := mqtt.NewClientOptions().AddBroker(host).SetClientID("homescript-test-shome").SetUsername(username).SetPassword(password)
//
// 		opts.SetKeepAlive(60 * time.Second)
// 		// Set the message callback handler
// 		opts.SetDefaultPublishHandler(f)
// 		opts.SetPingTimeout(1 * time.Second)
//
// 		c := mqtt.NewClient(opts)
// 		if token := c.Connect(); token.Wait() && token.Error() != nil {
// 			panic(token.Error())
// 		}
//
// 		var callBack mqtt.MessageHandler = func(c mqtt.Client, m mqtt.Message) {
// 			core := executor.(interpreterExecutor).vm.SpawnAsync(runtime.FunctionInvocation{
// 				Function:    "mqtt_recv",
// 				LiteralName: false,
// 				Args: []value.Value{
// 					*value.NewValueString(string(m.Topic())),
// 					*value.NewValueString(string(m.Payload())),
// 				},
// 				FunctionSignature: runtime.FunctionInvocationSignatureFromType(mqttCallbackFn(span).(ast.FunctionType)),
// 			}, nil)
//
// 			logger.Infof("Dispatched MQTT message to core %d", core.Corenum)
// 		}
//
// 		logger.Infof("Subscribed to MQTT topics: `%v`", topicArgs)
//
// 		// Subscribe to a topic
// 		if token := c.SubscribeMultiple(topicArgs, callBack); token.Wait() && token.Error() != nil {
// 			fmt.Println(token.Error())
// 			os.Exit(1)
// 		}
//
// 		// Publish a message
// 		// token := c.Publish("testtopic/1", 0, false, "Hello World")
// 		// token.Wait()
//
// 		// time.Sleep(6 * time.Second)
//
// 		// Unscribe
// 		// if token := c.Unsubscribe("testtopic/#"); token.Wait() && token.Error() != nil {
// 		// 	fmt.Println(token.Error())
// 		// 	os.Exit(1)
// 		// }
//
// 		// Disconnect
// 		// c.Disconnect(250)
// 		// time.Sleep(1 * time.Second)
//
// 		return value.NewValueNull(), nil
// 	})
// }

// func mqttPublish() value.Value {
// 	return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
// 		host := args[0].(value.ValueString).Inner
// 		username := args[1].(value.ValueString).Inner
// 		password := args[2].(value.ValueString).Inner
// 		topic := args[3].(value.ValueString).Inner
// 		payload := args[4].(value.ValueString).Inner
//
// 		// callBackArg := args[1].(value.ValueVMFunction)
//
// 		spew.Dump(args)
// 		// return value.NewValueNull(), nil
//
// 		mqtt.DEBUG = log.New(os.Stdout, "", 0)
// 		mqtt.ERROR = log.New(os.Stdout, "", 0)
// 		opts := mqtt.NewClientOptions().AddBroker(host).SetClientID("homescript-test-shome").SetUsername(username).SetPassword(password)
//
// 		opts.SetKeepAlive(60 * time.Second)
// 		// Set the message callback handler
// 		opts.SetDefaultPublishHandler(f)
// 		opts.SetPingTimeout(1 * time.Second)
//
// 		c := mqtt.NewClient(opts)
// 		if token := c.Connect(); token.Wait() && token.Error() != nil {
// 			panic(token.Error())
// 		}
//
// 		// Publish a message
// 		token := c.Publish(topic, 0, false, payload)
// 		token.Wait()
//
// 		// Disconnect
// 		c.Disconnect(250)
//
// 		return value.NewValueNull(), nil
// 	})
// }

func (self interpreterExecutor) execHelper(
	username,
	programID string,
	arguments map[string]string,
	span errors.Span,
) (*value.Value, *value.VmInterrupt) {
	res, err := HmsManager.RunUserScript(
		programID,
		username,
		nil,
		self.cancelation,
		self.ioWriter,
		nil,
	)

	if err != nil {
		return nil, value.NewVMThrowInterrupt(
			span,
			fmt.Sprintf("Failed to run program: `%s`", err.Error()),
		)
	}

	if !res.Errors.ContainsError {
		message := ""

		for _, err := range res.Errors.Diagnostics {
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

		return nil, value.NewVMThrowInterrupt(
			span,
			fmt.Sprintf("Invoked program crashed: `%s`", message),
		)
	}

	return value.NewValueNull(), nil
}

const execFnIdent = "exec"
const execUserFnIdent = "exec_user"

func (self interpreterExecutor) execBuiltin(usernameNeedsToBeSpecified bool) value.Value {
	return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
		var username *string
		// This will be 1 if the username is being read as the first argument.
		argumentIndexOffset := 0

		switch usernameNeedsToBeSpecified {
		case true:
			if self.context.Kind() == types.HMS_PROGRAM_KIND_USER {
				return nil, value.NewVMFatalException(
					fmt.Sprintf("The usage of the `%s` function in a user environment is not allowed", execUserFnIdent),
					value.Vm_HostErrorKind,
					span,
				)
			}

			usernameStr := args[0].(value.ValueString).Inner
			username = &usernameStr
			argumentIndexOffset = 1
		case false:
			if self.context.Username() == nil {
				return nil, value.NewVMFatalException(
					fmt.Sprintf("The usage of the `%s` function in a non-user environment is not possible", execFnIdent),
					value.Vm_HostErrorKind,
					span,
				)
			}
		}

		if username == nil {
			panic("Encountered internal bug: `username` cannot be <nil> at this point")
		}

		programID := args[argumentIndexOffset+0].(value.ValueString).Inner
		argOpt := args[argumentIndexOffset+1].(value.ValueOption)

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

		// TODO: remove this once argument support is implemented.
		if len(arguments) != 0 {
			return nil, value.NewVMFatalException(
				"BUG: Argument support is not implemented yet",
				value.Vm_HostErrorKind,
				span,
			)
		}

		return self.execHelper(
			*username,
			programID,
			arguments,
			span,
		)
	})
}

// if it exists, returns a value which is part of the host builtin modules
func (self interpreterExecutor) GetBuiltinImport(
	moduleName string,
	toImport string,
) (val value.Value, found bool) {
	switch moduleName {
	case "mqtt":
		switch toImport {
		case "subscribe":
			return mqttSubscribe(), true
		case "publish":
			return mqttPublish(), true
		default:
			return nil, false
		}
	case "hms":
		switch toImport {
		case execFnIdent:
			return self.execBuiltin(false), true
		case execUserFnIdent:
			return self.execBuiltin(true), true
		}
		return nil, false
	case "location":
		switch toImport {
		case "sun_times":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {

				serverConfig, found, err := database.GetServerConfiguration()
				if err != nil || !found {
					return nil, value.NewVMFatalException(
						"Could not retrieve system configuration",
						value.Vm_HostErrorKind,
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
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				weather, err := weather.GetCurrentWeather()
				if err != nil {
					return nil, value.NewVMThrowInterrupt(
						span,
						fmt.Sprintf("Could not load weather: %s", err.Error()),
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
	case "device":
		switch toImport {
		// case "get_switch":
		// 	return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
		// 		sw, found, err := database.GetDeviceById(args[0].(value.ValueString).Inner)
		// 		if err != nil {
		// 			return nil, value.NewVMFatalException(
		// 				err.Error(),
		// 				value.Vm_HostErrorKind,
		// 				span,
		// 			)
		// 		}
		//
		// 		if !found {
		// 			return value.NewNoneOption(), nil
		// 		}
		//
		// 		return value.NewValueOption(value.NewValueObject(map[string]*value.Value{
		// 			"name":      value.NewValueString(sw.Name),
		// 			"room_id":   value.NewValueString(sw.RoomId),
		// 			"vendor_id": value.NewValueString(sw.VendorId),
		// 			"model_id":  value.NewValueString(sw.VendorId),
		// 		})), nil
		// 	}), true
		case "set_power":
			return *value.NewValueBuiltinFunction(func(
				executor value.Executor,
				cancelCtx *context.Context,
				span errors.Span,
				args ...value.Value,
			) (*value.Value, *value.VmInterrupt) {
				deviceId := args[0].(value.ValueString).Inner
				powerOn := args[1].(value.ValueBool).Inner

				output, deviceFound, hmsErr, err := driver.Manager.SetDevicePower(deviceId, powerOn)
				if err != nil {
					return nil, value.NewVMFatalException(
						fmt.Sprintf("Backend failure during power action: %s", err.Error()),
						value.Vm_HostErrorKind,
						span,
					)
				}

				if hmsErr != nil {
					return nil, value.NewVMThrowInterrupt(
						span,
						fmt.Sprintf("Device malfunction: %s", hmsErr.String()),
					)
				}

				if !deviceFound {
					return nil, value.NewVMThrowInterrupt(
						span,
						fmt.Sprintf("No such device: `%s`", deviceId),
					)
				}

				return value.NewValueBool(output.Changed), nil
			}), true
		case "dim":
			return *value.NewValueBuiltinFunction(func(
				executor value.Executor,
				cancelCtx *context.Context,
				span errors.Span,
				args ...value.Value,
			) (*value.Value, *value.VmInterrupt) {
				deviceId := args[0].(value.ValueString).Inner
				function := args[1].(value.ValueString).Inner
				dimValue := args[2].(value.ValueInt).Inner

				output, deviceFound, hmsErr, err := driver.Manager.SetDeviceDim(deviceId, function, dimValue)
				if err != nil {
					return nil, value.NewVMFatalException(
						fmt.Sprintf("Backend failure during dim action: %s", err.Error()),
						value.Vm_HostErrorKind,
						span,
					)
				}

				if hmsErr != nil {
					return nil, value.NewVMThrowInterrupt(
						span,
						fmt.Sprintf("Device malfunction: %s", hmsErr.String()),
					)
				}

				if !deviceFound {
					return nil, value.NewVMThrowInterrupt(
						span,
						fmt.Sprintf("No such device: `%s`", deviceId),
					)
				}

				return value.NewValueBool(output.Changed), nil
			}), true
		default:
			return nil, true
		}
	case "widget":
		switch toImport {
		case "on_click_js": // TODO: remove these: it can be implemented better
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				targetCode := strings.ReplaceAll(args[0].(value.ValueString).Inner, "\"", "\\\"")
				inner := args[1].(value.ValueString).Inner
				wrapper := fmt.Sprintf("<span onclick=\"%s\">%s</span>", targetCode, inner)
				return value.NewValueString(wrapper), nil
			}), true
		case "on_click_hms":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
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
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				lhsDisp, i := args[0].Display()
				if i != nil {
					return nil, i
				}
				rhsDisp, i := args[1].Display()
				if i != nil {
					return nil, i
				}

				if args[0].Kind() != args[1].Kind() {
					return nil, value.NewVMThrowInterrupt(span, fmt.Sprintf("`%s` is not equal to `%s`", lhsDisp, rhsDisp))
				}

				isEqual, i := args[0].IsEqual(args[1])
				if i != nil {
					return nil, i
				}

				if !isEqual {
					return nil, value.NewVMThrowInterrupt(span, fmt.Sprintf("`%s` is not equal to `%s`", lhsDisp, rhsDisp))
				}

				return value.NewValueNull(), nil
			}), true
		}
	case "storage":
		switch toImport {
		case "set_storage":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				// TODO: use a macro here?
				if self.context.Username() == nil {
					return nil, value.NewVMFatalException(
						"The usage of the `storage` functions in a non-user environment is not possible",
						value.Vm_HostErrorKind,
						span,
					)
				}

				key := args[0].(value.ValueString).Inner
				disp, i := args[1].Display()
				if i != nil {
					return nil, i
				}

				if err := database.InsertHmsStorageEntry(*executor.(interpreterExecutor).context.Username(), key, disp); err != nil {
					return nil, value.NewVMFatalException(
						fmt.Sprintf("Could not set storage: %s", err.Error()),
						value.Vm_HostErrorKind,
						span,
					)
				}

				return value.NewValueNull(), nil
			}), true
		case "get_storage":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				// TODO: use a macro here?
				if self.context.Username() == nil {
					return nil, value.NewVMFatalException(
						"The usage of the `storage` functions in a non-user environment is not possible",
						value.Vm_HostErrorKind,
						span,
					)
				}

				key := args[0].(value.ValueString).Inner

				val, err := database.GetHmsStorageEntry(*executor.(interpreterExecutor).context.Username(), key)
				if err != nil {
					return nil, value.NewVMFatalException(
						fmt.Sprintf("Could not set storage: %s", err.Error()),
						value.Vm_HostErrorKind,
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
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				// TODO: use a macro here?
				if self.context.Username() == nil {
					return nil, value.NewVMFatalException(
						"The usage of the `remind` function in a non-user environment is not possible",
						value.Vm_HostErrorKind,
						span,
					)
				}

				fields := args[0].(value.ValueObject).FieldsInternal

				title := (*fields["title"]).(value.ValueString).Inner
				description := (*fields["title"]).(value.ValueString).Inner
				priority := (*fields["priority"]).(value.ValueInt).Inner
				dueDateDay := (*fields["due_date_day"]).(value.ValueInt).Inner
				dueDateMonth := (*fields["due_date_month"]).(value.ValueInt).Inner
				dueDateYear := (*fields["due_date_year"]).(value.ValueInt).Inner

				if priority < 0.0 || priority > 4.0 || float64(int64(priority)) != float64(priority) {
					return nil, value.NewVMThrowInterrupt(
						span,
						fmt.Sprintf("Reminder urgency has to an integer ( where 0 <= urgency <= 4 ), got %d", int(priority)),
					)
				}

				dueDate, valid := parseDate(int(dueDateYear), int(dueDateMonth), int(dueDateDay))
				if !valid {
					return nil, value.NewVMThrowInterrupt(
						span,
						"Invalid due date provided",
					)
				}

				newId, err := database.CreateNewReminder(
					title,
					description,
					dueDate,
					*executor.(interpreterExecutor).context.Username(),
					database.ReminderPriority(priority),
				)
				if err != nil {
					return nil, value.NewVMFatalException(
						fmt.Sprintf("Could not create reminder: %s", err.Error()),
						value.Vm_HostErrorKind,
						span,
					)
				}

				return value.NewValueInt(int64(newId)), nil
			}), true
		}
	case "net":
		switch toImport {
		case "ping":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				ip := args[0].(value.ValueString).Inner
				timeout := args[1].(value.ValueFloat).Inner

				pinger, err := ping.NewPinger(ip)
				if err != nil {
					return nil, value.NewVMFatalException(
						err.Error(),
						value.Vm_HostErrorKind,
						span,
					)
				}

				// perform the ping
				pinger.Count = 1
				pinger.Timeout = time.Millisecond * time.Duration(timeout*1000)
				err = pinger.Run() // blocks until the ping is finished or timed-out
				if err != nil {
					return nil, value.NewVMThrowInterrupt(
						span,
						err.Error(),
					)
				}
				stats := pinger.Statistics()
				return value.NewValueBool(stats.PacketsRecv > 0), nil // If at least 1 packet was received back, the host is considered online
			}), true
		case "http":
			return *value.NewValueObject(map[string]*value.Value{
				"get": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
					if self.context.Username() != nil {
						hasPermission, err := database.UserHasPermission(*self.context.Username(), database.PermissionHomescriptNetwork)
						if err != nil {
							return nil, value.NewVMFatalException(
								fmt.Sprintf("Could not send GET request: failed to validate user's permissions: %s", err.Error()),
								value.Vm_HostErrorKind,
								span,
							)
						}
						if !hasPermission {
							return nil, value.NewVMFatalException(
								fmt.Sprintf("will not send GET request: you lack permission to access the network via homescript. If this is unintentional, contact your administrator"),
								value.Vm_HostErrorKind,
								span,
							)
						}

					}

					url := args[0].(value.ValueString).Inner

					// Create a new request
					req, err := http.NewRequest(
						http.MethodGet,
						url,
						nil,
					)
					if err != nil {
						return nil, value.NewVMThrowInterrupt(span, err.Error())
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
						return nil, value.NewVMThrowInterrupt(span, err.Error())
					}

					// Read request response body
					defer res.Body.Close()
					resBody, err := io.ReadAll(res.Body)
					if err != nil {
						return nil, value.NewVMThrowInterrupt(span, err.Error())
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
				"generic": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
					if self.context.Username() != nil {
						hasPermission, err := database.UserHasPermission(*executor.(interpreterExecutor).context.Username(), database.PermissionHomescriptNetwork)
						if err != nil {
							return nil, value.NewVMFatalException(
								fmt.Sprintf("Could not perform request: failed to validate your permissions: %s", err.Error()),
								value.Vm_HostErrorKind,
								span,
							)
						}
						if !hasPermission {
							return nil, value.NewVMFatalException(
								fmt.Sprintf("Will not perform request: lacking permission to access the network via Homescript. If this is unintentional, contact your administrator"),
								value.Vm_HostErrorKind,
								span,
							)
						}
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
						return nil, value.NewVMThrowInterrupt(
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
						return nil, value.NewVMThrowInterrupt(
							span,
							err.Error(),
						)
					}

					// Read request response body
					defer res.Body.Close()
					resBody, err := io.ReadAll(res.Body)
					if err != nil {
						return nil, value.NewVMThrowInterrupt(
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
			testPermissions := func(username *string, span errors.Span) *value.VmInterrupt {
				return nil
			}

			if self.context.Username() != nil {
				testPermissions = func(username *string, span errors.Span) *value.VmInterrupt {
					hasPermission, err := database.UserHasPermission(*username, database.PermissionLogging)
					if err != nil {
						return value.NewVMFatalException(err.Error(), value.Vm_HostErrorKind, span)
					}
					if !hasPermission {
						return value.NewVMFatalException(fmt.Sprintf("Failed to add log event: lacking permission to add records to the internal logging system."), value.Vm_HostErrorKind, span)
					}
					return nil
				}
			}

			return *value.NewValueObject(map[string]*value.Value{
				"trace": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(executor.(interpreterExecutor).context.Username(), span); i != nil {
						return nil, i
					}
					event.Trace(title, description)
					return value.NewValueNull(), nil
				}),
				"debug": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(executor.(interpreterExecutor).context.Username(), span); i != nil {
						return nil, i
					}
					event.Debug(title, description)
					return value.NewValueNull(), nil
				}),
				"info": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(self.context.Username(), span); i != nil {
						return nil, i
					}
					event.Info(title, description)
					return value.NewValueNull(), nil
				}),
				"warn": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(self.context.Username(), span); i != nil {
						return nil, i
					}
					event.Warn(title, description)
					return value.NewValueNull(), nil
				}),
				"error": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(self.context.Username(), span); i != nil {
						return nil, i
					}
					event.Error(title, description)
					return value.NewValueNull(), nil
				}),
				"fatal": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
					title := args[0].(value.ValueString).Inner
					description := args[1].(value.ValueString).Inner
					if i := testPermissions(self.context.Username(), span); i != nil {
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
			// If this program was not triggered
			if self.context.Kind() != types.HMS_PROGRAM_KIND_AUTOMATION {
				panic(self.context.Kind())
				return *value.NewNoneOption(), true
			}

			if self.context.(types.ExecutionContextAutomation).Inner.NotificationContext == nil {
				panic("a")
			}

			automationContext := self.context.(types.ExecutionContextAutomation)

			return *value.NewValueOption(value.NewValueObject(map[string]*value.Value{
				"id":          value.NewValueInt(int64(automationContext.Inner.NotificationContext.Id)),
				"title":       value.NewValueString(automationContext.Inner.NotificationContext.Title),
				"description": value.NewValueString(automationContext.Inner.NotificationContext.Description),
				"level":       value.NewValueInt(int64(automationContext.Inner.NotificationContext.Level)),
			})), true

		}
	case "scheduler":
		switch toImport {
		case "create_schedule":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				if self.context.Username() == nil {
					return nil, value.NewVMFatalException(
						"The usage of the `scheduler` functions in a non-user environment is not possible",
						value.Vm_HostErrorKind,
						span,
					)
				}

				data := args[0].(value.ValueObject).FieldsInternal

				hour := (*data["hour"]).(value.ValueInt).Inner
				minute := (*data["minute"]).(value.ValueInt).Inner

				if hour < 0 || minute < 0 {
					return nil, value.NewVMThrowInterrupt(span, "Fields `hour` and `minute` have to be >= 0")
				}

				newId, err := scheduler.Manager.CreateNewSchedule(database.ScheduleData{
					Name:           (*data["name"]).(value.ValueString).Inner,
					Hour:           uint(hour),
					Minute:         uint(minute),
					TargetMode:     database.ScheduleTargetModeCode,
					HomescriptCode: (*data["code"]).(value.ValueString).Inner,
				}, *self.context.Username())

				if err != nil {
					return nil, value.NewVMFatalException(fmt.Sprintf("Backend error: %s", err.Error()), value.Vm_HostErrorKind, span)
				}

				return value.NewValueInt(int64(newId)), nil
			}), true
		case "delete_schedule":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				if self.context.Username() == nil {
					return nil, value.NewVMFatalException(
						"The usage of the `scheduler` functions in a non-user environment is not possible",
						value.Vm_HostErrorKind,
						span,
					)
				}

				id := args[0].(value.ValueInt).Inner

				if id < 0 {
					return nil, value.NewVMThrowInterrupt(span, fmt.Sprintf("IDs must be > 0, got %d", id))
				}

				_, found, err := scheduler.GetUserScheduleById(*self.context.Username(), uint(id))
				if err != nil {
					return nil, value.NewVMFatalException(
						fmt.Sprintf("Could not delete schedule: %s", err.Error()),
						value.Vm_HostErrorKind,
						span,
					)
				}

				if !found {
					return nil, value.NewVMThrowInterrupt(span, fmt.Sprintf("No schedule with ID %d exists", id))
				}

				if err := scheduler.Manager.RemoveScheduleById(uint(id)); err != nil {
					return nil, value.NewVMFatalException(
						fmt.Sprintf("Could not delete schedule: %s", err.Error()),
						value.Vm_HostErrorKind,
						span,
					)
				}

				return nil, nil
			}), true
		case "list_schedules":
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				if self.context.Username() == nil {
					return nil, value.NewVMFatalException(
						"The usage of the `scheulder` functions in a non-user environment is not possible",
						value.Vm_HostErrorKind,
						span,
					)
				}

				schedules, err := database.GetUserSchedules(*self.context.Username())
				if err != nil {
					return nil, value.NewVMFatalException(
						fmt.Sprintf("Could not list schedules: %s", err.Error()),
						value.Vm_HostErrorKind,
						span,
					)
				}

				list := make([]*value.Value, 0)

				for _, sched := range schedules {

					hmsId := value.NewNoneOption()
					switches := value.NewNoneOption()

					switch sched.Data.TargetMode {
					case database.ScheduleTargetModeDevices:
						innerValues := make([]*value.Value, 0)

						for _, job := range sched.Data.SwitchJobs {
							innerValues = append(innerValues, value.NewValueObject(map[string]*value.Value{
								"switch": value.NewValueString(job.DeviceId),
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
			return *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				if self.context.Username() == nil {
					return nil, value.NewVMFatalException(
						"The usage of this function in a non-user environment is not possible",
						value.Vm_HostErrorKind,
						span,
					)
				}

				title := args[0].(value.ValueString).Inner
				description := args[1].(value.ValueString).Inner
				level := args[2].(value.ValueInt).Inner

				hmsExecutor := executor.(interpreterExecutor)

				// Only run notification hooks if this homescript was NOT triggered due to a notification
				// this avoids unconditional recursion and thus prevents a crash.
				runHooks := hmsExecutor.context.Kind() != types.HMS_PROGRAM_KIND_AUTOMATION ||
					hmsExecutor.context.(types.ExecutionContextAutomation).Inner.NotificationContext == nil

				newId, err := notify.Manager.Notify(
					*self.context.Username(),
					title,
					description,
					notify.NotificationLevel(level),
					runHooks,
				)

				if err != nil {
					return nil, value.NewVMFatalException(
						fmt.Sprintf("Could not add notification: %s", err.Error()),
						value.Vm_HostErrorKind,
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
	if self.context.UserArgs() == nil {
		panic("Cannot access arguments in a non-user context")
	}

	result := make(map[string]*value.Value)

	for key, val := range self.context.UserArgs() {
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
	self.ioWriter.Write([]byte(input))
	return nil
}

func checkCancelation(ctx *context.Context, span errors.Span) *value.VmInterrupt {
	select {
	case <-(*ctx).Done():
		return value.NewVMTerminationInterrupt((*ctx).Err().Error(), span)
	default:
		// do nothing, this should not block the entire interpreter
		return nil
	}
}

func (e interpreterExecutor) genericPrinter(span errors.Span, args []value.Value, trailingNewLine bool) *value.VmInterrupt {
	output := make([]string, 0)
	for _, arg := range args {
		disp, i := arg.Display()
		if i != nil {
			return i
		}
		output = append(output, disp)
	}

	outStr := strings.Join(output, " ")
	if trailingNewLine {
		outStr = outStr + "\n"
	}

	if err := e.WriteStringTo(outStr); err != nil {
		return value.NewVMFatalException(
			err.Error(),
			value.Vm_HostErrorKind,
			span,
		)
	}

	return nil
}

func interpreterScopeAdditions() map[string]value.Value {
	// TODO: fill this
	return map[string]value.Value{
		"exit": *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
			code := args[0].(value.ValueInt).Inner
			return nil, value.NewVMExitInterrupt(code, span)
		}),
		"fmt": *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
			displays := make([]any, 0)

			for idx, arg := range args {
				if idx == 0 {
					continue
				}

				var out any

				if arg == nil {
					panic(fmt.Sprintf("One (or) more arguments to the `fmt` function were <nil> (pos=%d), (all=%v)", idx, args))
				}

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
		"print": *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
			return value.NewValueNull(), executor.(interpreterExecutor).genericPrinter(span, args, false)
		}),
		"println": *value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
			return value.NewValueNull(), executor.(interpreterExecutor).genericPrinter(span, args, true)
		}),
		"time": *value.NewValueObject(map[string]*value.Value{
			"sleep": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				durationSecs := args[0].(value.ValueFloat).Inner

				for i := 0; i < int(durationSecs*1000); i += 10 {
					if i := checkCancelation(cancelCtx, span); i != nil {
						return nil, i
					}
					time.Sleep(time.Millisecond * 10)
				}

				return value.NewValueNull(), nil
			}),
			"since": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				if args[0].Kind() != value.ObjectValueKind {
					fmt.Printf("illegal input: %v, %v\n", args, span)
				}

				milliObj := args[0].(value.ValueObject).FieldsInternal["unix_milli"]
				then := time.UnixMilli((*milliObj).(value.ValueInt).Inner)
				return createDurationObject(time.Since(then)), nil
			}),
			"add_days": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				base := createTimeStructFromObject(args[0])
				days := args[1].(value.ValueInt).Inner
				return createTimeObject(base.Add(time.Hour * 24 * time.Duration(days))), nil
			}),
			"add_hours": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				base := createTimeStructFromObject(args[0])
				hours := args[1].(value.ValueInt).Inner
				return createTimeObject(base.Add(time.Hour * time.Duration(hours))), nil
			}),
			"add_minutes": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
				base := createTimeStructFromObject(args[0])
				minutes := args[1].(value.ValueInt).Inner
				return createTimeObject(base.Add(time.Minute * time.Duration(minutes))), nil
			}),
			"now": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
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
		"display": value.NewValueBuiltinFunction(func(executor value.Executor, cancelCtx *context.Context, span errors.Span, args ...value.Value) (*value.Value, *value.VmInterrupt) {
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
