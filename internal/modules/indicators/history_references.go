package indicators

import (
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func indicatorHistoryReferences(mutation revisions.StoredMutation, _ uuid.UUID) ([]revisions.HistoryRecordReference, error) {
	var required, nullable string
	switch mutation.TargetKind {
	case "indicator_observation":
		required, nullable = "source_record_id", "resolved_indicator_record_id"
	case "indicator_state_interval":
		required = "indicator_record_id"
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
		incidentID, valid := canonical(value["incident_id"])
		if !valid {
			return nil, revisions.ErrInvalidTargetSemantics
		}
		recordID, valid := canonical(value[required])
		if !valid {
			return nil, revisions.ErrInvalidTargetSemantics
		}
		references = append(references, revisions.HistoryRecordReference{IncidentID: incidentID, RecordID: recordID})
		if mutation.TargetKind == "indicator_state_interval" {
			var support []any
			switch values := value["support_refs"].(type) {
			case []any:
				if values == nil {
					return nil, revisions.ErrInvalidTargetSemantics
				}
				support = values
			case []string:
				if values == nil {
					return nil, revisions.ErrInvalidTargetSemantics
				}
				support = make([]any, len(values))
				for index, text := range values {
					support[index] = text
				}
			default:
				return nil, revisions.ErrInvalidTargetSemantics
			}
			seen := make(map[uuid.UUID]bool, len(support))
			for _, raw := range support {
				id, valid := canonical(raw)
				if !valid || seen[id] {
					return nil, revisions.ErrInvalidTargetSemantics
				}
				seen[id] = true
				references = append(references, revisions.HistoryRecordReference{IncidentID: incidentID, RecordID: id})
			}
		}
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
