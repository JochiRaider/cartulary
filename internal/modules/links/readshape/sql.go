package readshape

const (
	ActiveRecordLinksV1 = "active_record_links_v1"
	ActiveRecordTagsV1  = "active_record_tags_v1"
)

func ActiveRecordLinksAlias(alias string) string {
	return ActiveRecordLinksV1 + " " + alias
}

func ActiveRecordTagsAlias(alias string) string {
	return ActiveRecordTagsV1 + " " + alias
}

func RecordRefItemRefSQL(recordIDExpr string) string {
	return "'record_ref:' || " + recordIDExpr + "::text"
}

func PartyRefItemRefSQL(recordIDExpr string) string {
	return "'party_ref:' || " + recordIDExpr + "::text"
}

func RecordTagItemRefSQL(recordIDExpr string, recordTagIDExpr string) string {
	return "'record_tag:' || " + recordIDExpr + "::text || ':' || " + recordTagIDExpr + "::text"
}
