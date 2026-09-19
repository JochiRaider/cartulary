package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

type savedGraphDeclarations interface {
	ListActiveGraphViewDeclarations(context.Context, uuid.UUID) ([]graphViewDeclaration, error)
	GetGraphViewDeclaration(context.Context, uuid.UUID, string) (graphViewDeclaration, error)
}
type savedGraphResults interface {
	ReadExactResult(context.Context, graphprojection.ResultBindingV2) (graphprojection.CompletedResultV2, error)
}
type savedGraphJobs interface {
	Get(context.Context, uuid.UUID) (jobs.Resource, error)
}
type savedGraphReadApplication struct {
	declarations    savedGraphDeclarations
	results         savedGraphResults
	jobs            savedGraphJobs
	incidentAccess  incidentAdmissionChecker
	composer        *graphSourceComposer
	verifier        *graphSourceComposer
	cursorProtector cursorProtector
	limits          EffectiveLimits
}
type savedGraphReadOutcome struct {
	Declaration graphViewDeclaration
	Composition graphComposition
}
type savedGraphContributorsOutcome struct {
	GraphViewID, ProjectionResultID string
	Page                            graphContributorsOutcome
}
type missingSelectedResultReason string

const (
	selectedResultPending missingSelectedResultReason = "initial_materialization_pending"
	selectedResultFailed  missingSelectedResultReason = "initial_materialization_failed"
)

