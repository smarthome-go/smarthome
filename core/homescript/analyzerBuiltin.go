package homescript

import (
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
	"github.com/smarthome-go/smarthome/core/device/driver"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

// A list of `known` object type annotations.
// The analyzer uses these in order to sanity-check every annotation.
var knownObjectTypeAnnotations = []string{driver.DRIVER_FIELD_REQUIRED_ANNOTATION}

type analyzerHost struct {
	context types.ExecutionContext
}

func (analyzerHost) GetKnownObjectTypeFieldAnnotations() []string {
	return knownObjectTypeAnnotations
}

func newAnalyzerHost(
	context types.ExecutionContext,
) analyzerHost {
	if context == nil {
		panic("Context cannot be <nil>")
	}

	return analyzerHost{
		context: context,
	}
}

func (self analyzerHost) PostValidationHook(
	analyzedModules map[string]ast.AnalyzedProgram,
	mainModule string,
	analyzer *analyzer.Analyzer,
) []diagnostic.Diagnostic {
	switch self.context.Kind() {
	case types.HMS_PROGRAM_KIND_DEVICE_DRIVER:
		_, diagnostics := driver.ExtractDriverInfo(analyzedModules, mainModule, true)
		return diagnostics
	default:
		// Forbid `trigger` annotations and singletons in non-driver code.
		diagnostics := make([]diagnostic.Diagnostic, 0)
		for _, mod := range analyzedModules {
			for _, singleton := range mod.Singletons {
				diagnostics = append(diagnostics, diagnostic.Diagnostic{
					Level:   diagnostic.DiagnosticLevelError,
					Message: "Singletons are not allowed in user programs",
					Notes:   []string{"To use singletons, create a driver Homescript"},
					Span:    singleton.Span(),
				})
			}

			for _, fn := range mod.Functions {
				if fn.Annotation == nil {
					continue
				}

				for _, ann := range fn.Annotation.Items {
					switch ann.(type) {
					case ast.AnalyzedAnnotationItemTrigger:
						diagnostics = append(diagnostics,
							diagnostic.Diagnostic{
								Level:   diagnostic.DiagnosticLevelError,
								Message: "Trigger annotations are not allowed in user programs",
								Notes: []string{
									"To use trigger annotations, create a driver Homescript",
								},
								Span: ann.Span(),
							},
						)
					default:
					}
				}
			}
		}
		return diagnostics
	}
}

func mqttCallbackFn(span errors.Span) ast.Type {
	return ast.NewFunctionType(
		ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
			ast.NewFunctionTypeParam(pAst.NewSpannedIdent("topic", span), ast.NewStringType(span), nil),
			ast.NewFunctionTypeParam(pAst.NewSpannedIdent("payload", span), ast.NewStringType(span), nil),
		}),
		span,
		ast.NewNullType(span),
		span,
	)
}

