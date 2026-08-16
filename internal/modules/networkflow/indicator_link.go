package networkflow

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	schemaIndicatorLinkRequest   = "cartulary.network_flow.indicator_link_request.v1"
	schemaIndicatorLinkResult    = "cartulary.network_flow.indicator_link_result.v1"
	routeKeyIndicatorLinksCreate = "nf.indicator_links.create"
)

type indicatorLinkRequest struct {
	ClientTxnID       string
	Selector          indicatorLinkSelector
	Target            indicatorLinkTarget
	ConfirmExactValue string
}

type indicatorLinkSelector struct {
	Kind             string
	TableID          string
	RowID            string
	FieldKey         string
	RowRefs          []NetworkFlowRowRef
	GraphQuery       graphSemanticRequest
	GraphQueryDigest string
	VertexID         string
	EdgeID           string
}

type indicatorLinkTarget struct {
	Mode          string
	IndicatorID   uuid.UUID
	IndicatorType string
}

type resolvedIndicatorLinkSelector struct {
	SelectorKind            string
	CandidateValue          string
	SourceRowRefs           []NetworkFlowRowRef
	SourceRowRefsTruncated  bool
	SourceRowRefsTotalCount int64
}

func (s *Service) handleIndicatorLinks(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := parseIncidentPathValue(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeIndicatorLinkRequest(r, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	payload, status, apiErr := s.commitIndicatorLinkRoute(r.Context(), incidentID, principal.User, request, indicatorLinkRequestHash(request), httpapi.RequestIDFromContext(r.Context()))
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, status, payload)
}

func decodeIndicatorLinkRequest(r *http.Request, limits EffectiveLimits) (indicatorLinkRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
	if apiErr != nil {
		return indicatorLinkRequest{}, apiErr
	}
	schemaID, apiErr := requiredJSONString(raw, "schema_id")
	if apiErr != nil {
		return indicatorLinkRequest{}, apiErr
	}
	if schemaID != schemaIndicatorLinkRequest {
		return indicatorLinkRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "client_txn_id", "selector", "target", "observation_mode", "confirm_exact_value"); apiErr != nil {
		return indicatorLinkRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredJSONString(raw, "client_txn_id")
	if apiErr != nil {
		return indicatorLinkRequest{}, apiErr
	}
	mode, apiErr := requiredJSONString(raw, "observation_mode")
	if apiErr != nil {
		return indicatorLinkRequest{}, apiErr
	}
	if mode != "binding_only" {
		return indicatorLinkRequest{}, invalidNetworkFlowRequest("observation_mode", "invalid_value")
	}
	confirm, apiErr := requiredJSONString(raw, "confirm_exact_value")
	if apiErr != nil {
		return indicatorLinkRequest{}, apiErr
	}
	selector, apiErr := decodeIndicatorSelector(raw["selector"], limits)
	if apiErr != nil {
		return indicatorLinkRequest{}, apiErr
	}
	target, apiErr := decodeIndicatorTarget(raw["target"])
	if apiErr != nil {
		return indicatorLinkRequest{}, apiErr
	}
	return indicatorLinkRequest{ClientTxnID: clientTxnID, Selector: selector, Target: target, ConfirmExactValue: confirm}, nil
}

func decodeIndicatorSelector(raw json.RawMessage, limits EffectiveLimits) (indicatorLinkSelector, *httpapi.APIError) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return indicatorLinkSelector{}, invalidIndicatorSelector("", "missing_member")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return indicatorLinkSelector{}, invalidIndicatorSelector("selector", "type_mismatch")
	}
	kind, apiErr := requiredJSONString(object, "kind")
	if apiErr != nil {
		return indicatorLinkSelector{}, invalidIndicatorSelector("kind", "unknown_selector_kind")
	}
	switch kind {
	case "row_field_value":
		if apiErr := ensureAllowedMembers(object, "kind", "network_flow_table_id", "network_flow_row_id", "field_key"); apiErr != nil {
			return indicatorLinkSelector{}, invalidIndicatorSelector("selector", "variant_member_conflict")
		}
		tableID, apiErr := requiredJSONString(object, "network_flow_table_id")
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		rowID, apiErr := requiredJSONString(object, "network_flow_row_id")
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		fieldKey, apiErr := requiredJSONString(object, "field_key")
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		if !networkFlowLinkableIPField(fieldKey) {
			return indicatorLinkSelector{}, invalidIndicatorSelector("field_key", "field_not_linkable")
		}
		return indicatorLinkSelector{Kind: kind, TableID: tableID, RowID: rowID, FieldKey: fieldKey}, nil
	case "row_refs":
		if apiErr := ensureAllowedMembers(object, "kind", "row_refs", "field_key"); apiErr != nil {
			return indicatorLinkSelector{}, invalidIndicatorSelector("selector", "variant_member_conflict")
		}
		fieldKey, apiErr := requiredJSONString(object, "field_key")
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		if !networkFlowLinkableIPField(fieldKey) {
			return indicatorLinkSelector{}, invalidIndicatorSelector("field_key", "field_not_linkable")
		}
		var refs []NetworkFlowRowRef
		if err := json.Unmarshal(object["row_refs"], &refs); err != nil {
			return indicatorLinkSelector{}, invalidIndicatorSelector("row_refs", "type_mismatch")
		}
		if len(refs) == 0 || int64(len(refs)) > limits.MaxBindingSourceRowRefs {
			return indicatorLinkSelector{}, invalidIndicatorSelector("row_refs", "row_ref_count_invalid")
		}
		seen := map[string]struct{}{}
		for _, ref := range refs {
			if ref.NetworkFlowTableID == "" || ref.NetworkFlowRowID == "" || ref.SourceRowNumber <= 0 || ref.MappingFingerprint == "" {
				return indicatorLinkSelector{}, invalidIndicatorSelector("row_refs", "invalid_row_ref")
			}
			if _, exists := seen[ref.NetworkFlowRowID]; exists {
				return indicatorLinkSelector{}, invalidIndicatorSelector("row_refs", "duplicate_row_ref")
			}
			seen[ref.NetworkFlowRowID] = struct{}{}
		}
		return indicatorLinkSelector{Kind: kind, RowRefs: refs, FieldKey: fieldKey}, nil
	case "graph_vertex":
		if apiErr := ensureAllowedMembers(object, "kind", "graph_query", "graph_query_digest", "vertex_id"); apiErr != nil {
			return indicatorLinkSelector{}, invalidIndicatorSelector("selector", "variant_member_conflict")
		}
		graphQuery, apiErr := decodeGraphSemanticRequest(object["graph_query"], limits)
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		if graphQuery.SchemaID != schemaGraphSemanticQueryV2 {
			return indicatorLinkSelector{}, invalidNetworkFlowRequest("selector.graph_query.schema_id", "invalid_schema_id")
		}
		digest, apiErr := requiredJSONString(object, "graph_query_digest")
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		vertexID, apiErr := requiredJSONString(object, "vertex_id")
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		return indicatorLinkSelector{Kind: kind, GraphQuery: graphQuery, GraphQueryDigest: digest, VertexID: vertexID}, nil
	case "graph_edge":
		if apiErr := ensureAllowedMembers(object, "kind", "graph_query", "graph_query_digest", "edge_id", "field_key"); apiErr != nil {
			return indicatorLinkSelector{}, invalidIndicatorSelector("selector", "variant_member_conflict")
		}
		graphQuery, apiErr := decodeGraphSemanticRequest(object["graph_query"], limits)
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		if graphQuery.SchemaID != schemaGraphSemanticQueryV2 {
			return indicatorLinkSelector{}, invalidNetworkFlowRequest("selector.graph_query.schema_id", "invalid_schema_id")
		}
		digest, apiErr := requiredJSONString(object, "graph_query_digest")
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		edgeID, apiErr := requiredJSONString(object, "edge_id")
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		fieldKey, apiErr := requiredJSONString(object, "field_key")
		if apiErr != nil {
			return indicatorLinkSelector{}, apiErr
		}
		if !networkFlowLinkableIPField(fieldKey) {
			return indicatorLinkSelector{}, invalidIndicatorSelector("field_key", "field_not_linkable")
		}
		return indicatorLinkSelector{Kind: kind, GraphQuery: graphQuery, GraphQueryDigest: digest, EdgeID: edgeID, FieldKey: fieldKey}, nil
	default:
		return indicatorLinkSelector{}, invalidIndicatorSelector("kind", "unknown_selector_kind")
	}
}

