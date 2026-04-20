package entities

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var ErrInvalidCreateRequest = errors.New("entities: invalid create request")

type Store struct {
	pool            *pgxpool.Pool
	authStore       *authn.Store
	revisionsStore  *revisions.Store
	projectionStore *projections.Store
	linkStore       *links.Store
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:            pool,
		authStore:       authn.NewStore(pool),
		revisionsStore:  revisions.NewStore(),
		projectionStore: projections.NewStore(pool),
		linkStore:       links.NewStore(),
	}
}

func (s *Store) QueryHostRows(ctx context.Context, incidentID uuid.UUID) ([]map[string]any, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("query host rows: store pool is nil")
	}

	rows, err := s.pool.Query(ctx, `
SELECT
    record_id,
    incident_id,
    display_name,
    aad_device_id,
    fqdn,
    hostname,
    host_state,
    merged_into_record_id,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
  FROM hosts
 WHERE incident_id = $1
   AND host_state IN ('stub', 'canonical')
 ORDER BY display_name ASC, record_id ASC
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("query host rows: %w", err)
	}
	defer rows.Close()

	aliasesByRecord, err := loadEntityAliasesByRecord(ctx, s.pool, incidentID, "host")
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, 0)
	for rows.Next() {
		record, err := scanHostRecord(rows)
		if err != nil {
			return nil, err
		}
		record.SuggestionOnlyAliases = aliasesByRecord[record.RecordID]
		result = append(result, BuildHostRow(record))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate host rows: %w", err)
	}
	return result, nil
}

func (s *Store) QueryIdentityRows(ctx context.Context, incidentID uuid.UUID) ([]map[string]any, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("query identity rows: store pool is nil")
	}

	rows, err := s.pool.Query(ctx, `
SELECT
    record_id,
    incident_id,
    display_name,
    aad_object_id,
    sid,
    upn,
    email,
    sam_account_name,
    identity_state,
    merged_into_record_id,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
  FROM identities
 WHERE incident_id = $1
   AND identity_state IN ('stub', 'canonical')
 ORDER BY display_name ASC, record_id ASC
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("query identity rows: %w", err)
	}
	defer rows.Close()

	aliasesByRecord, err := loadEntityAliasesByRecord(ctx, s.pool, incidentID, "identity")
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, 0)
	for rows.Next() {
		record, err := scanIdentityRecord(rows)
		if err != nil {
			return nil, err
		}
		record.SuggestionOnlyAliases = aliasesByRecord[record.RecordID]
		result = append(result, BuildIdentityRow(record))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identity rows: %w", err)
	}
	return result, nil
}

func (s *Store) QueryIndicatorRows(ctx context.Context, incidentID uuid.UUID) ([]map[string]any, error) {
	if s.pool == nil {
		return nil, fmt.Errorf("query indicator rows: store pool is nil")
	}

	rows, err := s.pool.Query(ctx, `
SELECT
    record_id::text,
    incident_id::text,
    row_version,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    dedupe_key,
    defanged_value,
    hash_algorithm,
    hash_value,
    stix_pattern,
    first_observed_at,
    last_observed_at,
    observation_count,
    lifecycle_summary,
    supporting_link_count,
    edited_at
  FROM indicator_grid_projection
 WHERE incident_id = $1
 ORDER BY indicator_type ASC, display_value ASC, record_id ASC
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("query indicator rows: %w", err)
	}
	defer rows.Close()

	result := make([]map[string]any, 0)
	for rows.Next() {
		record, err := scanIndicatorProjectionRecord(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, BuildIndicatorRow(record))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate indicator rows: %w", err)
	}
	return result, nil
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
		beforeVersion := record.RowVersion
		if !reflect.DeepEqual(beforeRow, afterRow) && record.RowVersion > 1 {
			beforeVersion = record.RowVersion - 1
		}
		value := entityVersionID("host", record.RecordID, beforeVersion)
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
		beforeVersion := record.RowVersion
		if !reflect.DeepEqual(beforeRow, afterRow) && record.RowVersion > 1 {
			beforeVersion = record.RowVersion - 1
		}
		value := entityVersionID("identity", record.RecordID, beforeVersion)
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

func loadHostByHostnameTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, hostname string) (HostRecord, error) {
	record, err := scanHostRecord(tx.QueryRow(ctx, `
SELECT
    record_id,
    incident_id,
    display_name,
    aad_device_id,
    fqdn,
    hostname,
    host_state,
    merged_into_record_id,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
  FROM hosts
 WHERE incident_id = $1
   AND hostname = $2
   AND host_state IN ('stub', 'canonical')
 ORDER BY updated_at DESC, record_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, hostname))
	return record, err
}

