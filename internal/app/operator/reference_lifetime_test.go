package operator

import (
	"bytes"
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/referencetest"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestReferenceCleanupGrace_Process(t *testing.T) {
	const child = "CARTULARY_TEST_OPERATOR_REFERENCE_CLEANUP"
	if os.Getenv(child) == "1" {
		lifetime, err := referenceVerificationLifetime(5*time.Millisecond, os.Stderr)
		if err != nil {
			t.Fatal(err)
		}
		referencetest.BlockedImportCleanup(t, lifetime)
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "-test.run=^TestReferenceCleanupGrace_Process$")
	cmd.Env = append(os.Environ(), child+"=1", "TMPDIR="+t.TempDir())
	var stderr bytes.Buffer
	cmd.Stdout = io.Discard
	cmd.Stderr = &stderr
	err = cmd.Run()
	var exit *exec.ExitError
	if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 70 {
		t.Fatal("fatal cleanup process", err, ctx.Err(), stderr.String())
	}
	if stderr.String() != "{\"code\":\"extension_integrity_failure\",\"reason_code\":\"operation_cancellation_grace_exhausted\"}\n" {
		t.Fatal("unsafe fatal diagnostic", stderr.String())
	}
}
