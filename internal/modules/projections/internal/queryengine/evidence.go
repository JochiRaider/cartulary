package queryengine

import "github.com/JochiRaider/cartulary/internal/platform/sqlshape"

const evidenceViewSchemaID = "cartulary.view.evidence.v1"

func EvidencePlans() []Surface {
	return []Surface{{
		ViewSchemaID: evidenceViewSchemaID,
		FromSQL:      sqlshape.Literal("FROM evidence_grid_projection p JOIN records r ON r.record_id = p.record_id"),
		RecordExpr:   sqlshape.Literal("p.record_id"),
		IncidentExpr: sqlshape.Literal("p.incident_id"),
		Fields: []Field{
			{Key: "evidence.title", Expr: sqlshape.Literal("p.title"), Kind: FieldKindText},
			{Key: "evidence.lifecycle_state", Expr: sqlshape.Literal("p.lifecycle_state"), Kind: FieldKindText},
			{Key: "evidence.requested_at", Expr: sqlshape.Literal("p.requested_at"), Kind: FieldKindTimestamp},
			{Key: "evidence.received_at", Expr: sqlshape.Literal("p.received_at"), Kind: FieldKindTimestamp},
			{Key: "evidence.storage_ref", Expr: sqlshape.Literal("p.storage_ref"), Kind: FieldKindText},
			{Key: "evidence.blob_hash", Expr: sqlshape.Literal("p.blob_hash"), Kind: FieldKindText},
			{Key: "evidence.collector_party_text", Expr: sqlshape.Literal("p.collector_party_text"), Kind: FieldKindText},
			{Key: "evidence.collector_party_id", Expr: sqlshape.Literal("p.collector_party_id"), Kind: FieldKindText},
			{Key: "evidence.source_party_text", Expr: sqlshape.Literal("p.source_party_text"), Kind: FieldKindText},
			{Key: "evidence.source_party_id", Expr: sqlshape.Literal("p.source_party_id"), Kind: FieldKindText},
			{Key: "evidence.owner_user_id", Expr: sqlshape.Literal("p.owner_user_id"), Kind: FieldKindText},
			{Key: "evidence.upload_state", Expr: sqlshape.Literal("p.upload_state"), Kind: FieldKindText},
			{Key: "evidence.linked_record_count", Expr: sqlshape.Literal("p.linked_record_count"), Kind: FieldKindNumber},
			{Key: "evidence.edited_at", Expr: sqlshape.Literal("p.edited_at"), Kind: FieldKindTimestamp},
		},
	}}
}
