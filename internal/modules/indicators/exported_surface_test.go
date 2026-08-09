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
	"CreateOutcome":                               {},
	"CreateOutcomeCreated":                        {},
	"CreateOutcomeReplayed":                       {},
	"CreateOutcomeReused":                         {},
	"CreateOutcomeUpdated":                        {},
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
	"IndicatorFindOrCreateParticipantV1":          {},
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

func TestIndicatorExportedSurfaceReachabilityLock(t *testing.T) {
	t.Parallel()

	actual := exportedRootDeclarations(t)
	missingClassifications := difference(actual, indicatorExportedSurfaceAllowlist)
	staleClassifications := difference(indicatorExportedSurfaceAllowlist, actual)
	if len(missingClassifications) != 0 || len(staleClassifications) != 0 {
		t.Fatalf("Indicator exported surface violates the production allowlist: unapproved=%v stale=%v", missingClassifications, staleClassifications)
	}
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
