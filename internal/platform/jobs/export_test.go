package jobs

import (
	"context"
	"github.com/google/uuid"
	"time"
)

// ConfigureRunnerSynchronizationForTest replaces the private renewal clock
// and select-boundary hook. It exists only in the package's test build.
func ConfigureRunnerSynchronizationForTest(
	runner *Runner,
	renewalTicks <-chan time.Time,
	beforeWait func(),
	renewExecution func(context.Context, Execution) error,
) {
	if runner == nil {
		return
	}
	runner.renewalTicks = func(time.Duration) (<-chan time.Time, func()) {
		return renewalTicks, func() {}
	}
	runner.beforeWait = beforeWait
	if renewExecution != nil {
		runner.renewExecution = renewExecution
	}
}

// ConfigureRunnerReleaseForTest replaces the private conditional-release
// operation. It exists only in the package's test build.
func ConfigureRunnerReleaseForTest(
	runner *Runner,
	releaseExecution func(context.Context, Execution) error,
) {
	if runner == nil || releaseExecution == nil {
		return
	}
	runner.releaseExecution = releaseExecution
}

// RecoverableJobsForTest exercises the current scanner without exporting a second production dispatcher.
func (m *Manager) RecoverableJobsForTest(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if err := m.ensureConfigured(); err != nil {
		return nil, err
	}
	candidates, err := m.recoverableCandidatesForSelection(ctx, limit, m.transactions.selection, m.transactions.selection.jobKinds(), nil)
	if err != nil {
		return nil, err
	}
	jobIDs := make([]uuid.UUID, 0, len(candidates))
	for _, candidate := range candidates {
		jobIDs = append(jobIDs, candidate.JobID)
	}
	return jobIDs, nil
}
