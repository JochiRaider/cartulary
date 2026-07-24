package incidentbundles

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var ErrNotFound = errors.New("incident bundle: not found")

type Store struct {
	pool *pgxpool.Pool
}

type JobAcceptedResult struct {
	Job      jobs.Resource
	Replayed bool
}

type JobPayload struct {
	JobID              uuid.UUID
	JobKind            string
	ActorUserID        uuid.UUID
	IncidentID         *uuid.UUID
	BundleID           *uuid.UUID
	UploadedSHA256     *string
	BundleStagingPath  *string
	ImportedIncidentID *uuid.UUID
	ManifestSHA256     *string
	RequestJSON        json.RawMessage
	CreatedAt          time.Time
}

type ExportAcceptedParams struct {
	ActorUserID       uuid.UUID
	Request           ExportRequest
	NormalizedRequest []byte
	Now               time.Time
}

type ImportAcceptedParams struct {
	ActorUserID       uuid.UUID
	Request           ImportMetadataRequest
	UploadedSHA256    string
	BundleStagingPath string
	NormalizedRequest []byte
	Now               time.Time
}

type ExportCompleteParams struct {
	JobID                uuid.UUID
	ActorUserID          uuid.UUID
	IncidentID           uuid.UUID
	BundleID             uuid.UUID
	ExportedAt           time.Time
	ManifestSHA256       string
	ReferencePackMode    string
	OptionalSections     []string
	RequiredCapabilities []string
	BundleSHA256         string
	BundleByteSize       int64
	BundleStoragePath    string
	ManifestFiles        []ManifestFile
}

type DescriptorRecord struct {
	BundleID             uuid.UUID
	IncidentID           uuid.UUID
	ExportedAt           time.Time
	ManifestSHA256       string
	ReferencePackMode    string
	HistoryMode          string
	BlobMode             string
	OptionalSections     []string
	RequiredCapabilities []string
	BundleSHA256         string
	BundleByteSize       int64
	BundleStoragePath    string
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) AcceptExport(ctx context.Context, params ExportAcceptedParams) (JobAcceptedResult, error) {
	requestHash := hashBytes(params.NormalizedRequest)
	key := authn.RouteIdempotencyKey{
		RouteKey:    "incident_bundles.export",
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.Request.IncidentID.String(),
		ClientTxnID: params.Request.ClientTxnID,
	}
	return s.acceptJob(ctx, jobAdmissionParams{
		Key:         key,
		RequestHash: requestHash,
		Create: func(ctx context.Context, tx pgx.Tx) (jobs.Resource, error) {
			scope := jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &params.Request.IncidentID}
			admission, err := jobs.NewExtensionJobAdmission(
				IncidentPortabilityProfileID,
				ExportJobKind,
				key,
				scope,
				params.NormalizedRequest,
			)
			if err != nil {
				return jobs.Resource{}, err
			}
			job, err := jobs.CreateQueuedTx(ctx, tx, jobs.CreateParams{
				Scope:             scope,
				SubmittedByUserID: params.ActorUserID,
				AuthPolicy:        jobs.AuthPolicyDeploymentAdminIncidentMembership,
				Cancelable:        true,
				Progress:          jobs.Progress{Completed: 0, Total: intPtr(1)},
				HandlerName:       incidentBundleJobHandlerName,
				Extension:         admission,
			}, params.Now)
			if err != nil {
				return jobs.Resource{}, err
			}
			jobID, err := uuid.Parse(job.JobID)
			if err != nil {
				return jobs.Resource{}, err
			}
			if err := insertJobPayloadTx(ctx, tx, jobID, "export", params.ActorUserID, &params.Request.IncidentID, nil, nil, nil, params.NormalizedRequest, params.Now); err != nil {
				return jobs.Resource{}, err
			}
			return job, nil
		},
	})
}

