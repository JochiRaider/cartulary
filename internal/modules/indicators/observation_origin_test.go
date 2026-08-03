package indicators

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"testing"

	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
)

func TestObservationOriginRegistryAndProducerMapping(t *testing.T) {
	t.Parallel()
	valid := []string{
		"manual_entry",
		"clipboard_paste",
		"csv_import",
		"xlsx_import",
		"api_import",
		"extraction",
		"system",
	}
	for _, raw := range valid {
		parsed, err := indicatororigin.Parse(raw)
		if err != nil || parsed.String() != raw {
			t.Fatalf("parse exact origin %q = %q, %v", raw, parsed, err)
		}
	}
	for _, raw := range []string{
		"",
		" manual_entry",
		"manual_entry ",
		"Manual_Entry",
		"interactive_cell",
		"auto_extract",
		"domain",
		"extension:manual_entry",
		"unknown",
	} {
		if _, err := indicatororigin.Parse(raw); !errors.Is(err, indicatororigin.ErrInvalidObservationOrigin) {
			t.Fatalf("parse invalid origin %q = %v", raw, err)
		}
	}
}

func TestObservationProducerSurfaceHasNoSystemConstructor(t *testing.T) {
	t.Parallel()
	parsed, err := parser.ParseDir(token.NewFileSet(), ".", nil, 0)
	if err != nil {
		t.Fatalf("parse Indicator owner surface: %v", err)
	}
	forbidden := []string{
		"ObservationOrigin",
		"ObservationProducerContext",
		"ParseObservationOrigin",
		"ManualEntryObservationProducer",
		"ClipboardPasteObservationProducer",
		"CSVImportObservationProducer",
		"XLSXImportObservationProducer",
		"APIImportObservationProducer",
		"ExtractionObservationProducer",
	}
	for _, file := range parsed["indicators"].Files {
		for _, declaration := range file.Decls {
			name := exportedDeclarationName(declaration)
			if slices.Contains(forbidden, name) {
				t.Fatalf("retired observation producer surface %s remains exported", name)
			}
		}
	}
}

func TestInvalidObservationProducerFailsBeforeTransaction(t *testing.T) {
	t.Parallel()
	parsed, err := parser.ParseFile(token.NewFileSet(), "store.go", nil, 0)
	if err != nil {
		t.Fatalf("parse Indicator observation command: %v", err)
	}
	forbiddenFields := []string{"Producer", "OriginKind", "MutationSource", "CreatedAt"}
	for _, declaration := range parsed.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, specification := range general.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "IndicatorObservationCreateParams" {
				continue
			}
			structure := typeSpec.Type.(*ast.StructType)
			for _, field := range structure.Fields.List {
				for _, name := range field.Names {
					if slices.Contains(forbiddenFields, name.Name) {
						t.Fatalf("caller-controlled observation field %s remains public", name.Name)
					}
				}
			}
			return
		}
	}
	t.Fatal("IndicatorObservationCreateParams declaration not found")
}

func exportedDeclarationName(declaration ast.Decl) string {
	switch typed := declaration.(type) {
	case *ast.FuncDecl:
		if typed.Name.IsExported() {
			return typed.Name.Name
		}
	case *ast.GenDecl:
		for _, specification := range typed.Specs {
			switch spec := specification.(type) {
			case *ast.TypeSpec:
				if spec.Name.IsExported() {
					return spec.Name.Name
				}
			case *ast.ValueSpec:
				for _, name := range spec.Names {
					if name.IsExported() {
						return name.Name
					}
				}
			}
		}
	}
	return ""
}
