package workbookassembly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type evidenceIdempotency struct {
	store   *authn.Store
	records *records.Store
}

type evidenceLifecycleIdempotency struct {
	store *authn.Store
}

func (adapter evidenceLifecycleIdempotency) Get(
	ctx context.Context,
	key evidence.LifecycleIdempotencyKey,
	requestHash []byte,
) (map[string]any, bool, error) {
	record, err := adapter.store.GetRouteIdempotency(ctx, evidenceLifecycleAuthKey(key))
	return decodeEvidenceLifecycleReplay(record, err, requestHash)
}

func (evidenceLifecycleIdempotency) GetTx(
	ctx context.Context,
	tx pgx.Tx,
	key evidence.LifecycleIdempotencyKey,
	requestHash []byte,
) (map[string]any, bool, error) {
	record, err := authn.GetRouteIdempotencyTx(ctx, tx, evidenceLifecycleAuthKey(key))
	return decodeEvidenceLifecycleReplay(record, err, requestHash)
}

func (evidenceLifecycleIdempotency) PutTx(
	ctx context.Context,
	tx pgx.Tx,
	key evidence.LifecycleIdempotencyKey,
	requestHash []byte,
	payload map[string]any,
) error {
	statusCode, ok := evidenceLifecycleStatus(key.OperationID)
	if !ok {
		return fmt.Errorf("unsupported Evidence lifecycle operation %q", key.OperationID)
	}
	err := authn.InsertRouteIdempotencyPayload(ctx, tx, evidenceLifecycleAuthKey(key), nil, requestHash, statusCode, payload)
	if authn.IsUniqueViolation(err) {
		return evidence.ErrClientTxnConflict
	}
	return err
}

func decodeEvidenceLifecycleReplay(
	record authn.RouteIdempotencyRecord,
	err error,
	requestHash []byte,
) (map[string]any, bool, error) {
	if errors.Is(err, authn.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !bytes.Equal(record.RequestHash, requestHash) {
		return nil, false, evidence.ErrClientTxnConflict
	}
	var payload map[string]any
	if err := json.Unmarshal(record.ResponseJSON, &payload); err != nil {
		return nil, false, err
	}
	return payload, true, nil
}

func evidenceLifecycleAuthKey(key evidence.LifecycleIdempotencyKey) authn.RouteIdempotencyKey {
	return authn.RouteIdempotencyKey{
		RouteKey: string(key.OperationID), ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}
}

func evidenceLifecycleStatus(operation evidence.LifecycleOperationID) (int, bool) {
	switch operation {
	case evidence.LifecycleOperationBlobCreate:
		return http.StatusCreated, true
	case evidence.LifecycleOperationBlobAttach:
		return http.StatusOK, true
	default:
		return 0, false
	}
}

func NewEvidenceLifecycleIdempotencyCapability(pool postgres.DB) evidence.LifecycleIdempotencyCapability {
	return evidenceLifecycleIdempotency{store: authn.NewStore(pool)}
}

func (adapter evidenceIdempotency) Get(
	ctx context.Context,
	key evidence.IdempotencyKey,
	requestHash []byte,
) (evidence.StoredMutationResult, bool, error) {
	record, err := adapter.store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
		RouteKey: string(key.OperationID), ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	})
	if errors.Is(err, authn.ErrNotFound) {
		return evidence.StoredMutationResult{}, false, nil
	}
	if err != nil {
		return evidence.StoredMutationResult{}, false, err
	}
	if !bytes.Equal(record.RequestHash, requestHash) {
		return evidence.StoredMutationResult{}, false, evidence.ErrClientTxnConflict
	}
	kind, ok := storedEvidenceKindForOperation(key.OperationID)
	if !ok || record.StatusCode != storedEvidenceStatus(kind) {
		return evidence.StoredMutationResult{}, false, evidence.ErrStoredMutationKindMismatch
	}
	recordID, err := storedEvidenceRecordID(record.ResponseJSON)
	if err != nil {
		return evidence.StoredMutationResult{}, false, fmt.Errorf("decode Evidence stored record identity: %w", err)
	}
	envelope, err := adapter.records.LoadEnvelope(ctx, recordID)
	if err != nil {
		return evidence.StoredMutationResult{}, false, fmt.Errorf("load Evidence stored record envelope: %w", err)
	}
	result, err := decodeStoredEvidenceResult(kind, record.ResponseJSON, envelope.IncidentID)
	if err != nil {
		return evidence.StoredMutationResult{}, false, fmt.Errorf("decode Evidence stored mutation result: %w", err)
	}
	return result, true, nil
}

func (evidenceIdempotency) PutTx(
	ctx context.Context,
	tx pgx.Tx,
	key evidence.IdempotencyKey,
	requestHash []byte,
	result evidence.StoredMutationResult,
) error {
	expected, ok := storedEvidenceKindForOperation(key.OperationID)
	if !ok || result.Kind() != expected {
		return evidence.ErrStoredMutationKindMismatch
	}
	payload, err := encodeStoredEvidenceResult(result)
	if err != nil {
		return err
	}
	err = authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{
		RouteKey: string(key.OperationID), ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}, nil, requestHash, storedEvidenceStatus(result.Kind()), payload)
	if authn.IsUniqueViolation(err) {
		return evidence.ErrClientTxnConflict
	}
	return err
}

func storedEvidenceKindForOperation(operation evidence.OperationID) (evidence.StoredMutationKind, bool) {
	switch operation {
	case evidence.OperationCreate:
		return evidence.StoredMutationCreate, true
	case evidence.OperationPatch, evidence.OperationConflictResolve:
		return evidence.StoredMutationPatch, true
	default:
		return "", false
	}
}

