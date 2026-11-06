package revisions

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func BeginIncidentBundleImportedRevisionSequenceTx(
	ctx context.Context,
	tx pgx.Tx,
) (int64, error) {
	var originalNext int64
	if err := tx.QueryRow(ctx, `
SELECT public.revisions_incident_bundle_sequence_begin_v1()
`).Scan(&originalNext); err != nil {
		return 0, err
	}
	return originalNext, nil
}

func FinishIncidentBundleImportedRevisionSequenceTx(
	ctx context.Context,
	tx pgx.Tx,
	originalNext int64,
) error {
	var repairedNext int64
	return tx.QueryRow(ctx, `
SELECT public.revisions_incident_bundle_sequence_finish_v1($1)
`, originalNext).Scan(&repairedNext)
}
