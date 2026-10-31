package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type graphSourceReader interface {
	tableScopeReader
	IterateRowsForTables(context.Context, uuid.UUID, []string, func(flowRow) error) error
	IterateGraphContributorRows(context.Context, uuid.UUID, []string, graphContributorPredicate, func(flowRow) error) error
}

type graphSourceComposer struct {
	store           graphSourceReader
	limits          EffectiveLimits
	graphProjection graphProjectionPort
	graphTelemetry  GraphTelemetryObserver
}

func newGraphSourceComposer(source graphSourceReader, limits EffectiveLimits, projection graphProjectionPort, observer GraphTelemetryObserver) (*graphSourceComposer, error) {
	if source == nil || projection == nil {
		return nil, errors.New("network flow graph composition dependencies are required")
	}
	if err := checkEffectiveLimits(limits); err != nil {
		return nil, err
	}
	return &graphSourceComposer{store: source, limits: limits, graphProjection: projection, graphTelemetry: observer}, nil
}

func (s *graphSourceComposer) composeGraph(ctx context.Context, incidentID uuid.UUID, request graphQueryRequest) (graphComposition, *semanticFailure) {
	composition, apiErr := s.composeGraphSource(ctx, incidentID, request)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	if err := ctx.Err(); err != nil {
		return graphComposition{}, graphProjectionFailedForContext(err)
	}
	sourceSnapshotID := graphSourceSnapshotDigest(incidentID, composition.SourceTables, composition.Digest)
	projectionStarted := time.Now()
	projection, apiErr := s.projectNetworkFlowGraph(ctx, sourceSnapshotID, composition)
	s.observeGraphPhase(ctx, graphTelemetryPhaseProjection, request.Aggregation.Mode, projectionStarted, apiErr)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	composition.GraphProjection = projection
	if apiErr := bindGraphV2ResponseMetadata(&composition); apiErr != nil {
		return graphComposition{}, apiErr
	}
	s.observeGraphComposition(ctx, composition)
	return composition, nil
}

func (s *graphSourceComposer) composeGraphSource(ctx context.Context, incidentID uuid.UUID, request graphQueryRequest) (graphComposition, *semanticFailure) {
	validationStarted := time.Now()
	tables, tableIDs, tableRanks, apiErr := s.resolveGraphTables(ctx, incidentID, request.TableScope)
	s.observeGraphPhase(ctx, graphTelemetryPhaseSourceValidation, request.Aggregation.Mode, validationStarted, apiErr)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	digest, err := graphQueryDigestV2(incidentID, tableIDs, request.Filters, request.TimeRange, request.Aggregation)
	if err != nil {
		return graphComposition{}, internalSemanticFailure(err)
	}
	composition := graphComposition{
		Aggregation:      request.Aggregation,
		Digest:           digest,
		ResultLimits:     request.Limits,
		SourceTables:     tables,
		SourceTableRefs:  graphSourceTableRefs(tables),
		TableRanks:       tableRanks,
		Vertices:         map[string]*graphVertex{},
		Edges:            map[string]*graphEdge{},
		SelectedTableIDs: tableIDs,
		IncludeExamples:  request.Aggregation.IncludeExampleRowRefs,
	}
	if request.Aggregation.Mode == "time_bucket_v1" {
		buckets, bucketErr := graphTimeBuckets(request.TimeRange, request.Aggregation.BucketWidthSeconds, request.Limits.MaxTimeBuckets)
		if bucketErr != nil {
			s.observeGraphPhase(ctx, graphTelemetryPhaseSourceValidation, request.Aggregation.Mode, validationStarted, bucketErr)
			return graphComposition{}, bucketErr
		}
		composition.TimeBuckets = buckets
	}
	tableByID := make(map[string]tableRecord, len(tables))
	for _, table := range tables {
		tableByID[table.TableID] = table
	}
	if err := ctx.Err(); err != nil {
		return graphComposition{}, graphProjectionFailedForContext(err)
	}
	var compositionErr *semanticFailure
	scanStarted := time.Now()
	err = s.store.IterateRowsForTables(ctx, incidentID, tableIDs, func(row flowRow) error {
		matched, rowErr := rowMatchesGraphQuery(row, request.Filters, request.TimeRange, request.Aggregation)
		if rowErr != nil {
			compositionErr = rowErr
			return errStopGraphIteration
		}
		if !matched {
			return nil
		}
		if rowErr = composeGraphRow(incidentID, row, tableByID, &composition); rowErr != nil {
			compositionErr = rowErr
			return errStopGraphIteration
		}
		return nil
	})
	if compositionErr != nil {
		s.observeGraphPhase(ctx, graphTelemetryPhaseSourceScan, request.Aggregation.Mode, scanStarted, compositionErr)
		return graphComposition{}, compositionErr
	}
	if err != nil {
		var scanErr *semanticFailure
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			scanErr = graphProjectionFailedForContext(err)
		} else {
			scanErr = internalSemanticFailure(err)
		}
		s.observeGraphPhase(ctx, graphTelemetryPhaseSourceScan, request.Aggregation.Mode, scanStarted, scanErr)
		return graphComposition{}, scanErr
	}
	if apiErr := validateGraphLimits(composition); apiErr != nil {
		s.observeGraphPhase(ctx, graphTelemetryPhaseSourceScan, request.Aggregation.Mode, scanStarted, apiErr)
		return graphComposition{}, apiErr
	}
	composition.SemanticQuery = graphSemanticQueryResource(tableIDs, request.Filters, request.TimeRange, request.Aggregation)
	s.observeGraphPhase(ctx, graphTelemetryPhaseSourceScan, request.Aggregation.Mode, scanStarted, nil)
	return composition, nil
}