func (s *Store) AcceptImport(ctx context.Context, params ImportAcceptedParams) (JobAcceptedResult, error) {
	requestHash := hashBytes(params.NormalizedRequest)
	key := authn.RouteIdempotencyKey{
		RouteKey:    "incident_bundles.import",
		ActorUserID: params.ActorUserID,
		ScopeKey:    "deployment",
		ClientTxnID: params.Request.ClientTxnID,
	}
	return s.acceptJob(ctx, jobAdmissionParams{
		Key:         key,
		RequestHash: requestHash,
		Create: func(ctx context.Context, tx pgx.Tx) (jobs.Resource, error) {
			scope := jobs.Scope{Kind: jobs.ScopeKindDeployment}
			admission, err := jobs.NewExtensionJobAdmission(
				IncidentPortabilityProfileID,
				ImportJobKind,
				key,
				scope,
				params.NormalizedRequest,
			)
			if err != nil {
				return jobs.Resource{}, err
			}
			job, err := jobs.CreateQueuedTx(ctx, tx, jobs.CreateParams{
				Scope:             scope,
				SubmittedByUserID: params.ActorUserID,
				AuthPolicy:        jobs.AuthPolicyDeploymentAdmin,
				Cancelable:        true,
				Progress:          jobs.Progress{Completed: 0, Total: intPtr(1)},
				HandlerName:       incidentBundleJobHandlerName,
				Extension:         admission,
			}, params.Now)
			if err != nil {
				return jobs.Resource{}, err
			}
			jobID, err := uuid.Parse(job.JobID)
			if err != nil {
				return jobs.Resource{}, err
			}
			uploadedSHA := params.UploadedSHA256
			stagingPath := params.BundleStagingPath
			if err := insertJobPayloadTx(ctx, tx, jobID, "import", params.ActorUserID, nil, nil, &uploadedSHA, &stagingPath, params.NormalizedRequest, params.Now); err != nil {
				return jobs.Resource{}, err
			}
			return job, nil
		},
	})
}

type jobAdmissionParams struct {
	Key         authn.RouteIdempotencyKey
	RequestHash []byte
	Create      func(context.Context, pgx.Tx) (jobs.Resource, error)
}

func (s *Store) acceptJob(ctx context.Context, params jobAdmissionParams) (JobAcceptedResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return JobAcceptedResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if existing, err := lookupRouteIdempotencyTx(ctx, tx, params.Key); err == nil {
		if !bytes.Equal(existing.RequestHash, params.RequestHash) {
			return JobAcceptedResult{}, authn.ErrClientTxnConflict
		}
		var job jobs.Resource
		if err := json.Unmarshal(existing.ResponseJSON, &job); err != nil {
			return JobAcceptedResult{}, err
		}
		return JobAcceptedResult{Job: job, Replayed: true}, tx.Commit(ctx)
	} else if !errors.Is(err, authn.ErrNotFound) {
		return JobAcceptedResult{}, err
	}
	job, err := params.Create(ctx, tx)
	if err != nil {
		return JobAcceptedResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, params.Key, nil, params.RequestHash, http.StatusAccepted, job); err != nil {
		return JobAcceptedResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return JobAcceptedResult{}, err
	}
	return JobAcceptedResult{Job: job}, nil
}

