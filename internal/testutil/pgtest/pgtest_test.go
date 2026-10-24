package pgtest

import (
	"bytes"
	"context"
	"crypto/x509"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	testcontainers "github.com/testcontainers/testcontainers-go"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/testcontainersx"
	"github.com/JochiRaider/cartulary/internal/testutil/testfailure"
)

func TestStartSharedUsesAttachEnvWithoutStartingOwnedHarness(t *testing.T) {
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	resetSharedHarness(t)

	t.Setenv(suiteservices.PGAdminDSNEnv, "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable")
	t.Setenv(suiteservices.PGDSNTemplateEnv, "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable")
	t.Setenv(suiteservices.PGTemplateDBEnv, "suite_template")

	oldPing := pingAdminDSNFn
	oldInspect := inspectClusterGenerationFn
	t.Cleanup(func() { pingAdminDSNFn = oldPing; inspectClusterGenerationFn = oldInspect })
	t.Setenv(suiteservices.PGSchemaHashEnv, pgschema.MustHash())
	t.Setenv(suiteservices.PGProvisioningDigestEnv, ProvisioningDigest())
	t.Setenv(suiteservices.PGTemplateDigestEnv, TemplateDigest(pgschema.MustHash()))
	t.Setenv(suiteservices.PGGenerationEnv, "12345")
	t.Setenv(suiteservices.PGIsolationEnv, "shared")
	t.Setenv(suiteservices.PGAllocationRefEnv, "pgtest-fixture")
	inspectClusterGenerationFn = func(context.Context, string) (string, error) { return "12345", nil }

	var pingDSN string
	pingAdminDSNFn = func(ctx context.Context, dsn string, _ func(context.Context)) error {
		pingDSN = dsn
		return nil
	}

	harness, err := StartShared(context.Background())
	if err != nil {
		t.Fatalf("start attached harness: %v", err)
	}
	if !harness.attached {
		t.Fatal("expected attached harness")
	}
	if harness.Container != nil {
		t.Fatal("expected attach mode to avoid creating a testcontainer")
	}
	if harness.templateDB != "suite_template" {
		t.Fatalf("unexpected template database: got %q", harness.templateDB)
	}
	if pingDSN != harness.AdminDSN() {
		t.Fatalf("unexpected admin dsn ping: got %q want %q", pingDSN, harness.AdminDSN())
	}
	if got := harness.dsnFor("tenant_one"); !strings.Contains(got, "/tenant_one?") {
		t.Fatalf("expected attach dsn template replacement, got %q", got)
	}
	for _, key := range []string{suiteservices.PGSchemaHashEnv, suiteservices.PGProvisioningDigestEnv, suiteservices.PGTemplateDigestEnv, suiteservices.PGGenerationEnv} {
		t.Run("reject_incompatible_"+key, func(t *testing.T) {
			t.Setenv(key, "superseded")
			if attached, _, err := startAttachedHarness(t.Context()); err == nil || attached != nil {
				t.Fatal("incompatible attachment repaired or accepted")
			}
		})
	}
}

func TestStartSharedRequiresExplicitAttachment(t *testing.T) {
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	resetSharedHarness(t)
	t.Setenv(suiteservices.PGAdminDSNEnv, "")
	t.Setenv(suiteservices.PGDSNTemplateEnv, "")
	if _, err := StartShared(t.Context()); err == nil {
		t.Fatal("missing attachment started a service")
	}
}

func TestCreateDatabaseRejectsOmittedServiceBeforeAcquisition(t *testing.T) {
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "object_store")

	oldCreate := createDatabaseFn
	createCalls := 0
	createDatabaseFn = func(context.Context, string, string, string) error {
		createCalls++
		return nil
	}
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
	})

	_, err := (&Harness{}).newDatabase(context.Background(), "guarded", suiteservices.FixtureReusePerTest, fixtureAttribution{})
	var dependencyErr *suiteservices.ServiceDependencyError
	if !errors.As(err, &dependencyErr) || dependencyErr.Service != "postgres" || dependencyErr.Reason != "omitted" {
		t.Fatalf("unexpected dependency error: %#v", err)
	}
	if createCalls != 0 {
		t.Fatalf("postgres acquisition ran before the dependency guard: calls=%d", createCalls)
	}
}

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

