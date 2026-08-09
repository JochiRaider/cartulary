package runtime

import (
	"reflect"
	"strings"
	"testing"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	projectionstorage "github.com/JochiRaider/cartulary/internal/modules/projections/internal/storage"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type catalogDB struct{ postgres.DB }

func TestCanonicalCatalogIsExecutableRuntimeGraph(t *testing.T) {
	t.Parallel()
	fixture := validCanonicalCatalogFixture()
	catalog, err := NewCatalog(fixture.contributions(t), validCatalogSources())
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	if catalog.registry == nil || len(catalog.registry.providers) != catalog.DescriptorSet().Len() {
		t.Fatalf("canonical registry is incomplete: %#v", catalog)
	}
	for _, provider := range catalog.registry.providers {
		descriptor, ok := catalog.DescriptorSet().Lookup(provider.descriptor.ProviderID)
		if !ok || !reflect.DeepEqual(descriptor, provider.descriptor) {
			t.Fatalf("runtime descriptor drift for %q: registry=%#v set=%#v", provider.descriptor.ProviderID, provider.descriptor, descriptor)
		}
		wantStrategy := queryStrategyCompiledPlan
		if provider.descriptor.ProviderID == "host" || provider.descriptor.ProviderID == "identity" {
			wantStrategy = queryStrategySourceOwnerHydration
		}
		if provider.queryStrategy != wantStrategy {
			t.Fatalf("provider %q strategy=%d want=%d", provider.descriptor.ProviderID, provider.queryStrategy, wantStrategy)
		}
	}
	wantOrder := []string{"timeline", "host", "identity", "indicator", "assessment", "artifact", "evidence", "party", "task_request", "decision"}
	if got := providerKeys(catalog.registry.rebuildOrder); !reflect.DeepEqual(got, wantOrder) {
		t.Fatalf("canonical rebuild order=%v want=%v", got, wantOrder)
	}
}

func TestCanonicalRuntimeServicesShareOneStore(t *testing.T) {
	t.Parallel()
	sources := validCatalogSources()
	catalog, err := NewCatalog(validCanonicalCatalogFixture().contributions(t), sources)
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	db := &catalogDB{}
	physical, err := projectionstorage.New(db)
	if err != nil {
		t.Fatalf("construct physical store: %v", err)
	}
	store, err := NewStore(db, catalog, physical)
	if err != nil {
		t.Fatalf("construct canonical Store: %v", err)
	}
	if store.registry != catalog.registry || store.physical != physical {
		t.Fatalf("canonical store is incomplete: %#v", store)
	}

	timelineRows := NewTimelineRowsFromStore(store)
	entityRows := NewEntityRowsFromStore(store, sources.Entities)
	indicatorRows := NewIndicatorRowsFromStore(store, sources.Indicators)
	assessmentRows := NewAssessmentRowsFromStore(store, sources.Assessments)
	artifactRows := NewArtifactRowsFromStore(store, sources.Artifacts)
	evidenceRows := NewEvidenceRowsFromStore(store, sources.Evidence)
	partyRows := NewPartyRowsFromStore(store, sources.Parties)
	taskDecisionRows := NewTaskDecisionRowsFromStore(store, sources.TaskRequests, sources.Decisions)
	if timelineRows.store != store || entityRows.store != store || indicatorRows.store != store ||
		assessmentRows.store != store || artifactRows.store != store || evidenceRows.store != store ||
		partyRows.store != store || taskDecisionRows.store != store || NewRestoreRebuilderFromStore(store).store != store {
		t.Fatal("projection service graph does not share one canonical store")
	}
}

func TestCanonicalStoreRejectsIncompleteConstruction(t *testing.T) {
	t.Parallel()
	db := &catalogDB{}
	physical, err := projectionstorage.New(db)
	if err != nil {
		t.Fatalf("construct physical store: %v", err)
	}
	catalog, err := NewCatalog(validCanonicalCatalogFixture().contributions(t), validCatalogSources())
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	for name, build := range map[string]func() error{
		"database": func() error { _, err := NewStore(nil, catalog, physical); return err },
		"catalog":  func() error { _, err := NewStore(db, nil, physical); return err },
		"physical": func() error { _, err := NewStore(db, catalog, nil); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := build(); err == nil {
				t.Fatalf("incomplete %s construction succeeded", name)
			}
		})
	}
}

