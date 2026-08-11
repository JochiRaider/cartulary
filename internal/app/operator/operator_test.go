package operator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
)

func TestOperatorObjectStoreInitCommand_U_DeploymentLocalResult(t *testing.T) {
	parsed, stop, parseExitCode := parseObjectStoreInitArgs([]string{"-config", "  /etc/cartulary/config.toml  "}, io.Discard)
	if stop || parseExitCode != 0 || parsed.sourceConfigPath != "/etc/cartulary/config.toml" {
		t.Fatalf("object-store parse got args=%#v stop=%t exit=%d", parsed, stop, parseExitCode)
	}
	if _, stop, exitCode := parseObjectStoreInitArgs([]string{"-help"}, io.Discard); !stop || exitCode != 0 {
		t.Fatalf("object-store help got stop=%t exit=%d", stop, exitCode)
	}
	if _, stop, exitCode := parseObjectStoreInitArgs([]string{"-unknown"}, io.Discard); !stop || exitCode != 2 {
		t.Fatalf("object-store invalid input got stop=%t exit=%d", stop, exitCode)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var gotConfigPath string
	var ensureCalls int
	objectStorageRoot := t.TempDir()
	runner := operatorRunner{
		objectStore: objectStoreExecutor{
			transport: operatorTransport{stdout: &stdout, stderr: &stderr},
			loadConfig: func(path string) (configassembly.Loaded, error) {
				gotConfigPath = path
				return objectStoreTestConfig(t, objectStorageRoot), nil
			},
			ensureObjectStoreBucket: func(context.Context, objectstore.Settings) (objectstore.EnsureBucketResult, error) {
				ensureCalls++
				return objectstore.EnsureBucketResult{Created: true}, nil
			},
		},
	}

	exitCode := runner.runCLI(context.Background(), []string{"object-store", "init", "-config", "/etc/cartulary/config.toml"})
	if exitCode != 0 {
		t.Fatalf("object-store init failed: exit=%d stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	if gotConfigPath != "/etc/cartulary/config.toml" {
		t.Fatalf("unexpected config path: got %q", gotConfigPath)
	}
	if ensureCalls != 1 {
		t.Fatalf("expected exactly one ensure call, got %d", ensureCalls)
	}

	var payload operatorObjectStoreInitResult
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode object-store init payload: %v\nstdout=%s", err, stdout.String())
	}
	if payload.SchemaID != operatorObjectStoreInitResultSchemaID || payload.Result != "created" || !payload.Created || payload.AlreadyExists {
		t.Fatalf("unexpected object-store init payload: %#v", payload)
	}
	if strings.Contains(stdout.String(), "bucket") || strings.Contains(stdout.String(), "endpoint") {
		t.Fatalf("object-store init result exposed storage details: %s", stdout.String())
	}
}

func TestOperatorCollaborationRequeueArgs_U_StrictCanonicalGrammar(t *testing.T) {
	incidentID := uuid.MustParse("20000000-0000-0000-0000-000000000002")
	for _, args := range [][]string{
		{"--incident-id", incidentID.String(), "--config", "/etc/cartulary/config.toml", "--timeout-seconds", "45"},
		{"--timeout-seconds=45", "--config=/etc/cartulary/config.toml", "--incident-id=" + incidentID.String()},
	} {
		parsed, failure := parseCollaborationRequeueArgs(args)
		if failure != nil || parsed.incidentID != incidentID || parsed.configPath != "/etc/cartulary/config.toml" || parsed.timeout != 45*time.Second {
			t.Fatalf("strict collaboration arguments failed: args=%v parsed=%#v failure=%#v", args, parsed, failure)
		}
	}
	parsed, failure := parseCollaborationRequeueArgs([]string{"--incident-id", incidentID.String()})
	if failure != nil || parsed.timeout != 30*time.Second || parsed.configPath != "" {
		t.Fatalf("default collaboration arguments changed: parsed=%#v failure=%#v", parsed, failure)
	}
}

func TestOperatorCollaborationRequeueArgs_U_RejectsClosedGrammarMatrix(t *testing.T) {
	incidentID := "20000000-0000-0000-0000-000000000002"
	tests := []struct {
		name       string
		args       []string
		wantReason string
	}{
		{name: "missing incident", args: nil, wantReason: "missing_required_flag"},
		{name: "single dash", args: []string{"-incident-id", incidentID}, wantReason: "unknown_flag"},
		{name: "duplicate", args: []string{"--incident-id", incidentID, "--incident-id", incidentID}, wantReason: "duplicate_flag"},
		{name: "terminator", args: []string{"--incident-id", incidentID, "--"}, wantReason: "unexpected_argument"},
		{name: "positional", args: []string{"--incident-id", incidentID, "extra"}, wantReason: "unexpected_argument"},
		{name: "unknown", args: []string{"--incident-id", incidentID, "--retry"}, wantReason: "unknown_flag"},
		{name: "empty", args: []string{"--incident-id="}, wantReason: "invalid_flag_value"},
		{name: "zero UUID", args: []string{"--incident-id", uuid.Nil.String()}, wantReason: "invalid_flag_value"},
		{name: "uppercase UUID", args: []string{"--incident-id", "20000000-0000-0000-0000-00000000000A"}, wantReason: "invalid_flag_value"},
		{name: "compact UUID", args: []string{"--incident-id", strings.ReplaceAll(incidentID, "-", "")}, wantReason: "invalid_flag_value"},
		{name: "URN UUID", args: []string{"--incident-id", "urn:uuid:" + incidentID}, wantReason: "invalid_flag_value"},
		{name: "braced UUID", args: []string{"--incident-id", "{" + incidentID + "}"}, wantReason: "invalid_flag_value"},
		{name: "relative config", args: []string{"--incident-id", incidentID, "--config", "config.toml"}, wantReason: "invalid_flag_value"},
		{name: "dot config", args: []string{"--incident-id", incidentID, "--config", "/etc/../cartulary.toml"}, wantReason: "invalid_flag_value"},
		{name: "tilde config", args: []string{"--incident-id", incidentID, "--config", "/etc/~secret.toml"}, wantReason: "invalid_flag_value"},
		{name: "variable config", args: []string{"--incident-id", incidentID, "--config", "/etc/$CONFIG.toml"}, wantReason: "invalid_flag_value"},
		{name: "timeout low", args: []string{"--incident-id", incidentID, "--timeout-seconds", "0"}, wantReason: "invalid_flag_value"},
		{name: "timeout high", args: []string{"--incident-id", incidentID, "--timeout-seconds", "301"}, wantReason: "invalid_flag_value"},
		{name: "timeout decimal", args: []string{"--incident-id", incidentID, "--timeout-seconds", "30.5"}, wantReason: "invalid_flag_value"},
		{name: "mixed help", args: []string{"--help", "--incident-id", incidentID}, wantReason: "unexpected_argument"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, failure := parseCollaborationRequeueArgs(test.args)
			if failure == nil || failure.reasonCode != test.wantReason {
				t.Fatalf("grammar result = %#v want %q", failure, test.wantReason)
			}
		})
	}
}

