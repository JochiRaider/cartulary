package httpapi_test

import (
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httpapiextensions"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReservedExtensionDispatchHonorsBaseRoutesClaimedFamiliesAndOutsideFallback_Unit(t *testing.T) {
	importProfile := extensionContract(t, "import")
	enterpriseProfile := extensionContract(t, "enterprise_authentication")

	baseProjections := httpapiextensions.New(nil)
	baseHandler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies: baseProjections.Dependencies(httpapi.DependencySet{}),
	})
	if err != nil {
		t.Fatalf("build base handler: %v", err)
	}
	ready := performRequest(t, baseHandler, http.MethodGet, "/readyz", nil)
	if ready.Code != http.StatusOK || !strings.Contains(ready.Body.String(), "ready") {
		t.Fatalf("expected base route to retain precedence, got status=%d body=%q", ready.Code, ready.Body.String())
	}

	claimedProjections := httpapiextensions.New([]httpapi.ExtensionProfile{
		{
			ProfileID:     importProfile.ProfileID,
			Claimed:       true,
			RouteFamilies: append([]string(nil), importProfile.RouteFamilies...),
		},
	})
	claimedHandler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies: claimedProjections.Dependencies(httpapi.DependencySet{}),
		AdditionalRoutes: []httpapi.RouteRegistrar{
			func(mux *http.ServeMux, deps httpapi.DependencySet) error {
				mux.HandleFunc(importProfile.RouteFamilies[0], func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})
				mux.HandleFunc(importProfile.RouteFamilies[0]+"/children", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusAccepted)
				})
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("build claimed handler: %v", err)
	}
	if got := performRequest(t, claimedHandler, http.MethodGet, importProfile.RouteFamilies[0], nil); got.Code != http.StatusNoContent {
		t.Fatalf("claimed family root must dispatch to the registered route, got %d", got.Code)
	}
	if got := performRequest(t, claimedHandler, http.MethodGet, importProfile.RouteFamilies[0]+"/children", nil); got.Code != http.StatusAccepted {
		t.Fatalf("claimed family descendant must dispatch to the registered route, got %d", got.Code)
	}

	unclaimedProjections := httpapiextensions.New([]httpapi.ExtensionProfile{
		{
			ProfileID:     importProfile.ProfileID,
			Claimed:       false,
			RouteFamilies: append([]string(nil), importProfile.RouteFamilies...),
		},
		{
			ProfileID:     enterpriseProfile.ProfileID,
			Claimed:       false,
			RouteFamilies: append([]string(nil), enterpriseProfile.RouteFamilies...),
		},
	})
	unclaimedHandler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies: unclaimedProjections.Dependencies(httpapi.DependencySet{}),
	})
	if err != nil {
		t.Fatalf("build unclaimed handler: %v", err)
	}

	rootReserved := performRequest(t, unclaimedHandler, http.MethodGet, importProfile.RouteFamilies[0], nil)
	rootBody := readJSONMap(t, rootReserved.Body.String())
	errorBody := rootBody["error"].(map[string]any)
	contracttest.RequireErrorContract(t, "extension_profile_not_claimed", http.StatusNotFound)
	if rootReserved.Code != http.StatusNotFound || errorBody["code"] != "extension_profile_not_claimed" {
		t.Fatalf("unexpected reserved root response: status=%d body=%#v", rootReserved.Code, rootBody)
	}
	rootDetails := errorBody["details"].(map[string]any)
	if rootDetails["profile_id"] != importProfile.ProfileID || rootDetails["route_family"] != importProfile.RouteFamilies[0] {
		t.Fatalf("unexpected reserved root details: %#v", rootDetails)
	}

	descendantReserved := performRequest(t, unclaimedHandler, http.MethodGet, strings.Replace(enterpriseProfile.RouteFamilies[0], "{user_id}", uuid.NewString(), 1)+"/provider", nil)
	descendantBody := readJSONMap(t, descendantReserved.Body.String())
	descendantError := descendantBody["error"].(map[string]any)
	if descendantReserved.Code != http.StatusNotFound || descendantError["code"] != "extension_profile_not_claimed" {
		t.Fatalf("unexpected reserved descendant response: status=%d body=%#v", descendantReserved.Code, descendantBody)
	}
	descendantDetails := descendantError["details"].(map[string]any)
	if descendantDetails["profile_id"] != enterpriseProfile.ProfileID || descendantDetails["route_family"] != enterpriseProfile.RouteFamilies[0] {
		t.Fatalf("unexpected reserved descendant details: %#v", descendantDetails)
	}

	outsideReserved := performRequest(t, unclaimedHandler, http.MethodGet, "/api/v1/outside-reserved-families", nil)
	if outsideReserved.Code != http.StatusNotFound || strings.Contains(outsideReserved.Body.String(), "extension_profile_not_claimed") {
		t.Fatalf("ordinary unknown paths must keep ordinary not-found handling: status=%d body=%q", outsideReserved.Code, outsideReserved.Body.String())
	}
}

func performRequest(t testing.TB, handler http.Handler, method string, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, body)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func readJSONMap(t testing.TB, body string) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode JSON body: %v", err)
	}
	return payload
}

func extensionContract(t testing.TB, profileID string) contracttest.ExtensionProfileContract {
	t.Helper()

	for _, profile := range contracttest.CurrentProfileExtensions(t) {
		if profile.ProfileID == profileID {
			return profile
		}
	}
	t.Fatalf("missing extension contract for %q", profileID)
	return contracttest.ExtensionProfileContract{}
}