func decodeIndicatorTarget(raw json.RawMessage) (indicatorLinkTarget, *httpapi.APIError) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return indicatorLinkTarget{}, invalidIndicatorTarget("", "missing_member")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return indicatorLinkTarget{}, invalidIndicatorTarget("target", "type_mismatch")
	}
	mode, apiErr := requiredJSONString(object, "mode")
	if apiErr != nil {
		return indicatorLinkTarget{}, invalidIndicatorTarget("mode", "unknown_target_mode")
	}
	switch mode {
	case "existing_indicator":
		if apiErr := ensureAllowedMembers(object, "mode", "indicator_id"); apiErr != nil {
			return indicatorLinkTarget{}, invalidIndicatorTarget("target", "variant_member_conflict")
		}
		indicatorText, apiErr := requiredJSONString(object, "indicator_id")
		if apiErr != nil {
			return indicatorLinkTarget{}, apiErr
		}
		indicatorID, err := uuid.Parse(indicatorText)
		if err != nil {
			return indicatorLinkTarget{}, invalidIndicatorTarget("indicator_id", "invalid_value")
		}
		return indicatorLinkTarget{Mode: mode, IndicatorID: indicatorID}, nil
	case "create_indicator":
		if apiErr := ensureAllowedMembers(object, "mode", "indicator_type"); apiErr != nil {
			return indicatorLinkTarget{}, invalidIndicatorTarget("target", "variant_member_conflict")
		}
		indicatorType, apiErr := requiredJSONString(object, "indicator_type")
		if apiErr != nil {
			return indicatorLinkTarget{}, apiErr
		}
		if indicatorType != "ipv4_addr" && indicatorType != "ipv6_addr" {
			return indicatorLinkTarget{}, invalidIndicatorTarget("indicator_type", "target_type_mismatch")
		}
		return indicatorLinkTarget{Mode: mode, IndicatorType: indicatorType}, nil
	default:
		return indicatorLinkTarget{}, invalidIndicatorTarget("mode", "unknown_target_mode")
	}
}

