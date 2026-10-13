package recoverybrowsertest

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

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
	dropCtx, cancelDrop := context.WithCancel(ctx)
	stopped := make(chan struct{})
	var dropErr error
	go func() {
		defer close(stopped)
		dropErr = postgrescleanup.ForceDropDatabase(dropCtx, h.AdminDSN(), database.Name)
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
	observationCtx, cancelObservation := context.WithTimeout(ctx, 2200*time.Millisecond)
	observation := observeDatabaseDrop(observationCtx, h.AdminDSN(), database.Name)
	cancelObservation()
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
		if dropErr != nil {
			t.Fatal(dropErr)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
