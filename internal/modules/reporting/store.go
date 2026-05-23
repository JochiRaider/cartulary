package reporting

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
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
	OutputMediaType              string
	OutputSHA256                 string
	RedactionManifestSHA256      string
	RedactionManifestJSON        []byte
	RenderedOutput               string
	CreateJobID                  uuid.UUID
	RenderFailedReasonCode       *string
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
	ActorUserID uuid.UUID
	Request     CreateReleaseRequest
	Rendered    RenderedRelease
	Now         time.Time
}

type CreateReleaseResult struct {
	Job       jobs.Resource
	ReleaseID uuid.UUID
	Replayed  bool
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
	tx, err := s.pool.Begin(ctx)
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
		snapshot, snapshotErr := getSnapshotRecordByCreateJobIDTx(ctx, tx, existingJobID)
		if snapshotErr != nil {
			return CreateSnapshotResult{}, snapshotErr
		}
		if params.Request.SourceChangeSetHighWatermark != nil && *params.Request.SourceChangeSetHighWatermark != snapshot.SourceChangeSetHighWatermark {
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
	watermark := SourceBoundaryForIncident(incident)
	if params.Request.SourceChangeSetHighWatermark != nil && *params.Request.SourceChangeSetHighWatermark != watermark {
		return CreateSnapshotResult{}, &SnapshotBoundaryConflictError{Expected: *params.Request.SourceChangeSetHighWatermark, Actual: watermark}
	}
	normalized, err := normalizeSnapshotRequest(params.Request.IncidentID, params.Request.ClientTxnID, &watermark)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	sum := sha256.Sum256(normalized)
	requestHash := sum[:]
	model, exportSHA, err := BuildExportModel(incident, params.Now.UTC(), watermark)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	exportJSON, err := canonicalJSON(model)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	job, err := jobs.CreateQueuedTx(ctx, tx, jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &params.Request.IncidentID},
		SubmittedByUserID: params.ActorUserID,
		Cancelable:        false,
		Progress:          jobs.Progress{Completed: 0},
	}, params.Now.UTC())
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	jobID := uuid.MustParse(job.JobID)
	snapshotRow, err := sqlc.New(tx).CreateReportingSnapshot(ctx, sqlc.CreateReportingSnapshotParams{
		IncidentID:                   pgUUID(params.Request.IncidentID),
		CreatedByUserID:              pgUUID(params.ActorUserID),
		ClientTxnID:                  params.Request.ClientTxnID,
		SnapshotAt:                   pgTimestamptz(params.Now),
		SourceChangeSetHighWatermark: watermark,
		DerivationVersion:            DerivationVersion,
		ExportModelSha256:            exportSHA,
		ExportModelJson:              exportJSON,
		CreateJobID:                  pgUUID(jobID),
	})
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	snapshot, err := snapshotRecordFromSQL(snapshotRow)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, job); err != nil {
		return CreateSnapshotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateSnapshotResult{}, err
	}
	return CreateSnapshotResult{Job: job, SnapshotID: snapshot.SnapshotID}, nil
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
	var model ExportModel
	if err := json.Unmarshal(record.ExportModelJSON, &model); err != nil {
		return SnapshotRecord{}, ExportModel{}, err
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
	initialState := ReleaseStatePendingApproval
	var approvedAt *time.Time
	if params.Request.ReleaseScope == ReleaseScopeInternalDraft {
		initialState = ReleaseStateApproved
		now := params.Now.UTC()
		approvedAt = &now
	}
	job, err := jobs.CreateQueuedTx(ctx, tx, jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &snapshot.IncidentID},
		SubmittedByUserID: params.ActorUserID,
		Cancelable:        false,
		Progress:          jobs.Progress{Completed: 0},
	}, params.Now.UTC())
	if err != nil {
		return CreateReleaseResult{}, err
	}
	jobID := uuid.MustParse(job.JobID)
	releaseRow, err := sqlc.New(tx).CreateReportingRelease(ctx, sqlc.CreateReportingReleaseParams{
		IncidentID:                   pgUUID(snapshot.IncidentID),
		SnapshotID:                   pgUUID(snapshot.SnapshotID),
		CreatedByUserID:              pgUUID(params.ActorUserID),
		ClientTxnID:                  params.Request.ClientTxnID,
		ReleaseScope:                 params.Request.ReleaseScope,
		ReleaseState:                 initialState,
		SnapshotAt:                   pgTimestamptz(snapshot.SnapshotAt),
		SourceChangeSetHighWatermark: snapshot.SourceChangeSetHighWatermark,
		DerivationVersion:            snapshot.DerivationVersion,
		ExportModelSha256:            snapshot.ExportModelSHA256,
		TemplateID:                   params.Request.TemplateID,
		TemplateVersion:              params.Request.TemplateVersion,
		RedactionProfileID:           params.Rendered.Profile.ProfileID,
		RedactionProfileVersion:      params.Rendered.Profile.Version,
		RedactionProfileSha256:       params.Rendered.ProfileSHA256,
		OutputKind:                   params.Request.OutputKind,
		OutputMediaType:              params.Rendered.OutputMediaType,
		OutputSha256:                 params.Rendered.OutputSHA256,
		RedactionManifestSha256:      params.Rendered.RedactionManifestSHA256,
		RedactionManifestJson:        params.Rendered.RedactionManifestJSON,
		RenderedOutput:               string(params.Rendered.Output),
		CreateJobID:                  pgUUID(jobID),
		ApprovedAt:                   optionalPGTimestamptz(approvedAt),
		CreatedAt:                    pgTimestamptz(params.Now),
	})
	if err != nil {
		return CreateReleaseResult{}, err
	}
	release, err := releaseRecordFromSQL(releaseRow)
	if err != nil {
		return CreateReleaseResult{}, err
	}
	if err := invalidatePriorCandidateTx(ctx, tx, release.ReleaseID, snapshot.SnapshotID, params.Request.OutputKind, params.Request.ReleaseScope, params.Request.TemplateID, params.Request.TemplateVersion, params.Rendered.Profile.ProfileID, params.Rendered.Profile.Version, params.Now.UTC()); err != nil {
		return CreateReleaseResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, job); err != nil {
		return CreateReleaseResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateReleaseResult{}, err
	}
	return CreateReleaseResult{Job: job, ReleaseID: release.ReleaseID}, nil
}

