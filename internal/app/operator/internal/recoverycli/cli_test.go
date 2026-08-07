package recoverycli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
)

func TestRecoveryCLIParserAndInvalidInvocationContract(t *testing.T) {
	const backupID = "00000000-0000-0000-0000-000000000428"
	t.Run("canonical operations use current defaults", func(t *testing.T) {
		tests := []struct {
			name      string
			args      []string
			operation string
			timeout   int
			target    string
			confirm   string
		}{
			{name: "backup inspect latest", args: []string{"backup", "inspect", "latest"}, operation: "backup_inspect_latest", timeout: 30},
			{name: "backup create", args: []string{"backup", "create"}, operation: "backup_create", timeout: 14400},
			{name: "restore latest", args: []string{"restore", "latest", "--target-config-file", "/tmp/cartulary-target.toml", "--confirm-backup-set-id", backupID}, operation: "restore_latest", timeout: 14400, target: "/tmp/cartulary-target.toml", confirm: backupID},
			{name: "restore verify latest", args: []string{"restore-verify", "latest", "--target-config-file", "/tmp/cartulary-target.toml"}, operation: "restore_verify_latest", timeout: 14400, target: "/tmp/cartulary-target.toml"},
			{name: "restore verify due", args: []string{"restore-verify", "due", "--target-config-file", "/tmp/cartulary-target.toml"}, operation: "restore_verify_due", timeout: 14400, target: "/tmp/cartulary-target.toml"},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				parsed := parseCommand(test.args)
				if !parsed.Handled || parsed.Invalid || parsed.Err != nil {
					t.Fatalf("canonical command got handled=%t invalid=%t err=%#v", parsed.Handled, parsed.Invalid, parsed.Err)
				}
				if parsed.Operation != test.operation || parsed.TimeoutSeconds != test.timeout {
					t.Fatalf("parsed command mismatch: %#v", parsed)
				}
				if parsed.Output != "json" || parsed.Progress != "" || parsed.SourceConfigPath != "" {
					t.Fatalf("canonical defaults mismatch: %#v", parsed)
				}
				if parsed.TargetConfigPath != test.target || parsed.ConfirmBackupSetID != test.confirm {
					t.Fatalf("canonical target fields mismatch: %#v", parsed)
				}
			})
		}
	})

	t.Run("generic unknown recovery subcommand", func(t *testing.T) {
		parsed := parseCommand([]string{"backup", "unsupported"})
		assertInvalidRecoveryCommand(t, parsed, "unknown_command")
	})

	t.Run("generic unknown flag", func(t *testing.T) {
		parsed := parseCommand([]string{"backup", "create", "--unknown-flag"})
		assertInvalidRecoveryCommand(t, parsed, "invalid_flag_value")
	})

	t.Run("literal absolute target path", func(t *testing.T) {
		parsed := parseCommand([]string{"restore-verify", "latest", "--target-config-file", "/tmp/cartulary-target.toml"})
		if !parsed.Handled || parsed.Invalid || parsed.TargetConfigPath != "/tmp/cartulary-target.toml" {
			t.Fatalf("literal absolute target path was not accepted: %#v", parsed)
		}
	})

	t.Run("unsafe target paths", func(t *testing.T) {
		tests := []struct {
			name string
			path string
		}{
			{name: "relative", path: "relative-target.toml"},
			{name: "tilde", path: "~/target.toml"},
			{name: "shell variable", path: "/tmp/$TARGET.toml"},
			{name: "NUL", path: "/tmp/target\x00.toml"},
			{name: "lexical parent", path: "/tmp/../target.toml"},
			{name: "lexical current", path: "/tmp/./target.toml"},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				parsed := parseCommand([]string{"restore-verify", "latest", "--target-config-file", test.path})
				assertInvalidRecoveryCommand(t, parsed, "invalid_flag_value")
			})
		}
	})

	t.Run("invalid invocation emits one closed failure line", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		handled, exitCode := (runner{stdout: &stdout, stderr: &stderr}).run(
			context.Background(),
			[]string{"backup", "unsupported"},
		)
		if !handled || exitCode != 2 {
			t.Fatalf("invalid invocation got handled=%t exit=%d stdout=%s stderr=%s", handled, exitCode, stdout.String(), stderr.String())
		}
		if stderr.Len() != 0 {
			t.Fatalf("invalid invocation wrote stderr: %q", stderr.String())
		}
		if got := stdout.String(); !strings.HasSuffix(got, "\n") || strings.Count(got, "\n") != 1 {
			t.Fatalf("stdout must be exactly one JSON line, got %q", got)
		}
		var payload result
		if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
			t.Fatalf("decode recovery result: %v\nstdout=%s", err, stdout.String())
		}
		if payload.SchemaID != resultSchemaID || payload.Result != "failed" || payload.Error == nil {
			t.Fatalf("unexpected recovery result: %#v", payload)
		}
		if payload.Error.Code != "invalid_operator_request" || payload.Error.ReasonCode != "unknown_command" {
			t.Fatalf("unexpected recovery error: %#v", payload.Error)
		}
		if len(payload.ArtifactRefs) != 0 {
			t.Fatalf("failure envelope fabricated artifacts: %#v", payload.ArtifactRefs)
		}
	})
}

func assertInvalidRecoveryCommand(t *testing.T, parsed command, reasonCode string) {
	t.Helper()
	if !parsed.Handled || !parsed.Invalid || parsed.Err == nil {
		t.Fatalf("command was not rejected as Recovery input: %#v", parsed)
	}
	if parsed.Err.ReasonCode != reasonCode {
		t.Fatalf("reason_code got %q want %q", parsed.Err.ReasonCode, reasonCode)
	}
}

