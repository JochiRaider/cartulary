package indicators

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestIndicatorApplicationCompositionAndRepositoryBoundaries(t *testing.T) {
	t.Parallel()
	t.Run("deployed replay preimages and digests", testIndicatorReplayHashCompatibility)

	fixedTime := time.Date(2026, 8, 23, 12, 34, 56, 789, time.UTC)
	complete := ApplicationDependencies{
		Postgres:        inertIndicatorDB{},
		Idempotency:     &inertIndicatorIdempotencyPort{},
		IncidentState:   &inertIndicatorIncidentStatePort{},
		Revisions:       &revisions.Appender{},
		RecordEnvelopes: records.NewStore(inertIndicatorDB{}),
		Projections:     inertIndicatorProjectionPort{},
		SourceText:      inertIndicatorSourceTextPort{},
		Collaboration:   &inertIndicatorPublicationPort{},
		Clock:           func() time.Time { return fixedTime },
	}
	var (
		typedNilPostgres      *inertIndicatorDB
		typedNilIdempotency   *inertIndicatorIdempotencyPort
		typedNilIncident      *inertIndicatorIncidentStatePort
		typedNilRevisions     *revisions.Appender
		typedNilRecords       *records.Store
		typedNilProjection    *inertIndicatorProjectionPort
		typedNilSourceText    *inertIndicatorSourceTextPort
		typedNilCollaboration *inertIndicatorPublicationPort
	)
	without := func(name string, value any) ApplicationDependencies {
		deps := complete
		switch name {
		case "Postgres":
			deps.Postgres, _ = value.(postgres.DB)
		case "Idempotency":
			deps.Idempotency, _ = value.(IdempotencyPort)
		case "IncidentState":
			deps.IncidentState, _ = value.(IncidentStatePort)
		case "Revisions":
			deps.Revisions, _ = value.(RevisionPort)
		case "RecordEnvelopes":
			deps.RecordEnvelopes, _ = value.(RecordEnvelopePort)
		case "Projections":
			deps.Projections, _ = value.(indicatorprojection.Rows)
		case "SourceText":
			deps.SourceText, _ = value.(SourceTextPort)
		case "Collaboration":
			deps.Collaboration, _ = value.(collaboration.RecordChangedAppender)
		case "Clock":
			deps.Clock, _ = value.(func() time.Time)
		}
		return deps
	}
	tests := []struct {
		name string
		deps ApplicationDependencies
		want string
	}{
		{name: "nil Postgres", deps: without("Postgres", nil), want: "Postgres is required"},
		{name: "typed nil Postgres", deps: without("Postgres", typedNilPostgres), want: "Postgres is required"},
		{name: "nil Idempotency", deps: without("Idempotency", nil), want: "Idempotency is required"},
		{name: "typed nil Idempotency", deps: without("Idempotency", typedNilIdempotency), want: "Idempotency is required"},
		{name: "nil IncidentState", deps: without("IncidentState", nil), want: "IncidentState is required"},
		{name: "typed nil IncidentState", deps: without("IncidentState", typedNilIncident), want: "IncidentState is required"},
		{name: "nil Revisions", deps: without("Revisions", nil), want: "Revisions is required"},
		{name: "typed nil Revisions", deps: without("Revisions", typedNilRevisions), want: "Revisions is required"},
		{name: "nil RecordEnvelopes", deps: without("RecordEnvelopes", nil), want: "RecordEnvelopes is required"},
		{name: "typed nil RecordEnvelopes", deps: without("RecordEnvelopes", typedNilRecords), want: "RecordEnvelopes is required"},
		{name: "nil Projections", deps: without("Projections", nil), want: "Projections is required"},
		{name: "typed nil Projections", deps: without("Projections", typedNilProjection), want: "Projections is required"},
		{name: "nil SourceText", deps: without("SourceText", nil), want: "SourceText is required"},
		{name: "typed nil SourceText", deps: without("SourceText", typedNilSourceText), want: "SourceText is required"},
		{name: "nil Collaboration", deps: without("Collaboration", nil), want: "Collaboration is required"},
		{name: "typed nil Collaboration", deps: without("Collaboration", typedNilCollaboration), want: "Collaboration is required"},
		{name: "nil Clock", deps: without("Clock", nil), want: "Clock is required"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			owner, err := NewApplication(test.deps)
			if err == nil || !strings.Contains(err.Error(), test.want) || owner != nil {
				t.Fatalf("compose result owner=%#v error=%v, want nil owner and %q", owner, err, test.want)
			}
		})
	}
	owner, err := NewApplication(complete)
	if err != nil || owner == nil {
		t.Fatalf("compose complete Indicators owner: owner=%#v err=%v", owner, err)
	}
	if got := owner.now(); !got.Equal(fixedTime) {
		t.Fatalf("injected Clock returned %s, want %s", got, fixedTime)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(loadIndicatorByDedupeSQL))); got != "d665f06c2526b0118e33eaa887da279ad54025967618662c2a0b47b7bfde857b" {
		t.Fatalf("canonical dedupe SQL digest = %s", got)
	}
	if _, statErr := os.Stat("repositories.go"); !os.IsNotExist(statErr) {
		t.Fatalf("obsolete repository namespace file remains: %v", statErr)
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

type inertIndicatorPublicationPort struct{}

func (*inertIndicatorPublicationPort) AppendRecordChangedTx(context.Context, pgx.Tx, collaboration.RecordChangeIntentInput) error {
	return nil
}

func TestIndicatorApplicationDelegatesProjectionRefreshAndLoad(t *testing.T) {
	t.Parallel()

	recordID := uuid.MustParse("00000000-0000-4000-8000-000000000404")
	wantRow := map[string]any{"record_id": recordID.String(), "row_version": int64(4)}
	port := &recordingIndicatorProjectionPort{row: wantRow}
	application := &Application{projections: port}
	gotRow, err := application.refreshAndLoadProjectionRowTx(context.Background(), nil, recordID)
	if err != nil {
		t.Fatalf("refresh and load Indicator projection row: %v", err)
	}
	if len(port.calls) != 2 || port.calls[0] != "refresh:"+recordID.String() || port.calls[1] != "load:"+recordID.String() {
		t.Fatalf("projection delegation calls = %#v", port.calls)
	}
	if gotRow["record_id"] != wantRow["record_id"] || gotRow["row_version"] != wantRow["row_version"] {
		t.Fatalf("projection row = %#v, want %#v", gotRow, wantRow)
	}
}

type inertIndicatorDB struct{}

type inertIndicatorProjectionPort struct{}

type inertIndicatorSourceTextPort struct{}

type inertIndicatorIdempotencyPort struct{}

type inertIndicatorIncidentStatePort struct{}

func (*inertIndicatorIdempotencyPort) GetRouteIdempotency(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
	panic("unexpected GetRouteIdempotency")
}

func (*inertIndicatorIdempotencyPort) InsertRouteIdempotencyPayload(context.Context, pgx.Tx, authn.RouteIdempotencyKey, []byte, int, any) error {
	panic("unexpected InsertRouteIdempotencyPayload")
}

func (*inertIndicatorIncidentStatePort) RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error {
	panic("unexpected RequireOpenTx")
}

type recordingIndicatorProjectionPort struct {
	calls []string
	row   map[string]any
}

func (port *recordingIndicatorProjectionPort) RefreshIndicatorTx(_ context.Context, _ pgx.Tx, recordID uuid.UUID) error {
	port.calls = append(port.calls, "refresh:"+recordID.String())
	return nil
}

func (port *recordingIndicatorProjectionPort) LoadIndicatorTx(_ context.Context, _ pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	port.calls = append(port.calls, "load:"+recordID.String())
	return port.row, nil
}

func (inertIndicatorProjectionPort) RefreshIndicatorTx(context.Context, pgx.Tx, uuid.UUID) error {
	panic("unexpected RefreshIndicatorTx")
}

func (inertIndicatorProjectionPort) LoadIndicatorTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error) {
	panic("unexpected LoadIndicatorTx")
}

func (inertIndicatorProjectionPort) DeleteIndicatorTx(context.Context, pgx.Tx, uuid.UUID) error {
	panic("unexpected DeleteIndicatorTx")
}

func (inertIndicatorProjectionPort) RebuildIndicatorsTx(context.Context, pgx.Tx, uuid.UUID) error {
	panic("unexpected RebuildIndicatorsTx")
}

func (port *recordingIndicatorProjectionPort) DeleteIndicatorTx(context.Context, pgx.Tx, uuid.UUID) error {
	panic("unexpected DeleteIndicatorTx")
}

func (port *recordingIndicatorProjectionPort) RebuildIndicatorsTx(context.Context, pgx.Tx, uuid.UUID) error {
	panic("unexpected RebuildIndicatorsTx")
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
