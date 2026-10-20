package recoverybrowsertest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
)

func TestDatabaseDropObservation_Integration(t *testing.T) {
	h := pgtest.Start(t)
	database := h.NewDatabaseT(t, "drop_observer")
	admin, err := sql.Open("pgx", h.AdminDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	t.Cleanup(cancel)
	transaction, err := admin.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = transaction.Rollback() })
	// Hold only this owned database's object lock. The uncommitted rename
	// leaves the original name visible to DROP until the transaction rolls back.
	name, err := pgIdentifier(database.Name)
	if err != nil {
		t.Fatal(err)
	}
	heldName, err := pgIdentifier(database.Name + "_held")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := transaction.ExecContext(ctx, "ALTER DATABASE "+name+" RENAME TO "+heldName); err != nil {
		t.Fatal(err)
	}
	dropCtx, cancelDrop := context.WithTimeout(ctx, 2500*time.Millisecond)
	stopped := make(chan struct{})
	var dropErr error
	go func() {
		defer close(stopped)
		dropErr = postgrescleanup.DropOwnedDatabase(dropCtx, h.AdminDSN(), database.Name)
	}()
	t.Cleanup(func() {
		cancelDrop()
		_ = transaction.Rollback()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			t.Error("database drop did not settle after cancellation and lock release")
		}
	})
	query := fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, database.Name)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked bool
		err := admin.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_catalog.pg_stat_activity WHERE state = 'active' AND query = $1 AND wait_event_type = 'Lock')`, query).Scan(&blocked)
		if err != nil {
			t.Fatal(err)
		}
		if blocked {
			break
		}
		select {
		case <-stopped:
			t.Fatalf("drop completed before the held object lock was observed: %v", dropErr)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
		}
	}
	select {
	case <-stopped:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	value := postgrescleanup.ObservationFromError(dropErr)
	if !errors.Is(dropErr, context.DeadlineExceeded) || value == nil {
		t.Fatalf("missing attributed drop observation: %v", dropErr)
	}
	if value.ServiceGeneration == nil || *value.ServiceGeneration != h.ProvisioningGeneration() {
		t.Fatal("drop wait lacks the actual owned PostgreSQL generation")
	}
	observation := value.Wait.String()
	for _, expected := range []string{"type=Lock event=object blockers=1", "checkpointer={", "requested=", "done=", "buffers="} {
		if !strings.Contains(observation, expected) {
			t.Errorf("drop observation %q lacks %q", observation, expected)
		}
	}
	if len(observation) > 1024 || strings.Contains(observation, database.Name) || strings.Contains(observation, h.AdminDSN()) {
		t.Error("database observer exceeded its bounded structural output")
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-stopped:
		if !errors.Is(dropErr, context.DeadlineExceeded) {
			t.Fatal(dropErr)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	t.Run("failed_drop_retains_shared_observation", func(t *testing.T) {
		held := h.NewDatabaseT(t, "drop_failure_observer")
		blocker, err := admin.BeginTx(t.Context(), nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = blocker.Rollback() })
		quoted, err := pgIdentifier(held.Name)
		if err != nil {
			t.Fatal(err)
		}
		renamed, err := pgIdentifier(held.Name + "_held")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := blocker.ExecContext(t.Context(), "ALTER DATABASE "+quoted+" RENAME TO "+renamed); err != nil {
			t.Fatal(err)
		}
		deadline, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		failure := postgrescleanup.DropOwnedDatabase(deadline, h.AdminDSN(), held.Name)
		if !errors.Is(failure, context.DeadlineExceeded) {
			t.Fatal("failed drop lost its original deadline")
		}
		var observed interface {
			PostgresCleanup() postgrescleanup.DropAttempt
		}
		if !errors.As(failure, &observed) {
			t.Fatal("failed drop lost the shared wait observation")
		}
		value := observed.PostgresCleanup().Wait
		if err := value.Validate(); err != nil {
			t.Fatal(err)
		}
		if value.Status != "sampled" || value.Last.Type != "Lock" || value.Last.Event != "object" || value.Last.Blockers != 1 {
			t.Fatalf("unexpected structural observation: %s", value.String())
		}
		if strings.Contains(value.String(), held.Name) || strings.Contains(value.String(), h.AdminDSN()) {
			t.Fatal("private drop identity escaped")
		}
		if err := blocker.Rollback(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("recovery_fixture_preserves_existing_targets_and_isolates_cleanup", func(t *testing.T) {
		existing := h.NewDatabaseT(t, "recovery_fixture_existing")
		template := os.Getenv("CARTULARY_PGTEST_TEMPLATE_DB")
		if template == "" {
			t.Fatal("owned integration fixture requires its migrated template")
		}
		var originalOID uint32
		if err := admin.QueryRowContext(t.Context(), `SELECT oid FROM pg_catalog.pg_database WHERE datname = $1`, existing.Name).Scan(&originalOID); err != nil {
			t.Fatal(err)
		}
		if created, err := createDatabase(t.Context(), h.AdminDSN(), existing.Name, template); err == nil || created {
			t.Fatal("fixture acquisition replaced an existing database")
		}
		assertExisting := func() {
			t.Helper()
			var currentOID uint32
			if err := admin.QueryRowContext(t.Context(), `SELECT oid FROM pg_catalog.pg_database WHERE datname = $1`, existing.Name).Scan(&currentOID); err != nil {
				t.Fatal(err)
			}
			if currentOID != originalOID {
				t.Fatal("fixture operation replaced its peer database")
			}
		}
		assertExisting()
		name := "recovery_fixture_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		dsn, created, err := createFromTemplateDB(t.Context(), h.AdminDSN(), name, template)
		if err != nil || !created {
			t.Fatal("clone ownership", created, err)
		}
		t.Cleanup(func() {
			if err := dropDatabase(context.Background(), h.AdminDSN(), name); err != nil {
				t.Error(err)
			}
		})
		borrower, err := sql.Open("pgx", dsn)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = borrower.Close() })
		connection, err := borrower.Conn(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = connection.Close() })
		if err := connection.PingContext(t.Context()); err != nil {
			t.Fatal(err)
		}
		// The active connection exercises atomic termination through the
		// common single bounded forced drop against only our proven-created clone.
		if err := dropDatabase(t.Context(), h.AdminDSN(), name); err != nil {
			t.Fatal(err)
		}
		if err := connection.PingContext(t.Context()); err == nil {
			t.Fatal("fixture cleanup left its target connection usable")
		}
		var exists bool
		if err := admin.QueryRowContext(t.Context(), `SELECT EXISTS (SELECT 1 FROM pg_catalog.pg_database WHERE datname = $1)`, name).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("fixture cleanup retained its target database")
		}
		assertExisting()
	})

}
