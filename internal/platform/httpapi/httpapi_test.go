package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contractextensions"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
	"github.com/JochiRaider/cartulary/internal/testutil/browserbuild"
)

var embeddedAssetPathPattern = regexp.MustCompile(`(?:src|href)="(/assets/[^"]+)"`)

func TestNewHandler_ServesEmbeddedRootAndAssets(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler(Options{Dependencies: testExtensionDependencies(t, nil)})
	if err != nil {
		t.Fatalf("NewHandler(): %v", err)
	}

	expectedRoot := browserbuild.New(t).RootHTML()

	rootRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	rootRecorder := httptest.NewRecorder()
	handler.ServeHTTP(rootRecorder, rootRequest)

	rootResponse := rootRecorder.Result()
	defer rootResponse.Body.Close()

	if rootResponse.StatusCode != http.StatusOK {
		t.Fatalf("unexpected root status: got %d want %d", rootResponse.StatusCode, http.StatusOK)
	}
	if contentType := rootResponse.Header.Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("unexpected root content type: %q", contentType)
	}

	rootBody, err := io.ReadAll(rootResponse.Body)
	if err != nil {
		t.Fatalf("read root body: %v", err)
	}
	if string(rootBody) != string(expectedRoot) {
		t.Fatalf("unexpected root body: got %q want %q", string(rootBody), string(expectedRoot))
	}

	assetMatch := embeddedAssetPathPattern.FindSubmatch(rootBody)
	if len(assetMatch) < 2 {
		return
	}

	assetRequest := httptest.NewRequest(http.MethodGet, string(assetMatch[1]), nil)
	assetRecorder := httptest.NewRecorder()
	handler.ServeHTTP(assetRecorder, assetRequest)

	assetResponse := assetRecorder.Result()
	defer assetResponse.Body.Close()

	if assetResponse.StatusCode != http.StatusOK {
		t.Fatalf("unexpected asset status for %s: got %d want %d", string(assetMatch[1]), assetResponse.StatusCode, http.StatusOK)
	}
	assetBody, err := io.ReadAll(assetResponse.Body)
	if err != nil {
		t.Fatalf("read asset body: %v", err)
	}
	if len(assetBody) == 0 {
		t.Fatalf("expected non-empty asset body for %s", string(assetMatch[1]))
	}
}

func TestNewHandler_HealthzRemainsLivenessAndReadyzIsStructured(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler(Options{
		Dependencies: testExtensionDependenciesWith(t, DependencySet{
			Readiness: ReadinessCheckFunc(func(context.Context) ReadinessState {
				return ReadyReadinessState()
			}),
		}, nil),
	})
	if err != nil {
		t.Fatalf("NewHandler(): %v", err)
	}

	healthRequest := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthRecorder := httptest.NewRecorder()
	handler.ServeHTTP(healthRecorder, healthRequest)
	healthResponse := healthRecorder.Result()
	defer healthResponse.Body.Close()
	healthBody, err := io.ReadAll(healthResponse.Body)
	if err != nil {
		t.Fatalf("read healthz body: %v", err)
	}
	if healthResponse.StatusCode != http.StatusOK || string(healthBody) != "ok\n" {
		t.Fatalf("unexpected healthz response: status=%d body=%q", healthResponse.StatusCode, string(healthBody))
	}
	if contentType := healthResponse.Header.Get("Content-Type"); !strings.Contains(contentType, "text/plain") {
		t.Fatalf("unexpected healthz content type: %q", contentType)
	}

	readyRequest := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	readyRecorder := httptest.NewRecorder()
	handler.ServeHTTP(readyRecorder, readyRequest)
	readyResponse := readyRecorder.Result()
	defer readyResponse.Body.Close()
	if readyResponse.StatusCode != http.StatusOK {
		t.Fatalf("unexpected readyz status: got %d want %d", readyResponse.StatusCode, http.StatusOK)
	}
	var payload ReadinessState
	if err := json.NewDecoder(readyResponse.Body).Decode(&payload); err != nil {
		t.Fatalf("decode readyz response: %v", err)
	}
	if payload.SchemaID != ReadinessSchemaID || payload.Status != ReadinessStatusReady {
		t.Fatalf("unexpected readyz payload: %#v", payload)
	}
}

