package evidence

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func exportIncidentBundleFiles(ctx context.Context, input sourceport.ExportContext) error {
	for _, spec := range []sourceport.AttributedRows{
		{Path: "data/evidence_records.ndjson", Table: "evidence", Identity: "record_id", Query: `SELECT to_jsonb(t) FROM evidence t WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT 128`},
		{Path: "data/evidence_custody_events.ndjson", Table: "evidence_custody_events", Identity: "custody_event_id", ActorColumns: []string{"actor_user_id"}, Query: `SELECT to_jsonb(t) FROM evidence_custody_events t WHERE incident_id=$1 AND ($2::uuid IS NULL OR custody_event_id>$2) ORDER BY custody_event_id LIMIT 128`},
		{Path: "data/object_blobs.ndjson", Table: "object_blobs", Identity: "object_blob_id", ActorColumns: []string{"created_by_user_id"}, Query: `SELECT to_jsonb(t) FROM object_blobs t WHERE incident_id=$1 AND ($2::uuid IS NULL OR object_blob_id>$2) ORDER BY object_blob_id LIMIT 128`},
	} {
		if err := sourceport.ExportAttributedRows(ctx, input, spec); err != nil {
			return err
		}
	}
	return nil
}

func importIncidentBundleFilesTx(ctx context.Context, tx pgx.Tx, files incidentportability.MemberReader, rewritten incidentportability.MemberReader, actorUserID uuid.UUID, attributions incidentportability.AttributionRecorder) error {
	specs := []incidentportability.FixedImportSpec{
		{LogicalBundlePath: "data/object_blobs.ndjson", AttributionTable: "object_blobs", StableIdentity: []string{"object_blob_id"}, RequiredColumns: []string{"object_blob_id", "incident_id"}, InsertSQL: `INSERT INTO object_blobs SELECT * FROM jsonb_populate_record(NULL::object_blobs, $1::jsonb)`},
		{LogicalBundlePath: "data/evidence_records.ndjson", AttributionTable: "evidence", StableIdentity: []string{"record_id"}, RequiredColumns: []string{"record_id", "incident_id"}, InsertSQL: `INSERT INTO evidence SELECT * FROM jsonb_populate_record(NULL::evidence, $1::jsonb)`},
		{LogicalBundlePath: "data/evidence_custody_events.ndjson", AttributionTable: "evidence_custody_events", StableIdentity: []string{"custody_event_id"}, RequiredColumns: []string{"custody_event_id", "evidence_record_id"}, InsertSQL: `INSERT INTO evidence_custody_events SELECT * FROM jsonb_populate_record(NULL::evidence_custody_events, $1::jsonb)`},
	}
	for _, spec := range specs {
		members := files
		if spec.LogicalBundlePath == "data/object_blobs.ndjson" {
			members = rewritten
		}
		if err := incidentportability.ImportFixedMemberNDJSON(ctx, tx, spec, members, actorUserID, attributions); err != nil {
			return err
		}
	}
	return nil
}
