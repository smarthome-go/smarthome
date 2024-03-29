package driver

import (
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
	"github.com/smarthome-go/smarthome/core/homescript/types"
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

type DeviceTemplate struct {
	Spec ast.TemplateSpec
	// Makes the HMS capability identifier to a `DriverCapability`.
	Capabilities map[string]DeviceCapability
}

func (self DeviceTemplate) Kind() types.TemplateKind {
	return types.TemplateKindDevice
}

func (self DeviceTemplate) GetSpec() ast.TemplateSpec {
	return self.Spec
}

type DriverTemplate struct {
	Spec ast.TemplateSpec
	// Makes the HMS capability identifier to a `DriverCapability`.
	Capabilities map[string]DriverCapability
}

func (self DriverTemplate) Kind() types.TemplateKind {
	return types.TemplateKindDriver
}

func (self DriverTemplate) GetSpec() ast.TemplateSpec {
	return self.Spec
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

const DeviceFunctionValidateDevice = "validate_device"
const DeviceFunctionValidateDriver = "validate_device"
const DeviceFunctionReportSensorReadings = "report_sensor_readings"
const DeviceFunctionReportPowerState = "report_power"
const DeviceFunctionReportPowerDraw = "report_power_draw"
const DeviceFunctionSetPower = "set_power"
const DeviceFunctionSetDim = "dim"
const DeviceFunctionReportDim = "report_dim"
const DeviceFunctionDim = "dim"

// TODO: maybe own submodule for templates?

func deviceValidateDeviceOrDriverSignature(span errors.Span) ast.TemplateMethod {
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

///
/// Generic sensor implementation
///

const ReportSensorTypeLabelIdent = "label"
const ReportSensorTypeValueIdent = "value"
const ReportSensorTypeUnitIdent = "unit"

func ReportSensorReadingType(span errors.Span) ast.Type {
	return ast.NewObjectType(
		[]ast.ObjectTypeField{
			ast.NewObjectTypeField(
				pAst.NewSpannedIdent(ReportSensorTypeLabelIdent, span),
				ast.NewStringType(span),
				span,
			),
			ast.NewObjectTypeField(
				pAst.NewSpannedIdent(ReportSensorTypeValueIdent, span),
				ast.NewAnyType(span), // TODO: is this the way?
				span,
			),
			ast.NewObjectTypeField(
				pAst.NewSpannedIdent(ReportSensorTypeUnitIdent, span),
				ast.NewStringType(span),
				span,
			),
		},
		span,
	)
}

func DeviceReportSensorReadingsSignature(span errors.Span) ast.TemplateMethod {
	return ast.TemplateMethod{
		Signature: ast.NewFunctionType(
			ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
			span,
			ast.NewListType(ReportSensorReadingType(span), span),
			span,
		).(ast.FunctionType),
		Modifier: pAst.FN_MODIFIER_PUB,
	}
}

///
/// Power signature(s)
///

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
				DeviceFunctionValidateDevice:       deviceValidateDeviceOrDriverSignature(span),
				DeviceFunctionReportSensorReadings: DeviceReportSensorReadingsSignature(span),
				DeviceFunctionReportPowerState:     DeviceReportPowerStateSignature(span),
				DeviceFunctionReportPowerDraw:      DeviceReportPowerDrawSignature(span),
				DeviceFunctionSetPower:             DeviceSetPowerSignature(span),
				DeviceFunctionReportDim:            DeviceReportDimSignature(span),
				DeviceFunctionSetDim:               DeviceDimSignature(span),
			},
			Capabilities: map[string]ast.TemplateCapability{
				DefaultCapabilityName: {
					RequiresMethods:           []string{DeviceFunctionValidateDevice},
					ConflictsWithCapabilities: []ast.TemplateConflict{},
				},
				"dimmable": {
					RequiresMethods: []string{
						DeviceFunctionReportDim,
						DeviceFunctionSetDim,
					},
					ConflictsWithCapabilities: []ast.TemplateConflict{},
				},
				"power": {
					RequiresMethods: []string{
						DeviceFunctionSetPower,
						DeviceFunctionReportPowerState,
						DeviceFunctionReportPowerDraw,
					},
					ConflictsWithCapabilities: []ast.TemplateConflict{},
				},
				"sensor": {
					RequiresMethods: []string{
						DeviceFunctionReportSensorReadings,
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
			"sensor":   DeviceCapabilitySensor,
		},
	}
}

// NOTE: here, all important templates are defined so that additional information can be attached to it.
// TODO: add integration tests for checking if all HMS template capabilities have a mapping.
func Templates(span errors.Span) map[types.ImportKey]types.Template {
	return map[types.ImportKey]types.Template{
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
