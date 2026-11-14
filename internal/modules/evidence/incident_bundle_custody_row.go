package evidence

import (
	"encoding/json"
	"maps"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
)

const portableCustodyPath = "data/evidence_custody_events.ndjson"

type portableCustodyRow struct {
	CustodyEventID   uuid.UUID       `json:"custody_event_id"`
	IncidentID       uuid.UUID       `json:"incident_id"`
	EvidenceRecordID uuid.UUID       `json:"evidence_record_id"`
	CustodyEventType string          `json:"custody_event_type"`
	ActorUserID      *uuid.UUID      `json:"actor_user_id"`
	OccurredAt       string          `json:"occurred_at"`
	LocationText     *string         `json:"location_text"`
	Note             *string         `json:"note"`
	Metadata         json.RawMessage `json:"metadata"`
}

func parsePortableCustodyShape(raw map[string]any) (portableCustodyRow, bool) {
	fields := strings.Fields("custody_event_id incident_id evidence_record_id custody_event_type actor_user_id occurred_at location_text note metadata")
	if len(raw) != len(fields) {
		return portableCustodyRow{}, false
	}
	for _, field := range fields {
		value, present := raw[field]
		if !present {
			return portableCustodyRow{}, false
		}
		if text, ok := value.(string); ok && !custodyString(text) {
			return portableCustodyRow{}, false
		}
	}
	for _, field := range []string{"custody_event_id", "incident_id", "evidence_record_id", "actor_user_id"} {
		if field == "actor_user_id" && raw[field] == nil {
			continue
		}
		text, ok := raw[field].(string)
		id, err := uuid.Parse(text)
		if !ok || err != nil || id == uuid.Nil || id.String() != text {
			return portableCustodyRow{}, false
		}
	}
	for _, field := range []string{"custody_event_type", "occurred_at", "location_text", "note"} {
		if (field == "location_text" || field == "note") && raw[field] == nil {
			continue
		}
		if _, ok := raw[field].(string); !ok {
			return portableCustodyRow{}, false
		}
	}
	switch raw["custody_event_type"] {
	case "requested", "received", "made_available", "transferred", "quarantined", "released":
	default:
		return portableCustodyRow{}, false
	}
	metadata, ok := custodyJSON(raw["metadata"])
	if !ok {
		return portableCustodyRow{}, false
	}
	source := maps.Clone(raw)
	source["metadata"] = metadata
	payload, err := json.Marshal(source)
	var row portableCustodyRow
	if err != nil || len(payload) > incidentportability.MaxNDJSONLogicalLineBytes || json.Unmarshal(payload, &row) != nil {
		return portableCustodyRow{}, false
	}
	return row, true
}