func (s *Service) commitIndicatorLinkRoute(ctx context.Context, incidentID uuid.UUID, actor authn.UserRecord, request indicatorLinkRequest, requestHash []byte, requestID string) (map[string]any, int, *httpapi.APIError) {
	idempotencyKey := indicatorLinkIdempotencyKey(actor.ID, incidentID, request.ClientTxnID)
	if payload, status, replayed, apiErr := s.replayIndicatorLinkIfPresent(ctx, idempotencyKey, requestHash, incidentID); replayed || apiErr != nil {
		return payload, status, apiErr
	}
	resolved, apiErr := s.resolveIndicatorSelector(ctx, incidentID, actor.ID, request.Selector)
	if apiErr != nil {
		return nil, 0, apiErr
	}
	if !canonicalIPLiteral(request.ConfirmExactValue) || request.ConfirmExactValue != resolved.CandidateValue {
		return nil, 0, indicatorLinkAmbiguous(resolved.SelectorKind, request.Selector.FieldKey, "candidate_mismatch", resolved.CandidateValue)
	}
	targetType := indicatorTypeForIP(resolved.CandidateValue)
	if request.Target.Mode == "create_indicator" && request.Target.IndicatorType != targetType {
		return nil, 0, invalidIndicatorTarget("indicator_type", "target_type_mismatch")
	}
	if s.transactions == nil {
		return nil, 0, httpapi.InternalAPIError(crossownertransaction.ErrUnavailable)
	}
	participant := &indicatorLinkParticipant{mutation: indicatorLinkMutation{
		IncidentID: incidentID, Actor: actor, Request: request, Resolved: resolved,
		TargetType: targetType, RequestHash: append([]byte(nil), requestHash...),
		RequestID: requestID, Now: s.now(), SafeDigester: s.safeDigester,
	}}
	result, err := s.transactions.Execute(ctx, crossownertransaction.Operation{
		OperationID:             "network-flow-indicator-link:" + incidentID.String() + ":" + request.ClientTxnID,
		NormalizedRequestSHA256: hex.EncodeToString(requestHash),
		Participants:            []crossownertransaction.Participant{participant},
	})
	if err != nil {
		if payload, status, replayed, replayErr := s.replayIndicatorLinkIfPresent(ctx, idempotencyKey, requestHash, incidentID); replayed || replayErr != nil {
			return payload, status, replayErr
		}
		if authn.IsUniqueViolation(err) {
			return nil, 0, httpapi.ClientTxnConflictError(request.ClientTxnID)
		}
		if errors.Is(err, indicators.ErrIndicatorNotFound) {
			return nil, 0, indicatorLinkForbidden(request.Selector.Kind, request.Selector.FieldKey, request.Target.Mode, "target_not_visible")
		}
		var validation *indicators.IndicatorCreateValidationError
		if errors.As(err, &validation) {
			return nil, 0, invalidIndicatorTarget(validation.Field, validation.ReasonCode)
		}
		if reason := indicatorParticipantReason(err); reason != "" {
			return nil, 0, invalidIndicatorTargetWithContext(request.Selector.Kind, request.Selector.FieldKey, request.Target.Mode, reason)
		}
		var conflict *crossownertransaction.ConflictError
		if errors.As(err, &conflict) {
			return nil, 0, &httpapi.APIError{
				Status: http.StatusConflict, Code: "transaction_conflict", Retryable: true,
				Details: map[string]any{"reason_code": conflict.ReasonCode},
			}
		}
		if errors.Is(err, crossownertransaction.ErrTimeout) {
			return nil, 0, &httpapi.APIError{
				Status: http.StatusServiceUnavailable, Code: "service_unavailable", Retryable: true,
				Details: map[string]any{"reason_code": "extension_transaction_timeout"},
			}
		}
		if errors.Is(err, ErrInvalidStorageArgument) {
			return nil, 0, invalidIndicatorTarget("target", "target_value_mismatch")
		}
		return nil, 0, httpapi.InternalAPIError(err)
	}
	committed, resultErr := participantResult[indicatorLinkCommitResult](result, IndicatorLinkParticipantID)
	if resultErr != nil {
		return nil, 0, httpapi.InternalAPIError(resultErr)
	}
	return committed.Payload, committed.Status, nil
}

