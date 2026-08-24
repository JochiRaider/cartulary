package httpapi

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func testIndicatorHTTPCompositionRejectsMissingDependencies(t *testing.T) {
	fixedTime := time.Date(2026, 8, 23, 13, 14, 15, 0, time.UTC)
	database := inertIndicatorHTTPDB{}
	owner := &indicators.Application{}
	recordEnvelopes := records.NewStore(database)
	incidents := admission.NewChecker(database)
	sessions := authn.NewStore(database)
	complete := platformhttpapi.DependencySet{
		Now: func() time.Time { return fixedTime },
	}
	var (
		typedNilOwner     *indicators.Application
		typedNilRecords   *records.Store
		typedNilIncidents *admission.Checker
		typedNilSessions  *authn.Store
	)
	tests := []struct {
		name      string
		deps      platformhttpapi.DependencySet
		owner     ownerApplication
		records   recordEnvelopeReader
		incidents incidentAdmission
		sessions  sessionStoreSlider
		want      string
	}{
		{name: "nil owner", deps: complete, records: recordEnvelopes, incidents: incidents, sessions: sessions, want: "owner is required"},
		{name: "typed nil owner", deps: complete, owner: typedNilOwner, records: recordEnvelopes, incidents: incidents, sessions: sessions, want: "owner is required"},
		{name: "nil RecordEnvelopes", deps: complete, owner: owner, incidents: incidents, sessions: sessions, want: "RecordEnvelopes is required"},
		{name: "typed nil RecordEnvelopes", deps: complete, owner: owner, records: typedNilRecords, incidents: incidents, sessions: sessions, want: "RecordEnvelopes is required"},
		{name: "nil IncidentAdmission", deps: complete, owner: owner, records: recordEnvelopes, sessions: sessions, want: "IncidentAdmission is required"},
		{name: "typed nil IncidentAdmission", deps: complete, owner: owner, records: recordEnvelopes, incidents: typedNilIncidents, sessions: sessions, want: "IncidentAdmission is required"},
		{name: "nil Sessions", deps: complete, owner: owner, records: recordEnvelopes, incidents: incidents, want: "Sessions is required"},
		{name: "typed nil Sessions", deps: complete, owner: owner, records: recordEnvelopes, incidents: incidents, sessions: typedNilSessions, want: "Sessions is required"},
		{name: "nil Now", deps: platformhttpapi.DependencySet{}, owner: owner, records: recordEnvelopes, incidents: incidents, sessions: sessions, want: "Now is required"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := newService(test.deps, test.owner, test.records, test.incidents, test.sessions)
			if err == nil || !strings.Contains(err.Error(), test.want) || service != nil {
				t.Fatalf("compose result service=%#v error=%v, want nil service and %q", service, err, test.want)
			}
		})
	}

	service, err := newService(complete, owner, recordEnvelopes, incidents, sessions)
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
