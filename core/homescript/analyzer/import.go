package analyzer

import (
	"github.com/smarthome-go/homescript/v3/homescript/analyzer"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
	"github.com/smarthome-go/smarthome/core/device/driver"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

type importHandler func(context types.ExecutionContext, span errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool)

func httpResponseType(span errors.Span) ast.Type {
	return ast.NewObjectType(
		[]ast.ObjectTypeField{
			ast.NewObjectTypeField(pAst.NewSpannedIdent("status", span), ast.NewStringType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("status_code", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("body", span), ast.NewStringType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("cookies", span), ast.NewAnyObjectType(span), span),
		},
		span,
	)
}

func deviceEventCallbackType(span errors.Span) ast.FunctionType {
	return ast.NewFunctionType(
		ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
			ast.NewFunctionTypeParam(pAst.NewSpannedIdent("data", span), ast.NewAnyObjectType(span), nil),
			ast.NewFunctionTypeParam(pAst.NewSpannedIdent("topic", span), ast.NewStringType(span), nil),
		}),
		span,
		ast.NewNullType(span),
		span,
	).(ast.FunctionType)
}

func deviceTriggerFilterParamType(span errors.Span) ast.FunctionTypeParam {
	return ast.NewFunctionTypeParam(
		pAst.NewSpannedIdent("topics", span),
		ast.NewOptionType(
			ast.NewListType(ast.NewStringType(span), span),
			span,
		),
		nil,
	)
}

