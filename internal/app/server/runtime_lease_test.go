package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

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
		return processLease.State() == processlease.StateUncertain && !lifecycle.AdmissionOpen()
	}, "application-process uncertainty did not close admission")

	processSession.setProof(processlease.ProofContinuous)
	waitForRuntimeLeaseCondition(t, func() bool {
		return processLease.State() == processlease.StateHeld && lifecycle.AdmissionOpen()
	}, "original application-process session did not restore admission")

	processSession.setIdentity("recreated-session")
	waitForRuntimeLeaseCondition(t, func() bool {
		return processLease.State() == processlease.StateLost &&
			lifecycle.FatalReason() == processlease.ErrApplicationProcessLeaseLost.FatalReasonCode() &&
			!lifecycle.AdmissionOpen()
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
		return servingLease.State() == processlease.StateUncertain && !lifecycle.AdmissionOpen()
	}, "Recovery-serving uncertainty did not close admission")

	servingSession.setProof(processlease.ProofContinuous)
	waitForRuntimeLeaseCondition(t, func() bool {
		return servingLease.State() == processlease.StateHeld && lifecycle.AdmissionOpen()
	}, "original Recovery-serving session did not restore admission")

	servingSession.setIdentity("recreated-serving-session")
	waitForRuntimeLeaseCondition(t, func() bool {
		return servingLease.State() == processlease.StateLost &&
			lifecycle.FatalReason() == processlease.ErrRecoveryServingLeaseLost.FatalReasonCode() &&
			!lifecycle.AdmissionOpen()
	}, "Recovery-serving loss did not use its distinct fatal outcome")
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
