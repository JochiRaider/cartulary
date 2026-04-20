package incidents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var (
	ErrIncidentNotFound          = errors.New("incidents: incident not found")
	ErrIncidentKeyConflict       = errors.New("incidents: incident key conflict")
	ErrIncidentVersionConflict   = errors.New("incidents: incident version conflict")
	ErrMembershipNotFound        = errors.New("incidents: membership not found")
	ErrMembershipExistsUsePatch  = errors.New("incidents: membership exists use patch")
	ErrMembershipVersionConflict = errors.New("incidents: membership version conflict")
	ErrLastIncidentAdmin         = errors.New("incidents: last incident admin")
)

type Store struct {
	pool      *pgxpool.Pool
	authStore *authn.Store
	hooks     StoreHooks
}

type IncidentRecord struct {
	ID                     uuid.UUID
	IncidentKey            string
	Title                  string
	Description            *string
	Status                 string
	Severity               *string
	TLP                    *string
	CurrentPhase           *string
	PrimaryExternalCaseRef *string
	CreatedByUserID        uuid.UUID
	CreatedAt              time.Time
	UpdatedAt              time.Time
	UpdatedByUserID        *uuid.UUID
	IncidentVersion        int64
	ClosedAt               *time.Time
}

type MembershipRecord struct {
	IncidentID        uuid.UUID
	UserID            uuid.UUID
	DisplayName       string
	Role              string
	JoinedAt          time.Time
	AddedByUserID     uuid.UUID
	UpdatedAt         time.Time
	UpdatedByUserID   *uuid.UUID
	MembershipVersion int64
}

type IncidentWorkbookPreferencesRecord struct {
	IncidentID      uuid.UUID
	DefaultSheetRef []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
	UpdatedByUserID *uuid.UUID
}

