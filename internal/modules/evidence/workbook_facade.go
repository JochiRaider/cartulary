package evidence

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type WorkbookFacade struct {
	pool              postgres.DB
	authStore         *authn.Store
	incidentAccess    incidents.Access
	recordStore       *records.Store
	projectionRows    evidenceprojection.Rows
	supportEffects    evidenceprojection.SupportProjectionEffectsTx
	revisionHistory   conflicttokens.RevisionWindowReader
	revisionAppender  *revisions.Appender
	sourceMutations   *SourceMutationService
	blobs             blobRepository
	blobLifecycle     blobLifecycleRepository
	conflictTokens    conflicttokens.ConflictTokenCodec
	conflictFields    conflicttokens.FieldResolver
	conflictSnapshots conflicttokens.RevisionSnapshotProjector
	keepSaved         conflicttokens.IdempotencyPort
	collaboration     collaboration.IntentAppender
	mutations         evidenceSourceMutationKernel
	objects           objectstore.Store
}

type WorkbookCreateRequest struct {
	ViewSchemaID        string
	ClientTxnID         string
	Values              map[string]WorkbookFieldValue
	InitialObjectBlobID *uuid.UUID
}

type WorkbookPatchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []WorkbookPatchChange
}

type WorkbookPatchChange struct {
	FieldKey       string
	Value          *WorkbookFieldValue
	CanonicalValue any
}

type WorkbookCreateCommand struct {
	Actor       authn.UserRecord
	IncidentID  uuid.UUID
	Request     WorkbookCreateRequest
	RequestHash []byte
	RequestID   string
	RouteKey    string
	Now         time.Time
}

type WorkbookPatchCommand struct {
	Actor            authn.UserRecord
	RecordID         uuid.UUID
	Request          WorkbookPatchRequest
	RequestHash      []byte
	RequestID        string
	RouteKey         string
	ConflictRouteKey string
	Now              time.Time
}

type WorkbookMutationResult struct {
	Payload          map[string]any
	StatusCode       int
	Replayed         bool
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	RowVersion       int64
	ViewSchemaID     string
	ChangedFieldKeys []string
}

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return "evidence: row version conflict"
}

type SameFieldConflictError struct {
	Conflict map[string]any
}

func (e *SameFieldConflictError) Error() string {
	return "evidence: same field conflict"
}

func newWorkbookFacade(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
	sourceMutations *SourceMutationService,
	objects objectstore.Store,
	conflictFields conflicttokens.FieldResolver,
	keepSaved conflicttokens.IdempotencyPort,
	projectionRows evidenceprojection.Rows,
	supportEffects evidenceprojection.SupportProjectionEffectsTx,
) (*WorkbookFacade, error) {
	switch {
	case pool == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: Postgres is required")
	case appender == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: Revisions is required")
	case intents == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: Collaboration is required")
	case sourceMutations == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: source mutations are required")
	case objects == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: object store is required")
	case conflictFields == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: conflict fields are required")
	case keepSaved == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: conflict idempotency is required")
	case projectionRows == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: projection rows are required")
	case supportEffects == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: support projection effects are required")
	}
	recordStore := records.NewStore()
	incidentAccess := incidents.NewAccess(pool)
	return &WorkbookFacade{
		pool:              pool,
		authStore:         authn.NewStore(pool),
		incidentAccess:    incidentAccess,
		recordStore:       recordStore,
		projectionRows:    projectionRows,
		supportEffects:    supportEffects,
		revisionHistory:   conflicttokens.NewRevisionWindowReader(),
		revisionAppender:  appender,
		sourceMutations:   sourceMutations,
		blobs:             blobRepository{db: pool},
		blobLifecycle:     blobLifecycleRepository{db: pool},
		conflictTokens:    conflictTokens,
		conflictFields:    conflictFields,
		conflictSnapshots: newEvidenceConflictSnapshotProjector(),
		keepSaved:         keepSaved,
		collaboration:     intents,
		objects:           objects,
		mutations:         sourceMutations.mutations,
	}, nil
}

