package httpapi

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func testIndicatorHTTPCompositionRejectsMissingDependencies(t *testing.T) {
	fixedTime := time.Date(2026, 8, 23, 13, 14, 15, 0, time.UTC)
	database := inertIndicatorHTTPDB{}
	owner := &indicators.Store{}
	recordEnvelopes := records.NewStore(database)
	complete := platformhttpapi.DependencySet{
		PostgresDB: database,
		Now:        func() time.Time { return fixedTime },
	}
	var (
		typedNilOwner    *indicators.Store
		typedNilRecords  *records.Store
		typedNilPostgres *inertIndicatorHTTPDB
	)
	tests := []struct {
		name    string
		deps    platformhttpapi.DependencySet
		owner   ownerApplication
		records recordEnvelopeReader
		want    string
	}{
		{name: "nil owner", deps: complete, records: recordEnvelopes, want: "owner is required"},
		{name: "typed nil owner", deps: complete, owner: typedNilOwner, records: recordEnvelopes, want: "owner is required"},
		{name: "nil RecordEnvelopes", deps: complete, owner: owner, want: "RecordEnvelopes is required"},
		{name: "typed nil RecordEnvelopes", deps: complete, owner: owner, records: typedNilRecords, want: "RecordEnvelopes is required"},
		{name: "nil Postgres", deps: platformhttpapi.DependencySet{Now: complete.Now}, owner: owner, records: recordEnvelopes, want: "Postgres is required"},
		{name: "typed nil Postgres", deps: platformhttpapi.DependencySet{PostgresDB: typedNilPostgres, Now: complete.Now}, owner: owner, records: recordEnvelopes, want: "Postgres is required"},
		{name: "nil Now", deps: platformhttpapi.DependencySet{PostgresDB: database}, owner: owner, records: recordEnvelopes, want: "Now is required"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := newService(test.deps, test.owner, test.records)
			if err == nil || !strings.Contains(err.Error(), test.want) || service != nil {
				t.Fatalf("compose result service=%#v error=%v, want nil service and %q", service, err, test.want)
			}
		})
	}

	service, err := newService(complete, owner, recordEnvelopes)
	if err != nil || service == nil {
		t.Fatalf("compose complete Indicator HTTP service: service=%#v error=%v", service, err)
	}
	if got := service.now(); !got.Equal(fixedTime) {
		t.Fatalf("injected HTTP Now returned %s, want %s", got, fixedTime)
	}
}

type inertIndicatorHTTPDB struct{}

func (inertIndicatorHTTPDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}

func (inertIndicatorHTTPDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query")
}

func (inertIndicatorHTTPDB) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow")
}

func (inertIndicatorHTTPDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	panic("unexpected BeginTx")
}
