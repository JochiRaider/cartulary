package suiteservices

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	LifecycleModeEnv = "CARTULARY_TEST_SERVICES_LIFECYCLE_MODE"

	LifecycleEventStartServices     = "start_services"
	LifecycleEventReadinessPassed   = "readiness_passed"
	LifecycleEventStartupFailed     = "startup_failed"
	LifecycleEventChildStarted      = "child_started"
	LifecycleEventChildFinished     = "child_finished"
	LifecycleEventInterruptReceived = "interrupt_received"
	LifecycleEventCleanupStarted    = "cleanup_started"
	LifecycleEventCleanupSucceeded  = "cleanup_succeeded"
	LifecycleEventCleanupFailed     = "cleanup_failed"
)

type LifecycleRecord struct {
	SchemaID         string              `json:"schema_id"`
	Target           string              `json:"target"`
	RunID            string              `json:"run_id"`
	RunRoot          string              `json:"run_root"`
	SuiteID          string              `json:"suite_id"`
	Mode             string              `json:"mode"`
	MachineID        string              `json:"machine_id"`
	Seq              int                 `json:"seq"`
	Event            string              `json:"event"`
	ChildKey         string              `json:"child_key,omitempty"`
	ActiveChildCount int                 `json:"active_child_count,omitempty"`
	FromState        string              `json:"from_state"`
	ToState          string              `json:"to_state"`
	TransitionStatus string              `json:"transition_status"`
	FailureClass     *string             `json:"failure_class"`
	FailureReason    *string             `json:"failure_reason"`
	EmittedAt        string              `json:"emitted_at"`
	MonotonicMS      int64               `json:"monotonic_ms"`
	ArtifactRefs     []map[string]string `json:"artifact_refs"`
	Extensions       map[string]any      `json:"extensions,omitempty"`
}

type lifecycleState struct {
	state    string
	children map[string]bool
	seq      int
}

func LifecycleEventsPath(env map[string]string) (string, bool, error) {
	suiteDir, ok, err := ResolveSuiteArtifactDir(env)
	if err != nil || !ok {
		return "", ok, err
	}
	return filepath.Join(suiteDir, "lifecycle-events.jsonl"), true, nil
}

func RecordLifecycleEvent(env map[string]string, event string, childKey string) error {
	return recordLifecycleEvent(env, event, childKey, "", "")
}

func RecordLifecycleFailureEvent(env map[string]string, event string, childKey string, failureClass string, failureReason string) error {
	return recordLifecycleEvent(env, event, childKey, failureClass, failureReason)
}

func recordLifecycleEvent(env map[string]string, event string, childKey string, failureClassValue string, failureReasonValue string) error {
	path, ok, err := LifecycleEventsPath(env)
	if err != nil || !ok {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create lifecycle artifact dir: %w", err)
	}
	state, err := readLifecycleState(path)
	if err != nil {
		return err
	}
	fromState := state.state
	toState, activeChildren, legal := applyLifecycleTransition(&state, event, childKey)
	status := "applied"
	var failureClass *string
	var failureReason *string
	if !legal {
		status = "illegal"
		toState = fromState
		class := FailureClassHelper
		reason := "scheduler_accounting_error"
		failureClass = &class
		failureReason = &reason
	} else if strings.TrimSpace(failureClassValue) != "" || strings.TrimSpace(failureReasonValue) != "" {
		class := strings.TrimSpace(failureClassValue)
		reason := strings.TrimSpace(failureReasonValue)
		failureClass = &class
		failureReason = &reason
	} else if class, reason, ok := defaultLifecycleFailure(event); ok {
		failureClass = &class
		failureReason = &reason
	}
	record := LifecycleRecord{
		SchemaID:         "cartulary.test_services.lifecycle.v1",
		Target:           firstNonEmpty(strings.TrimSpace(LookupEnvValue(env, TargetEnv)), "test-services"),
		RunID:            ResolveRunID(env),
		RunRoot:          lifecycleRunRoot(env),
		SuiteID:          SuiteID(env),
		Mode:             lifecycleMode(env),
		MachineID:        "test_services_suite_lifecycle_v1",
		Seq:              state.seq + 1,
		Event:            event,
		ChildKey:         strings.TrimSpace(childKey),
		ActiveChildCount: activeChildren,
		FromState:        fromState,
		ToState:          toState,
		TransitionStatus: status,
		FailureClass:     failureClass,
		FailureReason:    failureReason,
		EmittedAt:        time.Now().UTC().Format(time.RFC3339Nano),
		MonotonicMS:      time.Now().UnixNano() / int64(time.Millisecond),
		ArtifactRefs:     []map[string]string{},
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode lifecycle event: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open lifecycle events: %w", err)
	}
	if _, err := file.Write(append(payload, '\n')); err != nil {
		_ = file.Close()
		return fmt.Errorf("write lifecycle event: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close lifecycle events: %w", err)
	}
	if !legal {
		return fmt.Errorf("illegal lifecycle transition %s from %s", event, fromState)
	}
	return nil
}