func TestOperatorCollaborationRequeueCommand_U_V2DeliveryAndClosure(t *testing.T) {
	operationID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	incidentID := uuid.MustParse("20000000-0000-0000-0000-000000000002")
	t.Setenv("CARTULARY_POSTGRES_POSTGRES_PRIMARY_RUNTIME_DSN", "postgres://unit-test")

	t.Run("help is the sole non-envelope path", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		runner := collaborationV2TestRunner(t, &stdout, &stderr, nil, nil, operationID)
		if exitCode := runner.runCLI(context.Background(), []string{"collaboration", "requeue", "--help"}); exitCode != 0 {
			t.Fatalf("help exit=%d", exitCode)
		}
		if stdout.Len() != 0 || stderr.String() != "usage: "+collaborationRequeueUsage+"\n" {
			t.Fatalf("help transport changed: stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
	})

	t.Run("exact success envelope and acquired pool closure", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		pool := &collaborationRequeueFakePool{}
		var request collaboration.RequeueRequest
		port := collaborationRecoveryPortFunc(func(_ context.Context, got collaboration.RequeueRequest) (collaboration.RequeueResult, error) {
			request = got
			return collaboration.RequeueResult{RequeuedIntentCount: 2}, nil
		})
		runner := collaborationV2TestRunner(t, &stdout, &stderr, pool, port, operationID)

		exitCode := runner.runCLI(context.Background(), []string{"collaboration", "requeue", "--incident-id", incidentID.String()})
		if exitCode != 0 {
			t.Fatalf("collaboration requeue exit=%d stderr=%q", exitCode, stderr.String())
		}
		want := "{\"schema_id\":\"cartulary.operator.collaboration_requeue_result.v2\",\"operation_id\":\"10000000-0000-0000-0000-000000000001\",\"operation\":\"collaboration_requeue\",\"result\":\"succeeded\",\"started_at\":\"2026-08-06T17:00:00Z\",\"completed_at\":\"2026-08-06T17:00:02Z\",\"incident_id\":\"20000000-0000-0000-0000-000000000002\",\"requeued_intent_count\":2,\"error\":null}\n"
		if stdout.String() != want || stderr.Len() != 0 {
			t.Fatalf("v2 delivery changed: stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
		if request.OperationID != operationID || request.IncidentID != incidentID || !request.MutatedAt.Equal(time.Date(2026, 8, 6, 17, 0, 1, 0, time.UTC)) || !pool.closed {
			t.Fatalf("request/lifecycle changed: request=%#v pool_closed=%v", request, pool.closed)
		}
	})

	t.Run("invalid invocation emits one closed failure envelope", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		runner := collaborationV2TestRunner(t, &stdout, &stderr, nil, nil, operationID)
		exitCode := runner.runCLI(context.Background(), []string{"collaboration", "requeue", "--incident-id="})
		if exitCode != 2 || stderr.Len() != 0 || strings.Count(stdout.String(), "\n") != 1 {
			t.Fatalf("invalid transport changed: exit=%d stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
		}
		payload := decodeCollaborationRequeueResult(t, stdout.String())
		if payload.Error == nil || payload.Error.Code != "invalid_operator_request" || payload.Error.ReasonCode != "invalid_flag_value" || payload.IncidentID != nil {
			t.Fatalf("invalid envelope changed: %#v", payload)
		}
	})

	for _, test := range []struct {
		name       string
		kind       collaboration.RequeueFailureKind
		wantCode   string
		wantReason string
		wantExit   int
	}{
		{name: "not quarantined", kind: collaboration.RequeueFailureIncidentNotQuarantined, wantCode: "collaboration_requeue_rejected", wantReason: "incident_not_quarantined", wantExit: 3},
		{name: "repair not verified", kind: collaboration.RequeueFailureRepairNotVerified, wantCode: "collaboration_requeue_rejected", wantReason: "repair_not_verified", wantExit: 3},
		{name: "transaction", kind: collaboration.RequeueFailureTransaction, wantCode: "collaboration_requeue_failed", wantReason: "transaction_failed", wantExit: 4},
		{name: "commit unknown", kind: collaboration.RequeueFailureCommitOutcomeUnknown, wantCode: "collaboration_requeue_failed", wantReason: "commit_outcome_unknown", wantExit: 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			pool := &collaborationRequeueFakePool{}
			port := collaborationRecoveryPortFunc(func(context.Context, collaboration.RequeueRequest) (collaboration.RequeueResult, error) {
				return collaboration.RequeueResult{}, &collaboration.RequeueFailure{Kind: test.kind}
			})
			runner := collaborationV2TestRunner(t, &stdout, &stderr, pool, port, operationID)
			exitCode := runner.runCLI(context.Background(), []string{"collaboration", "requeue", "--incident-id", incidentID.String()})
			payload := decodeCollaborationRequeueResult(t, stdout.String())
			if exitCode != test.wantExit || payload.Error == nil || payload.Error.Code != test.wantCode || payload.Error.ReasonCode != test.wantReason || payload.RequeuedIntentCount != nil || stderr.Len() != 0 || !pool.closed {
				t.Fatalf("failure mapping changed: exit=%d payload=%#v stderr=%q pool_closed=%v", exitCode, payload, stderr.String(), pool.closed)
			}
		})
	}

	t.Run("configuration and Postgres failures are secret safe", func(t *testing.T) {
		for _, stage := range []string{"config", "postgres"} {
			t.Run(stage, func(t *testing.T) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				runner := collaborationV2TestRunner(t, &stdout, &stderr, nil, nil, operationID)
				if stage == "config" {
					runner.collaboration.loadConfig = func(string) (configassembly.Loaded, error) {
						return configassembly.Loaded{}, errors.New("forbidden-config-secret")
					}
				} else {
					runner.collaboration.setupPostgres = func(context.Context, postgres.Settings) (operatorPostgresPool, error) {
						return nil, errors.New("forbidden-postgres-secret")
					}
				}
				exitCode := runner.runCLI(context.Background(), []string{"collaboration", "requeue", "--incident-id", incidentID.String()})
				payload := decodeCollaborationRequeueResult(t, stdout.String())
				wantReason := "local_config_invalid"
				wantExit := 2
				if stage == "postgres" {
					wantReason = "postgres_unavailable"
					wantExit = 4
				}
				if exitCode != wantExit || payload.Error == nil || payload.Error.ReasonCode != wantReason || strings.Contains(stdout.String()+stderr.String(), "forbidden-") || stderr.Len() != 0 {
					t.Fatalf("safe %s failure changed: exit=%d payload=%#v stdout=%q stderr=%q", stage, exitCode, payload, stdout.String(), stderr.String())
				}
			})
		}
	})

	t.Run("caller cancellation and timeout remain distinct", func(t *testing.T) {
		for _, test := range []struct {
			name       string
			ctx        context.Context
			wantCode   string
			wantReason string
		}{
			{name: "cancelled", ctx: cancelledContext(), wantCode: "operation_cancelled", wantReason: "caller_cancelled"},
			{name: "timed out", ctx: expiredContext(), wantCode: "operation_timed_out", wantReason: "timeout_elapsed"},
		} {
			t.Run(test.name, func(t *testing.T) {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				runner := collaborationV2TestRunner(t, &stdout, &stderr, nil, nil, operationID)
				exitCode := runner.runCLI(test.ctx, []string{"collaboration", "requeue", "--incident-id", incidentID.String()})
				payload := decodeCollaborationRequeueResult(t, stdout.String())
				if exitCode != 4 || payload.Error == nil || payload.Error.Code != test.wantCode || payload.Error.ReasonCode != test.wantReason || stderr.Len() != 0 {
					t.Fatalf("context failure changed: exit=%d payload=%#v stderr=%q", exitCode, payload, stderr.String())
				}
			})
		}
	})

	t.Run("timeout flag bounds Postgres and transactional work", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		pool := &collaborationRequeueFakePool{}
		port := collaborationRecoveryPortFunc(func(ctx context.Context, _ collaboration.RequeueRequest) (collaboration.RequeueResult, error) {
			<-ctx.Done()
			return collaboration.RequeueResult{}, ctx.Err()
		})
		runner := collaborationV2TestRunner(t, &stdout, &stderr, pool, port, operationID)
		setup := runner.collaboration.setupPostgres
		var setupHadDeadline bool
		runner.collaboration.setupPostgres = func(ctx context.Context, settings postgres.Settings) (operatorPostgresPool, error) {
			_, setupHadDeadline = ctx.Deadline()
			return setup(ctx, settings)
		}
		started := time.Now()
		exitCode := runner.runCLI(context.Background(), []string{
			"collaboration", "requeue",
			"--incident-id", incidentID.String(),
			"--timeout-seconds", "1",
		})
		payload := decodeCollaborationRequeueResult(t, stdout.String())
		elapsed := time.Since(started)
		if exitCode != 4 || payload.Error == nil || payload.Error.Code != "operation_timed_out" || payload.Error.ReasonCode != "timeout_elapsed" || stderr.Len() != 0 || !pool.closed || !setupHadDeadline {
			t.Fatalf("operation timeout changed: exit=%d payload=%#v stderr=%q pool_closed=%v setup_deadline=%v", exitCode, payload, stderr.String(), pool.closed, setupHadDeadline)
		}
		if elapsed < 900*time.Millisecond || elapsed > 3*time.Second {
			t.Fatalf("operation timeout elapsed outside expected bound: %s", elapsed)
		}
	})

	t.Run("stdout failure after success reports only the delivery exception", func(t *testing.T) {
		var stderr bytes.Buffer
		pool := &collaborationRequeueFakePool{}
		port := collaborationRecoveryPortFunc(func(context.Context, collaboration.RequeueRequest) (collaboration.RequeueResult, error) {
			return collaboration.RequeueResult{RequeuedIntentCount: 1}, nil
		})
		runner := collaborationV2TestRunner(t, operatorFailWriter{}, &stderr, pool, port, operationID)
		exitCode := runner.runCLI(context.Background(), []string{"collaboration", "requeue", "--incident-id", incidentID.String()})
		if exitCode != 4 || stderr.String() != "operation_id=10000000-0000-0000-0000-000000000001 result_delivery_failed\n" || !pool.closed {
			t.Fatalf("delivery exception changed: exit=%d stderr=%q pool_closed=%v", exitCode, stderr.String(), pool.closed)
		}
	})
}

