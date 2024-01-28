package homescript

import (
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
)

// A list of `known` object type annotations
// The analyzer uses these in order to sanity-check every annotation
var knownObjectTypeAnnotations = []string{DRIVER_FIELD_REQUIRED_ANNOTATION}

type HMS_PROGRAM_KIND uint8

const (
	HMS_PROGRAM_KIND_NORMAL HMS_PROGRAM_KIND = iota
	HMS_PROGRAM_KIND_DEVICE_DRIVER
)

type AnalyzerDriverMetadata struct {
	VendorID string
	ModelID  string
}

type analyzerHost struct {
	username string
	// Depending on the program kind, the post-validation hook performs specific validations.
	programKind HMS_PROGRAM_KIND

	driverData *AnalyzerDriverMetadata
}

func (analyzerHost) GetKnownObjectTypeFieldAnnotations() []string {
	return knownObjectTypeAnnotations
}

func newAnalyzerHost(
	username string,
	programKind HMS_PROGRAM_KIND,
	driverData *AnalyzerDriverMetadata,
) analyzerHost {
	return analyzerHost{
		username:    username,
		programKind: programKind,
		driverData:  driverData,
	}
}

func (self analyzerHost) PostValidationHook(
	analyzedModules map[string]ast.AnalyzedProgram,
	mainModule string,
	analyzer *analyzer.Analyzer,
) []diagnostic.Diagnostic {
	switch self.programKind {
	case HMS_PROGRAM_KIND_DEVICE_DRIVER:
		// fmt.Printf("post validation driver hook: %s: %s\n", self.driverData.VendorId, self.driverData.ModelId)

		info, diagnostics := ExtractDriverInfo(analyzedModules, mainModule, true)
		fmt.Printf("post-validation: INFO: %v\n", info)
		return diagnostics
	case HMS_PROGRAM_KIND_NORMAL:
		// TODO: is there something to implement?
	}

	return nil
}

func (self analyzerHost) GetBuiltinImport(moduleName string, valueName string, span errors.Span, kind pAst.IMPORT_KIND) (result analyzer.BuiltinImport, moduleFound bool, valueFound bool) {
	// TODO: handle import kind

	switch moduleName {
	case "driver":
		switch kind {
		case pAst.IMPORT_KIND_TEMPLATE:
			switch valueName {
			case "Driver":
				return analyzer.BuiltinImport{
					Type: nil,
					Template: &ast.TemplateSpec{
						BaseMethods: map[string]ast.TemplateMethod{
							"validate_driver": {
								Signature: ast.NewFunctionType(
									ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)), span, ast.NewNullType(span), span,
								).(ast.FunctionType),
								Modifier: pAst.FN_MODIFIER_PUB,
							},
						},
						Capabilities: map[string]ast.TemplateCapability{
							"base": {
								RequiresMethods:           []string{"validate_driver"},
								ConflictsWithCapabilities: []ast.TemplateConflict{},
							},
						},
						DefaultCapabilities: []string{"base"},
						Span:                span,
					},
				}, true, true
			case "Device":
				return analyzer.BuiltinImport{
					Type: nil,
					Template: &ast.TemplateSpec{
						BaseMethods: map[string]ast.TemplateMethod{
							"validate_device": {
								Signature: ast.NewFunctionType(
									ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)), span, ast.NewNullType(span), span,
								).(ast.FunctionType),
								Modifier: pAst.FN_MODIFIER_PUB,
							},
							"dim": {
								Signature: ast.NewFunctionType(
									ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
										ast.NewFunctionTypeParam(
											pAst.NewSpannedIdent("percent", span),
											ast.NewIntType(span),
											nil,
										),
									}), span, ast.NewNullType(span), span,
								).(ast.FunctionType),
								Modifier: pAst.FN_MODIFIER_PUB,
							},
						},
						Capabilities: map[string]ast.TemplateCapability{
							"base": {
								RequiresMethods:           []string{"validate_device"},
								ConflictsWithCapabilities: []ast.TemplateConflict{},
							},
							"dimmable": {
								RequiresMethods:           []string{"dim"},
								ConflictsWithCapabilities: []ast.TemplateConflict{},
							},
						},
						DefaultCapabilities: []string{"base"},
						Span:                span,
					},
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
			default:
				return analyzer.BuiltinImport{}, true, false
			}
		case pAst.IMPORT_KIND_NORMAL:
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
	case "switch":
		switch valueName {
		case "get_switch":
			return analyzer.BuiltinImport{
				Type: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("id", span), ast.NewStringType(span), nil),
					}),
					span,
					ast.NewOptionType(
						ast.NewObjectType(
							[]ast.ObjectTypeField{
								ast.NewObjectTypeField(pAst.NewSpannedIdent("name", span), ast.NewStringType(span), span),
								ast.NewObjectTypeField(pAst.NewSpannedIdent("room_id", span), ast.NewStringType(span), span),
								ast.NewObjectTypeField(pAst.NewSpannedIdent("power", span), ast.NewBoolType(span), span),
								ast.NewObjectTypeField(pAst.NewSpannedIdent("watts", span), ast.NewIntType(span), span),
								ast.NewObjectTypeField(pAst.NewSpannedIdent("target_node", span), ast.NewOptionType(ast.NewStringType(span), span), span),
							},
							span),
						span),
					span,
				),
				Template: &ast.TemplateSpec{},
			}, true, true
		case "power":
			return analyzer.BuiltinImport{
				Type: ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("switch_id", span), ast.NewStringType(span), nil),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("power", span), ast.NewBoolType(span), nil),
					}),
					span,
					ast.NewNullType(span),
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
				Template: &ast.TemplateSpec{},
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
				Template: &ast.TemplateSpec{},
			}, true, true
		}
	}
	return analyzer.BuiltinImport{}, false, false
}

func (self analyzerHost) ResolveCodeModule(moduleName string) (code string, moduleFound bool, err error) {
	log.Trace(fmt.Sprintf("Resolving module `%s` by user `%s`", moduleName, self.username))
	script, found, err := GetPersonalScriptById(moduleName, self.username)
	if err != nil || !found {
		return "", found, err
	}
	return script.Data.Code, true, nil
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
