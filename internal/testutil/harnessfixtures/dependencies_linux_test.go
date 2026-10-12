//go:build linux

package harnessfixtures

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func dependencyFixture(kind, name string, dependencies []string) harnesssession.Resource {
	target, _ := json.Marshal(name)
	return harnesssession.Resource{Kind: kind, Target: target, RunID: "fixture", LeaseID: "00000000-0000-4000-8000-000000000000", State: "acquired", Dependencies: dependencies}
}

func testCleanupDependencies(t *testing.T) {
	for _, kind := range []string{"managed_suite", "borrowed_suite", "browser_stack", "object_store_proxy", "helper_process", "browser_process", "diagnostic_scope", "browser_renderer"} {
		t.Run("all_kinds_block_only_declared_dependencies_"+kind, func(t *testing.T) {
			suite := dependencyFixture("managed_suite", "suite", []string{})
			other := dependencyFixture("borrowed_suite", "independent", []string{})
			consumer := dependencyFixture(kind, "consumer", []string{suite.Reference()})
			for _, target := range []string{"", suite.Reference()} {
				calls := []string{}
				outcomes := map[string]int{}
				err := evaluateCleanup([]harnesssession.Resource{suite, consumer, other}, target, func(r harnesssession.Resource) error {
					calls = append(calls, r.Reference())
					if r.Reference() == consumer.Reference() {
						return errors.New("consumer failed")
					}
					return nil
				}, func(o CleanupOutcome) { outcomes[o.Outcome]++ })
				if !errors.Is(err, ErrCleanupBlocked) || slices.Contains(calls, suite.Reference()) || outcomes["failed"] != 1 || outcomes["blocked"] != 1 {
					t.Fatal("dependency decision diverged", calls, outcomes, err)
				}
				if (target == "") != slices.Contains(calls, other.Reference()) {
					t.Fatal("independent progress/selection lost", calls)
				}
			}
		})
	}
	t.Run("explicit_independence_does_not_block", func(t *testing.T) {
		consumer := dependencyFixture("diagnostic_scope", "scope", []string{})
		suite := dependencyFixture("managed_suite", "suite", []string{})
		calls := 0
		err := evaluateCleanup([]harnesssession.Resource{consumer, suite}, "", func(r harnesssession.Resource) error {
			if r.Kind == "diagnostic_scope" {
				return errors.New("scope failure")
			}
			calls++
			return nil
		}, nil)
		if err == nil || errors.Is(err, ErrCleanupBlocked) || calls != 1 {
			t.Fatal("independence was not preserved", calls, err)
		}
	})
	t.Run("unknown_and_foreign_edges_are_not_independence", func(t *testing.T) {
		for _, dependencies := range [][]string{nil, {"foreign-resource"}} {
			consumer := dependencyFixture("diagnostic_scope", "scope", dependencies)
			suite := dependencyFixture("managed_suite", "suite", []string{})
			err := evaluateCleanup([]harnesssession.Resource{consumer, suite}, "", func(harnesssession.Resource) error { t.Fatal("uncertain graph permitted cleanup"); return nil }, nil)
			if !errors.Is(err, errCleanupUncertain) || !errors.Is(err, ErrCleanupBlocked) {
				t.Fatal(err)
			}
		}
	})
	t.Run("cycles_preserve_connected_resources_and_allow_independent_progress", func(t *testing.T) {
		a := dependencyFixture("browser_renderer", "a", []string{})
		b := dependencyFixture("diagnostic_scope", "b", []string{a.Reference()})
		a.Dependencies = []string{b.Reference()}
		c := dependencyFixture("borrowed_suite", "independent", []string{})
		calls := 0
		err := evaluateCleanup([]harnesssession.Resource{a, b, c}, "", func(r harnesssession.Resource) error {
			if r.Reference() != c.Reference() {
				t.Fatal("cyclic resource closed")
			}
			calls++
			return nil
		}, nil)
		if !errors.Is(err, ErrCleanupBlocked) || calls != 1 {
			t.Fatal(calls, err)
		}
	})
	t.Run("consumer_first_across_partial_acquisition", func(t *testing.T) {
		a := dependencyFixture("managed_suite", "a", []string{})
		b := dependencyFixture("browser_stack", "b", []string{a.Reference()})
		b.State = "pending"
		c := dependencyFixture("browser_renderer", "c", []string{b.Reference()})
		calls := []string{}
		err := evaluateCleanup([]harnesssession.Resource{a, b, c}, "", func(r harnesssession.Resource) error { calls = append(calls, r.Reference()); return nil }, nil)
		if err != nil || !slices.Equal(calls, []string{c.Reference(), b.Reference(), a.Reference()}) {
			t.Fatal(calls, err)
		}
	})
}