func TestDatabaseNamesAreUniqueAcrossSimulatedProcesses(t *testing.T) {
	first := &Harness{suiteHash: "suitehash", processHash: "procaaaa"}
	second := &Harness{suiteHash: "suitehash", processHash: "procbbbb"}

	seen := make(map[string]struct{})
	validName := regexp.MustCompile(`^[a-z0-9_]+$`)
	for _, harness := range []*Harness{first, second} {
		for i := 0; i < 8; i++ {
			name := harness.nextDatabaseName("Prefix With Spaces-and-symbols!")
			if len(name) > 63 {
				t.Fatalf("database name exceeds postgres identifier limit: %q", name)
			}
			if !validName.MatchString(name) {
				t.Fatalf("database name must be sanitized, got %q", name)
			}
			if _, exists := seen[name]; exists {
				t.Fatalf("database name collision detected: %q", name)
			}
			seen[name] = struct{}{}
		}
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
	stubOwnedPostgresStartup(t)
	initializations := 0
	ready := false
	initializeClusterFn = func(context.Context, string) (string, error) {
		if !ready {
			t.Fatal("prerequisites initialized before readiness")
		}
		initializations++
		return "12345", nil
	}

	var gotRequest testcontainers.ContainerRequest
	startContainerFn = func(ctx context.Context, req testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
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
	waitReadyFn = func(context.Context, *Harness) error { ready = true; return nil }

	labels := map[string]string{"cartulary.test": "suite"}
	harness, err := StartOwnedWithLabels(context.Background(), labels)
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
		dsn, err := (&TestDatabase{DSN: harness.dsnFor("fixture")}).DSNForPurpose(purpose)
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
		stubOwnedPostgresStartup(t)
		startPreflightFn = func(context.Context) (string, error) { t.Fatal("nil context reached preflight"); return "", nil }
		startContainerFn = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
			t.Fatal("nil context reached provider")
			return nil, nil
		}
		//lint:ignore SA1012 This contract test intentionally exercises the public nil-context guard before acquisition.
		if _, err := StartOwned(nil); err == nil {
			t.Fatal("nil acquisition context accepted")
		}
	})
	t.Run("owner_caps_and_earlier_caller_bound_every_attempt", func(t *testing.T) {
		for _, earlier := range []bool{false, true} {
			t.Run(fmt.Sprint(earlier), func(t *testing.T) {
				stubOwnedPostgresStartup(t)
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
				startSleepFn = func(ctx context.Context, delay time.Duration) error {
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
				startContainerFn = func(ctx context.Context, _ testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
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
				if _, err := StartOwned(ctx); err == nil || attempts != 3 || sleeps != 2 {
					t.Fatal("startup retry ownership changed", err, attempts, sleeps)
				}
			})
		}
	})
	t.Run("endpoint_evidence", testEndpointDiagnostic)
	t.Run("cancel_drains_create_before_settlement", func(t *testing.T) {
		stubOwnedPostgresStartup(t)
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		deadline, _ := ctx.Deadline()
		acquired, settled := 0, 0
		startContainerFn = func(create context.Context, _ testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
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
		waitReadyFn = func(context.Context, *Harness) error { t.Fatal("cancelled acquisition became ready"); return nil }
		_, err := StartOwnedWithOptions(ctx, StartOptions{BeforeAcquire: func(*Harness) error { return nil }, Acquired: func(*Harness) error { acquired++; return nil }, Settled: func(*Harness) error { settled++; return nil }})
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
		stubOwnedPostgresStartup(t)
		oldPing := pingAdminDSNFn
		t.Cleanup(func() { pingAdminDSNFn = oldPing })
		pingAdminDSNFn = func(context.Context, string, func(context.Context)) error {
			return &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED}
		}
		waitReadyFn = func(ctx context.Context, h *Harness) error {
			probeCtx, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
			defer cancel()
			return h.WaitReady(probeCtx)
		}
		terminated, inspected, logged := false, false, false
		startContainerFn = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
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
		h, err := StartOwned(t.Context())
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
		stubOwnedPostgresStartup(t)
		cleanupErr := errors.New("cleanup receipt")
		startContainerFn = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
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
		waitReadyFn = func(context.Context, *Harness) error { return context.DeadlineExceeded }
		h, err := StartOwned(t.Context())
		d, ok := ReadinessDiagnosticFromError(err)
		if h == nil || !ok || !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, cleanupErr) || d.StartupLogStatus != "truncated" || len(d.StartupLogSignals) != 0 || d.ContainerState != "unavailable" || d.CleanupOutcome != "unresolved" {
			t.Fatal("diagnostic changed failure or exceeded bound", d, err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(h.tlsDirectory) })
	})
	t.Run("capability_validation_has_a_distinct_terminal_boundary", func(t *testing.T) {
		stubOwnedPostgresStartup(t)
		oldPing := pingAdminDSNFn
		t.Cleanup(func() { pingAdminDSNFn = oldPing })
		pingAdminDSNFn = func(ctx context.Context, _ string, connected func(context.Context)) error {
			connected(ctx)
			return errors.New(postgres.ReasonDataChecksumsDisabled)
		}
		waitReadyFn = func(ctx context.Context, h *Harness) error { return h.WaitReady(ctx) }
		starts := 0
		startContainerFn = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
			starts++
			return fakePostgresContainer{host: "127.0.0.1", port: network.MustParsePort("5432/tcp")}, nil
		}
		h, err := StartOwned(t.Context())
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
		stubOwnedPostgresStartup(t)
		acquired, starts, terminations := false, 0, 0
		publicationErr := errors.New("injected ownership publication failure")
		startContainerFn = func(_ context.Context, req testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
			if req.Started {
				t.Fatal("container started before cleanup authority publication")
			}
			return fakePostgresContainer{
				start:     func(context.Context) error { starts++; return nil },
				terminate: func(context.Context) error { terminations++; return nil },
			}, nil
		}
		harness, err := StartOwnedWithOptions(t.Context(), StartOptions{Acquired: func(h *Harness) error {
			acquired = h.Container != nil && len(h.CleanupFiles()) != 0
			return publicationErr
		}})
		if harness != nil || !errors.Is(err, publicationErr) || !acquired || starts != 0 || terminations != 1 {
			t.Fatal("failed publication lost ownership or started the container", starts, terminations, err)
		}
	})
	t.Run("failed_start_reserves_the_cleanup_budget_for_exact_removal", func(t *testing.T) {
		stubOwnedPostgresStartup(t)
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		primary := errors.New("injected readiness failure")
		waitReadyFn = func(context.Context, *Harness) error { cancel(); return primary }
		terminations, settlements := 0, 0
		startContainerFn = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
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
		h, err := StartOwnedWithOptions(ctx, StartOptions{Settled: func(*Harness) error { settlements++; return nil }})
		if h != nil || !errors.Is(err, primary) || terminations != 1 || settlements != 1 {
			t.Fatal("failed-start ownership or primary outcome lost", h, err, terminations, settlements)
		}
	})
	t.Run("published_harness_retains_ordinary_graceful_shutdown", func(t *testing.T) {
		stubOwnedPostgresStartup(t)
		waitReadyFn = func(context.Context, *Harness) error { return nil }
		terminations := 0
		startContainerFn = func(context.Context, testcontainers.GenericContainerRequest, testcontainersx.CreateObserver) (testcontainers.Container, error) {
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
		h, err := StartOwned(t.Context())
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
		stubOwnedPostgresStartup(t)
		starts, terminations := 0, 0
		cleanupErr := errors.New("injected cleanup failure")
		var directory string
		startContainerFn = func(_ context.Context, req testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
			starts++
			directory = filepath.Dir(req.Files[0].HostFilePath)
			return fakePostgresContainer{terminate: func(context.Context) error { terminations++; return cleanupErr }}, errors.New("connection reset by peer")
		}
		harness, err := StartOwned(t.Context())
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
			stubOwnedPostgresStartup(t)
			starts, terminations, initializations := 0, 0, 0
			var directory string
			injected := errors.New("injected generation failure at 5432: context deadline exceeded")
			startContainerFn = func(_ context.Context, req testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
				starts++
				directory = filepath.Dir(req.Files[0].HostFilePath)
				return fakePostgresContainer{host: "127.0.0.1", port: network.MustParsePort("5432/tcp"), terminate: func(context.Context) error { terminations++; return nil }}, nil
			}
			waitReadyFn = func(context.Context, *Harness) error {
				if stage == "readiness" {
					return &postgresReadinessError{LastErr: injected, DeadlineExpired: true}
				}
				return nil
			}
			initializeClusterFn = func(context.Context, string) (string, error) { initializations++; return "", injected }
			harness, err := StartOwned(t.Context())
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
	stubOwnedPostgresStartup(t)

	starts := 0
	terminations := 0
	startContainerFn = func(ctx context.Context, req testcontainers.GenericContainerRequest, _ testcontainersx.CreateObserver) (testcontainers.Container, error) {
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
	waitReadyFn = func(ctx context.Context, harness *Harness) error {
		return &postgresReadinessError{
			LastErr:         errors.New("password authentication failed for user cartulary"),
			DeadlineExpired: true,
		}
	}

	_, err := StartOwnedWithOptions(context.Background(), StartOptions{})
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
			oldPing := pingAdminDSNFn
			t.Cleanup(func() { pingAdminDSNFn = oldPing })
			started := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()
			calls := 0
			pingAdminDSNFn = func(probe context.Context, _ string, _ func(context.Context)) error {
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
			if err := (&Harness{}).WaitReady(ctx); err != nil || time.Since(started) != 16*time.Second || calls < 2 {
				t.Fatal("mapped endpoint incorrectly restarted the readiness deadline", err, time.Since(started), calls)
			}
		})
	})
	t.Run("parent_expiry_ends_readiness_without_a_new_window", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			oldPing := pingAdminDSNFn
			t.Cleanup(func() { pingAdminDSNFn = oldPing })
			pingAdminDSNFn = func(context.Context, string, func(context.Context)) error { return syscall.ECONNRESET }
			started := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()
			err := (&Harness{}).WaitReady(ctx)
			var readiness *postgresReadinessError
			if !errors.As(err, &readiness) || !readiness.DeadlineExpired || !errors.Is(err, context.DeadlineExceeded) || time.Since(started) != 35*time.Second {
				t.Fatal("readiness lost its owning deadline", err, time.Since(started))
			}
		})
	})
	t.Run("unbounded_or_missing_parent_cannot_launch_a_probe", func(t *testing.T) {
		oldPing := pingAdminDSNFn
		t.Cleanup(func() { pingAdminDSNFn = oldPing })
		pingAdminDSNFn = func(context.Context, string, func(context.Context)) error {
			t.Fatal("unbounded probe launched")
			return nil
		}
		for _, ctx := range []context.Context{nil, context.Background()} {
			err := (&Harness{}).WaitReady(ctx)
			var readiness *postgresReadinessError
			if !errors.As(err, &readiness) || readiness.Attempts != 0 || readiness.DeadlineExpired {
				t.Fatal("invalid parent was not rejected before execution", err)
			}
		}
	})
	t.Run("success_caps_each_probe", func(t *testing.T) {
		oldPing := pingAdminDSNFn
		t.Cleanup(func() { pingAdminDSNFn = oldPing })
		calls := 0
		pingAdminDSNFn = func(ctx context.Context, _ string, _ func(context.Context)) error {
			calls++
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > 5*time.Second {
				t.Fatal("probe was not capped at five seconds")
			}
			return nil
		}
		ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
		defer cancel()
		if err := (&Harness{}).WaitReady(ctx); err != nil || calls != 1 {
			t.Fatal("ready probe repeated or rejected", calls, err)
		}
	})
	t.Run("shorter_parent_deadline_bounds_first_probe_and_prevents_another", func(t *testing.T) {
		oldPing := pingAdminDSNFn
		t.Cleanup(func() { pingAdminDSNFn = oldPing })
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		parentDeadline, _ := ctx.Deadline()
		calls := 0
		pingAdminDSNFn = func(probe context.Context, _ string, _ func(context.Context)) error {
			calls++
			deadline, ok := probe.Deadline()
			if !ok || deadline.After(parentDeadline) {
				t.Fatal("probe outlived its parent")
			}
			<-probe.Done()
			return probe.Err()
		}
		err := (&Harness{}).WaitReady(ctx)
		var readiness *postgresReadinessError
		if calls != 1 || !errors.As(err, &readiness) || !readiness.DeadlineExpired || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("deadline lost or readiness repeated", calls, err)
		}
	})
	oldPing := pingAdminDSNFn
	t.Cleanup(func() {
		pingAdminDSNFn = oldPing
	})
	pingAdminDSNFn = func(ctx context.Context, dsn string, _ func(context.Context)) error {
		return ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := (&Harness{adminDSN: "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable"}).WaitReady(ctx)
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

func TestPrepareDatabaseTemplateModeClonesWithoutMigrationReplay(t *testing.T) {
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Setenv(suiteservices.SuiteIDEnv, "suite-template-mode")
	t.Setenv(suiteservices.TargetEnv, "backend-store")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "template-mode")

	oldCreate := createDatabaseFn
	oldMigrate := initializeDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		initializeDatabaseFn = oldMigrate
	})

	var createCalls []struct {
		Name     string
		Template string
	}
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		createCalls = append(createCalls, struct {
			Name     string
			Template string
		}{
			Name:     name,
			Template: templateDB,
		})
		return nil
	}

	migrateCalls := 0
	initializeDatabaseFn = func(ctx context.Context, db *sql.DB, _ string, source *database_migrations.Source) error {
		migrateCalls++
		return nil
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	testDB, err := harness.PrepareDatabase(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("prepare database from template: %v", err)
	}
	if len(createCalls) != 1 {
		t.Fatalf("expected one database create call, got %d", len(createCalls))
	}
	if createCalls[0].Template != "suite_template" {
		t.Fatalf("expected template clone create call, got template %q", createCalls[0].Template)
	}
	if migrateCalls != 0 {
		t.Fatalf("expected template mode to skip per-database migrations, got %d calls", migrateCalls)
	}
	if testDB.Name == "" || !strings.Contains(testDB.DSN, testDB.Name) {
		t.Fatalf("expected prepared database dsn to reference the cloned database, got %#v", testDB)
	}

	scope, ok, err := suiteservices.Summarize(nil)
	if err != nil {
		t.Fatalf("summarize suite service events: %v", err)
	}
	if !ok {
		t.Fatal("expected suite service summary")
	}
	if len(scope.Postgres.DatabasePreparations) != 1 {
		t.Fatalf("expected one database preparation, got %#v", scope.Postgres.DatabasePreparations)
	}
	preparation := scope.Postgres.DatabasePreparations[0]
	if preparation.Name != testDB.Name {
		t.Fatalf("unexpected prepared database name: got %q want %q", preparation.Name, testDB.Name)
	}
	if preparation.Strategy != suiteservices.PostgresPreparationTemplateClone {
		t.Fatalf("expected template-clone preparation, got %#v", preparation)
	}
	if preparation.TemplateDatabase != "suite_template" {
		t.Fatalf("unexpected template database: got %#v", preparation)
	}
	if preparation.Target != "backend-store" {
		t.Fatalf("unexpected preparation target: got %#v", preparation)
	}
}

