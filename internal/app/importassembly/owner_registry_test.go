package importassembly

import (
	"context"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

func TestOwnerCreateRegistryComposesEveryCurrentViewTarget(t *testing.T) {
	t.Parallel()

	registry, err := NewOwnerCreateRegistry(OwnerRegistryDependencies{
		Postgres:                inertOwnerRegistryDB{},
		RevisionAppender:        &revisions.Appender{},
		Intents:                 inertIntentAppender{},
		Timeline:                inertTimelineFacade(),
		EntityProjections:       inertEntityProjectionWriter{},
		AssessmentProjections:   inertAssessmentProjectionRows{},
		ArtifactProjections:     inertArtifactProjectionRows{},
		Evidence:                inertEvidenceOwnerFacade(t),
		PartyProjections:        inertPartyProjectionRows{},
		TaskDecisionProjections: inertTaskDecisionProjectionRows{},
		Indicators:              inertIndicatorOwner(),
	})
	if err != nil {
		t.Fatalf("compose owner-create registry: %v", err)
	}
	bindings := registry.Bindings()
	if len(bindings) != 14 {
		t.Fatalf("owner-create bindings = %d, want 14", len(bindings))
	}
	byTarget := make(map[string]string, len(bindings))
	for _, binding := range bindings {
		byTarget[binding.TargetViewSchemaID] = binding.FacadeID
	}
	for target, facade := range map[string]string{
		"cartulary.view.hosts.v1":         "entities.host.import_create",
		"cartulary.view.identities.v1":    "entities.identity.import_create",
		"cartulary.view.evidence.v1":      "evidence.import_create",
		"cartulary.view.indicators.v1":    "indicators.import_create",
		"cartulary.view.parties.v1":       "parties.import_create",
		"cartulary.view.task_requests.v1": "tasksdecisions.task_request.import_create",
		"cartulary.view.decisions.v1":     "tasksdecisions.decision.import_create",
		"cartulary.view.notes.v1":         "artifacts.note.import_create",
		"cartulary.view.timeline.v2":      "timeline.import_create",
		"cartulary.view.assessments.v1":   "assessments.import_create",
	} {
		if byTarget[target] != facade {
			t.Fatalf("binding for %s = %q, want %q", target, byTarget[target], facade)
		}
	}
}

func TestOwnerCreateRegistryConsumesNarrowEvidenceFacade(t *testing.T) {
	t.Parallel()

	dependencies := reflect.TypeOf(OwnerRegistryDependencies{})
	field, present := dependencies.FieldByName("Evidence")
	if !present {
		t.Fatal("owner registry dependencies omit the Evidence facade")
	}
	wantType := reflect.TypeOf((*ownerfacade.ImportOwnerCreateFacade)(nil)).Elem()
	if field.Type != wantType {
		t.Fatalf("Evidence dependency type = %v, want %v", field.Type, wantType)
	}
	if _, present := dependencies.FieldByName("EvidenceProjections"); present {
		t.Fatal("owner registry dependencies reconstruct Evidence from projection rows")
	}
}

func TestOwnerCreateRegistryRequiresCompositionDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		deps OwnerRegistryDependencies
	}{
		{
			name: "postgres",
			deps: OwnerRegistryDependencies{
				RevisionAppender:        &revisions.Appender{},
				Intents:                 inertIntentAppender{},
				Timeline:                inertTimelineFacade(),
				EntityProjections:       inertEntityProjectionWriter{},
				AssessmentProjections:   inertAssessmentProjectionRows{},
				ArtifactProjections:     inertArtifactProjectionRows{},
				Evidence:                inertEvidenceOwnerFacade(t),
				PartyProjections:        inertPartyProjectionRows{},
				TaskDecisionProjections: inertTaskDecisionProjectionRows{},
				Indicators:              inertIndicatorOwner(),
			},
		},
		{
			name: "revisions",
			deps: OwnerRegistryDependencies{
				Postgres:                inertOwnerRegistryDB{},
				Intents:                 inertIntentAppender{},
				Timeline:                inertTimelineFacade(),
				EntityProjections:       inertEntityProjectionWriter{},
				AssessmentProjections:   inertAssessmentProjectionRows{},
				ArtifactProjections:     inertArtifactProjectionRows{},
				Evidence:                inertEvidenceOwnerFacade(t),
				PartyProjections:        inertPartyProjectionRows{},
				TaskDecisionProjections: inertTaskDecisionProjectionRows{},
				Indicators:              inertIndicatorOwner(),
			},
		},
		{
			name: "intents",
			deps: OwnerRegistryDependencies{
				Postgres:                inertOwnerRegistryDB{},
				RevisionAppender:        &revisions.Appender{},
				Timeline:                inertTimelineFacade(),
				EntityProjections:       inertEntityProjectionWriter{},
				AssessmentProjections:   inertAssessmentProjectionRows{},
				ArtifactProjections:     inertArtifactProjectionRows{},
				Evidence:                inertEvidenceOwnerFacade(t),
				PartyProjections:        inertPartyProjectionRows{},
				TaskDecisionProjections: inertTaskDecisionProjectionRows{},
				Indicators:              inertIndicatorOwner(),
			},
		},
		{
			name: "timeline",
			deps: OwnerRegistryDependencies{
				Postgres:                inertOwnerRegistryDB{},
				RevisionAppender:        &revisions.Appender{},
				Intents:                 inertIntentAppender{},
				EntityProjections:       inertEntityProjectionWriter{},
				AssessmentProjections:   inertAssessmentProjectionRows{},
				ArtifactProjections:     inertArtifactProjectionRows{},
				Evidence:                inertEvidenceOwnerFacade(t),
				PartyProjections:        inertPartyProjectionRows{},
				TaskDecisionProjections: inertTaskDecisionProjectionRows{},
				Indicators:              inertIndicatorOwner(),
			},
		},
		{
			name: "entity projections",
			deps: OwnerRegistryDependencies{
				Postgres:                inertOwnerRegistryDB{},
				RevisionAppender:        &revisions.Appender{},
				Intents:                 inertIntentAppender{},
				Timeline:                inertTimelineFacade(),
				AssessmentProjections:   inertAssessmentProjectionRows{},
				ArtifactProjections:     inertArtifactProjectionRows{},
				Evidence:                inertEvidenceOwnerFacade(t),
				PartyProjections:        inertPartyProjectionRows{},
				TaskDecisionProjections: inertTaskDecisionProjectionRows{},
				Indicators:              inertIndicatorOwner(),
			},
		},
		{
			name: "assessment projections",
			deps: OwnerRegistryDependencies{
				Postgres:                inertOwnerRegistryDB{},
				RevisionAppender:        &revisions.Appender{},
				Intents:                 inertIntentAppender{},
				Timeline:                inertTimelineFacade(),
				EntityProjections:       inertEntityProjectionWriter{},
				ArtifactProjections:     inertArtifactProjectionRows{},
				Evidence:                inertEvidenceOwnerFacade(t),
				PartyProjections:        inertPartyProjectionRows{},
				TaskDecisionProjections: inertTaskDecisionProjectionRows{},
				Indicators:              inertIndicatorOwner(),
			},
		},
		{
			name: "artifact projections",
			deps: OwnerRegistryDependencies{
				Postgres:                inertOwnerRegistryDB{},
				RevisionAppender:        &revisions.Appender{},
				Intents:                 inertIntentAppender{},
				Timeline:                inertTimelineFacade(),
				EntityProjections:       inertEntityProjectionWriter{},
				AssessmentProjections:   inertAssessmentProjectionRows{},
				Evidence:                inertEvidenceOwnerFacade(t),
				PartyProjections:        inertPartyProjectionRows{},
				TaskDecisionProjections: inertTaskDecisionProjectionRows{},
				Indicators:              inertIndicatorOwner(),
			},
		},
		{
			name: "evidence owner facade",
			deps: OwnerRegistryDependencies{
				Postgres:                inertOwnerRegistryDB{},
				RevisionAppender:        &revisions.Appender{},
				Intents:                 inertIntentAppender{},
				Timeline:                inertTimelineFacade(),
				EntityProjections:       inertEntityProjectionWriter{},
				AssessmentProjections:   inertAssessmentProjectionRows{},
				ArtifactProjections:     inertArtifactProjectionRows{},
				PartyProjections:        inertPartyProjectionRows{},
				TaskDecisionProjections: inertTaskDecisionProjectionRows{},
				Indicators:              inertIndicatorOwner(),
			},
		},
		{
			name: "party projections",
			deps: OwnerRegistryDependencies{
				Postgres:                inertOwnerRegistryDB{},
				RevisionAppender:        &revisions.Appender{},
				Intents:                 inertIntentAppender{},
				Timeline:                inertTimelineFacade(),
				EntityProjections:       inertEntityProjectionWriter{},
				AssessmentProjections:   inertAssessmentProjectionRows{},
				ArtifactProjections:     inertArtifactProjectionRows{},
				Evidence:                inertEvidenceOwnerFacade(t),
				TaskDecisionProjections: inertTaskDecisionProjectionRows{},
				Indicators:              inertIndicatorOwner(),
			},
		},
		{
			name: "task/decision projections",
			deps: OwnerRegistryDependencies{
				Postgres:              inertOwnerRegistryDB{},
				RevisionAppender:      &revisions.Appender{},
				Intents:               inertIntentAppender{},
				Timeline:              inertTimelineFacade(),
				EntityProjections:     inertEntityProjectionWriter{},
				AssessmentProjections: inertAssessmentProjectionRows{},
				ArtifactProjections:   inertArtifactProjectionRows{},
				Evidence:              inertEvidenceOwnerFacade(t),
				PartyProjections:      inertPartyProjectionRows{},
				Indicators:            inertIndicatorOwner(),
			},
		},
		{
			name: "indicators",
			deps: OwnerRegistryDependencies{
				Postgres:                inertOwnerRegistryDB{},
				RevisionAppender:        &revisions.Appender{},
				Intents:                 inertIntentAppender{},
				Timeline:                inertTimelineFacade(),
				EntityProjections:       inertEntityProjectionWriter{},
				AssessmentProjections:   inertAssessmentProjectionRows{},
				ArtifactProjections:     inertArtifactProjectionRows{},
				Evidence:                inertEvidenceOwnerFacade(t),
				PartyProjections:        inertPartyProjectionRows{},
				TaskDecisionProjections: inertTaskDecisionProjectionRows{},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewOwnerCreateRegistry(test.deps); err == nil {
				t.Fatalf("missing %s dependency unexpectedly succeeded", test.name)
			}
		})
	}
}

