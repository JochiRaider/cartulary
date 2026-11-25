package harnessrows

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/testfailure"
)

// DiagnosticSchema identifies a row's typed Go failure diagnostic.
const DiagnosticSchema = "cartulary.harness_go_failure_diagnostic.v3"

// Outcome is one row's result. A passed row has no failure fields.
type Outcome struct {
	RowID         string
	TerminalState string
	DurationMS    int64
	ExitCode      int
	FailureClass  string
	FailureReason string
	Diagnostic    *Diagnostic
}

// Diagnostic is the typed setup and cleanup evidence behind a row failure.
type Diagnostic struct {
	SchemaID string                 `json:"schema_id"`
	Setup    *testfailure.Envelope  `json:"setup"`
	Cleanup  []testfailure.Envelope `json:"cleanup"`
}

func within(test string, selectors []string) bool {
	return slices.ContainsFunc(selectors, func(selector string) bool { return test == selector || strings.HasPrefix(test, selector+"/") })
}

func (s *Stream) outcomes() ([]Outcome, error) {
	outcomes := make([]Outcome, 0, len(s.rows))
	for _, row := range s.rows {
		missing, skipped, failed := false, false, false
		duration := int64(0)
		for _, selector := range row.Selectors {
			observation, ok := s.terminal[selector]
			missing = missing || !ok
			skipped = skipped || observation.Action == "skip"
			failed = failed || observation.Action == "fail"
			duration += observation.DurationMS
		}
		outcome := Outcome{RowID: row.ID, TerminalState: "passed", DurationMS: duration}
		switch {
		case missing:
			outcome.TerminalState, outcome.ExitCode, outcome.FailureClass, outcome.FailureReason = "infrastructure_failed", 3, "infra", "missing_selector_result"
		case skipped:
			outcome.TerminalState, outcome.ExitCode, outcome.FailureClass, outcome.FailureReason = "infrastructure_failed", 3, "infra", "unauthorized_skip"
		case failed:
			outcome.TerminalState, outcome.ExitCode, outcome.FailureClass, outcome.FailureReason = "failed", 10, "product", "test_assertion_failure"
		}
		if typed, ok := s.markerFailure(row.Selectors); ok {
			typed.RowID, typed.DurationMS = row.ID, duration
			outcome = typed
		}
		outcomes = append(outcomes, outcome)
	}
	return outcomes, nil
}

// markerFailure attributes a row's typed failure markers. Any malformed or
// conflicting marker is an accounting failure; a product assertion keeps its
// precedence over setup and cleanup failures.
func (s *Stream) markerFailure(selectors []string) (Outcome, bool) {
	var relevant []marker
	for _, m := range s.markers {
		if within(m.test, selectors) {
			relevant = append(relevant, m)
		}
	}
	if len(relevant) == 0 {
		return Outcome{}, false
	}
	accounting := Outcome{TerminalState: "failed", ExitCode: 11, FailureClass: "harness", FailureReason: "scheduler_accounting_error"}
	if slices.ContainsFunc(relevant, func(m marker) bool { return m.malformed }) {
		return accounting, true
	}
	key := func(e testfailure.Envelope) string {
		encoded, _ := json.Marshal(e)
		return string(encoded)
	}
	var setup []testfailure.Envelope
	var cleanup []testfailure.Envelope
	seen := map[string]bool{}
	var cleanupEntries []marker
	var known []string
	for _, m := range relevant {
		k := key(m.envelope)
		switch m.envelope.Phase {
		case "setup":
			known = append(known, m.test)
			if !seen[k] {
				setup = append(setup, m.envelope)
			}
		case "cleanup":
			cleanupEntries = append(cleanupEntries, m)
			if !seen[k] {
				cleanup = append(cleanup, m.envelope)
			}
		}
		seen[k] = true
	}
	if len(setup) > 1 || len(cleanup) > 32 {
		return accounting, true
	}
	// A later cleanup sees the failure raised by an earlier cleanup. Its prior
	// flag must not manufacture a product assertion. Descendant failures are
	// likewise already attributed before a parent's cleanup begins.
	accounted := func(test string) bool {
		return slices.ContainsFunc(known, func(k string) bool { return k == test || strings.HasPrefix(k, test+"/") })
	}
	prior := false
	for _, m := range cleanupEntries {
		if m.envelope.PriorFailure && !accounted(m.test) {
			prior = true
		}
		known = append(known, m.test)
	}
	diagnostic := &Diagnostic{SchemaID: DiagnosticSchema, Cleanup: cleanup}
	if diagnostic.Cleanup == nil {
		diagnostic.Cleanup = []testfailure.Envelope{}
	}
	if len(setup) == 1 {
		diagnostic.Setup = &setup[0]
	}
	if !slices.ContainsFunc(selectors, func(selector string) bool { return s.terminal[selector].Action == "fail" }) {
		return accounting, true
	}
	// Parent fail events can reflect their typed descendants. A failing sibling
	// or independent selector without a typed cause is still a product
	// assertion and keeps the ordinary class precedence.
	for _, test := range s.order {
		if s.terminal[test].Action == "fail" && within(test, selectors) && !slices.ContainsFunc(known, func(k string) bool { return k == test || strings.HasPrefix(k, test+"/") }) {
			prior = true
		}
	}
	switch {
	case prior:
		return Outcome{TerminalState: "failed", ExitCode: 10, FailureClass: "product", FailureReason: "test_assertion_failure", Diagnostic: diagnostic}, true
	case diagnostic.Setup != nil:
		if diagnostic.Setup.FailureClass == "interrupted" {
			return Outcome{TerminalState: "cancelled", ExitCode: 130, FailureClass: "interrupted", FailureReason: diagnostic.Setup.FailureReason, Diagnostic: diagnostic}, true
		}
		return Outcome{TerminalState: "infrastructure_failed", ExitCode: 3, FailureClass: diagnostic.Setup.FailureClass, FailureReason: diagnostic.Setup.FailureReason, Diagnostic: diagnostic}, true
	default:
		return Outcome{TerminalState: "failed", ExitCode: 12, FailureClass: "harness", FailureReason: "cleanup_error", Diagnostic: diagnostic}, true
	}
}
