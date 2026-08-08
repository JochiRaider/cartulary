package database_migrations

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/pressly/goose/v3"
)

func TestApplyCanceledContextSkipsSourceInspection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	status, err := Apply(ctx, nil, NewMigrationSource("/path/that/does/not/exist"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation error, got status=%#v err=%v", status, err)
	}
	if status.SourceName != "/path/that/does/not/exist" || status.Empty {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestProviderContextSourceLoggerIsolation(t *testing.T) {
	t.Parallel()

	type providerCase struct {
		name      string
		version   int64
		migration string
		marker    string
	}
	cases := []providerCase{
		{name: "alpha", version: 1, migration: "migrations/00001_alpha.sql", marker: "alpha-only"},
		{name: "beta", version: 2, migration: "migrations/00002_beta.sql", marker: "beta-only"},
	}

	var wait sync.WaitGroup
	for _, testCase := range cases {
		testCase := testCase
		wait.Add(1)
		go func() {
			defer wait.Done()

			source := MigrationSource{
				BaseFS: fstest.MapFS{
					testCase.migration: &fstest.MapFile{Data: []byte("-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n")},
				},
				Path: "migrations",
			}
			providerFS, err := migrationProviderFS(source)
			if err != nil {
				t.Errorf("%s: resolve source: %v", testCase.name, err)
				return
			}
			logPath := filepath.Join(t.TempDir(), testCase.name, "goose.log")
			logger, closer, err := newGooseLogger(logPath)
			if err != nil {
				t.Errorf("%s: create logger: %v", testCase.name, err)
				return
			}
			provider, err := newGooseProvider(&sql.DB{}, providerFS, logger)
			if err != nil {
				_ = closer.Close()
				t.Errorf("%s: create provider: %v", testCase.name, err)
				return
			}
			sources := provider.ListSources()
			if len(sources) != 1 || sources[0].Version != testCase.version {
				_ = closer.Close()
				t.Errorf("%s: provider source leak: %#v", testCase.name, sources)
				return
			}
			logger.Printf("%s", testCase.marker)
			if err := closer.Close(); err != nil {
				t.Errorf("%s: close logger: %v", testCase.name, err)
				return
			}
			body, err := os.ReadFile(logPath)
			if err != nil {
				t.Errorf("%s: read logger: %v", testCase.name, err)
				return
			}
			if !strings.Contains(string(body), testCase.marker) {
				t.Errorf("%s: marker absent from isolated logger: %q", testCase.name, body)
			}
		}()
	}
	wait.Wait()
}

func TestMigrationProviderFSRootsEmbeddedSources(t *testing.T) {
	sourceFS, err := migrationProviderFS(MigrationSource{
		BaseFS: fstest.MapFS{
			"nested/00001_test.sql": &fstest.MapFile{Data: []byte("-- +goose Up\nSELECT 1;\n")},
		},
		Path: "nested",
	})
	if err != nil {
		t.Fatalf("resolve embedded source: %v", err)
	}
	entries, err := fs.ReadDir(sourceFS, ".")
	if err != nil {
		t.Fatalf("read rooted source: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "00001_test.sql" {
		t.Fatalf("unexpected rooted source entries: %#v", entries)
	}
}

var _ goose.Logger = (*logProbe)(nil)

type logProbe struct{}

func (*logProbe) Fatalf(string, ...any) {}
func (*logProbe) Printf(string, ...any) {}
