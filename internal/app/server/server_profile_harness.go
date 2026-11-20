//go:build cartulary_harness

package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/JochiRaider/cartulary/internal/app/server/holdjob"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	networkflowharnesscontrol "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime/runevidence"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type harnessServerProfile struct{}

func newServerProfile() serverProfile {
	return harnessServerProfile{}
}

func (harnessServerProfile) validateEnvironment(func(string) (string, bool)) error {
	return nil
}

func (harnessServerProfile) runtimeOptions(lookup func(string) (string, bool)) Options {
	enabled, _ := lookup("CARTULARY_ENABLE_TEST_ROUTES")
	if enabled != "1" {
		return Options{}
	}

	testClock := httpapi.NewTestClock()
	networkFlowControls := networkflowharnesscontrol.NewControls()
	networkFlowControls.Crash = func() { os.Exit(86) }
	holdJobs := holdjob.NewControl()
	guardLog := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return Options{
		NetworkFlowComposition: networkFlowControls,
		Now:                    testClock.Now,
		HarnessJobs:            []jobs.Contribution{holdJobs.Contribution()},
		HTTP: httpapi.Options{
			// TH-HARNESS-REQ-858: an undeclared error code (Core 01
			// REQ-01-702) becomes 500 internal_error. Each violation is logged
			// to stdout with the server log stream for live diagnosis; stderr
			// carries only the one secret-safe lifecycle diagnostic. The same
			// finding is a contract evidence record (TH-HARNESS-REQ-860) that
			// fails the run at finalization. Every error a resolved operation
			// writes is also an error_emission coverage record, written once
			// per process (TH-HARNESS-REQ-862).
			ErrorCodeGuard: &httpapi.ErrorCodeGuard{Enforce: true, Report: func(operationID string, status int, code string) {
				guardLog.Warn("undeclared error code", slog.String("event", "undeclared_error_code"), slog.String("operation_id", operationID), slog.Int("status", status), slog.String("code", code))
				_ = runevidence.RecordTo(lookup, runevidence.Evidence{Kind: runevidence.UndeclaredErrorCode, OperationID: operationID, Status: status, Code: code, Process: runevidence.ServerHarness})
			}, Observe: func(operationID string, status int, code string) {
				_ = runevidence.RecordTo(lookup, runevidence.Evidence{Kind: runevidence.ErrorEmission, OperationID: operationID, Status: status, Code: code, Process: runevidence.ServerHarness})
			}},
			// TH-HARNESS-REQ-859: authenticating other than the declared
			// security (Core 01 REQ-01-703) becomes 500 internal_error when
			// nothing was written yet, with the same live log line and
			// contract evidence record.
			SecurityGuard: &httpapi.SecurityGuard{Enforce: true, Report: func(mismatch httpapi.SecurityMismatch) {
				guardLog.Warn("declared security mismatch", slog.String("event", "security_mismatch"), slog.String("operation_id", mismatch.OperationID), slog.Int("status", mismatch.Status), slog.String("declared", mismatch.Declared), slog.String("observed", mismatch.Observed))
				_ = runevidence.RecordTo(lookup, runevidence.Evidence{
					Kind: runevidence.SecurityMismatch, OperationID: mismatch.OperationID, Status: mismatch.Status,
					Security: &runevidence.SecurityFact{Declared: mismatch.Declared, Observed: mismatch.Observed}, Process: runevidence.ServerHarness,
				})
			}},
			AdditionalRoutes: append(
				harnessruntime.RegisterRoutes(testClock, networkFlowControls.Contribution(), harnessruntime.ControlContribution{Routes: []httpapi.RouteRegistrar{holdJobs.Routes()}}),
				auth.RegisterTestRoutes(),
				savedviews.RegisterTestRoutes(),
			),
		},
	}
}

// recordFatal writes the server-harness process's lifecycle fatal as one
// server_fatal contract evidence record carrying its reason code
// (TH-HARNESS-REQ-863); the run gate fails the run as server_fatal. The
// runner calls it once, beside the fatal diagnostic, before the process
// exits. A run without an explicit identity has no stream.
func (harnessServerProfile) recordFatal(lookup func(string) (string, bool), reasonCode string) {
	_ = runevidence.RecordTo(lookup, runevidence.Evidence{Kind: runevidence.ServerFatal, ReasonCode: reasonCode, Process: runevidence.ServerHarness})
}

func (harnessServerProfile) inheritedListenerFD(lookup func(string) (string, bool)) string {
	listenerFD, _ := lookup("CARTULARY_HTTP_LISTEN_FD")
	return listenerFD
}

func (harnessServerProfile) serve(ctx context.Context, handler http.Handler, options httpruntime.Options) error {
	return httpruntime.Serve(ctx, handler, options)
}
