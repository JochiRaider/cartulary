package reporting

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var ErrNotFound = errors.New("reporting: not found")
var ErrStateConflict = errors.New("reporting: state conflict")
var ErrApprovalRejected = errors.New("reporting: approval rejected")

type StateConflictError struct {
	ReasonCode string
}

func (e *StateConflictError) Error() string {
	return ErrStateConflict.Error()
}

func (e *StateConflictError) Unwrap() error {
	return ErrStateConflict
}

type ApprovalRejectedError struct {
	ReasonCode string
}

func (e *ApprovalRejectedError) Error() string {
	return ErrApprovalRejected.Error()
}

func (e *ApprovalRejectedError) Unwrap() error {
	return ErrApprovalRejected
}

type InvalidReleaseRequestError struct {
	Field      string
	ReasonCode string
}

func (e *InvalidReleaseRequestError) Error() string {
	if e.ReasonCode == "" {
		return "invalid release request"
	}
	return "invalid release request: " + e.ReasonCode
}

type UnsupportedSnapshotDerivationError struct {
	DerivationVersion string
	SchemaID          string
}

func (e *UnsupportedSnapshotDerivationError) Error() string {
	if e.SchemaID != "" {
		return fmt.Sprintf("unsupported snapshot export model %q derivation %q", e.SchemaID, e.DerivationVersion)
	}
	return fmt.Sprintf("unsupported snapshot derivation version %q", e.DerivationVersion)
}

type Store struct {
	pool               *pgxpool.Pool
	jobTransactions    reportingJobTransactions
	exportMaterializer reportingExportMaterializer
	graphSources       *GraphSourceRegistry
	referencePacks     reference_data.Retention
	templatePacks      *TemplatePackRegistry
}

type SnapshotRecord struct {
	SnapshotID                   uuid.UUID
	IncidentID                   uuid.UUID
	CreatedByUserID              uuid.UUID
	ClientTxnID                  string
	SnapshotAt                   time.Time
	SourceChangeSetHighWatermark string
	SourceBoundaryJSON           []byte
	DerivationVersion            string
	SourceCaptureSHA256          string
	SourceCaptureJSON            []byte
	CreateJobID                  uuid.UUID
	CreatedAt                    time.Time
}

type ReleaseRecord struct {
	ReleaseID                    uuid.UUID
	IncidentID                   uuid.UUID
	SnapshotID                   uuid.UUID
	CreatedByUserID              uuid.UUID
	ClientTxnID                  string
	ReleaseScope                 string
	ReleaseState                 string
	SnapshotAt                   time.Time
	SourceChangeSetHighWatermark string
	DerivationVersion            string
	ExportModelSHA256            string
	TemplateID                   string
	TemplateVersion              string
	RedactionProfileID           string
	RedactionProfileVersion      string
	RedactionProfileSHA256       string
	OutputKind                   string
	OutputOptions                json.RawMessage
	GraphProjectionRefs          json.RawMessage
	CompositionID                *uuid.UUID
	CompositionVersion           *string
	CompositionSHA256            *string
	RenderAdmittedAt             time.Time
	OutputMediaType              *string
	OutputSHA256                 *string
	RedactionManifestSHA256      *string
	RedactionManifestJSON        []byte
	CreateJobID                  uuid.UUID
	RenderFailedReasonCode       *string
	RecipientPartitionRefs       []string
	ApprovedAt                   *time.Time
	PublishedAt                  *time.Time
	InvalidatedAt                *time.Time
	InvalidationReason           *string
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

type CreateSnapshotParams struct {
	ActorUserID uuid.UUID
	Request     CreateSnapshotRequest
	Now         time.Time
}

type CreateSnapshotResult struct {
	Job        jobs.Resource
	JobID      uuid.UUID
	SnapshotID uuid.UUID
	Replayed   bool
}

type ResolvedSourceBoundary struct {
	Token         string
	CanonicalJSON []byte
}

type RenderedRelease struct {
	Profile                 RedactionProfile
	ProfileSHA256           string
	Redaction               RedactionResult
	OutputMediaType         string
	OutputSHA256            string
	RenderBundle            RenderBundle
	RedactionManifestSHA256 string
	RedactionManifestJSON   []byte
}

type CreateReleaseParams struct {
	ActorUserID      uuid.UUID
	Request          CreateReleaseRequest
	TemplateContract TemplateContract
	TemplateFiles    []TemplatePackFile
	Now              time.Time
}

type CreateReleaseResult struct {
	Job       jobs.Resource
	JobID     uuid.UUID
	ReleaseID uuid.UUID
	Replayed  bool
}

type snapshotCreateJobPayload struct {
	JobAdmittedAt                time.Time       `json:"job_admitted_at"`
	RequestedBoundary            *string         `json:"requested_boundary"`
	IncidentID                   string          `json:"incident_id"`
	ActorUserID                  string          `json:"actor_user_id"`
	ClientTxnID                  string          `json:"client_txn_id"`
	SnapshotAt                   time.Time       `json:"snapshot_at"`
	SourceChangeSetHighWatermark string          `json:"source_change_set_high_watermark"`
	SourceBoundaryJSON           json.RawMessage `json:"source_boundary_json"`
	SourceCapture                json.RawMessage `json:"source_capture"`
	SourceCaptureSHA256          string          `json:"source_capture_sha256"`
}

type releaseCreateJobPayload struct {
	renderInput
	ClientTxnID       string               `json:"client_txn_id"`
	NormalizedRequest []byte               `json:"normalized_request"`
	Request           CreateReleaseRequest `json:"-"`
}

type compositionPreviewJobPayload struct {
	renderInput
	PreviewAttemptID  uuid.UUID `json:"preview_attempt_id"`
	NormalizedRequest []byte    `json:"normalized_request"`
}

type sourceProjectionRef struct {
	SourceOwnerID                 string `json:"source_owner_id"`
	GraphViewID                   string `json:"graph_view_id"`
	ProjectionResultID            string `json:"projection_result_id"`
	SourceSnapshotID              string `json:"source_snapshot_id"`
	ProjectionSchemaID            string `json:"projection_schema_id"`
	ProjectionVersion             string `json:"projection_version"`
	NormalizedConfigurationSHA256 string `json:"normalized_configuration_sha256"`
	NormalizedSourceSHA256        string `json:"normalized_source_sha256"`
	CanonicalOutputSHA256         string `json:"canonical_output_sha256"`
}

func (ref sourceProjectionRef) binding() graphprojection.ResultBindingV2 {
	return graphprojection.ResultBindingV2{
		ProjectionResultID:            ref.ProjectionResultID,
		GraphViewID:                   ref.GraphViewID,
		SourceOwnerID:                 ref.SourceOwnerID,
		SourceSnapshotID:              ref.SourceSnapshotID,
		ProjectionSchemaID:            ref.ProjectionSchemaID,
		ProjectionVersion:             ref.ProjectionVersion,
		NormalizedConfigurationSHA256: ref.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        ref.NormalizedSourceSHA256,
		CanonicalOutputSHA256:         ref.CanonicalOutputSHA256,
	}
}

type ApproveReleaseParams struct {
	ActorUserID       uuid.UUID
	ActorIncidentRole string
	ReleaseID         uuid.UUID
	Request           ReleaseActionRequest
	Now               time.Time
}

type ReleaseActionParams struct {
	ActorUserID uuid.UUID
	ReleaseID   uuid.UUID
	Request     ReleaseActionRequest
	Now         time.Time
}

type ReleaseActionResult struct {
	Payload    map[string]any
	IncidentID uuid.UUID
	Replayed   bool
}

func newStore(
	pool *pgxpool.Pool,
	jobTransactions reportingJobTransactions,
	exportMaterializer reportingExportMaterializer,
	graphSources *GraphSourceRegistry,
	referencePacks reference_data.Retention,
) *Store {
	return &Store{
		pool:               pool,
		jobTransactions:    jobTransactions,
		exportMaterializer: exportMaterializer,
		graphSources:       graphSources,
		referencePacks:     referencePacks,
	}
}

func (s *Store) CreateSnapshot(ctx context.Context, params CreateSnapshotParams) (CreateSnapshotResult, error) {
	return s.admitSnapshotJob(ctx, params)
}

type SnapshotBoundaryConflictError struct {
	Expected string
	Actual   string
}

func (e *SnapshotBoundaryConflictError) Error() string {
	return "snapshot source boundary conflict"
}

func (s *Store) GetSnapshot(ctx context.Context, snapshotID uuid.UUID) (map[string]any, uuid.UUID, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, uuid.Nil, err
	}
	defer rollbackSnapshotTx(tx)
	record, err := readSnapshotDescriptorTx(ctx, tx, snapshotID)
	if err != nil {
		return nil, uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, uuid.Nil, err
	}
	return snapshotResource(record), record.Incident, nil
}

