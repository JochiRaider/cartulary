package jobs

import (
	"context"
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
