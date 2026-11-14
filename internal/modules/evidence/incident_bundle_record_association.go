package evidence

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p preparedEvidenceImport) validateBlobAssociation(ctx context.Context, tx pgx.Tx, row sourcevalue.Row, d sourceport.Descriptor) error {
	if row.ObjectBlobID == nil {
		return nil
	}
	var incident uuid.UUID
	var state string
	err := tx.QueryRow(ctx, `SELECT incident_id,upload_state FROM object_blobs WHERE object_blob_id=$1 FOR UPDATE`, row.ObjectBlobID).Scan(&incident, &state)
	if errors.Is(err, pgx.ErrNoRows) {
		return d.DeclaredFailure("evidence.object_metadata_agree")
	}
	if err != nil {
		return err
	}
	if incident != p.incidentID {
		return d.DeclaredFailure("evidence.object_metadata_agree")
	}
	var associated bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM evidence WHERE object_blob_id=$1 AND record_id<>$2)`, row.ObjectBlobID, row.RecordID).Scan(&associated); err != nil {
		return err
	}
	if associated {
		return d.DeclaredFailure("evidence.object_metadata_agree")
	}
	var deleted bool
	if err := tx.QueryRow(ctx, `SELECT deleted_at IS NOT NULL FROM records WHERE record_id=$1 AND incident_id=$2`, row.RecordID, p.incidentID).Scan(&deleted); err != nil {
		return err
	}
	if (state != policy.BlobAvailable && state != policy.BlobQuarantined) || (!deleted && policy.ViolatesEvidenceBlobBridge(row.LifecycleState, true, state)) {
		return d.DeclaredFailure("evidence.lifecycle_legal")
	}
	return nil
}
