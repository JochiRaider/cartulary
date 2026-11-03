package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

type publicationPreparation func(context.Context) error

func (p publicationPreparation) Prepare(ctx context.Context, _ <-chan struct{}) error { return p(ctx) }
func (p publicationPreparation) Close(context.Context) error                          { return nil }

func assertPublicationDeadlineBoundaries(t *testing.T) {
	for _, phase := range []string{"preparation", "activation", "cancel", "equal cancellation", "panic at expiry"} {
		t.Run("publication boundary "+phase, func(t *testing.T) {
			coordinator, plan := generatedPublicationPlan(t)
			lifecycle := processlifecycle.New()
			owner := &Runtime{lifecycle: lifecycle, publication: newPublicationController(lifecycle)}
			var clock atomic.Int64
			clock.Store(1)
			parent, cancel := context.WithCancel(t.Context())
			defer cancel()
			expire := func() { clock.Store(owner.publicationOperation.policy.Deadline.MonotonicNS) }
			if phase == "activation" {
				owner.evidenceCleanupDispatcher = publicationPreparation(func(context.Context) error { expire(); return nil })
			}
			err := owner.beginPublication(parent, coordinator, publicationTiming{seconds: 9, grace: time.Second, now: clock.Load}, func(context.Context) error {
				if err := owner.publication.prepare(plan); err != nil {
					return err
				}
				if err := owner.publication.commit(); err != nil {
					return err
				}
				acknowledgePreparedPublicationWorkers(t, owner.publication)
				switch phase {
				case "preparation":
					expire()
				case "cancel":
					cancel()
				case "equal cancellation":
					expire()
					cancel()
				case "panic at expiry":
					expire()
					panic("PRIVATE_PUBLICATION_PAYLOAD")
				}
				return nil
			})
			t.Cleanup(owner.Close)
			if phase == "activation" {
				if err != nil {
					t.Fatal(err)
				}
				err = owner.ActivatePublication()
			}
			var finding *extensions.StartupValidationError
			if !errors.As(err, &finding) || len(finding.Findings) != 1 {
				t.Fatalf("publication finding=%v", err)
			}
			f := finding.Findings[0]
			if f.ReasonCode != "extension_publication_failed" || f.Message != "Extension publication did not reach the serving state." || len(f.Details) != 2 || f.Details["phase"] != "publication" || f.Details["timeout_seconds"] != int64(9) {
				t.Fatalf("publication finding=%#v", f)
			}
			if phase == "cancel" {
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("cancellation=%v", err)
				}
			} else if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("deadline=%v", err)
			}
			if strings.Contains(err.Error(), "PRIVATE_") || lifecycle.AdmissionOpen() || owner.publication.currentState() != publicationFailed || lifecycle.FatalReason() != "" {
				t.Fatalf("failed publication escaped: %v", err)
			}
			if err := owner.ActivatePublication(); err == nil || lifecycle.AdmissionOpen() {
				t.Fatal("late activation escaped publication failure")
			}
		})
	}
	t.Run("serving lifetime outlives startup timer", func(t *testing.T) {
		coordinator, plan := generatedPublicationPlan(t)
		lifecycle := processlifecycle.New()
		owner := &Runtime{lifecycle: lifecycle, publication: newPublicationController(lifecycle)}
		if err := owner.beginPublication(t.Context(), coordinator, publicationTiming{seconds: 1, grace: time.Second}, func(context.Context) error {
			if err := owner.publication.prepare(plan); err != nil {
				return err
			}
			if err := owner.publication.commit(); err != nil {
				return err
			}
			acknowledgePreparedPublicationWorkers(t, owner.publication)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(owner.Close)
		if err := owner.ActivatePublication(); err != nil {
			t.Fatal(err)
		}
		select {
		case <-owner.publicationOperation.ctx.Done():
			t.Fatal("serving lifetime canceled by startup completion/timer")
		case <-time.After(1100 * time.Millisecond):
		}
		if !lifecycle.AdmissionOpen() {
			t.Fatal("publication lost its serving gate")
		}
		var fatal interface{ FatalReasonCode() string }
		if err := owner.ActivatePublication(); !errors.As(err, &fatal) || fatal.FatalReasonCode() != "published_component_lost" || lifecycle.AdmissionOpen() {
			t.Fatalf("duplicate activation retained a serving epoch: %v", err)
		}
	})
}

const publicationDeadlinePhaseEnv = "CARTULARY_TEST_PUBLICATION_DEADLINE_PHASE"

func TestServerRunnerPublicationDeadline_Process(t *testing.T) {
	if phase := os.Getenv(publicationDeadlinePhaseEnv); phase != "" {
		os.Exit(runPublicationDeadlineProcess(t, phase))
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"preparation", "listener", "activation", "blocked_preparation", "blocked_listener", "blocked_activation", "fatal_during_preparation"} {
		t.Run(phase, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, "-test.run=^TestServerRunnerPublicationDeadline_Process$")
			command.Env = append(os.Environ(), publicationDeadlinePhaseEnv+"="+phase, "TMPDIR="+t.TempDir())
			var stderr bytes.Buffer
			command.Stdout, command.Stderr = io.Discard, &stderr
			err := command.Run()
			var exit *exec.ExitError
			fatal := strings.HasPrefix(phase, "blocked_") || phase == "fatal_during_preparation"
			want := 2
			if fatal {
				want = 70
			}
			if ctx.Err() != nil || !errors.As(err, &exit) || exit.ExitCode() != want {
				t.Fatalf("exit=%v deadline=%v diagnostic=%q", err, ctx.Err(), stderr.String())
			}
			if fatal {
				reason := "operation_cancellation_grace_exhausted"
				if phase == "fatal_during_preparation" {
					reason = "published_component_lost"
				}
				want := "{\"code\":\"extension_integrity_failure\",\"reason_code\":\"" + reason + "\"}\n"
				if stderr.String() != want {
					t.Fatalf("fatal diagnostic=%q want=%q", stderr.String(), want)
				}
			} else {
				requirePublicationProcessDiagnostic(t, stderr.Bytes(), 1)
			}
		})
	}
}
func requirePublicationProcessDiagnostic(t *testing.T, data []byte, seconds float64) {
	t.Helper()
	var diagnostic struct {
		Error struct {
			Code    string `json:"code"`
			Details struct {
				Items []extensions.StartupFinding `json:"items"`
			} `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &diagnostic); err != nil || diagnostic.Error.Code != "invalid_deployment_config" || len(diagnostic.Error.Details.Items) != 1 {
		t.Fatalf("publication diagnostic=%q: %v", data, err)
	}
	f := diagnostic.Error.Details.Items[0]
	if f.Path != "$" || f.ReasonCode != "extension_publication_failed" || f.Message != "Extension publication did not reach the serving state." || len(f.Details) != 2 || f.Details["phase"] != "publication" || f.Details["timeout_seconds"] != seconds {
		t.Fatalf("publication finding=%#v", f)
	}
}
func runPublicationDeadlineProcess(t *testing.T, phase string) int {
	runner := newServerRunner(io.Discard, os.Stderr)
	runner.loadConfig = serverTestConfiguration(t)
	var owner *Runtime
	runner.buildRuntime = func(ctx context.Context, _ configassembly.Loaded, options Options) (serverRuntime, error) {
		coordinator, plan := generatedPublicationPlan(t)
		lifecycle := processlifecycle.New()
		owner = &Runtime{lifecycle: lifecycle, publication: newPublicationController(lifecycle), immediateFatal: options.immediateFatal}
		if strings.HasSuffix(phase, "activation") {
			owner.evidenceCleanupDispatcher = publicationPreparation(func(ctx context.Context) error {
				if phase == "blocked_activation" {
					select {}
				}
				<-ctx.Done()
				return errors.New("PRIVATE_ACTIVATION_PAYLOAD")
			})
		}
		grace := time.Second
		if strings.HasPrefix(phase, "blocked_") || phase == "fatal_during_preparation" {
			grace = 0
		}
		err := owner.beginPublication(ctx, coordinator, publicationTiming{seconds: 1, grace: grace}, func(ctx context.Context) error {
			if err := owner.publication.prepare(plan); err != nil {
				return err
			}
			if phase == "fatal_during_preparation" {
				owner.publication.componentLost("http")
				select {}
			}
			if phase == "blocked_preparation" {
				select {}
			}
			if phase == "preparation" {
				<-ctx.Done()
				return errors.New("PRIVATE_PREPARATION_PAYLOAD")
			}
			if err := owner.publication.commit(); err != nil {
				return err
			}
			acknowledgePreparedPublicationWorkers(t, owner.publication)
			return nil
		})
		if err != nil {
			owner.Close()
			return serverRuntime{}, err
		}
		runtime := testRuntime(http.NotFoundHandler(), owner.Close)
		runtime.publication = owner.publicationOperation
		runtime.ActivatePublication = owner.ActivatePublication
		runtime.FatalEvents = lifecycle.FatalEvents()
		return runtime, nil
	}
	runner.profile = publicationDeadlineProfile{phase: phase}
	result := runner.run(context.Background())
	if owner != nil && owner.lifecycle.AdmissionOpen() {
		return 99
	}
	return result
}

type publicationDeadlineProfile struct {
	failingServerProfile
	phase string
}

func (p publicationDeadlineProfile) serve(ctx context.Context, _ http.Handler, options httpruntime.Options) error {
	if p.phase == "blocked_listener" {
		select {}
	}
	if p.phase == "listener" {
		<-ctx.Done()
		return nil
	}
	if err := options.OnReady(); err != nil {
		return &httpruntime.StartupError{Err: err}
	}
	<-ctx.Done()
	return nil
}