var builtinImportHandlers = map[types.ImportKey]importHandler{
	{ModuleName: "triggers", ValueName: types.TriggerKillIdent}: func(_ types.ExecutionContext, span errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		if kind != pAst.IMPORT_KIND_TRIGGER {
			return analyzer.BuiltinImport{}, false
		}

		return analyzer.BuiltinImport{
			Trigger: &analyzer.TriggerFunction{
				TriggerFnType: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
					span,
					ast.NewNullType(span),
					span,
				).(ast.FunctionType),
				CallbackFnType: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
					span,
					ast.NewNullType(span),
					span,
				).(ast.FunctionType),
				Connective: pAst.OnTriggerDispatchKeyword,
				ImportedAt: span,
			},
			Type:     nil,
			Template: nil,
		}, true
	},
	{ModuleName: "triggers", ValueName: types.TriggerMinuteIdent}: func(_ types.ExecutionContext, span errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		if kind != pAst.IMPORT_KIND_TRIGGER {
			return analyzer.BuiltinImport{}, false
		}

		return analyzer.BuiltinImport{
			Trigger: &analyzer.TriggerFunction{
				TriggerFnType: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind(
						[]ast.FunctionTypeParam{ast.NewFunctionTypeParam(
							pAst.NewSpannedIdent("minutes", span),
							ast.NewIntType(span),
							nil,
						)},
					),
					span,
					ast.NewNullType(span),
					span,
				).(ast.FunctionType),
				CallbackFnType: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(
							pAst.NewSpannedIdent("elapsed", span),
							ast.NewIntType(span),
							nil,
						),
					}),
					span,
					ast.NewNullType(span),
					span,
				).(ast.FunctionType),
				Connective: pAst.AtTriggerDispatchKeyword,
				ImportedAt: span,
			},
			Type:     nil,
			Template: nil,
		}, true
	},
	{ModuleName: "driver", ValueName: "Driver"}: func(_ types.ExecutionContext, _ errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		if kind != pAst.IMPORT_KIND_TEMPLATE {
			return analyzer.BuiltinImport{}, false
		}

		return analyzer.BuiltinImport{
			Type:     nil,
			Template: nil,
		}, true
	},
	{ModuleName: "driver", ValueName: "Device"}: func(_ types.ExecutionContext, _ errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		if kind != pAst.IMPORT_KIND_TEMPLATE {
			return analyzer.BuiltinImport{}, false
		}

		return analyzer.BuiltinImport{
			Type:     nil,
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "driver", ValueName: "DriverMeta"}: func(_ types.ExecutionContext, span errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		if kind != pAst.IMPORT_KIND_TYPE {
			return analyzer.BuiltinImport{}, false
		}

		return analyzer.BuiltinImport{
			Type: ast.NewObjectType([]ast.ObjectTypeField{
				{
					FieldName: pAst.NewSpannedIdent("vendor_id", span),
					Type:      ast.NewStringType(span),
					Span:      span,
				},
				{
					FieldName: pAst.NewSpannedIdent("model_id", span),
					Type:      ast.NewStringType(span),
					Span:      span,
				},
				{
					FieldName: pAst.NewSpannedIdent("version", span),
					Type:      ast.NewStringType(span),
					Span:      span,
				},
			}, span),
			Template: nil,
		}, true
	},
	{ModuleName: "driver", ValueName: "Dimmable"}: func(_ types.ExecutionContext, span errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		if kind != pAst.IMPORT_KIND_TYPE {
			return analyzer.BuiltinImport{}, false
		}

		return analyzer.BuiltinImport{
			Type:     driver.ReportDimType(span),
			Template: nil,
		}, true
	},
	{ModuleName: "driver", ValueName: "Sensor"}: func(_ types.ExecutionContext, span errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		if kind != pAst.IMPORT_KIND_TYPE {
			return analyzer.BuiltinImport{}, false
		}

		return analyzer.BuiltinImport{
			Type:     driver.ReportSensorReadingType(span),
			Template: nil,
		}, true
	},
	{ModuleName: "driver", ValueName: "Color"}: func(_ types.ExecutionContext, span errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		if kind != pAst.IMPORT_KIND_TYPE {
			return analyzer.BuiltinImport{}, false
		}

		return analyzer.BuiltinImport{
			Type:     driver.RgbColorType(span),
			Template: nil,
		}, true
	},
	{ModuleName: "mqtt", ValueName: types.TriggerMqttMessageIdent}: func(_ types.ExecutionContext, span errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		if kind != pAst.IMPORT_KIND_TRIGGER {
			return analyzer.BuiltinImport{}, false
		}

		return analyzer.BuiltinImport{
			Type:     nil,
			Template: nil,
			Trigger: &analyzer.TriggerFunction{
				TriggerFnType: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind(
						[]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(
								pAst.NewSpannedIdent("topics", span),
								ast.NewListType(ast.NewStringType(span), span),
								nil,
							),
						},
					),
					span,
					ast.NewNullType(span),
					span,
				).(ast.FunctionType),
				CallbackFnType: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind(
						[]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(
								pAst.NewSpannedIdent("topic", span),
								ast.NewStringType(span),
								nil,
							),
							ast.NewFunctionTypeParam(
								pAst.NewSpannedIdent("payload", span),
								ast.NewStringType(span),
								nil,
							),
						},
					),
					span,
					ast.NewNullType(span),
					span,
				).(ast.FunctionType),
				Connective: pAst.OnTriggerDispatchKeyword,
				ImportedAt: span,
			},
		}, true
	},
	{ModuleName: "mqtt", ValueName: "subscribe"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("topics", span), ast.NewListType(ast.NewStringType(span), span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("callback", span), MqttCallbackFn(span), nil),
				}),
				span,
				ast.NewNullType(span),
				span,
			),
			Template: nil,
		}, true
	},
	{ModuleName: "mqtt", ValueName: "publish"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("topic", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("payload", span), ast.NewStringType(span), nil),
				}),
				span,
				ast.NewNullType(span),
				span,
			),
			Template: nil,
		}, true
	},
	{ModuleName: "hms", ValueName: "exec"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("script_id", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("arguments", span), ast.NewOptionType(ast.NewAnyObjectType(span), span), nil),
				}),
				span,
				ast.NewNullType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "hms", ValueName: "exec_user"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("username", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("script_id", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("arguments", span), ast.NewOptionType(ast.NewAnyObjectType(span), span), nil),
				}),
				span,
				ast.NewNullType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "location", ValueName: "sun_times"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		timeObjType := func(span errors.Span) ast.Type {
			return ast.NewObjectType([]ast.ObjectTypeField{
				ast.NewObjectTypeField(pAst.NewSpannedIdent("hour", span), ast.NewIntType(span), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("minute", span), ast.NewIntType(span), span),
			}, span)
		}

		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
				span,
				ast.NewObjectType([]ast.ObjectTypeField{
					ast.NewObjectTypeField(pAst.NewSpannedIdent("sunrise", span), timeObjType(span), span),
					ast.NewObjectTypeField(pAst.NewSpannedIdent("sunset", span), timeObjType(span), span),
				}, span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "location", ValueName: "weather"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
				span,

				// WeatherTitle       string  `json:"weatherTitle"`
				// WeatherDescription string  `json:"weatherDescription"`
				// Temperature        float32 `json:"temperature"`
				// FeelsLike          float32 `json:"feelsLike"`
				// Humidity           uint8   `json:"humidity"`

				ast.NewObjectType([]ast.ObjectTypeField{
					ast.NewObjectTypeField(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), span),
					ast.NewObjectTypeField(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), span),
					ast.NewObjectTypeField(pAst.NewSpannedIdent("temperature", span), ast.NewFloatType(span), span),
					ast.NewObjectTypeField(pAst.NewSpannedIdent("feels_like", span), ast.NewFloatType(span), span),
					ast.NewObjectTypeField(pAst.NewSpannedIdent("humidity", span), ast.NewIntType(span), span),
				}, span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "device", ValueName: "emit"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("topic", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("data", span), ast.NewAnyType(span), nil),
				}),
				span,
				ast.NewNullType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	// TODO: color
	{ModuleName: "device", ValueName: "set_color"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("device_id", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("color", span), ast.NewObjectType(
						[]ast.ObjectTypeField{
							ast.NewObjectTypeField(
								pAst.NewSpannedIdent("r", span),
								ast.NewIntType(span),
								span,
							),
							ast.NewObjectTypeField(
								pAst.NewSpannedIdent("g", span),
								ast.NewIntType(span),
								span,
							),
							ast.NewObjectTypeField(
								pAst.NewSpannedIdent("b", span),
								ast.NewIntType(span),
								span,
							)},
						span,
					), nil),
				}),
				span,
				ast.NewBoolType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "device", ValueName: "set_power"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("device_id", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("power", span), ast.NewBoolType(span), nil),
				}),
				span,
				ast.NewBoolType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "device", ValueName: "dim"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("device_id", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("function", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("value", span), ast.NewIntType(span), nil),
				}),
				span,
				ast.NewBoolType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "device", ValueName: types.TriggerDeviceEvent}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		deviceEventCallBack := deviceEventCallbackType(span)
		deviceTriggerFilterParam := deviceTriggerFilterParamType(span)

		return analyzer.BuiltinImport{
			Type:     nil,
			Template: nil,
			Trigger: &analyzer.TriggerFunction{
				TriggerFnType: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("device_id", span), ast.NewStringType(span), nil),
						deviceTriggerFilterParam,
					}),
					span,
					ast.NewNullType(span),
					span,
				).(ast.FunctionType),
				CallbackFnType: deviceEventCallBack,
				Connective:     pAst.OnTriggerDispatchKeyword,
				ImportedAt:     span,
			},
		}, true
	},
	{ModuleName: "device", ValueName: types.TriggerDeviceClassEvent}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		deviceEventCallBack := deviceEventCallbackType(span)
		deviceTriggerFilterParam := deviceTriggerFilterParamType(span)

		return analyzer.BuiltinImport{
			Type:     nil,
			Template: nil,
			Trigger: &analyzer.TriggerFunction{
				TriggerFnType: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("vendor", span), ast.NewStringType(span), nil),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("model", span), ast.NewStringType(span), nil),
						deviceTriggerFilterParam,
					}),
					span,
					ast.NewNullType(span),
					span,
				).(ast.FunctionType),
				CallbackFnType: deviceEventCallBack,
				Connective:     pAst.OnTriggerDispatchKeyword,
				ImportedAt:     span,
			},
		}, true
	},
	{ModuleName: "widget", ValueName: "on_click_js"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("base", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("js", span), ast.NewStringType(span), nil),
				}),
				span, ast.NewStringType(span), span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "widget", ValueName: "on_click_hms"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("base", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("js", span), ast.NewStringType(span), nil),
				}),
				span, ast.NewStringType(span), span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "testing", ValueName: "assert_eq"}: func(_ types.ExecutionContext, _ errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("lhs", errors.Span{}), ast.NewUnknownType(), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("rhs", errors.Span{}), ast.NewUnknownType(), nil),
				}),
				errors.Span{},
				ast.NewNullType(errors.Span{}),
				errors.Span{},
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "storage", ValueName: "set_storage"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(ast.NewNormalFunctionTypeParamKind(
				[]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("key", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("value", span), ast.NewUnknownType(), nil),
				},
			),
				span,
				ast.NewNullType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "storage", ValueName: "get_storage"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(ast.NewNormalFunctionTypeParamKind(
				[]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("key", span), ast.NewStringType(span), nil),
				},
			),
				span,
				ast.NewOptionType(ast.NewStringType(span), span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "reminder", ValueName: "remind"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("reminder", span),
						ast.NewObjectType([]ast.ObjectTypeField{
							ast.NewObjectTypeField(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("priority", span), ast.NewIntType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("due_date_day", span), ast.NewIntType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("due_date_month", span), ast.NewIntType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("due_date_year", span), ast.NewIntType(span), span),
						}, span), nil),
				}),
				span,
				ast.NewIntType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "net", ValueName: "ping"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("ip", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("timeout", span), ast.NewFloatType(span), nil),
				}),
				span,
				ast.NewBoolType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "net", ValueName: "udp_send"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("host", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("port", span), ast.NewIntType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("data", span), ast.NewStringType(span), nil),
				}),
				span,
				ast.NewNullType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "net", ValueName: "HttpResponse"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type:     httpResponseType(span),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "net", ValueName: "http"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewObjectType([]ast.ObjectTypeField{
				ast.NewObjectTypeField(pAst.NewSpannedIdent("get", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{ast.NewFunctionTypeParam(pAst.NewSpannedIdent("url", span), ast.NewStringType(span), nil)}),
					span,
					httpResponseType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("generic", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind(
						[]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("url", span), ast.NewStringType(span), nil),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("method", span), ast.NewStringType(span), nil),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("body", span), ast.NewOptionType(ast.NewStringType(span), span), nil),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("headers", span), ast.NewAnyObjectType(span), nil),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("cookies", span), ast.NewAnyObjectType(span), nil),
						},
					),
					span,
					httpResponseType(span),
					span,
				), span),
			}, span),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "log", ValueName: "logger"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewObjectType([]ast.ObjectTypeField{
				ast.NewObjectTypeField(pAst.NewSpannedIdent("trace", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), nil),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), nil),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("debug", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), nil),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), nil),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("info", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), nil),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), nil),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("warn", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), nil),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), nil),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("error", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), nil),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), nil),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("fatal", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), nil),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), nil),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
			}, span),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "context", ValueName: "args"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type:     ast.NewAnyObjectType(span),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "context", ValueName: "Notification"}: func(_ types.ExecutionContext, span errors.Span, kind pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		notificationType := ast.NewObjectType([]ast.ObjectTypeField{
			ast.NewObjectTypeField(pAst.NewSpannedIdent("id", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("level", span), ast.NewIntType(span), span),
		}, span)

		if kind != pAst.IMPORT_KIND_TYPE {
			return analyzer.BuiltinImport{}, true
		}

		return analyzer.BuiltinImport{
			Type:     notificationType,
			Template: nil,
			Trigger:  nil,
		}, true
	},
	{ModuleName: "context", ValueName: "notification"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		notificationType := ast.NewObjectType([]ast.ObjectTypeField{
			ast.NewObjectTypeField(pAst.NewSpannedIdent("id", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("level", span), ast.NewIntType(span), span),
		}, span)

		return analyzer.BuiltinImport{
			Type:     ast.NewOptionType(notificationType, span),
			Template: nil,
		}, true
	},
	{ModuleName: "scheduler", ValueName: "create_schedule"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("schedule", span),
						ast.NewObjectType(
							[]ast.ObjectTypeField{
								ast.NewObjectTypeField(pAst.NewSpannedIdent("name", span), ast.NewStringType(span), span),
								ast.NewObjectTypeField(pAst.NewSpannedIdent("hour", span), ast.NewIntType(span), span),
								ast.NewObjectTypeField(pAst.NewSpannedIdent("minute", span), ast.NewIntType(span), span),
								ast.NewObjectTypeField(pAst.NewSpannedIdent("code", span), ast.NewStringType(span), span),
							},
							span,
						),
						nil,
					),
				}),
				span,
				ast.NewIntType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "scheduler", ValueName: "delete_schedule"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("id", span), ast.NewIntType(span), nil),
				}),
				span,
				ast.NewNullType(span),
				span,
			),
			Template: &ast.TemplateSpec{},
		}, true
	},
	{ModuleName: "scheduler", ValueName: "list_schedules"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
				span,
				ast.NewListType(
					ast.NewObjectType(
						[]ast.ObjectTypeField{
							ast.NewObjectTypeField(pAst.NewSpannedIdent("id", span), ast.NewIntType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("name", span), ast.NewStringType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("hour", span), ast.NewIntType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("minute", span), ast.NewIntType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("target_mode", span), ast.NewStringType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("hms_id", span), ast.NewOptionType(ast.NewStringType(span), span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("switches", span), ast.NewOptionType(ast.NewListType(
								ast.NewObjectType([]ast.ObjectTypeField{
									ast.NewObjectTypeField(pAst.NewSpannedIdent("switch", span), ast.NewStringType(span), span),
									ast.NewObjectTypeField(pAst.NewSpannedIdent("power", span), ast.NewBoolType(span), span),
								}, span), span,
							), span), span),
						},
						span,
					),
					span,
				),
				span,
			),
			Template: nil,
		}, true
	},
	{ModuleName: "notification", ValueName: "notify"}: func(_ types.ExecutionContext, span errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type: ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), nil),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("level", span), ast.NewIntType(span), nil),
				}),
				span,
				ast.NewIntType(span),
				span,
			),
			Template: nil,
		}, true
	},
	{ModuleName: "time", ValueName: "Time"}: func(_ types.ExecutionContext, _ errors.Span, _ pAst.IMPORT_KIND) (analyzer.BuiltinImport, bool) {
		return analyzer.BuiltinImport{
			Type:     TimeObjType(errors.Span{}),
			Template: nil,
		}, true
	},
}

