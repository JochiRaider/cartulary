package projectionprovider

import "github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"

const partiesViewSchemaID = "cartulary.view.parties.v1"

func QuerySurfaces() []providercontract.QuerySurface {
	return []providercontract.QuerySurface{{
		ViewSchemaID: partiesViewSchemaID,
		FromSQL:      "FROM party_grid_projection p JOIN records r ON r.record_id = p.record_id",
		RecordExpr:   "p.record_id",
		IncidentExpr: "p.incident_id",
		Fields: []providercontract.QueryField{
			{Key: "party.display_name", Expr: "p.display_name", Kind: providercontract.FieldKindText},
			{Key: "party.party_kind", Expr: "p.party_kind", Kind: providercontract.FieldKindText},
			{Key: "party.organization_name", Expr: "p.organization_name", Kind: providercontract.FieldKindText},
			{Key: "party.role_title", Expr: "p.role_title", Kind: providercontract.FieldKindText},
			{Key: "party.primary_email", Expr: "p.primary_email", Kind: providercontract.FieldKindText},
			{Key: "party.timezone_name", Expr: "p.timezone_name", Kind: providercontract.FieldKindText},
			{Key: "party.external_ref", Expr: "p.external_ref", Kind: providercontract.FieldKindText},
			{Key: "party.notes", Expr: "p.notes", Kind: providercontract.FieldKindText},
			{Key: "party.updated_at", Expr: "p.updated_at", Kind: providercontract.FieldKindTimestamp},
		},
	}}
}
