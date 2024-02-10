package homescript

import (
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
)

const DefaultCapabilityName = "base"

//
//
//
// Driver template
//
//
//

const DriverModuleIdent = "driver"

//
// Driver functions.
//

const DriverFunctionValidate = "validate_driver"

func driverValidateDriverSignature(span errors.Span) ast.TemplateMethod {
	return ast.TemplateMethod{
		Signature: ast.NewFunctionType(
			ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)), span, ast.NewNullType(span), span,
		).(ast.FunctionType),
		Modifier: pAst.FN_MODIFIER_PUB,
	}
}

func driverTemplate(span errors.Span) DriverTemplate {
	return DriverTemplate{
		Spec: ast.TemplateSpec{
			BaseMethods: map[string]ast.TemplateMethod{
				DriverFunctionValidate: driverValidateDriverSignature(span),
			},
			Capabilities: map[string]ast.TemplateCapability{
				DefaultCapabilityName: {
					RequiresMethods:           []string{DriverFunctionValidate},
					ConflictsWithCapabilities: []ast.TemplateConflict{},
				},
			},
			DefaultCapabilities: []string{DefaultCapabilityName},
			Span:                span,
		},
		// TODO: implement this
		Capabilities: map[string]DriverCapability{
			DefaultCapabilityName: DriverCapabilityBase,
		},
	}
}

//
//
//
// Device template
//
//
//

//
// Device functions.
//

const DeviceFunctionValidate = "validate_device"
const DeviceFunctionReportPowerState = "report_power"
const DeviceFunctionReportPowerDraw = "report_power_draw"
const DeviceFuncionSetPower = "set_power"
const DeviceFuncionSetDim = "dim"
const DeviceFunctionReportDim = "report_dim"
const DeviceFunctionDim = "dim"

// TODO: maybe own submodule for templates?

func deviceValidateDeviceSignature(span errors.Span) ast.TemplateMethod {
	return ast.TemplateMethod{
		Signature: ast.NewFunctionType(
			ast.NewNormalFunctionTypeParamKind(
				make([]ast.FunctionTypeParam, 0)),
			span,
			ast.NewNullType(span),
			span,
		).(ast.FunctionType),
		Modifier: pAst.FN_MODIFIER_PUB,
	}
}

func DeviceReportPowerStateSignature(span errors.Span) ast.TemplateMethod {
	return ast.TemplateMethod{
		Signature: ast.NewFunctionType(
			ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
			span,
			ast.NewBoolType(span), // TODO: fix HMS type system, fuse with power_draw and return an object.
			span,
		).(ast.FunctionType),
		Modifier: pAst.FN_MODIFIER_PUB,
	}
}

func DeviceReportPowerDrawSignature(span errors.Span) ast.TemplateMethod {
	return ast.TemplateMethod{
		Signature: ast.NewFunctionType(
			ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
			span,
			ast.NewIntType(span), // TODO: fix HMS type system, fuse with power_draw and return an object.
			span,
		).(ast.FunctionType),
		Modifier: pAst.FN_MODIFIER_PUB,
	}
}

func DeviceSetPowerSignature(span errors.Span) ast.TemplateMethod {
	return ast.TemplateMethod{
		Signature: ast.NewFunctionType(
			ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
				ast.NewFunctionTypeParam(
					pAst.NewSpannedIdent("power_state", span),
					ast.NewBoolType(span),
					nil,
				),
			}), span, ast.NewBoolType(span), span,
		).(ast.FunctionType),
		Modifier: pAst.FN_MODIFIER_PUB,
	}
}

//
// Generic dimmer implementation
//

const ReportDimTypeLabelIdent = "label"
const ReportDimTypeRangeIdent = "range"
const ReportDimTypeValueIdent = "value"

func ReportDimType(span errors.Span) ast.Type {
	return ast.NewObjectType(
		[]ast.ObjectTypeField{
			ast.NewObjectTypeField(
				pAst.NewSpannedIdent(ReportDimTypeLabelIdent, span),
				ast.NewStringType(span),
				span,
			),
			ast.NewObjectTypeField(
				pAst.NewSpannedIdent(ReportDimTypeRangeIdent, span),
				ast.NewRangeType(span),
				span,
			),
			ast.NewObjectTypeField(
				pAst.NewSpannedIdent(ReportDimTypeValueIdent, span),
				ast.NewIntType(span),
				span,
			),
		},
		span,
	)
}

func DeviceReportDimSignature(span errors.Span) ast.TemplateMethod {
	return ast.TemplateMethod{
		Signature: ast.NewFunctionType(
			ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
			span,
			ast.NewListType(ReportDimType(span), span),
			span,
		).(ast.FunctionType),
		Modifier: pAst.FN_MODIFIER_PUB,
	}
}

func DeviceDimSignature(span errors.Span) ast.TemplateMethod {
	return ast.TemplateMethod{
		Signature: ast.NewFunctionType(
			ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
				ast.NewFunctionTypeParam(
					pAst.NewSpannedIdent("label", span),
					ast.NewStringType(span),
					nil,
				),
				ast.NewFunctionTypeParam(
					pAst.NewSpannedIdent("value", span),
					ast.NewIntType(span),
					nil,
				),
			}), span, ast.NewBoolType(span), span,
		).(ast.FunctionType),
		Modifier: pAst.FN_MODIFIER_PUB,
	}
}

func deviceTemplate(span errors.Span) DeviceTemplate {
	return DeviceTemplate{
		Spec: ast.TemplateSpec{
			BaseMethods: map[string]ast.TemplateMethod{
				DeviceFunctionValidate:         deviceValidateDeviceSignature(span),
				DeviceFunctionReportPowerState: DeviceReportPowerStateSignature(span),
				DeviceFunctionReportPowerDraw:  DeviceReportPowerDrawSignature(span),
				DeviceFuncionSetPower:          DeviceSetPowerSignature(span),
				DeviceFunctionReportDim:        DeviceReportDimSignature(span),
				DeviceFuncionSetDim:            DeviceDimSignature(span),
			},
			Capabilities: map[string]ast.TemplateCapability{
				DefaultCapabilityName: {
					RequiresMethods:           []string{DeviceFunctionValidate},
					ConflictsWithCapabilities: []ast.TemplateConflict{},
				},
				"dimmable": {
					RequiresMethods: []string{
						DeviceFunctionReportDim,
						DeviceFuncionSetDim,
					},
					ConflictsWithCapabilities: []ast.TemplateConflict{},
				},
				"power": {
					RequiresMethods: []string{
						DeviceFuncionSetPower,
						DeviceFunctionReportPowerState,
						DeviceFunctionReportPowerDraw,
					},
					ConflictsWithCapabilities: []ast.TemplateConflict{},
				},
			},
			DefaultCapabilities: []string{"base"},
			Span:                span,
		},
		// TODO: implement this
		Capabilities: map[string]DeviceCapability{
			"base":     DeviceCapabilityBase,
			"power":    DeviceCapabilityPower,
			"dimmable": DeviceCapabilityDimmable,
		},
	}
}

// NOTE: here, all important templates are defined so that additional information can be attached to it.
// TODO: add integration tests for checking if all HMS template capabilities have a mapping.
func Templates(span errors.Span) map[ImportKey]Template {
	return map[ImportKey]Template{
		{
			ModuleName: DriverModuleIdent,
			ValueName:  DRIVER_TEMPLATE_IDENT,
		}: driverTemplate(span),
		{
			ModuleName: DriverModuleIdent,
			ValueName:  DEVICE_TEMPLATE_IDENT,
		}: deviceTemplate(span),
	}
}
