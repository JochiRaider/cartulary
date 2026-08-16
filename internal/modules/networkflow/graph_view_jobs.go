package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const (
	GraphViewMaterializationJobKind = "network_flow_activity.graph_view_materialize_v1"
	GraphViewWorkerKind             = "network_flow_activity.graph_view_worker_v1"
	graphViewProgressUnitID         = "network_flow_activity.graph_view_materialize.projection_result.v1"
	graphViewResultResourceKind     = "network_flow_graph_view"
)

type GraphViewJobTransactions interface {
	CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error)
	CountNonterminalIncidentJobsTx(context.Context, pgx.Tx, uuid.UUID, string, int64) (int64, error)
}

type GraphViewJobManager interface {
	Get(context.Context, uuid.UUID) (jobs.Resource, error)
	ObserveExecution(context.Context, jobs.Execution) (jobs.Resource, error)
	HandlerPayload(context.Context, jobs.Execution) (json.RawMessage, error)
	RetainedHandlerPayload(context.Context, uuid.UUID) (json.RawMessage, error)
	CompleteCanceled(context.Context, jobs.Execution, jobs.CancellationCompletion) (jobs.Resource, error)
}

type GraphViewJobRunner interface {
	RegisterHandler(string, jobs.HandlerFunc) error
	Notify(uuid.UUID)
}

type GraphViewJobMutation func(context.Context, pgx.Tx) error

type GraphViewJobSuccessFinalization struct {
	Execution     jobs.Execution
	Completion    jobs.SuccessCompletion
	FinalCommitID string
	Mutate        GraphViewJobMutation
}

type GraphViewJobFailureFinalization struct {
	Execution  jobs.Execution
	Completion jobs.FailureCompletion
	Mutate     GraphViewJobMutation
}

type GraphViewJobFinalizer interface {
	FinalizeGraphViewJobSuccess(context.Context, GraphViewJobSuccessFinalization) (jobs.Resource, error)
	FinalizeGraphViewJobFailure(context.Context, GraphViewJobFailureFinalization) (jobs.Resource, error)
}

type graphViewMaterializationPayload struct {
	SchemaID                  string    `json:"schema_id"`
	IncidentID                uuid.UUID `json:"incident_id"`
	GraphViewID               string    `json:"graph_view_id"`
	MaterializationGeneration int64     `json:"materialization_generation"`
	SourceSnapshotID          string    `json:"source_snapshot_id"`
}

func (payload graphViewMaterializationPayload) valid() bool {
	return payload.SchemaID == "cartulary.network_flow.graph_view_materialization_payload.v1" &&
		payload.IncidentID != uuid.Nil && graphViewIDPattern.MatchString(payload.GraphViewID) &&
		payload.MaterializationGeneration > 0 && payload.SourceSnapshotID != ""
}

