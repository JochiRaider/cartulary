package reporting

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
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

type Store struct {
	pool *pgxpool.Pool
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
	ExportModelSHA256            string
	ExportModelJSON              []byte
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
	OutputMediaType              *string
	OutputSHA256                 *string
	RedactionManifestSHA256      *string
	RedactionManifestJSON        []byte
	RenderedOutput               *string
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
	SnapshotID uuid.UUID
	Replayed   bool
}

type SourceBoundaryState struct {
	IncidentID               string  `json:"incident_id"`
	IncidentVersion          int64   `json:"incident_version"`
	LatestChangeSetID        *string `json:"latest_change_set_id"`
	LatestChangeSetCreatedAt *string `json:"latest_change_set_created_at"`
}

type ResolvedSourceBoundary struct {
	Token         string
	CanonicalJSON []byte
	State         SourceBoundaryState
}

type RenderedRelease struct {
	Profile                 RedactionProfile
	ProfileSHA256           string
	Redaction               RedactionResult
	Output                  []byte
	OutputMediaType         string
	OutputSHA256            string
	RedactionManifestSHA256 string
	RedactionManifestJSON   []byte
}

type CreateReleaseParams struct {
	ActorUserID      uuid.UUID
	Request          CreateReleaseRequest
	TemplateContract TemplateContract
	Now              time.Time
}

type CreateReleaseResult struct {
	Job       jobs.Resource
	ReleaseID uuid.UUID
	Replayed  bool
}

type snapshotCreateJobPayload struct {
	IncidentID                   string          `json:"incident_id"`
	ActorUserID                  string          `json:"actor_user_id"`
	ClientTxnID                  string          `json:"client_txn_id"`
	SnapshotAt                   time.Time       `json:"snapshot_at"`
	SourceChangeSetHighWatermark string          `json:"source_change_set_high_watermark"`
	SourceBoundaryJSON           json.RawMessage `json:"source_boundary_json"`
	ExportModel                  ExportModel     `json:"export_model"`
	ExportModelSHA256            string          `json:"export_model_sha256"`
}

