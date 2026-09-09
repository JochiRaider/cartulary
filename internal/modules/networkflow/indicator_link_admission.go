package networkflow

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
)

var (
	linkTableIDPattern    = regexp.MustCompile(`^nft_[a-f0-9]{32}$`)
	linkRowIDPattern      = regexp.MustCompile(`^nfr_[a-f0-9]{64}$`)
	linkEndpointIDPattern = regexp.MustCompile(`^nfe_[a-f0-9]{64}$`)
	linkFlowEdgeIDPattern = regexp.MustCompile(`^nff_[a-f0-9]{64}$`)
	linkBindingIDPattern  = regexp.MustCompile(`^nfb_[a-f0-9]{32}$`)
	linkDigestPattern     = regexp.MustCompile(`^[a-f0-9]{64}$`)
	linkUUIDPattern       = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
)

// Link admission is closed and deterministic. Semantic checks deliberately
// follow committed replay and current resource checks, including field policy.
func decodeIndicatorLinkRequest(r *http.Request, limits EffectiveLimits) (indicatorLinkRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
	if apiErr != nil {
		return indicatorLinkRequest{}, completeLinkError(apiErr, indicatorLinkRequest{}, "")
	}
	return decodeIndicatorLinkObject(raw, limits)
}

func decodeIndicatorLinkObject(raw map[string]json.RawMessage, limits EffectiveLimits) (request indicatorLinkRequest, apiErr *httpapi.APIError) {
	defer func() { apiErr = completeLinkError(apiErr, request, "") }()
	if apiErr = linkMembers(raw, map[string]string{"schema_id": "string", "client_txn_id": "string", "selector": "object", "target": "object", "observation_mode": "string", "confirm_exact_value": "string"}); apiErr != nil {
		return
	}
	if linkString(raw, "schema_id") != schemaIndicatorLinkRequest {
		apiErr = invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
		return
	}
	request.ClientTxnID = linkString(raw, "client_txn_id")
	if request.ClientTxnID == "" || utf8.RuneCountInString(request.ClientTxnID) > 160 {
		apiErr = invalidNetworkFlowRequest("client_txn_id", "type_mismatch")
		return
	}
	if linkString(raw, "observation_mode") != "binding_only" {
		apiErr = invalidNetworkFlowRequest("observation_mode", "type_mismatch")
		return
	}
	request.ConfirmExactValue = linkString(raw, "confirm_exact_value")
	if n := utf8.RuneCountInString(request.ConfirmExactValue); n < 1 || n > 45 {
		apiErr = invalidNetworkFlowRequest("confirm_exact_value", "type_mismatch")
		return
	}
	var targetContext map[string]json.RawMessage
	_ = json.Unmarshal(raw["target"], &targetContext)
	request.Target.Mode = linkString(targetContext, "mode")
	request.Selector, apiErr = decodeIndicatorSelector(raw["selector"], limits)
	if apiErr != nil {
		return
	}
	request.Target, apiErr = decodeIndicatorTarget(raw["target"])
	return
}

