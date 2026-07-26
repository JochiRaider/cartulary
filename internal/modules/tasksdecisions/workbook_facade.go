package tasksdecisions

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflictmerge"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflictwindows"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/historyquery"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const TaskRequestsViewSchemaID = "cartulary.view.task_requests.v1"

type WorkbookFacade struct {
	pool             postgres.DB
	authStore        *authn.Store
	incidentAccess   incidents.Access
	recordStore      *records.Store
	linkStore        workbookCollectionLinkPort
	projectionRows   *projections.TaskDecisionRows
	revisionHistory  historyquery.Reader
	revisionAppender revisions.Appender
	store            *Store
	conflictTokens   conflicttokens.ConflictTokenCodec
}

type workbookCollectionLinkPort interface {
	ValidateRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionValidation) error
	ApplyRecordRefCollectionTx(context.Context, pgx.Tx, links.RecordRefCollectionCommand) (bool, error)
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
	return "tasksdecisions: row version conflict"
}

type SameFieldConflictError struct {
	Conflict map[string]any
}

func (e *SameFieldConflictError) Error() string {
	return "tasksdecisions: same field conflict"
}

func NewWorkbookFacade(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) *WorkbookFacade {
	return &WorkbookFacade{
		pool:             pool,
		authStore:        authn.NewStore(pool),
		incidentAccess:   incidents.NewAccess(pool),
		recordStore:      records.NewStore(),
		linkStore:        links.NewStore(),
		projectionRows:   newTaskDecisionProjectionRows(pool),
		revisionHistory:  historyquery.NewReader(),
		revisionAppender: revisions.NewAppender(),
		store:            NewStore(),
		conflictTokens:   conflictTokens,
	}
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
		payload, err := decodeSupersedeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return WorkbookMutationResult{}, fmt.Errorf("decode replayed task/decision create payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return WorkbookMutationResult{}, err
		}
		return WorkbookMutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, IncidentID: command.IncidentID, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
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
	if err := validateCreateReferencesTx(ctx, tx, f.linkStore, command.IncidentID, request); err != nil {
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
		CreatedByUserID: command.Actor.ID,
		CreatedAt:       now,
		UpdatedByUserID: command.Actor.ID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	switch request.ViewSchemaID {
	case TaskRequestsViewSchemaID:
		if err := f.store.InsertTaskRequestTx(ctx, tx, recordID, command.IncidentID, command.Actor.ID, TaskCreateParams{Values: request.Values}, now); err != nil {
			return WorkbookMutationResult{}, err
		}
		if value, ok := request.Values[TaskDecisionRecordFieldKey]; ok && value.UUID != nil {
			if _, err := f.store.ApplyTaskDirectChangeTx(ctx, tx, command.IncidentID, recordID, command.Actor.ID, TaskDecisionRecordFieldKey, value, now); err != nil {
				return WorkbookMutationResult{}, err
			}
		}
	case DecisionsViewSchemaID:
		if err := f.store.InsertDecisionTx(ctx, tx, recordID, command.IncidentID, command.Actor.ID, DecisionCreateParams{Values: request.Values}, now); err != nil {
			return WorkbookMutationResult{}, err
		}
	default:
		return WorkbookMutationResult{}, &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
	if err := f.applyCollectionPayloadsTx(ctx, tx, command.IncidentID, recordID, command.Actor.ID, request.Collections, now); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.refreshRowTx(ctx, tx, request.ViewSchemaID, recordID); err != nil {
		return WorkbookMutationResult{}, err
	}
	row, err := f.projectionRows.LoadTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	changeSetID, err := f.revisionAppender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  command.IncidentID,
		ActorUserID: command.Actor.ID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   now,
	})
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	afterVersionID := supersedeVersionID(recordID, 1)
	if err := f.revisionAppender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.revisionAppender.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  1,
		AfterValue:  row,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	payload := buildMutationPayload(request.ViewSchemaID, changeSetID, row)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, command.RequestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return WorkbookMutationResult{}, authn.ErrClientTxnConflict
		}
		return WorkbookMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("commit task/decision create transaction: %w", err)
	}
	return WorkbookMutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       command.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, row),
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
		payload, err := decodeSupersedeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return WorkbookMutationResult{}, fmt.Errorf("decode replayed task/decision patch payload: %w", err)
		}
		return WorkbookMutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: command.RecordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return WorkbookMutationResult{}, fmt.Errorf("query task/decision patch idempotency: %w", err)
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkbookMutationResult{}, fmt.Errorf("begin task/decision patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadSupersedeRecordMetaForUpdateTx(ctx, tx, command.RecordID)
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
		windowRows, err := f.revisionHistory.LoadRevisionWindowTx(ctx, tx, command.RecordID, request.BaseRowVersion, meta.RowVersion)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		fieldDescriptors := taskDecisionConflictFieldDescriptors(request.ViewSchemaID)
		window, err := conflictwindows.BuildPatchConflictWindowWithDescriptors(command.RecordID, request.BaseRowVersion, meta.RowVersion, windowRows, fieldDescriptors)
		if err != nil {
			return WorkbookMutationResult{}, adaptRevisionWindowError(command.RecordID, request.BaseRowVersion, meta.RowVersion, err)
		}
		if change, changed, ok := overlappingPatchChange(request.Changes, window.ChangedFields); ok {
			current, err := f.projectionRows.LoadTx(ctx, tx, request.ViewSchemaID, command.RecordID)
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
	beforeRow, err := f.projectionRows.LoadTx(ctx, tx, request.ViewSchemaID, command.RecordID)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := validatePatchReferencesTx(ctx, tx, f.linkStore, meta.IncidentID, request); err != nil {
		return WorkbookMutationResult{}, err
	}
	changed, err := f.applyPatchTx(ctx, tx, meta.IncidentID, command.RecordID, command.Actor.ID, request, command.Now.UTC())
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
	if err := f.touchSourceRowTx(ctx, tx, request.ViewSchemaID, command.RecordID, command.Now.UTC()); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.refreshRowTx(ctx, tx, request.ViewSchemaID, command.RecordID); err != nil {
		return WorkbookMutationResult{}, err
	}
	afterRow, err := f.projectionRows.LoadTx(ctx, tx, request.ViewSchemaID, command.RecordID)
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
	beforeVersionID := supersedeVersionID(command.RecordID, request.BaseRowVersion)
	if effectiveBeforeVersion != request.BaseRowVersion {
		beforeVersionID = supersedeVersionID(command.RecordID, effectiveBeforeVersion)
	}
	afterVersionID := supersedeVersionID(command.RecordID, rowVersion)
	if err := f.revisionAppender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		TargetID:        command.RecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return WorkbookMutationResult{}, err
	}
	if err := f.revisionAppender.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    command.RecordID,
		RowVersion:  rowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
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
		return WorkbookMutationResult{}, fmt.Errorf("commit task/decision patch transaction: %w", err)
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
		ChangedFieldKeys: changedFieldKeys(beforeRow, afterRow),
	}, nil
}

