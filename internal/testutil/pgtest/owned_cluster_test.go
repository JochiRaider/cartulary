package pgtest

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
	"github.com/jackc/pgx/v5/pgconn"
	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"testing/synctest"
	"time"
)

func TestPostgresSetupEnvelopeNormalizesReadinessAndCancellation(t *testing.T) {
	readiness := postgresSetupEnvelope("start", &postgresReadinessError{
		Attempts:        17,
		LastErr:         context.DeadlineExceeded,
		DeadlineExpired: true,
	})
	if readiness.FailureClass != "infra" || readiness.FailureReason != "service_readiness_timeout" ||
		readiness.Service != "postgres" || readiness.ReadinessStage != "start" || readiness.AttemptCount != 17 {
		t.Fatalf("unexpected postgres readiness envelope: %#v", readiness)
	}

	cancelled := postgresSetupEnvelope("start", context.Canceled)
	if cancelled.FailureClass != "interrupted" || cancelled.FailureReason != "cancelled_or_interrupted" {
		t.Fatalf("unexpected postgres cancellation envelope: %#v", cancelled)
	}
}

func TestPostgresContainerWaitStrategyOnlyWaitsForPortMapping(t *testing.T) {
	strategy := postgresPortWaitStrategy()
	if got := strategy.String(); !strings.Contains(got, "to be mapped") {
		t.Fatalf("expected mapped-port-only wait strategy, got %q", got)
	}
	if timeout := strategy.Timeout(); timeout == nil || *timeout != postgresPortMappingTimeout {
		t.Fatalf("unexpected port mapping timeout: got %v want %v", timeout, postgresPortMappingTimeout)
	}
}

func TestOwnedPostgresAppliesContainerLabels(t *testing.T) {
	startup := stubOwnedPostgresStartup(t)
	initializations := 0
	ready := false
	startup.initialize = func(context.Context, string) (string, error) {
		if !ready {
			t.Fatal("prerequisites initialized before readiness")
		}
		initializations++
		return "12345", nil
	}

	var gotRequest testcontainers.ContainerRequest
	startup.start = func(ctx context.Context, req testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
		deadline, bounded := ctx.Deadline()
		if remaining := time.Until(deadline); !bounded || remaining <= 30*time.Second || remaining > 35*time.Second {
			t.Fatalf("direct acquisition lost its attempt cap: bounded=%v remaining=%v", bounded, remaining)
		}
		gotRequest = req.ContainerRequest
		return fakePostgresContainer{
			host: "127.0.0.1",
			port: network.MustParsePort("5432/tcp"),
		}, nil
	}
	startup.wait = func(context.Context, *OwnedCluster) error { ready = true; return nil }

	labels := map[string]string{"cartulary.test": "suite"}
	harness, err := startOwned(context.Background(), StartOptions{Labels: labels}, startup)
	if err != nil {
		t.Fatalf("start labeled postgres: %v", err)
	}
	t.Cleanup(func() { _ = harness.Close(context.Background()) })
	if initializations != 1 || harness.ProvisioningGeneration() != "12345" {
		t.Fatalf("unpublished or repeated provisioning: count=%d generation=%q", initializations, harness.ProvisioningGeneration())
	}
	if gotRequest.Labels["cartulary.test"] != "suite" {
		t.Fatalf("container labels not applied: %#v", gotRequest.Labels)
	}
	if gotRequest.Image != postgresImage || ContainerImage() != postgresImage {
		t.Fatalf("postgres image mismatch: request=%q exported=%q want=%q", gotRequest.Image, ContainerImage(), postgresImage)
	}
	if gotRequest.Env["PGDATA"] != "/var/lib/postgresql/18/docker" ||
		gotRequest.Env["POSTGRES_INITDB_ARGS"] != "--data-checksums --auth-host=reject" {
		t.Fatalf("postgres initialization environment mismatch: %#v", gotRequest.Env)
	}
	if !strings.Contains(strings.Join(gotRequest.Cmd, " "), "ssl_min_protocol_version=TLSv1.3") ||
		!strings.Contains(strings.Join(gotRequest.Cmd, " "), "hba_file=") {
		t.Fatal("fixture omitted TLS and certificate-only HBA")
	}
	for _, purpose := range []postgres.Purpose{postgres.PurposeRuntime, postgres.PurposeMigration, postgres.PurposeRecovery} {
		dsn, err := (&TestDatabase{DSN: harness.Fixture().dsnFor("fixture")}).DSNForPurpose(purpose)
		if err != nil {
			t.Fatal(err)
		}
		settings, err := postgres.ResolveSettings(postgres.Binding{BindingKind: "managed_service", ServiceRef: "primary"}, purpose, map[string]string{mustPurposeKey(t, purpose): dsn})
		if err != nil || settings.DSN != dsn {
			t.Fatalf("fixture purpose admission: %v", err)
		}
	}
}

