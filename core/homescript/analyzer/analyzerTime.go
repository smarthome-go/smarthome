package analyzer

import (
	ast "github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
)

func DurationObjType(span errors.Span) ast.Type {
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

func TimeObjType(span errors.Span) ast.Type {
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
