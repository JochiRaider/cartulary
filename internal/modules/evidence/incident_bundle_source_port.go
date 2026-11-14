package evidence

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func incidentBundleSourceDescriptor() sourceport.Descriptor {
	return sourceport.Descriptor{
		FamilyID: "evidence", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.evidence", OwnerRelationIDs: []string{"evidence-source-and-handles"},
		Dependencies: []string{"tasks_decisions"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/evidence_records.ndjson", ContentRole: "source_rows", Versions: []int{8}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "evidence.source_identity_admitted"},
			{LogicalPath: "data/evidence_custody_events.ndjson", ContentRole: "source_rows", Versions: []int{8}, StableIdentity: []string{"custody_event_id"}, StableIdentityInvariantID: "evidence.source_identity_admitted"},
			{LogicalPath: "data/object_blobs.ndjson", ContentRole: "source_rows", Versions: []int{8}, StableIdentity: []string{"object_blob_id"}, StableIdentityInvariantID: "evidence.source_identity_admitted"},
		},
		InvariantIDs: []string{
			"evidence.envelope_type_scope", "evidence.object_metadata_agree",
			"evidence.storage_reference_legal", "evidence.byte_size_digest_agree",
			"evidence.lifecycle_legal", "evidence.staged_bytes_digest",
			"evidence.custody_ordered", "evidence.custody_same_incident",
			"evidence.source_identity_admitted",
			"evidence.row_shape_exact", "evidence.attribution_exact", "evidence.source_state_exact",
			"evidence.references_same_incident",
		},
	}
}

func NewIncidentBundleSourcePort() sourceport.Port {
	binding := &evidencePreparationBinding{contractMajor: sourceport.ContractMajor}
	descriptor := incidentBundleSourceDescriptor()
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: exportIncidentBundleFiles,
		Prepare: func(ctx context.Context, bundle sourceport.Bundle, input sourceport.ImportContext) (any, error) {
			return prepareEvidenceImport(ctx, bundle, input, descriptor, binding)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, input sourceport.ImportContext) error {
			prepared, ok := value.(preparedEvidenceImport)
			if !ok || !prepared.matches(input, binding) || tx == nil {
				return sourceport.ErrPreparedBinding
			}
			return prepared.apply(ctx, tx, input, descriptor)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedEvidenceImport)
			if !ok || !prepared.matches(importContext, binding) || tx == nil {
				return sourceport.ErrPreparedBinding
			}
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
			if err := prepared.compareRecords(ctx, tx, descriptor); err != nil {
				return err
			}
			if err := prepared.validateRecordAttributions(ctx, importContext.Attributions, descriptor); err != nil {
				return err
			}
			if err := prepared.validatePartyReferences(ctx, tx, descriptor); err != nil {
				return err
			}
			if err := prepared.compareBlobs(ctx, tx, importContext, descriptor); err != nil {
				return err
			}
			if err := prepared.validateBlobAttributions(ctx, importContext.Attributions, descriptor); err != nil {
				return err
			}
			if err := prepared.typed(ctx, descriptor, func(row sourcevalue.Row) error {
				return prepared.validateBlobAssociation(ctx, tx, row, descriptor)
			}); err != nil {
				return err
			}
			if err := prepared.compareCustody(ctx, tx, descriptor); err != nil {
				return err
			}
			return prepared.validateCustodyAttributions(ctx, importContext.Attributions, descriptor)
		},
	})
}
