package operator

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	operatorCollaborationRequeueResultSchemaID = "cartulary.operator.collaboration_requeue_result.v2"
	collaborationRequeueUsage                  = "operator collaboration requeue --incident-id <canonical-uuid> [--config <absolute-path>] [--timeout-seconds <seconds>]"
	defaultCollaborationRequeueTimeout         = 30 * time.Second
)

type operatorCollaborationRequeueResult struct {
	SchemaID            string                             `json:"schema_id"`
	OperationID         string                             `json:"operation_id"`
	Operation           string                             `json:"operation"`
	Result              string                             `json:"result"`
	StartedAt           string                             `json:"started_at"`
	CompletedAt         string                             `json:"completed_at"`
	IncidentID          *string                            `json:"incident_id"`
	RequeuedIntentCount *int                               `json:"requeued_intent_count"`
	Error               *operatorCollaborationRequeueError `json:"error"`
}

type operatorCollaborationRequeueError struct {
	Code       string `json:"code"`
	ReasonCode string `json:"reason_code"`
	Message    string `json:"message"`
}

type collaborationRequeueArgs struct {
	help        bool
	configPath  string
	incidentID  uuid.UUID
	hasIncident bool
	timeout     time.Duration
}

type collaborationRequeueParseFailure struct {
	reasonCode string
}

type collaborationRecoveryPort interface {
	RequeueIncident(context.Context, collaboration.RequeueRequest) (collaboration.RequeueResult, error)
}

type collaborationExecutor struct {
	transport       operatorTransport
	loadConfig      func(string) (configassembly.Loaded, error)
	setupPostgres   func(context.Context, postgres.Settings) (operatorPostgresPool, error)
	now             func() time.Time
	newOperationID  func() uuid.UUID
	newRecoveryPort func(postgres.DB) collaborationRecoveryPort
}

func newCollaborationRecoveryPort(db postgres.DB) collaborationRecoveryPort {
	return collaboration.NewRecoveryService(db)
}

func (executor collaborationExecutor) runCommand(ctx context.Context, args []string) int {
	operationID := executor.operationID()
	startedAt := executor.currentTime()
	parsed, parseFailure := parseCollaborationRequeueArgs(args[2:])
	if parsed.help {
		_, _ = fmt.Fprintf(normalizeOperatorWriter(executor.transport.stderr), "usage: %s\n", collaborationRequeueUsage)
		return 0
	}
	if parseFailure != nil {
		return executor.deliverCollaborationResult(collaborationFailureResult(
			operationID,
			startedAt,
			executor.currentTime(),
			collaborationIncidentID(parsed),
			"invalid_operator_request",
			parseFailure.reasonCode,
			collaborationRequeueMessage(parseFailure.reasonCode),
		), 2)
	}
	if failure := collaborationContextFailure(ctx, ctx); failure != nil {
		return executor.deliverMappedCollaborationFailure(operationID, startedAt, parsed, failure)
	}

	loaded, err := executor.loadConfig(parsed.configPath)
	if err != nil {
		failure := &operatorCollaborationRequeueError{
			Code:       "invalid_operator_request",
			ReasonCode: "local_config_invalid",
			Message:    collaborationRequeueMessage("local_config_invalid"),
		}
		return executor.deliverCollaborationFailure(operationID, startedAt, parsed, failure, 2)
	}
	settings, err := postgres.ResolveSettings(configassembly.PostgresBinding(loaded.Deployment()), nil)
	if err != nil {
		failure := &operatorCollaborationRequeueError{
			Code:       "invalid_operator_request",
			ReasonCode: "local_config_invalid",
			Message:    collaborationRequeueMessage("local_config_invalid"),
		}
		return executor.deliverCollaborationFailure(operationID, startedAt, parsed, failure, 2)
	}

	operationCtx, cancel := context.WithTimeout(ctx, parsed.timeout)
	defer cancel()
	pool, err := executor.setupPostgres(operationCtx, settings)
	if err != nil {
		if failure := collaborationContextFailure(operationCtx, ctx); failure != nil {
			return executor.deliverMappedCollaborationFailure(operationID, startedAt, parsed, failure)
		}
		failure := &operatorCollaborationRequeueError{
			Code:       "collaboration_requeue_failed",
			ReasonCode: "postgres_unavailable",
			Message:    collaborationRequeueMessage("postgres_unavailable"),
		}
		return executor.deliverCollaborationFailure(operationID, startedAt, parsed, failure, 4)
	}
	defer pool.Close()

	service := executor.newRecoveryPort(pool)
	result, err := service.RequeueIncident(operationCtx, collaboration.RequeueRequest{
		OperationID: operationID,
		IncidentID:  parsed.incidentID,
		MutatedAt:   executor.currentTime(),
	})
	if err != nil {
		failure, exitCode := mapCollaborationRequeueFailure(err, operationCtx, ctx)
		return executor.deliverCollaborationFailure(operationID, startedAt, parsed, failure, exitCode)
	}
	incidentID := parsed.incidentID.String()
	count := result.RequeuedIntentCount
	return executor.deliverCollaborationResult(operatorCollaborationRequeueResult{
		SchemaID:            operatorCollaborationRequeueResultSchemaID,
		OperationID:         operationID.String(),
		Operation:           "collaboration_requeue",
		Result:              "succeeded",
		StartedAt:           formatOperatorTimestamp(startedAt),
		CompletedAt:         formatOperatorTimestamp(executor.currentTime()),
		IncidentID:          &incidentID,
		RequeuedIntentCount: &count,
		Error:               nil,
	}, 0)
}

