package dispatcher

import (
	"errors"
	"fmt"
	"strings"

	"github.com/smarthome-go/homescript/v3/homescript"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	dispatcherTypes "github.com/smarthome-go/smarthome/core/homescript/dispatcher/types"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

type DeviceTarget struct {
	DeviceID string
	Driver   database.DeviceDriver
}

type HomescriptOrDriver struct {
	IsDriver   bool
	Homescript database.Homescript
	Device     DeviceTarget
}

func (i *InstanceT) RegisterDriverAnnotations() error {
	drivers, err := database.ListDeviceDrivers()
	if err != nil {
		return err
	}

	errs := make([]string, 0)

	for _, driver := range drivers {
		logger.Tracef("    => Registering driver `%s:%s`", driver.VendorID, driver.ModelID)
		if err := i.ReloadDriver(driver); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("Could not reload drivers: %s", strings.Join(errs, "; "))
	}

	return nil
}

func (i *InstanceT) ReloadDriver(driver database.DeviceDriver) error {
	errCnt := 0

	devices, err := database.ListAllDevices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		if device.VendorID != driver.VendorID || device.ModelID != driver.ModelID {
			continue
		}

		if err := i.RegisterDevice(driver, device.ID); err != nil {
			logger.Warnf("Failed to register device: %s", err.Error())
			errCnt++
		}
	}

	if err := database.ModifyDeviceDriverDirty(driver.VendorID, driver.ModelID, false); err != nil {
		return err
	}

	logger.Infof("Successfully reloaded driver `%s:%s`", driver.VendorID, driver.ModelID)

	if errCnt != 0 {
		return fmt.Errorf("%d device(s) could not be registered", errCnt)
	}

	return nil
}

func (i *InstanceT) RegisterUserScript(programID string, username string) error {
	//
	// First step: unregister any dispatcher hooks which are attached to this device.
	//

	logger.Tracef("Registering user script %s...", programID)

	//
	// Unregister all old registrations.
	//

	if err := i.UnregisterUserProgram(programID); err != nil {
		return err
	}

	context := types.NewExecutionContextUser(
		programID,
		username,
		nil,
	)

	program, diagnostics, err := i.Hms.AnalyzeUserScript(context)

	if err != nil {
		return err
	}

	// Skip further extraction of this device.
	if diagnostics.ContainsError {
		d := make([]string, 0)
		for _, di := range diagnostics.Diagnostics {
			d = append(d, di.String())
		}

		return fmt.Errorf(
			"Could not process user annotation: user `%s`, script `%s` extraction failed: %s",
			username,
			programID,
			strings.Join(d, ", "),
		)
	}

	compileOutput, err := i.Hms.Compile(program, programID)
	if err != nil {
		return err
	}

	triggers, err := i.Hms.ProcessAnnotations(
		compileOutput,
		context,
	)

	if err != nil {
		return err
	}

	for _, trigger := range triggers {
		switch trigger.Trigger {
		case types.TriggerDeviceEvent:
			containsEmpty := false

			deviceID := trigger.Args[0].(value.ValueString).Inner

			topicListRaw := trigger.Args[1].(value.ValueOption).Inner
			var topics *[]string

			if topicListRaw != nil {
				values := *(*topicListRaw).(value.ValueList).Values
				topicsR := make([]string, len(values))
				topics = &topicsR

				for idx, arg := range values {
					vString := (*arg).(value.ValueString).Inner
					(*topics)[idx] = vString

					if vString == "" && !containsEmpty {
						containsEmpty = true
					}
				}
			}

			// Sanity-check the arguments.
			if len(*topics) == 0 || containsEmpty {
				return errors.New("Empty lists or empty strings are not allowed as topics")
			}

			id, err := i.registerInternal(
				dispatcherTypes.RegisterInfo{
					ProgramID: programID,
					Function: &dispatcherTypes.CalledFunction{
						Ident:          trigger.CalledFnIdentMangled,
						IdentIsLiteral: true,
						CallMode: dispatcherTypes.CallModeAdaptive{
							AllocatingFallback: dispatcherTypes.CallModeAllocating{
								Context: context,
							},
						},
					},
					Trigger: dispatcherTypes.CallbackTriggerDeviceAction{
						FilterKind: dispatcherTypes.DeviceFilterKind(dispatcherTypes.DeviceFilterIndividual{
							ID: deviceID,
						}),
						Topics:        *topics,
						TopicWildcard: topics == nil,
					},
				},
			)

			if err != nil {
				return err
			}

			logger.Infof("Register user program: %d", id)
		case types.TriggerDeviceClassEvent:
			panic("HI")
		default:
			continue
		}
	}

	return nil
}

