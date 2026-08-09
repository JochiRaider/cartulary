package projectionprovider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
)

type Source struct{}

func NewSource() *Source { return &Source{} }

func (*Source) LoadProjectionInputTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
) (evidenceprojection.ProjectionInput, bool, error) {
	input, err := scanProjectionInput(tx.QueryRow(ctx, evidenceProjectionSourceSQL+`
 WHERE e.record_id = $1
`, recordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return evidenceprojection.ProjectionInput{}, false, nil
	}
	if err != nil {
		return evidenceprojection.ProjectionInput{}, false, fmt.Errorf("load Evidence projection input: %w", err)
	}
	return input, true, nil
}

func (*Source) ListProjectionInputsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	afterRecordID *uuid.UUID,
	limit int,
) (evidenceprojection.ProjectionInputPage, error) {
	if limit < 1 || limit > 1000 {
		return evidenceprojection.ProjectionInputPage{}, fmt.Errorf("evidence projection source page limit %d is outside 1..1000", limit)
	}
	rows, err := tx.Query(ctx, evidenceProjectionSourceSQL+`
 WHERE e.incident_id = $1
   AND ($2::uuid IS NULL OR e.record_id > $2)
 ORDER BY e.record_id
 LIMIT $3
`, incidentID, afterRecordID, limit+1)
	if err != nil {
		return evidenceprojection.ProjectionInputPage{}, fmt.Errorf("list evidence projection inputs: %w", err)
	}
	defer rows.Close()

	inputs := make([]evidenceprojection.ProjectionInput, 0, limit+1)
	for rows.Next() {
		input, scanErr := scanProjectionInput(rows)
		if scanErr != nil {
			return evidenceprojection.ProjectionInputPage{}, fmt.Errorf("scan Evidence projection input: %w", scanErr)
		}
		inputs = append(inputs, input)
	}
	if err := rows.Err(); err != nil {
		return evidenceprojection.ProjectionInputPage{}, fmt.Errorf("iterate Evidence projection inputs: %w", err)
	}
	page := evidenceprojection.ProjectionInputPage{Inputs: inputs}
	if len(inputs) > limit {
		page.Inputs = inputs[:limit]
		next := page.Inputs[len(page.Inputs)-1].RecordID
		page.NextRecordID = &next
	}
	return page, nil
}

func scanProjectionInput(scanner interface{ Scan(...any) error }) (evidenceprojection.ProjectionInput, error) {
	var (
		input              evidenceprojection.ProjectionInput
		title              pgtype.Text
		requestedAt        pgtype.Timestamptz
		receivedAt         pgtype.Timestamptz
		storageRef         pgtype.Text
		blobHash           pgtype.Text
		collectorPartyText pgtype.Text
		collectorPartyID   pgtype.UUID
		sourcePartyText    pgtype.Text
		sourcePartyID      pgtype.UUID
		linkedRecordCount  int32
	)
	if err := scanner.Scan(
		&input.RecordID, &input.IncidentID, &input.RowVersion, &title,
		&input.LifecycleState, &requestedAt, &receivedAt, &storageRef,
		&blobHash, &collectorPartyText, &collectorPartyID, &sourcePartyText,
		&sourcePartyID, &input.UploadState, &linkedRecordCount, &input.EditedAt,
	); err != nil {
		return evidenceprojection.ProjectionInput{}, err
	}
	input.Title = evidenceTextPointer(title)
	input.RequestedAt = evidenceTimePointer(requestedAt)
	input.ReceivedAt = evidenceTimePointer(receivedAt)
	input.StorageRef = evidenceTextPointer(storageRef)
	input.BlobHash = evidenceTextPointer(blobHash)
	input.CollectorPartyText = evidenceTextPointer(collectorPartyText)
	input.CollectorPartyID = evidenceUUIDPointer(collectorPartyID)
	input.SourcePartyText = evidenceTextPointer(sourcePartyText)
	input.SourcePartyID = evidenceUUIDPointer(sourcePartyID)
	input.LinkedRecordCount = int(linkedRecordCount)
	input.EditedAt = input.EditedAt.UTC()
	return input, nil
}

func evidenceTextPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func evidenceTimePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time.UTC()
	return &result
}

func evidenceUUIDPointer(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	result := uuid.UUID(value.Bytes)
	return &result
}

const evidenceProjectionSourceSQL = `
SELECT
    e.record_id,
    e.incident_id,
    r.row_version,
    e.title,
    e.lifecycle_state,
    e.requested_at,
    e.received_at,
    e.storage_ref,
    COALESCE(b.observed_sha256_hex, e.blob_hash),
    e.collector_party_text,
    e.collector_party_id,
    e.source_party_text,
    e.source_party_id,
    COALESCE(b.upload_state, e.upload_state),
    (
        SELECT COUNT(*)::integer
          FROM active_record_links_v1 rl
         WHERE rl.src_record_id = e.record_id OR rl.dst_record_id = e.record_id
    ),
    e.updated_at
  FROM evidence e
  JOIN records r
    ON r.incident_id = e.incident_id
   AND r.record_id = e.record_id
   AND r.deleted_at IS NULL
  LEFT JOIN object_blobs b
    ON b.object_blob_id = e.object_blob_id`
