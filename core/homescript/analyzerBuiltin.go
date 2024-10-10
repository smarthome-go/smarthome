package homescript

import (
	"fmt"
	"slices"
	"strings"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer"
	"github.com/smarthome-go/homescript/v3/homescript/analyzer/ast"
	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
	"github.com/smarthome-go/smarthome/core/device/driver"
	a "github.com/smarthome-go/smarthome/core/homescript/analyzer"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

// A list of `known` object type annotations.
// The analyzer uses these in order to sanity-check every annotation.
var knownObjectTypeAnnotations = []string{driver.DRIVER_FIELD_REQUIRED_ANNOTATION}

type analyzerHost struct {
	context     types.ExecutionContext
	diagnostics []diagnostic.Diagnostic
}

func (analyzerHost) GetKnownObjectTypeFieldAnnotations() []string {
	return knownObjectTypeAnnotations
}

func NewAnalyzerHost(
	context types.ExecutionContext,
) analyzerHost {
	if context == nil {
		panic("Context cannot be <nil>")
	}

	return analyzerHost{
		context:     context,
		diagnostics: make([]diagnostic.Diagnostic, 0),
	}
}

func (self analyzerHost) PostValidationHook(
	analyzedModules map[string]ast.AnalyzedProgram,
	mainModule string,
	analyzer *analyzer.Analyzer,
	hasPreviousError bool,
) []diagnostic.Diagnostic {
	diagnostics := self.diagnostics

	switch self.context.Kind() {
	case types.HMS_PROGRAM_KIND_DEVICE_DRIVER:
		_, diagnosticsDriver := driver.ExtractDriverInfo(analyzedModules, mainModule, true)
		diagnostics = append(diagnostics, diagnosticsDriver...)
		return diagnostics
	default:
		// Forbid `trigger` annotations and singletons in non-driver code.
		for _, mod := range analyzedModules {
			for _, singleton := range mod.Singletons {
				diagnostics = append(diagnostics, diagnostic.Diagnostic{
					Level:   diagnostic.DiagnosticLevelError,
					Message: "Singletons are not allowed in user programs",
					Notes:   []string{"To use singletons, create a driver Homescript"},
					Span:    singleton.Span(),
				})
			}

			for _, fn := range mod.Functions {
				if fn.Annotation == nil {
					continue
				}

				for _, ann := range fn.Annotation.Items {
					switch item := ann.(type) {
					case ast.AnalyzedAnnotationItemTrigger:
						triggerIdent := item.TriggerSource.Ident()
						if slices.Contains(types.ForbiddenUserTriggers, triggerIdent) {
							diagnostics = append(diagnostics,
								diagnostic.Diagnostic{
									Level:   diagnostic.DiagnosticLevelError,
									Message: "This trigger source can not be used in user programs",
									Notes: []string{
										fmt.Sprintf("To use the trigger source `%s`, create a driver Homescript", triggerIdent),
										fmt.Sprintf("The following trigger sources cannot be used in user mode: %s", strings.Join(types.ForbiddenUserTriggers, ",")),
									},
									Span: ann.Span(),
								},
							)
						}
					default:
					}
				}
			}
		}

		return diagnostics
	}
}

func (self analyzerHost) GetBuiltinImport(
	moduleName string,
	valueName string,
	span errors.Span,
	kind pAst.IMPORT_KIND,
) (result analyzer.BuiltinImport, moduleFound bool, valueFound bool) {
	return a.GetImport(
		self.context,
		moduleName,
		valueName,
		span,
		kind,
	)
}

func (self analyzerHost) ResolveCodeModule(moduleName string) (code string, moduleFound bool, err error) {
	logger.Trace(fmt.Sprintf("Resolving module `%s`", moduleName))

	if self.context.Username() != nil {
		script, found, err := HmsManager.GetPersonalScriptById(moduleName, *self.context.Username())
		if err != nil || !found {
			return "", found, err
		}
		return script.Data.Code, true, nil
	} else {
		script, found, err := HmsManager.GetScriptById(moduleName)
		if err != nil || !found {
			return "", found, err
		}
		return script.Data.Code, true, nil
	}
}

// TODO: fill this