func TestPostgresFixturePolicyResolutionUsesTopLevelTestAndPackage(t *testing.T) {
	t.Setenv(postgresFixturePolicyTestsEnv, "TestPostgresFixturePolicyResolutionUsesTopLevelTestAndPackage=template_clone")
	t.Setenv(postgresFixturePolicyPackagesEnv, "internal/modules/auth=transaction")
	t.Setenv(postgresFixturePolicyDefaultEnv, "transaction")

	policy := resolvePostgresFixturePolicy(fixtureAttribution{
		TestName:      "TestPostgresFixturePolicyResolutionUsesTopLevelTestAndPackage/subcase",
		CallerPackage: "internal/modules/auth",
	})
	if policy != postgresFixturePolicyTemplateClone {
		t.Fatalf("expected top-level test policy to win, got %q", policy)
	}
	t.Setenv(postgresFixturePolicyTestsEnv, "")
	policy = resolvePostgresFixturePolicy(fixtureAttribution{
		TestName:      "TestOther/subcase",
		CallerPackage: "./internal/modules/auth",
	})
	if policy != postgresFixturePolicyTransaction {
		t.Fatalf("expected package policy to win, got %q", policy)
	}

	t.Setenv(postgresFixturePolicyPackagesEnv, "")
	t.Setenv(postgresFixturePolicyDefaultEnv, "")
	policy = resolvePostgresFixturePolicy(fixtureAttribution{
		TestName:      "TestOther/subcase",
		CallerPackage: "internal/modules/timeline",
	})
	if policy != "" {
		t.Fatalf("expected no implicit fixture policy, got %q", policy)
	}
}

