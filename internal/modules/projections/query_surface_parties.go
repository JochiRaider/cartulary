package projections

func partyQuerySurfaces() []genericSurface {
	return []genericSurface{{
		viewSchemaID: partiesViewSchemaID,
		fromSQL:      "FROM party_grid_projection p JOIN records r ON r.record_id = p.record_id",
		recordExpr:   "p.record_id",
		incidentExpr: "p.incident_id",
		fields: []genericField{
			{key: "party.display_name", expr: "p.display_name", kind: fieldKindText},
			{key: "party.party_kind", expr: "p.party_kind", kind: fieldKindText},
			{key: "party.organization_name", expr: "p.organization_name", kind: fieldKindText},
			{key: "party.role_title", expr: "p.role_title", kind: fieldKindText},
			{key: "party.primary_email", expr: "p.primary_email", kind: fieldKindText},
			{key: "party.timezone_name", expr: "p.timezone_name", kind: fieldKindText},
			{key: "party.external_ref", expr: "p.external_ref", kind: fieldKindText},
			{key: "party.notes", expr: "p.notes", kind: fieldKindText},
			{key: "party.updated_at", expr: "p.updated_at", kind: fieldKindTimestamp},
		},
	}}
}