func (i *InstanceT) RegisterDevice(driver database.DeviceDriver, deviceID string) error {
	//
	// First step: unregister any dispatcher hooks which are attached to this device.
	//

	logger.Tracef("Registering device `%s:%s` (%s)...", driver.VendorID, driver.ModelID, deviceID)

	i.DoneRegistrations.Lock.Lock()
	defer i.DoneRegistrations.Lock.Unlock()

	set := i.DoneRegistrations.Set

	//
	// Unregister all old registrations.
	//

	for id, reg := range set {
		var ctx types.ExecutionContext

		if reg.Function == nil || reg.Function.CallMode == nil {
			continue
		}

		switch mode := reg.Function.CallMode.(type) {
		case dispatcherTypes.CallModeAllocating:
			ctx = mode.Context
		case dispatcherTypes.CallModeAdaptive:
			ctx = mode.AllocatingFallback.Context
		case dispatcherTypes.CallModeAttaching:
			continue
		}

		if ctx.Kind() != types.HMS_PROGRAM_KIND_DEVICE_DRIVER {
			continue
		}

		driverCtx := ctx.(types.ExecutionContextDriver)
		if driverCtx.DriverVendor != driver.VendorID ||
			driverCtx.DriverModel != driver.ModelID ||
			// NOTE: deref is ok because the support for running drivers without devices attached will be removed soon.
			*driverCtx.DeviceID != deviceID {
			continue
		}

		i.DoneRegistrations.Lock.Unlock()
		err := i.Unregister(id)
		i.DoneRegistrations.Lock.Lock()

		if err != nil {
			return err
		}

		logger.Tracef("[driver register] Removed previous dispatcher registration `%d`", id)
	}

	context := types.NewExecutionContextDriver(
		driver.VendorID,
		driver.ModelID,
		&deviceID,
	)

	filename := types.CreateDriverHmsId(database.DriverTuple{
		VendorID: driver.VendorID,
		ModelID:  driver.ModelID,
	})

	program, diagnostics, err := i.Hms.Analyze(
		homescript.InputProgram{
			ProgramText: driver.HomescriptCode,
			Filename:    filename,
		},
		context,
	)

	if err != nil {
		return err
	}

	// Skip further extraction of this device.
	if diagnostics.ContainsError {
		return fmt.Errorf(
			"Could not process driver annotation: driver `%s:%s` extraction failed",
			driver.VendorID,
			driver.ModelID,
		)
	}

	compileOutput, err := i.Hms.Compile(program, filename)
	if err != nil {
		return err
	}

	triggers, err := i.Hms.ProcessAnnotations(
		compileOutput,
		context,
	)

	if err != nil {
		return err
	}

	for _, trigger := range triggers {
		switch trigger.Trigger {
		case types.TriggerMqttMessageIdent:
			containsEmpty := false

			argList := trigger.Args[0].(value.ValueList).Values
			topics := make([]string, len(*argList))

			for idx, arg := range *argList {
				vString := (*arg).(value.ValueString).Inner
				topics[idx] = vString

				if vString == "" && !containsEmpty {
					containsEmpty = true
				}
			}

			// Sanity-check the arguments.
			if len(topics) == 0 || containsEmpty {
				return fmt.Errorf("Empty lists or empty strings are not allowed as topics")
			}

			i.DoneRegistrations.Lock.Unlock()
			err := i.RegisterMqttTriggerAnnotation(
				types.CreateDriverHmsId(database.DriverTuple{
					VendorID: driver.VendorID,
					ModelID:  driver.ModelID,
				}),
				dispatcherTypes.CalledFunction{
					Ident:          trigger.CalledFnIdentMangled,
					IdentIsLiteral: true,
					CallMode: dispatcherTypes.CallModeAdaptive{
						AllocatingFallback: dispatcherTypes.CallModeAllocating{
							Context: context,
						},
					},
				},
				topics,
			)
			i.DoneRegistrations.Lock.Lock()

			if err != nil {
				return err
			}
		default:
			continue
		}
	}

	logger.Infof("Successfully registered device `%s`", deviceID)
	return nil
}

func (i *InstanceT) RegisterMqttTriggerAnnotation(
	programID string,
	callback dispatcherTypes.CalledFunction,
	mqttTopics []string,
) error {
	_, err := i.Register(
		dispatcherTypes.RegisterInfo{
			ProgramID: programID,
			Function:  &callback,
			Trigger: dispatcherTypes.CallBackTriggerMqtt{
				Topics: mqttTopics,
			},
		},
		dispatcherTypes.ToleranceRetry,
	)

	return err
}
