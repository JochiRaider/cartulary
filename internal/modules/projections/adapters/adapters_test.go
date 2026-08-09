package adapters_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type inertDB struct {
	postgres.DB
}

type inertTimelineSource struct {
	timelineprojection.SourceReader
}
type inertEntitySource struct{ entityprojection.SourceReader }
type inertIndicatorSource struct {
	indicatorprojection.SourceReader
}
type inertAssessmentSource struct {
	assessmentprojection.SourceReader
}
type inertArtifactSource struct {
	artifactprojection.SourceReader
}
type inertEvidenceSource struct {
	evidenceprojection.SourceReader
}
type inertPartySource struct{ partyprojection.SourceReader }
type inertTaskRequestSource struct {
	taskdecisionprojection.TaskRequestSourceReader
}
type inertDecisionSource struct {
	taskdecisionprojection.DecisionSourceReader
}

func TestNewRejectsMissingDependencies(t *testing.T) {
	valid := validDependencies(t)

	t.Run("Postgres", func(t *testing.T) {
		dependencies := valid
		dependencies.Postgres = nil
		requireCompositionError(t, dependencies, "Postgres is required")
	})

	tests := map[string]func(*adapters.Dependencies){
		"Timeline": func(dependencies *adapters.Dependencies) {
			dependencies.Timeline = timelineprojection.Contribution{}
		},
		"Entities": func(dependencies *adapters.Dependencies) {
			dependencies.Entities = entityprojection.Contribution{}
		},
		"Indicators": func(dependencies *adapters.Dependencies) {
			dependencies.Indicators = indicatorprojection.Contribution{}
		},
		"Assessments": func(dependencies *adapters.Dependencies) {
			dependencies.Assessments = assessmentprojection.Contribution{}
		},
		"Artifacts": func(dependencies *adapters.Dependencies) {
			dependencies.Artifacts = artifactprojection.Contribution{}
		},
		"Evidence": func(dependencies *adapters.Dependencies) {
			dependencies.Evidence = evidenceprojection.Contribution{}
		},
		"Parties": func(dependencies *adapters.Dependencies) {
			dependencies.Parties = partyprojection.Contribution{}
		},
		"TasksDecisions": func(dependencies *adapters.Dependencies) {
			dependencies.TasksDecisions = taskdecisionprojection.Contribution{}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			dependencies := valid
			mutate(&dependencies)
			requireCompositionError(t, dependencies, name+" contribution is required")
		})
	}
}

func TestNewReturnsReadyPortsAndImmutableDescriptorSet(t *testing.T) {
	dependencies := validDependencies(t)
	ports, err := adapters.New(dependencies)
	if err != nil {
		t.Fatalf("compose projection foundation: %v", err)
	}
	if ports.DescriptorSet().Len() != 10 || !ports.RecoveryPorts().Ready() ||
		ports.Timeline().Writer == nil || ports.Timeline().Rebuilder == nil ||
		ports.Entities().Writer == nil || ports.Entities().Rebuilder == nil || ports.Entities().Reader == nil ||
		ports.Indicators().Rows == nil || ports.Indicators().Rebuilder == nil ||
		ports.Assessments().Rows == nil || ports.Assessments().Rebuilder == nil ||
		ports.Artifacts().Rows == nil || ports.Artifacts().Rebuilder == nil || ports.Artifacts().Reader == nil ||
		ports.Evidence().Rows == nil || ports.Evidence().Rebuilder == nil ||
		ports.Parties().Rows == nil || ports.Parties().Rebuilder == nil ||
		ports.TasksDecisions().Rows == nil || ports.TasksDecisions().Rebuilder == nil || ports.TasksDecisions().Reader == nil {
		t.Fatalf("projection ports are incomplete: %#v", ports)
	}

	first := ports.DescriptorSet().All()
	first[0].ViewSchemaIDs[0] = "mutated.returned.copy"
	second := ports.DescriptorSet().All()
	if reflect.DeepEqual(first, second) {
		t.Fatalf("descriptor set returned shared mutable data: %#v", second)
	}
	host, ok := ports.DescriptorSet().Lookup("host")
	if !ok || !reflect.DeepEqual(host.ViewSchemaIDs, []string{"cartulary.view.hosts.v1"}) {
		t.Fatalf("immutable host descriptor = %#v, found=%v", host, ok)
	}
	host.ViewSchemaIDs[0] = "mutated.lookup.copy"
	again, _ := ports.DescriptorSet().Lookup("host")
	if !reflect.DeepEqual(again.ViewSchemaIDs, []string{"cartulary.view.hosts.v1"}) {
		t.Fatalf("descriptor lookup exposed mutable state: %#v", again)
	}

	recoveryState := ports.RecoveryStateContribution()
	tableIDs := make([]string, 0, len(recoveryState.Tables))
	for _, table := range recoveryState.Tables {
		tableIDs = append(tableIDs, table.TableName)
	}
	slices.Sort(tableIDs)
	if !slices.Equal(tableIDs, providercontract.RecoveryProjectionTableIDs()) {
		t.Fatalf("recovery state tables = %v", tableIDs)
	}
	recoveryState.Tables[0].TableName = "mutated_grid_projection"
	if ports.RecoveryStateContribution().Tables[0].TableName == "mutated_grid_projection" {
		t.Fatalf("recovery state contribution exposed mutable adapter state")
	}
}

