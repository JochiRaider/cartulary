package incidentbundle

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

func exportIncidentBundleFiles(ctx context.Context, input sourceport.ExportContext) error {
	return sourceport.ExportAttributedRows(ctx, input, sourceport.AttributedRows{Path: assessmentsBundlePath, Table: "assessments", Identity: "record_id", ActorColumns: []string{"assessor_user_id", "deleted_by_user_id"}, Query: sqlshape.Literal(`SELECT jsonb_build_object(
'record_id',t.record_id,'incident_id',t.incident_id,'subject_record_id',t.subject_record_id,
'subject_type',t.subject_type,'assessment_state',t.assessment_state,'confidence_score',t.confidence_score,
'assessor_user_id',t.assessor_user_id,'assessed_at',t.assessed_at,'created_at',t.created_at,
'updated_at',t.updated_at,'deleted_at',t.deleted_at,'deleted_by_user_id',t.deleted_by_user_id,'rationale',t.rationale
) FROM assessments t WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT 128`)})
}
