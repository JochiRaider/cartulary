package evidence

import (
	"context"
	"errors"

	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/jackc/pgx/v5"
)

func (p preparedEvidenceImport) validateBlobAssociation(ctx context.Context, tx pgx.Tx, row sourcevalue.Row, d sourceport.Descriptor) error {
	if row.ObjectBlobID == nil {
		return nil
	}
	// The association locks and verifies the Blob through the single Blob
	// authority; a tuple violation is an invariant failure, not a declared one.
	locked, err := evidenceblob.LockTx(ctx, tx, *row.ObjectBlobID, evidenceblob.UpdateWait)
	if errors.Is(err, evidenceblob.ErrNotFound) {
		return d.DeclaredFailure("evidence.object_metadata_agree")
	}
	if err != nil {
		return err
	}
	state := locked.UploadState
	if locked.IncidentID != p.incidentID {
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
