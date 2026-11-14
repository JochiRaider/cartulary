package evidence

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

func exportIncidentBundleFiles(ctx context.Context, input sourceport.ExportContext) error {
	for _, spec := range []sourceport.AttributedRows{
		{Path: "data/evidence_records.ndjson", Table: "evidence", Identity: "record_id", ActorColumns: []string{"owner_user_id"}, Query: sqlshape.Literal(`SELECT to_jsonb(t) FROM evidence t WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT 128`)},
		{Path: "data/evidence_custody_events.ndjson", Table: "evidence_custody_events", Identity: "custody_event_id", ActorColumns: []string{"actor_user_id"}, Query: sqlshape.Literal(`SELECT jsonb_build_object('custody_event_id',custody_event_id,'incident_id',incident_id,'evidence_record_id',evidence_record_id,'custody_event_type',custody_event_type,'actor_user_id',actor_user_id,'occurred_at',occurred_at,'location_text',location_text,'note',note,'metadata',metadata) FROM evidence_custody_events WHERE incident_id=$1 AND ($2::uuid IS NULL OR custody_event_id>$2) ORDER BY custody_event_id LIMIT 128`)},
		{Path: "data/object_blobs.ndjson", Table: "object_blobs", Identity: "object_blob_id", ActorColumns: []string{"created_by_user_id"}, Query: sqlshape.Literal(`SELECT jsonb_build_object(
 'object_blob_id',object_blob_id,'incident_id',incident_id,'created_by_user_id',created_by_user_id,
 'upload_state',upload_state,'byte_size',byte_size,'filename_hint',filename_hint,'content_type_hint',content_type_hint,
 'expected_sha256_hex',expected_sha256_hex,'observed_size',observed_size,'observed_content_type',observed_content_type,
 'observed_sha256_hex',observed_sha256_hex,'target_expires_at',target_expires_at,'pending_expires_at',pending_expires_at,
 'finalized_at',finalized_at,'terminal_reason',terminal_reason,'failed_at',failed_at,'cleanup_due_at',cleanup_due_at,
 'cleaned_up_at',cleaned_up_at,'created_at',created_at,'updated_at',updated_at)
 FROM object_blobs WHERE incident_id=$1 AND ($2::uuid IS NULL OR object_blob_id>$2) ORDER BY object_blob_id LIMIT 128`)},
	} {
		if err := sourceport.ExportAttributedRows(ctx, input, spec); err != nil {
			return err
		}
	}
	return nil
}

func (p preparedEvidenceImport) apply(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, descriptor sourceport.Descriptor) error {
	if err := p.bindBlobRewrite(ctx, input, descriptor); err != nil {
		return err
	}
	if err := p.preflight(ctx, tx, descriptor); err != nil {
		return err
	}
	if err := p.insertBlobs(ctx, tx, input, descriptor); err != nil {
		return err
	}
	if err := p.insertRecords(ctx, tx, input, descriptor); err != nil {
		return err
	}
	return p.insertCustody(ctx, tx, input, descriptor)
}