func (s *Store) CreateRelease(ctx context.Context, params CreateReleaseParams) (CreateReleaseResult, error) {
	return s.admitReleaseJob(ctx, params)
}

func (s *Store) GetRelease(ctx context.Context, releaseID uuid.UUID) (map[string]any, uuid.UUID, error) {
	record, err := s.getReleaseRecord(ctx, releaseID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	return releaseResource(record), record.IncidentID, nil
}

// beginExecutionRead keeps the Jobs transition lock until all owned input and
// leased Graph bytes are copied. Rendering happens only after the transaction
// closes; it never retains a database lock across subprocess work.
func (s *Store) beginExecutionRead(ctx context.Context, execution jobs.Execution) (pgx.Tx, error) {
	if s == nil || s.pool == nil || s.jobTransactions == nil {
		return nil, jobs.ErrNotConfigured
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	if err := s.jobTransactions.ValidateExecutionTx(ctx, tx, execution); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}
	return tx, nil
}

func (s *Store) ReleasePayloadForJob(ctx context.Context, execution jobs.Execution) (releaseCreateJobPayload, error) {
	a, input, _, err := s.prepareRenderJob(ctx, execution)
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	return releasePayloadFromPrepared(a, input)
}

func (s *Store) ReportingJobKind(ctx context.Context, execution jobs.Execution) (string, error) {
	tx, err := s.beginExecutionRead(ctx, execution)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var kind string
	err = tx.QueryRow(ctx, `SELECT job_kind FROM reporting_job_payloads WHERE job_id=$1`, execution.JobID()).Scan(&kind)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return kind, nil
}

func (s *Store) CompositionPreviewPayloadForJob(ctx context.Context, execution jobs.Execution) (compositionPreviewJobPayload, error) {
	a, input, _, err := s.prepareRenderJob(ctx, execution)
	if err != nil {
		return compositionPreviewJobPayload{}, err
	}
	return previewPayloadFromPrepared(a, input)
}

func (s *Store) CompleteCompositionPreviewJobTx(
	ctx context.Context,
	tx pgx.Tx,
	jobID uuid.UUID,
	previewAttemptID uuid.UUID,
	candidate RenderCandidate,
	now time.Time,
) error {
	a, input, rendered, err := renderCandidateForCommitTx(ctx, tx, jobID, candidate)
	if err != nil {
		return err
	}
	payload, err := previewPayloadFromPrepared(a, input)
	if err != nil {
		return err
	}
	if payload.PreviewAttemptID != previewAttemptID {
		return errors.New("reporting: preview completion identity differs from admission")
	}
	if err := validateRetainedRenderFailureTx(ctx, tx, jobID, payload.renderInput, nil); err != nil {
		return err
	}
	stored, err := loadRetainedPreviewArtifactsTx(ctx, tx, payload, jobID)
	if err == nil {
		retained, err := jobs.ReadRetainedExtensionJob(ctx, tx, jobID)
		if err != nil {
			return err
		}
		if err := validatePreviewJobOutcome(retained); err != nil {
			return err
		}
		if !sameRetainedRenderArtifacts(stored, rendered) {
			return errors.New("reporting: repeated preview completion changed output")
		}
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if err := reportcomposition.LockPreviewAttemptForRenderTx(ctx, tx, previewAttemptID, jobID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO reporting_composition_preview_outputs (
    preview_attempt_id, render_attempt_id, release_scope, output_kind,
    output_media_type, output_sha256, redaction_profile_id, redaction_profile_version,
    redaction_profile_sha256, redaction_manifest_sha256, redaction_manifest_json,
    bundle_manifest_sha256, bundle_manifest_json, primary_bundle_path,
    primary_media_type, created_at
)
VALUES ($1, $2, 'internal_draft', $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`, previewAttemptID, jobID, rendered.RenderBundle.Manifest.OutputKind,
		rendered.OutputMediaType, rendered.OutputSHA256, rendered.Profile.ProfileID,
		rendered.Profile.Version, rendered.ProfileSHA256, rendered.RedactionManifestSHA256,
		rendered.RedactionManifestJSON, rendered.RenderBundle.ManifestSHA256,
		rendered.RenderBundle.ManifestJSON, rendered.RenderBundle.PrimaryPath,
		rendered.RenderBundle.PrimaryMedia, now.UTC()); err != nil {
		return err
	}
	for _, file := range rendered.RenderBundle.Files {
		if _, err := tx.Exec(ctx, `
INSERT INTO reporting_composition_preview_output_files (
    preview_attempt_id, bundle_path, role, media_type, file_sha256,
    size_bytes, storage_kind, object_ref, inline_bytes, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, $8, $9)
`, previewAttemptID, file.Path, file.Role, file.MediaType, file.SHA256,
			file.SizeBytes, file.StorageKind, file.Bytes, now.UTC()); err != nil {
			return err
		}
	}
	return insertPreviewRevealMapTx(ctx, tx, previewAttemptID, rendered, now)
}

func (s *Store) CompleteReleaseCreateJobTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, releaseID uuid.UUID, candidate RenderCandidate, now time.Time) error {
	a, input, rendered, err := renderCandidateForCommitTx(ctx, tx, jobID, candidate)
	if err != nil {
		return err
	}
	payload, err := releasePayloadFromPrepared(a, input)
	if err != nil {
		return err
	}
	if err := validateRetainedRenderFailureTx(ctx, tx, jobID, payload.renderInput, nil); err != nil {
		return err
	}
	if releaseID != reportingResourceID(jobID, "release") {
		return errors.New("reporting: completion identity changed")
	}
	if existing, err := getReleaseRecordByCreateJobIDTx(ctx, tx, jobID); err == nil {
		if err := validateReleaseAdmissionBinding(payload, jobID, existing); err != nil {
			return err
		}
		admission, err := jobs.ReadRetainedExtensionJob(ctx, tx, jobID)
		if err != nil {
			return err
		}
		if err := validateReleaseJobOutcome(existing, admission); err != nil {
			return err
		}
		if err := validateReleaseOutputBinding(existing, rendered); err != nil {
			return err
		}
		stored, err := loadRetainedReleaseArtifactsTx(ctx, tx, payload, existing)
		if err != nil {
			return err
		}
		if !sameRetainedRenderArtifacts(stored, rendered) {
			return errors.New("reporting: repeated release completion changed private artifact")
		}
		return nil
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}
	incidentID, snapshotID, actorID, err := payloadUUIDs(payload)
	if err != nil {
		return err
	}
	initialState := ReleaseStatePendingApproval
	var approvedAt *time.Time
	if payload.ReleaseScope == ReleaseScopeInternalDraft {
		initialState = ReleaseStateApproved
		approved := now.UTC()
		approvedAt = &approved
	}
	partitionJSON, err := canonicalStringArrayJSON(payload.RecipientPartitionRefs)
	if err != nil {
		return err
	}
	outputOptions := cloneRawMessageWithDefault(payload.OutputOptions, json.RawMessage(`{}`))
	graphProjectionRefs := cloneRawMessageWithDefault(payload.GraphProjectionRefs, json.RawMessage(`[]`))
	renderAdmittedAt := payload.RenderAdmittedAt.UTC()
	row, err := sqlc.New(tx).CreateReportingRelease(ctx, sqlc.CreateReportingReleaseParams{
		ReleaseID:                    pgUUID(releaseID),
		IncidentID:                   pgUUID(incidentID),
		SnapshotID:                   pgUUID(snapshotID),
		CreatedByUserID:              pgUUID(actorID),
		ClientTxnID:                  payload.ClientTxnID,
		ReleaseScope:                 payload.ReleaseScope,
		ReleaseState:                 initialState,
		SnapshotAt:                   pgTimestamptz(payload.SnapshotAt),
		SourceChangeSetHighWatermark: payload.SourceChangeSetHighWatermark,
		DerivationVersion:            payload.DerivationVersion,
		ExportModelSha256:            payload.ExportModelSHA256,
		TemplateID:                   payload.TemplateID,
		TemplateVersion:              payload.TemplateVersion,
		RedactionProfileID:           rendered.Profile.ProfileID,
		RedactionProfileVersion:      rendered.Profile.Version,
		RedactionProfileSha256:       rendered.ProfileSHA256,
		OutputKind:                   payload.OutputKind,
		OutputOptions:                []byte(outputOptions),
		GraphProjectionRefs:          []byte(graphProjectionRefs),
		CompositionID:                optionalPGUUIDString(payload.CompositionID),
		CompositionVersion:           optionalPGText(payload.CompositionVersion),
		CompositionSha256:            optionalPGText(payload.CompositionSHA256),
		RenderAdmittedAt:             pgTimestamptz(renderAdmittedAt),
		OutputMediaType:              requiredPGText(rendered.OutputMediaType),
		OutputSha256:                 requiredPGText(rendered.OutputSHA256),
		RedactionManifestSha256:      requiredPGText(rendered.RedactionManifestSHA256),
		RedactionManifestJson:        rendered.RedactionManifestJSON,
		CreateJobID:                  pgUUID(jobID),
		RecipientPartitionRefs:       partitionJSON,
		ApprovedAt:                   optionalPGTimestamptz(approvedAt),
		CreatedAt:                    pgTimestamptz(now),
	})
	if err != nil {
		return err
	}
	insertedReleaseID, err := uuidFromPG(row.ReleaseID)
	if err != nil {
		return err
	}
	if insertedReleaseID != releaseID {
		return fmt.Errorf("reporting release insert returned conflicting identity")
	}
	if err := insertRenderBundleTx(ctx, tx, releaseID, rendered.RenderBundle, now.UTC()); err != nil {
		return err
	}
	if err := insertReleaseRevealMapTx(ctx, tx, releaseID, rendered, now); err != nil {
		return err
	}
	if err := insertCompositionReleaseBindingTx(ctx, tx, releaseID, payload, now.UTC()); err != nil {
		return err
	}
	if err := invalidatePriorCandidateTx(ctx, tx, releaseID, snapshotID, payload.OutputKind, payload.ReleaseScope, payload.TemplateID, payload.TemplateVersion, rendered.Profile.ProfileID, rendered.Profile.Version, partitionJSON, outputOptions, graphProjectionRefs, payload.CompositionID, payload.CompositionVersion, payload.CompositionSHA256, now.UTC()); err != nil {
		return err
	}
	return nil
}

func insertRenderBundleTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, bundle RenderBundle, now time.Time) error {
	if bundle.ManifestSHA256 == "" || len(bundle.ManifestJSON) == 0 || bundle.PrimaryPath == "" || bundle.PrimaryMedia == "" {
		return fmt.Errorf("reporting render bundle is incomplete")
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO reporting_render_bundles (
    release_id, bundle_manifest_sha256, bundle_manifest_json,
    primary_bundle_path, primary_media_type, created_at
)
VALUES ($1, $2, $3, $4, $5, $6)
`, releaseID, bundle.ManifestSHA256, bundle.ManifestJSON, bundle.PrimaryPath, bundle.PrimaryMedia, now); err != nil {
		return err
	}
	for _, file := range bundle.Files {
		if file.Path == "" || file.Role == "" || file.MediaType == "" || file.SHA256 == "" {
			return fmt.Errorf("reporting render bundle file is incomplete")
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO reporting_render_bundle_files (
    release_id, bundle_path, role, media_type, file_sha256,
    size_bytes, storage_kind, inline_bytes, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, releaseID, file.Path, file.Role, file.MediaType, file.SHA256, file.SizeBytes, file.StorageKind, file.Bytes, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) completeReleaseRenderFailedJobTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, payload releaseCreateJobPayload, reasonCode string, now time.Time) (uuid.UUID, error) {
	profile, profileSHA := payload.RedactionProfile, payload.RedactionProfileSHA256
	incidentID, snapshotID, actorID, err := payloadUUIDs(payload)
	if err != nil {
		return uuid.UUID{}, err
	}
	partitionJSON, err := canonicalStringArrayJSON(payload.RecipientPartitionRefs)
	if err != nil {
		return uuid.UUID{}, err
	}
	outputOptions := cloneRawMessageWithDefault(payload.OutputOptions, json.RawMessage(`{}`))
	graphProjectionRefs := cloneRawMessageWithDefault(payload.GraphProjectionRefs, json.RawMessage(`[]`))
	renderAdmittedAt := payload.RenderAdmittedAt.UTC()
	if reasonCode == "" {
		return uuid.Nil, errors.New("reporting: failed release differs from admission")
	}
	if existing, err := getReleaseRecordByCreateJobIDTx(ctx, tx, jobID); err == nil {
		if err := validateReleaseAdmissionBinding(payload, jobID, existing); err != nil {
			return uuid.Nil, err
		}
		if existing.ReleaseState != ReleaseStateRenderFailed || !equalOptionalString(existing.RenderFailedReasonCode, &reasonCode) || existing.RedactionProfileSHA256 != profileSHA {
			return uuid.Nil, errors.New("reporting: repeated failure changed release outcome")
		}
		admission, err := jobs.ReadRetainedExtensionJob(ctx, tx, jobID)
		if err != nil {
			return uuid.Nil, err
		}
		if err := validateReleaseJobOutcome(existing, admission); err != nil {
			return uuid.Nil, err
		}
		if _, err := validateStoredReleaseOutputTx(ctx, tx, payload, existing); err != nil {
			return uuid.Nil, err
		}
		return existing.ReleaseID, nil
	} else if !errors.Is(err, ErrNotFound) {
		return uuid.Nil, err
	}
	row, err := sqlc.New(tx).CreateRenderFailedReportingRelease(ctx, sqlc.CreateRenderFailedReportingReleaseParams{
		ReleaseID:                    pgUUID(reportingResourceID(jobID, "release")),
		IncidentID:                   pgUUID(incidentID),
		SnapshotID:                   pgUUID(snapshotID),
		CreatedByUserID:              pgUUID(actorID),
		ClientTxnID:                  payload.ClientTxnID,
		ReleaseScope:                 payload.ReleaseScope,
		SnapshotAt:                   pgTimestamptz(payload.SnapshotAt),
		SourceChangeSetHighWatermark: payload.SourceChangeSetHighWatermark,
		DerivationVersion:            payload.DerivationVersion,
		ExportModelSha256:            payload.ExportModelSHA256,
		TemplateID:                   payload.TemplateID,
		TemplateVersion:              payload.TemplateVersion,
		RedactionProfileID:           profile.ProfileID,
		RedactionProfileVersion:      profile.Version,
		RedactionProfileSha256:       profileSHA,
		OutputKind:                   payload.OutputKind,
		OutputOptions:                []byte(outputOptions),
		GraphProjectionRefs:          []byte(graphProjectionRefs),
		CompositionID:                optionalPGUUIDString(payload.CompositionID),
		CompositionVersion:           optionalPGText(payload.CompositionVersion),
		CompositionSha256:            optionalPGText(payload.CompositionSHA256),
		RenderAdmittedAt:             pgTimestamptz(renderAdmittedAt),
		CreateJobID:                  pgUUID(jobID),
		RenderFailedReasonCode:       requiredPGText(reasonCode),
		RecipientPartitionRefs:       partitionJSON,
		CreatedAt:                    pgTimestamptz(now),
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	releaseID, err := uuidFromPG(row.ReleaseID)
	if err != nil {
		return uuid.UUID{}, err
	}
	if err := insertCompositionReleaseBindingTx(ctx, tx, releaseID, payload, now.UTC()); err != nil {
		return uuid.UUID{}, err
	}
	return releaseID, nil
}

func (s *Store) ApproveRelease(ctx context.Context, params ApproveReleaseParams) (ReleaseActionResult, error) {
	return s.withReleaseAction(ctx, "releases.approve", params.ActorUserID, params.ReleaseID, params.Request, func(ctx context.Context, tx pgx.Tx, release ReleaseRecord) (ReleaseRecord, error) {
		if release.ReleaseState == ReleaseStatePublished {
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "already_published"}
		}
		if release.ReleaseState == ReleaseStateInvalidated {
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "already_invalidated"}
		}
		if release.ReleaseState == ReleaseStateApproved {
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "already_approved"}
		}
		if release.ReleaseState == ReleaseStateRenderFailed {
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "render_failed"}
		}
		if release.ReleaseState != ReleaseStatePendingApproval {
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "approval_not_available"}
		}
		approvalRole, ok := approvalRoleForIncidentRole(params.ActorIncidentRole)
		if !ok {
			return ReleaseRecord{}, &ApprovalRejectedError{ReasonCode: "actor_lacks_approval_role"}
		}
		if release.ReleaseScope == ReleaseScopeInternalReview && approvalRole != "reviewer" {
			return ReleaseRecord{}, &ApprovalRejectedError{ReasonCode: "reviewer_approval_required"}
		}
		tupleJSON, err := approvalTupleJSON(release, params.ActorUserID, approvalRole)
		if err != nil {
			return ReleaseRecord{}, err
		}
		exists, err := approvalExistsTx(ctx, tx, release.ReleaseID, params.ActorUserID, approvalRole)
		if err != nil {
			return ReleaseRecord{}, err
		}
		if exists {
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "already_approved"}
		}
		if err := sqlc.New(tx).InsertReportingReleaseApproval(ctx, sqlc.InsertReportingReleaseApprovalParams{
			ReleaseID:               pgUUID(release.ReleaseID),
			ActorUserID:             pgUUID(params.ActorUserID),
			ApprovalRole:            approvalRole,
			Reason:                  optionalPGText(params.Request.Reason),
			ApprovalTupleJson:       tupleJSON,
			RedactionProfileSha256:  release.RedactionProfileSHA256,
			OutputSha256:            derefString(release.OutputSHA256),
			RedactionManifestSha256: derefString(release.RedactionManifestSHA256),
			CreatedAt:               pgTimestamptz(params.Now),
		}); err != nil {
			return ReleaseRecord{}, err
		}
		ready, err := approvalsSatisfiedTx(ctx, tx, release.ReleaseID, release.ReleaseScope)
		if err != nil {
			return ReleaseRecord{}, err
		}
		if !ready {
			return getReleaseRecordTx(ctx, tx, release.ReleaseID)
		}
		return updateReleaseStateTx(ctx, tx, release.ReleaseID, ReleaseStateApproved, params.Now.UTC(), "approved_at")
	}, params.Now)
}

func (s *Store) PublishRelease(ctx context.Context, params ReleaseActionParams) (ReleaseActionResult, error) {
	return s.withReleaseAction(ctx, "releases.publish", params.ActorUserID, params.ReleaseID, params.Request, func(ctx context.Context, tx pgx.Tx, release ReleaseRecord) (ReleaseRecord, error) {
		switch release.ReleaseState {
		case ReleaseStatePublished:
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "already_published"}
		case ReleaseStateInvalidated:
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "already_invalidated"}
		case ReleaseStatePendingApproval:
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "approval_required"}
		case ReleaseStateRenderFailed:
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "render_failed"}
		case ReleaseStateApproved:
			return updateReleaseStateTx(ctx, tx, release.ReleaseID, ReleaseStatePublished, params.Now.UTC(), "published_at")
		default:
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "invalid_state"}
		}
	}, params.Now)
}