func TestNewHandler_ReadyzDegradedStatusAndRedactsDependencyDetails(t *testing.T) {
	t.Parallel()

	leakedValues := []string{
		"http://127.0.0.1:9000",
		"secret-bucket",
		"AKIA-SECRET",
		"object/key.txt",
		"storage://unsafe/ref",
		"postgres://user:pass@db.example.test/cartulary",
	}
	store := readinessFailingStore{err: errors.New(strings.Join(leakedValues, " "))}
	handler, err := NewHandler(Options{
		Dependencies: testExtensionDependenciesWith(t, DependencySet{
			Readiness: NewDependencyReadinessChecker(nil, store),
		}, nil),
	})
	if err != nil {
		t.Fatalf("NewHandler(): %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	response := recorder.Result()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read readyz body: %v", err)
	}
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected degraded readyz status: got %d body=%q", response.StatusCode, string(body))
	}
	for _, value := range leakedValues {
		if strings.Contains(string(body), value) {
			t.Fatalf("readyz leaked forbidden value %q in body %q", value, string(body))
		}
	}

	var payload ReadinessState
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode degraded readyz response: %v", err)
	}
	if payload.Status != ReadinessStatusDegradedDependency {
		t.Fatalf("unexpected degraded readyz status payload: %#v", payload)
	}
	if len(payload.Dependencies) != 1 || payload.Dependencies[0].ReasonCode != ReadinessReasonDependencyUnavailable {
		t.Fatalf("unexpected degraded dependency payload: %#v", payload.Dependencies)
	}
}

func TestNewHandler_KeepsReservedExtensionRouting(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler(Options{
		Dependencies: testExtensionDependencies(t, testGeneratedExtensionProfiles(t)),
	})
	if err != nil {
		t.Fatalf("NewHandler(): %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/providers", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected reserved-route status: got %d want %d", response.StatusCode, http.StatusNotFound)
	}

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode reserved-route payload: %v", err)
	}

	errorPayload, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("missing error payload: %#v", payload)
	}
	if errorPayload["code"] != "extension_profile_not_claimed" {
		t.Fatalf("unexpected reserved-route code: %#v", errorPayload)
	}
}

