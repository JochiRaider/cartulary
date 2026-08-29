package collaboration

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

type failingSeedRuntimeDB struct {
	postgres.DB
	err error
}

func (db failingSeedRuntimeDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, db.err
}

func runtimeConstructionCatalog(t testing.TB) *PublicationCatalog {
	t.Helper()
	catalog, err := NewPublicationCatalog([]PublicationContribution{{
		ContributionID: "test.runtime.v1", SourceOwnerID: "test.runtime",
		AffectedViews: []ViewPublicationContribution{{
			ViewSchemaID: "cartulary.view.runtime.v1", RecordTypes: []string{"test.runtime"},
			PublicFieldKeys: []string{"runtime.value"}, PatchFieldKeys: []string{"runtime.value"},
		}},
	}}, []CanonicalPublicationView{{
		ViewSchemaID: "cartulary.view.runtime.v1", SourceOwnerID: "test.runtime", RecordTypes: []string{"test.runtime"},
	}})
	if err != nil {
		t.Fatalf("construct Runtime publication catalog: %v", err)
	}
	return catalog
}

func runtimeConstructionOptions(t testing.TB, db postgres.DB, onLoss func()) Options {
	t.Helper()
	return Options{
		Postgres: db,
		AcceptSocket: func(http.ResponseWriter, *http.Request) (protocol.Socket, error) {
			return nil, errors.New("unused test socket")
		},
		CheckBrowserOrigin:         func(http.ResponseWriter, *http.Request) bool { return true },
		ServiceVersion:             "1.2.3-test",
		Now:                        func() time.Time { return time.Date(2026, 8, 28, 19, 0, 0, 0, time.UTC) },
		PublicationCatalog:         runtimeConstructionCatalog(t),
		OnUnexpectedDispatcherLoss: onLoss,
	}
}

func TestRuntimeConstructionRejectsEveryMissingDependency_Unit(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Options)
		want string
	}{
		{name: "Postgres", edit: func(options *Options) { options.Postgres = nil }, want: "PostgreSQL"},
		{name: "socket accept", edit: func(options *Options) { options.AcceptSocket = nil }, want: "accept"},
		{name: "Origin check", edit: func(options *Options) { options.CheckBrowserOrigin = nil }, want: "Origin"},
		{name: "publication catalog", edit: func(options *Options) { options.PublicationCatalog = nil }, want: "catalog"},
		{name: "dispatcher loss", edit: func(options *Options) { options.OnUnexpectedDispatcherLoss = nil }, want: "loss callback"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := runtimeConstructionOptions(t, failingSeedRuntimeDB{}, func() {})
			test.edit(&options)
			runtime, err := NewRuntime(options)
			if err == nil || runtime != nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Runtime = %#v, error = %v; want missing dependency error containing %q", runtime, err, test.want)
			}
		})
	}
}

func TestRuntimeConstructionResolvesOneTelemetryVersion_Unit(t *testing.T) {
	runtime, err := NewRuntime(runtimeConstructionOptions(t, failingSeedRuntimeDB{}, func() {}))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.hub.serviceVersion != "1.2.3-test" || runtime.routes.serviceVersion != "1.2.3-test" {
		t.Fatalf("Collaboration service versions = hub:%q route:%q", runtime.hub.serviceVersion, runtime.routes.serviceVersion)
	}

	options := runtimeConstructionOptions(t, failingSeedRuntimeDB{}, func() {})
	options.ServiceVersion = " "
	unknownRuntime, err := NewRuntime(options)
	if err != nil {
		t.Fatal(err)
	}
	if unknownRuntime.hub.serviceVersion != telemetry.VersionUnknown || unknownRuntime.routes.serviceVersion != telemetry.VersionUnknown {
		t.Fatalf("default service versions = hub:%q route:%q", unknownRuntime.hub.serviceVersion, unknownRuntime.routes.serviceVersion)
	}
}

func TestRuntimeDispatcherStartupLossIsTerminalAndReportedOnce_Unit(t *testing.T) {
	var losses atomic.Int32
	runtime, err := NewRuntime(runtimeConstructionOptions(t, failingSeedRuntimeDB{err: errors.New("seed unavailable")}, func() {
		losses.Add(1)
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "seed collaboration replay tail cursor") {
		t.Fatalf("first dispatcher start error = %v", err)
	}
	if err := runtime.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "terminal") {
		t.Fatalf("terminal dispatcher restart error = %v", err)
	}
	runtime.reportUnexpectedDispatcherLoss()
	runtime.reportUnexpectedDispatcherLoss()
	if got := losses.Load(); got != 1 {
		t.Fatalf("unexpected dispatcher loss reports = %d want 1", got)
	}
	if err := runtime.Close(context.Background()); err != nil {
		t.Fatalf("close terminal dispatcher: %v", err)
	}
}
