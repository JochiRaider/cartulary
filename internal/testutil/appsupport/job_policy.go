package appsupport

import (
	"time"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// ShortJobRecoveryScan is a ConfigureRuntime hook for tests that restart a
// server holding recoverable jobs. A restarted runner's activation scan runs
// before admission opens, so recovered jobs otherwise wait one production 5 s
// recovery scan. Core 01 lets harness-owned composition inject shorter values;
// this shortens only the recovery-scan cadence and keeps every other
// production value, including leases and retry delays.
func ShortJobRecoveryScan(options *server.Options) {
	policy := jobs.ProductionRuntimePolicy()
	policy.RecoveryScan = 100 * time.Millisecond
	options.JobPolicy = &policy
}
