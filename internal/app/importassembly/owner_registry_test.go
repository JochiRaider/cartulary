package importassembly

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

func TestOwnerCreateRegistryComposesEveryCurrentViewTarget(t *testing.T) {
	t.Parallel()

	registry, err := NewOwnerCreateRegistry(OwnerRegistryDependencies{
		Postgres:          inertOwnerRegistryDB{},
		RevisionAppender:  &revisions.Appender{},
		Intents:           inertIntentAppender{},
		Timeline:          inertTimelineFacade(),
		ProjectionCatalog: &projections.Catalog{},
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

func TestOwnerCreateRegistryRequiresCompositionDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		deps OwnerRegistryDependencies
	}{
		{
			name: "postgres",
			deps: OwnerRegistryDependencies{
				RevisionAppender:  &revisions.Appender{},
				Intents:           inertIntentAppender{},
				Timeline:          inertTimelineFacade(),
				ProjectionCatalog: &projections.Catalog{},
			},
		},
		{
			name: "revisions",
			deps: OwnerRegistryDependencies{
				Postgres:          inertOwnerRegistryDB{},
				Intents:           inertIntentAppender{},
				Timeline:          inertTimelineFacade(),
				ProjectionCatalog: &projections.Catalog{},
			},
		},
		{
			name: "intents",
			deps: OwnerRegistryDependencies{
				Postgres:          inertOwnerRegistryDB{},
				RevisionAppender:  &revisions.Appender{},
				Timeline:          inertTimelineFacade(),
				ProjectionCatalog: &projections.Catalog{},
			},
		},
		{
			name: "timeline",
			deps: OwnerRegistryDependencies{
				Postgres:          inertOwnerRegistryDB{},
				RevisionAppender:  &revisions.Appender{},
				Intents:           inertIntentAppender{},
				ProjectionCatalog: &projections.Catalog{},
			},
		},
		{
			name: "projection catalog",
			deps: OwnerRegistryDependencies{
				Postgres:         inertOwnerRegistryDB{},
				RevisionAppender: &revisions.Appender{},
				Intents:          inertIntentAppender{},
				Timeline:         inertTimelineFacade(),
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

func inertTimelineFacade() *timeline.Facade {
	return timeline.NewFacade(
		inertOwnerRegistryDB{},
		timeline.Collaborators{},
		conflicttokens.ConflictTokenCodec{},
	)
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
