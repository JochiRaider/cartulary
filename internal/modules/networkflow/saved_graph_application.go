package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type savedGraphOutcomeKind uint8

const (
	savedGraphAccepted savedGraphOutcomeKind = iota + 1
	savedGraphChanged
	savedGraphRetired
)

type savedGraphOutcome struct {
	kind        savedGraphOutcomeKind
	declaration graphViewDeclaration
	jobID       uuid.UUID
	// Historical receipts preserve the originally committed response.
	receipt map[string]any
}
type savedGraphApplication struct {
	store          *store
	incidentAccess incidentAdmissionChecker
	receipts       savedGraphReceiptAdapter
	graphViewJobs  GraphViewJobTransactions
	jobRunner      GraphViewJobRunner
	now            func() time.Time
}

func (s *savedGraphApplication) commitGraphViewCreate(ctx context.Context, incidentID, actorUserID uuid.UUID, request graphViewCreateRequest, requestID string) (savedGraphOutcome, error) {
	if _, err := s.incidentAccess.Check(ctx, incidentID, actorUserID, admission.Requirement{AllowedRoles: admission.RolesEditorAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
		return savedGraphOutcome{}, err
	}
	if s.graphViewJobs == nil {
		return savedGraphOutcome{}, errors.New("graph view jobs unavailable")
	}
	displayName, err := normalizeGraphViewDisplayName(request.DisplayName)
	if err != nil {
		return savedGraphOutcome{}, err
	}
	normalizedRequest := graphViewMutationBytes(routeKeyGraphViewsCreate, "graph-views", map[string]any{
		"display_name": displayName, "semantic_query": request.Semantic.Raw,
	})
	requestHash := sha256Bytes(normalizedRequest)
	key := graphViewIdempotencyKey(routeKeyGraphViewsCreate, actorUserID, incidentID, "graph-views", request.ClientTxnID)
	if outcome, replayed, err := s.receipts.replay(ctx, key, requestHash); replayed || err != nil {
		return outcome, err
	}
	var outcome savedGraphOutcome
	var jobID uuid.UUID
	err = withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.store.lockIncidentTx(ctx, tx, incidentID); err != nil {
			return err
		}
		if _, err := s.incidentAccess.CheckTx(ctx, tx, incidentID, actorUserID, admission.Requirement{AllowedRoles: admission.RolesEditorAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
			return err
		}
		counts, err := s.store.CountGraphViewDeclarationsTx(ctx, tx, incidentID, s.store.limits.MaxRetainedGraphViewsPerIncident)
		if err != nil {
			return err
		}
		if counts.Active >= s.store.limits.MaxActiveGraphViewsPerIncident {
			return errGraphViewActiveLimit
		}
		if counts.Retained >= s.store.limits.MaxRetainedGraphViewsPerIncident {
			return errGraphViewRetainedLimit
		}
		if err := s.requireGraphViewJobCapacityTx(ctx, tx, incidentID); err != nil {
			return err
		}
		tables, err := graphViewSourceTablesTx(ctx, tx, incidentID, request.Semantic.SelectedTableIDs)
		if err != nil {
			return err
		}
		selectedTableIDs := make([]string, 0, len(tables))
		for _, table := range tables {
			selectedTableIDs = append(selectedTableIDs, table.TableID)
		}
		semantic := request.Semantic
		semantic.SelectedTableIDs = selectedTableIDs
		semantic.Raw = graphSemanticQueryResource(selectedTableIDs, semantic.Filters, semantic.TimeRange, semantic.Aggregation)
		semanticJSON := canonicalJSON(semantic.Raw)
		snapshotID, err := graphViewSourceSnapshotTx(ctx, tx, incidentID, semantic)
		if err != nil {
			return err
		}
		graphViewID, err := newGraphViewID()
		if err != nil {
			return err
		}
		now := s.now().UTC()
		declaration := graphViewDeclaration{
			GraphViewID: graphViewID, IncidentID: incidentID, DisplayName: displayName,
			NormalizedDisplayName: strings.ToLower(displayName), DeclarationState: graphViewDeclarationStateActive,
			SemanticQueryJSON: semanticJSON, SemanticQuerySHA256: graphViewSemanticQuerySHA256(semanticJSON),
			DesiredSourceSnapshotID: snapshotID, GraphViewVersion: 1, MaterializationGeneration: 1,
			CreatedByUserID: actorUserID, CreatedAt: now, UpdatedAt: now,
		}
		if err := s.store.InsertGraphViewDeclarationTx(ctx, tx, declaration); err != nil {
			return err
		}
		job, err := s.enqueueGraphViewMaterializationTx(ctx, tx, key, normalizedRequest, declaration, actorUserID, now)
		if err != nil {
			return err
		}
		jobID = uuid.MustParse(job.JobID)
		declaration, err = s.store.SetGraphViewLatestJobTx(ctx, tx, incidentID, graphViewID, jobID)
		if err != nil {
			return err
		}
		outcome = savedGraphOutcome{kind: savedGraphAccepted, declaration: declaration, jobID: jobID}
		if err := s.receipts.saveTx(ctx, tx, key, requestHash, outcome); err != nil {
			return err
		}
		return s.appendGraphViewAuditTx(ctx, tx, "network_flow_graph_view_created", actorUserID, incidentID, graphViewID, request.ClientTxnID, requestID, 1, 1)
	})
	if err != nil {
		if outcome, replayed, err := s.receipts.replay(ctx, key, requestHash); replayed || err != nil {
			return outcome, err
		}
		return savedGraphOutcome{}, err
	}
	if s.jobRunner != nil {
		s.jobRunner.Notify(jobID)
	}
	return outcome, nil
}

