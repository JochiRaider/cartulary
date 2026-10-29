package serverprocess

import (
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"net/http"
	"strings"
	"testing"
)

func TestUnclaimedReservedFamiliesReturnCanonical404AndOutsidePathsDoNot_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "incident_membership-i-2-06", TestRouteMode: httptestx.TestRouteModeDisabled})

	_, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	flowtest.SeedLocalUserFlags(t, harness.DB, "reserved-user@example.test", "Reserved User", "ReservedUser1!", false, false, true)
	userLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "reserved-user@example.test", "ReservedUser1!", nil)
	enterpriseProfile := ExtensionContract(t, "enterprise_authentication")
	networkFlowProfile := ExtensionContract(t, "network_flow_activity")

	rootReserved := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+enterpriseProfile.RouteFamilies[0], nil, httptestx.WithCookies(userLogin.Session))
	rootReservedBody := httptestx.RequireErrorEnvelope(t, rootReserved, http.StatusNotFound, "extension_profile_not_claimed")
	rootDetails := rootReservedBody["error"].(map[string]any)["details"].(map[string]any)
	if rootDetails["profile_id"] != enterpriseProfile.ProfileID || rootDetails["route_family"] != enterpriseProfile.RouteFamilies[0] {
		t.Fatalf("unexpected reserved root dispatch details: %#v", rootDetails)
	}

	networkFlowReserved := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+strings.Replace(networkFlowProfile.RouteFamilies[0], "{incident_id}", adminID.String(), 1),
		nil,
		httptestx.WithCookies(userLogin.Session),
	)
	networkFlowBody := httptestx.RequireErrorEnvelope(t, networkFlowReserved, http.StatusNotFound, "extension_profile_not_claimed")
	networkFlowDetails := networkFlowBody["error"].(map[string]any)["details"].(map[string]any)
	if networkFlowDetails["profile_id"] != networkFlowProfile.ProfileID || networkFlowDetails["route_family"] != networkFlowProfile.RouteFamilies[0] {
		t.Fatalf("unexpected Network Flow reserved root dispatch details: %#v", networkFlowDetails)
	}

	descendantReserved := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+strings.Replace(enterpriseProfile.RouteFamilies[0], "{user_id}", adminID.String(), 1)+"/provider",
		nil,
		httptestx.WithCookies(userLogin.Session),
	)
	descendantBody := httptestx.RequireErrorEnvelope(t, descendantReserved, http.StatusNotFound, "extension_profile_not_claimed")
	descendantDetails := descendantBody["error"].(map[string]any)["details"].(map[string]any)
	if descendantDetails["profile_id"] != enterpriseProfile.ProfileID || descendantDetails["route_family"] != enterpriseProfile.RouteFamilies[0] {
		t.Fatalf("unexpected reserved descendant dispatch details: %#v", descendantDetails)
	}

	outsideReserved := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/outside-reserved-families", nil, httptestx.WithCookies(userLogin.Session))
	bodyText := string(httptestx.ReadResponseBody(t, outsideReserved, http.StatusNotFound))
	if strings.Contains(bodyText, "extension_profile_not_claimed") {
		t.Fatalf("outside reserved families must keep ordinary 404 handling, got %q", bodyText)
	}
}

func ExtensionContract(t testing.TB, profileID string) contracttest.ExtensionProfileContract {
	t.Helper()
	for _, profile := range contracttest.CurrentProfileExtensions(t) {
		if profile.ProfileID == profileID {
			return profile
		}
	}
	t.Fatalf("missing extension contract for %q", profileID)
	return contracttest.ExtensionProfileContract{}
}