func TestValidateSelectedPostgresFixturePolicy(t *testing.T) {
	if err := validateSelectedPostgresFixturePolicy(postgresFixturePolicyTemplateClone, postgresFixturePolicyTemplateClone); err != nil {
		t.Fatalf("matching explicit policy failed: %v", err)
	}
	if err := validateSelectedPostgresFixturePolicy("", postgresFixturePolicyTemplateClone); err == nil {
		t.Fatal("expected every fixture helper to require an explicit policy")
	}
}

func TestPrepareIsolatedDatabaseTUsesFreshTemplateClones(t *testing.T) {
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Setenv(suiteservices.SuiteIDEnv, "")
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTemplateClone)

	oldCreate := createDatabaseFn
	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropOwnedDatabaseFn = oldDrop
	})

	var creates []struct {
		Name     string
		Template string
	}
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		creates = append(creates, struct {
			Name     string
			Template string
		}{Name: name, Template: templateDB})
		return nil
	}
	var drops []string
	dropOwnedDatabaseFn = func(ctx context.Context, adminDSN string, name string, _ func(postgrescleanup.DropAttempt)) error {
		drops = append(drops, name)
		return nil
	}
	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	var firstName string
	t.Run("first clone", func(t *testing.T) {
		firstName = harness.PrepareIsolatedDatabaseT(t, "clone-policy").Name
	})
	var secondName string
	t.Run("second clone", func(t *testing.T) {
		secondName = harness.PrepareIsolatedDatabaseT(t, "clone-policy").Name
	})

	if firstName == "" || secondName == "" || firstName == secondName {
		t.Fatalf("expected distinct template-clone database names, got first=%q second=%q", firstName, secondName)
	}
	if len(creates) != 2 {
		t.Fatalf("expected two database creates, got %#v", creates)
	}
	for _, create := range creates {
		if create.Template != "suite_template" {
			t.Fatalf("expected template clone create, got %#v", create)
		}
	}
	if len(drops) != 2 {
		t.Fatalf("expected cloned databases to be dropped outside active suite, got %v", drops)
	}
}

func TestPrepareIsolatedDatabaseTCleanupDropsStandaloneDatabase(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "")
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTemplateClone)

	oldCreate := createDatabaseFn
	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropOwnedDatabaseFn = oldDrop
	})

	var dropped []string
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		return nil
	}
	dropOwnedDatabaseFn = func(ctx context.Context, adminDSN string, name string, _ func(postgrescleanup.DropAttempt)) error {
		dropped = append(dropped, name)
		return nil
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	var preparedName string
	t.Run("prepare", func(t *testing.T) {
		preparedName = harness.PrepareIsolatedDatabaseT(t, "standalone-cleanup").Name
	})

	if len(dropped) != 1 || dropped[0] != preparedName {
		t.Fatalf("expected standalone PrepareIsolatedDatabaseT cleanup to drop %q, got %v", preparedName, dropped)
	}
}

func TestDatabaseCleanupUsesSingleDropOnlyForOwnedDatabase(t *testing.T) {
	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		dropOwnedDatabaseFn = oldDrop
	})

	harness := &Harness{
		adminDSN: "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		ownedDatabases: map[string]struct{}{
			"ct_deadbeef_cafe0001_000001_owned": {},
		},
	}
	dropCalls := 0
	dropOwnedDatabaseFn = func(ctx context.Context, _ string, _ string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		if ctx.Err() != nil {
			t.Fatalf("cleanup delegate received canceled parent: %v", ctx.Err())
		}
		return nil
	}

	if err := harness.dropDatabase(context.Background(), "ct_deadbeef_cafe0001_000001_owned", suiteservices.FixtureReusePerTest, fixtureAttribution{}); err != nil {
		t.Fatalf("drop owned active database: %v", err)
	}
	if dropCalls != 1 {
		t.Fatalf("expected one coordinated drop request, got %d", dropCalls)
	}

	if err := harness.dropDatabase(context.Background(), "ct_deadbeef_cafe0001_000002_unowned", suiteservices.FixtureReusePerTest, fixtureAttribution{}); err == nil {
		t.Fatal("expected unowned database cleanup to fail before deletion")
	}
	if dropCalls != 1 {
		t.Fatalf("unowned database reached deletion: calls=%d", dropCalls)
	}
}

