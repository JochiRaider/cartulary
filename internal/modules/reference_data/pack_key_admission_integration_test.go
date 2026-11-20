package reference_data_test

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestReferencePackMalformedPackKeys_Integration verifies that a pack key
// outside the canonical key syntax is a statement about the request, never a
// server failure (REQ-01-706). A malformed refresh selector member fails with
// 400 invalid_reference_pack_request/invalid_pack_keys (AC-369, AC-614). A
// malformed reverify path key violates the pack_key parameter's declared
// pattern and receives the route's not-found response, identical to a
// well-formed key that names no pack (REQ-01-707). Neither admits a Job or
// changes pack or admission state.
func TestReferencePackMalformedPackKeys_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	h := startReferencePackServer(t, runtime, "reference-malformed-pack-keys")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	malformed := []string{"TYPE_REGISTRY.HOST", "type_registry", "type-registry.host", "type_registry.", "a." + strings.Repeat("b", 33), "a.b.c.d.e.f.g.h.i"}
	before := packAdmissionState(t, h)

	for index, key := range malformed {
		t.Run("refresh/"+key, func(t *testing.T) {
			response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+"/api/v1/reference-packs/refresh", map[string]any{
				"client_txn_id": "malformed-refresh-" + string(rune('a'+index)),
				"pack_keys":     []string{"type_registry.host", key},
			}, csrfOptions(admin)...)
			details := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_reference_pack_request")["error"].(map[string]any)["details"]
			if want := map[string]any{"field": "pack_keys", "reason_code": "invalid_pack_keys"}; !reflect.DeepEqual(details, want) {
				t.Fatalf("refresh details = %#v, want %#v", details, want)
			}
		})
	}

	missing := httptestx.RequireErrorEnvelope(t, postAction(t, h, admin, "/api/v1/reference-packs/enrichment.absent/1/reverify", "missing-reverify", ""), http.StatusNotFound, "reference_pack_not_found")["error"].(map[string]any)["details"]
	for _, key := range malformed {
		t.Run("reverify/"+key, func(t *testing.T) {
			response := postAction(t, h, admin, "/api/v1/reference-packs/"+key+"/1/reverify", "malformed-reverify", "")
			details := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "reference_pack_not_found")["error"].(map[string]any)["details"]
			if !reflect.DeepEqual(details, missing) {
				t.Fatalf("malformed key details = %#v, want the missing-key details %#v", details, missing)
			}
		})
	}

	if after := packAdmissionState(t, h); after != before {
		t.Fatal("malformed pack keys changed pack or admission state")
	}
}
