package lifecyclecontract

import (
	"encoding/json"
	"fmt"
	"testing"
)

// recorder captures the first fatal failure of a validator call.
type recorder struct {
	testing.TB
	failure string
}

func (r *recorder) Helper() {}

func (r *recorder) Fatalf(format string, args ...any) {
	r.failure = fmt.Sprintf(format, args...)
	panic(r)
}

func failureOf(check func(testing.TB)) (failure string) {
	r := &recorder{}
	defer func() {
		if recovered := recover(); recovered != nil {
			if recovered != r {
				panic(recovered)
			}
			failure = r.failure
		}
	}()
	check(r)
	return ""
}

func ptr(value string) *string { return &value }

func TestValidatorsEnforceTheRegisteredVocabulary_Unit(t *testing.T) {
	for _, tc := range []struct {
		name  string
		check func(testing.TB)
		pass  bool
	}{
		{name: "registered guard of the machine", pass: true, check: func(tb testing.TB) {
			RequireTransition(tb, "timeline.capture_state", ptr("rough"), "superseded", "guards_not_satisfied", []string{"replacement_must_differ_from_target"})
		}},
		{name: "matrix rejection", pass: true, check: func(tb testing.TB) {
			RequireTransition(tb, "indicator_observation.disposition", ptr("dismissed"), "dismissed", "transition_not_allowed", []string{})
		}},
		{name: "record type without a machine", pass: true, check: func(tb testing.TB) {
			RequireTransition(tb, "artifact", nil, "superseded", "transition_not_allowed", []string{"record_type_has_machine"})
		}},
		{name: "unregistered reason", check: func(tb testing.TB) {
			RequireTransition(tb, "task_request.status", ptr("open"), "done", "from_status_not_allowed", []string{})
		}},
		{name: "guard of another machine", check: func(tb testing.TB) {
			RequireTransition(tb, "decision.status", ptr("proposed"), "approved", "guards_not_satisfied", []string{"active_requires_owner"})
		}},
		{name: "empty from_status", check: func(tb testing.TB) {
			RequireTransition(tb, "task_request.status", ptr(""), "done", "transition_not_allowed", []string{})
		}},
		{name: "empty to_status", check: func(tb testing.TB) {
			RequireTransition(tb, "task_request.status", ptr("open"), "", "transition_not_allowed", []string{})
		}},
		{name: "guards_not_satisfied without a guard", check: func(tb testing.TB) {
			RequireTransition(tb, "task_request.status", ptr("open"), "blocked", "guards_not_satisfied", []string{})
		}},
		{name: "null guard list", check: func(tb testing.TB) {
			RequireTransition(tb, "task_request.status", ptr("open"), "done", "transition_not_allowed", nil)
		}},
		{name: "wire details", pass: true, check: func(tb testing.TB) {
			var details map[string]any
			_ = json.Unmarshal([]byte(`{"from_status":"open","to_status":"blocked","violated_guards":["blocked_requires_reason"],"reason_code":"guards_not_satisfied"}`), &details)
			RequireDetails(tb, "task_request.status", details)
		}},
		{name: "wire details omit a member", check: func(tb testing.TB) {
			var details map[string]any
			_ = json.Unmarshal([]byte(`{"from_status":"open","to_status":"blocked","reason_code":"transition_not_allowed"}`), &details)
			RequireDetails(tb, "task_request.status", details)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			failure := failureOf(tc.check)
			if tc.pass && failure != "" {
				t.Fatalf("valid transition rejected: %s", failure)
			}
			if !tc.pass && failure == "" {
				t.Fatal("invalid transition accepted")
			}
		})
	}
}