func TestDatabaseCleanupPassesOwnerContext(t *testing.T) {
	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		dropOwnedDatabaseFn = oldDrop
	})

	const databaseName = "ct_deadbeef_cafe0001_000001_owned"
	harness := &Harness{
		adminDSN: "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		ownedDatabases: map[string]struct{}{
			databaseName: {},
		},
	}
	parent, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dropCalls := 0
	dropOwnedDatabaseFn = func(ctx context.Context, _ string, _ string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		if ctx != parent || ctx.Err() != nil {
			t.Fatalf("coordinated cleanup inherited an expired child context: %v", ctx.Err())
		}
		return nil
	}

	if err := harness.dropDatabase(parent, databaseName, suiteservices.FixtureReusePerTest, fixtureAttribution{}); err != nil {
		t.Fatalf("drop owned database: %v", err)
	}
	if dropCalls != 1 {
		t.Fatalf("coordinated cleanup calls = %d, want 1", dropCalls)
	}
}

func TestDatabaseCleanupRetainsFailedRecoveryAuthority(t *testing.T) {
	authorizePGSuite(t)
	resetSharedHarness(t)
	t.Setenv(suiteservices.SuiteIDEnv, "suite-cleanup-failure")
	t.Setenv(suiteservices.TargetEnv, "backend-integration")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "cleanup-failure")

	oldDrop := dropOwnedDatabaseFn
	oldCreate := createDatabaseFn
	t.Cleanup(func() {
		dropOwnedDatabaseFn = oldDrop
		createDatabaseFn = oldCreate
	})
	var createCalls atomic.Int32
	createDatabaseFn = func(context.Context, string, string, string) error {
		createCalls.Add(1)
		return nil
	}

	const databaseName = "ct_deadbeef_cafe0001_000001_owned"
	harness := &Harness{attached: true, templateDB: "suite_template", ownedDatabases: map[string]struct{}{databaseName: {}}}
	sharedHarness = harness
	if err := suiteservices.RecordEvent(nil, suiteservices.Event{Type: suiteservices.EventPostgresDBCreated, Name: databaseName, Kind: "template-clone"}); err != nil {
		t.Fatalf("record owned database: %v", err)
	}
	dropCalls := 0
	dropOwnedDatabaseFn = func(_ context.Context, _ string, _ string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		return errors.New("permission denied")
	}
	if err := harness.dropDatabase(context.Background(), databaseName, suiteservices.FixtureReusePerTest, fixtureAttribution{}); err == nil {
		t.Fatal("expected cleanup failure")
	}
	if dropCalls != 1 {
		t.Fatalf("non-connection failure delegated %d times", dropCalls)
	}
	ledger, found, err := suiteservices.CurrentResourceLedger(nil)
	if err != nil || !found {
		t.Fatalf("read cleanup-failure ledger: found=%t err=%v", found, err)
	}
	if len(ledger.Databases) != 1 || ledger.Databases[0] != databaseName {
		t.Fatalf("failed cleanup lost recovery authority: %#v", ledger.Databases)
	}

	dropOwnedDatabaseFn = func(_ context.Context, _ string, _ string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		return context.Canceled
	}
	if err := harness.dropDatabase(context.Background(), databaseName, suiteservices.FixtureReusePerTest, fixtureAttribution{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("interrupted cleanup error = %v, want context cancellation", err)
	}
	if dropCalls != 2 {
		t.Fatalf("interrupted cleanup delegated %d total times", dropCalls)
	}

	var preparations sync.WaitGroup
	errorsSeen := make(chan error, 16)
	for range 16 {
		preparations.Go(func() {
			_, err := harness.PrepareDatabase(t.Context(), "after_failed_cleanup")
			errorsSeen <- err
		})
	}
	preparations.Wait()
	close(errorsSeen)
	for err := range errorsSeen {
		if !errors.Is(err, errGenerationQuarantined) {
			t.Fatalf("preparation after failed cleanup = %v", err)
		}
	}
	if createCalls.Load() != 0 {
		t.Fatalf("quarantined generation performed %d creates", createCalls.Load())
	}

	// A logical detach cannot stand in for physical settlement by the owner.
	if err := StopShared(t.Context()); err != nil {
		t.Fatal(err)
	}
	if acquired, err := StartShared(t.Context()); acquired != nil || !errors.Is(err, errGenerationQuarantined) {
		t.Fatalf("reacquired quarantined attachment: %v, %v", acquired, err)
	}
	for _, entry := range []struct {
		name, policy string
		run          func(testing.TB)
	}{
		{"shared", "", func(tb testing.TB) { Start(tb) }},
		{"fresh", "", func(tb testing.TB) { harness.NewDatabaseT(tb, "rejected") }},
		{"clone", postgresFixturePolicyTemplateClone, func(tb testing.TB) { harness.PrepareIsolatedDatabaseT(tb, "rejected") }},
		{"open_clone", postgresFixturePolicyTemplateClone, func(tb testing.TB) { harness.OpenIsolatedDatabaseT(tb, "rejected", postgres.PurposeRuntime) }},
		{"migration", "", func(tb testing.TB) { harness.MigrationDatabaseT(tb) }},
		{"migration_through", "", func(tb testing.TB) { harness.MigrationDatabaseThroughT(tb, 1) }},
		{"rollback", postgresFixturePolicyTransaction, func(tb testing.TB) { harness.BeginRollbackDBT(tb, "rejected") }},
	} {
		t.Run(entry.name, func(t *testing.T) {
			t.Setenv(postgresFixturePolicyDefaultEnv, entry.policy)
			expectQuarantinedPreparation(t, entry.run)
		})
	}
	if createCalls.Load() != 0 {
		t.Fatalf("testing entry point prepared a quarantined generation: %d", createCalls.Load())
	}

	// Existing owned resources may finish cleanup, but success cannot restore
	// this generation's eligibility or erase its prior failed row outcome.
	dropOwnedDatabaseFn = func(_ context.Context, _ string, _ string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		return nil
	}
	if err := harness.DropDatabase(t.Context(), databaseName); err != nil {
		t.Fatal(err)
	}
	if _, err := harness.PrepareDatabase(t.Context(), "after_late_settlement"); !errors.Is(err, errGenerationQuarantined) {
		t.Fatalf("late settlement readmitted generation: %v", err)
	}
	independent := &Harness{attached: true, templateDB: "independent_template"}
	db, err := independent.PrepareDatabase(t.Context(), "independent_generation")
	if err != nil || db == nil || createCalls.Load() != 1 {
		t.Fatalf("unrelated generation failed: %v, %v, creates=%d", db, err, createCalls.Load())
	}
	if err := independent.DropDatabase(t.Context(), db.Name); err != nil {
		t.Fatal(err)
	}

	// Partial provisioning may also report failed cleanup after a committed
	// CREATE. Preserve that failure and reject subsequent preparation.
	partialFailure := errors.Join(errors.New("clone finalization failed"), context.DeadlineExceeded)
	createDatabaseFn = func(context.Context, string, string, string) error {
		createCalls.Add(1)
		return partialFailure
	}
	partial := &Harness{attached: true, templateDB: "partial_template"}
	if _, err := partial.PrepareDatabase(t.Context(), "partial"); !errors.Is(err, partialFailure) {
		t.Fatalf("partial preparation lost primary failure: %v", err)
	}
	if _, err := partial.PrepareDatabase(t.Context(), "after_partial"); !errors.Is(err, errGenerationQuarantined) || createCalls.Load() != 2 {
		t.Fatalf("partial preparation retried: %v, creates=%d", err, createCalls.Load())
	}
}

