package server

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
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
	registrars, err := newApplicationRouteCatalog(catalog).Bind(contributions, nil)
	if err != nil {
		t.Fatalf("built-in route contributions rejected: %v", err)
	}
	if len(registrars) != len(requiredBuiltInRouteContributionIDs) {
		t.Fatalf("built-in registrar count got %d want %d", len(registrars), len(requiredBuiltInRouteContributionIDs))
	}

	omitted := append([]routeContribution(nil), contributions[:len(contributions)-1]...)
	if _, err := newApplicationRouteCatalog(catalog).Bind(omitted, nil); err == nil {
		t.Fatal("expected omitted built-in route contribution to fail")
	}
	reordered := append([]routeContribution(nil), contributions...)
	reordered[0], reordered[1] = reordered[1], reordered[0]
	if _, err := newApplicationRouteCatalog(catalog).Bind(reordered, nil); err == nil {
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
	networkFlowRouteCount := 0
	bindings := make([]extensionRouteBinding, 0, len(routeIDs))
	for _, contributionID := range routeIDs {
		if contributionID == networkflow.RouteContributionID {
			networkFlowRouteCount++
		}
		bindings = append(bindings, extensionRouteBinding{
			id:              contributionID,
			contributionIDs: []string{contributionID},
			registrar:       registrar,
		})
	}
	if networkFlowRouteCount != 1 {
		t.Fatalf("Network Flow route contribution count got %d want 1", networkFlowRouteCount)
	}
	registrars, err := newApplicationRouteCatalog(catalog).Bind(base, bindings)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(registrars), len(base)+len(routeIDs); got != want {
		t.Fatalf("registrar count got %d want %d", got, want)
	}
	if _, err := newApplicationRouteCatalog(catalog).Bind(base, bindings[:len(bindings)-1]); err == nil {
		t.Fatal("expected missing claimed contribution binding to fail")
	}
	duplicate := append([]extensionRouteBinding(nil), bindings...)
	duplicate = append(duplicate, bindings[0])
	if _, err := newApplicationRouteCatalog(catalog).Bind(base, duplicate); err == nil {
		t.Fatal("expected duplicate claimed contribution binding to fail")
	}
}

func TestRuntimeCleanupIsReverseOrderedAndIdempotent(t *testing.T) {
	t.Parallel()
	var calls []string
	runtime := &Runtime{}
	runtime.own("first", func(context.Context) error { calls = append(calls, "first"); return nil })
	runtime.own("second", func(context.Context) error { calls = append(calls, "second"); return nil })
	runtime.Close()
	runtime.Close()
	if want := []string{"second", "first"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("cleanup calls got %#v want %#v", calls, want)
	}
}

