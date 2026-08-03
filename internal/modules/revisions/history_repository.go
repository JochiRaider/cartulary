package revisions

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type mutationHistoryRow struct {
	ChangeSetID     uuid.UUID
	ActorUserID     uuid.UUID
	CommittedAt     time.Time
	Source          string
	SequenceNo      int
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeValue     []byte
	AfterValue      []byte
	RevisionNo      *int64
	HistoryEntryRef *string
}

type revisionHistoryRow struct {
	ChangeSetID uuid.UUID
	ActorUserID uuid.UUID
	CommittedAt time.Time
	Source      string
	RevisionNo  int64
	BeforeValue []byte
	AfterValue  []byte
}

type historyQueryRepository struct{}

func (historyQueryRepository) LoadMutationRowsTx(ctx context.Context, tx pgx.Tx, record RecordHistoryRecord) ([]mutationHistoryRow, error) {
	rows, err := tx.Query(ctx, `
SELECT cs.change_set_id,
       cs.actor_user_id,
       cs.created_at,
       cs.source,
       csm.sequence_no,
       csm.target_kind,
       csm.target_id,
       csm.operation_kind,
       csm.before_value,
       csm.after_value,
       rr.row_version,
       href.history_entry_ref
  FROM change_sets cs
  JOIN change_set_mutations csm
    ON csm.change_set_id = cs.change_set_id
  LEFT JOIN record_revisions rr
    ON rr.change_set_id = cs.change_set_id
   AND rr.record_id = $1
  LEFT JOIN record_history_entry_refs href
    ON href.record_id = $1
	   AND href.change_set_id = csm.change_set_id
	   AND href.mutation_sequence_no = csm.sequence_no
	 WHERE cs.incident_id = $2
	   AND (
	       csm.target_id = $3
	       OR (
	           csm.target_kind = 'record_link'
	           AND (
	               csm.before_value ->> 'src_record_id' = $3
	               OR csm.before_value ->> 'dst_record_id' = $3
	               OR csm.after_value ->> 'src_record_id' = $3
	               OR csm.after_value ->> 'src_record_id' = $3
	               OR csm.after_value ->> 'dst_record_id' = $3
	           )
	       )
	       OR (
	           csm.target_kind = 'entity_mention'
	           AND (
	               csm.before_value ->> 'source_record_id' = $3
	               OR csm.after_value ->> 'source_record_id' = $3
	           )
	       )
	       OR (
	           csm.target_kind = 'record_tag'
	           AND (
	               csm.before_value ->> 'record_id' = $3
	               OR csm.after_value ->> 'record_id' = $3
	           )
	       )
	       OR (
	           csm.target_kind = 'indicator_observation'
	           AND (
	               csm.before_value ->> 'source_record_id' = $3
	               OR csm.after_value ->> 'source_record_id' = $3
	               OR csm.before_value ->> 'resolved_indicator_record_id' = $3
	               OR csm.after_value ->> 'resolved_indicator_record_id' = $3
	           )
	       )
	       OR (
	           csm.target_kind = 'indicator_state_interval'
	           AND (
	               csm.before_value ->> 'indicator_record_id' = $3
	               OR csm.after_value ->> 'indicator_record_id' = $3
	           )
	       )
	   )
	 ORDER BY cs.created_at DESC, cs.change_set_id DESC, csm.sequence_no ASC
	`, record.RecordID, record.IncidentID, record.RecordID.String())
	if err != nil {
		return nil, fmt.Errorf("query record history mutations: %w", err)
	}
	defer rows.Close()

	result := make([]mutationHistoryRow, 0)
	for rows.Next() {
		var row mutationHistoryRow
		var revisionNo sql.NullInt64
		var historyEntryRef sql.NullString
		if err := rows.Scan(
			&row.ChangeSetID,
			&row.ActorUserID,
			&row.CommittedAt,
			&row.Source,
			&row.SequenceNo,
			&row.TargetKind,
			&row.TargetID,
			&row.OperationKind,
			&row.BeforeValue,
			&row.AfterValue,
			&revisionNo,
			&historyEntryRef,
		); err != nil {
			return nil, fmt.Errorf("scan record history mutation: %w", err)
		}
		if revisionNo.Valid {
			value := revisionNo.Int64
			row.RevisionNo = &value
		}
		if historyEntryRef.Valid {
			value := historyEntryRef.String
			row.HistoryEntryRef = &value
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate record history mutations: %w", err)
	}
	rows.Close()
	return result, nil
}

func (historyQueryRepository) LoadRevisionRowsTx(ctx context.Context, tx pgx.Tx, record RecordHistoryRecord) ([]revisionHistoryRow, error) {
	rows, err := tx.Query(ctx, `
SELECT cs.change_set_id,
       cs.actor_user_id,
       cs.created_at,
       cs.source,
       rr.row_version,
       rr.before_json,
       rr.after_json
  FROM record_revisions rr
  JOIN change_sets cs
    ON cs.change_set_id = rr.change_set_id
 WHERE rr.record_id = $1
   AND cs.incident_id = $2
 ORDER BY cs.created_at DESC, cs.change_set_id DESC, rr.row_version DESC
`, record.RecordID, record.IncidentID)
	if err != nil {
		return nil, fmt.Errorf("query record history revisions: %w", err)
	}
	defer rows.Close()

	result := make([]revisionHistoryRow, 0)
	for rows.Next() {
		var row revisionHistoryRow
		if err := rows.Scan(
			&row.ChangeSetID,
			&row.ActorUserID,
			&row.CommittedAt,
			&row.Source,
			&row.RevisionNo,
			&row.BeforeValue,
			&row.AfterValue,
		); err != nil {
			return nil, fmt.Errorf("scan record history revision: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate record history revisions: %w", err)
	}
	return result, nil
}

func (historyQueryRepository) EnsureHistoryEntryRefTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changeSetID uuid.UUID, sequenceNo int) (string, error) {
	var existing string
	err := tx.QueryRow(ctx, `
SELECT history_entry_ref
  FROM record_history_entry_refs
 WHERE record_id = $1
   AND change_set_id = $2
   AND mutation_sequence_no = $3
`, recordID, changeSetID, sequenceNo).Scan(&existing)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("lookup history entry ref: %w", err)
	}

	for attempts := 0; attempts < 3; attempts++ {
		candidate, err := generateHistoryEntryRef()
		if err != nil {
			return "", err
		}
		err = tx.QueryRow(ctx, `
INSERT INTO record_history_entry_refs (history_entry_ref, record_id, change_set_id, mutation_sequence_no)
VALUES ($1, $2, $3, $4)
ON CONFLICT (record_id, change_set_id, mutation_sequence_no) DO UPDATE
SET created_at = record_history_entry_refs.created_at
RETURNING history_entry_ref
`, candidate, recordID, changeSetID, sequenceNo).Scan(&existing)
		if err == nil {
			return existing, nil
		}
	}
	return "", fmt.Errorf("insert history entry ref after retries: %w", err)
}

func generateHistoryEntryRef() (string, error) {
	var payload [16]byte
	if _, err := rand.Read(payload[:]); err != nil {
		return "", fmt.Errorf("generate history entry ref: %w", err)
	}
	return "href_" + base64.RawURLEncoding.EncodeToString(payload[:]), nil
}
