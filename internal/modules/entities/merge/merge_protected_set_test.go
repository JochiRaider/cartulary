package merge

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type mergeProjectionWriterStub struct {
	workbookprojection.Writer
}

type mergeStorePostgresStub struct{ postgres.DB }
type mergeTimelineEffectsStub struct{ TimelineEffectsPort }
type mergeCollaborationStub struct {
	collaboration.RecordChangedAppender
}
type mergeAssessmentEffectsStub struct{ AssessmentEffectsPort }

func TestMergeStoreCompositionRequiresCompleteDependencies_Unit(t *testing.T) {
	mentionStore := &noopMentionStore{}
	valid := func() StoreDependencies {
		return StoreDependencies{
			Postgres:      mergeStorePostgresStub{},
			Revisions:     &revisions.Appender{},
			HostIdentity:  hostidentity.NewMergeCapability(),
			Assessments:   mergeAssessmentEffectsStub{},
			Mentions:      mentionStore,
			Links:         noopLinkEffects{},
			Timeline:      mergeTimelineEffectsStub{},
			Projections:   mergeProjectionWriterStub{},
			Collaboration: mergeCollaborationStub{},
		}
	}
	tests := []struct {
		name       string
		dependency string
		mutate     func(*StoreDependencies)
	}{
		{name: "missing Postgres", dependency: "Postgres", mutate: func(dependencies *StoreDependencies) { dependencies.Postgres = nil }},
		{name: "typed-nil Postgres", dependency: "Postgres", mutate: func(dependencies *StoreDependencies) { dependencies.Postgres = (*mergeStorePostgresStub)(nil) }},
		{name: "missing Revisions", dependency: "Revisions", mutate: func(dependencies *StoreDependencies) { dependencies.Revisions = nil }},
		{name: "typed-nil Revisions", dependency: "Revisions", mutate: func(dependencies *StoreDependencies) { dependencies.Revisions = (*revisions.Appender)(nil) }},
		{name: "missing HostIdentity", dependency: "HostIdentity", mutate: func(dependencies *StoreDependencies) { dependencies.HostIdentity = nil }},
		{name: "typed-nil HostIdentity", dependency: "HostIdentity", mutate: func(dependencies *StoreDependencies) {
			dependencies.HostIdentity = (*hostidentity.MergeCapability)(nil)
		}},
		{name: "missing Assessments", dependency: "Assessments", mutate: func(dependencies *StoreDependencies) { dependencies.Assessments = nil }},
		{name: "typed-nil Assessments", dependency: "Assessments", mutate: func(dependencies *StoreDependencies) { dependencies.Assessments = (*mergeAssessmentEffectsStub)(nil) }},
		{name: "missing Mentions", dependency: "Mentions", mutate: func(dependencies *StoreDependencies) { dependencies.Mentions = nil }},
		{name: "typed-nil Mentions", dependency: "Mentions", mutate: func(dependencies *StoreDependencies) { dependencies.Mentions = (*noopMentionStore)(nil) }},
		{name: "missing Links", dependency: "Links", mutate: func(dependencies *StoreDependencies) { dependencies.Links = nil }},
		{name: "typed-nil Links", dependency: "Links", mutate: func(dependencies *StoreDependencies) { dependencies.Links = (*noopLinkEffects)(nil) }},
		{name: "missing Timeline", dependency: "Timeline", mutate: func(dependencies *StoreDependencies) { dependencies.Timeline = nil }},
		{name: "typed-nil Timeline", dependency: "Timeline", mutate: func(dependencies *StoreDependencies) { dependencies.Timeline = (*mergeTimelineEffectsStub)(nil) }},
		{name: "missing Projections", dependency: "Projections", mutate: func(dependencies *StoreDependencies) { dependencies.Projections = nil }},
		{name: "typed-nil Projections", dependency: "Projections", mutate: func(dependencies *StoreDependencies) { dependencies.Projections = (*mergeProjectionWriterStub)(nil) }},
		{name: "missing Collaboration", dependency: "Collaboration", mutate: func(dependencies *StoreDependencies) { dependencies.Collaboration = nil }},
		{name: "typed-nil Collaboration", dependency: "Collaboration", mutate: func(dependencies *StoreDependencies) { dependencies.Collaboration = (*mergeCollaborationStub)(nil) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dependencies := valid()
			test.mutate(&dependencies)
			store, err := NewStore(dependencies)
			if store != nil {
				t.Fatalf("NewStore result = %#v, want nil", store)
			}
			want := "compose Merge store: " + test.dependency + " is required"
			if err == nil || err.Error() != want {
				t.Fatalf("NewStore error = %v, want %q", err, want)
			}
		})
	}

	store, err := NewStore(valid())
	if err != nil || store == nil {
		t.Fatalf("valid NewStore result = %#v, %v", store, err)
	}
	adapter, ok := store.ports.mentions.(entityMentionAdapter)
	if !ok || adapter.store != mentionStore {
		t.Fatalf("mention port = %#v, want adapter retaining %p", store.ports.mentions, mentionStore)
	}
	store, err = NewStore(StoreDependencies{})
	if store != nil || err == nil || err.Error() != "compose Merge store: Postgres is required" {
		t.Fatalf("empty NewStore result = %#v, %v", store, err)
	}
}

func TestMergeAssessmentRepointRejectsUnprotectedAssessment(t *testing.T) {
	assessmentID := uuid.New()
	err := classifyAssessmentRepointError(&AssessmentProtectedSetChangedError{RecordID: assessmentID})
	var precondition *MergePreconditionError
	if !errors.As(err, &precondition) || precondition.ReasonCode != "protected_set_changed" || precondition.Details["record_id"] != assessmentID.String() {
		t.Fatalf("expected protected_set_changed precondition, got %T %[1]v", err)
	}
}

type noopLinkEffects struct{}

type noopMentionStore struct{}

func (noopMentionStore) RepointMergedMentionsTx(context.Context, pgx.Tx, mentions.RepointMergedMentionsCommand) (mentions.RepointMergedMentionsResult, error) {
	return mentions.RepointMergedMentionsResult{
		Mutations:             []mentions.MergeMutation{},
		TimelineInvalidations: map[uuid.UUID][]string{},
	}, nil
}

func (noopLinkEffects) RepointLinksTx(context.Context, pgx.Tx, RepointLinksCommand) (RepointLinksResult, error) {
	return RepointLinksResult{
		Mutations:                 []LinkEffectMutation{},
		LinkTypesBySourceRecordID: map[uuid.UUID][]string{},
	}, nil
}

func (noopLinkEffects) RepointTagsTx(context.Context, pgx.Tx, RepointTagsCommand) (RepointTagsResult, error) {
	return RepointTagsResult{Mutations: []LinkEffectMutation{}}, nil
}