// TestNewHandler_ExtensionAdmissionGateClosesRequestsAndReadiness covers the
// admission gate's service_unavailable reasons (Core 01 §3.3.6.2, REQ-01-707,
// Core 04 REQ-04-078, REQ-04-145, REQ-04-146a, EXT-REQ-213, AC-618) through the
// process lifecycle that drives it. Before startup publication opens admission
// the gate answers extension_publication_pending; while lease ownership is
// uncertain admission_paused, including when uncertainty preceded
// publication; after an ordinary stop shutdown_in_progress; and once the
// process is fatal extension_integrity_failure. Each refusal is a retryable
// 503 that also covers an /api path matching no operation, and the first of
// integrity failure, shutdown, pause and pending wins. /readyz reports the
// same state, and liveness stays open throughout.
func TestNewHandler_ExtensionAdmissionGateClosesRequestsAndReadiness(t *testing.T) {
	t.Parallel()

	newGate := func(t *testing.T) (*processlifecycle.Controller, http.Handler) {
		t.Helper()
		lifecycle := processlifecycle.New()
		handler, err := NewHandler(Options{Dependencies: testExtensionDependenciesWith(t, DependencySet{
			Admission: lifecycle,
			Readiness: ReadinessCheckFunc(func(context.Context) ReadinessState { return ReadyReadinessState() }),
		}, nil)})
		if err != nil {
			t.Fatal(err)
		}
		return lifecycle, handler
	}
	assertGate := func(t *testing.T, handler http.Handler, wantReason string, wantReadiness string) {
		t.Helper()
		// The refusal precedes route resolution, so a path matching no
		// operation receives it too (REQ-01-707).
		for _, path := range []string{"/api/v1/auth/providers", "/api/v1/no-such-route"} {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			var envelope ErrorEnvelope
			if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode gated response for %s = %d %s: %v", path, recorder.Code, recorder.Body.String(), err)
			}
			if recorder.Code != http.StatusServiceUnavailable || envelope.Error.Code != "service_unavailable" ||
				envelope.Error.Status != http.StatusServiceUnavailable || !envelope.Error.Retryable ||
				envelope.Error.Details["reason_code"] != wantReason || len(envelope.Error.Details) != 1 {
				t.Fatalf("gated response for %s = %d %s, want retryable 503 service_unavailable with reason_code %s", path, recorder.Code, recorder.Body.String(), wantReason)
			}
		}

		readyRecorder := httptest.NewRecorder()
		handler.ServeHTTP(readyRecorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		var readiness ReadinessState
		if err := json.Unmarshal(readyRecorder.Body.Bytes(), &readiness); err != nil {
			t.Fatalf("decode gated readiness %d %s: %v", readyRecorder.Code, readyRecorder.Body.String(), err)
		}
		if readyRecorder.Code != http.StatusServiceUnavailable || readiness.SchemaID != ReadinessSchemaID || readiness.Status != wantReadiness {
			t.Fatalf("gated readiness = %d %s, want 503 with status %s", readyRecorder.Code, readyRecorder.Body.String(), wantReadiness)
		}
		healthRecorder := httptest.NewRecorder()
		handler.ServeHTTP(healthRecorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		if healthRecorder.Code != http.StatusOK {
			t.Fatalf("liveness closed with admission: %d", healthRecorder.Code)
		}
	}
	assertOpen := func(t *testing.T, handler http.Handler) {
		t.Helper()
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("ready admission response = %d %s", recorder.Code, recorder.Body.String())
		}
	}
	publish := func(t *testing.T, lifecycle *processlifecycle.Controller) {
		t.Helper()
		if err := lifecycle.Publish(); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("before publication the gate is pending", func(t *testing.T) {
		t.Parallel()
		_, handler := newGate(t)
		assertGate(t, handler, "extension_publication_pending", "starting_dependency_probe")
	})
	t.Run("uncertainty before publication keeps admission paused after it", func(t *testing.T) {
		t.Parallel()
		lifecycle, handler := newGate(t)
		lifecycle.CloseAdmission()
		assertGate(t, handler, "extension_publication_pending", "starting_dependency_probe")
		publish(t, lifecycle)
		assertGate(t, handler, "admission_paused", "admission_paused")
		if !lifecycle.RestoreAdmission() {
			t.Fatal("the original lease session's proof did not open the gate")
		}
		assertOpen(t, handler)
	})
	t.Run("uncertainty after publication pauses until the original session proves ownership", func(t *testing.T) {
		t.Parallel()
		lifecycle, handler := newGate(t)
		publish(t, lifecycle)
		assertOpen(t, handler)
		lifecycle.CloseAdmission()
		assertGate(t, handler, "admission_paused", "admission_paused")
		if !lifecycle.RestoreAdmission() {
			t.Fatal("the original lease session did not restore admission")
		}
		assertOpen(t, handler)
	})
	t.Run("ordinary shutdown answers shutdown_in_progress and never reopens", func(t *testing.T) {
		t.Parallel()
		lifecycle, handler := newGate(t)
		publish(t, lifecycle)
		lifecycle.MarkTerminating()
		assertGate(t, handler, "shutdown_in_progress", "shutdown_in_progress")
		if lifecycle.RestoreAdmission() {
			t.Fatal("a stopping process restored admission")
		}
		assertGate(t, handler, "shutdown_in_progress", "shutdown_in_progress")
	})
	t.Run("shutdown outranks uncertain ownership", func(t *testing.T) {
		t.Parallel()
		lifecycle, handler := newGate(t)
		publish(t, lifecycle)
		lifecycle.CloseAdmission()
		lifecycle.MarkTerminating()
		assertGate(t, handler, "shutdown_in_progress", "shutdown_in_progress")
	})
	t.Run("a fatal outranks every other state and is permanent", func(t *testing.T) {
		t.Parallel()
		lifecycle, handler := newGate(t)
		publish(t, lifecycle)
		lifecycle.CloseAdmission()
		assertGate(t, handler, "admission_paused", "admission_paused")
		if !lifecycle.Fatal("application_process_lease_lost") {
			t.Fatal("lease loss did not enter the fatal lifecycle")
		}
		assertGate(t, handler, "extension_integrity_failure", "fatal_integrity_failure")
		lifecycle.MarkTerminating()
		if lifecycle.RestoreAdmission() {
			t.Fatal("a fatal process restored admission")
		}
		assertGate(t, handler, "extension_integrity_failure", "fatal_integrity_failure")
	})
}

type readinessFailingStore struct {
	err error
}

func (store readinessFailingStore) UploadTarget(context.Context, string, time.Time) (objectstore.UploadTarget, error) {
	return objectstore.UploadTarget{}, store.err
}

func (store readinessFailingStore) CompleteUploadTarget(context.Context, string, io.Reader, string) error {
	return store.err
}

func (store readinessFailingStore) PutObject(context.Context, string, io.Reader, int64, string) error {
	return store.err
}

func (store readinessFailingStore) ReadObject(context.Context, string, objectstore.ReadOptions) (io.ReadCloser, objectstore.ObjectInfo, error) {
	return nil, objectstore.ObjectInfo{}, store.err
}

func (store readinessFailingStore) StatObject(context.Context, string) (objectstore.ObjectInfo, error) {
	return objectstore.ObjectInfo{}, store.err
}

func (store readinessFailingStore) ListObjects(context.Context, string) ([]objectstore.ObjectInfo, error) {
	return nil, store.err
}

func (store readinessFailingStore) DeleteObject(context.Context, string) error {
	return store.err
}

func (store readinessFailingStore) Close() error {
	return nil
}

func TestNewHandler_KeepsUnclaimedReservedExtensionRootsUnavailable(t *testing.T) {
	t.Parallel()

	profiles := testGeneratedExtensionProfiles(t)
	handler, err := NewHandler(Options{Dependencies: testExtensionDependencies(t, profiles)})
	if err != nil {
		t.Fatalf("NewHandler(): %v", err)
	}

	for _, profile := range profiles {
		if profile.Claimed {
			continue
		}
		for _, routeFamily := range profile.RouteFamilies {
			path := strings.ReplaceAll(routeFamily, "{user_id}", "11111111-1111-1111-1111-111111111111")
			path = strings.ReplaceAll(path, "{incident_id}", "22222222-2222-2222-2222-222222222222")
			for _, path := range []string{
				path,
				strings.Replace(path, "/api/v1/", "/api//v1/", 1),
				strings.Replace(path, "/api/v1/", "/api/v1/other/../", 1),
				path + "/../outside",
			} {
				request := httptest.NewRequest(http.MethodGet, path, nil)
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, request)

				response := recorder.Result()
				body, err := io.ReadAll(response.Body)
				_ = response.Body.Close()
				if err != nil {
					t.Fatalf("read reserved-route response for %s: %v", path, err)
				}
				if response.StatusCode != http.StatusNotFound {
					t.Fatalf("reserved root %s returned status %d body %q", path, response.StatusCode, string(body))
				}

				var payload map[string]any
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatalf("decode reserved-route payload for %s: %v body=%q", path, err, string(body))
				}
				errorPayload, ok := payload["error"].(map[string]any)
				if !ok {
					t.Fatalf("missing error payload for %s: %#v", path, payload)
				}
				if errorPayload["code"] != "extension_profile_not_claimed" {
					t.Fatalf("reserved root %s returned unexpected error: %#v", path, errorPayload)
				}
				details, ok := errorPayload["details"].(map[string]any)
				if !ok {
					t.Fatalf("missing reserved-route details for %s: %#v", path, errorPayload)
				}
				if details["profile_id"] != profile.ProfileID || details["route_family"] != routeFamily {
					t.Fatalf("unexpected reserved-route details for %s: %#v", path, details)
				}
				if response.Header.Get("Location") != "" {
					t.Fatalf("unclaimed reserved path %s redirected", path)
				}
			}
		}
	}
}

