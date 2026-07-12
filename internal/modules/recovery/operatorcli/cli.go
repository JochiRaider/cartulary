package operatorcli

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

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

const (
	ResultSchemaID   = "cartulary.operator_recovery_result.v1"
	ProgressSchemaID = "cartulary.operator_recovery_progress.v1"
)

var (
	ErrConfirmationMismatch     = errors.New("confirmation_mismatch")
	ErrOperationLockUnavailable = errors.New("operation_lock_unavailable")
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

type Outcome struct {
	BackupSetID        *string
	ConsistencyPointAt *time.Time
	ArtifactRefs       []ArtifactRef
	Result             string
}

type Operations interface {
	BackupInspectLatest(context.Context, Command, ProgressEmitter) (Outcome, error)
	BackupCreate(context.Context, Command, ProgressEmitter) (Outcome, error)
	RestoreLatest(context.Context, Command, ProgressEmitter) (Outcome, error)
	RestoreVerifyLatest(context.Context, Command, ProgressEmitter) (Outcome, error)
	RestoreVerifyDue(context.Context, Command, ProgressEmitter) (Outcome, error)
}

type Runner struct {
	Stdout     io.Writer
	Stderr     io.Writer
	Now        func() time.Time
	Operations Operations
}

func (runner Runner) Run(ctx context.Context, args []string) (bool, int) {
	parsed := ParseCommand(args)
	if !parsed.Handled {
		return false, 0
	}

	now := runner.now()
	operationID := uuid.NewString()
	startedAt := now().UTC()
	parsed.OperationID = operationID
	if parsed.Invalid {
		errPayload := parsed.Err
		if errPayload == nil {
			errPayload = ErrorPayload("invalid_operator_request", "unknown_command", "unsupported recovery command")
		}
		return true, runner.emitResult(Result{
			SchemaID:     ResultSchemaID,
			OperationID:  operationID,
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
	if parsed.TimeoutSeconds > 0 {
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(parsed.TimeoutSeconds)*time.Second)
	}
	defer cancel()

	progress := ProgressEmitter{
		enabled:     parsed.Progress == "jsonl",
		writer:      normalizeWriter(runner.Stderr),
		operationID: operationID,
		now:         now,
	}
	outcome, err := runner.runOperation(runCtx, parsed, progress)
	completedAt := now().UTC()
	if err != nil {
		errPayload, exitCode := MapError(parsed.Operation, err)
		return true, runner.emitResult(Result{
			SchemaID:           ResultSchemaID,
			OperationID:        operationID,
			Operation:          parsed.Operation,
			Result:             "failed",
			StartedAt:          startedAt,
			CompletedAt:        completedAt,
			BackupSetID:        outcome.BackupSetID,
			ConsistencyPointAt: outcome.ConsistencyPointAt,
			ArtifactRefs:       sortedArtifactRefs(outcome.ArtifactRefs),
			Error:              errPayload,
		}, exitCode)
	}
	result := outcome.Result
	if result == "" {
		result = "succeeded"
	}
	return true, runner.emitResult(Result{
		SchemaID:           ResultSchemaID,
		OperationID:        operationID,
		Operation:          parsed.Operation,
		Result:             result,
		StartedAt:          startedAt,
		CompletedAt:        completedAt,
		BackupSetID:        outcome.BackupSetID,
		ConsistencyPointAt: outcome.ConsistencyPointAt,
		ArtifactRefs:       sortedArtifactRefs(outcome.ArtifactRefs),
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

func (emitter ProgressEmitter) Emit(phase string, completed int, total *int) {
	if !emitter.enabled {
		return
	}
	_ = json.NewEncoder(emitter.writer).Encode(Progress{
		SchemaID:    ProgressSchemaID,
		OperationID: emitter.operationID,
		Phase:       phase,
		Completed:   completed,
		Total:       total,
		EmittedAt:   emitter.now().UTC(),
	})
}

func (runner Runner) runOperation(ctx context.Context, parsed Command, progress ProgressEmitter) (Outcome, error) {
	if runner.Operations == nil {
		return Outcome{}, errors.New("unsupported recovery operation")
	}
	switch parsed.Operation {
	case "backup_inspect_latest":
		return runner.Operations.BackupInspectLatest(ctx, parsed, progress)
	case "backup_create":
		return runner.Operations.BackupCreate(ctx, parsed, progress)
	case "restore_latest":
		return runner.Operations.RestoreLatest(ctx, parsed, progress)
	case "restore_verify_latest":
		return runner.Operations.RestoreVerifyLatest(ctx, parsed, progress)
	case "restore_verify_due":
		return runner.Operations.RestoreVerifyDue(ctx, parsed, progress)
	default:
		return Outcome{}, errors.New("unsupported recovery operation")
	}
}

func OutcomeForBackupSet(backupSet recovery.BackupSet, kind string, schemaID string) Outcome {
	outcome := OutcomeForStoredBackupSet(backupSet)
	outcome.ArtifactRefs = append(outcome.ArtifactRefs, ArtifactRefFor(kind, schemaID, kind+":"+backupSet.BackupSetID.String(), outcome.BackupSetID))
	return outcome
}

func OutcomeForStoredBackupSet(backupSet recovery.BackupSet) Outcome {
	if backupSet.BackupSetID == uuid.Nil {
		return Outcome{ArtifactRefs: []ArtifactRef{}}
	}
	backupSetID := backupSet.BackupSetID.String()
	consistencyPointAt := backupSet.ConsistencyPointAt
	return Outcome{
		BackupSetID:        &backupSetID,
		ConsistencyPointAt: &consistencyPointAt,
		ArtifactRefs:       []ArtifactRef{},
	}
}

func OutcomeForCandidate(backupSetID uuid.UUID, consistencyPointAt time.Time) Outcome {
	id := backupSetID.String()
	at := consistencyPointAt.UTC()
	return Outcome{BackupSetID: &id, ConsistencyPointAt: &at, ArtifactRefs: []ArtifactRef{}}
}

func ArtifactRefFor(kind string, schemaID string, refID string, backupSetID *string) ArtifactRef {
	return ArtifactRef{Kind: kind, SchemaID: schemaID, RefID: refID, BackupSetID: backupSetID}
}

func MapError(operation string, err error) (*Error, int) {
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorPayload("operation_timed_out", "timeout_elapsed", "operation timed out"), 4
	}
	if errors.Is(err, ErrConfirmationMismatch) {
		return ErrorPayload("invalid_operator_request", "confirmation_mismatch", "confirmed backup_set_id does not match latest retained backup"), 2
	}
	if errors.Is(err, ErrOperationLockUnavailable) {
		return ErrorPayload("recovery_operation_in_progress", "operation_lock_unavailable", "another recovery operation is in progress"), 3
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "recovery master key required"):
		return ErrorPayload("recovery_key_unavailable", "secret_reference_missing", "recovery master key is unavailable"), 3
	case strings.Contains(message, "recovery key") && (strings.Contains(message, "invalid") || strings.Contains(message, "parse")):
		return ErrorPayload("recovery_key_unavailable", "recovery_key_invalid", "recovery master key is invalid"), 3
	case strings.Contains(message, "operator recovery journal"):
		return ErrorPayload("journal_write_failed", "journal_append_failed", "operator recovery journal write failed"), 4
	case strings.Contains(message, "operator recovery audit"):
		return ErrorPayload("audit_write_failed", "audit_append_failed", "operator recovery audit write failed"), 4
	case strings.Contains(message, "load config"):
		return ErrorPayload("invalid_operator_request", "local_config_invalid", "local configuration is invalid"), 2
	case strings.Contains(message, "no successful") || strings.Contains(message, "backup set not found"):
		return ErrorPayload("backup_set_not_found", "no_successful_retained_backup", "no successful retained backup is available"), 3
	case strings.Contains(message, "checksum") || strings.Contains(message, "integrity") || strings.Contains(message, "artifact"):
		return ErrorPayload("backup_integrity_failed", "artifact_missing", "backup artifact or integrity proof is unavailable"), 3
	case strings.Contains(message, "source-config and target-config must be different files"):
		return ErrorPayload("unsafe_restore_target", "same_database_binding", "restore target database binding is not distinct"), 3
	case strings.Contains(message, "source and target postgres") && strings.Contains(message, "differ"):
		return ErrorPayload("unsafe_restore_target", "same_database_binding", "restore target database binding is not distinct"), 3
	case strings.Contains(message, "same") && strings.Contains(message, "postgres"):
		return ErrorPayload("unsafe_restore_target", "same_database_binding", "restore target database binding is not distinct"), 3
	case strings.Contains(message, "object store") && strings.Contains(message, "differ"):
		return ErrorPayload("unsafe_restore_target", "same_object_store_binding", "restore target object-store binding is not distinct"), 3
	case strings.Contains(message, "target database is not empty"):
		return ErrorPayload("unsafe_restore_target", "target_database_not_fresh", "restore target database is not fresh"), 3
	case strings.Contains(message, "target object store is not empty"):
		return ErrorPayload("unsafe_restore_target", "target_object_namespace_not_fresh", "restore target object namespace is not fresh"), 3
	case strings.Contains(message, "read target marker"):
		return ErrorPayload("unsafe_restore_target", "target_marker_missing", "restore target marker is missing"), 3
	case strings.Contains(message, "target marker"):
		return ErrorPayload("unsafe_restore_target", "target_marker_invalid", "restore target marker is invalid"), 3
	}
	switch operation {
	case "backup_create":
		return ErrorPayload("backup_create_failed", "backup_publication_failed", "backup creation failed"), 4
	case "restore_latest":
		return ErrorPayload("restore_failed", "invariant_check_failed", "restore failed"), 4
	case "restore_verify_latest", "restore_verify_due":
		if strings.Contains(message, "workbook probe") {
			return ErrorPayload("verification_failed", "workbook_probe_failed", "restore verification workbook probe failed"), 4
		}
		return ErrorPayload("verification_failed", "invariant_check_failed", "restore verification failed"), 4
	default:
		return ErrorPayload("invalid_operator_request", "unknown_command", "unsupported recovery command"), 2
	}
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

func IntPtr(value int) *int {
	return &value
}