func (s *Service) replayIndicatorLinkIfPresent(ctx context.Context, key authn.RouteIdempotencyKey, requestHash []byte, incidentID uuid.UUID) (map[string]any, int, bool, *httpapi.APIError) {
	existing, err := s.authStore.GetRouteIdempotency(ctx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return nil, 0, true, httpapi.ClientTxnConflictError(key.ClientTxnID)
		}
		payload, err := decodeStoredNetworkFlowResponse(existing.ResponseJSON)
		if err != nil {
			return nil, 0, true, httpapi.InternalAPIError(err)
		}
		indicatorID, err := indicatorIDFromLinkPayload(payload)
		if err != nil {
			return nil, 0, true, httpapi.InternalAPIError(err)
		}
		if _, err := s.store.GetActiveIndicator(ctx, incidentID, indicatorID); err != nil {
			if errors.Is(err, ErrTableNotFound) {
				return nil, 0, true, indicatorLinkForbidden("", "", "existing_indicator", "target_not_visible")
			}
			return nil, 0, true, httpapi.InternalAPIError(err)
		}
		return payload, existing.StatusCode, true, nil
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return nil, 0, false, httpapi.InternalAPIError(err)
	}
	return nil, 0, false, nil
}

func (s *Service) resolveIndicatorSelector(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, selector indicatorLinkSelector) (resolvedIndicatorLinkSelector, *httpapi.APIError) {
	switch selector.Kind {
	case "row_field_value":
		row, apiErr := s.getAcceptedRowForLink(ctx, incidentID, selector.TableID, selector.RowID, nil)
		if apiErr != nil {
			return resolvedIndicatorLinkSelector{}, apiErr
		}
		candidate, apiErr := candidateValueFromRow(row, selector.FieldKey)
		if apiErr != nil {
			return resolvedIndicatorLinkSelector{}, apiErr
		}
		return resolvedIndicatorLinkSelector{
			SelectorKind:            selector.Kind,
			CandidateValue:          candidate,
			SourceRowRefs:           []NetworkFlowRowRef{rowRefFromRow(row)},
			SourceRowRefsTotalCount: 1,
		}, nil
	case "row_refs":
		return s.resolveRowRefsSelector(ctx, incidentID, selector)
	case "graph_vertex", "graph_edge":
		return s.resolveGraphSelector(ctx, incidentID, actorUserID, selector)
	default:
		return resolvedIndicatorLinkSelector{}, invalidIndicatorSelector("kind", "unknown_selector_kind")
	}
}

