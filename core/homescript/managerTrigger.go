package homescript

import (
	"bytes"
	"context"
	"time"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/compiler"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
	driverTypes "github.com/smarthome-go/smarthome/core/device/driver/types"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

func (m *Manager) ProcessAnnotations(
	compileOutput compiler.CompileOutput,
	context types.ExecutionContext,
) (triggers []types.TriggerAnnotation, err error) {
	triggers = make([]types.TriggerAnnotation, 0)

	for annotationFn, annotation := range compileOutput.Annotations {
		for _, item := range annotation.Items {
			switch itemS := item.(type) {
			case compiler.IdentCompiledAnnotation:
				panic("not implemented")
			case compiler.TriggerCompiledAnnotation:
				ident := itemS.ArgumentFunctionIdent

				args, err := m.ExtractTriggerAnnotationArgs(
					itemS,
					ident,
					context,
				)
				if err != nil {
					return nil, err
				}

				mangledFn := compileOutput.Mappings.Functions[annotationFn.UnmangledFunction]
				sourceSpan := compileOutput.SourceMap[mangledFn][0]

				triggers = append(triggers, types.TriggerAnnotation{
					CalledFnIdentMangled: mangledFn,
					Trigger:              itemS.TriggerSource,
					Args:                 args,
					Span:                 sourceSpan,
				})
			}
		}
	}

	return triggers, nil
}

func (m *Manager) ExtractTriggerAnnotationArgs(
	annotation compiler.TriggerCompiledAnnotation,
	// target *HomescriptOrDriver,
	targetFunctionName string,
	executionContext types.ExecutionContext,
) ([]value.Value, error) {
	logger.Tracef(
		"Processing trigger annotation with target `%v` for function `%s`...",
		executionContext.Kind(),
		targetFunctionName,
	)

	buffer := bytes.Buffer{}

	const maxRuntime = time.Second * 2
	ctx, cancelFunc := context.WithTimeout(context.Background(), maxRuntime)

	var res types.HmsRes
	var err error

	switch exec := executionContext.(type) {
	case types.ExecutionContextDriver:
		res, err = m.RunDriverScript(
			driverTypes.DriverInvocationIDs{
				DeviceID: exec.DeviceID,
				VendorID: exec.DriverVendor,
				ModelID:  exec.DriverModel,
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
	case types.ExecutionContextUser:
		res, err = m.RunUserScriptTweakable(
			exec.Filename,
			exec.UsernameData,
			&runtime.FunctionInvocation{
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
			nil,
			false,
			nil,
			nil,
			make(map[string]string),
		)
	case types.ExecutionContextAutomation:
		panic("not implemented")
	}

	if err != nil {
		panic(err.Error())
	}

	if res.Errors.ContainsError {
		panic(res.Errors.Diagnostics)
	}

	argList := res.ReturnValue.(value.ValueList)

	// Make args.
	args := make([]value.Value, len(*argList.Values))
	for idx, src := range *argList.Values {
		args[idx] = *src
	}

	return args, nil
}