func TestOperatorObjectStoreInitCommand_U_RedactsFailure(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	objectStorageRoot := t.TempDir()
	forbidden := []string{
		"http://127.0.0.1:9000",
		"secret-bucket",
		"AKIA-SECRET",
		"object/key.txt",
		"storage://unsafe/ref",
		"postgres://user:pass@db.example.test/cartulary",
	}
	runner := operatorRunner{
		objectStore: objectStoreExecutor{
			transport: operatorTransport{stdout: &stdout, stderr: &stderr},
			loadConfig: func(string) (configassembly.Loaded, error) {
				return objectStoreTestConfig(t, objectStorageRoot), nil
			},
			ensureObjectStoreBucket: func(context.Context, objectstore.Settings) (objectstore.EnsureBucketResult, error) {
				return objectstore.EnsureBucketResult{}, errors.New(strings.Join(forbidden, " "))
			},
		},
	}

	exitCode := runner.runCLI(context.Background(), []string{"object-store", "init"})
	if exitCode == 0 {
		t.Fatalf("expected object-store init failure, stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no success payload on failure, got %s", stdout.String())
	}
	for _, value := range forbidden {
		if strings.Contains(stderr.String(), value) {
			t.Fatalf("object-store init failure leaked forbidden value %q in stderr %q", value, stderr.String())
		}
	}
	if !strings.Contains(stderr.String(), "object-store init failed: reason_code=dependency_unavailable") {
		t.Fatalf("stderr did not include redacted failure reason: %s", stderr.String())
	}
}

