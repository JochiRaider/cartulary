package networkflow

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GraphRestoreReadSession borrows the caller's snapshot without granting
// mutation, transaction creation, commit or rollback capabilities.
type GraphRestoreReadSession interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type graphSourceReadStore struct{ reader GraphRestoreReadSession }

func (s *graphSourceReadStore) ListActiveTables(ctx context.Context, incidentID uuid.UUID) ([]tableRecord, error) {
	rows, err := s.reader.Query(ctx, tableSelectColumns()+`
  FROM network_flow_tables
 WHERE incident_id = $1
   AND table_status = 'active'
 ORDER BY created_at ASC, network_flow_table_id ASC
`, incidentID)
	if err != nil {
		return nil, fmt.Errorf("list active network flow tables: %w", err)
	}
	defer rows.Close()
	return scanTables(rows)
}

func (s *graphSourceReadStore) GetActiveTable(ctx context.Context, incidentID uuid.UUID, tableID string) (tableRecord, error) {
	table, err := getTable(ctx, s.reader, incidentID, tableID, false)
	if err != nil {
		return tableRecord{}, err
	}
	if table.TableStatus != tableStatusActive {
		return tableRecord{}, errTableNotActive
	}
	return table, nil
}

func (s *graphSourceReadStore) IterateRowsForTables(ctx context.Context, incidentID uuid.UUID, tableIDs []string, visit func(flowRow) error) error {
	if len(tableIDs) == 0 {
		return nil
	}
	rows, err := s.reader.Query(ctx, `
SELECT `+flowRowColumnList()+`
  FROM unnest($2::text[]) WITH ORDINALITY AS selected(selected_table_id, table_rank)
  JOIN network_flow_rows
    ON incident_id = $1
   AND network_flow_table_id = selected.selected_table_id
 ORDER BY selected.table_rank ASC, flow_start_utc ASC, flow_end_utc ASC,
          source_row_number ASC, network_flow_row_id ASC
`, incidentID, tableIDs)
	if err != nil {
		return fmt.Errorf("iterate network flow rows for tables: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		record, err := scanFlowRow(rows)
		if err != nil {
			return err
		}
		if err := visit(record); err != nil {
			return fmt.Errorf("visit network flow table row: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan network flow table rows: %w", err)
	}
	return ctx.Err()
}

func (s *graphSourceReadStore) IterateGraphContributorRows(ctx context.Context, incidentID uuid.UUID, tableIDs []string, predicate graphContributorPredicate, visit func(flowRow) error) error {
	if len(tableIDs) == 0 {
		return nil
	}
	base := `
SELECT ` + flowRowColumnList() + `
  FROM unnest($2::text[]) WITH ORDINALITY AS selected(selected_table_id, table_rank)
  JOIN network_flow_rows
    ON incident_id = $1
   AND network_flow_table_id = selected.selected_table_id
 WHERE `
	var query string
	var args []any
	switch predicate.Kind {
	case "vertex":
		query = base + `(src_ip = $3 OR dst_ip = $3)`
		args = []any{incidentID, tableIDs, predicate.EndpointValue}
	case "default_edge":
		query = base + `src_ip = $3
   AND dst_ip = $4
   AND ip_protocol = $5
   AND (($6::boolean AND dst_port = $7) OR (NOT $6::boolean AND dst_port IS NULL))`
		args = []any{incidentID, tableIDs, predicate.SourceEndpointValue, predicate.DestinationEndpointValue, predicate.Protocol, predicate.DestinationPort != nil, predicate.DestinationPort}
	case "time_bucket_edge":
		if predicate.BucketStartUTC == nil || predicate.BucketEndUTC == nil {
			return fmt.Errorf("incomplete time-bucket graph contributor predicate")
		}
		query = base + `src_ip = $3
   AND dst_ip = $4
   AND ip_protocol = $5
   AND (($6::boolean AND dst_port = $7) OR (NOT $6::boolean AND dst_port IS NULL))
   AND flow_start_utc >= $8
   AND flow_start_utc < $9`
		args = []any{
			incidentID, tableIDs, predicate.SourceEndpointValue, predicate.DestinationEndpointValue,
			predicate.Protocol, predicate.DestinationPort != nil, predicate.DestinationPort,
			predicate.BucketStartUTC.UTC(), predicate.BucketEndUTC.UTC(),
		}
	default:
		return fmt.Errorf("unknown graph contributor predicate")
	}
	query += `
 ORDER BY selected.table_rank ASC, flow_start_utc ASC, flow_end_utc ASC,
          source_row_number ASC, network_flow_row_id ASC`
	rows, err := s.reader.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("iterate graph contributor rows: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		record, err := scanFlowRow(rows)
		if err != nil {
			return err
		}
		if err := visit(record); err != nil {
			return fmt.Errorf("visit graph contributor row: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("scan graph contributor rows: %w", err)
	}
	return ctx.Err()
}
