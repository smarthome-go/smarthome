package homescript

import "github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"

type DriverInfo struct {
	DriverConfig ConfigInfoWrapper `json:"driver"`
	DeviceConfig ConfigInfoWrapper `json:"device"`
}

type ConfigInfoWrapper struct {
	Config  ConfigFieldDescriptorStruct
	HmsType ast.Type
}

type CONFIG_FIELD_TYPE string

const (
	CONFIG_FIELD_TYPE_INT    CONFIG_FIELD_TYPE = "INT"
	CONFIG_FIELD_TYPE_FLOAT                    = "FLOAT"
	CONFIG_FIELD_TYPE_BOOL                     = "BOOL"
	CONFIG_FIELD_TYPE_STRING                   = "STRING"
	CONFIG_FIELD_TYPE_LIST                     = "LIST"
	CONFIG_FIELD_TYPE_STRUCT                   = "STRUCT"
	CONFIG_FIELD_TYPE_OPTION                   = "OPTION"
)

type ConfigFieldDescriptor interface {
	Kind() CONFIG_FIELD_TYPE
}

//
// Atom: int, float, bool, string
//

type ConfigFieldDescriptorAtom struct {
	Type CONFIG_FIELD_TYPE `json:"type"`
}

func (self ConfigFieldDescriptorAtom) Kind() CONFIG_FIELD_TYPE {
	return self.Type
}

type ConfigFieldDescriptorWithInner struct {
	Type  CONFIG_FIELD_TYPE     `json:"type"`
	Inner ConfigFieldDescriptor `json:"inner"`
}

func (self ConfigFieldDescriptorWithInner) Kind() CONFIG_FIELD_TYPE {
	return CONFIG_FIELD_TYPE_LIST
}

//
// Struct
//

type ConfigFieldDescriptorStruct struct {
	Type   CONFIG_FIELD_TYPE `json:"type"`
	Fields []ConfigFieldItem `json:"fields"`
}

type ConfigFieldItem struct {
	Name string                `json:"name"`
	Type ConfigFieldDescriptor `json:"type"`
}

func (self ConfigFieldDescriptorStruct) Kind() CONFIG_FIELD_TYPE {
	return CONFIG_FIELD_TYPE_STRUCT
}
