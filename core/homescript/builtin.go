package homescript

import (
	"fmt"
	"strings"

	"github.com/smarthome-go/homescript/v2/homescript"
	"github.com/smarthome-go/homescript/v2/homescript/errors"
	hmsErrors "github.com/smarthome-go/homescript/v2/homescript/errors"
)

var scopeAdditions = map[string]homescript.Value{
	"on_click_hms": homescript.ValueBuiltinFunction{
		Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
			if err := checkArgs("on_click_hms", span, args, homescript.TypeString, homescript.TypeString); err != nil {
				return nil, nil, err
			}

			targetCode := strings.ReplaceAll(args[0].(homescript.ValueString).Value, "'", "\\'")
			targetCode = strings.ReplaceAll(targetCode, "\"", "\\\"")
			inner := args[1].(homescript.ValueString).Value

			callBackCode := fmt.Sprintf("fetch('/api/homescript/run/live', {method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ code: `%s`, args: [] }) })", targetCode)

			wrapper := fmt.Sprintf("<span onclick=\"%s\">%s</span>", callBackCode, inner)

			return homescript.ValueString{Value: wrapper, Range: span}, nil, nil
		},
	},
	"on_click_js": homescript.ValueBuiltinFunction{
		Callback: func(executor homescript.Executor, span hmsErrors.Span, args ...homescript.Value) (homescript.Value, *int, *hmsErrors.Error) {
			if err := checkArgs("on_click_js", span, args, homescript.TypeString, homescript.TypeString); err != nil {
				return nil, nil, err
			}

			targetCode := strings.ReplaceAll(args[0].(homescript.ValueString).Value, "\"", "\\\"")
			inner := args[1].(homescript.ValueString).Value

			wrapper := fmt.Sprintf("<span onclick=\"%s\">%s</span>", targetCode, inner)

			return homescript.ValueString{Value: wrapper, Range: span}, nil, nil
		},
	},
}

// Helper function which checks the validity of args provided to builtin functions
func checkArgs(name string, span errors.Span, args []homescript.Value, types ...homescript.ValueType) *errors.Error {
	if len(args) != len(types) {
		s := ""
		if len(types) != 1 {
			s = "s"
		}
		return errors.NewError(
			span,
			fmt.Sprintf("function '%s' takes %d argument%s but %d were given", name, len(types), s, len(args)),
			errors.TypeError,
		)
	}
	for i, typ := range types {
		if args[i].Type() != typ {
			return errors.NewError(
				span,
				fmt.Sprintf("Argument %d of function '%s' has to be of type %v", i+1, name, typ),
				errors.TypeError,
			)
		}
	}
	return nil
}
