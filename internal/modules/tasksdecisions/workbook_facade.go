package tasksdecisions

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const TaskRequestsViewSchemaID = "cartulary.view.task_requests.v1"

type MutationFacade struct {
	pool             postgres.DB
	idempotency      IdempotencyCapability
	keepSaved        conflicttokens.IdempotencyPort
	incidentAccess   IncidentStateCapability
	memberReferences MemberReferenceCapability
	recordStore      RecordEnvelopeCapability
	linkStore        LinkCapability
	projectionRows   taskdecisionprojection.Rows
	revisions        RevisionCapability
	conflictTokens   conflicttokens.ConflictTokenCodec
	conflictFields   conflicttokens.FieldResolver
}

type WorkbookCreateRequest struct {
	ViewSchemaID string
	ClientTxnID  string
	Values       map[string]FieldValue
	Collections  map[string]WorkbookCollectionActionPayload
}

type WorkbookPatchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []WorkbookPatchChange
}

type WorkbookPatchChange struct {
	FieldKey       string
	Value          *FieldValue
	Collection     *WorkbookCollectionActionPayload
	CanonicalValue any
}

type WorkbookCollectionActionPayload struct {
	Actions []WorkbookCollectionAction
}

type WorkbookCollectionAction struct {
	Op             string
	RawText        string
	LinkedRecordID *uuid.UUID
	PartyID        *uuid.UUID
	ItemRef        string
	RiskRefText    string
	NormalizedText string
}

type WorkbookCreateCommand struct {
	ActorUserID uuid.UUID
	IncidentID  uuid.UUID
	Request     WorkbookCreateRequest
	RequestHash []byte
	RequestID   string
	RouteKey    string
	Now         time.Time
}

type WorkbookPatchCommand struct {
	ActorUserID      uuid.UUID
	RecordID         uuid.UUID
	Request          WorkbookPatchRequest
	RequestHash      []byte
	RequestID        string
	RouteKey         string
	ConflictRouteKey string
	Now              time.Time
}

type WorkbookMutationResult struct {
	Row              map[string]any
	Created          bool
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
	return "tasksdecisions: row version conflict"
}

type SameFieldConflictError struct {
	Conflict map[string]any
}

func (e *SameFieldConflictError) Error() string {
	return "tasksdecisions: same field conflict"
}

func NewMutationContribution(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	dependencies MutationDependencies,
) (*MutationFacade, error) {
	if pool == nil {
		return nil, fmt.Errorf("tasks/decisions mutation composition: Postgres is required")
	}
	if err := dependencies.validate(); err != nil {
		return nil, err
	}
	return &MutationFacade{
		pool:             pool,
		idempotency:      dependencies.Idempotency,
		keepSaved:        dependencies.KeepSavedIdempotency,
		incidentAccess:   dependencies.IncidentState,
		memberReferences: dependencies.MemberReferences,
		recordStore:      dependencies.RecordEnvelopes,
		linkStore:        dependencies.Links,
		projectionRows:   dependencies.Projections,
		revisions:        dependencies.Revisions,
		conflictTokens:   conflictTokens,
		conflictFields:   dependencies.ConflictFields,
	}, nil
}

