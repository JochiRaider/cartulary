package timeline

import "github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"

func projectRecordHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("timeline", "text", "date_entered_text analyst_text mitre_stage_text device_object_text ip_address_text activity_utc_text activity_local_text raw_activity_text activity_synopsis_text data_source_text activity_time_pair_state capture_state reviewed_at reviewed_by_user_id superseded_at superseded_by_user_id")
	fields = append(fields, historycontract.Fields("timeline", "boolean", "activity_utc_generated activity_local_generated")...)
	for i := range fields {
		if fields[i].Member == "capture_state" {
			fields[i].Kind = "capture_state"
			fields[i].Required = true
		}
	}
	return historycontract.Row(facts, fields)
}
