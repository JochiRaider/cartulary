package records

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ValidateSnapshotEnvelope checks a historical Records row without changing it.
// Actor catalog membership is checked by the history import boundary; this
// validator owns the complete envelope's scalar, clock and deletion rules.
func ValidateSnapshotEnvelope(incidentID, recordID uuid.UUID, recordType string, value map[string]any) error {
	if incidentID == uuid.Nil || recordID == uuid.Nil {
		return recordsInvariantFailure("records.envelope_legal")
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return recordsInvariantFailure("records.envelope_legal")
	}
	var row map[string]any
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&row); err != nil {
		return recordsInvariantFailure("records.envelope_legal")
	}
	// PostgreSQL snapshots spell UTC as +00:00. Adapt only the known clock
	// fields for the portable row validator, preserving the actual history.
	for _, key := range []string{"created_at", "updated_at", "deleted_at"} {
		if value, ok := row[key].(string); ok {
			canonical := value
			if strings.HasSuffix(value, "+00:00") {
				canonical = strings.TrimSuffix(value, "+00:00") + "Z"
			}
			parsed, ok := parseCanonicalTimestamp(canonical)
			if !ok {
				return recordsInvariantFailure("records.envelope_legal")
			}
			row[key] = parsed.Format(time.RFC3339Nano)
		}
	}
	prepared, err := prepareRecordsRow(row, recordsImportContext{
		IncidentID:    incidentID,
		ActorAdmitted: func(value string) bool { return value != uuid.Nil.String() },
	})
	if err != nil {
		return err
	}
	if prepared.RecordID != recordID || string(prepared.RecordType) != recordType {
		return recordsInvariantFailure("records.envelope_legal")
	}
	return nil
}
