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
	store *authn.Store
}

func (a partyIdempotency) Get(
	ctx context.Context,
	key parties.IdempotencyKey,
	requestHash []byte,
) (parties.IdempotencyRecord, error) {
	record, err := a.store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	})
	if errors.Is(err, authn.ErrNotFound) {
		return parties.IdempotencyRecord{}, parties.ErrIdempotencyNotFound
	}
	if err != nil {
		return parties.IdempotencyRecord{}, err
	}
	if !bytes.Equal(record.RequestHash, requestHash) {
		return parties.IdempotencyRecord{RequestHash: record.RequestHash}, nil
	}
	kind, ok := partyStoredKindForRoute(key.RouteKey)
	if !ok {
		return parties.IdempotencyRecord{}, parties.ErrStoredMutationKindMismatch
	}
	result, err := decodePartyStoredResult(kind, record.ResponseJSON)
	if err != nil {
		return parties.IdempotencyRecord{}, fmt.Errorf("decode Parties stored mutation result: %w", err)
	}
	return parties.IdempotencyRecord{RequestHash: record.RequestHash, Result: result}, nil
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

func decodePartyStoredResult(
	kind parties.StoredMutationKind,
	data []byte,
) (parties.StoredMutationResult, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return parties.StoredMutationResult{}, err
	}
	viewSchemaID, ok := payload["view_schema_id"].(string)
	if !ok {
		return parties.StoredMutationResult{}, fmt.Errorf("view_schema_id is missing")
	}
	changeSetID, err := partyPayloadUUID(payload, "change_set_id")
	if err != nil {
		return parties.StoredMutationResult{}, err
	}
	row, ok := payload["row"].(map[string]any)
	if !ok {
		return parties.StoredMutationResult{}, fmt.Errorf("row is missing")
	}
	recordID, err := partyPayloadUUID(row, "record_id")
	if err != nil {
		return parties.StoredMutationResult{}, err
	}
	stored := parties.StoredRowMutationResult{
		ViewSchemaID: viewSchemaID, RecordID: recordID, ChangeSetID: changeSetID, Row: row,
	}
	switch kind {
	case parties.StoredMutationCreate:
		return parties.NewStoredCreateResult(stored), nil
	case parties.StoredMutationPatch:
		return parties.NewStoredPatchResult(stored), nil
	default:
		return parties.StoredMutationResult{}, parties.ErrStoredMutationKindMismatch
	}
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
		IncidentState:   admission.NewChecker(pool),
		Idempotency:     partyIdempotency{store: authn.NewStore(pool)},
		RecordEnvelopes: records.NewStore(),
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
