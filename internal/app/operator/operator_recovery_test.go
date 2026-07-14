package operator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/operatorcli"
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
			parsed := operatorcli.ParseCommand(test.args)
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
			parsed := operatorcli.ParseCommand(test.args)
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
	parsed := operatorcli.ParseCommand([]string{"backup-metadata", "latest"})
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
			parsed := operatorcli.ParseCommand([]string{"restore", "latest", "--target-config-file", targetPath, "--confirm-backup-set-id", backupID})
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
	var payload operatorcli.Result
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode recovery result: %v\nstdout=%s", err, stdout.String())
	}
	if payload.SchemaID != operatorcli.ResultSchemaID || payload.Result != "failed" || payload.Error == nil {
		t.Fatalf("unexpected recovery result: %#v", payload)
	}
	if payload.Error.Code != "invalid_operator_request" || payload.Error.ReasonCode != "unknown_command" {
		t.Fatalf("unexpected recovery error: %#v", payload.Error)
	}
	if len(payload.ArtifactRefs) != 0 {
		t.Fatalf("failure envelope must not fabricate artifacts: %#v", payload.ArtifactRefs)
	}
}

func TestOperatorRecoveryErrorMappingUsesClosedExitCodes(t *testing.T) {
	tests := []struct {
		name       string
		operation  string
		err        error
		code       string
		reasonCode string
		exitCode   int
	}{
		{
			name:       "timeout",
			operation:  "restore_latest",
			err:        context.DeadlineExceeded,
			code:       "operation_timed_out",
			reasonCode: "timeout_elapsed",
			exitCode:   4,
		},
		{
			name:       "lock contention",
			operation:  "restore_verify_due",
			err:        operatorcli.ErrOperationLockUnavailable,
			code:       "recovery_operation_in_progress",
			reasonCode: "operation_lock_unavailable",
			exitCode:   3,
		},
		{
			name:       "confirmation mismatch",
			operation:  "restore_latest",
			err:        operatorcli.ErrConfirmationMismatch,
			code:       "invalid_operator_request",
			reasonCode: "confirmation_mismatch",
			exitCode:   2,
		},
		{
			name:       "verification probe",
			operation:  "restore_verify_latest",
			err:        errors.New("workbook probe returned an error"),
			code:       "verification_failed",
			reasonCode: "workbook_probe_failed",
			exitCode:   4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload, exitCode := operatorcli.MapError(test.operation, test.err)
			if exitCode != test.exitCode {
				t.Fatalf("exit code got %d want %d", exitCode, test.exitCode)
			}
			if payload.Code != test.code || payload.ReasonCode != test.reasonCode {
				t.Fatalf("error got %#v", payload)
			}
		})
	}
}
