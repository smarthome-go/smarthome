package homescript

import (
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	herrors "github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/smarthome/core/database"
)

const DRIVER_SINGLETON_IDENT = "@Driver"
const DRIVER_DEVICE_SINGLETON_IDENT = "@Device"

func ExtractDriverInfoTotal(driver database.DeviceDriver) (info DriverInfo, hmsErrors []diagnostic.Diagnostic, err error) {
	filename := fmt.Sprintf("@%s:%s", driver.VendorId, driver.ModelId)

	analyzed, res, err := HmsManager.Analyze(
		"", // TODO: what to do with this field??
		filename,
		driver.HomescriptCode,
		HMS_PROGRAM_KIND_DEVICE_DRIVER,
		&AnalyzerDriverMetadata{
			VendorId: driver.VendorId,
			ModelId:  driver.ModelId,
		},
	)
	if err != nil {
		return DriverInfo{}, nil, err
	}

	if len(res.Errors) != 0 {
		err0 := res.Errors[0]
		log.Debugf("Could not extract driver info: %s", err0.String())

		diagnostics := make([]diagnostic.Diagnostic, len(res.Errors))
		for idx, err := range res.Errors {
			if err.SyntaxError != nil {
				diagnostics[idx] = diagnostic.Diagnostic{
					Level:   diagnostic.DiagnosticLevelError,
					Message: err.SyntaxError.Message,
					Notes:   make([]string, 0),
					Span:    err.Span,
				}
			} else if err.DiagnosticError != nil {
				diagnostics[idx] = diagnostic.Diagnostic{
					Level:   err.DiagnosticError.Level,
					Message: err.DiagnosticError.Message,
					Notes:   err.DiagnosticError.Notes,
					Span:    err.Span,
				}
			} else if err.RuntimeInterrupt != nil {
				panic("Unreachable state: this program is only analyzed, not executed")
			}
		}

		return DriverInfo{}, diagnostics, nil
	}

	info, diagnostics := ExtractDriverInfo(driver, analyzed, filename)
	if len(diagnostics) != 0 {
		return DriverInfo{}, diagnostics, nil
	}

	return info, nil, nil
}

func ExtractDriverInfo(driver database.DeviceDriver, analyzed map[string]ast.AnalyzedProgram, mainModule string) (DriverInfo, []diagnostic.Diagnostic) {
	diagnonstics := make([]diagnostic.Diagnostic, 0)

	driverSingleton, driverSingletonFound := ast.AnalyzedSingletonTypeDefinition{}, false
	deviceSingleton, deviceSingletonFound := ast.AnalyzedSingletonTypeDefinition{}, false

	// Iterate over singletons, assert that there is a `driver` singleton
	for _, singleton := range analyzed[mainModule].Singletons {
		if singleton.Ident.Ident() == DRIVER_SINGLETON_IDENT {
			driverSingleton = singleton
			driverSingletonFound = true

			if deviceSingletonFound && driverSingletonFound {
				break
			}

			continue
		}

		if singleton.Ident.Ident() == DRIVER_DEVICE_SINGLETON_IDENT {
			deviceSingleton = singleton
			deviceSingletonFound = true

			if deviceSingletonFound && driverSingletonFound {
				break
			}

			continue
		}
	}

	if !driverSingletonFound {
		diagnonstics = append(diagnonstics, diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Singleton `%s` not found", DRIVER_SINGLETON_IDENT),
			Notes: []string{
				fmt.Sprintf("A singleton named `%s` is required for every driver implementation", DRIVER_DEVICE_SINGLETON_IDENT),
				fmt.Sprintf("This singleton can be declared like this: `TODO, add final syntax`"),
			},
			Span: herrors.Span{
				Start:    herrors.Location{},
				End:      herrors.Location{},
				Filename: mainModule,
			},
		})
	}

	if !deviceSingletonFound {
		diagnonstics = append(diagnonstics, diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Singleton `%s` not found", DRIVER_DEVICE_SINGLETON_IDENT),
			Notes: []string{
				fmt.Sprintf("A singleton named `%s` is required for every driver implementation", DRIVER_DEVICE_SINGLETON_IDENT),
				fmt.Sprintf("This singleton can be declared like this: `TODO, add final syntax`"),
			},
			Span: herrors.Span{
				Start:    herrors.Location{},
				End:      herrors.Location{},
				Filename: mainModule,
			},
		})
	}

	// If one of the required singletons was not found, report the errors
	if !driverSingletonFound || !deviceSingletonFound {
		return DriverInfo{}, diagnonstics
	}

	driverConfig, err := singletonAsConfigField(driverSingleton)
	if err != nil {
		return DriverInfo{}, []diagnostic.Diagnostic{{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Cannot generate configuration interface from this type: %s", err.Error()),
			Notes: []string{
				"This type is not supported in the configuration of drivers",
			},
			Span: driverSingleton.TypeDef.Range,
		},
		}
	}

	const DRIVER_TEMPLATE_IDENT = "Driver"
	const DEVICE_TEMPLATE_IDENT = "Device"

	// TODO: validate that the driver implements all required templates
	if d := requireTemplateImplementation(driverSingleton, DRIVER_TEMPLATE_IDENT, "hardware driver"); d != nil {
		diagnonstics = append(diagnonstics, *d)
	}

	if d := requireTemplateImplementation(deviceSingleton, DEVICE_TEMPLATE_IDENT, "device"); d != nil {
		diagnonstics = append(diagnonstics, *d)
	}

	deviceConfig, err := singletonAsConfigField(deviceSingleton)
	if err != nil {
		diagnonstics = append(diagnonstics, diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Cannot generate configuration interface from this type: %s", err.Error()),
			Notes: []string{
				"This type is not supported in the configuration of drivers",
			},
			Span: deviceSingleton.TypeDef.Range,
		})
	}

	// TODO: validate that the device implements all required templates

	return DriverInfo{
		DriverConfig: driverConfig.(ConfigFieldDescriptorStruct),
		DeviceConfig: deviceConfig.(ConfigFieldDescriptorStruct),
	}, diagnonstics
}

