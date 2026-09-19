package networkflow

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
)

type graphSourceReader interface {
	ListActiveTables(context.Context, uuid.UUID) ([]tableRecord, error)
	GetActiveTable(context.Context, uuid.UUID, string) (tableRecord, error)
	IterateRowsForTables(context.Context, uuid.UUID, []string, func(flowRow) error) error
	IterateGraphContributorRows(context.Context, uuid.UUID, []string, graphContributorPredicate, func(flowRow) error) error
}

type graphSourceComposer struct {
	store           graphSourceReader
	limits          EffectiveLimits
	graphProjection graphProjectionPort
	now             func() time.Time
	graphTelemetry  GraphTelemetryObserver
}

func newGraphSourceComposer(source graphSourceReader, limits EffectiveLimits, projection graphProjectionPort, now func() time.Time, observer GraphTelemetryObserver) (*graphSourceComposer, error) {
	if source == nil || projection == nil || now == nil {
		return nil, errors.New("network flow graph composition dependencies are required")
	}
	if err := checkEffectiveLimits(limits); err != nil {
		return nil, err
	}
	return &graphSourceComposer{store: source, limits: limits, graphProjection: projection, now: now, graphTelemetry: observer}, nil
}

func (s *graphSourceComposer) composeGraph(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, request graphQueryRequest) (graphComposition, *semanticFailure) {
	composition, apiErr := s.composeGraphSource(ctx, incidentID, request)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	if err := ctx.Err(); err != nil {
		return graphComposition{}, graphProjectionFailedForContext(err)
	}
	sourceSnapshotID := graphSourceSnapshotDigest(incidentID, composition.SourceTables, composition.Digest)
	projectionStarted := time.Now()
	projection, apiErr := s.projectNetworkFlowGraph(ctx, actorUserID, sourceSnapshotID, composition, s.now())
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
	schemaID := schemaGraphSemanticQueryV2
	digest := graphQueryDigestV2(incidentID, tableIDs, request.Filters, request.TimeRange, request.Aggregation)
	composition := graphComposition{
		SemanticSchemaID: schemaID,
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
	err := s.store.IterateRowsForTables(ctx, incidentID, tableIDs, func(row flowRow) error {
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
	composition.SemanticQuery = graphSemanticQueryResource(schemaID, tableIDs, request.Filters, request.TimeRange, request.Aggregation, request.Limits)
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
	composition.SemanticSchemaID = semantic.SchemaID
	composition.SemanticQuery = semantic.Raw
	composition.Digest = graphQueryDigestForSemantic(incidentID, composition.SelectedTableIDs, semantic)
	return composition, nil
}

func (s *graphSourceComposer) resolveGraphTables(ctx context.Context, incidentID uuid.UUID, scope tableScope) ([]tableRecord, []string, map[string]int, *semanticFailure) {
	activeTables, err := s.store.ListActiveTables(ctx, incidentID)
	if err != nil {
		return nil, nil, nil, internalSemanticFailure(err)
	}
	byID := make(map[string]tableRecord, len(activeTables))
	for _, table := range activeTables {
		byID[table.TableID] = table
	}
	var selected []tableRecord
	switch scope.Mode {
	case "active_table":
		table, ok := byID[scope.ActiveTableID]
		if !ok {
			if _, err := s.store.GetActiveTable(ctx, incidentID, scope.ActiveTableID); err != nil {
				return nil, nil, nil, tableReadFailure(err)
			}
			return nil, nil, nil, newSemanticFailure(failureTableNotFound, "network_flow_table_id", "not_found")
		}
		selected = []tableRecord{table}
	case "selected_tables":
		if len(scope.SelectedTableIDs) == 0 {
			return nil, nil, nil, invalidTableScope("table_scope", "empty_resolved_scope")
		}
		selectedSet := stringSet(scope.SelectedTableIDs)
		for _, table := range activeTables {
			if _, ok := selectedSet[table.TableID]; ok {
				selected = append(selected, table)
				delete(selectedSet, table.TableID)
			}
		}
		if len(selectedSet) > 0 {
			missing := make([]string, 0, len(selectedSet))
			for tableID := range selectedSet {
				missing = append(missing, tableID)
			}
			sort.Strings(missing)
			if _, err := s.store.GetActiveTable(ctx, incidentID, missing[0]); err != nil {
				return nil, nil, nil, tableReadFailure(err)
			}
			return nil, nil, nil, newSemanticFailure(failureTableNotFound, "network_flow_table_id", "not_found")
		}
	case "all_active_tables":
		selected = activeTables
	default:
		return nil, nil, nil, invalidTableScope("mode", "unknown_mode")
	}
	if len(selected) == 0 {
		return nil, nil, nil, invalidTableScope("table_scope", "empty_resolved_scope")
	}
	tableIDs := make([]string, 0, len(selected))
	tableRanks := make(map[string]int, len(selected))
	for index, table := range selected {
		tableIDs = append(tableIDs, table.TableID)
		tableRanks[table.TableID] = index
	}
	return selected, tableIDs, tableRanks, nil
}

func (s *graphSourceComposer) projectNetworkFlowGraph(ctx context.Context, actorUserID uuid.UUID, sourceSnapshotID string, composition graphComposition, requestedAt time.Time) (map[string]any, *semanticFailure) {
	if err := ctx.Err(); err != nil {
		return nil, graphProjectionFailedForContext(err)
	}
	graphViewKeySnapshot := sourceSnapshotID
	if len(graphViewKeySnapshot) > len("nfsnap_") && graphViewKeySnapshot[:len("nfsnap_")] == "nfsnap_" {
		graphViewKeySnapshot = graphViewKeySnapshot[len("nfsnap_"):]
	}
	graphViewKey := "network_flow_activity:" + composition.SourceTables[0].IncidentID.String() + ":" + graphViewKeySnapshot
	projector := s.graphProjection
	if projector == nil {
		projector = newGraphProjectionAdapter()
	}
	graphViewID, err := deriveNetworkFlowGraphViewID(graphViewKey)
	if err != nil {
		return nil, graphProjectionFailed("adapter_contract_rejected")
	}
	input := networkFlowProjectionInput(sourceSnapshotID, composition)
	projectionResource, err := projector.ProjectEphemeral(ctx, graphViewID, canonicalJSON(input))
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			return nil, graphProjectionFailed("projection_cancelled")
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, graphProjectionFailed("projection_timeout")
		}
		var adapterErr *graphProjectionAdapterError
		if errors.As(err, &adapterErr) {
			return nil, graphProjectionFailed(adapterErr.reason)
		}
		return nil, graphProjectionFailed("projection_unavailable")
	}
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
	digest := graphQueryDigestForSemantic(incidentID, tableIDs, semantic)
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
	err := s.store.IterateGraphContributorRows(ctx, incidentID, tableIDs, predicate, func(row flowRow) error {
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
