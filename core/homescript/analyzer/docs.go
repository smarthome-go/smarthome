package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/smarthome-go/homescript/v3/homescript/analyzer"
	"github.com/smarthome-go/homescript/v3/homescript/errors"
	pAst "github.com/smarthome-go/homescript/v3/homescript/parser/ast"
	"github.com/smarthome-go/smarthome/core/device/driver"
	"github.com/smarthome-go/smarthome/core/homescript/types"
)

type builtinImportDoc struct {
	Module            string `json:"module"`
	Name              string `json:"name"`
	Kind              string `json:"kind"`
	Signature         string `json:"signature,omitempty"`
	TriggerSignature  string `json:"triggerSignature,omitempty"`
	CallbackSignature string `json:"callbackSignature,omitempty"`
}

type builtinImportDocs struct {
	Entries []builtinImportDoc `json:"entries"`
}

func BuildDocsJSON(context types.ExecutionContext) ([]byte, error) {
	entries := buildBuiltinImportDocs(context)
	data, err := json.MarshalIndent(builtinImportDocs{Entries: entries}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal builtin import docs: %w", err)
	}

	return data, nil
}

func BuildDocs(context types.ExecutionContext, outputPath string) error {
	if outputPath == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	data, err := BuildDocsJSON(context)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create docs directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("write docs file: %w", err)
	}

	return nil
}

func buildBuiltinImportDocs(context types.ExecutionContext) []builtinImportDoc {
	if context == nil {
		context = types.NewExecutionContextUserNoFilename("docs", map[string]string{})
	}

	entries := make([]builtinImportDoc, 0, len(builtinImportHandlers))
	keys := make([]types.ImportKey, 0, len(builtinImportHandlers))
	entrySet := make(map[types.ImportKey]struct{}, len(builtinImportHandlers))

	for key := range builtinImportHandlers {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].ModuleName == keys[j].ModuleName {
			return keys[i].ValueName < keys[j].ValueName
		}
		return keys[i].ModuleName < keys[j].ModuleName
	})

	for _, key := range keys {
		entry, ok := resolveBuiltinImportDoc(context, key, builtinImportHandlers[key])
		if !ok {
			continue
		}
		entries = append(entries, entry)
		entrySet[key] = struct{}{}
	}

	for key := range driver.Templates(errors.Span{}) {
		if _, exists := entrySet[key]; exists {
			continue
		}
		entries = append(entries, builtinImportDoc{
			Module: key.ModuleName,
			Name:   key.ValueName,
			Kind:   importKindToString(pAst.IMPORT_KIND_TEMPLATE),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Module == entries[j].Module {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].Module < entries[j].Module
	})

	return entries
}

func resolveBuiltinImportDoc(
	context types.ExecutionContext,
	key types.ImportKey,
	handler importHandler,
) (builtinImportDoc, bool) {
	kinds := []pAst.IMPORT_KIND{
		pAst.IMPORT_KIND_NORMAL,
		pAst.IMPORT_KIND_TRIGGER,
		pAst.IMPORT_KIND_TYPE,
		pAst.IMPORT_KIND_TEMPLATE,
	}

	for _, kind := range kinds {
		result, valueFound := handler(context, errors.Span{}, kind)
		if !valueFound {
			continue
		}
		if result.Type == nil && result.Template == nil && result.Trigger == nil {
			continue
		}
		return builtinImportDocFromResult(key, kind, result), true
	}

	return builtinImportDoc{}, false
}

func builtinImportDocFromResult(
	key types.ImportKey,
	kind pAst.IMPORT_KIND,
	result analyzer.BuiltinImport,
) builtinImportDoc {
	doc := builtinImportDoc{
		Module: key.ModuleName,
		Name:   key.ValueName,
		Kind:   importKindToString(kind),
	}

	if result.Trigger != nil {
		doc.TriggerSignature = result.Trigger.TriggerFnType.String()
		doc.CallbackSignature = result.Trigger.CallbackFnType.String()
		return doc
	}

	if result.Type != nil {
		doc.Signature = result.Type.String()
	}

	return doc
}

func importKindToString(kind pAst.IMPORT_KIND) string {
	switch kind {
	case pAst.IMPORT_KIND_NORMAL:
		return "normal"
	case pAst.IMPORT_KIND_TRIGGER:
		return "trigger"
	case pAst.IMPORT_KIND_TYPE:
		return "type"
	case pAst.IMPORT_KIND_TEMPLATE:
		return "template"
	default:
		return "unknown"
	}
}
