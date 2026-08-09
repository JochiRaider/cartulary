package workbookassembly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type taskDecisionIdempotency struct {
	store *authn.Store
}

func (a taskDecisionIdempotency) Get(
	ctx context.Context,
	key tasksdecisions.IdempotencyKey,
	requestHash []byte,
) (tasksdecisions.IdempotencyRecord, error) {
	record, err := a.store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	})
	if errors.Is(err, authn.ErrNotFound) {
		return tasksdecisions.IdempotencyRecord{}, tasksdecisions.ErrIdempotencyNotFound
	}
	if err != nil {
		return tasksdecisions.IdempotencyRecord{}, err
	}
	if !bytes.Equal(record.RequestHash, requestHash) {
		return tasksdecisions.IdempotencyRecord{RequestHash: record.RequestHash}, nil
	}
	kind, ok := taskDecisionStoredKindForRoute(key.RouteKey)
	if !ok {
		return tasksdecisions.IdempotencyRecord{}, tasksdecisions.ErrStoredMutationKindMismatch
	}
	result, err := decodeTaskDecisionStoredResult(kind, record.ResponseJSON)
	if err != nil {
		return tasksdecisions.IdempotencyRecord{}, fmt.Errorf("decode Tasks/Decisions stored mutation result: %w", err)
	}
	return tasksdecisions.IdempotencyRecord{
		RequestHash: record.RequestHash,
		Result:      result,
	}, nil
}

func (taskDecisionIdempotency) PutTx(
	ctx context.Context,
	tx pgx.Tx,
	key tasksdecisions.IdempotencyKey,
	requestHash []byte,
	result tasksdecisions.StoredMutationResult,
) error {
	expectedKind, ok := taskDecisionStoredKindForRoute(key.RouteKey)
	if !ok || result.Kind() != expectedKind {
		return tasksdecisions.ErrStoredMutationKindMismatch
	}
	payload, err := encodeTaskDecisionStoredResult(result)
	if err != nil {
		return err
	}
	status := http.StatusOK
	if result.Kind() == tasksdecisions.StoredMutationCreate {
		status = http.StatusCreated
	}
	err = authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}, nil, requestHash, status, payload)
	if authn.IsUniqueViolation(err) {
		return tasksdecisions.ErrClientTxnConflict
	}
	return err
}

func taskDecisionStoredKindForRoute(routeKey string) (tasksdecisions.StoredMutationKind, bool) {
	switch routeKey {
	case "workbook.rows.create":
		return tasksdecisions.StoredMutationCreate, true
	case "workbook.records.patch", "workbook.records.conflicts.resolve":
		return tasksdecisions.StoredMutationPatch, true
	case "workbook.records.supersede":
		return tasksdecisions.StoredMutationDecisionSupersession, true
	default:
		return "", false
	}
}

func decodeTaskDecisionStoredResult(
	kind tasksdecisions.StoredMutationKind,
	data []byte,
) (tasksdecisions.StoredMutationResult, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return tasksdecisions.StoredMutationResult{}, err
	}
	viewSchemaID, ok := payload["view_schema_id"].(string)
	if !ok {
		return tasksdecisions.StoredMutationResult{}, fmt.Errorf("view_schema_id is missing")
	}
	changeSetID, err := taskDecisionPayloadUUID(payload, "change_set_id")
	if err != nil {
		return tasksdecisions.StoredMutationResult{}, err
	}
	switch kind {
	case tasksdecisions.StoredMutationCreate, tasksdecisions.StoredMutationPatch:
		row, ok := payload["row"].(map[string]any)
		if !ok {
			return tasksdecisions.StoredMutationResult{}, fmt.Errorf("row is missing")
		}
		recordID, err := taskDecisionPayloadUUID(row, "record_id")
		if err != nil {
			return tasksdecisions.StoredMutationResult{}, err
		}
		stored := tasksdecisions.StoredWorkbookResult{
			ViewSchemaID: viewSchemaID,
			RecordID:     recordID,
			ChangeSetID:  changeSetID,
			Row:          row,
		}
		if kind == tasksdecisions.StoredMutationCreate {
			return tasksdecisions.NewStoredCreateResult(stored), nil
		}
		return tasksdecisions.NewStoredPatchResult(stored), nil
	case tasksdecisions.StoredMutationDecisionSupersession:
		targetRecordID, err := taskDecisionPayloadUUID(payload, "target_record_id")
		if err != nil {
			return tasksdecisions.StoredMutationResult{}, err
		}
		supersedingRecordID, err := taskDecisionPayloadUUID(payload, "superseding_record_id")
		if err != nil {
			return tasksdecisions.StoredMutationResult{}, err
		}
		targetVersion, err := taskDecisionPayloadInt64(payload, "target_row_version")
		if err != nil {
			return tasksdecisions.StoredMutationResult{}, err
		}
		supersedingVersion, err := taskDecisionPayloadInt64(payload, "superseding_row_version")
		if err != nil {
			return tasksdecisions.StoredMutationResult{}, err
		}
		targetStatus, statusOK := payload["target_status"].(string)
		reason, reasonOK := payload["reason"].(string)
		if !statusOK || !reasonOK {
			return tasksdecisions.StoredMutationResult{}, fmt.Errorf("supersession facts are incomplete")
		}
		return tasksdecisions.NewStoredDecisionSupersessionResult(tasksdecisions.StoredDecisionSupersessionResult{
			ViewSchemaID: viewSchemaID,
			ChangeSetID:  changeSetID,
			Facts: tasksdecisions.SupersedeFacts{
				TargetRecordID: targetRecordID, SupersedingRecordID: supersedingRecordID,
				TargetRowVersion: targetVersion, SupersedingRowVersion: supersedingVersion,
				TargetStatus: targetStatus, Reason: reason,
			},
		}), nil
	default:
		return tasksdecisions.StoredMutationResult{}, tasksdecisions.ErrStoredMutationKindMismatch
	}
}

