package testfailure

import (
	"encoding/json"
	"fmt"
	"testing"
)

const (
	SchemaID = "cartulary.harness_test_failure.v1"
	Marker   = "CARTULARY_HARNESS_TEST_FAILURE="
)

type Envelope struct {
	SchemaID       string `json:"schema_id"`
	FailureClass   string `json:"failure_class"`
	FailureReason  string `json:"failure_reason"`
	SetupSource    string `json:"setup_source"`
	Service        string `json:"service"`
	ReadinessStage string `json:"readiness_stage"`
	AttemptCount   int    `json:"attempt_count"`
	CleanupOutcome string `json:"cleanup_outcome"`
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
	}
}

func (e Envelope) Validate() error {
	if e.SchemaID != SchemaID {
		return fmt.Errorf("invalid test failure schema")
	}
	if !member(e.FailureClass, "harness", "infra", "interrupted") {
		return fmt.Errorf("invalid test failure class")
	}
	if !member(e.FailureReason, "cancelled_or_interrupted", "fixture_error", "service_readiness_timeout") {
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

func member(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
