package incidentbundles

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const (
	ImportTransactionParticipantID = "incident_portability.import_v2"
	importTransactionInputSchema   = "cartulary.incident_portability_import_transaction_input.v2"
)

func ImportTransactionDescriptor() crossownertransaction.Descriptor {
	descriptor := crossownertransaction.Descriptor{
		ParticipantID:         ImportTransactionParticipantID,
		OwnerProfileID:        ProfileID,
		InputSchemaID:         importTransactionInputSchema,
		PrepareAlgorithmID:    "incident_portability.import_prepare_v2",
		ValidationAlgorithmID: "incident_portability.import_validate_v2",
		WriteAlgorithmID:      "incident_portability.import_write_v2",
		SerializationKeyKinds: []string{"incident_portability.incident"},
		OwnedStateFamilyIDs:   []string{"core.incident", "incident_portability.job"},
	}
	descriptor.ContractSHA256 = crossownertransaction.DescriptorDigest(descriptor)
	return descriptor
}

// ValidateImportTransactionInput is the closed, side-effect-free schema
// resolver for the Core incident import participant.
func ValidateImportTransactionInput(value any) bool {
	object, ok := value.(map[string]any)
	if !ok || len(object) != 4 || object["schema_id"] != importTransactionInputSchema {
		return false
	}
	for _, key := range []string{"incident_id", "job_id"} {
		text, ok := object[key].(string)
		if !ok {
			return false
		}
		id, err := uuid.Parse(text)
		if err != nil || id == uuid.Nil || id.String() != text {
			return false
		}
	}
	digest, ok := object["manifest_sha256"].(string)
	if !ok || len(digest) != 64 || strings.ToLower(digest) != digest {
		return false
	}
	_, err := hex.DecodeString(digest)
	return err == nil
}

type importTransactionResult struct {
	IncidentID uuid.UUID
}

type importReadCapability interface {
	crossownertransaction.ReadCapability
	ValidateIncidentBundleImport(context.Context, uuid.UUID, jobs.Execution) error
}

type importWriteCapability interface {
	crossownertransaction.WriteCapability
	ApplyIncidentBundleImport(context.Context, *preparedImport, importParams, uuid.UUID, string) (importTransactionResult, error)
}

type importTransactionReader struct {
	tx   pgx.Tx
	jobs interface {
		ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
	}
}

type importTransactionWriter struct {
	tx       pgx.Tx
	importer importer
	now      func() time.Time
}

func (*importTransactionReader) ParticipantScope() string { return ImportTransactionParticipantID }
func (*importTransactionWriter) ParticipantScope() string { return ImportTransactionParticipantID }

