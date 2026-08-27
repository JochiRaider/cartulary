package runtime

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	commLogViewSchemaID              = "cartulary.view.comm_log.v1"
	findingsViewSchemaID             = "cartulary.view.findings.v1"
	forensicKeywordsViewSchemaID     = "cartulary.view.forensic_keywords.v1"
	handoffViewSchemaID              = "cartulary.view.handoff.v1"
	investigativeQueriesViewSchemaID = "cartulary.view.investigative_queries.v1"
	lessonViewSchemaID               = "cartulary.view.lesson.v1"
	statusReviewViewSchemaID         = "cartulary.view.status_review.v1"
)

func (s *Store) Supports(viewSchemaID string) bool {
	if s == nil || s.registry == nil {
		return false
	}
	_, ok := s.registry.querySurfaceForView(viewSchemaID)
	return ok
}

func (s *Store) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	page, err := s.QueryRowsPage(ctx, incidentID, viewSchemaID, query, querypage.Window{Limit: int(^uint(0)>>1) - 1})
	return page.Rows, err
}

func (s *Store) QueryRowsPage(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	registry, err := s.providerRegistry()
	if err != nil {
		return querypage.Result{}, err
	}
	definition, ok := registry.querySurfaceForView(viewSchemaID)
	if !ok {
		return querypage.Result{}, fmt.Errorf("projection query surface %q not mapped", viewSchemaID)
	}
	sqlText, args, err := queryengine.BuildQueryPageSQL(incidentID, definition, query, window)
	if err != nil {
		return querypage.Result{}, err
	}
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return querypage.Result{}, fmt.Errorf("query workbook rows: %w", err)
	}
	defer rows.Close()

	result, err := queryengine.ScanRows(rows, definition)
	if err != nil {
		return querypage.Result{}, err
	}
	return querypage.Finish(result, window.Limit), nil
}

func (s *Store) LoadRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	registry, err := s.providerRegistry()
	if err != nil {
		return nil, err
	}
	definition, ok := registry.querySurfaceForView(viewSchemaID)
	if !ok {
		return nil, fmt.Errorf("projection query surface %q not mapped", viewSchemaID)
	}
	return queryengine.LoadRowTx(ctx, tx, definition, recordID)
}
