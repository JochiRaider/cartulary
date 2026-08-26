package tasksdecisions

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

func AdmitSupersedeJSON(reader io.Reader) (SupersedeRequest, *AdmissionFailure) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return SupersedeRequest{}, invalidAdmission("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"base_row_version": {}, "client_txn_id": {}, "reason": {}, "replacement_record_id": {},
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return SupersedeRequest{}, invalidAdmission(key, "unknown_field")
		}
	}
	var request SupersedeRequest
	if value, present := raw["base_row_version"]; !present {
		return SupersedeRequest{}, invalidAdmission("base_row_version", "missing_required_field")
	} else if json.Unmarshal(value, &request.BaseRowVersion) != nil || request.BaseRowVersion < 1 {
		return SupersedeRequest{}, invalidAdmission("base_row_version", "invalid_base_row_version")
	}
	if value, present := raw["client_txn_id"]; !present {
		return SupersedeRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return SupersedeRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	}
	if value, present := raw["reason"]; !present {
		return SupersedeRequest{}, invalidAdmission("reason", "missing_required_field")
	} else {
		var rawReason string
		if json.Unmarshal(value, &rawReason) != nil {
			return SupersedeRequest{}, invalidAdmission("reason", "invalid_value")
		}
		reason, ok := fieldnorm.NormalizeNote(rawReason)
		if !ok {
			return SupersedeRequest{}, invalidAdmission("reason", "invalid_value")
		}
		request.Reason = reason
	}
	if value, present := raw["replacement_record_id"]; present {
		if string(value) == "null" {
			return SupersedeRequest{}, invalidAdmission("replacement_record_id", "field_not_nullable")
		}
		var text string
		if json.Unmarshal(value, &text) != nil {
			return SupersedeRequest{}, invalidAdmission("replacement_record_id", "invalid_value")
		}
		parsed, parseErr := uuid.Parse(text)
		if parseErr != nil {
			return SupersedeRequest{}, invalidAdmission("replacement_record_id", "invalid_value")
		}
		request.ReplacementRecordID = &parsed
	}
	return request, nil
}
