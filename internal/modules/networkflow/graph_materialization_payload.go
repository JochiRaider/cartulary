package networkflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/google/uuid"
	"regexp"
)

var sourceSnapshotIDPattern = regexp.MustCompile(`^nfsnap_[a-f0-9]{64}$`)
var errMaterializationPayload = errors.New("invalid Network Flow materialization payload")

// Rejected bytes never produce an identity usable by a failure mutation.
func decodeGraphViewMaterializationPayload(raw []byte, incident uuid.UUID) (graphViewMaterializationPayload, error) {
	object, err := strictjson.DecodeObject(bytes.NewReader(raw))
	if err != nil || len(object) != 5 || incident == uuid.Nil {
		return graphViewMaterializationPayload{}, errMaterializationPayload
	}
	for _, field := range []string{"schema_id", "incident_id", "graph_view_id", "materialization_generation", "source_snapshot_id"} {
		value, ok := object[field]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return graphViewMaterializationPayload{}, errMaterializationPayload
		}
	}
	var payload graphViewMaterializationPayload
	if json.Unmarshal(raw, &payload) != nil || !payload.valid() || payload.IncidentID != incident {
		return graphViewMaterializationPayload{}, errMaterializationPayload
	}
	var incidentText string
	if json.Unmarshal(object["incident_id"], &incidentText) != nil || incidentText != incident.String() {
		return graphViewMaterializationPayload{}, errMaterializationPayload
	}
	return payload, nil
}