func (s *savedGraphReadApplication) list(ctx context.Context, who readIdentity) ([]graphViewDeclaration, *semanticFailure) {
	if failure := admitRead(ctx, s.incidentAccess, who); failure != nil {
		return nil, failure
	}
	records, err := s.declarations.ListActiveGraphViewDeclarations(ctx, who.IncidentID)
	if err != nil {
		return nil, internalSemanticFailure(err)
	}
	return records, nil
}
func (s *savedGraphReadApplication) get(ctx context.Context, who readIdentity, id string) (graphViewDeclaration, *semanticFailure) {
	if failure := admitRead(ctx, s.incidentAccess, who); failure != nil {
		return graphViewDeclaration{}, failure
	}
	declaration, err := s.declarations.GetGraphViewDeclaration(ctx, who.IncidentID, id)
	if errors.Is(err, errGraphViewDeclarationNotFound) || err == nil && declaration.DeclarationState != graphViewDeclarationStateActive {
		return graphViewDeclaration{}, newSemanticFailure(failureGraphViewNotFound, "graph_view_id", "not_found")
	}
	if err != nil {
		return graphViewDeclaration{}, internalSemanticFailure(err)
	}
	return declaration, nil
}
func (s *savedGraphReadApplication) missingSelectedResult(ctx context.Context, declaration graphViewDeclaration) (missingSelectedResultReason, *semanticFailure) {
	if declaration.SelectedResult != nil {
		return "", internalSemanticFailure(errors.New("selected result classification requires absence"))
	}
	retained := func() missingSelectedResultReason {
		if declaration.LastFailureCode != nil {
			return selectedResultFailed
		}
		return selectedResultPending
	}
	if declaration.LatestJobID == nil {
		return retained(), nil
	}
	job, err := s.jobs.Get(ctx, *declaration.LatestJobID)
	if errors.Is(err, jobs.ErrNotFound) {
		return retained(), nil
	}
	if err != nil {
		return "", internalSemanticFailure(fmt.Errorf("lookup saved graph materialization: %w", err))
	}
	switch job.Status {
	case jobs.StatusQueued, jobs.StatusRunning, jobs.StatusCancelRequested:
		return selectedResultPending, nil
	case jobs.StatusFailed, jobs.StatusCanceled:
		return selectedResultFailed, nil
	default:
		return "", internalSemanticFailure(errors.New("saved graph job and selected result are inconsistent"))
	}
}
func (s *savedGraphReadApplication) selected(ctx context.Context, who readIdentity, id string) (graphViewDeclaration, *semanticFailure) {
	declaration, failure := s.get(ctx, who, id)
	if failure != nil {
		return graphViewDeclaration{}, failure
	}
	if declaration.SelectedResult == nil {
		reason, failure := s.missingSelectedResult(ctx, declaration)
		if failure != nil {
			return graphViewDeclaration{}, failure
		}
		return graphViewDeclaration{}, newSemanticFailure(failureGraphViewNotMaterialized, "graph_view_id", string(reason))
	}
	return declaration, nil
}
func (s *savedGraphReadApplication) result(ctx context.Context, who readIdentity, id string) (*savedGraphReadOutcome, *semanticFailure) {
	declaration, failure := s.selected(ctx, who, id)
	if failure != nil {
		return nil, failure
	}
	completed, err := s.results.ReadExactResult(ctx, graphViewResultBinding(declaration))
	if err != nil {
		return nil, internalSemanticFailure(err)
	}
	var result map[string]any
	if err := json.Unmarshal(completed.ResultJSON, &result); err != nil {
		return nil, internalSemanticFailure(err)
	}
	semantic, failure := decodeGraphSemanticRequest(declaration.SemanticQueryJSON, s.limits)
	if failure != nil {
		return nil, malformedStoredGraphResult()
	}
	composition, failure := s.verifier.composeGraphSourceFromSemantic(ctx, who.IncidentID, semantic)
	if failure != nil || graphSourceSnapshotDigest(who.IncidentID, composition.SourceTables, composition.Digest) != declaration.SelectedResult.SourceSnapshotID {
		failureOut := malformedStoredGraphResult()
		if failure != nil {
			failureOut.cause = failure
		}
		return nil, failureOut
	}
	composition.GraphProjection = result
	if semantic.Aggregation.Mode == "time_bucket_v1" {
		composition.TimeBuckets, failure = deriveTimeBucketIndexFromExactResult(semantic.TimeRange, semantic.Aggregation.BucketWidthSeconds, composition.ResultLimits.MaxTimeBuckets, result)
		if failure != nil {
			return nil, failure
		}
	}
	if failure := bindGraphV2ResponseMetadata(&composition); failure != nil {
		return nil, failure
	}
	return &savedGraphReadOutcome{Declaration: declaration, Composition: composition}, nil
}
func (s *savedGraphReadApplication) contributors(ctx context.Context, who readIdentity, id string, request graphViewContributorRequest) (*savedGraphContributorsOutcome, *semanticFailure) {
	declaration, failure := s.selected(ctx, who, id)
	if failure != nil {
		return nil, failure
	}
	// Contributors are bound to the exact immutable selected envelope as well as
	// its semantic/source identity; a matching ID alone cannot authorize rows.
	if _, err := s.results.ReadExactResult(ctx, graphViewResultBinding(declaration)); err != nil {
		return nil, internalSemanticFailure(err)
	}
	semantic, failure := decodeGraphSemanticRequest(declaration.SemanticQueryJSON, s.limits)
	if failure != nil {
		return nil, malformedStoredGraphResult()
	}
	digest := graphQueryDigestForSemantic(who.IncidentID, semantic.SelectedTableIDs, semantic)
	return s.querySavedGraphContributors(ctx, who.ActorID.String(), who.SessionID.String(), who.IncidentID, id, declaration.SelectedResult.ProjectionResultID, semantic, digest, request)
}
func (s *savedGraphReadApplication) querySavedGraphContributors(
	ctx context.Context,
	actorID string,
	sessionID string,
	incidentID uuid.UUID,
	graphViewID string,
	selectedResultID string,
	semantic graphSemanticRequest,
	digest string,
	request graphViewContributorRequest,
) (*savedGraphContributorsOutcome, *semanticFailure) {
	var position *contributorCursorPosition
	limit := request.Limit
	if request.Continuation {
		payload, reason := s.cursorProtector.Decode(request.CursorToken)
		if reason != "" {
			return nil, cursorInvalid(reason)
		}
		if payload.Route != routeKeyGraphViewContributorsQuery || payload.ActorUserID != actorID || payload.SessionID != sessionID || payload.IncidentID != incidentID.String() {
			return nil, cursorInvalid(payloadMismatchReason(payload, routeKeyGraphViewContributorsQuery, actorID, incidentID.String()))
		}
		if payload.Scope["graph_view_id"] != graphViewID || payload.Scope["projection_result_id"] != selectedResultID || payload.Scope["graph_query_digest"] != digest {
			return nil, cursorInvalid("scope_stale")
		}
		if payload.PositionKind != "contributor_keyset_v1" {
			return nil, cursorInvalid("malformed")
		}
		decodedPosition, err := decodeContributorCursorPosition(payload.Position)
		if err != nil || !sameSortSpecs(decodedPosition.Row.EffectiveSort, effectiveSort(nil)) {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
		position = &decodedPosition
		limit = payload.Limit
		var echo struct {
			GraphViewID        string          `json:"graph_view_id"`
			ProjectionResultID string          `json:"projection_result_id"`
			GraphQueryDigest   string          `json:"graph_query_digest"`
			Selector           json.RawMessage `json:"selector"`
		}
		if err := json.Unmarshal(payload.QueryEcho, &echo); err != nil || echo.GraphViewID != graphViewID || echo.ProjectionResultID != selectedResultID || echo.GraphQueryDigest != digest {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
		selector, apiErr := decodeGraphSelector(echo.Selector)
		if apiErr != nil {
			return nil, cursorInvalid("malformed")
		}
		request.ProjectionResultID = echo.ProjectionResultID
		request.Selector = selector
		request.Limit = limit
		if payload.QueryHash != queryHash(savedGraphContributorQueryEcho(graphViewID, digest, request)) {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
	} else if request.ProjectionResultID != selectedResultID {
		return nil, graphQueryStale("digest_mismatch", request.ProjectionResultID)
	}

	rows, hasMore, tableRanks, apiErr := s.composer.queryGraphContributorPage(ctx, incidentID, semantic, digest, request.Selector, position, limit)
	if apiErr != nil {
		return nil, apiErr
	}
	echo := savedGraphContributorQueryEcho(graphViewID, digest, request)
	echoRaw, _ := json.Marshal(echo)
	var nextToken *string
	if hasMore && len(rows) > 0 {
		token, err := s.cursorProtector.Encode(cursorBinding{
			Route: routeKeyGraphViewContributorsQuery, ActorUserID: actorID, SessionID: sessionID, IncidentID: incidentID.String(),
			Scope: map[string]string{
				"graph_view_id": graphViewID, "projection_result_id": selectedResultID, "graph_query_digest": digest,
			},
			QueryHash: queryHash(echo), QueryEcho: echoRaw, Limit: limit,
		}, "contributor_keyset_v1", newContributorCursorPosition(rows[len(rows)-1], tableRanks))
		if err != nil {
			return nil, internalSemanticFailure(err)
		}
		nextToken = &token
	}
	return &savedGraphContributorsOutcome{GraphViewID: graphViewID, ProjectionResultID: selectedResultID, Page: graphContributorsOutcome{Rows: rows, Selector: request.Selector, Limit: limit, NextToken: nextToken}}, nil
}

func savedGraphContributorQueryEcho(graphViewID string, digest string, request graphViewContributorRequest) map[string]any {
	return map[string]any{
		"graph_view_id": graphViewID, "projection_result_id": request.ProjectionResultID,
		"graph_query_digest": digest, "selector": graphSelectorResource(request.Selector),
	}
}
