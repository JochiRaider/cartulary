package parties

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	partysource "github.com/JochiRaider/cartulary/internal/modules/parties/internal/source"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type MutationFacade struct {
	pool              postgres.DB
	idempotency       IdempotencyCapability
	incidentAccess    IncidentStateCapability
	recordStore       RecordEnvelopeCapability
	projectionRows    partyprojection.Rows
	revisions         RevisionCapability
	conflictTokens    conflicttokens.ConflictTokenCodec
	conflictFields    conflicttokens.FieldResolver
	conflictSnapshots conflicttokens.RevisionSnapshotProjector
	keepSaved         KeepSavedCapability
}

type CreateCommand struct {
	ActorUserID uuid.UUID
	IncidentID  uuid.UUID
	Admission   CreateAdmission
	RequestID   string
	RouteKey    string
	Now         time.Time
}

type PatchCommand struct {
	ActorUserID      uuid.UUID
	RecordID         uuid.UUID
	Admission        PatchAdmission
	RequestID        string
	RouteKey         string
	ConflictRouteKey string
	Now              time.Time
}

type MutationOutcome string

const (
	MutationCreated   MutationOutcome = "created"
	MutationReused    MutationOutcome = "reused"
	MutationUpdated   MutationOutcome = "updated"
	MutationKeptSaved MutationOutcome = "kept_saved"
	MutationReplayed  MutationOutcome = "replayed"
)

type MutationResult struct {
	Outcome          MutationOutcome
	Row              map[string]any
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
	return "parties: row version conflict"
}

type SameFieldConflictError struct {
	Conflict SameFieldConflict
}

func (e *SameFieldConflictError) Error() string {
	return "parties: same field conflict"
}

type OptionalConflictValue struct {
	Present bool
	Value   any
}

type SameFieldConflict struct {
	ConflictToken           string
	RecordID                uuid.UUID
	FieldKey                string
	ConflictResolutionClass string
	BaseRowVersion          int64
	CurrentRowVersion       int64
	ClientValue             any
	ServerValue             any
	BaseValue               OptionalConflictValue
	ServerUpdatedBy         uuid.UUID
	ServerUpdatedAt         time.Time
	SuggestedMergedValue    OptionalConflictValue
}

func NewMutationContribution(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	dependencies MutationDependencies,
) (*MutationFacade, error) {
	if pool == nil {
		return nil, fmt.Errorf("parties mutation composition: Postgres is required")
	}
	if err := dependencies.validate(); err != nil {
		return nil, err
	}
	return &MutationFacade{
		pool:              pool,
		idempotency:       dependencies.Idempotency,
		incidentAccess:    dependencies.IncidentState,
		recordStore:       dependencies.RecordEnvelopes,
		projectionRows:    dependencies.Projections,
		revisions:         dependencies.Revisions,
		conflictTokens:    conflictTokens,
		conflictFields:    dependencies.ConflictFields,
		conflictSnapshots: newPartyConflictSnapshotProjector(),
		keepSaved:         dependencies.KeepSaved,
	}, nil
}

