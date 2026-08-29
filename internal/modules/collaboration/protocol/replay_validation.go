package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ValidateReplayablePayload validates one admitted durable family without
// requiring sequence metadata. Unknown families fail closed.
func ValidateReplayablePayload(incidentID uuid.UUID, family string, payload json.RawMessage) error {
	if incidentID == uuid.Nil {
		return errors.New("replayable payload incident identity is invalid")
	}
	var object map[string]json.RawMessage
	if len(payload) == 0 || json.Unmarshal(payload, &object) != nil || object == nil {
		return errors.New("replayable payload must be a JSON object")
	}
	switch family {
	case "record_changed":
		if err := validateRecordChangedPayload(payload); err != nil {
			return fmt.Errorf("invalid record_changed payload: %w", err)
		}
	case "job_progress":
		var progress JobProgressPayload
		if err := json.Unmarshal(payload, &progress); err != nil {
			return fmt.Errorf("invalid job_progress payload: %w", err)
		}
		if err := ValidateIncidentJobProgressPayload(incidentID, progress); err != nil {
			return fmt.Errorf("invalid job_progress payload: %w", err)
		}
	case "extension_resource_changed":
		var change ExtensionResourceChangePayload
		if err := json.Unmarshal(payload, &change); err != nil {
			return fmt.Errorf("invalid extension_resource_changed payload: %w", err)
		}
		if err := ValidateExtensionResourceChangePayload(change); err != nil {
			return fmt.Errorf("invalid extension_resource_changed payload: %w", err)
		}
	default:
		return fmt.Errorf("unsupported replayable message family %q", family)
	}
	return nil
}

// ValidateSequencedReplayableMessage validates the durable envelope and its
// family payload while retaining the protocol's additive-member behavior.
func ValidateSequencedReplayableMessage(message Message) error {
	if !IsReplayableMessageType(message.Type) {
		return fmt.Errorf("unsupported replayable message family %q", message.Type)
	}
	incidentID, err := uuid.Parse(message.IncidentID)
	if err != nil || incidentID == uuid.Nil {
		return errors.New("sequenced replayable incident identity is invalid")
	}
	eventID := strings.TrimSpace(message.EventID)
	if eventID == "" || eventID != message.EventID || len(eventID) > 512 {
		return errors.New("sequenced replayable event identity is invalid")
	}
	if message.StreamSeq == nil || *message.StreamSeq < 1 {
		return errors.New("sequenced replayable stream sequence is invalid")
	}
	emittedAt, err := time.Parse(time.RFC3339Nano, message.EmittedAt)
	if err != nil || emittedAt.IsZero() {
		return errors.New("sequenced replayable emitted_at is invalid")
	}
	_, offsetSeconds := emittedAt.Zone()
	if offsetSeconds != 0 {
		return errors.New("sequenced replayable emitted_at must be UTC")
	}
	return ValidateReplayablePayload(incidentID, message.Type, message.Payload)
}
