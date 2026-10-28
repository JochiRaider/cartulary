package httptestx

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func RegisterBootstrapRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		_ = deps

		mux.HandleFunc("/api/v1/test/success", func(w http.ResponseWriter, r *http.Request) {
			_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
				"service": "bootstrap",
				"status":  "ok",
			})
		})

		mux.HandleFunc("/api/v1/test/error", func(w http.ResponseWriter, r *http.Request) {
			_ = httpapi.WriteError(w, r, http.StatusServiceUnavailable, "bootstrap_error", "bootstrap harness error", map[string]any{
				"reason_code": "bootstrap_unavailable",
				"scope":       "harness",
			})
		})

		return nil
	}
}
