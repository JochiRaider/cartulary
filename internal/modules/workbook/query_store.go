package workbook

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type QueryStore struct {
	entityStore    entityQueryPort
	indicatorStore indicatorQueryPort
	projectionRows *projectionadapters.WorkbookRows
}

type entityQueryPort interface {
	QueryHostRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error)
	QueryIdentityRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error)
	QueryHostRowsPage(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error)
	QueryIdentityRowsPage(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error)
}

type indicatorQueryPort interface {
	QueryRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error)
	QueryRowsPage(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error)
}

func NewQueryStore(pool postgres.DB) *QueryStore {
	return &QueryStore{
		entityStore:    hostidentity.NewStore(pool),
		indicatorStore: indicators.NewStore(pool),
		projectionRows: projectionadapters.NewWorkbookRows(pool),
	}
}

func (s *Store) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	return queryStoreFromStore(s).QueryRows(ctx, incidentID, viewSchemaID, query)
}

func (s *Store) QueryRowsPage(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	return queryStoreFromStore(s).QueryRowsPage(ctx, incidentID, viewSchemaID, query, window)
}

func queryStoreFromStore(s *Store) *QueryStore {
	return &QueryStore{
		entityStore:    s.entityStore,
		indicatorStore: s.indicatorStore,
		projectionRows: s.projectionRows,
	}
}

func (s *QueryStore) QueryRows(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta) ([]map[string]any, error) {
	switch viewSchemaID {
	case hostidentity.HostsViewSchemaID:
		return s.entityStore.QueryHostRows(ctx, incidentID, query)
	case hostidentity.IdentitiesViewSchemaID:
		return s.entityStore.QueryIdentityRows(ctx, incidentID, query)
	case indicators.ViewSchemaID:
		return s.indicatorStore.QueryRows(ctx, incidentID, query)
	default:
		if !s.projectionRows.Supports(viewSchemaID) {
			return nil, fmt.Errorf("workbook query surface %q not mapped", viewSchemaID)
		}
		return s.projectionRows.QueryRows(ctx, incidentID, viewSchemaID, query)
	}
}

func (s *QueryStore) QueryRowsPage(ctx context.Context, incidentID uuid.UUID, viewSchemaID string, query viewschema.QueryMeta, window querypage.Window) (querypage.Result, error) {
	switch viewSchemaID {
	case hostidentity.HostsViewSchemaID:
		return s.entityStore.QueryHostRowsPage(ctx, incidentID, query, window)
	case hostidentity.IdentitiesViewSchemaID:
		return s.entityStore.QueryIdentityRowsPage(ctx, incidentID, query, window)
	case indicators.ViewSchemaID:
		return s.indicatorStore.QueryRowsPage(ctx, incidentID, query, window)
	default:
		if !s.projectionRows.Supports(viewSchemaID) {
			return querypage.Result{}, fmt.Errorf("workbook query surface %q not mapped", viewSchemaID)
		}
		return s.projectionRows.QueryRowsPage(ctx, incidentID, viewSchemaID, query, window)
	}
}