func TestExplicitExtensionProfilesMatchExtensionContractRegistry(t *testing.T) {
	t.Parallel()

	artifact, ok := contractsgen.Index["contracts/extensions/generated/profile-registry.json"]
	if !ok {
		t.Fatal("generated extension contract registry is not packaged")
	}
	var registry struct {
		Profiles []struct {
			ProfileID     string   `json:"profile_id"`
			Claimable     bool     `json:"claimable"`
			ContractMajor int      `json:"contract_major"`
			Families      []string `json:"route_families"`
			WorkspaceKeys []string `json:"workspace_keys"`
			Capabilities  []string `json:"capability_ids"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(artifact.JSON), &registry); err != nil {
		t.Fatalf("decode extension contract registry: %v", err)
	}
	runtimeProfiles := testGeneratedExtensionProfiles(t)
	if len(runtimeProfiles) != len(registry.Profiles) {
		t.Fatalf("extension contract profile count mismatch: runtime=%d contract=%d", len(runtimeProfiles), len(registry.Profiles))
	}
	for index, wantProfile := range registry.Profiles {
		gotProfile := runtimeProfiles[index]
		if gotProfile.ProfileID != wantProfile.ProfileID {
			t.Fatalf("extension contract profile id mismatch at %d: runtime=%s contract=%s", index, gotProfile.ProfileID, wantProfile.ProfileID)
		}
		if strings.Join(gotProfile.RouteFamilies, ",") != strings.Join(wantProfile.Families, ",") {
			t.Fatalf("extension contract route families mismatch for %s: runtime=%v contract=%v", wantProfile.ProfileID, gotProfile.RouteFamilies, wantProfile.Families)
		}
		if !gotProfile.Claimable || gotProfile.ContractMajor == nil || *gotProfile.ContractMajor != wantProfile.ContractMajor {
			t.Fatalf("extension contract recognition mismatch for %s: runtime=%#v contract=%#v", wantProfile.ProfileID, gotProfile, wantProfile)
		}
		if strings.Join(gotProfile.WorkspaceKeys, ",") != strings.Join(wantProfile.WorkspaceKeys, ",") || len(gotProfile.Capabilities) != len(wantProfile.Capabilities) {
			t.Fatalf("extension contract workspace/capability mismatch for %s", wantProfile.ProfileID)
		}
	}
}

func TestNewHandlerRequiresExplicitExtensionProjections(t *testing.T) {
	t.Parallel()

	_, err := NewHandler()
	if err == nil || !strings.Contains(err.Error(), "http extension projections must be explicit") {
		t.Fatalf("missing extension projections error = %v", err)
	}
}

type testExtensionProjectionSet struct {
	profiles []ExtensionProfile
}

func testExtensionDependencies(t testing.TB, profiles []ExtensionProfile) DependencySet {
	return testExtensionDependenciesWith(t, DependencySet{}, profiles)
}

func testExtensionDependenciesWith(t testing.TB, deps DependencySet, profiles []ExtensionProfile) DependencySet {
	if deps.BrowserBuild == nil {
		deps.BrowserBuild = browserbuild.New(t)
	}
	projections := testExtensionProjectionSet{profiles: cloneExtensionProfiles(profiles)}
	deps.ExtensionDiscovery = projections
	deps.ExtensionClaims = projections
	deps.ExtensionRoutes = projections
	deps.ExtensionWorkspaces = projections
	return deps
}

func (p testExtensionProjectionSet) ExtensionDiscoveryProfiles() []ExtensionProfile {
	return cloneExtensionProfiles(p.profiles)
}

func (p testExtensionProjectionSet) ExtensionClaims() []ExtensionClaim {
	claims := make([]ExtensionClaim, 0, len(p.profiles))
	for _, profile := range p.profiles {
		claims = append(claims, ExtensionClaim{ProfileID: profile.ProfileID, Claimed: profile.Claimed})
	}
	return claims
}

func (p testExtensionProjectionSet) ExtensionRoutes() []ExtensionRoute {
	routes := []ExtensionRoute{}
	for _, profile := range p.profiles {
		for _, routeFamily := range profile.RouteFamilies {
			routes = append(routes, ExtensionRoute{
				ProfileID: profile.ProfileID, RouteFamily: routeFamily, Claimed: profile.Claimed,
			})
		}
	}
	return routes
}

func (p testExtensionProjectionSet) ExtensionWorkspaces() []ExtensionWorkspacePublication {
	workspaces := []ExtensionWorkspacePublication{}
	for _, profile := range p.profiles {
		if !profile.Claimed {
			continue
		}
		for _, workspace := range profile.Workspaces {
			workspaces = append(workspaces, ExtensionWorkspacePublication{
				ProfileID: profile.ProfileID, WorkspaceKey: workspace.WorkspaceKey, MinimumRole: workspace.MinimumRole,
			})
		}
	}
	return workspaces
}

func testGeneratedExtensionProfiles(t testing.TB) []ExtensionProfile {
	t.Helper()
	artifact, ok := contractsgen.Index["contracts/extensions/generated/profile-registry.json"]
	if !ok {
		t.Fatal("generated extension contract registry is not packaged")
	}
	var registry struct {
		Profiles []struct {
			ProfileID     string   `json:"profile_id"`
			Claimable     bool     `json:"claimable"`
			ContractMajor int      `json:"contract_major"`
			RouteFamilies []string `json:"route_families"`
			WorkspaceKeys []string `json:"workspace_keys"`
			Capabilities  []string `json:"capability_ids"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(artifact.JSON), &registry); err != nil {
		t.Fatalf("decode extension contract registry fixture: %v", err)
	}
	profiles := make([]ExtensionProfile, 0, len(registry.Profiles))
	for _, descriptor := range registry.Profiles {
		contractMajor := descriptor.ContractMajor
		workspaces := make([]ExtensionWorkspace, 0, len(descriptor.WorkspaceKeys))
		for _, workspaceKey := range descriptor.WorkspaceKeys {
			workspaces = append(workspaces, ExtensionWorkspace{
				WorkspaceKey: workspaceKey,
				MinimumRole:  "viewer",
			})
		}
		profiles = append(profiles, ExtensionProfile{
			ProfileID:     descriptor.ProfileID,
			Claimable:     descriptor.Claimable,
			ContractMajor: &contractMajor,
			RouteFamilies: append([]string(nil), descriptor.RouteFamilies...),
			WorkspaceKeys: append([]string(nil), descriptor.WorkspaceKeys...),
			Capabilities:  append([]string(nil), descriptor.Capabilities...),
			Workspaces:    workspaces,
		})
	}
	return profiles
}
