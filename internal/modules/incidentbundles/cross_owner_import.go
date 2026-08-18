package incidentbundles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
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

type ImportTransactionResult struct {
	IncidentID uuid.UUID
}

type ImportReadCapability interface {
	crossownertransaction.ReadCapability
	ValidateIncidentBundleImport(context.Context, uuid.UUID, jobs.Execution) error
}

type ImportWriteCapability interface {
	crossownertransaction.WriteCapability
	ApplyIncidentBundleImport(context.Context, *PreparedImport, ImportParams, uuid.UUID, string) (ImportTransactionResult, error)
}

type jobRunnableValidator interface {
	ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
}

// ImportTransactionProvider owns the Incident Bundles logical capability over
// the one physical transaction supplied by app composition.
type ImportTransactionProvider struct {
	importer Importer
	jobs     jobRunnableValidator
	now      func() time.Time
}

func NewImportTransactionProvider(pool *pgxpool.Pool, blobPort blobPortability, finalizer incidents.IncidentBundleImportFinalizer, projectionRebuild importProjectionRebuilder, historicalIntents historicalIntentPolicy, jobs jobRunnableValidator, now func() time.Time) (*ImportTransactionProvider, error) {
	if pool == nil || blobPort == nil || finalizer == nil || projectionRebuild == nil || historicalIntents == nil || jobs == nil {
		return nil, errors.New("incident bundle import transaction provider is incomplete")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &ImportTransactionProvider{importer: Importer{
		pool: pool, blobPort: blobPort, finalizer: finalizer, projectionRebuild: projectionRebuild,
		historicalIntents: historicalIntents,
	}, jobs: jobs, now: now}, nil
}

func (p *ImportTransactionProvider) TransactionCapabilities(participantID string, tx pgx.Tx) (crossownertransaction.ReadCapability, crossownertransaction.WriteCapability, error) {
	if p == nil || tx == nil || participantID != ImportTransactionParticipantID {
		return nil, nil, fmt.Errorf("%w: %s", crossownertransaction.ErrParticipantSet, participantID)
	}
	capability := &importTransactionCapability{
		participantID: participantID, tx: tx, importer: p.importer, jobs: p.jobs, now: p.now,
	}
	return capability, capability, nil
}

type importTransactionCapability struct {
	participantID string
	tx            pgx.Tx
	importer      Importer
	jobs          jobRunnableValidator
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
		return &VerificationError{ReasonCode: "duplicate_incident_id"}
	}
	if err := c.jobs.ValidateExecutionTx(ctx, c.tx, execution); err != nil {
		if errors.Is(err, jobs.ErrCancellationRequested) || errors.Is(err, jobs.ErrInvalidTransition) {
			return crossownertransaction.ErrCanceled
		}
		return err
	}
	return nil
}

func (c *importTransactionCapability) ApplyIncidentBundleImport(ctx context.Context, prepared *PreparedImport, params ImportParams, jobID uuid.UUID, manifestSHA string) (ImportTransactionResult, error) {
	if c == nil || c.tx == nil || c.now == nil || prepared == nil {
		return ImportTransactionResult{}, crossownertransaction.ErrUnavailable
	}
	incidentID, err := c.importer.ApplyPreparedImportTx(ctx, c.tx, prepared, params)
	if err != nil {
		return ImportTransactionResult{}, err
	}
	now := c.now().UTC()
	if err := MarkImportCompleteTx(ctx, c.tx, jobID, incidentID, manifestSHA, now); err != nil {
		return ImportTransactionResult{}, err
	}
	return ImportTransactionResult{IncidentID: incidentID}, nil
}

type ImportTransactionParticipant struct {
	prepared    *PreparedImport
	params      ImportParams
	execution   jobs.Execution
	manifestSHA string
}

func NewImportTransactionParticipant(prepared *PreparedImport, params ImportParams, execution jobs.Execution, manifestSHA string) (*ImportTransactionParticipant, error) {
	if prepared == nil || prepared.IncidentID == uuid.Nil || params.ActorUserID == uuid.Nil ||
		execution.JobID() == uuid.Nil || manifestSHA == "" {
		return nil, ErrPortabilityPayload
	}
	return &ImportTransactionParticipant{
		prepared: prepared, params: params, execution: execution, manifestSHA: manifestSHA,
	}, nil
}

func (p *ImportTransactionParticipant) ID() string { return ImportTransactionParticipantID }

func (p *ImportTransactionParticipant) BuildInput(context.Context, crossownertransaction.OperationContext) (crossownertransaction.Input, error) {
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

func (p *ImportTransactionParticipant) Prepare(context.Context, crossownertransaction.Invocation) (crossownertransaction.PrepareResult, error) {
	if p == nil || p.prepared == nil {
		return crossownertransaction.PrepareResult{}, ErrPortabilityPayload
	}
	return crossownertransaction.PrepareResult{SerializationKeys: []crossownertransaction.SerializationKey{{
		KeyKind: "incident_portability.incident", Key: p.prepared.IncidentID.String(),
	}}}, nil
}

func (p *ImportTransactionParticipant) Validate(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.ValidationResult, error) {
	capability, ok := invocation.ReadAccess.(ImportReadCapability)
	if !ok {
		return crossownertransaction.ValidationResult{}, crossownertransaction.ErrUnavailable
	}
	if err := capability.ValidateIncidentBundleImport(ctx, p.prepared.IncidentID, p.execution); err != nil {
		return crossownertransaction.ValidationResult{}, err
	}
	return crossownertransaction.Valid(), nil
}

func (p *ImportTransactionParticipant) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	capability, ok := invocation.WriteAccess.(ImportWriteCapability)
	if !ok {
		return crossownertransaction.WriteResult{}, crossownertransaction.ErrUnavailable
	}
	result, err := capability.ApplyIncidentBundleImport(ctx, p.prepared, p.params, p.execution.JobID(), p.manifestSHA)
	if err != nil {
		return crossownertransaction.WriteResult{}, err
	}
	return crossownertransaction.Written(result), nil
}
