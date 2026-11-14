package crossownertransaction

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"
)

func assertPhaseInputOwnership(t *testing.T) {
	t.Helper()
	t.Run("each phase receives the original admitted input and scoped context", func(t *testing.T) {
		p := &inputOwnershipParticipant{t: t, original: []byte("{\"ok\":true}\n")}
		tx := &transactionFake{commitOutcome: CommitProven}
		c := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: tx}, func() int64 { return 100 }, nil)
		result, err := c.Execute(t.Context(), operation(p))
		if err != nil || tx.commitCalls != 1 || tx.rollbackCalls.Load() != 0 || result.ParticipantValues["a"] != "receipt" || !reflect.DeepEqual(p.phases, []string{"prepare", "validate", "write"}) {
			t.Fatalf("phase ownership outcome: err=%v commits=%d rollbacks=%d phases=%v result=%v", err, tx.commitCalls, tx.rollbackCalls.Load(), p.phases, result)
		}
		for i, retained := range p.retained {
			if retained[0] != byte('1'+i) {
				t.Errorf("phase %d retained bytes changed: %q", i, retained)
			}
		}
	})
}

type inputOwnershipParticipant struct {
	t        *testing.T
	original []byte
	phases   []string
	retained [][]byte
}

func (*inputOwnershipParticipant) ID() string { return "a" }
func (p *inputOwnershipParticipant) BuildInput(_ context.Context, op OperationContext) (Input, error) {
	if op.OperationID != "operation" || op.NormalizedRequestSHA256 != "digest" || op.DeadlineMonotonicNS != 100+int64(time.Minute) {
		p.t.Errorf("input context = %+v", op)
	}
	return Input{SchemaID: "a.input", CanonicalBytes: p.original}, nil
}
func (p *inputOwnershipParticipant) observe(inv Invocation, phase string) {
	p.t.Helper()
	if inv.SchemaID != "cartulary.extension_transaction_participant_context.v1" || inv.Phase != phase || inv.OperationID != "operation" || inv.ParticipantID != "a" || inv.OwnerProfileID != "profile" || inv.NormalizedRequestSHA256 != "digest" || inv.CancellationRequested || inv.DeadlineMonotonicNS != 100+int64(time.Minute) || inv.Input.SchemaID != "a.input" || !bytes.Equal(inv.Input.CanonicalBytes, []byte("{\"ok\":true}\n")) {
		p.t.Errorf("incorrect %s invocation: %+v input=%q", phase, inv, inv.Input.CanonicalBytes)
	}
	if phase == "prepare" {
		if inv.ReadAccess != nil || inv.WriteAccess != nil {
			p.t.Error("prepare received transaction authority")
		}
	} else {
		if inv.ReadAccess == nil || inv.ReadAccess.ParticipantScope() != "a" {
			p.t.Error("wrong read scope")
		}
		if phase == "validate" && inv.WriteAccess != nil {
			p.t.Error("validator received write authority")
		}
		if phase == "write" && (inv.WriteAccess == nil || inv.WriteAccess.ParticipantScope() != "a") {
			p.t.Error("wrong write scope")
		}
	}
	p.original[0] = '!'
	inv.Input.CanonicalBytes[0] = byte('1' + len(p.phases))
	p.retained = append(p.retained, inv.Input.CanonicalBytes)
	p.phases = append(p.phases, phase)
}
func (p *inputOwnershipParticipant) Prepare(_ context.Context, inv Invocation) (PrepareResult, error) {
	p.observe(inv, "prepare")
	return PrepareResult{SerializationKeys: []SerializationKey{}}, nil
}
func (p *inputOwnershipParticipant) Validate(_ context.Context, inv Invocation) (ValidationResult, error) {
	p.observe(inv, "validate")
	return Valid(), nil
}
func (p *inputOwnershipParticipant) Write(_ context.Context, inv Invocation) (WriteResult, error) {
	p.observe(inv, "write")
	return Written("receipt"), nil
}
