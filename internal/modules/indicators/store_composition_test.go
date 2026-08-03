package indicators

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func TestIndicatorStoreCompositionAndRepositoryBoundaries(t *testing.T) {
	t.Parallel()

	if _, err := NewStore(StoreDependencies{}); err == nil || !strings.Contains(err.Error(), "Postgres is required") {
		t.Fatalf("missing Postgres dependency error = %v", err)
	}
	if _, err := NewStore(StoreDependencies{Postgres: inertIndicatorDB{}}); err == nil || !strings.Contains(err.Error(), "Revisions is required") {
		t.Fatalf("missing Revisions dependency error = %v", err)
	}
	if _, err := NewStore(StoreDependencies{Postgres: inertIndicatorDB{}, Revisions: &revisions.Appender{}}); err == nil || !strings.Contains(err.Error(), "Projections is required") {
		t.Fatalf("missing Projections dependency error = %v", err)
	}
	if _, err := NewStore(StoreDependencies{Postgres: inertIndicatorDB{}, Revisions: &revisions.Appender{}, Projections: inertIndicatorProjectionPort{}}); err == nil || !strings.Contains(err.Error(), "SourceText is required") {
		t.Fatalf("missing SourceText dependency error = %v", err)
	}
	owner, err := NewStore(StoreDependencies{
		Postgres:    inertIndicatorDB{},
		Revisions:   &revisions.Appender{},
		Projections: inertIndicatorProjectionPort{},
		SourceText:  inertIndicatorSourceTextPort{},
	})
	if err != nil || owner == nil {
		t.Fatalf("compose complete Indicators owner: owner=%#v err=%v", owner, err)
	}

	for _, path := range []string{
		"source_repository.go",
		"observation_repository.go",
		"lifecycle_repository.go",
	} {
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		content := string(body)
		for _, forbidden := range []string{
			"BeginTx(",
			".Commit(",
			".Rollback(",
			"AppendRevision",
			"projection",
			"idempotency",
		} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s owns forbidden workflow concern %q", path, forbidden)
			}
		}
	}
}

func TestIndicatorStoreDelegatesProjectionRefreshAndLoad(t *testing.T) {
	t.Parallel()

	recordID := uuid.MustParse("00000000-0000-4000-8000-000000000404")
	wantRow := map[string]any{"record_id": recordID.String(), "row_version": int64(4)}
	port := &recordingIndicatorProjectionPort{row: wantRow}
	store := &Store{projections: port}
	gotRow, err := store.refreshAndLoadProjectionRowTx(context.Background(), nil, recordID)
	if err != nil {
		t.Fatalf("refresh and load Indicator projection row: %v", err)
	}
	if len(port.calls) != 2 || port.calls[0] != "refresh:"+ViewSchemaID+":"+recordID.String() || port.calls[1] != "load:"+ViewSchemaID+":"+recordID.String() {
		t.Fatalf("projection delegation calls = %#v", port.calls)
	}
	if gotRow["record_id"] != wantRow["record_id"] || gotRow["row_version"] != wantRow["row_version"] {
		t.Fatalf("projection row = %#v, want %#v", gotRow, wantRow)
	}
}

type inertIndicatorDB struct{}

type inertIndicatorProjectionPort struct{}

type inertIndicatorSourceTextPort struct{}

type recordingIndicatorProjectionPort struct {
	calls []string
	row   map[string]any
}

func (port *recordingIndicatorProjectionPort) RefreshRowTx(_ context.Context, _ pgx.Tx, viewSchemaID string, recordID uuid.UUID) error {
	port.calls = append(port.calls, "refresh:"+viewSchemaID+":"+recordID.String())
	return nil
}

func (port *recordingIndicatorProjectionPort) LoadRowTx(_ context.Context, _ pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	port.calls = append(port.calls, "load:"+viewSchemaID+":"+recordID.String())
	return port.row, nil
}

func (inertIndicatorProjectionPort) RefreshRowTx(context.Context, pgx.Tx, string, uuid.UUID) error {
	panic("unexpected RefreshRowTx")
}

func (inertIndicatorProjectionPort) LoadRowTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error) {
	panic("unexpected LoadRowTx")
}

func (inertIndicatorSourceTextPort) LoadTextTx(context.Context, pgx.Tx, uuid.UUID, string, string) (SourceTextValue, error) {
	panic("unexpected LoadTextTx")
}

func (inertIndicatorSourceTextPort) LoadRowTx(context.Context, pgx.Tx, uuid.UUID, string, string) (map[string]any, error) {
	panic("unexpected LoadRowTx")
}

func (inertIndicatorSourceTextPort) RefreshAndLoadRowTx(context.Context, pgx.Tx, uuid.UUID, string, string) (map[string]any, error) {
	panic("unexpected RefreshAndLoadRowTx")
}

func (inertIndicatorDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}

func (inertIndicatorDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query")
}

func (inertIndicatorDB) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow")
}

func (inertIndicatorDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	panic("unexpected BeginTx")
}
