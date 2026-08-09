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
	valid := validDependencies(t, validFoundationModel())

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

func TestNewRejectsInvalidCatalogAndIntentMatrix(t *testing.T) {
	tests := map[string]struct {
		mutate func(*foundationModel)
		want   string
	}{
		"duplicate provider": {
			mutate: func(model *foundationModel) {
				mutateFoundationDescriptor(model, "host", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.ProviderID = "timeline"
				})
			},
			want: "duplicate projection provider_id",
		},
		"duplicate table": {
			mutate: func(model *foundationModel) {
				mutateFoundationDescriptor(model, "identity", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.ProjectionTableIDs = []string{"host_grid_projection"}
				})
			},
			want: "duplicate projection table ownership",
		},
		"duplicate view": {
			mutate: func(model *foundationModel) {
				mutateFoundationDescriptor(model, "identity", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.ViewSchemaIDs = []string{"cartulary.view.hosts.v1"}
				})
			},
			want: "duplicate projection view ownership",
		},
		"missing provider": {
			mutate: func(model *foundationModel) {
				model.descriptors["entities"] = model.descriptors["entities"][:1]
			},
			want: "missing active projection provider \"identity\"",
		},
		"unsupported descriptor version": {
			mutate: func(model *foundationModel) {
				mutateFoundationDescriptor(model, "host", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.SchemaVersion = "projection_provider_descriptor.v2"
				})
			},
			want: "unsupported schema version",
		},
		"unresolved surface": {
			mutate: func(model *foundationModel) {
				model.intents["entities"] = model.intents["entities"][:1]
			},
			want: "has no semantic intent",
		},
		"invalid semantic fields": {
			mutate: func(model *foundationModel) {
				model.intents["entities"][0].FieldKeys = nil
			},
			want: "has no field keys",
		},
		"source ownership mismatch": {
			mutate: func(model *foundationModel) {
				mutateFoundationDescriptor(model, "host", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.SourceOwnerModule = "timeline"
				})
			},
			want: "source owner",
		},
		"storage ownership mismatch": {
			mutate: func(model *foundationModel) {
				mutateFoundationDescriptor(model, "host", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.ProjectionStorageOwnerModule = "entities"
				})
			},
			want: "storage owner",
		},
		"rebuild cycle": {
			mutate: func(model *foundationModel) {
				mutateFoundationDescriptor(model, "timeline", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.RebuildAfter = []string{"decision"}
				})
			},
			want: "rebuild graph has a cycle",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			model := validFoundationModel()
			test.mutate(&model)
			requireCompositionError(t, validDependencies(t, model), test.want)
		})
	}
}

