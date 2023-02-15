package homescript

import (
	"fmt"
	"strings"

	"github.com/smarthome-go/homescript/v2/homescript"
	"github.com/smarthome-go/homescript/v2/homescript/errors"
	hmsErrors "github.com/smarthome-go/homescript/v2/homescript/errors"
	"github.com/smarthome-go/smarthome/core/database"
)

func valPtr(input homescript.Value) *homescript.Value {
	return &input
}

func scopeAdditions() map[string]homescript.Value {
	return map[string]homescript.Value{
		"on_click_hms": homescript.ValueBuiltinFunction{
			Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
				if err := checkArgs("on_click_hms", span, args, homescript.TypeString, homescript.TypeString); err != nil {
					return nil, nil, err
				}

				targetCode := strings.ReplaceAll(args[0].(homescript.ValueString).Value, "'", "\\'")
				targetCode = strings.ReplaceAll(targetCode, "\"", "\\\"")
				inner := args[1].(homescript.ValueString).Value

				callBackCode := fmt.Sprintf("fetch('/api/homescript/run/live', {method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code: `%s`, args: [] }) })", targetCode)

				wrapper := fmt.Sprintf("<span onclick=\"%s\">%s</span>", callBackCode, inner)

				return homescript.ValueString{Value: wrapper, Range: span}, nil, nil
			},
		},
		"on_click_js": homescript.ValueBuiltinFunction{
			Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
				if err := checkArgs("on_click_js", span, args, homescript.TypeString, homescript.TypeString); err != nil {
					return nil, nil, err
				}

				targetCode := strings.ReplaceAll(args[0].(homescript.ValueString).Value, "\"", "\\\"")
				inner := args[1].(homescript.ValueString).Value

				wrapper := fmt.Sprintf("<span onclick=\"%s\">%s</span>", targetCode, inner)

				return homescript.ValueString{Value: wrapper, Range: span}, nil, nil
			},
		},
		"scheduler": homescript.ValueBuiltinVariable{
			Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
				return homescript.ValueObject{
					ObjFields: map[string]*homescript.Value{
						"new": valPtr(homescript.ValueBuiltinFunction{
							Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
								if err := checkArgs("new", span, args, homescript.TypeObject); err != nil {
									return nil, nil, err
								}

								obj := args[0].(homescript.ValueObject)
								if err := checkObj(span, obj, map[string]homescript.ValueType{
									"name":   homescript.TypeString,
									"hour":   homescript.TypeNumber,
									"minute": homescript.TypeNumber,
									"code":   homescript.TypeString,
								}); err != nil {
									return nil, nil, err
								}

								name := (*obj.Fields()["name"]).(homescript.ValueString)
								hour := (*obj.Fields()["hour"]).(homescript.ValueNumber)
								minute := (*obj.Fields()["minute"]).(homescript.ValueNumber)
								code := (*obj.Fields()["code"]).(homescript.ValueString)

								if err := checkInt(span, hour, "Field `hour`"); err != nil {
									return nil, nil, err
								}
								if err := checkInt(span, minute, "Field `minute`"); err != nil {
									return nil, nil, err
								}

								if executor.IsAnalyzer() {
									return homescript.ValueNull{}, nil, nil
								}

								if err := CreateNewSchedule(database.ScheduleData{
									Name:           name.Value,
									Hour:           uint(hour.Value),
									Minute:         uint(minute.Value),
									TargetMode:     database.ScheduleTargetModeCode,
									HomescriptCode: code.Value,
								}, executor.GetUser()); err != nil {
									return nil, nil, hmsErrors.NewError(span, err.Error(), errors.RuntimeError)
								}
								return homescript.ValueNull{}, nil, nil
							},
						}),
						"delete": valPtr(homescript.ValueBuiltinFunction{
							Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
								if err := checkArgs("delete", span, args, homescript.TypeNumber); err != nil {
									return nil, nil, err
								}

								id := args[0].(homescript.ValueNumber).Value

								if float64(int(id)) != id || id < 0.0 {
									return nil, nil, errors.NewError(span, fmt.Sprintf("Illegal value: ID needs to be a positive integer, got `%f`", id), errors.ValueError)
								}

								if executor.IsAnalyzer() {
									return homescript.ValueNull{}, nil, nil
								}

								if err := RemoveScheduleById(uint(id)); err != nil {
									return nil, nil, errors.NewError(span, err.Error(), errors.RuntimeError)
								}

								return homescript.ValueNull{}, nil, nil
							},
						}),
						"schedules": valPtr(homescript.ValueBuiltinFunction{
							Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
								if err := checkArgs("schedules", span, args); err != nil {
									return nil, nil, err
								}

								output := make([]*homescript.Value, 0)

								schedules, err := database.GetUserSchedules(executor.GetUser())
								if err != nil {
									return nil, nil, hmsErrors.NewError(span, err.Error(), errors.RuntimeError)
								}

								for _, schedule := range schedules {
									var hmsCode homescript.Value = homescript.ValueNull{}
									var hmsId homescript.Value = homescript.ValueNull{}
									var switchJobs homescript.Value = homescript.ValueNull{}

									switch schedule.Data.TargetMode {
									case database.ScheduleTargetModeCode:
										hmsCode = homescript.ValueString{Value: schedule.Data.HomescriptCode}
									case database.ScheduleTargetModeHMS:
										hmsId = homescript.ValueString{Value: schedule.Data.HomescriptTargetId}
									case database.ScheduleTargetModeSwitches:
										objType := homescript.TypeObject

										switches := make([]*homescript.Value, 0)
										for _, switchJob := range schedule.Data.SwitchJobs {
											switches = append(switches, valPtr(homescript.ValueObject{
												ObjFields: map[string]*homescript.Value{
													"id":    valPtr(homescript.ValueString{Value: switchJob.SwitchId}),
													"power": valPtr(homescript.ValueBool{Value: switchJob.PowerOn}),
												},
											}))
										}

										switchJobs = homescript.ValueList{Values: &switches, ValueType: &objType}
									}
									output = append(output, valPtr(homescript.ValueObject{
										DataType: "schedule",
										ObjFields: map[string]*homescript.Value{
											"id": valPtr(homescript.ValueNumber{Value: float64(schedule.Id)}),
											"name": valPtr(homescript.ValueString{
												Value: schedule.Data.Name,
											}),
											"target_mode": valPtr(homescript.ValueString{
												Value: string(schedule.Data.TargetMode),
											}),
											"hms_code": valPtr(hmsCode),
											"hms_id":   valPtr(hmsId),
											"switches": valPtr(switchJobs),
											"hour":     valPtr(homescript.ValueNumber{Value: float64(schedule.Data.Hour)}),
											"minute":   valPtr(homescript.ValueNumber{Value: float64(schedule.Data.Minute)}),
										},
									}))
								}

								type_ := homescript.TypeObject
								return homescript.ValueList{Values: &output, ValueType: &type_}, nil, nil
							},
						}),
					},
				}, nil
			},
		},
	}
}