func (f *MutationFacade) Create(ctx context.Context, command WorkbookCreateCommand) (WorkbookMutationResult, error) {
	request := command.Request
	idempotencyKey := IdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.IncidentID.String() + ":" + request.ViewSchemaID,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.idempotency.Get(ctx, idempotencyKey, command.RequestHash); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return WorkbookMutationResult{}, ErrClientTxnConflict
		}
		if existing.Result.Kind() != StoredMutationCreate {
			return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
		}
		stored, ok := existing.Result.WorkbookResult()
		if !ok || stored.ViewSchemaID != request.ViewSchemaID {
			return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
		}
		return WorkbookMutationResult{Row: stored.Row, Replayed: true, IncidentID: command.IncidentID, RecordID: stored.RecordID, ChangeSetID: stored.ChangeSetID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return WorkbookMutationResult{}, fmt.Errorf("query task/decision create idempotency: %w", err)
	}
	if err := validateCreateRequest(request); err != nil {
		return WorkbookMutationResult{}, err
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("begin task/decision create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, command.IncidentID); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := validateCreateReferencesTx(ctx, tx, f.memberReferences, f.linkStore, command.IncidentID, request); err != nil {
		return WorkbookMutationResult{}, err
	}
	now := command.Now.UTC()
	recordType := recordTypeForView(request.ViewSchemaID)
	if recordType == "" {
		return WorkbookMutationResult{}, &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
	recordID, err := f.recordStore.InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      command.IncidentID,
		RecordType:      recordType,
		CreatedByUserID: command.ActorUserID,
		CreatedAt:       now,
		UpdatedByUserID: command.ActorUserID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	directLinkMutations := make([]links.RecordLinkMutation, 0)
	switch request.ViewSchemaID {
	case TaskRequestsViewSchemaID:
		if err := tasksource.InsertTaskRequestTx(ctx, tx, recordID, command.IncidentID, command.ActorUserID, TaskCreateParams{Values: request.Values}, now); err != nil {
			return WorkbookMutationResult{}, err
		}
		if value, ok := request.Values[TaskDecisionRecordFieldKey]; ok && value.UUID != nil {
			directLinkMutations, err = syncTaskDecisionReferenceTx(ctx, tx, f.linkStore, command.IncidentID, recordID, command.ActorUserID, value.UUID, now)
			if err != nil {
				return WorkbookMutationResult{}, err
			}
		}
	case DecisionsViewSchemaID:
		if err := tasksource.InsertDecisionTx(ctx, tx, recordID, command.IncidentID, command.ActorUserID, DecisionCreateParams{Values: request.Values}, now); err != nil {
			return WorkbookMutationResult{}, err
		}
	default:
		return WorkbookMutationResult{}, &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
	collectionMutations, err := f.applyCollectionPayloadsTx(ctx, tx, command.IncidentID, recordID, command.ActorUserID, request.Collections, now)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	collectionMutations = append(directLinkMutations, collectionMutations...)
	afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.refreshRowTx(ctx, tx, request.ViewSchemaID, recordID); err != nil {
		return WorkbookMutationResult{}, err
	}
	row, err := f.loadProjectionRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	changeSetID, err := f.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  command.IncidentID,
		ActorUserID: command.ActorUserID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   now,
	})
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	afterVersionID := supersedeVersionID(recordID, 1)
	if err := f.revisions.AppendCapturedRecordMutationTx(ctx, tx, revisions.AppendCapturedRecordMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		RecordID:       recordID,
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterSnapshot:  &afterSnapshot,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.appendRecordLinkMutationsTx(ctx, tx, changeSetID, 2, collectionMutations); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.revisions.AppendCapturedRecordRevisionTx(ctx, tx, revisions.AppendCapturedRecordRevisionParams{
		ChangeSetID:   changeSetID,
		RecordID:      recordID,
		RowVersion:    1,
		AfterSnapshot: &afterSnapshot,
		LiveChange:    revisions.LiveRecordChange{AfterValue: row},
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	storedResult := NewStoredCreateResult(StoredWorkbookResult{
		ViewSchemaID: request.ViewSchemaID,
		RecordID:     recordID,
		ChangeSetID:  changeSetID,
		Row:          row,
	})
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, storedResult); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit task/decision create transaction: %w", err)
	}
	return WorkbookMutationResult{
		Row:              row,
		Created:          true,
		IncidentID:       command.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, row),
	}, nil
}

