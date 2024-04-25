package driver

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/smarthome-go/homescript/v3/homescript"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	herrors "github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/lexer"
	"github.com/smarthome-go/smarthome/core/database"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

var log *logrus.Logger

func InitLogger(logger *logrus.Logger) {
	log = logger
}

const DRIVER_TEMPLATE_IDENT = "Driver"
const DEVICE_TEMPLATE_IDENT = "Device"

const DRIVER_SINGLETON_IDENT = "Driver"
const DRIVER_DEVICE_SINGLETON_IDENT = "Device"

const DRIVER_FIELD_REQUIRED_ANNOTATION = "setting"

// TODO: make this const
var DriverSingletonIdent = fmt.Sprintf("%s%s", lexer.SINGLETON_TOKEN, DRIVER_SINGLETON_IDENT)
var DriverDeviceSingletonIdent = fmt.Sprintf("%s%s", lexer.SINGLETON_TOKEN, DRIVER_DEVICE_SINGLETON_IDENT)
var DriverFieldRequiredAnnotation = fmt.Sprintf("%s%s", lexer.TYPE_ANNOTATION_TOKEN, DRIVER_FIELD_REQUIRED_ANNOTATION)

type DriverManager struct {
	Hms types.Manager
}

// TODO: do this correctly
var Manager DriverManager

func InitManager(hmsManager types.Manager) {
	Manager = DriverManager{
		Hms: hmsManager,
	}
}

