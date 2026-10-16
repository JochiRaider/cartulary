package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/google/uuid"
)

const renderAdmissionSchema = "cartulary.reporting_render_admission.v1"
const renderPreparationAlgorithm = "materialize_retained_reporting_render_v1"

// Admission retains a recipe and exact immutable dependencies. It does not
// retain or eagerly construct the expanded presentation model. Preparation is
// cancelable worker work, and all derived bytes remain bound to this recipe.
type renderAdmission struct {
	SchemaID                     string                    `json:"schema_id"`
	Algorithm                    string                    `json:"preparation_algorithm"`
	Identity                     RenderIdentity            `json:"identity"`
	ActorUserID                  string                    `json:"actor_user_id"`
	IncidentID                   string                    `json:"incident_id"`
	SnapshotID                   string                    `json:"snapshot_id"`
	SnapshotAt                   time.Time                 `json:"snapshot_at"`
	SourceChangeSetHighWatermark string                    `json:"source_change_set_high_watermark"`
	DerivationVersion            string                    `json:"derivation_version"`
	SourceCaptureSHA256          string                    `json:"source_capture_sha256"`
	ReferencePacks               reference_data.SetBinding `json:"reference_packs"`
	TemplateContract             TemplateContract          `json:"template_contract"`
	TemplateFiles                []TemplatePackFile        `json:"template_files"`
	TemplateSHA256               string                    `json:"template_sha256"`
	Profile                      RedactionProfile          `json:"authored_redaction_profile"`
	ProfileSHA256                string                    `json:"authored_redaction_profile_sha256"`
	ExpectedProfileSHA256        *string                   `json:"expected_materialized_profile_sha256"`
	ReleaseScope                 string                    `json:"release_scope"`
	RecipientPartitionRefs       []string                  `json:"recipient_partition_refs"`
	OutputKind                   string                    `json:"output_kind"`
	OutputOptions                json.RawMessage           `json:"output_options"`
	GraphProjectionRefs          json.RawMessage           `json:"graph_projection_refs"`
	GraphSources                 frozenGraphSources        `json:"graph_sources"`
	GraphSourcesSHA256           string                    `json:"graph_sources_sha256"`
	CompositionJSON              json.RawMessage           `json:"composition_json"`
	CompositionID                *string                   `json:"composition_id"`
	CompositionVersion           *string                   `json:"composition_version"`
	CompositionSHA256            *string                   `json:"composition_sha256"`
	ClientTxnID                  string                    `json:"client_txn_id"`
	NormalizedRequest            []byte                    `json:"normalized_request"`
}

func decodeRenderAdmission(raw []byte, digest string, jobID, incidentID, actorID uuid.UUID, admittedAt time.Time) (renderAdmission, error) {
	var admission renderAdmission
	invalid := errors.New("reporting: invalid retained render admission")
	canonical, err := canonicalReportingBytes(raw)
	if err != nil || reportingObjectDigest(renderAdmissionSchema, canonical) != digest {
		return admission, invalid
	}
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.UseNumber()
	if decoder.Decode(&admission) != nil {
		return admission, invalid
	}
	encoded, err := marshalRenderAdmission(admission)
	if err != nil || !bytes.Equal(encoded, canonical) {
		return renderAdmission{}, invalid
	}
	if admission.SchemaID != renderAdmissionSchema || admission.Algorithm != renderPreparationAlgorithm ||
		jobID == uuid.Nil || actorID == uuid.Nil || incidentID == uuid.Nil ||
		admission.ActorUserID != actorID.String() || admission.IncidentID != incidentID.String() ||
		!canonicalUUID(admission.SnapshotID) || !sha256HexPattern.MatchString(admission.SourceCaptureSHA256) ||
		admission.DerivationVersion != DerivationVersion || admission.SourceChangeSetHighWatermark == "" ||
		admission.Identity.validate() != nil || admission.RecipientPartitionRefs == nil || admission.TemplateFiles == nil || admission.NormalizedRequest == nil {
		return renderAdmission{}, invalid
	}
	stamp, err := normalizedReportingTimestamp(admittedAt)
	if err != nil || stamp != admission.Identity.RenderAdmittedAt {
		return renderAdmission{}, invalid
	}
	if _, err := normalizedReportingTimestamp(admission.SnapshotAt); err != nil || admission.SnapshotAt.After(admittedAt) {
		return renderAdmission{}, invalid
	}
	if admission.Identity.ReleaseID != nil && *admission.Identity.ReleaseID != reportingResourceID(jobID, "release").String() {
		return renderAdmission{}, invalid
	}
	if admission.Identity.PreviewAttemptID != nil && (admission.ReleaseScope != ReleaseScopeInternalDraft || len(admission.RecipientPartitionRefs) != 0 || admission.ExpectedProfileSHA256 == nil) {
		return renderAdmission{}, invalid
	}
	if admission.ExpectedProfileSHA256 != nil && !sha256HexPattern.MatchString(*admission.ExpectedProfileSHA256) {
		return renderAdmission{}, invalid
	}
	if digest, err := ValidateRedactionProfile(admission.Profile); err != nil || digest != admission.ProfileSHA256 ||
		!profileSupportsScope(RedactionProfileView{ProfileStatus: admission.Profile.ProfileStatus, AllowedReleaseScopes: admission.Profile.AllowedReleaseScopes}, admission.ReleaseScope) {
		return renderAdmission{}, invalid
	}
	if err := validateTemplateContract(admission.TemplateContract, admission.OutputKind, admission.ReleaseScope); err != nil {
		return renderAdmission{}, invalid
	}
	if err := validateTemplatePackFiles(admission.TemplateContract, admission.TemplateFiles); err != nil {
		return renderAdmission{}, invalid
	}
	if digest, err := templateManifestDigest(admission.TemplateContract); err != nil || digest != admission.TemplateSHA256 {
		return renderAdmission{}, invalid
	}
	refs, err := decodeSourceProjectionRefs(admission.GraphProjectionRefs)
	if err != nil || validateFrozenGraphSources(admission.GraphSources, refs) != nil {
		return renderAdmission{}, invalid
	}
	if digest, err := graphSourcesDigest(admission.GraphSources); err != nil || digest != admission.GraphSourcesSHA256 {
		return renderAdmission{}, invalid
	}
	if _, err := parseRenderOutputOptions(admission.OutputOptions, admission.OutputKind, admission.ReleaseScope); err != nil {
		return renderAdmission{}, invalid
	}
	if err := validateRenderAdmissionRequest(admission); err != nil {
		return renderAdmission{}, err
	}
	return admission, nil
}

