package reference_data

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

var ErrNotFound = errors.New("reference_data: not found")

type Store struct {
	pool *pgxpool.Pool
}

type ImportAcceptedParams struct {
	ActorUserID       uuid.UUID
	Request           ImportMetadataRequest
	BundleSHA256      string
	BundleStagingPath string
	NormalizedRequest []byte
	Now               time.Time
}

type JobAcceptedResult struct {
	Job      jobs.Resource
	Replayed bool
}

type JobPayload struct {
	JobID             uuid.UUID
	JobKind           string
	ActorUserID       uuid.UUID
	PackKey           *string
	PackVersion       *string
	ResolvedPackKeys  []string
	BundleSHA256      *string
	BundleStagingPath *string
	RequestJSON       json.RawMessage
	CreatedAt         time.Time
}

type ActionParams struct {
	ActorUserID uuid.UUID
	PackKey     string
	PackVersion string
	Request     ActionRequest
	Now         time.Time
}

type ActionResult struct {
	Payload  map[string]any
	Replayed bool
}

type RefreshAcceptedParams struct {
	ActorUserID       uuid.UUID
	Request           RefreshRequest
	NormalizedRequest []byte
	Now               time.Time
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) ListVersions(ctx context.Context) ([]VersionRecord, error) {
	rows, err := s.pool.Query(ctx, versionSelectSQL()+`
ORDER BY rp.pack_key ASC, rp.version ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVersions(rows)
}

func (s *Store) ListVersionsForPackKeys(ctx context.Context, packKeys []string) ([]VersionRecord, error) {
	if len(packKeys) == 0 {
		return []VersionRecord{}, nil
	}
	rows, err := s.pool.Query(ctx, versionSelectSQL()+`
WHERE rp.pack_key = ANY($1)
ORDER BY rp.pack_key ASC, rp.version ASC`, packKeys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanVersions(rows)
}

func (s *Store) ListPackKeys(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT pack_key FROM reference_packs ORDER BY pack_key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var packKey string
		if err := rows.Scan(&packKey); err != nil {
			return nil, err
		}
		out = append(out, packKey)
	}
	return out, rows.Err()
}

func (s *Store) GetVersion(ctx context.Context, packKey string, packVersion string) (VersionRecord, error) {
	record, err := scanVersion(s.pool.QueryRow(ctx, versionSelectSQL()+`
WHERE rp.pack_key = $1 AND rp.version = $2`, packKey, packVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return VersionRecord{}, ErrNotFound
	}
	return record, err
}

func (s *Store) AcceptImport(ctx context.Context, params ImportAcceptedParams) (JobAcceptedResult, error) {
	requestHash := hashBytes(params.NormalizedRequest)
	key := authn.RouteIdempotencyKey{
		RouteKey:    "reference_packs.import",
		ActorUserID: params.ActorUserID,
		ScopeKey:    "deployment",
		ClientTxnID: params.Request.ClientTxnID,
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return JobAcceptedResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if existing, err := lookupRouteIdempotencyTx(ctx, tx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
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

	scope := jobs.Scope{Kind: jobs.ScopeKindDeployment}
	admission, err := jobs.NewExtensionJobAdmission(
		ProfileID,
		ImportJobKind,
		key,
		scope,
		params.NormalizedRequest,
	)
	if err != nil {
		return JobAcceptedResult{}, err
	}
	job, err := jobs.CreateQueuedTx(ctx, tx, jobs.CreateParams{
		Scope:             scope,
		SubmittedByUserID: params.ActorUserID,
		AuthPolicy:        jobs.AuthPolicyDeploymentAdmin,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0, Total: intPtr(1)},
		HandlerName:       LifecycleWorkerKind,
		Extension:         admission,
	}, params.Now)
	if err != nil {
		return JobAcceptedResult{}, err
	}

	bundleSHA := params.BundleSHA256
	bundleStagingPath := params.BundleStagingPath
	if err := insertJobPayloadTx(ctx, tx, job.JobID, "import", params.ActorUserID, nil, nil, nil, &bundleSHA, &bundleStagingPath, params.NormalizedRequest, params.Now); err != nil {
		return JobAcceptedResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, job); err != nil {
		return JobAcceptedResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return JobAcceptedResult{}, err
	}
	return JobAcceptedResult{Job: job}, nil
}

func (s *Store) Activate(ctx context.Context, params ActionParams) (ActionResult, error) {
	return s.applyAction(ctx, params, "reference_packs.activate", func(ctx context.Context, tx pgx.Tx) (map[string]any, error) {
		record, err := getVersionTx(ctx, tx, params.PackKey, params.PackVersion)
		if err != nil {
			return nil, err
		}
		state := publicCondition(record.StoredStatus, record.VerificationResult)
		if record.Active {
			return nil, wrapAPIError(referencePackActivationRejected("already_active"))
		}
		if state != ConditionVerifiedAvailable {
			return nil, wrapAPIError(referencePackActivationRejected("not_verified_available"))
		}
		previousActive, err := activeVersionTx(ctx, tx, params.PackKey)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `
INSERT INTO reference_pack_activation_state (
    pack_key, active_version, previous_active_version, activated_at, activated_by_user_id, operator_note
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (pack_key) DO UPDATE
   SET active_version = EXCLUDED.active_version,
       previous_active_version = EXCLUDED.previous_active_version,
       activated_at = EXCLUDED.activated_at,
       activated_by_user_id = EXCLUDED.activated_by_user_id,
       operator_note = EXCLUDED.operator_note
`, params.PackKey, params.PackVersion, previousActive, params.Now, params.ActorUserID, params.Request.Reason)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `
UPDATE reference_packs
   SET activated_at = $3,
       activated_by_user_id = $4,
       previous_active_version = $5
 WHERE pack_key = $1 AND version = $2
`, params.PackKey, params.PackVersion, params.Now, params.ActorUserID, previousActive)
		if err != nil {
			return nil, err
		}
		updated, err := getVersionTx(ctx, tx, params.PackKey, params.PackVersion)
		if err != nil {
			return nil, err
		}
		if err := insertAttestationTx(ctx, tx, attestationFromRecord(updated, "activate", params.ActorUserID, nil, params.Now, params.Request.Reason)); err != nil {
			return nil, err
		}
		return map[string]any{"pack_version": updated.Resource()}, nil
	})
}

func (s *Store) Disable(ctx context.Context, params ActionParams) (ActionResult, error) {
	return s.applyAction(ctx, params, "reference_packs.disable", func(ctx context.Context, tx pgx.Tx) (map[string]any, error) {
		record, err := getVersionTx(ctx, tx, params.PackKey, params.PackVersion)
		if err != nil {
			return nil, err
		}
		state := publicCondition(record.StoredStatus, record.VerificationResult)
		if state == ConditionDisabled {
			return nil, wrapAPIError(referencePackStateConflict("already_disabled"))
		}
		if state != ConditionVerifiedAvailable {
			return nil, wrapAPIError(referencePackStateConflict("not_disableable"))
		}
		_, err = tx.Exec(ctx, `
UPDATE reference_packs
   SET status = 'disabled'
 WHERE pack_key = $1 AND version = $2
`, params.PackKey, params.PackVersion)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `
UPDATE reference_pack_activation_state
   SET previous_active_version = active_version,
       active_version = NULL
 WHERE pack_key = $1 AND active_version = $2
`, params.PackKey, params.PackVersion)
		if err != nil {
			return nil, err
		}
		updated, err := getVersionTx(ctx, tx, params.PackKey, params.PackVersion)
		if err != nil {
			return nil, err
		}
		if err := insertAttestationTx(ctx, tx, attestationFromRecord(updated, "disable", params.ActorUserID, nil, params.Now, params.Request.Reason)); err != nil {
			return nil, err
		}
		return map[string]any{"pack_version": updated.Resource()}, nil
	})
}

func (s *Store) AcceptReverify(ctx context.Context, params ActionParams) (JobAcceptedResult, error) {
	record, err := s.GetVersion(ctx, params.PackKey, params.PackVersion)
	if errors.Is(err, ErrNotFound) {
		return JobAcceptedResult{}, err
	}
	if err != nil {
		return JobAcceptedResult{}, err
	}
	if publicCondition(record.StoredStatus, record.VerificationResult) == ConditionStaged {
		return JobAcceptedResult{}, wrapAPIError(referencePackStateConflict("verification_pending"))
	}
	return s.acceptJob(ctx, acceptJobParams{
		RouteKey:          "reference_packs.reverify",
		ScopeKey:          params.PackKey + "/" + params.PackVersion,
		ClientTxnID:       params.Request.ClientTxnID,
		ActorUserID:       params.ActorUserID,
		JobKind:           "reverify",
		PackKey:           &params.PackKey,
		PackVersion:       &params.PackVersion,
		NormalizedRequest: params.Request.Normalized,
		Now:               params.Now,
	})
}

func (s *Store) AcceptRefresh(ctx context.Context, params RefreshAcceptedParams) (JobAcceptedResult, error) {
	return s.acceptJob(ctx, acceptJobParams{
		RouteKey:          "reference_packs.refresh",
		ScopeKey:          "deployment",
		ClientTxnID:       params.Request.ClientTxnID,
		ActorUserID:       params.ActorUserID,
		JobKind:           "refresh",
		ResolvedPackKeys:  params.Request.ResolvedPackKeys,
		NormalizedRequest: params.NormalizedRequest,
		Now:               params.Now,
	})
}

func (s *Store) LookupRefreshReplay(ctx context.Context, actorUserID uuid.UUID, clientTxnID string) (JobAcceptedResult, JobPayload, bool, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey:    "reference_packs.refresh",
		ActorUserID: actorUserID,
		ScopeKey:    "deployment",
		ClientTxnID: clientTxnID,
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return JobAcceptedResult{}, JobPayload{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if errors.Is(err, authn.ErrNotFound) {
		return JobAcceptedResult{}, JobPayload{}, false, tx.Commit(ctx)
	}
	if err != nil {
		return JobAcceptedResult{}, JobPayload{}, false, err
	}
	var job jobs.Resource
	if err := json.Unmarshal(existing.ResponseJSON, &job); err != nil {
		return JobAcceptedResult{}, JobPayload{}, false, err
	}
	jobID, err := uuid.Parse(job.JobID)
	if err != nil {
		return JobAcceptedResult{}, JobPayload{}, false, err
	}
	payload, err := getJobPayloadTx(ctx, tx, jobID)
	if err != nil {
		return JobAcceptedResult{}, JobPayload{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return JobAcceptedResult{}, JobPayload{}, false, err
	}
	return JobAcceptedResult{Job: job, Replayed: true}, payload, true, nil
}

func (s *Store) JobPayload(ctx context.Context, jobID uuid.UUID) (JobPayload, error) {
	return getJobPayloadTx(ctx, s.pool, jobID)
}

func (s *Store) CompleteImportVerificationTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, actorUserID uuid.UUID, verification VerificationResult, bundleStoragePath string, now time.Time) (VersionRecord, error) {
	if err := upsertVerifiedPackTx(ctx, tx, actorUserID, verification, bundleStoragePath, now); err != nil {
		return VersionRecord{}, err
	}
	if err := insertAttestationTx(ctx, tx, attestationParams{
		PackKey: verification.PackKey, PackVersion: verification.PackVersion, PackKind: verification.PackKind,
		EventKind: "import", ManifestSHA256: verification.ManifestSHA256, PayloadSHA256: verification.PayloadSHA256,
		SourceIdentifier: verification.SourceIdentifier, VerificationMethod: verification.VerificationMethod,
		SignerKeyID: verification.SignerKeyID, VerificationResult: VerificationPassed, ActorUserID: &actorUserID,
		JobID: &jobID, OccurredAt: now, Metadata: verification.Metadata,
	}); err != nil {
		return VersionRecord{}, err
	}
	_, err := tx.Exec(ctx, `
UPDATE reference_pack_job_payloads
   SET pack_key = $2,
       pack_version = $3,
       bundle_sha256 = $4
 WHERE job_id = $1
`, jobID, verification.PackKey, verification.PackVersion, verification.BundleSHA256)
	if err != nil {
		return VersionRecord{}, err
	}
	return getVersionTx(ctx, tx, verification.PackKey, verification.PackVersion)
}

func (s *Store) ApplyVerificationResult(ctx context.Context, record VersionRecord, verification *VerificationResult, verificationErr *VerificationError, eventKind string, actorUserID uuid.UUID, jobID uuid.UUID, now time.Time) (VersionRecord, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return VersionRecord{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	updated, err := s.ApplyVerificationResultTx(ctx, tx, record, verification, verificationErr, eventKind, actorUserID, jobID, now)
	if err != nil {
		return VersionRecord{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return VersionRecord{}, err
	}
	return updated, nil
}

func (s *Store) ApplyVerificationResultTx(ctx context.Context, tx pgx.Tx, record VersionRecord, verification *VerificationResult, verificationErr *VerificationError, eventKind string, actorUserID uuid.UUID, jobID uuid.UUID, now time.Time) (VersionRecord, error) {
	if verificationErr != nil {
		status := "failed"
		if verificationErr.ReasonCode == "payload_missing" {
			status = "missing"
		}
		_, err := tx.Exec(ctx, `
UPDATE reference_packs
   SET status = $3,
       verification_result = 'failed'
 WHERE pack_key = $1 AND version = $2
`, record.PackKey, record.PackVersion, status)
		if err != nil {
			return VersionRecord{}, err
		}
		_, err = tx.Exec(ctx, `
UPDATE reference_pack_activation_state
   SET previous_active_version = active_version,
       active_version = NULL
 WHERE pack_key = $1 AND active_version = $2
`, record.PackKey, record.PackVersion)
		if err != nil {
			return VersionRecord{}, err
		}
		updated, err := getVersionTx(ctx, tx, record.PackKey, record.PackVersion)
		if err != nil {
			return VersionRecord{}, err
		}
		meta := map[string]any{"reason_code": verificationErr.ReasonCode}
		if err := insertAttestationTx(ctx, tx, attestationFromRecord(updated, eventKind, actorUserID, &jobID, now, nil, meta)); err != nil {
			return VersionRecord{}, err
		}
		return updated, nil
	}
	if verification == nil {
		return VersionRecord{}, fmt.Errorf("missing verification result")
	}
	_, err := tx.Exec(ctx, `
UPDATE reference_packs
   SET pack_kind = $3,
       source_identifier = $4,
       manifest_sha256 = $5,
       payload_sha256 = $6,
       pack_contract_version = $7,
       verification_method = $8,
       signer_key_id = $9,
       status = 'available',
       verification_result = 'passed',
       bundle_sha256 = $10,
       metadata = $11
 WHERE pack_key = $1 AND version = $2
`, record.PackKey, record.PackVersion, verification.PackKind, verification.SourceIdentifier, verification.ManifestSHA256, verification.PayloadSHA256, verification.PackContractVersion, verification.VerificationMethod, verification.SignerKeyID, verification.BundleSHA256, mustJSON(verification.Metadata))
	if err != nil {
		return VersionRecord{}, err
	}
	updated, err := getVersionTx(ctx, tx, record.PackKey, record.PackVersion)
	if err != nil {
		return VersionRecord{}, err
	}
	if err := insertAttestationTx(ctx, tx, attestationFromRecord(updated, eventKind, actorUserID, &jobID, now, nil)); err != nil {
		return VersionRecord{}, err
	}
	return updated, nil
}

func (s *Store) applyAction(ctx context.Context, params ActionParams, routeKey string, apply func(context.Context, pgx.Tx) (map[string]any, error)) (ActionResult, error) {
	requestHash := hashBytes(params.Request.Normalized)
	key := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.PackKey + "/" + params.PackVersion,
		ClientTxnID: params.Request.ClientTxnID,
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ActionResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if existing, err := lookupRouteIdempotencyTx(ctx, tx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return ActionResult{}, authn.ErrClientTxnConflict
		}
		var payload map[string]any
		if err := json.Unmarshal(existing.ResponseJSON, &payload); err != nil {
			return ActionResult{}, err
		}
		return ActionResult{Payload: payload, Replayed: true}, tx.Commit(ctx)
	} else if !errors.Is(err, authn.ErrNotFound) {
		return ActionResult{}, err
	}
	payload, err := apply(ctx, tx)
	if err != nil {
		return ActionResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload); err != nil {
		return ActionResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{Payload: payload}, nil
}

type acceptJobParams struct {
	RouteKey          string
	ScopeKey          string
	ClientTxnID       string
	ActorUserID       uuid.UUID
	JobKind           string
	PackKey           *string
	PackVersion       *string
	ResolvedPackKeys  []string
	NormalizedRequest []byte
	Now               time.Time
}

func (s *Store) acceptJob(ctx context.Context, params acceptJobParams) (JobAcceptedResult, error) {
	requestHash := hashBytes(params.NormalizedRequest)
	key := authn.RouteIdempotencyKey{
		RouteKey:    params.RouteKey,
		ActorUserID: params.ActorUserID,
		ScopeKey:    params.ScopeKey,
		ClientTxnID: params.ClientTxnID,
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return JobAcceptedResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if existing, err := lookupRouteIdempotencyTx(ctx, tx, key); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
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
	scope := jobs.Scope{Kind: jobs.ScopeKindDeployment}
	admission, err := jobs.NewExtensionJobAdmission(
		ProfileID,
		referencePackContractJobKind(params.JobKind),
		key,
		scope,
		params.NormalizedRequest,
	)
	if err != nil {
		return JobAcceptedResult{}, err
	}
	job, err := jobs.CreateQueuedTx(ctx, tx, jobs.CreateParams{
		Scope:             scope,
		SubmittedByUserID: params.ActorUserID,
		AuthPolicy:        jobs.AuthPolicyDeploymentAdmin,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0, Total: intPtr(1)},
		HandlerName:       LifecycleWorkerKind,
		Extension:         admission,
	}, params.Now)
	if err != nil {
		return JobAcceptedResult{}, err
	}
	if err := insertJobPayloadTx(ctx, tx, job.JobID, params.JobKind, params.ActorUserID, params.PackKey, params.PackVersion, params.ResolvedPackKeys, nil, nil, params.NormalizedRequest, params.Now); err != nil {
		return JobAcceptedResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, job); err != nil {
		return JobAcceptedResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return JobAcceptedResult{}, err
	}
	return JobAcceptedResult{Job: job}, nil
}

func referencePackContractJobKind(localKind string) string {
	switch localKind {
	case "import":
		return ImportJobKind
	case "reverify":
		return ReverifyJobKind
	case "refresh":
		return RefreshJobKind
	default:
		return ""
	}
}

func versionSelectSQL() string {
	return `
SELECT rp.pack_key,
       rp.pack_kind,
       rp.version,
       rp.status,
       COALESCE(rpas.active_version = rp.version, false) AS active,
       rp.source_identifier,
       rp.manifest_sha256,
       rp.payload_sha256,
       rp.pack_contract_version,
       rp.verification_method,
       rp.verification_result,
       rp.signer_key_id,
       rp.previous_active_version,
       rp.imported_by_user_id::text,
       rp.imported_at,
       rp.activated_by_user_id::text,
       rp.activated_at,
       rp.bundle_sha256,
       rp.bundle_storage_path
  FROM reference_packs rp
  LEFT JOIN reference_pack_activation_state rpas ON rpas.pack_key = rp.pack_key
`
}

func scanVersions(rows pgx.Rows) ([]VersionRecord, error) {
	var out []VersionRecord
	for rows.Next() {
		record, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, rows.Err()
}

type versionScanner interface {
	Scan(dest ...any) error
}

func scanVersion(row versionScanner) (VersionRecord, error) {
	var record VersionRecord
	var sourceIdentifier sql.NullString
	var signerKeyID sql.NullString
	var previousActive sql.NullString
	var importedBy sql.NullString
	var activatedBy sql.NullString
	var activatedAt sql.NullTime
	err := row.Scan(
		&record.PackKey,
		&record.PackKind,
		&record.PackVersion,
		&record.StoredStatus,
		&record.Active,
		&sourceIdentifier,
		&record.ManifestSHA256,
		&record.PayloadSHA256,
		&record.PackContractVersion,
		&record.VerificationMethod,
		&record.VerificationResult,
		&signerKeyID,
		&previousActive,
		&importedBy,
		&record.ImportedAt,
		&activatedBy,
		&activatedAt,
		&record.BundleSHA256,
		&record.BundleStoragePath,
	)
	if err != nil {
		return VersionRecord{}, err
	}
	record.SourceIdentifier = nullStringPtr(sourceIdentifier)
	record.SignerKeyID = nullStringPtr(signerKeyID)
	record.PreviousActive = nullStringPtr(previousActive)
	record.ImportedByUserID = nullStringPtr(importedBy)
	record.ActivatedByUserID = nullStringPtr(activatedBy)
	if activatedAt.Valid {
		record.ActivatedAt = &activatedAt.Time
	}
	return record, nil
}

func getVersionTx(ctx context.Context, tx pgx.Tx, packKey string, packVersion string) (VersionRecord, error) {
	record, err := scanVersion(tx.QueryRow(ctx, versionSelectSQL()+`
WHERE rp.pack_key = $1 AND rp.version = $2
FOR UPDATE OF rp`, packKey, packVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return VersionRecord{}, ErrNotFound
	}
	return record, err
}

func activeVersionTx(ctx context.Context, tx pgx.Tx, packKey string) (*string, error) {
	var active sql.NullString
	err := tx.QueryRow(ctx, `
SELECT active_version
  FROM reference_pack_activation_state
 WHERE pack_key = $1
 FOR UPDATE
`, packKey).Scan(&active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return nullStringPtr(active), nil
}

func upsertVerifiedPackTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, verification VerificationResult, bundleStoragePath string, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO reference_packs (
    pack_key, version, pack_kind, source_identifier, manifest_sha256, payload_sha256,
    pack_contract_version, verification_method, signer_key_id, status, imported_at,
    imported_by_user_id, verification_result, bundle_sha256, bundle_storage_path, metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'available', $10, $11, 'passed', $12, $13, $14)
ON CONFLICT (pack_key, version) DO UPDATE
   SET pack_kind = EXCLUDED.pack_kind,
       source_identifier = EXCLUDED.source_identifier,
       manifest_sha256 = EXCLUDED.manifest_sha256,
       payload_sha256 = EXCLUDED.payload_sha256,
       pack_contract_version = EXCLUDED.pack_contract_version,
       verification_method = EXCLUDED.verification_method,
       signer_key_id = EXCLUDED.signer_key_id,
       status = EXCLUDED.status,
       imported_at = EXCLUDED.imported_at,
       imported_by_user_id = EXCLUDED.imported_by_user_id,
       verification_result = EXCLUDED.verification_result,
       bundle_sha256 = EXCLUDED.bundle_sha256,
       bundle_storage_path = EXCLUDED.bundle_storage_path,
       metadata = EXCLUDED.metadata
`, verification.PackKey, verification.PackVersion, verification.PackKind, verification.SourceIdentifier, verification.ManifestSHA256, verification.PayloadSHA256, verification.PackContractVersion, verification.VerificationMethod, verification.SignerKeyID, now, actorUserID, verification.BundleSHA256, bundleStoragePath, mustJSON(verification.Metadata))
	return err
}

type attestationParams struct {
	PackKey               string
	PackVersion           string
	PackKind              string
	EventKind             string
	ManifestSHA256        string
	PayloadSHA256         string
	SourceIdentifier      *string
	VerificationMethod    string
	SignerKeyID           *string
	PreviousActiveVersion *string
	VerificationResult    string
	ActorUserID           *uuid.UUID
	JobID                 *uuid.UUID
	OccurredAt            time.Time
	OperatorNote          *string
	Metadata              map[string]any
}

func attestationFromRecord(record VersionRecord, eventKind string, actorUserID uuid.UUID, jobID *uuid.UUID, now time.Time, note *string, metadata ...map[string]any) attestationParams {
	meta := map[string]any{}
	if len(metadata) > 0 && metadata[0] != nil {
		meta = metadata[0]
	}
	return attestationParams{
		PackKey: record.PackKey, PackVersion: record.PackVersion, PackKind: record.PackKind,
		EventKind: eventKind, ManifestSHA256: record.ManifestSHA256, PayloadSHA256: record.PayloadSHA256,
		SourceIdentifier: record.SourceIdentifier, VerificationMethod: record.VerificationMethod,
		SignerKeyID: record.SignerKeyID, PreviousActiveVersion: record.PreviousActive,
		VerificationResult: record.VerificationResult, ActorUserID: &actorUserID, JobID: jobID,
		OccurredAt: now, OperatorNote: note, Metadata: meta,
	}
}

func insertAttestationTx(ctx context.Context, tx pgx.Tx, params attestationParams) error {
	_, err := tx.Exec(ctx, `
INSERT INTO reference_pack_attestations (
    pack_key, pack_version, pack_kind, event_kind, manifest_sha256, payload_sha256,
    source_identifier, verification_method, signer_key_id, previous_active_version,
    verification_result, actor_user_id, job_id, occurred_at, operator_note, metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
`, params.PackKey, params.PackVersion, params.PackKind, params.EventKind, params.ManifestSHA256, params.PayloadSHA256, params.SourceIdentifier, params.VerificationMethod, params.SignerKeyID, params.PreviousActiveVersion, params.VerificationResult, params.ActorUserID, params.JobID, params.OccurredAt, params.OperatorNote, mustJSON(params.Metadata))
	return err
}

func insertJobPayloadTx(ctx context.Context, tx pgx.Tx, jobID string, jobKind string, actorUserID uuid.UUID, packKey *string, packVersion *string, resolvedPackKeys []string, bundleSHA *string, bundleStagingPath *string, normalizedRequest []byte, now time.Time) error {
	if resolvedPackKeys == nil {
		resolvedPackKeys = []string{}
	}
	_, err := tx.Exec(ctx, `
INSERT INTO reference_pack_job_payloads (
    job_id, job_kind, actor_user_id, pack_key, pack_version, resolved_pack_keys,
    bundle_sha256, bundle_staging_path, request_json, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`, jobID, jobKind, actorUserID, packKey, packVersion, resolvedPackKeys, bundleSHA, bundleStagingPath, json.RawMessage(normalizedRequest), now)
	return err
}

type jobPayloadQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func getJobPayloadTx(ctx context.Context, q jobPayloadQuerier, jobID uuid.UUID) (JobPayload, error) {
	var payload JobPayload
	var packKey sql.NullString
	var packVersion sql.NullString
	var bundleSHA sql.NullString
	var stagingPath sql.NullString
	err := q.QueryRow(ctx, `
SELECT job_id, job_kind, actor_user_id, pack_key, pack_version, resolved_pack_keys,
       bundle_sha256, bundle_staging_path, request_json, created_at
  FROM reference_pack_job_payloads
 WHERE job_id = $1
`, jobID).Scan(&payload.JobID, &payload.JobKind, &payload.ActorUserID, &packKey, &packVersion, &payload.ResolvedPackKeys, &bundleSHA, &stagingPath, &payload.RequestJSON, &payload.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return JobPayload{}, ErrNotFound
	}
	if err != nil {
		return JobPayload{}, err
	}
	payload.PackKey = nullStringPtr(packKey)
	payload.PackVersion = nullStringPtr(packVersion)
	payload.BundleSHA256 = nullStringPtr(bundleSHA)
	payload.BundleStagingPath = nullStringPtr(stagingPath)
	if payload.ResolvedPackKeys == nil {
		payload.ResolvedPackKeys = []string{}
	}
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

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func mustJSON(value any) json.RawMessage {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return payload
}

func intPtr(value int) *int {
	return &value
}

func sortedRefs(records []VersionRecord) []jobs.ResourceRef {
	refs := make([]jobs.ResourceRef, 0, len(records))
	for _, record := range records {
		ref := referencePackResourceRef(record.PackKey, record.PackVersion)
		refs = append(refs, jobs.ResourceRef(ref))
	}
	sort.Slice(refs, func(i int, j int) bool {
		return refs[i].Route < refs[j].Route
	})
	return refs
}
