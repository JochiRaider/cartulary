package deleterestore

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

type source struct{}

var _ deleterestorecontract.DeleteRestoreSource = source{}

func NewSource() deleterestorecontract.DeleteRestoreSource {
	return source{}
}

func (source) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	return deleterestorecontract.ScanSnapshot(tx.QueryRow(ctx, `
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(e))
  FROM records r
  JOIN evidence e
    ON e.record_id = r.record_id
 WHERE r.record_id = $1
`, recordID))
}

func (source) UpdateSourceDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) error {
	return nil
}

func (source) ViewSchemaID(context.Context, pgx.Tx, uuid.UUID) (string, error) {
	return "cartulary.view.evidence.v1", nil
}

func (source) PrepareStateTransitionTx(context.Context, pgx.Tx, deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

func (source) ValidateStateTransitionTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest) (deleterestorecontract.StateTransitionPreparation, error) {
	if request.Kind != deleterestorecontract.StateTransitionRestore {
		return deleterestorecontract.StateTransitionPreparation{}, nil
	}
	var collector, origin, blob *uuid.UUID
	var lifecycle string
	if err := tx.QueryRow(ctx, `SELECT collector_party_id,source_party_id,object_blob_id,lifecycle_state FROM evidence WHERE record_id=$1 AND incident_id=$2`, request.RecordID, request.IncidentID).Scan(&collector, &origin, &blob, &lifecycle); err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, err
	}
	partyResult, err := validateRestorePartiesTx(ctx, tx, request, collector, origin)
	if err != nil || partyResult.Blocked != nil {
		return partyResult, err
	}
	blocked := deleterestorecontract.StateTransitionPreparation{Blocked: &deleterestorecontract.StateTransitionBlock{ReasonCode: "evidence_blob_state_conflict"}}
	blobState := ""
	if blob != nil {
		var incident uuid.UUID
		err := tx.QueryRow(ctx, `SELECT incident_id,upload_state FROM object_blobs WHERE object_blob_id=$1 FOR UPDATE NOWAIT`, *blob).Scan(&incident, &blobState)
		if errors.Is(err, pgx.ErrNoRows) {
			return blocked, nil
		}
		if err != nil {
			return deleterestorecontract.StateTransitionPreparation{}, restoreDependencyError(err, request.RecordID)
		}
		if incident != request.IncidentID || (blobState != policy.BlobAvailable && blobState != policy.BlobQuarantined) {
			return blocked, nil
		}
	}
	if policy.ViolatesEvidenceBlobBridge(lifecycle, blob != nil, blobState) {
		return blocked, nil
	}
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

func validateRestorePartiesTx(ctx context.Context, tx pgx.Tx, request deleterestorecontract.StateTransitionRequest, collector, origin *uuid.UUID) (deleterestorecontract.StateTransitionPreparation, error) {
	partyIDs := make([]uuid.UUID, 0, 2)
	if collector != nil {
		partyIDs = append(partyIDs, *collector)
	}
	if origin != nil && (collector == nil || *origin != *collector) {
		partyIDs = append(partyIDs, *origin)
	}
	if len(partyIDs) == 0 {
		return deleterestorecontract.StateTransitionPreparation{}, nil
	}
	rows, err := tx.Query(ctx, `
SELECT r.record_id
  FROM records r
  JOIN parties p ON p.record_id=r.record_id AND p.incident_id=r.incident_id
 WHERE r.record_id=ANY($1::uuid[]) AND r.incident_id=$2
   AND r.record_type='party' AND r.deleted_at IS NULL
 ORDER BY r.record_id
 FOR SHARE OF r NOWAIT
`, partyIDs, request.IncidentID)
	if err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, restoreDependencyError(err, request.RecordID)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, restoreDependencyError(err, request.RecordID)
	}
	if count != len(partyIDs) {
		return deleterestorecontract.StateTransitionPreparation{Blocked: &deleterestorecontract.StateTransitionBlock{
			ReasonCode: "inactive_evidence_party_reference",
		}}, nil
	}
	return deleterestorecontract.StateTransitionPreparation{}, nil
}

func restoreDependencyError(err error, recordID uuid.UUID) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "55P03" {
		return &revisions.RecordLockedError{RecordID: recordID}
	}
	return err
}