func TestCanonicalCatalogRejectsInvalidContractFacts(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*canonicalCatalogFixture)
		want   string
	}{
		"duplicate provider": {
			mutate: func(fixture *canonicalCatalogFixture) {
				mutateCanonicalDescriptor(fixture, "host", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.ProviderID = "timeline"
				})
			},
			want: "duplicate projection provider_id",
		},
		"duplicate table": {
			mutate: func(fixture *canonicalCatalogFixture) {
				mutateCanonicalDescriptor(fixture, "identity", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.ProjectionTableIDs = []string{"host_grid_projection"}
				})
			},
			want: "duplicate projection table ownership",
		},
		"duplicate view": {
			mutate: func(fixture *canonicalCatalogFixture) {
				mutateCanonicalDescriptor(fixture, "identity", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.ViewSchemaIDs = []string{"cartulary.view.hosts.v1"}
				})
			},
			want: "duplicate projection view ownership",
		},
		"missing provider": {
			mutate: func(fixture *canonicalCatalogFixture) {
				fixture.descriptors["entities"] = fixture.descriptors["entities"][:1]
			},
			want: "missing active projection provider \"identity\"",
		},
		"unsupported descriptor version": {
			mutate: func(fixture *canonicalCatalogFixture) {
				mutateCanonicalDescriptor(fixture, "host", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.SchemaVersion = "projection_provider_descriptor.v2"
				})
			},
			want: "unsupported schema_version",
		},
		"unresolved surface": {
			mutate: func(fixture *canonicalCatalogFixture) {
				fixture.intents["entities"] = fixture.intents["entities"][:1]
			},
			want: "has no semantic intent",
		},
		"invalid semantic fields": {
			mutate: func(fixture *canonicalCatalogFixture) {
				fixture.intents["entities"][0].FieldKeys = nil
			},
			want: "has no field keys",
		},
		"source ownership mismatch": {
			mutate: func(fixture *canonicalCatalogFixture) {
				mutateCanonicalDescriptor(fixture, "host", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.SourceOwnerModule = "timeline"
				})
			},
			want: "source owner",
		},
		"storage ownership mismatch": {
			mutate: func(fixture *canonicalCatalogFixture) {
				mutateCanonicalDescriptor(fixture, "host", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.ProjectionStorageOwnerModule = "entities"
				})
			},
			want: "projection_storage_owner_module",
		},
		"rebuild cycle": {
			mutate: func(fixture *canonicalCatalogFixture) {
				mutateCanonicalDescriptor(fixture, "timeline", func(descriptor *providercontract.ProviderDescriptor) {
					descriptor.RebuildAfter = []string{"decision"}
				})
			},
			want: "rebuild graph has a cycle",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixture := validCanonicalCatalogFixture()
			test.mutate(&fixture)
			err := validateCanonicalCatalogFixture(t, fixture)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewCatalog error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestCanonicalCatalogRejectsMissingProviderSources(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*ProviderSources)
	}{
		{name: "Timeline", mutate: func(sources *ProviderSources) { sources.Timeline = nil }},
		{name: "Entities", mutate: func(sources *ProviderSources) { sources.Entities = nil }},
		{name: "Indicators", mutate: func(sources *ProviderSources) { sources.Indicators = nil }},
		{name: "Assessments", mutate: func(sources *ProviderSources) { sources.Assessments = nil }},
		{name: "Artifacts", mutate: func(sources *ProviderSources) { sources.Artifacts = nil }},
		{name: "Evidence", mutate: func(sources *ProviderSources) { sources.Evidence = nil }},
		{name: "Parties", mutate: func(sources *ProviderSources) { sources.Parties = nil }},
		{name: "Task requests", mutate: func(sources *ProviderSources) { sources.TaskRequests = nil }},
		{name: "Decisions", mutate: func(sources *ProviderSources) { sources.Decisions = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			sources := validCatalogSources()
			test.mutate(&sources)
			_, err := NewCatalog(validCanonicalCatalogFixture().contributions(t), sources)
			if err == nil || !strings.Contains(err.Error(), "projection "+test.name+" source is required") {
				t.Fatalf("NewCatalog error = %v", err)
			}
		})
	}
}

