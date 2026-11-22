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

func ConditionalRoutes() map[string]http.HandlerFunc {
	return BindRoutes("", "", "example", map[string]http.HandlerFunc{
		"queryDeclared": queryDeclared, "queryWrongStatus": queryWrongStatus, "queryUnknown": queryUnknown,
		"infeasible": infeasible, "discarded": discarded, "reassigned": reassigned, "escaped": escaped,
		"otherRequest": otherRequest, "otherWriter": otherWriter, "earlierResponse": earlierResponse,
		"ambiguous": queryDeclared, "absentParameter": absentParameter,
	})
}

func AmbiguousRoutes() map[string]http.HandlerFunc {
	return BindRoutes("", "", "example", map[string]http.HandlerFunc{"ambiguous": queryWrongStatus})
}

func queryError(r *http.Request, key string, status int, code string) *httpapi.APIError {
	if _, ok := r.URL.Query()[key]; ok {
		return &httpapi.APIError{Status: status, Code: code}
	}
	return nil
}

func write(w http.ResponseWriter, r *http.Request, e *httpapi.APIError) {
	forward(w, r, e.Status, e.Code)
}

func queryDeclared(w http.ResponseWriter, r *http.Request) {
	if e := queryError(r, "limit", 404, "missing"); e != nil {
		write(w, r, e)
	}
}
func queryWrongStatus(w http.ResponseWriter, r *http.Request) {
	if e := queryError(r, "limit", 409, "missing"); e != nil {
		write(w, r, e)
	}
}
func queryUnknown(w http.ResponseWriter, r *http.Request) {
	if e := queryError(r, "limit", 404, "unknown"); e != nil {
		write(w, r, e)
	}
}
func infeasible(w http.ResponseWriter, r *http.Request) {
	if e := queryError(r, "absent", 409, "missing"); e != nil {
		write(w, r, e)
	}
}
func discarded(w http.ResponseWriter, r *http.Request) {
	_ = queryError(r, "limit", 409, "missing")
}
func reassigned(w http.ResponseWriter, r *http.Request) {
	if e := queryError(r, "limit", 404, "missing"); e != nil {
		e.Status = 409
		write(w, r, e)
	}
}

var saved *httpapi.APIError

func escaped(w http.ResponseWriter, r *http.Request) {
	if e := queryError(r, "limit", 404, "missing"); e != nil {
		saved = e
		write(w, r, e)
	}
}
func otherRequest(w http.ResponseWriter, r *http.Request) {
	if e := queryError(r, "limit", 409, "missing"); e != nil {
		write(w, &http.Request{}, e)
	}
}
func otherWriter(w http.ResponseWriter, r *http.Request) {
	if e := queryError(r, "limit", 409, "missing"); e != nil {
		write(nil, r, e)
	}
}
func earlierResponse(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	queryWrongStatus(w, r)
}

func PathString(*http.Request, string) (string, bool) { return "", false }

func absentParameter(w http.ResponseWriter, r *http.Request) {
	if _, ok := PathString(r, ""); ok {
		forward(w, r, 409, "missing")
	}
}
