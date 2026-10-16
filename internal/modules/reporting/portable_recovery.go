package reporting

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/artifactport"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ValidateRetainedSnapshotModels admits native snapshots before readiness.
// Keep this separate from incident portability: deployment history has no
// portable-catalog cardinality bound and needs no exported reference catalog.
func ValidateRetainedSnapshotModels(ctx context.Context, pool *pgxpool.Pool, references reference_data.RetainedBindingValidator) error {
	if pool == nil || references == nil {
		return errors.New("reporting: incomplete snapshot recovery dependencies")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Pending, interrupted and captured requests are authoritative recovery
	// state too. Validate them before a restored worker can be dispatched.
	var lastJob *uuid.UUID
	for {
		var id uuid.UUID
		var captured bool
		err := tx.QueryRow(ctx, `SELECT job_id,request_json IS NOT NULL FROM reporting_job_payloads WHERE job_kind='snapshot_create' AND ($1::uuid IS NULL OR job_id>$1) ORDER BY job_id LIMIT 1`, lastJob).Scan(&id, &captured)
		if errors.Is(err, pgx.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		admission, err := readSnapshotAdmissionTx(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := references.ValidateBindingTx(ctx, tx, admission.ReferencePacks); err != nil {
			return err
		}
		if captured {
			if _, err := getSnapshotPayloadTx(ctx, tx, id); err != nil {
				return err
			}
		}
		retained, err := jobs.ReadRetainedExtensionJob(ctx, tx, id)
		if err != nil {
			return err
		}
		if retained.Resource.Status == jobs.StatusSucceeded && !captured {
			return errors.New("reporting: successful snapshot Job has no capture")
		}
		lastJob = &id
	}
	var last *uuid.UUID
	for {
		var id uuid.UUID
		err := tx.QueryRow(ctx, `SELECT snapshot_id FROM reporting_snapshots WHERE ($1::uuid IS NULL OR snapshot_id>$1) ORDER BY snapshot_id LIMIT 1`, last).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		record, err := getSnapshotRecordTx(ctx, tx, id)
		if err != nil {
			return err
		}
		model, err := decodeSnapshotRecordCapture(record)
		if err != nil {
			return err
		}
		payload, err := getSnapshotPayloadTx(ctx, tx, record.CreateJobID)
		if err != nil {
			return err
		}
		if err := validateSnapshotAdmissionBinding(payload, record.CreateJobID, record); err != nil {
			return err
		}
		if err := references.ValidateBindingTx(ctx, tx, model.ReferencePacks); err != nil {
			return err
		}
		last = &id
	}
	return tx.Commit(ctx)
}

// ValidateRetainedArtifacts checks restored source evidence before readiness.
// It never requires source freshness at restore time, renders content, adopts
// source trust or resolves the destination's active selection.
func ValidateRetainedArtifacts(ctx context.Context, pool *pgxpool.Pool, references interface {
	ExportTx(context.Context, pgx.Tx, uuid.UUID, []reference_data.SetBinding) ([]byte, error)
}) error {
	if pool == nil || references == nil {
		return errors.New("reporting: incomplete artifact recovery dependencies")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	last := uuid.Nil
	for {
		var incident uuid.UUID
		err := tx.QueryRow(ctx, `SELECT incident_id FROM reporting_imported_artifact_files WHERE incident_id>$1 GROUP BY incident_id ORDER BY incident_id LIMIT 1`, last).Scan(&incident)
		if errors.Is(err, pgx.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		last = incident
		var operations int
		if err := tx.QueryRow(ctx, `SELECT count(DISTINCT operation_id) FROM reporting_imported_artifact_files WHERE incident_id=$1`, incident).Scan(&operations); err != nil {
			return err
		}
		if operations != 1 {
			return artifactport.ErrInvalid
		}
		files, err := loadImportedArtifactFiles(ctx, tx, incident)
		if err != nil {
			return err
		}
		refs, err := references.ExportTx(ctx, tx, incident, nil)
		if err != nil {
			return err
		}
		_, err = (PortableReferenceSource{}).PrepareArtifactImport(ctx, artifactport.ImportRequest{IncidentID: incident, OperationID: incident, References: refs, Bundle: artifactFiles(files)})
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
