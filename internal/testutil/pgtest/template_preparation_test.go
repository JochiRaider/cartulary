package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/testutil/postgrescatalog"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestTemplatePreparationCompletesCopyBeforePublication(t *testing.T) {
	if suiteservices.LookupEnvValue(nil, suiteservices.PGAdminDSNEnv) != "" {
		t.Fatal("self-owned lifecycle selection acquired an unused broker cluster")
	}
	for _, failure := range []struct {
		name string
		err  error
	}{
		{"seal_failure", errors.New("injected template sealing failure")},
		{"seal_cancellation", context.Canceled},
	} {
		t.Run(failure.name, func(t *testing.T) {
			h := startTemplatePreparationHarness(t)
			original := sealTemplateDBFn
			sealTemplateDBFn = func(ctx context.Context, adminDSN, name string) error {
				if errors.Is(failure.err, context.Canceled) {
					cancelled, cancel := context.WithCancel(ctx)
					cancel()
					return original(cancelled, adminDSN, name)
				}
				return failure.err
			}
			t.Cleanup(func() { sealTemplateDBFn = original })
			clone, err := h.PrepareDatabase(t.Context(), "unpublished_template")
			if !errors.Is(err, failure.err) || clone != nil || h.templateDB != "" {
				t.Fatalf("failed sealing published a template or clone: clone=%v template=%q error=%v", clone, h.templateDB, err)
			}
			name := templateDatabaseName(h.suiteHash, h.schemaHash)
			admin, err := sql.Open("pgx", h.AdminDSN())
			if err != nil {
				t.Fatal(err)
			}
			defer admin.Close()
			var exists bool
			if err := admin.QueryRowContext(t.Context(), `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname=$1)`, name).Scan(&exists); err != nil {
				t.Fatal(err)
			}
			// Ownership is historical proof retained for idempotent cleanup;
			// physical absence, not removal of that proof, establishes settlement.
			if exists {
				t.Fatal("failed sealing did not settle its exact unpublished template")
			}
		})
	}
	// A separate physical cluster makes WAL publication and ownership assertions
	// independent of other test processes and their template preparation.
	h := startTemplatePreparationHarness(t)
	ctx := t.Context()
	clone := func(prefix string) *TestDatabase {
		t.Helper()
		database, err := h.PrepareDatabase(ctx, prefix)
		if err != nil {
			t.Fatal(err)
		}
		h.registerDatabaseCleanupT(t, database.Name, suiteservices.FixtureReusePerTest, fixtureAttributionFor(t, "pgtest"))
		return database
	}
	admin, err := sql.Open("pgx", h.AdminDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close() })
	originalSeal := sealTemplateDBFn
	sealCalls := 0
	sealTemplateDBFn = func(ctx context.Context, adminDSN, name string) error {
		sealCalls++
		var initializedThrough string
		if err := admin.QueryRowContext(ctx, `SELECT pg_current_wal_insert_lsn()::text`).Scan(&initializedThrough); err != nil {
			return err
		}
		if err := originalSeal(ctx, adminDSN, name); err != nil {
			return err
		}
		var settled bool
		if err := admin.QueryRowContext(ctx, `SELECT redo_lsn >= $1::pg_lsn FROM pg_control_checkpoint()`, initializedThrough).Scan(&settled); err != nil {
			return err
		}
		if !settled {
			t.Fatal("template was sealed before its initialization checkpoint completed")
		}
		return nil
	}
	t.Cleanup(func() { sealTemplateDBFn = originalSeal })
	if err := h.ensureLocalTemplateDatabase(ctx); err != nil {
		t.Fatal(err)
	}
	var sealedCheckpoint string
	if err := admin.QueryRowContext(ctx, `SELECT checkpoint_lsn::text FROM pg_control_checkpoint()`).Scan(&sealedCheckpoint); err != nil {
		t.Fatal(err)
	}
	var durable bool
	if err := admin.QueryRowContext(ctx, `SELECT current_setting('fsync') = 'on'
AND current_setting('full_page_writes') = 'on' AND current_setting('synchronous_commit') = 'on'
AND current_setting('data_checksums') = 'on'`).Scan(&durable); err != nil {
		t.Fatal(err)
	}
	if !durable {
		t.Fatal("clone fixture disabled production durability settings")
	}
	var before string
	if err := admin.QueryRowContext(ctx, "SELECT pg_current_wal_insert_lsn()::text").Scan(&before); err != nil {
		t.Fatal(err)
	}
	first := clone("prepared_first")
	var walFlushed bool
	if err := admin.QueryRowContext(ctx, "SELECT pg_current_wal_flush_lsn() > $1::pg_lsn", before).Scan(&walFlushed); err != nil {
		t.Fatal(err)
	}
	if !walFlushed {
		t.Fatal("template copy was published without durable WAL progress")
	}
	second := clone("prepared_second")
	var unchangedCheckpoint bool
	if err := admin.QueryRowContext(ctx, `SELECT checkpoint_lsn = $1::pg_lsn FROM pg_control_checkpoint()`, sealedCheckpoint).Scan(&unchangedCheckpoint); err != nil {
		t.Fatal(err)
	}
	if sealCalls != 1 || !unchangedCheckpoint {
		t.Fatal("clone creation repeated template sealing or requested a checkpoint")
	}
	// Database ACLs are database-specific, not inherited template authority.
	for _, name := range []string{first.Name, second.Name} {
		for _, login := range []string{"cartulary_migration_login", "cartulary_runtime_login", "cartulary_recovery_login"} {
			var connect, temporary, create bool
			if err := admin.QueryRowContext(ctx, `SELECT has_database_privilege($1,$2,'CONNECT'),has_database_privilege($1,$2,'TEMPORARY'),has_database_privilege($1,$2,'CREATE')`, login, name).Scan(&connect, &temporary, &create); err != nil || !connect || temporary || create {
				t.Fatalf("clone grants %s: connect=%v temporary=%v create=%v error=%v", login, connect, temporary, create, err)
			}
		}
	}
	// Independent target locks permit concurrent clone finalization. The fixed
	// roles and privileges already exist; no consumer may repair them.
	var clones sync.WaitGroup
	cloneErrors := make(chan error, 2)
	for range 2 {
		name := h.nextDatabaseName("parallel_clone")
		clones.Go(func() {
			err := h.createDatabase(ctx, name, h.templateDB)
			if err == nil {
				err = h.DropDatabase(ctx, name)
			}
			cloneErrors <- err
		})
	}
	clones.Wait()
	close(cloneErrors)
	for err := range cloneErrors {
		if err != nil {
			t.Fatal("concurrent isolated clone", err)
		}
	}
	firstDB, err := sql.Open("pgx", first.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = firstDB.Close() })
	if _, err := firstDB.ExecContext(ctx, "CREATE TABLE public.clone_isolation_probe (id integer PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	secondDB, err := sql.Open("pgx", second.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = secondDB.Close() })
	var isolated, migrated bool
	if err := secondDB.QueryRowContext(ctx, "SELECT to_regclass('public.clone_isolation_probe') IS NULL, to_regclass('public.goose_db_version') IS NOT NULL").Scan(&isolated, &migrated); err != nil {
		t.Fatal(err)
	}
	if !isolated || !migrated {
		t.Fatal("prepared clones lost template schema or shared mutable state")
	}
	if err := h.createDatabase(ctx, second.Name, h.templateDB); err == nil {
		t.Fatal("template copy replaced an existing database")
	}
	if err := firstDB.Close(); err != nil {
		t.Fatal(err)
	}
	if err := h.DropDatabase(ctx, first.Name); err != nil {
		t.Fatal(err)
	}
	if err := secondDB.PingContext(ctx); err != nil {
		t.Fatalf("dropping one clone disrupted its live peer: %v", err)
	}
	var templatePresent bool
	if err := admin.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname=$1 AND datistemplate AND NOT datallowconn)", h.templateDB).Scan(&templatePresent); err != nil || !templatePresent {
		t.Fatalf("template changed after clone cleanup: present=%v error=%v", templatePresent, err)
	}
	t.Run("failed_preparation_removes_only_its_created_database", func(t *testing.T) {
		// The authenticated fixture administrator may assume this role, which
		// can create its own databases but cannot checkpoint or provision roles.
		if _, err := admin.ExecContext(ctx, `CREATE ROLE ct_clone_preparation_limited CREATEDB`); err != nil {
			t.Fatal(err)
		}
		limited, err := url.Parse(h.AdminDSN())
		if err != nil {
			t.Fatal(err)
		}
		query := limited.Query()
		query.Set("options", "-crole=ct_clone_preparation_limited")
		limited.RawQuery = query.Encode()
		probe, err := sql.Open("pgx", limited.String())
		if err != nil {
			t.Fatal(err)
		}
		defer probe.Close()
		var canCreate bool
		if err := probe.QueryRowContext(ctx, `SELECT rolcreatedb FROM pg_roles WHERE rolname = current_user`).Scan(&canCreate); err != nil || !canCreate {
			t.Fatalf("limited creation role is unavailable: %v", err)
		}
		_, err = probe.ExecContext(ctx, "CHECKPOINT")
		var permission *pgconn.PgError
		if !errors.As(err, &permission) || permission.Code != "42501" {
			t.Fatal("checkpoint permission boundary was not reached")
		}
		// CREATE's synchronous WAL commit is the publication boundary. An
		// unrelated cluster-wide checkpoint must not be a clone prerequisite.
		cloneName := h.nextDatabaseName("unprivileged_clone")
		created, cloneErr := postgrescatalog.CloneDatabase(ctx, limited.String(), cloneName, h.templateDB, postgresCatalogAdmissionTimeout, func(*sql.Conn) error { return nil })
		if created {
			t.Cleanup(func() {
				if err := postgrescleanup.DropOwnedDatabase(context.Background(), h.AdminDSN(), cloneName); err != nil {
					t.Error(err)
				}
			})
		}
		if !created || cloneErr != nil {
			t.Fatalf("durable clone requires unrelated checkpoint privilege: created=%v error=%v", created, cloneErr)
		}
		name := h.nextDatabaseName("failed_preparation")
		oldFinalizer := finalizeCloneFn
		injected := errors.New("injected clone finalization failure")
		finalizeCloneFn = func(ctx context.Context, connection *sql.Conn, target string) error {
			var connects bool
			if err := connection.QueryRowContext(ctx, `SELECT datallowconn FROM pg_database WHERE datname=$1`, target).Scan(&connects); err != nil || connects {
				t.Fatal("unpublished clone allowed connections", connects, err)
			}
			return injected
		}
		t.Cleanup(func() { finalizeCloneFn = oldFinalizer })
		err = createDatabase(ctx, h.AdminDSN(), name, h.templateDB)
		finalizeCloneFn = oldFinalizer
		if !errors.Is(err, injected) {
			t.Fatalf("lost preparation failure: %v", err)
		}
		var present bool
		if err := admin.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, name).Scan(&present); err != nil {
			t.Fatal(err)
		}
		if present {
			t.Fatal("failed preparation leaked its created database")
		}
		if err := secondDB.PingContext(ctx); err != nil {
			t.Fatalf("failed acquisition removed an existing peer: %v", err)
		}
	})

	// A failed copy must not publish a database or acquire ownership of it.
	failedName := h.nextDatabaseName("failed_copy")
	if err := h.createDatabase(ctx, failedName, "ct_missing_preparation_template"); err == nil {
		t.Fatal("missing template was accepted")
	}
	var exists bool
	if err := admin.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname=$1)", failedName).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists || h.ownsDatabase(failedName) {
		t.Fatalf("failed template copy was published: exists=%v owned=%v", exists, h.ownsDatabase(failedName))
	}
}

func startTemplatePreparationHarness(t *testing.T) *Harness {
	t.Helper()
	h, err := StartOwned(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := h.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	return h
}
