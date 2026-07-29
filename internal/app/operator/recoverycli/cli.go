package recoverycli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
)

const (
	ResultSchemaID   = "cartulary.operator_recovery_result.v1"
	ProgressSchemaID = "cartulary.operator_recovery_progress.v1"
)

type Result struct {
	SchemaID           string        `json:"schema_id"`
	OperationID        string        `json:"operation_id"`
	Operation          string        `json:"operation"`
	Result             string        `json:"result"`
	StartedAt          time.Time     `json:"started_at"`
	CompletedAt        time.Time     `json:"completed_at"`
	BackupSetID        *string       `json:"backup_set_id"`
	ConsistencyPointAt *time.Time    `json:"consistency_point_at"`
	ArtifactRefs       []ArtifactRef `json:"artifact_refs"`
	Error              *Error        `json:"error"`
}

type ArtifactRef struct {
	Kind        string  `json:"kind"`
	SchemaID    string  `json:"schema_id"`
	RefID       string  `json:"ref_id"`
	BackupSetID *string `json:"backup_set_id"`
}

type Error struct {
	Code       string `json:"code"`
	ReasonCode string `json:"reason_code"`
	Message    string `json:"message"`
}

type Progress struct {
	SchemaID    string    `json:"schema_id"`
	OperationID string    `json:"operation_id"`
	Phase       string    `json:"phase"`
	Completed   int       `json:"completed"`
	Total       *int      `json:"total"`
	EmittedAt   time.Time `json:"emitted_at"`
}

type Command struct {
	OperationID        string
	Handled            bool
	Invalid            bool
	Operation          string
	SourceConfigPath   string
	TargetConfigPath   string
	ConfirmBackupSetID string
	Output             string
	Progress           string
	TimeoutSeconds     int
	Err                *Error
}

type Runner struct {
	Stdout io.Writer
	Stderr io.Writer
	Now    func() time.Time
	Facade application.Facade
}

func (runner Runner) Run(ctx context.Context, args []string) (bool, int) {
	parsed := ParseCommand(args)
	if !parsed.Handled {
		return false, 0
	}

	now := runner.now()
	operationID := uuid.New()
	startedAt := now().UTC()
	parsed.OperationID = operationID.String()
	if parsed.Invalid {
		errPayload := parsed.Err
		if errPayload == nil {
			errPayload = ErrorPayload("invalid_operator_request", "unknown_command", "unsupported recovery command")
		}
		return true, runner.emitResult(Result{
			SchemaID:     ResultSchemaID,
			OperationID:  operationID.String(),
			Operation:    parsed.Operation,
			Result:       "failed",
			StartedAt:    startedAt,
			CompletedAt:  now().UTC(),
			ArtifactRefs: []ArtifactRef{},
			Error:        errPayload,
		}, 2)
	}

	runCtx := ctx
	cancel := func() {}
	if parsed.TimeoutSeconds > 0 && parsed.Operation != "restore_verify_due" {
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(parsed.TimeoutSeconds)*time.Second)
	}
	defer cancel()

	progress := ProgressEmitter{
		enabled:     parsed.Progress == "jsonl",
		writer:      normalizeWriter(runner.Stderr),
		operationID: operationID.String(),
		now:         now,
	}
	outcome, err := runner.runOperation(runCtx, operationID, parsed, progress)
	completedAt := now().UTC()
	if err != nil {
		kind, ok := application.FailureKindOf(err)
		if !ok {
			kind = fallbackFailureKind(parsed.Operation)
		}
		errPayload, exitCode := MapFailureKind(kind)
		backupSetID, consistencyPointAt, artifactRefs := wireResultFields(outcome)
		return true, runner.emitResult(Result{
			SchemaID:           ResultSchemaID,
			OperationID:        operationID.String(),
			Operation:          parsed.Operation,
			Result:             "failed",
			StartedAt:          startedAt,
			CompletedAt:        completedAt,
			BackupSetID:        backupSetID,
			ConsistencyPointAt: consistencyPointAt,
			ArtifactRefs:       sortedArtifactRefs(artifactRefs),
			Error:              errPayload,
		}, exitCode)
	}
	result := string(outcome.Status)
	if result == "" {
		result = "succeeded"
	}
	backupSetID, consistencyPointAt, artifactRefs := wireResultFields(outcome)
	return true, runner.emitResult(Result{
		SchemaID:           ResultSchemaID,
		OperationID:        operationID.String(),
		Operation:          parsed.Operation,
		Result:             result,
		StartedAt:          startedAt,
		CompletedAt:        completedAt,
		BackupSetID:        backupSetID,
		ConsistencyPointAt: consistencyPointAt,
		ArtifactRefs:       sortedArtifactRefs(artifactRefs),
		Error:              nil,
	}, 0)
}

