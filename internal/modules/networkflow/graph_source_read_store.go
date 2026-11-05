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
		query = base + `(endpoint_a_ip = $3 OR endpoint_b_ip = $3)`
		args = []any{incidentID, tableIDs, predicate.EndpointValue}
	case "default_edge", "time_bucket_edge":
		key := predicate.Semantics
		if !key.valid() {
			return fmt.Errorf("invalid graph contributor semantics")
		}
		firstIP, secondIP := "endpoint_a_ip::inet", "endpoint_b_ip::inet"
		secondPort := "endpoint_b_port"
		if key.DirectionKind != directionUnknown {
			firstIP, _ = rowSQLExpression(fieldFromIP)
			secondIP, _ = rowSQLExpression(fieldToIP)
			secondPort, _ = rowSQLExpression(fieldToPort)
		}
		query = base + "direction_kind = $3 AND ip_protocol = $4 AND " + firstIP + " = $5::inet AND " + secondIP + " = $6::inet AND " + secondPort + " IS NOT DISTINCT FROM $7::integer"
		args = []any{incidentID, tableIDs, key.DirectionKind, key.Protocol, key.First.IP, key.Second.IP, key.Second.Port}
		if key.DirectionKind == directionUnknown {
			query += " AND endpoint_a_port IS NOT DISTINCT FROM $8::integer"
		} else {
			query += " AND $8::integer IS NULL"
		}
		args = append(args, key.First.Port)
		query += " AND EXISTS (SELECT 1 FROM network_flow_tables t WHERE t.incident_id=network_flow_rows.incident_id AND t.network_flow_table_id=network_flow_rows.network_flow_table_id AND t.source_profile_id=ANY($9::text[]))"
		args = append(args, sourceProfilesForCounterSemantics(key.CounterSemantics))
		if predicate.Kind == "time_bucket_edge" {
			if predicate.BucketStartUTC == nil || predicate.BucketEndUTC == nil {
				return fmt.Errorf("incomplete time-bucket graph contributor predicate")
			}
			query += " AND flow_start_utc >= $10 AND flow_start_utc < $11"
			args = append(args, predicate.BucketStartUTC.UTC(), predicate.BucketEndUTC.UTC())
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
