package networkflow

func acceptedRowsResource(outcome *acceptedRowsOutcome) map[string]any {
	if outcome == nil {
		return nil
	}
	rows := make([]any, 0, len(outcome.Rows))
	for _, row := range outcome.Rows {
		rows = append(rows, rowResource(row))
	}
	return map[string]any{
		"table_scope": map[string]any{"mode": outcome.Mode, "table_ids": outcome.TableIDs},
		"rows":        rows,
		"meta": map[string]any{
			"query":  acceptedRowsQueryEcho(outcome.Filters, outcome.Sort, effectiveSort(outcome.Sort), outcome.TableIDs),
			"paging": map[string]any{"limit": outcome.Limit, "returned_count": len(rows), "next_cursor_token": outcome.NextToken},
		},
	}
}
func diagnosticsResource(tableID string, outcome *diagnosticsOutcome) map[string]any {
	if outcome == nil {
		return nil
	}
	rows := make([]any, 0, len(outcome.Diagnostics))
	for _, row := range outcome.Diagnostics {
		rows = append(rows, diagnosticResource(row))
	}
	return map[string]any{
		"schema_id": "cartulary.network_flow.rejected_rows_query_result.v1", "network_flow_table_id": tableID, "diagnostics": rows,
		"meta": map[string]any{"query": rejectedRowsQueryEcho(outcome.Request), "paging": map[string]any{"limit": outcome.Limit, "returned_count": len(rows), "next_cursor_token": outcome.NextToken}},
	}
}