func TestOperatorRecoveryFailureKindsMapExhaustivelyToClosedWirePairs(t *testing.T) {
	tests := []struct {
		kind       application.FailureKind
		code       string
		reasonCode string
		exitCode   int
	}{
		{application.FailureConfirmationMismatch, "invalid_operator_request", "confirmation_mismatch", 2},
		{application.FailureLocalConfigInvalid, "invalid_operator_request", "local_config_invalid", 2},
		{application.FailureSecretReferenceMissing, "recovery_key_unavailable", "secret_reference_missing", 3},
		{application.FailureSecretReferenceUnresolved, "recovery_key_unavailable", "secret_reference_unresolved", 3},
		{application.FailureRecoveryKeyInvalid, "recovery_key_unavailable", "recovery_key_invalid", 3},
		{application.FailureNoSuccessfulRetainedBackup, "backup_set_not_found", "no_successful_retained_backup", 3},
		{application.FailureSelectedBackupNotRetained, "backup_set_not_found", "selected_backup_not_retained", 3},
		{application.FailureArtifactMissing, "backup_integrity_failed", "artifact_missing", 3},
		{application.FailureIntegrityProofMissing, "backup_integrity_failed", "integrity_proof_missing", 3},
		{application.FailureChecksumMismatch, "backup_integrity_failed", "checksum_mismatch", 3},
		{application.FailureAttestationInvalid, "backup_integrity_failed", "attestation_invalid", 3},
		{application.FailureSameDatabaseBinding, "unsafe_restore_target", "same_database_binding", 3},
		{application.FailureSameObjectStoreBinding, "unsafe_restore_target", "same_object_store_binding", 3},
		{application.FailureTargetDatabaseNotFresh, "unsafe_restore_target", "target_database_not_fresh", 3},
		{application.FailureTargetObjectNamespaceNotFresh, "unsafe_restore_target", "target_object_namespace_not_fresh", 3},
		{application.FailureTargetServingTraffic, "unsafe_restore_target", "target_serving_traffic", 3},
		{application.FailureTargetMarkerMissing, "unsafe_restore_target", "target_marker_missing", 3},
		{application.FailureTargetMarkerInvalid, "unsafe_restore_target", "target_marker_invalid", 3},
		{application.FailureOperationLockUnavailable, "recovery_operation_in_progress", "operation_lock_unavailable", 3},
		{application.FailureTimeoutElapsed, "operation_timed_out", "timeout_elapsed", 4},
		{application.FailureBackupPostgres, "backup_create_failed", "postgres_backup_failed", 4},
		{application.FailureBackupObject, "backup_create_failed", "object_backup_failed", 4},
		{application.FailureBackupIntegrityProof, "backup_create_failed", "integrity_proof_failed", 4},
		{application.FailureBackupArtifactReadback, "backup_create_failed", "artifact_readback_failed", 4},
		{application.FailureBackupAttestationWrite, "backup_create_failed", "attestation_write_failed", 4},
		{application.FailureBackupPublication, "backup_create_failed", "backup_publication_failed", 4},
		{application.FailureBackupJournalWrite, "backup_create_failed", "journal_write_failed", 4},
		{application.FailureRestorePostgres, "restore_failed", "postgres_restore_failed", 4},
		{application.FailureRestoreObject, "restore_failed", "object_restore_failed", 4},
		{application.FailureRestoreProjectionRebuild, "restore_failed", "projection_rebuild_failed", 4},
		{application.FailureRestoreInvariantCheck, "restore_failed", "invariant_check_failed", 4},
		{application.FailureRestoreJournalWrite, "restore_failed", "journal_write_failed", 4},
		{application.FailureVerificationPostgres, "verification_failed", "postgres_restore_failed", 4},
		{application.FailureVerificationObject, "verification_failed", "object_restore_failed", 4},
		{application.FailureVerificationProjectionRebuild, "verification_failed", "projection_rebuild_failed", 4},
		{application.FailureVerificationInvariantCheck, "verification_failed", "invariant_check_failed", 4},
		{application.FailureVerificationWorkbookProbe, "verification_failed", "workbook_probe_failed", 4},
		{application.FailureVerificationAttestationUpdate, "verification_failed", "attestation_update_failed", 4},
		{application.FailureVerificationJournalWrite, "verification_failed", "journal_write_failed", 4},
	}

	if len(tests) != len(application.AllFailureKinds()) {
		t.Fatalf("typed failure mapping test count got %d want %d", len(tests), len(application.AllFailureKinds()))
	}
	seen := make(map[application.FailureKind]struct{}, len(tests))
	for _, test := range tests {
		t.Run(string(test.kind), func(t *testing.T) {
			if _, duplicate := seen[test.kind]; duplicate {
				t.Fatalf("duplicate failure kind %q", test.kind)
			}
			seen[test.kind] = struct{}{}
			payload, exitCode := mapFailureKind(test.kind)
			if exitCode != test.exitCode {
				t.Fatalf("exit code got %d want %d", exitCode, test.exitCode)
			}
			if payload.Code != test.code || payload.ReasonCode != test.reasonCode {
				t.Fatalf("error got %#v", payload)
			}
		})
	}
}

func TestRunnerLeavesDueInvocationUnboundedForPerAttemptTimeouts(t *testing.T) {
	facade := &deadlineInspectingFacade{t: t}
	var stdout bytes.Buffer
	handled, exitCode := (runner{
		stdout: &stdout,
		facade: facade,
	}).run(context.Background(), []string{
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
