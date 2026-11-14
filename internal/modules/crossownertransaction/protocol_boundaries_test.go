package crossownertransaction

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func assertProtocolCardinalityLimits(t *testing.T) {
	t.Helper()
	t.Run("participant cardinality and identity admission", func(t *testing.T) {
		events := []string{}
		catalog := make([]Descriptor, 16384)
		participants := make([]Participant, 16384)
		for i := range catalog {
			id := fmt.Sprintf("p%05d", i)
			catalog[i] = descriptor(id)
			participants[len(participants)-1-i] = &participantFake{id: id, events: &events}
		}
		backend := &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}
		c := newTestCoordinator(t, catalog, backend, nil, nil)
		for _, count := range []int{1, 16384} {
			got, descriptors, err := c.resolve(participants[:count])
			if err != nil || len(got) != count || len(descriptors) != count {
				t.Fatalf("admitted count=%d: %v", count, err)
			}
			for i := range got {
				if got[i].ID() != descriptors[i].ParticipantID || (i > 0 && got[i-1].ID() >= got[i].ID()) {
					t.Fatal("participant/descriptor order changed")
				}
			}
		}
		if participants[0].ID() != "p16383" || len(events) != 0 || backend.beginCalls != 0 {
			t.Fatal("resolution mutated caller order or invoked work")
		}
		for name, set := range map[string][]Participant{
			"zero": {}, "over maximum": append(append([]Participant(nil), participants...), participants[0]),
			"duplicate": {participants[0], participants[0]}, "unadmitted": {&participantFake{id: "unknown", events: &events}},
		} {
			t.Run(name, func(t *testing.T) {
				op := operation(nil)
				op.Participants = set
				_, err := c.Execute(t.Context(), op)
				if !errors.Is(err, ErrParticipantSet) || backend.beginCalls != 0 || len(events) != 0 {
					t.Fatalf("participant rejection = %v begins=%d events=%v", err, backend.beginCalls, events)
				}
			})
		}
	})
	for _, count := range []int{1024, 1025} {
		t.Run(fmt.Sprintf("participant keys %d", count), func(t *testing.T) {
			backend := &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, nil)
			_, err := c.Execute(t.Context(), operation(&participantFake{id: "a", keys: boundaryKeys("a", count)}))
			if count == 1024 {
				if err != nil || backend.tx.commitCalls != 1 {
					t.Fatalf("maximum keys = %v", err)
				}
			} else if !errors.Is(err, ErrSerializationKeys) || backend.beginCalls != 0 {
				t.Fatalf("key overflow = %v begins=%d", err, backend.beginCalls)
			}
		})
	}
	for _, extra := range []bool{false, true} {
		t.Run(fmt.Sprintf("aggregate keys extra=%v", extra), func(t *testing.T) {
			events := []string{}
			backend := &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}
			catalog := []Descriptor{}
			participants := []Participant{}
			for _, id := range []string{"a", "b", "c", "d", "e"} {
				catalog = append(catalog, descriptor(id))
				count := 1024
				if id == "e" {
					count = 0
					if extra {
						count = 1
					}
				}
				participants = append(participants, &participantFake{id: id, keys: boundaryKeys(id, count), events: &events})
			}
			c := newTestCoordinator(t, catalog, backend, nil, nil)
			op := operation(participants[0])
			op.Participants = participants
			_, err := c.Execute(t.Context(), op)
			if !extra {
				if err != nil || backend.tx.commitCalls != 1 {
					t.Fatalf("aggregate maximum = %v", err)
				}
			} else if !errors.Is(err, ErrSerializationKeys) || backend.beginCalls != 0 || contains(events, "validate:a") {
				t.Fatalf("aggregate overflow = %v begins=%d", err, backend.beginCalls)
			}
		})
	}
	for _, tc := range []struct {
		name, key string
		valid     bool
	}{
		{"maximum", strings.Repeat("x", 512), true}, {"over", strings.Repeat("x", 513), false}, {"empty", "", false}, {"invalid UTF8", "\xff", false}, {"NUL", "a\x00b", false},
	} {
		t.Run("key scalar "+tc.name, func(t *testing.T) {
			backend := &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, nil)
			_, err := c.Execute(t.Context(), operation(&participantFake{id: "a", keys: []SerializationKey{{KeyKind: "a.kind", Key: tc.key}}}))
			if tc.valid {
				if err != nil || backend.tx.commitCalls != 1 {
					t.Fatalf("valid key = %v", err)
				}
			} else if !errors.Is(err, ErrSerializationKeys) || backend.beginCalls != 0 {
				t.Fatalf("bad key = %v begins=%d", err, backend.beginCalls)
			}
		})
	}
}

func boundaryKeys(id string, count int) []SerializationKey {
	keys := make([]SerializationKey, count)
	for i := range keys {
		keys[i] = SerializationKey{KeyKind: id + ".kind", Key: fmt.Sprintf("%04d", i)}
	}
	return keys
}

func assertDeadlineArithmeticAndInheritance(t *testing.T) {
	t.Helper()
	for _, tc := range []struct {
		now      int64
		duration time.Duration
		want     int64
	}{
		{100, time.Second, 100 + 1_000_000_000}, {math.MaxInt64 - 10, 10, math.MaxInt64}, {math.MaxInt64 - 10, 11, math.MaxInt64}, {math.MaxInt64, time.Minute, math.MaxInt64},
	} {
		if got := saturatingAdd(tc.now, tc.duration); got != tc.want {
			t.Fatalf("deadline(%d,%d)=%d want=%d", tc.now, tc.duration, got, tc.want)
		}
	}
	t.Run("saturated deadline reaches participant input", func(t *testing.T) {
		tx := &transactionFake{commitOutcome: CommitProven}
		c := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: tx}, func() int64 { return math.MaxInt64 - int64(time.Minute) }, nil)
		p := &deadlineInputParticipant{participantFake: &participantFake{id: "a"}}
		op := operation(p)
		op.Timeout = 2 * time.Minute
		_, err := c.Execute(context.Background(), op)
		if err != nil || tx.commitCalls != 1 || p.deadline != math.MaxInt64 {
			t.Fatalf("saturated protocol = %v deadline=%d commits=%d", err, p.deadline, tx.commitCalls)
		}
	})
	for _, expired := range []bool{false, true} {
		t.Run(fmt.Sprintf("inherited minimum expired=%v", expired), func(t *testing.T) {
			parentBudget := time.Minute
			if expired {
				parentBudget = -time.Second
			}
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(parentBudget))
			defer cancel()
			backend := &backendFake{tx: &transactionFake{commitOutcome: CommitProven}}
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, func() int64 { return 100 }, nil)
			p := &deadlineInputParticipant{participantFake: &participantFake{id: "a"}}
			op := operation(p)
			op.Timeout = 2 * time.Minute
			_, err := c.Execute(ctx, op)
			if expired {
				if !errors.Is(err, ErrTimeout) || backend.beginCalls != 0 || p.deadline != 0 {
					t.Fatalf("expired parent = %v begins=%d deadline=%d", err, backend.beginCalls, p.deadline)
				}
			} else if err != nil || backend.tx.commitCalls != 1 || p.deadline <= 100 || p.deadline > 100+int64(time.Minute) {
				t.Fatalf("inherited deadline = %v deadline=%d", err, p.deadline)
			}
		})
	}
}

type deadlineInputParticipant struct {
	*participantFake
	deadline int64
}

func (p *deadlineInputParticipant) BuildInput(ctx context.Context, op OperationContext) (Input, error) {
	p.deadline = op.DeadlineMonotonicNS
	return p.participantFake.BuildInput(ctx, op)
}
