package incidentbundles

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const (
	ImportTransactionParticipantID = "incident_portability.import_v1"
	importTransactionInputSchema   = "cartulary.incident_portability_import_transaction_input.v1"
)

func ImportTransactionDescriptor() crossownertransaction.Descriptor {
	return crossownertransaction.Descriptor{
		ParticipantID:         ImportTransactionParticipantID,
		OwnerProfileID:        ProfileID,
		ContractSHA256:        digestBytes([]byte("cartulary.incident_portability.import_transaction.v1")),
		InputSchemaID:         importTransactionInputSchema,
		PrepareAlgorithmID:    "incident_portability.import_prepare_v1",
		ValidationAlgorithmID: "incident_portability.import_validate_v1",
		WriteAlgorithmID:      "incident_portability.import_write_v1",
		SerializationKeyKinds: []string{"incident_portability.incident"},
		OwnedStateFamilyIDs:   []string{"core.incident", "incident_portability.job"},
	}
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

type importTransactionCapability struct {
	participantID string
	tx            pgx.Tx
	importer      importer
	jobs          JobTransactions
	now           func() time.Time
}

func (c *importTransactionCapability) ParticipantScope() string {
	if c == nil {
		return ""
	}
	return c.participantID
}

func (c *importTransactionCapability) ValidateIncidentBundleImport(ctx context.Context, incidentID uuid.UUID, execution jobs.Execution) error {
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

func (c *importTransactionCapability) ApplyIncidentBundleImport(ctx context.Context, prepared *preparedImport, params importParams, jobID uuid.UUID, manifestSHA string) (importTransactionResult, error) {
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
	payload, err := json.Marshal(map[string]any{
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
