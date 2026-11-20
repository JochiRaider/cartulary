package server

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/server/holdjob"
	"github.com/JochiRaider/cartulary/internal/gen/networkflowroutes"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	networkflowharnesscontrol "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/browserbuild"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/reportingfixture"
)

// TestEveryBoundPublicOperationIsAdmittedCentrally_Integration covers Core 01
// REQ-01-707 for the assembled runtime: with every extension profile claimed,
// every public operation of the registry, OpenAPI and Network Flow alike, is
// bound, and each is admitted centrally. There is no allowlist; an operation
// bound in owner mode fails the test by name. Every other route is recorded by
// the registry, so none is registered outside it: the WebSocket upgrade
// (REQ-01-708) and, in a harness-owned runtime, the guarded test-only harness
// routes (Testing Harness NLSpec TH-HARNESS-REQ-450).
func TestEveryBoundPublicOperationIsAdmittedCentrally_Integration(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "central-admission-catalog")
	admittedPool := openAdmittedRuntimePool(t, ctx, testDB)
	defer admittedPool.Close()

	// The Network Flow key-ring fixture names these secrets.
	t.Setenv("CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR", "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE")
	t.Setenv("CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST", "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI")
	cfg := BindPostgres(t, RuntimeConfigWithOverlays(t, configtest.Overlay(
		"CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH", fixtures.Path("bootstrap-admin", "canonical.json"),
		"CARTULARY__ENTERPRISE_AUTHENTICATION__CLAIMED", "true",
		"CARTULARY__ENTERPRISE_AUTHENTICATION__PROVIDER_MANIFEST_PATH", fixtures.Path("enterprise-auth", "empty.json"),
		"CARTULARY__IMPORT__CLAIMED", "true",
		"CARTULARY__INCIDENT_PORTABILITY__CLAIMED", "true",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED", "true",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH", fixtures.Path("network-flow", "key-rings.json"),
		"CARTULARY__REFERENCE_PACK__CLAIMED", "true",
		"CARTULARY__SNAPSHOT_REPORTING__CLAIMED", "true",
	)), testDB.Env())

	socket := httpapi.SupplementaryRoute{Kind: httpapi.SupplementarySocketRoute, Method: http.MethodGet, Path: "/ws/v1/incidents/{incident_id}"}
	t.Run("production composition", func(t *testing.T) {
		runtime := startCatalogRuntime(t, ctx, cfg, admittedPool, Options{})
		registry := runtime.publicRoutes
		requireEveryOperationAdmittedCentrally(t, registry)
		// The socket is the only route outside the operation registry.
		if got := registry.SupplementaryRoutes(); !reflect.DeepEqual(got, []httpapi.SupplementaryRoute{socket}) {
			t.Fatalf("routes recorded outside the operation registry = %#v, want only the socket", got)
		}
	})

	t.Run("harness test routes are bound through the guarded helper", func(t *testing.T) {
		// The server-harness profile's test-route composition
		// (server_profile_harness.go, build-tagged out of this test).
		t.Setenv(httpapi.TestRoutesEnabledEnv, "1")
		t.Setenv(httpapi.TestRuntimeMarkerEnv, httpapi.TestRuntimeMarkerValue)
		t.Setenv(httpapi.TestRouteTokenEnv, "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG")
		testClock := httpapi.NewTestClock()
		networkFlowControls := networkflowharnesscontrol.NewControls()
		holdJobs := holdjob.NewControl()
		runtime := startCatalogRuntime(t, ctx, cfg, admittedPool, Options{
			NetworkFlowComposition: networkFlowControls,
			Now:                    testClock.Now,
			HarnessJobs:            []jobs.Contribution{holdJobs.Contribution()},
			HTTP: httpapi.Options{AdditionalRoutes: append(
				harnessruntime.RegisterRoutes(testClock, networkFlowControls.Contribution(), harnessruntime.ControlContribution{Routes: []httpapi.RouteRegistrar{holdJobs.Routes()}}),
				auth.RegisterTestRoutes(),
				savedviews.RegisterTestRoutes(),
			)},
		})
		registry := runtime.publicRoutes
		requireEveryOperationAdmittedCentrally(t, registry)
		test := func(method, path string) httpapi.SupplementaryRoute {
			return httpapi.SupplementaryRoute{Kind: httpapi.SupplementaryTestRoute, Method: method, Path: path}
		}
		want := []httpapi.SupplementaryRoute{
			test(http.MethodGet, "/api/v1/test/auth/touch"),
			test(http.MethodPost, "/api/v1/test/clock/reset"),
			test(http.MethodPost, "/api/v1/test/clock/set"),
			test(http.MethodGet, "/api/v1/test/clock/state"),
			test(http.MethodPost, "/api/v1/test/incidents/{incident_id}/saved-views/system"),
			test(http.MethodPost, "/api/v1/test/runtime/hold-jobs"),
			test(http.MethodGet, "/api/v1/test/runtime/identity"),
			test(http.MethodPost, "/api/v1/test/runtime/network-flow-audit-assertions"),
			test(http.MethodPost, "/api/v1/test/runtime/network-flow-auth-transitions"),
			test(http.MethodPost, "/api/v1/test/runtime/network-flow-faults"),
			test(http.MethodPost, "/api/v1/test/runtime/network-flow-randomness"),
			socket,
		}
		if got := registry.SupplementaryRoutes(); !reflect.DeepEqual(got, want) {
			t.Fatalf("routes recorded outside the operation registry = %#v, want %#v", got, want)
		}
		// No recorded route shares a path with a public operation.
		for _, route := range registry.SupplementaryRoutes() {
			for _, operation := range registry.Snapshot() {
				if operation.PathTemplate == route.Path {
					t.Fatalf("%s route %s shares its path with operation %s", route.Kind, route.Path, operation.OperationID)
				}
			}
		}
	})
}