type UserWorkbookPreferencesRecord struct {
	IncidentID   uuid.UUID
	UserID       uuid.UUID
	HomeSheetRef []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateIncidentResult struct {
	Incident   IncidentRecord
	Payload    map[string]any
	StatusCode int
	Location   string
}

type MembershipCreateResult struct {
	Membership MembershipRecord
	Payload    map[string]any
	StatusCode int
}

func NewStore(pool *pgxpool.Pool) *Store {
	return NewStoreWithHooks(pool, currentStoreHooks())
}

func NewStoreWithHooks(pool *pgxpool.Pool, hooks StoreHooks) *Store {
	return &Store{
		pool:      pool,
		authStore: authn.NewStore(pool),
		hooks:     hooks,
	}
}

func (s *Store) ListVisibleIncidents(ctx context.Context, userID uuid.UUID, snapshotAt time.Time, afterUpdatedAt *time.Time, afterIncidentID *uuid.UUID, limit int) ([]IncidentRecord, *incidentCursor, error) {
	query := `
SELECT i.id, i.incident_key, i.title, i.description, i.status, i.severity, i.tlp, i.current_phase,
       i.primary_external_case_ref, i.created_by_user_id, i.created_at, i.updated_at, i.updated_by_user_id,
       i.incident_version, i.closed_at
  FROM incidents i
  JOIN incident_memberships m ON m.incident_id = i.id
 WHERE m.user_id = $1
   AND i.updated_at <= $2
`
	args := []any{userID, snapshotAt.UTC()}
	if afterUpdatedAt != nil && afterIncidentID != nil {
		query += `
   AND (i.updated_at < $3 OR (i.updated_at = $3 AND i.id > $4))
`
		args = append(args, afterUpdatedAt.UTC(), *afterIncidentID)
	}
	query += fmt.Sprintf(" ORDER BY i.updated_at DESC, i.id ASC LIMIT $%d", len(args)+1)
	args = append(args, limit+1)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list visible incidents: %w", err)
	}
	defer rows.Close()

	records := make([]IncidentRecord, 0, limit+1)
	for rows.Next() {
		record, err := scanIncident(rows)
		if err != nil {
			return nil, nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate visible incidents: %w", err)
	}

	var next *incidentCursor
	if len(records) > limit {
		last := records[limit-1]
		next = &incidentCursor{
			Route:       "incidents.list",
			ActorUserID: userID.String(),
			Limit:       limit,
			SnapshotAt:  snapshotAt.UTC(),
			UpdatedAt:   last.UpdatedAt,
			IncidentID:  last.ID.String(),
		}
		records = records[:limit]
	}
	return records, next, nil
}

func (s *Store) GetVisibleIncident(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (IncidentRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT i.id, i.incident_key, i.title, i.description, i.status, i.severity, i.tlp, i.current_phase,
       i.primary_external_case_ref, i.created_by_user_id, i.created_at, i.updated_at, i.updated_by_user_id,
       i.incident_version, i.closed_at
  FROM incidents i
  JOIN incident_memberships m ON m.incident_id = i.id
 WHERE i.id = $1
   AND m.user_id = $2
`, incidentID, userID)
	record, err := scanIncident(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return IncidentRecord{}, ErrIncidentNotFound
	}
	return record, err
}

func (s *Store) GetIncidentMembershipForUser(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (MembershipRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT m.incident_id, m.user_id, u.display_name, m.role, m.joined_at, m.added_by_user_id,
       m.updated_at, m.updated_by_user_id, m.membership_version
  FROM incident_memberships m
  JOIN users u ON u.id = m.user_id
 WHERE m.incident_id = $1
   AND m.user_id = $2
`, incidentID, userID)
	record, err := scanMembership(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return MembershipRecord{}, ErrMembershipNotFound
	}
	return record, err
}

func (s *Store) ListMemberships(ctx context.Context, incidentID uuid.UUID, snapshotAt time.Time, afterJoinedAt *time.Time, afterUserID *uuid.UUID, limit int) ([]MembershipRecord, *membershipCursor, error) {
	query := `
SELECT m.incident_id, m.user_id, u.display_name, m.role, m.joined_at, m.added_by_user_id,
       m.updated_at, m.updated_by_user_id, m.membership_version
  FROM incident_memberships m
  JOIN users u ON u.id = m.user_id
 WHERE m.incident_id = $1
   AND m.joined_at <= $2
`
	args := []any{incidentID, snapshotAt.UTC()}
	if afterJoinedAt != nil && afterUserID != nil {
		query += `
   AND (m.joined_at > $3 OR (m.joined_at = $3 AND m.user_id > $4))
`
		args = append(args, afterJoinedAt.UTC(), *afterUserID)
	}
	query += fmt.Sprintf(" ORDER BY m.joined_at ASC, m.user_id ASC LIMIT $%d", len(args)+1)
	args = append(args, limit+1)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list memberships: %w", err)
	}
	defer rows.Close()

	records := make([]MembershipRecord, 0, limit+1)
	for rows.Next() {
		record, err := scanMembership(rows)
		if err != nil {
			return nil, nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate memberships: %w", err)
	}

	var next *membershipCursor
	if len(records) > limit {
		last := records[limit-1]
		next = &membershipCursor{
			Route:       "incident.memberships.list",
			ActorUserID: "",
			IncidentID:  incidentID.String(),
			Limit:       limit,
			SnapshotAt:  snapshotAt.UTC(),
			JoinedAt:    last.JoinedAt,
			UserID:      last.UserID.String(),
		}
		records = records[:limit]
	}
	return records, next, nil
}

func (s *Store) GetIncidentWorkbookPreferences(ctx context.Context, incidentID uuid.UUID) (IncidentWorkbookPreferencesRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT incident_id, default_sheet_ref, created_at, updated_at, updated_by_user_id
  FROM incident_workbook_preferences
 WHERE incident_id = $1
`, incidentID)
	var record IncidentWorkbookPreferencesRecord
	if err := row.Scan(&record.IncidentID, &record.DefaultSheetRef, &record.CreatedAt, &record.UpdatedAt, &record.UpdatedByUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return IncidentWorkbookPreferencesRecord{}, ErrIncidentNotFound
		}
		return IncidentWorkbookPreferencesRecord{}, err
	}
	return record, nil
}

func (s *Store) GetUserWorkbookPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (UserWorkbookPreferencesRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT incident_id, user_id, home_sheet_ref, created_at, updated_at
  FROM user_workbook_preferences
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, userID)
	var record UserWorkbookPreferencesRecord
	if err := row.Scan(&record.IncidentID, &record.UserID, &record.HomeSheetRef, &record.CreatedAt, &record.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserWorkbookPreferencesRecord{}, ErrIncidentNotFound
		}
		return UserWorkbookPreferencesRecord{}, err
	}
	return record, nil
}

func (s *Store) CreateIncident(ctx context.Context, actor authn.UserRecord, request CreateIncidentRequest, requestHash []byte, requestID string, now time.Time) (CreateIncidentResult, error) {
	scopeKey := IncidentCreateIdempotencyScope(actor.ID)
	if existing, err := s.authStore.GetRouteIdempotency(ctx, "incidents.create", scopeKey, request.ClientTxnID); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return CreateIncidentResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return CreateIncidentResult{}, fmt.Errorf("decode replayed incident payload: %w", err)
		}
		incidentID, err := extractUUID(payload["incident_id"])
		if err != nil {
			return CreateIncidentResult{}, fmt.Errorf("decode replayed incident id: %w", err)
		}
		return CreateIncidentResult{
			Incident:   IncidentRecord{ID: incidentID},
			Payload:    payload,
			StatusCode: http.StatusOK,
			Location:   incidentLocation(incidentID),
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return CreateIncidentResult{}, fmt.Errorf("query incident create idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CreateIncidentResult{}, fmt.Errorf("begin incident create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	bootstrap := DefaultIncidentCreateBootstrap()
	var incident IncidentRecord
	if err := tx.QueryRow(ctx, `
INSERT INTO incidents (
    incident_key, incident_key_canonical, title, description, status, severity, tlp, current_phase,
    primary_external_case_ref, created_by_user_id, created_at, updated_at, updated_by_user_id, incident_version
)
VALUES ($1, $2, $3, $4, 'active', $5, $6, $7, $8, $9, $10, $10, $9, 1)
RETURNING id, incident_key, title, description, status, severity, tlp, current_phase, primary_external_case_ref,
          created_by_user_id, created_at, updated_at, updated_by_user_id, incident_version, closed_at
`, request.IncidentKey, request.IncidentKey, request.Title, request.Description, request.Severity, request.TLP, request.CurrentPhase, request.PrimaryExternalCaseRef, actor.ID, now.UTC()).Scan(
		&incident.ID,
		&incident.IncidentKey,
		&incident.Title,
		&incident.Description,
		&incident.Status,
		&incident.Severity,
		&incident.TLP,
		&incident.CurrentPhase,
		&incident.PrimaryExternalCaseRef,
		&incident.CreatedByUserID,
		&incident.CreatedAt,
		&incident.UpdatedAt,
		&incident.UpdatedByUserID,
		&incident.IncidentVersion,
		&incident.ClosedAt,
	); err != nil {
		if isIncidentKeyConflict(err) {
			return CreateIncidentResult{}, ErrIncidentKeyConflict
		}
		return CreateIncidentResult{}, fmt.Errorf("insert incident: %w", err)
	}

	var membership MembershipRecord
	if err := tx.QueryRow(ctx, `
INSERT INTO incident_memberships (
    incident_id, user_id, role, joined_at, added_by_user_id, updated_at, updated_by_user_id, membership_version
)
VALUES ($1, $2, $4, $3, $2, $3, $2, 1)
RETURNING incident_id, user_id, $5::text AS display_name, role, joined_at, added_by_user_id, updated_at, updated_by_user_id, membership_version
`, incident.ID, actor.ID, now.UTC(), bootstrap.CreatorRole, actor.DisplayName).Scan(
		&membership.IncidentID,
		&membership.UserID,
		&membership.DisplayName,
		&membership.Role,
		&membership.JoinedAt,
		&membership.AddedByUserID,
		&membership.UpdatedAt,
		&membership.UpdatedByUserID,
		&membership.MembershipVersion,
	); err != nil {
		return CreateIncidentResult{}, fmt.Errorf("insert bootstrap membership: %w", err)
	}

	if bootstrap.CreatesIncidentWorkbookPreferences {
		if _, err := tx.Exec(ctx, `
INSERT INTO incident_workbook_preferences (incident_id, default_sheet_ref, created_at, updated_at, updated_by_user_id)
VALUES ($1, NULL, $2, $2, $3)
`, incident.ID, now.UTC(), actor.ID); err != nil {
			return CreateIncidentResult{}, fmt.Errorf("insert incident workbook preferences: %w", err)
		}
	}

	if bootstrap.CreatesUserWorkbookPreferences {
		if _, err := tx.Exec(ctx, `
INSERT INTO user_workbook_preferences (incident_id, user_id, home_sheet_ref, created_at, updated_at)
VALUES ($1, $2, NULL, $3, $3)
`, incident.ID, actor.ID, now.UTC()); err != nil {
			return CreateIncidentResult{}, fmt.Errorf("insert user workbook preferences: %w", err)
		}
	}

	incidentPayload := BuildIncidentResource(incident)
	membershipPayload := BuildMembershipResource(membership)
	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &actor.ID,
		IncidentID:   &incident.ID,
		EventSource:  "incidents",
		EventKind:    "incident_created",
		ClientTxnID:  &request.ClientTxnID,
		RequestID:    &requestID,
		AfterJSON:    incidentPayload,
	}); err != nil {
		return CreateIncidentResult{}, err
	}
	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &actor.ID,
		IncidentID:   &incident.ID,
		EventSource:  "incidents",
		EventKind:    "incident_membership_created",
		ClientTxnID:  &request.ClientTxnID,
		RequestID:    &requestID,
		AfterJSON:    membershipPayload,
	}); err != nil {
		return CreateIncidentResult{}, err
	}

	responseJSON, err := json.Marshal(incidentPayload)
	if err != nil {
		return CreateIncidentResult{}, fmt.Errorf("marshal incident payload: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key, scope_key, client_txn_id, actor_user_id, target_user_id, request_hash, status_code, response_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, "incidents.create", scopeKey, request.ClientTxnID, actor.ID, actor.ID, requestHash, http.StatusCreated, responseJSON); err != nil {
		if authn.IsUniqueViolation(err) {
			return CreateIncidentResult{}, authn.ErrClientTxnConflict
		}
		return CreateIncidentResult{}, fmt.Errorf("insert incident idempotency: %w", err)
	}

	if err := s.beforeCommit("incidents.create", incident.ID); err != nil {
		return CreateIncidentResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateIncidentResult{}, fmt.Errorf("commit incident create transaction: %w", err)
	}
	return CreateIncidentResult{
		Incident:   incident,
		Payload:    incidentPayload,
		StatusCode: http.StatusCreated,
		Location:   incidentLocation(incident.ID),
	}, nil
}

func (s *Store) UpdateIncident(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request IncidentPatchRequest, requestID string, now time.Time) (IncidentRecord, bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IncidentRecord{}, false, fmt.Errorf("begin incident patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	row := tx.QueryRow(ctx, `
SELECT id, incident_key, title, description, status, severity, tlp, current_phase, primary_external_case_ref,
       created_by_user_id, created_at, updated_at, updated_by_user_id, incident_version, closed_at
  FROM incidents
 WHERE id = $1
 FOR UPDATE
`, incidentID)
	current, err := scanIncident(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return IncidentRecord{}, false, ErrIncidentNotFound
	}
	if err != nil {
		return IncidentRecord{}, false, fmt.Errorf("query incident for patch: %w", err)
	}
	if current.IncidentVersion != request.BaseIncidentVersion {
		return IncidentRecord{}, false, ErrIncidentVersionConflict
	}

	next, changed := ApplyIncidentPatch(current, request, actor.ID, now)
	if !changed {
		if err := tx.Commit(ctx); err != nil {
			return IncidentRecord{}, false, fmt.Errorf("commit incident no-op patch transaction: %w", err)
		}
		return current, false, nil
	}

	var updated IncidentRecord
	if err := tx.QueryRow(ctx, `
UPDATE incidents
   SET tlp = $2,
       current_phase = $3,
       primary_external_case_ref = $4,
       updated_at = $5,
       updated_by_user_id = $6,
       incident_version = incident_version + 1
 WHERE id = $1
RETURNING id, incident_key, title, description, status, severity, tlp, current_phase, primary_external_case_ref,
          created_by_user_id, created_at, updated_at, updated_by_user_id, incident_version, closed_at
`, incidentID, next.TLP, next.CurrentPhase, next.PrimaryExternalCaseRef, next.UpdatedAt, actor.ID).Scan(
		&updated.ID,
		&updated.IncidentKey,
		&updated.Title,
		&updated.Description,
		&updated.Status,
		&updated.Severity,
		&updated.TLP,
		&updated.CurrentPhase,
		&updated.PrimaryExternalCaseRef,
		&updated.CreatedByUserID,
		&updated.CreatedAt,
		&updated.UpdatedAt,
		&updated.UpdatedByUserID,
		&updated.IncidentVersion,
		&updated.ClosedAt,
	); err != nil {
		return IncidentRecord{}, false, fmt.Errorf("update incident: %w", err)
	}

	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &actor.ID,
		IncidentID:   &incidentID,
		EventSource:  "incidents",
		EventKind:    "incident_updated",
		RequestID:    &requestID,
		BeforeJSON:   BuildIncidentResource(current),
		AfterJSON:    BuildIncidentResource(updated),
	}); err != nil {
		return IncidentRecord{}, false, err
	}

	if err := s.beforeCommit("incidents.patch", incidentID); err != nil {
		return IncidentRecord{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return IncidentRecord{}, false, fmt.Errorf("commit incident patch transaction: %w", err)
	}
	return updated, true, nil
}

func (s *Store) CreateMembership(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, targetUser authn.UserRecord, request MembershipCreateRequest, requestHash []byte, requestID string, now time.Time) (MembershipCreateResult, error) {
	scopeKey := actor.ID.String() + ":" + incidentID.String()
	if existing, err := s.authStore.GetRouteIdempotency(ctx, "incident.memberships.create", scopeKey, request.ClientTxnID); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MembershipCreateResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MembershipCreateResult{}, fmt.Errorf("decode replayed membership payload: %w", err)
		}
		record, err := s.GetMembership(ctx, incidentID, targetUser.ID)
		if err != nil {
			return MembershipCreateResult{}, err
		}
		return MembershipCreateResult{
			Membership: record,
			Payload:    payload,
			StatusCode: http.StatusOK,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MembershipCreateResult{}, fmt.Errorf("query membership create idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MembershipCreateResult{}, fmt.Errorf("begin membership create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	row := tx.QueryRow(ctx, `
SELECT m.incident_id, m.user_id, u.display_name, m.role, m.joined_at, m.added_by_user_id,
       m.updated_at, m.updated_by_user_id, m.membership_version
  FROM incident_memberships m
  JOIN users u ON u.id = m.user_id
 WHERE m.incident_id = $1
   AND m.user_id = $2
 FOR UPDATE
`, incidentID, targetUser.ID)
	current, err := scanMembership(row)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
	case err != nil:
		return MembershipCreateResult{}, fmt.Errorf("query existing membership: %w", err)
	default:
		if current.Role != request.Role {
			return MembershipCreateResult{}, ErrMembershipExistsUsePatch
		}
		payload := BuildMembershipResource(current)
		if err := insertRouteIdempotency(ctx, tx, "incident.memberships.create", scopeKey, request.ClientTxnID, &actor.ID, &targetUser.ID, requestHash, http.StatusOK, payload); err != nil {
			if authn.IsUniqueViolation(err) {
				return MembershipCreateResult{}, authn.ErrClientTxnConflict
			}
			return MembershipCreateResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return MembershipCreateResult{}, fmt.Errorf("commit existing membership transaction: %w", err)
		}
		return MembershipCreateResult{Membership: current, Payload: payload, StatusCode: http.StatusOK}, nil
	}

	var created MembershipRecord
	if err := tx.QueryRow(ctx, `
INSERT INTO incident_memberships (
    incident_id, user_id, role, joined_at, added_by_user_id, updated_at, updated_by_user_id, membership_version
)
VALUES ($1, $2, $3, $4, $5, $4, $5, 1)
RETURNING incident_id, user_id, $6::text AS display_name, role, joined_at, added_by_user_id, updated_at, updated_by_user_id, membership_version
`, incidentID, targetUser.ID, request.Role, now.UTC(), actor.ID, targetUser.DisplayName).Scan(
		&created.IncidentID,
		&created.UserID,
		&created.DisplayName,
		&created.Role,
		&created.JoinedAt,
		&created.AddedByUserID,
		&created.UpdatedAt,
		&created.UpdatedByUserID,
		&created.MembershipVersion,
	); err != nil {
		return MembershipCreateResult{}, fmt.Errorf("insert membership: %w", err)
	}

	payload := BuildMembershipResource(created)
	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &targetUser.ID,
		IncidentID:   &incidentID,
		EventSource:  "incidents",
		EventKind:    "incident_membership_created",
		ClientTxnID:  &request.ClientTxnID,
		RequestID:    &requestID,
		AfterJSON:    payload,
	}); err != nil {
		return MembershipCreateResult{}, err
	}
	if err := insertRouteIdempotency(ctx, tx, "incident.memberships.create", scopeKey, request.ClientTxnID, &actor.ID, &targetUser.ID, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MembershipCreateResult{}, authn.ErrClientTxnConflict
		}
		return MembershipCreateResult{}, err
	}

	if err := s.beforeCommit("incident.memberships.create", incidentID); err != nil {
		return MembershipCreateResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MembershipCreateResult{}, fmt.Errorf("commit membership create transaction: %w", err)
	}
	return MembershipCreateResult{Membership: created, Payload: payload, StatusCode: http.StatusCreated}, nil
}

func (s *Store) GetMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (MembershipRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT m.incident_id, m.user_id, u.display_name, m.role, m.joined_at, m.added_by_user_id,
       m.updated_at, m.updated_by_user_id, m.membership_version
  FROM incident_memberships m
  JOIN users u ON u.id = m.user_id
 WHERE m.incident_id = $1
   AND m.user_id = $2
`, incidentID, userID)
	record, err := scanMembership(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return MembershipRecord{}, ErrMembershipNotFound
	}
	return record, err
}

