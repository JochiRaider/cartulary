package networkflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/netip"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	schemaIndicatorLinkRequest   = "cartulary.network_flow.indicator_link_request.v1"
	schemaIndicatorLinkResult    = "cartulary.network_flow_indicator_link_result.v1"
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
	RowRefs          []networkFlowRowRef
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
	SourceRowRefs           []networkFlowRowRef
	SourceRowRefsTruncated  bool
	SourceRowRefsTotalCount int64
}

func (s *routeService) handleIndicatorLinks(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, pathErr := uuid.Parse(r.PathValue("incident_id"))
	if pathErr == nil {
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesEditorAdmin, "editor|admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
	}
	raw, apiErr := decodeNetworkFlowObjectHTTP(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, completeLinkError(apiErr, indicatorLinkRequest{}, ""))
		return
	}
	if pathErr != nil {
		writeAPIError(w, r, completeLinkError(invalidNetworkFlowRequestHTTP("incident_id", "type_mismatch"), indicatorLinkRequest{}, ""))
		return
	}
	if r.URL.RawQuery != "" {
		writeAPIError(w, r, completeLinkError(invalidNetworkFlowRequestHTTP("query", "unknown_member"), indicatorLinkRequest{}, ""))
		return
	}
	request, apiErr := decodeIndicatorLinkObject(raw, s.store.limits)
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

