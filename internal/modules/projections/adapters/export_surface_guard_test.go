package adapters

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type exportAllowance struct {
	production  map[string]struct{}
	testFixture map[string]struct{}
}

var projectionExportAllowlists = map[string]exportAllowance{
	".": newExportAllowance(`
		Dependencies MaintenanceRebuilder Ports SourceTextRows New
		Ports.DescriptorSet Ports.WorkbookQueryProvider Ports.RecoveryPorts
		Ports.MaintenanceRebuilder Ports.RestoreProbeQuery Ports.RevisionRebuilder
		Ports.RevisionLiveRecords Ports.SourceTextRows Ports.ImportRebuilder
		Ports.Timeline Ports.Entities Ports.Indicators Ports.Assessments
		Ports.Artifacts Ports.Evidence Ports.Parties Ports.TaskDecisionMutationRows
		Ports.TaskDecisionReportingReader
	`, ``),
	"../providercontract": newExportAllowance(`
		DescriptorSchemaVersion ProviderStatus ProviderStatusActive
		ProviderStatusDeprecated ProviderStatusExperimental RestoreRebuildParticipation
		RestoreRebuildRequired RestoreRebuildNonparticipating RestoreRebuildUnsupported
		ProviderCapabilities ProviderDescriptor ProviderDescriptor.Clone DescriptorSet
		NewDescriptorSet DescriptorSet.Len DescriptorSet.All DescriptorSet.Lookup
		DescriptorSet.RebuildOrder RecoveryRebuildAlgorithmID RecoveryProjectionTableIDs
		RecoveryStateContribution SurfaceIntent SourceFilterIntent SurfaceIntent.Clone
		Contribution NewContribution Contribution.IsZero Contribution.SourceOwnerModule
		Contribution.Descriptors Contribution.SurfaceIntents
	`, ``),
	"../internal/runtime": newExportAllowance(`
		Provider Catalog Catalog.DescriptorSet Store Store.RefreshRowTx
		Store.RebuildIncidentTx Store.RebuildImportedIncidentTx Store.Supports
		Store.QueryRows Store.QueryRowsPage Store.LoadRowTx Store.RebuildIncident
		RestoreRebuilder NewRestoreRebuilderFromStore
		RestoreRebuilder.RebuildRestoreProjections ProviderSources NewCatalog
		EvidenceAssociationEffects NewEvidenceAssociationEffectsFromStore
		EvidenceAssociationEffects.RefreshEvidenceAssociationEffects TimelineSource
		NewStore TimelineRows NewTimelineRowsFromStore TimelineRows.ApplyTimelineMutationTx
		EntityRows IndicatorRows NewIndicatorRowsFromStore NewEntityRowsFromStore
		IndicatorRows.RefreshIndicatorTx IndicatorRows.LoadIndicatorTx
		EntityRows.RefreshHostTx EntityRows.RefreshIdentityTx EntityRows.DeleteHostTx
		EntityRows.DeleteIdentityTx EntityRows.SelectHostQueryProjections
		EntityRows.CollectHostDerivedFactsTx EntityRows.SelectIdentityQueryProjections
		EntityRows.CollectIdentityDerivedFactsTx AssessmentRows NewAssessmentRowsFromStore
		AssessmentRows.RefreshAssessmentTx AssessmentRows.ApplyAssessmentMutationTx
		AssessmentRows.LoadAssessmentTx ArtifactRows NewArtifactRowsFromStore
		ArtifactRows.RefreshArtifactTx ArtifactRows.LoadArtifactTx
		ArtifactRows.CollectDerivedFactsTx EvidenceRows NewEvidenceRowsFromStore
		EvidenceRows.RefreshEvidenceTx EvidenceRows.LoadEvidenceTx PartyRows
		NewPartyRowsFromStore PartyRows.RefreshPartyTx PartyRows.LoadPartyTx
		TaskDecisionMutationRows NewTaskDecisionMutationRowsFromStore
		TaskDecisionMutationRows.RefreshTaskRequestTx TaskDecisionMutationRows.RefreshDecisionTx
		TaskDecisionMutationRows.LoadTaskRequestTx TaskDecisionMutationRows.LoadDecisionTx
		TaskDecisionReportingReader NewTaskDecisionReportingReader
		TaskDecisionReportingReader.CollectTaskDerivedFactsTx
		TaskDecisionReportingReader.CollectDecisionDerivedFactsTx
		TaskRequestSource DecisionSource
	`, `
		TimelineRows.ApplyTimelineFixtureBatchTx TimelineRows.CountTimelineFixtureRows
		TimelineRows.CountTimelineFixtureRowsTx
	`),
	"../internal/queryengine": newExportAllowance(`
		FieldKind FieldKindText FieldKindTimestamp FieldKindDate FieldKindBool
		FieldKindNumber FieldKindCollection Field Surface CompileSurface Surface.Field
		Field.OrderExpr ScanRows LoadRowTx BuildQueryPageSQL BuildRow TimelinePlans
		PartyPlans AssessmentPlans EvidencePlans IndicatorPlans ArtifactReader
		NewArtifactReader ArtifactReader.CollectDerivedFactsTx ArtifactPlans DecisionReader
		NewDecisionReader DecisionReader.CollectDecisionDerivedFactsTx DecisionPlans
		HostReader NewHostReader HostReader.SelectHostQueryProjections
		HostReader.CollectHostDerivedFactsTx IdentityReader NewIdentityReader
		IdentityReader.SelectIdentityQueryProjections IdentityReader.CollectIdentityDerivedFactsTx
		TaskReader NewTaskReader TaskReader.CollectTaskDerivedFactsTx TaskRequestPlans
	`, ``),
	"../internal/storage": newExportAllowance(`
		Store New Store.InsertPartyTx Store.DeletePartyRowTx Store.DeletePartyIncidentTx
		Store.InsertDecisionTx Store.DeleteDecisionRowTx Store.DeleteDecisionIncidentTx
		Store.UpsertAssessmentTx Store.DeleteAssessmentRowTx Store.DeleteAssessmentIncidentTx
		Store.UpsertHostTx Store.DeleteHostRowTx Store.DeleteHostIncidentTx
		Store.LoadHostEvidenceAssociationStateTx Store.InsertEvidenceTx
		Store.DeleteEvidenceRowTx Store.DeleteEvidenceIncidentTx Store.UpsertIndicatorTx
		Store.DeleteIndicatorRowTx Store.DeleteIndicatorIncidentTx Store.UpsertIdentityTx
		Store.DeleteIdentityRowTx Store.DeleteIdentityIncidentTx
		Store.LoadIdentityEvidenceAssociationStateTx Store.UpsertTimelineRowTx
		Store.DeleteTimelineRowTx Store.DeleteTimelineIncidentTx Store.InsertTaskRequestTx
		Store.DeleteTaskRequestRowTx Store.DeleteTaskRequestIncidentTx Store.InsertArtifactTx
		Store.DeleteArtifactRowTx Store.DeleteArtifactIncidentTx
	`, `
		Store.InsertTimelineFixtureBatchTx Store.CountTimelineFixtureRows
		Store.CountTimelineFixtureRowsTx
	`),
}

