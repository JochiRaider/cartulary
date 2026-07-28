package extensiondiscovery_test

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestExtensionDiscoverySessionSliding_ServiceBacked(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "extension-discovery-session")

	_, _ = flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	userID := flowtest.SeedLocalUserFlags(
		t,
		harness.DB,
		"extension-discovery@example.test",
		"Extension Discovery",
		"ExtensionDiscovery1!",
		false,
		false,
		true,
	)
	session, _ := flowtest.LoginLocalUser(
		t,
		harness.Server.HTTP.URL,
		"extension-discovery@example.test",
		"ExtensionDiscovery1!",
		nil,
	)
	before := flowtest.QuerySessionRow(t, harness.DB, userID)

	response := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/extensions",
		nil,
		httptestx.WithCookies(session),
	)
	body := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("extension discovery data = %#v", body["data"])
	}
	items, ok := data["extensions"].([]any)
	if !ok || len(items) != 6 {
		t.Fatalf("extension discovery items = %#v", data["extensions"])
	}
	after := flowtest.QuerySessionRow(t, harness.DB, userID)
	if !after.LastQualifyingActivityAt.After(before.LastQualifyingActivityAt) ||
		!after.IdleExpiresAt.After(before.IdleExpiresAt) {
		t.Fatalf("extension discovery did not slide the session: before=%#v after=%#v", before, after)
	}
}