func (f *WorkbookFacade) Create(ctx context.Context, command WorkbookCreateCommand) (WorkbookMutationResult, error) {
	request := command.Request
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.Actor.ID,
		ScopeKey:    command.IncidentID.String() + ":" + request.ViewSchemaID,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return WorkbookMutationResult{}, fmt.Errorf("decode replayed evidence create payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return WorkbookMutationResult{}, err
		}
		return WorkbookMutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, IncidentID: command.IncidentID, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return WorkbookMutationResult{}, fmt.Errorf("query evidence create idempotency: %w", err)
	}
	createParams := WorkbookCreateParams{
		Values:                 request.Values,
		InitialBlobWasSupplied: request.InitialObjectBlobID != nil,
	}
	if err := ValidateWorkbookCreateParams(createParams); err != nil {
		return WorkbookMutationResult{}, err
	}
	var observed *ObservedObject
	if request.InitialObjectBlobID != nil {
		var observeErr error
		observed, observeErr = f.observeInitialBlob(ctx, command.IncidentID, *request.InitialObjectBlobID)
		if observeErr != nil {
			return WorkbookMutationResult{}, observeErr
		}
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("begin evidence create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, command.IncidentID); err != nil {
		return WorkbookMutationResult{}, err
	}
	if request.InitialObjectBlobID != nil {
		initialBlob, commitRejection, finalizeErr := f.finalizeInitialBlobTx(
			ctx,
			tx,
			command.IncidentID,
			*request.InitialObjectBlobID,
			observed,
			command.Now.UTC(),
		)
		if finalizeErr != nil {
			if commitRejection {
				if err := tx.Commit(ctx); err != nil {
					return WorkbookMutationResult{}, fmt.Errorf("commit rejected evidence blob finalization: %w", err)
				}
			}
			return WorkbookMutationResult{}, finalizeErr
		}
		createParams.InitialBlob = initialBlob
		createParams.InitialBlobFinalized = true
	}
	if err := ValidateWorkbookCreateParams(createParams); err != nil {
		return WorkbookMutationResult{}, err
	}
	result, err := f.mutations.createTx(ctx, tx, evidenceCreateTxCommand{
		IncidentID:    command.IncidentID,
		ActorUserID:   command.Actor.ID,
		ViewSchemaID:  request.ViewSchemaID,
		ClientTxnID:   request.ClientTxnID,
		RequestID:     command.RequestID,
		Source:        command.RouteKey,
		MutationOrder: 1,
		Values:        request.Values,
		Now:           command.Now,
	}, createParams)
	if err != nil {
		if request.InitialObjectBlobID != nil && isEvidenceBlobUniqueViolation(err) {
			return WorkbookMutationResult{}, AttachRejectedError{ReasonCode: AttachReasonBlobNotVisible, Cause: ErrBlobNotAttachable}
		}
		return WorkbookMutationResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, command.RequestHash, http.StatusCreated, result.payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit evidence create transaction: %w", err)
	}
	return WorkbookMutationResult{
		Payload:          result.payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       command.IncidentID,
		RecordID:         result.recordID,
		ChangeSetID:      result.changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: result.changedFieldKeys,
	}, nil
}

