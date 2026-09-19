package networkflow

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type queryCompletionProbe struct {
	pgx.Tx
	graphQueryComposer
	httpauth.SessionStore
	committed, begins, computations, rollbacks int
	appendErr, commitErr, slideErr             error
	admissionErr                               error
	computeFailure                             *semanticFailure
	cancel                                     context.CancelFunc
	indeterminate                              bool
	pending                                    bool
}
type queryCompletionDB struct {
	postgres.DB
	probe *queryCompletionProbe
}

func (db *queryCompletionDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	db.probe.begins++
	return db.probe, nil
}
func (p *queryCompletionProbe) Commit(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.commitErr == nil || p.indeterminate {
		if p.pending {
			p.committed++
			p.pending = false
		}
	}
	return p.commitErr
}
func (p *queryCompletionProbe) Rollback(context.Context) error {
	p.pending = false
	p.rollbacks++
	return nil
}
func (p *queryCompletionProbe) AppendNetworkFlowEventTx(_ context.Context, _ pgx.Tx, _ *uuid.UUID, _ *uuid.UUID, event string, _ *string, correlation *string, _ any, after any) error {
	if event != "network_flow_graph_query_executed" {
		return errors.New("audit binding lost")
	}
	if after.(map[string]any)["graph_query_digest_safe"] == nil {
		return errors.New("digest lost")
	}
	p.pending = true
	return p.appendErr
}
func (p *queryCompletionProbe) Check(context.Context, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error) {
	return admission.Grant{Role: admission.RoleAdmin}, p.admissionErr
}
func (p *queryCompletionProbe) CheckTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error) {
	return admission.Grant{Role: admission.RoleAdmin}, p.admissionErr
}
func (p *queryCompletionProbe) composeGraph(context.Context, uuid.UUID, graphQueryRequest) (graphComposition, *semanticFailure) {
	p.computations++
	if p.cancel != nil {
		p.cancel()
	}
	return graphComposition{Digest: strings.Repeat("a", 64), SemanticQuery: map[string]any{}, GraphProjection: map[string]any{}}, p.computeFailure
}
func (p *queryCompletionProbe) GetSessionByFingerprint(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error) {
	now := time.Now()
	actor := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	return authn.SessionRecord{ID: uuid.New(), UserID: actor, AuthenticatedAt: now.Add(-time.Hour), LastQualifyingActivityAt: now, IdleExpiresAt: now.Add(time.Hour), AbsoluteExpiresAt: now.Add(2 * time.Hour), SessionExpiresAt: now.Add(time.Hour)}, authn.UserRecord{ID: actor, IsActive: true}, nil
}
func (p *queryCompletionProbe) SlideSession(_ context.Context, _ uuid.UUID, timing authn.SessionTiming) (authn.SessionTiming, error) {
	return timing, p.slideErr
}

type failingQueryWriter struct {
	*httptest.ResponseRecorder
	attempts int
}

func (w *failingQueryWriter) Write([]byte) (int, error) {
	w.attempts++
	return 0, errors.New("delivery lost")
}

func TestGraphQueryCompletion_Unit(t *testing.T) {
	sentinel := errors.New("SENTINEL secret failure")
	for _, name := range []string{"success retry", "admission", "compute", "limit", "cancellation", "audit rollback", "commit failure", "indeterminate commit", "session maintenance", "response delivery"} {
		t.Run(name, func(t *testing.T) {
			probe := &queryCompletionProbe{}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			want, status := 0, 500
			switch name {
			case "success retry":
				want, status = 2, 200
			case "admission":
				probe.admissionErr = &admission.Denied{Code: admission.DenialNotVisible}
				status = 404
			case "compute":
				probe.computeFailure = internalSemanticFailure(sentinel)
			case "limit":
				probe.computeFailure = graphLimitExceeded("vertex_limit_exceeded", "vertices", 1, 2)
				status = 413
			case "cancellation":
				probe.cancel = cancel
				status = 502
			case "audit rollback":
				probe.appendErr = sentinel
			case "commit failure":
				probe.commitErr = sentinel
			case "indeterminate commit":
				probe.commitErr = sentinel
				probe.indeterminate = true
				want = 1
			case "session maintenance":
				probe.slideErr = sentinel
				want = 1
			case "response delivery":
				want, status = 1, 200
			}
			store := &store{pool: &queryCompletionDB{probe: probe}, limits: defaultEffectiveLimits(), auditAppender: probe}
			app := &graphQueryApplication{store: store, incidentAccess: probe, composer: probe, safeDigester: probe}
			service := &routeService{limits: store.limits, incidentAccess: probe, authStore: probe, now: time.Now, graphQueries: app}
			attempts := 1
			if name == "success retry" {
				attempts = 2
			}
			for i := 0; i < attempts; i++ {
				r := httptest.NewRequest(http.MethodPost, "/api/v1/incidents/11111111-1111-4111-8111-111111111111/network-flow/graphs/query", strings.NewReader(`{"schema_id":"cartulary.network_flow.graph_query_request.v2","table_scope":{"mode":"all_active_tables"},"aggregation":{"mode":"default_flow_edge_v1"}}`)).WithContext(ctx)
				r.SetPathValue("incident_id", "11111111-1111-4111-8111-111111111111")
				r.AddCookie(&http.Cookie{Name: authn.SessionCookieName, Value: "session"})
				recorder := httptest.NewRecorder()
				var writer http.ResponseWriter = recorder
				broken := &failingQueryWriter{ResponseRecorder: recorder}
				if name == "response delivery" {
					writer = broken
				}
				service.handleGraphQuery(writer, r)
				if recorder.Code != status || strings.Contains(recorder.Body.String(), "SENTINEL") {
					t.Fatalf("response %d %s", recorder.Code, recorder.Body)
				}
				if name == "response delivery" && broken.attempts != 1 {
					t.Fatal("delivery path not exercised")
				}
			}
			if probe.committed != want {
				t.Fatalf("committed=%d want=%d", probe.committed, want)
			}
			if probe.begins > attempts || probe.computations > attempts {
				t.Fatal("implicit execution replay")
			}
			if (name == "admission" || name == "compute" || name == "limit" || name == "cancellation") && probe.begins != 0 {
				t.Fatal("audit began before successful computation")
			}
		})
	}
}

func (p *queryCompletionProbe) Digest(domain, value string) (string, string, error) {
	return strings.Repeat("b", 64), "test-key", nil
}