func (s *Store) GetRelease(ctx context.Context, releaseID uuid.UUID) (map[string]any, uuid.UUID, error) {
	record, err := s.getReleaseRecord(ctx, releaseID)
	if err != nil {
		return nil, uuid.UUID{}, err
	}
	return releaseResource(record), record.IncidentID, nil
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
			OutputSha256:            release.OutputSHA256,
			RedactionManifestSha256: release.RedactionManifestSHA256,
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

func snapshotRecordFromSQL(row sqlc.ReportingSnapshot) (SnapshotRecord, error) {
	snapshotID, err := uuidFromPG(row.SnapshotID)
	if err != nil {
		return SnapshotRecord{}, err
	}
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return SnapshotRecord{}, err
	}
	createdBy, err := uuidFromPG(row.CreatedByUserID)
	if err != nil {
		return SnapshotRecord{}, err
	}
	createJobID, err := uuidFromPG(row.CreateJobID)
	if err != nil {
		return SnapshotRecord{}, err
	}
	snapshotAt, err := timeFromPG(row.SnapshotAt)
	if err != nil {
		return SnapshotRecord{}, err
	}
	createdAt, err := timeFromPG(row.CreatedAt)
	if err != nil {
		return SnapshotRecord{}, err
	}
	return SnapshotRecord{
		SnapshotID:                   snapshotID,
		IncidentID:                   incidentID,
		CreatedByUserID:              createdBy,
		ClientTxnID:                  row.ClientTxnID,
		SnapshotAt:                   snapshotAt,
		SourceChangeSetHighWatermark: row.SourceChangeSetHighWatermark,
		DerivationVersion:            row.DerivationVersion,
		ExportModelSHA256:            row.ExportModelSha256,
		ExportModelJSON:              append([]byte(nil), row.ExportModelJson...),
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
		OutputMediaType:              row.OutputMediaType,
		OutputSHA256:                 row.OutputSha256,
		RedactionManifestSHA256:      row.RedactionManifestSha256,
		RedactionManifestJSON:        append([]byte(nil), row.RedactionManifestJson...),
		RenderedOutput:               row.RenderedOutput,
		CreateJobID:                  createJobID,
		RenderFailedReasonCode:       optionalStringFromPG(row.RenderFailedReasonCode),
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
		"output_sha256":                    record.OutputSHA256,
		"redaction_manifest_sha256":        record.RedactionManifestSHA256,
		"release_state":                    record.ReleaseState,
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
	return resource
}

func invalidatePriorCandidateTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, snapshotID uuid.UUID, outputKind string, releaseScope string, templateID string, templateVersion string, redactionProfileID string, redactionProfileVersion string, now time.Time) error {
	return sqlc.New(tx).InvalidateSupersededReportingReleases(ctx, sqlc.InvalidateSupersededReportingReleasesParams{
		SnapshotID:              pgUUID(snapshotID),
		OutputKind:              outputKind,
		ReleaseScope:            releaseScope,
		TemplateID:              templateID,
		TemplateVersion:         templateVersion,
		RedactionProfileID:      redactionProfileID,
		RedactionProfileVersion: redactionProfileVersion,
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
		"output_media_type":                release.OutputMediaType,
		"output_sha256":                    release.OutputSHA256,
		"redaction_manifest_sha256":        release.RedactionManifestSHA256,
		"release_scope":                    release.ReleaseScope,
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

func optionalStringFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	parsed := value.String
	return &parsed
}