func (self *DriverManager) ExtractDriverInfoTotal(
	vendorID string,
	modelID string,
	homescriptCode string,
) (info DriverInfo, hmsErrors []diagnostic.Diagnostic, err error) {
	// TODO: remove this hack for the filename
	// TODO: use the ananlyze with context function here
	filename := types.CreateDriverHmsId(database.DriverTuple{VendorID: vendorID, ModelID: modelID})

	analyzed, res, err := self.Hms.Analyze(
		homescript.InputProgram{
			ProgramText: homescriptCode,
			Filename:    filename,
		},
		types.NewExecutionContextDriver(
			vendorID,
			modelID,
			nil,
		),
	)
	if err != nil {
		return DriverInfo{}, nil, err
	}

	if res.ContainsError {
		// Only include actual errors, not other diagnostic messages
		errors := make([]types.HmsError, 0)
		for _, err := range res.Diagnostics {
			if err.DiagnosticError != nil && err.DiagnosticError.Level != diagnostic.DiagnosticLevelError {
				continue
			}
			errors = append(errors, err)
		}

		log.Debugf("Could not extract driver info: %s", errors[0].String())
		diagnostics := make([]diagnostic.Diagnostic, len(errors))
		for idx, err := range errors {
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

	info, diagnostics := ExtractDriverInfo(analyzed, filename, false)
	if len(diagnostics) != 0 {
		return DriverInfo{}, diagnostics, nil
	}

	return info, nil, nil
}

func ExtractDriverInfo(
	analyzed map[string]ast.AnalyzedProgram,
	mainModule string,
	emitWarnings bool,
) (DriverInfo, []diagnostic.Diagnostic) {
	diagnostics := make([]diagnostic.Diagnostic, 0)

	driverSingleton, driverSingletonFound := ast.AnalyzedSingletonTypeDefinition{}, false
	driverCapabilities := make(CapabilitySet[DriverCapability], 0)

	deviceSingleton, deviceSingletonFound := ast.AnalyzedSingletonTypeDefinition{}, false
	deviceCapabilities := make(CapabilitySet[DeviceCapability], 0)

	// Iterate over singletons, assert that there is a `driver` and a `device` singleton.
	for _, singleton := range analyzed[mainModule].Singletons {
		if singleton.Ident.Ident() == DriverSingletonIdent {
			driverSingleton = singleton
			driverSingletonFound = true

			// Map HMS capabilities to driver capabilities.
			for _, impl := range analyzed[mainModule].ImplBlocks {
				if impl.SingletonIdent.Ident() == DriverSingletonIdent {
					template := driverTemplate(herrors.Span{})
					for ident := range impl.FinalCapabilities {
						driverCapabilities = append(driverCapabilities, template.Capabilities[ident])
					}

				}
			}

			if deviceSingletonFound && driverSingletonFound {
				break
			}

			continue
		}

		if singleton.Ident.Ident() == DriverDeviceSingletonIdent {
			deviceSingleton = singleton
			deviceSingletonFound = true

			// Map HMS capabilities to device capabilities.
			for _, impl := range analyzed[mainModule].ImplBlocks {
				if impl.SingletonIdent.Ident() == DriverDeviceSingletonIdent {
					template := deviceTemplate(herrors.Span{})
					for ident := range impl.FinalCapabilities {
						deviceCapabilities = append(deviceCapabilities, template.Capabilities[ident])
					}

				}
			}

			if deviceSingletonFound && driverSingletonFound {
				break
			}
		}
	}

	if !driverSingletonFound {
		diagnostics = append(diagnostics, diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Singleton `%s` not found", DriverSingletonIdent),
			Notes: []string{
				fmt.Sprintf("A singleton named `%s` is required for every driver implementation", DriverDeviceSingletonIdent),
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
		diagnostics = append(diagnostics, diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Singleton `%s` not found", DriverDeviceSingletonIdent),
			Notes: []string{
				fmt.Sprintf("A singleton named `%s` is required for every driver implementation", DriverDeviceSingletonIdent),
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
		return DriverInfo{}, diagnostics
	}

	driverConfig, diagnosticsExtraction, err := typeToConfigField(driverSingleton.SingletonType, true, driverSingleton.Range)
	if err != nil {
		return DriverInfo{}, []diagnostic.Diagnostic{{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Cannot generate configuration interface from this type: %s", err.Error()),
			Notes: []string{
				"This type does not support driver implementation",
			},
			Span: driverSingleton.Type().Span(),
		},
		}
	}

	if diagnosticsExtraction != nil && emitWarnings {
		diagnostics = append(diagnostics, diagnosticsExtraction...)
	}

	// TODO: validate that the driver implements all required templates
	if d := requireTemplateImplementation(driverSingleton, DRIVER_TEMPLATE_IDENT, "hardware driver"); d != nil {
		diagnostics = append(diagnostics, *d)
	}

	if d := requireTemplateImplementation(deviceSingleton, DEVICE_TEMPLATE_IDENT, "device"); d != nil {
		diagnostics = append(diagnostics, *d)
	}

	deviceConfig, diagnosticsExtraction, err := typeToConfigField(deviceSingleton.SingletonType, true, deviceSingleton.Range)
	if err != nil {
		diagnostics = append(diagnostics, diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Cannot generate configuration interface from this type: %s", err.Error()),
			Notes: []string{
				"This type does not support driver implementation",
			},
			Span: deviceSingleton.Type().Span(),
		})
	}

	if diagnosticsExtraction != nil && emitWarnings {
		diagnostics = append(diagnostics, diagnosticsExtraction...)
	}

	// TODO: validate that the device implements all required templates

	incompatibleType := false
	if driverConfig.Kind() != CONFIG_FIELD_TYPE_STRUCT {
		diagnostics = append(diagnostics, diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Cannot generate configuration interface from this type: %s", err.Error()),
			Notes: []string{
				"This type does not support driver implementation",
			},
			Span: deviceSingleton.Type().Span(),
		})
		incompatibleType = true
	}

	if deviceConfig.Kind() != CONFIG_FIELD_TYPE_STRUCT {
		diagnostics = append(diagnostics, diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Cannot generate configuration interface from this type: %s", err.Error()),
			Notes: []string{
				"This type does not support driver implementation",
			},
			Span: deviceSingleton.Type().Span(),
		})
		incompatibleType = true
	}

	if incompatibleType {
		return DriverInfo{
			DriverConfig: ConfigInfoWrapperDriver{
				Capabilities: driverCapabilities,
				Info:         ConfigInfoWrapper{},
			},
			DeviceConfig: ConfigInfoWrapperDevice{
				Capabilities: deviceCapabilities,
				Info:         ConfigInfoWrapper{},
			},
		}, diagnostics
	}

	return DriverInfo{
		DriverConfig: ConfigInfoWrapperDriver{
			Capabilities: driverCapabilities,
			Info: ConfigInfoWrapper{
				Config:  driverConfig.(ConfigFieldDescriptorStruct),
				HmsType: driverSingleton.SingletonType.(ast.ObjectType),
			},
		},
		DeviceConfig: ConfigInfoWrapperDevice{
			Capabilities: deviceCapabilities,
			Info: ConfigInfoWrapper{
				Config:  deviceConfig.(ConfigFieldDescriptorStruct),
				HmsType: deviceSingleton.SingletonType.(ast.ObjectType),
			},
		},
	}, diagnostics
}

func requireTemplateImplementation(singleton ast.AnalyzedSingletonTypeDefinition, templateIdent string, usecase string) *diagnostic.Diagnostic {
	containsImpl := false
	for _, implementedTempl := range singleton.ImplementsTemplates {
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
			Span: singleton.Type().Span(),
		}
	}

	return nil
}

func typeToConfigField(from ast.Type, topLevel bool, contextSpan errors.Span) (ConfigFieldDescriptor, []diagnostic.Diagnostic, error) {
	if topLevel && from.Kind() != ast.ObjectTypeKind {
		panic(fmt.Sprintf("BUG warning: expected object type, found `%s`", from.Kind()))
	}

	switch from.Kind() {
	case ast.IntTypeKind:
		return ConfigFieldDescriptorAtom{
			Type: CONFIG_FIELD_TYPE_INT,
		}, nil, nil
	case ast.FloatTypeKind:
		return ConfigFieldDescriptorAtom{
			Type: CONFIG_FIELD_TYPE_FLOAT,
		}, nil, nil
	case ast.BoolTypeKind:
		return ConfigFieldDescriptorAtom{
			Type: CONFIG_FIELD_TYPE_BOOL,
		}, nil, nil
	case ast.StringTypeKind:
		return ConfigFieldDescriptorAtom{
			Type: CONFIG_FIELD_TYPE_STRING,
		}, nil, nil
	case ast.ListTypeKind:
		list := from.(ast.ListType)
		inner, diagnostics, err := typeToConfigField(list.Inner, false, from.Span())
		return ConfigFieldDescriptorWithInner{
			Type:  CONFIG_FIELD_TYPE_LIST,
			Inner: inner,
		}, diagnostics, err
	case ast.ObjectTypeKind:
		diagnostics := make([]diagnostic.Diagnostic, 0)
		obj := from.(ast.ObjectType)
		fields := make([]ConfigFieldItem, 0)

		for _, field := range obj.ObjFields {
			// If this field does not have the required annotation, do not add it
			// NOTE: this is only done if this is a top-level call.
			// For nested structures, all fields are taken into account.
			if topLevel && (field.Annotation == nil || field.Annotation.Ident() != DriverFieldRequiredAnnotation) {
				continue
			}

			fieldNew, diagnosticsRec, err := typeToConfigField(field.Type, false, field.Span)
			if err != nil {
				return nil, diagnosticsRec, err
			}
			diagnostics = append(diagnostics, diagnosticsRec...)

			fields = append(fields, ConfigFieldItem{
				Name: field.FieldName.Ident(),
				Type: fieldNew,
			})
		}

		// Guard case: do not enter error / warning handling code
		if len(fields) > 0 {
			return ConfigFieldDescriptorStruct{
				Type:   CONFIG_FIELD_TYPE_STRUCT,
				Fields: fields,
			}, diagnostics, nil
		}

		if topLevel {
			// For top-level structs (the whole driver / device), this is just a warning
			diagnostics = append(diagnostics, diagnostic.Diagnostic{
				Level:   diagnostic.DiagnosticLevelWarning,
				Message: "Cannot apply settings-based configuration on this singleton",
				Notes: []string{
					fmt.Sprintf(
						"A field can be used as a setting by prefixing it with the `%s%s` directive",
						lexer.TYPE_ANNOTATION_TOKEN,
						DRIVER_FIELD_REQUIRED_ANNOTATION),
				},
				Span: contextSpan,
				// In this case, the context span will be the span of the entire singleton, not just its type.
			})
		} else {
			// For nested elements of settings parameters, this is an error as this makes no sense
			diagnostics = append(diagnostics, diagnostic.Diagnostic{
				Level:   diagnostic.DiagnosticLevelError,
				Message: "Empty object type in a configuration parameter",
				Notes: []string{
					"This field is redundant and can be deleted",
				},
				Span: contextSpan,
				// The context span is used as this will be the span of the `current` parameter.
				// Therefore, the error span will include the whole parameter and not just its type.
			})
		}

		return ConfigFieldDescriptorStruct{
			Type:   CONFIG_FIELD_TYPE_STRUCT,
			Fields: fields,
		}, diagnostics, nil
	case ast.OptionTypeKind:
		option := from.(ast.OptionType)
		// NOTE: recursive calls do not return diagnostics
		inner, _, err := typeToConfigField(option.Inner, false, from.Span())
		return ConfigFieldDescriptorWithInner{
			Type:  CONFIG_FIELD_TYPE_OPTION,
			Inner: inner,
		}, nil, err
	default:
		return nil, nil, fmt.Errorf("Cannot derive user configuration from type `%s`", from)
	}
}