func encodeTaskDecisionStoredResult(result tasksdecisions.StoredMutationResult) (map[string]any, error) {
	switch result.Kind() {
	case tasksdecisions.StoredMutationCreate, tasksdecisions.StoredMutationPatch:
		stored, ok := result.WorkbookResult()
		if !ok {
			return nil, tasksdecisions.ErrStoredMutationKindMismatch
		}
		return map[string]any{
			"view_schema_id": stored.ViewSchemaID,
			"change_set_id":  stored.ChangeSetID.String(),
			"row":            stored.Row,
		}, nil
	case tasksdecisions.StoredMutationDecisionSupersession:
		stored, ok := result.DecisionSupersessionResult()
		if !ok {
			return nil, tasksdecisions.ErrStoredMutationKindMismatch
		}
		return map[string]any{
			"view_schema_id":          stored.ViewSchemaID,
			"change_set_id":           stored.ChangeSetID.String(),
			"target_record_id":        stored.Facts.TargetRecordID.String(),
			"superseding_record_id":   stored.Facts.SupersedingRecordID.String(),
			"target_row_version":      stored.Facts.TargetRowVersion,
			"superseding_row_version": stored.Facts.SupersedingRowVersion,
			"target_status":           stored.Facts.TargetStatus,
			"reason":                  stored.Facts.Reason,
		}, nil
	default:
		return nil, tasksdecisions.ErrStoredMutationKindMismatch
	}
}

func taskDecisionPayloadUUID(payload map[string]any, key string) (uuid.UUID, error) {
	value, ok := payload[key].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("%s is missing", key)
	}
	result, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s is invalid: %w", key, err)
	}
	return result, nil
}

func taskDecisionPayloadInt64(payload map[string]any, key string) (int64, error) {
	switch value := payload[key].(type) {
	case float64:
		return int64(value), nil
	case int64:
		return value, nil
	case int:
		return int64(value), nil
	default:
		return 0, fmt.Errorf("%s is missing", key)
	}
}

type taskDecisionRevisions struct {
	appender *revisions.Appender
	history  conflicttokens.RevisionWindowReader
}

func (a taskDecisionRevisions) AppendChangeSetTx(
	ctx context.Context,
	tx pgx.Tx,
	params revisions.AppendChangeSetParams,
) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, params)
}

func (a taskDecisionRevisions) AppendMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	params revisions.AppendMutationParams,
) error {
	return a.appender.AppendMutationTx(ctx, tx, params)
}

func (a taskDecisionRevisions) AppendRecordRevisionTx(
	ctx context.Context,
	tx pgx.Tx,
	params revisions.AppendRecordRevisionParams,
) error {
	return a.appender.AppendRecordRevisionTx(ctx, tx, params)
}

func (a taskDecisionRevisions) LoadRevisionWindowTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	baseVersion int64,
	currentVersion int64,
) ([]conflicttokens.RevisionWindowRow, error) {
	return a.history.LoadRevisionWindowTx(ctx, tx, recordID, baseVersion, currentVersion)
}

func newTaskDecisionMutationDependencies(
	pool postgres.DB,
	appender *revisions.Appender,
	conflictFields conflicttokens.FieldResolver,
	projectionRows taskdecisionprojection.Rows,
) tasksdecisions.MutationDependencies {
	authStore := authn.NewStore(pool)
	return tasksdecisions.MutationDependencies{
		IncidentState:        incidents.NewAccess(pool),
		MemberReferences:     tasksdecisions.NewMemberReferenceCapability(),
		Idempotency:          taskDecisionIdempotency{store: authStore},
		RecordEnvelopes:      records.NewStore(),
		Links:                links.NewStore(),
		Projections:          projectionRows,
		Revisions:            taskDecisionRevisions{appender: appender, history: conflicttokens.NewRevisionWindowReader()},
		ConflictFields:       conflictFields,
		KeepSavedIdempotency: NewConflictIdempotencyPort(pool),
	}
}

func NewTaskDecisionMutationContribution(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	conflictFields conflicttokens.FieldResolver,
	projectionRows taskdecisionprojection.Rows,
) (*tasksdecisions.MutationFacade, error) {
	if appender == nil {
		return nil, fmt.Errorf("compose Tasks/Decisions mutation contribution: Revisions appender is required")
	}
	facade, err := tasksdecisions.NewMutationContribution(
		pool,
		conflictTokens,
		newTaskDecisionMutationDependencies(pool, appender, conflictFields, projectionRows),
	)
	if err != nil {
		return nil, fmt.Errorf("compose Tasks/Decisions mutation contribution: %w", err)
	}
	return facade, nil
}
