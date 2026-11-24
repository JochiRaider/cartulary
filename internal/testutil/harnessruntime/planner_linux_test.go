//go:build linux

package harnessruntime

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessplan"
)

func TestGoPlannerAdoption(t *testing.T) {
	root := repositoryRootForPlanner(t)
	node := filepath.Join(t.TempDir(), "node")
	if err := os.WriteFile(node, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	const owner = "module.reportcomposition"
	environment := map[string]string{"NODE_BIN": node, "CARTULARY_TEST_RUN_ID": "adoption-test", "CARTULARY_HARNESS_CACHE_MODE": "off",
		"CARTULARY_MAKE_INPUT_SOURCES": "CARTULARY_HARNESS_CACHE_MODE=cli"}

	t.Run("a Go plan becomes the runtime's execution model", func(t *testing.T) {
		request, err := planRequest([]string{"--selection", "owner", "--target", "test-slice", "--owner", owner}, environment)
		if err != nil {
			t.Fatal(err)
		}
		source, err := harnessplan.Prepare(root, request, time.Now(), 42)
		if err != nil {
			t.Fatal(err)
		}
		plan, err := adoptPlan(source)
		if err != nil {
			t.Fatal(err)
		}
		if plan.published.name != "plan.json" || plan.published.schema != harnessplan.PlanSchema || plan.GraphDigest() != source.Graph.Digest {
			t.Fatalf("published %s %s", plan.published.name, plan.published.schema)
		}
		units := plan.facts.Graph.Units
		if len(units) != len(source.Graph.Units) || len(plan.facts.Cache) != len(units) || plan.facts.Manifest["run_id"] != "adoption-test" {
			t.Fatalf("adopted %d units and %d cache facts from %d", len(units), len(plan.facts.Cache), len(source.Graph.Units))
		}
		residents := 0
		for i, u := range units {
			if i > 0 && units[i-1].ID >= u.ID {
				t.Fatal("adopted units are not in ID order")
			}
			planned := source.Graph.Units[u.ID]
			if u.Residents == nil || len(u.Residents) != len(planned.Residents) || u.Physical != nil || !u.FailurePolicy.ContinueIndependent {
				t.Fatalf("%s residents %v from %v", u.ID, u.Residents, planned.Residents)
			}
			residents += len(u.Residents)
			if strings.HasPrefix(u.ID, "go:") && (len(u.Command.Args) != 3 || u.Command.Args[1] != "--row-ids") {
				t.Fatalf("%s command %v does not carry its rows", u.ID, u.Command.Args)
			}
			for _, name := range []string{"CARTULARY_SERVER_BIN", "CARTULARY_MIGRATE_BIN"} {
				if value, ok := u.Command.Environment[name]; !ok || value != "" && !filepath.IsAbs(value) {
					t.Fatalf("%s runtime binary %s is unbound: %q", u.ID, name, value)
				}
			}
		}
		if residents == 0 {
			t.Fatal("no adopted unit holds a resident")
		}
	})

	t.Run("graph arguments become a planner request", func(t *testing.T) {
		request, err := planRequest([]string{"--selection", "owner", "--target", "test-slice", "--owner", owner, "--rows", "b,a", "--service-backed-only"}, environment)
		if err != nil || request.Kind != "owner" || request.Owner != owner || !slices.Equal(request.Rows, []string{"b", "a"}) || !request.ServiceBackedOnly {
			t.Fatalf("request = %+v, %v", request, err)
		}
		if _, err := planRequest([]string{"--plan", "x"}, environment); err == nil {
			t.Fatal("an unknown argument was accepted")
		}
	})

	t.Run("planner rejections become closed preparation diagnostics", func(t *testing.T) {
		for err, want := range map[error]string{
			fmt.Errorf("%w: unknown owner", harnessplan.ErrSelection):         "failure_reason=usage_error cause=command_selection",
			fmt.Errorf("%w: duplicate row", harnessplan.ErrInput):             "failure_reason=configuration_error cause=catalog_shape",
			errors.New("a\nmultiline " + strings.Repeat("x", 400) + " error"): "failure_reason=configuration_error cause=unclassified",
		} {
			lines := PreparationDiagnosticLines(plannerFailure(err))
			if len(lines) != 2 || !strings.Contains(lines[0], want) || strings.ContainsAny(lines[1], "\n\r") || len(lines[1]) > len("[GRAPH-PREPARE] subject=")+256 {
				t.Errorf("%v: %q", err, lines)
			}
		}
	})

	t.Run("an attached session's shared allocations are not charged", func(t *testing.T) {
		attached := map[string]string{"CARTULARY_TEST_SERVICES_MODE": "attach", "CARTULARY_TEST_SERVICES_SESSION_FILE": filepath.Join(t.TempDir(), "session.json")}
		for name, value := range environment {
			attached[name] = value
		}
		attached["CARTULARY_MAKE_INPUT_SOURCES"] += " CARTULARY_TEST_SERVICES_MODE=cli CARTULARY_TEST_SERVICES_SESSION_FILE=cli"
		request, err := planRequest([]string{"--selection", "owner", "--target", "test-slice", "--owner", owner, "--service-backed-only"}, attached)
		if err != nil {
			t.Fatal(err)
		}
		source, err := harnessplan.Prepare(root, request, time.Now(), 42)
		if err != nil {
			t.Fatal(err)
		}
		plan, err := adoptPlan(source)
		if err != nil {
			t.Fatal(err)
		}
		scopes := map[string]string{}
		for _, u := range plan.facts.Graph.Units {
			for _, resident := range u.Residents {
				scopes[strings.SplitN(resident.Key, ":", 3)[1]] = resident.Scope
			}
		}
		if scopes["shared"] != harnessfixtures.ResidentAttached || scopes["unit"] != "" {
			t.Fatalf("resident scopes by kind = %v", scopes)
		}
	})
}

// repositoryRootForPlanner walks up from the package directory to the module root.
func repositoryRootForPlanner(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		if parent := filepath.Dir(dir); parent != dir {
			dir = parent
			continue
		}
		t.Fatal("module root not found")
	}
}