func TestOperatorObjectStoreInitCommand_U_UsesOnlyTypedFailureClassification(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantReason string
	}{
		{
			name: "typed config diagnostic",
			err: config.NewDiagnosticsError(config.Diagnostic{
				ReasonCode: "unsafe-secret-marker",
				Message:    "forbidden config detail",
			}),
			wantReason: config.InvalidDeploymentConfigCode,
		},
		{
			name: "typed adapter",
			err: &objectstore.AdapterError{
				Code:      objectstore.ErrorCodeUnavailable,
				Reason:    objectstore.ReasonBucketMissing,
				Operation: objectstore.OperationEnsureBucket,
				Message:   "forbidden bucket detail",
			},
			wantReason: string(objectstore.ReasonBucketMissing),
		},
		{
			name:       "untyped text resembling former matcher",
			err:        errors.New("service_ref missing object-store parse CARTULARY_S3_SECRET"),
			wantReason: "dependency_unavailable",
		},
		{
			name: "unknown typed reason fails generic",
			err: &objectstore.AdapterError{
				Code:    objectstore.ErrorCodeUnavailable,
				Reason:  objectstore.ReasonCode("forbidden-secret-reason"),
				Message: "forbidden adapter detail",
			},
			wantReason: "dependency_unavailable",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := sanitizeObjectStoreInitError(test.err)
			want := "object-store init failed: reason_code=" + test.wantReason
			if got == nil || got.Error() != want {
				t.Fatalf("typed classification = %v want %q", got, want)
			}
			for _, forbidden := range []string{"unsafe-secret-marker", "forbidden", "service_ref", "CARTULARY_S3_SECRET"} {
				if strings.Contains(got.Error(), forbidden) {
					t.Fatalf("classification leaked %q in %q", forbidden, got.Error())
				}
			}
		})
	}
}

