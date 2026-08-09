package httptestx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	stdhttptest "net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	projectiontestsupport "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

type Server struct {
	Config configassembly.Deployment
	HTTP   *stdhttptest.Server
	Clock  *httpapi.TestClock

	runtime       *server.Runtime
	jobs          *JobsCapability
	collaboration *CollaborationCapability
	revisions     *RevisionsCapability
	projections   *ProjectionCapability
	indicatorText indicators.SourceTextPort
}

type JobsCapability struct {
	manager      *jobs.Manager
	transactions *jobs.TransactionService
	pool         *pgxpool.Pool
	now          func() time.Time
}

func (capability *JobsCapability) Create(ctx context.Context, params jobs.EnqueueParams) (jobs.Resource, error) {
	if capability == nil || capability.transactions == nil || capability.pool == nil || capability.now == nil {
		return jobs.Resource{}, errors.New("jobs create capability is unavailable")
	}
	tx, err := capability.pool.Begin(ctx)
	if err != nil {
		return jobs.Resource{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	resource, err := capability.transactions.CreateQueuedTx(ctx, tx, params, capability.now())
	if err != nil {
		return jobs.Resource{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return jobs.Resource{}, err
	}
	return resource, nil
}

type CollaborationCapability struct {
	hub        *collaboration.Hub
	dispatcher *collaboration.Dispatcher
	intents    collaboration.IntentAppender
}

func (capability *CollaborationCapability) SubscribeIncident(incidentID uuid.UUID, buffer int) (<-chan collaboration.Message, func()) {
	if capability == nil || capability.hub == nil {
		return nil, func() {}
	}
	return capability.hub.SubscribeIncident(incidentID, buffer)
}

func (capability *CollaborationCapability) RevokeSession(sessionID uuid.UUID, reasonCode string) {
	if capability == nil || capability.hub == nil {
		return
	}
	capability.hub.RevokeSession(sessionID, reasonCode)
}

func (capability *CollaborationCapability) CloseDispatcher(ctx context.Context) error {
	if capability == nil || capability.dispatcher == nil {
		return errors.New("collaboration dispatcher capability is unavailable")
	}
	return capability.dispatcher.Close(ctx)
}

func (capability *CollaborationCapability) NewDispatcher(pool *pgxpool.Pool, now func() time.Time) *collaboration.Dispatcher {
	if capability == nil || capability.hub == nil || pool == nil {
		return nil
	}
	return collaboration.NewDispatcher(pool, capability.hub, now)
}

func (capability *CollaborationCapability) IntentAppender() collaboration.IntentAppender {
	if capability == nil {
		return nil
	}
	return capability.intents
}

type RevisionsCapability struct {
	runtime *revisionassembly.Runtime
}

func (capability *RevisionsCapability) Appender() *revisions.Appender {
	if capability == nil || capability.runtime == nil {
		return nil
	}
	return capability.runtime.Appender()
}

type ProjectionCapability = projectiontestsupport.Capability

func (s *Server) JobsCapability() *JobsCapability {
	if s == nil {
		return nil
	}
	return s.jobs
}

func (s *Server) CollaborationCapability() *CollaborationCapability {
	if s == nil {
		return nil
	}
	return s.collaboration
}

func (s *Server) RevisionsCapability() *RevisionsCapability {
	if s == nil {
		return nil
	}
	return s.revisions
}

func (s *Server) ProjectionCapability() *ProjectionCapability {
	if s == nil {
		return nil
	}
	return s.projections
}

func (s *Server) IndicatorSourceTextPort() indicators.SourceTextPort {
	if s == nil {
		return nil
	}
	return s.indicatorText
}

type TestRouteMode string

const (
	TestRouteModeHarnessOwned TestRouteMode = "harness_owned"
	TestRouteModeDisabled     TestRouteMode = "disabled"
	TestRouteModeCustomEnv    TestRouteMode = "custom_env"
)

type ServerOptions struct {
	Loaded           *configassembly.Loaded
	Env              map[string]string
	Dependencies     httpapi.DependencySet
	AdditionalRoutes []httpapi.RouteRegistrar
	Postgres         *pgxpool.Pool
	ObjectStore      objectstore.Store
	TestRouteMode    TestRouteMode
}

type AuthCookies = authcookietest.AuthCookies

const TestRouteToken = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"

func StartServer(t testing.TB, options ServerOptions) *Server {
	t.Helper()

	env := make(map[string]string, len(options.Env)+3)
	for key, value := range options.Env {
		env[key] = value
	}
	if err := applyTestRouteMode(env, options.TestRouteMode); err != nil {
		t.Fatalf("configure test routes: %v", err)
	}
	configtest.EnsureRevisionsConflictTokenTestEnvironment(env)

	var loaded configassembly.Loaded
	if options.Loaded == nil {
		tempRoots := configtest.SetupTempRoots(t)
		for key, value := range tempRoots.Paths {
			if _, exists := env[key]; !exists {
				env[key] = value
			}
		}
		configtest.BindPostgresEnvToDatabaseRoot(t, tempRoots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], env)
		if _, exists := env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"]; !exists {
			env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")
		}
		loaded = configtest.LoadFixture(t, []string{"config", "valid.toml"}, env)
	} else {
		loaded = *options.Loaded
	}
	cfg := loaded.Deployment()

	clock := httpapi.NewTestClock()
	routes := append([]httpapi.RouteRegistrar{RegisterBootstrapRoutes(), httpapi.RegisterTestClockRoutes(clock)}, options.AdditionalRoutes...)
	var (
		jobsCapability          *JobsCapability
		collaborationCapability *CollaborationCapability
		revisionsCapability     *RevisionsCapability
		projectionCapability    *ProjectionCapability
		indicatorSourceText     indicators.SourceTextPort
	)
	runtime, err := server.NewRuntime(context.Background(), loaded, server.Options{
		Env:         env,
		Now:         clock.Now,
		Postgres:    options.Postgres,
		ObjectStore: options.ObjectStore,
		ObserveJobs: func(manager *jobs.Manager, transactions *jobs.TransactionService, _ *jobs.Runner, pool *pgxpool.Pool) {
			jobsCapability = &JobsCapability{manager: manager, transactions: transactions, pool: pool, now: clock.Now}
		},
		ObserveCollaboration: func(hub *collaboration.Hub, dispatcher *collaboration.Dispatcher, intents collaboration.IntentAppender) {
			collaborationCapability = &CollaborationCapability{hub: hub, dispatcher: dispatcher, intents: intents}
		},
		ObserveProjections: func(runtime *projectionassembly.Runtime) {
			projectionCapability = projectiontestsupport.New(projectiontestsupport.Dependencies{
				TimelineRebuilder:  runtime.TimelinePorts().Rebuilder,
				EntityRebuilder:    runtime.EntityPorts().Rebuilder,
				IndicatorRebuilder: runtime.IndicatorPorts().Rebuilder,
				IndicatorRows:      runtime.IndicatorPorts().Rows,
				EvidenceRows:       runtime.EvidencePorts().Rows,
			})
			indicatorSourceText = indicatorassembly.NewSourceTextPort(runtime.SourceTextRows())
		},
		ObserveRevisions: func(runtime *revisionassembly.Runtime) {
			revisionsCapability = &RevisionsCapability{runtime: runtime}
		},
		HTTP: httpapi.Options{
			Dependencies:     options.Dependencies,
			AdditionalRoutes: routes,
		},
	})
	if err != nil {
		t.Fatalf("start app runtime: %v", err)
	}
	if err := runtime.ActivatePublication(); err != nil {
		runtime.Close()
		t.Fatalf("activate app runtime publication: %v", err)
	}

	server := &Server{
		Config:        cfg,
		HTTP:          stdhttptest.NewServer(runtime.HTTPHandler()),
		Clock:         clock,
		runtime:       runtime,
		jobs:          jobsCapability,
		collaboration: collaborationCapability,
		revisions:     revisionsCapability,
		projections:   projectionCapability,
		indicatorText: indicatorSourceText,
	}
	t.Cleanup(func() {
		server.Close()
	})

	return server
}

func applyTestRouteMode(env map[string]string, mode TestRouteMode) error {
	switch mode {
	case TestRouteModeHarnessOwned:
		if _, exists := env[httpapi.TestRoutesEnabledEnv]; !exists {
			env[httpapi.TestRoutesEnabledEnv] = "1"
		}
		if _, exists := env[httpapi.TestRuntimeMarkerEnv]; !exists {
			env[httpapi.TestRuntimeMarkerEnv] = httpapi.TestRuntimeMarkerValue
		}
		if _, exists := env[httpapi.TestRouteTokenEnv]; !exists {
			env[httpapi.TestRouteTokenEnv] = TestRouteToken
		}
	case TestRouteModeDisabled:
		delete(env, httpapi.TestRoutesEnabledEnv)
		delete(env, httpapi.TestRuntimeMarkerEnv)
		delete(env, httpapi.TestRouteTokenEnv)
		delete(env, httpapi.TestRuntimeAPIOriginEnv)
		delete(env, httpapi.TestRuntimePublicOriginEnv)
	case TestRouteModeCustomEnv:
	default:
		return fmt.Errorf("test route mode must be explicit and valid, got %q", mode)
	}
	return nil
}

func (s *Server) Close() {
	if s == nil {
		return
	}
	if s.HTTP != nil {
		s.HTTP.Close()
	}
	if s.runtime != nil {
		s.runtime.Close()
	}
}

func SetClockFixed(t testing.TB, server *Server, now time.Time) time.Time {
	t.Helper()
	if server == nil || server.Clock == nil {
		t.Fatal("test server clock is required")
	}
	return server.Clock.SetFixed(now)
}

func SetClockAfter(t testing.TB, server *Server, baseline time.Time, delta time.Duration) time.Time {
	t.Helper()
	return SetClockFixed(t, server, baseline.UTC().Add(delta))
}

func AdvanceClock(t testing.TB, server *Server, delta time.Duration) time.Time {
	t.Helper()
	if server == nil || server.Clock == nil {
		t.Fatal("test server clock is required")
	}
	return server.Clock.Advance(delta)
}

func NewJSONRequest(t testing.TB, method string, url string, body any) *http.Request {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func NewAuthenticatedJSONRequest(t testing.TB, method string, url string, token string, body any) *http.Request {
	t.Helper()

	req := NewJSONRequest(t, method, url, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func Do(t testing.TB, client *http.Client, req *http.Request) *http.Response {
	t.Helper()

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func DoJSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := NewJSONRequest(t, method, url, body)
	for _, option := range options {
		option(req)
	}
	return Do(t, http.DefaultClient, req)
}

func DoRawJSON(t testing.TB, method string, url string, body string, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("create raw JSON request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, option := range options {
		option(req)
	}
	return Do(t, http.DefaultClient, req)
}

func WithCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}
	}
}

func WithHeader(key string, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func ReadBodyString(t testing.TB, body io.ReadCloser) string {
	t.Helper()
	defer body.Close()

	payload, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return string(payload)
}

func ReadJSONBody(t testing.TB, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()

	var payload map[string]any
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return payload
}

func RequireStatus(t testing.TB, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("unexpected status: got %d want %d; read body: %v", resp.StatusCode, want, err)
		}
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(body))
		t.Fatalf("unexpected status: got %d want %d body=%s", resp.StatusCode, want, body)
	}
}

