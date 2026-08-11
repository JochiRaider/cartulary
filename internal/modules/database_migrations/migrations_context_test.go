package database_migrations

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/JochiRaider/cartulary/internal/modules/database_migrations/sourcecatalog"
)

func TestApplyCanceledContextSkipsDatabaseAccess(t *testing.T) {
	source, err := buildSource(
		fstest.MapFS{"00001_valid.sql": &fstest.MapFile{Data: []byte(validMigrationBody)}},
		".",
		"lineage",
		"boundary",
	)
	if err != nil {
		t.Fatalf("construct source: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := Apply(ctx, nil, source); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation error, got %v", err)
	}
}

func TestProviderSourceIsolation(t *testing.T) {
	t.Parallel()

	type providerCase struct {
		name      string
		migration string
	}
	cases := []providerCase{
		{name: "alpha", migration: "migrations/00001_alpha.sql"},
		{name: "beta", migration: "migrations/00001_beta.sql"},
	}

	var wait sync.WaitGroup
	for _, testCase := range cases {
		testCase := testCase
		wait.Add(1)
		go func() {
			defer wait.Done()

			source, err := buildSource(
				fstest.MapFS{testCase.migration: &fstest.MapFile{Data: []byte(validMigrationBody)}},
				"migrations",
				testCase.name+"-lineage",
				testCase.name+"-boundary",
			)
			if err != nil {
				t.Errorf("%s: construct source: %v", testCase.name, err)
				return
			}
			locker, err := sourcecatalog.NewSessionLocker()
			if err != nil {
				t.Errorf("%s: create provider locker: %v", testCase.name, err)
				return
			}
			provider, err := sourcecatalog.NewProvider(&sql.DB{}, source, locker)
			if err != nil {
				t.Errorf("%s: create provider: %v", testCase.name, err)
				return
			}
			sources := provider.ListSources()
			if len(sources) != 1 || sources[0].Version != 1 {
				t.Errorf("%s: provider source leak: %#v", testCase.name, sources)
			}
		}()
	}
	wait.Wait()
}

func TestMigrationProviderUsesImmutableRoot(t *testing.T) {
	source, err := buildSource(
		fstest.MapFS{"nested/00001_test.sql": &fstest.MapFile{Data: []byte(validMigrationBody)}},
		"nested",
		"lineage",
		"boundary",
	)
	if err != nil {
		t.Fatalf("construct source: %v", err)
	}
	inspection, err := InspectSource(source)
	if err != nil {
		t.Fatalf("inspect rooted source: %v", err)
	}
	if len(inspection.Entries) != 1 || inspection.Entries[0].Filename != "00001_test.sql" {
		t.Fatalf("unexpected rooted source entries: %#v", inspection.Entries)
	}
}
