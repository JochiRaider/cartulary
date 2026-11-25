package testfailure

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

const (
	SchemaID = "cartulary.harness_test_failure.v4"
	Marker   = "CARTULARY_HARNESS_TEST_FAILURE="
)

type Envelope struct {
	SchemaID           string              `json:"schema_id"`
	FailureClass       string              `json:"failure_class"`
	FailureReason      string              `json:"failure_reason"`
	SetupSource        string              `json:"setup_source"`
	Service            string              `json:"service"`
	ReadinessStage     string              `json:"readiness_stage"`
	AttemptCount       int                 `json:"attempt_count"`
	CleanupOutcome     string              `json:"cleanup_outcome"`
	Phase              string              `json:"phase"`
	PriorFailure       bool                `json:"prior_failure"`
	CleanupDiagnostics []CleanupDiagnostic `json:"cleanup_diagnostics,omitempty"`
}

func NewEnvelope(
	failureClass string,
	failureReason string,
	setupSource string,
	service string,
	readinessStage string,
	attemptCount int,
	cleanupOutcome string,
) Envelope {
	return Envelope{
		SchemaID:       SchemaID,
		FailureClass:   failureClass,
		FailureReason:  failureReason,
		SetupSource:    setupSource,
		Service:        service,
		ReadinessStage: readinessStage,
		AttemptCount:   attemptCount,
		CleanupOutcome: cleanupOutcome,
		Phase:          "setup",
	}
}

func (e Envelope) Validate() error {
	if e.SchemaID != SchemaID {
		return fmt.Errorf("invalid test failure schema")
	}
	if !member(e.FailureClass, "harness", "infra", "interrupted") {
		return fmt.Errorf("invalid test failure class")
	}
	if !member(e.FailureReason, "cancelled_or_interrupted", "cleanup_error", "fixture_error", "service_readiness_timeout") {
		return fmt.Errorf("invalid test failure reason")
	}
	if !member(e.SetupSource, "appsupport", "pgtest", "s3test") {
		return fmt.Errorf("invalid test failure setup source")
	}
	if !member(e.Service, "object_store", "postgres") {
		return fmt.Errorf("invalid test failure service")
	}
	if !member(e.ReadinessStage, "attach", "capability", "create_bucket", "delete", "dependency_guard", "head", "not_found", "put", "start") {
		return fmt.Errorf("invalid test failure readiness stage")
	}
	if e.AttemptCount < 0 || e.AttemptCount > 1000 {
		return fmt.Errorf("invalid test failure attempt count")
	}
	if !member(e.CleanupOutcome, "completed", "failed", "not_required") {
		return fmt.Errorf("invalid test failure cleanup outcome")
	}
	if !member(e.Phase, "setup", "cleanup") {
		return fmt.Errorf("invalid test failure phase")
	}
	if e.Phase == "cleanup" {
		if e.FailureClass != "harness" || e.FailureReason != "cleanup_error" || e.CleanupOutcome != "failed" {
			return fmt.Errorf("invalid cleanup failure classification")
		}
		if len(e.CleanupDiagnostics) > 8 {
			return fmt.Errorf("too many cleanup diagnostics")
		}
		for _, diagnostic := range e.CleanupDiagnostics {
			if !validCleanupStage(diagnostic.Stage) || !validCleanupCause(diagnostic.Cause) || (diagnostic.Postgres != nil && diagnostic.Postgres.Validate() != nil) {
				return fmt.Errorf("invalid cleanup diagnostic")
			}
		}
	} else if e.PriorFailure || e.FailureReason == "cleanup_error" || len(e.CleanupDiagnostics) != 0 {
		return fmt.Errorf("invalid setup failure classification")
	}
	return nil
}

func Encode(e Envelope) (string, error) {
	if err := e.Validate(); err != nil {
		return "", err
	}
	payload, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("encode test failure: %w", err)
	}
	return Marker + string(payload), nil
}

// requiredFields are the members every envelope states, so a decoded zero
// value is never mistaken for a reported one.
var requiredFields = []string{"schema_id", "failure_class", "failure_reason", "setup_source", "service", "readiness_stage", "attempt_count", "cleanup_outcome", "phase", "prior_failure"}

// Decode reads one encoded envelope (the text after Marker) as strictly as
// Encode writes it: no unknown or missing members, no trailing data, at most
// eight distinct cleanup diagnostics, and every value Validate accepts. The
// Go row runner uses it to tell a typed failure from a malformed one.
func Decode(data []byte) (Envelope, error) {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(data, &members); err != nil {
		return Envelope{}, errors.New("malformed test failure envelope")
	}
	for _, name := range requiredFields {
		if _, ok := members[name]; !ok {
			return Envelope{}, errors.New("incomplete test failure envelope")
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var e Envelope
	if err := decoder.Decode(&e); err != nil {
		return Envelope{}, errors.New("malformed test failure envelope")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return Envelope{}, errors.New("trailing test failure data")
	}
	if _, ok := members["cleanup_diagnostics"]; ok {
		if len(e.CleanupDiagnostics) == 0 {
			return Envelope{}, errors.New("empty cleanup diagnostics")
		}
		for i := range e.CleanupDiagnostics {
			for j := range i {
				if reflect.DeepEqual(e.CleanupDiagnostics[i], e.CleanupDiagnostics[j]) {
					return Envelope{}, errors.New("repeated cleanup diagnostic")
				}
			}
		}
	}
	return e, e.Validate()
}

func Fail(t testing.TB, e Envelope) {
	t.Helper()
	encoded, err := Encode(e)
	if err != nil {
		t.Fatalf("invalid harness test failure envelope: %v", err)
	}
	t.Fatal(encoded)
}

// CleanupFailure records normalized cleanup evidence independently of an
// existing assertion or setup failure. Error lets the remaining cleanups run.
// No arbitrary error text or resource identity enters the typed envelope.
func CleanupFailure(t interface {
	Helper()
	Failed() bool
	Error(args ...any)
}, source, service, stage string, cause error) {
	t.Helper()
	e := NewEnvelope("harness", "cleanup_error", source, service, stage, 1, "failed")
	e.Phase = "cleanup"
	e.PriorFailure = t.Failed()
	e.CleanupDiagnostics = describeCleanupFailure(cause)
	encoded, err := Encode(e)
	if err != nil {
		t.Error("invalid harness cleanup failure envelope:", err)
		return
	}
	t.Error(encoded)
}

func member(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