func (s *savedGraphApplication) commitGraphViewRefresh(ctx context.Context, incidentID uuid.UUID, graphViewID string, actorUserID uuid.UUID, request graphViewVersionRequest, requestID string) (savedGraphOutcome, error) {
	if _, err := s.incidentAccess.Check(ctx, incidentID, actorUserID, admission.Requirement{AllowedRoles: admission.RolesEditorAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
		return savedGraphOutcome{}, err
	}
	if s.graphViewJobs == nil {
		return savedGraphOutcome{}, errors.New("graph view jobs unavailable")
	}
	normalizedRequest := graphViewMutationBytes(routeKeyGraphViewsRefresh, "graph_view_id:"+graphViewID, map[string]any{"base_graph_view_version": request.BaseGraphViewVersion})
	requestHash := sha256Bytes(normalizedRequest)
	key := graphViewIdempotencyKey(routeKeyGraphViewsRefresh, actorUserID, incidentID, graphViewID, request.ClientTxnID)
	if outcome, replayed, err := s.receipts.replay(ctx, key, requestHash); replayed || err != nil {
		return outcome, err
	}
	var outcome savedGraphOutcome
	var jobID uuid.UUID
	err := withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.store.lockIncidentTx(ctx, tx, incidentID); err != nil {
			return err
		}
		if _, err := s.incidentAccess.CheckTx(ctx, tx, incidentID, actorUserID, admission.Requirement{AllowedRoles: admission.RolesEditorAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
			return err
		}
		declaration, err := s.store.GetGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, true)
		if err != nil {
			return err
		}
		if declaration.DeclarationState != graphViewDeclarationStateActive {
			return errGraphViewDeclarationNotActive
		}
		if declaration.GraphViewVersion != request.BaseGraphViewVersion {
			return &graphViewVersionConflictError{Current: declaration.GraphViewVersion, Base: request.BaseGraphViewVersion}
		}
		if err := s.requireGraphViewJobCapacityTx(ctx, tx, incidentID); err != nil {
			return err
		}
		semantic, apiErr := decodeGraphSemanticRequest(declaration.SemanticQueryJSON, s.store.limits)
		if apiErr != nil {
			return errGraphViewDeclarationInvalid
		}
		snapshotID, err := graphViewSourceSnapshotTx(ctx, tx, incidentID, semantic)
		if err != nil {
			return err
		}
		now := s.now().UTC()
		declaration, err = s.store.RefreshGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, request.BaseGraphViewVersion, snapshotID, now)
		if err != nil {
			return err
		}
		job, err := s.enqueueGraphViewMaterializationTx(ctx, tx, key, normalizedRequest, declaration, actorUserID, now)
		if err != nil {
			return err
		}
		jobID = uuid.MustParse(job.JobID)
		declaration, err = s.store.SetGraphViewLatestJobTx(ctx, tx, incidentID, graphViewID, jobID)
		if err != nil {
			return err
		}
		outcome = savedGraphOutcome{kind: savedGraphAccepted, declaration: declaration, jobID: jobID}
		if err := s.receipts.saveTx(ctx, tx, key, requestHash, outcome); err != nil {
			return err
		}
		return s.appendGraphViewAuditTx(ctx, tx, "network_flow_graph_view_refresh_requested", actorUserID, incidentID, graphViewID, request.ClientTxnID, requestID, declaration.GraphViewVersion, declaration.MaterializationGeneration)
	})
	if err != nil {
		if outcome, replayed, err := s.receipts.replay(ctx, key, requestHash); replayed || err != nil {
			return outcome, err
		}
		return savedGraphOutcome{}, err
	}
	if s.jobRunner != nil {
		s.jobRunner.Notify(jobID)
	}
	return outcome, nil
}

