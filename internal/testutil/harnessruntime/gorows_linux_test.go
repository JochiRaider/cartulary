//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRuntimeGoRows runs Go units the way the runtime runs them: the planned
// go test command under a bound environment, its event stream settled into
// row results the runtime publishes itself (TH-HARNESS-REQ-872).
func TestRuntimeGoRows(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatal(err)
	}
	key := filepath.Join(root, masterKeyFixture)
	if err := os.MkdirAll(filepath.Dir(key), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(key, []byte("dGVzdC1rZXk=\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// The stand-in go binary records its arguments and environment, then
	// replays an event stream.
	goBinary := filepath.Join(root, "go")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" > \"$FAKE_RECORD\"\nenv >> \"$FAKE_RECORD\"\ncat \"$FAKE_EVENTS\"\nprintf 'build note\\n' >&2\nexit \"${FAKE_EXIT:-0}\"\n"
	if err := os.WriteFile(goBinary, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	environment := map[string]string{"PATH": "/usr/bin:/bin", "GO": goBinary, "GO_CACHE_DIR": "/machine/go/build", "GO_MOD_CACHE_DIR": "/machine/go/mod",
		"GO_TMP_DIR": filepath.Join(root, "tmp"), "GOROOT": "/inherited/root", "CARTULARY_HARNESS_EXECUTION_ENABLED": "1"}
	goUnit := func(args ...string) unit {
		u := engineUnit("go:unit", 1)
		u.Command = command{Executable: "go", Args: append(args, "test", "-json", "-count=1", "-p=1", "-run", "^(?:TestA|TestB)$", "./internal/fixture"), Environment: map[string]string{}}
		u.Rows = []rowBinding{{ID: "harness.fixture.unit.a", Runner: "go", Selectors: []string{"TestA"}}, {ID: "harness.fixture.unit.b", Runner: "go", Selectors: []string{"TestB"}}}
		u.Evidence = []string{"rows/harness.fixture.unit.a.json", "rows/harness.fixture.unit.b.json", "unit-results/go-unit.json"}
		u.Digest = "sha256:" + strings.Repeat("0", 64)
		return u
	}
	run := func(t *testing.T, u unit, exit string, events ...string) (unitResult, string) {
		t.Helper()
		scratch := t.TempDir()
		record, stream := filepath.Join(scratch, "record"), filepath.Join(scratch, "events")
		if err := os.WriteFile(stream, []byte(strings.Join(events, "\n")+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		u.Command.Environment = map[string]string{"FAKE_RECORD": record, "FAKE_EVENTS": stream, "FAKE_EXIT": exit}
		p := processRunner{Root: root, Environment: environment, KillGrace: 50 * time.Millisecond}
		r := p.Execute(context.Background(), executionRequest{Unit: u})
		recorded, _ := os.ReadFile(record)
		return r, string(recorded)
	}
	pass := func(test string) string { return `{"Action":"pass","Test":"` + test + `","Elapsed":0.25}` }

	t.Run("a passing unit settles its rows and retains no event output", func(t *testing.T) {
		r, recorded := run(t, goUnit(), "0", pass("TestA"), pass("TestB"), `{"Action":"pass","Package":"example/fixture","Elapsed":0.6}`)
		if r.Status != "passed" || r.ExitCode != 0 || r.Failure != nil || r.Stdout != nil || r.GoRows == nil || len(r.GoRows.Outcomes) != 2 || r.GoRows.Durations == nil {
			t.Fatalf("result %+v", r)
		}
		for _, line := range []string{"test -json -count=1 -p=1 -run ^(?:TestA|TestB)$ ./internal/fixture", "GOCACHE=/machine/go/build", "GOMODCACHE=/machine/go/mod",
			"GOTMPDIR=" + environment["GO_TMP_DIR"], "GOROOT=\n", "CARTULARY_AUTH_MASTER_KEY=dGVzdC1rZXk=\n"} {
			if !strings.Contains(recorded, line) {
				t.Fatalf("the go command lacks %q:\n%s", line, recorded)
			}
		}
	})
	t.Run("a failed row fails the unit with its detail", func(t *testing.T) {
		r, _ := run(t, goUnit(), "1", `{"Action":"output","Test":"TestA","Output":"assertion detail\n"}`, `{"Action":"fail","Test":"TestA"}`, pass("TestB"))
		if r.Status != "failed" || r.Failure.Class != "product" || r.Failure.Reason != "test_assertion_failure" || r.ExitCode != 10 || r.Stdout == nil {
			t.Fatalf("result %+v", r)
		}
		for _, want := range []string{"build note\n", "[GO FAILURE] TestA\nassertion detail\n", "[FAIL] row=harness.fixture.unit.a runner=go failure_class=product failure_reason=test_assertion_failure"} {
			if !strings.Contains(string(r.Stderr), want) {
				t.Fatalf("stderr lacks %q: %q", want, r.Stderr)
			}
		}
	})
	t.Run("an exit status alone does not fail settled rows", func(t *testing.T) {
		if r, _ := run(t, goUnit(), "1", pass("TestA"), pass("TestB")); r.Status != "passed" || r.ExitCode != 0 {
			t.Fatalf("result %+v", r)
		}
	})
	t.Run("a stream over the line contract settles no row", func(t *testing.T) {
		r, _ := run(t, goUnit(), "0", `{"Action":"output","Test":"TestA","Output":"`+strings.Repeat("x", 1<<20)+`"}`, pass("TestA"), pass("TestB"))
		if r.Status != "failed" || r.Failure.Reason != "artifact_error" || r.GoRows != nil {
			t.Fatalf("result %+v", r)
		}
	})
	t.Run("an independent module runs without the repository's test key", func(t *testing.T) {
		_, recorded := run(t, goUnit("-C", "referencepacks/format"), "0", pass("TestA"), pass("TestB"))
		if !strings.Contains(recorded, "-C referencepacks/format test -json") || strings.Contains(recorded, "CARTULARY_AUTH_MASTER_KEY=") {
			t.Fatalf("module invocation:\n%s", recorded)
		}
	})
	t.Run("an unbound machine cache is a configuration error", func(t *testing.T) {
		saved := environment["GO_CACHE_DIR"]
		environment["GO_CACHE_DIR"] = "relative/cache"
		defer func() { environment["GO_CACHE_DIR"] = saved }()
		if r, _ := run(t, goUnit(), "0", pass("TestA"), pass("TestB")); r.Status != "failed" || r.Failure.Reason != "configuration_error" {
			t.Fatalf("result %+v", r)
		}
	})
	t.Run("publication writes the rows and their durations", func(t *testing.T) {
		u := goUnit()
		r, _ := run(t, u, "1", `{"Action":"output","Test":"TestA","Output":"CARTULARY_HARNESS_TEST_FAILURE={\"broken\"\n"}`, `{"Action":"fail","Test":"TestA","Elapsed":1}`, pass("TestB"))
		runRoot := t.TempDir()
		if err := os.Chmod(runRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		publisher := &graphPublication{root: runRoot}
		if err := publisher.publishGoRows(u, &r); err != nil {
			t.Fatal(err)
		}
		var rowA, rowB struct {
			State      string  `json:"terminal_state"`
			Reason     *string `json:"failure_reason"`
			ExitCode   int     `json:"exit_code"`
			DurationMS int64   `json:"duration_ms"`
			Runner     string  `json:"runner"`
		}
		for name, into := range map[string]any{"harness.fixture.unit.a": &rowA, "harness.fixture.unit.b": &rowB} {
			data, err := os.ReadFile(filepath.Join(runRoot, "rows", name+".json"))
			if err == nil {
				err = json.Unmarshal(data, into)
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		if rowA.State != "failed" || *rowA.Reason != "scheduler_accounting_error" || rowA.ExitCode != 11 || rowA.DurationMS != 1000 || rowB.State != "passed" || rowB.Reason != nil || rowB.Runner != "go" {
			t.Fatalf("rows %+v %+v", rowA, rowB)
		}
		var durations struct {
			UnitID      string `json:"unit_id"`
			Invocations []struct {
				Package string `json:"package"`
				Tests   []any  `json:"tests"`
			} `json:"invocations"`
		}
		data, err := os.ReadFile(filepath.Join(runRoot, "unit-logs", "row-harness.fixture.unit.a", "go-test-durations.json"))
		if err == nil {
			err = json.Unmarshal(data, &durations)
		}
		if err != nil || durations.UnitID != "go:unit" || len(durations.Invocations) != 1 || durations.Invocations[0].Package != "./internal/fixture" || len(durations.Invocations[0].Tests) != 2 {
			t.Fatalf("durations %+v: %v", durations, err)
		}
	})
}