func insertHostTx(ctx context.Context, tx pgx.Tx, record *HostRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO hosts (
    incident_id,
    display_name,
    aad_device_id,
    fqdn,
    hostname,
    host_state,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, $11, $11)
RETURNING record_id
`, record.IncidentID, record.DisplayName, record.AADDeviceID, record.FQDN, record.Hostname, record.HostState, record.EntityOrigin, record.SeedMentionID, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func updateHostTx(ctx context.Context, tx pgx.Tx, record HostRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE hosts
   SET display_name = $2,
       aad_device_id = $3,
       fqdn = $4,
       hostname = $5,
       host_state = $6,
       merged_into_record_id = $7,
       row_version = $8,
       updated_at = $9,
       updated_by_user_id = $10
 WHERE record_id = $1
`, record.RecordID, record.DisplayName, record.AADDeviceID, record.FQDN, record.Hostname, record.HostState, record.MergedIntoRecordID, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
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
	record, err := scanIdentityRecord(tx.QueryRow(ctx, `
SELECT
    record_id,
    incident_id,
    display_name,
    aad_object_id,
    sid,
    upn,
    email::text,
    sam_account_name,
    identity_state,
    merged_into_record_id,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
  FROM identities
 WHERE incident_id = $1
   AND email = $2
   AND identity_state IN ('stub', 'canonical')
 ORDER BY updated_at DESC, record_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, email))
	return record, err
}

func insertIdentityTx(ctx context.Context, tx pgx.Tx, record *IdentityRecord) error {
	return tx.QueryRow(ctx, `
INSERT INTO identities (
    incident_id,
    display_name,
    aad_object_id,
    sid,
    upn,
    email,
    sam_account_name,
    identity_state,
    entity_origin,
    seed_entity_mention_id,
    row_version,
    created_at,
    updated_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12, $13, $13)
RETURNING record_id
`, record.IncidentID, record.DisplayName, record.AADObjectID, record.SID, record.UPN, record.Email, record.SamAccountName, record.IdentityState, record.EntityOrigin, record.SeedMentionID, record.RowVersion, record.CreatedAt.UTC(), record.CreatedByUser).Scan(&record.RecordID)
}

func updateIdentityTx(ctx context.Context, tx pgx.Tx, record IdentityRecord) error {
	_, err := tx.Exec(ctx, `
UPDATE identities
   SET display_name = $2,
       aad_object_id = $3,
       sid = $4,
       upn = $5,
       email = $6,
       sam_account_name = $7,
       identity_state = $8,
       merged_into_record_id = $9,
       row_version = $10,
       updated_at = $11,
       updated_by_user_id = $12
 WHERE record_id = $1
`, record.RecordID, record.DisplayName, record.AADObjectID, record.SID, record.UPN, record.Email, record.SamAccountName, record.IdentityState, record.MergedIntoRecordID, record.RowVersion, record.UpdatedAt.UTC(), record.UpdatedByUser)
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

type entityAliasQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func loadEntityAliasesByRecord(ctx context.Context, querier entityAliasQueryer, incidentID uuid.UUID, entityType string) (map[uuid.UUID][]string, error) {
	rows, err := querier.Query(ctx, `
SELECT record_id, raw_text
  FROM entity_aliases
 WHERE incident_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY record_id ASC, normalized_text ASC, created_at ASC, entity_alias_id ASC
`, incidentID, entityType)
	if err != nil {
		return nil, fmt.Errorf("query entity aliases by record: %w", err)
	}
	defer rows.Close()

	aliasesByRecord := make(map[uuid.UUID][]string)
	for rows.Next() {
		var (
			recordID uuid.UUID
			rawText  string
		)
		if err := rows.Scan(&recordID, &rawText); err != nil {
			return nil, fmt.Errorf("scan entity alias by record: %w", err)
		}
		aliasesByRecord[recordID] = append(aliasesByRecord[recordID], rawText)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entity aliases by record: %w", err)
	}

	for recordID, aliases := range aliasesByRecord {
		slices.Sort(aliases)
		aliasesByRecord[recordID] = aliases
	}
	return aliasesByRecord, nil
}

func scanIndicatorProjectionRecord(scanner interface {
	Scan(dest ...any) error
}) (IndicatorProjectionRecord, error) {
	var (
		record          IndicatorProjectionRecord
		recordIDRaw     string
		incidentIDRaw   string
		normalizedValue sql.NullString
		defangedValue   sql.NullString
		hashAlgorithm   sql.NullString
		hashValue       sql.NullString
		stixPattern     sql.NullString
		firstObserved   sql.NullTime
		lastObserved    sql.NullTime
		lifecycle       sql.NullString
		editedAt        sql.NullTime
	)
	if err := scanner.Scan(
		&recordIDRaw,
		&incidentIDRaw,
		&record.RowVersion,
		&record.IndicatorType,
		&record.ValueKind,
		&record.DisplayValue,
		&normalizedValue,
		&record.DedupeKey,
		&defangedValue,
		&hashAlgorithm,
		&hashValue,
		&stixPattern,
		&firstObserved,
		&lastObserved,
		&record.ObservationCount,
		&lifecycle,
		&record.SupportingLinkCnt,
		&editedAt,
	); err != nil {
		return IndicatorProjectionRecord{}, fmt.Errorf("scan indicator projection record: %w", err)
	}

	record.RecordID = uuid.MustParse(recordIDRaw)
	record.IncidentID = uuid.MustParse(incidentIDRaw)
	if normalizedValue.Valid {
		record.NormalizedValue = stringPointer(normalizedValue.String)
	}
	if defangedValue.Valid {
		record.DefangedValue = stringPointer(defangedValue.String)
	}
	if hashAlgorithm.Valid {
		record.HashAlgorithm = stringPointer(hashAlgorithm.String)
	}
	if hashValue.Valid {
		record.HashValue = stringPointer(hashValue.String)
	}
	if stixPattern.Valid {
		record.STIXPattern = stringPointer(stixPattern.String)
	}
	if firstObserved.Valid {
		value := firstObserved.Time.UTC()
		record.FirstObservedAt = &value
	}
	if lastObserved.Valid {
		value := lastObserved.Time.UTC()
		record.LastObservedAt = &value
	}
	if lifecycle.Valid {
		record.LifecycleSummary = stringPointer(lifecycle.String)
	}
	if editedAt.Valid {
		record.UpdatedAt = editedAt.Time.UTC()
	}
	return record, nil
}
