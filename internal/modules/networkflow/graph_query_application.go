package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type graphQueryComposer interface {
	composeGraph(context.Context, uuid.UUID, graphQueryRequest) (graphComposition, *semanticFailure)
	queryGraphContributorPage(context.Context, uuid.UUID, graphSemanticRequest, string, graphSelector, *contributorCursorPosition, int) ([]flowRow, bool, map[string]int, *semanticFailure)
}
type graphQueryApplication struct {
	store           *store
	incidentAccess  incidentAdmissionChecker
	composer        graphQueryComposer
	cursorProtector cursorProtector
	safeDigester    safeDigester
}
type graphContributorsOutcome struct {
	Rows      []flowRow
	Digest    string
	Selector  graphSelector
	Limit     int
	NextToken *string
}

func (s *graphQueryApplication) execute(ctx context.Context, who readIdentity, request graphQueryRequest, correlation string) (graphComposition, *semanticFailure) {
	if failure := admitRead(ctx, s.incidentAccess, who); failure != nil {
		return graphComposition{}, failure
	}
	composition, failure := s.composer.composeGraph(ctx, who.IncidentID, request)
	if failure != nil {
		return graphComposition{}, failure
	}
	if failure := s.recordGraphQueryAudit(ctx, who.IncidentID, who.ActorID, composition, correlation); failure != nil {
		return graphComposition{}, failure
	}
	// Commitment completes this occurrence; transport maintenance/delivery is independent.
	return composition, nil
}
func (s *graphQueryApplication) contributors(ctx context.Context, who readIdentity, request graphContributorQueryRequest) (*graphContributorsOutcome, *semanticFailure) {
	if failure := admitRead(ctx, s.incidentAccess, who); failure != nil {
		return nil, failure
	}
	return s.queryGraphContributors(ctx, who.ActorID.String(), who.SessionID.String(), who.IncidentID, request)
}
func (s *graphQueryApplication) queryGraphContributors(ctx context.Context, actorID string, sessionID string, incidentID uuid.UUID, request graphContributorQueryRequest) (*graphContributorsOutcome, *semanticFailure) {
	var position *contributorCursorPosition
	limit := request.Limit
	if request.Continuation {
		payload, reason := s.cursorProtector.Decode(request.CursorToken)
		if reason != "" {
			return nil, cursorInvalid(reason)
		}
		if payload.Route != routeKeyGraphsContributorsQuery || payload.ActorUserID != actorID || payload.SessionID != sessionID || payload.IncidentID != incidentID.String() {
			return nil, cursorInvalid(payloadMismatchReason(payload, routeKeyGraphsContributorsQuery, actorID, incidentID.String()))
		}
		limit = payload.Limit
		if payload.PositionKind != "contributor_keyset_v1" {
			return nil, cursorInvalid("malformed")
		}
		decodedPosition, err := decodeContributorCursorPosition(payload.Position)
		if err != nil {
			return nil, cursorInvalid("malformed")
		}
		if !sameSortSpecs(decodedPosition.Row.EffectiveSort, effectiveSort(nil)) {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
		position = &decodedPosition
		var echo struct {
			GraphQuery       graphSemanticRequest `json:"-"`
			GraphQueryDigest string               `json:"graph_query_digest"`
			Selector         graphSelector        `json:"selector"`
		}
		var rawEcho map[string]json.RawMessage
		if err := json.Unmarshal(payload.QueryEcho, &rawEcho); err != nil {
			return nil, cursorInvalid("malformed")
		}
		semantic, apiErr := decodeGraphSemanticRequest(rawEcho["graph_query"], s.store.limits)
		if apiErr != nil {
			return nil, cursorInvalid("malformed")
		}
		echo.GraphQuery = semantic
		if err := json.Unmarshal(rawEcho["graph_query_digest"], &echo.GraphQueryDigest); err != nil {
			return nil, cursorInvalid("malformed")
		}
		if err := json.Unmarshal(rawEcho["selector"], &echo.Selector); err != nil {
			return nil, cursorInvalid("malformed")
		}
		request.GraphQuery = echo.GraphQuery
		request.GraphQueryDigest = echo.GraphQueryDigest
		request.Selector = echo.Selector
		request.Limit = limit
		if payload.QueryHash != queryHash(graphContributorQueryEcho(request)) {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
	}
	rows, hasMore, tableRanks, apiErr := s.composer.queryGraphContributorPage(ctx, incidentID, request.GraphQuery, request.GraphQueryDigest, request.Selector, position, limit)
	if apiErr != nil {
		return nil, apiErr
	}
	queryEcho := graphContributorQueryEcho(request)
	queryEchoRaw, _ := json.Marshal(queryEcho)
	var nextToken *string
	if hasMore && len(rows) > 0 {
		token, err := s.cursorProtector.Encode(cursorBinding{
			Route:       routeKeyGraphsContributorsQuery,
			ActorUserID: actorID,
			SessionID:   sessionID,
			IncidentID:  incidentID.String(),
			Scope:       map[string]string{"graph_query_digest": request.GraphQueryDigest},
			QueryHash:   queryHash(queryEcho),
			QueryEcho:   queryEchoRaw,
			Limit:       limit,
		}, "contributor_keyset_v1", newContributorCursorPosition(rows[len(rows)-1], tableRanks))
		if err != nil {
			return nil, internalSemanticFailure(err)
		}
		nextToken = &token
	}
	return &graphContributorsOutcome{Rows: rows, Digest: request.GraphQueryDigest, Selector: request.Selector, Limit: limit, NextToken: nextToken}, nil
}

func (s *graphQueryApplication) recordGraphQueryAudit(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, composition graphComposition, requestID string) *semanticFailure {
	graphDigestSafe, keyID, err := s.safeDigester.Digest("graph_query_digest", composition.Digest)
	if err != nil {
		return internalSemanticFailure(err)
	}
	truncatedCount := 0
	for _, edge := range composition.Edges {
		limit := composition.ResultLimits.MaxExampleRowRefsPerEdge
		if !composition.IncludeExamples {
			limit = 0
		}
		if edge.FlowRowCount > limit {
			truncatedCount += edge.FlowRowCount - limit
		}
	}
	if err := ctx.Err(); err != nil {
		return graphProjectionFailedForContext(err)
	}
	err = withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if _, err := s.incidentAccess.CheckTx(ctx, tx, incidentID, actorUserID, admission.Requirement{AllowedRoles: admission.RolesMember, Lifecycle: admission.LifecycleOpen}); err != nil {
			return applicationAdmissionFailure(err, "")
		}

		return s.store.appendAuditEventTx(ctx, tx, networkFlowAuditEvent{
			ActorUserID: &actorUserID,
			IncidentID:  &incidentID,
			EventKind:   "network_flow_graph_query_executed",
			RequestID:   optionalStringPtr(requestID),
			AfterJSON: map[string]any{
				"incident_id":                    incidentID.String(),
				"actor_user_id":                  actorUserID.String(),
				"graph_query_digest_safe":        graphDigestSafe,
				"graph_query_digest_safe_key_id": keyID,
				"selected_table_count":           len(composition.SelectedTableIDs),
				"result_vertex_count":            len(composition.Vertices),
				"result_edge_count":              len(composition.Edges),
				"truncated_example_ref_count":    truncatedCount,
				"network_flow.audit_event_code":  "network_flow_graph_query_executed",
				"network_flow.audit_resource_id": composition.Digest,
			},
		})
	})
	if err != nil {
		var failure *semanticFailure
		if errors.As(err, &failure) {
			return failure
		}
		return internalSemanticFailure(fmt.Errorf("record network flow graph query audit: %w", err))
	}
	return nil
}
