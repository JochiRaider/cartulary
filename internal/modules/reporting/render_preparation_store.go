package reporting

import (
	"context"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) prepareRenderJob(ctx context.Context, execution jobs.Execution) (renderAdmission, renderInput, renderPreparation, error) {
	started := time.Now()
	var a renderAdmission
	var input renderInput
	var prepared renderPreparation
	tx, err := s.beginExecutionRead(ctx, execution)
	if err != nil {
		return a, input, prepared, err
	}
	defer rollbackSnapshotTx(tx)
	a, digest, err := readRenderAdmissionTx(ctx, tx, execution.JobID())
	if err != nil {
		return a, input, prepared, err
	}
	limits, err := resolveRenderLimits(a.TemplateContract.DeclaredLimits)
	if err != nil {
		return a, input, prepared, err
	}
	ctx, cancel := withRenderDeadlineFrom(ctx, started, limits.timeout("total_attempt"), "release_state", "timeout.total_attempt")
	defer cancel()
	if s.referencePacks == nil {
		return a, input, prepared, errors.New("reporting: Reference Pack retention is required")
	}
	if err := s.referencePacks.ValidateBindingTx(ctx, tx, a.ReferencePacks); err != nil {
		var missing *reference_data.ConsumerError
		if errors.As(err, &missing) {
			err = renderFailureAt(ErrRequiredReferencePackUnavailable, "export_model_materialization", "export_model_invalid", "required_reference_pack_unavailable")
		}
		return a, input, prepared, err
	}
	refs, err := decodeSourceProjectionRefs(a.GraphProjectionRefs)
	if err != nil {
		return a, input, prepared, err
	}
	graphs, err := s.graphSources.ReadAndRenewTx(ctx, tx, execution.JobID(), refs, time.Now().UTC())
	if err != nil {
		return a, input, prepared, err
	}
	frozen, err := freezeGraphSources(graphs)
	if err != nil || !sameFrozenGraphSources(frozen, a.GraphSources) {
		return a, input, prepared, errors.New("reporting: graph lease differs from render admission")
	}
	if err := tx.Commit(ctx); err != nil {
		return a, input, prepared, err
	}

	// No Jobs lock is retained while reading incident-sized bytes or deriving
	// presentation, profile rules and model hashes. The worker supervises this
	// phase and drains it before it publishes a cancellation or failure.
	source, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return a, input, prepared, err
	}
	defer rollbackSnapshotTx(source)
	record, err := getSnapshotRecordTx(ctx, source, uuid.MustParse(a.SnapshotID))
	if err != nil {
		return a, input, prepared, err
	}
	capture, err := getSnapshotPayloadTx(ctx, source, record.CreateJobID)
	if err != nil {
		return a, input, prepared, err
	}
	if err := validateSnapshotAdmissionBinding(capture, record.CreateJobID, record); err != nil {
		return a, input, prepared, err
	}
	if err := source.Commit(ctx); err != nil {
		return a, input, prepared, err
	}
	input, prepared, err = materializeRenderAdmission(ctx, a, digest, record)
	if err != nil {
		return a, input, prepared, err
	}
	raw, err := canonicalJSON(prepared)
	if err != nil {
		return a, input, prepared, err
	}
	preparedSHA := reportingObjectDigest(renderPreparationSchema, raw)
	admissionBytes, err := marshalRenderAdmission(a)
	if err != nil {
		return a, input, prepared, err
	}
	commit, err := s.pool.Begin(ctx)
	if err != nil {
		return a, input, prepared, err
	}
	defer rollbackSnapshotTx(commit)
	// Match the finalizer's lock order. Only the small preparation seal is
	// written under the execution fence; never the expanded case-content model.
	if err := s.jobTransactions.ValidateExecutionTx(ctx, commit, execution); err != nil {
		return a, input, prepared, err
	}
	tag, err := commit.Exec(ctx, `UPDATE reporting_job_payloads SET render_preparation_json=$2,render_preparation_sha256=$3 WHERE job_id=$1 AND render_admission_sha256=$4 AND request_json=$5 AND (render_preparation_json IS NULL OR (render_preparation_json=$2 AND render_preparation_sha256=$3))`, execution.JobID(), raw, preparedSHA, digest, admissionBytes)
	if err != nil {
		return a, input, prepared, err
	}
	if tag.RowsAffected() != 1 {
		return a, input, prepared, errors.New("reporting: preparation changed retained inputs")
	}
	if err := commit.Commit(ctx); err != nil {
		return a, input, prepared, err
	}
	return a, input, prepared, nil
}

func readRenderPreparationTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, a renderAdmission, admissionSHA string) (*renderPreparation, string, error) {
	var raw []byte
	var digest *string
	if err := tx.QueryRow(ctx, `SELECT render_preparation_json,render_preparation_sha256 FROM reporting_job_payloads WHERE job_id=$1`, id).Scan(&raw, &digest); err != nil {
		return nil, "", err
	}
	if raw == nil && digest == nil {
		return nil, "", nil
	}
	if digest == nil || raw == nil {
		return nil, "", errors.New("reporting: incomplete render preparation")
	}
	prepared, err := decodeRenderPreparation(raw, *digest, a, admissionSHA)
	if err != nil {
		return nil, "", err
	}
	return &prepared, *digest, nil
}
