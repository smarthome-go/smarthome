package driver

import "github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"

type DeviceCapability string

const (
	DeviceCapabilityBase     DeviceCapability = "base"
	DeviceCapabilityPower    DeviceCapability = "power"
	DeviceCapabilityDimmable DeviceCapability = "dimmable"
	DeviceCapabilityColor    DeviceCapability = "color"
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

func (set CapabilitySet[T]) Has(check T) bool {
	for _, elem := range set {
		if elem == check {
			return true
		}
	}

	return false
}

func (set *CapabilitySet[T]) Add(add T) {
	if set.Has(add) {
		return
	}

	*set = append(*set, add)
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
	CONFIG_FIELD_TYPE_FLOAT  CONFIG_FIELD_TYPE = "FLOAT"
	CONFIG_FIELD_TYPE_BOOL   CONFIG_FIELD_TYPE = "BOOL"
	CONFIG_FIELD_TYPE_STRING CONFIG_FIELD_TYPE = "STRING"
	CONFIG_FIELD_TYPE_LIST   CONFIG_FIELD_TYPE = "LIST"
	CONFIG_FIELD_TYPE_STRUCT CONFIG_FIELD_TYPE = "STRUCT"
	CONFIG_FIELD_TYPE_OPTION CONFIG_FIELD_TYPE = "OPTION"
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

func (descriptor ConfigFieldDescriptorAtom) Kind() CONFIG_FIELD_TYPE {
	return descriptor.Type
}

type ConfigFieldDescriptorWithInner struct {
	Type  CONFIG_FIELD_TYPE     `json:"type"`
	Inner ConfigFieldDescriptor `json:"inner"`
}

func (descriptor ConfigFieldDescriptorWithInner) Kind() CONFIG_FIELD_TYPE {
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

func (descriptor ConfigFieldDescriptorStruct) Kind() CONFIG_FIELD_TYPE {
	return CONFIG_FIELD_TYPE_STRUCT
}