type inertEntityProjectionWriter struct {
	entityprojection.Writer
}

type inertAssessmentProjectionRows struct {
	assessmentprojection.Rows
}

type inertArtifactProjectionRows struct {
	artifactprojection.Rows
}

type inertPartyProjectionRows struct {
	partyprojection.Rows
}

type inertTaskDecisionProjectionRows struct {
	taskdecisionprojection.Rows
}

type inertIndicatorProjectionRows struct {
	indicatorprojection.Rows
}

type inertIndicatorSourceText struct {
	indicators.SourceTextPort
}

func inertIndicatorOwner() *indicators.Store {
	owner, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:    inertOwnerRegistryDB{},
		Revisions:   &revisions.Appender{},
		Projections: inertIndicatorProjectionRows{},
		SourceText:  inertIndicatorSourceText{},
	})
	if err != nil {
		panic(err)
	}
	return owner
}

func inertTimelineFacade() *timeline.Facade {
	return timeline.NewFacade(
		inertOwnerRegistryDB{},
		timeline.Collaborators{},
		conflicttokens.ConflictTokenCodec{},
	)
}

func inertEvidenceOwnerFacade(t testing.TB) ownerfacade.ImportOwnerCreateFacade {
	t.Helper()
	facade, err := ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: "cartulary.view.evidence.v1",
			FacadeID:           "evidence.import_create",
		},
		func(context.Context, pgx.Tx, ownerfacade.ImportOwnerCreateCommand) (ownerfacade.ImportOwnerCreateResponse, error) {
			panic("owner registry construction must not invoke the Evidence facade")
		},
	)
	if err != nil {
		t.Fatalf("compose inert Evidence owner facade: %v", err)
	}
	return facade
}

type inertOwnerRegistryDB struct{}

func (inertOwnerRegistryDB) Exec(
	context.Context,
	string,
	...any,
) (pgconn.CommandTag, error) {
	panic("owner registry construction must not execute SQL")
}

func (inertOwnerRegistryDB) Query(
	context.Context,
	string,
	...any,
) (pgx.Rows, error) {
	panic("owner registry construction must not query SQL")
}

func (inertOwnerRegistryDB) QueryRow(
	context.Context,
	string,
	...any,
) pgx.Row {
	panic("owner registry construction must not query SQL")
}

func (inertOwnerRegistryDB) BeginTx(
	context.Context,
	pgx.TxOptions,
) (pgx.Tx, error) {
	panic("owner registry construction must not begin a transaction")
}

type inertIntentAppender struct{}

func (inertIntentAppender) AppendIntentTx(
	context.Context,
	pgx.Tx,
	collaboration.EventIntent,
) error {
	panic("owner registry construction must not append intents")
}
