package testfailure

import (
	"encoding/json"
	"fmt"
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
