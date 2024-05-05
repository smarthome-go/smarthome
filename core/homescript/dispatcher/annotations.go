package dispatcher

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/smarthome-go/homescript/v3/homescript"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/compiler"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	"github.com/smarthome-go/smarthome/core/database"
	driverTypes "github.com/smarthome-go/smarthome/core/device/driver/types"
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

	for _, driver := range drivers {
		if err := i.ReloadDriver(driver); err != nil {
			return err
		}
	}

	return nil
}

func (i *InstanceT) ReloadDriver(driver database.DeviceDriver) error {
	devices, err := database.ListAllDevices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		if device.VendorID != driver.VendorID || device.ModelID != driver.ModelID {
			continue
		}

		if err := i.RegisterDevice(driver, device.ID); err != nil {
			logger.Errorf("Failed to register device: %s", err.Error())
			return err
		}
	}

	if err := database.ModifyDeviceDriverDirty(driver.VendorID, driver.ModelID, false); err != nil {
		return err
	}

	logger.Infof("Successfully reloaded driver `%s:%s`", driver.VendorID, driver.ModelID)

	return nil
}

func (i *InstanceT) RegisterDevice(driver database.DeviceDriver, deviceID string) error {
	//
	// First step: unregister any dispatcher hooks which are attached to this device.
	//

	i.DoneRegistrations.Lock.Lock()
	doneRegs := i.DoneRegistrations.Set
	i.DoneRegistrations.Lock.Unlock()

	for id, reg := range doneRegs {
		var ctx types.ExecutionContext

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

		if err := i.Unregister(id); err != nil {
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

	for annotationFn, annotation := range compileOutput.Annotations {
		fmt.Printf("============= %v\n", annotation)
		for _, item := range annotation.Items {
			switch itemS := item.(type) {
			case compiler.IdentCompiledAnnotation:
			case compiler.TriggerCompiledAnnotation:
				ident := itemS.ArgumentFunctionIdent
				args, err := i.ExtractDriverTriggerAnnotationArgs(itemS,
					&HomescriptOrDriver{
						IsDriver: true,
						//nolint:exhaustruct
						Homescript: database.Homescript{},
						Device: DeviceTarget{
							DeviceID: deviceID,
							Driver:   driver,
						},
					},
					ident,
					context,
				)
				if err != nil {
					return err
				}

				// Transform the argument list to a string list.
				untypedList := args[0].(value.ValueList).Values

				topics := make([]string, len(*untypedList))
				for idx, arg := range *untypedList {
					topics[idx] = (*arg).(value.ValueString).Inner
				}

				if err := i.RegisterTriggerAnnotation(
					types.CreateDriverHmsId(database.DriverTuple{
						VendorID: driver.VendorID,
						ModelID:  driver.ModelID,
					}),
					dispatcherTypes.CalledFunction{
						Ident:          annotationFn.UnmangledFunction,
						IdentIsLiteral: false,
						// TODO: decide on the callmode
						CallMode: dispatcherTypes.CallModeAllocating{
							Context: context,
						},
					},
					topics,
				); err != nil {
					// TODO: what kind of error is this?
					// Is this fatal?
					return err
				}
			}
		}
	}

	logger.Infof("Successfully registered device `%s`", deviceID)
	return nil
}

func (i *InstanceT) RegisterTriggerAnnotation(
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

func (i *InstanceT) ExtractDriverTriggerAnnotationArgs(
	annotation compiler.TriggerCompiledAnnotation,
	target *HomescriptOrDriver,
	targetFunctionName string,
	executionContext types.ExecutionContext,
) ([]value.Value, error) {
	logger.Tracef(
		"Processing trigger annotation with target `%s:%s` for function `%s`...",
		target.Device.Driver.VendorID,
		target.Device.Driver.ModelID,
		targetFunctionName,
	)

	buffer := bytes.Buffer{}
	startTrigger := time.Now()

	const maxRuntime = time.Second * 2
	ctx, cancelFunc := context.WithTimeout(context.Background(), maxRuntime)

	res, err := i.Hms.RunDriverScript(
		driverTypes.DriverInvocationIDs{
			DeviceID: &target.Device.DeviceID,
			VendorID: target.Device.Driver.VendorID,
			ModelID:  target.Device.Driver.ModelID,
		},
		runtime.FunctionInvocation{
			Function:    annotation.ArgumentFunctionIdent,
			LiteralName: true,
			Args:        []value.Value{},
			FunctionSignature: runtime.FunctionInvocationSignature{
				Params:     []runtime.FunctionInvocationSignatureParam{},
				ReturnType: ast.NewListType(ast.NewAnyType(errors.Span{}), errors.Span{}),
			},
		},
		types.Cancelation{
			Context:    ctx,
			CancelFunc: cancelFunc,
		},
		&buffer,
	)

	if res.Errors.ContainsError {
		panic(res.Errors.Diagnostics)
	}

	argList := res.ReturnValue.(value.ValueList)

	disp, erR := argList.Display()
	if err != nil {
		panic((*erR).Message())
	}

	fmt.Printf(
		"====> (%v) FN = `%s:%s` | ARGS = `%s`\n",
		time.Since(startTrigger),
		annotation.CallbackFnIdent,
		annotation.ArgumentFunctionIdent,
		disp,
	)

	// Make args.
	args := make([]value.Value, len(*argList.Values))
	for idx, src := range *argList.Values {
		args[idx] = *src
	}

	return args, nil
}