func (f *MutationFacade) Patch(ctx context.Context, command WorkbookPatchCommand) (WorkbookMutationResult, error) {
	request := command.Request
	idempotencyKey := IdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.RecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := f.idempotency.Get(ctx, idempotencyKey, command.RequestHash); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return WorkbookMutationResult{}, ErrClientTxnConflict
		}
		if existing.Result.Kind() != StoredMutationPatch {
			return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
		}
		stored, ok := existing.Result.WorkbookResult()
		if !ok || stored.ViewSchemaID != request.ViewSchemaID || stored.RecordID != command.RecordID {
			return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
		}
		return WorkbookMutationResult{Row: stored.Row, Replayed: true, RecordID: command.RecordID, ChangeSetID: stored.ChangeSetID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return WorkbookMutationResult{}, fmt.Errorf("query task/decision patch idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("begin task/decision patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadSupersedeRecordMetaForUpdateTx(ctx, tx, f.recordStore, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if !recordTypeMatchesView(meta.RecordType, request.ViewSchemaID) {
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
		windowRows, err := f.revisions.LoadRevisionWindowTx(ctx, tx, command.RecordID, request.BaseRowVersion, meta.RowVersion)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		fieldDescriptors, err := f.conflictFields.ResolveViewSchema(request.ViewSchemaID)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		projector, ok := taskDecisionConflictSnapshotProjector(request.ViewSchemaID)
		if !ok {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, conflicttokens.ErrInvalidSnapshotProjector)
		}
		window, err := conflicttokens.BuildCanonicalPatchConflictWindow(command.RecordID, request.BaseRowVersion, meta.RowVersion, windowRows, fieldDescriptors, projector)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		if change, changed, ok := overlappingPatchChange(request.Changes, window.ChangedFields); ok {
			current, err := f.loadProjectionRowTx(ctx, tx, request.ViewSchemaID, command.RecordID)
			if err != nil {
				return WorkbookMutationResult{}, err
			}
			conflictPayload, err := buildSameFieldConflict(sameFieldConflictParams{
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
	beforeRow, err := f.loadProjectionRowTx(ctx, tx, request.ViewSchemaID, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	beforeSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := validatePatchReferencesTx(ctx, tx, f.memberReferences, f.linkStore, meta.IncidentID, request); err != nil {
		return WorkbookMutationResult{}, err
	}
	changed, collectionMutations, err := f.applyPatchTx(ctx, tx, meta.IncidentID, command.RecordID, command.ActorUserID, request, command.Now.UTC())
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if !changed {
		return WorkbookMutationResult{}, &ValidationError{Field: "changes", ReasonCode: "no_effective_change"}
	}
	rowVersion, err := f.recordStore.AdvanceVersionTx(ctx, tx, command.RecordID, command.ActorUserID, command.Now.UTC())
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.touchSourceRowTx(ctx, tx, request.ViewSchemaID, command.RecordID, command.Now.UTC()); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.refreshRowTx(ctx, tx, request.ViewSchemaID, command.RecordID); err != nil {
		return WorkbookMutationResult{}, err
	}
	afterRow, err := f.loadProjectionRowTx(ctx, tx, request.ViewSchemaID, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	changeSetID, err := f.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  meta.IncidentID,
		ActorUserID: command.ActorUserID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   command.Now.UTC(),
	})
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	beforeVersionID := supersedeVersionID(command.RecordID, request.BaseRowVersion)
	if effectiveBeforeVersion != request.BaseRowVersion {
		beforeVersionID = supersedeVersionID(command.RecordID, effectiveBeforeVersion)
	}
	afterVersionID := supersedeVersionID(command.RecordID, rowVersion)
	if err := f.revisions.AppendCapturedRecordMutationTx(ctx, tx, revisions.AppendCapturedRecordMutationParams{
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
	if err := f.appendRecordLinkMutationsTx(ctx, tx, changeSetID, 2, collectionMutations); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.revisions.AppendCapturedRecordRevisionTx(ctx, tx, revisions.AppendCapturedRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       command.RecordID,
		RowVersion:     rowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		LiveChange:     revisions.LiveRecordChange{BeforeValue: beforeRow, AfterValue: afterRow},
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	storedResult := NewStoredPatchResult(StoredWorkbookResult{
		ViewSchemaID: request.ViewSchemaID,
		RecordID:     command.RecordID,
		ChangeSetID:  changeSetID,
		Row:          afterRow,
	})
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, storedResult); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit task/decision patch transaction: %w", err)
	}
	return WorkbookMutationResult{
		Row:              afterRow,
		IncidentID:       meta.IncidentID,
		RecordID:         command.RecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeRow, afterRow),
	}, nil
}

func validateCreateRequest(request WorkbookCreateRequest) error {
	switch request.ViewSchemaID {
	case TaskRequestsViewSchemaID:
		return policy.ValidateTaskCreateParams(TaskCreateParams{Values: request.Values})
	case DecisionsViewSchemaID:
		return policy.ValidateDecisionCreateParams(DecisionCreateParams{Values: request.Values})
	default:
		return &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func validateCreateReferencesTx(
	ctx context.Context,
	tx pgx.Tx,
	members MemberReferenceCapability,
	linkStore LinkCapability,
	incidentID uuid.UUID,
	request WorkbookCreateRequest,
) error {
	for fieldKey, value := range request.Values {
		if value.UUID != nil && policy.IsMemberUserReferenceField(fieldKey) {
			if err := members.ValidateIncidentMemberUserTx(ctx, tx, incidentID, *value.UUID, fieldKey); err != nil {
				return err
			}
		}
		if value.UUID != nil {
			if err := validateDirectReferenceTx(ctx, tx, incidentID, fieldKey, *value.UUID); err != nil {
				return err
			}
		}
	}
	for fieldKey, payload := range request.Collections {
		if err := validateCollectionPayloadTx(ctx, tx, linkStore, incidentID, fieldKey, payload); err != nil {
			return err
		}
	}
	return nil
}

func validatePatchReferencesTx(
	ctx context.Context,
	tx pgx.Tx,
	members MemberReferenceCapability,
	linkStore LinkCapability,
	incidentID uuid.UUID,
	request WorkbookPatchRequest,
) error {
	for _, change := range request.Changes {
		if change.Value != nil && change.Value.UUID != nil && policy.IsMemberUserReferenceField(change.FieldKey) {
			if err := members.ValidateIncidentMemberUserTx(ctx, tx, incidentID, *change.Value.UUID, change.FieldKey); err != nil {
				return err
			}
		}
		if change.Value != nil && change.Value.UUID != nil {
			if err := validateDirectReferenceTx(ctx, tx, incidentID, change.FieldKey, *change.Value.UUID); err != nil {
				return err
			}
		}
		if change.Collection != nil {
			if err := validateCollectionPayloadTx(ctx, tx, linkStore, incidentID, change.FieldKey, *change.Collection); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateDirectReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey string, recordID uuid.UUID) error {
	recordType, ok := policy.DirectReferenceRecordType(fieldKey)
	if !ok {
		return nil
	}
	return validateTargetRecordTx(ctx, tx, incidentID, recordID, recordType, fieldKey)
}

func validateIncidentMemberUserTx(ctx context.Context, tx pgx.Tx, incidentID, userID uuid.UUID, field string) error {
	return tasksource.ValidateMemberUserTx(ctx, tx, incidentID, userID, field)
}

func validateTargetRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
	return tasksource.ValidateTargetRecordTx(ctx, tx, incidentID, recordID, expectedType, field)
}

func (f *MutationFacade) applyPatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, request WorkbookPatchRequest, now time.Time) (bool, []links.RecordLinkMutation, error) {
	changed := false
	collectionMutations := make([]links.RecordLinkMutation, 0)
	var beforeTask TaskLifecycleState
	var beforeDecisionStatus string
	var err error
	if request.ViewSchemaID == TaskRequestsViewSchemaID && touchesAnyField(request.Changes, "task.status", "task.blocked_reason", "task.completed_at", "task.owner_user_id") {
		beforeTask, err = tasksource.LoadTaskLifecycleStateTx(ctx, tx, recordID)
		if err != nil {
			return false, nil, err
		}
	}
	if request.ViewSchemaID == DecisionsViewSchemaID {
		if err := validateDecisionMachineConsistentTx(ctx, tx, recordID); err != nil {
			return false, nil, err
		}
		if touchesField(request.Changes, "decision.status") {
			beforeDecisionStatus, err = tasksource.LoadDecisionStatusTx(ctx, tx, recordID)
			if err != nil {
				return false, nil, err
			}
		}
	}
	for _, change := range request.Changes {
		if change.Value != nil {
			applied, mutations, err := f.applyDirectChangeTx(ctx, tx, incidentID, recordID, actorID, request.ViewSchemaID, change, now)
			if err != nil {
				return false, nil, err
			}
			changed = changed || applied
			collectionMutations = append(collectionMutations, mutations...)
			continue
		}
		if change.Collection != nil {
			applied, mutations, err := f.applyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, change.FieldKey, *change.Collection, now)
			if err != nil {
				return false, nil, err
			}
			changed = changed || applied
			collectionMutations = append(collectionMutations, mutations...)
		}
	}
	if request.ViewSchemaID == TaskRequestsViewSchemaID && touchesAnyField(request.Changes, "task.status", "task.blocked_reason", "task.completed_at", "task.owner_user_id") {
		applied, err := tasksource.NormalizeTaskLifecycleTx(ctx, tx, recordID, beforeTask, touchesField(request.Changes, "task.completed_at"), now)
		if err != nil {
			return false, nil, err
		}
		changed = changed || applied
	}
	if request.ViewSchemaID == DecisionsViewSchemaID && touchesField(request.Changes, "decision.status") {
		afterDecisionStatus, err := tasksource.LoadDecisionStatusTx(ctx, tx, recordID)
		if err != nil {
			return false, nil, err
		}
		if err := policy.ValidateDecisionStatusTransition(beforeDecisionStatus, afterDecisionStatus); err != nil {
			return false, nil, err
		}
		if err := validateDecisionMachineConsistentTx(ctx, tx, recordID); err != nil {
			return false, nil, err
		}
	}
	return changed, collectionMutations, nil
}

func (f *MutationFacade) applyDirectChangeTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, viewSchemaID string, change WorkbookPatchChange, now time.Time) (bool, []links.RecordLinkMutation, error) {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		if err := policy.ValidateTaskDirectPatchChange(change.FieldKey, *change.Value); err != nil {
			return false, nil, err
		}
		return applyTaskDirectChangeTx(ctx, tx, f.linkStore, incidentID, recordID, actorID, change.FieldKey, *change.Value, now)
	case DecisionsViewSchemaID:
		if err := policy.ValidateDecisionDirectPatchChange(change.FieldKey, *change.Value); err != nil {
			return false, nil, err
		}
		changed, err := tasksource.ApplyDecisionDirectChangeTx(ctx, tx, recordID, change.FieldKey, *change.Value, now)
		return changed, nil, err
	default:
		return false, nil, &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func (f *MutationFacade) applyCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, collections map[string]WorkbookCollectionActionPayload, now time.Time) ([]links.RecordLinkMutation, error) {
	fieldKeys := make([]string, 0, len(collections))
	for fieldKey := range collections {
		fieldKeys = append(fieldKeys, fieldKey)
	}
	slices.Sort(fieldKeys)
	mutations := make([]links.RecordLinkMutation, 0)
	for _, fieldKey := range fieldKeys {
		_, applied, err := f.applyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, fieldKey, collections[fieldKey], now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}

func validateCollectionPayloadTx(ctx context.Context, tx pgx.Tx, linkStore LinkCapability, incidentID uuid.UUID, fieldKey string, payload WorkbookCollectionActionPayload) error {
	descriptor, ok := lookupCollectionDescriptor(fieldKey)
	if !ok {
		return &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
	}
	command, err := recordRefValidation(incidentID, descriptor, payload)
	if err != nil {
		return err
	}
	return linkStore.ValidateRecordRefCollectionTx(ctx, tx, command)
}

func (f *MutationFacade) applyCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, fieldKey string, payload WorkbookCollectionActionPayload, now time.Time) (bool, []links.RecordLinkMutation, error) {
	descriptor, ok := lookupCollectionDescriptor(fieldKey)
	if !ok {
		return false, nil, &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
	}
	adds, removes, err := recordRefActions(descriptor, payload)
	if err != nil {
		return false, nil, err
	}
	result, err := f.linkStore.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID:         incidentID,
		SourceRecordID:     recordID,
		ActorUserID:        actorID,
		FieldKey:           descriptor.FieldKey,
		LinkType:           links.LinkType(descriptor.LinkType),
		ExpectedTargetType: descriptor.ExpectedTargetType,
		AddRecordIDs:       adds,
		RemoveRecordIDs:    removes,
		Now:                now,
	})
	if err != nil {
		return false, nil, err
	}
	return len(result.RecordLinks) > 0, result.RecordLinks, nil
}

func (f *MutationFacade) appendRecordLinkMutationsTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, startSequence int, mutations []links.RecordLinkMutation) error {
	for index, mutation := range mutations {
		if err := f.revisions.AppendMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    startSequence + index,
			TargetKind:    "record_link",
			TargetID:      mutation.RecordLinkID.String(),
			OperationKind: mutation.Operation,
			BeforeValue:   mutation.BeforeValue,
			AfterValue:    mutation.AfterValue,
		}); err != nil {
			return err
		}
	}
	return nil
}

