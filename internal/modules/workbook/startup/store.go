package startup

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var (
	ErrPreferencesNotFound = errors.New("workbook startup: preferences not found")
	ErrStoreUnavailable    = errors.New("workbook startup: persistence is unavailable")
)

// PreferenceStore is the non-transactional preference surface Workbook needs
// for its public preference routes. The PostgreSQL implementation is owned by
// the startup/postgres leaf package.
type PreferenceStore interface {
	GetDefaultPreferences(ctx context.Context, incidentID uuid.UUID) (DefaultPreferencesRecord, error)
	PutDefaultPreferences(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, defaultSheetRef []byte, now time.Time) (DefaultPreferencesRecord, error)
	GetUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (UserPreferencesRecord, error)
	PutUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, homeSheetRef []byte, now time.Time) (UserPreferencesRecord, error)
}

// Session is the deliberately small transaction-scoped capability required by
// startup selection. It exposes neither pgx nor a general database handle.
type Session interface {
	UserPreferenceRef(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) ([]byte, error)
	DefaultPreferenceRef(ctx context.Context, incidentID uuid.UUID) ([]byte, error)
	ClearUserPreferenceIfCurrent(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, expected []byte, now time.Time) (bool, error)
	ClearDefaultPreferenceIfCurrent(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, expected []byte, now time.Time) (bool, error)
	ResolveSavedView(ctx context.Context, incidentID uuid.UUID, savedViewID uuid.UUID, userID uuid.UUID) (SavedViewRecord, string, error)
}

// UnitOfWork joins Workbook preferences and Saved Views resolution over one
// transaction without leaking transaction mechanics into the Workbook core.
type UnitOfWork interface {
	Run(ctx context.Context, operation func(Session) (Record, error)) (Record, error)
}

type Store struct {
	preferences           PreferenceStore
	unitOfWork            UnitOfWork
	workspaceAvailability WorkspaceResolver
}

func NewStore(preferences PreferenceStore, unitOfWork UnitOfWork, workspaceResolvers ...WorkspaceResolver) *Store {
	resolver := WorkspaceResolver(NewWorkspaceRegistryFromPublication(nil))
	if len(workspaceResolvers) > 0 && workspaceResolvers[0] != nil {
		resolver = workspaceResolvers[0]
	}
	return &Store{
		preferences:           preferences,
		unitOfWork:            unitOfWork,
		workspaceAvailability: resolver,
	}
}

func (s *Store) ValidatePreferenceSheetRef(raw []byte, role string, field string) *httpapi.APIError {
	if len(raw) == 0 {
		return nil
	}
	ref, reasonCode := decodeSheetRef(raw)
	if reasonCode != "" {
		return invalidMutationPayload(field, reasonCode)
	}
	if ref.Kind != "extension_workspace" {
		return nil
	}
	if reasonCode = s.workspaceAvailability.ResolveWorkspace(ref, role); reasonCode == "" {
		return nil
	}
	return invalidMutationPayload(extensionReasonField(field, reasonCode), reasonCode)
}

func (s *Store) ValidateExplicitSheetRef(raw []byte, role string) *httpapi.APIError {
	if len(raw) == 0 {
		return nil
	}
	ref, reasonCode := decodeSheetRef(raw)
	if reasonCode != "" {
		return invalidStartupRequest("sheet_ref", reasonCode)
	}
	if ref.Kind != "extension_workspace" {
		return nil
	}
	if reasonCode = s.workspaceAvailability.ResolveWorkspace(ref, role); reasonCode == "" {
		return nil
	}
	return invalidStartupRequest(extensionReasonField("sheet_ref", reasonCode), reasonCode)
}

func extensionReasonField(prefix string, reasonCode string) string {
	switch reasonCode {
	case "invalid_extension_profile_id", "extension_profile_not_claimed":
		if prefix == "sheet_ref" {
			return "extension_profile_id"
		}
		return prefix + ".extension_profile_id"
	case "invalid_extension_workspace_key", "extension_workspace_unavailable", "extension_workspace_not_visible":
		if prefix == "sheet_ref" {
			return "sheet_ref_id"
		}
		return prefix + ".workspace_key"
	default:
		return prefix
	}
}

func (s *Store) GetDefaultPreferences(ctx context.Context, incidentID uuid.UUID) (DefaultPreferencesRecord, error) {
	if s.preferences == nil {
		return DefaultPreferencesRecord{}, ErrStoreUnavailable
	}
	return s.preferences.GetDefaultPreferences(ctx, incidentID)
}

func (s *Store) PutDefaultPreferences(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, defaultSheetRef []byte, now time.Time) (DefaultPreferencesRecord, error) {
	if s.preferences == nil {
		return DefaultPreferencesRecord{}, ErrStoreUnavailable
	}
	return s.preferences.PutDefaultPreferences(ctx, incidentID, actorUserID, defaultSheetRef, now)
}

func (s *Store) GetUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (UserPreferencesRecord, error) {
	if s.preferences == nil {
		return UserPreferencesRecord{}, ErrStoreUnavailable
	}
	return s.preferences.GetUserPreferences(ctx, incidentID, userID)
}

func (s *Store) PutUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, homeSheetRef []byte, now time.Time) (UserPreferencesRecord, error) {
	if s.preferences == nil {
		return UserPreferencesRecord{}, ErrStoreUnavailable
	}
	return s.preferences.PutUserPreferences(ctx, incidentID, userID, homeSheetRef, now)
}

