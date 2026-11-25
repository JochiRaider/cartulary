// Package testdeadline bounds test work by the test's own deadline.
//
// TH-HARNESS-REQ-864: a hang guard only stops a stuck test, so load alone
// must not fail it. A guard therefore ends with the test's deadline, less a
// settlement margin in which the failure, its evidence and the test's
// cleanup still finish before the test binary's own timeout. A fixed
// deadline remains right for a wait the test asserts.
package testdeadline

import (
	"context"
	"testing"
	"time"
)

// Settlement is left between a guard's deadline and the test's.
const Settlement = 30 * time.Second

// Context returns the test's context bounded by its deadline less
// Settlement, never below one second. A test without a deadline keeps only
// its context, which ends with the test.
func Context(t testing.TB) (context.Context, context.CancelFunc) {
	if test, ok := t.(interface{ Deadline() (time.Time, bool) }); ok {
		if deadline, bounded := test.Deadline(); bounded {
			return context.WithTimeout(t.Context(), max(time.Until(deadline)-Settlement, time.Second))
		}
	}
	return context.WithCancel(t.Context())
}
