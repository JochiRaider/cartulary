package workbook_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/jackc/pgx/v5"
)

// The wrapper observes the production catalog's database boundary. A write
// after the first SELECT closes is deterministic and requires no timing sleeps.
type locatorReadDB struct {
	postgres.DB
	afterRead                    func()
	once                         sync.Once
	beginErr, readErr, commitErr error
	begins, commits, rollbacks   int
	options                      pgx.TxOptions
}

func (db *locatorReadDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	db.begins++
	db.options = options
	if db.beginErr != nil {
		return nil, db.beginErr
	}
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return &locatorReadTx{Tx: tx, db: db}, nil
}

func (db *locatorReadDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("locator escaped its transaction through the pool")
}

type locatorReadTx struct {
	pgx.Tx
	db *locatorReadDB
}

func (tx *locatorReadTx) Begin(context.Context) (pgx.Tx, error) {
	return nil, errors.New("locator nested its borrowed transaction")
}

func (tx *locatorReadTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if tx.db.readErr != nil {
		return nil, tx.db.readErr
	}
	rows, err := tx.Tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &locatorReadRows{Rows: rows, db: tx.db}, nil
}

func (tx *locatorReadTx) Commit(ctx context.Context) error {
	tx.db.commits++
	if tx.db.commitErr != nil {
		_ = tx.Tx.Rollback(ctx)
		return tx.db.commitErr
	}
	return tx.Tx.Commit(ctx)
}

func (tx *locatorReadTx) Rollback(ctx context.Context) error {
	tx.db.rollbacks++
	return tx.Tx.Rollback(ctx)
}

type locatorReadRows struct {
	pgx.Rows
	db *locatorReadDB
}

func (rows *locatorReadRows) Close() {
	rows.Rows.Close()
	rows.db.once.Do(func() {
		if rows.db.afterRead != nil {
			rows.db.afterRead()
		}
	})
}

func TestWorkbookLocatorConsistentReadAndTransactionLifetime_Integration(t *testing.T) {
	h, login, _, incidentID := ConflictFixture(t, "locator-snapshot", "IR-LOCATOR-SNAPSHOT")
	for _, test := range []struct{ name, view, field string }{
		{"host", entitycontract.HostsViewSchemaID, "host.display_name"},
		{"identity", entitycontract.IdentitiesViewSchemaID, "identity.display_name"},
		{"note", artifacts.NotesViewSchemaID, "note.title"},
	} {
		t.Run(test.name, func(t *testing.T) {
			created := requireWorkbookCreate(t, h, login, incidentID, test.view, map[string]any{
				"client_txn_id": "snapshot-create-" + test.name, test.field: "Before",
			})["row"].(map[string]any)
			recordID := appsupport.MustUUID(t, created["record_id"].(string))
			schema, ok := viewschema.Lookup(test.view)
			if !ok {
				t.Fatal("missing schema")
			}
			command := workbook.LocateCommand{IncidentID: incidentID, ViewSchemaID: test.view, RecordID: recordID, Query: schema.DefaultQueryMeta()}
			locate := func(ctx context.Context, db *locatorReadDB) (string, error) {
				t.Helper()
				catalog := appsupport.NewWorkbookCatalog(db, workbookTestConflictTokens())
				provider, ok := catalog.QueryFor(test.view)
				if !ok {
					t.Fatal("missing locator provider")
				}
				location, err := provider.LocateRows(ctx, command)
				if err != nil {
					return "", err
				}
				if location.Outcome != "located" || len(location.Page.Rows) == 0 || location.Page.Rows[0]["record_id"] != recordID.String() {
					t.Fatalf("unexpected location: %#v", location)
				}
				return location.Page.Rows[0]["cells"].(map[string]any)[test.field].(map[string]any)["value"].(string), nil
			}
			assertLifetime := func(db *locatorReadDB, commits, rollbacks int) {
				t.Helper()
				if db.begins != 1 || db.commits != commits || db.rollbacks != rollbacks ||
					db.options.IsoLevel != pgx.RepeatableRead || db.options.AccessMode != pgx.ReadOnly {
					t.Fatalf("invalid transaction lifetime: begins=%d commits=%d rollbacks=%d options=%#v", db.begins, db.commits, db.rollbacks, db.options)
				}
			}
			db := &locatorReadDB{DB: h.Pool}
			db.afterRead = func() {
				if db.commits != 0 || db.rollbacks != 0 {
					t.Fatal("reader ended the caller transaction")
				}
				requireWorkbookPatch(t, h, login, recordID, map[string]any{
					"view_schema_id": test.view, "base_row_version": 1,
					"client_txn_id": "snapshot-update-" + test.name,
					"changes":       []map[string]any{{"field_key": test.field, "value": "After"}},
				})
			}
			value, err := locate(t.Context(), db)
			if err != nil || value != "Before" {
				t.Fatalf("snapshot changed during location: value=%q error=%v", value, err)
			}
			assertLifetime(db, 1, 1)
			fresh := &locatorReadDB{DB: h.Pool}
			value, err = locate(t.Context(), fresh)
			if err != nil || value != "After" {
				t.Fatalf("new location missed committed change: value=%q error=%v", value, err)
			}
			assertLifetime(fresh, 1, 1)

			for _, stage := range []string{"begin", "read", "commit", "cancel"} {
				t.Run(stage, func(t *testing.T) {
					failure := fmt.Errorf("injected locator %s failure", stage)
					db := &locatorReadDB{DB: h.Pool}
					ctx, cancel := context.WithCancel(t.Context())
					defer cancel()
					commits, rollbacks := 0, 1
					switch stage {
					case "begin":
						db.beginErr, rollbacks = failure, 0
					case "read":
						db.readErr = failure
					case "commit":
						db.commitErr, commits = failure, 1
					case "cancel":
						db.afterRead, failure = cancel, context.Canceled
					}
					_, err := locate(ctx, db)
					if !errors.Is(err, failure) {
						t.Fatalf("operational failure became a location outcome: %v", err)
					}
					assertLifetime(db, commits, rollbacks)
				})
			}
		})
	}
}
