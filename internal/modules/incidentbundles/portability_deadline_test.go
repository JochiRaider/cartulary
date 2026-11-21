package incidentbundles

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// blockingPortabilityParticipant completes its real work, then holds its
// normal result until the invocation context ends or the test releases it.
// The fallback keeps an unbounded implementation from hanging the test.
type blockingPortabilityParticipant struct {
	*portabilityParticipantFake
	ignoreCancellation bool
	release            chan struct{}
	returned           chan struct{}
}

func (p *blockingPortabilityParticipant) hold(ctx context.Context) {
	defer close(p.returned)
	if p.ignoreCancellation {
		select {
		case <-p.release:
		case <-time.After(5 * time.Second):
		}
		return
	}
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
	}
}

func (p *blockingPortabilityParticipant) Export(ctx context.Context, input ExportInvocation) (ExportResult, error) {
	result, err := p.portabilityParticipantFake.Export(ctx, input)
	p.hold(ctx)
	return result, err
}

func (p *blockingPortabilityParticipant) PrepareImport(ctx context.Context, invocation ImportInvocation, scope *PortabilityStagedOutputScope) (ImportPreparation, error) {
	result, err := p.portabilityParticipantFake.PrepareImport(ctx, invocation, scope)
	p.hold(ctx)
	return result, err
}

func deadlineTestPolicy() ExtensionPolicy {
	return ExtensionPolicy{ProfileID: "profile", ClaimState: ClaimStateClaimed, ContractMajor: 1,
		Mode: PortabilityModeParticipant, ParticipantID: "profile.portability_v2",
		ParticipantSHA256: strings.Repeat("a", 64), ParticipantSchemaID: "test.specialization.v2",
		Export:                 PortabilityOperationPolicy{OutputSchemaID: "profile.state.v1", StateFamilyIDs: []string{"profile.state"}, MaximumInputBytes: PortabilityParticipantByteLimit, MaximumOutputBytes: PortabilityParticipantByteLimit, MaximumItems: 10000},
		Import:                 PortabilityOperationPolicy{OutputSchemaID: "profile.import.v1", StateFamilyIDs: []string{"profile.state"}, MaximumInputBytes: PortabilityParticipantByteLimit, MaximumOutputBytes: PortabilityParticipantByteLimit, MaximumItems: 10000},
		AuthoritativeFamilyIDs: []string{"profile.state"}}
}

