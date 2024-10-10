package analyzer

import (
	"github.com/smarthome-go/homescript/v3/homescript/analyzer"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
)

const PrintfnBuiltinIdent = "printfn"
const InputBuiltinIdent = "input"
const PollInputBuiltinIdent = "poll_input"

func AnalyzerScopeAdditions() map[string]analyzer.Variable {
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
		PrintfnBuiltinIdent: analyzer.NewBuiltinVar(
			ast.NewFunctionType(
				ast.NewVarArgsFunctionTypeParamKind([]ast.Type{}, ast.NewUnknownType()),
				errors.Span{},
				ast.NewNullType(errors.Span{}),
				errors.Span{},
			),
		),
		InputBuiltinIdent: analyzer.NewBuiltinVar(
			ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
				errors.Span{},
				ast.NewStringType(errors.Span{}),
				errors.Span{},
			),
		),
		PollInputBuiltinIdent: analyzer.NewBuiltinVar(
			ast.NewFunctionType(
				ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
				errors.Span{},
				ast.NewOptionType(ast.NewStringType(errors.Span{}), errors.Span{}),
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
								TimeObjType(errors.Span{}),
								nil)}),
						errors.Span{},
						DurationObjType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
				ast.NewObjectTypeField(
					pAst.NewSpannedIdent("now", errors.Span{}),
					ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind(make([]ast.FunctionTypeParam, 0)),
						errors.Span{},
						TimeObjType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
				ast.NewObjectTypeField(
					pAst.NewSpannedIdent("add_days", errors.Span{}),
					ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("time", errors.Span{}), TimeObjType(errors.Span{}), nil),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("days", errors.Span{}), ast.NewIntType(errors.Span{}), nil),
						}),
						errors.Span{},
						TimeObjType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
				ast.NewObjectTypeField(
					pAst.NewSpannedIdent("add_hours", errors.Span{}),
					ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("time", errors.Span{}), TimeObjType(errors.Span{}), nil),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("hours", errors.Span{}), ast.NewIntType(errors.Span{}), nil),
						}),
						errors.Span{},
						TimeObjType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
				ast.NewObjectTypeField(
					pAst.NewSpannedIdent("add_minutes", errors.Span{}),
					ast.NewFunctionType(
						ast.NewNormalFunctionTypeParamKind([]ast.FunctionTypeParam{
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("time", errors.Span{}), TimeObjType(errors.Span{}), nil),
							ast.NewFunctionTypeParam(pAst.NewSpannedIdent("hours", errors.Span{}), ast.NewIntType(errors.Span{}), nil),
						}),
						errors.Span{},
						TimeObjType(errors.Span{}),
						errors.Span{},
					),
					errors.Span{},
				),
			},
			errors.Span{},
		)),
	}
}
