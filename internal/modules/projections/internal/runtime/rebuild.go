package runtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
)

type RestoreRebuilder struct {
	store *Store
}

func NewRestoreRebuilderFromStore(store *Store) *RestoreRebuilder {
	return &RestoreRebuilder{store: store}
}

func (r *RestoreRebuilder) RebuildRestoreProjections(ctx context.Context, request restorecontract.ProjectionRebuildRequest) (restorecontract.ProjectionRebuildResult, error) {
	var store *Store
	if r != nil {
		store = r.store
	}
	return rebuildRestoreProjections(ctx, store, request)
}

func rebuildRestoreProjections(ctx context.Context, s *Store, request restorecontract.ProjectionRebuildRequest) (result restorecontract.ProjectionRebuildResult, err error) {
	ctx, finishTelemetry := s.startProjectionSpan(ctx, "unknown")
	defer func() { finishTelemetry(err) }()

	result = restorecontract.ProjectionRebuildResult{
		RestoreOperationID: request.RestoreOperationID,
		Status:             restorecontract.ProjectionRebuildStatusFailed,
		ReadinessOutcome:   restorecontract.ProjectionReadinessIncomplete,
		ProviderResults:    []restorecontract.ProjectionProviderResult{},
		Warnings:           []restorecontract.ProjectionRebuildMessage{},
		Errors:             []restorecontract.ProjectionRebuildMessage{},
	}
	if validationErr := validateRestoreProjectionRebuildRequest(request); validationErr != nil {
		result.Errors = append(result.Errors, restorecontract.ProjectionRebuildMessage{
			Code:    "invalid_restore_projection_rebuild_request",
			Message: validationErr.Error(),
		})
		return result, fmt.Errorf("rebuild restore projections: %w", validationErr)
	}
	if s == nil || s.pool == nil {
		result.Errors = append(result.Errors, restorecontract.ProjectionRebuildMessage{
			Code:    "projection_store_required",
			Message: "projection store is required",
		})
		return result, fmt.Errorf("rebuild restore projections: projection store is required")
	}
	if s.registry == nil || len(s.registry.providers) == 0 {
		result.Status = restorecontract.ProjectionRebuildStatusNotApplicable
		result.ReadinessOutcome = restorecontract.ProjectionReadinessNotApplicable
		return result, nil
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		result.Errors = append(result.Errors, restorecontract.ProjectionRebuildMessage{
			Code:    "begin_restore_projection_rebuild_failed",
			Message: err.Error(),
		})
		return result, fmt.Errorf("begin restore projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	incidentIDs, err := listRestoreProjectionIncidentIDs(ctx, tx)
	if err != nil {
		result.Errors = append(result.Errors, restorecontract.ProjectionRebuildMessage{
			Code:    "list_restore_projection_incidents_failed",
			Message: err.Error(),
		})
		return result, err
	}
	registry, registryErr := s.providerRegistry()
	if registryErr != nil {
		result.Errors = append(result.Errors, restorecontract.ProjectionRebuildMessage{
			Code:    "projection_catalog_required",
			Message: registryErr.Error(),
		})
		return result, fmt.Errorf("rebuild restore projections: %w", registryErr)
	}
	participatingProviders := 0
	type pendingProviderSuccess struct {
		resultIndex   int
		viewSchemaIDs []string
		tableResults  []restorecontract.ProjectionTableResult
	}
	pendingSuccesses := make([]pendingProviderSuccess, 0, len(registry.rebuildOrder))
	for _, provider := range registry.rebuildOrder {
		providerResult := restorecontract.ProjectionProviderResult{
			ProviderKey:             provider.descriptor.ProviderID,
			Status:                  restorecontract.ProjectionProviderResultFailed,
			RebuiltViewSchemaIDs:    []string{},
			RebuiltProjectionTables: []restorecontract.ProjectionTableResult{},
			Warnings:                []string{},
		}
		if provider.descriptor.RestoreRebuild == providercontract.RestoreRebuildNonparticipating {
			providerResult.Status = restorecontract.ProjectionProviderResultSkippedNonparticipating
			result.ProviderResults = append(result.ProviderResults, providerResult)
			continue
		}
		if provider.descriptor.RestoreRebuild != providercontract.RestoreRebuildRequired ||
			!provider.descriptor.Capabilities.RestoreRebuild ||
			!provider.descriptor.Capabilities.IncidentRebuild ||
			provider.rebuildIncidentTx == nil {
			err := fmt.Errorf("projection provider %q is not restore-rebuild ready", provider.descriptor.ProviderID)
			providerResult.Status = restorecontract.ProjectionProviderResultFailed
			providerResult.Error = err.Error()
			result.ProviderResults = append(result.ProviderResults, providerResult)
			markRestoreProjectionResultsRolledBack(result.ProviderResults, "restore projection rebuild rolled back before commit")
			result.Errors = append(result.Errors, restorecontract.ProjectionRebuildMessage{
				Code:        "restore_projection_provider_not_ready",
				Message:     err.Error(),
				ProviderKey: provider.descriptor.ProviderID,
			})
			return result, err
		}
		participatingProviders++
		providerResult.IncidentCount = len(incidentIDs)
		for _, incidentID := range incidentIDs {
			if err := provider.rebuildIncidentTx(ctx, s, tx, incidentID); err != nil {
				wrapped := fmt.Errorf("rebuild %s projection for incident %s: %w", provider.descriptor.ProviderID, incidentID, err)
				providerResult.Status = restorecontract.ProjectionProviderResultFailed
				providerResult.Error = wrapped.Error()
				result.ProviderResults = append(result.ProviderResults, providerResult)
				markRestoreProjectionResultsRolledBack(result.ProviderResults, "restore projection rebuild rolled back before commit")
				result.Errors = append(result.Errors, restorecontract.ProjectionRebuildMessage{
					Code:        "restore_projection_provider_rebuild_failed",
					Message:     wrapped.Error(),
					ProviderKey: provider.descriptor.ProviderID,
				})
				return result, wrapped
			}
		}
		tableResults, err := restoreProjectionProviderTableResults(ctx, tx, provider.descriptor.ProjectionTableIDs)
		if err != nil {
			wrapped := fmt.Errorf("count %s projection rows after restore rebuild: %w", provider.descriptor.ProviderID, err)
			providerResult.Status = restorecontract.ProjectionProviderResultFailed
			providerResult.Error = wrapped.Error()
			result.ProviderResults = append(result.ProviderResults, providerResult)
			markRestoreProjectionResultsRolledBack(result.ProviderResults, "restore projection rebuild rolled back before commit")
			result.Errors = append(result.Errors, restorecontract.ProjectionRebuildMessage{
				Code:        "restore_projection_provider_row_count_failed",
				Message:     wrapped.Error(),
				ProviderKey: provider.descriptor.ProviderID,
			})
			return result, wrapped
		}
		result.ProviderResults = append(result.ProviderResults, providerResult)
		pendingSuccesses = append(pendingSuccesses, pendingProviderSuccess{
			resultIndex:   len(result.ProviderResults) - 1,
			viewSchemaIDs: append([]string(nil), provider.descriptor.ViewSchemaIDs...),
			tableResults:  tableResults,
		})
	}
	if participatingProviders == 0 {
		result.Status = restorecontract.ProjectionRebuildStatusNotApplicable
		result.ReadinessOutcome = restorecontract.ProjectionReadinessNotApplicable
		return result, nil
	}
	if err := tx.Commit(ctx); err != nil {
		markRestoreProjectionResultsRolledBack(result.ProviderResults, "restore projection rebuild commit failed")
		result.Errors = append(result.Errors, restorecontract.ProjectionRebuildMessage{
			Code:    "commit_restore_projection_rebuild_failed",
			Message: err.Error(),
		})
		return result, fmt.Errorf("commit restore projection rebuild: %w", err)
	}
	for _, pending := range pendingSuccesses {
		providerResult := &result.ProviderResults[pending.resultIndex]
		providerResult.Status = restorecontract.ProjectionProviderResultSucceeded
		providerResult.RebuiltViewSchemaIDs = pending.viewSchemaIDs
		providerResult.RebuiltProjectionTables = pending.tableResults
		providerResult.Error = ""
	}
	result.Status = restorecontract.ProjectionRebuildStatusSucceeded
	result.ReadinessOutcome = restorecontract.ProjectionReadinessReady
	return result, nil
}

func validateRestoreProjectionRebuildRequest(request restorecontract.ProjectionRebuildRequest) error {
	if err := request.Validate(); err != nil {
		return err
	}
	if request.ProviderRegistryRef != restorecontract.ProviderRegistryRefCodeBacked {
		return fmt.Errorf("unsupported provider_registry_ref %q", request.ProviderRegistryRef)
	}
	return nil
}

func restoreProjectionProviderTableResults(ctx context.Context, tx pgx.Tx, tableFamilies []string) ([]restorecontract.ProjectionTableResult, error) {
	results := make([]restorecontract.ProjectionTableResult, 0, len(tableFamilies))
	for _, tableFamily := range tableFamilies {
		query := fmt.Sprintf("SELECT count(*) FROM %s", pgx.Identifier{tableFamily}.Sanitize())
		var count int64
		if err := tx.QueryRow(ctx, query).Scan(&count); err != nil {
			return nil, fmt.Errorf("count %s rows: %w", tableFamily, err)
		}
		results = append(results, restorecontract.ProjectionTableResult{
			ProjectionTableID: tableFamily,
			RowCount:          count,
		})
	}
	return results, nil
}

func markRestoreProjectionResultsRolledBack(results []restorecontract.ProjectionProviderResult, message string) {
	for index := range results {
		providerResult := &results[index]
		if providerResult.Status == restorecontract.ProjectionProviderResultSkippedNonparticipating {
			continue
		}
		providerResult.Status = restorecontract.ProjectionProviderResultFailed
		providerResult.RebuiltViewSchemaIDs = []string{}
		providerResult.RebuiltProjectionTables = []restorecontract.ProjectionTableResult{}
		if providerResult.Error == "" {
			providerResult.Error = message
		}
	}
}

func listRestoreProjectionIncidentIDs(ctx context.Context, tx pgx.Tx) ([]uuid.UUID, error) {
	rows, err := tx.Query(ctx, `SELECT id FROM incidents ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list restore projection incidents: %w", err)
	}
	defer rows.Close()

	incidentIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var rawIncidentID pgtype.UUID
		if err := rows.Scan(&rawIncidentID); err != nil {
			return nil, fmt.Errorf("scan restore projection incident: %w", err)
		}
		incidentID, err := uuidFromPG(rawIncidentID)
		if err != nil {
			return nil, fmt.Errorf("decode restore projection incident: %w", err)
		}
		incidentIDs = append(incidentIDs, incidentID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate restore projection incidents: %w", err)
	}
	return incidentIDs, nil
}
