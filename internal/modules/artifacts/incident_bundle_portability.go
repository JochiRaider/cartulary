package artifacts

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

func exportIncidentBundleFiles(ctx context.Context, input sourceport.ExportContext) error {
	manifest, err := loadSourceStateManifest()
	if err != nil {
		return err
	}
	actors := map[string][]string{
		"artifacts":         {"outgoing_owner_user_id", "incoming_owner_user_id", "review_owner_user_id", "owner_user_id", "created_by_user_id"},
		"artifact_findings": {"owner_user_id"}, "artifact_investigative_queries": {"created_by_user_id"}, "handoff_risk_refs": {"created_by_user_id", "deleted_by_user_id"},
	}
	for _, relation := range manifest.relations {
		if len(relation.stableIdentity) != 1 {
			return errInvalidSourceStateManifest
		}
		identity := relation.stableIdentity[0]
		query, err := sourceExportQuery(relation.tableName, identity)
		if err != nil {
			return err
		}
		if err := sourceport.ExportAttributedRows(ctx, input, sourceport.AttributedRows{Path: relation.logicalBundlePath, Table: relation.tableName, Identity: identity, ActorColumns: actors[relation.tableName], Query: query}); err != nil {
			return err
		}
	}
	return nil
}

func importIncidentBundleFilesTx(ctx context.Context, tx pgx.Tx, files incidentportability.MemberReader, actorUserID uuid.UUID, attributions incidentportability.AttributionRecorder) error {
	manifest, err := loadSourceStateManifest()
	if err != nil {
		return err
	}
	specs := manifest.importSpecifications()
	for _, spec := range specs {
		if err := incidentportability.ImportFixedMemberNDJSON(ctx, tx, spec, files, actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}

func sourceExportQuery(table, identity string) (sqlshape.Fragment, error) {
	column := sqlshape.Identifier(identity)
	switch table {
	case "artifacts":
		return sqlshape.Format(`SELECT to_jsonb(t) FROM "artifacts" t WHERE incident_id=$1 AND ($2::uuid IS NULL OR %s>$2) ORDER BY %s LIMIT 128`, column, column), nil
	case "artifact_findings":
		return sqlshape.Format(`SELECT to_jsonb(t) FROM "artifact_findings" t WHERE incident_id=$1 AND ($2::uuid IS NULL OR %s>$2) ORDER BY %s LIMIT 128`, column, column), nil
	case "artifact_investigative_queries":
		return sqlshape.Format(`SELECT to_jsonb(t) FROM "artifact_investigative_queries" t WHERE incident_id=$1 AND ($2::uuid IS NULL OR %s>$2) ORDER BY %s LIMIT 128`, column, column), nil
	case "artifact_forensic_keywords":
		return sqlshape.Format(`SELECT to_jsonb(t) FROM "artifact_forensic_keywords" t WHERE incident_id=$1 AND ($2::uuid IS NULL OR %s>$2) ORDER BY %s LIMIT 128`, column, column), nil
	case "handoff_risk_refs":
		return sqlshape.Format(`SELECT to_jsonb(t) FROM "handoff_risk_refs" t WHERE incident_id=$1 AND ($2::uuid IS NULL OR %s>$2) ORDER BY %s LIMIT 128`, column, column), nil
	default:
		return sqlshape.Fragment{}, errInvalidSourceStateManifest
	}
}