func (c *importTransactionReader) ValidateIncidentBundleImport(ctx context.Context, incidentID uuid.UUID, execution jobs.Execution) error {
	if c == nil || c.tx == nil || c.jobs == nil || incidentID == uuid.Nil || execution.JobID() == uuid.Nil {
		return crossownertransaction.ErrUnavailable
	}
	var incidentExists bool
	if err := c.tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM incidents WHERE id = $1)`, incidentID).Scan(&incidentExists); err != nil {
		return err
	}
	if incidentExists {
		return &verificationError{ReasonCode: "duplicate_incident_id"}
	}
	if err := c.jobs.ValidateExecutionTx(ctx, c.tx, execution); err != nil {
		if errors.Is(err, jobs.ErrCancellationRequested) || errors.Is(err, jobs.ErrInvalidTransition) {
			return crossownertransaction.ErrCanceled
		}
		return err
	}
	return nil
}

func (c *importTransactionWriter) ApplyIncidentBundleImport(ctx context.Context, prepared *preparedImport, params importParams, jobID uuid.UUID, manifestSHA string) (importTransactionResult, error) {
	if c == nil || c.tx == nil || c.now == nil || prepared == nil {
		return importTransactionResult{}, crossownertransaction.ErrUnavailable
	}
	incidentID, err := c.importer.applyPreparedImportTx(ctx, c.tx, prepared, params)
	if err != nil {
		return importTransactionResult{}, err
	}
	now := c.now().UTC()
	if err := markImportCompleteTx(ctx, c.tx, jobID, incidentID, manifestSHA, now); err != nil {
		return importTransactionResult{}, err
	}
	return importTransactionResult{IncidentID: incidentID}, nil
}

type importTransactionParticipant struct {
	prepared    *preparedImport
	params      importParams
	execution   jobs.Execution
	manifestSHA string
}

func newImportTransactionParticipant(prepared *preparedImport, params importParams, execution jobs.Execution, manifestSHA string) (*importTransactionParticipant, error) {
	if prepared == nil || prepared.IncidentID == uuid.Nil || params.ActorUserID == uuid.Nil ||
		execution.JobID() == uuid.Nil || manifestSHA == "" {
		return nil, ErrPortabilityPayload
	}
	return &importTransactionParticipant{
		prepared: prepared, params: params, execution: execution, manifestSHA: manifestSHA,
	}, nil
}

func (p *importTransactionParticipant) ID() string { return ImportTransactionParticipantID }

func (p *importTransactionParticipant) BuildInput(context.Context, crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
	if p == nil || p.prepared == nil {
		return crossownertransaction.Input{}, ErrPortabilityPayload
	}
	payload, err := canonicaljson.MarshalExtension(map[string]any{
		"schema_id":       importTransactionInputSchema,
		"incident_id":     p.prepared.IncidentID.String(),
		"job_id":          p.execution.JobID().String(),
		"manifest_sha256": p.manifestSHA,
	})
	return crossownertransaction.Input{SchemaID: importTransactionInputSchema, CanonicalBytes: payload}, err
}

func (p *importTransactionParticipant) Prepare(context.Context, crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	if p == nil || p.prepared == nil {
		return crossownertransaction.PrepareResult{}, ErrPortabilityPayload
	}
	return crossownertransaction.PrepareResult{SerializationKeys: []crossownertransaction.SerializationKey{{
		KeyKind: "incident_portability.incident", Key: p.prepared.IncidentID.String(),
	}}}, nil
}

func (p *importTransactionParticipant) Validate(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.ValidationResult, error) {
	capability, ok := invocation.ReadAccess.(importReadCapability)
	if !ok {
		return crossownertransaction.ValidationResult{}, crossownertransaction.ErrUnavailable
	}
	if err := capability.ValidateIncidentBundleImport(ctx, p.prepared.IncidentID, p.execution); err != nil {
		return crossownertransaction.ValidationResult{}, err
	}
	return crossownertransaction.Valid(), nil
}

func (p *importTransactionParticipant) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	capability, ok := invocation.WriteAccess.(importWriteCapability)
	if !ok {
		return crossownertransaction.WriteResult{}, crossownertransaction.ErrUnavailable
	}
	result, err := capability.ApplyIncidentBundleImport(ctx, p.prepared, p.params, p.execution.JobID(), p.manifestSHA)
	if err != nil {
		return crossownertransaction.WriteResult{}, err
	}
	return crossownertransaction.Written(result), nil
}

// deadlineFailure is a deadline expiry whose commit is proven absent. Its Job
// terminal error is service_unavailable with the registered reason and the
// Extensions Table 26-B details operation_id and timeout_seconds.
type deadlineFailure struct {
	reasonCode     string
	timeoutSeconds int64
}

// importDeadlineFailure classifies only the two portability deadlines: the
// participant boundary (Table 9-E) and the cross-owner transaction (EXT-REQ-219).
// Fatal and cancellation outcomes are classified before this is consulted.
func importDeadlineFailure(cause error, transactions any) (deadlineFailure, bool) {
	var participant *PortabilityTimeoutError
	if errors.As(cause, &participant) {
		return deadlineFailure{reasonCode: "extension_participant_timeout", timeoutSeconds: participant.TimeoutSeconds}, true
	}
	if errors.Is(cause, crossownertransaction.ErrTimeout) {
		var seconds int64
		if timed, ok := transactions.(interface{ Timeout() time.Duration }); ok {
			seconds = int64(timed.Timeout() / time.Second)
		}
		return deadlineFailure{reasonCode: "extension_transaction_timeout", timeoutSeconds: seconds}, true
	}
	return deadlineFailure{}, false
}

func deadlineFailureCompletion(operationID uuid.UUID, failure deadlineFailure) jobs.FailureCompletion {
	return jobs.FailureCompletion{
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ErrorSummary: jobs.ErrorSummary{
			Code:      "service_unavailable",
			Message:   "service_unavailable",
			Retryable: true,
			Details: map[string]any{
				"reason_code":     failure.reasonCode,
				"operation_id":    operationID.String(),
				"timeout_seconds": failure.timeoutSeconds,
			},
		},
	}
}
