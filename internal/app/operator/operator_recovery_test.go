package operator

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/operator/recoverycli"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
)

func TestOperatorRecoveryParserAcceptsCanonicalCommands(t *testing.T) {
	backupID := "00000000-0000-0000-0000-000000000428"
	tests := []struct {
		name      string
		args      []string
		operation string
		timeout   int
		progress  string
	}{
		{
			name:      "backup inspect latest",
			args:      []string{"backup", "inspect", "latest", "--source-config-file", "source.toml", "--progress", "jsonl", "--timeout-seconds", "10"},
			operation: "backup_inspect_latest",
			timeout:   10,
			progress:  "jsonl",
		},
		{
			name:      "backup create",
			args:      []string{"backup", "create"},
			operation: "backup_create",
			timeout:   14400,
		},
		{
			name:      "restore latest",
			args:      []string{"restore", "latest", "--target-config-file", "/tmp/cartulary-target.toml", "--confirm-backup-set-id", backupID, "--timeout-seconds", "120"},
			operation: "restore_latest",
			timeout:   120,
		},
		{
			name:      "restore verify latest",
			args:      []string{"restore-verify", "latest", "--target-config-file", "/tmp/cartulary-target.toml"},
			operation: "restore_verify_latest",
			timeout:   14400,
		},
		{
			name:      "restore verify due",
			args:      []string{"restore-verify", "due", "--target-config-file", "/tmp/cartulary-target.toml"},
			operation: "restore_verify_due",
			timeout:   14400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed := recoverycli.ParseCommand(test.args)
			if !parsed.Handled || parsed.Invalid {
				t.Fatalf("parse canonical command got handled=%t invalid=%t err=%#v", parsed.Handled, parsed.Invalid, parsed.Err)
			}
			if parsed.Operation != test.operation || parsed.TimeoutSeconds != test.timeout || parsed.Progress != test.progress {
				t.Fatalf("parsed command mismatch: %#v", parsed)
			}
		})
	}
}

func TestOperatorRecoveryParserRejectsLegacyRecoverySurface(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		reasonCode string
	}{
		{
			name:       "legacy backup capture command",
			args:       []string{"backup", "capture"},
			reasonCode: "unknown_command",
		},
		{
			name:       "deployment admin flag",
			args:       []string{"backup", "create", "--deployment-admin-email", "admin@example.test"},
			reasonCode: "invalid_flag_value",
		},
		{
			name:       "legacy target flag",
			args:       []string{"restore", "latest", "--target-config", "/tmp/target.toml", "--confirm-backup-set-id", "00000000-0000-0000-0000-000000000428"},
			reasonCode: "invalid_flag_value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed := recoverycli.ParseCommand(test.args)
			if !parsed.Handled || !parsed.Invalid || parsed.Err == nil {
				t.Fatalf("legacy command was not rejected as recovery: %#v", parsed)
			}
			if parsed.Err.ReasonCode != test.reasonCode {
				t.Fatalf("reason_code got %q want %q", parsed.Err.ReasonCode, test.reasonCode)
			}
		})
	}
}

func TestOperatorRecoveryParserLeavesRetiredTopLevelNamesForRegistryUsage(t *testing.T) {
	parsed := recoverycli.ParseCommand([]string{"backup-metadata", "latest"})
	if parsed.Handled {
		t.Fatalf("retired top-level name was claimed by recovery: %#v", parsed)
	}
}

func TestOperatorRecoveryParserRejectsUnsafeTargetPaths(t *testing.T) {
	backupID := "00000000-0000-0000-0000-000000000428"
	tests := []string{
		"relative-target.toml",
		"~/target.toml",
		"/tmp/$TARGET.toml",
		"/tmp/../target.toml",
		"/tmp/./target.toml",
	}
	for _, targetPath := range tests {
		t.Run(targetPath, func(t *testing.T) {
			parsed := recoverycli.ParseCommand([]string{"restore", "latest", "--target-config-file", targetPath, "--confirm-backup-set-id", backupID})
			if !parsed.Handled || !parsed.Invalid || parsed.Err == nil {
				t.Fatalf("unsafe target path was accepted: %#v", parsed)
			}
			if parsed.Err.ReasonCode != "invalid_flag_value" {
				t.Fatalf("reason_code got %q want invalid_flag_value", parsed.Err.ReasonCode)
			}
		})
	}
}

func TestOperatorRecoveryCLIEmitsSingleFailureEnvelopeForLegacyCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := RunOperatorCLIContext(context.Background(), []string{"backup", "capture"}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("exit code got %d want 2; stdout=%s stderr=%s", exitCode, stdout.String(), stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("legacy recovery rejection wrote stderr: %s", stderr.String())
	}
	if got := stdout.String(); !strings.HasSuffix(got, "\n") || strings.Count(got, "\n") != 1 {
		t.Fatalf("stdout must be exactly one JSON line, got %q", got)
	}
	var payload recoverycli.Result
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode recovery result: %v\nstdout=%s", err, stdout.String())
	}
	if payload.SchemaID != recoverycli.ResultSchemaID || payload.Result != "failed" || payload.Error == nil {
		t.Fatalf("unexpected recovery result: %#v", payload)
	}
	if payload.Error.Code != "invalid_operator_request" || payload.Error.ReasonCode != "unknown_command" {
		t.Fatalf("unexpected recovery error: %#v", payload.Error)
	}
	if len(payload.ArtifactRefs) != 0 {
		t.Fatalf("failure envelope must not fabricate artifacts: %#v", payload.ArtifactRefs)
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
			payload, exitCode := recoverycli.MapFailureKind(test.kind)
			if exitCode != test.exitCode {
				t.Fatalf("exit code got %d want %d", exitCode, test.exitCode)
			}
			if payload.Code != test.code || payload.ReasonCode != test.reasonCode {
				t.Fatalf("error got %#v", payload)
			}
		})
	}
}
