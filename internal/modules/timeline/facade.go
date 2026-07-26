package timeline

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Facade struct {
	store *store
}

func NewFacade(pool postgres.DB, dependencies Dependencies) *Facade {
	return newFacadeWithStore(newStore(pool, dependencies))
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
	rows, err := buildClipboardOwnerRows(command.Plan)
	if err != nil {
		return ClipboardPasteResult{}, err
	}
	return f.store.applyOwnerBatchV1(ctx, command.Actor, command.IncidentID, ownerBatchApplyV1{
		ClientTxnID: command.ClientTxnID,
		Operation:   OwnerBatchOperationClipboardPasteV1,
		Targets:     command.Targets,
		Rows:        rows,
		RequestHash: command.RequestHash,
		RequestID:   command.RequestID,
		Now:         command.Now,
	})
}

func (f *Facade) ApplyFillDown(ctx context.Context, command FillDownCommand) (ClipboardPasteResult, error) {
	rows, err := buildFillDownOwnerRows(command.FieldKey, command.Value, len(command.Targets))
	if err != nil {
		return ClipboardPasteResult{}, err
	}
	return f.store.applyOwnerBatchV1(ctx, command.Actor, command.IncidentID, ownerBatchApplyV1{
		ClientTxnID: command.ClientTxnID,
		Operation:   OwnerBatchOperationFillDownV1,
		Targets:     command.Targets,
		Rows:        rows,
		RequestHash: command.RequestHash,
		RequestID:   command.RequestID,
		Now:         command.Now,
	})
}

func (f *Facade) ApplyMultiRowTagAssignment(ctx context.Context, command MultiRowTagAssignmentCommand) (ClipboardPasteResult, error) {
	rows, err := buildTagAssignmentOwnerRows(command.TagName, command.NormalizedTag, len(command.Targets))
	if err != nil {
		return ClipboardPasteResult{}, err
	}
	return f.store.applyOwnerBatchV1(ctx, command.Actor, command.IncidentID, ownerBatchApplyV1{
		ClientTxnID: command.ClientTxnID,
		Operation:   OwnerBatchOperationMultiRowTagAssignmentV1,
		Targets:     command.Targets,
		Rows:        rows,
		RequestHash: command.RequestHash,
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