func mustPurposeKey(t *testing.T, purpose postgres.Purpose) string {
	t.Helper()
	key, err := postgres.EnvKeyForServiceRef("primary", purpose)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func TestOwnedPostgresReadinessAndInitializationFailuresAreTerminal(t *testing.T) {
	t.Run("nil_context_rejected_before_preflight", func(t *testing.T) {
		startup := stubOwnedPostgresStartup(t)
		startup.preflight = func(context.Context) (string, error) { t.Fatal("nil context reached preflight"); return "", nil }
		startup.start = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
			t.Fatal("nil context reached provider")
			return nil, nil
		}
		//lint:ignore SA1012 This contract test intentionally exercises the public nil-context guard before acquisition.
		if _, err := startOwned(nil, StartOptions{}, startup); err == nil {
			t.Fatal("nil acquisition context accepted")
		}
	})
	t.Run("owner_caps_and_earlier_caller_bound_every_attempt", func(t *testing.T) {
		for _, earlier := range []bool{false, true} {
			t.Run(fmt.Sprint(earlier), func(t *testing.T) {
				startup := stubOwnedPostgresStartup(t)
				ctx := context.Background()
				var callerDeadline time.Time
				if earlier {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
					defer cancel()
					callerDeadline, _ = ctx.Deadline()
				}
				attempts, sleeps := 0, 0
				var outerDeadline time.Time
				startup.sleep = func(ctx context.Context, delay time.Duration) error {
					sleeps++
					deadline, bounded := ctx.Deadline()
					if !bounded || delay != 500*time.Millisecond {
						t.Fatal("retry lost its outer deadline or backoff", bounded, delay)
					}
					if earlier {
						if !deadline.Equal(callerDeadline) {
							t.Fatal("outer cap extended caller deadline")
						}
					} else if remaining := time.Until(deadline); remaining <= 175*time.Second || remaining > 180*time.Second {
						t.Fatal("wrong outer cap", remaining)
					}
					if !outerDeadline.IsZero() && !deadline.Equal(outerDeadline) {
						t.Fatal("retry renewed outer deadline")
					}
					outerDeadline = deadline
					return nil
				}
				startup.start = func(ctx context.Context, _ testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
					attempts++
					deadline, bounded := ctx.Deadline()
					if !bounded {
						t.Fatal("unbounded startup attempt")
					}
					if earlier {
						if !deadline.Equal(callerDeadline) {
							t.Fatal("attempt extended caller deadline")
						}
					} else if remaining := time.Until(deadline); remaining <= 30*time.Second || remaining > 35*time.Second {
						t.Fatal("wrong attempt cap", remaining)
					}
					return fakePostgresContainer{}, errors.New("mapped port 5432: context deadline exceeded")
				}
				if _, err := startOwned(ctx, StartOptions{}, startup); err == nil || attempts != 3 || sleeps != 2 {
					t.Fatal("startup retry ownership changed", err, attempts, sleeps)
				}
			})
		}
	})
	t.Run("endpoint_evidence", testEndpointDiagnostic)
	t.Run("cancel_drains_create_before_settlement", func(t *testing.T) {
		startup := stubOwnedPostgresStartup(t)
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		deadline, _ := ctx.Deadline()
		acquired, settled := 0, 0
		startup.start = func(create context.Context, _ testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
			cancel()
			if create.Err() != nil {
				t.Fatal("cancel discarded in-flight create")
			}
			got, _ := create.Deadline()
			if !got.Equal(deadline) {
				t.Fatal("create deadline changed")
			}
			return fakePostgresContainer{host: "127.0.0.1", port: network.MustParsePort("5432/tcp")}, nil
		}
		startup.wait = func(context.Context, *OwnedCluster) error { t.Fatal("cancelled acquisition became ready"); return nil }
		_, err := startOwned(ctx, StartOptions{BeforeAcquire: func(*OwnedCluster) error { return nil }, Acquired: func(*OwnedCluster) error { acquired++; return nil }, Settled: func(*OwnedCluster) error { settled++; return nil }}, startup)
		if !errors.Is(err, context.Canceled) || acquired != 1 || settled != 1 {
			t.Fatal("late acquisition lost", err, acquired, settled)
		}
	})
	t.Run("timestamped_bootstrap_and_final_server_events_remain_distinct", func(t *testing.T) {
		d := newStartupDiagnostic(t.Context())
		started := "2026-10-20T00:00:00Z"
		d.ContainerStartedAt = &started
		d.StartupLogStatus = "complete"
		d.observeStartupLogEvents(strings.Join([]string{
			"2026-10-20T00:00:01Z The files belonging to this database system will be owned by user secret",
			"2026-10-20T00:00:09Z Success. You can now start the database server",
			"2026-10-20T00:00:10Z database system is ready to accept connections",
			"2026-10-20T00:00:11Z database system is shut down",
			"2026-10-20T00:00:15Z PostgreSQL init process complete; ready for start up.",
			"2026-10-20T00:00:17Z listening on IPv4 address secret-endpoint",
			"2026-10-20T00:00:18Z database system is ready to accept connections",
		}, "\n"))
		if len(d.StartupLogEvents) != 7 {
			t.Fatal(d.StartupLogEvents)
		}
		for i, ms := range []int64{1000, 9000, 10000, 11000, 15000, 17000, 18000} {
			event := d.StartupLogEvents[i]
			phase := "bootstrap"
			if i >= 4 {
				phase = "final_server"
			}
			if event.Phase != phase || event.Timestamp == nil || event.ContainerElapsedMS == nil || *event.ContainerElapsedMS != ms {
				t.Fatal("phase timing lost", event)
			}
		}
		if d.StartupLogEvents[2].Signal != "server_ready" || d.StartupLogEvents[6].Signal != "server_ready" {
			t.Fatal("ready events collapsed")
		}
		encoded, _ := json.Marshal(d)
		if bytes.Contains(encoded, []byte("secret")) {
			t.Fatal("raw log content retained")
		}
	})
	t.Run("missing_or_regressing_clocks_and_bounds_never_fabricate_phase_timing", func(t *testing.T) {
		d := newStartupDiagnostic(t.Context())
		started := "2026-10-20T00:00:00Z"
		d.ContainerStartedAt = &started
		d.observeStartupLogEvents("database system is ready to accept connections\n2026-10-20T00:00:10Z Skipping initialization\n2026-10-20T00:00:09Z database system is ready to accept connections\n")
		if d.StartupLogEvents[0].Phase != "unknown" || d.StartupLogEvents[0].Timestamp != nil || d.StartupLogEvents[0].ContainerElapsedMS != nil || d.StartupLogEvents[2].ContainerElapsedMS != nil {
			t.Fatal(d.StartupLogEvents)
		}
		d.observeStartupLogEvents(strings.Repeat("FATAL: secret\n", 129))
		if len(d.StartupLogEvents) != 128 || d.StartupLogStatus != "truncated" {
			t.Fatal("event cap lost")
		}
		for range 129 {
			d.recordProbeFailure(syscall.ECONNREFUSED)
		}
		if len(d.ProbeObservations) != 128 || !d.ProbeObservationsTruncated || d.ProbeFailures["connection_refused"] != 129 {
			t.Fatal("probe cap changed counts")
		}
	})
	t.Run("Docker_frames_are_bounded_before_provider_length_allocation", func(t *testing.T) {
		frame := func(stream byte, length uint32, content string) []byte {
			header := make([]byte, 8)
			header[0] = stream
			binary.BigEndian.PutUint32(header[4:], length)
			return append(header, []byte(content)...)
		}
		data, truncated, err := readPostgresLogFrames(bytes.NewReader(append(frame(1, 4, "one\n"), frame(2, 4, "two\n")...)))
		if err != nil || truncated || string(data) != "one\ntwo\n" {
			t.Fatal(string(data), truncated, err)
		}
		data, truncated, err = readPostgresLogFrames(bytes.NewReader(frame(1, ^uint32(0), strings.Repeat("x", 65536))))
		if err != nil || !truncated || len(data) != 65536 {
			t.Fatal("unbounded frame", len(data), truncated, err)
		}
		if _, _, err := readPostgresLogFrames(bytes.NewReader(frame(3, 0, ""))); err == nil {
			t.Fatal("invalid stream accepted")
		}
		if _, _, err := readPostgresLogFrames(bytes.NewReader(frame(1, 5, "x"))); err == nil {
			t.Fatal("incomplete frame accepted")
		}
	})
	t.Run("probe_cause_and_container_evidence_survive_deadline_and_cleanup", func(t *testing.T) {
		startup := stubOwnedPostgresStartup(t)
		startup.ping = func(context.Context, string, func(context.Context)) error {
			return &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED}
		}
		startup.wait = func(ctx context.Context, h *OwnedCluster) error {
			probeCtx, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
			defer cancel()
			return h.WaitReady(probeCtx)
		}
		terminated, inspected, logged := false, false, false
		startup.start = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
			return fakePostgresContainer{host: "127.0.0.1", port: network.MustParsePort("5432/tcp"),
				state: func(ctx context.Context) (*dockercontainer.State, error) {
					deadline, ok := ctx.Deadline()
					if terminated || ctx.Err() != nil || !ok || time.Until(deadline) > 2*time.Second {
						t.Fatal("inspection lost independent budget or ran after termination")
					}
					inspected = true
					return &dockercontainer.State{Status: dockercontainer.StateExited, ExitCode: 1, OOMKilled: false}, nil
				},
				logs: func(context.Context) (io.ReadCloser, error) {
					if terminated {
						t.Fatal("logs read after termination")
					}
					logged = true
					return io.NopCloser(strings.NewReader("initdb: No space left on device /private/secret-path\nsecret-password\n")), nil
				},
				terminate: func(context.Context) error { terminated = true; return nil },
			}, nil
		}
		h, err := startOwned(t.Context(), StartOptions{}, startup)
		d, ok := ReadinessDiagnosticFromError(err)
		if h != nil || !ok || !terminated || !inspected || !logged || d.Outcome != "deadline_expired" || d.ProbeCount != 1 || d.ProbeFailures["connection_refused"] != 1 || d.LastProbeFailure != "connection_refused" || d.CleanupOutcome != "completed" || d.ContainerState != "exited" || d.ContainerExitCode == nil || *d.ContainerExitCode != 1 || d.ContainerOOMKilled == nil || *d.ContainerOOMKilled || d.StartupLogStatus != "complete" || !reflect.DeepEqual(d.StartupLogSignals, []string{"storage_exhausted"}) {
			t.Fatal("causal observation lost", d, err)
		}
		lastPhase := d.Phases[len(d.Phases)-1]
		if lastPhase.Stage != "sql_connectivity" || lastPhase.ConfiguredMS != 0 || lastPhase.LimitingDeadline != "parent" || lastPhase.EffectiveMS <= 0 || lastPhase.EffectiveMS > 35000 {
			t.Fatal("readiness budget attribution is not owned by the attempt", lastPhase)
		}
		encoded, err := json.Marshal(d)
		if err != nil || strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "private") {
			t.Fatal("sensitive diagnostic", err)
		}
		value, err := harnessschema.Decode(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if err := harnessschema.Validate("cartulary.postgres_readiness_diagnostic.v6", value); err != nil {
			t.Fatal(err)
		}
		object := value.(map[string]any)
		object["schema_id"] = "cartulary.postgres_readiness_diagnostic.v4"
		if err := harnessschema.Validate("cartulary.postgres_readiness_diagnostic.v6", object); err == nil {
			t.Fatal("historical readiness diagnostic accepted as current")
		}
		object["schema_id"] = "cartulary.postgres_readiness_diagnostic.v6"
		object["sql_window_ms"] = 15000
		if err := harnessschema.Validate("cartulary.postgres_readiness_diagnostic.v6", object); err == nil {
			t.Fatal("retired independent readiness window accepted")
		}
		delete(object, "sql_window_ms")
		phases := object["phases"].([]any)
		phase := phases[len(phases)-1].(map[string]any)
		phase["configured_ms"] = float64(15000)
		if err := harnessschema.Validate("cartulary.postgres_readiness_diagnostic.v6", object); err == nil {
			t.Fatal("contradictory independent SQL phase cap accepted")
		}
		phase["configured_ms"] = float64(0)
		phase["limiting_deadline"] = "phase"
		if err := harnessschema.Validate("cartulary.postgres_readiness_diagnostic.v6", object); err == nil {
			t.Fatal("SQL phase reported a non-parent deadline")
		}
	})
	t.Run("probe_classification_keeps_safe_causes_only", func(t *testing.T) {
		for _, tc := range []struct {
			err          error
			cause, state string
		}{
			{&pgconn.PgError{Code: "57P03", Message: "secret-host"}, "server_starting", "57P03"},
			{&pgconn.PgError{Code: "53300", Detail: "secret-user"}, "connection_limit", "53300"},
			{&pgconn.PgError{Code: "secret"}, "server_error", ""},
			{x509.UnknownAuthorityError{}, "tls_rejected", ""},
			{&os.PathError{Op: "open", Path: "/secret/key", Err: os.ErrNotExist}, "local_configuration", ""},
			{syscall.ECONNRESET, "connection_reset", ""},
			{io.EOF, "connection_closed", ""},
			{context.DeadlineExceeded, "probe_timeout", ""},
			{errors.New("secret-raw-provider-error"), "unknown", ""},
		} {
			d := newStartupDiagnostic(t.Context())
			d.recordProbeFailure(fmt.Errorf("secret-envelope: %w", tc.err))
			if d.LastProbeFailure != tc.cause || d.LastSQLState != tc.state || d.ProbeFailures[tc.cause] != 1 {
				t.Fatal("incorrect normalized cause", d)
			}
		}
	})
	t.Run("observation_is_bounded_and_cannot_replace_primary_failure", func(t *testing.T) {
		startup := stubOwnedPostgresStartup(t)
		cleanupErr := errors.New("cleanup receipt")
		startup.start = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
			return fakePostgresContainer{
				state: func(context.Context) (*dockercontainer.State, error) {
					return nil, errors.New("secret-inspect-failure")
				},
				logs: func(context.Context) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(strings.Repeat("x", 65536) + "FATAL: secret\n")), nil
				},
				terminate: func(context.Context) error { return cleanupErr },
			}, nil
		}
		startup.wait = func(context.Context, *OwnedCluster) error { return context.DeadlineExceeded }
		h, err := startOwned(t.Context(), StartOptions{}, startup)
		d, ok := ReadinessDiagnosticFromError(err)
		if h == nil || !ok || !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, cleanupErr) || d.StartupLogStatus != "truncated" || len(d.StartupLogSignals) != 0 || d.ContainerState != "unavailable" || d.CleanupOutcome != "unresolved" {
			t.Fatal("diagnostic changed failure or exceeded bound", d, err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(h.tlsDirectory) })
	})
	t.Run("capability_validation_has_a_distinct_terminal_boundary", func(t *testing.T) {
		startup := stubOwnedPostgresStartup(t)
		startup.ping = func(ctx context.Context, _ string, connected func(context.Context)) error {
			connected(ctx)
			return errors.New(postgres.ReasonDataChecksumsDisabled)
		}
		startup.wait = func(ctx context.Context, h *OwnedCluster) error { return h.WaitReady(ctx) }
		starts := 0
		startup.start = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
			starts++
			return fakePostgresContainer{host: "127.0.0.1", port: network.MustParsePort("5432/tcp")}, nil
		}
		h, err := startOwned(t.Context(), StartOptions{}, startup)
		diagnostic, ok := ReadinessDiagnosticFromError(err)
		if h != nil || !ok || starts != 1 || diagnostic.Stage != "security_capability" || diagnostic.Outcome != "capability_rejected" || diagnostic.ProbeCount != 1 {
			t.Fatal("capability rejection lost or retried", starts, diagnostic, err)
		}
		last := diagnostic.Phases[len(diagnostic.Phases)-1]
		if last.ConfiguredMS != 5000 || last.EffectiveMS > 5000 || last.EffectiveMS <= 0 {
			t.Fatal("validation exceeded the SQL probe deadline", last)
		}
	})
	t.Run("authority_publication_precedes_start_and_failure_settles", func(t *testing.T) {
		startup := stubOwnedPostgresStartup(t)
		acquired, starts, terminations := false, 0, 0
		publicationErr := errors.New("injected ownership publication failure")
		startup.start = func(_ context.Context, req testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
			if req.Started {
				t.Fatal("container started before cleanup authority publication")
			}
			return fakePostgresContainer{
				start:     func(context.Context) error { starts++; return nil },
				terminate: func(context.Context) error { terminations++; return nil },
			}, nil
		}
		harness, err := startOwned(t.Context(), StartOptions{Acquired: func(h *OwnedCluster) error {
			acquired = h.Container != nil && len(h.CleanupFiles()) != 0
			if h.Fixture() != nil {
				t.Fatal("acquisition published fixture before readiness")
			}
			if db, err := h.PrepareTemplate(t.Context(), ""); db != nil || err == nil {
				t.Fatal("unready owner prepared a template")
			}
			return publicationErr
		}}, startup)
		if harness != nil || !errors.Is(err, publicationErr) || !acquired || starts != 0 || terminations != 1 {
			t.Fatal("failed publication lost ownership or started the container", starts, terminations, err)
		}
	})
	t.Run("failed_start_reserves_the_cleanup_budget_for_exact_removal", func(t *testing.T) {
		startup := stubOwnedPostgresStartup(t)
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		primary := errors.New("injected readiness failure")
		startup.wait = func(context.Context, *OwnedCluster) error { cancel(); return primary }
		terminations, settlements := 0, 0
		startup.start = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
			return fakePostgresContainer{
				host: "127.0.0.1", port: network.MustParsePort("5432/tcp"),
				terminateOptions: func(ctx context.Context, options []testcontainers.TerminateOption) error {
					terminations++
					deadline, bounded := ctx.Deadline()
					stop := testcontainers.NewTerminateOptions(ctx, options...).StopTimeout()
					if ctx.Err() != nil || !bounded || time.Until(deadline) <= 0 || time.Until(deadline) > 10*time.Second || stop == nil || *stop != 0 {
						t.Fatal("failed start spent the independent removal budget on graceful stop")
					}
					return nil
				},
			}, nil
		}
		h, err := startOwned(ctx, StartOptions{Settled: func(*OwnedCluster) error { settlements++; return nil }}, startup)
		if h != nil || !errors.Is(err, primary) || terminations != 1 || settlements != 1 {
			t.Fatal("failed-start ownership or primary outcome lost", h, err, terminations, settlements)
		}
	})
	t.Run("published_harness_retains_ordinary_graceful_shutdown", func(t *testing.T) {
		startup := stubOwnedPostgresStartup(t)
		startup.wait = func(context.Context, *OwnedCluster) error { return nil }
		terminations := 0
		startup.start = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
			return fakePostgresContainer{
				host: "127.0.0.1", port: network.MustParsePort("5432/tcp"),
				terminateOptions: func(ctx context.Context, options []testcontainers.TerminateOption) error {
					terminations++
					stop := testcontainers.NewTerminateOptions(ctx, options...).StopTimeout()
					if stop == nil || *stop != 10*time.Second {
						t.Fatal("ordinary shutdown was changed by failed-start policy")
					}
					return nil
				},
			}, nil
		}
		h, err := startOwned(t.Context(), StartOptions{}, startup)
		if err != nil {
			t.Fatal(err)
		}
		for range 2 {
			if err := h.Close(t.Context()); err != nil {
				t.Fatal(err)
			}
		}
		if terminations != 1 {
			t.Fatal("ordinary cleanup repeated", terminations)
		}
	})
	t.Run("partial_container_cleanup_failure_preserves_one_attempt", func(t *testing.T) {
		startup := stubOwnedPostgresStartup(t)
		starts, terminations := 0, 0
		cleanupErr := errors.New("injected cleanup failure")
		var directory string
		startup.start = func(_ context.Context, req testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
			starts++
			directory = filepath.Dir(req.Files[0].HostFilePath)
			return fakePostgresContainer{terminate: func(context.Context) error { terminations++; return cleanupErr }}, errors.New("connection reset by peer")
		}
		harness, err := startOwned(t.Context(), StartOptions{}, startup)
		if harness == nil || !errors.Is(err, cleanupErr) || starts != 1 || terminations != 1 {
			t.Fatal("partial ownership discarded or retried", starts, terminations, err)
		}
		for range 2 {
			if !errors.Is(harness.Close(t.Context()), cleanupErr) {
				t.Fatal("cleanup receipt lost")
			}
		}
		if terminations != 1 {
			t.Fatal("ordinary cleanup repeated", terminations)
		}
		diagnostic, ok := ReadinessDiagnosticFromError(err)
		if !ok || diagnostic.Stage != "container_start" || diagnostic.CleanupOutcome != "unresolved" || strings.Contains(err.Error(), "connection reset") {
			t.Fatal("unbounded or missing diagnostic", diagnostic, err)
		}
		if _, err := os.Stat(directory); err != nil {
			t.Fatal("unsettled credentials removed", err)
		}
		// The fake container has no physical resources; only its test keys remain.
		t.Cleanup(func() { _ = os.RemoveAll(directory) })
	})
	for _, stage := range []string{"readiness", "initialization"} {
		t.Run(stage, func(t *testing.T) {
			startup := stubOwnedPostgresStartup(t)
			starts, terminations, initializations := 0, 0, 0
			var directory string
			injected := errors.New("injected generation failure at 5432: context deadline exceeded")
			startup.start = func(_ context.Context, req testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
				starts++
				directory = filepath.Dir(req.Files[0].HostFilePath)
				return fakePostgresContainer{host: "127.0.0.1", port: network.MustParsePort("5432/tcp"), terminate: func(context.Context) error { terminations++; return nil }}, nil
			}
			startup.wait = func(context.Context, *OwnedCluster) error {
				if stage == "readiness" {
					return &postgresReadinessError{LastErr: injected, DeadlineExpired: true}
				}
				return nil
			}
			startup.initialize = func(context.Context, string) (string, error) { initializations++; return "", injected }
			harness, err := startOwned(t.Context(), StartOptions{}, startup)
			if harness != nil || err == nil || starts != 1 || terminations != 1 || (stage == "initialization" && initializations != 1) || (stage == "readiness" && initializations != 0) {
				t.Fatalf("failed generation was published or retried: harness=%v error=%v starts=%d terminations=%d initializations=%d", harness, err, starts, terminations, initializations)
			}
			if _, err := os.Stat(directory); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("failed generation retained keys", err)
			}
			diagnostic, ok := ReadinessDiagnosticFromError(err)
			if !ok || diagnostic.CleanupOutcome != "completed" || diagnostic.SchemaID != "cartulary.postgres_readiness_diagnostic.v6" || diagnostic.PollMS != 250 || diagnostic.ProbeCapMS != 5000 {
				t.Fatal("readiness evidence missing", diagnostic)
			}
		})
	}
}

