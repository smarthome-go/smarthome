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

type HMS_PROGRAM_KIND uint8

const (
	HMS_PROGRAM_KIND_NORMAL HMS_PROGRAM_KIND = iota
	HMS_PROGRAM_KIND_DEVICE_DRIVER
)
