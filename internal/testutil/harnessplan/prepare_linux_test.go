//go:build linux

package harnessplan

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

func TestPlanPreparation(t *testing.T) {
	root := repositoryRoot(t)
	node := filepath.Join(t.TempDir(), "node")
	if err := os.WriteFile(node, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	const row = "harness.command_surface.unit.work_graph_planner"
	// environment builds a runtime environment whose named inputs come from
	// the Make command line.
	environment := func(cli map[string]string) map[string]string {
		values := map[string]string{"NODE_BIN": node, runIDEnvironment: "prepare-test"}
		named := map[string]string{cacheModeInput: "off"}
		maps.Copy(named, cli)
		var sources []string
		for _, name := range slices.Sorted(maps.Keys(named)) {
			values[name] = named[name]
			sources = append(sources, name+"=cli")
		}
		values[makeInputSources] = strings.Join(sources, " ")
		return values
	}
	command := func(cli map[string]string) Request {
		return Request{Kind: "command", Target: "test-slice", Environment: environment(cli)}
	}

	t.Run("an owner slice of test-slice prepares a valid plan", func(t *testing.T) {
		plan, err := Prepare(root, command(map[string]string{"OWNER": "harness.command_surface", "ROWS": row}), now, 42)
		if err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(plan)
		if err != nil {
			t.Fatal(err)
		}
		document, err := harnessschema.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		if err := harnessschema.Validate(PlanSchema, document); err != nil {
			t.Fatalf("plan does not satisfy %s: %v", PlanSchema, err)
		}
		bound := 0
		for _, unit := range plan.Graph.Units {
			for _, binding := range unit.RowBindings {
				if binding.RowID == row {
					bound++
				}
			}
		}
		if bound != 1 {
			t.Fatalf("row bound %d times", bound)
		}
		m := plan.Manifest
		if m.RunID != "prepare-test" || m.CommandID == "" || m.Target != "test-slice" || m.CacheMode != "off" || m.Instrumentation.Mode != "off" {
			t.Fatalf("manifest identity = %+v", m)
		}
		for name, want := range map[string]string{"OWNER": "harness.command_surface", "ROWS": row, "SCOPE": "all", "JSON": "false", "BROWSER_LANES": "4", cacheModeInput: "off"} {
			if m.DeclaredInputs[name] != want {
				t.Errorf("declared %s = %q, want %q", name, m.DeclaredInputs[name], want)
			}
		}
		clear := plan.InputEnvironment.Clear
		if !slices.Contains(clear, "OWNER") || !slices.Contains(clear, makeInputSources) || plan.InputEnvironment.Forward[cacheModeInput] != "off" {
			t.Fatalf("input environment = %+v", plan.InputEnvironment)
		}
		if _, forwarded := plan.InputEnvironment.Forward["OWNER"]; forwarded {
			t.Fatal("a selection control reached the unit environment")
		}
		if plan.Bindings["NODE_BIN"] != node || plan.Bindings[servicesModeInput] != "owned" || !filepath.IsAbs(plan.Bindings["CARTULARY_SERVER_BIN"]) {
			t.Fatalf("bindings = %v", plan.Bindings)
		}
		if plan.HostProfile.Platform != "linux" || len(plan.HostProfile.CPUModels) == 0 || plan.Manifest.SourceCommit == "" {
			t.Fatalf("host %+v, commit %q", plan.HostProfile, plan.Manifest.SourceCommit)
		}
		var policy struct {
			RestrictedRoots []string `json:"restricted_roots"`
		}
		data, err = os.ReadFile(filepath.Join(root, executableInputPolicyPath))
		if err != nil || json.Unmarshal(data, &policy) != nil {
			t.Fatal("cannot read the executable input policy")
		}
		for _, input := range plan.SourceInputs {
			if documentation(input.Path, policy.RestrictedRoots) {
				t.Fatalf("documentation %s is in the source snapshot", input.Path)
			}
		}

		owner, err := Prepare(root, Request{Kind: "owner", Target: "test-slice", Owner: "harness.command_surface", Rows: []string{row}, Environment: environment(nil)}, now, 42)
		if err != nil {
			t.Fatal(err)
		}
		if owner.Graph.Digest != plan.Graph.Digest {
			t.Fatal("an owner request and the equivalent command compile different graphs")
		}
	})

	t.Run("a public operation command selects its operation", func(t *testing.T) {
		plan, err := Prepare(root, Request{Kind: "target", Target: "go-vulncheck", Environment: environment(nil)}, now, 42)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := plan.Graph.Units["operation:harness.go_vulncheck"]; !ok {
			t.Fatalf("units = %v", slices.Sorted(maps.Keys(plan.Graph.Units)))
		}
	})

	t.Run("normal cache mode keys units by their closures", func(t *testing.T) {
		normal := command(map[string]string{"OWNER": "harness.command_surface", "ROWS": row, cacheModeInput: "normal"})
		plan, err := Prepare(root, normal, now, 42)
		if err != nil {
			t.Fatal(err)
		}
		var fact CacheFact
		for i, unit := range plan.Graph.Sorted() {
			if len(unit.RowBindings) == 1 && unit.RowBindings[0].RowID == row {
				fact = plan.Cache[i]
			}
		}
		if !fact.Eligible || *fact.DependencyStrategy != strategyGoPackages {
			t.Fatalf("planner row fact = %+v", fact)
		}
		var closure []string
		for _, index := range plan.CacheInputs[*fact.DependencyDigest] {
			closure = append(closure, plan.SourceInputs[index].Path)
		}
		for _, member := range []string{"internal/testutil/harnessplan/compile.go", "internal/testutil/harnessschema/schema_generated.go", "go.mod", "tools/harness_cache_registry.json", "tools/test_families/harness.command_surface.json"} {
			if !slices.Contains(closure, member) {
				t.Errorf("closure lacks %s", member)
			}
		}
		if slices.ContainsFunc(closure, func(path string) bool { return strings.HasPrefix(path, "internal/modules/") }) {
			t.Error("closure includes packages the row does not import")
		}
		again, err := Prepare(root, normal, now, 42)
		if err != nil || *again.Cache[0].InputDigest != *plan.Cache[0].InputDigest {
			t.Fatal("cache keys are not deterministic")
		}

		vulncheck, err := Prepare(root, Request{Kind: "target", Target: "go-vulncheck", Environment: environment(map[string]string{cacheModeInput: "normal"})}, now, 42)
		if err != nil {
			t.Fatal(err)
		}
		// The scan declares no reusable outputs (its reports are fresh run
		// evidence), so it always runs.
		for i, unit := range vulncheck.Graph.Sorted() {
			if fact := vulncheck.Cache[i]; unit.ID == "operation:harness.go_vulncheck" && (fact.Eligible || *fact.Reason != "reusable_outputs_incomplete") {
				t.Errorf("scan fact = %+v", fact)
			}
		}
	})

	t.Run("the runtime's imposed modes override the command's inputs", func(t *testing.T) {
		request := command(map[string]string{"OWNER": "harness.command_surface", "ROWS": row, cacheModeInput: "normal"})
		request.CacheMode, request.InstrumentationMode = "off", "basic"
		plan, err := Prepare(root, request, now, 42)
		if err != nil {
			t.Fatal(err)
		}
		if plan.Manifest.CacheMode != "off" || plan.Manifest.Instrumentation.Mode != "basic" || plan.Manifest.DeclaredInputs[cacheModeInput] != "normal" {
			t.Fatalf("manifest modes %s/%s, declared %s", plan.Manifest.CacheMode, plan.Manifest.Instrumentation.Mode, plan.Manifest.DeclaredInputs[cacheModeInput])
		}
	})

	t.Run("invalid and unsupported requests are rejected before work", func(t *testing.T) {
		owned := map[string]string{"OWNER": "harness.command_surface", "ROWS": row}
		with := func(request Request, change func(map[string]string)) Request {
			request.Environment = maps.Clone(request.Environment)
			change(request.Environment)
			return request
		}
		for name, test := range map[string]struct {
			request Request
			want    error
		}{
			"aggregate selection":          {Request{Kind: "aggregate", Target: "check", Environment: environment(nil)}, ErrUnsupported},
			"unknown command":              {Request{Kind: "command", Target: "no-such-command", Environment: environment(nil)}, ErrSelection},
			"unknown owner":                {command(map[string]string{"OWNER": "harness.no_such_owner"}), ErrSelection},
			"row of another owner":         {command(map[string]string{"OWNER": "harness.test_catalog", "ROWS": row}), ErrSelection},
			"no service-backed rows":       {command(map[string]string{"OWNER": "harness.command_surface", "SCOPE": "services"}), ErrSelection},
			"shuffle seed with shell rows": {command(map[string]string{"OWNER": "harness.command_surface", "ROWS": "harness.command_surface.behavior.bootstrap_shellcheck", "GO_SHUFFLE_SEED": "7"}), ErrSelection},
			"attach outside service scope": {command(map[string]string{"OWNER": "harness.command_surface", "ROWS": row, servicesModeInput: "attach"}), ErrSelection},
			"session file in owned mode":   {command(map[string]string{"OWNER": "harness.command_surface", "ROWS": row, servicesSessionFileInput: "/tmp/session.json"}), ErrSelection},
			"reserved service state":       {with(command(owned), func(e map[string]string) { e["CARTULARY_TEST_SERVICES_SUITE_ID"] = "x" }), ErrSelection},
			"invalid run ID":               {with(command(owned), func(e map[string]string) { e[runIDEnvironment] = "../escape" }), ErrSelection},
			"missing Node runtime":         {with(command(owned), func(e map[string]string) { e["NODE_BIN"] = filepath.Join(root, "no-such-node") }), ErrSelection},
			"nested budget without tokens": {with(command(owned), func(e map[string]string) { e[hostAdmissionLease] = "lease" }), ErrSelection},
		} {
			_, err := Prepare(root, test.request, now, 42)
			if !errors.Is(err, test.want) {
				t.Errorf("%s: Prepare = %v, want %v", name, err, test.want)
			}
			t.Logf("%s: %v", name, err)
		}
	})
}
