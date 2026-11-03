package crossownertransaction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func protocolLimitsFixture() Limits {
	return Limits{MinimumParticipants: 1, Participants: 16384, ParticipantInputBytes: 67108864, AggregateInputBytes: 67108864,
		PrepareResultBytes: 67108864, AggregatePrepareResultBytes: 67108864, ResultBytes: 1048576,
		SerializationKeysPerParticipant: 1024, AggregateSerializationKeys: 4096, ValidationFindings: 256}
}

func assertProtocolByteLimits(t *testing.T) {
	assertValidationByteLimits(t)
	for _, ceiling := range []int64{67108864, 1048576, math.MaxInt64} {
		for _, total := range []int64{0, ceiling / 2} {
			for _, delta := range []int64{-1, 0, 1} {
				if ceiling == math.MaxInt64 && total == 0 && delta == 1 {
					continue
				}
				size := ceiling - total + delta
				got, ok := boundedTotal(total, size, ceiling, ceiling)
				if ok != (delta <= 0) || ok && got != total+size {
					t.Fatalf("boundedTotal(%d,%d,%d) = %d,%v", total, size, ceiling, got, ok)
				}
			}
		}
	}
	for _, delta := range []int{-1, 0, 1} {
		for _, aggregate := range []bool{false, true} {
			t.Run(fmt.Sprintf("64MiB delta=%d aggregate=%v", delta, aggregate), func(t *testing.T) {
				events := []string{}
				tx := &transactionFake{commitOutcome: CommitProven}
				backend := &backendFake{tx: tx}
				coordinator := newTestCoordinator(t, []Descriptor{descriptor("a"), descriptor("b")}, backend, nil, nil)
				size := 67108864
				if aggregate {
					size /= 2
				}
				input := canonicalStringInput(size + delta)
				participants := []Participant{&participantFake{id: "a", input: input, events: &events}}
				if aggregate {
					participants = append(participants, &participantFake{id: "b", input: canonicalStringInput(size), events: &events})
				}
				op := operation(participants[0])
				op.Participants = participants
				_, err := coordinator.Execute(context.Background(), op)
				if delta <= 0 {
					if err != nil || tx.commitCalls != 1 {
						t.Fatalf("accepted boundary = %v, commits=%d", err, tx.commitCalls)
					}
					return
				}
				if !errors.Is(err, ErrInput) || backend.beginCalls != 0 || contains(events, "prepare:a") {
					t.Fatalf("overflow = %v, begins=%d, events=%v", err, backend.beginCalls, events)
				}
			})
		}
	}
	t.Run("prepare canonical bytes exceed validation budget", func(t *testing.T) {
		coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}, nil, nil)
		keys := make([]SerializationKey, 1024)
		for i := range keys {
			keys[i] = SerializationKey{KeyKind: "a.kind", Key: fmt.Sprintf("%04d", i) + strings.Repeat("\x01", 508)}
		}
		ordered, size, err := coordinator.validatePrepare(descriptor("a"), PrepareResult{SerializationKeys: keys})
		if err != nil || size <= 1048576 || size >= 67108864 {
			t.Fatalf("prepare size=%d err=%v", size, err)
		}
		canonical, err := canonicaljson.MarshalExtension(map[string]any{"schema_id": PrepareResultSchema, "participant_id": "a", "serialization_keys": prepareKeyObjects(ordered)})
		if err != nil || int64(len(canonical)) != size || !bytes.Contains(canonical, []byte(`"participant_id":"a"`)) || bytes.Contains(canonical, []byte(`"ParticipantID"`)) || canonical[len(canonical)-1] != '\n' {
			t.Fatal("prepare bytes do not use the closed canonical envelope")
		}
		if _, err := coordinator.Execute(context.Background(), operation(&participantFake{id: "a", keys: keys})); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("closed write envelope byte gate", func(t *testing.T) {
		coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{}, nil, nil)
		canonical, err := canonicaljson.MarshalExtension(map[string]any{"schema_id": WriteResultSchema, "participant_id": "a", "status": "written"})
		if err != nil {
			t.Fatal(err)
		}
		for _, delta := range []int64{-1, 0, 1} {
			coordinator.limits.ResultBytes = int64(len(canonical)) + delta
			if coordinator.validWriteResult(descriptor("a"), Written(make(chan int))) != (delta >= 0) {
				t.Fatalf("write envelope budget delta=%d", delta)
			}
		}
	})
}

func canonicalStringInput(size int) []byte {
	value := bytes.Repeat([]byte{'x'}, size)
	value[0] = '"'
	value[size-2] = '"'
	value[size-1] = '\n'
	return value
}

func assertValidationByteLimits(t *testing.T) {
	for _, delta := range []int{-1, 0, 1} {
		t.Run(fmt.Sprintf("1MiB validation delta=%d", delta), func(t *testing.T) {
			result := ValidationResult{Status: "invalid", Findings: make([]Finding, 256)}
			for i := range result.Findings {
				result.Findings[i] = Finding{Path: fmt.Sprintf("$.value_%04d", i), ReasonCode: "fixture_invalid", Message: "Invalid value.", Details: []byte(`{"padding":""}`)}
			}
			encode := func() []byte {
				value, err := canonicaljson.MarshalExtension(map[string]any{"schema_id": ValidationResultSchema, "participant_id": "a", "status": result.Status, "findings": result.Findings})
				if err != nil {
					t.Fatal(err)
				}
				return value
			}
			extra := 1048576 + delta - len(encode())
			for i := range result.Findings {
				n := extra / (len(result.Findings) - i)
				if n > 4096 {
					t.Fatal("fixture exceeded its declared detail bound")
				}
				result.Findings[i].Details = []byte(`{"padding":"` + strings.Repeat("x", n) + `"}`)
				extra -= n
			}
			if len(encode()) != 1048576+delta {
				t.Fatal("incorrect validation boundary fixture")
			}
			events := []string{}
			tx := &transactionFake{events: &events, commitOutcome: CommitProven}
			coordinator := newTestCoordinator(t, []Descriptor{descriptor("a"), descriptor("b")}, &backendFake{tx: tx}, nil, nil)
			coordinator.validationPolicy = boundedFindingPolicy{}
			op := operation(&participantFake{id: "a", validation: result, events: &events})
			op.Participants = append(op.Participants, &participantFake{id: "b", events: &events})
			_, err := coordinator.Execute(context.Background(), op)
			var failure *ValidationError
			want := "cross_owner_transaction_validation_failed"
			if delta > 0 {
				want = "extension_validation_result_invalid"
			}
			if !errors.As(err, &failure) || failure.ReasonCode != want || !tx.rolledBack || tx.commitCalls != 0 || contains(events, "validate:b") || contains(events, "write:a") {
				t.Fatalf("validation boundary=%v events=%v", err, events)
			}
		})
	}
}

// The fixture supplies a closed, bounded owner finding policy. Actual packaged
// owner policy admission is covered by application composition integration.
type boundedFindingPolicy struct{ validationPolicyFixture }

func (boundedFindingPolicy) AcceptFinding(_ Descriptor, f Finding) bool {
	var details map[string]string
	if json.Unmarshal(f.Details, &details) != nil || len(details) != 1 {
		return false
	}
	padding, ok := details["padding"]
	return ok && len(padding) <= 4096 && strings.Trim(padding, "x") == "" && strings.HasPrefix(f.Path, "$.value_") && f.ReasonCode == "fixture_invalid" && f.Message == "Invalid value."
}
