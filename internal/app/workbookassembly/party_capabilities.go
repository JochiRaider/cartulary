package workbookassembly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type partyIdempotency struct {
	store    *authn.Store
	metadata partyReplayMetadataLoader
}

func (a partyIdempotency) Get(
	ctx context.Context,
	key parties.IdempotencyKey,
	requestHash []byte,
) (parties.StoredMutationResult, bool, error) {
	record, err := a.store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	})
	if errors.Is(err, authn.ErrNotFound) {
		return parties.StoredMutationResult{}, false, nil
	}
	if err != nil {
		return parties.StoredMutationResult{}, false, err
	}
	if !bytes.Equal(record.RequestHash, requestHash) {
		return parties.StoredMutationResult{}, false, parties.ErrClientTxnConflict
	}
	kind, ok := partyStoredKindForRoute(key.RouteKey)
	if !ok {
		return parties.StoredMutationResult{}, false, parties.ErrStoredMutationKindMismatch
	}
	stored, err := decodePartyStoredRow(record.ResponseJSON)
	if err != nil {
		return parties.StoredMutationResult{}, false, fmt.Errorf("decode Parties stored mutation result: %w", err)
	}
	metadata, err := a.metadata.Load(ctx, stored.ChangeSetID, stored.RecordID)
	if err != nil {
		return parties.StoredMutationResult{}, false, fmt.Errorf("load Parties stored mutation semantics: %w", err)
	}
	if metadata.ActorUserID != key.ActorUserID || metadata.Source != key.RouteKey ||
		metadata.ClientTxnID != key.ClientTxnID || metadata.IncidentID == uuid.Nil {
		return parties.StoredMutationResult{}, false, parties.ErrStoredMutationKindMismatch
	}
	stored.IncidentID = metadata.IncidentID
	stored.ChangedFieldKeys = append([]string(nil), metadata.ChangedFieldKeys...)
	switch {
	case kind == parties.StoredMutationCreate && record.StatusCode == http.StatusCreated &&
		metadata.OperationKind == "create" && metadata.RowVersion != nil && *metadata.RowVersion == stored.RowVersion:
		stored.Outcome = parties.MutationCreated
		return parties.NewStoredCreateResult(stored), true, nil
	case kind == parties.StoredMutationCreate && record.StatusCode == http.StatusOK &&
		metadata.OperationKind == "reuse" && metadata.RowVersion == nil && len(metadata.ChangedFieldKeys) == 0:
		stored.Outcome = parties.MutationReused
		return parties.NewStoredCreateResult(stored), true, nil
	case kind == parties.StoredMutationPatch && record.StatusCode == http.StatusOK &&
		metadata.OperationKind == "patch" && metadata.RowVersion != nil && *metadata.RowVersion == stored.RowVersion:
		stored.Outcome = parties.MutationUpdated
		return parties.NewStoredPatchResult(stored), true, nil
	default:
		return parties.StoredMutationResult{}, false, parties.ErrStoredMutationKindMismatch
	}
}

func (partyIdempotency) PutTx(
	ctx context.Context,
	tx pgx.Tx,
	key parties.IdempotencyKey,
	requestHash []byte,
	result parties.StoredMutationResult,
) error {
	expectedKind, ok := partyStoredKindForRoute(key.RouteKey)
	if !ok || result.Kind() != expectedKind {
		return parties.ErrStoredMutationKindMismatch
	}
	payload, status, err := encodePartyStoredResult(result)
	if err != nil {
		return err
	}
	err = authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}, nil, requestHash, status, payload)
	if authn.IsUniqueViolation(err) {
		return parties.ErrClientTxnConflict
	}
	return err
}

func partyStoredKindForRoute(routeKey string) (parties.StoredMutationKind, bool) {
	switch routeKey {
	case workbookCreateOperation:
		return parties.StoredMutationCreate, true
	case workbookPatchOperation, workbookConflictResolveOperation:
		return parties.StoredMutationPatch, true
	default:
		return "", false
	}
}