func (s *Store) InvalidateRelease(ctx context.Context, params ReleaseActionParams) (ReleaseActionResult, error) {
	return s.withReleaseAction(ctx, "releases.invalidate", params.ActorUserID, params.ReleaseID, params.Request, func(ctx context.Context, tx pgx.Tx, release ReleaseRecord) (ReleaseRecord, error) {
		switch release.ReleaseState {
		case ReleaseStateInvalidated:
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "already_invalidated"}
		case ReleaseStateRenderFailed:
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "render_failed"}
		case ReleaseStatePendingApproval, ReleaseStateApproved, ReleaseStatePublished:
		default:
			return ReleaseRecord{}, &StateConflictError{ReasonCode: "invalid_state"}
		}
		reason := params.Request.Reason
		return invalidateReleaseTx(ctx, tx, release.ReleaseID, params.Now.UTC(), reason)
	}, params.Now)
}

func (s *Store) withReleaseAction(
	ctx context.Context,
	routeKey string,
	actorUserID uuid.UUID,
	releaseID uuid.UUID,
	request ReleaseActionRequest,
	mutate func(context.Context, pgx.Tx, ReleaseRecord) (ReleaseRecord, error),
	now time.Time,
) (ReleaseActionResult, error) {
	sum := sha256.Sum256(request.Normalized)
	requestHash := sum[:]
	key := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actorUserID,
		ScopeKey:    releaseID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ReleaseActionResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return ReleaseActionResult{}, authn.ErrClientTxnConflict
		}
		var payload map[string]any
		if err := json.Unmarshal(existing.ResponseJSON, &payload); err != nil {
			return ReleaseActionResult{}, err
		}
		release, err := getReleaseRecordTx(ctx, tx, releaseID)
		if err != nil {
			return ReleaseActionResult{}, err
		}
		return ReleaseActionResult{Payload: payload, IncidentID: release.IncidentID, Replayed: true}, tx.Commit(ctx)
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return ReleaseActionResult{}, err
	}
	release, err := getReleaseRecordForUpdateTx(ctx, tx, releaseID)
	if err != nil {
		return ReleaseActionResult{}, err
	}
	updated, err := mutate(ctx, tx, release)
	if err != nil {
		return ReleaseActionResult{}, err
	}
	payload := releaseResource(updated)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload); err != nil {
		return ReleaseActionResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ReleaseActionResult{}, err
	}
	_ = now
	return ReleaseActionResult{Payload: payload, IncidentID: updated.IncidentID}, nil
}