func validateCreateRequest(request WorkbookCreateRequest) error {
	switch request.ViewSchemaID {
	case TaskRequestsViewSchemaID:
		return ValidateTaskCreateParams(TaskCreateParams{Values: request.Values})
	case DecisionsViewSchemaID:
		return ValidateDecisionCreateParams(DecisionCreateParams{Values: request.Values})
	default:
		return &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func validateCreateReferencesTx(ctx context.Context, tx pgx.Tx, linkStore workbookCollectionLinkPort, incidentID uuid.UUID, request WorkbookCreateRequest) error {
	for fieldKey, value := range request.Values {
		if value.UUID != nil && strings.HasSuffix(fieldKey, "_user_id") {
			if err := validateIncidentMemberUserTx(ctx, tx, incidentID, *value.UUID, fieldKey); err != nil {
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

func validatePatchReferencesTx(ctx context.Context, tx pgx.Tx, linkStore workbookCollectionLinkPort, incidentID uuid.UUID, request WorkbookPatchRequest) error {
	for _, change := range request.Changes {
		if change.Value != nil && change.Value.UUID != nil && strings.HasSuffix(change.FieldKey, "_user_id") {
			if err := validateIncidentMemberUserTx(ctx, tx, incidentID, *change.Value.UUID, change.FieldKey); err != nil {
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
	switch fieldKey {
	case "task.requester_party_id":
		return validateTargetRecordTx(ctx, tx, incidentID, recordID, "party", fieldKey)
	case TaskDecisionRecordFieldKey:
		return validateTargetRecordTx(ctx, tx, incidentID, recordID, "decision", fieldKey)
	default:
		return nil
	}
}

func validateIncidentMemberUserTx(ctx context.Context, tx pgx.Tx, incidentID, userID uuid.UUID, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
  SELECT 1
    FROM users u
    JOIN incident_memberships m ON m.user_id = u.id
   WHERE u.id = $1
     AND u.is_active = true
     AND m.incident_id = $2
)`, userID, incidentID).Scan(&exists); err != nil {
		return fmt.Errorf("validate user: %w", err)
	}
	if !exists {
		return &ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}

func validateTargetRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
	var exists bool
	if expectedType == "" {
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND deleted_at IS NULL
)
`, incidentID, recordID).Scan(&exists); err != nil {
			return fmt.Errorf("validate collection target: %w", err)
		}
	} else if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND record_type = $3
       AND deleted_at IS NULL
)
`, incidentID, recordID, expectedType).Scan(&exists); err != nil {
		return fmt.Errorf("validate collection target: %w", err)
	}
	if !exists {
		return &ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}

func (f *WorkbookFacade) applyPatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, request WorkbookPatchRequest, now time.Time) (bool, error) {
	changed := false
	var beforeTask TaskLifecycleState
	var beforeDecisionStatus string
	var err error
	if request.ViewSchemaID == TaskRequestsViewSchemaID && touchesAnyField(request.Changes, "task.status", "task.blocked_reason", "task.completed_at", "task.owner_user_id") {
		beforeTask, err = f.store.LoadTaskLifecycleStateTx(ctx, tx, recordID)
		if err != nil {
			return false, err
		}
	}
	if request.ViewSchemaID == DecisionsViewSchemaID {
		if err := f.store.ValidateDecisionMachineConsistentTx(ctx, tx, recordID); err != nil {
			return false, err
		}
		if touchesField(request.Changes, "decision.status") {
			beforeDecisionStatus, err = f.store.LoadDecisionStatusTx(ctx, tx, recordID)
			if err != nil {
				return false, err
			}
		}
	}
	for _, change := range request.Changes {
		if change.Value != nil {
			applied, err := f.applyDirectChangeTx(ctx, tx, incidentID, recordID, actorID, request.ViewSchemaID, change, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
			continue
		}
		if change.Collection != nil {
			applied, err := f.applyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, change.FieldKey, *change.Collection, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		}
	}
	if request.ViewSchemaID == TaskRequestsViewSchemaID && touchesAnyField(request.Changes, "task.status", "task.blocked_reason", "task.completed_at", "task.owner_user_id") {
		applied, err := f.store.NormalizeTaskLifecycleTx(ctx, tx, recordID, beforeTask, touchesField(request.Changes, "task.completed_at"), now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	if request.ViewSchemaID == DecisionsViewSchemaID && touchesField(request.Changes, "decision.status") {
		afterDecisionStatus, err := f.store.LoadDecisionStatusTx(ctx, tx, recordID)
		if err != nil {
			return false, err
		}
		if err := ValidateDecisionStatusTransition(beforeDecisionStatus, afterDecisionStatus); err != nil {
			return false, err
		}
		if err := f.store.ValidateDecisionMachineConsistentTx(ctx, tx, recordID); err != nil {
			return false, err
		}
	}
	return changed, nil
}

func (f *WorkbookFacade) applyDirectChangeTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, viewSchemaID string, change WorkbookPatchChange, now time.Time) (bool, error) {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		if err := ValidateTaskDirectPatchChange(change.FieldKey, *change.Value); err != nil {
			return false, err
		}
		return f.store.ApplyTaskDirectChangeTx(ctx, tx, incidentID, recordID, actorID, change.FieldKey, *change.Value, now)
	case DecisionsViewSchemaID:
		if err := ValidateDecisionDirectPatchChange(change.FieldKey, *change.Value); err != nil {
			return false, err
		}
		return f.store.ApplyDecisionDirectChangeTx(ctx, tx, recordID, change.FieldKey, *change.Value, now)
	default:
		return false, &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func (f *WorkbookFacade) applyCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, collections map[string]WorkbookCollectionActionPayload, now time.Time) error {
	for fieldKey, payload := range collections {
		if _, err := f.applyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, fieldKey, payload, now); err != nil {
			return err
		}
	}
	return nil
}

func validateCollectionPayloadTx(ctx context.Context, tx pgx.Tx, linkStore workbookCollectionLinkPort, incidentID uuid.UUID, fieldKey string, payload WorkbookCollectionActionPayload) error {
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

func (f *WorkbookFacade) applyCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, fieldKey string, payload WorkbookCollectionActionPayload, now time.Time) (bool, error) {
	descriptor, ok := lookupCollectionDescriptor(fieldKey)
	if !ok {
		return false, &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
	}
	adds, removes, err := recordRefActions(descriptor, payload)
	if err != nil {
		return false, err
	}
	return f.linkStore.ApplyRecordRefCollectionTx(ctx, tx, links.RecordRefCollectionCommand{
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

func (f *WorkbookFacade) touchSourceRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID, now time.Time) error {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		return f.store.TouchTaskRequestTx(ctx, tx, recordID, now)
	case DecisionsViewSchemaID:
		return f.store.TouchDecisionTx(ctx, tx, recordID, now)
	default:
		return &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
}

func (f *WorkbookFacade) refreshRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		return f.projectionRows.RefreshTaskRequestTx(ctx, tx, recordID)
	case DecisionsViewSchemaID:
		return f.projectionRows.RefreshDecisionTx(ctx, tx, recordID)
	default:
		return &ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
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
	var windowErr *conflictwindows.RevisionWindowError
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
	Window            conflictwindows.PatchConflictWindow
	Change            WorkbookPatchChange
	Changed           conflictwindows.PatchChangedField
	CurrentRow        map[string]any
	FieldDescriptors  conflictwindows.FieldDescriptorSet
	Codec             conflicttokens.ConflictTokenCodec
}

func overlappingPatchChange(changes []WorkbookPatchChange, changedFields map[string]conflictwindows.PatchChangedField) (WorkbookPatchChange, conflictwindows.PatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return WorkbookPatchChange{}, conflictwindows.PatchChangedField{}, false
}

func buildSameFieldConflict(params sameFieldConflictParams) (map[string]any, error) {
	baseValue, ok := rowCellValue(params.Window.BaseRow, params.Change.FieldKey)
	if !ok {
		return nil, &conflictwindows.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	serverValue, ok := rowCellValue(params.CurrentRow, params.Change.FieldKey)
	if !ok {
		return nil, &conflictwindows.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	clientValue, err := patchClientConflictValue(params.Change, baseValue)
	if err != nil {
		return nil, &conflictwindows.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	conflictClass := params.FieldDescriptors.ConflictResolutionClass(params.Change.FieldKey)
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflict := map[string]any{
		"conflict_token":            conflictToken(params.RouteKey, params.RecordID, params.ViewSchemaID, params.Change.FieldKey, conflictClass, params.BaseRowVersion, params.CurrentRowVersion, params.RequestHash, params.Codec),
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
		if suggested, ok := conflictmerge.SuggestedTextMergeValue(baseValue, serverValue, clientValue); ok {
			conflict["suggested_merged_value"] = suggested
		}
	}
	return conflict, nil
}

func taskDecisionConflictFieldDescriptors(viewSchemaID string) conflictwindows.FieldDescriptorSet {
	return conflictwindows.ViewSchemaFieldDescriptors(viewSchemaID)
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

func conflictToken(routeKey string, recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, codec conflicttokens.ConflictTokenCodec) string {
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
