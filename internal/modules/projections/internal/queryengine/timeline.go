package queryengine

const timelineViewSchemaID = "cartulary.view.timeline.v2"

func TimelinePlans() []Surface {
	return []Surface{{
		ViewSchemaID: timelineViewSchemaID,
		FromSQL:      `FROM timeline_grid_projection t JOIN records r ON r.record_id = t.record_id`,
		RecordExpr:   "t.record_id",
		IncidentExpr: "t.incident_id",
		Fields: []Field{
			{Key: "timeline.date_entered_text", Expr: "t.date_entered_text", Kind: FieldKindText},
			{Key: "timeline.analyst_text", Expr: "t.analyst_text", Kind: FieldKindText},
			{Key: "timeline.mitre_stage_text", Expr: "t.mitre_stage_text", Kind: FieldKindText},
			{Key: "timeline.device_object_text", Expr: "t.device_object_text", Kind: FieldKindText},
			{Key: "timeline.ip_address_text", Expr: "t.ip_address_text", Kind: FieldKindText},
			{Key: "timeline.activity_utc_text", Expr: "t.activity_utc_text", Kind: FieldKindText},
			{Key: "timeline.activity_local_text", Expr: "t.activity_local_text", Kind: FieldKindText},
			{Key: "timeline.raw_activity_text", Expr: "t.raw_activity_text", Kind: FieldKindText},
			{Key: "timeline.activity_synopsis_text", Expr: "t.activity_synopsis_text", Kind: FieldKindText},
			{Key: "timeline.data_source_text", Expr: "t.data_source_text", Kind: FieldKindText},
			{Key: "timeline.host_refs", Expr: "t.host_refs", Kind: FieldKindCollection, Ordered: true},
			{Key: "timeline.identity_refs", Expr: "t.identity_refs", Kind: FieldKindCollection, Ordered: true},
			{Key: "timeline.tags", Expr: "t.tags", Kind: FieldKindCollection},
			{Key: "timeline.attached_evidence_ids", Expr: "t.attached_evidence_refs", Kind: FieldKindCollection},
			{Key: "timeline.evidence_count", Expr: "t.evidence_count", Kind: FieldKindNumber},
			{Key: "timeline.recorded_at", Expr: "t.recorded_at", Kind: FieldKindTimestamp},
			{Key: "timeline.edited_at", Expr: "t.edited_at", Kind: FieldKindTimestamp},
			{Key: "timeline.activity_sort_ts", Expr: "t.activity_sort_ts", Kind: FieldKindTimestamp},
			{Key: "timeline.date_entered_sort_day", Expr: "t.date_entered_sort_day", Kind: FieldKindDate},
			{Key: "timeline.activity_time_pair_state", Expr: "t.activity_time_pair_state", Kind: FieldKindText},
			{Key: "timeline.capture_state", Expr: "t.capture_state", Kind: FieldKindText},
			{Key: "timeline.replacement_record_id", Expr: "t.replacement_record_id", Kind: FieldKindText},
			{Key: "timeline.has_evidence", Expr: "t.has_evidence", Kind: FieldKindBool},
			{Key: "timeline.has_unresolved_mentions", Expr: "t.has_unresolved_mentions", Kind: FieldKindBool},
		},
	}}
}