type preparationFailureProbe struct{ testing.TB }
type capturedPreparationFailure struct{ message string }

func (p preparationFailureProbe) Fatal(args ...any) {
	panic(capturedPreparationFailure{message: fmt.Sprint(args...)})
}

func (p preparationFailureProbe) Fatalf(format string, args ...any) {
	panic(capturedPreparationFailure{message: fmt.Sprintf(format, args...)})
}

func expectQuarantinedPreparation(t *testing.T, run func(testing.TB)) {
	t.Helper()
	var captured *capturedPreparationFailure
	func() {
		defer func() {
			if value := recover(); value != nil {
				failure, ok := value.(capturedPreparationFailure)
				if !ok {
					panic(value)
				}
				captured = &failure
			}
		}()
		run(preparationFailureProbe{TB: t})
	}()
	if captured == nil || !strings.HasPrefix(captured.message, testfailure.Marker) {
		t.Fatalf("preparation did not publish typed fixture failure: %#v", captured)
	}
	var envelope testfailure.Envelope
	if err := json.Unmarshal([]byte(strings.TrimPrefix(captured.message, testfailure.Marker)), &envelope); err != nil {
		t.Fatal(err)
	}
	if err := envelope.Validate(); err != nil || envelope.Phase != "setup" || envelope.FailureClass != "harness" || envelope.FailureReason != "fixture_error" || envelope.ReadinessStage != "capability" || envelope.AttemptCount != 0 {
		t.Fatalf("quarantine setup classification: %#v, %v", envelope, err)
	}
}

func TestPrepareIsolatedDatabaseTCleanupDropsAttachedSuiteTemplateClone(t *testing.T) {
	authorizePGSuite(t)
	t.Setenv(postgresFixturePolicyDefaultEnv, postgresFixturePolicyTemplateClone)
	t.Setenv(suiteservices.SuiteIDEnv, "suite-retained-template")
	t.Setenv(suiteservices.TargetEnv, "backend-integration")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "retained-template")

	oldCreate := createDatabaseFn
	oldDrop := dropOwnedDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropOwnedDatabaseFn = oldDrop
	})

	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		return nil
	}
	dropCalls := 0
	dropOwnedDatabaseFn = func(ctx context.Context, adminDSN string, name string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		return nil
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
		attached:    true,
	}

	var preparedName string
	t.Run("prepare", func(t *testing.T) {
		preparedName = harness.PrepareIsolatedDatabaseT(t, "attached-cleanup").Name
	})

	if dropCalls != 1 {
		t.Fatalf("expected attached suite template cleanup to drop %q, got %d drops", preparedName, dropCalls)
	}

	t.Setenv(suiteservices.PersistentBorrowerEnv, "1")
	t.Run("persistent session borrower", func(t *testing.T) {
		harness.PrepareIsolatedDatabaseT(t, "persistent-borrower-cleanup")
	})
	if dropCalls != 2 {
		t.Fatalf("expected a persistent session borrower to use the same eager cleanup path, got %d drops", dropCalls)
	}
}

func TestMigrationDatabaseTCleanupDropsStandaloneScratchDatabase(t *testing.T) {
	t.Setenv(suiteservices.SuiteIDEnv, "")
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")

	oldCreate := createDatabaseFn
	oldDrop := dropOwnedDatabaseFn
	oldMigrate := migrateDatabaseFn
	oldOpen := openMigrationDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropOwnedDatabaseFn = oldDrop
		migrateDatabaseFn = oldMigrate
		openMigrationDatabaseFn = oldOpen
	})

	var scratchName string
	var dropped []string
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		scratchName = name
		return nil
	}
	dropOwnedDatabaseFn = func(ctx context.Context, adminDSN string, name string, _ func(postgrescleanup.DropAttempt)) error {
		dropped = append(dropped, name)
		return nil
	}
	migrateDatabaseFn = func(ctx context.Context, db *sql.DB, source *database_migrations.Source) error {
		return nil
	}
	openMigrationDatabaseFn = func(string, postgres.Purpose) (*sql.DB, error) {
		return sql.Open("pgx", "postgres://fixture/unused")
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
	}

	t.Run("migrate scratch", func(t *testing.T) {
		harness.MigrationDatabaseT(t)
	})

	if len(dropped) != 1 || dropped[0] != scratchName {
		t.Fatalf("expected standalone migration scratch cleanup to drop %q, got %v", scratchName, dropped)
	}
}

