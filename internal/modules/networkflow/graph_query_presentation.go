package networkflow

func graphContributorsResource(outcome *graphContributorsOutcome) map[string]any {
	if outcome == nil {
		return nil
	}
	contributors := make([]any, 0, len(outcome.Rows))
	for _, row := range outcome.Rows {
		contributors = append(contributors, map[string]any{"row_ref": rowRefResource(row), "row": rowResource(row)})
	}
	return map[string]any{"schema_id": schemaGraphContributorQueryResult, "graph_query_digest": outcome.Digest, "selector": graphSelectorResource(outcome.Selector), "contributors": contributors, "meta": map[string]any{"paging": map[string]any{"limit": outcome.Limit, "returned_count": len(contributors), "next_cursor_token": outcome.NextToken}}}
}

func savedGraphContributorsResource(outcome *savedGraphContributorsOutcome) map[string]any {
	if outcome == nil {
		return nil
	}
	result := graphContributorsResource(&outcome.Page)
	result["schema_id"] = "cartulary.network_flow.graph_view_contributor_query_result.v2"
	result["graph_view_id"], result["projection_result_id"] = outcome.GraphViewID, outcome.ProjectionResultID
	delete(result, "graph_query_digest")
	return result
}