// checkPortabilityParticipantDeadlines covers Extensions Table 9-C/9-E: each
// participant invocation runs under portability_participant_seconds, a late
// normal result is discarded, nothing is published, and a callback still
// active after the cancellation grace selects the fatal outcome.
func checkPortabilityParticipantDeadlines(t *testing.T) {
	policy := deadlineTestPolicy()
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	input := PortabilityImportContext{OperationID: "operation", IncidentID: incidentID, ActorUserID: uuid.MustParse("33333333-3333-4333-8333-333333333333")}
	budget := 50 * time.Millisecond
	newBlocking := func(ignore bool) *blockingPortabilityParticipant {
		return &blockingPortabilityParticipant{
			portabilityParticipantFake: &portabilityParticipantFake{id: policy.ParticipantID, allocate: true},
			ignoreCancellation:         ignore, release: make(chan struct{}), returned: make(chan struct{}),
		}
	}
	orchestrate := func(t *testing.T, participant ExtensionParticipant, allocator *portabilityAllocatorFake, grace time.Duration, fatal func(error)) *PortabilityOrchestrator {
		t.Helper()
		orchestrator, err := NewPortabilityOrchestrator(PortabilityOptions{
			Policies: []ExtensionPolicy{policy}, Presence: portabilityPresenceFake{present: true},
			Participants: []ExtensionParticipant{participant}, Allocator: allocator, SnapshotReads: portabilityReadFactoryFake{},
			ParticipantTimeout: budget, CancellationGrace: grace, FatalSink: fatal,
		})
		if err != nil {
			t.Fatal(err)
		}
		return orchestrator
	}
	requireTimeout := func(t *testing.T, err error, operation string) {
		t.Helper()
		var timeout *PortabilityTimeoutError
		if !errors.Is(err, ErrPortabilityTimeout) || !errors.As(err, &timeout) || timeout.ProfileID != "profile" || timeout.Operation != operation {
			t.Fatalf("%s participant outcome = %v, want extension_participant_timeout", operation, err)
		}
		if fatalImportFailure(err) || errors.Is(err, ErrPortabilityResult) {
			t.Fatalf("%s timeout was classified as a fatal or result failure: %v", operation, err)
		}
	}

	t.Run("participants require the Table 9-C deadline", func(t *testing.T) {
		_, err := NewPortabilityOrchestrator(PortabilityOptions{
			Policies: []ExtensionPolicy{policy}, Presence: portabilityPresenceFake{present: true},
			Participants: []ExtensionParticipant{&portabilityParticipantFake{id: policy.ParticipantID}}, Allocator: &portabilityAllocatorFake{}, SnapshotReads: portabilityReadFactoryFake{},
		})
		if !errors.Is(err, ErrPortabilityUnavailable) {
			t.Fatalf("orchestrator without a participant deadline: %v", err)
		}
		orchestrator := orchestrate(t, &portabilityParticipantFake{id: policy.ParticipantID}, &portabilityAllocatorFake{}, time.Second, func(error) {})
		if orchestrator.Timeout() != budget || orchestrator.timeoutSeconds() != 0 {
			t.Fatalf("Timeout() = %v seconds %d", orchestrator.Timeout(), orchestrator.timeoutSeconds())
		}
	})

	t.Run("export expiry discards the late result and writes no catalog", func(t *testing.T) {
		participant := newBlocking(false)
		orchestrator := orchestrate(t, participant, &portabilityAllocatorFake{}, 2*time.Second, func(err error) { t.Errorf("unexpected fatal outcome: %v", err) })
		members := newArchiveTestMembers(t, Limits{})
		err := orchestrator.Export(t.Context(), portabilityQueryFake{}, incidentID, members, newActorReferenceSet())
		requireTimeout(t, err, "export")
		<-participant.returned
		for _, member := range members.Members() {
			if member.Path == extensionCatalogPath("profile") {
				t.Fatal("timed-out export published its extension catalog")
			}
		}
	})

	t.Run("import expiry abandons staged outputs and prepares no participant", func(t *testing.T) {
		members := newArchiveTestMembers(t, Limits{})
		exporter := orchestrate(t, &portabilityParticipantFake{id: policy.ParticipantID}, &portabilityAllocatorFake{}, time.Second, func(error) {})
		if err := exporter.Export(t.Context(), portabilityQueryFake{}, incidentID, members, newActorReferenceSet()); err != nil {
			t.Fatal(err)
		}
		members.Seal()
		participant := newBlocking(false)
		allocator := &portabilityAllocatorFake{}
		orchestrator := orchestrate(t, participant, allocator, 2*time.Second, func(err error) { t.Errorf("unexpected fatal outcome: %v", err) })
		prepared, err := orchestrator.PrepareImport(t.Context(), input, members)
		requireTimeout(t, err, "import")
		<-participant.returned
		if len(prepared.Participants) != 0 || len(prepared.Publications) != 0 {
			t.Fatal("timed-out import returned a prepared participant or publication")
		}
		if len(allocator.allocated) != 1 || len(allocator.abandoned) != 1 || allocator.allocated[0] != allocator.abandoned[0] {
			t.Fatalf("staged outputs allocated=%v abandoned=%v", allocator.allocated, allocator.abandoned)
		}
	})

	t.Run("a callback active after the grace selects the fatal outcome", func(t *testing.T) {
		participant := newBlocking(true)
		var once sync.Once
		fatal := make(chan error, 1)
		orchestrator := orchestrate(t, participant, &portabilityAllocatorFake{}, 20*time.Millisecond, func(err error) {
			once.Do(func() { fatal <- err })
		})
		done := make(chan error, 1)
		go func() {
			done <- orchestrator.Export(t.Context(), portabilityQueryFake{}, incidentID, newArchiveTestMembers(t, Limits{}), newActorReferenceSet())
		}()
		select {
		case err := <-fatal:
			if !fatalImportFailure(err) {
				t.Fatalf("grace exhaustion reason: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("no fatal outcome after the cancellation grace")
		}
		close(participant.release)
		if err := <-done; !fatalImportFailure(err) {
			t.Fatalf("late completion replaced the fatal outcome: %v", err)
		}
	})
}

// testDeadlineFailureResults: a participant or transaction deadline expiry is
// a terminal timeout with service_unavailable and its registered reason, not
// internal_error (Extensions Table 9-C, EXT-REQ-219, Table 26-B).
func testDeadlineFailureResults(t *testing.T) {
	t.Run("deadline failures finalize as service_unavailable timeouts", func(t *testing.T) {
		jobID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
		for _, tc := range []struct {
			cause   error
			reason  string
			seconds int64
		}{
			{&PortabilityTimeoutError{ProfileID: "profile", Operation: "export", TimeoutSeconds: 300}, "extension_participant_timeout", 300},
			{&PortabilityTimeoutError{ProfileID: "profile", Operation: "import", TimeoutSeconds: 7}, "extension_participant_timeout", 7},
		} {
			finalizer := &deadlineRecordingFinalizer{}
			sink := incidentBundleJobResultSink{manager: &recordingJobOperations{}, finalizer: finalizer, now: time.Now}
			sink.completeFailedFromError(t.Context(), jobs.Execution{}, "incident_bundle_export_rejected", tc.cause)
			if finalizer.timeouts != 1 || finalizer.failures != 0 {
				t.Fatalf("finalizer calls timeout=%d failure=%d", finalizer.timeouts, finalizer.failures)
			}
			summary := finalizer.completion.ErrorSummary
			if summary.Code != "service_unavailable" || !summary.Retryable || summary.Details["reason_code"] != tc.reason ||
				summary.Details["timeout_seconds"] != tc.seconds || summary.Details["operation_id"] != uuid.Nil.String() || len(summary.Details) != 3 {
				t.Fatalf("deadline completion = %#v", summary)
			}
		}
		failure, ok := importDeadlineFailure(fmtTimeout(crossownertransaction.ErrTimeout), timedCoordinatorFake{timeout: 30 * time.Second})
		if !ok || failure.reasonCode != "extension_transaction_timeout" || failure.timeoutSeconds != 30 {
			t.Fatalf("coordinator timeout classification = %#v %v", failure, ok)
		}
		completion := deadlineFailureCompletion(jobID, failure)
		if completion.ErrorSummary.Details["operation_id"] != jobID.String() {
			t.Fatalf("coordinator timeout operation_id = %#v", completion.ErrorSummary.Details)
		}
		if _, ok := importDeadlineFailure(ErrPortabilityResult, nil); ok {
			t.Fatal("an ordinary participant failure was classified as a deadline")
		}
	})
}

type deadlineRecordingFinalizer struct {
	JobSuccessFinalizer
	timeouts, failures int
	completion         jobs.FailureCompletion
}

func (f *deadlineRecordingFinalizer) FinalizeIncidentBundleJobTimeout(_ context.Context, request JobFailureFinalization) (jobs.Resource, error) {
	f.timeouts++
	f.completion = request.Completion
	return jobs.Resource{}, nil
}

func (f *deadlineRecordingFinalizer) FinalizeIncidentBundleJobFailure(_ context.Context, request JobFailureFinalization) (jobs.Resource, error) {
	f.failures++
	f.completion = request.Completion
	return jobs.Resource{}, nil
}

func fmtTimeout(err error) error {
	return errors.Join(errors.New("rollback proved commit absent"), err)
}

// timedCoordinatorFake exposes only the coordinator's effective timeout.
type timedCoordinatorFake struct{ timeout time.Duration }

func (f timedCoordinatorFake) Execute(context.Context, crossownertransaction.Operation) (crossownertransaction.Result, error) {
	return crossownertransaction.Result{}, crossownertransaction.ErrTimeout
}

func (f timedCoordinatorFake) Timeout() time.Duration { return f.timeout }
