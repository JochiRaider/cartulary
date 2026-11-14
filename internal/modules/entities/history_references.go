package entities

import (
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func entityHistoryReferences(mutation revisions.StoredMutation, importIncident uuid.UUID) ([]revisions.HistoryRecordReference, error) {
	var required, nullable string
	switch mutation.TargetKind {
	case "entity_mention":
		required, nullable = "source_record_id", "resolved_record_id"
	case "entity_alias", "entity_preserved_identifier":
		required = "record_id"
	default:
		return nil, revisions.ErrInvalidTargetSemantics
	}
	canonical := func(value any) (uuid.UUID, bool) {
		text, ok := value.(string)
		if !ok {
			return uuid.Nil, false
		}
		id, err := uuid.Parse(text)
		return id, err == nil && id != uuid.Nil && id.String() == text
	}
	var references []revisions.HistoryRecordReference
	for _, value := range []map[string]any{mutation.BeforeValue, mutation.AfterValue} {
		if value == nil {
			continue
		}
		incidentID := importIncident
		if mutation.TargetKind != "entity_mention" {
			var valid bool
			incidentID, valid = canonical(value["incident_id"])
			if !valid {
				return nil, revisions.ErrInvalidTargetSemantics
			}
		}
		if incidentID == uuid.Nil {
			return nil, revisions.ErrInvalidTargetSemantics
		}
		recordID, valid := canonical(value[required])
		if !valid {
			return nil, revisions.ErrInvalidTargetSemantics
		}
		references = append(references, revisions.HistoryRecordReference{IncidentID: incidentID, RecordID: recordID})
		if nullable != "" {
			raw, present := value[nullable]
			if !present {
				return nil, revisions.ErrInvalidTargetSemantics
			}
			if raw != nil {
				recordID, valid := canonical(raw)
				if !valid {
					return nil, revisions.ErrInvalidTargetSemantics
				}
				references = append(references, revisions.HistoryRecordReference{IncidentID: incidentID, RecordID: recordID})
			}
		}
	}
	return references, nil
}
