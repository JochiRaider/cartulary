package httptestx

import (
	"bytes"
	"context"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestProjectionCapabilityCallerMatrix(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	harnessSource, err := os.ReadFile(filepath.Join(repoRoot, "internal", "testutil", "httptestx", "httptestx.go"))
	if err != nil {
		t.Fatalf("read httptestx source: %v", err)
	}
	for _, forbidden := range []string{
		"type ProjectionCapability struct",
		"ProjectionCapability{bundle:",
		"ProjectionCatalog",
		"EvidenceProjectionPortFor",
	} {
		if bytes.Contains(harnessSource, []byte(forbidden)) {
			t.Fatalf("httptestx retains forbidden Timeline-owned projection capability %q", forbidden)
		}
	}
	if !bytes.Contains(harnessSource, []byte("projectiontestsupport.New(")) {
		t.Fatal("httptestx does not construct the typed Projections-owned test capability")
	}

	capabilityType := reflect.TypeOf(ProjectionCapability{})
	for index := 0; index < capabilityType.NumField(); index++ {
		if capabilityType.Field(index).IsExported() {
			t.Fatalf("projection test capability unexpectedly exports field %q", capabilityType.Field(index).Name)
		}
	}

	want := map[string][]string{
		"internal/modules/entities/resolution_integration_test.go": {
			".Projections.RebuildHosts(",
		},
		"internal/modules/evidence/integration_test.go": {
			".Projections.EvidencePort(",
			".Projections.RebuildTimeline(",
		},
		"internal/modules/indicators/resolution_integration_test.go": {
			".Projections.RebuildIndicators(",
		},
		"internal/modules/revisions/indicator_children_test.go": {
			".Projections.IndicatorProjectionPort(",
		},
		"internal/modules/timeline/resolution_integration_test.go": {
			".Projections.RebuildTimeline(",
		},
		"internal/modules/timeline/timeline_event_integration_test.go": {
			".Projections.RebuildTimeline(",
		},
		"internal/testutil/appsupport/runtime.go": {
			"server.ProjectionCapability()",
		},
	}

	got := map[string][]string{}
	err = filepath.WalkDir(filepath.Join(repoRoot, "internal"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		relative, relativeErr := filepath.Rel(repoRoot, path)
		if relativeErr != nil {
			return relativeErr
		}
		relative = filepath.ToSlash(relative)
		if relative == "internal/testutil/httptestx/httptestx_test.go" {
			return nil
		}
		for _, marker := range []string{
			".Projections.RebuildHosts(",
			".Projections.RebuildIdentities(",
			".Projections.RebuildIndicators(",
			".Projections.RebuildTimeline(",
			".Projections.IndicatorProjectionPort(",
			".Projections.EvidencePort(",
			"server.ProjectionCapability()",
		} {
			if bytes.Contains(content, []byte(marker)) {
				got[relative] = append(got[relative], marker)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inventory projection capability callers: %v", err)
	}
	for path := range want {
		sort.Strings(want[path])
	}
	for path := range got {
		sort.Strings(got[path])
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("projection capability caller matrix changed:\ngot  %#v\nwant %#v", got, want)
	}
}

func TestHarnessBootsServerAndAssertsEnvelopes(t *testing.T) {
	serverType := reflect.TypeOf(Server{})
	wantExportedFields := map[string]struct{}{"Clock": {}, "Config": {}, "HTTP": {}}
	for index := 0; index < serverType.NumField(); index++ {
		field := serverType.Field(index)
		if !field.IsExported() {
			continue
		}
		if _, expected := wantExportedFields[field.Name]; !expected {
			t.Fatalf("httptestx.Server unexpectedly exports field %q", field.Name)
		}
		delete(wantExportedFields, field.Name)
	}
	if len(wantExportedFields) != 0 {
		t.Fatalf("httptestx.Server is missing exported fields: %#v", wantExportedFields)
	}

	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "httptestx")

	s3Harness := s3test.Start(t)
	bucket, err := s3Harness.BootstrapBucket(context.Background(), "httptestx")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}

	server := StartServer(t, ServerOptions{Env: env, TestRouteMode: TestRouteModeHarnessOwned})

	successResp := Do(t, server.HTTP.Client(), NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/success", nil))
	successBody := RequireSuccessEnvelope(t, successResp, http.StatusOK)
	data := successBody["data"].(map[string]any)
	if data["service"] != "bootstrap" || data["status"] != "ok" {
		t.Fatalf("unexpected success payload: %#v", data)
	}

	errorResp := Do(t, server.HTTP.Client(), NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/error", nil))
	errorBody := RequireErrorEnvelope(t, errorResp, http.StatusServiceUnavailable, "bootstrap_error")
	errorDetails := errorBody["error"].(map[string]any)["details"].(map[string]any)
	if errorDetails["reason_code"] != "bootstrap_unavailable" {
		t.Fatalf("unexpected error details: %#v", errorDetails)
	}
}

func TestStartServerHonorsTestRouteMode(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "httptestx-test-route-mode")

	s3Harness := s3test.Start(t)
	bucket, err := s3Harness.BootstrapBucket(context.Background(), "httptestx-test-route-mode")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}

	t.Run("harness owned", func(t *testing.T) {
		server := StartServer(t, ServerOptions{Env: env, TestRouteMode: TestRouteModeHarnessOwned})
		req := NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/clock/state", nil)
		req.Header.Set("X-Cartulary-Test-Route-Token", TestRouteToken)

		resp := Do(t, server.HTTP.Client(), req)
		RequireSuccessEnvelope(t, resp, http.StatusOK)
	})

	t.Run("disabled", func(t *testing.T) {
		server := StartServer(t, ServerOptions{
			Env:           env,
			TestRouteMode: TestRouteModeDisabled,
		})
		req := NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/clock/state", nil)
		req.Header.Set("X-Cartulary-Test-Route-Token", TestRouteToken)

		resp := Do(t, server.HTTP.Client(), req)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("disabled test route mode must leave clock routes unavailable: got %d want %d", resp.StatusCode, http.StatusNotFound)
		}
		_ = resp.Body.Close()
	})

	t.Run("custom env", func(t *testing.T) {
		customEnv := make(map[string]string, len(env)+3)
		for key, value := range env {
			customEnv[key] = value
		}
		customEnv["CARTULARY_ENABLE_TEST_ROUTES"] = "1"
		customEnv["CARTULARY_TEST_RUNTIME_MARKER"] = "harness-owned"
		customEnv["CARTULARY_TEST_ROUTE_TOKEN"] = TestRouteToken

		server := StartServer(t, ServerOptions{
			Env:           customEnv,
			TestRouteMode: TestRouteModeCustomEnv,
		})
		req := NewJSONRequest(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/clock/state", nil)
		req.Header.Set("X-Cartulary-Test-Route-Token", TestRouteToken)

		resp := Do(t, server.HTTP.Client(), req)
		RequireSuccessEnvelope(t, resp, http.StatusOK)
	})
}

func TestApplyTestRouteModeRejectsMissingAndUnknownModes(t *testing.T) {
	for _, mode := range []TestRouteMode{"", "unknown"} {
		env := map[string]string{}
		if err := applyTestRouteMode(env, mode); err == nil {
			t.Fatalf("expected mode %q to fail", mode)
		}
		if len(env) != 0 {
			t.Fatalf("invalid mode %q mutated env: %#v", mode, env)
		}
	}
}

func TestRequireAuthCookies(t *testing.T) {
	authCookies := RequireAuthCookies(t, []*http.Cookie{
		{
			Name:     authn.SessionCookieName,
			Value:    "session-token",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		},
		{
			Name:     authn.CSRFCookieName,
			Value:    "csrf-token",
			Path:     "/",
			HttpOnly: false,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		},
	})

	if authCookies.Session == nil || authCookies.Session.Value != "session-token" {
		t.Fatalf("unexpected session cookie: %#v", authCookies.Session)
	}
	if authCookies.CSRF == nil || authCookies.CSRF.Value != "csrf-token" {
		t.Fatalf("unexpected csrf cookie: %#v", authCookies.CSRF)
	}
}