func decodePartyStoredRow(data []byte) (parties.StoredRowMutationResult, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var payload map[string]any
	if err := decoder.Decode(&payload); err != nil {
		return parties.StoredRowMutationResult{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return parties.StoredRowMutationResult{}, fmt.Errorf("stored payload has trailing content")
	}
	viewSchemaID, ok := payload["view_schema_id"].(string)
	if !ok || viewSchemaID != parties.ViewSchemaID {
		return parties.StoredRowMutationResult{}, fmt.Errorf("view_schema_id is invalid")
	}
	changeSetID, err := partyPayloadUUID(payload, "change_set_id")
	if err != nil {
		return parties.StoredRowMutationResult{}, err
	}
	row, ok := payload["row"].(map[string]any)
	if !ok {
		return parties.StoredRowMutationResult{}, fmt.Errorf("row is missing")
	}
	recordID, err := partyPayloadUUID(row, "record_id")
	if err != nil {
		return parties.StoredRowMutationResult{}, err
	}
	rawRowVersion, ok := row["row_version"].(json.Number)
	if !ok {
		return parties.StoredRowMutationResult{}, fmt.Errorf("row_version is missing or is not an integer")
	}
	rowVersion, err := rawRowVersion.Int64()
	if err != nil || rowVersion < 1 {
		return parties.StoredRowMutationResult{}, fmt.Errorf("row_version is invalid")
	}
	row["row_version"] = rowVersion
	return parties.StoredRowMutationResult{
		ViewSchemaID: viewSchemaID, RecordID: recordID, ChangeSetID: changeSetID,
		RowVersion: rowVersion, Row: row,
	}, nil
}

func encodePartyStoredResult(result parties.StoredMutationResult) (map[string]any, int, error) {
	stored, ok := result.RowMutationResult()
	if !ok {
		return nil, 0, parties.ErrStoredMutationKindMismatch
	}
	status := http.StatusOK
	if result.Kind() == parties.StoredMutationCreate && stored.Outcome == parties.MutationCreated {
		status = http.StatusCreated
	}
	return map[string]any{
		"view_schema_id": stored.ViewSchemaID,
		"change_set_id":  stored.ChangeSetID.String(),
		"row":            stored.Row,
	}, status, nil
}

func partyPayloadUUID(payload map[string]any, key string) (uuid.UUID, error) {
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

type partyReplayMetadata struct {
	IncidentID       uuid.UUID
	ActorUserID      uuid.UUID
	Source           string
	ClientTxnID      string
	OperationKind    string
	RowVersion       *int64
	ChangedFieldKeys []string
}

type partyReplayMetadataLoader interface {
	Load(context.Context, uuid.UUID, uuid.UUID) (partyReplayMetadata, error)
}

type postgresPartyReplayMetadata struct{ pool postgres.DB }

func (loader postgresPartyReplayMetadata) Load(
	ctx context.Context,
	changeSetID uuid.UUID,
	recordID uuid.UUID,
) (partyReplayMetadata, error) {
	var metadata partyReplayMetadata
	var clientTxnID *string
	err := loader.pool.QueryRow(ctx, `
SELECT change_set.incident_id,
       change_set.actor_user_id,
       change_set.source,
       change_set.client_txn_id,
       mutation.operation_kind,
       revision.row_version,
       COALESCE(array_agg(fact.field_key ORDER BY fact.field_key)
           FILTER (WHERE fact.field_key IS NOT NULL), '{}'::text[])
  FROM change_sets AS change_set
  JOIN change_set_mutations AS mutation
    ON mutation.change_set_id = change_set.change_set_id
   AND mutation.sequence_no = 1
   AND mutation.target_kind = 'record'
   AND mutation.target_id = $2
  LEFT JOIN record_revisions AS revision
    ON revision.change_set_id = change_set.change_set_id
   AND revision.record_id = $3
  LEFT JOIN record_revision_conflict_facts AS fact
    ON fact.revision_id = revision.revision_id
 WHERE change_set.change_set_id = $1
 GROUP BY change_set.incident_id,
          change_set.actor_user_id,
          change_set.source,
          change_set.client_txn_id,
          mutation.operation_kind,
          revision.row_version
`, changeSetID, recordID.String(), recordID).Scan(
		&metadata.IncidentID,
		&metadata.ActorUserID,
		&metadata.Source,
		&clientTxnID,
		&metadata.OperationKind,
		&metadata.RowVersion,
		&metadata.ChangedFieldKeys,
	)
	if err != nil {
		return partyReplayMetadata{}, err
	}
	if clientTxnID == nil || strings.TrimSpace(*clientTxnID) == "" {
		return partyReplayMetadata{}, fmt.Errorf("change set client transaction is missing")
	}
	metadata.ClientTxnID = *clientTxnID
	return metadata, nil
}

type partyRevisions struct {
	appender *revisions.Appender
	history  conflicttokens.RevisionWindowReader
}

type partyKeepSaved struct {
	idempotency conflicttokens.IdempotencyPort
}

func (a partyKeepSaved) KeepSaved(
	ctx context.Context,
	transactions conflicttokens.TransactionRunner,
	command conflicttokens.Command,
	load conflicttokens.TargetLoader,
) (parties.KeepSavedResult, error) {
	result, err := conflicttokens.KeepSaved(ctx, transactions, a.idempotency, command, load)
	if err != nil {
		return parties.KeepSavedResult{}, err
	}
	row, ok := result.Payload["row"].(map[string]any)
	if !ok {
		return parties.KeepSavedResult{}, fmt.Errorf("translate Party keep-saved result: semantic row is missing")
	}
	return parties.KeepSavedResult{
		Replayed: result.Replayed, IncidentID: result.IncidentID, RecordID: result.RecordID,
		ClientTxnID: result.ClientTxnID, RowVersion: result.RowVersion,
		ViewSchemaID: result.ViewSchemaID, Row: row,
	}, nil
}

func (a partyRevisions) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (revisions.RecordSnapshot, error) {
	return a.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

func (a partyRevisions) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params revisions.AppendChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, params)
}

func (a partyRevisions) AppendRecordMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordMutationParams) error {
	return a.appender.AppendRecordMutationTx(ctx, tx, params)
}

