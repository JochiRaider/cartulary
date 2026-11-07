package referenceassembly

import (
	"context"
	"errors"

	referenceadmission "github.com/JochiRaider/cartulary/internal/modules/reference_data/contractadmission"
	reportingadmission "github.com/JochiRaider/cartulary/internal/modules/reporting/referenceadmission"
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
