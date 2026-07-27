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
	"github.com/jackc/pgx/v5/pgtype"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/workbookpreferences"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var (
	ErrIncidentNotFound          = errors.New("incidents: incident not found")
	ErrIncidentKeyConflict       = errors.New("incidents: incident key conflict")
	ErrIncidentVersionConflict   = errors.New("incidents: incident version conflict")
	ErrMembershipNotFound        = errors.New("incidents: membership not found")
	ErrMembershipExistsUsePatch  = errors.New("incidents: membership exists use patch")
	ErrMembershipVersionConflict = errors.New("incidents: membership version conflict")
	ErrLastIncidentAdmin         = errors.New("incidents: last incident admin")
	ErrIncidentClosed            = errors.New("incidents: incident closed")
	ErrIncidentIllegalTransition = errors.New("incidents: illegal incident transition")
)

type Store struct {
	pool                postgres.DB
	authStore           *authn.Store
	preferenceBootstrap PreferenceBootstrapPort
}

func (s *Store) ListAdministrativeAuditEvents(
	ctx context.Context,
	filter administrativeaudit.ListFilter,
) ([]administrativeaudit.Record, error) {
	return administrativeaudit.List(ctx, s.pool, filter)
}

// IncidentVersionConflictError carries the optimistic-concurrency values needed
// for clients to reconcile a stale incident metadata patch.
type IncidentVersionConflictError struct {
	IncidentID             uuid.UUID
	BaseIncidentVersion    int64
	CurrentIncidentVersion int64
}

func (e *IncidentVersionConflictError) Error() string {
	return ErrIncidentVersionConflict.Error()
}

func (e *IncidentVersionConflictError) Unwrap() error {
	return ErrIncidentVersionConflict
}

func (e *IncidentVersionConflictError) Details() map[string]any {
	return map[string]any{
		"incident_id":              e.IncidentID.String(),
		"base_incident_version":    e.BaseIncidentVersion,
		"current_incident_version": e.CurrentIncidentVersion,
	}
}

type IncidentListPosition struct {
	UpdatedAt time.Time
	ID        uuid.UUID
}

