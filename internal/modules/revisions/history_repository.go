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
	ChangeSetID             uuid.UUID
	ActorUserID             uuid.UUID
	CommittedAt             time.Time
	Source                  string
	SequenceNo              int
	TargetKind              string
	TargetID                string
	OperationKind           string
	BeforeValue             []byte
	AfterValue              []byte
	RevisionNo              *int64
	HistoryEntryRef         *string
	HistoryEntryAddressable bool
	CoalesceRevision        bool
	RevisionCount           int64
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

// Only ordering metadata is fetched for the limit+1 descriptor selection. Snapshot
// and projection work starts after the lookahead descriptor has been discarded.
const historyDescriptorsSQL = `
WITH events AS (
 SELECT cs.created_at, cs.change_set_id, 0 AS event_rank, csm.sequence_no, 0::bigint AS revision_no
 FROM change_sets cs JOIN change_set_mutations csm ON csm.change_set_id = cs.change_set_id
 WHERE cs.incident_id = $2 AND csm.history_record_ids @> ARRAY[$1]::uuid[]
 UNION ALL
 SELECT cs.created_at, cs.change_set_id, 1, 0, rr.row_version
 FROM record_revisions rr JOIN change_sets cs ON cs.change_set_id = rr.change_set_id
 WHERE rr.record_id = $1 AND cs.incident_id = $2
   AND NOT EXISTS (SELECT 1 FROM change_set_mutations csm
                   WHERE csm.change_set_id = rr.change_set_id
                     AND csm.history_record_ids @> ARRAY[$1]::uuid[])
)
SELECT created_at, change_set_id, event_rank, sequence_no, revision_no FROM events
WHERE $3::timestamptz IS NULL OR created_at < $3
 OR (created_at = $3 AND change_set_id < $4)
 OR (created_at = $3 AND change_set_id = $4 AND event_rank > $5)
 OR (created_at = $3 AND change_set_id = $4 AND event_rank = $5 AND sequence_no > $6)
 OR (created_at = $3 AND change_set_id = $4 AND event_rank = $5 AND sequence_no = $6 AND revision_no < $7)
ORDER BY created_at DESC, change_set_id DESC, event_rank ASC, sequence_no ASC, revision_no DESC
LIMIT $8`

func (historyQueryRepository) SelectDescriptorsTx(ctx context.Context, tx pgx.Tx, record RecordHistoryRecord, query HistoryQuery) ([]HistoryPosition, error) {
	var afterTime *time.Time
	var afterID uuid.UUID
	var rank, sequence int
	var revision int64
	if query.After != nil {
		afterTime, afterID, sequence, revision = &query.After.CommittedAt, query.After.ChangeSetID, query.After.SequenceNo, query.After.RevisionNo
		if query.After.Kind == HistoryRevision {
			rank = 1
		}
	}
	rows, err := tx.Query(ctx, historyDescriptorsSQL, record.RecordID, record.IncidentID, afterTime, afterID, rank, sequence, revision, query.Limit+1)
	if err != nil {
		return nil, fmt.Errorf("select history descriptors: %w", err)
	}
	defer rows.Close()
	result := make([]HistoryPosition, 0, query.Limit+1)
	for rows.Next() {
		var position HistoryPosition
		var eventRank int
		if err := rows.Scan(&position.CommittedAt, &position.ChangeSetID, &eventRank, &position.SequenceNo, &position.RevisionNo); err != nil {
			return nil, err
		}
		position.Kind = HistoryMutation
		if eventRank == 1 {
			position.Kind = HistoryRevision
		}
		if err := position.Validate(); err != nil {
			return nil, fmt.Errorf("invalid stored history descriptor: %w", err)
		}
		result = append(result, position)
	}
	return result, rows.Err()
}

func selectedMutationKeys(selected []HistoryPosition) ([]uuid.UUID, []int) {
	ids, sequences := make([]uuid.UUID, 0, len(selected)), make([]int, 0, len(selected))
	for _, position := range selected {
		if position.Kind == HistoryMutation {
			ids = append(ids, position.ChangeSetID)
			sequences = append(sequences, position.SequenceNo)
		}
	}
	return ids, sequences
}

func (historyQueryRepository) LoadMutationRowsTx(ctx context.Context, tx pgx.Tx, record RecordHistoryRecord, selected []HistoryPosition, rowKinds []string) ([]mutationHistoryRow, error) {
	ids, sequences := selectedMutationKeys(selected)
	if len(ids) == 0 {
		return nil, nil
	}
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
       rr.revision_count,
       href.history_entry_ref,
       $1 = ANY(csm.history_entry_record_ids),
       csm.sequence_no = (SELECT min(associated.sequence_no) FROM change_set_mutations associated
                         WHERE associated.change_set_id = cs.change_set_id
                           AND associated.history_record_ids @> ARRAY[$1]::uuid[])
       AND NOT EXISTS (SELECT 1 FROM change_set_mutations associated
                       WHERE associated.change_set_id = cs.change_set_id
                         AND associated.history_record_ids @> ARRAY[$1]::uuid[]
                         AND associated.target_kind = ANY($5::text[]))
  FROM unnest($3::uuid[], $4::int[]) selected(change_set_id, sequence_no)
  JOIN change_sets cs ON cs.change_set_id = selected.change_set_id
  JOIN change_set_mutations csm
    ON csm.change_set_id = cs.change_set_id AND csm.sequence_no = selected.sequence_no
  LEFT JOIN LATERAL (
    SELECT max(row_version) AS row_version, count(*) AS revision_count
    FROM record_revisions WHERE change_set_id = cs.change_set_id AND record_id = $1
  ) rr ON true
  LEFT JOIN record_history_entry_refs href
    ON href.record_id = $1
	   AND href.change_set_id = csm.change_set_id
	   AND href.mutation_sequence_no = csm.sequence_no
	 WHERE cs.incident_id = $2
	   AND csm.history_record_ids @> ARRAY[$1]::uuid[]
	 ORDER BY cs.created_at DESC, cs.change_set_id DESC, csm.sequence_no ASC
	`, record.RecordID, record.IncidentID, ids, sequences, rowKinds)
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
			&row.RevisionCount,
			&historyEntryRef,
			&row.HistoryEntryAddressable,
			&row.CoalesceRevision,
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

func (historyQueryRepository) LoadRevisionRowsTx(ctx context.Context, tx pgx.Tx, record RecordHistoryRecord, selected []HistoryPosition) ([]revisionHistoryRow, error) {
	ids, versions := make([]uuid.UUID, 0, len(selected)), make([]int64, 0, len(selected))
	for _, position := range selected {
		ids = append(ids, position.ChangeSetID)
		versions = append(versions, position.RevisionNo)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := tx.Query(ctx, `
SELECT cs.change_set_id,
       cs.actor_user_id,
       cs.created_at,
       cs.source,
       rr.row_version,
       rr.before_json,
       rr.after_json
  FROM unnest($3::uuid[], $4::bigint[]) selected(change_set_id, row_version)
  JOIN record_revisions rr ON rr.change_set_id = selected.change_set_id AND rr.row_version = selected.row_version
  JOIN change_sets cs
    ON cs.change_set_id = rr.change_set_id
 WHERE rr.record_id = $1
   AND cs.incident_id = $2
 ORDER BY cs.created_at DESC, cs.change_set_id DESC, rr.row_version DESC
`, record.RecordID, record.IncidentID, ids, versions)
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
