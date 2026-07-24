// Package deadline owns the Extensions monotonic deadline and terminal-outcome
// policy shared by every timed extension protocol.
package deadline

import "math"

type Source string

const (
	SourceLocal     Source = "local"
	SourceInherited Source = "inherited"
)

type Deadline struct {
	MonotonicNS int64
	Source      Source
}

// New uses checked, saturating signed-64-bit arithmetic. Equality deliberately
// selects the inherited deadline so deadline ownership remains stable.
func New(nowMonotonicNS int64, timeoutSeconds int64, inherited *Deadline) Deadline {
	candidate := saturatingMonotonicNS(nowMonotonicNS, timeoutSeconds)
	if inherited != nil && inherited.MonotonicNS <= candidate {
		return Deadline{MonotonicNS: inherited.MonotonicNS, Source: SourceInherited}
	}
	return Deadline{MonotonicNS: candidate, Source: SourceLocal}
}

func (d Deadline) Expired(nowMonotonicNS int64) bool {
	return nowMonotonicNS >= d.MonotonicNS
}

func saturatingMonotonicNS(nowMonotonicNS int64, timeoutSeconds int64) int64 {
	if timeoutSeconds <= 0 {
		return nowMonotonicNS
	}
	if timeoutSeconds > math.MaxInt64/1_000_000_000 {
		return math.MaxInt64
	}
	delta := timeoutSeconds * 1_000_000_000
	if nowMonotonicNS > math.MaxInt64-delta {
		return math.MaxInt64
	}
	return nowMonotonicNS + delta
}

type CommitProof string

const (
	CommitProvenSuccessful CommitProof = "committed"
	CommitProvenAbsent     CommitProof = "absent"
	CommitIndeterminate    CommitProof = "indeterminate"
)

type Outcome string

const (
	OutcomeCommitted  Outcome = "committed"
	OutcomeCanceled   Outcome = "canceled"
	OutcomeTimedOut   Outcome = "timed_out"
	OutcomeFatal      Outcome = "fatal_indeterminate_commit"
	OutcomeInProgress Outcome = "in_progress"
)

// Classify gives commit proof precedence over cancellation and timeout. When
// cancellation and expiry have equal samples, expiry wins.
func Classify(proof CommitProof, cancellationSampleNS *int64, expirySampleNS *int64) Outcome {
	switch proof {
	case CommitProvenSuccessful:
		return OutcomeCommitted
	case CommitIndeterminate:
		return OutcomeFatal
	case CommitProvenAbsent:
		if cancellationSampleNS != nil && (expirySampleNS == nil || *cancellationSampleNS < *expirySampleNS) {
			return OutcomeCanceled
		}
		if expirySampleNS != nil {
			return OutcomeTimedOut
		}
	}
	return OutcomeInProgress
}
