package extensions

import "math"

type DeadlineSource string

const (
	DeadlineLocal     DeadlineSource = "local"
	DeadlineInherited DeadlineSource = "inherited"
)

type Deadline struct {
	MonotonicNS int64
	Source      DeadlineSource
}

// NewDeadline uses checked, saturating signed-64-bit arithmetic. Equality
// deliberately selects the inherited deadline so ownership remains stable.
func NewDeadline(nowMonotonicNS int64, timeoutSeconds int64, inherited *Deadline) Deadline {
	candidate := saturatingDeadline(nowMonotonicNS, timeoutSeconds)
	if inherited != nil && inherited.MonotonicNS <= candidate {
		return Deadline{MonotonicNS: inherited.MonotonicNS, Source: DeadlineInherited}
	}
	return Deadline{MonotonicNS: candidate, Source: DeadlineLocal}
}

func (d Deadline) Expired(nowMonotonicNS int64) bool {
	return nowMonotonicNS >= d.MonotonicNS
}

func saturatingDeadline(nowMonotonicNS int64, timeoutSeconds int64) int64 {
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

type DeadlineOutcome string

const (
	DeadlineCommitted  DeadlineOutcome = "committed"
	DeadlineCanceled   DeadlineOutcome = "canceled"
	DeadlineTimedOut   DeadlineOutcome = "timed_out"
	DeadlineFatal      DeadlineOutcome = "fatal_indeterminate_commit"
	DeadlineInProgress DeadlineOutcome = "in_progress"
)

func ClassifyDeadlineOutcome(proof CommitProof, cancellationSampleNS *int64, expirySampleNS *int64) DeadlineOutcome {
	switch proof {
	case CommitProvenSuccessful:
		return DeadlineCommitted
	case CommitIndeterminate:
		return DeadlineFatal
	case CommitProvenAbsent:
		if cancellationSampleNS != nil && (expirySampleNS == nil || *cancellationSampleNS < *expirySampleNS) {
			return DeadlineCanceled
		}
		if expirySampleNS != nil {
			return DeadlineTimedOut
		}
	}
	return DeadlineInProgress
}