func (self analyzerHost) GetBuiltinImport(
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

	switch moduleName {
	case "triggers":
		if kind != pAst.IMPORT_KIND_TRIGGER {
			return analyzer.BuiltinImport{}, true, false
		}

		switch valueName {
		case "minute":
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
			}, true, true
		default:
			return analyzer.BuiltinImport{}, true, false
		}
	case "driver":
		switch kind {
		case pAst.IMPORT_KIND_TEMPLATE:
			switch valueName {
			case "Driver":
				return analyzer.BuiltinImport{
					Type:     nil,
					Template: nil,
				}, true, true
			case "Device":
				return analyzer.BuiltinImport{
					Type:     nil,
					Template: &ast.TemplateSpec{},
				}, true, true
			default:
				return analyzer.BuiltinImport{}, true, false
			}
		case pAst.IMPORT_KIND_TYPE:
			switch valueName {
			case "DriverMeta":
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
				}, true, true
			case "Dimmable":
				return analyzer.BuiltinImport{
					Type:     driver.ReportDimType(span),
					Template: nil,
				}, true, true
			case "Sensor":
				return analyzer.BuiltinImport{
					Type:     driver.ReportSensorReadingType(span),
					Template: nil,
				}, true, true
			default:
				return analyzer.BuiltinImport{}, true, false
			}
		case pAst.IMPORT_KIND_NORMAL:
			return analyzer.BuiltinImport{}, true, false
		}
	case "mqtt":
		if kind == pAst.IMPORT_KIND_TRIGGER {
			switch valueName {
			case "message":
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
				}, true, true
			}

			return analyzer.BuiltinImport{}, true, false
		}

		switch valueName {
		case "subscribe":
			return analyzer.BuiltinImport{
				Type: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("topics", span), ast.NewListType(ast.NewStringType(span), span), nil),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("callback", span), ast.NewFunctionType(
							ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{}),
							span,
							ast.NewNullType(span),
							span,
						), nil),
					}),
					span,
					ast.NewNullType(span),
					span,
				),
				Template: nil,
			}, true, true
		case "publish":
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
			}, true, true
		default:
			return analyzer.BuiltinImport{}, true, false
		}
	case "hms":
		switch valueName {
		case "exec":
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
			}, true, true
		case "exec_user":
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
			}, true, true
		}
		return analyzer.BuiltinImport{}, true, false
	case "location":
		switch valueName {
		case "sun_times":
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
			}, true, true
		case "weather":
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
			}, true, true
		}
		return analyzer.BuiltinImport{}, true, false
	case "device":
		switch valueName {
		// TODO: port to device
		// case "get_switch":
		// 	return analyzer.BuiltinImport{
		// 		Type: ast.NewFunctionType(
		// 			ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
		// 				ast.NewFunctionTypeParam(pAst.NewSpannedIdent("id", span), ast.NewStringType(span), nil),
		// 			}),
		// 			span,
		// 			ast.NewOptionType(
		// 				ast.NewObjectType(
		// 					[]ast.ObjectTypeField{
		// 						ast.NewObjectTypeField(pAst.NewSpannedIdent("name", span), ast.NewStringType(span), span),
		// 						ast.NewObjectTypeField(pAst.NewSpannedIdent("room_id", span), ast.NewStringType(span), span),
		// 						ast.NewObjectTypeField(pAst.NewSpannedIdent("power", span), ast.NewBoolType(span), span),
		// 						ast.NewObjectTypeField(pAst.NewSpannedIdent("watts", span), ast.NewIntType(span), span),
		// 						ast.NewObjectTypeField(pAst.NewSpannedIdent("target_node", span), ast.NewOptionType(ast.NewStringType(span), span), span),
		// 					},
		// 					span),
		// 				span),
		// 			span,
		// 		),
		// 		Template: &ast.TemplateSpec{},
		// 	}, true, true
		case "set_power":
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
			}, true, true
		case "dim":
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
			}, true, true
		default:
			return analyzer.BuiltinImport{}, true, false
		}
	case "widget":
		switch valueName {
		case "on_click_js", "on_click_hms":
			return analyzer.BuiltinImport{
				Type: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("base", span), ast.NewStringType(span), nil),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("js", span), ast.NewStringType(span), nil),
					}),
					span, ast.NewStringType(span), span,
				),
				Template: &ast.TemplateSpec{},
			}, true, true
		default:
			return analyzer.BuiltinImport{}, true, false
		}
	case "testing":
		switch valueName {
		case "assert_eq":
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
			}, true, true
		default:
			return analyzer.BuiltinImport{}, true, false
		}
	case "storage":
		switch valueName {
		case "set_storage":
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
			}, true, true
		case "get_storage":
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
			}, true, true
		default:
			return analyzer.BuiltinImport{}, true, false
		}
	case "reminder":
		switch valueName {
		case "remind":
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
			}, true, true
		default:
			return analyzer.BuiltinImport{}, true, false
		}
	case "net":
		newHttpResponse := func() ast.Type {
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

		switch valueName {
		case "ping":
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
			}, true, true
		case "HttpResponse":
			return analyzer.BuiltinImport{
				Type:     newHttpResponse(),
				Template: &ast.TemplateSpec{},
			}, true, true
		case "http":
			return analyzer.BuiltinImport{
				Type: ast.NewObjectType([]ast.ObjectTypeField{
					ast.NewObjectTypeField(pAst.NewSpannedIdent("get", span), ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{ast.NewFunctionTypeParam(pAst.NewSpannedIdent("url", span), ast.NewStringType(span), nil)}),
						span,
						newHttpResponse(),
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
						newHttpResponse(),
						span,
					), span),
				}, span),
				Template: &ast.TemplateSpec{},
			}, true, true
		default:
			return analyzer.BuiltinImport{}, true, false
		}
	case "log":
		switch valueName {
		case "logger":
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
			}, true, true
		default:
			return analyzer.BuiltinImport{}, true, false
		}
	case "context":
		switch valueName {
		case "args":
			return analyzer.BuiltinImport{
				Type:     ast.NewAnyObjectType(span),
				Template: &ast.TemplateSpec{},
			}, true, true
		case "notification":
			return analyzer.BuiltinImport{
				Type: ast.NewObjectType([]ast.ObjectTypeField{
					ast.NewObjectTypeField(pAst.NewSpannedIdent("id", span), ast.NewIntType(span), span),
					ast.NewObjectTypeField(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), span),
					ast.NewObjectTypeField(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), span),
					ast.NewObjectTypeField(pAst.NewSpannedIdent("level", span), ast.NewIntType(span), span),
				}, span),
				Template: &ast.TemplateSpec{},
			}, true, true
		}
		return analyzer.BuiltinImport{}, true, false
	case "scheduler":
		switch valueName {
		case "create_schedule":
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
			}, true, true
		case "delete_schedule":
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
			}, true, true
		case "list_schedules":
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
			}, true, true
		}
		return analyzer.BuiltinImport{}, true, false
	case "notification":
		switch valueName {
		case "notify":
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
			}, true, true
		}
	case "time":
		switch valueName {
		case "Time":
			return analyzer.BuiltinImport{
				Type:     timeObjType(errors.Span{}),
				Template: nil,
			}, true, true
		}

		return analyzer.BuiltinImport{}, true, false
	}
	return analyzer.BuiltinImport{}, false, false
}