// Private transport bytes (normalized requests and template files) use JSON's
// explicit base64 representation before Reporting canonicalization. They are
// not source scalar values in a case-content model.
func marshalRenderAdmission(admission renderAdmission) ([]byte, error) {
	raw, err := json.Marshal(admission)
	if err != nil {
		return nil, err
	}
	return canonicalReportingBytes(raw)
}

func validateRenderAdmissionRequest(a renderAdmission) error {
	invalid := errors.New("reporting: render admission differs from request")
	if a.Identity.ReleaseID != nil {
		request, apiErr := DecodeCreateReleaseRequest(bytes.NewReader(a.NormalizedRequest))
		if apiErr != nil || !bytes.Equal(request.Normalized, a.NormalizedRequest) ||
			request.SnapshotID.String() != a.SnapshotID || request.ClientTxnID != a.ClientTxnID ||
			request.TemplateID != a.TemplateContract.TemplateID || request.TemplateVersion != a.TemplateContract.TemplateVersion ||
			request.RedactionProfileID != a.Profile.ProfileID || request.RedactionProfileVersion != a.Profile.Version ||
			request.ReleaseScope != a.ReleaseScope || request.OutputKind != a.OutputKind ||
			!slices.Equal(request.RecipientPartitionRefs, a.RecipientPartitionRefs) ||
			!equalFrozenJSON(request.OutputOptions, a.OutputOptions) || !equalFrozenJSON(request.GraphProjectionRefs, a.GraphProjectionRefs) ||
			!equalOptionalString(optionalUUIDStringPointer(request.CompositionID), a.CompositionID) ||
			!equalOptionalString(request.CompositionVersion, a.CompositionVersion) || !equalOptionalString(request.CompositionSHA256, a.CompositionSHA256) ||
			a.ExpectedProfileSHA256 != nil {
			return invalid
		}
		if a.CompositionID == nil {
			if len(a.CompositionJSON) != 0 && !bytes.Equal(a.CompositionJSON, []byte("null")) {
				return invalid
			}
		} else {
			source, err := decodeCompositionRenderInput(a.CompositionJSON)
			if err != nil || source.SchemaID != reportcomposition.CompositionSchemaID || source.CompositionID != *a.CompositionID ||
				source.IncidentID != a.IncidentID || source.TemplateID != a.TemplateContract.TemplateID || source.TemplateVersion != a.TemplateContract.TemplateVersion ||
				!equalOptionalString(source.CompositionVersion, a.CompositionVersion) || !equalOptionalString(source.CompositionSHA256, a.CompositionSHA256) {
				return invalid
			}
		}
		return nil
	}
	request, apiErr := reportcomposition.DecodeNormalizedPreviewRequest(a.NormalizedRequest)
	if apiErr != nil || !bytes.Equal(request.Normalized, a.NormalizedRequest) || request.ClientTxnID != a.ClientTxnID ||
		request.SnapshotID != a.SnapshotID || request.DerivationVersion != a.DerivationVersion ||
		request.TemplateID != a.TemplateContract.TemplateID || request.TemplateVersion != a.TemplateContract.TemplateVersion ||
		request.RedactionProfileID != a.Profile.ProfileID || request.RedactionProfileVersion != a.Profile.Version ||
		a.ExpectedProfileSHA256 == nil || request.RedactionProfileSHA256 != *a.ExpectedProfileSHA256 ||
		request.OutputKind != a.OutputKind || request.RenderEnvironmentProfile != "cartulary.reporting.local.v1" ||
		!equalFrozenJSON(request.OutputOptions, a.OutputOptions) || !equalFrozenJSON(request.GraphProjectionRefs, a.GraphProjectionRefs) ||
		!equalFrozenJSON(request.RecipientPartitionRefs, []byte(`[]`)) {
		return invalid
	}
	var requestedVersion *string
	if request.CompositionVersion != nil {
		value := fmt.Sprintf("v%d", *request.CompositionVersion)
		requestedVersion = &value
	}
	if !equalOptionalString(requestedVersion, a.CompositionVersion) {
		return invalid
	}
	source, err := reportcomposition.DecodeRenderSource(a.CompositionJSON)
	if err != nil || source.SchemaID != reportcomposition.PreviewSourceSchemaID || a.CompositionID == nil || source.CompositionID != *a.CompositionID ||
		source.IncidentID != a.IncidentID || source.TemplateID != a.TemplateContract.TemplateID || source.TemplateVersion != a.TemplateContract.TemplateVersion ||
		source.SourceKind != request.SourceKind || !equalOptionalString(source.CompositionVersion, a.CompositionVersion) ||
		!equalOptionalString(source.CompositionSHA256, a.CompositionSHA256) {
		return invalid
	}
	return nil
}
