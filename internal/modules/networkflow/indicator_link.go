package networkflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
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

type indicatorLinkApplication struct {
	store          *store
	incidentAccess incidentAdmissionChecker
	receipts       indicatorLinkReceiptAdapter
	safeDigester   safeDigester
	now            func() time.Time
	graphComposer  *graphSourceComposer
	transactions   *crossownertransaction.Coordinator
}
type indicatorLinkOutcome struct {
	binding   indicatorBindingRecord
	duplicate bool
	receipt   *storedMutationReceipt
}

func (s *indicatorLinkApplication) execute(ctx context.Context, incidentID uuid.UUID, actor authn.UserRecord, request indicatorLinkRequest, requestHash []byte, requestID string) (outcome indicatorLinkOutcome, apiErr *semanticFailure) {
	candidate := ""
	defer func() { apiErr = completeLinkError(apiErr, request, candidate) }()
	if _, err := s.incidentAccess.Check(ctx, incidentID, actor.ID, admission.Requirement{AllowedRoles: admission.RolesEditorAdmin, Lifecycle: admission.LifecycleOpen}); err != nil {
		return indicatorLinkOutcome{}, applicationAdmissionFailure(err, "editor|admin")
	}
	idempotencyKey := indicatorLinkIdempotencyKey(actor.ID, incidentID, request.ClientTxnID)
	if outcome, replayed, apiErr := s.receipts.replay(ctx, idempotencyKey, requestHash, incidentID); replayed || apiErr != nil {
		return outcome, apiErr
	}
	resolved, apiErr := s.resolveIndicatorSelector(ctx, incidentID, actor.ID, request.Selector)
	if apiErr != nil {
		return indicatorLinkOutcome{}, apiErr
	}
	candidate = resolved.CandidateValue
	if !canonicalIPLiteral(request.ConfirmExactValue) || request.ConfirmExactValue != resolved.CandidateValue {
		return indicatorLinkOutcome{}, indicatorLinkAmbiguous(resolved.SelectorKind, request.Selector.FieldKey, "candidate_mismatch", resolved.CandidateValue)
	}
	targetType, representable := indicators.CanonicalIPIndicatorType(resolved.CandidateValue)
	if !representable {
		return indicatorLinkOutcome{}, invalidIndicatorTarget("indicator_type", "core_ip_indicator_type_unavailable")
	}
	if request.Target.Mode == "create_indicator" && request.Target.IndicatorType != targetType {
		return indicatorLinkOutcome{}, invalidIndicatorTarget("indicator_type", "target_type_mismatch")
	}
	if request.Target.Mode == "existing_indicator" {
		target, err := s.store.GetActiveIndicator(ctx, incidentID, request.Target.IndicatorID)
		if errors.Is(err, errTableNotFound) {
			return indicatorLinkOutcome{}, indicatorLinkForbidden(request.Selector.Kind, request.Selector.FieldKey, request.Target.Mode, "target_not_visible")
		}
		if err != nil {
			return indicatorLinkOutcome{}, internalSemanticFailure(err)
		}
		if err := validateIndicatorTargetLogical(target, candidate, targetType); err != nil {
			return indicatorLinkOutcome{}, invalidIndicatorTarget("target", indicatorParticipantReason(err))
		}
	}
	if int64(len(resolved.SourceRowRefs)) > s.store.limits.MaxBindingSourceRowRefs {
		return indicatorLinkOutcome{}, &semanticFailure{kind: failureResourceLimit, reason: "row_limit_exceeded", details: failureDetails{LimitKey: "max_binding_source_row_refs", Limit: int(s.store.limits.MaxBindingSourceRowRefs), Actual: len(resolved.SourceRowRefs), Phase: "indicator_link"}}
	}
	if int64(len(request.Selector.GraphQuery.SelectedTableIDs)) > s.store.limits.MaxSelectedTablesPerQuery {
		return indicatorLinkOutcome{}, &semanticFailure{kind: failureInvalidTableScope, reason: "selected_table_limit_exceeded", details: failureDetails{TableIDs: request.Selector.GraphQuery.SelectedTableIDs, LimitKey: "network_flow.max_selected_tables_per_query", LinkGraph: true}}
	}
	if int64(len(request.Selector.GraphQuery.Filters)) > s.store.limits.MaxFiltersPerQuery {
		return indicatorLinkOutcome{}, invalidFilter("", "too_many_filters")
	}
	if s.transactions == nil {
		return indicatorLinkOutcome{}, internalSemanticFailure(crossownertransaction.ErrUnavailable)
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
		if outcome, replayed, replayErr := s.receipts.replay(ctx, idempotencyKey, requestHash, incidentID); replayed || replayErr != nil {
			return outcome, replayErr
		}
		if authn.IsUniqueViolation(err) {
			return indicatorLinkOutcome{}, clientTxnFailure(request.ClientTxnID)
		}
		var currentError *semanticFailure
		if errors.As(err, &currentError) {
			return indicatorLinkOutcome{}, currentError
		}
		var denied *admission.Denied
		if errors.As(err, &denied) {
			return indicatorLinkOutcome{}, applicationAdmissionFailure(err, "editor|admin")
		}
		if errors.Is(err, indicators.ErrIndicatorNotFound) {
			return indicatorLinkOutcome{}, indicatorLinkForbidden(request.Selector.Kind, request.Selector.FieldKey, request.Target.Mode, "target_not_visible")
		}
		var validation *indicators.IndicatorCreateValidationError
		if errors.As(err, &validation) {
			return indicatorLinkOutcome{}, invalidIndicatorTarget("target", "target_value_mismatch")
		}
		if reason := indicatorParticipantReason(err); reason != "" {
			return indicatorLinkOutcome{}, invalidIndicatorTargetWithContext(request.Selector.Kind, request.Selector.FieldKey, request.Target.Mode, reason)
		}
		var conflict *crossownertransaction.ConflictError
		if errors.As(err, &conflict) {
			return indicatorLinkOutcome{}, &semanticFailure{kind: failureTransactionConflict, reason: conflict.ReasonCode}
		}
		if errors.Is(err, crossownertransaction.ErrTimeout) {
			return indicatorLinkOutcome{}, &semanticFailure{kind: failureTransactionTimeout, reason: "extension_transaction_timeout", details: failureDetails{OperationID: "network-flow-indicator-link:" + incidentID.String() + ":" + request.ClientTxnID, TimeoutSeconds: s.transactions.Timeout().Seconds()}}
		}
		if errors.Is(err, errInvalidStorageArgument) {
			return indicatorLinkOutcome{}, invalidIndicatorTarget("target", "target_value_mismatch")
		}
		return indicatorLinkOutcome{}, internalSemanticFailure(err)
	}
	committed, resultErr := participantResult[indicatorLinkCommitResult](result, IndicatorLinkParticipantID)
	if resultErr != nil {
		return indicatorLinkOutcome{}, internalSemanticFailure(resultErr)
	}
	return indicatorLinkOutcome{binding: committed.Binding, duplicate: committed.Duplicate}, nil
}

