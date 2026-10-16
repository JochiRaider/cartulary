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

// ValidateRetainedPreviews validates immutable admission and output in one
// consistent read. It never renders, selects a current draft/profile or leases
// graph results. Pending, failed and canceled attempts must have no output.
func ValidateRetainedPreviews(ctx context.Context, pool *pgxpool.Pool, references reference_data.RetainedBindingValidator) error {
	if pool == nil || references == nil {
		return errors.New("reporting: incomplete preview recovery dependencies")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var last *uuid.UUID
	var admissions, outputs int64
	for {
		source, present, err := reportcomposition.NextRetainedPreviewTx(ctx, tx, last)
		if err != nil {
			return err
		}
		if !present {
			break
		}
		a, digest, err := readRenderAdmissionTx(ctx, tx, source.RenderAttemptID)
		if err != nil {
			return err
		}
		seal, _, err := readRenderPreparationTx(ctx, tx, source.RenderAttemptID, a, digest)
		if err != nil {
			return err
		}
		if err := references.ValidateBindingTx(ctx, tx, a.ReferencePacks); err != nil {
			return err
		}
		retained, err := jobs.ReadRetainedExtensionJob(ctx, tx, source.RenderAttemptID)
		if err != nil {
			return err
		}
		metadata, err := renderAdmissionMetadata(a, seal)
		if err != nil {
			return err
		}
		payload, err := previewPayloadFromPrepared(a, metadata)
		if err != nil {
			return err
		}
		if seal != nil {
			payload, err = getPreviewPayloadTx(ctx, tx, source.RenderAttemptID)
			if err != nil {
				return err
			}
		}
		if seal == nil && retained.Resource.Status == jobs.StatusSucceeded {
			return errors.New("reporting: successful preview lacks preparation")
		}
		if err := validateRetainedPreviewFailureTx(ctx, tx, source.RenderAttemptID, payload.renderInput, retained); err != nil {
			return err
		}
		_, err = loadRetainedPreviewOutputTx(ctx, tx, payload, source.RenderAttemptID)
		if retained.Resource.Status == jobs.StatusSucceeded {
			if err != nil {
				return err
			}
			if err := validatePreviewJobOutcome(retained); err != nil {
				return err
			}
			outputs++
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return errors.New("reporting: unfinished preview has output or invalid retained output")
		}
		admissions++
		id := source.RenderAttemptID
		last = &id
	}
	// Every Composition admission already matched a unique private Job input.
	// Extra owner rows indicate orphaned inputs/outputs, not additional coverage.
	var storedInputs, storedOutputs int64
	if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM reporting_job_payloads WHERE job_kind='composition_preview'),(SELECT count(*) FROM reporting_composition_preview_outputs)`).Scan(&storedInputs, &storedOutputs); err != nil {
		return err
	}
	if storedInputs != admissions || storedOutputs != outputs {
		return errors.New("reporting: orphaned preview input or output")
	}
	return tx.Commit(ctx)
}

func validatePreviewJobOutcome(retained jobs.RetainedExtensionJob) error {
	r := retained.Resource
	if r.Status != jobs.StatusSucceeded || retained.ExpiredAt == nil && (r.ResultSummary == nil || r.ResultSummary.Code != reportingJobPreviewCreatedCode || len(r.ResultSummary.ResourceRefs) != 0) {
		return errors.New("reporting: preview output differs from Jobs outcome")
	}
	return nil
}

func loadRetainedPreviewOutputTx(ctx context.Context, tx pgx.Tx, payload compositionPreviewJobPayload, jobID uuid.UUID) (RenderedRelease, error) {
	result, err := loadRetainedPreviewArtifactsTx(ctx, tx, payload, jobID)
	if err != nil {
		return result, err
	}
	if err := validateRenderCandidateBindingContext(ctx, payload.renderInput, result); err != nil {
		return result, err
	}
	return result, nil
}

func loadRetainedPreviewArtifactsTx(ctx context.Context, tx pgx.Tx, payload compositionPreviewJobPayload, jobID uuid.UUID) (RenderedRelease, error) {
	var result RenderedRelease
	b := &result.RenderBundle
	var id uuid.UUID
	var scope, kind, profileID, profileVersion string
	err := tx.QueryRow(ctx, `SELECT preview_attempt_id,release_scope,output_kind,output_media_type,output_sha256,redaction_profile_id,redaction_profile_version,redaction_profile_sha256,redaction_manifest_sha256,redaction_manifest_json,bundle_manifest_sha256,bundle_manifest_json,primary_bundle_path,primary_media_type FROM reporting_composition_preview_outputs WHERE render_attempt_id=$1`, jobID).Scan(
		&id, &scope, &kind, &result.OutputMediaType, &result.OutputSHA256, &profileID, &profileVersion, &result.ProfileSHA256, &result.RedactionManifestSHA256, &result.RedactionManifestJSON, &b.ManifestSHA256, &b.ManifestJSON, &b.PrimaryPath, &b.PrimaryMedia)
	if err != nil {
		return result, err
	}
	if id != payload.PreviewAttemptID || scope != payload.ReleaseScope || kind != payload.OutputKind || profileID != payload.RedactionProfileID || profileVersion != payload.RedactionProfileVersion {
		return result, errors.New("reporting: retained preview row differs from admission")
	}
	result.Profile = payload.RedactionProfile
	b.Manifest, b.ManifestJSON, err = decodePortableObject[RenderBundleManifest](b.ManifestJSON, b.ManifestSHA256)
	if err != nil {
		return result, err
	}
	result.RedactionManifestJSON, err = canonicalReportingBytes(result.RedactionManifestJSON)
	if err != nil {
		return result, err
	}
	rows, err := tx.Query(ctx, `SELECT bundle_path,role,media_type,file_sha256,size_bytes,storage_kind,inline_bytes FROM reporting_composition_preview_output_files WHERE preview_attempt_id=$1 ORDER BY bundle_path`, id)
	if err != nil {
		return result, err
	}
	if err := loadRetainedRenderFiles(rows, &result); err != nil {
		return result, err
	}
	err = tx.QueryRow(ctx, `SELECT reveal_map_sha256,canonical_bytes FROM reporting_preview_reveal_maps WHERE preview_attempt_id=$1`, id).Scan(&result.Redaction.RevealMapSHA256, &result.Redaction.RevealMapJSON)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}
	return result, nil
}