func defaultLifecycleFailure(event string) (string, string, bool) {
	switch event {
	case LifecycleEventStartupFailed:
		return "unknown", "unknown_failure", true
	case LifecycleEventInterruptReceived:
		return "interrupted", "cancelled_or_interrupted", true
	case LifecycleEventCleanupFailed:
		return FailureClassHelper, "cleanup_error", true
	default:
		return "", "", false
	}
}

func readLifecycleState(path string) (lifecycleState, error) {
	state := lifecycleState{state: "requested", children: map[string]bool{}}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return state, nil
	}
	if err != nil {
		return state, fmt.Errorf("read lifecycle events: %w", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var record LifecycleRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return state, fmt.Errorf("decode lifecycle event: %w", err)
		}
		if record.Seq > state.seq {
			state.seq = record.Seq
		}
		if record.TransitionStatus != "applied" {
			continue
		}
		state.state = record.ToState
		switch record.Event {
		case LifecycleEventChildStarted:
			state.children[record.ChildKey] = true
		case LifecycleEventChildFinished:
			delete(state.children, record.ChildKey)
		}
	}
	return state, nil
}

func applyLifecycleTransition(state *lifecycleState, event string, childKey string) (string, int, bool) {
	childKey = strings.TrimSpace(childKey)
	legal := false
	toState := state.state
	switch event {
	case LifecycleEventStartServices:
		legal = state.state == "requested"
		toState = "starting"
	case LifecycleEventReadinessPassed:
		legal = state.state == "starting"
		toState = "ready"
	case LifecycleEventStartupFailed:
		legal = state.state == "requested" || state.state == "starting"
		toState = "failed_start"
	case LifecycleEventChildStarted:
		legal = childKey != "" && (state.state == "ready" || state.state == "running_child") && !state.children[childKey]
		toState = "running_child"
		if legal {
			state.children[childKey] = true
		}
	case LifecycleEventChildFinished:
		legal = childKey != "" && state.children[childKey]
		if legal {
			delete(state.children, childKey)
			if len(state.children) == 0 {
				toState = "ready"
			} else {
				toState = "running_child"
			}
		}
	case LifecycleEventInterruptReceived:
		legal = state.state == "starting" || state.state == "ready" || state.state == "running_child"
		toState = "interrupted"
	case LifecycleEventCleanupStarted:
		legal = state.state == "starting" || state.state == "ready" || state.state == "running_child" || state.state == "interrupted"
		toState = "cleaning"
	case LifecycleEventCleanupSucceeded:
		legal = state.state == "cleaning"
		toState = "cleaned"
	case LifecycleEventCleanupFailed:
		legal = state.state == "cleaning"
		toState = "cleanup_failed"
	}
	if legal {
		state.state = toState
	}
	return toState, len(state.children), legal
}

func CurrentLifecycleState(env map[string]string) (string, bool, error) {
	path, ok, err := LifecycleEventsPath(env)
	if err != nil || !ok {
		return "", ok, err
	}
	state, err := readLifecycleState(path)
	if err != nil {
		return "", true, err
	}
	return state.state, true, nil
}

func lifecycleRunRoot(env map[string]string) string {
	root, err := ResolveResultsRoot(env)
	if err != nil {
		return ""
	}
	return filepath.Join(root, ResolveRunID(env))
}

func lifecycleMode(env map[string]string) string {
	mode := strings.TrimSpace(LookupEnvValue(env, LifecycleModeEnv))
	if mode == "owned" || mode == "attach" {
		return mode
	}
	if SuiteActive(env) {
		return "attach"
	}
	return "owned"
}

func ReadLifecycleEvents(env map[string]string) ([]LifecycleRecord, error) {
	path, ok, err := LifecycleEventsPath(env)
	if err != nil || !ok {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var records []LifecycleRecord
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var record LifecycleRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Seq < records[j].Seq })
	return records, nil
}
