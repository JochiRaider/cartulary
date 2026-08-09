package queryengine

const partyViewSchemaID = "cartulary.view.parties.v1"

func PartyPlans() []Surface {
	return []Surface{{
		ViewSchemaID: partyViewSchemaID,
		FromSQL:      "FROM party_grid_projection p JOIN records r ON r.record_id = p.record_id",
		RecordExpr:   "p.record_id",
		IncidentExpr: "p.incident_id",
		Fields: []Field{
			{Key: "party.display_name", Expr: "p.display_name", Kind: FieldKindText},
			{Key: "party.party_kind", Expr: "p.party_kind", Kind: FieldKindText},
			{Key: "party.organization_name", Expr: "p.organization_name", Kind: FieldKindText},
			{Key: "party.role_title", Expr: "p.role_title", Kind: FieldKindText},
			{Key: "party.primary_email", Expr: "p.primary_email", Kind: FieldKindText},
			{Key: "party.timezone_name", Expr: "p.timezone_name", Kind: FieldKindText},
			{Key: "party.external_ref", Expr: "p.external_ref", Kind: FieldKindText},
			{Key: "party.notes", Expr: "p.notes", Kind: FieldKindText},
			{Key: "party.updated_at", Expr: "p.updated_at", Kind: FieldKindTimestamp},
		},
	}}
}
