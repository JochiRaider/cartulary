package pgtest

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/testutil/postgrescatalog"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
)

func TestTemplatePreparationCompletesCopyBeforePublication(t *testing.T) {
	// A separate physical cluster makes WAL publication and ownership assertions
	// independent of other test processes and their template preparation.
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
	ctx := t.Context()
	if err := h.ensureLocalTemplateDatabase(ctx); err != nil {
		t.Fatal(err)
	}
	admin, err := sql.Open("pgx", h.AdminDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close() })
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
	first := h.PrepareIsolatedDatabaseT(t, "prepared_first")
	var walFlushed bool
	if err := admin.QueryRowContext(ctx, "SELECT pg_current_wal_flush_lsn() > $1::pg_lsn", before).Scan(&walFlushed); err != nil {
		t.Fatal(err)
	}
	if !walFlushed {
		t.Fatal("template copy was published without durable WAL progress")
	}
	second := h.PrepareIsolatedDatabaseT(t, "prepared_second")
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
		created, cloneErr := postgrescatalog.CloneDatabase(ctx, limited.String(), cloneName, h.templateDB, postgresCatalogAdmissionTimeout)
		if created {
			t.Cleanup(func() {
				if _, err := postgrescleanup.DropOwnedDatabase(context.Background(), h.AdminDSN(), cloneName,
					postgresDatabaseNormalDropTimeout, postgresDatabaseForcedDropTimeout); err != nil {
					t.Error(err)
				}
			})
		}
		if !created || cloneErr != nil {
			t.Fatalf("durable clone requires unrelated checkpoint privilege: created=%v error=%v", created, cloneErr)
		}
		name := h.nextDatabaseName("failed_preparation")
		err = createDatabase(ctx, limited.String(), name, h.templateDB)
		if !errors.As(err, &permission) || permission.Code != "42501" {
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