// startCatalogRuntime starts and publishes an all-claimed runtime. A startup
// failure reports only its closed diagnostic (Extensions EXT-REQ-160), so the
// test also prints the retained cause, such as a route-parity failure.
func startCatalogRuntime(t *testing.T, ctx context.Context, cfg configassembly.Loaded, pool postgres.AdmittedPool, options Options) *Runtime {
	t.Helper()
	store, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("open object store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	options.BrowserBuild = browserbuild.New(t)
	options.ReportingRuntime = reportingfixture.Runtime{}
	options.Postgres = pool
	options.ObjectStore = store
	runtime, err := NewRuntime(ctx, cfg, options)
	if err != nil {
		t.Fatalf("start all-claimed runtime: %v (cause: %v)", err, startupCause(err))
	}
	t.Cleanup(runtime.Close)
	if err := runtime.ActivatePublication(); err != nil {
		t.Fatalf("activate all-claimed publication: %v (cause: %v)", err, startupCause(err))
	}
	claims := runtime.publication.claims()
	if len(claims) != 6 {
		t.Fatalf("claim count = %d: %#v", len(claims), claims)
	}
	for _, claim := range claims {
		if !claim.Claimed {
			t.Fatalf("profile %q is not claimed, so its operations are not enumerated", claim.ProfileID)
		}
	}
	if runtime.publicRoutes == nil {
		t.Fatal("the runtime retained no public route registry")
	}
	return runtime
}

// startupCause unwraps a startup failure to the cause its diagnostic keeps
// private.
func startupCause(err error) error {
	cause := errors.Unwrap(err)
	if cause == nil {
		return err
	}
	return cause
}

func requireEveryOperationAdmittedCentrally(t *testing.T, registry *httpapi.RouteRegistry) {
	t.Helper()
	diagnostics := registry.Diagnostics()
	if diagnostics.SupportedOperationCount == 0 || diagnostics.ActiveOperationCount != diagnostics.SupportedOperationCount {
		t.Fatalf("active operations = %d, supported = %d; every claimed operation must be bound", diagnostics.ActiveOperationCount, diagnostics.SupportedOperationCount)
	}
	active := registry.Snapshot()
	modes := registry.AdmissionModes()
	if len(modes) != len(active) {
		t.Fatalf("admission modes = %d, active operations = %d", len(modes), len(active))
	}
	var offenders []string
	networkFlow := 0
	for _, operation := range active {
		mode, recorded := modes[operation.OperationID]
		if !recorded || mode != httpapi.AdmissionCentral {
			offenders = append(offenders, operation.OperationID+": "+string(mode))
		}
		if operation.OwnerID == "module.networkflow" {
			networkFlow++
		}
	}
	sort.Strings(offenders)
	if len(offenders) != 0 {
		t.Fatalf("operations not admitted centrally (REQ-01-707):\n%s", strings.Join(offenders, "\n"))
	}
	if want := len(networkflowroutes.All()); networkFlow != want {
		t.Fatalf("bound Network Flow operations = %d, want %d", networkFlow, want)
	}
}
