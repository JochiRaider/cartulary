package harnessrows

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

type event map[string]any

func stream(rows []Row, events ...any) (Result, error) {
	s := NewStream(rows)
	for _, e := range events {
		var line []byte
		switch v := e.(type) {
		case string:
			line = []byte(v)
		default:
			line, _ = json.Marshal(v)
		}
		_, _ = s.Write(append(line, '\n'))
	}
	return s.Finish("./internal/fixture")
}

var rowA = []Row{{ID: "row-a", Selectors: []string{"TestA"}}}

func only(t *testing.T, rows []Row, events ...any) Outcome {
	t.Helper()
	result, err := stream(rows, events...)
	if err != nil {
		t.Fatal(err)
	}
	return result.Outcomes[0]
}

func markerOutput(test string, envelope any) event {
	encoded, _ := json.Marshal(envelope)
	return event{"Action": "output", "Test": test, "Output": "CARTULARY_HARNESS_TEST_FAILURE=" + string(encoded) + "\n"}
}

func with(base map[string]any, changes map[string]any) map[string]any {
	result := maps.Clone(base)
	maps.Copy(result, changes)
	return result
}

var readiness = map[string]any{
	"schema_id": "cartulary.harness_test_failure.v4", "failure_class": "infra", "failure_reason": "service_readiness_timeout",
	"setup_source": "s3test", "service": "object_store", "readiness_stage": "put", "attempt_count": 25,
	"cleanup_outcome": "completed", "phase": "setup", "prior_failure": false,
}

var cleanup = with(readiness, map[string]any{
	"failure_class": "harness", "failure_reason": "cleanup_error", "setup_source": "pgtest", "service": "postgres",
	"readiness_stage": "delete", "attempt_count": 1, "cleanup_outcome": "failed", "phase": "cleanup",
	"cleanup_diagnostics": []any{map[string]any{"stage": "forced_drop", "cause": "deadline_exceeded"}},
})

var waitSnapshot = map[string]any{
	"type": "IPC", "event": "CheckpointDone", "blockers": 0, "checkpointer_type": "IO", "checkpointer_event": "DataFileSync",
	"checkpointer_blockers": 0, "requested": 7, "done": 6, "buffers": 200, "write_ms": 1234, "sync_ms": 56,
	"barrier": map[string]any{"operation": "select", "database_scope": "other", "operation_age_ms": 123, "status": "sampled",
		"backend": "client backend", "state": "idle", "type": "Client", "event": "ClientRead", "blockers": 0},
}

var dropAttempt = map[string]any{"duration_ms": 5000, "admission_ms": 1, "service_generation": nil, "admission_budget_ms": 15000,
	"operation_ms": 4999, "operation_budget_ms": 5000, "admission": "acquired", "outcome": "deadline_exceeded"}

func observedCleanup(wait map[string]any, attempt map[string]any) map[string]any {
	postgres := with(with(dropAttempt, attempt), map[string]any{"wait": wait})
	return with(cleanup, map[string]any{"cleanup_diagnostics": []any{map[string]any{"stage": "forced_drop", "cause": "deadline_exceeded", "postgres": postgres}}})
}

var sampledWait = map[string]any{"status": "sampled", "first": waitSnapshot, "last": waitSnapshot}

type placed struct {
	test     string
	envelope map[string]any
}

// withCleanup emits envelopes then fails every selector.
func withCleanup(t *testing.T, envelopes []placed, rows ...Row) Outcome {
	t.Helper()
	if len(rows) == 0 {
		rows = rowA
	}
	var events []any
	for _, p := range envelopes {
		test := p.test
		if test == "" {
			test = "TestA"
		}
		events = append(events, markerOutput(test, p.envelope))
	}
	for _, row := range rows {
		for _, selector := range row.Selectors {
			events = append(events, event{"Action": "fail", "Test": selector})
		}
	}
	return only(t, rows, events...)
}

