package reporting

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ValidateRetainedReleases admits native release history without rendering,
// changing source, or imposing incident-portability cardinality limits.
func ValidateRetainedReleases(ctx context.Context, pool *pgxpool.Pool, references reference_data.RetainedBindingValidator) error {
	if pool == nil || references == nil {
		return errors.New("reporting: incomplete release recovery dependencies")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var last *uuid.UUID
	var materialized int64
	for {
		var id uuid.UUID
		err := tx.QueryRow(ctx, `SELECT job_id FROM reporting_job_payloads WHERE job_kind='release_create' AND ($1::uuid IS NULL OR job_id>$1) ORDER BY job_id LIMIT 1`, last).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		a, digest, err := readRenderAdmissionTx(ctx, tx, id)
		if err != nil {
			return err
		}
		seal, _, err := readRenderPreparationTx(ctx, tx, id, a, digest)
		if err != nil {
			return err
		}
		retained, err := jobs.ReadRetainedExtensionJob(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := references.ValidateBindingTx(ctx, tx, a.ReferencePacks); err != nil {
			return err
		}
		metadata, err := renderAdmissionMetadata(a, seal)
		if err != nil {
			return err
		}
		if seal == nil {
			if _, err := getReleaseRecordByCreateJobIDTx(ctx, tx, id); !errors.Is(err, ErrNotFound) {
				return errors.New("reporting: unprepared render has release output")
			}
			if retained.Resource.Status == jobs.StatusSucceeded {
				return errors.New("reporting: successful render lacks preparation")
			}
			if err := validateRetainedPreviewFailureTx(ctx, tx, id, metadata, retained); err != nil {
				return err
			}
			last = &id
			continue
		}
		payload, err := getReleasePayloadTx(ctx, tx, id)
		if err != nil {
			return err
		}
		record, err := getReleaseRecordByCreateJobIDTx(ctx, tx, id)
		if errors.Is(err, ErrNotFound) {
			switch retained.Resource.Status {
			case jobs.StatusQueued, jobs.StatusRunning, jobs.StatusCancelRequested:
			default:
				return errors.New("reporting: terminal admitted render has no release outcome")
			}
			if err := validateRetainedRenderFailureTx(ctx, tx, id, payload.renderInput, nil); err != nil {
				return err
			}
			last = &id
			continue
		}
		if err != nil {
			return err
		}
		if err := validateReleaseAdmissionBinding(payload, id, record); err != nil {
			return err
		}
		if err := validateReleaseJobOutcome(record, retained); err != nil {
			return err
		}
		if _, err := validateStoredReleaseOutputTx(ctx, tx, payload, record); err != nil {
			return err
		}
		materialized++
		last = &id
	}
	var stored int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM reporting_releases`).Scan(&stored); err != nil {
		return err
	}
	if stored != materialized {
		return errors.New("reporting: orphaned release outcome")
	}
	return tx.Commit(ctx)
}

func validateStoredReleaseOutputTx(ctx context.Context, tx pgx.Tx, payload releaseCreateJobPayload, record ReleaseRecord) (RenderedRelease, error) {
	if err := validateRetainedRenderFailureTx(ctx, tx, record.CreateJobID, payload.renderInput, record.RenderFailedReasonCode); err != nil {
		return RenderedRelease{}, err
	}
	var composition *reportcomposition.ReleaseTuple
	if record.CompositionID != nil {
		composition = &reportcomposition.ReleaseTuple{CompositionID: *record.CompositionID, CompositionVersion: derefString(record.CompositionVersion), CompositionSHA256: derefString(record.CompositionSHA256)}
	}
	if err := reportcomposition.ValidateRetainedReleaseBindingTx(ctx, tx, record.ReleaseID, record.ReleaseScope, composition); err != nil {
		return RenderedRelease{}, err
	}
	if record.ReleaseState == ReleaseStateRenderFailed {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reporting_render_bundles WHERE release_id=$1)`, record.ReleaseID).Scan(&exists); err != nil {
			return RenderedRelease{}, err
		}
		if exists || record.RenderFailedReasonCode == nil || *record.RenderFailedReasonCode == "" ||
			record.OutputSHA256 != nil || record.OutputMediaType != nil || record.RedactionManifestSHA256 != nil || len(record.RedactionManifestJSON) != 0 {
			return RenderedRelease{}, errors.New("reporting: failed release has output or lacks failure identity")
		}
		if payload.RedactionProfileSHA256 != record.RedactionProfileSHA256 {
			return RenderedRelease{}, errors.New("reporting: failed release profile differs from admission")
		}
		return RenderedRelease{}, nil
	}
	rendered, err := loadRetainedReleaseArtifactsTx(ctx, tx, payload, record)
	if err != nil {
		return RenderedRelease{}, err
	}
	if err := validateReleaseCandidateBinding(payload, record.CreateJobID, record.ReleaseID, rendered); err != nil {
		return RenderedRelease{}, err
	}
	if err := validateReleaseOutputBinding(record, rendered); err != nil {
		return RenderedRelease{}, err
	}
	return rendered, nil
}

func loadRetainedReleaseArtifactsTx(ctx context.Context, tx pgx.Tx, payload releaseCreateJobPayload, record ReleaseRecord) (RenderedRelease, error) {
	var result RenderedRelease
	b := &result.RenderBundle
	err := tx.QueryRow(ctx, `SELECT bundle_manifest_sha256,bundle_manifest_json,primary_bundle_path,primary_media_type FROM reporting_render_bundles WHERE release_id=$1`, record.ReleaseID).Scan(&b.ManifestSHA256, &b.ManifestJSON, &b.PrimaryPath, &b.PrimaryMedia)
	if err != nil {
		return result, err
	}
	b.Manifest, b.ManifestJSON, err = decodePortableObject[RenderBundleManifest](b.ManifestJSON, b.ManifestSHA256)
	if err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, `SELECT bundle_path,role,media_type,file_sha256,size_bytes,storage_kind,inline_bytes FROM reporting_render_bundle_files WHERE release_id=$1 ORDER BY bundle_path`, record.ReleaseID)
	if err != nil {
		return result, err
	}
	if err := loadRetainedRenderFiles(rows, &result); err != nil {
		return result, err
	}
	err = tx.QueryRow(ctx, `SELECT reveal_map_sha256,canonical_bytes FROM reporting_release_reveal_maps WHERE release_id=$1`, record.ReleaseID).Scan(&result.Redaction.RevealMapSHA256, &result.Redaction.RevealMapJSON)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}
	result.Profile = payload.RedactionProfile
	result.ProfileSHA256 = record.RedactionProfileSHA256
	result.OutputSHA256 = derefString(record.OutputSHA256)
	result.OutputMediaType = derefString(record.OutputMediaType)
	result.RedactionManifestSHA256 = derefString(record.RedactionManifestSHA256)
	result.RedactionManifestJSON, err = canonicalReportingBytes(record.RedactionManifestJSON)
	return result, err
}
