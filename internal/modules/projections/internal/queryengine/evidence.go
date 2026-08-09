package queryengine

const evidenceViewSchemaID = "cartulary.view.evidence.v1"

func EvidencePlans() []Surface {
	return []Surface{{
		ViewSchemaID: evidenceViewSchemaID,
		FromSQL:      "FROM evidence_grid_projection p JOIN records r ON r.record_id = p.record_id",
		RecordExpr:   "p.record_id",
		IncidentExpr: "p.incident_id",
		Fields: []Field{
			{Key: "evidence.title", Expr: "p.title", Kind: FieldKindText},
			{Key: "evidence.lifecycle_state", Expr: "p.lifecycle_state", Kind: FieldKindText},
			{Key: "evidence.requested_at", Expr: "p.requested_at", Kind: FieldKindTimestamp},
			{Key: "evidence.received_at", Expr: "p.received_at", Kind: FieldKindTimestamp},
			{Key: "evidence.storage_ref", Expr: "p.storage_ref", Kind: FieldKindText},
			{Key: "evidence.blob_hash", Expr: "p.blob_hash", Kind: FieldKindText},
			{Key: "evidence.collector_party_text", Expr: "p.collector_party_text", Kind: FieldKindText},
			{Key: "evidence.collector_party_id", Expr: "p.collector_party_id", Kind: FieldKindText},
			{Key: "evidence.source_party_text", Expr: "p.source_party_text", Kind: FieldKindText},
			{Key: "evidence.source_party_id", Expr: "p.source_party_id", Kind: FieldKindText},
			{Key: "evidence.upload_state", Expr: "p.upload_state", Kind: FieldKindText},
			{Key: "evidence.linked_record_count", Expr: "p.linked_record_count", Kind: FieldKindNumber},
			{Key: "evidence.edited_at", Expr: "p.edited_at", Kind: FieldKindTimestamp},
		},
	}}
}
