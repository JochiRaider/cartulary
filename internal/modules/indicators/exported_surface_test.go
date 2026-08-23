package indicators

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// indicatorExportedSurfaceAllowlist is a production boundary, not a historical
// inventory. New root exports require an explicit owner-contract decision.
var indicatorExportedSurfaceAllowlist = map[string]struct{}{
	"AffectedRecordVersion":                       {},
	"CreateCommand":                               {},
	"CreateResult":                                {},
	"ErrIllegalTransition":                        {},
	"ErrIndicatorNotFound":                        {},
	"ErrIndicatorObservationNotFound":             {},
	"ErrIndicatorSourceNotFound":                  {},
	"ErrInvalidCreateRequest":                     {},
	"ErrResolvedIndicatorNotFound":                {},
	"ErrRowVersionConflict":                       {},
	"ErrSourceTextUnavailable":                    {},
	"IncidentBundleContribution":                  {},
	"IndicatorCreateValidationError":              {},
	"IndicatorCreateValidationError.Error":        {},
	"IndicatorFindOrCreateParticipantCommand":     {},
	"IndicatorFindOrCreateParticipantResult":      {},
	"IndicatorLifecycleAppendParams":              {},
	"IndicatorLifecycleIntervalRecord":            {},
	"IndicatorLifecycleMutationResult":            {},
	"IndicatorObservationActionParams":            {},
	"IndicatorObservationCreateParams":            {},
	"IndicatorObservationMutationResult":          {},
	"IndicatorObservationRecord":                  {},
	"IndicatorObservationResolveParams":           {},
	"IndicatorReference":                          {},
	"NewImportCreateFacade":                       {},
	"NewIncidentBundleContribution":               {},
	"NewRevisionContribution":                     {},
	"NewStore":                                    {},
	"RecoveryStateContribution":                   {},
	"RecordEnvelopePort":                          {},
	"SourceTextPort":                              {},
	"SourceTextValue":                             {},
	"Store":                                       {},
	"Store.AppendIndicatorLifecycleInterval":      {},
	"Store.DismissIndicatorObservation":           {},
	"Store.GetIndicatorObservation":               {},
	"Store.ListIndicatorLifecycleIntervals":       {},
	"Store.ListIndicatorObservations":             {},
	"Store.ListSourceRecordIndicatorObservations": {},
	"Store.ResolveIndicatorObservation":           {},
	"Store.RestoreIndicatorObservation":           {},
	"Store.CreateIndicatorObservation":            {},
	"Store.CreateIndicatorRow":                    {},
	"Store.FindOrCreateIndicatorParticipantTx":    {},
	"Store.GetActiveIndicatorParticipant":         {},
	"Store.GetActiveIndicatorParticipantTx":       {},
	"StoreDependencies":                           {},
	"ValidateCreateCommand":                       {},
	"ViewSchemaID":                                {},
}

// indicatorExportRoles records a production responsibility for every retained
// declaration in the guarded surface.
var indicatorExportRoles = indicatorExportRoleInventory(map[string]string{
	"owner application command, result, validation, or classified error contract": `
		AffectedRecordVersion CreateCommand CreateResult ErrIllegalTransition ErrIndicatorNotFound
		ErrIndicatorObservationNotFound ErrIndicatorSourceNotFound ErrInvalidCreateRequest
		ErrResolvedIndicatorNotFound ErrRowVersionConflict ErrSourceTextUnavailable
		IndicatorCreateValidationError IndicatorCreateValidationError.Error
		IndicatorFindOrCreateParticipantCommand IndicatorFindOrCreateParticipantResult
		IndicatorLifecycleAppendParams IndicatorLifecycleIntervalRecord IndicatorLifecycleMutationResult
		IndicatorObservationActionParams IndicatorObservationCreateParams IndicatorObservationMutationResult
		IndicatorObservationRecord IndicatorObservationResolveParams IndicatorReference
		SourceTextPort SourceTextValue ValidateCreateCommand ViewSchemaID
	`,
	"typed source-owner contribution consumed by application assembly": `
		IncidentBundleContribution NewImportCreateFacade NewIncidentBundleContribution
		NewRevisionContribution RecoveryStateContribution
	`,
	"complete Indicators store construction and transaction participant capability": `
		NewStore RecordEnvelopePort Store StoreDependencies
	`,
	"live Indicators application operation consumed by HTTP, Workbook, Imports, or Network Flow": `
		Store.AppendIndicatorLifecycleInterval Store.CreateIndicatorObservation Store.CreateIndicatorRow
		Store.DismissIndicatorObservation Store.FindOrCreateIndicatorParticipantTx
		Store.GetActiveIndicatorParticipant Store.GetActiveIndicatorParticipantTx
		Store.GetIndicatorObservation Store.ListIndicatorLifecycleIntervals
		Store.ListIndicatorObservations Store.ListSourceRecordIndicatorObservations
		Store.ResolveIndicatorObservation Store.RestoreIndicatorObservation
	`,
})

