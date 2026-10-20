package postgrescleanup

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

func testCheckpointDiagnostics(t *testing.T) {
	stamp := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	first := &DropWaitSnapshot{Requested: 2, Done: 1, Buffers: 10, WriteMS: 100, SyncMS: 20, statsReset: stamp}
	last := *first
	last.Requested++
	last.Done++
	last.Buffers += 3
	last.WriteMS += 25
	last.SyncMS += 2
	interval := checkpointInterval(first, &last)
	if !interval.valid() || interval.Status != "sampled" || interval.WriteMS != 25 || interval.SyncMS != 2 || interval.Buffers != 3 {
		t.Fatalf("invalid delta: %+v", interval)
	}
	for _, mutate := range []func(*DropWaitSnapshot){
		func(s *DropWaitSnapshot) { s.statsReset = s.statsReset.Add(time.Second) },
		func(s *DropWaitSnapshot) { s.Requested = 0 },
		func(s *DropWaitSnapshot) { s.SyncMS = 0 },
	} {
		reset := last
		mutate(&reset)
		if got := checkpointInterval(first, &reset); got.Status != "reset" || !got.valid() {
			t.Fatal("reset accepted as measured interval")
		}
	}
	if checkpointInterval(nil, &last).Status != "not_observed" {
		t.Fatal("missing sample claimed")
	}
	missing := last
	missing.statsReset = time.Time{}
	if checkpointInterval(first, &missing).Status != "unavailable" {
		t.Fatal("unknown epoch claimed")
	}
	for _, invalid := range []float64{-1, math.NaN(), math.Inf(1), float64(maxDiagnosticCounter) * 2} {
		last.SyncMS = invalid
		if last.valid() {
			t.Fatal("invalid statistic admitted")
		}
	}

	line := "2026-10-19T00:00:03.000000000Z 2026-10-19 00:00:03.000 UTC [123] LOG:  checkpoint complete: wrote 10 buffers (1.0%), wrote 0 SLRU buffers; 0 WAL file(s) added, 0 removed, 0 recycled; write=1.234 s, sync=0.056 s, total=1.300 s; sync files=4, longest=0.040 s, average=0.014 s; distance=2 kB, estimate=3 kB; lsn=0/10, redo lsn=0/F"
	value, ok := ParseCheckpointCompletion(line, stamp)
	if !ok || value.ElapsedMS != 3000 || value.WriteMS != 1234 || value.SyncMS != 56 || value.Files != 4 || value.LongestSyncMS != 40 || value.AverageSyncMS != 14 {
		t.Fatal("checkpoint numeric projection")
	}
	encoded, _ := json.Marshal(value)
	for _, private := range []string{"UTC", "LOG:", "lsn", "[123]", "2026"} {
		if strings.Contains(string(encoded), private) {
			t.Fatal("raw server text retained")
		}
	}
	for _, invalid := range []string{line + " private", strings.Replace(line, "sync=0.056", "sync=NaN", 1), strings.Replace(line, "sync=0.056", "sync=-0.056", 1), strings.Replace(line, "checkpoint complete:", "unknown complete:", 1), strings.Repeat("x", 4097)} {
		if _, ok := ParseCheckpointCompletion(invalid, stamp); ok {
			t.Fatal("unknown log format admitted")
		}
	}
	if _, ok := ParseCheckpointCompletion(line, stamp.Add(time.Hour)); ok {
		t.Fatal("negative relative time admitted")
	}
	attempt := UnavailableAttempt()
	for _, invalid := range []string{"", "0", "private DSN", strings.Repeat("1", 21)} {
		attempt.ServiceGeneration = &invalid
		if attempt.Validate() == nil {
			t.Fatal("unbounded service generation admitted")
		}
	}
	generation := "12345"
	attempt.ServiceGeneration = &generation
	if attempt.Validate() != nil {
		t.Fatal("valid service generation rejected")
	}

	failure := &attemptError{err: context.DeadlineExceeded, observation: UnavailableAttempt()}
	if !errors.Is(failure, context.DeadlineExceeded) || ObservationFromError(failure).Wait.Status != "unavailable" {
		t.Fatal("diagnostics changed failure")
	}
}
