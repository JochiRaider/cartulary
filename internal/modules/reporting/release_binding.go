package reporting

import (
	"bytes"
	"context"
	"errors"
	"slices"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

// Release lifecycle timestamps and approval state may evolve. Admission facts
// and completed output do not: both completion replay and recovery use these
// comparisons against the independently retained owner input.
func validateReleaseAdmissionBinding(payload releaseCreateJobPayload, jobID uuid.UUID, record ReleaseRecord) error {
	if record.CreateJobID != jobID || record.ReleaseID != reportingResourceID(jobID, "release") ||
		payload.IncidentID != record.IncidentID.String() || payload.SnapshotID != record.SnapshotID.String() ||
		payload.ActorUserID != record.CreatedByUserID.String() || payload.ClientTxnID != record.ClientTxnID ||
		payload.ReleaseScope != record.ReleaseScope || !payload.SnapshotAt.Equal(record.SnapshotAt) ||
		payload.SourceChangeSetHighWatermark != record.SourceChangeSetHighWatermark || payload.DerivationVersion != record.DerivationVersion ||
		payload.ExportModelSHA256 != record.ExportModelSHA256 || payload.TemplateID != record.TemplateID || payload.TemplateVersion != record.TemplateVersion ||
		payload.RedactionProfileID != record.RedactionProfileID || payload.RedactionProfileVersion != record.RedactionProfileVersion || payload.RedactionProfileSHA256 != record.RedactionProfileSHA256 ||
		payload.OutputKind != record.OutputKind || !equalFrozenJSON(payload.OutputOptions, record.OutputOptions) ||
		!equalFrozenJSON(payload.GraphProjectionRefs, record.GraphProjectionRefs) ||
		!equalOptionalString(payload.CompositionID, optionalUUIDStringPointer(record.CompositionID)) ||
		!equalOptionalString(payload.CompositionVersion, record.CompositionVersion) || !equalOptionalString(payload.CompositionSHA256, record.CompositionSHA256) ||
		!payload.RenderAdmittedAt.Equal(record.RenderAdmittedAt) || !slices.Equal(payload.RecipientPartitionRefs, record.RecipientPartitionRefs) {
		return errors.New("reporting: retained release differs from admission")
	}
	return nil
}

func validateReleaseCandidateBinding(payload releaseCreateJobPayload, jobID, releaseID uuid.UUID, rendered RenderedRelease) error {
	if releaseID != reportingResourceID(jobID, "release") {
		return errors.New("reporting: release identity differs from admission")
	}
	return validateRenderCandidateBinding(payload.renderInput, rendered)
}

func validateRenderCandidateBinding(payload renderInput, rendered RenderedRelease) error {
	return validateRenderCandidateBindingContext(context.Background(), payload, rendered)
}

func validateRenderCandidateBindingContext(ctx context.Context, payload renderInput, rendered RenderedRelease) error {
	invalid := errors.New("reporting: rendered output differs from admission")
	if err := validateShareableBundle(rendered.RenderBundle); err != nil {
		return err
	}
	if err := validateRevealMap(rendered); err != nil {
		return err
	}
	m := rendered.RenderBundle.Manifest
	if m.RendererID != renderBundleRendererID || m.RendererVersion != renderBundleRendererVersion ||
		!sameReportingValue(m.RenderIdentity, payload.ExportModel.RenderIdentity) ||
		m.SnapshotID != payload.SnapshotID || m.ExportModelID != payload.ExportModel.ExportModelID || m.AdmittedExportModelSHA256 != payload.ExportModelSHA256 ||
		!sameReportingValue(m.ReferencePacks, payload.ExportModel.ReferencePacks) ||
		m.TemplateID != payload.TemplateID || m.TemplateVersion != payload.TemplateVersion || m.OutputKind != payload.OutputKind || m.ReleaseScope != payload.ReleaseScope ||
		rendered.OutputSHA256 != rendered.RenderBundle.ManifestSHA256 || rendered.OutputMediaType != rendered.RenderBundle.PrimaryMedia ||
		rendered.RedactionManifestSHA256 != m.RedactionManifestSHA256 {
		return invalid
	}
	template, err := canonicalJSON(templateManifestView(payload.TemplateContract))
	if err != nil || reportingObjectDigest("cartulary.reporting_template_pack_manifest.v3", template) != m.TemplateManifestSHA256 {
		return invalid
	}
	profile, profileSHA := payload.RedactionProfile, payload.RedactionProfileSHA256
	if validateFrozenRenderProfile(payload) != nil || rendered.ProfileSHA256 != profileSHA || !sameReportingValue(profile, rendered.Profile) {
		return invalid
	}
	// Reuse the pure redaction owner against the independently frozen input.
	// Rehashing mutually consistent artifacts cannot replace admitted values or
	// selected rules. This does not rerun rendering or read live source state.
	expected, err := redactRenderInput(ctx, payload)
	if err != nil || m.ExportModelSHA256 != expected.ModelSHA256 || rendered.RedactionManifestSHA256 != expected.ManifestSHA256 || !bytes.Equal(rendered.RedactionManifestJSON, expected.ManifestJSON) ||
		rendered.Redaction.TokenManifestSHA256 != expected.TokenManifestSHA256 || !bytes.Equal(rendered.Redaction.TokenManifestJSON, expected.TokenManifestJSON) ||
		rendered.Redaction.RevealMapSHA256 != expected.RevealMapSHA256 || !bytes.Equal(rendered.Redaction.RevealMapJSON, expected.RevealMapJSON) {
		return invalid
	}
	if payload.OutputKind == OutputKindSlidev {
		if err := validateSlidevCandidateBinding(payload, expected.Model, rendered.RenderBundle); err != nil {
			return err
		}
	} else if payload.OutputKind == OutputKindMermaid {
		if err := validateMermaidCandidateBinding(payload, expected.Model, rendered.RenderBundle); err != nil {
			return err
		}
	}
	// The owner row, root manifest and actual shareable members must identify
	// the same artifacts. A valid standalone manifest is insufficient.
	counts := map[string]int{}
	for _, file := range rendered.RenderBundle.Files {
		switch file.Role {
		case renderBundleRoleValidationSummary:
			counts[file.Role]++
			if file.Path != "validation/summary.json" || validatePassedRenderSummary(file.Bytes, m) != nil {
				return invalid
			}
		case renderBundleRoleRedactionManifest:
			counts[file.Role]++
			if !bytes.Equal(file.Bytes, expected.ManifestJSON) {
				return invalid
			}
		case renderBundleRoleRedactionProfileView:
			counts[file.Role]++
			if !bytes.Equal(file.Bytes, expected.ProfileViewJSON) {
				return invalid
			}
		case renderBundleRoleTokenManifest:
			if !bytes.Equal(file.Bytes, expected.TokenManifestJSON) {
				return invalid
			}
		}
	}
	if counts[renderBundleRoleValidationSummary] != 1 || counts[renderBundleRoleRedactionManifest] != 1 || counts[renderBundleRoleRedactionProfileView] != 1 {
		return invalid
	}
	return nil
}

func validateReleaseOutputBinding(record ReleaseRecord, rendered RenderedRelease) error {
	if record.ReleaseState == ReleaseStateRenderFailed || record.RenderFailedReasonCode != nil ||
		!equalOptionalString(record.OutputSHA256, &rendered.OutputSHA256) || !equalOptionalString(record.OutputMediaType, &rendered.OutputMediaType) ||
		!equalOptionalString(record.RedactionManifestSHA256, &rendered.RedactionManifestSHA256) ||
		!equalFrozenJSON(record.RedactionManifestJSON, rendered.RedactionManifestJSON) || record.RedactionProfileSHA256 != rendered.ProfileSHA256 {
		return errors.New("reporting: retained release output binding differs")
	}
	return nil
}

func validateReleaseJobOutcome(record ReleaseRecord, retained jobs.RetainedExtensionJob) error {
	invalid := errors.New("reporting: retained release differs from Jobs outcome")
	resource := retained.Resource
	if record.ReleaseState == ReleaseStateRenderFailed {
		if (resource.Status != jobs.StatusFailed && resource.Status != jobs.StatusCanceled) || record.RenderFailedReasonCode == nil || *record.RenderFailedReasonCode == "" ||
			record.OutputSHA256 != nil || record.OutputMediaType != nil || record.RedactionManifestSHA256 != nil || len(record.RedactionManifestJSON) != 0 {
			return invalid
		}
		if resource.Status == jobs.StatusCanceled {
			if *record.RenderFailedReasonCode != "render_canceled" && *record.RenderFailedReasonCode != "partial_output_cleanup_failed" {
				return invalid
			}
			if retained.ExpiredAt == nil && (resource.ResultSummary == nil || resource.ResultSummary.Code != "job_canceled") {
				return invalid
			}
			return nil
		}
		if retained.ExpiredAt == nil && (resource.ErrorSummary == nil || resource.ErrorSummary.Code != renderFailureJobCode(*record.RenderFailedReasonCode) ||
			resource.ErrorSummary.Details["release_id"] != record.ReleaseID.String() || resource.ErrorSummary.Details["reason_code"] != *record.RenderFailedReasonCode) {
			return invalid
		}
		return nil
	}
	switch record.ReleaseState {
	case ReleaseStatePendingApproval, ReleaseStateApproved, ReleaseStatePublished, ReleaseStateInvalidated:
	default:
		return invalid
	}
	if resource.Status != jobs.StatusSucceeded || record.RenderFailedReasonCode != nil {
		return invalid
	}
	if retained.ExpiredAt == nil {
		expected := []jobs.ResourceRef{{Kind: "release", ID: record.ReleaseID.String(), Route: "/api/v1/releases/" + record.ReleaseID.String()}}
		if resource.ResultSummary == nil || resource.ResultSummary.Code != reportingJobReleaseCreatedCode || !slices.Equal(resource.ResultSummary.ResourceRefs, expected) {
			return invalid
		}
	}
	return nil
}