type releaseCreateJobPayload struct {
	ActorUserID                  string               `json:"actor_user_id"`
	ClientTxnID                  string               `json:"client_txn_id"`
	SnapshotID                   string               `json:"snapshot_id"`
	IncidentID                   string               `json:"incident_id"`
	SnapshotAt                   time.Time            `json:"snapshot_at"`
	SourceChangeSetHighWatermark string               `json:"source_change_set_high_watermark"`
	DerivationVersion            string               `json:"derivation_version"`
	ExportModelSHA256            string               `json:"export_model_sha256"`
	ExportModel                  ExportModel          `json:"export_model"`
	TemplateID                   string               `json:"template_id"`
	TemplateVersion              string               `json:"template_version"`
	TemplateContract             TemplateContract     `json:"template_contract"`
	RedactionProfileID           string               `json:"redaction_profile_id"`
	RedactionProfileVersion      string               `json:"redaction_profile_version"`
	OutputKind                   string               `json:"output_kind"`
	ReleaseScope                 string               `json:"release_scope"`
	RecipientPartitionRefs       []string             `json:"recipient_partition_refs"`
	NormalizedRequest            []byte               `json:"normalized_request"`
	Request                      CreateReleaseRequest `json:"-"`
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

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateSnapshot(ctx context.Context, params CreateSnapshotParams) (CreateSnapshotResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey:    "snapshots.create",
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.Request.IncidentID.String(),
		ClientTxnID: params.Request.ClientTxnID,
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if err == nil {
		existingJobID, parseErr := existingResponseJobID(existing.ResponseJSON)
		if parseErr != nil {
			return CreateSnapshotResult{}, parseErr
		}
		payload, payloadErr := getSnapshotPayloadTx(ctx, tx, existingJobID)
		if payloadErr != nil {
			return CreateSnapshotResult{}, payloadErr
		}
		if params.Request.SourceChangeSetHighWatermark != nil && *params.Request.SourceChangeSetHighWatermark != payload.SourceChangeSetHighWatermark {
			return CreateSnapshotResult{}, authn.ErrClientTxnConflict
		}
		var resource jobs.Resource
		if err := json.Unmarshal(existing.ResponseJSON, &resource); err != nil {
			return CreateSnapshotResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return CreateSnapshotResult{}, err
		}
		return CreateSnapshotResult{Job: resource, Replayed: true}, nil
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return CreateSnapshotResult{}, err
	}

	incident, err := getIncidentTx(ctx, tx, params.Request.IncidentID)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	boundary, err := resolveSourceBoundaryTx(ctx, tx, incident)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	if params.Request.SourceChangeSetHighWatermark != nil && *params.Request.SourceChangeSetHighWatermark != boundary.Token {
		return CreateSnapshotResult{}, &SnapshotBoundaryConflictError{Expected: *params.Request.SourceChangeSetHighWatermark, Actual: boundary.Token}
	}
	normalized, err := normalizeSnapshotRequest(params.Request.IncidentID, params.Request.ClientTxnID, &boundary.Token)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	sum := sha256.Sum256(normalized)
	requestHash := sum[:]
	workbookFields, err := collectWorkbookExportFieldsTx(ctx, tx, params.Request.IncidentID)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	model, exportSHA, err := BuildExportModel(incident, params.Now.UTC(), boundary.Token, workbookFields)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	payloadJSON, err := canonicalJSON(snapshotCreateJobPayload{
		IncidentID:                   params.Request.IncidentID.String(),
		ActorUserID:                  params.ActorUserID.String(),
		ClientTxnID:                  params.Request.ClientTxnID,
		SnapshotAt:                   params.Now.UTC(),
		SourceChangeSetHighWatermark: boundary.Token,
		SourceBoundaryJSON:           append(json.RawMessage(nil), boundary.CanonicalJSON...),
		ExportModel:                  model,
		ExportModelSHA256:            exportSHA,
	})
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	job, err := jobs.CreateQueuedTx(ctx, tx, jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &params.Request.IncidentID},
		SubmittedByUserID: params.ActorUserID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	}, params.Now.UTC())
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	jobID := uuid.MustParse(job.JobID)
	if err := sqlc.New(tx).CreateReportingJobPayload(ctx, sqlc.CreateReportingJobPayloadParams{
		JobID:       pgUUID(jobID),
		JobKind:     "snapshot_create",
		IncidentID:  pgUUID(params.Request.IncidentID),
		ActorUserID: pgUUID(params.ActorUserID),
		RequestJson: payloadJSON,
		CreatedAt:   pgTimestamptz(params.Now),
	}); err != nil {
		return CreateSnapshotResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, job); err != nil {
		return CreateSnapshotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateSnapshotResult{}, err
	}
	return CreateSnapshotResult{Job: job}, nil
}

type SnapshotBoundaryConflictError struct {
	Expected string
	Actual   string
}

func (e *SnapshotBoundaryConflictError) Error() string {
	return "snapshot source boundary conflict"
}

