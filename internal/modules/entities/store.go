package entities

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var ErrInvalidCreateRequest = errors.New("entities: invalid create request")

type Store struct {
	pool           *pgxpool.Pool
	authStore      *authn.Store
	revisionsStore *revisions.Store
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:           pool,
		authStore:      authn.NewStore(pool),
		revisionsStore: revisions.NewStore(),
	}
}

func (s *Store) CreateHostRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + HostsViewSchemaID
	if existing, err := s.authStore.GetRouteIdempotency(ctx, hostCreateRouteKey, scopeKey, request.ClientTxnID); err == nil {
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

	record, beforeRow, operationKind, statusCode, err := s.upsertHostTx(ctx, tx, actor, incidentID, request, now)
	if err != nil {
		return MutationResult{}, err
	}
	if err := upsertHostProjectionTx(ctx, tx, record); err != nil {
		return MutationResult{}, err
	}

	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
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

	afterRow := BuildHostRow(record)
	var beforeVersionID *string
	if beforeRow != nil {
		value := entityVersionID("host", record.RecordID, record.RowVersion-1)
		beforeVersionID = &value
	}
	afterVersionID := entityVersionID("host", record.RecordID, record.RowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "host",
		TargetID:        record.RecordID.String(),
		OperationKind:   operationKind,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}

	payload := BuildMutationPayload(HostsViewSchemaID, changeSetID, afterRow)
	if err := insertRouteIdempotency(ctx, tx, hostCreateRouteKey, scopeKey, request.ClientTxnID, actor.ID, requestHash, statusCode, payload); err != nil {
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
	scopeKey := incidentID.String() + ":" + IdentitiesViewSchemaID
	if existing, err := s.authStore.GetRouteIdempotency(ctx, identityCreateRouteKey, scopeKey, request.ClientTxnID); err == nil {
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

	record, beforeRow, operationKind, statusCode, err := s.upsertIdentityTx(ctx, tx, actor, incidentID, request, now)
	if err != nil {
		return MutationResult{}, err
	}
	if err := upsertIdentityProjectionTx(ctx, tx, record); err != nil {
		return MutationResult{}, err
	}

	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
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

	afterRow := BuildIdentityRow(record)
	var beforeVersionID *string
	if beforeRow != nil {
		value := entityVersionID("identity", record.RecordID, record.RowVersion-1)
		beforeVersionID = &value
	}
	afterVersionID := entityVersionID("identity", record.RecordID, record.RowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "identity",
		TargetID:        record.RecordID.String(),
		OperationKind:   operationKind,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}

	payload := BuildMutationPayload(IdentitiesViewSchemaID, changeSetID, afterRow)
	if err := insertRouteIdempotency(ctx, tx, identityCreateRouteKey, scopeKey, request.ClientTxnID, actor.ID, requestHash, statusCode, payload); err != nil {
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

func (s *Store) upsertHostTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, now time.Time) (HostRecord, map[string]any, string, int, error) {
	displayName := request.Values["host.display_name"]
	hostname := optionalValue(request.Values, "host.hostname")
	if strings.TrimSpace(displayName) == "" {
		if hostname == nil {
			return HostRecord{}, nil, "", 0, ErrInvalidCreateRequest
		}
		displayName = *hostname
	}

	if hostname != nil {
		current, err := loadHostByHostnameTx(ctx, tx, incidentID, *hostname)
		switch {
		case err == nil:
			beforeRow := BuildHostRow(current)
			next := current
			next.DisplayName = displayName
			next.Hostname = hostname
			next.HostState = "stub"
			next.RowVersion = current.RowVersion + 1
			next.UpdatedAt = now.UTC()
			next.UpdatedByUser = actor.ID
			if err := updateHostTx(ctx, tx, next); err != nil {
				return HostRecord{}, nil, "", 0, err
			}
			return next, beforeRow, "patch", http.StatusOK, nil
		case !errors.Is(err, pgx.ErrNoRows):
			return HostRecord{}, nil, "", 0, fmt.Errorf("lookup host exact match: %w", err)
		}
	}

	record := HostRecord{
		IncidentID:    incidentID,
		DisplayName:   displayName,
		Hostname:      hostname,
		HostState:     "stub",
		RowVersion:    1,
		CreatedAt:     now.UTC(),
		UpdatedAt:     now.UTC(),
		CreatedByUser: actor.ID,
		UpdatedByUser: actor.ID,
	}
	if err := insertHostTx(ctx, tx, &record); err != nil {
		return HostRecord{}, nil, "", 0, err
	}
	return record, nil, "create", http.StatusCreated, nil
}

func (s *Store) upsertIdentityTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, now time.Time) (IdentityRecord, map[string]any, string, int, error) {
	displayName := request.Values["identity.display_name"]
	upn := optionalValue(request.Values, "identity.upn")
	email := optionalValue(request.Values, "identity.email")
	samAccountName := optionalValue(request.Values, "identity.sam_account_name")
	if strings.TrimSpace(displayName) == "" {
		switch {
		case email != nil:
			displayName = *email
		case upn != nil:
			displayName = *upn
		case samAccountName != nil:
			displayName = *samAccountName
		default:
			return IdentityRecord{}, nil, "", 0, ErrInvalidCreateRequest
		}
	}

	if email != nil {
		current, err := loadIdentityByEmailTx(ctx, tx, incidentID, *email)
		switch {
		case err == nil:
			beforeRow := BuildIdentityRow(current)
			next := current
			next.DisplayName = displayName
			next.UPN = upn
			next.Email = email
			next.SamAccountName = samAccountName
			next.IdentityState = "stub"
			next.RowVersion = current.RowVersion + 1
			next.UpdatedAt = now.UTC()
			next.UpdatedByUser = actor.ID
			if err := updateIdentityTx(ctx, tx, next); err != nil {
				return IdentityRecord{}, nil, "", 0, err
			}
			return next, beforeRow, "patch", http.StatusOK, nil
		case !errors.Is(err, pgx.ErrNoRows):
			return IdentityRecord{}, nil, "", 0, fmt.Errorf("lookup identity exact match: %w", err)
		}
	}

	record := IdentityRecord{
		IncidentID:     incidentID,
		DisplayName:    displayName,
		UPN:            upn,
		Email:          email,
		SamAccountName: samAccountName,
		IdentityState:  "stub",
		RowVersion:     1,
		CreatedAt:      now.UTC(),
		UpdatedAt:      now.UTC(),
		CreatedByUser:  actor.ID,
		UpdatedByUser:  actor.ID,
	}
	if err := insertIdentityTx(ctx, tx, &record); err != nil {
		return IdentityRecord{}, nil, "", 0, err
	}
	return record, nil, "create", http.StatusCreated, nil
}

func loadHostByHostnameTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, hostname string) (HostRecord, error) {
	var (
		record      HostRecord
		rawHostname pgtype.Text
	)
	err := tx.QueryRow(ctx, `
SELECT record_id, incident_id, display_name, hostname, host_state, row_version, created_at, updated_at, created_by_user_id, updated_by_user_id
  FROM hosts
 WHERE incident_id = $1
   AND hostname = $2
   AND host_state IN ('stub', 'canonical')
 ORDER BY updated_at DESC, record_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, hostname).Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DisplayName,
		&rawHostname,
		&record.HostState,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
	)
	if rawHostname.Valid {
		value := rawHostname.String
		record.Hostname = &value
	} else {
		record.Hostname = nil
	}
	return record, err
}

func insertHostTx(ctx context.Context, tx pgx.Tx, record *HostRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO hosts (
    incident_id,
    display_name,
    hostname,
    host_state,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $6, $7, $7)
RETURNING record_id
`, record.IncidentID, record.DisplayName, record.Hostname, record.HostState, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func updateHostTx(ctx context.Context, tx pgx.Tx, record HostRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE hosts
   SET display_name = $2,
       hostname = $3,
       host_state = $4,
       row_version = $5,
       updated_at = $6,
       updated_by_user_id = $7
 WHERE record_id = $1
`, record.RecordID, record.DisplayName, record.Hostname, record.HostState, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
	if err != nil {
		return fmt.Errorf("update host: %w", err)
	}
	return nil
}

func upsertHostProjectionTx(ctx context.Context, tx pgx.Tx, record HostRecord) error {
	_, err := tx.Exec(ctx, `
INSERT INTO host_grid_projection (
    record_id,
    incident_id,
    row_version,
    display_name,
    hostname,
    host_state,
    linked_event_count,
    evidence_count,
    location,
    os_platform,
    business_owner,
    criticality,
    containment_status,
    edited_at
)
VALUES ($1, $2, $3, $4, $5, $6, 0, 0, NULL, NULL, NULL, NULL, NULL, $7)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    display_name = EXCLUDED.display_name,
    hostname = EXCLUDED.hostname,
    host_state = EXCLUDED.host_state,
    edited_at = EXCLUDED.edited_at
`, record.RecordID, record.IncidentID, record.RowVersion, record.DisplayName, record.Hostname, record.HostState, record.UpdatedAt.UTC())
	if err != nil {
		return fmt.Errorf("upsert host projection: %w", err)
	}
	return nil
}

func loadIdentityByEmailTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, email string) (IdentityRecord, error) {
	var (
		record            IdentityRecord
		rawUPN            pgtype.Text
		rawEmail          pgtype.Text
		rawSamAccountName pgtype.Text
	)
	err := tx.QueryRow(ctx, `
SELECT record_id, incident_id, display_name, upn, email::text, sam_account_name, identity_state, row_version, created_at, updated_at, created_by_user_id, updated_by_user_id
  FROM identities
 WHERE incident_id = $1
   AND email = $2
   AND identity_state IN ('stub', 'canonical')
 ORDER BY updated_at DESC, record_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, email).Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DisplayName,
		&rawUPN,
		&rawEmail,
		&rawSamAccountName,
		&record.IdentityState,
		&record.RowVersion,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.CreatedByUser,
		&record.UpdatedByUser,
	)
	if rawUPN.Valid {
		value := rawUPN.String
		record.UPN = &value
	} else {
		record.UPN = nil
	}
	if rawEmail.Valid {
		value := rawEmail.String
		record.Email = &value
	} else {
		record.Email = nil
	}
	if rawSamAccountName.Valid {
		value := rawSamAccountName.String
		record.SamAccountName = &value
	} else {
		record.SamAccountName = nil
	}
	return record, err
}

func insertIdentityTx(ctx context.Context, tx pgx.Tx, record *IdentityRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO identities (
    incident_id,
    display_name,
    upn,
    email,
    sam_account_name,
    identity_state,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, $9, $9)
RETURNING record_id
`, record.IncidentID, record.DisplayName, record.UPN, record.Email, record.SamAccountName, record.IdentityState, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func updateIdentityTx(ctx context.Context, tx pgx.Tx, record IdentityRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE identities
   SET display_name = $2,
       upn = $3,
       email = $4,
       sam_account_name = $5,
       identity_state = $6,
       row_version = $7,
       updated_at = $8,
       updated_by_user_id = $9
 WHERE record_id = $1
`, record.RecordID, record.DisplayName, record.UPN, record.Email, record.SamAccountName, record.IdentityState, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
	if err != nil {
		return fmt.Errorf("update identity: %w", err)
	}
	return nil
}

func upsertIdentityProjectionTx(ctx context.Context, tx pgx.Tx, record IdentityRecord) error {
	_, err := tx.Exec(ctx, `
INSERT INTO identity_grid_projection (
    record_id,
    incident_id,
    row_version,
    display_name,
    upn,
    email,
    sam_account_name,
    identity_state,
    linked_event_count,
    evidence_count,
    privilege_level,
    mfa_state,
    reset_status,
    edited_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, 0, NULL, NULL, NULL, $9)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    display_name = EXCLUDED.display_name,
    upn = EXCLUDED.upn,
    email = EXCLUDED.email,
    sam_account_name = EXCLUDED.sam_account_name,
    identity_state = EXCLUDED.identity_state,
    edited_at = EXCLUDED.edited_at
`, record.RecordID, record.IncidentID, record.RowVersion, record.DisplayName, record.UPN, record.Email, record.SamAccountName, record.IdentityState, record.UpdatedAt.UTC())
	if err != nil {
		return fmt.Errorf("upsert identity projection: %w", err)
	}
	return nil
}

func insertRouteIdempotency(ctx context.Context, tx pgx.Tx, routeKey string, scopeKey string, clientTxnID string, actorUserID uuid.UUID, requestHash []byte, statusCode int, payload map[string]any) error {
	responseJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal idempotency payload: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key,
    scope_key,
    client_txn_id,
    actor_user_id,
    target_user_id,
    request_hash,
    status_code,
    response_json
)
VALUES ($1, $2, $3, $4, NULL, $5, $6, $7)
`, routeKey, scopeKey, clientTxnID, actorUserID, requestHash, statusCode, responseJSON); err != nil {
		return fmt.Errorf("insert route idempotency: %w", err)
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
