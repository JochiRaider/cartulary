package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Admission freezes selectors and owner bytes without expanding source content.
// The worker verifies the requested materialized profile digest during preparation.
func buildPreviewAdmissionTx(ctx context.Context, tx pgx.Tx, source reportcomposition.PreviewRenderSource, normalized []byte, templates *TemplatePackRegistry) (renderAdmission, error) {
	var empty renderAdmission
	if source.RenderEnvironmentProfile != "cartulary.reporting.local.v1" {
		return empty, reportcomposition.ErrPreviewInputRejected
	}
	digest, err := reportcomposition.PreviewSourceDigest(source.PreviewSourceJSON)
	if err != nil || source.PreviewAttemptID == uuid.Nil || digest != source.PreviewSourceSHA256 || source.DerivationVersion != DerivationVersion {
		return empty, errors.New("reporting: invalid preview source binding")
	}
	snapshotID, err := uuid.Parse(source.SnapshotID)
	if err != nil {
		return empty, reportcomposition.ErrPreviewInputRejected
	}
	snapshot, err := readSnapshotDescriptorTx(ctx, tx, snapshotID)
	if errors.Is(err, ErrNotFound) {
		return empty, reportcomposition.ErrPreviewInputRejected
	}
	if err != nil {
		return empty, err
	}
	if snapshot.Incident != source.IncidentID || snapshot.Derivation != source.DerivationVersion {
		return empty, reportcomposition.ErrPreviewInputRejected
	}
	var recipients []string
	if json.Unmarshal(source.RecipientPartitionRefs, &recipients) != nil || recipients == nil || len(recipients) != 0 {
		return empty, reportcomposition.ErrPreviewInputRejected
	}
	pack, ok := templates.resolve(source.TemplateID, source.TemplateVersion)
	if !ok || !pack.Manifest.SupportsOutputKind(source.OutputKind) || !pack.Manifest.SupportsReleaseScope(ReleaseScopeInternalDraft) {
		return empty, reportcomposition.ErrPreviewInputRejected
	}
	profile, err := selectedRedactionProfile(source.RedactionProfileID, source.RedactionProfileVersion, nil)
	if err != nil || !profileSupportsScope(RedactionProfileView{ProfileStatus: profile.ProfileStatus, AllowedReleaseScopes: profile.AllowedReleaseScopes}, ReleaseScopeInternalDraft) {
		return empty, reportcomposition.ErrPreviewInputRejected
	}
	profileSHA, err := ValidateRedactionProfile(profile)
	if err != nil {
		return empty, err
	}
	if _, err := parseRenderOutputOptions(source.OutputOptions, source.OutputKind, ReleaseScopeInternalDraft); err != nil {
		return empty, reportcomposition.ErrPreviewInputRejected
	}
	templateSHA, err := templateManifestDigest(pack.Manifest)
	if err != nil {
		return empty, err
	}
	compositionID, previewID := source.CompositionID.String(), source.PreviewAttemptID.String()
	var version *string
	if source.CompositionVersion != nil {
		value := fmt.Sprintf("v%d", *source.CompositionVersion)
		version = &value
	}
	stamp, err := normalizedReportingTimestamp(source.CreatedAt)
	if err != nil {
		return empty, err
	}
	request, apiErr := reportcomposition.DecodeNormalizedPreviewRequest(normalized)
	if apiErr != nil {
		return empty, reportcomposition.ErrPreviewInputRejected
	}
	return renderAdmission{
		SchemaID: renderAdmissionSchema, Algorithm: renderPreparationAlgorithm,
		Identity:    RenderIdentity{PreviewAttemptID: &previewID, RenderAdmittedAt: stamp},
		ActorUserID: source.CreatedByUserID.String(), IncidentID: snapshot.Incident.String(), SnapshotID: snapshot.ID.String(), SnapshotAt: snapshot.At,
		SourceChangeSetHighWatermark: snapshot.Boundary, DerivationVersion: snapshot.Derivation, SourceCaptureSHA256: snapshot.CaptureSHA256, ReferencePacks: snapshot.ReferencePacks,
		TemplateContract: pack.Manifest, TemplateFiles: append([]TemplatePackFile{}, pack.Files...), TemplateSHA256: templateSHA,
		Profile: profile, ProfileSHA256: profileSHA, ExpectedProfileSHA256: &source.RedactionProfileSHA256,
		ReleaseScope: ReleaseScopeInternalDraft, RecipientPartitionRefs: []string{}, OutputKind: source.OutputKind,
		OutputOptions: append(json.RawMessage(nil), source.OutputOptions...), GraphProjectionRefs: append(json.RawMessage(nil), source.GraphProjectionRefs...),
		CompositionJSON: append(json.RawMessage(nil), source.PreviewSourceJSON...), CompositionID: &compositionID, CompositionVersion: version, CompositionSHA256: cloneStringPtr(source.CompositionSHA256),
		ClientTxnID: request.ClientTxnID, NormalizedRequest: append([]byte(nil), normalized...),
	}, nil
}
