package scenariotest

import (
	"github.com/google/uuid"
)

const (
	TimelineHostRefsFieldKey = "timeline.host_refs"
	MentionActionResolve     = "resolve_item"
	MentionActionDismiss     = "dismiss_item"
)

func HostCreatePayload(clientTxnID string) map[string]any {
	return map[string]any{
		"client_txn_id":     clientTxnID,
		"host.display_name": "VPN Gateway",
		"host.hostname":     "vpn-gateway",
	}
}

func IdentityCreatePayload(clientTxnID string) map[string]any {
	return map[string]any{
		"client_txn_id":             clientTxnID,
		"identity.display_name":     "VPN User",
		"identity.email":            "vpn.user@example.test",
		"identity.sam_account_name": "VPNUSER",
	}
}

func IndicatorCreatePayload(clientTxnID string) map[string]any {
	return map[string]any{
		"client_txn_id":              clientTxnID,
		"indicator.indicator_type":   "ipv4_addr",
		"indicator.value_kind":       "atomic",
		"indicator.display_value":    "203.0.113.24",
		"indicator.normalized_value": "203.0.113.24",
		"indicator.defanged_value":   "203[.]0[.]113[.]24",
	}
}

func CollectionActions(actions ...map[string]any) map[string]any {
	return map[string]any{
		"kind":    "collection_actions_v1",
		"actions": append([]map[string]any(nil), actions...),
	}
}

func AddResolvedRefAction(rawText string, resolvedRecordID uuid.UUID) map[string]any {
	return map[string]any{
		"op":                 "add_resolved_ref",
		"raw_text":           rawText,
		"resolved_record_id": resolvedRecordID.String(),
	}
}

func TimelineCollectionPatchPayload(
	fieldKey string,
	baseRowVersion int64,
	clientTxnID string,
	actionPayload map[string]any,
) map[string]any {
	return map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v2",
		"base_row_version": baseRowVersion,
		"client_txn_id":    clientTxnID,
		"changes": []map[string]any{{
			"field_key":      fieldKey,
			"action_payload": actionPayload,
		}},
	}
}

func MentionResolveRoutePayload(
	baseMentionRowVersion int64,
	clientTxnID string,
	action string,
	resolvedRecordID *uuid.UUID,
	reason *string,
) map[string]any {
	payload := map[string]any{
		"base_mention_row_version": baseMentionRowVersion,
		"client_txn_id":            clientTxnID,
		"action":                   action,
	}
	if resolvedRecordID != nil {
		payload["resolved_record_id"] = resolvedRecordID.String()
	}
	if reason != nil {
		payload["reason"] = *reason
	}
	return payload
}
