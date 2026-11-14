package rollback

import (
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/google/uuid"
)

// parseCompleteIndicatorSource owns source admission; Revisions validates the
// complete Records envelope and schema before invoking the row provider.
func parseCompleteIndicatorSource(value map[string]any) (rowState, error) {
	record, recordOK := value["record"].(map[string]any)
	source, sourceOK := value["source"].(map[string]any)
	if !recordOK || !sourceOK || sourcevalue.ValidateSnapshotSource(record, source) != nil {
		return rowState{}, rollbackcontract.ErrTargetNotReversible
	}
	nullable := func(key string) *string {
		if source[key] == nil {
			return nil
		}
		text := source[key].(string)
		return &text
	}
	return rowState{recordID: uuid.MustParse(source["record_id"].(string)), incidentID: uuid.MustParse(source["incident_id"].(string)),
		indicatorType: source["indicator_type"].(string), valueKind: source["value_kind"].(string), displayValue: source["display_value"].(string),
		normalized: nullable("normalized_value"), dedupeKey: source["dedupe_key"].(string), defanged: nullable("defanged_value"), hashAlgorithm: nullable("hash_algorithm"), hashValue: nullable("hash_value"), stixPattern: nullable("stix_pattern")}, nil
}
