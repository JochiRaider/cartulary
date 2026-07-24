package server

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
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

	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := coordinator.ResolveClaims(nil)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := extensionassembly.NewPublicationCatalog(plan, coordinator.ParticipantContracts())
	if err != nil {
		t.Fatal(err)
	}
	registrars, err := applicationRouteRegistrars(contributions, nil, catalog)
	if err != nil {
		t.Fatalf("built-in route contributions rejected: %v", err)
	}
	if len(registrars) != len(requiredBuiltInRouteContributionIDs) {
		t.Fatalf("built-in registrar count got %d want %d", len(registrars), len(requiredBuiltInRouteContributionIDs))
	}

	omitted := append([]routeContribution(nil), contributions[:len(contributions)-1]...)
	if _, err := applicationRouteRegistrars(omitted, nil, catalog); err == nil {
		t.Fatal("expected omitted built-in route contribution to fail")
	}
	reordered := append([]routeContribution(nil), contributions...)
	reordered[0], reordered[1] = reordered[1], reordered[0]
	if _, err := applicationRouteRegistrars(reordered, nil, catalog); err == nil {
		t.Fatal("expected reordered built-in route contribution to fail")
	}
}

func TestApplicationRouteContributionsAreExactCatalogProjection(t *testing.T) {
	t.Parallel()
	registrar := httpapi.RouteRegistrar(func(*http.ServeMux, httpapi.DependencySet) error { return nil })
	base := make([]routeContribution, 0, len(requiredBuiltInRouteContributionIDs))
	for _, id := range requiredBuiltInRouteContributionIDs {
		base = append(base, routeContribution{id: id, registrar: registrar})
	}
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	profileIDs := []string{}
	for _, descriptor := range coordinator.Descriptors() {
		profileIDs = append(profileIDs, descriptor.ProfileID)
	}
	resolution, err := coordinator.ResolveClaims(profileIDs)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := extensionassembly.NewPublicationCatalog(plan, coordinator.ParticipantContracts())
	if err != nil {
		t.Fatal(err)
	}
	routeIDs := catalog.ContributionIDs("http_route_family")
	bindings := make([]extensionRouteBinding, 0, len(routeIDs))
	for _, contributionID := range routeIDs {
		bindings = append(bindings, extensionRouteBinding{
			id:              contributionID,
			contributionIDs: []string{contributionID},
			registrar:       registrar,
		})
	}
	registrars, err := applicationRouteRegistrars(base, bindings, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(registrars), len(base)+len(routeIDs); got != want {
		t.Fatalf("registrar count got %d want %d", got, want)
	}
	if _, err := applicationRouteRegistrars(base, bindings[:len(bindings)-1], catalog); err == nil {
		t.Fatal("expected missing claimed contribution binding to fail")
	}
	duplicate := append([]extensionRouteBinding(nil), bindings...)
	duplicate = append(duplicate, bindings[0])
	if _, err := applicationRouteRegistrars(base, duplicate, catalog); err == nil {
		t.Fatal("expected duplicate claimed contribution binding to fail")
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
