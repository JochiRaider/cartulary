package stream

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

const (
	EventFamilyRecordChanged           = "record_changed"
	EventFamilyJobProgress             = "job_progress"
	EventFamilyExtensionResourceChange = "extension_resource_changed"

	maxIntentPayload = 256 * 1024
)

var errIntentKeyCollision = errors.New("collaboration intent key collision")

type EventIntent struct {
	IntentKey         string
	IncidentID        uuid.UUID
	EventFamily       string
	CanonicalPayload  json.RawMessage
	SourceChangeSetID *uuid.UUID
	SourceRecordID    *uuid.UUID
	SourceRowVersion  *int64
	SourceIdentity    string
	MutationOrdinal   int
	CreatedAt         time.Time
}

// IntentWriter appends canonical event intents through transactions borrowed
// from source owners. Its zero value is ready for use and owns no database,
// clock, replay, dispatcher, or lifecycle state.
type IntentWriter struct{}

func NewEventIntent(
	intentKey string,
	incidentID uuid.UUID,
	eventFamily string,
	payload any,
	sourceIdentity string,
	mutationOrdinal int,
	createdAt time.Time,
) (EventIntent, error) {
	canonicalPayload, err := canonicalObject(payload)
	if err != nil {
		return EventIntent{}, err
	}
	intent := EventIntent{
		IntentKey:        intentKey,
		IncidentID:       incidentID,
		EventFamily:      eventFamily,
		CanonicalPayload: canonicalPayload,
		SourceIdentity:   sourceIdentity,
		MutationOrdinal:  mutationOrdinal,
		CreatedAt:        createdAt.UTC(),
	}
	if err := validateEventIntent(intent); err != nil {
		return EventIntent{}, err
	}
	return intent, nil
}

