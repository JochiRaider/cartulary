package evidence

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

// quarantineAssociationAttempts bounds how often quarantine repeats the
// unlocked association lookup when a competing association committed between
// the lookup and the Blob lock (REQ-02-190a).
const quarantineAssociationAttempts = 3

// errQuarantineAssociationUnstable fails quarantine closed (internal_error)
// when the Blob's association kept changing across every bounded attempt.
var errQuarantineAssociationUnstable = errors.New("evidence: blob association changed during quarantine")

func (s *blobLifecycleService) QuarantineBlob(ctx context.Context, actorUserID uuid.UUID, objectBlobID uuid.UUID, trigger string, requestID string, now time.Time) (quarantineBlobResult, error) {
	if !evidencepolicy.ValidQuarantineEntryTrigger(trigger) {
		return quarantineBlobResult{}, errIllegalBlobTransition
	}
	for attempt := 0; attempt < quarantineAssociationAttempts; attempt++ {
		result, err := s.quarantineBlobAttempt(ctx, actorUserID, objectBlobID, trigger, requestID, now)
		if errors.Is(err, errQuarantineAssociationChanged) {
			continue
		}
		return result, err
	}
	return quarantineBlobResult{}, errQuarantineAssociationUnstable
}

// errQuarantineAssociationChanged reports that the association seen by the
// unlocked lookup changed before the Blob lock; the attempt commits nothing.
var errQuarantineAssociationChanged = errors.New("evidence: blob association changed before the blob lock")