func (s *Service) resolveRowRefsSelector(ctx context.Context, incidentID uuid.UUID, selector indicatorLinkSelector) (resolvedIndicatorLinkSelector, *httpapi.APIError) {
	activeTables, err := s.store.ListActiveTables(ctx, incidentID)
	if err != nil {
		return resolvedIndicatorLinkSelector{}, httpapi.InternalAPIError(err)
	}
	tableRanks := make(map[string]int, len(activeTables))
	for index, table := range activeTables {
		tableRanks[table.TableID] = index
	}
	rows := make([]FlowRow, 0, len(selector.RowRefs))
	candidate := ""
	for _, ref := range selector.RowRefs {
		row, apiErr := s.getAcceptedRowForLink(ctx, incidentID, ref.NetworkFlowTableID, ref.NetworkFlowRowID, &ref)
		if apiErr != nil {
			return resolvedIndicatorLinkSelector{}, apiErr
		}
		value, apiErr := candidateValueFromRow(row, selector.FieldKey)
		if apiErr != nil {
			return resolvedIndicatorLinkSelector{}, apiErr
		}
		if candidate == "" {
			candidate = value
		} else if candidate != value {
			return resolvedIndicatorLinkSelector{}, indicatorLinkAmbiguous(selector.Kind, selector.FieldKey, "candidate_mismatch", candidate)
		}
		rows = append(rows, row)
	}
	sortContributorRows(rows, tableRanks)
	refs := make([]NetworkFlowRowRef, 0, len(rows))
	for _, row := range rows {
		refs = append(refs, rowRefFromRow(row))
	}
	return resolvedIndicatorLinkSelector{
		SelectorKind:            selector.Kind,
		CandidateValue:          candidate,
		SourceRowRefs:           refs,
		SourceRowRefsTotalCount: int64(len(refs)),
	}, nil
}

func (s *Service) resolveGraphSelector(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, selector indicatorLinkSelector) (resolvedIndicatorLinkSelector, *httpapi.APIError) {
	_ = actorUserID
	composition, apiErr := s.composeGraphSourceFromSemantic(ctx, incidentID, selector.GraphQuery)
	if apiErr != nil {
		return resolvedIndicatorLinkSelector{}, apiErr
	}
	if composition.Digest != selector.GraphQueryDigest {
		return resolvedIndicatorLinkSelector{}, graphQueryStale("digest_mismatch", selector.GraphQueryDigest)
	}
	var candidate string
	var predicate graphContributorPredicate
	switch selector.Kind {
	case "graph_vertex":
		vertex := composition.Vertices[selector.VertexID]
		if vertex == nil {
			return resolvedIndicatorLinkSelector{}, graphQueryStale("vertex_not_found", selector.GraphQueryDigest)
		}
		candidate = vertex.EndpointValue
		predicate = graphContributorPredicate{Kind: "vertex", EndpointValue: vertex.EndpointValue}
	case "graph_edge":
		edge := composition.Edges[selector.EdgeID]
		if edge == nil {
			return resolvedIndicatorLinkSelector{}, graphQueryStale("edge_not_found", selector.GraphQueryDigest)
		}
		switch selector.FieldKey {
		case FieldSrcIP:
			candidate = edge.SrcEndpointValue
		case FieldDstIP:
			candidate = edge.DstEndpointValue
		default:
			return resolvedIndicatorLinkSelector{}, invalidIndicatorSelector("field_key", "field_not_linkable")
		}
		predicate = graphContributorPredicate{
			Kind:                     "default_edge",
			SourceEndpointValue:      edge.SrcEndpointValue,
			DestinationEndpointValue: edge.DstEndpointValue,
			Protocol:                 edge.IPProtocol,
			DestinationPort:          cloneInt32(edge.DstPort),
		}
	default:
		return resolvedIndicatorLinkSelector{}, invalidIndicatorSelector("kind", "unknown_selector_kind")
	}
	limit := int(s.store.limits.MaxBindingSourceRowRefs)
	refs := make([]NetworkFlowRowRef, 0, limit)
	var total int64
	err := s.store.IterateGraphContributorRows(ctx, incidentID, composition.SelectedTableIDs, predicate, func(row FlowRow) error {
		matched, matchErr := rowMatchesGraphQuery(row, selector.GraphQuery.Filters, selector.GraphQuery.TimeRange, selector.GraphQuery.Aggregation)
		if matchErr != nil {
			apiErr = matchErr
			return errStopGraphIteration
		}
		if !matched {
			return nil
		}
		total++
		if len(refs) < limit {
			refs = append(refs, rowRefFromRow(row))
		}
		return nil
	})
	if apiErr != nil {
		return resolvedIndicatorLinkSelector{}, apiErr
	}
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return resolvedIndicatorLinkSelector{}, graphProjectionFailedForContext(err)
		}
		return resolvedIndicatorLinkSelector{}, httpapi.InternalAPIError(err)
	}
	if total == 0 {
		return resolvedIndicatorLinkSelector{}, invalidIndicatorSelector("selector", "row_not_accepted")
	}
	return resolvedIndicatorLinkSelector{
		SelectorKind:            selector.Kind,
		CandidateValue:          candidate,
		SourceRowRefs:           refs,
		SourceRowRefsTruncated:  int64(len(refs)) < total,
		SourceRowRefsTotalCount: total,
	}, nil
}

