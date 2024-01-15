package homescript

type DriverInfo struct {
	DriverConfig ConfigFieldDescriptorStruct `json:"driver"`
	DeviceConfig ConfigFieldDescriptorStruct `json:"device"`
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

// //
// // Float
// //
//
// type ConfigFieldDescriptorFloat struct {
// 	Value float64
// }
//
// func (self ConfigFieldDescriptorFloat) Kind() CONFIG_FIELD_TYPE {
// 	return CONFIG_FIELD_TYPE_FLOAT
// }
//
// //
// // Bool
// //
//
// type ConfigFieldDescriptorBool struct {
// 	Value bool
// }
//
// func (self ConfigFieldDescriptorBool) Kind() CONFIG_FIELD_TYPE {
// 	return CONFIG_FIELD_TYPE_BOOL
// }
//
// //
// // String
// //
//
// type ConfigFieldDescriptorString struct {
// 	Value string
// }
//
// func (self ConfigFieldDescriptorString) Kind() CONFIG_FIELD_TYPE {
// 	return CONFIG_FIELD_TYPE_STRING
// }

//
// List / Option
//

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

// TODO: what to do here?
// Transforms a `ConfigField` into native Go types so that it can be encoded as JSON.
// func MarshalConfigField(self ConfigField) interface{} {
// 	switch self.Kind() {
// 	case CONFIG_FIELD_TYPE_INT:
// 		return self.(ConfigFieldInt).Value
// 	case CONFIG_FIELD_TYPE_FLOAT:
// 		return self.(ConfigFieldFloat).Value
// 	case CONFIG_FIELD_TYPE_BOOL:
// 		return self.(ConfigFieldBool).Value
// 	case CONFIG_FIELD_TYPE_STRING:
// 		return self.(ConfigFieldString).Value
// 	case CONFIG_FIELD_TYPE_LIST:
// 		valuesRaw := self.(ConfigFieldList).Values
// 		valuesNew := make([]interface{}, len(valuesRaw))
// 		for idx, value := range valuesRaw {
// 			valuesNew[idx] = MarshalConfigField(value)
// 		}
// 		return valuesNew
// 	case CONFIG_FIELD_TYPE_STRUCT:
// 		fieldsRaw := self.(ConfigFieldStruct).Fields
// 		mapNew := make(map[string]interface{})
// 		for key, value := range fieldsRaw {
// 			mapNew[key] = MarshalConfigField(value)
// 		}
// 		return mapNew
// 	default:
// 		panic("A new config field type was introduced without updating this code")
// 	}
// }