func TestOwnedPostgresDoesNotRetryAuthenticationReadinessFailure(t *testing.T) {
	startup := stubOwnedPostgresStartup(t)

	starts := 0
	terminations := 0
	startup.start = func(ctx context.Context, req testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
		starts++
		return fakePostgresContainer{
			host: "127.0.0.1",
			port: network.MustParsePort("5432/tcp"),
			terminate: func(context.Context) error {
				terminations++
				return nil
			},
		}, nil
	}
	startup.wait = func(ctx context.Context, harness *OwnedCluster) error {
		return &postgresReadinessError{
			LastErr:         errors.New("password authentication failed for user cartulary"),
			DeadlineExpired: true,
		}
	}

	_, err := startOwned(context.Background(), StartOptions{}, startup)
	if err == nil {
		t.Fatal("expected authentication readiness failure")
	}
	if starts != 1 {
		t.Fatalf("expected no retry for auth failure, got %d starts", starts)
	}
	if terminations != 1 {
		t.Fatalf("expected failed auth attempt to terminate its container once, got %d", terminations)
	}

	var startFailure *testcontainersx.StartFailure
	if !errors.As(err, &startFailure) {
		t.Fatalf("expected StartFailure, got %T", err)
	}
	if startFailure.Retryable {
		t.Fatal("auth readiness failure must not be retryable")
	}
	if startFailure.AttemptsStarted != 1 || startFailure.MaxAttempts != 3 {
		t.Fatalf("unexpected attempts: got %d/%d", startFailure.AttemptsStarted, startFailure.MaxAttempts)
	}
}

