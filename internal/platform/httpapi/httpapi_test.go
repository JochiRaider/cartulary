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

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contracts"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi/webassets"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

var embeddedAssetPathPattern = regexp.MustCompile(`(?:src|href)="(/assets/[^"]+)"`)

func TestNewHandler_ServesEmbeddedRootAndAssets(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler(): %v", err)
	}

	expectedRoot, err := webassets.ReadBrowserRootHTML()
	if err != nil {
		t.Fatalf("ReadBrowserRootHTML(): %v", err)
	}

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
		Dependencies: DependencySet{
			Readiness: ReadinessCheckFunc(func(context.Context) ReadinessState {
				return ReadyReadinessState()
			}),
		},
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
		Dependencies: DependencySet{
			Readiness: NewDependencyReadinessChecker(nil, store),
		},
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

	handler, err := NewHandler()
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

func TestNewHandler_ExtensionAdmissionGateClosesRequestsAndReadiness(t *testing.T) {
	t.Parallel()

	gate := &testAdmissionGate{}
	handler, err := NewHandler(Options{Dependencies: DependencySet{
		Admission: gate,
		Readiness: ReadinessCheckFunc(func(context.Context) ReadinessState { return ReadyReadinessState() }),
	}})
	if err != nil {
		t.Fatal(err)
	}

	assertGate := func(wantReason string, wantReadiness string) {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/providers", nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusServiceUnavailable || !strings.Contains(recorder.Body.String(), `"reason_code":"`+wantReason+`"`) {
			t.Fatalf("gated response = %d %s", recorder.Code, recorder.Body.String())
		}

		readyRequest := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		readyRecorder := httptest.NewRecorder()
		handler.ServeHTTP(readyRecorder, readyRequest)
		if readyRecorder.Code != http.StatusServiceUnavailable || !strings.Contains(readyRecorder.Body.String(), `"status":"`+wantReadiness+`"`) {
			t.Fatalf("gated readiness = %d %s", readyRecorder.Code, readyRecorder.Body.String())
		}
		healthRecorder := httptest.NewRecorder()
		handler.ServeHTTP(healthRecorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		if healthRecorder.Code != http.StatusOK {
			t.Fatalf("liveness closed with admission: %d", healthRecorder.Code)
		}
	}

	assertGate("extension_publication_pending", ReadinessStatusStartingDependencyProbe)
	gate.reason = "published_component_lost"
	assertGate("extension_integrity_failure", ReadinessStatusFatalIntegrityFailure)
	gate.open = true
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("ready admission response = %d %s", recorder.Code, recorder.Body.String())
	}
}

type testAdmissionGate struct {
	open   bool
	reason string
}

func (g *testAdmissionGate) AdmissionOpen() bool { return g.open }
func (g *testAdmissionGate) FatalReason() string { return g.reason }

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

	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler(): %v", err)
	}

	for _, profile := range CurrentExtensionProfiles() {
		if profile.Claimed {
			continue
		}
		for _, routeFamily := range profile.RouteFamilies {
			path := strings.ReplaceAll(routeFamily, "{user_id}", "11111111-1111-1111-1111-111111111111")
			path = strings.ReplaceAll(path, "{incident_id}", "22222222-2222-2222-2222-222222222222")
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
		}
	}
}

func TestCurrentExtensionProfilesMatchExtensionContractRegistry(t *testing.T) {
	t.Parallel()

	artifact, ok := contractsgen.ExtensionArtifactsIndex["contracts/extensions/generated/profile-registry.json"]
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
	runtimeProfiles := CurrentExtensionProfiles()
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
