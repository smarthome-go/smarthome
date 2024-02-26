package homescript

import "github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"

type DeviceCapability string

const (
	DeviceCapabilityBase     DeviceCapability = "base"
	DeviceCapabilityPower    DeviceCapability = "power"
	DeviceCapabilityDimmable DeviceCapability = "dimmable"
	DeviceCapabilitySensor   DeviceCapability = "sensor"
)

type DriverCapability string

const (
	// TODO: Add more
	DriverCapabilityBase DriverCapability = "base"
)

type DriverInfo struct {
	DriverConfig ConfigInfoWrapperDriver `json:"driver"`
	DeviceConfig ConfigInfoWrapperDevice `json:"device"`
}

//
// Begin capability set
//

type CapabilitySet[T comparable] []T

func (self CapabilitySet[T]) Has(check T) bool {
	for _, elem := range self {
		if elem == check {
			return true
		}
	}

	return false
}

func (self *CapabilitySet[T]) Add(add T) {
	if self.Has(add) {
		return
	}

	*self = append(*self, add)
}

//
// End capability set
//

type ConfigInfoWrapperDevice struct {
	Capabilities CapabilitySet[DeviceCapability] `json:"capabilities"`
	Info         ConfigInfoWrapper               `json:"info"`
}

type ConfigInfoWrapperDriver struct {
	Capabilities CapabilitySet[DriverCapability] `json:"capabilities"`
	Info         ConfigInfoWrapper               `json:"info"`
}

type ConfigInfoWrapper struct {
	Config ConfigFieldDescriptorStruct `json:"config"`
	// This field is ignored as it would add redundant bloat to HTTP responses
	HmsType ast.ObjectType `json:"-"`
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
