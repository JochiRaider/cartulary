package testfailure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
)

func TestEncodeProducesBoundedMarkerWithoutDiagnosticText(t *testing.T) {
	envelope := NewEnvelope(
		"infra",
		"service_readiness_timeout",
		"s3test",
		"object_store",
		"put",
		25,
		"completed",
	)
	encoded, err := Encode(envelope)
	if err != nil {
		t.Fatalf("encode envelope: %v", err)
	}
	if !strings.HasPrefix(encoded, Marker) {
		t.Fatalf("marker prefix missing: %q", encoded)
	}
	if strings.Contains(encoded, "transport") || strings.Contains(encoded, "credential") {
		t.Fatalf("marker leaked diagnostic text: %q", encoded)
	}
	for _, prior := range []bool{false, true} {
		probe := &cleanupProbe{failed: prior}
		CleanupFailure(probe, "pgtest", "postgres", "delete", &diagnosticError{stage: "forced_drop", cause: context.DeadlineExceeded})
		var diagnostic Envelope
		if err := json.Unmarshal([]byte(strings.TrimPrefix(probe.output, Marker)), &diagnostic); err != nil {
			t.Fatal(err)
		}
		if err := diagnostic.Validate(); err != nil {
			t.Fatal(err)
		}
		if diagnostic.PriorFailure != prior || diagnostic.Phase != "cleanup" || diagnostic.FailureReason != "cleanup_error" {
			t.Fatalf("cleanup lost independent attribution: %+v", diagnostic)
		}
		if !probe.failed {
			t.Fatal("cleanup did not fail the test")
		}
		if !reflect.DeepEqual(diagnostic.CleanupDiagnostics, []CleanupDiagnostic{{Stage: "forced_drop", Cause: "deadline_exceeded"}}) {
			t.Fatalf("cleanup lost normalized cause: %+v", diagnostic)
		}
	}
	t.Run("drop wait retains exact failure and rejects private values", func(t *testing.T) {
		sample := &postgrescleanup.DropWaitSnapshot{Barrier: &postgrescleanup.BarrierWaitSnapshot{Status: "sampled", Operation: "select", DatabaseScope: "other", OperationAgeMS: 123, Backend: "client backend", State: "idle", Type: "Client", Event: "ClientRead"}, Type: "IPC", Event: "CheckpointDone", CheckpointerType: "IO", CheckpointerEvent: "DataFileSync", Requested: 7, Done: 6, Buffers: 200}
		observation := postgrescleanup.DropAttempt{Admission: "acquired", Outcome: "deadline_exceeded", Wait: postgrescleanup.DropObservation{Status: "sampled", First: sample, Last: sample}}
		failure := &observedDiagnosticError{diagnosticError: &diagnosticError{stage: "forced_drop", cause: context.DeadlineExceeded}, observation: observation}
		probe := &cleanupProbe{}
		CleanupFailure(probe, "pgtest", "postgres", "delete", failure)
		var envelope map[string]any
		if err := json.Unmarshal([]byte(strings.TrimPrefix(probe.output, Marker)), &envelope); err != nil {
			t.Fatal(err)
		}
		diagnostic := envelope["cleanup_diagnostics"].([]any)[0].(map[string]any)
		wait, ok := diagnostic["postgres"].(map[string]any)["wait"].(map[string]any)
		if !ok || wait["status"] != "sampled" || diagnostic["stage"] != "forced_drop" || diagnostic["cause"] != "deadline_exceeded" {
			t.Fatal("lost structural wait or original failure")
		}
		if !errors.Is(failure, context.DeadlineExceeded) {
			t.Fatal("lost original cause")
		}
		original := *sample.Barrier
		for _, mutate := range []func(*postgrescleanup.BarrierWaitSnapshot){
			func(b *postgrescleanup.BarrierWaitSnapshot) { b.Backend = "PrivateCredentialAlpha123" },
			func(b *postgrescleanup.BarrierWaitSnapshot) { b.Operation = "PrivateCredentialAlpha123" },
			func(b *postgrescleanup.BarrierWaitSnapshot) { b.DatabaseScope = "PrivateCredentialAlpha123" },
			func(b *postgrescleanup.BarrierWaitSnapshot) { b.OperationAgeMS = -1 },
			func(b *postgrescleanup.BarrierWaitSnapshot) { b.OperationAgeMS = 86400001 },
		} {
			*sample.Barrier = original
			mutate(sample.Barrier)
			CleanupFailure(probe, "pgtest", "postgres", "delete", failure)
			if strings.Contains(probe.output, "PrivateCredential") || !strings.Contains(probe.output, `"status":"unavailable"`) {
				t.Fatal("invalid observation was not closed")
			}
		}
	})
	t.Run("joined causes survive without error text", func(t *testing.T) {
		failure := errors.Join(
			&diagnosticError{stage: "forced_drop", cause: &sqlStateError{state: "55006"}},
			&diagnosticError{stage: "catalog_unlock", cause: context.DeadlineExceeded},
			&diagnosticError{stage: "catalog_unlock", cause: context.DeadlineExceeded},
		)
		want := []CleanupDiagnostic{{Stage: "forced_drop", Cause: "database_in_use"}, {Stage: "catalog_unlock", Cause: "deadline_exceeded"}}
		if got := describeCleanupFailure(failure); !reflect.DeepEqual(got, want) {
			t.Fatalf("joined failure diagnostics: got %+v, want %+v", got, want)
		}
		if !errors.Is(failure, context.DeadlineExceeded) {
			t.Fatal("typed stage lost the wrapped cause")
		}
	})
	t.Run("unknown values cannot leak and cycles are bounded", func(t *testing.T) {
		failure := &diagnosticError{stage: "private-database-secret", cause: &sqlStateError{state: "private-transport-secret"}}
		want := []CleanupDiagnostic{{Stage: "unknown", Cause: "unknown"}}
		if got := describeCleanupFailure(failure); !reflect.DeepEqual(got, want) {
			t.Fatalf("unregistered data retained: %+v", got)
		}
		failure.cause = failure
		want = []CleanupDiagnostic{{Stage: "unknown", Cause: "diagnostic_limit"}}
		if got := describeCleanupFailure(failure); !reflect.DeepEqual(got, want) {
			t.Fatalf("cycle did not retain bounded diagnostic: %+v", got)
		}
	})
}