func (s *Service) getAcceptedRowForLink(ctx context.Context, incidentID uuid.UUID, tableID string, rowID string, suppliedRef *NetworkFlowRowRef) (FlowRow, *httpapi.APIError) {
	if apiErr := s.ensureActiveTables(ctx, incidentID, []string{tableID}); apiErr != nil {
		return FlowRow{}, apiErr
	}
	rows, err := s.store.ListRows(ctx, incidentID, tableID)
	if err != nil {
		return FlowRow{}, tableReadError(err)
	}
	for _, row := range rows {
		if row.RowID != rowID {
			continue
		}
		if suppliedRef != nil && (row.SourceRowNumber != suppliedRef.SourceRowNumber || row.MappingFingerprint != suppliedRef.MappingFingerprint) {
			return FlowRow{}, invalidIndicatorSelector("row_refs", "row_not_accepted")
		}
		return row, nil
	}
	return FlowRow{}, invalidIndicatorSelector("network_flow_row_id", "row_not_accepted")
}

func candidateValueFromRow(row FlowRow, fieldKey string) (string, *httpapi.APIError) {
	switch fieldKey {
	case FieldSrcIP:
		return row.SrcIP, nil
	case FieldDstIP:
		return row.DstIP, nil
	default:
		return "", invalidIndicatorSelector("field_key", "field_not_linkable")
	}
}

func rowRefFromRow(row FlowRow) NetworkFlowRowRef {
	return NetworkFlowRowRef{
		NetworkFlowTableID: row.NetworkFlowTableID,
		NetworkFlowRowID:   row.RowID,
		SourceRowNumber:    row.SourceRowNumber,
		MappingFingerprint: row.MappingFingerprint,
	}
}

func indicatorLinkPayload(binding IndicatorBindingRecord, duplicate bool) map[string]any {
	return map[string]any{
		"schema_id": schemaIndicatorLinkResult,
		"binding":   indicatorBindingResource(binding),
		"duplicate": duplicate,
	}
}

func indicatorLinkIdempotencyKey(actorUserID uuid.UUID, incidentID uuid.UUID, clientTxnID string) authn.RouteIdempotencyKey {
	return authn.RouteIdempotencyKey{
		RouteKey:    routeKeyIndicatorLinksCreate,
		ActorUserID: actorUserID,
		ScopeKey:    incidentID.String(),
		ClientTxnID: clientTxnID,
	}
}

func indicatorLinkRequestHash(request indicatorLinkRequest) []byte {
	return networkFlowRequestHash(map[string]any{
		"route_key":           routeKeyIndicatorLinksCreate,
		"selector":            indicatorSelectorHashResource(request.Selector),
		"target":              indicatorTargetHashResource(request.Target),
		"observation_mode":    "binding_only",
		"confirm_exact_value": request.ConfirmExactValue,
	})
}