// asJSON normalizes a diagnostic for comparison with authored maps.
func asJSON(t *testing.T, value any) any {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func TestGoRowOutcomes(t *testing.T) {
	t.Run("terminal events settle rows", func(t *testing.T) {
		if o := only(t, rowA, event{"Action": "pass", "Test": "TestA", "Elapsed": 0.01}); o.TerminalState != "passed" || o.ExitCode != 0 || o.FailureClass != "" || o.DurationMS != 10 {
			t.Fatalf("passed %+v", o)
		}
		if o := only(t, rowA); o.TerminalState != "infrastructure_failed" || o.ExitCode != 3 || o.FailureReason != "missing_selector_result" {
			t.Fatalf("missing %+v", o)
		}
		if o := only(t, rowA, event{"Action": "skip", "Test": "TestA"}); o.TerminalState != "infrastructure_failed" || o.FailureReason != "unauthorized_skip" {
			t.Fatalf("skipped %+v", o)
		}
		if o := only(t, rowA, event{"Action": "fail", "Test": "TestA", "Elapsed": 0.01}); o.TerminalState != "failed" || o.ExitCode != 10 || o.FailureClass != "product" {
			t.Fatalf("failed %+v", o)
		}
	})
	t.Run("a typed setup failure is infrastructure with its diagnostic", func(t *testing.T) {
		o := only(t, rowA, markerOutput("TestA/readiness", readiness), event{"Action": "fail", "Test": "TestA", "Elapsed": 120})
		if o.TerminalState != "infrastructure_failed" || o.FailureClass != "infra" || o.FailureReason != "service_readiness_timeout" || o.ExitCode != 3 {
			t.Fatalf("readiness %+v", o)
		}
		want := map[string]any{"schema_id": DiagnosticSchema, "setup": asJSON(t, readiness), "cleanup": []any{}}
		if got := asJSON(t, o.Diagnostic); !reflect.DeepEqual(got, asJSON(t, want)) {
			t.Fatalf("diagnostic %v", got)
		}
		duplicate := only(t, rowA, markerOutput("TestA", readiness), markerOutput("TestA/readiness", readiness), event{"Action": "fail", "Test": "TestA"})
		if duplicate.FailureReason != "service_readiness_timeout" || !reflect.DeepEqual(duplicate.Diagnostic, o.Diagnostic) {
			t.Fatalf("duplicate %+v", duplicate)
		}
		conflicting := only(t, rowA, markerOutput("TestA", readiness), markerOutput("TestA/readiness", with(readiness, map[string]any{"cleanup_outcome": "failed"})), event{"Action": "fail", "Test": "TestA"})
		if conflicting.FailureReason != "scheduler_accounting_error" || conflicting.ExitCode != 11 || conflicting.Diagnostic != nil {
			t.Fatalf("conflicting %+v", conflicting)
		}
		cancelled := only(t, rowA, markerOutput("TestA/readiness", with(readiness, map[string]any{"failure_class": "interrupted", "failure_reason": "cancelled_or_interrupted"})), event{"Action": "fail", "Test": "TestA"})
		if cancelled.TerminalState != "cancelled" || cancelled.FailureClass != "interrupted" || cancelled.ExitCode != 130 {
			t.Fatalf("cancelled %+v", cancelled)
		}
		unrelated := only(t, rowA, markerOutput("TestOther/readiness", readiness), event{"Action": "fail", "Test": "TestA"})
		if unrelated.FailureClass != "product" || unrelated.ExitCode != 10 || unrelated.Diagnostic != nil {
			t.Fatalf("unrelated %+v", unrelated)
		}
		malformed := only(t, rowA, event{"Action": "output", "Test": "TestA", "Output": "CARTULARY_HARNESS_TEST_FAILURE={not-json}\n"}, event{"Action": "fail", "Test": "TestA"})
		if malformed.FailureReason != "scheduler_accounting_error" || malformed.ExitCode != 11 {
			t.Fatalf("malformed %+v", malformed)
		}
	})
	t.Run("cleanup failures keep product and setup precedence", func(t *testing.T) {
		if o := withCleanup(t, []placed{{envelope: cleanup}}); o.FailureReason != "cleanup_error" || o.ExitCode != 12 || !reflect.DeepEqual(asJSON(t, o.Diagnostic.Cleanup), []any{asJSON(t, cleanup)}) {
			t.Fatalf("cleanup only %+v", o)
		}
		observed := observedCleanup(sampledWait, nil)
		if o := withCleanup(t, []placed{{envelope: observed}}); o.ExitCode != 12 || !reflect.DeepEqual(asJSON(t, o.Diagnostic.Cleanup), []any{asJSON(t, observed)}) {
			t.Fatalf("observed %+v", o)
		}
		after := withCleanup(t, []placed{{envelope: with(cleanup, map[string]any{"prior_failure": true})}})
		if after.FailureReason != "test_assertion_failure" || after.ExitCode != 10 || after.Diagnostic.Cleanup[0].FailureReason != "cleanup_error" {
			t.Fatalf("after assertion %+v", after)
		}
		afterSetup := withCleanup(t, []placed{{envelope: readiness}, {envelope: with(cleanup, map[string]any{"prior_failure": true})}})
		if afterSetup.FailureReason != "service_readiness_timeout" || afterSetup.ExitCode != 3 || len(afterSetup.Diagnostic.Cleanup) != 1 || afterSetup.Diagnostic.Setup == nil {
			t.Fatalf("after setup %+v", afterSetup)
		}
		if o := withCleanup(t, []placed{{envelope: cleanup}, {envelope: cleanup}, {envelope: with(cleanup, map[string]any{"prior_failure": true})}}); o.FailureReason != "cleanup_error" || o.ExitCode != 12 {
			t.Fatalf("a later cleanup manufactured a product failure: %+v", o)
		}
		if o := withCleanup(t, []placed{{test: "TestA/child", envelope: cleanup}, {envelope: with(cleanup, map[string]any{"prior_failure": true})}}); o.ExitCode != 12 {
			t.Fatalf("after child cleanup %+v", o)
		}
		two := Row{ID: "row-a", Selectors: []string{"TestA", "TestB"}}
		if o := withCleanup(t, []placed{{envelope: cleanup}}, two); o.ExitCode != 10 {
			t.Fatalf("a separate assertion was lost: %+v", o)
		}
		if o := withCleanup(t, []placed{{envelope: readiness}, {envelope: with(cleanup, map[string]any{"prior_failure": true})}}, two); o.ExitCode != 10 || o.Diagnostic.Setup.FailureReason != "service_readiness_timeout" {
			t.Fatalf("product precedence was lost: %+v", o)
		}
		sibling := only(t, rowA, markerOutput("TestA/setup", readiness), event{"Action": "fail", "Test": "TestA/setup"}, event{"Action": "fail", "Test": "TestA/assertion"}, event{"Action": "fail", "Test": "TestA"})
		if sibling.ExitCode != 10 {
			t.Fatalf("a failing sibling lost product precedence: %+v", sibling)
		}
		if o := withCleanup(t, []placed{{test: "TestOther", envelope: cleanup}}); o.ExitCode != 10 || o.Diagnostic != nil {
			t.Fatalf("unrelated cleanup %+v", o)
		}
		if o := withCleanup(t, []placed{{envelope: with(cleanup, map[string]any{"failure_class": "product"})}}); o.FailureReason != "scheduler_accounting_error" {
			t.Fatalf("malformed cleanup %+v", o)
		}
		unsafe := with(cleanup, map[string]any{"cleanup_diagnostics": []any{map[string]any{"stage": "forced_drop", "cause": "private-transport-text"}}})
		if o := withCleanup(t, []placed{{envelope: unsafe}}); o.FailureReason != "scheduler_accounting_error" || o.Diagnostic != nil {
			t.Fatalf("unsafe cleanup %+v", o)
		}
		misplaced := observedCleanup(sampledWait, nil)
		misplaced["cleanup_diagnostics"].([]any)[0].(map[string]any)["stage"] = "catalog_unlock"
		if o := withCleanup(t, []placed{{envelope: misplaced}}); o.FailureReason != "cleanup_error" {
			t.Fatalf("attempt evidence lost with an unlock failure: %+v", o)
		}
	})
	t.Run("drop observations are validated before attribution", func(t *testing.T) {
		for _, c := range []struct {
			label   string
			changes map[string]any
			wait    map[string]any
			valid   bool
		}{
			{"remaining owner budget", map[string]any{"operation_budget_ms": 25000}, sampledWait, true},
			{"maximum owner budget", map[string]any{"operation_budget_ms": 35000}, sampledWait, true},
			{"negative operation budget", map[string]any{"operation_budget_ms": -1}, sampledWait, false},
			{"oversized operation budget", map[string]any{"operation_budget_ms": 35001}, sampledWait, false},
			{"oversized admission budget", map[string]any{"admission_budget_ms": 15001}, sampledWait, false},
			{"negative admission budget", map[string]any{"admission_budget_ms": -1}, sampledWait, false},
			{"unadmitted operation", map[string]any{"admission": "rejected"}, sampledWait, false},
			{"unadmitted budget", map[string]any{"admission": "rejected", "operation_ms": 0}, sampledWait, false},
			{"unadmitted sampled operation", map[string]any{"admission": "rejected", "operation_ms": 0, "operation_budget_ms": 0}, sampledWait, false},
			{"unadmitted zero values", map[string]any{"admission": "rejected", "operation_ms": 0, "operation_budget_ms": 0}, map[string]any{"status": "not_observed", "first": nil, "last": nil}, true},
		} {
			envelope := observedCleanup(c.wait, c.changes)
			o := withCleanup(t, []placed{{envelope: envelope}})
			if want := map[bool]int{true: 12, false: 11}[c.valid]; o.ExitCode != want || (o.Diagnostic != nil) != c.valid {
				t.Fatalf("%s: %+v", c.label, o)
			}
			if c.valid && !reflect.DeepEqual(asJSON(t, o.Diagnostic.Cleanup), []any{asJSON(t, envelope)}) {
				t.Fatalf("%s: diagnostic %v", c.label, asJSON(t, o.Diagnostic))
			}
		}
		barrier := func(changes map[string]any) map[string]any {
			return with(waitSnapshot, map[string]any{"barrier": with(waitSnapshot["barrier"].(map[string]any), changes)})
		}
		for i, wait := range []map[string]any{
			{"status": "sampled", "first": waitSnapshot, "last": barrier(map[string]any{"operation": "PrivateCredentialAlpha123"})},
			{"status": "sampled", "first": waitSnapshot, "last": barrier(map[string]any{"database_scope": "private_database"})},
			{"status": "sampled", "first": waitSnapshot, "last": barrier(map[string]any{"operation_age_ms": -1})},
			{"status": "sampled", "first": waitSnapshot, "last": barrier(map[string]any{"operation_age_ms": 86400001})},
			{"status": "sampled", "first": waitSnapshot, "last": barrier(map[string]any{"operation": ""})},
			{"status": "sampled", "first": waitSnapshot, "last": barrier(map[string]any{"database_scope": ""})},
			{"status": "sampled", "first": waitSnapshot, "last": barrier(map[string]any{"backend": "PrivateCredentialAlpha123"})},
			{"status": "sampled", "first": waitSnapshot, "last": barrier(map[string]any{"pid": 123})},
			{"status": "sampled", "first": waitSnapshot, "last": barrier(map[string]any{"status": "unavailable"})},
			{"status": "sampled", "first": waitSnapshot, "last": with(waitSnapshot, map[string]any{"event": "PrivateCredentialAlpha123"})},
			{"status": "sampled", "first": nil, "last": waitSnapshot},
			{"status": "not_observed", "first": waitSnapshot, "last": nil},
			{"status": "sampled", "first": waitSnapshot, "last": with(waitSnapshot, map[string]any{"buffers": -1})},
		} {
			envelope := with(cleanup, map[string]any{"cleanup_diagnostics": []any{map[string]any{"stage": "forced_drop", "cause": "deadline_exceeded", "postgres": with(dropAttempt, map[string]any{"wait": wait})}}})
			if o := withCleanup(t, []placed{{envelope: envelope}}); o.FailureReason != "scheduler_accounting_error" || o.Diagnostic != nil {
				t.Fatalf("invalid wait %d was attributed: %+v", i, o)
			}
		}
	})
}

func TestGoRowMarkerFraming(t *testing.T) {
	encoded, _ := json.Marshal(observedCleanup(sampledWait, nil))
	line := "    pgtest.go:676: CARTULARY_HARNESS_TEST_FAILURE=" + string(encoded) + "\n"
	fragments := func(parts ...string) []any {
		var events []any
		for _, part := range parts {
			events = append(events, event{"Action": "output", "Test": "TestA", "Output": part}, event{"Action": "output", "Test": "TestOther", "Output": "unrelated interleaved output\n"})
		}
		return events
	}
	check := func(label string, want int, events ...any) {
		t.Helper()
		events = append(events, event{"Action": "fail", "Test": "TestA", "Elapsed": 0.01})
		if o := only(t, rowA, events...); o.ExitCode != want {
			t.Fatalf("%s: %+v", label, o)
		}
	}
	cut := strings.Index(line, "CARTULARY_HARNESS_TEST_FAILURE") + 11
	check("fragmented cleanup keeps its class", 12, fragments(line[:1024], line[1024:])...)
	check("non-event JSON does not alter framing", 12, append([]any{"null", "[]", `"process diagnostic"`, "not json"}, fragments(line[:1024], line[1024:])...)...)
	check("another package cannot complete an envelope", 11,
		event{"Action": "output", "Package": "first", "Test": "TestA", "Output": line[:1024]},
		event{"Action": "output", "Package": "second", "Test": "TestA", "Output": line[1024:]})
	check("a split marker is reassembled within its test", 12, fragments(line[:cut], line[cut:1024], line[1024:])...)
	check("an unterminated envelope fails closed", 11, fragments(strings.TrimRight(line, "\n"))...)
	check("a truncated envelope fails closed", 11, fragments(line[:1024])...)
	check("another test cannot complete an envelope", 11,
		event{"Action": "output", "Test": "TestA", "Output": line[:1024]}, event{"Action": "output", "Test": "TestOther", "Output": line[1024:]})
	check("an oversized envelope fails closed", 11, fragments(strings.TrimRight(line, "\n"), strings.Repeat(" ", 65536), "\n")...)
	check("several diagnostic lines keep their framing", 12, event{"Action": "output", "Test": "TestA", "Output": line + line})
	var crowd []any
	for i := range 128 {
		crowd = append(crowd, event{"Action": "output", "Test": fmt.Sprintf("TestOther%d", i), "Output": "partial"})
	}
	check("the unfinished-line bound cannot discard a diagnostic", 11, append(crowd, event{"Action": "output", "Test": "TestA", "Output": line[:1024]})...)

	t.Run("an event over the line bound leaves outcomes unknown", func(t *testing.T) {
		s := NewStream(rowA)
		_, _ = s.Write([]byte(`{"Action":"output","Test":"TestA","Output":"` + strings.Repeat("x", MaxEventBytes) + `"}` + "\n"))
		_, _ = s.Write([]byte(`{"Action":"pass","Test":"TestA"}` + "\n"))
		if _, err := s.Finish("./internal/fixture"); err != ErrEventTooLong {
			t.Fatalf("error %v", err)
		}
	})
	t.Run("events split across writes are reassembled", func(t *testing.T) {
		s := NewStream(rowA)
		whole := []byte(`{"Action":"pass","Test":"TestA","Elapsed":0.5}` + "\n")
		for _, b := range whole {
			_, _ = s.Write([]byte{b})
		}
		result, err := s.Finish("./internal/fixture")
		if err != nil || result.Outcomes[0].TerminalState != "passed" || result.Outcomes[0].DurationMS != 500 {
			t.Fatalf("result %+v: %v", result, err)
		}
	})
}

func TestGoRowFailureDetails(t *testing.T) {
	var events []any
	for i := range 12 {
		name := fmt.Sprintf("TestA/case%d", i)
		events = append(events, event{"Action": "output", "Test": name, "Output": strings.Repeat("x", 5000) + fmt.Sprintf("early assertion %d\n", i)}, event{"Action": "fail", "Test": name})
	}
	events = append(events, event{"Action": "output", "Test": "TestOther", "Output": "unselected failure detail"}, event{"Action": "fail", "Test": "TestOther"})
	for range 100 {
		events = append(events, event{"Action": "output", "Test": "TestLater", "Output": strings.Repeat("passing noise", 100)})
	}
	result, err := stream(rowA, events...)
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcomes[0].TerminalState != "infrastructure_failed" {
		t.Fatalf("diagnostic detail invented parent accounting: %+v", result.Outcomes[0])
	}
	output := result.FailureDetails
	if !strings.Contains(output, "early assertion 0") || !strings.Contains(output, "selected=8 omitted=4") || strings.Contains(output, "passing noise") || strings.Contains(output, "unselected failure detail") || len(output) >= 40*1024 {
		t.Fatalf("details %q", output)
	}
	if result, _ := stream(rowA, event{"Action": "pass", "Test": "TestA"}); result.FailureDetails != "" {
		t.Fatalf("a passing stream rendered details %q", result.FailureDetails)
	}
}

// TH-HARNESS-AC-165: Go test durations come from the outcome parse, keep
// terminal-event order, redact and bound names, and never change outcomes.
func TestGoRowDurations(t *testing.T) {
	rows := []Row{{ID: "row-b", Selectors: []string{"TestB"}}, {ID: "row-a", Selectors: []string{"TestA"}}}
	long := "TestA/" + strings.Repeat("x", 600)
	pkg := "example.test/fixture"
	result, err := stream(rows,
		event{"Action": "run", "Package": pkg, "Test": "TestA"},
		event{"Action": "pass", "Package": pkg, "Test": "TestA/one", "Elapsed": 1.25},
		event{"Action": "pass", "Package": pkg, "Test": "TestA/two/nested", "Elapsed": 0.004},
		event{"Action": "fail", "Package": pkg, "Test": "TestA/two", "Elapsed": 0.5},
		event{"Action": "skip", "Package": pkg, "Test": "TestA/password=hunter2", "Elapsed": 0},
		event{"Action": "pass", "Package": pkg, "Test": long, "Elapsed": 0.1},
		event{"Action": "fail", "Package": pkg, "Test": "TestA", "Elapsed": 2},
		event{"Action": "pass", "Package": pkg, "Test": "TestB", "Elapsed": 0.25},
		event{"Action": "fail", "Package": pkg, "Elapsed": 2.6},
	)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"package": pkg, "rows": []any{map[string]any{"row_id": "row-b", "selectors": []any{"TestB"}}, map[string]any{"row_id": "row-a", "selectors": []any{"TestA"}}},
		"package_action": "fail", "package_elapsed_ms": 2600.0, "tests": []any{
			map[string]any{"name": "TestA/one", "action": "pass", "duration_ms": 1250.0},
			map[string]any{"name": "TestA/two/nested", "action": "pass", "duration_ms": 4.0},
			map[string]any{"name": "TestA/two", "action": "fail", "duration_ms": 500.0},
			map[string]any{"name": "TestA/password=hunter2", "action": "skip", "duration_ms": 0.0},
			map[string]any{"name": long, "action": "pass", "duration_ms": 100.0},
			map[string]any{"name": "TestA", "action": "fail", "duration_ms": 2000.0},
			map[string]any{"name": "TestB", "action": "pass", "duration_ms": 250.0},
		}}
	if got := asJSON(t, result.Durations); !reflect.DeepEqual(got, want) {
		t.Fatalf("durations %v", got)
	}
	redact := regexp.MustCompile(`password=[^/]*`)
	document := NewDurationsDocument("go:unit", []Durations{result.Durations}, func(s string) string { return redact.ReplaceAllString(s, "password=<redacted>") })
	var names []string
	for _, test := range document.Invocations[0].Tests {
		names = append(names, test.Name)
	}
	if document.Omitted != 1 || !reflect.DeepEqual(names, []string{"TestA/one", "TestA/two/nested", "TestA/two", "TestA/password=<redacted>", "TestA", "TestB"}) || *document.UnitID != "go:unit" {
		t.Fatalf("document omitted=%d names=%v", document.Omitted, names)
	}
	many := result.Durations
	many.Tests = nil
	for i := range 4100 {
		many.Tests = append(many.Tests, TestDuration{Name: fmt.Sprintf("TestA/case%d", i), Action: "pass", DurationMS: int64(i)})
	}
	bounded := NewDurationsDocument("", []Durations{many, result.Durations}, func(s string) string { return s })
	if len(bounded.Invocations[0].Tests) != 4096 || len(bounded.Invocations[1].Tests) != 0 || bounded.Omitted != 4+len(result.Durations.Tests) || bounded.UnitID != nil {
		t.Fatalf("bounded omitted=%d", bounded.Omitted)
	}
	if empty, _ := stream(rowA); empty.Durations.Package != "./internal/fixture" || empty.Durations.PackageAction != nil || len(empty.Durations.Tests) != 0 {
		t.Fatalf("empty durations %+v", empty.Durations)
	}
}