func (a partyRevisions) AppendRecordRevisionAndIntentTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordRevisionParams) error {
	return a.appender.AppendRecordRevisionAndIntentTx(ctx, tx, params)
}

func (a partyRevisions) LoadRevisionWindowTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	baseVersion int64,
	currentVersion int64,
) ([]conflicttokens.RevisionWindowRow, error) {
	return a.history.LoadRevisionWindowTx(ctx, tx, recordID, baseVersion, currentVersion)
}

func newPartyMutationDependencies(
	pool postgres.DB,
	appender *revisions.Appender,
	conflictFields conflicttokens.FieldResolver,
	projectionRows partyprojection.Rows,
) parties.MutationDependencies {
	return parties.MutationDependencies{
		IncidentState: admission.NewChecker(pool),
		Idempotency: partyIdempotency{
			store: authn.NewStore(pool), metadata: postgresPartyReplayMetadata{pool: pool},
		},
		RecordEnvelopes: records.NewStore(pool),
		Projections:     projectionRows,
		Revisions:       partyRevisions{appender: appender, history: conflicttokens.NewRevisionWindowReader()},
		ConflictFields:  conflictFields,
		KeepSaved:       partyKeepSaved{idempotency: NewConflictIdempotencyPort(pool)},
	}
}

func NewPartyMutationContribution(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	conflictFields conflicttokens.FieldResolver,
	projectionRows partyprojection.Rows,
) (*parties.MutationFacade, error) {
	if appender == nil {
		return nil, fmt.Errorf("compose Parties mutation contribution: Revisions appender is required")
	}
	facade, err := parties.NewMutationContribution(
		pool,
		conflictTokens,
		newPartyMutationDependencies(pool, appender, conflictFields, projectionRows),
	)
	if err != nil {
		return nil, fmt.Errorf("compose Parties mutation contribution: %w", err)
	}
	return facade, nil
}