// TestRuntimeShutdownReportsComponentsOpenAtTheDeadline covers REQ-04-146a:
// every component close shares the caller's deadline, and the components not
// closed by then are reported in closing order instead of being waited on.
func TestRuntimeShutdownReportsComponentsOpenAtTheDeadline(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	defer close(release)
	var calls []string
	runtime := &Runtime{}
	runtime.own("postgres", func(context.Context) error { calls = append(calls, "postgres"); return nil })
	runtime.own("jobs_runner", func(context.Context) error { <-release; return nil })
	runtime.own("collaboration", func(ctx context.Context) error {
		if _, ok := ctx.Deadline(); !ok {
			t.Error("component close received no deadline")
		}
		calls = append(calls, "collaboration")
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	open := runtime.shutdown(ctx)
	if want := []string{"jobs_runner", "postgres"}; !reflect.DeepEqual(open, want) {
		t.Fatalf("open components got %#v want %#v", open, want)
	}
	if want := []string{"collaboration"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("closed components got %#v want %#v", calls, want)
	}
	if again := runtime.shutdown(context.Background()); again != nil {
		t.Fatalf("a repeated shutdown must be idempotent, got %#v", again)
	}
}

// TestRuntimeShutdownStartsBackgroundClosesAtTheStopRequest covers
// REQ-04-146a: background components start closing when the HTTP drain begins,
// a failed close leaves its component open, and a close already finished
// counts as closed even when the deadline has expired.
func TestRuntimeShutdownStartsBackgroundClosesAtTheStopRequest(t *testing.T) {
	t.Parallel()
	started := make(chan struct{})
	finished := make(chan struct{})
	runtime := &Runtime{}
	runtime.own("postgres", func(context.Context) error { return nil })
	runtime.own("telemetry", func(context.Context) error { return errors.New("flush failed") })
	runtime.ownBackground("jobs_runner", func(ctx context.Context) error {
		close(started)
		defer close(finished)
		if _, ok := ctx.Deadline(); !ok {
			t.Error("background close received no application deadline")
		}
		return nil
	})
	beginCtx, cancelBegin := context.WithTimeout(context.Background(), time.Minute)
	defer cancelBegin()
	runtime.beginShutdown(beginCtx)
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("background component did not start closing at the stop request")
	}
	<-finished
	// The deadline has already expired: a finished close still counts as
	// closed, and a component that was never closed is open.
	expired, cancel := context.WithCancel(context.Background())
	cancel()
	open := runtime.shutdown(expired)
	if want := []string{"telemetry", "postgres"}; !reflect.DeepEqual(open, want) {
		t.Fatalf("open components got %#v want %#v", open, want)
	}

	failing := &Runtime{}
	failing.own("postgres", func(context.Context) error { return nil })
	failing.own("telemetry", func(context.Context) error { return errors.New("flush failed") })
	ctx, cancelCtx := context.WithTimeout(context.Background(), time.Minute)
	defer cancelCtx()
	if open := failing.shutdown(ctx); !reflect.DeepEqual(open, []string{"telemetry"}) {
		t.Fatalf("a failed close must leave its component open, got %#v", open)
	}
}

// TestRuntimeStopRequestClosesAdmission covers Core 04 REQ-04-146a and
// EXT-REQ-193: at the stop request, before any drain, the process closes
// admission and readiness. A serving process moves to quiescing, never
// reopens admission and still lets a fatal take precedence without leaving
// the drain; a process still starting goes straight to terminating and can no
// longer publish. A repeated stop request changes nothing.
func TestRuntimeStopRequestClosesAdmission(t *testing.T) {
	t.Parallel()
	stop := func(runtime *Runtime) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		runtime.beginShutdown(ctx)
	}

	t.Run("a serving process stops admitting at the stop request", func(t *testing.T) {
		t.Parallel()
		lifecycle := processlifecycle.New()
		if err := lifecycle.Publish(); err != nil {
			t.Fatal(err)
		}
		runtime := &Runtime{lifecycle: lifecycle}
		stop(runtime)
		if lifecycle.AdmissionOpen() || lifecycle.State() != processlifecycle.StateQuiescing {
			t.Fatalf("after the stop request admission open=%t state=%s, want closed and quiescing", lifecycle.AdmissionOpen(), lifecycle.State())
		}
		stop(runtime)
		if lifecycle.RestoreAdmission() || lifecycle.AdmissionOpen() || lifecycle.AdmissionPaused() {
			t.Fatal("a stopping process reopened or paused admission")
		}
		if !lifecycle.Fatal("published_component_lost") || lifecycle.State() != processlifecycle.StateQuiescing {
			t.Fatalf("a fatal during ordinary shutdown left state %s, want quiescing", lifecycle.State())
		}
	})
	t.Run("a starting process terminates at the stop request", func(t *testing.T) {
		t.Parallel()
		lifecycle := processlifecycle.New()
		runtime := &Runtime{lifecycle: lifecycle}
		stop(runtime)
		if lifecycle.State() != processlifecycle.StateTerminating {
			t.Fatalf("after the stop request a starting process is %s, want terminating", lifecycle.State())
		}
		if err := lifecycle.Publish(); err == nil || lifecycle.AdmissionOpen() {
			t.Fatal("publication opened admission after the stop request")
		}
	})
}
