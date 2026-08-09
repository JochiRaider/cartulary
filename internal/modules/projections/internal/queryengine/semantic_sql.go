package queryengine

func activeRecordLinksAlias(alias string) string {
	return "active_record_links_v1 " + alias
}

func activeRecordTagsAlias(alias string) string {
	return "active_record_tags_v1 " + alias
}

func recordRefItemRefSQL(recordIDExpr string) string {
	return "'record_ref:' || " + recordIDExpr + "::text"
}

func partyRefItemRefSQL(recordIDExpr string) string {
	return "'party_ref:' || " + recordIDExpr + "::text"
}

func recordTagItemRefSQL(recordIDExpr string, recordTagIDExpr string) string {
	return "'record_tag:' || " + recordIDExpr + "::text || ':' || " + recordTagIDExpr + "::text"
}