func (s *Store) UpdateMembership(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, userID uuid.UUID, request MembershipPatchRequest, requestID string, now time.Time) (MembershipRecord, bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MembershipRecord{}, false, fmt.Errorf("begin membership patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	row := tx.QueryRow(ctx, `
SELECT m.incident_id, m.user_id, u.display_name, m.role, m.joined_at, m.added_by_user_id,
       m.updated_at, m.updated_by_user_id, m.membership_version
  FROM incident_memberships m
  JOIN users u ON u.id = m.user_id
 WHERE m.incident_id = $1
   AND m.user_id = $2
 FOR UPDATE
`, incidentID, userID)
	current, err := scanMembership(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return MembershipRecord{}, false, ErrMembershipNotFound
	}
	if err != nil {
		return MembershipRecord{}, false, fmt.Errorf("query membership for patch: %w", err)
	}
	if current.MembershipVersion != request.BaseMembershipVersion {
		return MembershipRecord{}, false, ErrMembershipVersionConflict
	}
	if current.Role == request.Role {
		if err := tx.Commit(ctx); err != nil {
			return MembershipRecord{}, false, fmt.Errorf("commit membership no-op patch transaction: %w", err)
		}
		return current, false, nil
	}

	adminCount, err := countIncidentAdminsTx(ctx, tx, incidentID)
	if err != nil {
		return MembershipRecord{}, false, fmt.Errorf("count incident admins: %w", err)
	}
	if WouldLeaveNoIncidentAdmins(current.Role, adminCount, &request.Role, false) {
		return MembershipRecord{}, false, ErrLastIncidentAdmin
	}

	var updated MembershipRecord
	if err := tx.QueryRow(ctx, `
UPDATE incident_memberships
   SET role = $3,
       updated_at = $4,
       updated_by_user_id = $5,
       membership_version = membership_version + 1
 WHERE incident_id = $1
   AND user_id = $2
RETURNING incident_id, user_id, $6::text AS display_name, role, joined_at, added_by_user_id, updated_at, updated_by_user_id, membership_version
`, incidentID, userID, request.Role, now.UTC(), actor.ID, current.DisplayName).Scan(
		&updated.IncidentID,
		&updated.UserID,
		&updated.DisplayName,
		&updated.Role,
		&updated.JoinedAt,
		&updated.AddedByUserID,
		&updated.UpdatedAt,
		&updated.UpdatedByUserID,
		&updated.MembershipVersion,
	); err != nil {
		return MembershipRecord{}, false, fmt.Errorf("update membership: %w", err)
	}

	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &userID,
		IncidentID:   &incidentID,
		EventSource:  "incidents",
		EventKind:    "incident_membership_updated",
		RequestID:    &requestID,
		BeforeJSON:   BuildMembershipResource(current),
		AfterJSON:    BuildMembershipResource(updated),
	}); err != nil {
		return MembershipRecord{}, false, err
	}

	if err := s.beforeCommit("incident.memberships.patch", incidentID); err != nil {
		return MembershipRecord{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MembershipRecord{}, false, fmt.Errorf("commit membership patch transaction: %w", err)
	}
	return updated, true, nil
}

func (s *Store) DeleteMembership(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, userID uuid.UUID, request MembershipDeleteRequest, requestID string) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin membership delete transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	row := tx.QueryRow(ctx, `
SELECT m.incident_id, m.user_id, u.display_name, m.role, m.joined_at, m.added_by_user_id,
       m.updated_at, m.updated_by_user_id, m.membership_version
  FROM incident_memberships m
  JOIN users u ON u.id = m.user_id
 WHERE m.incident_id = $1
   AND m.user_id = $2
 FOR UPDATE
`, incidentID, userID)
	current, err := scanMembership(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrMembershipNotFound
	}
	if err != nil {
		return fmt.Errorf("query membership for delete: %w", err)
	}
	if current.MembershipVersion != request.BaseMembershipVersion {
		return ErrMembershipVersionConflict
	}

	adminCount, err := countIncidentAdminsTx(ctx, tx, incidentID)
	if err != nil {
		return fmt.Errorf("count incident admins: %w", err)
	}
	if WouldLeaveNoIncidentAdmins(current.Role, adminCount, nil, true) {
		return ErrLastIncidentAdmin
	}

	if _, err := tx.Exec(ctx, `
DELETE FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, userID); err != nil {
		return fmt.Errorf("delete membership: %w", err)
	}

	deleted := map[string]any{"deleted": true}
	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &userID,
		IncidentID:   &incidentID,
		EventSource:  "incidents",
		EventKind:    "incident_membership_deleted",
		RequestID:    &requestID,
		BeforeJSON:   BuildMembershipResource(current),
		AfterJSON:    deleted,
	}); err != nil {
		return err
	}

	if err := s.beforeCommit("incident.memberships.delete", incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit membership delete transaction: %w", err)
	}
	return nil
}

