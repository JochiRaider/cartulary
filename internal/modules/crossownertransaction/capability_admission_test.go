package crossownertransaction

import (
	"context"
	"errors"
	"testing"
	"time"
)

func assertNilCapabilityAdmission(t *testing.T) {
	t.Helper()
	for _, typed := range []bool{false, true} {
		for _, dependency := range []string{"backend", "policy"} {
			t.Run(dependency+nilCaseName(typed), func(t *testing.T) {
				opts := Options{Limits: protocolLimitsFixture(), ValidationPolicy: validationPolicyFixture{}, Backend: &backendFake{}, Timeout: time.Minute, FatalSink: func(error) { t.Fatal("unexpected fatal") }}
				switch dependency {
				case "backend":
					opts.Backend = nil
					if typed {
						opts.Backend = (*backendFake)(nil)
					}
				case "policy":
					opts.ValidationPolicy = nil
					if typed {
						opts.ValidationPolicy = (*validationPolicyFixture)(nil)
					}
				}
				if got, err := New(opts); got != nil || !errors.Is(err, ErrUnavailable) {
					t.Fatalf("incomplete construction = %v, %v", got, err)
				}
			})
		}
		for _, slot := range []int{0, 1} {
			t.Run("participant"+nilCaseName(typed)+string(rune('0'+slot)), func(t *testing.T) {
				events := []string{}
				backend := &backendFake{tx: &transactionFake{}}
				c := newTestCoordinator(t, []Descriptor{descriptor("a"), descriptor("b")}, backend, nil, nil)
				p := []Participant{&participantFake{id: "a", events: &events}, &participantFake{id: "b", events: &events}}
				p[slot] = nil
				if typed {
					p[slot] = (*participantFake)(nil)
				}
				op := operation(p[0])
				op.Participants = p
				if _, err := c.Execute(t.Context(), op); !errors.Is(err, ErrParticipantSet) || backend.beginCalls != 0 || len(events) != 0 {
					t.Fatalf("participant admission = %v begins=%d events=%v", err, backend.beginCalls, events)
				}
			})
		}
	}
	for _, port := range []string{"preconditions", "finalizer"} {
		t.Run("optional typed nil "+port, func(t *testing.T) {
			events := []string{}
			backend := &backendFake{tx: &transactionFake{}}
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, nil)
			op := operation(&participantFake{id: "a", events: &events})
			if port == "preconditions" {
				op.Preconditions = (*preconditionsFake)(nil)
			} else {
				op.Finalizer = (*finalizerFake)(nil)
			}
			if _, err := c.Execute(t.Context(), op); !errors.Is(err, ErrInput) || backend.beginCalls != 0 || len(events) != 0 {
				t.Fatalf("optional port = %v begins=%d events=%v", err, backend.beginCalls, events)
			}
		})
	}
	for _, typed := range []bool{false, true} {
		t.Run("transaction"+nilCaseName(typed), func(t *testing.T) {
			backend := &admissionBackend{}
			if typed {
				backend.tx = (*transactionFake)(nil)
			}
			fatalCalls := 0
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, func(error) { fatalCalls++ })
			result, err := c.Execute(t.Context(), operation(&participantFake{id: "a"}))
			if !IsFatalIntegrity(err) || fatalCalls != 1 || backend.calls != 1 || len(result.ParticipantValues) != 0 {
				t.Fatalf("unusable transaction = %v fatal=%d begins=%d", err, fatalCalls, backend.calls)
			}
		})
		for _, phase := range []string{"preconditions", "read", "write", "finalize"} {
			t.Run("capability "+phase+nilCaseName(typed), func(t *testing.T) {
				events := []string{}
				tx := &admissionTransaction{transactionFake: &transactionFake{events: &events}, phase: phase, typed: typed}
				backend := &admissionBackend{tx: tx}
				c := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, nil)
				op := operation(&participantFake{id: "a", events: &events})
				op.Preconditions = preconditionsFake{events: &events}
				op.Finalizer = finalizerFake{events: &events}
				result, err := c.Execute(t.Context(), op)
				want := ErrValidation
				forbidden := "validate:a"
				if phase == "write" {
					want, forbidden = ErrWrite, "write:a"
				} else if phase == "finalize" {
					want, forbidden = ErrWrite, "finalize"
				}
				if !errors.Is(err, want) || backend.calls != 1 || tx.rollbackCalls.Load() != 1 || tx.commitCalls != 0 || contains(events, forbidden) || len(result.ParticipantValues) != 0 {
					t.Fatalf("capability admission = %v begins=%d rollback=%d commit=%d events=%v", err, backend.calls, tx.rollbackCalls.Load(), tx.commitCalls, events)
				}
			})
		}
	}
}

func nilCaseName(typed bool) string {
	if typed {
		return " typed nil"
	}
	return " nil"
}

type admissionBackend struct {
	tx    Transaction
	calls int
}

func (b *admissionBackend) Begin(context.Context, []Descriptor) (Transaction, error) {
	b.calls++
	return b.tx, nil
}

type admissionTransaction struct {
	*transactionFake
	phase string
	typed bool
}

func (t *admissionTransaction) ReadCapability(id string) (ReadCapability, error) {
	if t.phase == "read" {
		if t.typed {
			return (*capabilityFake)(nil), nil
		}
		return nil, nil
	}
	return t.transactionFake.ReadCapability(id)
}
func (t *admissionTransaction) WriteCapability(id string) (WriteCapability, error) {
	if t.phase == "write" {
		if t.typed {
			return (*capabilityFake)(nil), nil
		}
		return nil, nil
	}
	return t.transactionFake.WriteCapability(id)
}
func (t *admissionTransaction) PreconditionCapability() (PreconditionCapability, error) {
	if t.phase == "preconditions" {
		if t.typed {
			return (*preconditionCapabilityFake)(nil), nil
		}
		return nil, nil
	}
	return t.transactionFake.PreconditionCapability()
}
func (t *admissionTransaction) FinalizationCapability() (FinalizationCapability, error) {
	if t.phase == "finalize" {
		if t.typed {
			return (*finalizationCapabilityFake)(nil), nil
		}
		return nil, nil
	}
	return t.transactionFake.FinalizationCapability()
}
