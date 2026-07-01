package revisions

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const rollbackRouteKey = "records.rollback"

type RollbackTarget struct {
	Kind                string
	HistoryEntryRef     string
	ChangeSetID         string
	RestoreToRevisionNo int64
}

type RollbackRequest struct {
	BaseRowVersion int64
	ClientTxnID    string
	Reason         *string
	Target         RollbackTarget
}

func DecodeRollbackRequest(reader io.Reader) (RollbackRequest, *httpapi.APIError) {
	raw, apiErr := decodeObjectRollback(reader)
	if apiErr != nil {
		return RollbackRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"base_row_version": {},
		"client_txn_id":    {},
		"reason":           {},
		"target":           {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return RollbackRequest{}, invalidRollbackRequest(key, "unknown_field")
		}
	}

	var request RollbackRequest
	if value, ok := raw["base_row_version"]; !ok {
		return RollbackRequest{}, invalidRollbackRequest("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return RollbackRequest{}, invalidRollbackRequest("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return RollbackRequest{}, invalidRollbackRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return RollbackRequest{}, invalidRollbackRequest("client_txn_id", "missing_required_field")
	}
	reason, ok := normalizeRollbackReason(raw, "reason")
	if !ok {
		return RollbackRequest{}, invalidRollbackRequest("reason", "invalid_value")
	}
	request.Reason = reason

	targetRaw, ok := raw["target"]
	if !ok {
		return RollbackRequest{}, invalidRollbackRequest("target", "missing_required_field")
	}
	target, apiErr := decodeRollbackTarget(targetRaw)
	if apiErr != nil {
		return RollbackRequest{}, apiErr
	}
	request.Target = target
	return request, nil
}

func decodeObjectRollback(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
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

func decodeRollbackTarget(value json.RawMessage) (RollbackTarget, *httpapi.APIError) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(value, &raw); err != nil || raw == nil {
		return RollbackTarget{}, invalidRollbackRequest("target", "target_not_object")
	}
	kindValue, ok := raw["kind"]
	if !ok {
		return RollbackTarget{}, invalidRollbackRequest("target.kind", "missing_required_field")
	}
	var kind string
	if err := json.Unmarshal(kindValue, &kind); err != nil || kind == "" {
		return RollbackTarget{}, invalidRollbackRequest("target.kind", "invalid_value")
	}
	target := RollbackTarget{Kind: kind}
	var allowed map[string]struct{}
	switch kind {
	case "history_entry":
		allowed = map[string]struct{}{"kind": {}, "history_entry_ref": {}}
		value, ok := raw["history_entry_ref"]
		if !ok {
			return RollbackTarget{}, invalidRollbackRequest("target.history_entry_ref", "missing_required_field")
		}
		if err := json.Unmarshal(value, &target.HistoryEntryRef); err != nil || strings.TrimSpace(target.HistoryEntryRef) == "" {
			return RollbackTarget{}, invalidRollbackRequest("target.history_entry_ref", "invalid_value")
		}
	case "change_set":
		allowed = map[string]struct{}{"kind": {}, "change_set_id": {}}
		value, ok := raw["change_set_id"]
		if !ok {
			return RollbackTarget{}, invalidRollbackRequest("target.change_set_id", "missing_required_field")
		}
		if err := json.Unmarshal(value, &target.ChangeSetID); err != nil || strings.TrimSpace(target.ChangeSetID) == "" {
			return RollbackTarget{}, invalidRollbackRequest("target.change_set_id", "invalid_value")
		}
		parsed, err := uuid.Parse(strings.TrimSpace(target.ChangeSetID))
		if err != nil {
			return RollbackTarget{}, invalidRollbackRequest("target.change_set_id", "invalid_value")
		}
		target.ChangeSetID = parsed.String()
	case "row_restore":
		allowed = map[string]struct{}{"kind": {}, "restore_to_revision_no": {}}
		value, ok := raw["restore_to_revision_no"]
		if !ok {
			return RollbackTarget{}, invalidRollbackRequest("target.restore_to_revision_no", "missing_required_field")
		}
		if err := json.Unmarshal(value, &target.RestoreToRevisionNo); err != nil || target.RestoreToRevisionNo < 1 {
			return RollbackTarget{}, invalidRollbackRequest("target.restore_to_revision_no", "invalid_value")
		}
	default:
		return RollbackTarget{}, invalidRollbackRequest("target.kind", "unsupported_target_kind")
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return RollbackTarget{}, invalidRollbackRequest("target."+key, "unknown_field")
		}
	}
	return target, nil
}

func RollbackRequestHash(request RollbackRequest) []byte {
	payload := map[string]any{
		"base_row_version": request.BaseRowVersion,
		"reason":           stringOrNil(request.Reason),
		"target":           request.Target.Normalized(),
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func (target RollbackTarget) Normalized() map[string]any {
	switch target.Kind {
	case "history_entry":
		return map[string]any{"kind": "history_entry", "history_entry_ref": target.HistoryEntryRef}
	case "change_set":
		return map[string]any{"kind": "change_set", "change_set_id": target.ChangeSetID}
	case "row_restore":
		return map[string]any{"kind": "row_restore", "restore_to_revision_no": target.RestoreToRevisionNo}
	default:
		return map[string]any{"kind": target.Kind}
	}
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

func invalidRollbackRequest(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_rollback_request",
		Message: "invalid rollback request",
		Details: details,
	}
}
