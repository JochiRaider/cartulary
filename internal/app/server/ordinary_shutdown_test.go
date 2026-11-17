package server

import (
	"bytes"
	"context"
	"fmt"
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
	onShutdown chan func()
}

func (profile stoppingServerProfile) serve(ctx context.Context, _ http.Handler, options httpruntime.Options) error {
	if err := options.OnReady(); err != nil {
		return &httpruntime.StartupError{Err: err}
	}
	profile.onShutdown <- options.OnShutdown
	profile.stop()
	<-ctx.Done()
	return profile.drainErr
}

func runOrdinaryShutdown(t *testing.T, drain time.Duration, drainErr error, closeRuntime func(context.Context) []string) (int, string, string, func(), time.Duration) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	onShutdown := make(chan func(), 1)
	shutdownHook := func() {}
	runner := newServerRunner(&stdout, &stderr)
	runner.loadConfig = serverTestConfiguration(t)
	runner.buildRuntime = func(context.Context, configassembly.Loaded, Options) (serverRuntime, error) {
		runtime := testRuntime(http.NotFoundHandler(), nil)
		runtime.Close = closeRuntime
		runtime.OnShutdown = shutdownHook
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
	return exitCode, stdout.String(), stderr.String(), <-onShutdown, elapsed
}

// TestServerRunnerOrdinaryShutdownHonoursOneDeadline covers REQ-04-146a and
// AC-604 at the runner: a stop request enters ordinary shutdown, which exits 0
// when every component closes in time and exits 1 naming the components still
// open at the one deadline. Neither case reports published_component_lost.
func TestServerRunnerOrdinaryShutdownHonoursOneDeadline(t *testing.T) {
	t.Run("complete shutdown exits zero and logs timed phases", func(t *testing.T) {
		var closeDeadline time.Time
		exitCode, stdout, stderr, hook, _ := runOrdinaryShutdown(t, time.Second, nil, func(ctx context.Context) []string {
			closeDeadline, _ = ctx.Deadline()
			return nil
		})
		if exitCode != 0 || stderr != "" {
			t.Fatalf("exit=%d stderr=%q, want 0 and no diagnostic", exitCode, stderr)
		}
		if hook == nil {
			t.Fatal("the WebSocket shutdown hook was not passed to the HTTP runtime")
		}
		if closeDeadline.IsZero() {
			t.Fatal("component close received no application shutdown deadline")
		}
		for _, phase := range []string{"stop_requested", "http_drain", "components", "complete"} {
			if !strings.Contains(stdout, "phase="+phase) {
				t.Fatalf("missing shutdown phase %q in log: %s", phase, stdout)
			}
		}
		if !strings.Contains(stdout, "time=") {
			t.Fatalf("shutdown phase lines are not timestamped: %s", stdout)
		}
	})
	t.Run("components open at the deadline exit one and are named", func(t *testing.T) {
		const drain = 200 * time.Millisecond
		exitCode, _, stderr, _, elapsed := runOrdinaryShutdown(t, drain, nil, func(ctx context.Context) []string {
			<-ctx.Done()
			return []string{"jobs_runner", "postgres"}
		})
		if exitCode != 1 {
			t.Fatalf("exit=%d, want 1", exitCode)
		}
		if want := "{\"code\":\"shutdown_incomplete\",\"open_components\":[\"jobs_runner\",\"postgres\"]}\n"; stderr != want {
			t.Fatalf("diagnostic=%q want %q", stderr, want)
		}
		if limit := drain + shutdownCloseGuard + time.Second; elapsed > limit {
			t.Fatalf("shutdown took %s, beyond the deadline bound %s", elapsed, limit)
		}
	})
	t.Run("a stop-initiated drain failure is incomplete shutdown, not component loss", func(t *testing.T) {
		exitCode, _, stderr, _, _ := runOrdinaryShutdown(t, time.Second, fmt.Errorf("shutdown http server: %w", context.DeadlineExceeded), func(context.Context) []string {
			return nil
		})
		if exitCode != 1 {
			t.Fatalf("exit=%d, want 1", exitCode)
		}
		if want := "{\"code\":\"shutdown_incomplete\",\"open_components\":[\"http_server\"]}\n"; stderr != want {
			t.Fatalf("diagnostic=%q want %q", stderr, want)
		}
		if strings.Contains(stderr, "published_component_lost") {
			t.Fatalf("ordinary shutdown reported component loss: %q", stderr)
		}
	})
}