func (s *routeService) commitIndicatorLinkRoute(ctx context.Context, incidentID uuid.UUID, actor authn.UserRecord, request indicatorLinkRequest, requestHash []byte, requestID string) (payload map[string]any, status int, apiErr *httpapi.APIError) {
	candidate := ""
	defer func() { apiErr = completeLinkError(apiErr, request, candidate) }()
	idempotencyKey := indicatorLinkIdempotencyKey(actor.ID, incidentID, request.ClientTxnID)
	if payload, status, replayed, apiErr := s.replayIndicatorLinkIfPresent(ctx, idempotencyKey, requestHash, incidentID); replayed || apiErr != nil {
		return payload, status, apiErr
	}
	if _, err := s.incidentAccess.Check(ctx, incidentID, actor.ID, admission.Requirement{AllowedRoles: admission.RolesEditorAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
		return nil, 0, savedGraphAdmissionError(err, "editor|admin")
	}
	resolved, apiErr := s.resolveIndicatorSelector(ctx, incidentID, actor.ID, request.Selector)
	if apiErr != nil {
		return nil, 0, apiErr
	}
	candidate = resolved.CandidateValue
	if !canonicalIPLiteral(request.ConfirmExactValue) || request.ConfirmExactValue != resolved.CandidateValue {
		return nil, 0, indicatorLinkAmbiguous(resolved.SelectorKind, request.Selector.FieldKey, "candidate_mismatch", resolved.CandidateValue)
	}
	targetType, representable := indicators.CanonicalIPIndicatorType(resolved.CandidateValue)
	if !representable {
		return nil, 0, invalidIndicatorTarget("indicator_type", "core_ip_indicator_type_unavailable")
	}
	if request.Target.Mode == "create_indicator" && request.Target.IndicatorType != targetType {
		return nil, 0, invalidIndicatorTarget("indicator_type", "target_type_mismatch")
	}
	if request.Target.Mode == "existing_indicator" {
		target, err := s.store.GetActiveIndicator(ctx, incidentID, request.Target.IndicatorID)
		if errors.Is(err, errTableNotFound) {
			return nil, 0, indicatorLinkForbidden(request.Selector.Kind, request.Selector.FieldKey, request.Target.Mode, "target_not_visible")
		}
		if err != nil {
			return nil, 0, httpapi.InternalAPIError(err)
		}
		if err := validateIndicatorTargetLogical(target, candidate, targetType); err != nil {
			return nil, 0, invalidIndicatorTarget("target", indicatorParticipantReason(err))
		}
	}
	if int64(len(resolved.SourceRowRefs)) > s.store.limits.MaxBindingSourceRowRefs {
		return nil, 0, &httpapi.APIError{Status: http.StatusRequestEntityTooLarge, Code: "network_flow_resource_limit_exceeded", Details: map[string]any{"reason_code": "row_limit_exceeded", "limit_key": "max_binding_source_row_refs", "limit": s.store.limits.MaxBindingSourceRowRefs, "actual": len(resolved.SourceRowRefs), "phase": "indicator_link", "retry_action": "reduce_scope_or_limits"}}
	}
	if int64(len(request.Selector.GraphQuery.SelectedTableIDs)) > s.store.limits.MaxSelectedTablesPerQuery {
		return nil, 0, &httpapi.APIError{Status: 400, Code: "network_flow_invalid_table_scope", Details: map[string]any{"reason_code": "selected_table_limit_exceeded", "mode": "selected_tables", "table_ids": request.Selector.GraphQuery.SelectedTableIDs, "limit_key": "network_flow.max_selected_tables_per_query", "retry_action": "correct_request"}}
	}
	if int64(len(request.Selector.GraphQuery.Filters)) > s.store.limits.MaxFiltersPerQuery {
		return nil, 0, &httpapi.APIError{Status: 400, Code: "network_flow_invalid_filter", Details: map[string]any{"reason_code": "too_many_filters", "field_key": nil, "op": nil, "filter_index": nil, "retry_action": "correct_request"}}
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
		var currentError *indicatorLinkPreconditionError
		if errors.As(err, &currentError) {
			return nil, 0, currentError.APIError
		}
		var denied *admission.Denied
		if errors.As(err, &denied) {
			return nil, 0, savedGraphAdmissionError(err, "editor|admin")
		}
		if errors.Is(err, indicators.ErrIndicatorNotFound) {
			return nil, 0, indicatorLinkForbidden(request.Selector.Kind, request.Selector.FieldKey, request.Target.Mode, "target_not_visible")
		}
		var validation *indicators.IndicatorCreateValidationError
		if errors.As(err, &validation) {
			return nil, 0, invalidIndicatorTarget("target", "target_value_mismatch")
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
				Details: map[string]any{"reason_code": "extension_transaction_timeout", "operation_id": "network-flow-indicator-link:" + incidentID.String() + ":" + request.ClientTxnID, "timeout_seconds": s.transactions.Timeout().Seconds()},
			}
		}
		if errors.Is(err, errInvalidStorageArgument) {
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

func (s *routeService) replayIndicatorLinkIfPresent(ctx context.Context, key authn.RouteIdempotencyKey, requestHash []byte, incidentID uuid.UUID) (map[string]any, int, bool, *httpapi.APIError) {
	existing, err := s.authStore.GetRouteIdempotency(ctx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return nil, 0, true, httpapi.ClientTxnConflictError(key.ClientTxnID)
		}
		payload, err := decodeStoredNetworkFlowResponse(existing.ResponseJSON)
		if err != nil {
			return nil, 0, true, httpapi.InternalAPIError(err)
		}
		if err := validateIndicatorLinkReceipt(payload, existing.StatusCode, incidentID); err != nil {
			return nil, 0, true, httpapi.InternalAPIError(err)
		}
		indicatorID, err := indicatorIDFromLinkPayload(payload)
		if err != nil {
			return nil, 0, true, httpapi.InternalAPIError(err)
		}
		if _, err := s.store.GetActiveIndicator(ctx, incidentID, indicatorID); err != nil {
			if errors.Is(err, errTableNotFound) {
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

func (s *routeService) resolveIndicatorSelector(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, selector indicatorLinkSelector) (resolvedIndicatorLinkSelector, *httpapi.APIError) {
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
			SourceRowRefs:           []networkFlowRowRef{rowRefFromRow(row)},
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

func (s *routeService) resolveRowRefsSelector(ctx context.Context, incidentID uuid.UUID, selector indicatorLinkSelector) (resolvedIndicatorLinkSelector, *httpapi.APIError) {
	activeTables, err := s.store.ListActiveTables(ctx, incidentID)
	if err != nil {
		return resolvedIndicatorLinkSelector{}, httpapi.InternalAPIError(err)
	}
	tableRanks := make(map[string]int, len(activeTables))
	for index, table := range activeTables {
		tableRanks[table.TableID] = index
	}
	rows := make([]flowRow, 0, len(selector.RowRefs))
	candidate := ""
	for _, ref := range selector.RowRefs {
		row, apiErr := s.getAcceptedRowForLink(ctx, incidentID, ref.NetworkFlowTableID, ref.NetworkFlowRowID, &ref)
		if apiErr != nil {
			return resolvedIndicatorLinkSelector{}, apiErr
		}
		rows = append(rows, row)
	}
	for _, row := range rows {
		value, apiErr := candidateValueFromRow(row, selector.FieldKey)
		if apiErr != nil {
			return resolvedIndicatorLinkSelector{}, apiErr
		}
		if candidate == "" {
			candidate = value
		} else if candidate != value {
			return resolvedIndicatorLinkSelector{}, indicatorLinkAmbiguous(selector.Kind, selector.FieldKey, "candidate_mismatch", candidate)
		}
	}
	sortContributorRows(rows, tableRanks)
	refs := make([]networkFlowRowRef, 0, len(rows))
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

func (s *routeService) resolveGraphSelector(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, selector indicatorLinkSelector) (resolvedIndicatorLinkSelector, *httpapi.APIError) {
	_ = actorUserID
	for _, tableID := range selector.GraphQuery.SelectedTableIDs {
		if apiErr := s.ensureActiveTables(ctx, incidentID, []string{tableID}); apiErr != nil {
			return resolvedIndicatorLinkSelector{}, linkSourceTableError(apiErr, tableID)
		}
	}
	composition, apiErr := s.composeGraphSourceFromSemanticHTTP(ctx, incidentID, selector.GraphQuery)
	if apiErr != nil {
		return resolvedIndicatorLinkSelector{}, apiErr
	}
	if composition.Digest != selector.GraphQueryDigest {
		return resolvedIndicatorLinkSelector{}, graphQueryStaleHTTP("digest_mismatch", selector.GraphQueryDigest)
	}
	var candidate string
	var predicate graphContributorPredicate
	switch selector.Kind {
	case "graph_vertex":
		vertex := composition.Vertices[selector.VertexID]
		if vertex == nil {
			return resolvedIndicatorLinkSelector{}, graphQueryStaleHTTP("vertex_not_found", selector.GraphQueryDigest)
		}
		candidate = vertex.EndpointValue
		predicate = graphContributorPredicate{Kind: "vertex", EndpointValue: vertex.EndpointValue}
	case "graph_edge":
		edge := composition.Edges[selector.EdgeID]
		if edge == nil {
			return resolvedIndicatorLinkSelector{}, graphQueryStaleHTTP("edge_not_found", selector.GraphQueryDigest)
		}
		switch selector.FieldKey {
		case fieldSrcIP:
			candidate = edge.SrcEndpointValue
		case fieldDstIP:
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
	refs := make([]networkFlowRowRef, 0, limit)
	var total int64
	err := s.store.IterateGraphContributorRows(ctx, incidentID, composition.SelectedTableIDs, predicate, func(row flowRow) error {
		matched, matchErr := rowMatchesGraphQueryHTTP(row, selector.GraphQuery.Filters, selector.GraphQuery.TimeRange, selector.GraphQuery.Aggregation)
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
			return resolvedIndicatorLinkSelector{}, graphProjectionFailedForContextHTTP(err)
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

func (s *routeService) getAcceptedRowForLink(ctx context.Context, incidentID uuid.UUID, tableID string, rowID string, suppliedRef *networkFlowRowRef) (flowRow, *httpapi.APIError) {
	if apiErr := s.ensureActiveTables(ctx, incidentID, []string{tableID}); apiErr != nil {
		return flowRow{}, linkSourceTableError(apiErr, tableID)
	}
	rows, err := s.store.ListRows(ctx, incidentID, tableID)
	if err != nil {
		return flowRow{}, tableReadError(err)
	}
	for _, row := range rows {
		if row.RowID != rowID {
			continue
		}
		if suppliedRef != nil && (row.SourceRowNumber != suppliedRef.SourceRowNumber || row.MappingFingerprint != suppliedRef.MappingFingerprint) {
			return flowRow{}, invalidIndicatorSelector("row_refs", "row_not_accepted")
		}
		return row, nil
	}
	return flowRow{}, invalidIndicatorSelector("network_flow_row_id", "row_not_accepted")
}

func candidateValueFromRow(row flowRow, fieldKey string) (string, *httpapi.APIError) {
	switch fieldKey {
	case fieldSrcIP:
		return row.SrcIP, nil
	case fieldDstIP:
		return row.DstIP, nil
	default:
		return "", invalidIndicatorSelector("field_key", "field_not_linkable")
	}
}

func rowRefFromRow(row flowRow) networkFlowRowRef {
	return networkFlowRowRef{
		NetworkFlowTableID: row.NetworkFlowTableID,
		NetworkFlowRowID:   row.RowID,
		SourceRowNumber:    row.SourceRowNumber,
		MappingFingerprint: row.MappingFingerprint,
	}
}

func indicatorLinkPayload(binding indicatorBindingRecord, duplicate bool) map[string]any {
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
	digest := sha256.Sum256(graphViewMutationBytes(routeKeyIndicatorLinksCreate, "indicator-links", map[string]any{
		"selector":            indicatorSelectorHashResource(request.Selector),
		"target":              indicatorTargetHashResource(request.Target),
		"observation_mode":    "binding_only",
		"confirm_exact_value": request.ConfirmExactValue,
	}))
	return digest[:]
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
	return fieldKey == fieldSrcIP || fieldKey == fieldDstIP
}

func canonicalIPLiteral(value string) bool {
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return false
	}
	return addr.String() == value
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
