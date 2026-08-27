package startup

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestWorkbookStartupCompareAndClearRestartsAfterConcurrentReplacement_Integration(t *testing.T) {
	t.Run("constructor rejects missing and typed-nil dependencies", func(t *testing.T) {
		validPreferences := noopPreferenceStore{}
		validUnitOfWork := replacementUnitOfWork{}
		validWorkspaces := NewWorkspaceRegistryFromPublication(nil)
		var typedNilPreferences *noopPreferenceStore
		var typedNilUnitOfWork *replacementUnitOfWork
		var typedNilWorkspaces *WorkspaceRegistry
		for _, test := range []struct {
			name         string
			dependencies Dependencies
		}{
			{name: "nil preferences", dependencies: Dependencies{UnitOfWork: validUnitOfWork, Workspaces: validWorkspaces}},
			{name: "typed-nil preferences", dependencies: Dependencies{Preferences: typedNilPreferences, UnitOfWork: validUnitOfWork, Workspaces: validWorkspaces}},
			{name: "nil unit of work", dependencies: Dependencies{Preferences: validPreferences, Workspaces: validWorkspaces}},
			{name: "typed-nil unit of work", dependencies: Dependencies{Preferences: validPreferences, UnitOfWork: typedNilUnitOfWork, Workspaces: validWorkspaces}},
			{name: "nil workspaces", dependencies: Dependencies{Preferences: validPreferences, UnitOfWork: validUnitOfWork}},
			{name: "typed-nil workspaces", dependencies: Dependencies{Preferences: validPreferences, UnitOfWork: validUnitOfWork, Workspaces: typedNilWorkspaces}},
		} {
			t.Run(test.name, func(t *testing.T) {
				store, err := NewStore(test.dependencies)
				if err == nil || store != nil {
					t.Fatalf("invalid dependencies published a store: store=%#v err=%v", store, err)
				}
			})
		}
	})

	incidentID := uuid.New()
	userID := uuid.New()
	invalid := []byte(`{"kind":"view_schema","id":"cartulary.view.removed.v1"}`)
	replacement := []byte(`{"kind":"view_schema","id":"cartulary.view.timeline.v2"}`)
	session := &replacementSession{
		initial:     invalid,
		replacement: replacement,
	}
	store, err := NewStore(Dependencies{
		Preferences: noopPreferenceStore{},
		UnitOfWork:  replacementUnitOfWork{session: session},
		Workspaces:  NewWorkspaceRegistryFromPublication(nil),
	})
	if err != nil {
		t.Fatalf("construct startup store: %v", err)
	}

	result, err := store.Resolve(
		context.Background(),
		incidentID,
		userID,
		"viewer",
		nil,
		time.Date(2026, 7, 26, 23, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("resolve after conditional-clear comparison miss: %v", err)
	}
	if session.userReads != 2 {
		t.Fatalf("startup selection must restart after comparison miss, reads=%d", session.userReads)
	}
	if session.clearAttempts != 1 {
		t.Fatalf("startup must attempt exactly one clear of the stale pointer, attempts=%d", session.clearAttempts)
	}
	if result.Source != SourceHome || !bytes.Equal(result.SelectedSheetRef, replacement) {
		t.Fatalf("startup must preserve and select the concurrent replacement: %#v", result)
	}
	if !bytes.Equal(result.HomeSheetRef, replacement) || len(result.ClearedPointers) != 0 {
		t.Fatalf("comparison miss must not report or expose a clear: %#v", result)
	}
}

type noopPreferenceStore struct{}

func (noopPreferenceStore) GetDefaultPreferences(context.Context, uuid.UUID) (DefaultPreferencesRecord, error) {
	return DefaultPreferencesRecord{}, nil
}

func (noopPreferenceStore) PutDefaultPreferences(context.Context, uuid.UUID, uuid.UUID, []byte, time.Time) (DefaultPreferencesRecord, error) {
	return DefaultPreferencesRecord{}, nil
}

func (noopPreferenceStore) GetUserPreferences(context.Context, uuid.UUID, uuid.UUID) (UserPreferencesRecord, error) {
	return UserPreferencesRecord{}, nil
}

func (noopPreferenceStore) PutUserPreferences(context.Context, uuid.UUID, uuid.UUID, []byte, time.Time) (UserPreferencesRecord, error) {
	return UserPreferencesRecord{}, nil
}

type replacementUnitOfWork struct {
	session Session
}

func (u replacementUnitOfWork) Run(
	ctx context.Context,
	operation func(Session) (Record, error),
) (Record, error) {
	return operation(u.session)
}

type replacementSession struct {
	Session
	initial       []byte
	replacement   []byte
	userReads     int
	clearAttempts int
}

func (r *replacementSession) UserPreferenceRef(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) ([]byte, error) {
	r.userReads++
	if r.userReads == 1 {
		return cloneBytes(r.initial), nil
	}
	return cloneBytes(r.replacement), nil
}

func (r *replacementSession) DefaultPreferenceRef(
	context.Context,
	uuid.UUID,
) ([]byte, error) {
	return nil, nil
}

func (r *replacementSession) ClearUserPreferenceIfCurrent(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	expected []byte,
	_ time.Time,
) (bool, error) {
	r.clearAttempts++
	if !bytes.Equal(expected, r.initial) {
		return false, nil
	}
	return false, nil
}