type collectionDescriptor struct {
	FieldKey           string
	LinkType           string
	ExpectedTargetType string
}

func IsWorkbookRecordRefCollectionField(fieldKey string) bool {
	_, ok := lookupCollectionDescriptor(fieldKey)
	return ok
}

func AllowsWorkbookCollectionOp(fieldKey string, op string) bool {
	if _, ok := lookupCollectionDescriptor(fieldKey); !ok {
		return false
	}
	return op == "add_record_ref" || op == "remove_record_ref"
}

func lookupCollectionDescriptor(fieldKey string) (collectionDescriptor, bool) {
	switch fieldKey {
	case "task.linked_record_ids":
		return collectionDescriptor{FieldKey: fieldKey, LinkType: links.LinkTypeReferencesRecord}, true
	case "decision.affected_record_ids":
		return collectionDescriptor{FieldKey: fieldKey, LinkType: links.LinkTypeReferencesRecord}, true
	case "decision.support_refs":
		return collectionDescriptor{FieldKey: fieldKey, LinkType: links.LinkTypeSupportedBy}, true
	default:
		return collectionDescriptor{}, false
	}
}

func recordRefValidation(incidentID uuid.UUID, descriptor collectionDescriptor, payload WorkbookCollectionActionPayload) (links.RecordRefCollectionValidation, error) {
	adds, removes, err := recordRefActions(descriptor, payload)
	return links.RecordRefCollectionValidation{IncidentID: incidentID, FieldKey: descriptor.FieldKey, LinkType: links.LinkType(descriptor.LinkType), ExpectedTargetType: descriptor.ExpectedTargetType, AddRecordIDs: adds, RemoveRecordIDs: removes}, err
}

