package driver

import (
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
)

// This type specifies the configuration interface of a driver / device.
type ConfigSpec struct {
	HelperText string        `json:"helperText"`
	Fields     []ConfigField `json:"fields"`
}

// This type specific each configurable field of the driver
// TODO: figure out how to serialize the type?
type ConfigField struct {
	Name string   `json:"name"`
	Key  string   `json:"key"`
	Type ast.Type `json:"type"`
}
