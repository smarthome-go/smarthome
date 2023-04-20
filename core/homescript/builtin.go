package homescript

import (
	"fmt"
	"strings"

	"github.com/smarthome-go/homescript/v2/homescript"
	"github.com/smarthome-go/homescript/v2/homescript/errors"
	hmsErrors "github.com/smarthome-go/homescript/v2/homescript/errors"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/automation"
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
		"system": homescript.ValueBuiltinVariable{
			Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
				return homescript.ValueObject{
					IsProtected: true,
					DataType:    "system",
					ObjFields: map[string]*homescript.Value{
						"scheduler_enabled": valPtr(homescript.ValueBuiltinVariable{Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
							serverConfig, found, err := database.GetServerConfiguration()
							if err != nil || !found {
								return nil, errors.NewError(span, "Could not retrieve system configuration", errors.RuntimeError)
							}
							return homescript.ValueBool{Value: serverConfig.AutomationEnabled}, nil
						}}),
						"lockdown_mode_enabled": valPtr(homescript.ValueBuiltinVariable{Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
							serverConfig, found, err := database.GetServerConfiguration()
							if err != nil || !found {
								return nil, errors.NewError(span, "Could not retrieve system configuration", errors.RuntimeError)
							}
							return homescript.ValueBool{Value: serverConfig.LockDownMode}, nil
						}}),
						"location": valPtr(homescript.ValueBuiltinVariable{Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
							serverConfig, found, err := database.GetServerConfiguration()
							if err != nil || !found {
								return nil, errors.NewError(span, "Could not retrieve system configuration", errors.RuntimeError)
							}
							return homescript.ValueObject{DataType: "location", ObjFields: map[string]*homescript.Value{
								"lat": valPtr(homescript.ValueNumber{Value: float64(serverConfig.Latitude)}),
								"lon": valPtr(homescript.ValueNumber{Value: float64(serverConfig.Longitude)}),
							}}, nil
						}}),
						"sun_times": valPtr(homescript.ValueBuiltinVariable{Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
							serverConfig, found, err := database.GetServerConfiguration()
							if err != nil || !found {
								return nil, errors.NewError(span, "Could not retrieve system configuration", errors.RuntimeError)
							}

							rise, set := automation.CalculateSunRiseSet(serverConfig.Latitude, serverConfig.Longitude)

							return homescript.ValueObject{DataType: "sun_times", ObjFields: map[string]*homescript.Value{
								"sunrise": valPtr(homescript.ValueObject{DataType: "time_simple", ObjFields: map[string]*homescript.Value{
									"hour":   valPtr(homescript.ValueNumber{Value: float64(rise.Hour)}),
									"minute": valPtr(homescript.ValueNumber{Value: float64(rise.Minute)}),
								}}),
								"sunset": valPtr(homescript.ValueObject{DataType: "time_simple", ObjFields: map[string]*homescript.Value{
									"hour":   valPtr(homescript.ValueNumber{Value: float64(set.Hour)}),
									"minute": valPtr(homescript.ValueNumber{Value: float64(set.Minute)}),
								}}),
							}}, nil
						}}),
						"hardware": valPtr(homescript.ValueBuiltinVariable{Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
							hasPermission, err := database.UserHasPermission(executor.GetUser(), database.PermissionSystemConfig)
							if err != nil {
								return nil, errors.NewError(span, err.Error(), errors.RuntimeError)
							}

							if !hasPermission {
								return nil, errors.NewError(span, fmt.Sprintf("Permission denied: you lack the permission `%s`", database.PermissionSystemConfig), errors.RuntimeError)
							}

							outList := make([]*homescript.Value, 0)
							typeObj := homescript.TypeObject

							if executor.IsAnalyzer() {
								return homescript.ValueList{ValueType: &typeObj, Values: &outList}, nil
							}

							nodes, err := database.GetHardwareNodes()
							if err != nil {
								return nil, errors.NewError(span, err.Error(), errors.RuntimeError)
							}

							for _, node := range nodes {
								nameCopy := node.Name
								urlCopy := node.Url
								tokenCopy := node.Token
								onlineCopy := node.Online
								enabledCopy := node.Enabled

								currObj := homescript.ValueObject{
									DataType:  "hw_node",
									IsDynamic: false,
									ObjFields: map[string]*homescript.Value{
										"name":    valPtr(homescript.ValueString{Value: nameCopy}),
										"online":  valPtr(homescript.ValueBool{Value: onlineCopy}),
										"enabled": valPtr(homescript.ValueBool{Value: enabledCopy}),
										"url":     valPtr(homescript.ValueString{Value: urlCopy}),
										"token":   valPtr(homescript.ValueString{Value: tokenCopy}),
										"set_enabled": valPtr(homescript.ValueBuiltinFunction{Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
											if err := checkArgs("set_enabled", span, args, homescript.TypeBoolean); err != nil {
												return nil, nil, err
											}

											shouldEnable := args[0].(homescript.ValueBool).Value

											fmt.Printf("setting %s to %t\n", urlCopy, shouldEnable)

											if err := database.ModifyHardwareNode(urlCopy, shouldEnable, nameCopy, tokenCopy); err != nil {
												return nil, nil, errors.NewError(span, err.Error(), errors.RuntimeError)
											}

											return homescript.ValueNull{}, nil, nil
										}}),
									},

									Range:       span,
									IsProtected: true,
								}

								outList = append(outList, valPtr(currObj))
							}

							return homescript.ValueList{
								Values:      &outList,
								ValueType:   &typeObj,
								Range:       span,
								IsProtected: true,
							}, nil
						}}),
					},
				}, nil
			},
		},
		"automation": homescript.ValueBuiltinVariable{
			Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
				return homescript.ValueObject{
					DataType: "automation",
					ObjFields: map[string]*homescript.Value{
						"new": valPtr(homescript.ValueBuiltinFunction{
							Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
								if err := checkArgs("new", span, args, homescript.TypeObject); err != nil {
									return nil, nil, err
								}

								obj := args[0].(homescript.ValueObject)
								valErr, stop := checkObj(span, obj, map[string]homescript.ValueType{
									"name":        homescript.TypeString,
									"description": homescript.TypeString,
									"hour":        homescript.TypeNumber,
									"minute":      homescript.TypeNumber,
									"hms_id":      homescript.TypeString,
									"days":        homescript.TypeList,
								}, executor)
								if valErr != nil {
									return nil, nil, valErr
								}
								if stop {
									return homescript.ValueNumber{Value: 0.0}, nil, nil
								}

								fields, fieldErr := obj.Fields(executor, span)
								if fieldErr != nil {
									return nil, nil, fieldErr
								}

								name := (*fields["name"]).(homescript.ValueString)
								description := (*fields["description"]).(homescript.ValueString)
								hour := (*fields["hour"]).(homescript.ValueNumber)
								minute := (*fields["minute"]).(homescript.ValueNumber)
								hmsId := (*fields["hms_id"]).(homescript.ValueString)
								days := (*fields["days"]).(homescript.ValueList)

								if err := checkInt(span, hour, "Field `hour`"); err != nil {
									return nil, nil, err
								}
								if err := checkInt(span, minute, "Field `minute`"); err != nil {
									return nil, nil, err
								}

								if hour.Value < 0.0 || hour.Value > 24.0 {
									return nil, nil, errors.NewError(span, "Hour must be => 0 and <= 24", errors.ValueError)
								}

								if minute.Value < 0.0 || minute.Value > 60.0 {
									return nil, nil, errors.NewError(span, "Minute must be => 0 and <= 60", errors.ValueError)
								}

								data, exists, err := database.GetUserHomescriptById(hmsId.Value, executor.GetUser())
								if err != nil {
									return nil, nil, errors.NewError(span, err.Error(), errors.RuntimeError)
								}

								if !exists {
									return nil, nil, errors.NewError(span, fmt.Sprintf("Homescript with ID `%s` does not exist", hmsId.Value), errors.ValueError)
								}

								if data.Data.SchedulerEnabled {
									return nil, nil, errors.NewError(span, fmt.Sprintf("Homescript with ID `%s` cannot be used as an automation / scheduler target", hmsId.Value), errors.ValueError)
								}

								if len(*days.Values) == 0 || len(*days.Values) > 7 {
									return nil, nil, errors.NewError(span, fmt.Sprintf("Invalid `days` list: expected >= 0 and <=7, got `%d`", len(*days.Values)), errors.ValueError)
								}

								// Check for duplicates and if each provided day is valid
								containsDays := make([]uint8, 0) // Contains the days, is used to check if there are duplicates in the days
								for idx, day := range *days.Values {
									if err := checkInt(span, (*day).(homescript.ValueNumber), fmt.Sprintf("Day at index `%d` invalid: ", idx)); err != nil {
										return nil, nil, err
									}

									dayInt := int((*day).(homescript.ValueNumber).Value)
									if dayInt > 6 {
										return nil, nil, errors.NewError(span, fmt.Sprintf("invalid day in `days`: day must be >= 0 and <= 6, found `%d`", day), errors.ValueError)
									}
									dayIsAlreadyPresend := false
									for _, dayTemp := range containsDays {
										if dayTemp == uint8(dayInt) {
											dayIsAlreadyPresend = true
										}
									}
									if dayIsAlreadyPresend {
										return nil, nil, errors.NewError(span, fmt.Sprintf("Duplicate entry in `days` list: `%d`", dayInt), errors.ValueError)
									}
									containsDays = append(containsDays, uint8(dayInt)) // If the day is not already present, add it
								}

								if executor.IsAnalyzer() {
									return homescript.ValueNumber{Value: 0.0}, nil, nil
								}

								id, err := CreateNewAutomation(name.Value, description.Value, uint8(hour.Value), uint8(minute.Value), containsDays, hmsId.Value, executor.GetUser(), true, database.TimingNormal)
								if err != nil {
									return nil, nil, hmsErrors.NewError(span, err.Error(), errors.RuntimeError)
								}
								return homescript.ValueNumber{Value: float64(id)}, nil, nil
							},
						}),
						"automations": valPtr(homescript.ValueBuiltinFunction{
							Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
								if err := checkArgs("automations", span, args); err != nil {
									return nil, nil, err
								}

								output := make([]*homescript.Value, 0)

								automations, err := database.GetUserAutomations(executor.GetUser())
								if err != nil {
									return nil, nil, hmsErrors.NewError(span, err.Error(), errors.RuntimeError)
								}

								for _, automationItem := range automations {
									cronData, err := automation.GetDataFromCronExpression(automationItem.Data.CronExpression)
									if err != nil {
										return nil, nil, hmsErrors.NewError(span, err.Error(), errors.RuntimeError)
									}

									numberType := homescript.TypeNumber
									days := make([]*homescript.Value, 0)
									for _, day := range cronData.Days {
										days = append(days, valPtr(homescript.ValueNumber{Value: float64(day)}))
									}

									output = append(output, valPtr(homescript.ValueObject{
										DataType: "automation",
										ObjFields: map[string]*homescript.Value{
											"id": valPtr(homescript.ValueNumber{Value: float64(automationItem.Id)}),
											"name": valPtr(homescript.ValueString{
												Value: automationItem.Data.Name,
											}),
											"timing_mode": valPtr(homescript.ValueString{
												Value: string(automationItem.Data.TimingMode),
											}),
											"hms_id": valPtr(homescript.ValueString{Value: automationItem.Data.HomescriptId}),
											"hour":   valPtr(homescript.ValueNumber{Value: float64(cronData.Hour)}),
											"minute": valPtr(homescript.ValueNumber{Value: float64(cronData.Minute)}),
											"days":   valPtr(homescript.ValueList{ValueType: &numberType, Values: &days}),
										},
									}))
								}

								type_ := homescript.TypeObject
								return homescript.ValueList{Values: &output, ValueType: &type_}, nil, nil
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

								automationData, found, err := database.GetAutomationById(uint(id))
								if err != nil {
									return nil, nil, errors.NewError(span, err.Error(), errors.RuntimeError)
								}

								if !found || automationData.Owner != executor.GetUser() {
									return nil, nil, errors.NewError(span, fmt.Sprintf("Automation with ID `%d` does not exist", int(id)), errors.ValueError)
								}

								if executor.IsAnalyzer() {
									return homescript.ValueNull{}, nil, nil
								}

								if err := RemoveAutomation(uint(id)); err != nil {
									return nil, nil, errors.NewError(span, err.Error(), errors.RuntimeError)
								}

								return homescript.ValueNull{}, nil, nil
							},
						}),
					},
				}, nil
			},
		},
		"scheduler": homescript.ValueBuiltinVariable{
			Callback: func(executor homescript.Executor, span hmsErrors.Span) (homescript.Value, *hmsErrors.Error) {
				return homescript.ValueObject{

					DataType: "scheduler",
					ObjFields: map[string]*homescript.Value{
						"new": valPtr(homescript.ValueBuiltinFunction{
							Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
								if err := checkArgs("new", span, args, homescript.TypeObject); err != nil {
									return nil, nil, err
								}

								obj := args[0].(homescript.ValueObject)
								valErr, stop := checkObj(span, obj, map[string]homescript.ValueType{
									"name":   homescript.TypeString,
									"hour":   homescript.TypeNumber,
									"minute": homescript.TypeNumber,
									"code":   homescript.TypeString,
								}, executor)
								if valErr != nil {
									return nil, nil, valErr
								}
								if stop {
									return homescript.ValueNumber{Value: 0.0}, nil, nil
								}

								fields, fieldErr := obj.Fields(executor, span)
								if fieldErr != nil {
									return nil, nil, fieldErr
								}

								name := (*fields["name"]).(homescript.ValueString)
								hour := (*fields["hour"]).(homescript.ValueNumber)
								minute := (*fields["minute"]).(homescript.ValueNumber)
								code := (*fields["code"]).(homescript.ValueString)

								if err := checkInt(span, hour, "Field `hour`"); err != nil {
									return nil, nil, err
								}
								if err := checkInt(span, minute, "Field `minute`"); err != nil {
									return nil, nil, err
								}

								if hour.Value < 0.0 || hour.Value > 24.0 {
									return nil, nil, errors.NewError(span, "Hour must be => 0 and <= 24", errors.ValueError)
								}

								if minute.Value < 0.0 || minute.Value > 60.0 {
									return nil, nil, errors.NewError(span, "Minute must be => 0 and <= 60", errors.ValueError)
								}

								if executor.IsAnalyzer() {
									return homescript.ValueNumber{Value: 0.0}, nil, nil
								}

								id, err := CreateNewSchedule(database.ScheduleData{
									Name:           name.Value,
									Hour:           uint(hour.Value),
									Minute:         uint(minute.Value),
									TargetMode:     database.ScheduleTargetModeCode,
									HomescriptCode: code.Value,
								}, executor.GetUser())
								if err != nil {
									return nil, nil, hmsErrors.NewError(span, err.Error(), errors.RuntimeError)
								}
								return homescript.ValueNumber{Value: float64(id)}, nil, nil
							},
						}),
						"modify": valPtr(homescript.ValueBuiltinFunction{
							Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
								if err := checkArgs("modify", span, args, homescript.TypeNumber, homescript.TypeObject); err != nil {
									return nil, nil, err
								}

								obj := args[1].(homescript.ValueObject)
								valErr, stop := checkObj(span, obj, map[string]homescript.ValueType{
									"name":   homescript.TypeString,
									"hour":   homescript.TypeNumber,
									"minute": homescript.TypeNumber,
									"code":   homescript.TypeString,
								}, executor)
								if valErr != nil {
									return nil, nil, valErr
								}
								if stop {
									return homescript.ValueNull{}, nil, nil
								}

								fields, fieldErr := obj.Fields(executor, span)
								if fieldErr != nil {
									return nil, nil, fieldErr
								}

								name := (*fields["name"]).(homescript.ValueString)
								hour := (*fields["hour"]).(homescript.ValueNumber)
								minute := (*fields["minute"]).(homescript.ValueNumber)
								code := (*fields["code"]).(homescript.ValueString)

								if err := checkInt(span, hour, "Field `hour`"); err != nil {
									return nil, nil, err
								}
								if err := checkInt(span, minute, "Field `minute`"); err != nil {
									return nil, nil, err
								}

								if hour.Value < 0.0 || hour.Value > 24.0 {
									return nil, nil, errors.NewError(span, "Hour must be => 0 and <= 24", errors.ValueError)
								}
								if minute.Value < 0.0 || minute.Value > 60.0 {
									return nil, nil, errors.NewError(span, "Minute must be => 0 and <= 60", errors.ValueError)
								}

								id := args[0].(homescript.ValueNumber)
								if err := checkInt(span, id, "argument `id` is not an integer"); err != nil {
									return nil, nil, err
								}

								if executor.IsAnalyzer() {
									return homescript.ValueNull{}, nil, nil
								}

								schedulerData, found, err := database.GetScheduleById(uint(id.Value))
								if err != nil {
									return nil, nil, errors.NewError(span, err.Error(), errors.RuntimeError)
								}

								if !found || schedulerData.Owner != executor.GetUser() {
									return nil, nil, errors.NewError(span, fmt.Sprintf("Schedule with ID `%d` does not exist", int(id.Value)), errors.ValueError)
								}

								if err := ModifyScheduleById(uint(id.Value), database.ScheduleData{
									Name:           name.Value,
									Hour:           uint(hour.Value),
									Minute:         uint(minute.Value),
									TargetMode:     database.ScheduleTargetModeCode,
									HomescriptCode: code.Value,
								}); err != nil {
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

								schedulerData, found, err := database.GetScheduleById(uint(id))
								if err != nil {
									return nil, nil, errors.NewError(span, err.Error(), errors.RuntimeError)
								}

								if !found || schedulerData.Owner != executor.GetUser() {
									return nil, nil, errors.NewError(span, fmt.Sprintf("Schedule with ID `%d` does not exist", int(id)), errors.ValueError)
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
func checkObj(span errors.Span, obj homescript.ValueObject, check map[string]homescript.ValueType, executor homescript.Executor) (*errors.Error, bool) {
	fields, err := obj.Fields(executor, span)
	if err != nil {
		return err, false
	}

	for key, type_ := range check {

		if fields[key] == nil {
			return errors.NewError(span, fmt.Sprintf("Key `%s` of type `%v` not found in object", key, type_.String()), errors.TypeError), false
		}

		if (*fields[key]) == nil {
			return nil, true
		}

		if (*fields[key]).Type() != type_ {
			return errors.NewError(span, fmt.Sprintf("Key `%s` has type `%v`, however `%v` was expected", key, (*fields[key]).Type(), type_.String()), errors.TypeError), false
		}
	}

	return nil, false
}

func checkInt(span errors.Span, num homescript.ValueNumber, errPrefix string) *errors.Error {
	if float64(int(num.Value)) != num.Value {
		return errors.NewError(span, fmt.Sprintf("%s: expected integer, found `%f`", errPrefix, num.Value), errors.ValueError)
	}
	return nil
}
