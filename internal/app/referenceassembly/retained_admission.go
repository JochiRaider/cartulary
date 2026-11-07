package referenceassembly

import (
	"context"
	"database/sql"
	"errors"

	referenceadmission "github.com/JochiRaider/cartulary/internal/modules/reference_data/contractadmission"
	reportingadmission "github.com/JochiRaider/cartulary/internal/modules/reporting/migrationadmission"
	"github.com/jackc/pgx/v5"
)

// ValidateRetainedRuntimeState coordinates owner inspections before application
// leases, reconciliation, input publication or readiness. Both owners observe
// the same snapshot and receive no write capability.
func ValidateRetainedRuntimeState(ctx context.Context, db interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}) (resultErr error) {
	if db == nil {
		return errors.New("retained state admission database required")
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, tx.Rollback(ctx)) }()
	if err := referenceadmission.Validate(ctx, tx); err != nil {
		return err
	}
	return reportingadmission.ValidateReferenceBindings(ctx, tx, func(b []byte) error {
		return referenceadmission.ValidateBinding(ctx, tx, b)
	})
}

// Migration owns exclusion and connection lifetime; this composition owns only
// the order of read-only owner inspections before any migration writes.
func ValidateRetainedMigrationState(ctx context.Context, conn *sql.Conn) error {
	if err := referenceadmission.ValidateMigration(ctx, conn); err != nil {
		return err
	}
	return reportingadmission.ValidateReferenceBindingsMigration(ctx, conn, func(b []byte) error {
		return referenceadmission.ValidateBindingMigration(ctx, conn, b)
	})
}