func RequireRequestID(t testing.TB, resp *http.Response) string {
	t.Helper()
	requestID := resp.Header.Get(httpapi.RequestIDHeader)
	if requestID == "" {
		t.Fatal("missing request id header")
	}
	return requestID
}

func RequireSuccessEnvelope(t testing.TB, resp *http.Response, wantStatus int) map[string]any {
	t.Helper()
	RequireStatus(t, resp, wantStatus)
	requestID := RequireRequestID(t, resp)
	body := ReadJSONBody(t, resp)

	metaValue, ok := body["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected success envelope meta object, got %T", body["meta"])
	}
	if metaValue["request_id"] != requestID {
		t.Fatalf("body meta.request_id mismatch: got %v want %s", metaValue["request_id"], requestID)
	}
	if _, ok := body["data"].(map[string]any); !ok {
		t.Fatalf("expected success envelope data object, got %T", body["data"])
	}
	return body
}

func RequireErrorEnvelope(t testing.TB, resp *http.Response, wantStatus int, wantCode string) map[string]any {
	t.Helper()
	RequireStatus(t, resp, wantStatus)
	requestID := RequireRequestID(t, resp)
	body := ReadJSONBody(t, resp)

	errorValue, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %T", body["error"])
	}
	if errorValue["request_id"] != requestID {
		t.Fatalf("error.request_id mismatch: got %v want %s", errorValue["request_id"], requestID)
	}
	if errorValue["code"] != wantCode {
		t.Fatalf("unexpected error code: got %v want %s", errorValue["code"], wantCode)
	}
	if _, ok := errorValue["retryable"].(bool); !ok {
		t.Fatalf("expected error.retryable boolean, got %T", errorValue["retryable"])
	}
	if _, ok := errorValue["details"].(map[string]any); !ok {
		t.Fatalf("expected error details object, got %T", errorValue["details"])
	}
	return body
}

