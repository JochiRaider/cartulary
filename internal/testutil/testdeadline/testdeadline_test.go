package testdeadline

import (
	"testing"
	"time"
)

type deadlineTB struct {
	testing.TB
	deadline time.Time
	bounded  bool
}

func (d deadlineTB) Deadline() (time.Time, bool) { return d.deadline, d.bounded }

func TestContextFollowsTheTestDeadline(t *testing.T) {
	for _, tc := range []struct {
		name      string
		remaining time.Duration
		bounded   bool
		want      time.Duration
	}{
		{"settlement is left before a long deadline", 10 * time.Minute, true, 10*time.Minute - Settlement},
		{"a near deadline keeps one second", 5 * time.Second, true, time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := Context(deadlineTB{TB: t, deadline: time.Now().Add(tc.remaining), bounded: tc.bounded})
			defer cancel()
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) > tc.want || time.Until(deadline) < tc.want-5*time.Second {
				t.Fatalf("deadline in %v, want about %v", time.Until(deadline), tc.want)
			}
		})
	}
	t.Run("an unbounded test keeps only its context", func(t *testing.T) {
		ctx, cancel := Context(deadlineTB{TB: t})
		defer cancel()
		if _, ok := ctx.Deadline(); ok {
			t.Fatal("unbounded test got a deadline")
		}
		cancel()
		if ctx.Err() == nil {
			t.Fatal("cancel did not end the context")
		}
	})
}