func (s *graphSourceComposer) composeGraphSourceFromSemantic(ctx context.Context, incidentID uuid.UUID, semantic graphSemanticRequest) (graphComposition, *semanticFailure) {
	if semantic.ResultLimits.MaxContributingRows == 0 {
		semantic.ResultLimits.MaxContributingRows = int(s.limits.MaxContributingRowsPerGraph)
	}
	if semantic.ResultLimits.MaxTimeBuckets == 0 {
		semantic.ResultLimits.MaxTimeBuckets = int(s.limits.MaxTimeBucketsPerGraph)
	}
	request := graphQueryRequest{
		TableScope:  tableScope{Mode: "selected_tables", SelectedTableIDs: semantic.SelectedTableIDs},
		Filters:     semantic.Filters,
		TimeRange:   semantic.TimeRange,
		Aggregation: semantic.Aggregation,
		Limits:      semantic.ResultLimits,
	}
	composition, apiErr := s.composeGraphSource(ctx, incidentID, request)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	composition.SemanticQuery = semantic.Raw
	digest, err := graphQueryDigestForSemantic(incidentID, composition.SelectedTableIDs, semantic)
	if err != nil {
		return graphComposition{}, internalSemanticFailure(err)
	}
	composition.Digest = digest
	return composition, nil
}

func (s *graphSourceComposer) resolveGraphTables(ctx context.Context, incidentID uuid.UUID, scope tableScope) ([]tableRecord, []string, map[string]int, *semanticFailure) {
	return resolveActiveTableScope(ctx, s.store, s.limits, incidentID, scope)
}

