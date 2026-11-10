package queryengine

import "github.com/JochiRaider/cartulary/internal/platform/sqlshape"

const partyViewSchemaID = "cartulary.view.parties.v1"

func PartyPlans() []Surface {
	return []Surface{{
		ViewSchemaID: partyViewSchemaID,
		FromSQL:      sqlshape.Literal("FROM party_grid_projection p JOIN records r ON r.record_id = p.record_id"),
		RecordExpr:   sqlshape.Literal("p.record_id"),
		IncidentExpr: sqlshape.Literal("p.incident_id"),
		Fields: []Field{
			{Key: "party.display_name", Expr: sqlshape.Literal("p.display_name"), Kind: FieldKindText},
			{Key: "party.party_kind", Expr: sqlshape.Literal("p.party_kind"), Kind: FieldKindText},
			{Key: "party.organization_name", Expr: sqlshape.Literal("p.organization_name"), Kind: FieldKindText},
			{Key: "party.role_title", Expr: sqlshape.Literal("p.role_title"), Kind: FieldKindText},
			{Key: "party.primary_email", Expr: sqlshape.Literal("p.primary_email"), Kind: FieldKindText},
			{Key: "party.timezone_name", Expr: sqlshape.Literal("p.timezone_name"), Kind: FieldKindText},
			{Key: "party.external_ref", Expr: sqlshape.Literal("p.external_ref"), Kind: FieldKindText},
			{Key: "party.notes", Expr: sqlshape.Literal("p.notes"), Kind: FieldKindText},
			{Key: "party.updated_at", Expr: sqlshape.Literal("p.updated_at"), Kind: FieldKindTimestamp},
		},
	}}
}
