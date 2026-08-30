package recoveryassembly

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresrestore"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type graphProjectionDerivedStateReconciler struct {
	now func() time.Time
}

func (reconciler graphProjectionDerivedStateReconciler) ReconcileGraphProjectionDerivedState(
	ctx context.Context,
	tx pgx.Tx,
	bindings []graphprojection.ResultBindingV2,
) (int, int, error) {
	networkFlowJobs, err := networkflow.ReconcileGraphRestoreJobsTx(ctx, tx)
	if err != nil {
		return 0, 0, err
	}
	observedAt := time.Now().UTC()
	if reconciler.now != nil {
		observedAt = reconciler.now().UTC()
	}
	reportingJobs, leases, err := reporting.ReconcileGraphProjectionRestoreTx(ctx, tx, bindings, observedAt)
	if err != nil {
		return 0, 0, err
	}
	return networkFlowJobs + reportingJobs, leases, nil
}

// NewGraphProjectionRestoreParticipant composes source-owner registrations
// with Graph's pure v2 rebuild engine and narrow borrowed-Postgres writer.
func NewGraphProjectionRestoreParticipant(db postgres.DB) (restorecontract.GraphProjectionParticipant, error) {
	registration, err := networkflow.NewGraphRestoreSourceRegistration(db)
	if err != nil {
		return nil, err
	}
	registry, err := graphrestore.NewCurrentRestoreSourceRegistry(registration)
	if err != nil {
		return nil, err
	}
	writer, err := postgresrestore.New(db, graphProjectionDerivedStateReconciler{})
	if err != nil {
		return nil, err
	}
	return graphrestore.NewRestoreService(writer, registry)
}