func (f *MutationFacade) Create(ctx context.Context, command CreateCommand) (MutationResult, error) {
	request := command.Admission
	if command.RouteKey != workbookCreateOperation {
		return MutationResult{}, &ValidationError{Field: "operation_id", ReasonCode: "invalid_value"}
	}
	idempotencyKey := IdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.IncidentID.String() + ":" + ViewSchemaID,
		ClientTxnID: request.clientTxnID,
	}
	if existing, err := f.idempotency.Get(ctx, idempotencyKey, request.requestHash[:]); err == nil {
		if !bytes.Equal(existing.RequestHash, request.requestHash[:]) {
			return MutationResult{}, ErrClientTxnConflict
		}
		if existing.Result.Kind() != StoredMutationCreate {
			return MutationResult{}, ErrStoredMutationKindMismatch
		}
		stored, ok := existing.Result.RowMutationResult()
		if !ok || stored.ViewSchemaID != ViewSchemaID {
			return MutationResult{}, ErrStoredMutationKindMismatch
		}
		return MutationResult{
			Outcome: MutationReplayed, Row: stored.Row, IncidentID: command.IncidentID,
			RecordID: stored.RecordID, ChangeSetID: stored.ChangeSetID,
			ViewSchemaID: ViewSchemaID, ClientTxnID: request.clientTxnID,
		}, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return MutationResult{}, fmt.Errorf("query party create idempotency: %w", err)
	}
	params := partysource.CreateParams{Values: request.values}
	if err := validateCreateParams(params); err != nil {
		return MutationResult{}, err
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin party create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := f.incidentAccess.RequireOpenTx(ctx, tx, command.IncidentID); err != nil {
		return MutationResult{}, err
	}
	now := command.Now.UTC()
	if recordID, found, err := partysource.FindReusablePartyTx(ctx, tx, command.IncidentID, params); err != nil || found {
		if err != nil {
			return MutationResult{}, err
		}
		result, err := f.reuseCreateTx(ctx, tx, command, idempotencyKey, recordID, now)
		if err != nil {
			return MutationResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return MutationResult{}, fmt.Errorf("commit party reuse transaction: %w", err)
		}
		return result, nil
	}

	recordID, err := f.recordStore.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      command.IncidentID,
		RecordType:      "party",
		CreatedByUserID: command.ActorUserID,
		CreatedAt:       now,
		UpdatedByUserID: command.ActorUserID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return MutationResult{}, err
	}
	if err := partysource.InsertPartyTx(ctx, tx, recordID, command.IncidentID, params, now); err != nil {
		return MutationResult{}, err
	}
	if err := f.projectionRows.RefreshPartyTx(ctx, tx, recordID); err != nil {
		return MutationResult{}, err
	}
	row, err := f.projectionRows.LoadPartyTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := f.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  command.IncidentID,
		ActorUserID: command.ActorUserID,
		Source:      command.RouteKey,
		ClientTxnID: &request.clientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   now,
	})
	if err != nil {
		return MutationResult{}, err
	}
	afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		RecordID:       recordID,
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterSnapshot:  &afterSnapshot,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := f.revisions.AppendRecordRevisionAndIntentTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID:   changeSetID,
		RecordID:      recordID,
		RowVersion:    1,
		AfterSnapshot: &afterSnapshot,
		LiveChange:    revisions.LiveRecordChange{AfterValue: row},
	}); err != nil {
		return MutationResult{}, err
	}
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, request.requestHash[:], NewStoredCreateResult(StoredRowMutationResult{
		Outcome: MutationCreated, ViewSchemaID: ViewSchemaID, RecordID: recordID, ChangeSetID: changeSetID, Row: row,
	})); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit party create transaction: %w", err)
	}
	return MutationResult{
		Outcome:          MutationCreated,
		Row:              row,
		IncidentID:       command.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.clientTxnID,
		RowVersion:       1,
		ViewSchemaID:     ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, row),
	}, nil
}

func (f *MutationFacade) reuseCreateTx(ctx context.Context, tx pgx.Tx, command CreateCommand, idempotencyKey IdempotencyKey, recordID uuid.UUID, now time.Time) (MutationResult, error) {
	request := command.Admission
	if err := f.projectionRows.RefreshPartyTx(ctx, tx, recordID); err != nil {
		return MutationResult{}, err
	}
	row, err := f.projectionRows.LoadPartyTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := f.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  command.IncidentID,
		ActorUserID: command.ActorUserID,
		Source:      command.RouteKey,
		ClientTxnID: &request.clientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   now,
	})
	if err != nil {
		return MutationResult{}, err
	}
	rowVersion, err := rowVersionFromGenericRow(row)
	if err != nil {
		return MutationResult{}, err
	}
	afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, rowVersion)
	if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		RecordID:       recordID,
		OperationKind:  "reuse",
		AfterVersionID: &afterVersionID,
		AfterSnapshot:  &afterSnapshot,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, request.requestHash[:], NewStoredCreateResult(StoredRowMutationResult{
		Outcome: MutationReused, ViewSchemaID: ViewSchemaID, RecordID: recordID, ChangeSetID: changeSetID, Row: row,
	})); err != nil {
		return MutationResult{}, err
	}
	return MutationResult{
		Outcome:          MutationReused,
		Row:              row,
		IncidentID:       command.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.clientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     ViewSchemaID,
		ChangedFieldKeys: []string{},
	}, nil
}