func ParseCommand(args []string) Command {
	if len(args) == 0 {
		return Command{Handled: false}
	}
	if len(args) < 2 {
		switch args[0] {
		case "backup", "restore", "restore-verify":
			return invalidCommand("unknown", "unknown_command", "unsupported recovery command")
		default:
			return Command{Handled: false}
		}
	}
	switch {
	case args[0] == "backup" && args[1] == "inspect" && len(args) >= 3 && args[2] == "latest":
		return parseFlags("backup_inspect_latest", args[3:], false, false)
	case args[0] == "backup" && args[1] == "create":
		return parseFlags("backup_create", args[2:], false, false)
	case args[0] == "restore" && args[1] == "latest":
		return parseFlags("restore_latest", args[2:], true, true)
	case args[0] == "restore-verify" && args[1] == "latest":
		return parseFlags("restore_verify_latest", args[2:], true, false)
	case args[0] == "restore-verify" && args[1] == "due":
		return parseFlags("restore_verify_due", args[2:], true, false)
	case args[0] == "backup" || args[0] == "restore" || args[0] == "restore-verify":
		return invalidCommand("unknown", "unknown_command", "unsupported recovery command")
	default:
		return Command{Handled: false}
	}
}

func parseFlags(operation string, args []string, requiresTarget bool, requiresConfirm bool) Command {
	flags := flag.NewFlagSet("operator "+operation, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	output := flags.String("output", "json", "output mode")
	progress := flags.String("progress", "", "progress mode")
	timeoutRaw := flags.String("timeout-seconds", "", "operation timeout")
	sourceConfig := flags.String("source-config-file", "", "source config file")
	targetConfig := flags.String("target-config-file", "", "target config file")
	confirmBackupSetID := flags.String("confirm-backup-set-id", "", "confirmed backup set id")
	if err := flags.Parse(args); err != nil {
		return invalidCommand(operation, "invalid_flag_value", "invalid recovery flag")
	}
	if flags.NArg() != 0 {
		return invalidCommand(operation, "invalid_flag_value", "unexpected recovery argument")
	}
	if *output != "json" {
		return invalidCommand(operation, "unsupported_output_mode", "unsupported output mode")
	}
	if *progress != "" && *progress != "jsonl" {
		return invalidCommand(operation, "unsupported_progress_mode", "unsupported progress mode")
	}
	timeout, errPayload := ParseTimeout(operation, *timeoutRaw)
	if errPayload != nil {
		return Command{Handled: true, Invalid: true, Operation: operation, Err: errPayload}
	}
	target := strings.TrimSpace(*targetConfig)
	if requiresTarget {
		if target == "" {
			return invalidCommand(operation, "missing_required_flag", "target-config-file is required")
		}
		if err := ValidateTargetConfigPath(target); err != nil {
			return invalidCommand(operation, "invalid_flag_value", err.Error())
		}
	}
	confirm := strings.TrimSpace(*confirmBackupSetID)
	if requiresConfirm {
		if confirm == "" {
			return invalidCommand(operation, "missing_required_flag", "confirm-backup-set-id is required")
		}
		if _, err := uuid.Parse(confirm); err != nil {
			return invalidCommand(operation, "invalid_flag_value", "confirm-backup-set-id must be an exact UUID")
		}
	}
	return Command{
		Handled:            true,
		Operation:          operation,
		SourceConfigPath:   strings.TrimSpace(*sourceConfig),
		TargetConfigPath:   target,
		ConfirmBackupSetID: confirm,
		Output:             *output,
		Progress:           *progress,
		TimeoutSeconds:     timeout,
	}
}

func ParseTimeout(operation string, raw string) (int, *Error) {
	defaultValue, minimum, maximum := TimeoutBounds(operation)
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, ErrorPayload("invalid_operator_request", "invalid_flag_value", "timeout-seconds must be an integer")
	}
	if value < minimum {
		return 0, ErrorPayload("invalid_operator_request", "timeout_below_minimum", "timeout-seconds is below the operation minimum")
	}
	if value > maximum {
		return 0, ErrorPayload("invalid_operator_request", "timeout_above_maximum", "timeout-seconds is above the operation maximum")
	}
	return value, nil
}