func (s *Store) GetSnapshot(ctx context.Context, snapshotID uuid.UUID) (map[string]any, uuid.UUID, error) {
	record, err := s.getSnapshotRecord(ctx, snapshotID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	return snapshotResource(record), record.IncidentID, nil
}

func (s *Store) GetSnapshotForRender(ctx context.Context, snapshotID uuid.UUID) (SnapshotRecord, ExportModel, error) {
	record, err := s.getSnapshotRecord(ctx, snapshotID)
	if err != nil {
		return SnapshotRecord{}, ExportModel{}, err
	}
	if record.DerivationVersion != DerivationVersion {
		return SnapshotRecord{}, ExportModel{}, fmt.Errorf("unsupported snapshot derivation version %q", record.DerivationVersion)
	}
	var model ExportModel
	if err := json.Unmarshal(record.ExportModelJSON, &model); err != nil {
		return SnapshotRecord{}, ExportModel{}, err
	}
	if model.SchemaID != ExportModelSchemaID || model.DerivationVersion != DerivationVersion {
		return SnapshotRecord{}, ExportModel{}, fmt.Errorf("unsupported snapshot export model %q derivation %q", model.SchemaID, model.DerivationVersion)
	}
	return record, model, nil
}

func (s *Store) CreateRelease(ctx context.Context, params CreateReleaseParams) (CreateReleaseResult, error) {
	sum := sha256.Sum256(params.Request.Normalized)
	requestHash := sum[:]
	key := authn.RouteIdempotencyKey{
		RouteKey:    "releases.create",
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.Request.SnapshotID.String(),
		ClientTxnID: params.Request.ClientTxnID,
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return CreateReleaseResult{}, authn.ErrClientTxnConflict
		}
		var resource jobs.Resource
		if err := json.Unmarshal(existing.ResponseJSON, &resource); err != nil {
			return CreateReleaseResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return CreateReleaseResult{}, err
		}
		return CreateReleaseResult{Job: resource, Replayed: true}, nil
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return CreateReleaseResult{}, err
	}

	snapshot, err := getSnapshotRecordTx(ctx, tx, params.Request.SnapshotID)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	if snapshot.DerivationVersion != DerivationVersion {
		return CreateReleaseResult{}, fmt.Errorf("unsupported snapshot derivation version %q", snapshot.DerivationVersion)
	}
	var model ExportModel
	if err := json.Unmarshal(snapshot.ExportModelJSON, &model); err != nil {
		return CreateReleaseResult{}, err
	}
	if model.SchemaID != ExportModelSchemaID || model.DerivationVersion != DerivationVersion {
		return CreateReleaseResult{}, fmt.Errorf("unsupported snapshot export model %q derivation %q", model.SchemaID, model.DerivationVersion)
	}
	payloadJSON, err := canonicalJSON(releaseCreateJobPayload{
		ActorUserID:                  params.ActorUserID.String(),
		ClientTxnID:                  params.Request.ClientTxnID,
		SnapshotID:                   snapshot.SnapshotID.String(),
		IncidentID:                   snapshot.IncidentID.String(),
		SnapshotAt:                   snapshot.SnapshotAt.UTC(),
		SourceChangeSetHighWatermark: snapshot.SourceChangeSetHighWatermark,
		DerivationVersion:            snapshot.DerivationVersion,
		ExportModelSHA256:            snapshot.ExportModelSHA256,
		ExportModel:                  model,
		TemplateID:                   params.Request.TemplateID,
		TemplateVersion:              params.Request.TemplateVersion,
		TemplateContract:             params.TemplateContract,
		RedactionProfileID:           params.Request.RedactionProfileID,
		RedactionProfileVersion:      params.Request.RedactionProfileVersion,
		OutputKind:                   params.Request.OutputKind,
		ReleaseScope:                 params.Request.ReleaseScope,
		RecipientPartitionRefs:       cloneStrings(params.Request.RecipientPartitionRefs),
		NormalizedRequest:            append([]byte(nil), params.Request.Normalized...),
	})
	if err != nil {
		return CreateReleaseResult{}, err
	}
	job, err := jobs.CreateQueuedTx(ctx, tx, jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &snapshot.IncidentID},
		SubmittedByUserID: params.ActorUserID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	}, params.Now.UTC())
	if err != nil {
		return CreateReleaseResult{}, err
	}
	jobID := uuid.MustParse(job.JobID)
	if err := sqlc.New(tx).CreateReportingJobPayload(ctx, sqlc.CreateReportingJobPayloadParams{
		JobID:       pgUUID(jobID),
		JobKind:     "release_create",
		IncidentID:  pgUUID(snapshot.IncidentID),
		ActorUserID: pgUUID(params.ActorUserID),
		RequestJson: payloadJSON,
		CreatedAt:   pgTimestamptz(params.Now),
	}); err != nil {
		return CreateReleaseResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, job); err != nil {
		return CreateReleaseResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateReleaseResult{}, err
	}
	return CreateReleaseResult{Job: job}, nil
}

