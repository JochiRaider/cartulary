package collaboration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const historicalIntentSuppressionSetting = "cartulary.collaboration.suppress_historical_intents"

// HistoricalIntentPolicy owns the transaction-local exception that prevents
// imported history from being projected as current live Collaboration events.
type HistoricalIntentPolicy struct{}

func NewHistoricalIntentPolicy() *HistoricalIntentPolicy {
	return &HistoricalIntentPolicy{}
}

func (*HistoricalIntentPolicy) SuppressTx(ctx context.Context, tx pgx.Tx) error {
	if tx == nil {
		return errors.New("historical restore transaction is required")
	}
	_, err := tx.Exec(
		ctx,
		`SELECT set_config($1, 'on', true)`,
		historicalIntentSuppressionSetting,
	)
	if err != nil {
		return fmt.Errorf("suppress historical collaboration intents: %w", err)
	}
	return nil
}

// SuppressSQLTx supports repository fixtures and migration-oriented stdlib
// callers without exposing the owned setting name outside Collaboration.
func (*HistoricalIntentPolicy) SuppressSQLTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return errors.New("historical restore transaction is required")
	}
	if _, err := tx.ExecContext(
		ctx,
		`SELECT set_config($1, 'on', true)`,
		historicalIntentSuppressionSetting,
	); err != nil {
		return fmt.Errorf("suppress historical collaboration intents: %w", err)
	}
	return nil
}

func (*HistoricalIntentPolicy) IsSuppressedTx(ctx context.Context, tx pgx.Tx) (bool, error) {
	if tx == nil {
		return false, errors.New("historical restore transaction is required")
	}
	var suppressed bool
	if err := tx.QueryRow(
		ctx,
		`SELECT COALESCE(current_setting($1, true), '') = 'on'`,
		historicalIntentSuppressionSetting,
	).Scan(&suppressed); err != nil {
		return false, fmt.Errorf("read historical collaboration intent suppression: %w", err)
	}
	return suppressed, nil
}
