package projections

func evidenceQuerySurfaces() []genericSurface {
	return []genericSurface{{
		viewSchemaID: evidenceViewSchemaID,
		fromSQL:      "FROM evidence_grid_projection p JOIN records r ON r.record_id = p.record_id",
		recordExpr:   "p.record_id",
		incidentExpr: "p.incident_id",
		fields: []genericField{
			{key: "evidence.title", expr: "p.title", kind: fieldKindText},
			{key: "evidence.lifecycle_state", expr: "p.lifecycle_state", kind: fieldKindText},
			{key: "evidence.requested_at", expr: "p.requested_at", kind: fieldKindTimestamp},
			{key: "evidence.received_at", expr: "p.received_at", kind: fieldKindTimestamp},
			{key: "evidence.storage_ref", expr: "p.storage_ref", kind: fieldKindText},
			{key: "evidence.blob_hash", expr: "p.blob_hash", kind: fieldKindText},
			{key: "evidence.collector_party_text", expr: "p.collector_party_text", kind: fieldKindText},
			{key: "evidence.collector_party_id", expr: "p.collector_party_id", kind: fieldKindText},
			{key: "evidence.source_party_text", expr: "p.source_party_text", kind: fieldKindText},
			{key: "evidence.source_party_id", expr: "p.source_party_id", kind: fieldKindText},
			{key: "evidence.upload_state", expr: "p.upload_state", kind: fieldKindText},
			{key: "evidence.linked_record_count", expr: "p.linked_record_count", kind: fieldKindNumber},
			{key: "evidence.edited_at", expr: "p.edited_at", kind: fieldKindTimestamp},
		},
	}}
}