func (s *savedGraphApplication) commitGraphViewRename(ctx context.Context, incidentID uuid.UUID, graphViewID string, actorUserID uuid.UUID, request graphViewRenameRequest, requestID string) (savedGraphOutcome, error) {
	if _, err := s.incidentAccess.Check(ctx, incidentID, actorUserID, admission.Requirement{AllowedRoles: admission.RolesEditorAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
		return savedGraphOutcome{}, err
	}
	displayName, err := normalizeGraphViewDisplayName(request.DisplayName)
	if err != nil {
		return savedGraphOutcome{}, err
	}
	normalizedRequest := graphViewMutationBytes(routeKeyGraphViewsPatch, "graph_view_id:"+graphViewID, map[string]any{"base_graph_view_version": request.BaseGraphViewVersion, "display_name": displayName})
	requestHash := sha256Bytes(normalizedRequest)
	key := graphViewIdempotencyKey(routeKeyGraphViewsPatch, actorUserID, incidentID, graphViewID, request.ClientTxnID)
	if outcome, replayed, err := s.receipts.replay(ctx, key, requestHash); replayed || err != nil {
		return outcome, err
	}
	var outcome savedGraphOutcome
	err = withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.store.lockIncidentTx(ctx, tx, incidentID); err != nil {
			return err
		}
		if _, err := s.incidentAccess.CheckTx(ctx, tx, incidentID, actorUserID, admission.Requirement{AllowedRoles: admission.RolesEditorAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
			return err
		}
		declaration, err := s.store.RenameGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, request.BaseGraphViewVersion, displayName, strings.ToLower(displayName), s.now())
		if err != nil {
			return err
		}
		outcome = savedGraphOutcome{kind: savedGraphChanged, declaration: declaration}
		if err := s.receipts.saveTx(ctx, tx, key, requestHash, outcome); err != nil {
			return err
		}
		if declaration.GraphViewVersion == request.BaseGraphViewVersion {
			return nil
		}
		return s.appendGraphViewAuditTx(ctx, tx, "network_flow_graph_view_renamed", actorUserID, incidentID, graphViewID, request.ClientTxnID, requestID, declaration.GraphViewVersion, declaration.MaterializationGeneration)
	})
	if err != nil {
		if outcome, replayed, err := s.receipts.replay(ctx, key, requestHash); replayed || err != nil {
			return outcome, err
		}
		return savedGraphOutcome{}, err
	}
	return outcome, nil
}

