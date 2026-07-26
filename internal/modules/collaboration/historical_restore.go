package collaboration

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// SuppressHistoricalIntentsTx marks a restore transaction as historical source
// replay. Collaboration capture triggers then avoid backfilling restored
// revisions into the live incident stream.
func SuppressHistoricalIntentsTx(ctx context.Context, tx pgx.Tx) error {
	if tx == nil {
		return errors.New("historical restore transaction is required")
	}
	_, err := tx.Exec(
		ctx,
		`SELECT set_config('cartulary.collaboration.suppress_historical_intents', 'on', true)`,
	)
	return err
}
