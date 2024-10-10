package analyzer

import (
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
)

func MqttCallbackFn(span errors.Span) ast.Type {
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