func (s *graphSourceComposer) projectNetworkFlowGraph(ctx context.Context, sourceSnapshotID string, composition graphComposition) (map[string]any, *semanticFailure) {
	if err := ctx.Err(); err != nil {
		return nil, graphProjectionFailedForContext(err)
	}
	graphViewKeySnapshot := sourceSnapshotID
	if len(graphViewKeySnapshot) > len("nfsnap_") && graphViewKeySnapshot[:len("nfsnap_")] == "nfsnap_" {
		graphViewKeySnapshot = graphViewKeySnapshot[len("nfsnap_"):]
	}
	graphViewKey := "network_flow_activity:" + composition.SourceTables[0].IncidentID.String() + ":" + graphViewKeySnapshot
	graphViewID, err := deriveNetworkFlowGraphViewID(graphViewKey)
	if err != nil {
		return nil, graphProjectionFailed("adapter_contract_rejected")
	}
	input := networkFlowProjectionInput(sourceSnapshotID, composition)
	encoded, err := json.Marshal(input)
	if err != nil {
		return nil, internalSemanticFailure(err)
	}
	projection, err := s.graphProjection.Project(ctx, graphViewID, encoded, nil)
	if err != nil {
		failure := graphProjectionFailed("projection_unavailable")
		switch {
		case errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled):
			failure.reason = "projection_cancelled"
		case errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded):
			failure.reason = "projection_timeout"
		default:
			var adapterErr *graphProjectionAdapterError
			if errors.As(err, &adapterErr) {
				failure.reason = adapterErr.reason
			}
		}
		failure.cause = err
		return nil, failure
	}
	projectionResource := graphProjectionResource(projection)
	summary, ok := projectionResource["validation_summary"].(map[string]any)
	if !ok || summary["fatal_count"] != 0 || summary["error_count"] != 0 || summary["warning_count"] != 0 || summary["info_count"] != 0 {
		return nil, graphProjectionFailed("adapter_contract_rejected")
	}
	return projectionResource, nil
}

func (s *graphSourceComposer) queryGraphContributorPage(ctx context.Context, incidentID uuid.UUID, semantic graphSemanticRequest, expectedDigest string, selector graphSelector, position *contributorCursorPosition, limit int) ([]flowRow, bool, map[string]int, *semanticFailure) {
	_, tableIDs, tableRanks, apiErr := s.resolveGraphTables(ctx, incidentID, tableScope{Mode: "selected_tables", SelectedTableIDs: semantic.SelectedTableIDs})
	if apiErr != nil {
		return nil, false, nil, apiErr
	}
	digest, err := graphQueryDigestForSemantic(incidentID, tableIDs, semantic)
	if err != nil {
		return nil, false, nil, internalSemanticFailure(err)
	}
	if digest != expectedDigest {
		return nil, false, nil, graphQueryStale("digest_mismatch", expectedDigest)
	}
	if apiErr := validateGraphSelectorForSemantic(semantic, selector); apiErr != nil {
		return nil, false, nil, apiErr
	}
	predicate, apiErr := canonicalGraphContributorPredicate(incidentID, selector)
	if apiErr != nil {
		return nil, false, nil, apiErr
	}

	rows := make([]flowRow, 0, limit+1)
	matchedAny := false
	err = s.store.IterateGraphContributorRows(ctx, incidentID, tableIDs, predicate, func(row flowRow) error {
		matched, matchErr := rowMatchesGraphQuery(row, semantic.Filters, semantic.TimeRange, semantic.Aggregation)
		if matchErr != nil {
			apiErr = matchErr
			return errStopGraphIteration
		}
		if !matched {
			return nil
		}
		matchedAny = true
		if position != nil {
			rank := tableRanks[row.NetworkFlowTableID]
			if rank < position.WorkspaceTableOrder || rank == position.WorkspaceTableOrder && compareRowToPosition(row, position.Row) <= 0 {
				return nil
			}
		}
		rows = append(rows, row)
		if len(rows) > limit {
			return errStopGraphIteration
		}
		return nil
	})
	if apiErr != nil {
		return nil, false, nil, apiErr
	}
	if err != nil && !errors.Is(err, errStopGraphIteration) {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, false, nil, graphProjectionFailedForContext(err)
		}
		return nil, false, nil, internalSemanticFailure(err)
	}
	if !matchedAny {
		reason := "edge_not_found"
		if predicate.Kind == "vertex" {
			reason = "vertex_not_found"
		}
		return nil, false, nil, graphQueryStale(reason, digest)
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, tableRanks, nil
}
