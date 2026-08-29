package hostidentity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *Store) CreateHostRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + entitycontract.HostsViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    hostCreateRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed host create payload: %w", err)
		}
		recordID, err := extractUUIDFromPayload(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query host create idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin host create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.incidentAccess.RequireOpenTx(ctx, tx, incidentID); err != nil {
		return MutationResult{}, err
	}
	record, beforeRow, operationKind, statusCode, beforeSnapshot, err := s.upsertHostTx(ctx, tx, actor, incidentID, request, now)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.ports.projections.RefreshHostTx(ctx, tx, record.RecordID); err != nil {
		return MutationResult{}, err
	}
	afterSnapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return MutationResult{}, err
	}

	changeSetID, err := s.ports.revisions.AppendChangeSetTx(ctx, tx, entityChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      hostCreateRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}

	afterRow := buildHostRow(record)
	var beforeVersionID *string
	if beforeRow != nil {
		beforeVersion := record.RowVersion
		if !reflect.DeepEqual(beforeRow, afterRow) && record.RowVersion > 1 {
			beforeVersion = record.RowVersion - 1
		}
		value := entityVersionID("host", record.RecordID, beforeVersion)
		beforeVersionID = &value
	}
	afterVersionID := entityVersionID("host", record.RecordID, record.RowVersion)
	if err := s.ports.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "host",
		RecordID:        record.RecordID,
		OperationKind:   operationKind,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.appendAliasCreateMutationsTx(ctx, tx, changeSetID, 2, record.AliasMutations); err != nil {
		return MutationResult{}, err
	}
	if beforeRow == nil || !reflect.DeepEqual(beforeRow, afterRow) {
		changedFields := entityChangedFieldKeys(beforeRow, afterRow)
		if err := s.ports.revisions.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
			ChangeSetID:    changeSetID,
			RecordID:       record.RecordID,
			RowVersion:     record.RowVersion,
			BeforeSnapshot: beforeSnapshot,
			AfterSnapshot:  &afterSnapshot,
			ConflictFacts:  entityRevisionFacts(beforeRow, afterRow, changedFields),
		}); err != nil {
			return MutationResult{}, err
		}
		if err := s.appendRecordChangedTx(ctx, tx, incidentID, actor.ID, request.ClientTxnID, changeSetID, record.RecordID, record.RowVersion, 0, now, entitycontract.HostsViewSchemaID, afterRow, changedFields); err != nil {
			return MutationResult{}, err
		}
	}

	payload := buildMutationPayload(entitycontract.HostsViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, statusCode, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit host create transaction: %w", err)
	}

	return MutationResult{
		Payload:     payload,
		StatusCode:  statusCode,
		RecordID:    record.RecordID,
		ChangeSetID: changeSetID,
		RowVersion:  record.RowVersion,
	}, nil
}

func (s *Store) CreateIdentityRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + entitycontract.IdentitiesViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    identityCreateRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed identity create payload: %w", err)
		}
		recordID, err := extractUUIDFromPayload(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query identity create idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin identity create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.incidentAccess.RequireOpenTx(ctx, tx, incidentID); err != nil {
		return MutationResult{}, err
	}
	record, beforeRow, operationKind, statusCode, beforeSnapshot, err := s.upsertIdentityTx(ctx, tx, actor, incidentID, request, now)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.ports.projections.RefreshIdentityTx(ctx, tx, record.RecordID); err != nil {
		return MutationResult{}, err
	}
	afterSnapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return MutationResult{}, err
	}

	changeSetID, err := s.ports.revisions.AppendChangeSetTx(ctx, tx, entityChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      identityCreateRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}

	afterRow := buildIdentityRow(record)
	var beforeVersionID *string
	if beforeRow != nil {
		beforeVersion := record.RowVersion
		if !reflect.DeepEqual(beforeRow, afterRow) && record.RowVersion > 1 {
			beforeVersion = record.RowVersion - 1
		}
		value := entityVersionID("identity", record.RecordID, beforeVersion)
		beforeVersionID = &value
	}
	afterVersionID := entityVersionID("identity", record.RecordID, record.RowVersion)
	if err := s.ports.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "identity",
		RecordID:        record.RecordID,
		OperationKind:   operationKind,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.appendAliasCreateMutationsTx(ctx, tx, changeSetID, 2, record.AliasMutations); err != nil {
		return MutationResult{}, err
	}
	if beforeRow == nil || !reflect.DeepEqual(beforeRow, afterRow) {
		changedFields := entityChangedFieldKeys(beforeRow, afterRow)
		if err := s.ports.revisions.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
			ChangeSetID:    changeSetID,
			RecordID:       record.RecordID,
			RowVersion:     record.RowVersion,
			BeforeSnapshot: beforeSnapshot,
			AfterSnapshot:  &afterSnapshot,
			ConflictFacts:  entityRevisionFacts(beforeRow, afterRow, changedFields),
		}); err != nil {
			return MutationResult{}, err
		}
		if err := s.appendRecordChangedTx(ctx, tx, incidentID, actor.ID, request.ClientTxnID, changeSetID, record.RecordID, record.RowVersion, 0, now, entitycontract.IdentitiesViewSchemaID, afterRow, changedFields); err != nil {
			return MutationResult{}, err
		}
	}

	payload := buildMutationPayload(entitycontract.IdentitiesViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, statusCode, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit identity create transaction: %w", err)
	}

	return MutationResult{
		Payload:     payload,
		StatusCode:  statusCode,
		RecordID:    record.RecordID,
		ChangeSetID: changeSetID,
		RowVersion:  record.RowVersion,
	}, nil
}

func (s *Store) appendAliasCreateMutationsTx(ctx context.Context, tx pgx.Tx, changeSetID uuid.UUID, startSequence int, aliases []AliasMutationValue) error {
	for index, alias := range aliases {
		if err := s.ports.revisions.AppendMutationTx(ctx, tx, entityMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    startSequence + index,
			TargetKind:    "entity_alias",
			TargetID:      "entity_alias:" + alias.EntityAliasID.String(),
			OperationKind: "create",
			AfterValue:    alias.MutationValue(),
		}); err != nil {
			return err
		}
	}
	return nil
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractUUIDFromPayload(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok || text == "" {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsed, nil
}

func entityVersionID(prefix string, recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("%s:%s:%d", prefix, recordID.String(), rowVersion)
}

func optionalValue(values map[string]string, key string) *string {
	value, ok := values[key]
	if !ok {
		return nil
	}
	cloned := value
	return &cloned
}