func (s *Store) Resolve(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, role string, explicitSheetRef []byte, now time.Time) (Record, error) {
	if s.unitOfWork == nil {
		return Record{}, ErrStoreUnavailable
	}
	availability := ExtensionWorkspaceAvailability{
		SchemaID:   ExtensionWorkspaceAvailabilitySchemaID,
		IncidentID: incidentID.String(),
		Workspaces: s.workspaceAvailability.AvailableWorkspaces(role),
	}
	return s.unitOfWork.Run(ctx, func(session Session) (Record, error) {
		return s.resolveInSession(ctx, session, incidentID, userID, role, explicitSheetRef, now, availability)
	})
}

func (s *Store) resolveInSession(
	ctx context.Context,
	session Session,
	incidentID uuid.UUID,
	userID uuid.UUID,
	role string,
	explicitSheetRef []byte,
	now time.Time,
	availability ExtensionWorkspaceAvailability,
) (Record, error) {
	for {
		homeRef, err := session.UserPreferenceRef(ctx, incidentID, userID)
		if err != nil {
			return Record{}, err
		}
		defaultRef, err := session.DefaultPreferenceRef(ctx, incidentID)
		if err != nil {
			return Record{}, err
		}

		record := Record{
			IncidentID:                     incidentID,
			HomeSheetRef:                   cloneBytes(homeRef),
			DefaultSheetRef:                cloneBytes(defaultRef),
			ExtensionWorkspaceAvailability: availability,
		}
		restartSelection := false
		for _, candidate := range []struct {
			source string
			ref    []byte
		}{
			{source: SourceExplicit, ref: explicitSheetRef},
			{source: SourceHome, ref: homeRef},
			{source: SourceDefault, ref: defaultRef},
		} {
			if len(candidate.ref) == 0 {
				continue
			}
			resolved, invalidReason, err := resolveCandidate(
				ctx,
				session,
				incidentID,
				userID,
				role,
				candidate.ref,
				s.workspaceAvailability,
			)
			if err != nil {
				return Record{}, err
			}
			if invalidReason == "" {
				record.SelectedSheetRef = mustSheetRefJSON(resolved.SheetRef)
				if resolved.ViewSchemaID != "" {
					viewSchemaID := resolved.ViewSchemaID
					record.SelectedViewSchemaID = &viewSchemaID
				}
				record.SelectedSavedView = resolved.SavedView
				record.Source = candidate.source
				return record, nil
			}
			if candidate.source == SourceExplicit {
				continue
			}
			cleared := false
			if candidate.source == SourceHome {
				cleared, err = session.ClearUserPreferenceIfCurrent(ctx, incidentID, userID, candidate.ref, now)
			}
			if candidate.source == SourceDefault {
				cleared, err = session.ClearDefaultPreferenceIfCurrent(ctx, incidentID, userID, candidate.ref, now)
			}
			if err != nil {
				return Record{}, err
			}
			if !cleared {
				restartSelection = true
				break
			}
			record.ClearedPointers = append(record.ClearedPointers, ClearedPointer{
				Source:     candidate.source,
				SheetRef:   cloneBytes(candidate.ref),
				ReasonCode: invalidReason,
			})
			if candidate.source == SourceHome {
				record.HomeSheetRef = nil
			}
			if candidate.source == SourceDefault {
				record.DefaultSheetRef = nil
			}
		}

		if restartSelection {
			continue
		}
		record.SelectedSheetRef = mustSheetRefJSON(SheetRef{Kind: "view_schema", ID: TimelineViewSchemaID})
		timelineViewSchemaID := TimelineViewSchemaID
		record.SelectedViewSchemaID = &timelineViewSchemaID
		record.Source = SourceTimeline
		return record, nil
	}
}

type resolvedCandidate struct {
	ViewSchemaID string
	SavedView    *SavedViewRecord
	SheetRef     SheetRef
}

func resolveCandidate(
	ctx context.Context,
	session Session,
	incidentID uuid.UUID,
	userID uuid.UUID,
	role string,
	raw []byte,
	workspaceAvailability WorkspaceResolver,
) (resolvedCandidate, string, error) {
	ref, reasonCode := decodeSheetRef(raw)
	if reasonCode != "" {
		return resolvedCandidate{}, reasonCode, nil
	}
	switch ref.Kind {
	case "view_schema":
		if _, ok := viewschema.Lookup(ref.ID); !ok {
			return resolvedCandidate{}, "unknown_view_schema", nil
		}
		return resolvedCandidate{ViewSchemaID: ref.ID, SheetRef: ref}, "", nil
	case "saved_view":
		savedViewID, err := uuid.Parse(ref.ID)
		if err != nil {
			return resolvedCandidate{}, "invalid_saved_view_id", nil
		}
		savedView, reasonCode, err := session.ResolveSavedView(ctx, incidentID, savedViewID, userID)
		if err != nil {
			return resolvedCandidate{}, "", err
		}
		if reasonCode != "" {
			return resolvedCandidate{}, reasonCode, nil
		}
		if _, ok := viewschema.Lookup(savedView.ViewSchemaID); !ok {
			return resolvedCandidate{}, "unknown_view_schema", nil
		}
		return resolvedCandidate{ViewSchemaID: savedView.ViewSchemaID, SavedView: &savedView, SheetRef: ref}, "", nil
	case "extension_workspace":
		if reasonCode := workspaceAvailability.ResolveWorkspace(ref, role); reasonCode != "" {
			return resolvedCandidate{}, reasonCode, nil
		}
		return resolvedCandidate{SheetRef: ref}, "", nil
	default:
		return resolvedCandidate{}, "unsupported_sheet_ref_kind", nil
	}
}

func cloneBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	return append([]byte(nil), value...)
}
