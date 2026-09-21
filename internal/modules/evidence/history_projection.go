package evidence

import "github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"

func projectRecordHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("evidence", "text", "title lifecycle_state requested_at received_at storage_ref blob_hash collector_party_text collector_party_id source_party_text source_party_id upload_state")
	for i := range fields {
		if fields[i].Member == "lifecycle_state" || fields[i].Member == "upload_state" {
			fields[i].Required = true
		}
	}
	return historycontract.Row(facts, fields)
}