func decodeIndicatorSelector(raw json.RawMessage, limits EffectiveLimits) (selector indicatorLinkSelector, apiErr *httpapi.APIError) {
	object, apiErr := linkObject(raw, "selector")
	if apiErr != nil {
		return selector, apiErr
	}
	selector.Kind = linkString(object, "kind")
	members := map[string]string{"kind": "string"}
	switch selector.Kind {
	case "row_field_value":
		members["network_flow_table_id"], members["network_flow_row_id"], members["field_key"] = "string", "string", "string"
	case "row_refs":
		members["row_refs"], members["field_key"] = "array", "string"
	case "graph_vertex", "graph_edge":
		members["graph_query"], members["graph_query_digest"] = "object", "string"
		if selector.Kind == "graph_vertex" {
			members["vertex_id"] = "string"
		} else {
			members["edge_id"], members["field_key"] = "string", "string"
		}
	default:
		return selector, invalidIndicatorSelector("kind", "unknown_selector_kind")
	}
	if apiErr := linkMembers(object, members); apiErr != nil {
		if apiErr.Details["reason_code"] == "unknown_member" {
			apiErr = invalidIndicatorSelector("selector", "variant_member_conflict")
		}
		return selector, apiErr
	}
	selector.FieldKey = linkString(object, "field_key")
	switch selector.Kind {
	case "row_field_value":
		selector.TableID, selector.RowID = linkString(object, "network_flow_table_id"), linkString(object, "network_flow_row_id")
		if !linkRowIDPattern.MatchString(selector.RowID) {
			return selector, invalidNetworkFlowRequest("network_flow_row_id", "type_mismatch")
		}
		if !linkTableIDPattern.MatchString(selector.TableID) {
			return selector, invalidNetworkFlowRequest("network_flow_table_id", "type_mismatch")
		}
	case "row_refs":
		var refs []json.RawMessage
		_ = json.Unmarshal(object["row_refs"], &refs)
		if len(refs) == 0 || len(refs) > 1000 {
			return selector, invalidNetworkFlowRequest("row_refs", "type_mismatch")
		}
		seen := map[string]bool{}
		for _, rawRef := range refs {
			ref, apiErr := decodeLinkRowRef(rawRef)
			if apiErr != nil {
				return selector, apiErr
			}
			if seen[ref.NetworkFlowRowID] {
				return selector, invalidIndicatorSelector("row_refs", "variant_member_conflict")
			}
			seen[ref.NetworkFlowRowID] = true
			selector.RowRefs = append(selector.RowRefs, ref)
		}
	case "graph_vertex", "graph_edge":
		selector.GraphQuery, apiErr = decodeLinkGraphQuery(object["graph_query"], limits)
		if apiErr != nil {
			return selector, apiErr
		}
		if selector.GraphQuery.SchemaID != schemaGraphSemanticQueryV2 {
			return selector, invalidNetworkFlowRequest("selector.graph_query.schema_id", "invalid_schema_id")
		}
		selector.GraphQueryDigest = linkString(object, "graph_query_digest")
		if !linkDigestPattern.MatchString(selector.GraphQueryDigest) {
			return selector, invalidNetworkFlowRequest("graph_query_digest", "type_mismatch")
		}
		selector.VertexID, selector.EdgeID = linkString(object, "vertex_id"), linkString(object, "edge_id")
		if selector.Kind == "graph_vertex" && !linkEndpointIDPattern.MatchString(selector.VertexID) {
			return selector, invalidNetworkFlowRequest("vertex_id", "type_mismatch")
		}
		if selector.Kind == "graph_edge" && !linkFlowEdgeIDPattern.MatchString(selector.EdgeID) {
			return selector, invalidNetworkFlowRequest("edge_id", "type_mismatch")
		}
	}
	return selector, nil
}

func decodeIndicatorTarget(raw json.RawMessage) (target indicatorLinkTarget, apiErr *httpapi.APIError) {
	object, apiErr := linkObject(raw, "target")
	if apiErr != nil {
		return target, apiErr
	}
	target.Mode = linkString(object, "mode")
	members := map[string]string{"mode": "string"}
	switch target.Mode {
	case "existing_indicator":
		members["indicator_id"] = "string"
	case "create_indicator":
		members["indicator_type"] = "string"
	default:
		return target, invalidNetworkFlowRequest("target.mode", "type_mismatch")
	}
	if apiErr := linkMembers(object, members); apiErr != nil {
		return target, apiErr
	}
	if target.Mode == "existing_indicator" {
		var err error
		target.IndicatorID, err = uuid.Parse(linkString(object, "indicator_id"))
		if err != nil || target.IndicatorID == uuid.Nil || !linkUUIDPattern.MatchString(linkString(object, "indicator_id")) {
			return target, invalidNetworkFlowRequest("target.indicator_id", "type_mismatch")
		}
	} else {
		target.IndicatorType = linkString(object, "indicator_type")
	}
	return target, nil
}

func decodeLinkRowRef(raw json.RawMessage) (NetworkFlowRowRef, *httpapi.APIError) {
	object, apiErr := linkObject(raw, "row_refs")
	if apiErr != nil {
		return NetworkFlowRowRef{}, apiErr
	}
	if apiErr := linkMembers(object, map[string]string{"network_flow_table_id": "string", "network_flow_row_id": "string", "source_row_number": "number", "mapping_fingerprint": "string"}); apiErr != nil {
		return NetworkFlowRowRef{}, apiErr
	}
	var ref NetworkFlowRowRef
	if json.Unmarshal(raw, &ref) != nil || !validLinkRowRef(ref) {
		return ref, invalidNetworkFlowRequest("row_refs", "type_mismatch")
	}
	return ref, nil
}

