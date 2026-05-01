package postgres

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"
)

func TestMigrateCanceledContextSkipsInspectionAndGoose(t *testing.T) {
	oldRunContext := gooseRunContext
	t.Cleanup(func() {
		gooseRunContext = oldRunContext
	})

	gooseCalls := 0
	gooseRunContext = func(ctx context.Context, command string, db *sql.DB, dir string, args ...string) error {
		gooseCalls++
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	status, err := Migrate(ctx, nil, NewMigrationSource("/path/that/does/not/exist"), "up")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation error, got status=%#v err=%v", status, err)
	}
	if status.Command != "up" || status.Directory != "/path/that/does/not/exist" || status.Empty {
		t.Fatalf("unexpected status: %#v", status)
	}
	if gooseCalls != 0 {
		t.Fatalf("expected goose not to run, got %d calls", gooseCalls)
	}
}

func TestRunGooseEmbeddedCanceledContextSkipsBaseFSAndGoose(t *testing.T) {
	oldRunContext := gooseRunContext
	oldSetBaseFS := gooseSetBaseFS
	t.Cleanup(func() {
		gooseRunContext = oldRunContext
		gooseSetBaseFS = oldSetBaseFS
	})

	gooseCalls := 0
	setBaseFSCalls := 0
	gooseRunContext = func(ctx context.Context, command string, db *sql.DB, dir string, args ...string) error {
		gooseCalls++
		return nil
	}
	gooseSetBaseFS = func(fsys fs.FS) {
		setBaseFSCalls++
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runGoose(ctx, "up", nil, MigrationSource{
		BaseFS: fstest.MapFS{"migrations/00001_test.sql": &fstest.MapFile{Data: []byte("-- +goose Up\nSELECT 1;\n")}},
		Path:   "migrations",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation error, got %v", err)
	}
	if setBaseFSCalls != 0 {
		t.Fatalf("expected BaseFS not to be set, got %d calls", setBaseFSCalls)
	}
	if gooseCalls != 0 {
		t.Fatalf("expected goose not to run, got %d calls", gooseCalls)
	}
}

func TestRunGooseEmbeddedCanceledContextReturnsWhileGuardHeld(t *testing.T) {
	oldRunContext := gooseRunContext
	oldSetBaseFS := gooseSetBaseFS
	t.Cleanup(func() {
		gooseRunContext = oldRunContext
		gooseSetBaseFS = oldSetBaseFS
	})

	gooseRunContext = func(ctx context.Context, command string, db *sql.DB, dir string, args ...string) error {
		t.Fatal("goose should not run while the BaseFS guard is held")
		return nil
	}
	gooseSetBaseFS = func(fsys fs.FS) {
		t.Fatal("BaseFS should not be set while the guard is held")
	}

	gooseBaseFSSem <- struct{}{}
	t.Cleanup(func() {
		select {
		case <-gooseBaseFSSem:
		default:
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runGoose(ctx, "up", nil, MigrationSource{
			BaseFS: fstest.MapFS{"migrations/00001_test.sql": &fstest.MapFile{Data: []byte("-- +goose Up\nSELECT 1;\n")}},
			Path:   "migrations",
		})
	}()

	time.AfterFunc(10*time.Millisecond, cancel)

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected cancellation error, got %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("runGoose did not return promptly after context cancellation")
	}
}

func TestRunGoosePassesContextToGoose(t *testing.T) {
	oldRunContext := gooseRunContext
	t.Cleanup(func() {
		gooseRunContext = oldRunContext
	})

	type markerKey struct{}
	ctx := context.WithValue(context.Background(), markerKey{}, "marker")

	var gotMarker any
	gooseRunContext = func(ctx context.Context, command string, db *sql.DB, dir string, args ...string) error {
		gotMarker = ctx.Value(markerKey{})
		return context.Canceled
	}

	err := runGoose(ctx, "up", nil, NewMigrationSource("."))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected goose error to be returned, got %v", err)
	}
	if gotMarker != "marker" {
		t.Fatalf("goose did not receive caller context marker: got %#v", gotMarker)
	}
}