type IncidentListPageRequest struct {
	AnchorUpdatedAt *time.Time
	After           *IncidentListPosition
	Limit           int
	SearchTokens    []string
	Status          *string
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

type IncidentLifecycleResult struct {
	Incident   IncidentRecord
	Payload    map[string]any
	StatusCode int
	Replayed   bool
}

func NewStore(pool postgres.DB) *Store {
	return NewStoreWithOptions(pool, StoreOptions{})
}

func NewStoreWithOptions(pool postgres.DB, options StoreOptions) *Store {
	preferenceBootstrap := options.PreferenceBootstrap
	if preferenceBootstrap == nil {
		preferenceBootstrap = workbookpreferences.NewBootstrap()
	}
	return &Store{
		pool:                pool,
		authStore:           authn.NewStore(pool),
		preferenceBootstrap: preferenceBootstrap,
	}
}

func (s *Store) ListVisibleIncidents(ctx context.Context, userID uuid.UUID, page IncidentListPageRequest) ([]IncidentRecord, error) {
	if page.Limit < 1 {
		page.Limit = 1
	}
	if len(page.SearchTokens) > 0 {
		return s.listVisibleIncidentsWithSearch(ctx, userID, page)
	}
	return s.listVisibleIncidentCandidates(ctx, userID, page, page.After, page.Limit)
}

func (s *Store) listVisibleIncidentsWithSearch(ctx context.Context, userID uuid.UUID, page IncidentListPageRequest) ([]IncidentRecord, error) {
	records := make([]IncidentRecord, 0, page.Limit)
	after := page.After
	candidateLimit := page.Limit
	if candidateLimit < 100 {
		candidateLimit = 100
	}
	for {
		candidates, err := s.listVisibleIncidentCandidates(ctx, userID, page, after, candidateLimit)
		if err != nil {
			return nil, err
		}
		if len(candidates) == 0 {
			return records, nil
		}
		for _, candidate := range candidates {
			if !listquery.MatchSearchTokens(page.SearchTokens,
				candidate.IncidentKey,
				candidate.Title,
				optionalStringValue(candidate.Severity),
				optionalStringValue(candidate.TLP),
				optionalStringValue(candidate.CurrentPhase),
				optionalStringValue(candidate.PrimaryExternalCaseRef),
			) {
				continue
			}
			records = append(records, candidate)
			if len(records) >= page.Limit {
				return records, nil
			}
		}
		if len(candidates) < candidateLimit {
			return records, nil
		}
		last := candidates[len(candidates)-1]
		after = &IncidentListPosition{UpdatedAt: last.UpdatedAt, ID: last.ID}
	}
}

func (s *Store) listVisibleIncidentCandidates(ctx context.Context, userID uuid.UUID, page IncidentListPageRequest, after *IncidentListPosition, limit int) ([]IncidentRecord, error) {
	var afterUpdatedAt *time.Time
	var afterID *uuid.UUID
	if after != nil {
		updatedAt := after.UpdatedAt.UTC()
		afterUpdatedAt = &updatedAt
		id := after.ID
		afterID = &id
	}
	filterByStatus := false
	status := ""
	if page.Status != nil {
		filterByStatus = true
		status = *page.Status
	}
	rows, err := sqlc.New(s.pool).ListVisibleIncidents(ctx, sqlc.ListVisibleIncidentsParams{
		UserID:  pgUUID(userID),
		Column2: pgOptionalTimestamptzPtr(page.AnchorUpdatedAt),
		Column3: pgOptionalTimestamptzPtr(afterUpdatedAt),
		Column4: pgOptionalUUIDPtr(afterID),
		Limit:   int32(limit),
		Column6: filterByStatus,
		Status:  status,
	})
	if err != nil {
		return nil, fmt.Errorf("list visible incidents: %w", err)
	}

	records := make([]IncidentRecord, 0, page.Limit)
	for _, row := range rows {
		record, err := incidentRecordFromSQL(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *Store) GetVisibleIncident(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (IncidentRecord, error) {
	row, err := sqlc.New(s.pool).GetVisibleIncidentByID(ctx, sqlc.GetVisibleIncidentByIDParams{
		ID:     pgUUID(incidentID),
		UserID: pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return IncidentRecord{}, ErrIncidentNotFound
	}
	if err != nil {
		return IncidentRecord{}, err
	}
	return incidentRecordFromSQL(row)
}

func (s *Store) GetIncidentMembershipForUser(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (MembershipRecord, error) {
	row, err := sqlc.New(s.pool).GetIncidentMembershipForActor(ctx, sqlc.GetIncidentMembershipForActorParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MembershipRecord{}, ErrMembershipNotFound
	}
	if err != nil {
		return MembershipRecord{}, err
	}
	return membershipRecordFromSQL(row)
}

func (s *Store) ListMemberships(ctx context.Context, incidentID uuid.UUID) ([]MembershipRecord, error) {
	rows, err := sqlc.New(s.pool).ListAllIncidentMemberships(ctx, pgUUID(incidentID))
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}
	records := make([]MembershipRecord, 0, 16)
	for _, row := range rows {
		record, err := membershipRecordFromSQL(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *Store) CreateIncident(ctx context.Context, actor authn.UserRecord, request CreateIncidentRequest, requestHash []byte, requestID string, now time.Time) (CreateIncidentResult, error) {
	key := authn.ActorOnlyRouteIdempotencyKey("incidents.create", actor.ID, request.ClientTxnID)
	if existing, err := s.authStore.GetRouteIdempotency(ctx, key); err == nil {
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
	q := sqlc.New(tx)
	incidentRow, err := q.CreateIncident(ctx, sqlc.CreateIncidentParams{
		IncidentKey:            request.IncidentKey,
		IncidentKeyCanonical:   request.IncidentKey,
		Title:                  request.Title,
		Description:            pgTextPtr(request.Description),
		Severity:               pgTextPtr(request.Severity),
		Tlp:                    pgTextPtr(request.TLP),
		CurrentPhase:           pgTextPtr(request.CurrentPhase),
		PrimaryExternalCaseRef: pgTextPtr(request.PrimaryExternalCaseRef),
		CreatedByUserID:        pgUUID(actor.ID),
		CreatedAt:              pgTimestamptz(now),
	})
	if err != nil {
		if isIncidentKeyConflict(err) {
			return CreateIncidentResult{}, ErrIncidentKeyConflict
		}
		return CreateIncidentResult{}, fmt.Errorf("insert incident: %w", err)
	}
	incident, err := incidentRecordFromSQL(incidentRow)
	if err != nil {
		return CreateIncidentResult{}, err
	}

	membershipRow, err := q.CreateBootstrapIncidentMembership(ctx, sqlc.CreateBootstrapIncidentMembershipParams{
		IncidentID: pgUUID(incident.ID),
		UserID:     pgUUID(actor.ID),
		JoinedAt:   pgTimestamptz(now),
		Role:       bootstrap.CreatorRole,
		Column5:    actor.DisplayName,
	})
	if err != nil {
		return CreateIncidentResult{}, fmt.Errorf("insert bootstrap membership: %w", err)
	}
	membership, err := membershipRecordFromSQL(membershipRow)
	if err != nil {
		return CreateIncidentResult{}, err
	}

	if err := s.preferenceBootstrap.BootstrapIncidentPreferencesTx(ctx, tx, incident.ID, actor.ID, now); err != nil {
		return CreateIncidentResult{}, err
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
	if err := authn.InsertRouteIdempotency(ctx, tx, key, &actor.ID, requestHash, http.StatusCreated, responseJSON); err != nil {
		if authn.IsUniqueViolation(err) {
			return CreateIncidentResult{}, authn.ErrClientTxnConflict
		}
		return CreateIncidentResult{}, fmt.Errorf("insert incident idempotency: %w", err)
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

	q := sqlc.New(tx)
	currentRow, err := q.GetIncidentForUpdate(ctx, pgUUID(incidentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IncidentRecord{}, false, ErrIncidentNotFound
	}
	if err != nil {
		return IncidentRecord{}, false, fmt.Errorf("query incident for patch: %w", err)
	}
	current, err := incidentRecordFromSQL(currentRow)
	if err != nil {
		return IncidentRecord{}, false, err
	}
	if current.IncidentVersion != request.BaseIncidentVersion {
		return IncidentRecord{}, false, &IncidentVersionConflictError{
			IncidentID:             incidentID,
			BaseIncidentVersion:    request.BaseIncidentVersion,
			CurrentIncidentVersion: current.IncidentVersion,
		}
	}
	if current.Status == "closed" {
		return IncidentRecord{}, false, ErrIncidentClosed
	}

	next, changed := ApplyIncidentPatch(current, request, actor.ID, now)
	if !changed {
		if err := tx.Commit(ctx); err != nil {
			return IncidentRecord{}, false, fmt.Errorf("commit incident no-op patch transaction: %w", err)
		}
		return current, false, nil
	}

	updatedRow, err := q.UpdateIncidentMetadata(ctx, sqlc.UpdateIncidentMetadataParams{
		ID:                     pgUUID(incidentID),
		Description:            pgTextPtr(next.Description),
		Severity:               pgTextPtr(next.Severity),
		Tlp:                    pgTextPtr(next.TLP),
		CurrentPhase:           pgTextPtr(next.CurrentPhase),
		PrimaryExternalCaseRef: pgTextPtr(next.PrimaryExternalCaseRef),
		UpdatedAt:              pgTimestamptz(next.UpdatedAt),
		UpdatedByUserID:        pgUUID(actor.ID),
	})
	if err != nil {
		return IncidentRecord{}, false, fmt.Errorf("update incident: %w", err)
	}
	updated, err := incidentRecordFromSQL(updatedRow)
	if err != nil {
		return IncidentRecord{}, false, err
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

	if err := tx.Commit(ctx); err != nil {
		return IncidentRecord{}, false, fmt.Errorf("commit incident patch transaction: %w", err)
	}
	return updated, true, nil
}

func (s *Store) TransitionIncidentLifecycle(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, action string, request IncidentLifecycleRequest, requestHash []byte, requestID string, now time.Time) (IncidentLifecycleResult, error) {
	routeKey := "incidents." + action
	key := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    incidentID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, key); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return IncidentLifecycleResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return IncidentLifecycleResult{}, fmt.Errorf("decode replayed incident lifecycle payload: %w", err)
		}
		return IncidentLifecycleResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return IncidentLifecycleResult{}, fmt.Errorf("query incident lifecycle idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IncidentLifecycleResult{}, fmt.Errorf("begin incident lifecycle transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := sqlc.New(tx)
	currentRow, err := q.GetIncidentForUpdate(ctx, pgUUID(incidentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IncidentLifecycleResult{}, ErrIncidentNotFound
	}
	if err != nil {
		return IncidentLifecycleResult{}, fmt.Errorf("query incident for lifecycle: %w", err)
	}
	current, err := incidentRecordFromSQL(currentRow)
	if err != nil {
		return IncidentLifecycleResult{}, err
	}
	if current.IncidentVersion != request.BaseIncidentVersion {
		return IncidentLifecycleResult{}, &IncidentVersionConflictError{
			IncidentID:             incidentID,
			BaseIncidentVersion:    request.BaseIncidentVersion,
			CurrentIncidentVersion: current.IncidentVersion,
		}
	}

	nextStatus := ""
	nextClosedAt := pgtype.Timestamptz{}
	switch action {
	case "close":
		if current.Status != "active" {
			return IncidentLifecycleResult{}, ErrIncidentIllegalTransition
		}
		nextStatus = "closed"
		nextClosedAt = pgTimestamptz(now)
	case "reopen":
		if current.Status != "closed" {
			return IncidentLifecycleResult{}, ErrIncidentIllegalTransition
		}
		nextStatus = "active"
	default:
		return IncidentLifecycleResult{}, ErrIncidentIllegalTransition
	}

	updatedRow, err := q.UpdateIncidentLifecycle(ctx, sqlc.UpdateIncidentLifecycleParams{
		ID:              pgUUID(incidentID),
		Status:          nextStatus,
		ClosedAt:        nextClosedAt,
		UpdatedAt:       pgTimestamptz(now),
		UpdatedByUserID: pgUUID(actor.ID),
	})
	if err != nil {
		return IncidentLifecycleResult{}, fmt.Errorf("update incident lifecycle: %w", err)
	}
	updated, err := incidentRecordFromSQL(updatedRow)
	if err != nil {
		return IncidentLifecycleResult{}, err
	}

	payload := BuildIncidentResource(updated)
	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &actor.ID,
		IncidentID:   &incidentID,
		EventSource:  "incidents",
		EventKind:    "incident_" + action,
		ReasonCode:   &request.Reason,
		ClientTxnID:  &request.ClientTxnID,
		RequestID:    &requestID,
		BeforeJSON:   BuildIncidentResource(current),
		AfterJSON:    payload,
	}); err != nil {
		return IncidentLifecycleResult{}, err
	}

	responseJSON, err := json.Marshal(payload)
	if err != nil {
		return IncidentLifecycleResult{}, fmt.Errorf("marshal incident lifecycle payload: %w", err)
	}
	if err := authn.InsertRouteIdempotency(ctx, tx, key, &actor.ID, requestHash, http.StatusOK, responseJSON); err != nil {
		if authn.IsUniqueViolation(err) {
			return IncidentLifecycleResult{}, authn.ErrClientTxnConflict
		}
		return IncidentLifecycleResult{}, fmt.Errorf("insert incident lifecycle idempotency: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return IncidentLifecycleResult{}, fmt.Errorf("commit incident lifecycle transaction: %w", err)
	}
	return IncidentLifecycleResult{
		Incident:   updated,
		Payload:    payload,
		StatusCode: http.StatusOK,
	}, nil
}

func (s *Store) CreateMembership(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, targetUser authn.UserRecord, request MembershipCreateRequest, requestHash []byte, requestID string, now time.Time) (MembershipCreateResult, error) {
	scopeKey := incidentID.String()
	key := authn.RouteIdempotencyKey{
		RouteKey:    "incident.memberships.create",
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, key); err == nil {
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

	q := sqlc.New(tx)
	currentRow, err := q.GetIncidentMembershipForUpdate(ctx, sqlc.GetIncidentMembershipForUpdateParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(targetUser.ID),
	})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
	case err != nil:
		return MembershipCreateResult{}, fmt.Errorf("query existing membership: %w", err)
	default:
		current, err := membershipRecordFromSQL(currentRow)
		if err != nil {
			return MembershipCreateResult{}, err
		}
		if current.Role != request.Role {
			return MembershipCreateResult{}, ErrMembershipExistsUsePatch
		}
		payload := BuildMembershipResource(current)
		if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, &targetUser.ID, requestHash, http.StatusOK, payload); err != nil {
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

	createdRow, err := q.CreateIncidentMembership(ctx, sqlc.CreateIncidentMembershipParams{
		IncidentID:    pgUUID(incidentID),
		UserID:        pgUUID(targetUser.ID),
		Role:          request.Role,
		JoinedAt:      pgTimestamptz(now),
		AddedByUserID: pgUUID(actor.ID),
		Column6:       targetUser.DisplayName,
	})
	if err != nil {
		return MembershipCreateResult{}, fmt.Errorf("insert membership: %w", err)
	}
	created, err := membershipRecordFromSQL(createdRow)
	if err != nil {
		return MembershipCreateResult{}, err
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
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, &targetUser.ID, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MembershipCreateResult{}, authn.ErrClientTxnConflict
		}
		return MembershipCreateResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return MembershipCreateResult{}, fmt.Errorf("commit membership create transaction: %w", err)
	}
	return MembershipCreateResult{Membership: created, Payload: payload, StatusCode: http.StatusCreated}, nil
}

func (s *Store) GetMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (MembershipRecord, error) {
	row, err := sqlc.New(s.pool).GetIncidentMembershipForActor(ctx, sqlc.GetIncidentMembershipForActorParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MembershipRecord{}, ErrMembershipNotFound
	}
	if err != nil {
		return MembershipRecord{}, err
	}
	return membershipRecordFromSQL(row)
}

func (s *Store) UpdateMembership(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, userID uuid.UUID, request MembershipPatchRequest, requestID string, now time.Time) (MembershipRecord, bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MembershipRecord{}, false, fmt.Errorf("begin membership patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := sqlc.New(tx)
	currentRow, err := q.GetIncidentMembershipForUpdate(ctx, sqlc.GetIncidentMembershipForUpdateParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return MembershipRecord{}, false, ErrMembershipNotFound
	}
	if err != nil {
		return MembershipRecord{}, false, fmt.Errorf("query membership for patch: %w", err)
	}
	current, err := membershipRecordFromSQL(currentRow)
	if err != nil {
		return MembershipRecord{}, false, err
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

	updatedRow, err := q.UpdateIncidentMembershipRole(ctx, sqlc.UpdateIncidentMembershipRoleParams{
		IncidentID:      pgUUID(incidentID),
		UserID:          pgUUID(userID),
		Role:            request.Role,
		UpdatedAt:       pgTimestamptz(now),
		UpdatedByUserID: pgUUID(actor.ID),
		Column6:         current.DisplayName,
	})
	if err != nil {
		return MembershipRecord{}, false, fmt.Errorf("update membership: %w", err)
	}
	updated, err := membershipRecordFromSQL(updatedRow)
	if err != nil {
		return MembershipRecord{}, false, err
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

	q := sqlc.New(tx)
	currentRow, err := q.GetIncidentMembershipForUpdate(ctx, sqlc.GetIncidentMembershipForUpdateParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrMembershipNotFound
	}
	if err != nil {
		return fmt.Errorf("query membership for delete: %w", err)
	}
	current, err := membershipRecordFromSQL(currentRow)
	if err != nil {
		return err
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

	if err := q.DeleteIncidentMembership(ctx, sqlc.DeleteIncidentMembershipParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	}); err != nil {
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
	PublicSource string
}

func insertAuditEvent(ctx context.Context, tx pgx.Tx, event auditEvent) error {
	occurredAt := time.Now().UTC()
	raw := administrativeaudit.RawEvent{
		ActorUserID:  event.ActorUserID,
		TargetUserID: event.TargetUserID,
		IncidentID:   event.IncidentID,
		EventSource:  event.EventSource,
		EventKind:    event.EventKind,
		ReasonCode:   event.ReasonCode,
		ClientTxnID:  event.ClientTxnID,
		RequestID:    event.RequestID,
		Before:       event.BeforeJSON,
		After:        event.AfterJSON,
		OccurredAt:   occurredAt,
	}
	actionCode, changes, projected := membershipAuditProjection(event)
	if !projected {
		if _, err := administrativeaudit.AppendRawTx(ctx, tx, raw); err != nil {
			return fmt.Errorf("insert incident audit event: %w", err)
		}
		return nil
	}
	if event.IncidentID == nil || event.ActorUserID == nil || event.TargetUserID == nil {
		return errors.New("insert incident membership audit event: projection identifiers are incomplete")
	}
	source := event.PublicSource
	if source == "" {
		source = administrativeaudit.SourceAPI
	}
	targetID := event.TargetUserID.String()
	if _, err := administrativeaudit.AppendTx(ctx, tx, raw, administrativeaudit.Event{
		ScopeKind:   administrativeaudit.ScopeIncident,
		ScopeID:     event.IncidentID,
		OccurredAt:  occurredAt,
		ActorKind:   administrativeaudit.ActorUser,
		ActorUserID: event.ActorUserID,
		Source:      source,
		ActionCode:  actionCode,
		TargetKind:  administrativeaudit.TargetIncidentMembership,
		TargetID:    &targetID,
		Changes:     changes,
		ReasonCode:  event.ReasonCode,
	}); err != nil {
		return fmt.Errorf("insert incident audit event: %w", err)
	}
	return nil
}

func membershipAuditProjection(event auditEvent) (string, []administrativeaudit.Change, bool) {
	beforeRole := membershipRole(event.BeforeJSON)
	afterRole := membershipRole(event.AfterJSON)
	switch event.EventKind {
	case "incident_membership_created":
		return administrativeaudit.ActionMembershipCreated, []administrativeaudit.Change{
			administrativeaudit.Visible("role", nil, afterRole),
		}, true
	case "incident_membership_updated":
		return administrativeaudit.ActionMembershipRoleChanged, []administrativeaudit.Change{
			administrativeaudit.Visible("role", beforeRole, afterRole),
		}, true
	case "incident_membership_deleted":
		return administrativeaudit.ActionMembershipDeleted, []administrativeaudit.Change{
			administrativeaudit.Visible("role", beforeRole, nil),
		}, true
	default:
		return "", nil, false
	}
}

func membershipRole(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return typed["role"]
	case map[string]string:
		return typed["role"]
	default:
		payload, err := json.Marshal(value)
		if err != nil {
			return nil
		}
		var resource struct {
			Role any `json:"role"`
		}
		if err := json.Unmarshal(payload, &resource); err != nil {
			return nil
		}
		return resource.Role
	}
}

func countIncidentAdminsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (int, error) {
	count, err := sqlc.New(tx).CountIncidentAdmins(ctx, pgUUID(incidentID))
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func incidentLocation(incidentID uuid.UUID) string {
	return "/api/v1/incidents/" + incidentID.String()
}

func extractUUID(value any) (uuid.UUID, error) {
	text, ok := value.(string)
	if !ok || text == "" {
		return uuid.UUID{}, errors.New("missing uuid string")
	}
	return uuid.Parse(text)
}

func incidentRecordFromSQL(row sqlc.Incident) (IncidentRecord, error) {
	id, err := uuidFromPG(row.ID)
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("incident id: %w", err)
	}
	createdBy, err := uuidFromPG(row.CreatedByUserID)
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("incident created by: %w", err)
	}
	createdAt, err := timeFromPG(row.CreatedAt)
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("incident created at: %w", err)
	}
	updatedAt, err := timeFromPG(row.UpdatedAt)
	if err != nil {
		return IncidentRecord{}, fmt.Errorf("incident updated at: %w", err)
	}
	return IncidentRecord{
		ID:                     id,
		IncidentKey:            row.IncidentKey,
		Title:                  row.Title,
		Description:            optionalStringFromPG(row.Description),
		Status:                 row.Status,
		Severity:               optionalStringFromPG(row.Severity),
		TLP:                    optionalStringFromPG(row.Tlp),
		CurrentPhase:           optionalStringFromPG(row.CurrentPhase),
		PrimaryExternalCaseRef: optionalStringFromPG(row.PrimaryExternalCaseRef),
		CreatedByUserID:        createdBy,
		CreatedAt:              createdAt,
		UpdatedAt:              updatedAt,
		UpdatedByUserID:        optionalUUIDFromPG(row.UpdatedByUserID),
		IncidentVersion:        row.IncidentVersion,
		ClosedAt:               optionalTimeFromPG(row.ClosedAt),
	}, nil
}

func membershipRecordFromSQL(row any) (MembershipRecord, error) {
	switch r := row.(type) {
	case sqlc.GetIncidentMembershipForActorRow:
		return membershipRecordFromSQLFields(r.IncidentID, r.UserID, r.DisplayName, r.Role, r.JoinedAt, r.AddedByUserID, r.UpdatedAt, r.UpdatedByUserID, r.MembershipVersion)
	case sqlc.GetIncidentMembershipForUpdateRow:
		return membershipRecordFromSQLFields(r.IncidentID, r.UserID, r.DisplayName, r.Role, r.JoinedAt, r.AddedByUserID, r.UpdatedAt, r.UpdatedByUserID, r.MembershipVersion)
	case sqlc.ListAllIncidentMembershipsRow:
		return membershipRecordFromSQLFields(r.IncidentID, r.UserID, r.DisplayName, r.Role, r.JoinedAt, r.AddedByUserID, r.UpdatedAt, r.UpdatedByUserID, r.MembershipVersion)
	case sqlc.ListIncidentMembershipsRow:
		return membershipRecordFromSQLFields(r.IncidentID, r.UserID, r.DisplayName, r.Role, r.JoinedAt, r.AddedByUserID, r.UpdatedAt, r.UpdatedByUserID, r.MembershipVersion)
	case sqlc.CreateBootstrapIncidentMembershipRow:
		return membershipRecordFromSQLFields(r.IncidentID, r.UserID, r.DisplayName, r.Role, r.JoinedAt, r.AddedByUserID, r.UpdatedAt, r.UpdatedByUserID, r.MembershipVersion)
	case sqlc.CreateIncidentMembershipRow:
		return membershipRecordFromSQLFields(r.IncidentID, r.UserID, r.DisplayName, r.Role, r.JoinedAt, r.AddedByUserID, r.UpdatedAt, r.UpdatedByUserID, r.MembershipVersion)
	case sqlc.UpdateIncidentMembershipRoleRow:
		return membershipRecordFromSQLFields(r.IncidentID, r.UserID, r.DisplayName, r.Role, r.JoinedAt, r.AddedByUserID, r.UpdatedAt, r.UpdatedByUserID, r.MembershipVersion)
	default:
		return MembershipRecord{}, fmt.Errorf("unsupported membership SQL row %T", row)
	}
}

func membershipRecordFromSQLFields(
	rowIncidentID pgtype.UUID,
	rowUserID pgtype.UUID,
	rowDisplayName string,
	rowRole string,
	rowJoinedAt pgtype.Timestamptz,
	rowAddedByUserID pgtype.UUID,
	rowUpdatedAt pgtype.Timestamptz,
	rowUpdatedByUserID pgtype.UUID,
	rowMembershipVersion int64,
) (MembershipRecord, error) {
	incidentID, err := uuidFromPG(rowIncidentID)
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("membership incident id: %w", err)
	}
	userID, err := uuidFromPG(rowUserID)
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("membership user id: %w", err)
	}
	addedBy, err := uuidFromPG(rowAddedByUserID)
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("membership added by: %w", err)
	}
	joinedAt, err := timeFromPG(rowJoinedAt)
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("membership joined at: %w", err)
	}
	updatedAt, err := timeFromPG(rowUpdatedAt)
	if err != nil {
		return MembershipRecord{}, fmt.Errorf("membership updated at: %w", err)
	}
	return MembershipRecord{
		IncidentID:        incidentID,
		UserID:            userID,
		DisplayName:       rowDisplayName,
		Role:              rowRole,
		JoinedAt:          joinedAt,
		AddedByUserID:     addedBy,
		UpdatedAt:         updatedAt,
		UpdatedByUserID:   optionalUUIDFromPG(rowUpdatedByUserID),
		MembershipVersion: rowMembershipVersion,
	}, nil
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}

func pgOptionalUUIDPtr(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return pgUUID(*value)
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.FromBytes(value.Bytes[:])
}

func optionalUUIDFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	id, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return nil
	}
	return &id
}

func pgTextPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func optionalStringFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func pgTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func pgOptionalTimestamptzPtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgTimestamptz(*value)
}

func timeFromPG(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, errors.New("missing timestamp")
	}
	return value.Time.UTC(), nil
}

func optionalTimeFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	timestamp := value.Time.UTC()
	return &timestamp
}

func stringPointersEqual(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func isIncidentKeyConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && pgErr.ConstraintName == "incidents_incident_key_canonical_key"
}