func (s *savedGraphApplication) commitGraphViewRetire(ctx context.Context, incidentID uuid.UUID, graphViewID string, actorUserID uuid.UUID, request graphViewVersionRequest, requestID string) (savedGraphOutcome, error) {
	if _, err := s.incidentAccess.Check(ctx, incidentID, actorUserID, admission.Requirement{AllowedRoles: admission.RolesReviewerAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
		return savedGraphOutcome{}, err
	}
	normalizedRequest := graphViewMutationBytes(routeKeyGraphViewsDelete, "graph_view_id:"+graphViewID, map[string]any{"base_graph_view_version": request.BaseGraphViewVersion})
	requestHash := sha256Bytes(normalizedRequest)
	key := graphViewIdempotencyKey(routeKeyGraphViewsDelete, actorUserID, incidentID, graphViewID, request.ClientTxnID)
	if outcome, replayed, err := s.receipts.replay(ctx, key, requestHash); replayed || err != nil {
		return outcome, err
	}
	var outcome savedGraphOutcome
	err := withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.store.lockIncidentTx(ctx, tx, incidentID); err != nil {
			return err
		}
		if _, err := s.incidentAccess.CheckTx(ctx, tx, incidentID, actorUserID, admission.Requirement{AllowedRoles: admission.RolesReviewerAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
			return err
		}
		declaration, err := s.store.RetireGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, request.BaseGraphViewVersion, s.now())
		if err != nil {
			return err
		}
		outcome = savedGraphOutcome{kind: savedGraphRetired}
		if err := s.receipts.saveTx(ctx, tx, key, requestHash, outcome); err != nil {
			return err
		}
		return s.appendGraphViewAuditTx(ctx, tx, "network_flow_graph_view_retired", actorUserID, incidentID, graphViewID, request.ClientTxnID, requestID, declaration.GraphViewVersion, declaration.MaterializationGeneration)
	})
	if err != nil {
		if outcome, replayed, err := s.receipts.replay(ctx, key, requestHash); replayed || err != nil {
			return outcome, err
		}
		return savedGraphOutcome{}, err
	}
	return outcome, nil
}

func (s *savedGraphApplication) enqueueGraphViewMaterializationTx(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, normalizedRequest []byte, declaration graphViewDeclaration, actorUserID uuid.UUID, now time.Time) (jobs.Resource, error) {
	scope := jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &declaration.IncidentID}
	admission, err := jobs.NewExtensionJobAdmission(ProfileID, jobs.NewRouteIdempotencyKey(key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID), scope, normalizedRequest)
	if err != nil {
		return jobs.Resource{}, err
	}
	handlerPayload, err := json.Marshal(graphViewMaterializationPayload{
		SchemaID:   "cartulary.network_flow.graph_view_materialization_payload.v1",
		IncidentID: declaration.IncidentID, GraphViewID: declaration.GraphViewID,
		MaterializationGeneration: declaration.MaterializationGeneration,
		SourceSnapshotID:          declaration.DesiredSourceSnapshotID,
	})
	if err != nil {
		return jobs.Resource{}, err
	}
	total := 1
	return s.graphViewJobs.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{
		JobKind: GraphViewMaterializationJobKind, Scope: scope, SubmittedByUserID: actorUserID,
		AuthPolicy: jobs.AuthPolicyIncidentMembership, Cancelable: true,
		Progress: jobs.Progress{Completed: 0, Total: &total}, HandlerPayload: handlerPayload, Extension: admission,
	}, now.UTC())
}

func (s *savedGraphApplication) requireGraphViewJobCapacityTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	count, err := s.graphViewJobs.CountNonterminalIncidentJobsTx(ctx, tx, incidentID, GraphViewMaterializationJobKind, s.store.limits.MaxNonterminalGraphJobsPerIncident)
	if err != nil {
		return err
	}
	if count >= s.store.limits.MaxNonterminalGraphJobsPerIncident {
		return errGraphViewJobLimit
	}
	return nil
}

func (s *savedGraphApplication) appendGraphViewAuditTx(ctx context.Context, tx pgx.Tx, eventKind string, actorUserID, incidentID uuid.UUID, graphViewID, clientTxnID, requestID string, version, generation int64) error {
	return s.store.appendAuditEventTx(ctx, tx, networkFlowAuditEvent{
		ActorUserID: &actorUserID, IncidentID: &incidentID, EventKind: eventKind,
		ClientTxnID: optionalStringPtr(clientTxnID), RequestID: optionalStringPtr(requestID),
		AfterJSON: map[string]any{
			"incident_id": incidentID.String(), "actor_user_id": actorUserID.String(),
			"graph_view_id": graphViewID, "graph_view_version": version,
			"materialization_generation":    generation,
			"network_flow.audit_event_code": eventKind, "network_flow.audit_resource_id": graphViewID,
		},
	})
}

type graphViewCreateRequest struct {
	ClientTxnID string
	DisplayName string
	Semantic    graphSemanticRequest
}

type graphViewVersionRequest struct {
	ClientTxnID          string
	BaseGraphViewVersion int64
}

type graphViewRenameRequest struct {
	graphViewVersionRequest
	DisplayName string
}

var errGraphViewJobLimit = errors.New("network flow graph materialization job limit exceeded")