func (s *Store) CompleteExportDescriptorTx(ctx context.Context, tx pgx.Tx, params ExportCompleteParams) (DescriptorRecord, error) {
	if tx == nil {
		return DescriptorRecord{}, errors.New("incident bundle export transaction unavailable")
	}
	_, err := tx.Exec(ctx, `
INSERT INTO incident_bundle_exports (
    bundle_id, incident_id, export_job_id, exported_by_user_id, exported_at, manifest_sha256,
    reference_pack_mode, history_mode, blob_mode, optional_sections, required_capabilities,
    bundle_sha256, bundle_byte_size, bundle_storage_path, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'full', 'full', $8, $9, $10, $11, $12, $5)
`, params.BundleID, params.IncidentID, params.JobID, params.ActorUserID, params.ExportedAt, params.ManifestSHA256, params.ReferencePackMode, params.OptionalSections, params.RequiredCapabilities, params.BundleSHA256, params.BundleByteSize, params.BundleStoragePath)
	if err != nil {
		return DescriptorRecord{}, err
	}
	for _, file := range params.ManifestFiles {
		sha := strings.TrimPrefix(file.SHA256, "sha256:")
		_, err := tx.Exec(ctx, `
INSERT INTO incident_bundle_manifest_files (bundle_id, path, sha256, size_bytes, required)
VALUES ($1, $2, $3, $4, $5)
`, params.BundleID, file.Path, sha, file.SizeBytes, file.Required)
		if err != nil {
			return DescriptorRecord{}, err
		}
	}
	tag, err := tx.Exec(ctx, `
UPDATE incident_bundle_job_payloads
   SET bundle_id = $2,
       manifest_sha256 = $3,
       updated_at = $4
 WHERE job_id = $1
`, params.JobID, params.BundleID, params.ManifestSHA256, params.ExportedAt)
	if err != nil {
		return DescriptorRecord{}, err
	}
	if tag.RowsAffected() != 1 {
		return DescriptorRecord{}, ErrNotFound
	}
	return getDescriptorTx(ctx, tx, params.BundleID)
}

func (s *Store) GetDescriptor(ctx context.Context, bundleID uuid.UUID) (DescriptorRecord, error) {
	record, err := getDescriptorTx(ctx, s.pool, bundleID)
	if errors.Is(err, pgx.ErrNoRows) {
		return DescriptorRecord{}, ErrNotFound
	}
	return record, err
}

func (s *Store) GetJobPayload(ctx context.Context, jobID uuid.UUID) (JobPayload, error) {
	payload, err := getJobPayloadTx(ctx, s.pool, jobID)
	if errors.Is(err, pgx.ErrNoRows) {
		return JobPayload{}, ErrNotFound
	}
	return payload, err
}

func (s *Store) ListRecoverableJobPayloads(ctx context.Context) ([]JobPayload, error) {
	rows, err := s.pool.Query(ctx, `
SELECT p.job_id
  FROM incident_bundle_job_payloads p
  JOIN jobs j ON j.job_id = p.job_id
 WHERE j.status IN ('queued', 'running', 'cancel_requested')
 ORDER BY p.created_at ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var payloads []JobPayload
	for rows.Next() {
		var jobID uuid.UUID
		if err := rows.Scan(&jobID); err != nil {
			return nil, err
		}
		payload, err := s.GetJobPayload(ctx, jobID)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, payload)
	}
	return payloads, rows.Err()
}

func MarkImportCompleteTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, incidentID uuid.UUID, manifestSHA string, now time.Time) error {
	if tx == nil {
		return errors.New("incident bundle import transaction unavailable")
	}
	return markImportComplete(ctx, tx, jobID, incidentID, manifestSHA, now)
}

type commandExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func markImportComplete(ctx context.Context, executor commandExecutor, jobID uuid.UUID, incidentID uuid.UUID, manifestSHA string, now time.Time) error {
	tag, err := executor.Exec(ctx, `
UPDATE incident_bundle_job_payloads
   SET imported_incident_id = $2,
       manifest_sha256 = $3,
       updated_at = $4
 WHERE job_id = $1
`, jobID, incidentID, manifestSHA, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) MarkJobFailure(ctx context.Context, jobID uuid.UUID, reason string, now time.Time) {
	_, _ = s.pool.Exec(ctx, `
UPDATE incident_bundle_job_payloads
   SET failure_reason = $2,
       updated_at = $3
 WHERE job_id = $1
`, jobID, reason, now)
}

func (r DescriptorRecord) Resource() map[string]any {
	return map[string]any{
		"bundle_id":             r.BundleID.String(),
		"incident_id":           r.IncidentID.String(),
		"exported_at":           r.ExportedAt,
		"manifest_sha256":       r.ManifestSHA256,
		"reference_pack_mode":   r.ReferencePackMode,
		"history_mode":          r.HistoryMode,
		"blob_mode":             r.BlobMode,
		"optional_sections":     r.OptionalSections,
		"required_capabilities": r.RequiredCapabilities,
	}
}

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func getDescriptorTx(ctx context.Context, q rowQuerier, bundleID uuid.UUID) (DescriptorRecord, error) {
	var record DescriptorRecord
	err := q.QueryRow(ctx, `
SELECT bundle_id, incident_id, exported_at, manifest_sha256, reference_pack_mode,
       history_mode, blob_mode, optional_sections, required_capabilities,
       bundle_sha256, bundle_byte_size, bundle_storage_path
  FROM incident_bundle_exports
 WHERE bundle_id = $1
`, bundleID).Scan(&record.BundleID, &record.IncidentID, &record.ExportedAt, &record.ManifestSHA256, &record.ReferencePackMode, &record.HistoryMode, &record.BlobMode, &record.OptionalSections, &record.RequiredCapabilities, &record.BundleSHA256, &record.BundleByteSize, &record.BundleStoragePath)
	if record.OptionalSections == nil {
		record.OptionalSections = []string{}
	}
	if record.RequiredCapabilities == nil {
		record.RequiredCapabilities = []string{}
	}
	return record, err
}

func insertJobPayloadTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, jobKind string, actorUserID uuid.UUID, incidentID *uuid.UUID, bundleID *uuid.UUID, uploadedSHA *string, stagingPath *string, normalizedRequest []byte, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO incident_bundle_job_payloads (
    job_id, job_kind, actor_user_id, incident_id, bundle_id, uploaded_sha256,
    bundle_staging_path, request_json, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
`, jobID, jobKind, actorUserID, incidentID, bundleID, uploadedSHA, stagingPath, json.RawMessage(normalizedRequest), now)
	return err
}