func TestPostgresWaitReadyRespectsContextCancellation(t *testing.T) {
	t.Run("readiness_uses_the_bounded_parent_after_slow_bootstrap", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			startup := defaultStartupOperations()
			started := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()
			calls := 0
			startup.ping = func(probe context.Context, _ string, _ func(context.Context)) error {
				calls++
				deadline, ok := probe.Deadline()
				if !ok || time.Until(deadline) > 5*time.Second || deadline.After(started.Add(35*time.Second)) {
					t.Fatal("probe exceeded the existing attempt or probe cap")
				}
				if time.Since(started) < 16*time.Second {
					return syscall.ECONNRESET
				}
				return nil
			}
			if err := (&OwnedCluster{fixture: &Harness{}, operations: startup}).WaitReady(ctx); err != nil || time.Since(started) != 16*time.Second || calls < 2 {
				t.Fatal("mapped endpoint incorrectly restarted the readiness deadline", err, time.Since(started), calls)
			}
		})
	})
	t.Run("parent_expiry_ends_readiness_without_a_new_window", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			startup := defaultStartupOperations()
			startup.ping = func(context.Context, string, func(context.Context)) error { return syscall.ECONNRESET }
			started := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()
			err := (&OwnedCluster{fixture: &Harness{}, operations: startup}).WaitReady(ctx)
			var readiness *postgresReadinessError
			if !errors.As(err, &readiness) || !readiness.DeadlineExpired || !errors.Is(err, context.DeadlineExceeded) || time.Since(started) != 35*time.Second {
				t.Fatal("readiness lost its owning deadline", err, time.Since(started))
			}
		})
	})
	t.Run("unbounded_or_missing_parent_cannot_launch_a_probe", func(t *testing.T) {
		startup := defaultStartupOperations()
		startup.ping = func(context.Context, string, func(context.Context)) error {
			t.Fatal("unbounded probe launched")
			return nil
		}
		for _, ctx := range []context.Context{nil, context.Background()} {
			err := (&OwnedCluster{fixture: &Harness{}, operations: startup}).WaitReady(ctx)
			var readiness *postgresReadinessError
			if !errors.As(err, &readiness) || readiness.Attempts != 0 || readiness.DeadlineExpired {
				t.Fatal("invalid parent was not rejected before execution", err)
			}
		}
	})
	t.Run("success_caps_each_probe", func(t *testing.T) {
		startup := defaultStartupOperations()
		calls := 0
		startup.ping = func(ctx context.Context, _ string, _ func(context.Context)) error {
			calls++
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > 5*time.Second {
				t.Fatal("probe was not capped at five seconds")
			}
			return nil
		}
		ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
		defer cancel()
		if err := (&OwnedCluster{fixture: &Harness{}, operations: startup}).WaitReady(ctx); err != nil || calls != 1 {
			t.Fatal("ready probe repeated or rejected", calls, err)
		}
	})
	t.Run("shorter_parent_deadline_bounds_first_probe_and_prevents_another", func(t *testing.T) {
		startup := defaultStartupOperations()
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		parentDeadline, _ := ctx.Deadline()
		calls := 0
		startup.ping = func(probe context.Context, _ string, _ func(context.Context)) error {
			calls++
			deadline, ok := probe.Deadline()
			if !ok || deadline.After(parentDeadline) {
				t.Fatal("probe outlived its parent")
			}
			<-probe.Done()
			return probe.Err()
		}
		err := (&OwnedCluster{fixture: &Harness{}, operations: startup}).WaitReady(ctx)
		var readiness *postgresReadinessError
		if calls != 1 || !errors.As(err, &readiness) || !readiness.DeadlineExpired || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("deadline lost or readiness repeated", calls, err)
		}
	})
	startup := defaultStartupOperations()
	startup.ping = func(ctx context.Context, dsn string, _ func(context.Context)) error {
		return ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := (&OwnedCluster{operations: startup, fixture: &Harness{adminDSN: "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable"}}).WaitReady(ctx)
	if err == nil {
		t.Fatal("expected context cancellation readiness error")
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("WaitReady did not respect context cancellation promptly: %v", elapsed)
	}
	var readinessErr *postgresReadinessError
	if !errors.As(err, &readinessErr) {
		t.Fatalf("expected postgresReadinessError, got %T", err)
	}
	if readinessErr.DeadlineExpired {
		t.Fatal("context cancellation must not be marked as deadline expiry")
	}
}