func TimeoutBounds(operation string) (int, int, int) {
	switch operation {
	case "backup_inspect_latest":
		return 30, 1, 3600
	case "backup_create", "restore_latest", "restore_verify_latest", "restore_verify_due":
		return 14400, 60, 86400
	default:
		return 30, 1, 3600
	}
}

func ValidateTargetConfigPath(path string) error {
	if strings.ContainsRune(path, '\x00') {
		return errors.New("target-config-file must not contain NUL")
	}
	if path == "~" || strings.HasPrefix(path, "~/") || strings.Contains(path, "$") {
		return errors.New("target-config-file must be a literal absolute path")
	}
	if !filepath.IsAbs(path) {
		return errors.New("target-config-file must be absolute")
	}
	for _, part := range strings.Split(path, string(filepath.Separator)) {
		if part == "." || part == ".." {
			return errors.New("target-config-file must not contain . or .. segments")
		}
	}
	return nil
}

func invalidCommand(operation string, reason string, message string) Command {
	return Command{
		Handled:   true,
		Invalid:   true,
		Operation: operation,
		Err:       ErrorPayload("invalid_operator_request", reason, message),
	}
}

func ErrorPayload(code string, reasonCode string, message string) *Error {
	return &Error{Code: code, ReasonCode: reasonCode, Message: message}
}

type ProgressEmitter struct {
	enabled     bool
	writer      io.Writer
	operationID string
	now         func() time.Time
}

func (emitter ProgressEmitter) ReportProgress(progress application.Progress) {
	if !emitter.enabled {
		return
	}
	_ = json.NewEncoder(emitter.writer).Encode(Progress{
		SchemaID:    ProgressSchemaID,
		OperationID: emitter.operationID,
		Phase:       progress.Phase,
		Completed:   progress.Completed,
		Total:       progress.Total,
		EmittedAt:   emitter.now().UTC(),
	})
}

func (runner Runner) runOperation(ctx context.Context, operationID uuid.UUID, parsed Command, progress application.ProgressSink) (application.Result, error) {
	if runner.Facade == nil {
		return application.Result{}, application.NewFailure(
			fallbackFailureKind(parsed.Operation),
			errors.New("operator recovery facade is unavailable"),
		)
	}
	switch parsed.Operation {
	case "backup_inspect_latest":
		return runner.Facade.BackupInspectLatest(ctx, application.BackupInspectLatestRequest{
			OperationID:      operationID,
			SourceConfigPath: parsed.SourceConfigPath,
		}, progress)
	case "backup_create":
		return runner.Facade.BackupCreate(ctx, application.BackupCreateRequest{
			OperationID:      operationID,
			SourceConfigPath: parsed.SourceConfigPath,
		}, progress)
	case "restore_latest":
		confirmed, err := uuid.Parse(parsed.ConfirmBackupSetID)
		if err != nil {
			return application.Result{}, application.NewFailure(application.FailureConfirmationMismatch, err)
		}
		return runner.Facade.RestoreLatest(ctx, application.RestoreLatestRequest{
			OperationID:        operationID,
			SourceConfigPath:   parsed.SourceConfigPath,
			TargetConfigPath:   parsed.TargetConfigPath,
			ConfirmedBackupSet: confirmed,
		}, progress)
	case "restore_verify_latest":
		return runner.Facade.RestoreVerifyLatest(ctx, application.RestoreVerifyLatestRequest{
			OperationID:      operationID,
			SourceConfigPath: parsed.SourceConfigPath,
			TargetConfigPath: parsed.TargetConfigPath,
		}, progress)
	case "restore_verify_due":
		return runner.Facade.RestoreVerifyDue(ctx, application.RestoreVerifyDueRequest{
			OperationID:      operationID,
			SourceConfigPath: parsed.SourceConfigPath,
			TargetConfigPath: parsed.TargetConfigPath,
			AttemptTimeout:   time.Duration(parsed.TimeoutSeconds) * time.Second,
		}, progress)
	default:
		return application.Result{}, application.NewFailure(
			fallbackFailureKind(parsed.Operation),
			errors.New("unsupported recovery operation"),
		)
	}
}

type failureMapping struct {
	Code       string
	ReasonCode string
	Message    string
	ExitCode   int
}

func MapFailureKind(kind application.FailureKind) (*Error, int) {
	mapping, ok := failureMappingForKind(kind)
	if !ok {
		return ErrorPayload("invalid_operator_request", "unknown_command", "unsupported recovery failure"), 2
	}
	return ErrorPayload(mapping.Code, mapping.ReasonCode, mapping.Message), mapping.ExitCode
}

