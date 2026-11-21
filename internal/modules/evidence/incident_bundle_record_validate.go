package evidence

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/jackc/pgx/v5"
	"reflect"
)

func (p preparedEvidenceImport) compareRecords(ctx context.Context, tx pgx.Tx, d sourceport.Descriptor) error {
	rows, err := tx.Query(ctx, `SELECT record_id,incident_id,title,lifecycle_state,requested_at,received_at,storage_ref,blob_hash,collector_party_text,collector_party_id,source_party_text,source_party_id,upload_state,object_blob_id,owner_user_id,created_at,updated_at FROM evidence WHERE incident_id=$1 ORDER BY record_id`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = p.typed(ctx, d, func(expected sourcevalue.Row) error {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return d.DeclaredFailure("evidence.source_state_exact")
		}
		var row sourcevalue.Row
		if err := rows.Scan(&row.RecordID, &row.IncidentID, &row.Title, &row.LifecycleState, &row.RequestedAt, &row.ReceivedAt, &row.StorageRef, &row.BlobHash, &row.CollectorPartyText, &row.CollectorPartyID, &row.SourcePartyText, &row.SourcePartyID, &row.UploadState, &row.ObjectBlobID, &row.OwnerUserID, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return err
		}
		row.UTCClocks()
		if !reflect.DeepEqual(row, expected.WithLocalActor(p.actorID)) {
			return d.DeclaredFailure("evidence.source_state_exact")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return d.DeclaredFailure("evidence.source_state_exact")
	}
	return rows.Err()
}
