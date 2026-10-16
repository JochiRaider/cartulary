package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"
)

const renderPreparationSchema = "cartulary.reporting_render_preparation.v1"

// Preparation retains the identities of derived inputs, not another copy of
// the expanded case model. A restart deterministically derives and compares
// this seal before rendering; publication checks its independent stored digest.
type renderPreparation struct {
	SchemaID            string `json:"schema_id"`
	AdmissionSHA256     string `json:"admission_sha256"`
	SourceCaptureSHA256 string `json:"source_capture_sha256"`
	SnapshotModelID     string `json:"snapshot_model_id"`
	ExportModelID       string `json:"export_model_id"`
	ExportModelSHA256   string `json:"export_model_sha256"`
	SourceFactsSHA256   string `json:"source_facts_sha256"`
	ProfileSHA256       string `json:"redaction_profile_sha256"`
}

func materializeRenderAdmission(ctx context.Context, a renderAdmission, admissionSHA string, record SnapshotRecord) (renderInput, renderPreparation, error) {
	var empty renderInput
	var noPreparation renderPreparation
	if err := ctx.Err(); err != nil {
		return empty, noPreparation, err
	}
	raw, err := marshalRenderAdmission(a)
	if err != nil || reportingObjectDigest(renderAdmissionSchema, raw) != admissionSHA {
		return empty, noPreparation, errors.New("reporting: altered render recipe")
	}
	capture, err := decodeSnapshotRecordCapture(record)
	if err != nil {
		return empty, noPreparation, err
	}
	if record.SourceCaptureSHA256 != a.SourceCaptureSHA256 || capture.SnapshotID != a.SnapshotID || capture.IncidentID != a.IncidentID ||
		!capture.SnapshotAt.Equal(a.SnapshotAt) || capture.SourceChangeSetHighWatermark != a.SourceChangeSetHighWatermark ||
		capture.DerivationVersion != a.DerivationVersion || !sameReportingValue(capture.ReferencePacks, a.ReferencePacks) {
		return empty, noPreparation, errors.New("reporting: retained capture differs from render admission")
	}
	snapshot, err := materializeSnapshotCapture(ctx, capture)
	if err != nil {
		return empty, noPreparation, err
	}
	if err := ctx.Err(); err != nil {
		return empty, noPreparation, err
	}
	if a.Identity.ReleaseID != nil {
		request, apiErr := DecodeCreateReleaseRequest(bytes.NewReader(a.NormalizedRequest))
		if apiErr != nil {
			return empty, noPreparation, errors.New("reporting: invalid retained release request")
		}
		if err := validateRenderRecipientPartitions(request.ReleaseScope, request.RecipientPartitionRefs, snapshot.SnapshotContent, a.Profile); err != nil {
			return empty, noPreparation, err
		}
	}
	model, digest, err := bindRenderModel(snapshot, a.Identity, a.ReleaseScope, a.RecipientPartitionRefs)
	if err != nil {
		return empty, noPreparation, err
	}
	// Derivation may append exact path rules. Do not alias its authored policy.
	profileBytes, err := json.Marshal(a.Profile)
	if err != nil {
		return empty, noPreparation, err
	}
	var authored RedactionProfile
	if err := json.Unmarshal(profileBytes, &authored); err != nil {
		return empty, noPreparation, err
	}
	profile, profileSHA, err := materializeRedactionProfile(authored, snapshot.SnapshotContent)
	if err != nil {
		return empty, noPreparation, err
	}
	if a.ExpectedProfileSHA256 != nil && *a.ExpectedProfileSHA256 != profileSHA {
		return empty, noPreparation, renderFailureAt(errors.New("reporting: prepared profile differs from preview request"), "redaction", "redaction_manifest_invalid", "redaction_profile_invalid")
	}
	sourceSHA, err := sourceFactsDigest(snapshot.SourceFacts)
	if err != nil {
		return empty, noPreparation, err
	}
	at, err := time.Parse(reportingTimestampLayout, a.Identity.RenderAdmittedAt)
	if err != nil {
		return empty, noPreparation, err
	}
	composition := bytes.Clone(a.CompositionJSON)
	if bytes.Equal(composition, []byte("null")) {
		composition = nil
	}
	input := renderInput{
		SourceFacts: snapshot.SourceFacts, SourceFactsSHA256: sourceSHA,
		ActorUserID: a.ActorUserID, IncidentID: a.IncidentID, SnapshotID: a.SnapshotID, SnapshotAt: a.SnapshotAt,
		SourceChangeSetHighWatermark: a.SourceChangeSetHighWatermark, DerivationVersion: a.DerivationVersion,
		ExportModel: model, ExportModelSHA256: digest,
		TemplateID: a.TemplateContract.TemplateID, TemplateVersion: a.TemplateContract.TemplateVersion,
		TemplateContract: a.TemplateContract, TemplateFiles: a.TemplateFiles,
		RedactionProfileID: profile.ProfileID, RedactionProfileVersion: profile.Version, RedactionProfile: profile, RedactionProfileSHA256: profileSHA,
		OutputKind: a.OutputKind, OutputOptions: bytes.Clone(a.OutputOptions), ReleaseScope: a.ReleaseScope,
		RecipientPartitionRefs: append([]string{}, a.RecipientPartitionRefs...), GraphProjectionRefs: bytes.Clone(a.GraphProjectionRefs),
		CompositionJSON: composition, CompositionID: cloneStringPtr(a.CompositionID), CompositionVersion: cloneStringPtr(a.CompositionVersion), CompositionSHA256: cloneStringPtr(a.CompositionSHA256),
		RenderAdmittedAt: at, GraphSources: a.GraphSources,
	}
	if err := ctx.Err(); err != nil {
		return empty, noPreparation, err
	}
	if err := validateRenderInputIdentity(input); err != nil {
		return empty, noPreparation, err
	}
	if err := validateFrozenRenderProfile(input); err != nil {
		return empty, noPreparation, err
	}
	prepared := renderPreparation{SchemaID: renderPreparationSchema, AdmissionSHA256: admissionSHA, SourceCaptureSHA256: a.SourceCaptureSHA256,
		SnapshotModelID: model.SnapshotModelID, ExportModelID: model.ExportModelID, ExportModelSHA256: digest, SourceFactsSHA256: sourceSHA,
		ProfileSHA256: profileSHA}
	input.preparation = &prepared
	return input, prepared, nil
}

