package graphassembly

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresrestore"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

type graphProjectionDerivedStateReconciler struct{}

func (graphProjectionDerivedStateReconciler) ReconcileGraphProjectionDerivedState(ctx context.Context, tx pgx.Tx) (int, error) {
	return networkflow.ReconcileGraphRestoreJobsTx(ctx, tx)
}

// RebuildClaimedGraph is invoked once after successful profile admission. All
// source reads, reconstruction and publication share its repeatable-read
// transaction; the writer owns the final commit and committed-state proof.
func RebuildClaimedGraph(ctx context.Context, db dbport.DB, request graphrestore.RestoreRebuildRequest) (graphrestore.RestoreRebuildResult, error) {
	if db == nil || ctx == nil {
		return graphrestore.RestoreRebuildResult{}, graphrestore.NewRestoreError(graphrestore.RestoreErrorInvalidRequest)
	}
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return graphrestore.RestoreRebuildResult{}, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	registration, err := networkflow.NewGraphRestoreSourceRegistration(tx)
	if err != nil {
		return graphrestore.RestoreRebuildResult{}, err
	}
	registry, err := graphrestore.NewCurrentRestoreSourceRegistry(registration)
	if err != nil {
		return graphrestore.RestoreRebuildResult{}, err
	}
	writer, err := postgresrestore.New(tx, db, graphProjectionDerivedStateReconciler{})
	if err != nil {
		return graphrestore.RestoreRebuildResult{}, err
	}
	service, err := graphrestore.NewRestoreService(writer, registry)
	if err != nil {
		return graphrestore.RestoreRebuildResult{}, err
	}
	return service.Rebuild(ctx, request)
}