func requireTemplateImplementation(singleton ast.AnalyzedSingletonTypeDefinition, templateIdent string, usecase string) *diagnostic.Diagnostic {
	fmt.Printf("Checking singleton `%s`: %v\n", singleton.ImplementsTemplates, singleton.Ident)

	containsImpl := false
	for _, implementedTempl := range singleton.ImplementsTemplates {
		fmt.Printf("found impl: %v\n", implementedTempl)
		if implementedTempl.Template.Ident() == templateIdent {
			containsImpl = true
			break
		}
	}

	if !containsImpl {
		return &diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Template `%s` is not implemented for this Singleton", singleton.Ident),
			Notes: []string{
				fmt.Sprintf("In order to use this singleton as a %s, it must implement the template `%s` with at least default capabilities", usecase, templateIdent),
				fmt.Sprintf("It can be implemented like this: `impl %s for %s with { ... } { ... }`", templateIdent, singleton.Ident),
			},
			Span: singleton.TypeDef.Range,
		}
	}

	return nil
}

func singletonAsConfigField(from ast.AnalyzedSingletonTypeDefinition) (ConfigFieldDescriptor, error) {
	return typeToConfigField(from.TypeDef.RhsType)
}

func typeToConfigField(from ast.Type) (ConfigFieldDescriptor, error) {
	switch from.Kind() {
	case ast.IntTypeKind:
		return ConfigFieldDescriptorAtom{
			Type: CONFIG_FIELD_TYPE_INT,
		}, nil
	case ast.FloatTypeKind:
		return ConfigFieldDescriptorAtom{
			Type: CONFIG_FIELD_TYPE_FLOAT,
		}, nil
	case ast.BoolTypeKind:
		return ConfigFieldDescriptorAtom{
			Type: CONFIG_FIELD_TYPE_BOOL,
		}, nil
	case ast.StringTypeKind:
		return ConfigFieldDescriptorAtom{
			Type: CONFIG_FIELD_TYPE_STRING,
		}, nil
	case ast.ListTypeKind:
		list := from.(ast.ListType)
		inner, err := typeToConfigField(list.Inner)
		return ConfigFieldDescriptorWithInner{
			Self:  CONFIG_FIELD_TYPE_LIST,
			Inner: inner,
		}, err
	case ast.ObjectTypeKind:
		obj := from.(ast.ObjectType)
		fields := make(map[string]ConfigFieldDescriptor)

		for _, field := range obj.ObjFields {
			fieldNew, err := typeToConfigField(field.Type)
			if err != nil {
				return nil, err
			}

			fields[field.FieldName.Ident()] = fieldNew
		}

		return ConfigFieldDescriptorStruct{
			Self:   CONFIG_FIELD_TYPE_STRUCT,
			Fields: fields,
		}, nil
	case ast.OptionTypeKind:
		option := from.(ast.OptionType)
		inner, err := typeToConfigField(option.Inner)
		return ConfigFieldDescriptorWithInner{
			Self:  CONFIG_FIELD_TYPE_OPTION,
			Inner: inner,
		}, err
	default:
		return nil, fmt.Errorf("Cannot derive user configuration from type `%s`", from)
	}
}