func decodeRenderPreparation(raw []byte, digest string, a renderAdmission, admissionSHA string) (renderPreparation, error) {
	var prepared renderPreparation
	invalid := errors.New("reporting: invalid retained render preparation")
	canonical, err := canonicalReportingBytes(raw)
	if err != nil || reportingObjectDigest(renderPreparationSchema, canonical) != digest {
		return prepared, invalid
	}
	if json.Unmarshal(canonical, &prepared) != nil {
		return prepared, invalid
	}
	encoded, err := canonicalJSON(prepared)
	if err != nil || !bytes.Equal(encoded, canonical) || prepared.SchemaID != renderPreparationSchema ||
		prepared.AdmissionSHA256 != admissionSHA || prepared.SourceCaptureSHA256 != a.SourceCaptureSHA256 ||
		!sha256HexPattern.MatchString(prepared.ExportModelSHA256) || !sha256HexPattern.MatchString(prepared.SourceFactsSHA256) {
		return renderPreparation{}, invalid
	}
	modelID, err := snapshotModelID(a.SnapshotID, a.DerivationVersion, a.SnapshotAt)
	if err != nil || prepared.SnapshotModelID != modelID {
		return renderPreparation{}, invalid
	}
	exportID, err := a.Identity.generatedID("export_model", "expm_", map[string]any{"snapshot_id": a.SnapshotID, "derivation_version": a.DerivationVersion, "render_admitted_at": a.Identity.RenderAdmittedAt})
	if err != nil || prepared.ExportModelID != exportID {
		return renderPreparation{}, invalid
	}
	if !sha256HexPattern.MatchString(prepared.ProfileSHA256) || a.ExpectedProfileSHA256 != nil && *a.ExpectedProfileSHA256 != prepared.ProfileSHA256 {
		return renderPreparation{}, invalid
	}

	return prepared, nil
}
