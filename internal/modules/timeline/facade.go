package timeline

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Facade struct {
	store *store
}

var errRequestFingerprintRequired = errors.New("timeline mutation request fingerprint is required")

func NewFacade(pool postgres.DB, collaborators Collaborators, conflictTokens conflicttokens.ConflictTokenCodec) *Facade {
	return newFacadeWithStore(newStore(pool, collaborators, conflictTokens))
}

func newFacadeWithStore(store *store) *Facade {
	return &Facade{store: store}
}

func (f *Facade) ParseConflictToken(token string) (conflicttokens.ConflictTokenClaims, bool) {
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
	if err := requireRequestFingerprint(command.RequestHash); err != nil {
		return MutationResult{}, err
	}
	return f.store.CreateRow(ctx, command.Actor, command.IncidentID, command.Request, command.RequestHash, command.RequestID, command.Now)
}

func (f *Facade) CreateImportedRow(ctx context.Context, command CreateRowCommand) (MutationResult, error) {
	if err := requireRequestFingerprint(command.RequestHash); err != nil {
		return MutationResult{}, err
	}
	return f.store.CreateImportedRow(ctx, command.Actor, command.IncidentID, command.Request, command.RequestHash, command.RequestID, command.Now)
}

func (f *Facade) PatchRow(ctx context.Context, command PatchRowCommand) (MutationResult, error) {
	if err := requireRequestFingerprint(command.RequestHash); err != nil {
		return MutationResult{}, err
	}
	return f.store.PatchRow(ctx, command.Actor, command.RecordID, command.Request, command.RequestHash, command.RequestID, command.Now)
}

func (f *Facade) ResolveConflict(ctx context.Context, command ConflictResolveCommand) (MutationResult, error) {
	if err := requireRequestFingerprint(command.RequestHash); err != nil {
		return MutationResult{}, err
	}
	return f.store.ResolveConflict(ctx, command.Actor, command.RecordID, command.Claims, command.Request, command.RequestHash, command.RequestID, command.Now)
}

func (f *Facade) ApplyClipboardPaste(ctx context.Context, command ClipboardPasteCommand) (ClipboardPasteResult, error) {
	if err := requireRequestFingerprint(command.RequestHash); err != nil {
		return ClipboardPasteResult{}, err
	}
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
	if err := requireRequestFingerprint(command.RequestHash); err != nil {
		return ClipboardPasteResult{}, err
	}
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
	if err := requireRequestFingerprint(command.RequestHash); err != nil {
		return ClipboardPasteResult{}, err
	}
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
	if err := requireRequestFingerprint(command.RequestHash); err != nil {
		return MutationResult{}, err
	}
	return f.store.MarkReviewed(ctx, command.Actor, command.RecordID, command.Request, command.RequestHash, command.RequestID, command.Now)
}

func (f *Facade) SupersedeRow(ctx context.Context, command SupersedeCommand) (MutationResult, error) {
	if err := requireRequestFingerprint(command.RequestHash); err != nil {
		return MutationResult{}, err
	}
	return f.store.Supersede(ctx, command.Actor, command.RecordID, command.Request, command.RequestHash, command.RequestID, command.Now)
}

func requireRequestFingerprint(fingerprint []byte) error {
	if len(fingerprint) == 0 {
		return errRequestFingerprintRequired
	}
	return nil
}
