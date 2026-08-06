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

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
)

func (a *Application) ListVisibleIncidents(
	ctx context.Context,
	userID uuid.UUID,
	page IncidentListPageRequest,
) ([]IncidentRecord, error) {
	if page.Limit < 1 {
		page.Limit = 1
	}
	if len(page.SearchTokens) > 0 {
		return a.listVisibleIncidentsWithSearch(ctx, userID, page)
	}
	return a.repository.listVisibleIncidentCandidates(ctx, userID, page, page.After, page.Limit)
}

func (a *Application) listVisibleIncidentsWithSearch(
	ctx context.Context,
	userID uuid.UUID,
	page IncidentListPageRequest,
) ([]IncidentRecord, error) {
	records := make([]IncidentRecord, 0, page.Limit)
	after := page.After
	candidateLimit := page.Limit
	if candidateLimit < 100 {
		candidateLimit = 100
	}
	for {
		candidates, err := a.repository.listVisibleIncidentCandidates(ctx, userID, page, after, candidateLimit)
		if err != nil {
			return nil, err
		}
		if len(candidates) == 0 {
			return records, nil
		}
		for _, candidate := range candidates {
			if !listquery.MatchSearchTokens(
				page.SearchTokens,
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

func (a *Application) GetVisibleIncident(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) (IncidentRecord, error) {
	return a.repository.getVisibleIncident(ctx, incidentID, userID)
}

func (a *Application) GetIncidentMembershipForUser(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) (MembershipRecord, error) {
	return a.repository.getMembership(ctx, incidentID, userID)
}

func (a *Application) CreateIncident(
	ctx context.Context,
	actor authn.UserRecord,
	request CreateIncidentRequest,
	requestHash []byte,
	requestID string,
	now time.Time,
) (CreateIncidentResult, error) {
	key := authn.ActorOnlyRouteIdempotencyKey("incidents.create", actor.ID, request.ClientTxnID)
	if existing, err := a.authStore.GetRouteIdempotency(ctx, key); err == nil {
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

	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CreateIncidentResult{}, fmt.Errorf("begin incident create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txRepository := newRepository(tx)
	bootstrap := DefaultIncidentCreateBootstrap()
	incident, err := txRepository.createIncident(ctx, createIncidentPersistenceParams{
		IncidentKey:            request.IncidentKey,
		Title:                  request.Title,
		Description:            request.Description,
		Severity:               request.Severity,
		TLP:                    request.TLP,
		CurrentPhase:           request.CurrentPhase,
		PrimaryExternalCaseRef: request.PrimaryExternalCaseRef,
		CreatedByUserID:        actor.ID,
		CreatedAt:              now,
	})
	if err != nil {
		return CreateIncidentResult{}, err
	}

	membership, err := txRepository.createBootstrapMembership(ctx, createBootstrapMembershipPersistenceParams{
		IncidentID:  incident.ID,
		UserID:      actor.ID,
		JoinedAt:    now,
		Role:        bootstrap.CreatorRole,
		DisplayName: actor.DisplayName,
	})
	if err != nil {
		return CreateIncidentResult{}, err
	}

	if err := a.preferenceBootstrap.BootstrapIncidentPreferencesTx(ctx, tx, incident.ID, actor.ID, now); err != nil {
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
	if err := authn.InsertRouteIdempotency(
		ctx,
		tx,
		key,
		&actor.ID,
		requestHash,
		http.StatusCreated,
		responseJSON,
	); err != nil {
		if authn.IsUniqueViolation(err) {
			return CreateIncidentResult{}, authn.ErrClientTxnConflict
		}
		return CreateIncidentResult{}, fmt.Errorf("insert incident idempotency: %w", err)
	}

	if err := a.incidentCreateCommit.CommitIncidentCreate(ctx, tx); err != nil {
		return CreateIncidentResult{}, fmt.Errorf("commit incident create transaction: %w", err)
	}
	return CreateIncidentResult{
		Incident:   incident,
		Payload:    incidentPayload,
		StatusCode: http.StatusCreated,
		Location:   incidentLocation(incident.ID),
	}, nil
}

func (a *Application) UpdateIncident(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	request IncidentPatchRequest,
	requestID string,
	now time.Time,
) (IncidentRecord, bool, error) {
	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IncidentRecord{}, false, fmt.Errorf("begin incident patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txRepository := newRepository(tx)
	current, err := txRepository.getIncidentForUpdate(ctx, incidentID)
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

	updated, err := txRepository.updateIncidentMetadata(ctx, incidentID, next, actor.ID)
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

func (a *Application) TransitionIncidentLifecycle(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	action string,
	request IncidentLifecycleRequest,
	requestHash []byte,
	requestID string,
	now time.Time,
) (IncidentLifecycleResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey:    "incidents." + action,
		ActorUserID: actor.ID,
		ScopeKey:    incidentID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := a.authStore.GetRouteIdempotency(ctx, key); err == nil {
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

	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IncidentLifecycleResult{}, fmt.Errorf("begin incident lifecycle transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txRepository := newRepository(tx)
	current, err := txRepository.getIncidentForUpdate(ctx, incidentID)
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
	var nextClosedAt *time.Time
	switch action {
	case "close":
		if current.Status != "active" {
			return IncidentLifecycleResult{}, ErrIncidentIllegalTransition
		}
		nextStatus = "closed"
		closedAt := now
		nextClosedAt = &closedAt
	case "reopen":
		if current.Status != "closed" {
			return IncidentLifecycleResult{}, ErrIncidentIllegalTransition
		}
		nextStatus = "active"
	default:
		return IncidentLifecycleResult{}, ErrIncidentIllegalTransition
	}

	updated, err := txRepository.updateIncidentLifecycle(ctx, updateIncidentLifecyclePersistenceParams{
		IncidentID:      incidentID,
		Status:          nextStatus,
		ClosedAt:        nextClosedAt,
		UpdatedAt:       now,
		UpdatedByUserID: actor.ID,
	})
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
	if err := authn.InsertRouteIdempotency(
		ctx,
		tx,
		key,
		&actor.ID,
		requestHash,
		http.StatusOK,
		responseJSON,
	); err != nil {
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