func objectStoreTestConfig(t testing.TB, rootPath string) configassembly.Loaded {
	t.Helper()
	return configtest.LoadFixture(t, []string{"config", "valid.toml"}, map[string]string{
		"CARTULARY__ROOTS__OBJECT_STORAGE__BINDING_KIND": "filesystem_root",
		"CARTULARY__ROOTS__OBJECT_STORAGE__PATH":         rootPath,
		"CARTULARY__ROOTS__OBJECT_STORAGE__SERVICE_REF":  "",
	})
}

func collaborationV2TestRunner(
	t testing.TB,
	stdout io.Writer,
	stderr io.Writer,
	pool operatorPostgresPool,
	port collaborationRecoveryPort,
	operationID uuid.UUID,
) operatorRunner {
	t.Helper()
	clockIndex := 0
	clock := []time.Time{
		time.Date(2026, 8, 6, 17, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 6, 17, 0, 1, 0, time.UTC),
		time.Date(2026, 8, 6, 17, 0, 2, 0, time.UTC),
	}
	return operatorRunner{
		stderr: stderr,
		collaboration: collaborationExecutor{
			transport: operatorTransport{stdout: stdout, stderr: stderr},
			loadConfig: func(string) (configassembly.Loaded, error) {
				return migrationEvidenceTestDeployment(t), nil
			},
			setupPostgres: func(context.Context, postgres.Settings) (operatorPostgresPool, error) {
				if pool == nil {
					return nil, errors.New("missing test pool")
				}
				return pool, nil
			},
			now: func() time.Time {
				if clockIndex >= len(clock) {
					return clock[len(clock)-1]
				}
				value := clock[clockIndex]
				clockIndex++
				return value
			},
			newOperationID: func() uuid.UUID {
				return operationID
			},
			newRecoveryPort: func(postgres.DB) collaborationRecoveryPort {
				if port == nil {
					return collaborationRecoveryPortFunc(func(context.Context, collaboration.RequeueRequest) (collaboration.RequeueResult, error) {
						return collaboration.RequeueResult{}, errors.New("missing test recovery port")
					})
				}
				return port
			},
		},
	}
}

