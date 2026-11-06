package networkflow

import (
	"context"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
)

// Attribution is resolved only after the owner read has admitted the incident
// and closed its cursor. Historical UUIDs identify authors, never local accounts.
// Keep storage foreign keys and authorization identities unchanged.
func resolveHistoricalAuthors[T any](ctx context.Context, s *store, query incidentportability.Queryer, incidentID uuid.UUID, table string, values []T, identity func(*T) (string, *uuid.UUID)) error {
	if len(values) == 0 || s.importedAttributions == nil {
		return nil
	}
	ids := make([]string, len(values))
	for i := range values {
		ids[i], _ = identity(&values[i])
	}
	resolved, err := s.importedAttributions.ResolvePortableSourceActors(ctx, query, incidentID, table, "created_by_user_id", ids)
	if err != nil {
		return err
	}
	for i := range values {
		id, actor := identity(&values[i])
		if original, ok := resolved[id]; ok {
			value, err := uuid.Parse(original)
			if err != nil || value == uuid.Nil {
				return fmt.Errorf("invalid historical Network Flow author")
			}
			*actor = value
		}
	}
	return nil
}

func (s *store) resolveTableAuthor(ctx context.Context, query incidentportability.Queryer, value tableRecord) (tableRecord, error) {
	values := []tableRecord{value}
	err := resolveHistoricalAuthors(ctx, s, query, value.IncidentID, "network_flow_tables", values, tableAuthor)
	return values[0], err
}
func tableAuthor(v *tableRecord) (string, *uuid.UUID) { return v.TableID, &v.CreatedByUserID }
func graphAuthor(v *graphViewDeclaration) (string, *uuid.UUID) {
	return v.GraphViewID, &v.CreatedByUserID
}
func (s *store) resolveGraphAuthor(ctx context.Context, query incidentportability.Queryer, value graphViewDeclaration) (graphViewDeclaration, error) {
	values := []graphViewDeclaration{value}
	err := resolveHistoricalAuthors(ctx, s, query, value.IncidentID, "network_flow_graph_views", values, graphAuthor)
	return values[0], err
}
func (s *store) resolveRowAuthors(ctx context.Context, incidentID uuid.UUID, values []flowRow) error {
	return resolveHistoricalAuthors(ctx, s, s.pool, incidentID, "network_flow_rows", values, func(v *flowRow) (string, *uuid.UUID) { return v.RowID, &v.CreatedByUserID })
}

func (s *store) ListActiveTables(ctx context.Context, incidentID uuid.UUID) ([]tableRecord, error) {
	values, err := s.graphSourceReadStore.ListActiveTables(ctx, incidentID)
	if err == nil {
		err = resolveHistoricalAuthors(ctx, s, s.pool, incidentID, "network_flow_tables", values, tableAuthor)
	}
	return values, err
}
func (s *store) GetActiveTable(ctx context.Context, incidentID uuid.UUID, id string) (tableRecord, error) {
	value, err := s.graphSourceReadStore.GetActiveTable(ctx, incidentID, id)
	if err != nil {
		return value, err
	}
	return s.resolveTableAuthor(ctx, s.pool, value)
}
