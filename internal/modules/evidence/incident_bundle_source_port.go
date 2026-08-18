package evidence

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func NewIncidentBundleSourcePort() sourceport.Port {
	descriptor := sourceport.Descriptor{
		FamilyID: "evidence", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.evidence", OwnerRelationIDs: []string{"evidence-source-and-handles"},
		Dependencies: []string{"tasks_decisions"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/evidence_records.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "evidence.source_identity_admitted"},
			{LogicalPath: "data/evidence_custody_events.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"custody_event_id"}, StableIdentityInvariantID: "evidence.source_identity_admitted"},
			{LogicalPath: "data/object_blobs.ndjson", ContentRole: "source_rows", Versions: []int{1, 2}, StableIdentity: []string{"object_blob_id"}, StableIdentityInvariantID: "evidence.source_identity_admitted"},
		},
		InvariantIDs: []string{
			"evidence.envelope_type_scope", "evidence.object_metadata_agree",
			"evidence.storage_reference_legal", "evidence.byte_size_digest_agree",
			"evidence.lifecycle_legal", "evidence.staged_bytes_digest",
			"evidence.custody_ordered", "evidence.custody_same_incident",
			"evidence.source_identity_admitted",
		},
	}
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: sourceport.QueryExport(ExportIncidentBundleFiles),
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return sourceport.PrepareFiles(descriptor, bundle, importContext.BundleVersion)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			files := map[string][]byte(value.(sourceport.PreparedFiles))
			files["data/object_blobs.ndjson"] = importContext.RewrittenObjectBlobs
			return ImportIncidentBundleFilesTx(ctx, tx, files, importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, _ any, importContext sourceport.ImportContext) error {
			var invalid bool
			if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1 FROM evidence evidence_row
    LEFT JOIN records record ON record.record_id = evidence_row.record_id
    WHERE evidence_row.incident_id = $1
      AND (record.record_id IS NULL OR record.incident_id <> $1 OR record.record_type <> 'evidence')
)`, importContext.IncidentID).Scan(&invalid); err != nil {
				return err
			}
			if invalid {
				return descriptor.DeclaredFailure("evidence.envelope_type_scope")
			}
			return nil
		},
	})
}