func FailureEvidenceFields(kind application.FailureKind) (string, string) {
	mapping, ok := failureMappingForKind(kind)
	if !ok {
		return "", ""
	}
	return mapping.Code, mapping.ReasonCode
}

func failureMappingForKind(kind application.FailureKind) (failureMapping, bool) {
	switch kind {
	case application.FailureConfirmationMismatch:
		return failureMapping{"invalid_operator_request", "confirmation_mismatch", "confirmed backup_set_id does not match latest retained backup", 2}, true
	case application.FailureLocalConfigInvalid:
		return failureMapping{"invalid_operator_request", "local_config_invalid", "local configuration is invalid", 2}, true
	case application.FailureSecretReferenceMissing:
		return failureMapping{"recovery_key_unavailable", "secret_reference_missing", "recovery master key is unavailable", 3}, true
	case application.FailureSecretReferenceUnresolved:
		return failureMapping{"recovery_key_unavailable", "secret_reference_unresolved", "recovery master key reference cannot be resolved", 3}, true
	case application.FailureRecoveryKeyInvalid:
		return failureMapping{"recovery_key_unavailable", "recovery_key_invalid", "recovery master key is invalid", 3}, true
	case application.FailureNoSuccessfulRetainedBackup:
		return failureMapping{"backup_set_not_found", "no_successful_retained_backup", "no successful retained backup is available", 3}, true
	case application.FailureSelectedBackupNotRetained:
		return failureMapping{"backup_set_not_found", "selected_backup_not_retained", "selected backup is not retained", 3}, true
	case application.FailureArtifactMissing:
		return failureMapping{"backup_integrity_failed", "artifact_missing", "backup artifact is unavailable", 3}, true
	case application.FailureIntegrityProofMissing:
		return failureMapping{"backup_integrity_failed", "integrity_proof_missing", "backup integrity proof is unavailable", 3}, true
	case application.FailureChecksumMismatch:
		return failureMapping{"backup_integrity_failed", "checksum_mismatch", "backup checksum verification failed", 3}, true
	case application.FailureAttestationInvalid:
		return failureMapping{"backup_integrity_failed", "attestation_invalid", "backup attestation is invalid", 3}, true
	case application.FailureSameDatabaseBinding:
		return failureMapping{"unsafe_restore_target", "same_database_binding", "restore target database binding is not distinct", 3}, true
	case application.FailureSameObjectStoreBinding:
		return failureMapping{"unsafe_restore_target", "same_object_store_binding", "restore target object-store binding is not distinct", 3}, true
	case application.FailureTargetDatabaseNotFresh:
		return failureMapping{"unsafe_restore_target", "target_database_not_fresh", "restore target database is not fresh", 3}, true
	case application.FailureTargetObjectNamespaceNotFresh:
		return failureMapping{"unsafe_restore_target", "target_object_namespace_not_fresh", "restore target object namespace is not fresh", 3}, true
	case application.FailureTargetServingTraffic:
		return failureMapping{"unsafe_restore_target", "target_serving_traffic", "restore target is serving traffic", 3}, true
	case application.FailureTargetMarkerMissing:
		return failureMapping{"unsafe_restore_target", "target_marker_missing", "restore target marker is missing", 3}, true
	case application.FailureTargetMarkerInvalid:
		return failureMapping{"unsafe_restore_target", "target_marker_invalid", "restore target marker is invalid", 3}, true
	case application.FailureOperationLockUnavailable:
		return failureMapping{"recovery_operation_in_progress", "operation_lock_unavailable", "another recovery operation is in progress", 3}, true
	case application.FailureTimeoutElapsed:
		return failureMapping{"operation_timed_out", "timeout_elapsed", "operation timed out", 4}, true
	case application.FailureBackupPostgres:
		return failureMapping{"backup_create_failed", "postgres_backup_failed", "PostgreSQL backup failed", 4}, true
	case application.FailureBackupObject:
		return failureMapping{"backup_create_failed", "object_backup_failed", "object-store backup failed", 4}, true
	case application.FailureBackupIntegrityProof:
		return failureMapping{"backup_create_failed", "integrity_proof_failed", "backup integrity proof creation failed", 4}, true
	case application.FailureBackupArtifactReadback:
		return failureMapping{"backup_create_failed", "artifact_readback_failed", "backup artifact readback failed", 4}, true
	case application.FailureBackupAttestationWrite:
		return failureMapping{"backup_create_failed", "attestation_write_failed", "backup attestation write failed", 4}, true
	case application.FailureBackupPublication:
		return failureMapping{"backup_create_failed", "backup_publication_failed", "backup publication failed", 4}, true
	case application.FailureBackupJournalWrite:
		return failureMapping{"backup_create_failed", "journal_write_failed", "backup terminal evidence write failed", 4}, true
	case application.FailureRestorePostgres:
		return failureMapping{"restore_failed", "postgres_restore_failed", "PostgreSQL restore failed", 4}, true
	case application.FailureRestoreObject:
		return failureMapping{"restore_failed", "object_restore_failed", "object-store restore failed", 4}, true
	case application.FailureRestoreProjectionRebuild:
		return failureMapping{"restore_failed", "projection_rebuild_failed", "restore projection rebuild failed", 4}, true
	case application.FailureRestoreInvariantCheck:
		return failureMapping{"restore_failed", "invariant_check_failed", "restore invariant check failed", 4}, true
	case application.FailureRestoreJournalWrite:
		return failureMapping{"restore_failed", "journal_write_failed", "restore terminal evidence write failed", 4}, true
	case application.FailureVerificationPostgres:
		return failureMapping{"verification_failed", "postgres_restore_failed", "verification PostgreSQL restore failed", 4}, true
	case application.FailureVerificationObject:
		return failureMapping{"verification_failed", "object_restore_failed", "verification object-store restore failed", 4}, true
	case application.FailureVerificationProjectionRebuild:
		return failureMapping{"verification_failed", "projection_rebuild_failed", "verification projection rebuild failed", 4}, true
	case application.FailureVerificationInvariantCheck:
		return failureMapping{"verification_failed", "invariant_check_failed", "restore verification invariant check failed", 4}, true
	case application.FailureVerificationWorkbookProbe:
		return failureMapping{"verification_failed", "workbook_probe_failed", "restore verification workbook probe failed", 4}, true
	case application.FailureVerificationAttestationUpdate:
		return failureMapping{"verification_failed", "attestation_update_failed", "restore verification attestation update failed", 4}, true
	case application.FailureVerificationJournalWrite:
		return failureMapping{"verification_failed", "journal_write_failed", "verification terminal evidence write failed", 4}, true
	default:
		return failureMapping{}, false
	}
}

