package server

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

// TestRuntimeApplicationProcessLeaseClosesAndRestoresAdmissionForOriginalSessionOnly
// covers Core 04 REQ-04-145: uncertain ownership of the application-process
// lease pauses admission, so the gate answers admission_paused (AC-618); only
// the original session's proof reopens it, and a recreated session is lost.
func TestRuntimeApplicationProcessLeaseClosesAndRestoresAdmissionForOriginalSessionOnly(t *testing.T) {
	processSession := newRuntimeLeaseSession("application-session")
	processLease := acquireRuntimeTestLease(t, processSession)
	servingLease := acquireRuntimeTestLease(t, newRuntimeLeaseSession("serving-session"))
	lifecycle := processlifecycle.New()
	if err := lifecycle.Publish(); err != nil {
		t.Fatal(err)
	}
	runtime := &Runtime{
		processLease: &processlease.ApplicationProcessLease{Lease: processLease},
		servingLease: &processlease.ApplicationRecoveryServingLease{Lease: servingLease},
		lifecycle:    lifecycle,
	}
	monitorCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	processLease.StartMonitor(monitorCtx)
	go runtime.watchProcessLease(monitorCtx)

	processSession.setProof(processlease.ProofUncertain)
	waitForRuntimeLeaseCondition(t, func() bool {
		return processLease.State() == processlease.StateUncertain && !lifecycle.AdmissionOpen() && lifecycle.AdmissionPaused()
	}, "application-process uncertainty did not pause admission")

	processSession.setProof(processlease.ProofContinuous)
	waitForRuntimeLeaseCondition(t, func() bool {
		return processLease.State() == processlease.StateHeld && lifecycle.AdmissionOpen() && !lifecycle.AdmissionPaused()
	}, "original application-process session did not restore admission")

	processSession.setIdentity("recreated-session")
	waitForRuntimeLeaseCondition(t, func() bool {
		return processLease.State() == processlease.StateLost &&
			lifecycle.FatalReason() == processlease.ErrApplicationProcessLeaseLost.FatalReasonCode() &&
			!lifecycle.AdmissionOpen() && !lifecycle.AdmissionPaused()
	}, "recreated application-process session was not rejected irreversibly")
}

func TestRuntimeRecoveryServingLeaseHasDistinctFatalOutcome(t *testing.T) {
	processLease := acquireRuntimeTestLease(t, newRuntimeLeaseSession("application-session"))
	servingSession := newRuntimeLeaseSession("serving-session")
	servingLease := acquireRuntimeTestLease(t, servingSession)
	lifecycle := processlifecycle.New()
	if err := lifecycle.Publish(); err != nil {
		t.Fatal(err)
	}
	runtime := &Runtime{
		processLease: &processlease.ApplicationProcessLease{Lease: processLease},
		servingLease: &processlease.ApplicationRecoveryServingLease{Lease: servingLease},
		lifecycle:    lifecycle,
	}
	monitorCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	servingLease.StartMonitor(monitorCtx)
	go runtime.watchServingLease(monitorCtx)

	servingSession.setProof(processlease.ProofUncertain)
	waitForRuntimeLeaseCondition(t, func() bool {
		return servingLease.State() == processlease.StateUncertain && !lifecycle.AdmissionOpen() && lifecycle.AdmissionPaused()
	}, "Recovery-serving uncertainty did not pause admission")

	servingSession.setProof(processlease.ProofContinuous)
	waitForRuntimeLeaseCondition(t, func() bool {
		return servingLease.State() == processlease.StateHeld && lifecycle.AdmissionOpen() && !lifecycle.AdmissionPaused()
	}, "original Recovery-serving session did not restore admission")

	servingSession.setIdentity("recreated-serving-session")
	waitForRuntimeLeaseCondition(t, func() bool {
		return servingLease.State() == processlease.StateLost &&
			lifecycle.FatalReason() == processlease.ErrRecoveryServingLeaseLost.FatalReasonCode() &&
			!lifecycle.AdmissionOpen() && !lifecycle.AdmissionPaused()
	}, "Recovery-serving loss did not use its distinct fatal outcome")
}

