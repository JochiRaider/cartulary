package extensionstore

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// WithAdmissionRead supplies a stable, database-enforced read-only snapshot.
// Admission cannot commit changes, including on rejection or cancellation.
func (s *Store) WithAdmissionRead(ctx context.Context, read func(Querier) error) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	return read(tx)
}

func ReadJobCommitProofPage(ctx context.Context, reader Querier, profileID string, after uuid.UUID) ([]uuid.UUID, error) {
	rows, err := reader.Query(ctx, `SELECT job_id FROM extension_job_commit_proofs WHERE owner_profile_id=$1 AND job_id>$2 ORDER BY job_id LIMIT 128`, profileID, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]uuid.UUID, 0, 128)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