func (s *Store) getReleaseRecord(ctx context.Context, releaseID uuid.UUID) (ReleaseRecord, error) {
	row, err := sqlc.New(s.pool).GetReportingRelease(ctx, pgUUID(releaseID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ReleaseRecord{}, ErrNotFound
	}
	if err != nil {
		return ReleaseRecord{}, err
	}
	return releaseRecordFromSQL(row)
}

func getSnapshotRecordTx(ctx context.Context, tx pgx.Tx, snapshotID uuid.UUID) (SnapshotRecord, error) {
	row, err := sqlc.New(tx).GetReportingSnapshot(ctx, pgUUID(snapshotID))
	if errors.Is(err, pgx.ErrNoRows) {
		return SnapshotRecord{}, ErrNotFound
	}
	if err != nil {
		return SnapshotRecord{}, err
	}
	return snapshotRecordFromSQL(row)
}

func getReleaseRecordByCreateJobIDTx(ctx context.Context, tx pgx.Tx, createJobID uuid.UUID) (ReleaseRecord, error) {
	row, err := sqlc.New(tx).GetReportingReleaseByCreateJob(ctx, pgUUID(createJobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ReleaseRecord{}, ErrNotFound
	}
	if err != nil {
		return ReleaseRecord{}, err
	}
	releaseID, err := uuidFromPG(row.ReleaseID)
	if err != nil {
		return ReleaseRecord{}, err
	}
	return getReleaseRecordTx(ctx, tx, releaseID)
}

func getReleaseRecordTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID) (ReleaseRecord, error) {
	row, err := sqlc.New(tx).GetReportingRelease(ctx, pgUUID(releaseID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ReleaseRecord{}, ErrNotFound
	}
	if err != nil {
		return ReleaseRecord{}, err
	}
	return releaseRecordFromSQL(row)
}

func getReleaseRecordForUpdateTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID) (ReleaseRecord, error) {
	row, err := sqlc.New(tx).GetReportingReleaseForUpdate(ctx, pgUUID(releaseID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ReleaseRecord{}, ErrNotFound
	}
	if err != nil {
		return ReleaseRecord{}, err
	}
	return releaseRecordFromSQL(row)
}

func resolveReleaseCompositionTupleTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request CreateReleaseRequest) (*reportcomposition.ResolvedReleaseTuple, error) {
	if request.CompositionID == nil {
		return nil, nil
	}
	resolved, err := reportcomposition.ResolveReleaseTupleTx(ctx, tx, incidentID, request.TemplateID, request.TemplateVersion, reportcomposition.ReleaseTuple{
		CompositionID:      *request.CompositionID,
		CompositionVersion: derefString(request.CompositionVersion),
		CompositionSHA256:  derefString(request.CompositionSHA256),
	})
	if err != nil {
		var tupleErr *reportcomposition.ReleaseTupleError
		if errors.As(err, &tupleErr) {
			return nil, &InvalidReleaseRequestError{Field: tupleErr.Field, ReasonCode: tupleErr.ReasonCode}
		}
		return nil, err
	}
	return &resolved, nil
}

func invalidGraphProjectionRef(reasonCode string) error {
	return &InvalidReleaseRequestError{Field: "graph_projection_refs", ReasonCode: reasonCode}
}

func decodeSourceProjectionRefs(raw json.RawMessage) ([]sourceProjectionRef, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var refs []sourceProjectionRef
	if err := json.Unmarshal(raw, &refs); err != nil {
		return nil, err
	}
	for i, ref := range refs {
		if strings.TrimSpace(ref.SourceOwnerID) == "" ||
			strings.TrimSpace(ref.GraphViewID) == "" ||
			strings.TrimSpace(ref.ProjectionResultID) == "" ||
			strings.TrimSpace(ref.SourceSnapshotID) == "" ||
			strings.TrimSpace(ref.ProjectionSchemaID) == "" ||
			strings.TrimSpace(ref.ProjectionVersion) == "" ||
			!sha256HexPattern.MatchString(ref.NormalizedConfigurationSHA256) ||
			!sha256HexPattern.MatchString(ref.NormalizedSourceSHA256) ||
			!sha256HexPattern.MatchString(ref.CanonicalOutputSHA256) {
			return nil, fmt.Errorf("invalid source projection ref at %d", i)
		}
	}
	return refs, nil
}

func insertCompositionReleaseBindingTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, payload releaseCreateJobPayload, now time.Time) error {
	if payload.CompositionID == nil {
		return nil
	}
	compositionID, err := uuid.Parse(*payload.CompositionID)
	if err != nil {
		return err
	}
	return reportcomposition.BindReleaseTupleTx(ctx, tx, releaseID, reportcomposition.ReleaseTuple{
		CompositionID:      compositionID,
		CompositionVersion: derefString(payload.CompositionVersion),
		CompositionSHA256:  derefString(payload.CompositionSHA256),
	}, payload.ReleaseScope, now.UTC())
}