// TestRuntimePublicationHonoursLeaseOwnership covers Core 04 REQ-04-145,
// EXT-REQ-213 and AC-618: the gate opens only while every lease is held. A
// lease uncertain when publication completes leaves admission paused, and
// serving-gated work waits, until the original session proves ownership.
// Uncertainty is not a publication failure, confirmed loss before the gate
// opens is the fatal lease loss, and a restore never reopens admission over
// the other lease's uncertainty. No lease watcher runs here, so the gate
// itself must consult ownership.
func TestRuntimePublicationHonoursLeaseOwnership(t *testing.T) {
	type publicationLeases struct {
		processSession *runtimeLeaseSession
		servingSession *runtimeLeaseSession
		process        *processlease.Lease
		serving        *processlease.Lease
	}
	acquire := func(t *testing.T) publicationLeases {
		t.Helper()
		leases := publicationLeases{
			processSession: newRuntimeLeaseSession("application-session"),
			servingSession: newRuntimeLeaseSession("serving-session"),
		}
		// A long detection deadline keeps an uncertain lease uncertain for
		// the whole subtest instead of letting it expire to lost.
		leases.process = acquireRuntimeTestLeaseWithin(t, leases.processSession, 20*time.Second)
		leases.serving = acquireRuntimeTestLeaseWithin(t, leases.servingSession, 20*time.Second)
		monitorCtx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		leases.process.StartMonitor(monitorCtx)
		leases.serving.StartMonitor(monitorCtx)
		return leases
	}
	publishingRuntime := func(t *testing.T, leases publicationLeases, preparation *leaseOwnershipPreparation) (*Runtime, *processlifecycle.Controller) {
		t.Helper()
		controller, _, lifecycle := preparedPublicationController(t)
		if err := controller.commit(); err != nil {
			t.Fatal(err)
		}
		acknowledgePreparedPublicationWorkers(t, controller)
		runtime := &Runtime{
			publication:               controller,
			lifecycle:                 lifecycle,
			processLease:              &processlease.ApplicationProcessLease{Lease: leases.process},
			servingLease:              &processlease.ApplicationRecoveryServingLease{Lease: leases.serving},
			evidenceCleanupDispatcher: preparation,
		}
		beginTestPublication(t, runtime)
		return runtime, lifecycle
	}
	requireServingWorkWaits := func(t *testing.T, preparation *leaseOwnershipPreparation) {
		t.Helper()
		select {
		case <-preparation.started:
			t.Fatal("serving-gated work started while lease ownership was uncertain")
		default:
		}
	}

	t.Run("uncertainty during preparation publishes with admission paused", func(t *testing.T) {
		leases := acquire(t)
		preparation := newLeaseOwnershipPreparation(func() bool {
			leases.processSession.setProof(processlease.ProofUncertain)
			return awaitLeaseState(leases.process, processlease.StateUncertain)
		})
		runtime, lifecycle := publishingRuntime(t, leases, preparation)
		if err := runtime.ActivatePublication(); err != nil {
			t.Fatalf("uncertain ownership failed publication: %v", err)
		}
		if !preparation.reached.Load() {
			t.Fatal("the application-process lease never became uncertain during preparation")
		}
		if lifecycle.AdmissionOpen() || !lifecycle.AdmissionPaused() {
			t.Fatalf("publication opened admission over an uncertain application-process lease: open=%t paused=%t", lifecycle.AdmissionOpen(), lifecycle.AdmissionPaused())
		}
		requireServingWorkWaits(t, preparation)

		leases.processSession.setProof(processlease.ProofContinuous)
		waitForRuntimeLeaseConditionWithin(t, 5*time.Second, func() bool {
			return leases.process.State() == processlease.StateHeld
		}, "the original application-process session did not prove ownership")
		if !lifecycle.RestoreAdmission() || !lifecycle.AdmissionOpen() {
			t.Fatal("the original session's proof did not open the gate")
		}
		select {
		case <-preparation.started:
		case <-time.After(5 * time.Second):
			t.Fatal("serving-gated work did not start when the gate opened")
		}
	})

	t.Run("uncertainty before activation is not a publication failure", func(t *testing.T) {
		leases := acquire(t)
		leases.servingSession.setProof(processlease.ProofUncertain)
		waitForRuntimeLeaseConditionWithin(t, 5*time.Second, func() bool {
			return leases.serving.State() == processlease.StateUncertain
		}, "the recovery-serving lease did not become uncertain")
		preparation := newLeaseOwnershipPreparation(nil)
		runtime, lifecycle := publishingRuntime(t, leases, preparation)
		if err := runtime.ActivatePublication(); err != nil {
			t.Fatalf("uncertain ownership failed publication: %v", err)
		}
		if lifecycle.AdmissionOpen() || !lifecycle.AdmissionPaused() {
			t.Fatalf("publication opened admission over an uncertain recovery-serving lease: open=%t paused=%t", lifecycle.AdmissionOpen(), lifecycle.AdmissionPaused())
		}
		requireServingWorkWaits(t, preparation)
	})

	for _, tc := range []struct {
		name   string
		lose   func(publicationLeases)
		wait   func(publicationLeases) *processlease.Lease
		reason string
	}{
		{
			name:   "confirmed application-process loss before activation is fatal",
			lose:   func(leases publicationLeases) { leases.processSession.setIdentity("recreated-session") },
			wait:   func(leases publicationLeases) *processlease.Lease { return leases.process },
			reason: processlease.ErrApplicationProcessLeaseLost.FatalReasonCode(),
		},
		{
			name:   "confirmed recovery-serving loss before activation is fatal",
			lose:   func(leases publicationLeases) { leases.servingSession.setIdentity("recreated-serving-session") },
			wait:   func(leases publicationLeases) *processlease.Lease { return leases.serving },
			reason: processlease.ErrRecoveryServingLeaseLost.FatalReasonCode(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			leases := acquire(t)
			tc.lose(leases)
			lost := tc.wait(leases)
			waitForRuntimeLeaseConditionWithin(t, 5*time.Second, func() bool {
				return lost.State() == processlease.StateLost
			}, "the lease was not lost")
			runtime, lifecycle := publishingRuntime(t, leases, newLeaseOwnershipPreparation(nil))
			err := runtime.ActivatePublication()
			var fatal interface{ FatalReasonCode() string }
			if !errors.As(err, &fatal) || fatal.FatalReasonCode() != tc.reason {
				t.Fatalf("activation over a lost lease = %v, want fatal %s (exit 70)", err, tc.reason)
			}
			if lifecycle.AdmissionOpen() {
				t.Fatal("a lost lease opened admission")
			}
		})
	}

	t.Run("a restore never reopens admission over the other lease's uncertainty", func(t *testing.T) {
		leases := acquire(t)
		preparation := newLeaseOwnershipPreparation(nil)
		runtime, lifecycle := publishingRuntime(t, leases, preparation)
		if err := runtime.ActivatePublication(); err != nil {
			t.Fatal(err)
		}
		if !lifecycle.AdmissionOpen() {
			t.Fatal("held leases did not open admission")
		}
		leases.servingSession.setProof(processlease.ProofUncertain)
		waitForRuntimeLeaseConditionWithin(t, 5*time.Second, func() bool {
			return leases.serving.State() == processlease.StateUncertain
		}, "the recovery-serving lease did not become uncertain")
		// The recovery-serving watcher pauses admission; a process-lease
		// watcher acting on a stale view of both leases then restores.
		lifecycle.CloseAdmission()
		if lifecycle.RestoreAdmission() || lifecycle.AdmissionOpen() || !lifecycle.AdmissionPaused() {
			t.Fatal("a restore reopened admission while the recovery-serving lease was uncertain")
		}
		leases.servingSession.setProof(processlease.ProofContinuous)
		waitForRuntimeLeaseConditionWithin(t, 5*time.Second, func() bool {
			return leases.serving.State() == processlease.StateHeld
		}, "the original recovery-serving session did not prove ownership")
		if !lifecycle.RestoreAdmission() || !lifecycle.AdmissionOpen() {
			t.Fatal("proof from every lease did not reopen admission")
		}
	})
}

