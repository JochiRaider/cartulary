package savedviews

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSavedViewApplicationOrderingAndNoOp_Unit(t *testing.T) {
	t.Parallel()

	actorUserID := uuid.MustParse("00000000-0000-4000-8000-000000111101")
	otherUserID := uuid.MustParse("00000000-0000-4000-8000-000000111102")
	incidentID := uuid.MustParse("00000000-0000-4000-8000-000000111103")
	savedViewID := uuid.MustParse("00000000-0000-4000-8000-000000111104")
	now := time.Date(2026, time.July, 29, 8, 30, 0, 0, time.UTC)
	base := savedViewRecord{
		SavedViewID:      savedViewID,
		IncidentID:       incidentID,
		ViewSchemaID:     "cartulary.view.timeline.v1",
		Scope:            scopePrivate,
		DisplayName:      "Current",
		QueryJSON:        []byte(`{"filters":[],"sort":[]}`),
		LayoutJSON:       []byte(`{"column_order":[]}`),
		OwnerUserID:      &actorUserID,
		CreatedAt:        now.Add(-time.Hour),
		UpdatedAt:        now.Add(-time.Minute),
		SavedViewVersion: 4,
	}

	t.Run("mutation denial precedes version conflict", func(t *testing.T) {
		repository := &fakeSavedViewRepository{current: base}
		application := newSavedViewApplication(repository)
		_, err := application.patch(
			context.Background(),
			incidentID,
			savedViewID,
			otherUserID,
			"member",
			patchRequest{BaseSavedViewVersion: 3},
			now,
		)
		if !errors.Is(err, errSavedViewMutationDenied) {
			t.Fatalf("patch error = %v, want mutation denied", err)
		}
		assertSavedViewRepositoryEvents(t, repository.events, "lock")
	})

	t.Run("version conflict precedes patch application", func(t *testing.T) {
		repository := &fakeSavedViewRepository{current: base}
		application := newSavedViewApplication(repository)
		_, err := application.patch(
			context.Background(),
			incidentID,
			savedViewID,
			actorUserID,
			"member",
			patchRequest{BaseSavedViewVersion: 3},
			now,
		)
		var conflict *savedViewVersionConflictError
		if !errors.As(err, &conflict) {
			t.Fatalf("patch error = %v, want version conflict", err)
		}
		assertSavedViewRepositoryEvents(t, repository.events, "lock")
	})

	t.Run("structural no-op performs no update", func(t *testing.T) {
		repository := &fakeSavedViewRepository{current: base}
		application := newSavedViewApplication(repository)
		got, err := application.patch(
			context.Background(),
			incidentID,
			savedViewID,
			actorUserID,
			"member",
			patchRequest{
				BaseSavedViewVersion: 4,
				QueryJSON: optionalJSON{
					Present: true,
					Value:   []byte("{ \"sort\": [], \"filters\": [] }"),
				},
			},
			now,
		)
		if err != nil {
			t.Fatalf("patch no-op: %v", err)
		}
		if got.SavedViewVersion != base.SavedViewVersion || !got.UpdatedAt.Equal(base.UpdatedAt) {
			t.Fatalf("no-op changed record: %#v", got)
		}
		assertSavedViewRepositoryEvents(t, repository.events, "lock")
	})

	t.Run("changed patch updates exactly once", func(t *testing.T) {
		repository := &fakeSavedViewRepository{current: base}
		application := newSavedViewApplication(repository)
		got, err := application.patch(
			context.Background(),
			incidentID,
			savedViewID,
			actorUserID,
			"member",
			patchRequest{
				BaseSavedViewVersion: 4,
				DisplayName:          optionalString{Present: true, Value: "Changed"},
			},
			now,
		)
		if err != nil {
			t.Fatalf("patch change: %v", err)
		}
		if got.DisplayName != "Changed" || got.SavedViewVersion != 5 || !got.UpdatedAt.Equal(now) {
			t.Fatalf("changed record = %#v", got)
		}
		assertSavedViewRepositoryEvents(t, repository.events, "lock", "update")
	})

	t.Run("delete denial performs no delete", func(t *testing.T) {
		repository := &fakeSavedViewRepository{current: base}
		application := newSavedViewApplication(repository)
		err := application.delete(context.Background(), incidentID, savedViewID, otherUserID, "member")
		if !errors.Is(err, errSavedViewMutationDenied) {
			t.Fatalf("delete error = %v, want mutation denied", err)
		}
		assertSavedViewRepositoryEvents(t, repository.events, "lock")
	})
}

type fakeSavedViewRepository struct {
	current savedViewRecord
	events  []string
}

func (f *fakeSavedViewRepository) create(context.Context, uuid.UUID, uuid.UUID, createRequest, time.Time) (savedViewRecord, error) {
	panic("unexpected create")
}

func (f *fakeSavedViewRepository) createSystemFixture(context.Context, uuid.UUID, createRequest, time.Time) (savedViewRecord, error) {
	panic("unexpected fixture create")
}

func (f *fakeSavedViewRepository) listVisible(context.Context, uuid.UUID, uuid.UUID, listPageRequest) ([]savedViewRecord, error) {
	panic("unexpected list")
}

func (f *fakeSavedViewRepository) getVisibleForUpdate(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (savedViewRecord, error) {
	panic("unexpected standalone lookup")
}

func (f *fakeSavedViewRepository) patchVisible(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
	mutate func(savedViewRecord) (savedViewRecord, bool, error),
) (savedViewRecord, error) {
	f.events = append(f.events, "lock")
	next, changed, err := mutate(f.current)
	if err != nil {
		return savedViewRecord{}, err
	}
	if !changed {
		return f.current, nil
	}
	f.events = append(f.events, "update")
	f.current = next
	return next, nil
}

func (f *fakeSavedViewRepository) deleteVisible(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
	authorize func(savedViewRecord) error,
) error {
	f.events = append(f.events, "lock")
	if err := authorize(f.current); err != nil {
		return err
	}
	f.events = append(f.events, "delete")
	return nil
}

func assertSavedViewRepositoryEvents(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("repository events = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("repository events = %v, want %v", got, want)
		}
	}
}
