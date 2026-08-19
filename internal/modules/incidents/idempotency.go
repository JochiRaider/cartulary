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

func incidentCreateRequestHash(request CreateIncidentRequest) []byte {
	return hashRequestPayload(incidentCreateIdempotencyPayload{
		ClientTxnID:            request.ClientTxnID,
		CurrentPhase:           request.CurrentPhase,
		Description:            request.Description,
		IncidentKey:            request.IncidentKey,
		PrimaryExternalCaseRef: request.PrimaryExternalCaseRef,
		Severity:               request.Severity,
		Title:                  request.Title,
		TLP:                    request.TLP,
	})
}

func incidentLifecycleRequestHash(action string, request IncidentLifecycleRequest) []byte {
	return hashRequestPayload(incidentLifecycleIdempotencyPayload{
		ActionRoute:         action,
		BaseIncidentVersion: request.BaseIncidentVersion,
		Reason:              request.Reason,
	})
}

func membershipCreateRequestHash(request MembershipCreateRequest) []byte {
	return hashRequestPayload(membershipCreateIdempotencyPayload{
		ClientTxnID: request.ClientTxnID,
		Email:       request.Email,
		Role:        request.Role,
		UserID:      request.UserID,
	})
}

func hashRequestPayload(payload any) []byte {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
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
