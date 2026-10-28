package httptestx

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/reportingfixture"
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
	runtime       *collaboration.Runtime
	recordChanges collaboration.RecordChangedAppender
}

func (capability *CollaborationCapability) RevokeSession(sessionID uuid.UUID, reasonCode string) {
	if capability == nil || capability.runtime == nil {
		return
	}
	capability.runtime.RevokeSession(sessionID, reasonCode)
}

func (capability *CollaborationCapability) CloseDispatcher(ctx context.Context) error {
	if capability == nil || capability.runtime == nil {
		return errors.New("collaboration dispatcher capability is unavailable")
	}
	return capability.runtime.Close(ctx)
}

func (capability *CollaborationCapability) RecordChanges() collaboration.RecordChangedAppender {
	if capability == nil {
		return nil
	}
	return capability.recordChanges
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

func (capability *RevisionsCapability) ConflictFieldResolver() conflicts.FieldResolver {
	if capability == nil || capability.runtime == nil {
		return nil
	}
	return capability.runtime.ConflictFieldResolver()
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
	ConfigureRuntime func(*server.Options)
	Loaded           *configassembly.Loaded
	Env              map[string]string
	Dependencies     httpapi.DependencySet
	AdditionalRoutes []httpapi.RouteRegistrar
	Postgres         postgres.AdmittedPool
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
	configtest.EnsureCryptographicTestEnvironment(env)

	var loaded configassembly.Loaded
	if options.Loaded == nil {
		if env["CARTULARY__REFERENCE_PACK__CLAIMED"] != "false" {
			if _, exists := env["CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH"]; !exists {
				env["CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH"] = fixtures.Path("reference-packs", "trust-bootstrap.json")
			}
		}
		tempRoots := configtest.SetupTempRoots(t)
		for key, value := range tempRoots.Paths {
			if _, exists := env[key]; !exists {
				env[key] = value
			}
		}
		configtest.BindPostgresEnvToDatabaseRoot(t, tempRoots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], env, postgres.PurposeRuntime)
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
	runtimeOptions := server.Options{
		ReportingRuntime: reportingfixture.Runtime{},
		Env:              env,
		Now:              clock.Now,
		Postgres:         options.Postgres,
		ObjectStore:      options.ObjectStore,
		ObserveJobs: func(manager *jobs.Manager, transactions *jobs.TransactionService, _ *jobs.Runner, pool *pgxpool.Pool) {
			jobsCapability = &JobsCapability{manager: manager, transactions: transactions, pool: pool, now: clock.Now}
		},
		ObserveCollaboration: func(runtime *collaboration.Runtime) {
			collaborationCapability = &CollaborationCapability{
				runtime:       runtime,
				recordChanges: runtime.RecordChanges(),
			}
		},
		ObserveProjections: func(runtime *projectionassembly.Runtime) {
			projectionCapability = projectiontestsupport.New(runtime)
			indicatorSourceText = indicatorassembly.NewSourceTextPort(runtime.SourceTextRows())
		},
		ObserveRevisions: func(runtime *revisionassembly.Runtime) {
			revisionsCapability = &RevisionsCapability{runtime: runtime}
		},
		HTTP: httpapi.Options{
			Dependencies:     options.Dependencies,
			AdditionalRoutes: routes,
		},
	}
	if options.ConfigureRuntime != nil {
		options.ConfigureRuntime(&runtimeOptions)
	}
	runtime, err := server.NewRuntime(context.Background(), loaded, runtimeOptions)
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
