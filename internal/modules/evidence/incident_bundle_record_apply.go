package evidence

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/jackc/pgx/v5"
)

func (p preparedEvidenceImport) preflight(ctx context.Context, tx pgx.Tx, d sourceport.Descriptor) error {
	if err := p.typed(ctx, d, func(row sourcevalue.Row) error {
		var valid bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM records WHERE record_id=$1 AND incident_id=$2 AND record_type='evidence')`, row.RecordID, row.IncidentID).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return d.DeclaredFailure("evidence.envelope_type_scope")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := p.typed(ctx, d, func(row sourcevalue.Row) error {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM evidence WHERE record_id=$1)`, row.RecordID).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return d.DeclaredFailure("evidence.source_identity_admitted")
		}
		return nil
	}); err != nil {
		return err
	}
	return p.validatePartyReferences(ctx, tx, d)
}

func (p preparedEvidenceImport) validatePartyReferences(ctx context.Context, tx pgx.Tx, d sourceport.Descriptor) error {
	return p.typed(ctx, d, func(row sourcevalue.Row) error {
		if row.CollectorPartyID == nil && row.SourcePartyID == nil {
			return nil
		}
		var valid bool
		if err := tx.QueryRow(ctx, `SELECT NOT EXISTS (
 SELECT 1 FROM unnest(ARRAY[$3::uuid,$4::uuid]) AS ref(party_id)
 LEFT JOIN records party_record ON party_record.record_id=ref.party_id
 LEFT JOIN parties party_source ON party_source.record_id=ref.party_id
 WHERE ref.party_id IS NOT NULL AND (
  party_record.record_id IS NULL OR party_record.incident_id<>$2 OR party_record.record_type<>'party'
  OR party_source.record_id IS NULL OR party_source.incident_id<>$2
  OR (evidence_record.deleted_at IS NULL AND party_record.deleted_at IS NOT NULL)
 )
) FROM records evidence_record WHERE evidence_record.record_id=$1 AND evidence_record.incident_id=$2`,
			row.RecordID, row.IncidentID, row.CollectorPartyID, row.SourcePartyID).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return d.DeclaredFailure("evidence.references_same_incident")
		}
		return nil
	})
}
func (p preparedEvidenceImport) insertRecords(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, d sourceport.Descriptor) error {
	return p.typed(ctx, d, func(source sourcevalue.Row) error {
		row := source.WithLocalActor(p.actorID)
		if err := p.validateBlobAssociation(ctx, tx, row, d); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `INSERT INTO evidence(record_id,incident_id,title,lifecycle_state,requested_at,received_at,storage_ref,blob_hash,collector_party_text,collector_party_id,source_party_text,source_party_id,upload_state,object_blob_id,owner_user_id,created_at,updated_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
			row.RecordID, row.IncidentID, row.Title, row.LifecycleState, row.RequestedAt, row.ReceivedAt, row.StorageRef, row.BlobHash, row.CollectorPartyText, row.CollectorPartyID, row.SourcePartyText, row.SourcePartyID, row.UploadState, row.ObjectBlobID, row.OwnerUserID, row.CreatedAt, row.UpdatedAt)
		if err != nil {
			if isEvidenceBlobUniqueViolation(err) {
				return d.DeclaredFailure("evidence.object_metadata_agree")
			}
			return err
		}
		if tag.RowsAffected() != 1 {
			return d.DeclaredFailure("evidence.source_identity_admitted")
		}
		if source.OwnerUserID != nil {
			return input.Attributions.RecordImportedAttribution("evidence", source.RecordID.String(), "owner_user_id", source.OwnerUserID.String())
		}
		return nil
	})
}
