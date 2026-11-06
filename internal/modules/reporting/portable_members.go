package reporting

import (
	"context"
	"errors"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/artifactport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// A release retains only metadata and its bounded cross-file validation inputs.
// Large binary outputs are verified incrementally; source and SVG validation
// uses one owner-bounded member at a time, independently of archive size.
func validatePortableReleaseMembers(ctx context.Context, members *incidentportability.BoundMembers, manifest RenderBundleManifest, releaseID string, used map[string]bool) error {
	previous := ""
	primary, validationCount := false, 0
	localCount, localBytes := int64(0), int64(0)
	checks := make([]RenderBundleFile, 0, len(manifest.Files))
	controlBytes := int64(0)
	for _, f := range manifest.Files {
		if !portableMemberPath(f.Path) || f.Path <= previous || f.SizeBytes < 0 || !portableDigestPattern.MatchString(f.SHA256) || !shareableBundleMember(f) {
			return artifactport.ErrInvalid
		}
		previous = f.Path
		name := portableReleasePath(releaseID, f.Path)
		expected := incidentportability.Member{Path: name, SHA256: f.SHA256, ByteSize: f.SizeBytes}
		source, err := incidentportability.OpenVerifiedMember(ctx, members, expected)
		if err != nil {
			return artifactport.ErrInvalid
		}
		var data []byte
		maximum := int64(0)
		keep := false
		switch f.Role {
		case renderBundleRoleLocalAsset, renderBundleRoleLocalTheme:
			localCount++
			localBytes += f.SizeBytes
			if localCount > renderLimitRules["local_assets.count"].hardMax || localBytes > renderLimitRules["local_assets.total_bytes"].hardMax {
				_ = source.Close()
				return artifactport.ErrInvalid
			}
			maximum = renderLimitRules["local_asset.byte_size"].hardMax
			keep = f.Role == renderBundleRoleLocalTheme
		case renderBundleRoleRenderedSVG:
			maximum = 500 << 20 // Existing renderer supervisor bound.
		case renderBundleRoleSourceSlidev, renderBundleRoleSourceMermaid:
			maximum = incidentportability.MaxControlMemberBytes
		case renderBundleRoleValidationSummary, renderBundleRoleToolchainSnapshot,
			renderBundleRoleRedactionManifest, renderBundleRoleRedactionProfileView, renderBundleRoleTokenManifest:
			maximum = incidentportability.MaxControlMemberBytes
			controlBytes += f.SizeBytes
			keep = true
			if controlBytes > incidentportability.MaxControlMemberBytes {
				_ = source.Close()
				return artifactport.ErrInvalid
			}
		case renderBundleRoleRenderedPDF:
			if f.SizeBytes > renderLimitRules["rendered_pdf.byte_size"].hardMax {
				_ = source.Close()
				return artifactport.ErrInvalid
			}
		}
		if maximum > 0 {
			if f.SizeBytes > maximum {
				_ = source.Close()
				return artifactport.ErrInvalid
			}
			data, err = io.ReadAll(io.LimitReader(source, maximum+1))
		} else {
			_, err = io.CopyBuffer(io.Discard, source, make([]byte, 64<<10))
		}
		err = errors.Join(err, source.Close())
		if err != nil {
			return artifactport.ErrInvalid
		}
		switch f.Role {
		case renderBundleRoleSourceSlidev:
			err = ValidateSelfContainedOutput(OutputKindSlidev, data)
		case renderBundleRoleSourceMermaid:
			err = ValidateSelfContainedOutput(OutputKindMermaid, data)
		case renderBundleRoleRenderedSVG:
			err = validateRenderedSVG(data)
		case renderBundleRoleLocalAsset, renderBundleRoleLocalTheme:
			err = validateTemplateFileContent(f.MediaType, data)
		case renderBundleRoleValidationSummary:
			validationCount++
			if f.Path != "validation/summary.json" {
				return artifactport.ErrInvalid
			}
			err = validatePassedRenderSummary(data, manifest)
		}
		if err != nil {
			return artifactport.ErrInvalid
		}
		check := RenderBundleFile{RenderBundleManifestFile: f}
		if keep {
			check.Bytes = data
		}
		checks = append(checks, check)
		used[name] = true
		if f.Path == manifest.PrimaryPath && f.MediaType == manifest.PrimaryMediaType {
			primary = true
		}
	}
	if !primary || validationCount != 1 || validatePublicRedactionArtifacts(manifest, checks) != nil || validateBundleToolchain(manifest, checks) != nil || validateThemeBundleFiles(checks) != nil {
		return artifactport.ErrInvalid
	}
	return nil
}