func snapshotRecordFromSQL(r sqlc.GetReportingSnapshotRow) (SnapshotRecord, error) {
	return snapshotRecordFromSQLFields(r.SnapshotID, r.IncidentID, r.CreatedByUserID, r.ClientTxnID, r.SnapshotAt, r.SourceChangeSetHighWatermark, r.SourceBoundaryJson, r.DerivationVersion, r.SourceCaptureSha256, r.SourceCaptureJson, r.CreateJobID, r.CreatedAt)
}

func snapshotRecordFromSQLFields(
	rowSnapshotID pgtype.UUID,
	rowIncidentID pgtype.UUID,
	rowCreatedByUserID pgtype.UUID,
	rowClientTxnID string,
	rowSnapshotAt pgtype.Timestamptz,
	rowSourceChangeSetHighWatermark string,
	rowSourceBoundaryJSON []byte,
	rowDerivationVersion string,
	rowSourceCaptureSHA256 string,
	rowSourceCaptureJSON []byte,
	rowCreateJobID pgtype.UUID,
	rowCreatedAt pgtype.Timestamptz,
) (SnapshotRecord, error) {
	snapshotID, err := uuidFromPG(rowSnapshotID)
	if err != nil {
		return SnapshotRecord{}, err
	}
	incidentID, err := uuidFromPG(rowIncidentID)
	if err != nil {
		return SnapshotRecord{}, err
	}
	createdBy, err := uuidFromPG(rowCreatedByUserID)
	if err != nil {
		return SnapshotRecord{}, err
	}
	createJobID, err := uuidFromPG(rowCreateJobID)
	if err != nil {
		return SnapshotRecord{}, err
	}
	snapshotAt, err := timeFromPG(rowSnapshotAt)
	if err != nil {
		return SnapshotRecord{}, err
	}
	createdAt, err := timeFromPG(rowCreatedAt)
	if err != nil {
		return SnapshotRecord{}, err
	}
	return SnapshotRecord{
		SnapshotID:                   snapshotID,
		IncidentID:                   incidentID,
		CreatedByUserID:              createdBy,
		ClientTxnID:                  rowClientTxnID,
		SnapshotAt:                   snapshotAt,
		SourceChangeSetHighWatermark: rowSourceChangeSetHighWatermark,
		SourceBoundaryJSON:           append([]byte(nil), rowSourceBoundaryJSON...),
		DerivationVersion:            rowDerivationVersion,
		SourceCaptureSHA256:          rowSourceCaptureSHA256,
		SourceCaptureJSON:            append([]byte(nil), rowSourceCaptureJSON...),
		CreateJobID:                  createJobID,
		CreatedAt:                    createdAt,
	}, nil
}

