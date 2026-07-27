package appsupport

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func IncidentCreateCommitFaultDependencies() httpapi.DependencySet {
	return server.DependencySetWithPostgresDBDecoratorForTesting(func(base postgres.DB) postgres.DB {
		return incidentCreateCommitFaultDB{DB: base}
	})
}

type incidentCreateCommitFaultDB struct {
	postgres.DB
}

func (db incidentCreateCommitFaultDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return &incidentCreateCommitFaultTx{Tx: tx}, nil
}

type incidentCreateCommitFaultTx struct {
	pgx.Tx
	incidentCreate bool
}

func (tx *incidentCreateCommitFaultTx) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	tx.observe(query)
	return tx.Tx.Exec(ctx, query, args...)
}

func (tx *incidentCreateCommitFaultTx) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	tx.observe(query)
	return tx.Tx.Query(ctx, query, args...)
}

func (tx *incidentCreateCommitFaultTx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	tx.observe(query)
	return tx.Tx.QueryRow(ctx, query, args...)
}

func (tx *incidentCreateCommitFaultTx) Commit(ctx context.Context) error {
	if !tx.incidentCreate {
		return tx.Tx.Commit(ctx)
	}
	_ = tx.Tx.Rollback(ctx)
	return errors.New("forced incident create commit failure")
}

func (tx *incidentCreateCommitFaultTx) observe(query string) {
	normalized := strings.ToLower(query)
	if strings.Contains(normalized, "insert into incidents") {
		tx.incidentCreate = true
	}
}