type diagnosticError struct {
	stage string
	cause error
}

func (*diagnosticError) Error() string          { panic("arbitrary error text must never be read") }
func (e *diagnosticError) CleanupStage() string { return e.stage }
func (e *diagnosticError) Unwrap() error        { return e.cause }

type observedDiagnosticError struct {
	*diagnosticError
	observation postgrescleanup.DropAttempt
}

func (e *observedDiagnosticError) PostgresCleanup() postgrescleanup.DropAttempt { return e.observation }

type sqlStateError struct{ state string }

func (*sqlStateError) Error() string      { panic("database error text must never be read") }
func (e *sqlStateError) SQLState() string { return e.state }

type cleanupProbe struct {
	failed bool
	output string
}

func (*cleanupProbe) Helper()        {}
func (p *cleanupProbe) Failed() bool { return p.failed }
func (p *cleanupProbe) Error(args ...any) {
	p.output = fmt.Sprint(args...)
	p.failed = true
}

func TestEnvelopeRejectsUnregisteredValues(t *testing.T) {
	envelope := NewEnvelope("infra", "other", "s3test", "object_store", "put", 1, "completed")
	if err := envelope.Validate(); err == nil {
		t.Fatal("expected unregistered reason to fail")
	}
	envelope = NewEnvelope("harness", "cleanup_error", "pgtest", "postgres", "delete", 1, "failed")
	if err := envelope.Validate(); err == nil {
		t.Fatal("cleanup reason accepted as setup")
	}
	envelope.Phase = "cleanup"
	envelope.FailureClass = "product"
	if err := envelope.Validate(); err == nil {
		t.Fatal("cleanup accepted product classification")
	}
	envelope.FailureClass = "harness"
	envelope.CleanupDiagnostics = []CleanupDiagnostic{{Stage: "catalog_unlock", Cause: "deadline_exceeded", Postgres: &postgrescleanup.DropAttempt{Admission: "unregistered", Wait: postgrescleanup.DropObservation{Status: "not_observed"}}}}
	if err := envelope.Validate(); err == nil {
		t.Fatal("invalid admission observation accepted")
	}
	envelope.CleanupDiagnostics = []CleanupDiagnostic{{Stage: "private-resource", Cause: "unknown"}}
	if err := envelope.Validate(); err == nil {
		t.Fatal("cleanup accepted an unregistered stage")
	}
}
