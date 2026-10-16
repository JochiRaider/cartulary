package reporting

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// A reveal map is an internal adjunct. It cannot be supplied to the renderer
// or enumerated by any ordinary bundle reader, including incident portability.
func validateRevealMap(rendered RenderedRelease) error {
	r := rendered.Redaction
	m := rendered.RenderBundle.Manifest
	invalid := func() error { return fmt.Errorf("reporting sensitive artifact binding invalid") }
	if m.TokenManifestSHA256 == nil {
		if len(r.RevealMapJSON) != 0 || r.RevealMapSHA256 != "" || len(r.TokenManifestJSON) != 0 || r.TokenManifestSHA256 != "" {
			return invalid()
		}
		return nil
	}
	tokens, tokenBytes, err := decodePortableObject[RedactionTokenManifest](r.TokenManifestJSON, *m.TokenManifestSHA256)
	if err != nil || !bytes.Equal(tokenBytes, r.TokenManifestJSON) || r.TokenManifestSHA256 != *m.TokenManifestSHA256 || tokens.SchemaID != RedactionTokenManifestSchemaID || len(tokens.Entries) == 0 {
		return invalid()
	}
	redaction, redactionBytes, err := decodePortableObject[RedactionManifest](rendered.RedactionManifestJSON, m.RedactionManifestSHA256)
	if err != nil || !bytes.Equal(redactionBytes, rendered.RedactionManifestJSON) || rendered.RedactionManifestSHA256 != m.RedactionManifestSHA256 ||
		redaction.SnapshotID != m.SnapshotID || !sameReportingValue(redaction.RenderIdentity, m.RenderIdentity) ||
		validateTokenManifestBindings(tokens, redaction, m.RedactionManifestSHA256) != nil {
		return invalid()
	}
	reveal, revealBytes, err := decodePortableObject[RedactionRevealMap](r.RevealMapJSON, r.RevealMapSHA256)
	if err != nil || !bytes.Equal(revealBytes, r.RevealMapJSON) || reveal.SchemaID != RedactionRevealMapSchemaID ||
		reveal.TokenManifestSHA256 != *m.TokenManifestSHA256 || reveal.ProfileSHA256 != tokens.ProfileSHA256 ||
		reveal.SnapshotID != m.SnapshotID || reveal.CreatedAt != m.RenderAdmittedAt ||
		!sameReportingValue(reveal.RenderIdentity, m.RenderIdentity) || len(reveal.Entries) != len(tokens.Entries) {
		return invalid()
	}
	// The private artifact contains one canonical display value per subject,
	// never a second inventory of arbitrary original field values. Its digest
	// must equal the independently bound public subject digest.
	for i, entry := range reveal.Entries {
		token := tokens.Entries[i]
		digest, err := subjectDisplayValueDigest(entry.StableSubjectRef, &entry.CanonicalDisplayValue)
		if err != nil || !validReportingText(entry.CanonicalDisplayValue, true) || digest != token.CanonicalDisplayValueSHA256 ||
			entry.TokenID != token.TokenID || entry.DisplayToken != token.DisplayToken || entry.StableSubjectRef != token.StableSubjectRef ||
			entry.SubjectFamily != token.SubjectFamily || !equalOptionalString(entry.SourceRecordID, token.SourceRecordID) ||
			!equalOptionalString(entry.EntityMentionID, token.EntityMentionID) || entry.CanonicalDisplayValueSHA256 != token.CanonicalDisplayValueSHA256 ||
			entry.RecipientPartitionRefs == nil || !slices.Equal(entry.RecipientPartitionRefs, token.RecipientPartitionRefs) {
			return invalid()
		}
	}

	return nil
}

func insertReleaseRevealMapTx(ctx context.Context, tx pgx.Tx, releaseID uuid.UUID, rendered RenderedRelease, now time.Time) error {
	if len(rendered.Redaction.RevealMapJSON) == 0 {
		return nil
	}
	_, err := tx.Exec(ctx, `INSERT INTO reporting_release_reveal_maps
    (release_id,reveal_map_sha256,canonical_bytes,created_at) VALUES ($1,$2,$3,$4)`,
		releaseID, rendered.Redaction.RevealMapSHA256, rendered.Redaction.RevealMapJSON, now.UTC())
	return err
}

func insertPreviewRevealMapTx(ctx context.Context, tx pgx.Tx, previewID uuid.UUID, rendered RenderedRelease, now time.Time) error {
	if len(rendered.Redaction.RevealMapJSON) == 0 {
		return nil
	}
	_, err := tx.Exec(ctx, `INSERT INTO reporting_preview_reveal_maps
    (preview_attempt_id,reveal_map_sha256,canonical_bytes,created_at) VALUES ($1,$2,$3,$4)`,
		previewID, rendered.Redaction.RevealMapSHA256, rendered.Redaction.RevealMapJSON, now.UTC())
	return err
}

