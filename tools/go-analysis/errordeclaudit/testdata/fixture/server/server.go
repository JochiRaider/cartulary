// Package server is a fixture whose handlers produce declared codes, one
// seeded undeclared code reached through a forwarding function, and an
// operation the binding never roots.
package server

import (
	"net/http"

	"example.test/fixture/httpapi"
)

type service struct{}

func (s *service) operationHandlers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"getThing": s.handleGet,
		"putThing": s.handlePut,
	}
}

func (s *service) handleGet(w http.ResponseWriter, r *http.Request) {
	code := "thing_not_found"
	if r.URL.Path == "" {
		code = "internal_error"
	}
	httpapi.WriteAPIError(w, r, httpapi.NotFound(code))
	_ = httpapi.WriteError(w, r, http.StatusInternalServerError, "internal_error", "internal error", nil)
}

func (s *service) handlePut(w http.ResponseWriter, r *http.Request) {
	httpapi.WriteAPIError(w, r, reject("seeded_undeclared"))
}

// reject forwards its code through a second carrier.
func reject(code string) *httpapi.APIError {
	return httpapi.NotFound(code)
}

// Handlers is the fixture program's entry point.
func Handlers() map[string]http.HandlerFunc {
	return (&service{}).operationHandlers()
}
