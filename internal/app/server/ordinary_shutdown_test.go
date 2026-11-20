package server

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
)

// stoppingServerProfile publishes, requests an operator stop, and then returns
// the configured drain outcome, as httpruntime does after Shutdown.
type stoppingServerProfile struct {
	failingServerProfile
	stop       context.CancelFunc
	drainErr   error
	onShutdown chan httpruntime.Options
}

// serve publishes, requests an operator stop, starts HTTP shutdown as
// httpruntime does (running OnShutdown), and returns the configured drain
// outcome.
func (profile stoppingServerProfile) serve(ctx context.Context, _ http.Handler, options httpruntime.Options) error {
	if err := options.OnReady(); err != nil {
		return &httpruntime.StartupError{Err: err}
	}
	profile.stop()
	<-ctx.Done()
	if options.OnShutdown != nil {
		options.OnShutdown()
	}
	profile.onShutdown <- options
	return profile.drainErr
}

type ordinaryShutdownObservation struct {
	exitCode       int
	stdout, stderr string
	options        httpruntime.Options
	beginDeadline  time.Time
	elapsed        time.Duration
}

func runOrdinaryShutdown(t *testing.T, drain time.Duration, drainErr error, closeRuntime func(context.Context) []string) ordinaryShutdownObservation {
	t.Helper()
	var stdout, stderr bytes.Buffer
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	onShutdown := make(chan httpruntime.Options, 1)
	var beginDeadline time.Time
	runner := newServerRunner(&stdout, &stderr)
	runner.loadConfig = serverTestConfiguration(t)
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		runtime := testRuntime(http.NotFoundHandler(), nil)
		runtime.Close = closeRuntime
		runtime.BeginShutdown = func(ctx context.Context) { beginDeadline, _ = ctx.Deadline() }
		runtime.ShutdownDrainTimeout = drain
		runtime.Fatal = func(reason string) bool {
			t.Errorf("ordinary shutdown reported fatal %q", reason)
			return true
		}
		return runtime, nil
	}
	runner.profile = stoppingServerProfile{stop: stop, drainErr: drainErr, onShutdown: onShutdown}
	started := time.Now()
	exitCode := runner.run(ctx)
	elapsed := time.Since(started)
	return ordinaryShutdownObservation{exitCode, stdout.String(), stderr.String(), <-onShutdown, beginDeadline, elapsed}
}