// One policy governs producer persistence, portable import/export and recovery.
// A role is not permission to publish arbitrary paths or media. New producers
// must admit their complete role/path/media contract here before publishing.
func shareableBundleMember(file RenderBundleManifestFile) bool {
	if !portableMemberPath(file.Path) || !file.RequiredForRelease {
		return false
	}
	exact := func(path, media string) bool { return file.Path == path && file.MediaType == media }
	diagram := func(prefix, suffix, media string) bool {
		id := strings.TrimSuffix(strings.TrimPrefix(file.Path, prefix), suffix)
		return file.MediaType == media && validReportingIdentifier(id) && file.Path == prefix+id+suffix
	}
	switch file.Role {
	case renderBundleRoleValidationSummary:
		return exact("validation/summary.json", "application/vnd.cartulary.reporting-validation+json")
	case renderBundleRoleToolchainSnapshot:
		return exact("validation/toolchain.json", "application/vnd.cartulary.reporting-toolchain+json")
	case renderBundleRoleRedactionManifest:
		return exact("validation/redaction-manifest.json", "application/vnd.cartulary.redaction-manifest+json")
	case renderBundleRoleRedactionProfileView:
		return exact("validation/redaction-profile-view.json", "application/vnd.cartulary.redaction-profile-view+json")
	case renderBundleRoleTokenManifest:
		return exact("validation/token-manifest.json", "application/vnd.cartulary.reporting-token-manifest+json")
	case renderBundleRoleSourceSlidev:
		return exact("slides.md", "text/markdown; charset=utf-8")
	case renderBundleRoleSourceMermaid:
		return diagram("diagrams/", ".mmd", "text/vnd.cartulary.mermaid; charset=utf-8")
	case renderBundleRoleRenderedPDF:
		return exact("deck.pdf", "application/pdf")
	case renderBundleRoleRenderedSVG:
		return diagram("diagrams/", ".svg", "image/svg+xml")
	case renderBundleRoleDeckModel:
		return exact("intermediate/deck-model.json", "application/vnd.cartulary.reporting-deck-model+json")
	case renderBundleRoleDiagramModel:
		return diagram("intermediate/diagrams/", ".json", "application/vnd.cartulary.reporting-diagram+json")
	case renderBundleRoleLocalAsset:
		parts := strings.Split(file.Path, "/")
		return len(parts) == 3 && validReportingIdentifier(parts[1]) && templateAssetMember(TemplateAsset{AssetID: parts[1], AssetRole: file.Role, BundlePath: file.Path, MediaType: file.MediaType})
	case renderBundleRoleLocalTheme:
		return templateThemeMember(themeFileDeclaration(file.Path, file.MediaType))
	default:
		return false
	}
}

func validateShareableBundle(bundle RenderBundle) error {
	invalid := func() error { return fmt.Errorf("reporting shareable bundle invalid") }
	manifest, encoded, err := decodePortableObject[RenderBundleManifest](bundle.ManifestJSON, bundle.ManifestSHA256)
	projected, projectionErr := canonicalJSON(bundle.Manifest)
	if err != nil || projectionErr != nil || !bytes.Equal(encoded, bundle.ManifestJSON) || !bytes.Equal(encoded, projected) || validateBundleIdentity(manifest) != nil || manifest.SchemaID != RenderBundleManifestSchemaID || manifest.ManifestVersion != RenderBundleManifestVersion || manifest.PrimaryPath != bundle.PrimaryPath || manifest.PrimaryMediaType != bundle.PrimaryMedia || len(manifest.Files) == 0 || len(manifest.Files) != len(bundle.Files) {
		return invalid()
	}
	previous := ""
	primary := false
	tokenCount := 0
	for i, file := range bundle.Files {
		if !shareableBundleMember(file.RenderBundleManifestFile) || file.Path <= previous || file.RenderBundleManifestFile != manifest.Files[i] || file.StorageKind != renderBundleStorageInline || file.SizeBytes != int64(len(file.Bytes)) || file.SHA256 != hashHex(file.Bytes) {
			return invalid()
		}
		if file.Role == renderBundleRoleTokenManifest {
			tokenCount++
			if manifest.TokenManifestSHA256 == nil || reportingObjectDigest(RedactionTokenManifestSchemaID, file.Bytes) != *manifest.TokenManifestSHA256 {
				return invalid()
			}
		}
		previous = file.Path
		primary = primary || (file.Path == bundle.PrimaryPath && file.MediaType == bundle.PrimaryMedia)
	}
	if !primary || tokenCount > 1 || ((manifest.TokenManifestSHA256 != nil) != (tokenCount == 1)) {
		return invalid()
	}
	if err := validateBundleToolchain(manifest, bundle.Files); err != nil {
		return err
	}
	if err := validateBundleFilesSelfContained(bundle.Files); err != nil {
		return err
	}
	return validatePublicRedactionArtifacts(manifest, bundle.Files)
}
