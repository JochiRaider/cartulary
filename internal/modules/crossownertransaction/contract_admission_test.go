package crossownertransaction

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func assertProtocolContractAdmission(t *testing.T) {
	assertCanonicalTransactionInputs(t)
	for _, test := range []struct{ name, input string }{
		{"malformed", `{"ok":`},
		{"duplicate", "{\"ok\":true,\"ok\":false}\n"},
		{"unpaired surrogate", "{\"ok\":\"\\ud800\"}\n"},
		{"unknown field", "{\"extra\":1,\"ok\":true}\n"},
		{"wrong type", "{\"ok\":1}\n"},
	} {
		t.Run("input admission "+test.name, func(t *testing.T) {
			events := []string{}
			backend := &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}
			coordinator := newTestCoordinator(t, []Descriptor{descriptor("a"), descriptor("b")}, backend, nil, nil)
			op := operation(&participantFake{id: "a", input: []byte(test.input), events: &events})
			op.Participants = append(op.Participants, &participantFake{id: "b", events: &events})
			_, err := coordinator.Execute(context.Background(), op)
			if !errors.Is(err, ErrInput) || backend.beginCalls != 0 || contains(events, "input:b") || contains(events, "prepare:a") {
				t.Fatalf("input admission = %v begins=%d events=%v", err, backend.beginCalls, events)
			}
		})
	}
	t.Run("prepare order is checked before union sorting", func(t *testing.T) {
		events := []string{}
		backend := &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}
		coordinator := newTestCoordinator(t, []Descriptor{descriptor("a"), descriptor("b")}, backend, nil, nil)
		op := operation(&participantFake{id: "a", keys: []SerializationKey{{KeyKind: "a.kind", Key: "z"}, {KeyKind: "a.kind", Key: "a"}}, events: &events})
		op.Participants = append(op.Participants, &participantFake{id: "b", events: &events})
		_, err := coordinator.Execute(context.Background(), op)
		if !errors.Is(err, ErrSerializationKeys) || backend.beginCalls != 0 || contains(events, "prepare:b") {
			t.Fatalf("prepare admission = %v begins=%d events=%v", err, backend.beginCalls, events)
		}
	})
	for _, absent := range []bool{false, true} {
		t.Run(fmt.Sprintf("prepare array absent=%v", absent), func(t *testing.T) {
			backend := &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}
			coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, nil)
			keys := []SerializationKey{}
			if absent {
				keys = nil
			}
			_, err := coordinator.Execute(context.Background(), operation(prepareArrayParticipant{participantFake: &participantFake{id: "a"}, keys: keys}))
			if absent {
				if !errors.Is(err, ErrPrepare) || backend.beginCalls != 0 {
					t.Fatalf("absent array = %v, begins=%d", err, backend.beginCalls)
				}
			} else if err != nil || backend.tx.commitCalls != 1 {
				t.Fatalf("empty array = %v", err)
			}
		})
	}

	for _, status := range []string{"", "success", "Written"} {
		t.Run("invalid write status "+status, func(t *testing.T) {
			events := []string{}
			backend := &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}
			coordinator := newTestCoordinator(t, []Descriptor{descriptor("a"), descriptor("b")}, backend, nil, nil)
			op := operation(invalidWriteParticipant{participantFake: &participantFake{id: "a", events: &events}, status: status})
			op.Participants = append(op.Participants, &participantFake{id: "b", events: &events})
			_, err := coordinator.Execute(context.Background(), op)
			if !errors.Is(err, ErrWrite) || !backend.tx.rolledBack || backend.tx.commitCalls != 0 || contains(events, "write:b") {
				t.Fatalf("write result = %v rollback=%v commits=%d events=%v", err, backend.tx.rolledBack, backend.tx.commitCalls, events)
			}
		})
	}

}

type prepareArrayParticipant struct {
	*participantFake
	keys []SerializationKey
}

func (p prepareArrayParticipant) Prepare(context.Context, Invocation) (PrepareResult, error) {
	return PrepareResult{SerializationKeys: p.keys}, nil
}

type invalidWriteParticipant struct {
	*participantFake
	status string
}

func (p invalidWriteParticipant) Write(context.Context, Invocation) (WriteResult, error) {
	p.event("write:" + p.id)
	return WriteResult{Status: p.status}, nil
}

// The test input schema admits all JSON values so these vectors isolate byte
// admission from schema rejection. Expected bytes are independently authored.
type canonicalInputPolicy struct{ validationPolicyFixture }

func (canonicalInputPolicy) AcceptInput(Descriptor, any) bool { return true }
func assertCanonicalTransactionInputs(t *testing.T) {
	for _, tt := range []struct {
		name, input string
		valid       bool
	}{
		{"exact nested unicode and integer", "{\"a\":[null,true,-123,9007199254740993],\"z\":\"</&é😀\"}\n", true},
		{"empty array", "[]\n", true},
		{"missing LF", "{\"ok\":true}", false},
		{"extra LF", "{\"ok\":true}\n\n", false},
		{"space", "{ \"ok\":true}\n", false},
		{"BOM", "\xef\xbb\xbf{\"ok\":true}\n", false},
		{"reordered keys", "{\"z\":1,\"a\":2}\n", false},
		{"alternate escape", "\"\\u0061\"\n", false},
		{"negative zero", "-0\n", false},
		{"fraction", "1.0\n", false},
		{"exponent", "1e0\n", false},
		{"trailing value", "{\"ok\":true}\n{}", false},
		{"invalid UTF8", "\"\xff\"\n", false},
	} {
		t.Run("canonical input "+tt.name, func(t *testing.T) {
			events := []string{}
			backend := &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}
			c := newTestCoordinator(t, []Descriptor{descriptor("a"), descriptor("b")}, backend, nil, nil)
			c.validationPolicy = canonicalInputPolicy{}
			op := operation(&participantFake{id: "a", input: []byte(tt.input), events: &events})
			op.Participants = append(op.Participants, &participantFake{id: "b", events: &events})
			_, err := c.Execute(context.Background(), op)
			if tt.valid {
				if err != nil || backend.tx.commitCalls != 1 {
					t.Fatalf("valid canonical input = %v", err)
				}
				return
			}
			if !errors.Is(err, ErrInput) || backend.beginCalls != 0 || contains(events, "input:b") || contains(events, "prepare:a") {
				t.Fatalf("noncanonical input = %v begins=%d events=%v", err, backend.beginCalls, events)
			}
		})
	}
}
