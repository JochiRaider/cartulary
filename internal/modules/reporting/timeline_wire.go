package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
)

func (event *ReportingTimelineEvent) UnmarshalJSON(raw []byte) error {
	type wireEvent ReportingTimelineEvent
	var decoded wireEvent
	if err := decodeTimelineObject(raw, map[string]bool{
		"schema_id": false, "record_id": false, "activity_sort_key": false,
		"activity_sort_ts": true, "activity_parse_state": false, "activity_precision_rank": false,
		"date_entered_sort_key": true, "record_created_at": false, "display_times": false,
		"fields": false, "source_refs": false, "support_refs": false, "disclosure_partition_refs": false,
	}, &decoded); err != nil {
		return err
	}
	*event = ReportingTimelineEvent(decoded)
	return nil
}

func (display *ReportingDisplayTimes) UnmarshalJSON(raw []byte) error {
	type wireDisplay ReportingDisplayTimes
	var decoded wireDisplay
	if err := decodeTimelineObject(raw, map[string]bool{
		"primary_display": true, "precision_label": true, "source_text_display": true,
	}, &decoded); err != nil {
		return err
	}
	*display = ReportingDisplayTimes(decoded)
	return nil
}

// Nullable is distinct from optional: every member must appear, including null
// values. Decode into a fresh value so failed admission cannot mutate a caller.
func decodeTimelineObject(raw []byte, nullable map[string]bool, destination any) error {
	canonical, err := canonicalReportingBytes(raw)
	if err != nil {
		return err
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(canonical, &members); err != nil {
		return err
	}
	if len(members) != len(nullable) {
		return errors.New("invalid closed Timeline member set")
	}
	for key, value := range members {
		allowNull, exists := nullable[key]
		if !exists || (!allowNull && bytes.Equal(value, []byte("null"))) {
			return errors.New("unknown or null required Timeline member")
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	return decoder.Decode(destination)
}