func getJobPayloadTx(ctx context.Context, q rowQuerier, jobID uuid.UUID) (JobPayload, error) {
	var payload JobPayload
	var incidentID sql.NullString
	var bundleID sql.NullString
	var uploadedSHA sql.NullString
	var stagingPath sql.NullString
	var importedIncidentID sql.NullString
	var manifestSHA sql.NullString
	err := q.QueryRow(ctx, `
SELECT job_id, job_kind, actor_user_id, incident_id::text, bundle_id::text, uploaded_sha256,
       bundle_staging_path, imported_incident_id::text, manifest_sha256, request_json, created_at
  FROM incident_bundle_job_payloads
 WHERE job_id = $1
`, jobID).Scan(&payload.JobID, &payload.JobKind, &payload.ActorUserID, &incidentID, &bundleID, &uploadedSHA, &stagingPath, &importedIncidentID, &manifestSHA, &payload.RequestJSON, &payload.CreatedAt)
	if err != nil {
		return JobPayload{}, err
	}
	payload.IncidentID = uuidPtrFromNullString(incidentID)
	payload.BundleID = uuidPtrFromNullString(bundleID)
	payload.ImportedIncidentID = uuidPtrFromNullString(importedIncidentID)
	payload.UploadedSHA256 = stringPtrFromNull(uploadedSHA)
	payload.BundleStagingPath = stringPtrFromNull(stagingPath)
	payload.ManifestSHA256 = stringPtrFromNull(manifestSHA)
	return payload, nil
}

func lookupRouteIdempotencyTx(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
	var record authn.RouteIdempotencyRecord
	err := tx.QueryRow(ctx, `
SELECT route_key, scope_key, client_txn_id, actor_user_id, request_hash, status_code, response_json
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID).Scan(&record.RouteKey, &record.ScopeKey, &record.ClientTxnID, &record.ActorUserID, &record.RequestHash, &record.StatusCode, &record.ResponseJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return authn.RouteIdempotencyRecord{}, authn.ErrNotFound
	}
	return record, err
}

func uuidPtrFromNullString(value sql.NullString) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed, err := uuid.Parse(value.String)
	if err != nil {
		return nil
	}
	return &parsed
}

func stringPtrFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func intPtr(value int) *int {
	return &value
}
