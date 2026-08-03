package httpapi

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func decodeRollbackRequest(reader io.Reader) (revisions.RollbackRequest, *platformhttpapi.APIError) {
	raw, apiErr := decodeObjectRollback(reader)
	if apiErr != nil {
		return revisions.RollbackRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"base_row_version": {},
		"client_txn_id":    {},
		"reason":           {},
		"target":           {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return revisions.RollbackRequest{}, invalidRollbackRequest(key, "unknown_field")
		}
	}

	var request revisions.RollbackRequest
	if value, ok := raw["base_row_version"]; !ok {
		return revisions.RollbackRequest{}, invalidRollbackRequest("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return revisions.RollbackRequest{}, invalidRollbackRequest("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return revisions.RollbackRequest{}, invalidRollbackRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return revisions.RollbackRequest{}, invalidRollbackRequest("client_txn_id", "missing_required_field")
	}
	reason, ok := normalizeRollbackReason(raw, "reason")
	if !ok {
		return revisions.RollbackRequest{}, invalidRollbackRequest("reason", "invalid_value")
	}
	request.Reason = reason

	targetRaw, ok := raw["target"]
	if !ok {
		return revisions.RollbackRequest{}, invalidRollbackRequest("target", "missing_required_field")
	}
	target, apiErr := decodeRollbackTarget(targetRaw)
	if apiErr != nil {
		return revisions.RollbackRequest{}, apiErr
	}
	request.Target = target
	return request, nil
}

func decodeObjectRollback(reader io.Reader) (map[string]json.RawMessage, *platformhttpapi.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil || raw == nil {
		return nil, invalidRollbackRequest("", "request_not_object")
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return nil, invalidRollbackRequest("", "request_not_object")
	}
	return raw, nil
}

func decodeRollbackTarget(value json.RawMessage) (revisions.RollbackTarget, *platformhttpapi.APIError) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(value, &raw); err != nil || raw == nil {
		return revisions.RollbackTarget{}, invalidRollbackRequest("target", "target_not_object")
	}
	kindValue, ok := raw["kind"]
	if !ok {
		return revisions.RollbackTarget{}, invalidRollbackRequest("target.kind", "missing_required_field")
	}
	var kind string
	if err := json.Unmarshal(kindValue, &kind); err != nil || kind == "" {
		return revisions.RollbackTarget{}, invalidRollbackRequest("target.kind", "invalid_value")
	}
	target := revisions.RollbackTarget{Kind: kind}
	var allowed map[string]struct{}
	switch kind {
	case "history_entry":
		allowed = map[string]struct{}{"kind": {}, "history_entry_ref": {}}
		value, ok := raw["history_entry_ref"]
		if !ok {
			return revisions.RollbackTarget{}, invalidRollbackRequest("target.history_entry_ref", "missing_required_field")
		}
		if err := json.Unmarshal(value, &target.HistoryEntryRef); err != nil || strings.TrimSpace(target.HistoryEntryRef) == "" {
			return revisions.RollbackTarget{}, invalidRollbackRequest("target.history_entry_ref", "invalid_value")
		}
	case "change_set":
		allowed = map[string]struct{}{"kind": {}, "change_set_id": {}}
		value, ok := raw["change_set_id"]
		if !ok {
			return revisions.RollbackTarget{}, invalidRollbackRequest("target.change_set_id", "missing_required_field")
		}
		if err := json.Unmarshal(value, &target.ChangeSetID); err != nil || strings.TrimSpace(target.ChangeSetID) == "" {
			return revisions.RollbackTarget{}, invalidRollbackRequest("target.change_set_id", "invalid_value")
		}
		parsed, err := uuid.Parse(strings.TrimSpace(target.ChangeSetID))
		if err != nil {
			return revisions.RollbackTarget{}, invalidRollbackRequest("target.change_set_id", "invalid_value")
		}
		target.ChangeSetID = parsed.String()
	case "row_restore":
		allowed = map[string]struct{}{"kind": {}, "restore_to_revision_no": {}}
		value, ok := raw["restore_to_revision_no"]
		if !ok {
			return revisions.RollbackTarget{}, invalidRollbackRequest("target.restore_to_revision_no", "missing_required_field")
		}
		if err := json.Unmarshal(value, &target.RestoreToRevisionNo); err != nil || target.RestoreToRevisionNo < 1 {
			return revisions.RollbackTarget{}, invalidRollbackRequest("target.restore_to_revision_no", "invalid_value")
		}
	default:
		return revisions.RollbackTarget{}, invalidRollbackRequest("target.kind", "unsupported_target_kind")
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return revisions.RollbackTarget{}, invalidRollbackRequest("target."+key, "unknown_field")
		}
	}
	return target, nil
}

func normalizeRollbackReason(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok || string(value) == "null" {
		return nil, true
	}
	var input string
	if err := json.Unmarshal(value, &input); err != nil {
		return nil, false
	}
	normalized, ok := fieldnorm.NormalizeNote(input)
	if !ok {
		return nil, true
	}
	if len([]rune(normalized)) > 4096 {
		return nil, false
	}
	return &normalized, true
}

func invalidRollbackRequest(field string, reasonCode string) *platformhttpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &platformhttpapi.APIError{
		Status:  400,
		Code:    "invalid_rollback_request",
		Message: "invalid rollback request",
		Details: details,
	}
}
