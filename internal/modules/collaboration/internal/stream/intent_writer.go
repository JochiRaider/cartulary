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
	intentKey         string
	incidentID        uuid.UUID
	eventFamily       string
	canonicalPayload  json.RawMessage
	sourceChangeSetID *uuid.UUID
	sourceRecordID    *uuid.UUID
	sourceRowVersion  *int64
	sourceIdentity    string
	mutationOrdinal   int
	createdAt         time.Time
}

// IntentWriter appends canonical event intents through transactions borrowed
// from source owners. Its zero value is ready for use and owns no database,
// clock, replay, dispatcher, or lifecycle state.
type IntentWriter struct{}

func NewRecordChangedIntent(
	intentKey string,
	incidentID uuid.UUID,
	payload any,
	sourceChangeSetID uuid.UUID,
	sourceRecordID uuid.UUID,
	sourceRowVersion int64,
	sourceIdentity string,
	mutationOrdinal int,
	createdAt time.Time,
) (EventIntent, error) {
	if sourceChangeSetID == uuid.Nil || sourceRecordID == uuid.Nil || sourceRowVersion < 1 {
		return EventIntent{}, errors.New("collaboration record-change identity is invalid")
	}
	canonicalPayload, err := canonicalObject(payload)
	if err != nil {
		return EventIntent{}, err
	}
	intent := EventIntent{
		intentKey: intentKey, incidentID: incidentID, eventFamily: EventFamilyRecordChanged,
		canonicalPayload: canonicalPayload, sourceChangeSetID: &sourceChangeSetID,
		sourceRecordID: &sourceRecordID, sourceRowVersion: &sourceRowVersion,
		sourceIdentity: sourceIdentity, mutationOrdinal: mutationOrdinal, createdAt: createdAt.UTC(),
	}
	if err := validateEventIntent(intent); err != nil {
		return EventIntent{}, err
	}
	return intent, nil
}

func NewJobProgressIntent(
	intentKey string,
	incidentID uuid.UUID,
	payload any,
	sourceIdentity string,
	createdAt time.Time,
) (EventIntent, error) {
	canonicalPayload, err := canonicalObject(payload)
	if err != nil {
		return EventIntent{}, err
	}
	intent := EventIntent{
		intentKey: intentKey, incidentID: incidentID, eventFamily: EventFamilyJobProgress,
		canonicalPayload: canonicalPayload, sourceIdentity: sourceIdentity, createdAt: createdAt.UTC(),
	}
	if err := validateEventIntent(intent); err != nil {
		return EventIntent{}, err
	}
	return intent, nil
}

func NewExtensionResourceChangedIntent(
	intentKey string,
	incidentID uuid.UUID,
	payload any,
	sourceIdentity string,
	createdAt time.Time,
) (EventIntent, error) {
	canonicalPayload, err := canonicalObject(payload)
	if err != nil {
		return EventIntent{}, err
	}
	intent := EventIntent{
		intentKey: intentKey, incidentID: incidentID, eventFamily: EventFamilyExtensionResourceChange,
		canonicalPayload: canonicalPayload, sourceIdentity: sourceIdentity, createdAt: createdAt.UTC(),
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
	createdAt := intent.createdAt.UTC()
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
`, intent.intentKey, intent.incidentID, intent.eventFamily, string(intent.canonicalPayload),
		intent.sourceChangeSetID, intent.sourceRecordID, intent.sourceRowVersion,
		intent.sourceIdentity, intent.mutationOrdinal, createdAt)
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
`, intent.intentKey, intent.incidentID, intent.eventFamily, string(intent.canonicalPayload),
		intent.sourceChangeSetID, intent.sourceRecordID, intent.sourceRowVersion,
		intent.sourceIdentity, intent.mutationOrdinal).Scan(&exactDuplicate); err != nil {
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
`, intent.intentKey).Scan(&existingPayload, &existingIdentity, &existingOrdinal); err != nil {
		return fmt.Errorf("inspect collaboration intent collision: %w", err)
	}
	existingDigest := sha256.Sum256(existingPayload)
	incomingDigest := sha256.Sum256(intent.canonicalPayload)
	return fmt.Errorf(
		"%w: %s existing_payload_sha256=%x incoming_payload_sha256=%x payload_mismatch_keys=%v existing_source_identity=%q incoming_source_identity=%q existing_ordinal=%d incoming_ordinal=%d",
		errIntentKeyCollision,
		intent.intentKey,
		existingDigest,
		incomingDigest,
		payloadMismatchKeys(existingPayload, intent.canonicalPayload),
		existingIdentity,
		intent.sourceIdentity,
		existingOrdinal,
		intent.mutationOrdinal,
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
	if strings.TrimSpace(intent.intentKey) == "" || len(intent.intentKey) > 512 {
		return errors.New("collaboration event intent key is invalid")
	}
	if intent.incidentID == uuid.Nil || strings.TrimSpace(intent.sourceIdentity) == "" || len(intent.sourceIdentity) > 512 {
		return errors.New("collaboration event intent source identity is invalid")
	}
	switch intent.eventFamily {
	case EventFamilyRecordChanged:
		if intent.sourceChangeSetID == nil || *intent.sourceChangeSetID == uuid.Nil ||
			intent.sourceRecordID == nil || *intent.sourceRecordID == uuid.Nil ||
			intent.sourceRowVersion == nil || *intent.sourceRowVersion < 1 {
			return errors.New("collaboration record-change identity is invalid")
		}
	case EventFamilyJobProgress, EventFamilyExtensionResourceChange:
		if intent.sourceChangeSetID != nil || intent.sourceRecordID != nil || intent.sourceRowVersion != nil {
			return errors.New("collaboration non-record intent has record identity")
		}
	default:
		return fmt.Errorf("unsupported collaboration event family %q", intent.eventFamily)
	}
	if intent.mutationOrdinal < 0 || intent.createdAt.IsZero() {
		return errors.New("collaboration event intent ordering is invalid")
	}
	var payload map[string]any
	if len(intent.canonicalPayload) == 0 || json.Unmarshal(intent.canonicalPayload, &payload) != nil || payload == nil {
		return errors.New("collaboration event intent payload must be a JSON object")
	}
	if len(intent.canonicalPayload) > maxIntentPayload {
		return fmt.Errorf("collaboration event intent payload exceeds %d bytes", maxIntentPayload)
	}
	return protocol.ValidateReplayablePayload(intent.incidentID, intent.eventFamily, intent.canonicalPayload)
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