func storedEvidenceStatus(kind evidence.StoredMutationKind) int {
	if kind == evidence.StoredMutationCreate {
		return http.StatusCreated
	}
	return http.StatusOK
}

func decodeStoredEvidenceResult(
	kind evidence.StoredMutationKind,
	data []byte,
	incidentID uuid.UUID,
) (evidence.StoredMutationResult, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return evidence.StoredMutationResult{}, err
	}
	if len(payload) != 3 {
		return evidence.StoredMutationResult{}, evidence.ErrStoredMutationKindMismatch
	}
	viewSchemaID, viewOK := payload["view_schema_id"].(string)
	row, rowOK := payload["row"].(map[string]any)
	if !viewOK || !rowOK || viewSchemaID != evidence.ViewSchemaID {
		return evidence.StoredMutationResult{}, evidence.ErrStoredMutationKindMismatch
	}
	for key := range payload {
		if key != "view_schema_id" && key != "change_set_id" && key != "row" {
			return evidence.StoredMutationResult{}, evidence.ErrStoredMutationKindMismatch
		}
	}
	recordID, err := evidencePayloadUUID(row, "record_id")
	if err != nil {
		return evidence.StoredMutationResult{}, err
	}
	changeSetID, err := evidencePayloadUUID(payload, "change_set_id")
	if err != nil {
		return evidence.StoredMutationResult{}, err
	}
	rowVersion, err := evidencePayloadPositiveInt64(row, "row_version")
	if err != nil {
		return evidence.StoredMutationResult{}, err
	}
	stored := evidence.StoredMutationPayload{
		ViewSchemaID: viewSchemaID, IncidentID: incidentID, RecordID: recordID,
		RowVersion: rowVersion, ChangeSetID: &changeSetID, Row: row,
	}
	if kind == evidence.StoredMutationCreate {
		return evidence.NewStoredCreateResult(stored), nil
	}
	if kind == evidence.StoredMutationPatch {
		return evidence.NewStoredPatchResult(stored), nil
	}
	return evidence.StoredMutationResult{}, evidence.ErrStoredMutationKindMismatch
}

func encodeStoredEvidenceResult(result evidence.StoredMutationResult) (map[string]any, error) {
	stored, ok := result.Payload()
	if !ok || stored.ViewSchemaID != evidence.ViewSchemaID || stored.IncidentID == uuid.Nil ||
		stored.RecordID == uuid.Nil || stored.RowVersion < 1 || stored.ChangeSetID == nil ||
		*stored.ChangeSetID == uuid.Nil || stored.Row == nil {
		return nil, evidence.ErrStoredMutationKindMismatch
	}
	return map[string]any{
		"view_schema_id": stored.ViewSchemaID,
		"change_set_id":  stored.ChangeSetID.String(),
		"row":            stored.Row,
	}, nil
}

func storedEvidenceRecordID(data []byte) (uuid.UUID, error) {
	var payload struct {
		Row map[string]any `json:"row"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return uuid.Nil, err
	}
	return evidencePayloadUUID(payload.Row, "record_id")
}

func evidencePayloadPositiveInt64(payload map[string]any, key string) (int64, error) {
	value, ok := payload[key].(float64)
	if !ok || value < 1 || value > math.MaxInt64 || value != float64(int64(value)) {
		return 0, fmt.Errorf("%s is invalid", key)
	}
	return int64(value), nil
}

func evidencePayloadUUID(payload map[string]any, key string) (uuid.UUID, error) {
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

type evidenceRevisions struct {
	appender *revisions.Appender
	history  conflicttokens.RevisionWindowReader
}

func (adapter evidenceRevisions) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (revisions.RecordSnapshot, error) {
	return adapter.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

func (adapter evidenceRevisions) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params revisions.AppendChangeSetParams) (uuid.UUID, error) {
	return adapter.appender.AppendChangeSetTx(ctx, tx, params)
}

func (adapter evidenceRevisions) AppendRecordMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordMutationParams) error {
	return adapter.appender.AppendRecordMutationTx(ctx, tx, params)
}

func (adapter evidenceRevisions) AppendLiveRevisionTx(ctx context.Context, tx pgx.Tx, input revisions.LiveRevisionInput) error {
	return adapter.appender.AppendLiveRevisionTx(ctx, tx, input)
}

func (adapter evidenceRevisions) LoadRevisionWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, baseVersion int64, currentVersion int64) ([]conflicttokens.RevisionWindowRow, error) {
	return adapter.history.LoadRevisionWindowTx(ctx, tx, recordID, baseVersion, currentVersion)
}

func NewEvidenceMutationDependencies(
	pool postgres.DB,
	appender *revisions.Appender,
	conflictFields conflicttokens.FieldResolver,
	keepSaved conflicttokens.IdempotencyPort,
	projectionRows evidenceprojection.MutationRows,
	associationEffects evidenceprojection.AssociationEffects,
	publications collaboration.RecordChangedAppender,
) (evidence.MutationDependencies, error) {
	if appender == nil {
		return evidence.MutationDependencies{}, fmt.Errorf("compose Evidence mutation dependencies: Revisions appender is required")
	}
	authStore := authn.NewStore(pool)
	recordStore := records.NewStore(pool)
	return evidence.MutationDependencies{
		IncidentState:        admission.NewChecker(pool),
		Idempotency:          evidenceIdempotency{store: authStore, records: recordStore},
		LifecycleIdempotency: NewEvidenceLifecycleIdempotencyCapability(pool),
		RecordEnvelopes:      recordStore,
		Revisions:            evidenceRevisions{appender: appender, history: conflicttokens.NewRevisionWindowReader()},
		ProjectionRows:       projectionRows,
		AssociationEffects:   associationEffects,
		ConflictFields:       conflictFields,
		KeepSavedIdempotency: keepSaved,
		Collaboration:        publications,
	}, nil
}