var builtinImportModules = func() map[string]struct{} {
	modules := make(map[string]struct{}, len(builtinImportHandlers))
	for key := range builtinImportHandlers {
		if key.ModuleName == "notification" {
			continue
		}
		modules[key.ModuleName] = struct{}{}
	}
	return modules
}()

func GetImport(
	context types.ExecutionContext,
	moduleName string,
	valueName string,
	span errors.Span,
	kind pAst.IMPORT_KIND,
) (result analyzer.BuiltinImport, moduleFound bool, valueFound bool) {
	// TODO: differentiate between no such module and no such value?
	if kind == pAst.IMPORT_KIND_TEMPLATE {
		templ, found := driver.Templates(span)[types.ImportKey{
			ModuleName: moduleName,
			ValueName:  valueName,
		}]

		if !found {
			return analyzer.BuiltinImport{}, false, false
		}

		spec := templ.GetSpec()
		return analyzer.BuiltinImport{
			Type:     nil,
			Template: &spec,
		}, true, true
	}

	key := types.ImportKey{ModuleName: moduleName, ValueName: valueName}
	if handler, found := builtinImportHandlers[key]; found {
		result, valueFound := handler(context, span, kind)
		return result, true, valueFound
	}

	if _, found := builtinImportModules[moduleName]; found {
		return analyzer.BuiltinImport{}, true, false
	}

	return analyzer.BuiltinImport{}, false, false
}