func (f *MutationFacade) Patch(ctx context.Context, command PatchCommand) (MutationResult, error) {
	request := command.Admission
	if command.RouteKey != request.operationID ||
		(request.operationID != workbookPatchOperation && request.operationID != workbookConflictResolveOperation) {
		return MutationResult{}, &ValidationError{Field: "operation_id", ReasonCode: "invalid_value"}
	}
	idempotencyKey := IdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.RecordID.String(),
		ClientTxnID: request.clientTxnID,
	}
	if existing, err := f.idempotency.Get(ctx, idempotencyKey, request.requestHash[:]); err == nil {
		if !bytes.Equal(existing.RequestHash, request.requestHash[:]) {
			return MutationResult{}, ErrClientTxnConflict
		}
		if existing.Result.Kind() != StoredMutationPatch {
			return MutationResult{}, ErrStoredMutationKindMismatch
		}
		stored, ok := existing.Result.RowMutationResult()
		if !ok || stored.ViewSchemaID != ViewSchemaID || stored.RecordID != command.RecordID {
			return MutationResult{}, ErrStoredMutationKindMismatch
		}
		return MutationResult{
			Outcome: MutationReplayed, Row: stored.Row, RecordID: command.RecordID,
			ChangeSetID: stored.ChangeSetID, ViewSchemaID: ViewSchemaID,
			ClientTxnID: request.clientTxnID,
		}, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return MutationResult{}, fmt.Errorf("query party patch idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin party patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := partysource.LoadRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if meta.RecordType != "party" {
		return MutationResult{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.RequireOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return MutationResult{}, err
	}
	effectiveBeforeVersion := request.baseRowVersion
	if meta.RowVersion != request.baseRowVersion {
		if meta.RowVersion < request.baseRowVersion {
			return MutationResult{}, &RowVersionConflictError{RecordID: command.RecordID, BaseRowVersion: request.baseRowVersion, CurrentRowVersion: meta.RowVersion}
		}
		windowRows, err := f.revisions.LoadRevisionWindowTx(ctx, tx, command.RecordID, request.baseRowVersion, meta.RowVersion)
		if err != nil {
			return MutationResult{}, adaptRevisionWindowError(command.RecordID, request.baseRowVersion, meta.RowVersion, err)
		}
		fieldDescriptors, err := f.conflictFields.ResolveViewSchema(ViewSchemaID)
		if err != nil {
			return MutationResult{}, adaptRevisionWindowError(command.RecordID, request.baseRowVersion, meta.RowVersion, err)
		}
		window, err := conflicttokens.BuildCanonicalPatchConflictWindow(command.RecordID, request.baseRowVersion, meta.RowVersion, windowRows, fieldDescriptors, f.conflictSnapshots)
		if err != nil {
			return MutationResult{}, adaptRevisionWindowError(command.RecordID, request.baseRowVersion, meta.RowVersion, err)
		}
		if change, changed, ok := overlappingPartyPatchChange(request.changes, window.ChangedFields); ok {
			current, err := f.projectionRows.LoadPartyTx(ctx, tx, command.RecordID)
			if err != nil {
				return MutationResult{}, err
			}
			conflictPayload, err := buildPartySameFieldConflict(partySameFieldConflictParams{
				RouteKey:          command.ConflictRouteKey,
				RecordID:          command.RecordID,
				ViewSchemaID:      ViewSchemaID,
				BaseRowVersion:    request.baseRowVersion,
				CurrentRowVersion: meta.RowVersion,
				RequestHash:       request.requestHash[:],
				Window:            window,
				Change:            change,
				Changed:           changed,
				CurrentRow:        current,
				FieldDescriptors:  fieldDescriptors,
				Codec:             f.conflictTokens,
			})
			if err != nil {
				return MutationResult{}, adaptRevisionWindowError(command.RecordID, request.baseRowVersion, meta.RowVersion, err)
			}
			return MutationResult{}, &SameFieldConflictError{Conflict: conflictPayload}
		}
		effectiveBeforeVersion = meta.RowVersion
	}
	beforeRow, err := f.projectionRows.LoadPartyTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	beforeSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := partysource.PreparePatchActiveKeysTx(ctx, tx, meta.IncidentID, command.RecordID, sourcePatchChanges(request.changes)); err != nil {
		return MutationResult{}, err
	}
	changed, err := f.applyPatchTx(ctx, tx, command.RecordID, request, command.Now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if !changed {
		return MutationResult{}, &ValidationError{Field: "changes", ReasonCode: "no_effective_change"}
	}
	rowVersion, err := f.recordStore.AdvanceVersionTx(ctx, tx, command.RecordID, command.ActorUserID, command.Now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if err := partysource.TouchPartyTx(ctx, tx, command.RecordID, command.Now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := f.projectionRows.RefreshPartyTx(ctx, tx, command.RecordID); err != nil {
		return MutationResult{}, err
	}
	afterRow, err := f.projectionRows.LoadPartyTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := f.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  meta.IncidentID,
		ActorUserID: command.ActorUserID,
		Source:      command.RouteKey,
		ClientTxnID: &request.clientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   command.Now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	beforeVersionID := workbookVersionID(command.RecordID, request.baseRowVersion)
	if effectiveBeforeVersion != request.baseRowVersion {
		beforeVersionID = workbookVersionID(command.RecordID, effectiveBeforeVersion)
	}
	afterVersionID := workbookVersionID(command.RecordID, rowVersion)
	if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
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
		return MutationResult{}, err
	}
	if err := f.revisions.AppendRecordRevisionAndIntentTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       command.RecordID,
		RowVersion:     rowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		LiveChange: revisions.LiveRecordChange{
			BeforeValue: beforeRow,
			AfterValue:  afterRow,
		},
	}); err != nil {
		return MutationResult{}, err
	}
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, request.requestHash[:], NewStoredPatchResult(StoredRowMutationResult{
		Outcome: MutationUpdated, ViewSchemaID: ViewSchemaID, RecordID: command.RecordID, ChangeSetID: changeSetID, Row: afterRow,
	})); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit party patch transaction: %w", err)
	}
	return MutationResult{
		Outcome:          MutationUpdated,
		Row:              afterRow,
		IncidentID:       meta.IncidentID,
		RecordID:         command.RecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.clientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeRow, afterRow),
	}, nil
}