func (m *Module) handleGraphViewMaterialization(ctx context.Context, execution jobs.Execution) error {
	if m == nil || m.store == nil || m.jobManager == nil || m.jobFinalizer == nil || m.graphProjection == nil {
		return errors.New("network flow graph materialization unavailable")
	}
	materializationCtx, cancel := context.WithTimeout(ctx, time.Duration(m.limits.GraphMaterializationTimeoutSeconds)*time.Second)
	defer cancel()
	ctx = materializationCtx
	var payload graphViewMaterializationPayload
	job, err := m.jobManager.ObserveExecution(ctx, execution)
	if err != nil {
		if graphViewMaterializationTimedOut(ctx) {
			return m.failGraphViewMaterialization(context.WithoutCancel(ctx), execution, payload, "timeout", false)
		}
		return err
	}
	if job.Status == jobs.StatusCancelRequested {
		return m.cancelGraphViewMaterialization(ctx, execution)
	}
	rawPayload, err := m.jobManager.HandlerPayload(ctx, execution)
	if err != nil {
		if graphViewMaterializationTimedOut(ctx) {
			return m.failGraphViewMaterialization(context.WithoutCancel(ctx), execution, payload, "timeout", false)
		}
		return err
	}
	if err := json.Unmarshal(rawPayload, &payload); err != nil || !payload.valid() {
		return m.failGraphViewMaterialization(ctx, execution, payload, "source_invalid", false)
	}
	declaration, err := m.store.GetGraphViewDeclaration(ctx, payload.IncidentID, payload.GraphViewID)
	if graphViewMaterializationTimedOut(ctx) {
		return m.failGraphViewMaterialization(context.WithoutCancel(ctx), execution, payload, "timeout", false)
	}
	if err != nil || declaration.DeclarationState != GraphViewDeclarationStateActive ||
		declaration.MaterializationGeneration != payload.MaterializationGeneration ||
		declaration.DesiredSourceSnapshotID != payload.SourceSnapshotID || declaration.LatestJobID == nil ||
		*declaration.LatestJobID != execution.JobID() {
		return m.failGraphViewMaterialization(ctx, execution, payload, "publication_conflict", false)
	}
	semantic, apiErr := decodeGraphSemanticRequest(declaration.SemanticQueryJSON, m.limits)
	if apiErr != nil {
		return m.failGraphViewMaterialization(ctx, execution, payload, "source_invalid", false)
	}
	composer := &Service{store: m.store, graphProjection: m.graphProjection, now: m.now, graphTelemetry: m.graphTelemetry}
	composition, apiErr := composer.composeGraphSourceFromSemantic(ctx, payload.IncidentID, semantic)
	if graphViewMaterializationTimedOut(ctx) {
		return m.failGraphViewMaterialization(context.WithoutCancel(ctx), execution, payload, "timeout", false)
	}
	if apiErr != nil {
		return m.failGraphViewMaterialization(ctx, execution, payload, "source_invalid", false)
	}
	sourceSnapshotID := graphSourceSnapshotDigest(payload.IncidentID, composition.SourceTables, composition.Digest)
	if sourceSnapshotID != payload.SourceSnapshotID {
		return m.failGraphViewMaterialization(ctx, execution, payload, "source_invalid", false)
	}
	if graphViewMaterializationTimedOut(ctx) {
		return m.failGraphViewMaterialization(context.WithoutCancel(ctx), execution, payload, "timeout", false)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	projectionStarted := time.Now()
	projection, err := m.graphProjection.ProjectSaved(ctx, payload.GraphViewID, canonicalJSON(networkFlowProjectionInput(sourceSnapshotID, composition)), func(checkCtx context.Context) error {
		observed, observeErr := m.jobManager.ObserveExecution(checkCtx, execution)
		if observeErr != nil {
			return observeErr
		}
		if observed.Status == jobs.StatusCancelRequested {
			return jobs.ErrCancellationRequested
		}
		return nil
	})
	if err != nil {
		m.observeGraphPhase(ctx, graphTelemetryPhaseProjection, semantic.Aggregation.Mode, projectionStarted, err)
		if graphViewMaterializationTimedOut(ctx) {
			return m.failGraphViewMaterialization(context.WithoutCancel(ctx), execution, payload, "timeout", false)
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			if observed, observeErr := m.jobManager.ObserveExecution(context.WithoutCancel(ctx), execution); observeErr == nil && observed.Status == jobs.StatusCancelRequested {
				return m.cancelGraphViewMaterialization(context.WithoutCancel(ctx), execution)
			}
		}
		return err
	}
	completed, err := projection.CompletedResult()
	m.observeGraphPhase(ctx, graphTelemetryPhaseProjection, semantic.Aggregation.Mode, projectionStarted, err)
	if err != nil {
		return m.failGraphViewMaterialization(ctx, execution, payload, "projection_rejected", false)
	}
	m.observeGraphComposition(ctx, composition)
	if observed, observeErr := m.jobManager.ObserveExecution(ctx, execution); observeErr != nil {
		if graphViewMaterializationTimedOut(ctx) {
			return m.failGraphViewMaterialization(context.WithoutCancel(ctx), execution, payload, "timeout", false)
		}
		return observeErr
	} else if observed.Status == jobs.StatusCancelRequested {
		return m.cancelGraphViewMaterialization(context.WithoutCancel(ctx), execution)
	}
	if graphViewMaterializationTimedOut(ctx) {
		return m.failGraphViewMaterialization(context.WithoutCancel(ctx), execution, payload, "timeout", false)
	}
	completed.PublishedAt = m.now().UTC()
	selected := graphViewSelectedResultFromBinding(completed.Binding)
	total := 1
	publicationStarted := time.Now()
	_, err = m.jobFinalizer.FinalizeGraphViewJobSuccess(context.WithoutCancel(ctx), GraphViewJobSuccessFinalization{
		Execution: execution,
		Completion: jobs.SuccessCompletion{
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: jobs.ResultSummary{
				Code:    "network_flow_graph_view_materialized",
				Message: "Saved graph materialized.",
				ResourceRefs: []jobs.ResourceRef{{
					Kind:  graphViewResultResourceKind,
					ID:    payload.GraphViewID,
					Route: graphViewRoute(payload.IncidentID, payload.GraphViewID),
				}},
			},
		},
		FinalCommitID: completed.Binding.ProjectionResultID + ":" + execution.JobID().String(),
		Mutate: func(finalizeCtx context.Context, tx pgx.Tx) error {
			currentSourceSnapshot, sourceErr := m.graphViewSourceSnapshotTx(finalizeCtx, tx, payload.IncidentID, semantic)
			if sourceErr != nil || currentSourceSnapshot != payload.SourceSnapshotID {
				return ErrGraphViewPublicationStale
			}
			publisher, publisherErr := postgresresult.NewPublisher(tx)
			if publisherErr != nil {
				return publisherErr
			}
			if publisherErr = publisher.PublishResult(finalizeCtx, completed); publisherErr != nil {
				return publisherErr
			}
			_, publisherErr = m.store.PublishGraphViewResultTx(
				finalizeCtx, tx, payload.IncidentID, payload.GraphViewID,
				payload.MaterializationGeneration, payload.SourceSnapshotID,
				execution.JobID(), selected, completed.PublishedAt,
			)
			return publisherErr
		},
	})
	m.observeGraphPhase(context.WithoutCancel(ctx), graphTelemetryPhasePublication, semantic.Aggregation.Mode, publicationStarted, err)
	if errors.Is(err, jobs.ErrCancellationRequested) {
		return m.cancelGraphViewMaterialization(context.WithoutCancel(ctx), execution)
	}
	if errors.Is(err, ErrGraphViewPublicationStale) {
		return m.failGraphViewMaterialization(context.WithoutCancel(ctx), execution, payload, "publication_conflict", false)
	}
	return err
}

func (m *Module) graphViewSourceSnapshotTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, semantic graphSemanticRequest) (string, error) {
	return graphViewSourceSnapshotTx(ctx, tx, incidentID, semantic)
}

func graphViewSourceSnapshotTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, semantic graphSemanticRequest) (string, error) {
	tables, err := graphViewSourceTablesTx(ctx, tx, incidentID, semantic.SelectedTableIDs)
	if err != nil {
		return "", err
	}
	digest := graphQueryDigestForSemantic(incidentID, semantic.SelectedTableIDs, semantic)
	return graphSourceSnapshotDigest(incidentID, tables, digest), nil
}

func graphViewSourceTablesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, selectedTableIDs []string) ([]TableRecord, error) {
	if tx == nil || incidentID == uuid.Nil || len(selectedTableIDs) == 0 {
		return nil, ErrGraphViewDeclarationInvalid
	}
	rows, err := tx.Query(ctx, tableSelectColumns()+`
  FROM network_flow_tables
 WHERE incident_id = $1
   AND network_flow_table_id = ANY($2)
   AND table_status = 'active'
 ORDER BY created_at ASC, network_flow_table_id ASC
`, incidentID, selectedTableIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	selected := stringSet(selectedTableIDs)
	tables := make([]TableRecord, 0, len(selectedTableIDs))
	for rows.Next() {
		table, scanErr := scanTable(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if _, present := selected[table.TableID]; present {
			tables = append(tables, table)
			delete(selected, table.TableID)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(selected) != 0 || len(tables) != len(selectedTableIDs) {
		return nil, ErrGraphViewPublicationStale
	}
	return tables, nil
}

func (m *Module) failGraphViewMaterialization(ctx context.Context, execution jobs.Execution, payload graphViewMaterializationPayload, reasonCode string, retryable bool) error {
	if m.jobFinalizer == nil {
		return errors.New("network flow graph materialization finalizer unavailable")
	}
	total := 1
	_, err := m.jobFinalizer.FinalizeGraphViewJobFailure(ctx, GraphViewJobFailureFinalization{
		Execution: execution,
		Completion: jobs.FailureCompletion{
			Progress: jobs.Progress{Completed: 0, Total: &total},
			ErrorSummary: jobs.ErrorSummary{
				Code: "network_flow_graph_materialization_failed", Message: "Saved graph materialization failed.", Retryable: retryable,
				Details: map[string]any{"reason_code": reasonCode},
			},
		},
		Mutate: func(finalizeCtx context.Context, tx pgx.Tx) error {
			if !payload.valid() {
				return nil
			}
			return m.store.RecordGraphViewMaterializationFailureTx(
				finalizeCtx, tx, payload.IncidentID, payload.GraphViewID,
				payload.MaterializationGeneration, execution.JobID(), graphViewMaterializationFailureCode(reasonCode), m.now().UTC(),
			)
		},
	})
	return err
}

func graphViewMaterializationTimedOut(ctx context.Context) bool {
	return errors.Is(ctx.Err(), context.DeadlineExceeded)
}

func graphViewMaterializationFailureCode(reasonCode string) string {
	return "network_flow_graph_materialization_" + reasonCode
}

func (m *Module) cancelGraphViewMaterialization(ctx context.Context, execution jobs.Execution) error {
	total := 1
	_, err := m.jobManager.CompleteCanceled(ctx, execution, jobs.CancellationCompletion{
		Progress:      jobs.Progress{Completed: 0, Total: &total},
		ResultSummary: jobs.ResultSummary{Code: "network_flow_graph_materialization_cancelled", Message: "Saved graph materialization canceled."},
	})
	return err
}

func graphViewSelectedResultFromBinding(binding graphprojection.ResultBindingV2) GraphViewSelectedResultBinding {
	return GraphViewSelectedResultBinding{
		ProjectionResultID: binding.ProjectionResultID, SourceSnapshotID: binding.SourceSnapshotID,
		ProjectionSchemaID: binding.ProjectionSchemaID, ProjectionVersion: binding.ProjectionVersion,
		NormalizedConfigurationSHA256: binding.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        binding.NormalizedSourceSHA256, CanonicalOutputSHA256: binding.CanonicalOutputSHA256,
	}
}

func graphViewRoute(incidentID uuid.UUID, graphViewID string) string {
	return "/api/v1/incidents/" + incidentID.String() + "/network-flow/graph-views/" + graphViewID
}
