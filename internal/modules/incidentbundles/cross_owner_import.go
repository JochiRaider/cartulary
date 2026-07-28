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
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
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
	ValidateIncidentBundleImport(context.Context, uuid.UUID, uuid.UUID) error
}

type ImportWriteCapability interface {
	crossownertransaction.WriteCapability
	ApplyIncidentBundleImport(context.Context, *PreparedImport, ImportParams, uuid.UUID, string) (ImportTransactionResult, error)
}

// ImportTransactionProvider owns the Incident Bundles logical capability over
// the one physical transaction supplied by app composition.
type ImportTransactionProvider struct {
	importer Importer
	now      func() time.Time
}

func NewImportTransactionProvider(pool *pgxpool.Pool, objects objectstore.Store, finalizer incidents.IncidentBundleImportFinalizer, projectionRebuild importProjectionRebuilder, historicalIntents historicalIntentPolicy, now func() time.Time) (*ImportTransactionProvider, error) {
	if pool == nil || objects == nil || finalizer == nil || projectionRebuild == nil || historicalIntents == nil {
		return nil, errors.New("incident bundle import transaction provider is incomplete")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &ImportTransactionProvider{importer: Importer{
		pool: pool, objectStore: objects, finalizer: finalizer, projectionRebuild: projectionRebuild,
		historicalIntents: historicalIntents,
	}, now: now}, nil
}

func (p *ImportTransactionProvider) TransactionCapabilities(participantID string, tx pgx.Tx) (crossownertransaction.ReadCapability, crossownertransaction.WriteCapability, error) {
	if p == nil || tx == nil || participantID != ImportTransactionParticipantID {
		return nil, nil, fmt.Errorf("%w: %s", crossownertransaction.ErrParticipantSet, participantID)
	}
	capability := &importTransactionCapability{
		participantID: participantID, tx: tx, importer: p.importer, now: p.now,
	}
	return capability, capability, nil
}

type importTransactionCapability struct {
	participantID string
	tx            pgx.Tx
	importer      Importer
	now           func() time.Time
}

func (c *importTransactionCapability) ParticipantScope() string {
	if c == nil {
		return ""
	}
	return c.participantID
}

func (c *importTransactionCapability) ValidateIncidentBundleImport(ctx context.Context, incidentID, jobID uuid.UUID) error {
	if c == nil || c.tx == nil || incidentID == uuid.Nil || jobID == uuid.Nil {
		return crossownertransaction.ErrUnavailable
	}
	var incidentExists bool
	if err := c.tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM incidents WHERE id = $1)`, incidentID).Scan(&incidentExists); err != nil {
		return err
	}
	if incidentExists {
		return &VerificationError{ReasonCode: "duplicate_incident_id"}
	}
	var jobState string
	if err := c.tx.QueryRow(ctx, `SELECT status FROM jobs WHERE job_id = $1 FOR UPDATE`, jobID).Scan(&jobState); err != nil {
		return err
	}
	if jobState != jobs.StatusRunning && jobState != jobs.StatusQueued {
		return crossownertransaction.ErrCanceled
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
	jobID       uuid.UUID
	manifestSHA string
}

func NewImportTransactionParticipant(prepared *PreparedImport, params ImportParams, jobID uuid.UUID, manifestSHA string) (*ImportTransactionParticipant, error) {
	if prepared == nil || prepared.IncidentID == uuid.Nil || params.ActorUserID == uuid.Nil ||
		jobID == uuid.Nil || manifestSHA == "" {
		return nil, ErrPortabilityPayload
	}
	return &ImportTransactionParticipant{
		prepared: prepared, params: params, jobID: jobID, manifestSHA: manifestSHA,
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
		"job_id":          p.jobID.String(),
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
	if err := capability.ValidateIncidentBundleImport(ctx, p.prepared.IncidentID, p.jobID); err != nil {
		return crossownertransaction.ValidationResult{}, err
	}
	return crossownertransaction.Valid(), nil
}

func (p *ImportTransactionParticipant) Write(ctx context.Context, invocation crossownertransaction.Invocation) (crossownertransaction.WriteResult, error) {
	capability, ok := invocation.WriteAccess.(ImportWriteCapability)
	if !ok {
		return crossownertransaction.WriteResult{}, crossownertransaction.ErrUnavailable
	}
	result, err := capability.ApplyIncidentBundleImport(ctx, p.prepared, p.params, p.jobID, p.manifestSHA)
	if err != nil {
		return crossownertransaction.WriteResult{}, err
	}
	return crossownertransaction.Written(result), nil
}
