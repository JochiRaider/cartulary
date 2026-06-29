package timeline

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type Facade struct {
	store *Store
}

func NewFacade(pool postgres.DB) *Facade {
	return NewFacadeWithStore(NewStore(pool))
}

func NewFacadeWithStore(store *Store) *Facade {
	return &Facade{store: store}
}

func (f *Facade) RecordIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	return f.store.GetRecordIncident(ctx, recordID)
}

func (f *Facade) GetTimeConversionProfile(ctx context.Context, incidentID uuid.UUID, now time.Time) (TimeConversionProfile, error) {
	return f.store.GetTimeConversionProfile(ctx, incidentID, now)
}

func (f *Facade) PutTimeConversionProfile(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request TimeConversionProfilePutRequest, now time.Time) (TimeConversionProfile, error) {
	return f.store.PutTimeConversionProfile(ctx, actor, incidentID, request, now)
}

func (f *Facade) QueryTimelineRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error) {
	return f.store.QueryRows(ctx, incidentID, query)
}

func (f *Facade) CreateTimelineRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return f.store.CreateRow(ctx, actor, incidentID, request, requestHash, requestID, now)
}

func (f *Facade) CreateImportedTimelineRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return f.store.CreateImportedRow(ctx, actor, incidentID, request, requestHash, requestID, now)
}

func (f *Facade) PatchTimelineRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return f.store.PatchRow(ctx, actor, recordID, request, requestHash, requestID, now)
}

func (f *Facade) ResolveTimelineConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims TimelineConflictTokenClaims, request ConflictResolveRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return f.store.ResolveConflict(ctx, actor, recordID, claims, request, requestHash, requestID, now)
}

func (f *Facade) ClipboardPaste(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request ClipboardPasteRequest, requestHash []byte, requestID string, now time.Time) (ClipboardPasteResult, error) {
	return f.store.ClipboardPaste(ctx, actor, incidentID, request, requestHash, requestID, now)
}

func (f *Facade) MarkReviewed(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request ActionRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return f.store.MarkReviewed(ctx, actor, recordID, request, requestHash, requestID, now)
}

func (f *Facade) Supersede(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request SupersedeRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return f.store.Supersede(ctx, actor, recordID, request, requestHash, requestID, now)
}

func (f *Facade) SnapshotRecordSubstrate(ctx context.Context, recordID uuid.UUID) (RecordSubstrateSnapshot, error) {
	return f.store.SnapshotRecordSubstrate(ctx, recordID)
}
