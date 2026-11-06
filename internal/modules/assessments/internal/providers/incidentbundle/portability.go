package incidentbundle

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func exportIncidentBundleFiles(ctx context.Context, input sourceport.ExportContext) error {
	return sourceport.ExportAttributedRows(ctx, input, sourceport.AttributedRows{Path: assessmentsBundlePath, Table: "assessments", Identity: "record_id", ActorColumns: []string{"assessor_user_id", "deleted_by_user_id"}, Query: `SELECT to_jsonb(t) FROM assessments t WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT 128`})
}

func importIncidentBundleFilesTx(
	ctx context.Context,
	tx pgx.Tx,
	files incidentportability.MemberReader,
	actorUserID uuid.UUID,
	attributions incidentportability.AttributionRecorder,
) error {
	return incidentportability.ImportFixedMemberNDJSON(ctx, tx, incidentportability.FixedImportSpec{
		LogicalBundlePath: assessmentsBundlePath,
		AttributionTable:  "assessments",
		StableIdentity:    []string{"record_id"},
		RequiredColumns:   []string{"record_id", "incident_id"},
		InsertSQL:         `INSERT INTO assessments SELECT * FROM jsonb_populate_record(NULL::assessments, $1::jsonb)`,
	}, files, actorUserID, attributions)
}