func (f *MutationFacade) applyPatchTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, request PatchAdmission, now time.Time) (bool, error) {
	changed := false
	for _, change := range request.changes {
		applied, err := partysource.ApplyDirectChangeTx(ctx, tx, recordID, partysource.PatchChange{
			FieldKey: change.fieldKey,
			Value:    change.value,
		}, now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	return changed, nil
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

func rowVersionFromGenericRow(row map[string]any) (int64, error) {
	raw, ok := row["row_version"]
	if !ok {
		return 0, fmt.Errorf("generic row missing row_version")
	}
	switch value := raw.(type) {
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case float64:
		return int64(value), nil
	default:
		return 0, fmt.Errorf("generic row has unexpected row_version type %T", value)
	}
}

func workbookVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}

func validateCreateParams(params partysource.CreateParams) error {
	if !partysource.HasStoredText(params.Values, "party.display_name") {
		return &ValidationError{Field: "party.display_name", ReasonCode: "missing_required_field"}
	}
	if !partysource.HasStoredText(params.Values, "party.party_kind") {
		return &ValidationError{Field: "party.party_kind", ReasonCode: "missing_required_field"}
	}
	return nil
}

func sourcePatchChanges(changes []patchChange) []partysource.PatchChange {
	result := make([]partysource.PatchChange, 0, len(changes))
	for _, change := range changes {
		result = append(result, partysource.PatchChange{FieldKey: change.fieldKey, Value: change.value})
	}
	return result
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

type partySameFieldConflictParams struct {
	RouteKey          string
	RecordID          uuid.UUID
	ViewSchemaID      string
	BaseRowVersion    int64
	CurrentRowVersion int64
	RequestHash       []byte
	Window            conflicttokens.PatchConflictWindow
	Change            patchChange
	Changed           conflicttokens.PatchChangedField
	CurrentRow        map[string]any
	FieldDescriptors  conflicttokens.FieldDescriptorSet
	Codec             conflicttokens.ConflictTokenCodec
}

func overlappingPartyPatchChange(changes []patchChange, changedFields map[string]conflicttokens.PatchChangedField) (patchChange, conflicttokens.PatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.fieldKey]
		if ok {
			return change, changed, true
		}
	}
	return patchChange{}, conflicttokens.PatchChangedField{}, false
}

func buildPartySameFieldConflict(params partySameFieldConflictParams) (SameFieldConflict, error) {
	baseValue, ok := rowCellValue(params.Window.BaseRow, params.Change.fieldKey)
	if !ok {
		return SameFieldConflict{}, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	serverValue, ok := rowCellValue(params.CurrentRow, params.Change.fieldKey)
	if !ok {
		return SameFieldConflict{}, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	clientValue, present := params.Change.value.StoredValue()
	var admittedClientValue any
	if present {
		admittedClientValue = clientValue
	}
	conflictClass := params.FieldDescriptors.ConflictResolutionClass(params.Change.fieldKey)
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflictToken, err := partyConflictToken(params.RouteKey, params.RecordID, params.ViewSchemaID, params.Change.fieldKey, conflictClass, params.BaseRowVersion, params.CurrentRowVersion, params.RequestHash, params.Codec)
	if err != nil {
		return SameFieldConflict{}, err
	}
	conflict := SameFieldConflict{
		ConflictToken:           conflictToken,
		RecordID:                params.RecordID,
		FieldKey:                params.Change.fieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          params.BaseRowVersion,
		CurrentRowVersion:       params.CurrentRowVersion,
		ClientValue:             admittedClientValue,
		ServerValue:             serverValue,
		BaseValue:               OptionalConflictValue{Present: true, Value: baseValue},
		ServerUpdatedBy:         params.Changed.ServerUpdatedBy,
		ServerUpdatedAt:         params.Changed.ServerUpdatedAt.UTC(),
	}
	if conflictClass == "text_compare_merge" {
		if suggested, ok := conflicttokens.SuggestedTextMergeValue(baseValue, serverValue, admittedClientValue); ok {
			conflict.SuggestedMergedValue = OptionalConflictValue{Present: true, Value: suggested}
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

func partyConflictToken(routeKey string, recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, codec conflicttokens.ConflictTokenCodec) (string, error) {
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