func (f *WorkbookFacade) Patch(ctx context.Context, command WorkbookPatchCommand) (WorkbookMutationResult, error) {
	request := command.Request
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.Actor.ID,
		ScopeKey:    command.RecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredPayload(existing.ResponseJSON)
		if err != nil {
			return WorkbookMutationResult{}, fmt.Errorf("decode replayed evidence patch payload: %w", err)
		}
		return WorkbookMutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: command.RecordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return WorkbookMutationResult{}, fmt.Errorf("query evidence patch idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("begin evidence patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadEvidenceRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if meta.RecordType != "evidence" {
		return WorkbookMutationResult{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return WorkbookMutationResult{}, err
	}
	effectiveBeforeVersion := request.BaseRowVersion
	if meta.RowVersion != request.BaseRowVersion {
		if meta.RowVersion < request.BaseRowVersion {
			return WorkbookMutationResult{}, &RowVersionConflictError{RecordID: command.RecordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
		}
		windowRows, err := f.revisionHistory.LoadRevisionWindowTx(ctx, tx, command.RecordID, request.BaseRowVersion, meta.RowVersion)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		fieldDescriptors, err := f.conflictFields.ResolveViewSchema(request.ViewSchemaID)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		window, err := conflicttokens.BuildCanonicalPatchConflictWindow(command.RecordID, request.BaseRowVersion, meta.RowVersion, windowRows, fieldDescriptors, f.conflictSnapshots)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		if change, changed, ok := overlappingEvidencePatchChange(request.Changes, window.ChangedFields); ok {
			current, err := f.projectionRows.LoadEvidenceTx(ctx, tx, command.RecordID)
			if err != nil {
				return WorkbookMutationResult{}, err
			}
			conflictPayload, err := buildEvidenceSameFieldConflict(evidenceSameFieldConflictParams{
				RouteKey:          command.ConflictRouteKey,
				RecordID:          command.RecordID,
				ViewSchemaID:      request.ViewSchemaID,
				BaseRowVersion:    request.BaseRowVersion,
				CurrentRowVersion: meta.RowVersion,
				RequestHash:       command.RequestHash,
				Window:            window,
				Change:            change,
				Changed:           changed,
				CurrentRow:        current,
				FieldDescriptors:  fieldDescriptors,
				Codec:             f.conflictTokens,
			})
			if err != nil {
				return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
			}
			return WorkbookMutationResult{}, &SameFieldConflictError{Conflict: conflictPayload}
		}
		effectiveBeforeVersion = meta.RowVersion
	}
	beforeRow, err := f.projectionRows.LoadEvidenceTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	beforeSnapshot, err := f.revisionAppender.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := validateEvidencePatchReferencesTx(ctx, tx, meta.IncidentID, request); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.validateLifecyclePatchTx(ctx, tx, command.RecordID, request); err != nil {
		return WorkbookMutationResult{}, err
	}
	changed, err := f.applyPatchTx(ctx, tx, command.RecordID, request, command.Now.UTC())
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if !changed {
		return WorkbookMutationResult{}, &ValidationError{Field: "changes", ReasonCode: "no_effective_change"}
	}
	rowVersion, err := f.recordStore.AdvanceVersionTx(ctx, tx, command.RecordID, command.Actor.ID, command.Now.UTC())
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.sourceMutations.TouchWorkbookRowTx(ctx, tx, command.RecordID, command.Now.UTC()); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.projectionRows.RefreshEvidenceTx(ctx, tx, command.RecordID); err != nil {
		return WorkbookMutationResult{}, err
	}
	afterRow, err := f.projectionRows.LoadEvidenceTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	afterSnapshot, err := f.revisionAppender.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	changeSetID, err := f.revisionAppender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  meta.IncidentID,
		ActorUserID: command.Actor.ID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   command.Now.UTC(),
	})
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	beforeVersionID := workbookVersionID(command.RecordID, request.BaseRowVersion)
	if effectiveBeforeVersion != request.BaseRowVersion {
		beforeVersionID = workbookVersionID(command.RecordID, effectiveBeforeVersion)
	}
	afterVersionID := workbookVersionID(command.RecordID, rowVersion)
	if err := f.revisionAppender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		RecordID:        command.RecordID,
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  &beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.revisionAppender.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       command.RecordID,
		RowVersion:     rowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		LiveChange:     revisions.LiveRecordChange{BeforeValue: beforeRow, AfterValue: afterRow},
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	patchChangedFieldKeys := changedFieldKeys(beforeRow, afterRow)
	var affectedChanges []AttachRecordChange
	if slices.Contains(patchChangedFieldKeys, "evidence.lifecycle_state") {
		affectedChanges, err = refreshEvidenceSupportProjectionsTx(
			ctx,
			tx,
			f.supportEffects,
			meta.IncidentID,
			command.RecordID,
		)
		if err != nil {
			return WorkbookMutationResult{}, err
		}
	}
	if err := appendEvidenceRecordChangeIntentsTx(
		ctx,
		tx,
		f.collaboration,
		meta.IncidentID,
		command.Actor.ID,
		request.ClientTxnID,
		changeSetID,
		AttachRecordChange{
			RecordID:         command.RecordID,
			RowVersion:       rowVersion,
			ViewSchemaID:     request.ViewSchemaID,
			ChangedFieldKeys: patchChangedFieldKeys,
		},
		afterRow,
		affectedChanges,
		command.Now.UTC(),
	); err != nil {
		return WorkbookMutationResult{}, err
	}
	payload := buildMutationPayload(request.ViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, command.RequestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit evidence patch transaction: %w", err)
	}
	return WorkbookMutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       meta.IncidentID,
		RecordID:         command.RecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: patchChangedFieldKeys,
	}, nil
}