func (s *Store) GetRelease(ctx context.Context, releaseID uuid.UUID) (map[string]any, uuid.UUID, error) {
	record, err := s.getReleaseRecord(ctx, releaseID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	return releaseResource(record), record.IncidentID, nil
}

func (s *Store) CompleteSnapshotCreateJob(ctx context.Context, jobID uuid.UUID) (uuid.UUID, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if existing, err := getSnapshotRecordByCreateJobIDTx(ctx, tx, jobID); err == nil {
		return existing.SnapshotID, tx.Commit(ctx)
	} else if !errors.Is(err, ErrNotFound) {
		return uuid.UUID{}, err
	}
	payload, err := getSnapshotPayloadTx(ctx, tx, jobID)
	if err != nil {
		return uuid.UUID{}, err
	}
	incidentID, err := uuid.Parse(payload.IncidentID)
	if err != nil {
		return uuid.UUID{}, err
	}
	actorID, err := uuid.Parse(payload.ActorUserID)
	if err != nil {
		return uuid.UUID{}, err
	}
	exportJSON, err := canonicalJSON(payload.ExportModel)
	if err != nil {
		return uuid.UUID{}, err
	}
	row, err := sqlc.New(tx).CreateReportingSnapshot(ctx, sqlc.CreateReportingSnapshotParams{
		IncidentID:                   pgUUID(incidentID),
		CreatedByUserID:              pgUUID(actorID),
		ClientTxnID:                  payload.ClientTxnID,
		SnapshotAt:                   pgTimestamptz(payload.SnapshotAt),
		SourceChangeSetHighWatermark: payload.SourceChangeSetHighWatermark,
		SourceBoundaryJson:           []byte(payload.SourceBoundaryJSON),
		DerivationVersion:            DerivationVersion,
		ExportModelSha256:            payload.ExportModelSHA256,
		ExportModelJson:              exportJSON,
		CreateJobID:                  pgUUID(jobID),
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	record, err := snapshotRecordFromSQL(row)
	if err != nil {
		return uuid.UUID{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, err
	}
	return record.SnapshotID, nil
}

func (s *Store) ReleasePayloadForJob(ctx context.Context, jobID uuid.UUID) (releaseCreateJobPayload, error) {
	row, err := sqlc.New(s.pool).GetReportingJobPayload(ctx, pgUUID(jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return releaseCreateJobPayload{}, ErrNotFound
	}
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	if row.JobKind != "release_create" {
		return releaseCreateJobPayload{}, fmt.Errorf("reporting job %s has kind %q", jobID, row.JobKind)
	}
	var payload releaseCreateJobPayload
	if err := json.Unmarshal(row.RequestJson, &payload); err != nil {
		return releaseCreateJobPayload{}, err
	}
	snapshotID, err := uuid.Parse(payload.SnapshotID)
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	payload.Request = CreateReleaseRequest{
		SnapshotID:              snapshotID,
		ClientTxnID:             payload.ClientTxnID,
		TemplateID:              payload.TemplateID,
		TemplateVersion:         payload.TemplateVersion,
		RedactionProfileID:      payload.RedactionProfileID,
		RedactionProfileVersion: payload.RedactionProfileVersion,
		OutputKind:              payload.OutputKind,
		ReleaseScope:            payload.ReleaseScope,
		RecipientPartitionRefs:  cloneStrings(payload.RecipientPartitionRefs),
		Normalized:              append([]byte(nil), payload.NormalizedRequest...),
	}
	return payload, nil
}

func (s *Store) ReportingJobKind(ctx context.Context, jobID uuid.UUID) (string, error) {
	row, err := sqlc.New(s.pool).GetReportingJobPayload(ctx, pgUUID(jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return row.JobKind, nil
}

func (s *Store) CompleteReleaseCreateJob(ctx context.Context, jobID uuid.UUID, rendered RenderedRelease, now time.Time) (uuid.UUID, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if existing, err := getReleaseRecordByCreateJobIDTx(ctx, tx, jobID); err == nil {
		return existing.ReleaseID, tx.Commit(ctx)
	} else if !errors.Is(err, ErrNotFound) {
		return uuid.UUID{}, err
	}
	payload, err := getReleasePayloadTx(ctx, tx, jobID)
	if err != nil {
		return uuid.UUID{}, err
	}
	incidentID, snapshotID, actorID, err := payloadUUIDs(payload)
	if err != nil {
		return uuid.UUID{}, err
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
		return uuid.UUID{}, err
	}
	row, err := sqlc.New(tx).CreateReportingRelease(ctx, sqlc.CreateReportingReleaseParams{
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
		OutputMediaType:              requiredPGText(rendered.OutputMediaType),
		OutputSha256:                 requiredPGText(rendered.OutputSHA256),
		RedactionManifestSha256:      requiredPGText(rendered.RedactionManifestSHA256),
		RedactionManifestJson:        rendered.RedactionManifestJSON,
		RenderedOutput:               requiredPGText(string(rendered.Output)),
		CreateJobID:                  pgUUID(jobID),
		RecipientPartitionRefs:       partitionJSON,
		ApprovedAt:                   optionalPGTimestamptz(approvedAt),
		CreatedAt:                    pgTimestamptz(now),
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	releaseID, err := uuidFromPG(row.ReleaseID)
	if err != nil {
		return uuid.UUID{}, err
	}
	if err := invalidatePriorCandidateTx(ctx, tx, releaseID, snapshotID, payload.OutputKind, payload.ReleaseScope, payload.TemplateID, payload.TemplateVersion, rendered.Profile.ProfileID, rendered.Profile.Version, partitionJSON, now.UTC()); err != nil {
		return uuid.UUID{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, err
	}
	return releaseID, nil
}

func (s *Store) CompleteReleaseRenderFailedJob(ctx context.Context, jobID uuid.UUID, profile RedactionProfile, profileSHA string, reasonCode string, now time.Time) (uuid.UUID, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if existing, err := getReleaseRecordByCreateJobIDTx(ctx, tx, jobID); err == nil {
		return existing.ReleaseID, tx.Commit(ctx)
	} else if !errors.Is(err, ErrNotFound) {
		return uuid.UUID{}, err
	}
	payload, err := getReleasePayloadTx(ctx, tx, jobID)
	if err != nil {
		return uuid.UUID{}, err
	}
	incidentID, snapshotID, actorID, err := payloadUUIDs(payload)
	if err != nil {
		return uuid.UUID{}, err
	}
	partitionJSON, err := canonicalStringArrayJSON(payload.RecipientPartitionRefs)
	if err != nil {
		return uuid.UUID{}, err
	}
	if profile.ProfileID == "" {
		profile.ProfileID = payload.RedactionProfileID
		profile.Version = payload.RedactionProfileVersion
	}
	if profileSHA == "" {
		profileSHA = strings.Repeat("0", 64)
	}
	row, err := sqlc.New(tx).CreateRenderFailedReportingRelease(ctx, sqlc.CreateRenderFailedReportingReleaseParams{
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
	if err := tx.Commit(ctx); err != nil {
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

func (s *Store) getSnapshotRecord(ctx context.Context, snapshotID uuid.UUID) (SnapshotRecord, error) {
	row, err := sqlc.New(s.pool).GetReportingSnapshot(ctx, pgUUID(snapshotID))
	if errors.Is(err, pgx.ErrNoRows) {
		return SnapshotRecord{}, ErrNotFound
	}
	if err != nil {
		return SnapshotRecord{}, err
	}
	return snapshotRecordFromSQL(row)
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

func getSnapshotRecordByCreateJobIDTx(ctx context.Context, tx pgx.Tx, createJobID uuid.UUID) (SnapshotRecord, error) {
	row, err := sqlc.New(tx).GetReportingSnapshotByCreateJob(ctx, pgUUID(createJobID))
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

func getIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (incidents.IncidentRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT id, incident_key, title, description, status, severity, tlp, current_phase,
       primary_external_case_ref, created_by_user_id, created_at, updated_at, updated_by_user_id,
       incident_version, closed_at
  FROM incidents
 WHERE id = $1
`, incidentID)
	var record incidents.IncidentRecord
	if err := row.Scan(
		&record.ID,
		&record.IncidentKey,
		&record.Title,
		&record.Description,
		&record.Status,
		&record.Severity,
		&record.TLP,
		&record.CurrentPhase,
		&record.PrimaryExternalCaseRef,
		&record.CreatedByUserID,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.UpdatedByUserID,
		&record.IncidentVersion,
		&record.ClosedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return incidents.IncidentRecord{}, ErrNotFound
		}
		return incidents.IncidentRecord{}, err
	}
	return record, nil
}

func resolveSourceBoundaryTx(ctx context.Context, tx pgx.Tx, incident incidents.IncidentRecord) (ResolvedSourceBoundary, error) {
	var latestID pgtype.Text
	var latestCreated pgtype.Timestamptz
	err := tx.QueryRow(ctx, `
SELECT change_set_id::text, created_at
  FROM change_sets
 WHERE incident_id = $1
 ORDER BY created_at DESC, change_set_id DESC
 LIMIT 1
`, incident.ID).Scan(&latestID, &latestCreated)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ResolvedSourceBoundary{}, err
	}

	var latestIDPtr *string
	var latestCreatedPtr *string
	if err == nil {
		id := latestID.String
		latestIDPtr = &id
		created := latestCreated.Time.UTC().Format(time.RFC3339Nano)
		latestCreatedPtr = &created
	}
	state := SourceBoundaryState{
		IncidentID:               incident.ID.String(),
		IncidentVersion:          incident.IncidentVersion,
		LatestChangeSetID:        latestIDPtr,
		LatestChangeSetCreatedAt: latestCreatedPtr,
	}
	encoded, err := canonicalJSON(state)
	if err != nil {
		return ResolvedSourceBoundary{}, err
	}
	return ResolvedSourceBoundary{
		Token:         SourceBoundaryTokenPrefix + hashHex(encoded),
		CanonicalJSON: encoded,
		State:         state,
	}, nil
}

type workbookExportQuery struct {
	Prefix string
	SQL    string
}

func collectWorkbookExportFieldsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) ([]ExportField, error) {
	supportRefs, err := collectSupportRefsTx(ctx, tx, incidentID)
	if err != nil {
		return nil, err
	}
	queries := []workbookExportQuery{
		{
			Prefix: "record_envelopes",
			SQL: `SELECT r.record_id::text, 'record_envelope'::text, 'derived_analytic'::text, to_jsonb(r) - 'incident_id'
  FROM records r
 WHERE r.incident_id = $1 AND r.deleted_at IS NULL`,
		},
		{
			Prefix: "timeline",
			SQL: `SELECT t.record_id::text, 'timeline_event'::text, 'source_evidence'::text, to_jsonb(t) - 'incident_id'
  FROM timeline_events t
  JOIN records r ON r.incident_id = t.incident_id AND r.record_id = t.record_id AND r.deleted_at IS NULL
 WHERE t.incident_id = $1`,
		},
		{
			Prefix: "hosts",
			SQL: `SELECT h.record_id::text, 'host'::text, 'derived_analytic'::text, to_jsonb(h) - 'incident_id'
  FROM host_grid_projection h
  JOIN records r ON r.incident_id = h.incident_id AND r.record_id = h.record_id AND r.deleted_at IS NULL
 WHERE h.incident_id = $1`,
		},
		{
			Prefix: "identities",
			SQL: `SELECT i.record_id::text, 'identity'::text, 'derived_analytic'::text, to_jsonb(i) - 'incident_id'
  FROM identity_grid_projection i
  JOIN records r ON r.incident_id = i.incident_id AND r.record_id = i.record_id AND r.deleted_at IS NULL
 WHERE i.incident_id = $1`,
		},
		{
			Prefix: "parties",
			SQL: `SELECT p.record_id::text, 'party'::text, 'source_evidence'::text, to_jsonb(p) - 'incident_id'
  FROM parties p
  JOIN records r ON r.incident_id = p.incident_id AND r.record_id = p.record_id AND r.deleted_at IS NULL
 WHERE p.incident_id = $1`,
		},
		{
			Prefix: "evidence",
			SQL: `SELECT e.record_id::text, 'evidence'::text, 'source_evidence'::text, to_jsonb(e) - 'incident_id' - 'blob_hash' - 'storage_ref' - 'object_blob_id'
  FROM evidence e
  JOIN records r ON r.incident_id = e.incident_id AND r.record_id = e.record_id AND r.deleted_at IS NULL
 WHERE e.incident_id = $1`,
		},
		{
			Prefix: "task_requests",
			SQL: `SELECT t.record_id::text, 'task_request'::text, 'working_material'::text, to_jsonb(t) - 'incident_id'
  FROM task_request_grid_projection t
  JOIN records r ON r.incident_id = t.incident_id AND r.record_id = t.record_id AND r.deleted_at IS NULL
 WHERE t.incident_id = $1`,
		},
		{
			Prefix: "decisions",
			SQL: `SELECT d.record_id::text, 'decision'::text, 'working_material'::text, to_jsonb(d) - 'incident_id'
  FROM decision_grid_projection d
  JOIN records r ON r.incident_id = d.incident_id AND r.record_id = d.record_id AND r.deleted_at IS NULL
 WHERE d.incident_id = $1`,
		},
		{
			Prefix: "notes",
			SQL: `SELECT a.record_id::text, 'note'::text, 'working_material'::text, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'note'`,
		},
		{
			Prefix: "findings",
			SQL: `SELECT a.record_id::text, 'finding_hypothesis'::text,
       CASE WHEN a.finding_kind = 'finding' THEN 'curated_narrative'::text ELSE 'working_material'::text END,
       to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'finding'`,
		},
		{
			Prefix: "comm_log",
			SQL: `SELECT a.record_id::text, 'comm_log'::text, 'working_material'::text, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'comm_log'`,
		},
		{
			Prefix: "handoffs",
			SQL: `SELECT a.record_id::text, 'handoff'::text, 'working_material'::text, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'handoff'`,
		},
		{
			Prefix: "status_reviews",
			SQL: `SELECT a.record_id::text, 'status_review'::text, 'working_material'::text, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'status_review'`,
		},
		{
			Prefix: "lessons",
			SQL: `SELECT a.record_id::text, 'lesson'::text, 'working_material'::text, to_jsonb(a) - 'incident_id'
  FROM artifact_grid_projection a
  JOIN records r ON r.incident_id = a.incident_id AND r.record_id = a.record_id AND r.deleted_at IS NULL
 WHERE a.incident_id = $1 AND a.artifact_type = 'lesson'`,
		},
		{
			Prefix: "relationships",
			SQL: `SELECT rl.record_link_id::text, 'record_link'::text, 'derived_analytic'::text, to_jsonb(rl) - 'incident_id'
  FROM record_links rl
 WHERE rl.incident_id = $1 AND rl.deleted_at IS NULL`,
		},
		{
			Prefix: "tags",
			SQL: `SELECT rt.record_tag_id::text, 'record_tag'::text, 'derived_analytic'::text, to_jsonb(rt) - 'incident_id'
  FROM record_tags rt
 WHERE rt.incident_id = $1 AND rt.deleted_at IS NULL`,
		},
		{
			Prefix: "entity_mentions",
			SQL: `SELECT em.entity_mention_id::text, 'entity_mention'::text, 'source_evidence'::text, to_jsonb(em)
  FROM entity_mentions em
  JOIN records r ON r.record_id = em.source_record_id AND r.deleted_at IS NULL
 WHERE r.incident_id = $1`,
		},
	}
	fields := []ExportField{}
	for _, query := range queries {
		rows, err := tx.Query(ctx, query.SQL, incidentID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			var sourceFamily string
			var contentClass string
			var raw []byte
			if err := rows.Scan(&id, &sourceFamily, &contentClass, &raw); err != nil {
				rows.Close()
				return nil, err
			}
			var value any
			if err := json.Unmarshal(raw, &value); err != nil {
				rows.Close()
				return nil, err
			}
			field := ExportField{
				Path:         fmt.Sprintf("/%s/%s", query.Prefix, id),
				ContentClass: contentClass,
				SourceFamily: sourceFamily,
				Value:        value,
				SupportRefs:  cloneStrings(supportRefs[id]),
			}
			if sourceFamily == "party" {
				field.DisclosurePartitionRefs = []string{"party:" + id}
			}
			fields = append(fields, field)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})
	return fields, nil
}

func collectSupportRefsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) (map[string][]string, error) {
	rows, err := tx.Query(ctx, `
SELECT src_record_id::text, dst_record_id::text
  FROM record_links
 WHERE incident_id = $1
   AND deleted_at IS NULL
   AND link_type IN ('supported_by', 'references_record', 'attached_evidence')
 ORDER BY src_record_id::text ASC, dst_record_id::text ASC
`, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]string{}
	for rows.Next() {
		var src string
		var dst string
		if err := rows.Scan(&src, &dst); err != nil {
			return nil, err
		}
		out[src] = append(out[src], "/record_envelopes/"+dst)
	}
	return out, rows.Err()
}

func snapshotRecordFromSQL(row any) (SnapshotRecord, error) {
	switch r := row.(type) {
	case sqlc.ReportingSnapshot:
		return snapshotRecordFromSQLFields(r.SnapshotID, r.IncidentID, r.CreatedByUserID, r.ClientTxnID, r.SnapshotAt, r.SourceChangeSetHighWatermark, r.SourceBoundaryJson, r.DerivationVersion, r.ExportModelSha256, r.ExportModelJson, r.CreateJobID, r.CreatedAt)
	case sqlc.CreateReportingSnapshotRow:
		return snapshotRecordFromSQLFields(r.SnapshotID, r.IncidentID, r.CreatedByUserID, r.ClientTxnID, r.SnapshotAt, r.SourceChangeSetHighWatermark, r.SourceBoundaryJson, r.DerivationVersion, r.ExportModelSha256, r.ExportModelJson, r.CreateJobID, r.CreatedAt)
	case sqlc.GetReportingSnapshotRow:
		return snapshotRecordFromSQLFields(r.SnapshotID, r.IncidentID, r.CreatedByUserID, r.ClientTxnID, r.SnapshotAt, r.SourceChangeSetHighWatermark, r.SourceBoundaryJson, r.DerivationVersion, r.ExportModelSha256, r.ExportModelJson, r.CreateJobID, r.CreatedAt)
	case sqlc.GetReportingSnapshotByCreateJobRow:
		return snapshotRecordFromSQLFields(r.SnapshotID, r.IncidentID, r.CreatedByUserID, r.ClientTxnID, r.SnapshotAt, r.SourceChangeSetHighWatermark, r.SourceBoundaryJson, r.DerivationVersion, r.ExportModelSha256, r.ExportModelJson, r.CreateJobID, r.CreatedAt)
	default:
		return SnapshotRecord{}, fmt.Errorf("unsupported snapshot SQL row %T", row)
	}
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
	rowExportModelSHA256 string,
	rowExportModelJSON []byte,
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
		ExportModelSHA256:            rowExportModelSHA256,
		ExportModelJSON:              append([]byte(nil), rowExportModelJSON...),
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
		OutputMediaType:              optionalStringFromPG(row.OutputMediaType),
		OutputSHA256:                 optionalStringFromPG(row.OutputSha256),
		RedactionManifestSHA256:      optionalStringFromPG(row.RedactionManifestSha256),
		RedactionManifestJSON:        append([]byte(nil), row.RedactionManifestJson...),
		RenderedOutput:               optionalStringFromPG(row.RenderedOutput),
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

func existingResponseJobID(responseJSON []byte) (uuid.UUID, error) {
	var resource jobs.Resource
	if err := json.Unmarshal(responseJSON, &resource); err != nil {
		return uuid.UUID{}, err
	}
	jobID, err := uuid.Parse(resource.JobID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return jobID, nil
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
	var payload snapshotCreateJobPayload
	if err := json.Unmarshal(row.RequestJson, &payload); err != nil {
		return snapshotCreateJobPayload{}, err
	}
	return payload, nil
}

func getReleasePayloadTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (releaseCreateJobPayload, error) {
	row, err := sqlc.New(tx).GetReportingJobPayload(ctx, pgUUID(jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return releaseCreateJobPayload{}, ErrNotFound
	}
	if err != nil {
		return releaseCreateJobPayload{}, err
	}
	if row.JobKind != "release_create" {
		return releaseCreateJobPayload{}, fmt.Errorf("reporting job %s has kind %q", jobID, row.JobKind)
	}
	var payload releaseCreateJobPayload
	if err := json.Unmarshal(row.RequestJson, &payload); err != nil {
		return releaseCreateJobPayload{}, err
	}
	return payload, nil
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

func snapshotResource(record SnapshotRecord) map[string]any {
	return map[string]any{
		"snapshot_id":                      record.SnapshotID.String(),
		"incident_id":                      record.IncidentID.String(),
		"created_by_user_id":               record.CreatedByUserID.String(),
		"created_at":                       record.CreatedAt.UTC(),
		"snapshot_at":                      record.SnapshotAt.UTC(),
		"source_change_set_high_watermark": record.SourceChangeSetHighWatermark,
		"derivation_version":               record.DerivationVersion,
		"export_model_sha256":              record.ExportModelSHA256,
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
		"output_media_type":                record.OutputMediaType,
		"release_scope":                    record.ReleaseScope,
		"recipient_partition_refs":         stringArrayForResource(record.RecipientPartitionRefs),
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

func invalidatePriorCandidateTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, snapshotID uuid.UUID, outputKind string, releaseScope string, templateID string, templateVersion string, redactionProfileID string, redactionProfileVersion string, recipientPartitionRefs []byte, now time.Time) error {
	return sqlc.New(tx).InvalidateSupersededReportingReleases(ctx, sqlc.InvalidateSupersededReportingReleasesParams{
		SnapshotID:              pgUUID(snapshotID),
		OutputKind:              outputKind,
		ReleaseScope:            releaseScope,
		TemplateID:              templateID,
		TemplateVersion:         templateVersion,
		RedactionProfileID:      redactionProfileID,
		RedactionProfileVersion: redactionProfileVersion,
		RecipientPartitionRefs:  recipientPartitionRefs,
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
		"output_media_type":                derefString(release.OutputMediaType),
		"output_sha256":                    derefString(release.OutputSHA256),
		"redaction_manifest_sha256":        derefString(release.RedactionManifestSHA256),
		"release_scope":                    release.ReleaseScope,
		"recipient_partition_refs":         stringArrayForResource(release.RecipientPartitionRefs),
		"source_change_set_high_watermark": release.SourceChangeSetHighWatermark,
	})
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.FromBytes(value.Bytes[:])
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
