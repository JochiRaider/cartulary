package networkflow

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const refreshIntentSchema = "cartulary.network_flow.refresh_intent.v1"

type graphRefreshIntent struct {
	SchemaID      string          `json:"schema_id"`
	Scope         string          `json:"scope"`
	SemanticQuery json.RawMessage `json:"semantic_query"`
}

func encodeRefreshIntent(scope string, semantic json.RawMessage) (json.RawMessage, error) {
	if scope != "selected_tables" && scope != "all_active_tables" {
		return nil, errGraphViewDeclarationInvalid
	}
	return json.Marshal(graphRefreshIntent{SchemaID: refreshIntentSchema, Scope: scope, SemanticQuery: semantic})
}

func decodeRefreshIntent(raw json.RawMessage, limits EffectiveLimits) (graphRefreshIntent, graphSemanticRequest, error) {
	var intent graphRefreshIntent
	if _, err := decodeClosedImportObject(raw, "schema_id scope semantic_query", "", "", &intent); err != nil || intent.SchemaID != refreshIntentSchema {
		return intent, graphSemanticRequest{}, errGraphViewDeclarationInvalid
	}
	semantic, failure := decodeGraphSemanticRequest(intent.SemanticQuery, limits)
	if failure != nil || (intent.Scope != "selected_tables" && intent.Scope != "all_active_tables") {
		return intent, graphSemanticRequest{}, errGraphViewDeclarationInvalid
	}
	return intent, semantic, nil
}

func resolveRefreshIntentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, raw json.RawMessage, limits EffectiveLimits) (graphSemanticRequest, error) {
	intent, semantic, err := decodeRefreshIntent(raw, limits)
	if err != nil {
		return semantic, err
	}
	scope := tableScope{Mode: intent.Scope}
	switch intent.Scope {
	case "selected_tables":
		scope.SelectedTableIDs = semantic.SelectedTableIDs
	case "all_active_tables":
	default:
		return graphSemanticRequest{}, errGraphViewDeclarationInvalid
	}
	_, ids, _, failure := resolveActiveTableScope(ctx, &graphSourceReadStore{reader: tx}, limits, incidentID, scope)
	if failure != nil {
		return graphSemanticRequest{}, failure
	}
	semantic.SelectedTableIDs = ids
	semantic.Raw = graphSemanticQueryResource(ids, semantic.Filters, semantic.TimeRange, semantic.Aggregation)
	return semantic, nil
}