func TestProjectionInternalExportSurface(t *testing.T) {
	t.Parallel()
	for directory, allowance := range projectionExportAllowlists {
		directory := directory
		allowance := allowance
		t.Run(filepath.Base(directory), func(t *testing.T) {
			t.Parallel()
			if overlap := exportDifference(allowance.production, exportDifferenceSet(allowance.production, allowance.testFixture)); len(overlap) != 0 {
				t.Fatalf("production and retained test-fixture export classifications overlap: %v", overlap)
			}
			actual := exportedPackageDeclarations(t, directory)
			expected := exportUnion(allowance.production, allowance.testFixture)
			unexpected, missing := exportSurfaceDelta(actual, expected)
			if len(unexpected) != 0 || len(missing) != 0 {
				t.Fatalf("projection export surface drift: unexpected=%v missing=%v", unexpected, missing)
			}
		})
	}

	t.Run("synthetic unexpected export", func(t *testing.T) {
		actual := map[string]struct{}{"Expected": {}, "UnexpectedIteration3Export": {}}
		expected := map[string]struct{}{"Expected": {}}
		unexpected, missing := exportSurfaceDelta(actual, expected)
		if !reflect.DeepEqual(unexpected, []string{"UnexpectedIteration3Export"}) || len(missing) != 0 {
			t.Fatalf("synthetic unexpected-export result: unexpected=%v missing=%v", unexpected, missing)
		}
	})

	t.Run("synthetic missing export", func(t *testing.T) {
		actual := map[string]struct{}{"Retained": {}}
		expected := map[string]struct{}{"RequiredIteration3Export": {}, "Retained": {}}
		unexpected, missing := exportSurfaceDelta(actual, expected)
		if len(unexpected) != 0 || !reflect.DeepEqual(missing, []string{"RequiredIteration3Export"}) {
			t.Fatalf("synthetic missing-export result: unexpected=%v missing=%v", unexpected, missing)
		}
	})
}

func newExportAllowance(production string, testFixture string) exportAllowance {
	return exportAllowance{
		production:  exportWords(production),
		testFixture: exportWords(testFixture),
	}
}

func exportWords(value string) map[string]struct{} {
	result := make(map[string]struct{})
	for _, name := range strings.Fields(value) {
		result[name] = struct{}{}
	}
	return result
}

func exportedPackageDeclarations(t testing.TB, directory string) map[string]struct{} {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read projection package %s: %v", directory, err)
	}
	result := make(map[string]struct{})
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
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

func exportUnion(left map[string]struct{}, right map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{}, len(left)+len(right))
	for name := range left {
		result[name] = struct{}{}
	}
	for name := range right {
		result[name] = struct{}{}
	}
	return result
}

func exportDifferenceSet(left map[string]struct{}, right map[string]struct{}) map[string]struct{} {
	result := make(map[string]struct{}, len(left))
	for name := range left {
		if _, found := right[name]; !found {
			result[name] = struct{}{}
		}
	}
	return result
}

func exportDifference(left map[string]struct{}, right map[string]struct{}) []string {
	result := make([]string, 0)
	for name := range left {
		if _, found := right[name]; !found {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}

func exportSurfaceDelta(actual map[string]struct{}, expected map[string]struct{}) ([]string, []string) {
	return exportDifference(actual, expected), exportDifference(expected, actual)
}
