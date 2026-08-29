package evidence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func (s *blobLifecycleService) QuarantineBlob(ctx context.Context, actorUserID uuid.UUID, objectBlobID uuid.UUID, trigger string, requestID string, now time.Time) (quarantineBlobResult, error) {
	if !evidencepolicy.ValidQuarantineEntryTrigger(trigger) {
		return quarantineBlobResult{}, errIllegalBlobTransition
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return quarantineBlobResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	blob, err := loadBlobForUpdateTx(ctx, tx, objectBlobID)
	if err != nil {
		return quarantineBlobResult{}, err
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, blob.IncidentID); err != nil {
		return quarantineBlobResult{}, err
	}
	if !evidencepolicy.LegalBlobTransition(blob.UploadState, evidencepolicy.BlobQuarantined, trigger) {
		return quarantineBlobResult{}, errIllegalBlobTransition
	}

	rows, err := tx.Query(ctx, `
SELECT e.record_id
  FROM evidence e
  JOIN records r ON r.record_id = e.record_id
 WHERE e.object_blob_id = $1
   AND e.lifecycle_state IN ('available', 'released')
   AND r.deleted_at IS NULL
 ORDER BY e.record_id
 FOR UPDATE OF e, r
`, objectBlobID)
	if err != nil {
		return quarantineBlobResult{}, err
	}
	recordIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var recordID uuid.UUID
		if err := rows.Scan(&recordID); err != nil {
			rows.Close()
			return quarantineBlobResult{}, err
		}
		recordIDs = append(recordIDs, recordID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return quarantineBlobResult{}, err
	}
	rows.Close()

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