func (executor collaborationExecutor) deliverMappedCollaborationFailure(
	operationID uuid.UUID,
	startedAt time.Time,
	parsed collaborationRequeueArgs,
	failure *operatorCollaborationRequeueError,
) int {
	return executor.deliverCollaborationFailure(operationID, startedAt, parsed, failure, 4)
}

func (executor collaborationExecutor) deliverCollaborationFailure(
	operationID uuid.UUID,
	startedAt time.Time,
	parsed collaborationRequeueArgs,
	failure *operatorCollaborationRequeueError,
	exitCode int,
) int {
	return executor.deliverCollaborationResult(collaborationFailureResult(
		operationID,
		startedAt,
		executor.currentTime(),
		collaborationIncidentID(parsed),
		failure.Code,
		failure.ReasonCode,
		failure.Message,
	), exitCode)
}

func (executor collaborationExecutor) deliverCollaborationResult(result operatorCollaborationRequeueResult, exitCode int) int {
	if err := executor.transport.encodeJSON(result); err != nil {
		_, _ = fmt.Fprintf(
			normalizeOperatorWriter(executor.transport.stderr),
			"operation_id=%s result_delivery_failed\n",
			result.OperationID,
		)
		return 4
	}
	return exitCode
}

func collaborationFailureResult(
	operationID uuid.UUID,
	startedAt time.Time,
	completedAt time.Time,
	incidentID *string,
	code string,
	reasonCode string,
	message string,
) operatorCollaborationRequeueResult {
	return operatorCollaborationRequeueResult{
		SchemaID:            operatorCollaborationRequeueResultSchemaID,
		OperationID:         operationID.String(),
		Operation:           "collaboration_requeue",
		Result:              "failed",
		StartedAt:           formatOperatorTimestamp(startedAt),
		CompletedAt:         formatOperatorTimestamp(completedAt),
		IncidentID:          incidentID,
		RequeuedIntentCount: nil,
		Error: &operatorCollaborationRequeueError{
			Code:       code,
			ReasonCode: reasonCode,
			Message:    message,
		},
	}
}

func parseCollaborationRequeueArgs(args []string) (collaborationRequeueArgs, *collaborationRequeueParseFailure) {
	parsed := collaborationRequeueArgs{timeout: defaultCollaborationRequeueTimeout}
	if len(args) == 1 && args[0] == "--help" {
		parsed.help = true
		return parsed, nil
	}
	for _, arg := range args {
		if arg == "--help" {
			return parsed, &collaborationRequeueParseFailure{reasonCode: "unexpected_argument"}
		}
	}
	seen := make(map[string]struct{}, 3)
	for index := 0; index < len(args); index++ {
		token := args[index]
		if token == "--" || !strings.HasPrefix(token, "-") {
			return parsed, &collaborationRequeueParseFailure{reasonCode: "unexpected_argument"}
		}
		if !strings.HasPrefix(token, "--") || len(token) == 2 {
			return parsed, &collaborationRequeueParseFailure{reasonCode: "unknown_flag"}
		}
		nameValue := strings.TrimPrefix(token, "--")
		name, value, hasEquals := strings.Cut(nameValue, "=")
		switch name {
		case "incident-id", "config", "timeout-seconds":
		default:
			return parsed, &collaborationRequeueParseFailure{reasonCode: "unknown_flag"}
		}
		if _, duplicate := seen[name]; duplicate {
			return parsed, &collaborationRequeueParseFailure{reasonCode: "duplicate_flag"}
		}
		seen[name] = struct{}{}
		if !hasEquals {
			if index+1 >= len(args) || strings.HasPrefix(args[index+1], "--") {
				return parsed, &collaborationRequeueParseFailure{reasonCode: "invalid_flag_value"}
			}
			index++
			value = args[index]
		}
		if value == "" {
			return parsed, &collaborationRequeueParseFailure{reasonCode: "invalid_flag_value"}
		}
		switch name {
		case "incident-id":
			incidentID, err := uuid.Parse(value)
			if err != nil || incidentID == uuid.Nil || incidentID.String() != value {
				return parsed, &collaborationRequeueParseFailure{reasonCode: "invalid_flag_value"}
			}
			parsed.incidentID = incidentID
			parsed.hasIncident = true
		case "config":
			if !validExplicitOperatorConfigPath(value) {
				return parsed, &collaborationRequeueParseFailure{reasonCode: "invalid_flag_value"}
			}
			parsed.configPath = value
		case "timeout-seconds":
			seconds, err := strconv.Atoi(value)
			if err != nil || strconv.Itoa(seconds) != value || seconds < 1 || seconds > 300 {
				return parsed, &collaborationRequeueParseFailure{reasonCode: "invalid_flag_value"}
			}
			parsed.timeout = time.Duration(seconds) * time.Second
		}
	}
	if !parsed.hasIncident {
		return parsed, &collaborationRequeueParseFailure{reasonCode: "missing_required_flag"}
	}
	return parsed, nil
}

