package homescript

import (
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
	"github.com/smarthome-go/smarthome/core/database"
)

type analyzerHost struct {
	username string
}

func newAnalyzerHost(username string) analyzerHost {
	return analyzerHost{
		username: username,
	}
}

func (self analyzerHost) GetBuiltinImport(moduleName string, valueName string, span errors.Span) (valueType ast.Type, moduleFound bool, valueFound bool) {
	switch moduleName {
	case "hms":
		switch valueName {
		case "exec":
			return ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("script_id", span), ast.NewStringType(span)),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("arguments", span), ast.NewOptionType(ast.NewAnyObjectType(span), span)),
				}),
				span,
				ast.NewNullType(span),
				span,
			), true, true
		}
		return nil, true, false
	case "location":
		switch valueName {
		case "sun_times":
			timeObjType := func(span errors.Span) ast.Type {
				return ast.NewObjectType([]ast.ObjectTypeField{
					ast.NewObjectTypeField(pAst.NewSpannedIdent("hour", span), ast.NewIntType(span), span),
					ast.NewObjectTypeField(pAst.NewSpannedIdent("minute", span), ast.NewIntType(span), span),
				}, span)
			}

			return ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
				span,
				ast.NewObjectType([]ast.ObjectTypeField{
					ast.NewObjectTypeField(pAst.NewSpannedIdent("sunrise", span), timeObjType(span), span),
					ast.NewObjectTypeField(pAst.NewSpannedIdent("sunset", span), timeObjType(span), span),
				}, span),
				span,
			), true, true
		case "weather":
			return ast.NewFunctionType(
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
			), true, true
		}
		return nil, true, false
	case "switch":
		switch valueName {
		case "get_switch":
			return ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("id", span), ast.NewStringType(span)),
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
			), true, true
		case "power":
			return ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("switch_id", span), ast.NewStringType(span)),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("power", span), ast.NewBoolType(span)),
				}),
				span,
				ast.NewNullType(span),
				span,
			), true, true
		default:
			return nil, true, false
		}
	case "widget":
		switch valueName {
		case "on_click_js", "on_click_hms":
			return ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("base", span), ast.NewStringType(span)),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("js", span), ast.NewStringType(span)),
				}),
				span, ast.NewStringType(span), span,
			), true, true
		default:
			return nil, true, false
		}
	case "testing":
		switch valueName {
		case "assert_eq":
			return ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("lhs", errors.Span{}), ast.NewUnknownType()),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("rhs", errors.Span{}), ast.NewUnknownType()),
				}),
				errors.Span{},
				ast.NewNullType(errors.Span{}),
				errors.Span{},
			), true, true
		default:
			return nil, true, false
		}
	case "storage":
		switch valueName {
		case "set_storage":
			return ast.NewFunctionType(ast.NewNormalFunctionTypeParamKind(
				[]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("key", span), ast.NewStringType(span)),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("value", span), ast.NewUnknownType()),
				},
			),
				span,
				ast.NewNullType(span),
				span,
			), true, true
		case "get_storage":
			return ast.NewFunctionType(ast.NewNormalFunctionTypeParamKind(
				[]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("key", span), ast.NewStringType(span)),
				},
			),
				span,
				ast.NewOptionType(ast.NewStringType(span), span),
				span,
			), true, true
		default:
			return nil, true, false
		}
	case "reminder":
		switch valueName {
		case "remind":
			return ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("reminder", span),
						ast.NewObjectType([]ast.ObjectTypeField{
							ast.NewObjectTypeField(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("priority", span), ast.NewIntType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("due_date_day", span), ast.NewIntType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("due_date_month", span), ast.NewIntType(span), span),
							ast.NewObjectTypeField(pAst.NewSpannedIdent("due_date_year", span), ast.NewIntType(span), span),
						}, span)),
				}),
				span,
				ast.NewIntType(span),
				span,
			), true, true
		default:
			return nil, true, false
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
			return ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("ip", span), ast.NewStringType(span)),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("timeout", span), ast.NewFloatType(span)),
				}),
				span,
				ast.NewBoolType(span),
				span,
			), true, true
		case "HttpResponse":
			return newHttpResponse(), true, true
		case "http":
			return ast.NewObjectType([]ast.ObjectTypeField{
				ast.NewObjectTypeField(pAst.NewSpannedIdent("get", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{ast.NewFunctionTypeParam(pAst.NewSpannedIdent("url", span), ast.NewStringType(span))}),
					span,
					newHttpResponse(),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("generic", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind(
						[]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("url", span), ast.NewStringType(span)),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("method", span), ast.NewStringType(span)),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("body", span), ast.NewOptionType(ast.NewStringType(span), span)),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("headers", span), ast.NewAnyObjectType(span)),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("cookies", span), ast.NewAnyObjectType(span)),
						},
					),
					span,
					newHttpResponse(),
					span,
				), span),
			}, span), true, true
		default:
			return nil, true, false
		}
	case "log":
		switch valueName {
		case "logger":
			return ast.NewObjectType([]ast.ObjectTypeField{
				ast.NewObjectTypeField(pAst.NewSpannedIdent("trace", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span)),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span)),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("debug", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span)),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span)),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("info", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span)),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span)),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("warn", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span)),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span)),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("error", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span)),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span)),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("fatal", span), ast.NewFunctionType(
					ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span)),
						ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span)),
					}),
					span,
					ast.NewNullType(span),
					span,
				), span),
			}, span), true, true
		default:
			return nil, true, false
		}
	case "context":
		switch valueName {
		case "args":
			return ast.NewAnyObjectType(span), true, true
		case "notification":
			return ast.NewObjectType([]ast.ObjectTypeField{
				ast.NewObjectTypeField(pAst.NewSpannedIdent("id", span), ast.NewIntType(span), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("title", span), ast.NewStringType(span), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("description", span), ast.NewStringType(span), span),
				ast.NewObjectTypeField(pAst.NewSpannedIdent("level", span), ast.NewIntType(span), span),
			}, span), true, true
		}
		return nil, true, false
	case "scheduler":
		switch valueName {
		case "create_schedule":
			return ast.NewFunctionType(
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
					),
				}),
				span,
				ast.NewIntType(span),
				span,
			), true, true

		case "delete_schedule":
			return ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("id", span), ast.NewIntType(span)),
				}),
				span,
				ast.NewNullType(span),
				span,
			), true, true
		case "list_schedules":
			return ast.NewFunctionType(
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
			), true, true
		}
		return nil, true, false
	case "notification":
		switch valueName {
		case "notify":
			return ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("title", span), ast.NewStringType(span)),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("description", span), ast.NewStringType(span)),
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("level", span), ast.NewIntType(span)),
				}),
				span,
				ast.NewIntType(span),
				span,
			), true, true
		}
	}
	return nil, false, false
}

func (self analyzerHost) ResolveCodeModule(moduleName string) (code string, moduleFound bool, err error) {
	log.Trace(fmt.Sprintf("Resolving module `%s` by user `%s`", moduleName, self.username))
	script, found, err := database.GetUserHomescriptById(moduleName, self.username)
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
					ast.NewFunctionTypeParam(pAst.NewSpannedIdent("code", errors.Span{}), ast.NewIntType(errors.Span{})),
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
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("seconds", errors.Span{}), ast.NewFloatType(errors.Span{})),
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
							)}),
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
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("time", errors.Span{}), timeObjType(errors.Span{})),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("days", errors.Span{}), ast.NewIntType(errors.Span{})),
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
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("time", errors.Span{}), timeObjType(errors.Span{})),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("hours", errors.Span{}), ast.NewIntType(errors.Span{})),
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
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("time", errors.Span{}), timeObjType(errors.Span{})),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("hours", errors.Span{}), ast.NewIntType(errors.Span{})),
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