// AppendTx persists one validated intent through the caller's transaction.
func (IntentWriter) AppendTx(ctx context.Context, tx pgx.Tx, intent EventIntent) error {
	if tx == nil {
		return errors.New("collaboration intent transaction is not configured")
	}
	if err := validateEventIntent(intent); err != nil {
		return err
	}
	createdAt := intent.CreatedAt.UTC()
	tag, err := tx.Exec(ctx, `
INSERT INTO collaboration_event_intents (
    intent_key,
    incident_id,
    event_family,
    canonical_payload,
    source_change_set_id,
    source_record_id,
    source_row_version,
    source_identity,
    mutation_ordinal,
    next_attempt_at,
    created_at,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10, $10)
ON CONFLICT (intent_key) DO NOTHING
`, intent.IntentKey, intent.IncidentID, intent.EventFamily, string(intent.CanonicalPayload),
		intent.SourceChangeSetID, intent.SourceRecordID, intent.SourceRowVersion,
		intent.SourceIdentity, intent.MutationOrdinal, createdAt)
	if err != nil {
		return fmt.Errorf("append collaboration event intent: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	var exactDuplicate bool
	if err := tx.QueryRow(ctx, `
SELECT incident_id = $2
   AND event_family = $3
   AND canonical_payload = $4::jsonb
   AND source_change_set_id IS NOT DISTINCT FROM $5
   AND source_record_id IS NOT DISTINCT FROM $6
   AND source_row_version IS NOT DISTINCT FROM $7
   AND source_identity = $8
   AND mutation_ordinal = $9
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, intent.IntentKey, intent.IncidentID, intent.EventFamily, string(intent.CanonicalPayload),
		intent.SourceChangeSetID, intent.SourceRecordID, intent.SourceRowVersion,
		intent.SourceIdentity, intent.MutationOrdinal).Scan(&exactDuplicate); err != nil {
		return fmt.Errorf("verify collaboration event intent replay: %w", err)
	}
	if exactDuplicate {
		return nil
	}
	var (
		existingPayload  []byte
		existingIdentity string
		existingOrdinal  int
	)
	if err := tx.QueryRow(ctx, `
SELECT canonical_payload::text, source_identity, mutation_ordinal
  FROM collaboration_event_intents
 WHERE intent_key = $1
`, intent.IntentKey).Scan(&existingPayload, &existingIdentity, &existingOrdinal); err != nil {
		return fmt.Errorf("inspect collaboration intent collision: %w", err)
	}
	existingDigest := sha256.Sum256(existingPayload)
	incomingDigest := sha256.Sum256(intent.CanonicalPayload)
	return fmt.Errorf(
		"%w: %s existing_payload_sha256=%x incoming_payload_sha256=%x payload_mismatch_keys=%v existing_source_identity=%q incoming_source_identity=%q existing_ordinal=%d incoming_ordinal=%d",
		errIntentKeyCollision,
		intent.IntentKey,
		existingDigest,
		incomingDigest,
		payloadMismatchKeys(existingPayload, intent.CanonicalPayload),
		existingIdentity,
		intent.SourceIdentity,
		existingOrdinal,
		intent.MutationOrdinal,
	)
}

func payloadMismatchKeys(existing []byte, incoming []byte) []string {
	var existingObject map[string]json.RawMessage
	var incomingObject map[string]json.RawMessage
	if json.Unmarshal(existing, &existingObject) != nil || json.Unmarshal(incoming, &incomingObject) != nil {
		return []string{"<malformed>"}
	}
	keys := make(map[string]struct{}, len(existingObject)+len(incomingObject))
	for key := range existingObject {
		keys[key] = struct{}{}
	}
	for key := range incomingObject {
		keys[key] = struct{}{}
	}
	mismatches := make([]string, 0)
	for key := range keys {
		if !bytes.Equal(existingObject[key], incomingObject[key]) {
			mismatches = append(mismatches, key)
		}
	}
	slices.Sort(mismatches)
	return mismatches
}

func validateEventIntent(intent EventIntent) error {
	if strings.TrimSpace(intent.IntentKey) == "" || len(intent.IntentKey) > 512 {
		return errors.New("collaboration event intent key is invalid")
	}
	if intent.IncidentID == uuid.Nil || strings.TrimSpace(intent.SourceIdentity) == "" || len(intent.SourceIdentity) > 512 {
		return errors.New("collaboration event intent source identity is invalid")
	}
	switch intent.EventFamily {
	case EventFamilyRecordChanged, EventFamilyJobProgress, EventFamilyExtensionResourceChange:
	default:
		return fmt.Errorf("unsupported collaboration event family %q", intent.EventFamily)
	}
	if intent.MutationOrdinal < 0 || intent.CreatedAt.IsZero() {
		return errors.New("collaboration event intent ordering is invalid")
	}
	var payload map[string]any
	if len(intent.CanonicalPayload) == 0 || json.Unmarshal(intent.CanonicalPayload, &payload) != nil || payload == nil {
		return errors.New("collaboration event intent payload must be a JSON object")
	}
	if len(intent.CanonicalPayload) > maxIntentPayload {
		return fmt.Errorf("collaboration event intent payload exceeds %d bytes", maxIntentPayload)
	}
	return ValidateEventFamilyPayload(intent.IncidentID, intent.EventFamily, intent.CanonicalPayload)
}

func ValidateEventFamilyPayload(incidentID uuid.UUID, family string, payload json.RawMessage) error {
	switch family {
	case EventFamilyRecordChanged:
		message := replayMessage(
			uuid.MustParse("00000000-0000-4000-8000-000000000001"),
			incidentID,
			family,
			1,
			payload,
			time.Unix(0, 0).UTC(),
		)
		if err := validateRecordChangeMessage(message); err != nil {
			return fmt.Errorf("invalid record-change payload: %w", err)
		}
	case EventFamilyJobProgress:
		var progress protocol.JobProgressPayload
		if err := json.Unmarshal(payload, &progress); err != nil {
			return fmt.Errorf("invalid job-progress payload: %w", err)
		}
		if err := protocol.ValidateIncidentJobProgressPayload(incidentID, progress); err != nil {
			return fmt.Errorf("invalid job-progress payload: %w", err)
		}
	case EventFamilyExtensionResourceChange:
		var change protocol.ExtensionResourceChangePayload
		if err := json.Unmarshal(payload, &change); err != nil {
			return fmt.Errorf("invalid extension-resource-change payload: %w", err)
		}
		if err := protocol.ValidateExtensionResourceChangePayload(change); err != nil {
			return fmt.Errorf("invalid extension-resource-change payload: %w", err)
		}
	default:
		return fmt.Errorf("unsupported collaboration event family %q", family)
	}
	return nil
}

func canonicalObject(payload any) (json.RawMessage, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode collaboration event payload: %w", err)
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil || object == nil {
		return nil, errors.New("collaboration event payload must be an object")
	}
	canonical, err := json.Marshal(object)
	if err != nil {
		return nil, fmt.Errorf("canonicalize collaboration event payload: %w", err)
	}
	return canonical, nil
}