func TestIndicatorExportedSurfaceReachabilityLock(t *testing.T) {
	t.Parallel()

	actual := exportedRootDeclarations(t)
	missingClassifications := difference(actual, indicatorExportedSurfaceAllowlist)
	staleClassifications := difference(indicatorExportedSurfaceAllowlist, actual)
	if len(missingClassifications) != 0 || len(staleClassifications) != 0 {
		t.Fatalf("Indicator exported surface violates the production allowlist: unapproved=%v stale=%v", missingClassifications, staleClassifications)
	}
	missingRoles := difference(indicatorExportedSurfaceAllowlist, indicatorExportRoleKeys())
	staleRoles := difference(indicatorExportRoleKeys(), indicatorExportedSurfaceAllowlist)
	if len(missingRoles) != 0 || len(staleRoles) != 0 {
		t.Fatalf("Indicator exported surface role inventory disagrees with allowlist: missing=%v stale=%v", missingRoles, staleRoles)
	}
	if len(actual) != 50 {
		t.Fatalf("Indicator exported surface contains %d declarations, want exact reviewed surface of 50", len(actual))
	}
	for declaration, role := range indicatorExportRoles {
		if strings.TrimSpace(role) == "" {
			t.Fatalf("Indicator export %s has no production-role reason", declaration)
		}
	}

	t.Run("production import topology", func(t *testing.T) {
		assertIndicatorsProductionImportBoundaries(t)
	})
	t.Run("owner-local helpers remain private", func(t *testing.T) {
		assertIndicatorOwnerLocalHelpersPrivate(t)
	})
	t.Run("test fixtures are value-returning and minimal", func(t *testing.T) {
		assertIndicatorTestFixturesMinimal(t)
	})
}

func assertIndicatorOwnerLocalHelpersPrivate(t testing.TB) {
	t.Helper()
	assertFunctionInventory(t, filepath.Join("internal", "identity", "identity.go"),
		[]string{"normalizeIndicatorType", "normalizeValueKind", "normalizeValue", "isIPType", "dedupeKey"},
		[]string{"NormalizeIndicatorType", "NormalizeValueKind", "NormalizeValue", "IsIPType", "DedupeKey"},
	)
	assertFunctionInventory(t, filepath.Join("workbookprojection", "contribution.go"),
		[]string{"descriptor", "surfaceIntent"},
		[]string{"Descriptor", "SurfaceIntent"},
	)
}

func assertFunctionInventory(t testing.TB, path string, required []string, forbidden []string) {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(path), nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	functions := make(map[string]struct{})
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv == nil {
			functions[function.Name.Name] = struct{}{}
		}
	}
	for _, name := range required {
		if _, ok := functions[name]; !ok {
			t.Fatalf("%s is missing private helper %s", path, name)
		}
	}
	for _, name := range forbidden {
		if _, ok := functions[name]; ok {
			t.Fatalf("%s retains test-only exported helper %s", path, name)
		}
	}
}

func assertIndicatorTestFixturesMinimal(t testing.TB) {
	t.Helper()
	path := filepath.Join("testsupport", "fixtures.go")
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(path), nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	functions := make(map[string]struct{})
	var exampleFields []string
	for _, declaration := range parsed.Decls {
		switch typed := declaration.(type) {
		case *ast.FuncDecl:
			if typed.Recv == nil {
				functions[typed.Name.Name] = struct{}{}
			}
		case *ast.GenDecl:
			for _, specification := range typed.Specs {
				switch spec := specification.(type) {
				case *ast.ValueSpec:
					for _, name := range spec.Names {
						switch name.Name {
						case "Examples", "BaseTime", "PastTime":
							t.Fatalf("%s retains mutable fixture global %s", path, name.Name)
						}
					}
				case *ast.TypeSpec:
					if spec.Name.Name != "Example" {
						continue
					}
					structure, ok := spec.Type.(*ast.StructType)
					if !ok {
						t.Fatalf("%s Example is not a struct", path)
					}
					for _, field := range structure.Fields.List {
						for _, name := range field.Names {
							exampleFields = append(exampleFields, name.Name)
						}
					}
				}
			}
		}
	}
	wantFields := []string{"IndicatorType", "ValueKind", "DisplayValue", "NormalizedValue"}
	if strings.Join(exampleFields, " ") != strings.Join(wantFields, " ") {
		t.Fatalf("%s Example fields = %v, want %v", path, exampleFields, wantFields)
	}
	for _, name := range []string{"PrimaryExample", "BaseTime", "PastTime"} {
		if _, ok := functions[name]; !ok {
			t.Fatalf("%s is missing value-returning fixture %s", path, name)
		}
	}
}

func indicatorExportRoleInventory(groups map[string]string) map[string]string {
	roles := make(map[string]string)
	for role, declarations := range groups {
		for _, declaration := range strings.Fields(declarations) {
			if _, exists := roles[declaration]; exists {
				panic("duplicate Indicator export role: " + declaration)
			}
			roles[declaration] = role
		}
	}
	return roles
}

func indicatorExportRoleKeys() map[string]struct{} {
	result := make(map[string]struct{}, len(indicatorExportRoles))
	for declaration := range indicatorExportRoles {
		result[declaration] = struct{}{}
	}
	return result
}

func exportedRootDeclarations(t *testing.T) map[string]struct{} {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Indicator package: %v", err)
	}
	fset := token.NewFileSet()
	result := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Clean(entry.Name()), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		for _, declaration := range file.Decls {
			switch typed := declaration.(type) {
			case *ast.GenDecl:
				for _, specification := range typed.Specs {
					switch spec := specification.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(spec.Name.Name) {
							result[spec.Name.Name] = struct{}{}
						}
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							if ast.IsExported(name.Name) {
								result[name.Name] = struct{}{}
							}
						}
					}
				}
			case *ast.FuncDecl:
				if !ast.IsExported(typed.Name.Name) {
					continue
				}
				name := typed.Name.Name
				if typed.Recv != nil {
					receiver := exportedReceiverName(typed.Recv.List[0].Type)
					if receiver == "" {
						continue
					}
					name = receiver + "." + name
				}
				result[name] = struct{}{}
			}
		}
	}
	return result
}

func exportedReceiverName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		if ast.IsExported(typed.Name) {
			return typed.Name
		}
	case *ast.StarExpr:
		return exportedReceiverName(typed.X)
	}
	return ""
}

func difference(left map[string]struct{}, right map[string]struct{}) []string {
	var result []string
	for key := range left {
		if _, ok := right[key]; !ok {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}