func (f *WorkbookFacade) validateLifecyclePatchTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, request WorkbookPatchRequest) error {
	changes := make([]WorkbookLifecyclePatchChange, 0, len(request.Changes))
	for _, change := range request.Changes {
		var text *string
		if change.Value != nil {
			text = change.Value.Text
		}
		changes = append(changes, WorkbookLifecyclePatchChange{FieldKey: change.FieldKey, Text: text})
	}
	return f.sourceMutations.ValidateWorkbookLifecyclePatchTx(ctx, tx, recordID, changes)
}

func (f *WorkbookFacade) applyPatchTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, request WorkbookPatchRequest, now time.Time) (bool, error) {
	changed := false
	for _, change := range request.Changes {
		if change.Value == nil {
			continue
		}
		if err := ValidateWorkbookDirectPatchChange(change.FieldKey, *change.Value); err != nil {
			return false, err
		}
		applied, err := f.sourceMutations.ApplyWorkbookDirectChangeTx(ctx, tx, recordID, change.FieldKey, *change.Value, now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
}

func validateEvidenceReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, values map[string]WorkbookFieldValue) error {
	for fieldKey, value := range values {
		if value.UUID == nil {
			continue
		}
		switch fieldKey {
		case "evidence.collector_party_id", "evidence.source_party_id":
			if err := validateEvidenceTargetRecordTx(ctx, tx, incidentID, *value.UUID, "party", fieldKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateEvidencePatchReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request WorkbookPatchRequest) error {
	for _, change := range request.Changes {
		if change.Value == nil || change.Value.UUID == nil {
			continue
		}
		switch change.FieldKey {
		case "evidence.collector_party_id", "evidence.source_party_id":
			if err := validateEvidenceTargetRecordTx(ctx, tx, incidentID, *change.Value.UUID, "party", change.FieldKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateEvidenceTargetRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND record_type = $3
       AND deleted_at IS NULL
)
`, incidentID, recordID, expectedType).Scan(&exists); err != nil {
		return fmt.Errorf("validate evidence reference target: %w", err)
	}
	if !exists {
		return &ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}

type evidenceRecordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func loadEvidenceRecordMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (evidenceRecordMeta, error) {
	var meta evidenceRecordMeta
	var deletedAt sql.NullTime
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&meta.IncidentID, &meta.RecordType, &meta.RowVersion, &deletedAt)
	if err != nil {
		return evidenceRecordMeta{}, err
	}
	if deletedAt.Valid {
		return evidenceRecordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return meta, nil
}

func changedFieldKeys(before map[string]any, after map[string]any) []string {
	afterCells, _ := after["cells"].(map[string]any)
	beforeCells := map[string]any{}
	if before != nil {
		beforeCells, _ = before["cells"].(map[string]any)
	}
	keys := make([]string, 0)
	for fieldKey, afterValue := range afterCells {
		if beforeValue, ok := beforeCells[fieldKey]; !ok || !reflect.DeepEqual(beforeValue, afterValue) {
			keys = append(keys, fieldKey)
		}
	}
	slices.Sort(keys)
	return keys
}

func workbookVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}

func buildMutationPayload(viewSchemaID string, changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{"view_schema_id": viewSchemaID, "change_set_id": changeSetID.String(), "row": row}
}

func extractPayloadUUID(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsed, nil
}

func adaptRevisionWindowError(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64, err error) error {
	if err == nil {
		return nil
	}
	var windowErr *conflicttokens.RevisionWindowError
	if errors.As(err, &windowErr) {
		return &RowVersionConflictError{RecordID: windowErr.RecordID, BaseRowVersion: windowErr.BaseRowVersion, CurrentRowVersion: windowErr.CurrentRowVersion}
	}
	return &RowVersionConflictError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
}

type evidenceSameFieldConflictParams struct {
	RouteKey          string
	RecordID          uuid.UUID
	ViewSchemaID      string
	BaseRowVersion    int64
	CurrentRowVersion int64
	RequestHash       []byte
	Window            conflicttokens.PatchConflictWindow
	Change            WorkbookPatchChange
	Changed           conflicttokens.PatchChangedField
	CurrentRow        map[string]any
	FieldDescriptors  conflicttokens.FieldDescriptorSet
	Codec             conflicttokens.ConflictTokenCodec
}

func overlappingEvidencePatchChange(changes []WorkbookPatchChange, changedFields map[string]conflicttokens.PatchChangedField) (WorkbookPatchChange, conflicttokens.PatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return WorkbookPatchChange{}, conflicttokens.PatchChangedField{}, false
}

func buildEvidenceSameFieldConflict(params evidenceSameFieldConflictParams) (map[string]any, error) {
	baseValue, ok := rowCellValue(params.Window.BaseRow, params.Change.FieldKey)
	if !ok {
		return nil, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	serverValue, ok := rowCellValue(params.CurrentRow, params.Change.FieldKey)
	if !ok {
		return nil, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	clientValue := params.Change.CanonicalValue
	conflictClass := params.FieldDescriptors.ConflictResolutionClass(params.Change.FieldKey)
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflictToken, err := evidenceConflictToken(params.RouteKey, params.RecordID, params.ViewSchemaID, params.Change.FieldKey, conflictClass, params.BaseRowVersion, params.CurrentRowVersion, params.RequestHash, params.Codec)
	if err != nil {
		return nil, err
	}
	conflict := map[string]any{
		"conflict_token":            conflictToken,
		"record_id":                 params.RecordID.String(),
		"field_key":                 params.Change.FieldKey,
		"conflict_resolution_class": conflictClass,
		"base_row_version":          params.BaseRowVersion,
		"current_row_version":       params.CurrentRowVersion,
		"client_value":              clientValue,
		"server_value":              serverValue,
		"server_updated_by":         params.Changed.ServerUpdatedBy.String(),
		"server_updated_at":         params.Changed.ServerUpdatedAt.UTC().Format(time.RFC3339Nano),
		"base_value":                baseValue,
	}
	if conflictClass == "text_compare_merge" {
		if suggested, ok := conflicttokens.SuggestedTextMergeValue(baseValue, serverValue, clientValue); ok {
			conflict["suggested_merged_value"] = suggested
		}
	}
	return conflict, nil
}

func rowCellValue(row map[string]any, fieldKey string) (any, bool) {
	cells, _ := row["cells"].(map[string]any)
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return nil, false
	}
	value, ok := cell["value"]
	return value, ok
}

func evidenceConflictToken(routeKey string, recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, codec conflicttokens.ConflictTokenCodec) (string, error) {
	return codec.Issue(conflicttokens.ConflictTokenClaims{
		RouteKey:                routeKey,
		RecordID:                recordID.String(),
		ViewSchemaID:            viewSchemaID,
		FieldKey:                fieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          baseRowVersion,
		CurrentRowVersion:       currentRowVersion,
		RequestHash:             conflicttokens.RequestHashTokenValue(requestHash),
	})
}