func (s *indicatorLinkApplication) resolveIndicatorSelector(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, selector indicatorLinkSelector) (resolvedIndicatorLinkSelector, *semanticFailure) {
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

func (s *indicatorLinkApplication) resolveRowRefsSelector(ctx context.Context, incidentID uuid.UUID, selector indicatorLinkSelector) (resolvedIndicatorLinkSelector, *semanticFailure) {
	activeTables, err := s.store.ListActiveTables(ctx, incidentID)
	if err != nil {
		return resolvedIndicatorLinkSelector{}, internalSemanticFailure(err)
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

func (s *indicatorLinkApplication) resolveGraphSelector(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, selector indicatorLinkSelector) (resolvedIndicatorLinkSelector, *semanticFailure) {
	_ = actorUserID
	for _, tableID := range selector.GraphQuery.SelectedTableIDs {
		if apiErr := s.store.ensureActiveTables(ctx, incidentID, []string{tableID}); apiErr != nil {
			return resolvedIndicatorLinkSelector{}, linkSourceTableError(apiErr, tableID)
		}
	}
	composition, apiErr := s.graphComposer.composeGraphSourceFromSemantic(ctx, incidentID, selector.GraphQuery)
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
		return resolvedIndicatorLinkSelector{}, internalSemanticFailure(err)
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

func (s *indicatorLinkApplication) getAcceptedRowForLink(ctx context.Context, incidentID uuid.UUID, tableID string, rowID string, suppliedRef *networkFlowRowRef) (flowRow, *semanticFailure) {
	if apiErr := s.store.ensureActiveTables(ctx, incidentID, []string{tableID}); apiErr != nil {
		return flowRow{}, linkSourceTableError(apiErr, tableID)
	}
	rows, err := s.store.ListRows(ctx, incidentID, tableID)
	if err != nil {
		return flowRow{}, tableReadFailure(err)
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

func candidateValueFromRow(row flowRow, fieldKey string) (string, *semanticFailure) {
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

func canonicalIPLiteral(value string) bool {
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return false
	}
	return addr.String() == value
}

func invalidIndicatorSelector(field, reason string) *semanticFailure {
	return newSemanticFailure(failureInvalidIndicatorSelector, field, reason)
}
func invalidIndicatorTarget(field, reason string) *semanticFailure {
	return newSemanticFailure(failureInvalidIndicatorTarget, field, reason)
}
func invalidIndicatorTargetWithContext(selector, field, target, reason string) *semanticFailure {
	return &semanticFailure{kind: failureInvalidIndicatorTarget, reason: reason, details: failureDetails{SelectorKind: selector, LinkFieldKey: field, TargetMode: target}}
}
func indicatorLinkForbidden(selector, field, target, reason string) *semanticFailure {
	return &semanticFailure{kind: failureIndicatorLinkForbidden, reason: reason, details: failureDetails{SelectorKind: selector, LinkFieldKey: field, TargetMode: target}}
}
func indicatorLinkAmbiguous(selector, field, reason, candidate string) *semanticFailure {
	return &semanticFailure{kind: failureIndicatorLinkAmbiguous, reason: reason, details: failureDetails{SelectorKind: selector, LinkFieldKey: field, Candidate: candidate}}
}
