package deleterestore

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

type source struct{}

var _ deleterestorecontract.DeleteRestoreSource = source{}

func NewSource() deleterestorecontract.DeleteRestoreSource {
	return source{}
}

func (s source) SnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	snapshots, err := s.SnapshotsTx(ctx, tx, []uuid.UUID{recordID})
	return deleterestorecontract.SnapshotOf(snapshots, recordID, err)
}

func (source) SnapshotsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]map[string]any, error) {
	return deleterestorecontract.CollectSnapshots(tx.Query(ctx, `
SELECT r.record_id, jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(e))
  FROM records r
  JOIN evidence e
    ON e.record_id = r.record_id
 WHERE r.record_id = ANY($1)
`, recordIDs))
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
	// REQ-02-190: the prospective live state links a same-incident finalized
	// Blob, verified and locked non-waitingly through commit, and satisfies the
	// bridge. A tuple or lifecycle-vocabulary violation fails as internal_error.
	admission, err := evidenceblob.AdmitLinkTx(ctx, tx, evidenceblob.LinkRequest{
		EvidenceID: request.RecordID, IncidentID: request.IncidentID, Lifecycle: lifecycle, BlobID: blob, Mode: evidenceblob.ShareNoWait,
	})
	if err != nil {
		return deleterestorecontract.StateTransitionPreparation{}, restoreDependencyError(err, request.RecordID)
	}
	switch admission.Outcome {
	case evidenceblob.Admitted:
		return deleterestorecontract.StateTransitionPreparation{}, nil
	case evidenceblob.Contended:
		return deleterestorecontract.StateTransitionPreparation{}, &revisions.RecordLockedError{RecordID: request.RecordID}
	default:
		return deleterestorecontract.StateTransitionPreparation{Blocked: &deleterestorecontract.StateTransitionBlock{ReasonCode: "evidence_blob_state_conflict"}}, nil
	}
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
