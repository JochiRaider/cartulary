package artifacts

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
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
		query := fmt.Sprintf("SELECT to_jsonb(t) FROM %s t WHERE incident_id=$1 AND ($2::uuid IS NULL OR %s>$2) ORDER BY %s LIMIT 128", quoteSourceStateIdentifier(relation.tableName), quoteSourceStateIdentifier(identity), quoteSourceStateIdentifier(identity))
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