func (self analyzerHost) ResolveCodeModule(moduleName string) (code string, moduleFound bool, err error) {
	logger.Trace(fmt.Sprintf("Resolving module `%s`", moduleName))

	if self.context.Username() != nil {
		script, found, err := HmsManager.GetPersonalScriptById(moduleName, *self.context.Username())
		if err != nil || !found {
			return "", found, err
		}
		return script.Data.Code, true, nil
	} else {
		script, found, err := HmsManager.GetScriptById(moduleName)
		if err != nil || !found {
			return "", found, err
		}
		return script.Data.Code, true, nil
	}
}

// TODO: fill this
func analyzerScopeAdditions() map[string]analyzer.Variable {
	return map[string]analyzer.Variable{
		"exit": analyzer.NewBuiltinVar(
			ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("code", errors.Span{}), ast.NewIntType(errors.Span{}), nil),
				}),
				errors.Span{},
				ast.NewNeverType(),
				errors.Span{},
			),
		),
		"fmt": analyzer.NewBuiltinVar(
			ast.NewFunctionType(
				ast.NewVarArgsFunctionTypeParamKind([]ast.Type{ast.NewStringType(errors.Span{})}, ast.NewUnknownType()),
				errors.Span{},
				ast.NewStringType(errors.Span{}),
				errors.Span{},
			),
		),
		"print": analyzer.NewBuiltinVar(
			ast.NewFunctionType(
				ast.NewVarArgsFunctionTypeParamKind([]ast.Type{}, ast.NewUnknownType()),
				errors.Span{},
				ast.NewNullType(errors.Span{}),
				errors.Span{},
			),
		),
		"println": analyzer.NewBuiltinVar(
			ast.NewFunctionType(
				ast.NewVarArgsFunctionTypeParamKind([]ast.Type{}, ast.NewUnknownType()),
				errors.Span{},
				ast.NewNullType(errors.Span{}),
				errors.Span{},
			),
		),
		"time": analyzer.NewBuiltinVar(ast.NewObjectType(
			[]ast.ObjectTypeField{
				ast.NewObjectTypeField(
					pAst.NewSpannedIdent("sleep", errors.Span{}),
					ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("seconds", errors.Span{}), ast.NewFloatType(errors.Span{}), nil),
						}),
						errors.Span{},
						ast.NewNullType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
				ast.NewObjectTypeField(
					pAst.NewSpannedIdent("since", errors.Span{}),
					ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("when", errors.Span{}),
								timeObjType(errors.Span{}),
								nil)}),
						errors.Span{},
						durationObjType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
				ast.NewObjectTypeField(
					pAst.NewSpannedIdent("now", errors.Span{}),
					ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
						errors.Span{},
						timeObjType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
				ast.NewObjectTypeField(
					pAst.NewSpannedIdent("add_days", errors.Span{}),
					ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("time", errors.Span{}), timeObjType(errors.Span{}), nil),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("days", errors.Span{}), ast.NewIntType(errors.Span{}), nil),
						}),
						errors.Span{},
						timeObjType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
				ast.NewObjectTypeField(
					pAst.NewSpannedIdent("add_hours", errors.Span{}),
					ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("time", errors.Span{}), timeObjType(errors.Span{}), nil),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("hours", errors.Span{}), ast.NewIntType(errors.Span{}), nil),
						}),
						errors.Span{},
						timeObjType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
				ast.NewObjectTypeField(
					pAst.NewSpannedIdent("add_minutes", errors.Span{}),
					ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("time", errors.Span{}), timeObjType(errors.Span{}), nil),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("hours", errors.Span{}), ast.NewIntType(errors.Span{}), nil),
						}),
						errors.Span{},
						timeObjType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
			},
			errors.Span{},
		)),
	}
}

func durationObjType(span errors.Span) ast.Type {
	return ast.NewObjectType([]ast.ObjectTypeField{
		ast.NewObjectTypeField(pAst.NewSpannedIdent("hours", span), ast.NewFloatType(span), span),
		ast.NewObjectTypeField(pAst.NewSpannedIdent("minutes", span), ast.NewFloatType(span), span),
		ast.NewObjectTypeField(pAst.NewSpannedIdent("seconds", span), ast.NewFloatType(span), span),
		ast.NewObjectTypeField(pAst.NewSpannedIdent("millis", span), ast.NewIntType(span), span),
		ast.NewObjectTypeField(pAst.NewSpannedIdent("display", span), ast.NewFunctionType(
			ast.NormalFunctionTypeParamKindIdentifier{}, span, ast.NewStringType(span), span,
		), span),
	}, span)
}

func timeObjType(span errors.Span) ast.Type {
	return ast.NewObjectType(
		[]ast.ObjectTypeField{
			ast.NewObjectTypeField(pAst.NewSpannedIdent("year", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("month", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("year_day", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("hour", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("minute", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("second", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("month_day", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("week_day", span), ast.NewIntType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("week_day_text", span), ast.NewStringType(span), span),
			ast.NewObjectTypeField(pAst.NewSpannedIdent("unix_milli", span), ast.NewIntType(span), span),
		},
		span,
	)
}
