package policy

import "time"

const (
	// CleanupDelay leaves one scheduler interval of margin before the adopted
	// one-hour object-byte deletion deadline.
	CleanupDelay = 45 * time.Minute

	AllowedNonTerminalFinalizeFailures = 3
	TerminalFinalizeAttempt            = AllowedNonTerminalFinalizeFailures + 1
)

type FailureSchedule struct {
	FailedAt     time.Time
	CleanupDueAt time.Time
}

func ScheduleFailure(now time.Time) FailureSchedule {
	failedAt := now.UTC()
	return FailureSchedule{
		FailedAt:     failedAt,
		CleanupDueAt: failedAt.Add(CleanupDelay),
	}
}

func FinalizeAttemptIsTerminal(nextAttemptCount int) bool {
	return nextAttemptCount >= TerminalFinalizeAttempt
}