func releaseRecordFromSQL(row sqlc.ReportingRelease) (ReleaseRecord, error) {
	releaseID, err := uuidFromPG(row.ReleaseID)
	if err != nil {
		return ReleaseRecord{}, err
	}
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return ReleaseRecord{}, err
	}
	snapshotID, err := uuidFromPG(row.SnapshotID)
	if err != nil {
		return ReleaseRecord{}, err
	}
	createdBy, err := uuidFromPG(row.CreatedByUserID)
	if err != nil {
		return ReleaseRecord{}, err
	}
	createJobID, err := uuidFromPG(row.CreateJobID)
	if err != nil {
		return ReleaseRecord{}, err
	}
	snapshotAt, err := timeFromPG(row.SnapshotAt)
	if err != nil {
		return ReleaseRecord{}, err
	}
	createdAt, err := timeFromPG(row.CreatedAt)
	if err != nil {
		return ReleaseRecord{}, err
	}
	updatedAt, err := timeFromPG(row.UpdatedAt)
	if err != nil {
		return ReleaseRecord{}, err
	}
	renderAdmittedAt, err := timeFromPG(row.RenderAdmittedAt)
	if err != nil {
		return ReleaseRecord{}, err
	}
	return ReleaseRecord{
		ReleaseID:                    releaseID,
		IncidentID:                   incidentID,
		SnapshotID:                   snapshotID,
		CreatedByUserID:              createdBy,
		ClientTxnID:                  row.ClientTxnID,
		ReleaseScope:                 row.ReleaseScope,
		ReleaseState:                 row.ReleaseState,
		SnapshotAt:                   snapshotAt,
		SourceChangeSetHighWatermark: row.SourceChangeSetHighWatermark,
		DerivationVersion:            row.DerivationVersion,
		ExportModelSHA256:            row.ExportModelSha256,
		TemplateID:                   row.TemplateID,
		TemplateVersion:              row.TemplateVersion,
		RedactionProfileID:           row.RedactionProfileID,
		RedactionProfileVersion:      row.RedactionProfileVersion,
		RedactionProfileSHA256:       row.RedactionProfileSha256,
		OutputKind:                   row.OutputKind,
		OutputOptions:                cloneRawMessageWithDefault(row.OutputOptions, json.RawMessage(`{}`)),
		GraphProjectionRefs:          cloneRawMessageWithDefault(row.GraphProjectionRefs, json.RawMessage(`[]`)),
		CompositionID:                optionalUUIDFromPG(row.CompositionID),
		CompositionVersion:           optionalStringFromPG(row.CompositionVersion),
		CompositionSHA256:            optionalStringFromPG(row.CompositionSha256),
		RenderAdmittedAt:             renderAdmittedAt,
		OutputMediaType:              optionalStringFromPG(row.OutputMediaType),
		OutputSHA256:                 optionalStringFromPG(row.OutputSha256),
		RedactionManifestSHA256:      optionalStringFromPG(row.RedactionManifestSha256),
		RedactionManifestJSON:        append([]byte(nil), row.RedactionManifestJson...),
		CreateJobID:                  createJobID,
		RenderFailedReasonCode:       optionalStringFromPG(row.RenderFailedReasonCode),
		RecipientPartitionRefs:       decodeStringArrayJSON(row.RecipientPartitionRefs),
		ApprovedAt:                   optionalTimeFromPG(row.ApprovedAt),
		PublishedAt:                  optionalTimeFromPG(row.PublishedAt),
		InvalidatedAt:                optionalTimeFromPG(row.InvalidatedAt),
		InvalidationReason:           optionalStringFromPG(row.InvalidationReason),
		CreatedAt:                    createdAt,
		UpdatedAt:                    updatedAt,
	}, nil
}