func fallbackFailureKind(operation string) application.FailureKind {
	switch operation {
	case "backup_inspect_latest":
		return application.FailureArtifactMissing
	case "backup_create":
		return application.FailureBackupPublication
	case "restore_latest":
		return application.FailureRestoreInvariantCheck
	case "restore_verify_latest", "restore_verify_due":
		return application.FailureVerificationInvariantCheck
	default:
		return application.FailureLocalConfigInvalid
	}
}

func wireResultFields(result application.Result) (*string, *time.Time, []ArtifactRef) {
	var backupSetID *string
	if result.BackupSetID != nil {
		value := result.BackupSetID.String()
		backupSetID = &value
	}
	var consistencyPointAt *time.Time
	if result.ConsistencyPointAt != nil {
		value := result.ConsistencyPointAt.UTC()
		consistencyPointAt = &value
	}
	artifactRefs := make([]ArtifactRef, 0, len(result.ArtifactRefs))
	for _, ref := range result.ArtifactRefs {
		var refBackupSetID *string
		if ref.BackupSetID != nil {
			value := ref.BackupSetID.String()
			refBackupSetID = &value
		}
		artifactRefs = append(artifactRefs, ArtifactRef{
			Kind:        ref.Kind,
			SchemaID:    ref.SchemaID,
			RefID:       ref.RefID,
			BackupSetID: refBackupSetID,
		})
	}
	return backupSetID, consistencyPointAt, artifactRefs
}

func sortedArtifactRefs(refs []ArtifactRef) []ArtifactRef {
	if refs == nil {
		return []ArtifactRef{}
	}
	out := append([]ArtifactRef(nil), refs...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if artifactRefLess(out[j], out[i]) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func artifactRefLess(left ArtifactRef, right ArtifactRef) bool {
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	if left.SchemaID != right.SchemaID {
		return left.SchemaID < right.SchemaID
	}
	return left.RefID < right.RefID
}

func (runner Runner) emitResult(payload Result, exitCode int) int {
	encoder := json.NewEncoder(normalizeWriter(runner.Stdout))
	if err := encoder.Encode(payload); err != nil {
		return 4
	}
	return exitCode
}

func (runner Runner) now() func() time.Time {
	if runner.Now != nil {
		return runner.Now
	}
	return func() time.Time { return time.Now().UTC() }
}

func normalizeWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}
