package types

import (
	"fmt"

	"github.com/smarthome-go/homescript/v3/homescript/diagnostic"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	"github.com/smarthome-go/homescript/v3/homescript/runtime/value"
)

//
// Results and errors.
//

type HmsDiagnosticsContainer struct {
	ContainsError bool
	Diagnostics   []HmsError
	FileContents  map[string]string
}

type HmsRes struct {
	Errors HmsDiagnosticsContainer
	// The state of the used singletons after execution.
	Singletons map[string]value.Value
	// This is `nil` if no additional function was invoced or the called function did not return a value.
	ReturnValue value.Value
	// This is non zero-valued if an additional function is called.
	// Required so that errors caused by user-implementations can be correctly displayed.
	CalledFunctionSpan errors.Span
}

type HmsError struct {
	SyntaxError      *HmsSyntaxError      `json:"syntaxError"`
	DiagnosticError  *HmsDiagnosticError  `json:"diagnosticError"`
	RuntimeInterrupt *HmsRuntimeInterrupt `json:"runtimeError"`
	Span             errors.Span          `json:"span"`
}

func HmsErrorsFromDiagnostics(input []diagnostic.Diagnostic) []HmsError {
	output := make([]HmsError, 0)

	for _, diagnosticMsg := range input {
		if diagnosticMsg.Level != diagnostic.DiagnosticLevelError {
			continue
		}

		output = append(output, HmsError{
			SyntaxError: nil,
			DiagnosticError: &HmsDiagnosticError{
				Level:   diagnostic.DiagnosticLevelError,
				Message: diagnosticMsg.Message,
				Notes:   diagnosticMsg.Notes,
			},
			RuntimeInterrupt: nil,
			Span:             diagnosticMsg.Span,
		})
	}

	return output
}

func (e HmsError) String() string {
	spanDisplay := fmt.Sprintf("%s:%d:%d", e.Span.Filename, e.Span.Start.Line, e.Span.Start.Column)
	switch {
	case e.SyntaxError != nil:
		return fmt.Sprintf("Syntax error at %s: `%s`", spanDisplay, e.SyntaxError.Message)
	case e.DiagnosticError != nil:
		return fmt.Sprintf("Semantic error at %s: `%s`", spanDisplay, e.DiagnosticError.Message)
	case e.RuntimeInterrupt != nil:
		return fmt.Sprintf("%s at %s: `%s`", e.RuntimeInterrupt, spanDisplay, e.RuntimeInterrupt.Message)
	}

	panic("Illegal HmsError")
}

type HmsSyntaxError struct {
	Message string `json:"message"`
}

type HmsDiagnosticError struct {
	Level   diagnostic.DiagnosticLevel `json:"kind"`
	Message string                     `json:"message"`
	Notes   []string                   `json:"notes"`
}

type HmsRuntimeInterrupt struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

func (e HmsRuntimeInterrupt) String() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}