func TestMigrationDatabaseTCleanupDropsAttachedSuiteScratchDatabase(t *testing.T) {
	authorizePGSuite(t)
	t.Setenv(suiteservices.SuiteIDEnv, "suite-retained-migration-scratch")
	t.Setenv(suiteservices.TargetEnv, "backend-integration")
	t.Setenv("CARTULARY_TEST_RESULTS_DIR", t.TempDir())
	t.Setenv("CARTULARY_TEST_RUN_ID", "retained-migration-scratch")

	oldCreate := createDatabaseFn
	oldDrop := dropOwnedDatabaseFn
	oldMigrate := applyMigrationsThrough
	oldOpen := openMigrationDatabaseFn
	t.Cleanup(func() {
		createDatabaseFn = oldCreate
		dropOwnedDatabaseFn = oldDrop
		applyMigrationsThrough = oldMigrate
		openMigrationDatabaseFn = oldOpen
	})

	var scratchName string
	dropCalls := 0
	createDatabaseFn = func(ctx context.Context, adminDSN string, name string, templateDB string) error {
		scratchName = name
		return nil
	}
	dropOwnedDatabaseFn = func(ctx context.Context, adminDSN string, name string, _ func(postgrescleanup.DropAttempt)) error {
		dropCalls++
		return nil
	}
	applyMigrationsThrough = func(ctx context.Context, db *sql.DB, version int64) error {
		return nil
	}
	openMigrationDatabaseFn = func(string, postgres.Purpose) (*sql.DB, error) {
		return sql.Open("pgx", "postgres://fixture/unused")
	}

	harness := &Harness{
		adminDSN:    "postgres://cartulary:cartulary@127.0.0.1:5432/postgres?sslmode=disable",
		dsnTemplate: "postgres://cartulary:cartulary@127.0.0.1:5432/{database}?sslmode=disable",
		templateDB:  "suite_template",
		suiteHash:   "suitehash",
		processHash: "procaaaa",
		attached:    true,
	}

	t.Run("migrate scratch", func(t *testing.T) {
		harness.MigrationDatabaseThroughT(t, 1)
	})

	if dropCalls != 1 {
		t.Fatalf("expected attached suite migration scratch cleanup to drop %q, got %d drops", scratchName, dropCalls)
	}
}

func TestMigrationDatabaseTargetedOperationValidation(t *testing.T) {
	oldApply := applyMigrationsThrough
	oldRollback := rollbackMigrationsThrough
	t.Cleanup(func() {
		applyMigrationsThrough = oldApply
		rollbackMigrationsThrough = oldRollback
	})

	applyCalls := 0
	rollbackCalls := 0
	applyMigrationsThrough = func(context.Context, *sql.DB, int64) error {
		applyCalls++
		return nil
	}
	rollbackMigrationsThrough = func(context.Context, *sql.DB, int64) error {
		rollbackCalls++
		return nil
	}

	capabilityType := reflect.TypeOf(MigrationDatabase{})
	for index := 0; index < capabilityType.NumField(); index++ {
		if capabilityType.Field(index).PkgPath == "" {
			t.Fatalf("migration capability field %q must be unexported", capabilityType.Field(index).Name)
		}
	}
	methodType := reflect.TypeOf((*MigrationDatabase)(nil))
	if methodType.NumMethod() != 3 {
		t.Fatalf("migration capability method count = %d, want 3", methodType.NumMethod())
	}
	for index, name := range []string{"ApplyThrough", "RollbackThrough", "SQL"} {
		if got := methodType.Method(index).Name; got != name {
			t.Fatalf("migration capability method %d = %q, want %q", index, got, name)
		}
	}

	var zero MigrationDatabase
	if err := zero.ApplyThrough(context.Background(), 0); err == nil || err.Error() != "migration apply-through version must be positive" {
		t.Fatalf("unexpected apply target validation error: %v", err)
	}
	if err := zero.RollbackThrough(context.Background(), -1); err == nil || err.Error() != "migration rollback-through version must be non-negative" {
		t.Fatalf("unexpected rollback target validation error: %v", err)
	}
	if applyCalls != 0 || rollbackCalls != 0 {
		t.Fatalf("invalid targets reached migration access: apply=%d rollback=%d", applyCalls, rollbackCalls)
	}
	if zero.SQL() != nil {
		t.Fatal("zero capability exposed a database handle")
	}
	if err := zero.ApplyThrough(context.Background(), 1); err == nil || err.Error() != "migration database capability is not harness-issued" {
		t.Fatalf("unexpected zero capability error: %v", err)
	}

	issued := &MigrationDatabase{db: &sql.DB{}, identity: issuedMigrationDatabaseIdentity}
	if issued.SQL() == nil {
		t.Fatal("issued capability did not expose its scratch handle")
	}
	if err := issued.ApplyThrough(context.Background(), 1); err != nil {
		t.Fatalf("apply through issued capability: %v", err)
	}
	if err := issued.RollbackThrough(context.Background(), 0); err != nil {
		t.Fatalf("rollback through zero issued capability: %v", err)
	}
	if applyCalls != 1 || rollbackCalls != 1 {
		t.Fatalf("unexpected targeted call counts: apply=%d rollback=%d", applyCalls, rollbackCalls)
	}
}

func TestHarnessStartsPostgresAndRunsCurrentMigrationPath(t *testing.T) {
	harness := Start(t)

	testDB, err := harness.PrepareDatabase(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("prepare database: %v", err)
	}
	defer func() {
		if err := harness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop database: %v", err)
		}
	}()

	if testDB.DSN == "" {
		t.Fatal("expected database dsn")
	}
}

func TestPrepareIsolatedDatabaseTReturnsMigratedDatabase(t *testing.T) {
	harness := Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "bootstrap_t")

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	var exists bool
	if err := db.QueryRowContext(context.Background(), `
SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'users'
)`).Scan(&exists); err != nil {
		t.Fatalf("query users table: %v", err)
	}
	if !exists {
		t.Fatal("expected PrepareIsolatedDatabaseT to return a migrated database")
	}
	var format string
	if err := db.QueryRowContext(context.Background(), `SELECT format_id FROM public.application_crypto_format WHERE singleton`).Scan(&format); err != nil || format != database_migrations.ApplicationCryptoFormat {
		t.Fatalf("fresh fixture format admission = %q, %v", format, err)
	}
}