// quarantineBlobAttempt takes the REQ-02-190a locks in order: the Evidence
// linked to the Blob (whatever its lifecycle or deletion state), then the
// incident (shared), then the Blob. Associations change only under the Blob
// lock, so a link that differs under it means a competing association
// committed first and this attempt rolls back without effect.
func (s *blobLifecycleService) quarantineBlobAttempt(ctx context.Context, actorUserID uuid.UUID, objectBlobID uuid.UUID, trigger string, requestID string, now time.Time) (quarantineBlobResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return quarantineBlobResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	incidentID, linked, err := quarantineAssociationTx(ctx, tx, objectBlobID)
	if err != nil {
		return quarantineBlobResult{}, err
	}
	rows, err := tx.Query(ctx, `
SELECT e.record_id, e.lifecycle_state, r.deleted_at IS NULL
  FROM evidence e
  JOIN records r ON r.record_id = e.record_id
 WHERE e.record_id = ANY($1::uuid[])
 ORDER BY e.record_id
 FOR UPDATE OF r, e
`, linked)
	if err != nil {
		return quarantineBlobResult{}, err
	}
	recordIDs := make([]uuid.UUID, 0, len(linked))
	for rows.Next() {
		var recordID uuid.UUID
		var lifecycle string
		var live bool
		if err := rows.Scan(&recordID, &lifecycle, &live); err != nil {
			rows.Close()
			return quarantineBlobResult{}, err
		}
		// Only live accessible Evidence transitions (REQ-02-190); deleted
		// Evidence keeps its retained lifecycle but is still locked first.
		if live && (lifecycle == evidencepolicy.EvidenceAvailable || lifecycle == evidencepolicy.EvidenceReleased) {
			recordIDs = append(recordIDs, recordID)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return quarantineBlobResult{}, err
	}
	rows.Close()
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, incidentID); err != nil {
		return quarantineBlobResult{}, err
	}
	blob, err := loadBlobForUpdateTx(ctx, tx, objectBlobID)
	if err != nil {
		return quarantineBlobResult{}, err
	}
	_, current, err := quarantineAssociationTx(ctx, tx, objectBlobID)
	if err != nil {
		return quarantineBlobResult{}, err
	}
	if blob.IncidentID != incidentID || !slices.Equal(current, linked) {
		return quarantineBlobResult{}, errQuarantineAssociationChanged
	}
	if !evidencepolicy.LegalBlobTransition(blob.UploadState, evidencepolicy.BlobQuarantined, trigger) {
		return quarantineBlobResult{}, &blobTransitionError{fromStatus: blob.UploadState, toStatus: evidencepolicy.BlobQuarantined}
	}

	beforeRows := make(map[uuid.UUID]map[string]any, len(recordIDs))
	beforeSnapshots := make(map[uuid.UUID]revisions.RecordSnapshot, len(recordIDs))
	beforeVersions := make(map[uuid.UUID]int64, len(recordIDs))
	for _, recordID := range recordIDs {
		row, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		beforeRows[recordID] = row
		snapshot, err := s.revisionStore.CaptureRecordSnapshotTx(ctx, tx, recordID)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		beforeSnapshots[recordID] = snapshot
		beforeVersions[recordID] = int64FromAny(row["row_version"])
	}

	_, err = tx.Exec(ctx, `
UPDATE object_blobs
   SET upload_state = 'quarantined',
       updated_at = $2
 WHERE object_blob_id = $1
   AND upload_state = 'available'
`, objectBlobID, now.UTC())
	if err != nil {
		return quarantineBlobResult{}, err
	}

	var changeSetID uuid.UUID
	changedRows := make([]attachRecordChange, 0, len(recordIDs))
	if len(recordIDs) > 0 {
		reason := trigger
		requestIDPtr := &requestID
		if requestID == "" {
			requestIDPtr = nil
		}
		changeSetID, err = s.revisionStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
			IncidentID: blob.IncidentID, ActorUserID: actorUserID, Source: "evidence.blob.quarantine",
			Reason: &reason, RequestID: requestIDPtr, CreatedAt: now.UTC(),
		})
		if err != nil {
			return quarantineBlobResult{}, err
		}
	}

	for idx, recordID := range recordIDs {
		_, err := tx.Exec(ctx, `
UPDATE evidence
   SET lifecycle_state = 'quarantined',
       upload_state = 'quarantined',
       updated_at = $2
 WHERE record_id = $1
`, recordID, now.UTC())
		if err != nil {
			return quarantineBlobResult{}, err
		}
		if err := insertEvidenceCustodyEventTx(ctx, tx, evidenceCustodyEventParams{
			IncidentID:       blob.IncidentID,
			EvidenceRecordID: recordID,
			CustodyEventType: "quarantined",
			ActorUserID:      &actorUserID,
			OccurredAt:       now.UTC(),
			Metadata:         map[string]any{"object_blob_id": objectBlobID.String(), "trigger": trigger},
		}); err != nil {
			return quarantineBlobResult{}, err
		}
		rowVersion, err := s.records.AdvanceVersionTx(ctx, tx, recordID, actorUserID, now)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		if err := s.projections.RefreshEvidenceTx(ctx, tx, recordID); err != nil {
			return quarantineBlobResult{}, err
		}
		afterRow, err := s.projections.LoadEvidenceTx(ctx, tx, recordID)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		afterSnapshot, err := s.revisionStore.CaptureRecordSnapshotTx(ctx, tx, recordID)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		beforeVersionID := fmt.Sprintf("%s:%d", recordID, beforeVersions[recordID])
		afterVersionID := fmt.Sprintf("%s:%d", recordID, rowVersion)
		beforeSnapshot := beforeSnapshots[recordID]
		if err := s.revisionStore.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
			ChangeSetID: changeSetID, SequenceNo: idx + 1, TargetKind: "record", RecordID: recordID,
			OperationKind: "patch", BeforeVersionID: &beforeVersionID, AfterVersionID: &afterVersionID,
			BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot,
		}); err != nil {
			return quarantineBlobResult{}, err
		}
		changedFieldKeys := sortedChangedKeys(beforeRows[recordID], afterRow)
		if err := s.revisionStore.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
			ChangeSetID: changeSetID, RecordID: recordID, RowVersion: rowVersion,
			BeforeSnapshot: &beforeSnapshot, AfterSnapshot: &afterSnapshot,
			ConflictFacts: evidenceRevisionFacts(beforeRows[recordID], afterRow, changedFieldKeys),
		}); err != nil {
			return quarantineBlobResult{}, err
		}
		projectionChanges, err := s.refreshEvidenceSupportProjectionsTx(ctx, tx, blob.IncidentID, recordID)
		if err != nil {
			return quarantineBlobResult{}, err
		}
		primaryChange := attachRecordChange{
			RecordID: recordID, RowVersion: rowVersion, ViewSchemaID: ViewSchemaID,
			ChangedFieldKeys: changedFieldKeys,
		}
		if err := appendEvidenceRecordChangeIntentsTx(
			ctx,
			tx,
			s.collaboration,
			blob.IncidentID,
			actorUserID,
			changeSetID.String(),
			changeSetID,
			primaryChange,
			afterRow,
			projectionChanges,
			now,
		); err != nil {
			return quarantineBlobResult{}, err
		}
		changedRows = append(changedRows, primaryChange)
		changedRows = append(changedRows, projectionChanges...)
	}

	if err := tx.Commit(ctx); err != nil {
		return quarantineBlobResult{}, err
	}
	return quarantineBlobResult{
		IncidentID: blob.IncidentID, ObjectBlobID: objectBlobID, ChangeSetID: changeSetID,
		ChangedEvidenceRows: changedRows, ChangedEvidenceRecord: len(recordIDs),
	}, nil
}

// quarantineAssociationTx reads, without locking, the Blob's incident and the
// Evidence records linked to it in ascending record_id.
func quarantineAssociationTx(ctx context.Context, tx pgx.Tx, objectBlobID uuid.UUID) (uuid.UUID, []uuid.UUID, error) {
	var incidentID uuid.UUID
	var linked []uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT b.incident_id,
       COALESCE((
           SELECT array_agg(e.record_id ORDER BY e.record_id)
             FROM evidence e
            WHERE e.object_blob_id = b.object_blob_id
       ), '{}'::uuid[])
  FROM object_blobs b
 WHERE b.object_blob_id = $1
`, objectBlobID).Scan(&incidentID, &linked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, nil, ErrBlobNotFound
		}
		return uuid.Nil, nil, err
	}
	return incidentID, linked, nil
}
