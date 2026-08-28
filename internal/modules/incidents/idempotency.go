package incidents

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/google/uuid"
)

type incidentCreateIdempotencyPayload struct {
	ClientTxnID            string  `json:"client_txn_id"`
	CurrentPhase           *string `json:"current_phase"`
	Description            *string `json:"description"`
	IncidentKey            string  `json:"incident_key"`
	PrimaryExternalCaseRef *string `json:"primary_external_case_ref"`
	Severity               *string `json:"severity"`
	Title                  string  `json:"title"`
	TLP                    *string `json:"tlp"`
}

type incidentLifecycleIdempotencyPayload struct {
	ActionRoute         string `json:"action_route"`
	BaseIncidentVersion int64  `json:"base_incident_version"`
	Reason              string `json:"reason"`
}

type membershipCreateIdempotencyPayload struct {
	ClientTxnID string     `json:"client_txn_id"`
	Email       *string    `json:"email"`
	Role        string     `json:"role"`
	UserID      *uuid.UUID `json:"user_id"`
}

func hashRequestPayload(payload any) [sha256.Size]byte {
	data, _ := json.Marshal(payload)
	return sha256.Sum256(data)
}

func hashesEqual(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