func TestFoundationProviderContractContainsNoPhysicalQueryFields(t *testing.T) {
	for _, test := range []struct {
		contractType reflect.Type
		wantFields   []string
	}{
		{
			contractType: reflect.TypeOf(providercontract.ProviderDescriptor{}),
			wantFields: []string{
				"Capabilities",
				"CharacterizationRefs",
				"FacadePackages",
				"ProjectionStorageOwnerModule",
				"ProjectionTableIDs",
				"ProviderID",
				"RebuildAfter",
				"RestoreRebuild",
				"SchemaVersion",
				"SourceAuthorityModules",
				"SourceOwnerModule",
				"SourceRecordTypes",
				"Status",
				"ViewSchemaIDs",
			},
		},
		{
			contractType: reflect.TypeOf(providercontract.SurfaceIntent{}),
			wantFields:   []string{"CanonicalSourceFilter", "FieldKeys", "ViewSchemaID"},
		},
	} {
		gotFields := make([]string, 0, test.contractType.NumField())
		for index := 0; index < test.contractType.NumField(); index++ {
			gotFields = append(gotFields, test.contractType.Field(index).Name)
		}
		slices.Sort(gotFields)
		slices.Sort(test.wantFields)
		if !reflect.DeepEqual(gotFields, test.wantFields) {
			t.Fatalf("%s fields changed:\ngot  %#v\nwant %#v", test.contractType.Name(), gotFields, test.wantFields)
		}
	}
}

func validDependencies(t testing.TB) adapters.Dependencies {
	t.Helper()
	return adapters.Dependencies{
		Postgres:       &inertDB{},
		Timeline:       mustTimelineContribution(t),
		Entities:       mustEntitiesContribution(t),
		Indicators:     mustIndicatorsContribution(t),
		Assessments:    mustAssessmentsContribution(t),
		Artifacts:      mustArtifactsContribution(t),
		Evidence:       mustEvidenceContribution(t),
		Parties:        mustPartiesContribution(t),
		TasksDecisions: mustTasksDecisionsContribution(t),
	}
}

func mustTimelineContribution(t testing.TB) timelineprojection.Contribution {
	t.Helper()
	contribution, err := timelineprojection.NewContribution(&inertTimelineSource{})
	if err != nil {
		t.Fatalf("timeline contribution: %v", err)
	}
	return contribution
}

func mustEntitiesContribution(t testing.TB) entityprojection.Contribution {
	t.Helper()
	contribution, err := entityprojection.NewContribution(&inertEntitySource{})
	if err != nil {
		t.Fatalf("entities contribution: %v", err)
	}
	return contribution
}

func mustIndicatorsContribution(t testing.TB) indicatorprojection.Contribution {
	t.Helper()
	contribution, err := indicatorprojection.NewContribution(&inertIndicatorSource{})
	if err != nil {
		t.Fatalf("indicators contribution: %v", err)
	}
	return contribution
}

func mustAssessmentsContribution(t testing.TB) assessmentprojection.Contribution {
	t.Helper()
	contribution, err := assessmentprojection.NewContribution(&inertAssessmentSource{})
	if err != nil {
		t.Fatalf("assessments contribution: %v", err)
	}
	return contribution
}

func mustArtifactsContribution(t testing.TB) artifactprojection.Contribution {
	t.Helper()
	contribution, err := artifactprojection.NewContribution(&inertArtifactSource{})
	if err != nil {
		t.Fatalf("artifacts contribution: %v", err)
	}
	return contribution
}

func mustEvidenceContribution(t testing.TB) evidenceprojection.Contribution {
	t.Helper()
	contribution, err := evidenceprojection.NewContribution(&inertEvidenceSource{})
	if err != nil {
		t.Fatalf("evidence contribution: %v", err)
	}
	return contribution
}

func mustPartiesContribution(t testing.TB) partyprojection.Contribution {
	t.Helper()
	contribution, err := partyprojection.NewContribution(&inertPartySource{})
	if err != nil {
		t.Fatalf("parties contribution: %v", err)
	}
	return contribution
}

func mustTasksDecisionsContribution(t testing.TB) taskdecisionprojection.Contribution {
	t.Helper()
	contribution, err := taskdecisionprojection.NewContribution(
		&inertTaskRequestSource{},
		&inertDecisionSource{},
	)
	if err != nil {
		t.Fatalf("tasks/decisions contribution: %v", err)
	}
	return contribution
}

func requireCompositionError(t testing.TB, dependencies adapters.Dependencies, want string) {
	t.Helper()
	ports, err := adapters.New(dependencies)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("adapters.New ports=%#v error=%v, want containing %q", ports, err, want)
	}
	if ports.DescriptorSet().Len() != 0 || ports.RecoveryPorts().Ready() {
		t.Fatalf("failed composition returned usable ports: %#v", ports)
	}
}
