// Package qualified exercises operation-to-writer proof independently of the
// broad call-graph fixture.
package qualified

import (
	"example.test/fixture/httpapi"
	"net/http"
)

func BindRoutes(_, _, _ string, operations map[string]http.HandlerFunc) map[string]http.HandlerFunc {
	return operations
}

func Routes() map[string]http.HandlerFunc {
	return BindRoutes("", "", "example", map[string]http.HandlerFunc{
		"declared": declared, "wrongStatus": wrongStatus, "undeclared": undeclared, "conditional": conditional,
	})
}

func forward(w http.ResponseWriter, r *http.Request, status int, code string) {
	_ = httpapi.WriteError(w, r, status, code, "", nil)
}

func declared(w http.ResponseWriter, r *http.Request)    { forward(w, r, 404, "missing") }
func wrongStatus(w http.ResponseWriter, r *http.Request) { forward(w, r, 409, "missing") }
func undeclared(w http.ResponseWriter, r *http.Request)  { forward(w, r, 404, "unknown") }
func conditional(w http.ResponseWriter, r *http.Request) {
	_ = httpapi.NotFound("unused")
	if r.Method == "DELETE" {
		forward(w, r, 400, "possible")
	}
}