func RequireErrorDetails(t testing.TB, body map[string]any) map[string]any {
	t.Helper()
	errorValue, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %T", body["error"])
	}
	details, ok := errorValue["details"].(map[string]any)
	if !ok {
		t.Fatalf("expected error details object, got %T", errorValue["details"])
	}
	return details
}

func RequireErrorDetail(t testing.TB, body map[string]any, key string, want any) {
	t.Helper()
	details := RequireErrorDetails(t, body)
	if got := details[key]; got != want {
		t.Fatalf("unexpected error.details.%s: got %#v want %#v in %#v", key, got, want, details)
	}
}

func RequireErrorDetailStrings(t testing.TB, body map[string]any, key string, want []string) {
	t.Helper()
	details := RequireErrorDetails(t, body)
	raw, ok := details[key].([]any)
	if !ok {
		t.Fatalf("expected error.details.%s array, got %T in %#v", key, details[key], details)
	}
	if len(raw) != len(want) {
		t.Fatalf("unexpected error.details.%s length: got %#v want %#v", key, raw, want)
	}
	for index, value := range raw {
		if value != want[index] {
			t.Fatalf("unexpected error.details.%s[%d]: got %#v want %#v in %#v", key, index, value, want[index], raw)
		}
	}
}

func RequireAuthCookies(t testing.TB, cookies []*http.Cookie) AuthCookies {
	t.Helper()
	return authcookietest.RequireAuthCookies(t, cookies)
}