type auditEvent struct {
	ActorUserID  *uuid.UUID
	TargetUserID *uuid.UUID
	IncidentID   *uuid.UUID
	EventSource  string
	EventKind    string
	ReasonCode   *string
	ClientTxnID  *string
	RequestID    *string
	BeforeJSON   any
	AfterJSON    any
}

func insertAuditEvent(ctx context.Context, tx pgx.Tx, event auditEvent) error {
	_, err := tx.Exec(ctx, `
INSERT INTO deployment_admin_audit_events (
    actor_user_id, target_user_id, incident_id, event_source, event_kind, reason_code, client_txn_id, request_id, before_json, after_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`, event.ActorUserID, event.TargetUserID, event.IncidentID, event.EventSource, event.EventKind, event.ReasonCode, event.ClientTxnID, event.RequestID, jsonOrNil(event.BeforeJSON), jsonOrNil(event.AfterJSON))
	if err != nil {
		return fmt.Errorf("insert incident audit event: %w", err)
	}
	return nil
}

func insertRouteIdempotency(ctx context.Context, tx pgx.Tx, routeKey string, scopeKey string, clientTxnID string, actorUserID *uuid.UUID, targetUserID *uuid.UUID, requestHash []byte, statusCode int, payload map[string]any) error {
	responseJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal idempotency payload: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key, scope_key, client_txn_id, actor_user_id, target_user_id, request_hash, status_code, response_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, routeKey, scopeKey, clientTxnID, actorUserID, targetUserID, requestHash, statusCode, responseJSON); err != nil {
		return fmt.Errorf("insert route idempotency: %w", err)
	}
	return nil
}

func countIncidentAdminsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (int, error) {
	var count int
	if err := tx.QueryRow(ctx, `
SELECT COUNT(*)
  FROM incident_memberships
 WHERE incident_id = $1
   AND role = 'admin'
`, incidentID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func incidentLocation(incidentID uuid.UUID) string {
	return "/api/v1/incidents/" + incidentID.String()
}

func (s *Store) beforeCommit(routeKey string, incidentID uuid.UUID) error {
	if s == nil || s.hooks.BeforeCommit == nil {
		return nil
	}
	return s.hooks.BeforeCommit(routeKey, incidentID)
}

func jsonOrNil(value any) any {
	if value == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return payload
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanIncident(row rowScanner) (IncidentRecord, error) {
	var record IncidentRecord
	err := row.Scan(
		&record.ID,
		&record.IncidentKey,
		&record.Title,
		&record.Description,
		&record.Status,
		&record.Severity,
		&record.TLP,
		&record.CurrentPhase,
		&record.PrimaryExternalCaseRef,
		&record.CreatedByUserID,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.UpdatedByUserID,
		&record.IncidentVersion,
		&record.ClosedAt,
	)
	return record, err
}

func scanMembership(row rowScanner) (MembershipRecord, error) {
	var record MembershipRecord
	err := row.Scan(
		&record.IncidentID,
		&record.UserID,
		&record.DisplayName,
		&record.Role,
		&record.JoinedAt,
		&record.AddedByUserID,
		&record.UpdatedAt,
		&record.UpdatedByUserID,
		&record.MembershipVersion,
	)
	return record, err
}

func stringPointersEqual(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func extractUUID(value any) (uuid.UUID, error) {
	text, ok := value.(string)
	if !ok || text == "" {
		return uuid.UUID{}, errors.New("missing uuid string")
	}
	return uuid.Parse(text)
}

func isIncidentKeyConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && pgErr.ConstraintName == "incidents_incident_key_canonical_key"
}
