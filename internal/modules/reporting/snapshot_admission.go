package reporting

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const snapshotAdmissionSchema = "cartulary.reporting_snapshot_admission.v1"

// Request admission is small and independent of incident size. A capture is
// immutable only after its complete model has been durably retained. A worker
// restarting before that point may capture again; it cannot replace a capture.
type snapshotAdmission struct {
	SchemaID          string                    `json:"schema_id"`
	JobID             string                    `json:"job_id"`
	IncidentID        string                    `json:"incident_id"`
	ActorUserID       string                    `json:"actor_user_id"`
	ClientTxnID       string                    `json:"client_txn_id"`
	AdmittedAt        time.Time                 `json:"admitted_at"`
	RequestedBoundary *string                   `json:"requested_boundary"`
	ReferencePacks    reference_data.SetBinding `json:"reference_packs"`
}

func (s *Store) admitSnapshotJob(ctx context.Context, params CreateSnapshotParams) (CreateSnapshotResult, error) {
	params.Now = params.Now.UTC().Truncate(time.Microsecond)
	key := authn.RouteIdempotencyKey{RouteKey: "snapshots.create", ActorUserID: params.ActorUserID, ScopeKey: params.Request.IncidentID.String(), ClientTxnID: params.Request.ClientTxnID}
	normalized, err := normalizeSnapshotRequest(params.Request.IncidentID, params.Request.ClientTxnID, params.Request.SourceChangeSetHighWatermark)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	hash := sha256.Sum256(normalized)
	tx, release, err := authn.BeginRepeatableRouteIdempotencyTx(ctx, s.pool, key)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	defer release()
	existing, err := lookupRouteIdempotencyTx(ctx, tx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, hash[:]) {
			return CreateSnapshotResult{}, authn.ErrClientTxnConflict
		}
		var job jobs.Resource
		if err := json.Unmarshal(existing.ResponseJSON, &job); err != nil {
			return CreateSnapshotResult{}, err
		}
		id, err := uuid.Parse(job.JobID)
		if err != nil {
			return CreateSnapshotResult{}, err
		}
		if _, err := readSnapshotAdmissionTx(ctx, tx, id); err != nil {
			return CreateSnapshotResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return CreateSnapshotResult{}, err
		}
		return CreateSnapshotResult{Job: job, JobID: id, Replayed: true}, nil
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return CreateSnapshotResult{}, err
	}
	incident, err := getIncidentTx(ctx, tx, params.Request.IncidentID)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	// An explicit stale boundary can be rejected immediately. The worker also
	// checks it in the capture transaction; a later change is a terminal conflict.
	if expected := params.Request.SourceChangeSetHighWatermark; expected != nil {
		boundary, err := s.exportMaterializer.ResolveSourceBoundaryTx(ctx, tx, incident)
		if err != nil {
			return CreateSnapshotResult{}, err
		}
		if *expected != boundary.Token {
			return CreateSnapshotResult{}, &SnapshotBoundaryConflictError{Expected: *expected, Actual: boundary.Token}
		}
	}
	scope := jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &params.Request.IncidentID}
	extension, err := jobs.NewExtensionJobAdmission(ProfileID, jobs.NewRouteIdempotencyKey(key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID), scope, normalized)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	job, err := s.jobTransactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{JobKind: SnapshotCreateJobKind, Scope: scope, SubmittedByUserID: params.ActorUserID, Cancelable: true, Progress: jobs.Progress{Completed: 0}, Extension: extension}, params.Now)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	id := uuid.MustParse(job.JobID)
	if s.referencePacks == nil {
		return CreateSnapshotResult{}, errors.New("reporting: Reference Pack retention is required")
	}
	binding, err := s.referencePacks.CaptureCurrentTx(ctx, tx, "reporting_snapshot_job", id.String(), id)
	if err != nil {
		var failure *reference_data.ConsumerError
		if errors.As(err, &failure) {
			return CreateSnapshotResult{}, ErrRequiredReferencePackUnavailable
		}
		return CreateSnapshotResult{}, err
	}
	admission := snapshotAdmission{SchemaID: snapshotAdmissionSchema, JobID: id.String(), IncidentID: incident.ID, ActorUserID: params.ActorUserID.String(), ClientTxnID: params.Request.ClientTxnID, AdmittedAt: params.Now, RequestedBoundary: params.Request.SourceChangeSetHighWatermark, ReferencePacks: binding}
	raw, err := canonicalJSON(admission)
	if err != nil {
		return CreateSnapshotResult{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reporting_job_payloads(job_id,job_kind,incident_id,actor_user_id,request_json,snapshot_admission_json,created_at,updated_at,snapshot_admission_sha256) VALUES($1,'snapshot_create',$2,$3,NULL,$4,$5,$5,$6)`, id, params.Request.IncidentID, params.ActorUserID, raw, params.Now, reportingObjectDigest(snapshotAdmissionSchema, raw)); err != nil {
		return CreateSnapshotResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, hash[:], http.StatusAccepted, job); err != nil {
		return CreateSnapshotResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return CreateSnapshotResult{}, err
	}
	return CreateSnapshotResult{Job: job, JobID: id}, nil
}

func readSnapshotAdmissionTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (snapshotAdmission, error) {
	var raw []byte
	var digest string
	var incidentID, actorID uuid.UUID
	var at time.Time
	if err := tx.QueryRow(ctx, `SELECT incident_id,actor_user_id,snapshot_admission_json,created_at,snapshot_admission_sha256 FROM reporting_job_payloads WHERE job_id=$1 AND job_kind='snapshot_create'`, jobID).Scan(&incidentID, &actorID, &raw, &at, &digest); err != nil {
		return snapshotAdmission{}, err
	}
	var input snapshotAdmission
	canonical, err := canonicalReportingBytes(raw)
	if err != nil || json.Unmarshal(canonical, &input) != nil {
		return input, errors.New("reporting: invalid snapshot admission")
	}
	encoded, err := canonicalJSON(input)
	if err != nil || !bytes.Equal(canonical, encoded) || reportingObjectDigest(snapshotAdmissionSchema, canonical) != digest || input.SchemaID != snapshotAdmissionSchema || input.JobID != jobID.String() || input.IncidentID != incidentID.String() || input.ActorUserID != actorID.String() || !input.AdmittedAt.Equal(at) {
		return input, errors.New("reporting: invalid snapshot admission")
	}
	retained, err := jobs.ReadRetainedExtensionJob(ctx, tx, jobID)
	if err != nil {
		return input, err
	}
	payload := snapshotCreateJobPayload{IncidentID: input.IncidentID, ActorUserID: input.ActorUserID, ClientTxnID: input.ClientTxnID, JobAdmittedAt: input.AdmittedAt, RequestedBoundary: input.RequestedBoundary}
	if err := validateSnapshotJobAdmission(payload, jobID, retained); err != nil {
		return input, err
	}
	return input, nil
}

func (s *Store) PrepareSnapshotJob(ctx context.Context, execution jobs.Execution) (resultErr error) {
	stage := "request_validation"
	canonicalBytes := 0
	defer func() {
		if resultErr != nil {
			resultErr = &snapshotWorkError{stage: stage, cause: resultErr, canonicalBytes: canonicalBytes}
		}
	}()

	tx, err := s.beginExecutionRead(ctx, execution)
	if err != nil {
		return err
	}
	defer rollbackSnapshotTx(tx)
	admission, err := readSnapshotAdmissionTx(ctx, tx, execution.JobID())
	if err != nil {
		return err
	}
	var captured bool
	if err := tx.QueryRow(ctx, `SELECT request_json IS NOT NULL FROM reporting_job_payloads WHERE job_id=$1`, execution.JobID()).Scan(&captured); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if captured {
		return nil
	} // Completion independently validates the retained capture.

	// No Jobs lock is held while reading or materializing incident-sized input.
	stage = "source_capture"
	capture, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer rollbackSnapshotTx(capture)
	var at time.Time
	if err := capture.QueryRow(ctx, `SELECT transaction_timestamp()`).Scan(&at); err != nil {
		return err
	}
	if err := s.referencePacks.ValidateBindingTx(ctx, capture, admission.ReferencePacks); err != nil {
		return err
	}
	incidentID := uuid.MustParse(admission.IncidentID)
	incident, err := getIncidentTx(ctx, capture, incidentID)
	if err != nil {
		return err
	}
	boundary, err := s.exportMaterializer.ResolveSourceBoundaryTx(ctx, capture, incident)
	if err != nil {
		return err
	}
	if expected := admission.RequestedBoundary; expected != nil && *expected != boundary.Token {
		return &SnapshotBoundaryConflictError{Expected: *expected, Actual: boundary.Token}
	}
	source, err := s.exportMaterializer.CollectSourceTx(ctx, capture, incidentID)
	if err != nil {
		return err
	}
	if err := capture.Commit(ctx); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	stage = "model_serialization"
	encoded, digest, err := freezeSnapshotCapture(incident, reportingResourceID(execution.JobID(), "snapshot").String(), at, boundary.Token, admission.ReferencePacks, source)
	if err != nil {
		return err
	}
	payload := snapshotCreateJobPayload{JobAdmittedAt: admission.AdmittedAt, RequestedBoundary: admission.RequestedBoundary, IncidentID: admission.IncidentID, ActorUserID: admission.ActorUserID, ClientTxnID: admission.ClientTxnID, SnapshotAt: at, SourceChangeSetHighWatermark: boundary.Token, SourceBoundaryJSON: boundary.CanonicalJSON, SourceCapture: encoded, SourceCaptureSHA256: digest}
	raw, err := canonicalJSON(payload)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	canonicalBytes = len(raw)
	stage = "capture_persistence"
	commit, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollbackSnapshotTx(commit)
	tag, err := commit.Exec(ctx, `UPDATE reporting_job_payloads SET request_json=$2 WHERE job_id=$1 AND request_json IS NULL`, execution.JobID(), raw)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errors.New("reporting: snapshot capture already retained")
	}

	// The current execution is fenced only at publication. Cancellation can
	// always obtain the Jobs lock while the large model is being prepared.
	if err := s.jobTransactions.ValidateExecutionTx(ctx, commit, execution); err != nil {
		return err
	}
	return commit.Commit(ctx)
}

func rollbackSnapshotTx(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}
