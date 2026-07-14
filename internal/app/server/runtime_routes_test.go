package server

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestRouteContributionValidation(t *testing.T) {
	t.Parallel()
	registrar := httpapi.RouteRegistrar(func(*http.ServeMux, httpapi.DependencySet) error { return nil })

	tests := []struct {
		name          string
		contributions []routeContribution
		wantError     bool
	}{
		{name: "valid", contributions: []routeContribution{{id: "one", registrar: registrar}, {id: "two", registrar: registrar}}},
		{name: "missing id", contributions: []routeContribution{{registrar: registrar}}, wantError: true},
		{name: "missing registrar", contributions: []routeContribution{{id: "one"}}, wantError: true},
		{name: "duplicate id", contributions: []routeContribution{{id: "one", registrar: registrar}, {id: "one", registrar: registrar}}, wantError: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			registrars, err := routeRegistrars(test.contributions)
			if test.wantError {
				if err == nil {
					t.Fatal("expected route contribution error")
				}
				return
			}
			if err != nil {
				t.Fatalf("route contributions rejected: %v", err)
			}
			if len(registrars) != len(test.contributions) {
				t.Fatalf("registrars got %d want %d", len(registrars), len(test.contributions))
			}
		})
	}
}

func TestBuiltInRouteContributionMembershipAndOrder(t *testing.T) {
	t.Parallel()
	registrar := httpapi.RouteRegistrar(func(*http.ServeMux, httpapi.DependencySet) error { return nil })
	contributions := make([]routeContribution, 0, len(requiredBuiltInRouteContributionIDs))
	for _, id := range requiredBuiltInRouteContributionIDs {
		contributions = append(contributions, routeContribution{id: id, registrar: registrar})
	}

	registrars, err := builtInRouteRegistrars(contributions)
	if err != nil {
		t.Fatalf("built-in route contributions rejected: %v", err)
	}
	if len(registrars) != 19 {
		t.Fatalf("built-in registrar count got %d want 19", len(registrars))
	}

	omitted := append([]routeContribution(nil), contributions[:len(contributions)-1]...)
	if _, err := builtInRouteRegistrars(omitted); err == nil {
		t.Fatal("expected omitted built-in route contribution to fail")
	}
	reordered := append([]routeContribution(nil), contributions...)
	reordered[0], reordered[1] = reordered[1], reordered[0]
	if _, err := builtInRouteRegistrars(reordered); err == nil {
		t.Fatal("expected reordered built-in route contribution to fail")
	}
}

func TestRuntimeCleanupIsReverseOrderedAndIdempotent(t *testing.T) {
	t.Parallel()
	var calls []string
	runtime := &Runtime{}
	runtime.own(func() { calls = append(calls, "first") })
	runtime.own(func() { calls = append(calls, "second") })
	runtime.Close()
	runtime.Close()
	if want := []string{"second", "first"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("cleanup calls got %#v want %#v", calls, want)
	}
}
