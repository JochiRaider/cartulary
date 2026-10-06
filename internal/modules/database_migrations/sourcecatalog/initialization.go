package sourcecatalog

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3/database"
	gooselock "github.com/pressly/goose/v3/lock"
)

// NewInitializationLocker protects fresh admission through format commit. The
// existing provider lock separately serializes migration work on its own session.
func NewInitializationLocker() (gooselock.SessionLocker, error) {
	return gooselock.NewPostgresSessionLocker(gooselock.WithLockID(4097083627), gooselock.WithLockTimeout(1, 300), gooselock.WithUnlockTimeout(1, 30))
}

// InitializeLedger uses Goose's supported store implementation on the admission
// session. A committed empty ledger is interrupted initialization, not freshness.
// If this session is lost later, another initializer rejects that retained state.
func InitializeLedger(ctx context.Context, conn *sql.Conn) error {
	store, err := database.NewStore(database.DialectPostgres, "public.goose_db_version")
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := store.CreateVersionTable(ctx, tx); err != nil {
		return err
	}
	if err := store.Insert(ctx, tx, database.InsertRequest{Version: 0}); err != nil {
		return err
	}
	return tx.Commit()
}
