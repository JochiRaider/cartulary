package indicators

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// indicatorExportedSurfaceAllowlist is a production boundary, not a historical
// inventory. New root exports require an explicit owner-contract decision.
var indicatorExportedSurfaceAllowlist = map[string]struct{}{
	"AffectedRecordVersion":                             {},
	"CreateCommand":                                     {},
	"CreateResult":                                      {},
	"ErrIllegalTransition":                              {},
	"ErrIndicatorNotFound":                              {},
	"ErrIndicatorObservationNotFound":                   {},
	"ErrIndicatorSourceNotFound":                        {},
	"ErrInvalidCreateRequest":                           {},
	"ErrResolvedIndicatorNotFound":                      {},
	"ErrRowVersionConflict":                             {},
	"ErrSourceTextUnavailable":                          {},
	"IncidentBundleContribution":                        {},
	"IndicatorCreateValidationError":                    {},
	"IndicatorCreateValidationError.Error":              {},
	"IndicatorFindOrCreateParticipantCommand":           {},
	"IndicatorFindOrCreateParticipantResult":            {},
	"IndicatorLifecycleAppendParams":                    {},
	"IndicatorLifecycleIntervalRecord":                  {},
	"IndicatorLifecycleMutationResult":                  {},
	"IndicatorObservationActionParams":                  {},
	"IndicatorObservationCreateParams":                  {},
	"IndicatorObservationMutationResult":                {},
	"IndicatorObservationRecord":                        {},
	"IndicatorObservationResolveParams":                 {},
	"IndicatorReference":                                {},
	"IdempotencyPort":                                   {},
	"IncidentStatePort":                                 {},
	"NewImportContribution":                             {},
	"NewIncidentBundleContribution":                     {},
	"NewProjectionContribution":                         {},
	"NewRevisionContribution":                           {},
	"NewApplication":                                    {},
	"RecoveryStateContribution":                         {},
	"RecordEnvelopePort":                                {},
	"RevisionPort":                                      {},
	"SourceTextPort":                                    {},
	"SourceTextValue":                                   {},
	"Application":                                       {},
	"Application.AppendIndicatorLifecycleInterval":      {},
	"Application.DismissIndicatorObservation":           {},
	"Application.GetIndicatorObservation":               {},
	"Application.ListIndicatorLifecycleIntervals":       {},
	"Application.ListIndicatorObservations":             {},
	"Application.ListSourceRecordIndicatorObservations": {},
	"Application.ResolveIndicatorObservation":           {},
	"Application.RestoreIndicatorObservation":           {},
	"Application.CreateIndicatorObservation":            {},
	"Application.CreateIndicatorRow":                    {},
	"Application.FindOrCreateIndicatorParticipantTx":    {},
	"Application.GetActiveIndicatorParticipant":         {},
	"Application.GetActiveIndicatorParticipantTx":       {},
	"ApplicationDependencies":                           {},
	"ValidateCreateCommand":                             {},
	"ViewSchemaID":                                      {},
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
		IdempotencyPort IncidentStatePort RevisionPort SourceTextPort SourceTextValue
		ValidateCreateCommand ViewSchemaID
	`,
	"typed source-owner contribution consumed by application assembly": `
		IncidentBundleContribution NewImportContribution NewIncidentBundleContribution
		NewProjectionContribution NewRevisionContribution RecoveryStateContribution
	`,
	"complete Indicators application construction and transaction participant capability": `
		Application ApplicationDependencies NewApplication RecordEnvelopePort
	`,
	"live Indicators application operation consumed by HTTP, Workbook, Imports, or Network Flow": `
		Application.AppendIndicatorLifecycleInterval Application.CreateIndicatorObservation Application.CreateIndicatorRow
		Application.DismissIndicatorObservation Application.FindOrCreateIndicatorParticipantTx
		Application.GetActiveIndicatorParticipant Application.GetActiveIndicatorParticipantTx
		Application.GetIndicatorObservation Application.ListIndicatorLifecycleIntervals
		Application.ListIndicatorObservations Application.ListSourceRecordIndicatorObservations
		Application.ResolveIndicatorObservation Application.RestoreIndicatorObservation
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
	if len(actual) != 54 {
		t.Fatalf("Indicator exported surface contains %d declarations, want exact reviewed Iteration 3 surface of 54", len(actual))
	}
	for declaration, role := range indicatorExportRoles {
		if strings.TrimSpace(role) == "" {
			t.Fatalf("Indicator export %s has no production-role reason", declaration)
		}
	}
	t.Run("synthetic export is rejected", func(t *testing.T) {
		synthetic := make(map[string]struct{}, len(actual)+1)
		for declaration := range actual {
			synthetic[declaration] = struct{}{}
		}
		synthetic["UnapprovedIteration3Export"] = struct{}{}
		unapproved := difference(synthetic, indicatorExportedSurfaceAllowlist)
		if len(unapproved) != 1 || unapproved[0] != "UnapprovedIteration3Export" {
			t.Fatalf("synthetic export guard reported %v, want the injected declaration", unapproved)
		}
	})

	t.Run("predecessor topology is unreachable", func(t *testing.T) {
		assertIndicatorPredecessorTopologyAbsent(t)
	})
	t.Run("owner-local helpers remain private", func(t *testing.T) {
		assertIndicatorOwnerLocalHelpersPrivate(t)
	})
	t.Run("test fixtures are value-returning and minimal", func(t *testing.T) {
		assertIndicatorTestFixturesMinimal(t)
	})
}

