package harnessruntime

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestRetiredPublicErrorControlUnavailable(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[enabled], func(t *testing.T) {
			env := map[string]string{}
			if enabled {
				env = testRuntimeEnabledEnv()
			}
			server := startTestRuntimeIdentityServer(t, env, RegisterRoutes(httpapi.NewTestClock())...)
			// A valid enabled identity route proves the control plane was actually mounted.
			identity := doTestRuntimeResetRequest(t, server.Client(), authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodGet, server.URL+"/api/v1/test/runtime/identity", nil)))
			if enabled {
				requireTestRuntimeResetSuccessEnvelope(t, identity, http.StatusOK)
			} else {
				requireTestRuntimeResetStatus(t, identity, http.StatusNotFound)
				identity.Body.Close()
			}
			for _, authorized := range []bool{false, true} {
				request := newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/public-error-faults", map[string]any{})
				if authorized {
					request = authorizeTestRuntimeResetRequest(request)
				}
				response := doTestRuntimeResetRequest(t, server.Client(), request)
				requireTestRuntimeResetStatus(t, response, http.StatusNotFound)
				response.Body.Close()
			}
		})
	}
}