func recordRefActions(descriptor collectionDescriptor, payload WorkbookCollectionActionPayload) ([]uuid.UUID, []uuid.UUID, error) {
	adds := make([]uuid.UUID, 0)
	removes := make([]uuid.UUID, 0)
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			if action.LinkedRecordID == nil {
				return nil, nil, &links.CollectionValidationError{Field: descriptor.FieldKey, ReasonCode: "invalid_value"}
			}
			adds = append(adds, *action.LinkedRecordID)
		case "remove_record_ref":
			recordID, err := links.ParseRecordRefItemRef(action.ItemRef)
			if err != nil {
				return nil, nil, &links.CollectionValidationError{Field: descriptor.FieldKey, ReasonCode: "invalid_value"}
			}
			removes = append(removes, recordID)
		default:
			return nil, nil, &links.CollectionValidationError{Field: descriptor.FieldKey, ReasonCode: "invalid_value"}
		}
	}
	return adds, removes, nil
}

func (f *MutationFacade) touchSourceRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID, now time.Time) error {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		return tasksource.TouchTaskRequestTx(ctx, tx, recordID, now)
	case DecisionsViewSchemaID:
		return tasksource.TouchDecisionTx(ctx, tx, recordID, now)
	default:
		return &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func (f *MutationFacade) refreshRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		return f.projectionRows.RefreshTaskRequestTx(ctx, tx, recordID)
	case DecisionsViewSchemaID:
		return f.projectionRows.RefreshDecisionTx(ctx, tx, recordID)
	default:
		return &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func (f *MutationFacade) loadProjectionRowTx(
	ctx context.Context,
	tx pgx.Tx,
	viewSchemaID string,
	recordID uuid.UUID,
) (map[string]any, error) {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		return f.projectionRows.LoadTaskRequestTx(ctx, tx, recordID)
	case DecisionsViewSchemaID:
		return f.projectionRows.LoadDecisionTx(ctx, tx, recordID)
	default:
		return nil, &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
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

type sameFieldConflictParams struct {
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

func overlappingPatchChange(changes []WorkbookPatchChange, changedFields map[string]conflicttokens.PatchChangedField) (WorkbookPatchChange, conflicttokens.PatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return WorkbookPatchChange{}, conflicttokens.PatchChangedField{}, false
}

func buildSameFieldConflict(params sameFieldConflictParams) (map[string]any, error) {
	baseValue, ok := rowCellValue(params.Window.BaseRow, params.Change.FieldKey)
	if !ok {
		return nil, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	serverValue, ok := rowCellValue(params.CurrentRow, params.Change.FieldKey)
	if !ok {
		return nil, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	clientValue, err := patchClientConflictValue(params.Change, baseValue)
	if err != nil {
		return nil, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	conflictClass := params.FieldDescriptors.ConflictResolutionClass(params.Change.FieldKey)
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflictToken, err := conflictToken(params.RouteKey, params.RecordID, params.ViewSchemaID, params.Change.FieldKey, conflictClass, params.BaseRowVersion, params.CurrentRowVersion, params.RequestHash, params.Codec)
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

func patchClientConflictValue(change WorkbookPatchChange, baseValue any) (any, error) {
	if change.Collection == nil {
		return change.CanonicalValue, nil
	}
	return applyCollectionConflictActions(change.FieldKey, baseValue, *change.Collection)
}

func applyCollectionConflictActions(fieldKey string, baseValue any, payload WorkbookCollectionActionPayload) (map[string]any, error) {
	ordered, items, ok := cloneCollectionConflictValue(baseValue)
	if !ok {
		return nil, fmt.Errorf("invalid base collection value for %s", fieldKey)
	}
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			if action.LinkedRecordID == nil {
				return nil, fmt.Errorf("missing linked record")
			}
			items = upsertCollectionConflictItem(items, newRecordRefConflictItem(fieldKey, *action.LinkedRecordID))
		case "remove_record_ref":
			items = removeCollectionConflictItem(items, action.ItemRef)
		default:
			return nil, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	if !ordered {
		slices.SortFunc(items, func(left map[string]any, right map[string]any) int {
			return strings.Compare(collectionSortKey(left), collectionSortKey(right))
		})
	}
	return map[string]any{"kind": "collection_value_v1", "ordered": ordered, "items": items}, nil
}

func cloneCollectionConflictValue(value any) (bool, []map[string]any, bool) {
	object, ok := value.(map[string]any)
	if !ok || object["kind"] != "collection_value_v1" {
		return false, nil, false
	}
	ordered, ok := object["ordered"].(bool)
	if !ok {
		return false, nil, false
	}
	items := make([]map[string]any, 0)
	switch rawItems := object["items"].(type) {
	case []any:
		for _, rawItem := range rawItems {
			item, ok := rawItem.(map[string]any)
			if !ok {
				return false, nil, false
			}
			items = append(items, cloneMap(item))
		}
	case []map[string]any:
		for _, item := range rawItems {
			items = append(items, cloneMap(item))
		}
	default:
		return false, nil, false
	}
	return ordered, items, true
}

func newRecordRefConflictItem(fieldKey string, linkedRecordID uuid.UUID) map[string]any {
	targetType := collectionDisplayTargetType(fieldKey)
	if targetType == "" {
		targetType = "record"
	}
	return map[string]any{
		"item_ref":         links.RecordRefItemRef(linkedRecordID),
		"item_kind":        "record_ref",
		"display_text":     targetType + ":" + linkedRecordID.String(),
		"linked_record_id": linkedRecordID.String(),
	}
}

func collectionDisplayTargetType(fieldKey string) string {
	descriptor, ok := lookupCollectionDescriptor(fieldKey)
	if !ok {
		return ""
	}
	return descriptor.ExpectedTargetType
}

func upsertCollectionConflictItem(items []map[string]any, item map[string]any) []map[string]any {
	ref, _ := item["item_ref"].(string)
	for index, existing := range items {
		if existingRef, _ := existing["item_ref"].(string); existingRef == ref {
			items[index] = item
			return items
		}
	}
	return append(items, item)
}

func removeCollectionConflictItem(items []map[string]any, itemRef string) []map[string]any {
	result := items[:0]
	for _, item := range items {
		if existingRef, _ := item["item_ref"].(string); existingRef != itemRef {
			result = append(result, item)
		}
	}
	return result
}

func cloneMap(input map[string]any) map[string]any {
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func collectionSortKey(item map[string]any) string {
	if text, _ := item["display_text"].(string); text != "" {
		return text
	}
	if ref, _ := item["item_ref"].(string); ref != "" {
		return ref
	}
	return fmt.Sprint(item)
}

func conflictToken(routeKey string, recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, codec conflicttokens.ConflictTokenCodec) (string, error) {
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

func touchesField(changes []WorkbookPatchChange, fieldKey string) bool {
	for _, change := range changes {
		if change.FieldKey == fieldKey {
			return true
		}
	}
	return false
}

func touchesAnyField(changes []WorkbookPatchChange, fieldKeys ...string) bool {
	for _, fieldKey := range fieldKeys {
		if touchesField(changes, fieldKey) {
			return true
		}
	}
	return false
}

func recordTypeMatchesView(recordType string, viewSchemaID string) bool {
	return recordType == recordTypeForView(viewSchemaID)
}

func recordTypeForView(viewSchemaID string) string {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		return "task_request"
	case DecisionsViewSchemaID:
		return "decision"
	default:
		return ""
	}
}