func lookupRouteIdempotencyTx(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
	if err := authn.LockRouteIdempotencyTx(ctx, tx, key); err != nil {
		return authn.RouteIdempotencyRecord{}, err
	}
	row := tx.QueryRow(ctx, `
SELECT route_key, scope_key, client_txn_id, actor_user_id, request_hash, status_code, response_json
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID)
	var record authn.RouteIdempotencyRecord
	if err := row.Scan(&record.RouteKey, &record.ScopeKey, &record.ClientTxnID, &record.ActorUserID, &record.RequestHash, &record.StatusCode, &record.ResponseJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authn.RouteIdempotencyRecord{}, authn.ErrNotFound
		}
		return authn.RouteIdempotencyRecord{}, err
	}
	return record, nil
}

func getSnapshotPayloadTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (snapshotCreateJobPayload, error) {
	row, err := sqlc.New(tx).GetReportingJobPayload(ctx, pgUUID(jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return snapshotCreateJobPayload{}, ErrNotFound
	}
	if err != nil {
		return snapshotCreateJobPayload{}, err
	}
	if row.JobKind != "snapshot_create" {
		return snapshotCreateJobPayload{}, fmt.Errorf("reporting job %s has kind %q", jobID, row.JobKind)
	}
	payload, err := decodeSnapshotJobPayload(row.RequestJson, jobID, row.IncidentID.Bytes, row.ActorUserID.Bytes, row.CreatedAt.Time)
	if err != nil {
		return snapshotCreateJobPayload{}, err
	}
	admission, err := jobs.ReadRetainedExtensionJob(ctx, tx, jobID)
	if err != nil {
		return snapshotCreateJobPayload{}, err
	}
	if err := validateSnapshotJobAdmission(payload, jobID, admission); err != nil {
		return snapshotCreateJobPayload{}, err
	}
	seed, err := readSnapshotAdmissionTx(ctx, tx, jobID)
	if err != nil {
		return snapshotCreateJobPayload{}, err
	}
	if seed.ClientTxnID != payload.ClientTxnID || !sameReportingValue(seed.RequestedBoundary, payload.RequestedBoundary) || !sameSnapshotCaptureBinding(seed.ReferencePacks, payload.SourceCapture, payload.SourceCaptureSHA256) {
		return snapshotCreateJobPayload{}, errors.New("reporting: capture differs from admitted request")
	}
	return payload, nil
}

func getReleasePayloadTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (releaseCreateJobPayload, error) {
	a, input, err := materializeRetainedRenderTx(ctx, tx, jobID)
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	return releasePayloadFromPrepared(a, input)
}

func payloadUUIDs(payload releaseCreateJobPayload) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	incidentID, err := uuid.Parse(payload.IncidentID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, uuid.UUID{}, err
	}
	snapshotID, err := uuid.Parse(payload.SnapshotID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, uuid.UUID{}, err
	}
	actorID, err := uuid.Parse(payload.ActorUserID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, uuid.UUID{}, err
	}
	return incidentID, snapshotID, actorID, nil
}

func snapshotResource(record snapshotDescriptor) map[string]any {
	return map[string]any{
		"snapshot_id":                      record.ID.String(),
		"incident_id":                      record.Incident.String(),
		"created_by_user_id":               record.Actor.String(),
		"created_at":                       record.CreatedAt.UTC(),
		"snapshot_at":                      record.At.UTC(),
		"source_change_set_high_watermark": record.Boundary,
		"derivation_version":               record.Derivation,
		"source_capture_sha256":            record.CaptureSHA256,
	}
}

func releaseResource(record ReleaseRecord) map[string]any {
	resource := map[string]any{
		"release_id":                       record.ReleaseID.String(),
		"incident_id":                      record.IncidentID.String(),
		"snapshot_id":                      record.SnapshotID.String(),
		"snapshot_at":                      record.SnapshotAt.UTC(),
		"source_change_set_high_watermark": record.SourceChangeSetHighWatermark,
		"derivation_version":               record.DerivationVersion,
		"export_model_sha256":              record.ExportModelSHA256,
		"template_id":                      record.TemplateID,
		"template_version":                 record.TemplateVersion,
		"redaction_profile_id":             record.RedactionProfileID,
		"redaction_profile_version":        record.RedactionProfileVersion,
		"redaction_profile_sha256":         record.RedactionProfileSHA256,
		"output_kind":                      record.OutputKind,
		"output_options":                   rawJSONValue(record.OutputOptions),
		"output_media_type":                record.OutputMediaType,
		"release_scope":                    record.ReleaseScope,
		"recipient_partition_refs":         stringArrayForResource(record.RecipientPartitionRefs),
		"graph_projection_refs":            rawJSONValue(record.GraphProjectionRefs),
		"composition_id":                   optionalUUIDForResource(record.CompositionID),
		"composition_version":              optionalStringForResource(record.CompositionVersion),
		"composition_sha256":               optionalStringForResource(record.CompositionSHA256),
		"render_admitted_at":               record.RenderAdmittedAt.UTC(),
		"output_sha256":                    record.OutputSHA256,
		"redaction_manifest_sha256":        record.RedactionManifestSHA256,
		"release_state":                    record.ReleaseState,
		"render_failed_reason_code":        nil,
		"created_by_user_id":               record.CreatedByUserID.String(),
		"created_at":                       record.CreatedAt.UTC(),
		"approved_at":                      nil,
		"invalidated_at":                   nil,
		"published_at":                     nil,
		"invalidation_reason":              nil,
	}
	if record.ApprovedAt != nil {
		resource["approved_at"] = record.ApprovedAt.UTC()
	}
	if record.PublishedAt != nil {
		resource["published_at"] = record.PublishedAt.UTC()
	}
	if record.InvalidatedAt != nil {
		resource["invalidated_at"] = record.InvalidatedAt.UTC()
	}
	if record.InvalidationReason != nil {
		resource["invalidation_reason"] = *record.InvalidationReason
	}
	if record.RenderFailedReasonCode != nil {
		resource["render_failed_reason_code"] = *record.RenderFailedReasonCode
	}
	return resource
}

func invalidatePriorCandidateTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, snapshotID uuid.UUID, outputKind string, releaseScope string, templateID string, templateVersion string, redactionProfileID string, redactionProfileVersion string, recipientPartitionRefs []byte, outputOptions json.RawMessage, graphProjectionRefs json.RawMessage, compositionID *string, compositionVersion *string, compositionSHA256 *string, now time.Time) error {
	return sqlc.New(tx).InvalidateSupersededReportingReleases(ctx, sqlc.InvalidateSupersededReportingReleasesParams{
		SnapshotID:              pgUUID(snapshotID),
		OutputKind:              outputKind,
		ReleaseScope:            releaseScope,
		TemplateID:              templateID,
		TemplateVersion:         templateVersion,
		RedactionProfileID:      redactionProfileID,
		RedactionProfileVersion: redactionProfileVersion,
		RecipientPartitionRefs:  recipientPartitionRefs,
		OutputOptions:           []byte(outputOptions),
		GraphProjectionRefs:     []byte(graphProjectionRefs),
		CompositionID:           optionalPGUUIDString(compositionID),
		CompositionVersion:      optionalPGText(compositionVersion),
		CompositionSha256:       optionalPGText(compositionSHA256),
		ReleaseID:               pgUUID(releaseID),
		UpdatedAt:               pgTimestamptz(now),
	})
}

func approvalsSatisfiedTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, releaseScope string) (bool, error) {
	rows, err := sqlc.New(tx).ListReportingReleaseApprovals(ctx, pgUUID(releaseID))
	if err != nil {
		return false, err
	}
	var reviewer *uuid.UUID
	var admin *uuid.UUID
	for _, row := range rows {
		actorID, err := uuidFromPG(row.ActorUserID)
		if err != nil {
			return false, err
		}
		switch row.ApprovalRole {
		case "reviewer":
			id := actorID
			reviewer = &id
		case "admin":
			id := actorID
			admin = &id
		}
	}
	switch releaseScope {
	case ReleaseScopeInternalReview:
		return reviewer != nil, nil
	case ReleaseScopeExternal:
		return reviewer != nil && admin != nil && *reviewer != *admin, nil
	default:
		return false, nil
	}
}

func approvalExistsTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, actorID uuid.UUID, role string) (bool, error) {
	return sqlc.New(tx).ReportingReleaseApprovalExists(ctx, sqlc.ReportingReleaseApprovalExistsParams{
		ReleaseID:    pgUUID(releaseID),
		ActorUserID:  pgUUID(actorID),
		ApprovalRole: role,
	})
}

func updateReleaseStateTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, state string, now time.Time, timestampColumn string) (ReleaseRecord, error) {
	if timestampColumn != "approved_at" && timestampColumn != "published_at" {
		return ReleaseRecord{}, errors.New("unsupported release timestamp column")
	}
	row, err := sqlc.New(tx).UpdateReportingReleaseState(ctx, sqlc.UpdateReportingReleaseStateParams{
		ReleaseID:    pgUUID(releaseID),
		ReleaseState: state,
		UpdatedAt:    pgTimestamptz(now),
	})
	if err != nil {
		return ReleaseRecord{}, err
	}
	return releaseRecordFromSQL(row)
}

func invalidateReleaseTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, now time.Time, reason *string) (ReleaseRecord, error) {
	row, err := sqlc.New(tx).InvalidateReportingRelease(ctx, sqlc.InvalidateReportingReleaseParams{
		ReleaseID:          pgUUID(releaseID),
		UpdatedAt:          pgTimestamptz(now),
		InvalidationReason: optionalPGText(reason),
	})
	if err != nil {
		return ReleaseRecord{}, err
	}
	return releaseRecordFromSQL(row)
}

func approvalRoleForIncidentRole(role string) (string, bool) {
	switch role {
	case "reviewer":
		return "reviewer", true
	case "admin":
		return "admin", true
	default:
		return "", false
	}
}

func approvalTupleJSON(release ReleaseRecord, actorUserID uuid.UUID, approvalRole string) ([]byte, error) {
	return canonicalJSON(map[string]any{
		"release_id":                       release.ReleaseID.String(),
		"snapshot_id":                      release.SnapshotID.String(),
		"actor_user_id":                    actorUserID.String(),
		"approval_role":                    approvalRole,
		"template_id":                      release.TemplateID,
		"template_version":                 release.TemplateVersion,
		"redaction_profile_id":             release.RedactionProfileID,
		"redaction_profile_version":        release.RedactionProfileVersion,
		"redaction_profile_sha256":         release.RedactionProfileSHA256,
		"export_model_sha256":              release.ExportModelSHA256,
		"output_kind":                      release.OutputKind,
		"output_options":                   rawJSONValue(release.OutputOptions),
		"output_media_type":                derefString(release.OutputMediaType),
		"output_sha256":                    derefString(release.OutputSHA256),
		"redaction_manifest_sha256":        derefString(release.RedactionManifestSHA256),
		"release_scope":                    release.ReleaseScope,
		"recipient_partition_refs":         stringArrayForResource(release.RecipientPartitionRefs),
		"graph_projection_refs":            rawJSONValue(release.GraphProjectionRefs),
		"composition_id":                   optionalUUIDForResource(release.CompositionID),
		"composition_version":              optionalStringForResource(release.CompositionVersion),
		"composition_sha256":               optionalStringForResource(release.CompositionSHA256),
		"render_admitted_at":               release.RenderAdmittedAt.UTC(),
		"source_change_set_high_watermark": release.SourceChangeSetHighWatermark,
	})
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}

func optionalPGUUIDString(value *string) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	parsed, err := uuid.Parse(*value)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgUUID(parsed)
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.FromBytes(value.Bytes[:])
}

func optionalUUIDFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return nil
	}
	return &parsed
}

func pgTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func optionalPGTimestamptz(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgTimestamptz(*value)
}

func timeFromPG(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, errors.New("missing timestamp")
	}
	return value.Time.UTC(), nil
}

func optionalTimeFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	parsed := value.Time.UTC()
	return &parsed
}

func optionalPGText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func requiredPGText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}

func optionalStringFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	parsed := value.String
	return &parsed
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func decodeStringArrayJSON(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	return cloneStrings(values)
}

func canonicalStringArrayJSON(values []string) ([]byte, error) {
	out := cloneStrings(values)
	if out == nil {
		out = []string{}
	}
	return canonicalJSON(out)
}

func stringArrayForResource(values []string) []string {
	out := cloneStrings(values)
	if out == nil {
		return []string{}
	}
	return out
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func optionalUUIDStringPointer(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	parsed := value.String()
	return &parsed
}

func optionalUUIDForResource(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func optionalStringForResource(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func cloneRawMessageWithDefault(value json.RawMessage, fallback json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return append(json.RawMessage(nil), fallback...)
	}
	return append(json.RawMessage(nil), value...)
}
