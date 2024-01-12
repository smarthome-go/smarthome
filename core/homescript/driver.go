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

	if !res.Success || len(res.Errors) != 0 {
		err0 := res.Errors[0]
		log.Debugf("Could not extract driver info: %s", err0.String())

		diagnostics := make([]diagnostic.Diagnostic, len(res.Errors))
		for idx, err := range res.Errors {
			diagnostics[idx] = diagnostic.Diagnostic{
				Level:   err.DiagnosticError.Level,
				Message: err.DiagnosticError.Message,
				Notes:   err.DiagnosticError.Notes,
				Span:    err.Span,
			}
		}

		return DriverInfo{}, diagnostics, nil
	}

	info, diagnosticErr := ExtractDriverInfo(driver, analyzed, filename)
	if diagnosticErr != nil {
		return DriverInfo{}, []diagnostic.Diagnostic{*diagnosticErr}, nil
	}

	return info, nil, nil
}

func ExtractDriverInfo(driver database.DeviceDriver, analyzed map[string]ast.AnalyzedProgram, mainModule string) (DriverInfo, *diagnostic.Diagnostic) {
	driverSingleton, driverSingletonFound := ast.AnalyzedSingletonTypeDefinition{}, false
	deviceSingleton, deviceSingletonFound := ast.AnalyzedSingletonTypeDefinition{}, false

	// Iterate over singletons, assert that there is a `driver` singleton
	for _, singleton := range analyzed[mainModule].Singletons {
		if singleton.Ident.Ident() == DRIVER_SINGLETON_IDENT {
			driverSingleton = singleton
			driverSingletonFound = true
			continue
		}

		if singleton.Ident.Ident() == DRIVER_DEVICE_SINGLETON_IDENT {
			deviceSingleton = singleton
			deviceSingletonFound = true
			continue
		}
	}

	if !driverSingletonFound {
		return DriverInfo{}, &diagnostic.Diagnostic{
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
		}
	}

	if !deviceSingletonFound {
		return DriverInfo{}, &diagnostic.Diagnostic{
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
		}
	}

	driverConfig, err := singletonAsConfigField(driverSingleton)
	if err != nil {
		return DriverInfo{}, &diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Cannot generate configuration interface from this type: %s", err.Error()),
			Notes: []string{
				"This type is not supported in the configuration of drivers",
			},
			Span: driverSingleton.TypeDef.Range,
		}
	}

	// TODO: validate that the driver implements all required templates

	deviceConfig, err := singletonAsConfigField(deviceSingleton)
	if err != nil {
		return DriverInfo{}, &diagnostic.Diagnostic{
			Level:   diagnostic.DiagnosticLevelError,
			Message: fmt.Sprintf("Cannot generate configuration interface from this type: %s", err.Error()),
			Notes: []string{
				"This type is not supported in the configuration of drivers",
			},
			Span: deviceSingleton.TypeDef.Range,
		}
	}

	// TODO: validate that the device implements all required templates

	return DriverInfo{
		DriverConfig: driverConfig.(ConfigFieldDescriptorStruct),
		DeviceConfig: deviceConfig.(ConfigFieldDescriptorStruct),
	}, nil
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