// TestServerRunnerOrdinaryShutdownHonoursOneDeadline covers REQ-04-146a and
// AC-604 at the runner: a stop request enters ordinary shutdown, which exits 0
// when every component closes in time and exits 1 naming the components still
// open at the one deadline. Neither case reports published_component_lost.
func TestServerRunnerOrdinaryShutdownHonoursOneDeadline(t *testing.T) {
	t.Run("complete shutdown exits zero and logs timed phases", func(t *testing.T) {
		var closeDeadline time.Time
		got := runOrdinaryShutdown(t, 8*time.Second, nil, func(ctx context.Context) []string {
			closeDeadline, _ = ctx.Deadline()
			return nil
		})
		if got.exitCode != 0 || got.stderr != "" {
			t.Fatalf("exit=%d stderr=%q, want 0 and no diagnostic", got.exitCode, got.stderr)
		}
		if got.options.OnShutdown == nil || got.beginDeadline.IsZero() {
			t.Fatal("HTTP shutdown did not begin the runtime shutdown with the application deadline")
		}
		if closeDeadline.IsZero() || !closeDeadline.Equal(got.beginDeadline) {
			t.Fatalf("component close deadline %v, want the application deadline %v", closeDeadline, got.beginDeadline)
		}
		// The HTTP drain ends at the deadline minus the close reserve:
		// min(25% of 8 s, 5 s) = 2 s.
		if got.options.ShutdownDeadline == nil {
			t.Fatal("the HTTP drain received no absolute deadline")
		}
		if reserve := got.beginDeadline.Sub(got.options.ShutdownDeadline()); reserve != 2*time.Second {
			t.Fatalf("close reserve = %s, want 2s", reserve)
		}
		for _, phase := range []string{"stop_requested", "http_drain", "components", "complete"} {
			if !strings.Contains(got.stdout, "phase="+phase) {
				t.Fatalf("missing shutdown phase %q in log: %s", phase, got.stdout)
			}
		}
		if !strings.Contains(got.stdout, "time=") {
			t.Fatalf("shutdown phase lines are not timestamped: %s", got.stdout)
		}
	})
	t.Run("work cancelled at the close reserve boundary still exits zero", func(t *testing.T) {
		got := runOrdinaryShutdown(t, time.Second, &httpruntime.DrainDeadlineError{Err: context.DeadlineExceeded}, func(context.Context) []string {
			return nil
		})
		if got.exitCode != 0 || got.stderr != "" {
			t.Fatalf("exit=%d stderr=%q, want 0: cancelled in-flight work completes the HTTP phase", got.exitCode, got.stderr)
		}
		if !strings.Contains(got.stdout, "cancelled_in_flight=true") {
			t.Fatalf("the drain phase does not record the cancellation: %s", got.stdout)
		}
	})
	t.Run("components open at the deadline exit one and are named", func(t *testing.T) {
		const drain = 200 * time.Millisecond
		got := runOrdinaryShutdown(t, drain, nil, func(ctx context.Context) []string {
			<-ctx.Done()
			return []string{"jobs_runner", "postgres"}
		})
		if got.exitCode != 1 {
			t.Fatalf("exit=%d, want 1", got.exitCode)
		}
		if want := "{\"code\":\"shutdown_incomplete\",\"open_components\":[\"jobs_runner\",\"postgres\"]}\n"; got.stderr != want {
			t.Fatalf("diagnostic=%q want %q", got.stderr, want)
		}
		if limit := drain + shutdownCloseGuard + time.Second; got.elapsed > limit {
			t.Fatalf("shutdown took %s, beyond the deadline bound %s", got.elapsed, limit)
		}
	})
	t.Run("any other stop-initiated drain failure is incomplete shutdown, not component loss", func(t *testing.T) {
		got := runOrdinaryShutdown(t, time.Second, errors.New("shutdown http server: listener close failed"), func(context.Context) []string {
			return nil
		})
		if got.exitCode != 1 {
			t.Fatalf("exit=%d, want 1", got.exitCode)
		}
		if want := "{\"code\":\"shutdown_incomplete\",\"open_components\":[\"http_server\"]}\n"; got.stderr != want {
			t.Fatalf("diagnostic=%q want %q", got.stderr, want)
		}
		if strings.Contains(got.stderr, "published_component_lost") {
			t.Fatalf("ordinary shutdown reported component loss: %q", got.stderr)
		}
	})
}

// TestShutdownClockNeverRestartsAndReservesCloseTime covers REQ-04-146 and
// REQ-04-146a: the deadline is fixed by the first stop request (the fatal
// lifecycle and repeated stops reuse it), and the close reserve is
// min(25% of the deadline, 5 s).
func TestShutdownClockNeverRestartsAndReservesCloseTime(t *testing.T) {
	clock := newShutdownClock(40 * time.Second)
	clock.start()
	first := clock.deadline()
	time.Sleep(5 * time.Millisecond)
	clock.start()
	if again := clock.deadline(); !again.Equal(first) {
		t.Fatalf("a later stop or fatal signal moved the deadline from %v to %v", first, again)
	}
	if reserve := first.Sub(clock.httpDrainDeadline()); reserve != 5*time.Second {
		t.Fatalf("40 s deadline reserve = %s, want the 5 s cap", reserve)
	}
	for drain, want := range map[time.Duration]time.Duration{time.Second: 250 * time.Millisecond, 10 * time.Second: 2500 * time.Millisecond, 30 * time.Second: 5 * time.Second} {
		if got := shutdownCloseReserve(drain); got != want {
			t.Fatalf("reserve(%s) = %s, want %s", drain, got, want)
		}
	}
}
