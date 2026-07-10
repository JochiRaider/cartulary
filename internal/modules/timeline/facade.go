package timeline

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type Facade struct {
	store *store
}

type CreateRowCommand struct {
	Actor       authn.UserRecord
	IncidentID  uuid.UUID
	Request     CreateRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type PatchRowCommand struct {
	Actor       authn.UserRecord
	RecordID    uuid.UUID
	Request     PatchRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type ConflictResolveCommand struct {
	Actor       authn.UserRecord
	RecordID    uuid.UUID
	Claims      TimelineConflictTokenClaims
	Request     ConflictResolveRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type ClipboardPasteCommand struct {
	Actor       authn.UserRecord
	IncidentID  uuid.UUID
	Request     ClipboardPasteRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type BulkMutationCommand struct {
	Actor       authn.UserRecord
	IncidentID  uuid.UUID
	Request     BulkMutationRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type MarkReviewedCommand struct {
	Actor       authn.UserRecord
	RecordID    uuid.UUID
	Request     ActionRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

type SupersedeCommand struct {
	Actor       authn.UserRecord
	RecordID    uuid.UUID
	Request     SupersedeRequest
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

func NewFacade(pool postgres.DB) *Facade {
	return newFacadeWithStore(newStore(pool))
}

func newFacadeWithStore(store *store) *Facade {
	return &Facade{store: store}
}

func (f *Facade) SetConflictTokenCodec(codec conflicttokens.ConflictTokenCodec) {
	f.store.setConflictTokenCodec(codec)
}

func (f *Facade) ParseConflictToken(token string) (TimelineConflictTokenClaims, bool) {
	return f.store.parseConflictToken(token)
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

func (f *Facade) CreateRow(ctx context.Context, command CreateRowCommand) (MutationResult, error) {
	requestHash := command.RequestHash
	if requestHash == nil {
		requestHash = TimelineCreateRequestHash(command.Request)
	}
	return f.store.CreateRow(ctx, command.Actor, command.IncidentID, command.Request, requestHash, command.RequestID, command.Now)
}

func (f *Facade) CreateImportedRow(ctx context.Context, command CreateRowCommand) (MutationResult, error) {
	requestHash := command.RequestHash
	if requestHash == nil {
		requestHash = TimelineCreateRequestHash(command.Request)
	}
	return f.store.CreateImportedRow(ctx, command.Actor, command.IncidentID, command.Request, requestHash, command.RequestID, command.Now)
}

func (f *Facade) PatchRow(ctx context.Context, command PatchRowCommand) (MutationResult, error) {
	requestHash := command.RequestHash
	if requestHash == nil {
		requestHash = TimelinePatchRequestHash(command.Request)
	}
	return f.store.PatchRow(ctx, command.Actor, command.RecordID, command.Request, requestHash, command.RequestID, command.Now)
}

func (f *Facade) ResolveConflict(ctx context.Context, command ConflictResolveCommand) (MutationResult, error) {
	requestHash := command.RequestHash
	if requestHash == nil {
		requestHash = TimelineConflictResolveRequestHash(command.Claims, command.Request)
	}
	return f.store.ResolveConflict(ctx, command.Actor, command.RecordID, command.Claims, command.Request, requestHash, command.RequestID, command.Now)
}

func (f *Facade) ApplyClipboardPaste(ctx context.Context, command ClipboardPasteCommand) (ClipboardPasteResult, error) {
	requestHash := command.RequestHash
	if requestHash == nil {
		requestHash = TimelineClipboardPasteRequestHash(command.Request)
	}
	return f.store.ClipboardPaste(ctx, command.Actor, command.IncidentID, command.Request, requestHash, command.RequestID, command.Now)
}

func (f *Facade) ApplyBulkMutation(ctx context.Context, command BulkMutationCommand) (ClipboardPasteResult, error) {
	requestHash := command.RequestHash
	if requestHash == nil {
		requestHash = BulkMutationRequestHash(command.Request)
	}
	return f.ApplyClipboardPaste(ctx, ClipboardPasteCommand{
		Actor:       command.Actor,
		IncidentID:  command.IncidentID,
		Request:     BulkMutationClipboardRequest(command.Request),
		RequestHash: requestHash,
		RequestID:   command.RequestID,
		Now:         command.Now,
	})
}

func (f *Facade) MarkReviewedRow(ctx context.Context, command MarkReviewedCommand) (MutationResult, error) {
	requestHash := command.RequestHash
	if requestHash == nil {
		requestHash = TimelineActionRequestHash(command.Request.BaseRowVersion, command.Request.ClientTxnID, command.Request.Reason, nil)
	}
	return f.store.MarkReviewed(ctx, command.Actor, command.RecordID, command.Request, requestHash, command.RequestID, command.Now)
}

func (f *Facade) SupersedeRow(ctx context.Context, command SupersedeCommand) (MutationResult, error) {
	requestHash := command.RequestHash
	if requestHash == nil {
		requestHash = TimelineActionRequestHash(command.Request.BaseRowVersion, command.Request.ClientTxnID, &command.Request.Reason, command.Request.ReplacementRecordID)
	}
	return f.store.Supersede(ctx, command.Actor, command.RecordID, command.Request, requestHash, command.RequestID, command.Now)
}

func (f *Facade) SnapshotRecordSubstrate(ctx context.Context, recordID uuid.UUID) (RecordSubstrateSnapshot, error) {
	return f.store.SnapshotRecordSubstrate(ctx, recordID)
}