func indicatorSelectorHashResource(selector indicatorLinkSelector) map[string]any {
	out := map[string]any{"kind": selector.Kind}
	switch selector.Kind {
	case "row_field_value":
		out["network_flow_table_id"] = selector.TableID
		out["network_flow_row_id"] = selector.RowID
		out["field_key"] = selector.FieldKey
	case "row_refs":
		out["row_refs"] = selector.RowRefs
		out["field_key"] = selector.FieldKey
	case "graph_vertex":
		out["graph_query"] = selector.GraphQuery.Raw
		out["graph_query_digest"] = selector.GraphQueryDigest
		out["vertex_id"] = selector.VertexID
	case "graph_edge":
		out["graph_query"] = selector.GraphQuery.Raw
		out["graph_query_digest"] = selector.GraphQueryDigest
		out["edge_id"] = selector.EdgeID
		out["field_key"] = selector.FieldKey
	}
	return out
}

func indicatorTargetHashResource(target indicatorLinkTarget) map[string]any {
	if target.Mode == "existing_indicator" {
		return map[string]any{"mode": target.Mode, "indicator_id": target.IndicatorID.String()}
	}
	return map[string]any{"mode": target.Mode, "indicator_type": target.IndicatorType}
}

func indicatorIDFromLinkPayload(payload map[string]any) (uuid.UUID, error) {
	binding, ok := payload["binding"].(map[string]any)
	if !ok {
		return uuid.Nil, fmt.Errorf("network flow indicator link replay missing binding")
	}
	targetRef, ok := binding["target_indicator_ref"].(map[string]any)
	if !ok {
		return uuid.Nil, fmt.Errorf("network flow indicator link replay missing target indicator ref")
	}
	text, ok := targetRef["indicator_id"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("network flow indicator link replay target indicator id invalid")
	}
	indicatorID, err := uuid.Parse(text)
	if err != nil {
		return uuid.Nil, fmt.Errorf("network flow indicator link replay target indicator id parse: %w", err)
	}
	return indicatorID, nil
}

func networkFlowLinkableIPField(fieldKey string) bool {
	return fieldKey == FieldSrcIP || fieldKey == FieldDstIP
}

func canonicalIPLiteral(value string) bool {
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return false
	}
	return addr.String() == value
}

func indicatorTypeForIP(value string) string {
	addr, err := netip.ParseAddr(value)
	if err == nil && addr.Is4() {
		return "ipv4_addr"
	}
	return "ipv6_addr"
}

func invalidIndicatorSelector(field string, reason string) *httpapi.APIError {
	return indicatorLinkError(http.StatusBadRequest, "network_flow_invalid_indicator_selector", field, reason, "", "", "", "")
}

func invalidIndicatorTarget(field string, reason string) *httpapi.APIError {
	return indicatorLinkError(http.StatusBadRequest, "network_flow_invalid_indicator_target", field, reason, "", "", "", "")
}

func invalidIndicatorTargetWithContext(selectorKind string, fieldKey string, targetMode string, reason string) *httpapi.APIError {
	return indicatorLinkError(http.StatusBadRequest, "network_flow_invalid_indicator_target", "", reason, selectorKind, fieldKey, targetMode, "")
}

func indicatorLinkForbidden(selectorKind string, fieldKey string, targetMode string, reason string) *httpapi.APIError {
	return indicatorLinkError(http.StatusForbidden, "network_flow_indicator_link_forbidden", "", reason, selectorKind, fieldKey, targetMode, "")
}

func indicatorLinkAmbiguous(selectorKind string, fieldKey string, reason string, resolvedCandidate string) *httpapi.APIError {
	return indicatorLinkError(http.StatusBadRequest, "network_flow_indicator_link_ambiguous", "", reason, selectorKind, fieldKey, "", resolvedCandidate)
}

func indicatorLinkError(status int, code string, field string, reason string, selectorKind string, fieldKey string, targetMode string, resolvedCandidate string) *httpapi.APIError {
	details := map[string]any{"reason_code": reason}
	if field != "" {
		details["field"] = field
	}
	if selectorKind != "" {
		details["selector_kind"] = selectorKind
	}
	if fieldKey != "" {
		details["field_key"] = fieldKey
	}
	if targetMode != "" {
		details["target_mode"] = targetMode
	}
	if resolvedCandidate != "" {
		details["resolved_candidate_value"] = resolvedCandidate
	}
	return &httpapi.APIError{Status: status, Code: code, Message: code, Details: details}
}