func stubOwnedPostgresStartup(t testing.TB) startupOperations {
	t.Helper()
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Setenv(suiteservices.SuiteIDEnv, "")
	ops := defaultStartupOperations()
	ops.initialize = func(context.Context, string) (string, error) { return "12345", nil }
	ops.preflight = func(context.Context) (string, error) { return "unix:///var/run/docker.sock", nil }
	ops.sleep = func(context.Context, time.Duration) error { return nil }
	ops.logs = func(ctx context.Context, container testcontainers.Container) ([]postgresLogChunk, bool, error) {
		logs, err := container.Logs(ctx)
		if err != nil {
			return nil, false, err
		}
		defer logs.Close()
		data, err := io.ReadAll(io.LimitReader(logs, 65537))
		truncated := len(data) > 65536
		if truncated {
			data = data[:65536]
		}
		return []postgresLogChunk{{stream: "unknown", data: data}}, truncated, err
	}
	return ops
}

type fakePostgresContainer struct {
	testcontainers.Container
	host             string
	port             network.Port
	start            func(context.Context) error
	terminate        func(context.Context) error
	terminateOptions func(context.Context, []testcontainers.TerminateOption) error
	state            func(context.Context) (*dockercontainer.State, error)
	logs             func(context.Context) (io.ReadCloser, error)
	inspect          func(context.Context) (*dockercontainer.InspectResponse, error)
}

