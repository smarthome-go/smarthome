package types

import "github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"

type TemplateKind uint8

const (
	TemplateKindDriver TemplateKind = iota
	TemplateKindDevice
)

type Template interface {
	Kind() TemplateKind
	GetSpec() ast.TemplateSpec
}