// Helper function which checks the validity of args provided to builtin functions
func checkArgs(name string, span errors.Span, args []homescript.Value, types ...homescript.ValueType) *errors.Error {
	if len(args) != len(types) {
		s := ""
		if len(types) != 1 {
			s = "s"
		}
		return errors.NewError(
			span,
			fmt.Sprintf("function '%s' takes %d argument%s but %d were given", name, len(types), s, len(args)),
			errors.TypeError,
		)
	}
	for i, typ := range types {
		if args[i].Type() != typ {
			return errors.NewError(
				span,
				fmt.Sprintf("Argument %d of function '%s' has to be of type %v", i+1, name, typ),
				errors.TypeError,
			)
		}
	}
	return nil
}

// Helper function which checks that the passed object contains the correct keys
func checkObj(span errors.Span, obj homescript.ValueObject, check map[string]homescript.ValueType) *errors.Error {
	for key, type_ := range check {
		if obj.Fields()[key] == nil {
			return errors.NewError(span, fmt.Sprintf("Key `%s` of type `%v` not found in object", key, type_.String()), errors.TypeError)
		}

		if (*obj.Fields()[key]).Type() != type_ {
			return errors.NewError(span, fmt.Sprintf("Key `%s` has type `%v`, however `%v` was expected", key, (*obj.Fields()[key]).Type(), type_.String()), errors.TypeError)
		}
	}

	return nil
}

func checkInt(span errors.Span, num homescript.ValueNumber, errPrefix string) *errors.Error {
	if float64(int(num.Value)) != num.Value {
		return errors.NewError(span, fmt.Sprintf("%s: expected integer, found `%f`", errPrefix, num.Value), errors.ValueError)
	}
	return nil
}
