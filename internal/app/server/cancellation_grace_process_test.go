package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

const graceProcessPhaseEnv = "CARTULARY_TEST_GRACE_PROCESS_PHASE"

// The child runs the production runner/lifecycle/invocation code with private
// fault-injected ports. Its callback never completes: only process exit can
// settle it. This is not evidence of a deployed server or a live HTTP listener.
func TestServerRunnerCancellationGrace_Process(t *testing.T) {
	if phase := os.Getenv(graceProcessPhaseEnv); phase != "" {
		os.Exit(runBlockedGraceProcess(t, phase))
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"construction", "serving", "cleanup", "draining"} {
		t.Run(phase, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, "-test.run=^TestServerRunnerCancellationGrace_Process$")
			command.Env = append(os.Environ(), graceProcessPhaseEnv+"="+phase, "TMPDIR="+t.TempDir())
			var stderr bytes.Buffer
			command.Stdout = io.Discard
			command.Stderr = &stderr
			err := command.Run()
			var exit *exec.ExitError
			if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != 70 {
				t.Fatalf("child exit = %v, deadline=%v, diagnostic=%q", err, ctx.Err(), stderr.String())
			}
			reason := "operation_cancellation_grace_exhausted"
			if phase == "draining" {
				reason = "published_component_lost"
			}
			want := "{\"code\":\"extension_integrity_failure\",\"reason_code\":\"" + reason + "\"}\n"
			if stderr.String() != want {
				t.Fatalf("child fatal diagnostic=%q, want %q", stderr.String(), want)
			}
		})
	}
}

func runBlockedGraceProcess(t *testing.T, phase string) int {
	runner := newServerRunner(io.Discard, os.Stderr)
	runner.loadConfig = serverTestConfiguration(t)
	var block func()
	runner.buildRuntime = func(_ context.Context, _ configassembly.Loaded, options Options) (serverRuntime, error) {
		lifecycle := processlifecycle.New()
		owner := &Runtime{lifecycle: lifecycle, immediateFatal: options.immediateFatal}
		block = func() {
			ctx, cancel := context.WithCancel(context.Background())
			entered := make(chan struct{})
			go func() { <-entered; cancel() }()
			_, _ = extensiondeadline.Invoke(ctx, extensiondeadline.InvocationPolicy{Grace: 5 * time.Millisecond, Exhausted: owner.fatalIntegrity}, func(context.Context) (struct{}, error) { close(entered); select {} })
		}
		if phase == "construction" {
			block()
		}
		runtime := testRuntime(http.NotFoundHandler(), func() {})
		runtime.FatalEvents = lifecycle.FatalEvents()
		runtime.ActivatePublication = lifecycle.Publish
		runtime.ShutdownDrainTimeout = time.Minute
		if phase == "cleanup" {
			runtime.Close = block
		}
		if phase == "draining" {
			runtime.ActivatePublication = func() error {
				if err := lifecycle.Publish(); err != nil {
					return err
				}
				lifecycle.Fatal("published_component_lost")
				return nil
			}
		}
		return runtime, nil
	}
	runner.profile = graceProcessProfile{phase: phase, block: func() { block() }}
	return runner.run(context.Background())
}

type graceProcessProfile struct {
	failingServerProfile
	phase string
	block func()
}

func (p graceProcessProfile) serve(ctx context.Context, _ http.Handler, options httpruntime.Options) error {
	if err := options.OnReady(); err != nil {
		return err
	}
	switch p.phase {
	case "serving":
		p.block()
	case "draining":
		<-ctx.Done()
		p.block()
	}
	return nil
}