func assertIndicatorPredecessorTopologyAbsent(t testing.TB) {
	t.Helper()

	for _, path := range []string{
		"projectionprovider",
		"store_composition.go",
		"store_composition_test.go",
		"store_test_helpers_test.go",
	} {
		if _, err := os.Stat(filepath.Clean(path)); !os.IsNotExist(err) {
			t.Fatalf("retired Indicator path %s remains: %v", path, err)
		}
	}

	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	internalRoot := filepath.Join(repoRoot, "internal")
	indicatorRootImport := indicatorsRepoImportPrefix + "internal/modules/indicators"
	indicatorHTTPImport := indicatorRootImport + "/httpapi"
	projectionWrapperImport := indicatorRootImport + "/projectionprovider"
	peerConstructors := map[string]map[string]struct{}{
		indicatorsRepoImportPrefix + "internal/platform/authn": {
			"NewStore": {},
		},
		indicatorsRepoImportPrefix + "internal/modules/incidents/admission": {
			"NewChecker": {},
		},
		indicatorsRepoImportPrefix + "internal/modules/records": {
			"NewStore": {},
		},
		indicatorsRepoImportPrefix + "internal/modules/revisions": {
			"NewAppender": {},
		},
	}
	legacySelectors := map[string]struct{}{
		"Store":                 {},
		"StoreDependencies":     {},
		"NewStore":              {},
		"NewImportCreateFacade": {},
	}
	registerCalls := 0
	err := filepath.WalkDir(internalRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		relativePath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		relativePath = filepath.ToSlash(relativePath)
		production := !strings.HasSuffix(path, "_test.go")
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		imports := indicatorParsedImportNames(t, parsed, relativePath)
		if alias, ok := imports[projectionWrapperImport]; ok {
			t.Fatalf("%s imports retired Indicator projection wrapper as %s", relativePath, alias)
		}
		indicatorAlias := imports[indicatorRootImport]
		httpAlias := imports[indicatorHTTPImport]
		authnAlias := imports[indicatorsRepoImportPrefix+"internal/platform/authn"]
		rootProduction := production && filepath.ToSlash(filepath.Dir(relativePath)) == "internal/modules/indicators"

		for _, declaration := range parsed.Decls {
			if !rootProduction {
				break
			}
			switch typed := declaration.(type) {
			case *ast.FuncDecl:
				if _, forbidden := legacySelectors[typed.Name.Name]; forbidden {
					t.Fatalf("%s declares retired Indicator function %s", relativePath, typed.Name.Name)
				}
			case *ast.GenDecl:
				for _, specification := range typed.Specs {
					typeSpec, ok := specification.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if _, forbidden := legacySelectors[typeSpec.Name.Name]; forbidden {
						t.Fatalf("%s declares retired Indicator type %s", relativePath, typeSpec.Name.Name)
					}
					if typeSpec.Assign.IsValid() {
						if target, ok := typeSpec.Type.(*ast.Ident); ok && target.Name == "Application" {
							t.Fatalf("%s aliases Application as %s", relativePath, typeSpec.Name.Name)
						}
					}
				}
			}
		}

		ast.Inspect(parsed, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if ok {
				qualifier, qualified := selector.X.(*ast.Ident)
				if qualified && indicatorAlias != "" && qualifier.Name == indicatorAlias {
					if _, forbidden := legacySelectors[selector.Sel.Name]; forbidden {
						t.Fatalf("%s reaches retired indicators.%s", relativePath, selector.Sel.Name)
					}
				}
			}

			literal, ok := node.(*ast.CompositeLit)
			if ok && production && indicatorAlias != "" && authnAlias != "" {
				if selector, ok := literal.Type.(*ast.SelectorExpr); ok {
					qualifier, qualified := selector.X.(*ast.Ident)
					if qualified && qualifier.Name == authnAlias && selector.Sel.Name == "UserRecord" {
						t.Fatalf("%s constructs a synthetic authn.UserRecord while consuming Indicators", relativePath)
					}
				}
			}

			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			called, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			qualifier, ok := called.X.(*ast.Ident)
			if !ok {
				return true
			}
			if production && httpAlias != "" && qualifier.Name == httpAlias && called.Sel.Name == "RegisterRoutes" {
				registerCalls++
			}
			if !rootProduction {
				return true
			}
			for importPath, constructors := range peerConstructors {
				if qualifier.Name != imports[importPath] {
					continue
				}
				if _, forbidden := constructors[called.Sel.Name]; forbidden {
					t.Fatalf("%s constructs concrete peer %s.%s", relativePath, qualifier.Name, called.Sel.Name)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("audit Indicator predecessor reachability: %v", err)
	}
	if registerCalls != 1 {
		t.Fatalf("Indicator HTTP route registration count = %d, want one composition-owned path", registerCalls)
	}
}

func indicatorParsedImportNames(t testing.TB, file *ast.File, relativePath string) map[string]string {
	t.Helper()
	result := make(map[string]string, len(file.Imports))
	for _, importSpec := range file.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import in %s: %v", relativePath, err)
		}
		name := filepath.Base(importPath)
		if importSpec.Name != nil {
			name = importSpec.Name.Name
		}
		result[importPath] = name
	}
	return result
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
