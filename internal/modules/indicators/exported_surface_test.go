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

type iteration2SurfaceDisposition string

const (
	iteration2RetainedContract       iteration2SurfaceDisposition = "retained_contract"
	iteration2RequiredProductionPath iteration2SurfaceDisposition = "required_production_path"
	iteration2InternalOnlyCandidate  iteration2SurfaceDisposition = "internal_only_candidate"
	iteration2RemovableDeadSurface   iteration2SurfaceDisposition = "removable_dead_surface"
)

// iteration2ExportedSurface is the SL-09 reachability decision ledger. It is
// deliberately exhaustive: adding or removing a root export requires an
// explicit classification change. SL-11 replaces this current-state ledger
// with the smaller target-state allowlist.
var iteration2ExportedSurface = map[string]iteration2SurfaceDisposition{
	"APIImportObservationOrigin":               iteration2RemovableDeadSurface,
	"APIImportObservationProducer":             iteration2RemovableDeadSurface,
	"BuildMutationPayload":                     iteration2RemovableDeadSurface,
	"CSVImportObservationOrigin":               iteration2RemovableDeadSurface,
	"CSVImportObservationProducer":             iteration2RemovableDeadSurface,
	"ClipboardPasteObservationOrigin":          iteration2RemovableDeadSurface,
	"ClipboardPasteObservationProducer":        iteration2RemovableDeadSurface,
	"CreateCommand":                            iteration2RetainedContract,
	"ErrIndicatorNotFound":                     iteration2RetainedContract,
	"ErrIndicatorObservationNotFound":          iteration2RequiredProductionPath,
	"ErrInvalidCreateRequest":                  iteration2RetainedContract,
	"ErrInvalidObservationOrigin":              iteration2InternalOnlyCandidate,
	"ExtractionObservationOrigin":              iteration2RemovableDeadSurface,
	"ExtractionObservationProducer":            iteration2RemovableDeadSurface,
	"ImportCreateCommand":                      iteration2RemovableDeadSurface,
	"IncidentBundleContribution":               iteration2RetainedContract,
	"IndicatorCreateValidationError":           iteration2RetainedContract,
	"IndicatorCreateValidationError.Error":     iteration2RetainedContract,
	"IndicatorFindOrCreateParticipantCommand":  iteration2RetainedContract,
	"IndicatorFindOrCreateParticipantResult":   iteration2RetainedContract,
	"IndicatorFindOrCreateParticipantV1":       iteration2RetainedContract,
	"IndicatorLifecycleAppendParams":           iteration2RequiredProductionPath,
	"IndicatorLifecycleIntervalRecord":         iteration2RequiredProductionPath,
	"IndicatorObservationCreateParams":         iteration2RequiredProductionPath,
	"IndicatorObservationRecord":               iteration2RequiredProductionPath,
	"IndicatorObservationResolveParams":        iteration2RequiredProductionPath,
	"IndicatorRecord":                          iteration2RemovableDeadSurface,
	"ManualEntryObservationOrigin":             iteration2RemovableDeadSurface,
	"ManualEntryObservationProducer":           iteration2RequiredProductionPath,
	"MutationResult":                           iteration2RemovableDeadSurface,
	"NewImportCreateFacade":                    iteration2RetainedContract,
	"NewIncidentBundleContribution":            iteration2RetainedContract,
	"NewProjectionContribution":                iteration2RetainedContract,
	"NewRevisionContribution":                  iteration2RetainedContract,
	"NewStore":                                 iteration2RetainedContract,
	"ObservationOrigin":                        iteration2InternalOnlyCandidate,
	"ObservationProducerContext":               iteration2InternalOnlyCandidate,
	"ParseObservationOrigin":                   iteration2InternalOnlyCandidate,
	"ProjectionContribution":                   iteration2RetainedContract,
	"ProjectionContribution.QuerySurfaces":     iteration2RetainedContract,
	"ProjectionContribution.Source":            iteration2RetainedContract,
	"ProjectionPort":                           iteration2RetainedContract,
	"RecoveryStateContribution":                iteration2RetainedContract,
	"Store":                                    iteration2RetainedContract,
	"Store.AppendIndicatorLifecycleInterval":   iteration2RequiredProductionPath,
	"Store.CreateImportRowTx":                  iteration2RemovableDeadSurface,
	"Store.CreateIndicatorObservation":         iteration2RequiredProductionPath,
	"Store.CreateIndicatorRow":                 iteration2RetainedContract,
	"Store.FindOrCreateIndicatorParticipantTx": iteration2RetainedContract,
	"Store.GetActiveIndicatorParticipant":      iteration2RetainedContract,
	"Store.GetActiveIndicatorParticipantTx":    iteration2RetainedContract,
	"Store.ResolveIndicatorObservation":        iteration2RequiredProductionPath,
	"StoreDependencies":                        iteration2RetainedContract,
	"SystemObservationOrigin":                  iteration2RemovableDeadSurface,
	"ValidateCreateCommand":                    iteration2RetainedContract,
	"ViewSchemaID":                             iteration2RetainedContract,
	"XLSXImportObservationOrigin":              iteration2RemovableDeadSurface,
	"XLSXImportObservationProducer":            iteration2RemovableDeadSurface,
}

func TestIndicatorExportedSurfaceReachabilityLock(t *testing.T) {
	t.Parallel()

	actual := exportedRootDeclarations(t)
	missingClassifications := difference(actual, iteration2ExportedSurface)
	staleClassifications := difference(iteration2ExportedSurface, actual)
	if len(missingClassifications) != 0 || len(staleClassifications) != 0 {
		t.Fatalf("Indicator exported surface changed without an SL-09 disposition: unclassified=%v stale=%v", missingClassifications, staleClassifications)
	}

	counts := map[iteration2SurfaceDisposition]int{}
	for symbol, disposition := range iteration2ExportedSurface {
		switch disposition {
		case iteration2RetainedContract, iteration2RequiredProductionPath, iteration2InternalOnlyCandidate, iteration2RemovableDeadSurface:
			counts[disposition]++
		default:
			t.Fatalf("exported symbol %s has invalid disposition %q", symbol, disposition)
		}
	}
	for _, disposition := range []iteration2SurfaceDisposition{
		iteration2RetainedContract,
		iteration2RequiredProductionPath,
		iteration2InternalOnlyCandidate,
		iteration2RemovableDeadSurface,
	} {
		if counts[disposition] == 0 {
			t.Fatalf("SL-09 disposition %q has no classified symbol", disposition)
		}
	}
}

func exportedRootDeclarations(t *testing.T) map[string]iteration2SurfaceDisposition {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Indicator package: %v", err)
	}
	fset := token.NewFileSet()
	result := map[string]iteration2SurfaceDisposition{}
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
							result[spec.Name.Name] = ""
						}
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							if ast.IsExported(name.Name) {
								result[name.Name] = ""
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
				result[name] = ""
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

func difference(left map[string]iteration2SurfaceDisposition, right map[string]iteration2SurfaceDisposition) []string {
	var result []string
	for key := range left {
		if _, ok := right[key]; !ok {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}