type collaborationRecoveryPortFunc func(context.Context, collaboration.RequeueRequest) (collaboration.RequeueResult, error)

func (port collaborationRecoveryPortFunc) RequeueIncident(ctx context.Context, request collaboration.RequeueRequest) (collaboration.RequeueResult, error) {
	return port(ctx, request)
}

func decodeCollaborationRequeueResult(t testing.TB, encoded string) operatorCollaborationRequeueResult {
	t.Helper()
	var result operatorCollaborationRequeueResult
	if err := json.Unmarshal([]byte(encoded), &result); err != nil {
		t.Fatalf("decode collaboration requeue result: %v\nencoded=%s", err, encoded)
	}
	return result
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func expiredContext() context.Context {
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	cancel()
	return ctx
}

type operatorFailWriter struct{}

func (operatorFailWriter) Write([]byte) (int, error) {
	return 0, errors.New("injected writer failure")
}

type collaborationRequeueFakePool struct {
	closed bool
}

func (pool *collaborationRequeueFakePool) Close() {
	pool.closed = true
}

func (*collaborationRequeueFakePool) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected pool Exec")
}

func (*collaborationRequeueFakePool) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected pool Query")
}

func (*collaborationRequeueFakePool) QueryRow(context.Context, string, ...any) pgx.Row {
	return migrationEvidenceFakeRow{err: errors.New("unexpected pool QueryRow")}
}

func (*collaborationRequeueFakePool) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errors.New("unexpected pool BeginTx")
}