func validateCanonicalCatalogFixture(t testing.TB, fixture canonicalCatalogFixture) error {
	t.Helper()
	_, err := NewCatalog(fixture.contributions(t), validCatalogSources())
	return err
}

func validCatalogSources() ProviderSources {
	return ProviderSources{
		Timeline:     &catalogTimelineSource{},
		Entities:     &catalogEntitySource{},
		Indicators:   &catalogIndicatorSource{},
		Assessments:  &catalogAssessmentSource{},
		Artifacts:    &catalogArtifactSource{},
		Evidence:     &catalogEvidenceSource{},
		Parties:      &catalogPartySource{},
		TaskRequests: &catalogTaskRequestSource{},
		Decisions:    &catalogDecisionSource{},
	}
}

type catalogTimelineSource struct{ TimelineSource }
type catalogEntitySource struct{ entityprojection.SourceReader }
type catalogIndicatorSource struct {
	indicatorprojection.SourceReader
}
type catalogAssessmentSource struct {
	assessmentprojection.SourceReader
}
type catalogArtifactSource struct {
	artifactprojection.SourceReader
}
type catalogEvidenceSource struct {
	evidenceprojection.SourceReader
}
type catalogPartySource struct{ partyprojection.SourceReader }
type catalogTaskRequestSource struct {
	taskdecisionprojection.TaskRequestSourceReader
}
type catalogDecisionSource struct {
	taskdecisionprojection.DecisionSourceReader
}

type canonicalCatalogFixture struct {
	descriptors map[string][]providercontract.ProviderDescriptor
	intents     map[string][]providercontract.SurfaceIntent
}

func validCanonicalCatalogFixture() canonicalCatalogFixture {
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
	fixture := canonicalCatalogFixture{
		descriptors: map[string][]providercontract.ProviderDescriptor{},
		intents:     map[string][]providercontract.SurfaceIntent{},
	}
	for _, provider := range providers {
		fixture.descriptors[provider.owner] = append(fixture.descriptors[provider.owner], providercontract.ProviderDescriptor{
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
				Query: true, RefreshRow: true, RestoreRebuild: true, IncidentRebuild: true,
			},
			RestoreRebuild: providercontract.RestoreRebuildRequired,
			FacadePackages: []string{"internal/modules/" + provider.owner + "/workbookprojection"},
			RebuildAfter:   append([]string(nil), provider.after...),
		})
		for _, viewSchemaID := range provider.views {
			schema, ok := viewschema.Lookup(viewSchemaID)
			if !ok {
				panic("missing projection schema fixture " + viewSchemaID)
			}
			fieldKeys := make([]string, 0, len(schema.Fields()))
			for fieldKey := range schema.Fields() {
				fieldKeys = append(fieldKeys, fieldKey)
			}
			fixture.intents[provider.owner] = append(fixture.intents[provider.owner], providercontract.SurfaceIntent{
				ViewSchemaID: viewSchemaID,
				FieldKeys:    fieldKeys,
			})
		}
	}
	return fixture
}

func (fixture canonicalCatalogFixture) contributions(t testing.TB) []providercontract.Contribution {
	t.Helper()
	owners := []string{"timeline", "entities", "indicators", "assessments", "artifacts", "evidence", "parties", "tasksdecisions"}
	contributions := make([]providercontract.Contribution, 0, len(owners))
	for _, owner := range owners {
		contribution, err := providercontract.NewContribution(owner, fixture.descriptors[owner], fixture.intents[owner])
		if err != nil {
			t.Fatalf("construct %s fixture contribution: %v", owner, err)
		}
		contributions = append(contributions, contribution)
	}
	return contributions
}

func mutateCanonicalDescriptor(
	fixture *canonicalCatalogFixture,
	providerID string,
	mutate func(*providercontract.ProviderDescriptor),
) {
	for owner, descriptors := range fixture.descriptors {
		for index := range descriptors {
			if descriptors[index].ProviderID == providerID {
				mutate(&descriptors[index])
				fixture.descriptors[owner] = descriptors
				return
			}
		}
	}
	panic("missing projection descriptor fixture " + providerID)
}
