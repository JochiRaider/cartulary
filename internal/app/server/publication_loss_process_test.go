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
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
)

const publicationLossPhaseEnv = "CARTULARY_TEST_PUBLICATION_LOSS_PHASE"

// Real subprocess exits exercise the production runner, publication controller,
// and lifecycle through private fault injection. Live listener readiness has a
// separate witness; this profile intentionally does not open a socket.
func TestServerRunnerPublicationLoss_Process(t *testing.T) {
	if phase := os.Getenv(publicationLossPhaseEnv); phase != "" {
		os.Exit(runPublicationLossProcess(t, phase))
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"prepared", "committed", "serving", "private", "activation_failure"} {
		t.Run(phase, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, "-test.run=^TestServerRunnerPublicationLoss_Process$")
			command.Env = append(os.Environ(), publicationLossPhaseEnv+"="+phase, "TMPDIR="+t.TempDir())
			var stderr bytes.Buffer
			command.Stdout, command.Stderr = io.Discard, &stderr
			err := command.Run()
			var exit *exec.ExitError
			wantExit := 70
			if phase == "activation_failure" {
				wantExit = 2
			}
			if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != wantExit {
				t.Fatalf("child exit=%v deadline=%v diagnostic=%q", err, ctx.Err(), stderr.String())
			}
			if phase == "activation_failure" {
				requirePublicationProcessDiagnostic(t, stderr.Bytes(), 30)
				return
			}
			if want := "{\"code\":\"extension_integrity_failure\",\"reason_code\":\"published_component_lost\"}\n"; stderr.String() != want {
				t.Fatalf("child diagnostic=%q want=%q", stderr.String(), want)
			}
		})
	}
}

func runPublicationLossProcess(t *testing.T, phase string) int {
	runner := newServerRunner(io.Discard, os.Stderr)
	runner.loadConfig = serverTestConfiguration(t)
	controller, _, lifecycle := preparedPublicationController(t)
	if phase != "prepared" {
		if err := controller.commit(); err != nil {
			t.Fatal(err)
		}
		acknowledgePreparedPublicationWorkers(t, controller)
	}
	runner.buildRuntime = func(_ context.Context, _ configassembly.Loaded, options Options) (serverRuntime, error) {
		owner := &Runtime{publication: controller, lifecycle: lifecycle, immediateFatal: options.immediateFatal}
		if phase == "activation_failure" {
			owner.evidenceCleanupDispatcher = &serverCleanupLifecycle{startErr: errors.New("preparation failed"), readiness: lifecycle.AdmissionOpen}
		}
		beginTestPublication(t, owner)
		runtime := testRuntime(http.NotFoundHandler(), owner.Close)
		runtime.publication = owner.publicationOperation
		runtime.ActivatePublication = owner.ActivatePublication
		runtime.FatalEvents = lifecycle.FatalEvents()
		runtime.ShutdownDrainTimeout = time.Second
		switch phase {
		case "prepared":
			controller.componentLost("http")
		case "committed":
			controller.componentLost("job_dequeue")
		case "private":
			controller.privateComponentLost()
		}
		return runtime, nil
	}
	runner.profile = publicationLossProfile{controller: controller}
	return runner.run(context.Background())
}

type publicationLossProfile struct {
	failingServerProfile
	controller *publicationController
}

func (p publicationLossProfile) serve(ctx context.Context, _ http.Handler, options httpruntime.Options) error {
	if err := options.OnReady(); err != nil {
		return &httpruntime.StartupError{Err: err}
	}
	p.controller.componentLost("websocket")
	<-ctx.Done()
	return nil
}
