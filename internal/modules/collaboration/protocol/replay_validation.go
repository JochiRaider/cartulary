package protocol

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ValidateReplayablePayload validates one admitted durable family without
// requiring sequence metadata. Unknown families fail closed.
func ValidateReplayablePayload(incidentID uuid.UUID, family string, payload json.RawMessage) error {
	if incidentID == uuid.Nil {
		return errors.New("replayable payload incident identity is invalid")
	}
	if err := validateReplayablePayloadShape(family, payload); err != nil {
		return err
	}

	switch family {
	case "record_changed":
		if err := validateRecordChangedPayload(payload); err != nil {
			return fmt.Errorf("invalid record_changed payload: %w", err)
		}
	case "job_progress":
		var progress JobProgressPayload
		var scope, counters json.RawMessage
		if err := decodeKnownPayloadFields(payload, payloadField{"job_id", &progress.JobID}, payloadField{"scope", &scope}, payloadField{"status", &progress.Status}, payloadField{"progress", &counters}, payloadField{"updated_at", &progress.UpdatedAt}); err != nil {
			return fmt.Errorf("invalid job_progress payload: %w", err)
		}
		if err := decodeKnownPayloadFields(scope, payloadField{"kind", &progress.Scope.Kind}, payloadField{"incident_id", &progress.Scope.IncidentID}); err != nil {
			return err
		}
		if err := decodeKnownPayloadFields(counters, payloadField{"completed", &progress.Progress.Completed}, payloadField{"total", &progress.Progress.Total}); err != nil {
			return err
		}
		if err := ValidateIncidentJobProgressPayload(incidentID, progress); err != nil {
			return fmt.Errorf("invalid job_progress payload: %w", err)
		}
	case "extension_resource_changed":
		var change ExtensionResourceChangePayload
		var refs []json.RawMessage
		if err := decodeKnownPayloadFields(payload, payloadField{"extension_profile_id", &change.ExtensionProfileID}, payloadField{"resource_kind", &change.ResourceKind}, payloadField{"resource_id", &change.ResourceID}, payloadField{"change_kind", &change.ChangeKind}, payloadField{"reason_code", &change.ReasonCode}, payloadField{"workspace_refs", &refs}); err != nil {
			return fmt.Errorf("invalid extension_resource_changed payload: %w", err)
		}
		for _, raw := range refs {
			var ref ExtensionWorkspaceRef
			if err := decodeKnownPayloadFields(raw, payloadField{"kind", &ref.Kind}, payloadField{"extension_profile_id", &ref.ExtensionProfileID}, payloadField{"workspace_key", &ref.WorkspaceKey}); err != nil {
				return err
			}
			change.WorkspaceRefs = append(change.WorkspaceRefs, ref)
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
	emittedAt, err := parseUTCInstant(message.EmittedAt)
	if err != nil || emittedAt.IsZero() {
		return errors.New("sequenced replayable emitted_at is invalid")
	}
	return ValidateReplayablePayload(incidentID, message.Type, message.Payload)
}