func validLinkRowRef(ref NetworkFlowRowRef) bool {
	return linkTableIDPattern.MatchString(ref.NetworkFlowTableID) && linkRowIDPattern.MatchString(ref.NetworkFlowRowID) && linkDigestPattern.MatchString(ref.MappingFingerprint) && ref.SourceRowNumber > 0
}

func linkObject(raw json.RawMessage, field string) (map[string]json.RawMessage, *httpapi.APIError) {
	if len(raw) == 0 {
		return nil, invalidNetworkFlowRequest(field, "missing_member")
	}
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, invalidNetworkFlowRequest(field, "explicit_null")
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || object == nil {
		return nil, invalidNetworkFlowRequest(field, "type_mismatch")
	}
	return object, nil
}

func linkMembers(object map[string]json.RawMessage, members map[string]string) *httpapi.APIError {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, exists := members[key]; !exists {
			return invalidNetworkFlowRequest(key, "unknown_member")
		}
	}
	keys = keys[:0]
	for key := range members {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, exists := object[key]; !exists {
			return invalidNetworkFlowRequest(key, "missing_member")
		}
	}
	for _, key := range keys {
		if bytes.Equal(bytes.TrimSpace(object[key]), []byte("null")) && !strings.HasPrefix(members[key], "nullable:") {
			return invalidNetworkFlowRequest(key, "explicit_null")
		}
	}
	for _, key := range keys {
		raw := object[key]
		var value any
		_ = json.Unmarshal(raw, &value)
		if value == nil {
			continue
		}
		kind := ""
		switch value.(type) {
		case string:
			kind = "string"
		case float64:
			kind = "number"
		case map[string]any:
			kind = "object"
		case []any:
			kind = "array"
		case bool:
			kind = "boolean"
		}
		expected := strings.TrimPrefix(members[key], "nullable:")
		if expected != "any" && kind != expected {
			err := invalidNetworkFlowRequest(key, "type_mismatch")
			err.Details["actual_kind"] = kind
			return err
		}
	}
	return nil
}

func linkString(object map[string]json.RawMessage, key string) string {
	var value string
	_ = json.Unmarshal(object[key], &value)
	return value
}

func completeLinkError(apiErr *httpapi.APIError, request indicatorLinkRequest, candidate string) *httpapi.APIError {
	if apiErr == nil {
		return nil
	}
	d := apiErr.Details
	if d == nil {
		d = map[string]any{}
		apiErr.Details = d
	}
	if strings.HasPrefix(apiErr.Code, "network_flow_") {
		if _, ok := d["retry_action"]; !ok {
			d["retry_action"] = "correct_request"
		}
	}
	if apiErr.Code == "network_flow_invalid_table_scope" {
		completeLinkGraphError(apiErr, request.Selector.GraphQuery.SelectedTableIDs)
	}
	switch apiErr.Code {
	case "network_flow_invalid_request":
		for _, key := range []string{"field", "actual_kind"} {
			if _, exists := d[key]; !exists {
				d[key] = nil
			}
		}
		d["expected_contract"] = schemaIndicatorLinkRequest
		d["retry_action"] = "correct_request"
	case "network_flow_invalid_indicator_selector", "network_flow_invalid_indicator_target", "network_flow_indicator_link_ambiguous", "network_flow_indicator_link_forbidden":
		d["selector_kind"], d["field_key"], d["target_mode"] = linkNullable(request.Selector.Kind), linkNullable(request.Selector.FieldKey), linkNullable(request.Target.Mode)
		if existing, ok := d["resolved_candidate_value"].(string); ok && candidate == "" {
			candidate = existing
		}
		d["resolved_candidate_value"] = linkNullable(candidate)
		d["retry_action"] = "correct_request"
		if apiErr.Status == http.StatusForbidden {
			d["retry_action"] = "do_not_retry"
		}
	}
	return apiErr
}

func linkNullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func linkSourceTableError(apiErr *httpapi.APIError, tableID string) *httpapi.APIError {
	if apiErr.Code == "network_flow_table_not_active" || apiErr.Code == "network_flow_table_not_found" {
		apiErr.Details["network_flow_table_id"] = tableID
		apiErr.Details["table_status"] = nil
		apiErr.Details["allowed_states"] = []string{"active"}
		apiErr.Details["retry_action"] = "refresh_resource"
		if apiErr.Code == "network_flow_table_not_active" {
			apiErr.Details["table_status"] = TableStatusSoftDeleted
		}
	}
	return apiErr
}
