package recoverycli

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
)

func TestRunnerLeavesDueInvocationUnboundedForPerAttemptTimeouts(t *testing.T) {
	facade := &deadlineInspectingFacade{t: t}
	var stdout bytes.Buffer
	handled, exitCode := (Runner{
		Stdout: &stdout,
		Facade: facade,
	}).Run(context.Background(), []string{
		"restore-verify",
		"due",
		"--target-config-file",
		"/tmp/cartulary-target.toml",
		"--timeout-seconds",
		"60",
	})
	if !handled || exitCode != 0 {
		t.Fatalf("due runner got handled=%t exit=%d stdout=%s", handled, exitCode, stdout.String())
	}
	if !facade.called {
		t.Fatal("due facade was not called")
	}
}

type deadlineInspectingFacade struct {
	t      *testing.T
	called bool
}

func (facade *deadlineInspectingFacade) BackupInspectLatest(context.Context, application.BackupInspectLatestRequest, application.ProgressSink) (application.Result, error) {
	panic("unexpected BackupInspectLatest call")
}

func (facade *deadlineInspectingFacade) BackupCreate(context.Context, application.BackupCreateRequest, application.ProgressSink) (application.Result, error) {
	panic("unexpected BackupCreate call")
}

func (facade *deadlineInspectingFacade) RestoreLatest(context.Context, application.RestoreLatestRequest, application.ProgressSink) (application.Result, error) {
	panic("unexpected RestoreLatest call")
}

func (facade *deadlineInspectingFacade) RestoreVerifyLatest(context.Context, application.RestoreVerifyLatestRequest, application.ProgressSink) (application.Result, error) {
	panic("unexpected RestoreVerifyLatest call")
}

func (facade *deadlineInspectingFacade) RestoreVerifyDue(ctx context.Context, request application.RestoreVerifyDueRequest, _ application.ProgressSink) (application.Result, error) {
	facade.called = true
	if _, ok := ctx.Deadline(); ok {
		facade.t.Fatal("restore_verify_due received an outer CLI deadline")
	}
	if request.AttemptTimeout != time.Minute {
		facade.t.Fatalf("attempt timeout got %s want 1m", request.AttemptTimeout)
	}
	if request.OperationID == uuid.Nil {
		facade.t.Fatal("operation ID is empty")
	}
	return application.Result{ArtifactRefs: []application.ArtifactRef{}, Status: application.ResultNoOp}, nil
}
