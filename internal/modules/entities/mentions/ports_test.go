package mentions

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type mentionStorePostgresStub struct{ postgres.DB }
type mentionLinkOperationsStub struct{ LinkOperationsPort }
type mentionTimelineEffectsStub struct{ TimelineEffectsPort }
type mentionCollaborationStub struct {
	collaboration.RecordChangedAppender
}

func TestMentionStoreCompositionRequiresCompleteDependencies_Unit(t *testing.T) {
	valid := func() StoreDependencies {
		return StoreDependencies{
			Postgres:      mentionStorePostgresStub{},
			Revisions:     &revisions.Appender{},
			Links:         mentionLinkOperationsStub{},
			Projections:   noopProjectionWriter{},
			Timeline:      mentionTimelineEffectsStub{},
			Collaboration: mentionCollaborationStub{},
		}
	}
	tests := []struct {
		name       string
		dependency string
		mutate     func(*StoreDependencies)
	}{
		{name: "missing Postgres", dependency: "Postgres", mutate: func(dependencies *StoreDependencies) { dependencies.Postgres = nil }},
		{name: "typed-nil Postgres", dependency: "Postgres", mutate: func(dependencies *StoreDependencies) { dependencies.Postgres = (*mentionStorePostgresStub)(nil) }},
		{name: "missing Revisions", dependency: "Revisions", mutate: func(dependencies *StoreDependencies) { dependencies.Revisions = nil }},
		{name: "typed-nil Revisions", dependency: "Revisions", mutate: func(dependencies *StoreDependencies) { dependencies.Revisions = (*revisions.Appender)(nil) }},
		{name: "missing Links", dependency: "Links", mutate: func(dependencies *StoreDependencies) { dependencies.Links = nil }},
		{name: "typed-nil Links", dependency: "Links", mutate: func(dependencies *StoreDependencies) { dependencies.Links = (*mentionLinkOperationsStub)(nil) }},
		{name: "missing Projections", dependency: "Projections", mutate: func(dependencies *StoreDependencies) { dependencies.Projections = nil }},
		{name: "typed-nil Projections", dependency: "Projections", mutate: func(dependencies *StoreDependencies) { dependencies.Projections = (*noopProjectionWriter)(nil) }},
		{name: "missing Timeline", dependency: "Timeline", mutate: func(dependencies *StoreDependencies) { dependencies.Timeline = nil }},
		{name: "typed-nil Timeline", dependency: "Timeline", mutate: func(dependencies *StoreDependencies) { dependencies.Timeline = (*mentionTimelineEffectsStub)(nil) }},
		{name: "missing Collaboration", dependency: "Collaboration", mutate: func(dependencies *StoreDependencies) { dependencies.Collaboration = nil }},
		{name: "typed-nil Collaboration", dependency: "Collaboration", mutate: func(dependencies *StoreDependencies) { dependencies.Collaboration = (*mentionCollaborationStub)(nil) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dependencies := valid()
			test.mutate(&dependencies)
			store, err := NewStore(dependencies)
			if store != nil {
				t.Fatalf("NewStore result = %#v, want nil", store)
			}
			want := "compose Mention store: " + test.dependency + " is required"
			if err == nil || err.Error() != want {
				t.Fatalf("NewStore error = %v, want %q", err, want)
			}
		})
	}

	store, err := NewStore(valid())
	if err != nil || store == nil {
		t.Fatalf("valid NewStore result = %#v, %v", store, err)
	}
	store, err = NewStore(StoreDependencies{})
	if store != nil || err == nil || err.Error() != "compose Mention store: Postgres is required" {
		t.Fatalf("empty NewStore result = %#v, %v", store, err)
	}
}

type noopProjectionWriter struct{}

func (noopProjectionWriter) RefreshHostTx(context.Context, pgx.Tx, uuid.UUID) error     { return nil }
func (noopProjectionWriter) RefreshIdentityTx(context.Context, pgx.Tx, uuid.UUID) error { return nil }
func (noopProjectionWriter) DeleteHostTx(context.Context, pgx.Tx, uuid.UUID) error      { return nil }
func (noopProjectionWriter) DeleteIdentityTx(context.Context, pgx.Tx, uuid.UUID) error  { return nil }
func (noopProjectionWriter) RebuildHostsTx(context.Context, pgx.Tx, uuid.UUID) error    { return nil }
func (noopProjectionWriter) RebuildIdentitiesTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}
