package mentions

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const mentionActionRouteKey = "entities.entity_mentions.resolve"

type MentionActionRequest struct {
	BaseMentionRowVersion int64
	ClientTxnID           string
	Action                string
	ResolvedRecordID      *uuid.UUID
	Reason                *string
}

type MentionActionResult struct {
	Payload                map[string]any
	StatusCode             int
	Replayed               bool
	IncidentID             uuid.UUID
	SourceRecordID         uuid.UUID
	SourceRecordRowVersion int64
	ChangeSetID            uuid.UUID
	ClientTxnID            string
	ChangedFieldKeys       []string
	EntityInvalidations    []MentionEntityInvalidation
}

type MentionEntityInvalidation struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ViewSchemaID     string
	ChangedFieldKeys []string
}

func DecodeMentionActionRequest(reader io.Reader) (MentionActionRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return MentionActionRequest{}, apiErr
	}

	allowed := map[string]struct{}{
		"base_mention_row_version": {},
		"client_txn_id":            {},
		"action":                   {},
		"resolved_record_id":       {},
		"reason":                   {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return MentionActionRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request MentionActionRequest
	if value, ok := raw["base_mention_row_version"]; !ok {
		return MentionActionRequest{}, invalidMutationPayload("base_mention_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseMentionRowVersion); err != nil || request.BaseMentionRowVersion < 1 {
		return MentionActionRequest{}, invalidMutationPayload("base_mention_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return MentionActionRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return MentionActionRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["action"]; !ok {
		return MentionActionRequest{}, invalidMutationPayload("action", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Action); err != nil {
		return MentionActionRequest{}, invalidMutationPayload("action", "invalid_value")
	}

	switch request.Action {
	case "resolve_item":
		value, ok := raw["resolved_record_id"]
		if !ok || string(value) == "null" {
			return MentionActionRequest{}, invalidMutationPayload("resolved_record_id", "missing_required_field")
		}
		var rawID string
		if err := json.Unmarshal(value, &rawID); err != nil {
			return MentionActionRequest{}, invalidMutationPayload("resolved_record_id", "invalid_value")
		}
		recordID, err := uuid.Parse(rawID)
		if err != nil {
			return MentionActionRequest{}, invalidMutationPayload("resolved_record_id", "invalid_value")
		}
		request.ResolvedRecordID = &recordID
	case "dismiss_item", "revert_to_unresolved":
		if _, ok := raw["resolved_record_id"]; ok {
			return MentionActionRequest{}, invalidMutationPayload("resolved_record_id", "field_forbidden")
		}
	default:
		return MentionActionRequest{}, invalidMutationPayload("action", "invalid_value")
	}

	if value, ok := raw["reason"]; ok {
		if string(value) == "null" {
			request.Reason = nil
		} else {
			var rawReason string
			if err := json.Unmarshal(value, &rawReason); err != nil {
				return MentionActionRequest{}, invalidMutationPayload("reason", "invalid_value")
			}
			request.Reason = authn.NormalizeReasonNote(&rawReason)
		}
	}
	return request, nil
}

func MentionActionRequestHash(request MentionActionRequest) []byte {
	payload := map[string]any{
		"base_mention_row_version": request.BaseMentionRowVersion,
		"client_txn_id":            request.ClientTxnID,
		"action":                   request.Action,
		"resolved_record_id":       nil,
		"reason":                   derefString(request.Reason),
	}
	if request.ResolvedRecordID != nil {
		payload["resolved_record_id"] = request.ResolvedRecordID.String()
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}
