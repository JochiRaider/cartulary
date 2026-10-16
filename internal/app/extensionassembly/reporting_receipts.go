package extensionassembly

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// ReportingReceiptReconciler composes the route owners used by Reporting's
// workers. Only the ordinary snapshot/release routes return a common Job.
type ReportingReceiptReconciler struct{}

func (ReportingReceiptReconciler) ReconcileFinalIdempotencyOutcomeTx(ctx context.Context, tx pgx.Tx, key jobs.RouteIdempotencyKey, requestHash []byte, resource jobs.Resource) (bool, error) {
	switch key.RouteKey {
	case "report_compositions.preview":
		return (reportcomposition.PreviewReceiptReconciler{}).ReconcileFinalIdempotencyOutcomeTx(ctx, tx, key, requestHash, resource)
	case "snapshots.create", "releases.create":
		return (JobOwnerTransactionAdapters{}).ReconcileFinalIdempotencyOutcomeTx(ctx, tx, key, requestHash, resource)
	default:
		return false, errors.New("reporting finalizer has no receipt owner for route")
	}
}