func TestBeginRollbackDBTIsolatesRowsPerTransaction(t *testing.T) {
	harness := Start(t)
	var originalRole, originalReplicationRole string

	t.Run("insert rows inside rollback fixture", func(t *testing.T) {
		db := harness.BeginRollbackDBT(t, "rollback-db")
		if err := db.QueryRow(context.Background(), `SELECT current_user, current_setting('session_replication_role')`).Scan(&originalRole, &originalReplicationRole); err != nil {
			t.Fatalf("read initial transaction roles: %v", err)
		}
		if _, err := db.Exec(context.Background(), `INSERT INTO users (email, display_name, password_hash, mfa_required) VALUES ($1, $2, $3, false)`, "rollback@example.test", "Rollback", "hash"); err != nil {
			t.Fatalf("seed rollback transaction: %v", err)
		}
		// A committed nested transaction is only a released savepoint. Its writes
		// and transaction-local roles must not survive the outer fixture rollback.
		nested, err := db.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin nested transaction: %v", err)
		}
		if _, err := nested.Exec(context.Background(), `INSERT INTO users (email, display_name, password_hash, mfa_required) VALUES ($1, $2, $3, false)`, "nested-rollback@example.test", "Nested rollback", "hash"); err != nil {
			t.Fatalf("seed nested transaction: %v", err)
		}
		if err := nested.Commit(context.Background()); err != nil {
			t.Fatalf("commit nested savepoint: %v", err)
		}
		for _, statement := range []string{`SET LOCAL ROLE cartulary_recovery`, `SET LOCAL session_replication_role = replica`} {
			if _, err := db.Exec(context.Background(), statement); err != nil {
				t.Fatalf("set transaction-local state: %v", err)
			}
		}
	})

	t.Run("next rollback fixture starts clean", func(t *testing.T) {
		db := harness.BeginRollbackDBT(t, "rollback-db")
		var count int
		if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM users WHERE email IN ($1, $2)`, "rollback@example.test", "nested-rollback@example.test").Scan(&count); err != nil {
			t.Fatalf("count rollback rows: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected rollback fixture to discard rows, got %d", count)
		}
		var role, replicationRole string
		if err := db.QueryRow(context.Background(), `SELECT current_user, current_setting('session_replication_role')`).Scan(&role, &replicationRole); err != nil {
			t.Fatalf("read next borrower's transaction roles: %v", err)
		}
		if role != originalRole || replicationRole != originalReplicationRole {
			t.Fatalf("transaction-local state leaked: role=%q replication=%q", role, replicationRole)
		}
	})

	scope, ok, err := suiteservices.Summarize(nil)
	if err != nil {
		t.Fatalf("summarize suite service events: %v", err)
	}
	if !ok {
		t.Fatal("expected suite service summary")
	}
	transactionCount := 0
	for _, activity := range scope.Fixture.ByTest {
		if !strings.HasPrefix(activity.TestName, t.Name()+"/") {
			continue
		}
		if activity.Operation == "database-reset" {
			t.Fatalf("rollback fixture must not emit database-reset activity: %#v", activity)
		}
		if activity.Operation == "transaction" && activity.FixturePolicy == postgresFixturePolicyTransaction {
			transactionCount += activity.Count
		}
	}
	if transactionCount != 2 {
		t.Fatalf("expected two attributed rollback transaction events, got %d", transactionCount)
	}
}

func resetSharedHarness(t testing.TB) {
	t.Helper()

	sharedHarnessMu.Lock()
	sharedHarness = nil
	sharedHarnessMu.Unlock()

	t.Cleanup(func() {
		sharedHarnessMu.Lock()
		sharedHarness = nil
		sharedHarnessMu.Unlock()
	})
}

func authorizePGSuite(t *testing.T) {
	t.Helper()
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Setenv(suiteservices.CallModeEnv, "owned")
	t.Setenv(suiteservices.SuiteRuntimeRootEnv, "/private/pgtest-suite-runtime")
	t.Setenv(suiteservices.SuiteRuntimeLeaseIDEnv, "00000000-0000-4000-8000-000000000001")
	t.Setenv(suiteservices.SuiteRuntimeRunIDEnv, "pgtest-suite")
}

func stubOwnedPostgresStartup(t testing.TB) {
	t.Helper()
	// These unit scenarios substitute every service-starting operation. Their
	// certificate workspace belongs to the individual fake lifecycle, not to
	// an attached suite service or a retained result artifact.
	t.Setenv(suiteservices.HarnessServiceDependenciesEnv, "postgres")
	t.Setenv(suiteservices.SuiteIDEnv, "")

	oldInitialize := initializeClusterFn
	initializeClusterFn = func(context.Context, string) (string, error) { return "12345", nil }
	t.Cleanup(func() { initializeClusterFn = oldInitialize })
	oldStartContainer := startContainerFn
	oldWaitReady := waitReadyFn
	oldPreflight := startPreflightFn
	oldSleep := startSleepFn
	oldLogs := postgresStartupLogsFn
	postgresStartupLogsFn = func(ctx context.Context, container testcontainers.Container) ([]postgresLogChunk, bool, error) {
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
	t.Cleanup(func() {
		startContainerFn = oldStartContainer
		waitReadyFn = oldWaitReady
		startPreflightFn = oldPreflight
		startSleepFn = oldSleep
		postgresStartupLogsFn = oldLogs
	})

	startPreflightFn = func(context.Context) (string, error) {
		return "unix:///var/run/docker.sock", nil
	}
	startSleepFn = func(context.Context, time.Duration) error {
		return nil
	}
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

func TestRollbackFixtureCleanupRejectsUnprovenIsolation(t *testing.T) {
	for _, failed := range []string{"", "rollback", "close"} {
		t.Run("failure_"+failed, func(t *testing.T) {
			fixture := &transactionDatabase{db: &TestDatabase{Name: "retained-owned-database"}}
			if err := fixture.lockForBorrow(); err != nil {
				t.Fatal(err)
			}
			cancelled, cancel := context.WithCancel(t.Context())
			cancel()
			var contexts []context.Context
			calls := []string{}
			operation := func(name string) func(context.Context) error {
				return func(ctx context.Context) error {
					if ctx.Err() != nil {
						t.Fatal("cleanup inherited test cancellation")
					}
					deadline, ok := ctx.Deadline()
					if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > 5*time.Second {
						t.Fatal("missing bounded cleanup budget")
					}
					if len(contexts) > 0 && contexts[0].Err() == nil {
						t.Fatal("previous budget was not settled")
					}
					contexts = append(contexts, ctx)
					calls = append(calls, name)
					if failed == name {
						return context.DeadlineExceeded
					}
					return nil
				}
			}
			err := fixture.settle(cancelled, operation("rollback"), operation("close"))
			fixture.mu.Unlock()
			if !reflect.DeepEqual(calls, []string{"rollback", "close"}) {
				t.Fatal(calls)
			}
			if (err != nil) != (failed != "") {
				t.Fatal("incorrect cleanup result", err)
			}
			if fixture.db.Name != "retained-owned-database" {
				t.Fatal("cleanup discarded ownership")
			}
			borrowErr := fixture.lockForBorrow()
			if failed == "" {
				if borrowErr != nil {
					t.Fatal(borrowErr)
				}
				fixture.mu.Unlock()
			} else {
				if !errors.Is(borrowErr, context.DeadlineExceeded) {
					t.Fatal("poisoned fixture admitted another borrower", borrowErr)
				}
				if !fixture.mu.TryLock() {
					t.Fatal("rejected borrower retained fixture mutex")
				}
				fixture.mu.Unlock()
			}
		})
	}
}
