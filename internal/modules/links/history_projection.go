package links

import "github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"

func projectLinkHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("link", "text", "record_link_id src_record_id dst_record_id link_type field_key provenance owner_user_id decided_at deleted_at")
	fields = append(fields, historycontract.Fields("link", "number", "confidence")...)
	kind := "link"
	value := facts.After
	if value == nil {
		value = facts.Before
	}
	if value["link_type"] == "attached_evidence" {
		kind = "evidence_association"
	}
	return historycontract.Collection(facts, kind, fields, []string{"record_link_id"}, []string{"src_record_id", "dst_record_id"})
}

func projectTagHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("tag", "text", "record_tag_id record_id tag_name normalized_tag_name deleted_at")
	return historycontract.Collection(facts, "tag", fields, []string{"record_tag_id"}, []string{"record_id"})
}