func (c fakePostgresContainer) GetContainerID() string { return "fixture-container" }
func (c fakePostgresContainer) Inspect(ctx context.Context) (*dockercontainer.InspectResponse, error) {
	if c.inspect != nil {
		return c.inspect(ctx)
	}
	return nil, errors.New("inspection unavailable")
}

func (c fakePostgresContainer) State(ctx context.Context) (*dockercontainer.State, error) {
	if c.state != nil {
		return c.state(ctx)
	}
	return nil, errors.New("state not observed")
}

func (c fakePostgresContainer) Logs(ctx context.Context) (io.ReadCloser, error) {
	if c.logs != nil {
		return c.logs(ctx)
	}
	return nil, errors.New("logs not observed")
}

func (c fakePostgresContainer) Host(context.Context) (string, error) {
	return c.host, nil
}

func (c fakePostgresContainer) Start(ctx context.Context) error {
	if c.start != nil {
		return c.start(ctx)
	}
	return nil
}

func (c fakePostgresContainer) MappedPort(context.Context, string) (network.Port, error) {
	return c.port, nil
}

func (c fakePostgresContainer) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	if c.terminateOptions != nil {
		return c.terminateOptions(ctx, opts)
	}
	if c.terminate == nil {
		return nil
	}
	return c.terminate(ctx)
}