// leaseOwnershipPreparation is a serving-gated component whose preparation
// can change lease ownership before publication completes.
type leaseOwnershipPreparation struct {
	during  func() bool
	reached atomic.Bool
	started chan struct{}
	cancel  context.CancelFunc
	done    chan struct{}
}

func newLeaseOwnershipPreparation(during func() bool) *leaseOwnershipPreparation {
	return &leaseOwnershipPreparation{during: during, started: make(chan struct{})}
}

func (preparation *leaseOwnershipPreparation) Prepare(parent context.Context, serving <-chan struct{}) error {
	if preparation.during != nil {
		preparation.reached.Store(preparation.during())
	}
	ctx, cancel := context.WithCancel(parent)
	preparation.cancel = cancel
	preparation.done = make(chan struct{})
	go func() {
		defer close(preparation.done)
		select {
		case <-ctx.Done():
		case <-serving:
			close(preparation.started)
		}
	}()
	return nil
}

func (preparation *leaseOwnershipPreparation) Close(context.Context) error {
	if preparation.cancel != nil {
		preparation.cancel()
		<-preparation.done
	}
	return nil
}

// awaitLeaseState waits, without a testing.TB, for a lease state; it runs on
// the publication goroutine.
func awaitLeaseState(lease *processlease.Lease, state processlease.State) bool {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if lease.State() == state {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return false
}

func acquireRuntimeTestLeaseWithin(t testing.TB, session *runtimeLeaseSession, lossDetection time.Duration) *processlease.Lease {
	t.Helper()
	lease, err := processlease.Acquire(
		context.Background(),
		runtimeLeaseBackend{session: session},
		100*time.Millisecond,
		lossDetection,
	)
	if err != nil {
		t.Fatalf("acquire runtime test lease: %v", err)
	}
	t.Cleanup(lease.Close)
	return lease
}

func waitForRuntimeLeaseConditionWithin(t testing.TB, within time.Duration, condition func() bool, failure string) {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal(failure)
}

func acquireRuntimeTestLease(t testing.TB, session *runtimeLeaseSession) *processlease.Lease {
	t.Helper()
	lease, err := processlease.Acquire(
		context.Background(),
		runtimeLeaseBackend{session: session},
		100*time.Millisecond,
		30*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("acquire runtime test lease: %v", err)
	}
	t.Cleanup(lease.Close)
	return lease
}

func waitForRuntimeLeaseCondition(t testing.TB, condition func() bool, failure string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal(failure)
}

type runtimeLeaseBackend struct {
	session *runtimeLeaseSession
}

func (backend runtimeLeaseBackend) Open(context.Context) (processlease.Session, error) {
	return backend.session, nil
}

type runtimeLeaseSession struct {
	mu       sync.RWMutex
	identity string
	proof    processlease.Proof
}

func newRuntimeLeaseSession(identity string) *runtimeLeaseSession {
	return &runtimeLeaseSession{identity: identity, proof: processlease.ProofContinuous}
}

func (session *runtimeLeaseSession) Identity() string {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.identity
}

func (session *runtimeLeaseSession) TryAcquire(context.Context) (bool, error) {
	return true, nil
}

func (session *runtimeLeaseSession) Prove(context.Context) processlease.Proof {
	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.proof
}

func (session *runtimeLeaseSession) Release(context.Context) error {
	return nil
}

func (session *runtimeLeaseSession) Close() {}

func (session *runtimeLeaseSession) setProof(proof processlease.Proof) {
	session.mu.Lock()
	session.proof = proof
	session.mu.Unlock()
}

func (session *runtimeLeaseSession) setIdentity(identity string) {
	session.mu.Lock()
	session.identity = identity
	session.mu.Unlock()
}