func TestNewReturnsReadyPortsAndImmutableDescriptorSet(t *testing.T) {
	model := validFoundationModel()
	dependencies := validDependencies(t, model)
	ports, err := adapters.New(dependencies)
	if err != nil {
		t.Fatalf("compose projection foundation: %v", err)
	}
	if !ports.Ready() || ports.DescriptorSet().Len() != 10 {
		t.Fatalf("projection ports are incomplete: ready=%v descriptors=%d", ports.Ready(), ports.DescriptorSet().Len())
	}

	model.descriptors["entities"][0].ViewSchemaIDs[0] = "mutated.before.compose.return"
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

type foundationModel struct {
	descriptors map[string][]providercontract.ProviderDescriptor
	intents     map[string][]providercontract.SurfaceIntent
}

func validFoundationModel() foundationModel {
	providers := []struct {
		owner string
		id    string
		table string
		views []string
		after []string
	}{
		{owner: "timeline", id: "timeline", table: "timeline_grid_projection", views: []string{"cartulary.view.timeline.v2"}},
		{owner: "entities", id: "host", table: "host_grid_projection", views: []string{"cartulary.view.hosts.v1"}, after: []string{"timeline"}},
		{owner: "entities", id: "identity", table: "identity_grid_projection", views: []string{"cartulary.view.identities.v1"}, after: []string{"host"}},
		{owner: "indicators", id: "indicator", table: "indicator_grid_projection", views: []string{"cartulary.view.indicators.v1"}, after: []string{"identity"}},
		{owner: "assessments", id: "assessment", table: "assessment_grid_projection", views: []string{"cartulary.view.assessments.v1"}, after: []string{"indicator"}},
		{owner: "artifacts", id: "artifact", table: "artifact_grid_projection", views: []string{
			"cartulary.view.notes.v1",
			"cartulary.view.comm_log.v1",
			"cartulary.view.handoff.v1",
			"cartulary.view.status_review.v1",
			"cartulary.view.lesson.v1",
			"cartulary.view.findings.v1",
			"cartulary.view.investigative_queries.v1",
			"cartulary.view.forensic_keywords.v1",
		}, after: []string{"assessment"}},
		{owner: "evidence", id: "evidence", table: "evidence_grid_projection", views: []string{"cartulary.view.evidence.v1"}, after: []string{"artifact"}},
		{owner: "parties", id: "party", table: "party_grid_projection", views: []string{"cartulary.view.parties.v1"}, after: []string{"evidence"}},
		{owner: "tasksdecisions", id: "task_request", table: "task_request_grid_projection", views: []string{"cartulary.view.task_requests.v1"}, after: []string{"party"}},
		{owner: "tasksdecisions", id: "decision", table: "decision_grid_projection", views: []string{"cartulary.view.decisions.v1"}, after: []string{"task_request"}},
	}
	model := foundationModel{
		descriptors: map[string][]providercontract.ProviderDescriptor{},
		intents:     map[string][]providercontract.SurfaceIntent{},
	}
	for _, provider := range providers {
		model.descriptors[provider.owner] = append(model.descriptors[provider.owner], providercontract.ProviderDescriptor{
			SchemaVersion:                providercontract.DescriptorSchemaVersion,
			Status:                       providercontract.ProviderStatusActive,
			ProviderID:                   provider.id,
			SourceOwnerModule:            provider.owner,
			ViewSchemaIDs:                append([]string(nil), provider.views...),
			SourceRecordTypes:            []string{provider.id},
			SourceAuthorityModules:       []string{provider.owner},
			ProjectionTableIDs:           []string{provider.table},
			ProjectionStorageOwnerModule: "projections",
			Capabilities: providercontract.ProviderCapabilities{
				Query:           true,
				RefreshRow:      true,
				RestoreRebuild:  true,
				IncidentRebuild: true,
			},
			RestoreRebuild: providercontract.RestoreRebuildRequired,
			FacadePackages: []string{"internal/modules/" + provider.owner + "/workbookprojection"},
			RebuildAfter:   append([]string(nil), provider.after...),
		})
		for _, viewSchemaID := range provider.views {
			model.intents[provider.owner] = append(model.intents[provider.owner], providercontract.SurfaceIntent{
				ViewSchemaID: viewSchemaID,
				FieldKeys:    []string{provider.id + ".record_id"},
			})
		}
	}
	return model
}

func validDependencies(t testing.TB, model foundationModel) adapters.Dependencies {
	t.Helper()
	return adapters.Dependencies{
		Postgres:       &inertDB{},
		Timeline:       mustTimelineContribution(t, model),
		Entities:       mustEntitiesContribution(t, model),
		Indicators:     mustIndicatorsContribution(t, model),
		Assessments:    mustAssessmentsContribution(t, model),
		Artifacts:      mustArtifactsContribution(t, model),
		Evidence:       mustEvidenceContribution(t, model),
		Parties:        mustPartiesContribution(t, model),
		TasksDecisions: mustTasksDecisionsContribution(t, model),
	}
}

func mustTimelineContribution(t testing.TB, model foundationModel) timelineprojection.Contribution {
	t.Helper()
	contribution, err := timelineprojection.NewContribution(model.descriptors["timeline"], model.intents["timeline"], &inertTimelineSource{})
	if err != nil {
		t.Fatalf("timeline contribution: %v", err)
	}
	return contribution
}

func mustEntitiesContribution(t testing.TB, model foundationModel) entityprojection.Contribution {
	t.Helper()
	contribution, err := entityprojection.NewContribution(model.descriptors["entities"], model.intents["entities"], &inertEntitySource{})
	if err != nil {
		t.Fatalf("entities contribution: %v", err)
	}
	return contribution
}

func mustIndicatorsContribution(t testing.TB, model foundationModel) indicatorprojection.Contribution {
	t.Helper()
	contribution, err := indicatorprojection.NewContribution(model.descriptors["indicators"], model.intents["indicators"], &inertIndicatorSource{})
	if err != nil {
		t.Fatalf("indicators contribution: %v", err)
	}
	return contribution
}

func mustAssessmentsContribution(t testing.TB, model foundationModel) assessmentprojection.Contribution {
	t.Helper()
	contribution, err := assessmentprojection.NewContribution(model.descriptors["assessments"], model.intents["assessments"], &inertAssessmentSource{})
	if err != nil {
		t.Fatalf("assessments contribution: %v", err)
	}
	return contribution
}

func mustArtifactsContribution(t testing.TB, model foundationModel) artifactprojection.Contribution {
	t.Helper()
	contribution, err := artifactprojection.NewContribution(model.descriptors["artifacts"], model.intents["artifacts"], &inertArtifactSource{})
	if err != nil {
		t.Fatalf("artifacts contribution: %v", err)
	}
	return contribution
}

func mustEvidenceContribution(t testing.TB, model foundationModel) evidenceprojection.Contribution {
	t.Helper()
	contribution, err := evidenceprojection.NewContribution(model.descriptors["evidence"], model.intents["evidence"], &inertEvidenceSource{})
	if err != nil {
		t.Fatalf("evidence contribution: %v", err)
	}
	return contribution
}

func mustPartiesContribution(t testing.TB, model foundationModel) partyprojection.Contribution {
	t.Helper()
	contribution, err := partyprojection.NewContribution(model.descriptors["parties"], model.intents["parties"], &inertPartySource{})
	if err != nil {
		t.Fatalf("parties contribution: %v", err)
	}
	return contribution
}

func mustTasksDecisionsContribution(t testing.TB, model foundationModel) taskdecisionprojection.Contribution {
	t.Helper()
	contribution, err := taskdecisionprojection.NewContribution(
		model.descriptors["tasksdecisions"],
		model.intents["tasksdecisions"],
		taskdecisionprojection.Sources{
			TaskRequests: &inertTaskRequestSource{},
			Decisions:    &inertDecisionSource{},
		},
	)
	if err != nil {
		t.Fatalf("tasks/decisions contribution: %v", err)
	}
	return contribution
}

func mutateFoundationDescriptor(
	model *foundationModel,
	providerID string,
	mutate func(*providercontract.ProviderDescriptor),
) {
	for owner, descriptors := range model.descriptors {
		for index := range descriptors {
			if descriptors[index].ProviderID == providerID {
				mutate(&descriptors[index])
				model.descriptors[owner] = descriptors
				return
			}
		}
	}
	panic("missing projection descriptor fixture " + providerID)
}

func requireCompositionError(t testing.TB, dependencies adapters.Dependencies, want string) {
	t.Helper()
	ports, err := adapters.New(dependencies)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("adapters.New ports=%#v error=%v, want containing %q", ports, err, want)
	}
	if ports.Ready() || ports.DescriptorSet().Len() != 0 {
		t.Fatalf("failed composition returned usable ports: %#v", ports)
	}
}