func validExplicitOperatorConfigPath(path string) bool {
	if !filepath.IsAbs(path) || strings.ContainsRune(path, '\x00') || strings.ContainsAny(path, "~$`") {
		return false
	}
	for _, segment := range strings.Split(path, string(filepath.Separator)) {
		if segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func mapCollaborationRequeueFailure(
	err error,
	operationCtx context.Context,
	callerCtx context.Context,
) (*operatorCollaborationRequeueError, int) {
	var serviceFailure *collaboration.RequeueFailure
	if errors.As(err, &serviceFailure) {
		switch serviceFailure.Kind {
		case collaboration.RequeueFailureIncidentNotQuarantined:
			return collaborationError("collaboration_requeue_rejected", "incident_not_quarantined"), 3
		case collaboration.RequeueFailureRepairNotVerified:
			return collaborationError("collaboration_requeue_rejected", "repair_not_verified"), 3
		case collaboration.RequeueFailureCommitOutcomeUnknown:
			return collaborationError("collaboration_requeue_failed", "commit_outcome_unknown"), 4
		case collaboration.RequeueFailureCancelled:
			return collaborationError("operation_cancelled", "caller_cancelled"), 4
		case collaboration.RequeueFailureTimedOut:
			return collaborationError("operation_timed_out", "timeout_elapsed"), 4
		case collaboration.RequeueFailureTransaction:
			if contextFailure := collaborationContextFailure(operationCtx, callerCtx); contextFailure != nil {
				return contextFailure, 4
			}
			return collaborationError("collaboration_requeue_failed", "transaction_failed"), 4
		}
	}
	if contextFailure := collaborationContextFailure(operationCtx, callerCtx); contextFailure != nil {
		return contextFailure, 4
	}
	return collaborationError("collaboration_requeue_failed", "transaction_failed"), 4
}

func collaborationContextFailure(operationCtx context.Context, callerCtx context.Context) *operatorCollaborationRequeueError {
	if errors.Is(callerCtx.Err(), context.Canceled) {
		return collaborationError("operation_cancelled", "caller_cancelled")
	}
	if errors.Is(callerCtx.Err(), context.DeadlineExceeded) {
		return collaborationError("operation_timed_out", "timeout_elapsed")
	}
	if errors.Is(operationCtx.Err(), context.DeadlineExceeded) {
		return collaborationError("operation_timed_out", "timeout_elapsed")
	}
	if errors.Is(operationCtx.Err(), context.Canceled) {
		return collaborationError("operation_cancelled", "caller_cancelled")
	}
	return nil
}

func collaborationError(code string, reasonCode string) *operatorCollaborationRequeueError {
	return &operatorCollaborationRequeueError{
		Code:       code,
		ReasonCode: reasonCode,
		Message:    collaborationRequeueMessage(reasonCode),
	}
}

func collaborationRequeueMessage(reasonCode string) string {
	switch reasonCode {
	case "missing_required_flag":
		return "Required operator input is missing."
	case "invalid_flag_value":
		return "An operator flag value is invalid."
	case "duplicate_flag":
		return "An operator flag was provided more than once."
	case "unknown_flag":
		return "An unknown operator flag was provided."
	case "unexpected_argument":
		return "An unexpected operator argument was provided."
	case "local_config_invalid":
		return "The local operator configuration is invalid."
	case "incident_not_quarantined":
		return "The collaboration incident is not quarantined."
	case "repair_not_verified":
		return "Pending collaboration payload repair could not be verified."
	case "postgres_unavailable":
		return "Postgres is unavailable for the operator action."
	case "transaction_failed":
		return "The collaboration requeue transaction failed."
	case "commit_outcome_unknown":
		return "The collaboration requeue commit outcome is unknown."
	case "timeout_elapsed":
		return "The collaboration requeue operation timed out."
	case "caller_cancelled":
		return "The collaboration requeue operation was cancelled."
	default:
		return "The collaboration requeue operation failed."
	}
}

func collaborationIncidentID(parsed collaborationRequeueArgs) *string {
	if !parsed.hasIncident {
		return nil
	}
	incidentID := parsed.incidentID.String()
	return &incidentID
}

func (executor collaborationExecutor) operationID() uuid.UUID {
	if executor.newOperationID != nil {
		if operationID := executor.newOperationID(); operationID != uuid.Nil {
			return operationID
		}
	}
	return uuid.New()
}

func (executor collaborationExecutor) currentTime() time.Time {
	if executor.now == nil {
		return time.Now().UTC()
	}
	return executor.now().UTC()
}

func formatOperatorTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
