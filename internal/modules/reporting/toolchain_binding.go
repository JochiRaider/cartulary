package reporting

import (
	"bytes"
	"errors"
	"slices"
	"strings"
)

func diagramToolUsage(files []RenderBundleFile) (browser, mermaid bool) {
	auto := map[string]bool{}
	for _, f := range files {
		if f.Role == renderBundleRoleSourceMermaid {
			mermaid = true
			auto[strings.TrimSuffix(f.Path, ".mmd")+".svg"] = true
		}
	}
	for _, f := range files {
		if f.Role == renderBundleRoleRenderedPDF || f.Role == renderBundleRoleRenderedSVG && auto[f.Path] {
			browser = true
		}
	}
	return
}

// Historical validation checks recorded bytes and the declared environment
// contract. It neither probes the host nor substitutes today's installation.
func validateBundleToolchain(manifest RenderBundleManifest, files []RenderBundleFile) error {
	invalid := errors.New("reporting: invalid retained toolchain artifact")
	if manifest.ToolchainSnapshotSHA256 == nil || !portableDigestPattern.MatchString(*manifest.ToolchainSnapshotSHA256) ||
		!slices.Equal(manifest.ValidationArtifacts, []string{"validation/summary.json", "validation/toolchain.json"}) {
		return invalid
	}
	var artifact *RenderBundleFile
	browserUsed, mermaidUsed := diagramToolUsage(files)
	for i := range files {
		file := &files[i]
		switch file.Role {
		case renderBundleRoleToolchainSnapshot:
			if artifact != nil {
				return invalid
			}
			artifact = file
		}
	}
	if artifact == nil || artifact.Path != "validation/toolchain.json" || artifact.MediaType != "application/vnd.cartulary.reporting-toolchain+json" || !artifact.RequiredForRelease {
		return invalid
	}
	snapshot, encoded, err := decodePortableObject[ToolchainSnapshot](artifact.Bytes, *manifest.ToolchainSnapshotSHA256)
	if err != nil || !bytes.Equal(encoded, artifact.Bytes) || snapshot.SchemaID != "cartulary.reporting_toolchain_snapshot.v2" || snapshot.CreatedAt != manifest.RenderAdmittedAt ||
		snapshot.NodeVersion == "" || snapshot.PackageManager == "" || snapshot.OSImageID == "" ||
		!validReportingText(snapshot.NodeVersion, false) || !validReportingText(snapshot.PackageManager, false) || !validReportingText(snapshot.OSImageID, false) ||
		!portableDigestPattern.MatchString(snapshot.LockfileSHA256) || !portableDigestPattern.MatchString(snapshot.PackageStoreDigest) || !portableDigestPattern.MatchString(snapshot.FontManifestSHA256) ||
		snapshot.Locale != "C.UTF-8" || snapshot.Timezone != "UTC" || snapshot.ViewportCSSPX != "1280x720" || snapshot.DeviceScaleFactor != 1 ||
		(snapshot.ColorScheme != "light" && snapshot.ColorScheme != "dark") || snapshot.NetworkPolicyID != "render_sandbox_policy_v1" ||
		snapshot.BrowserLaunchArgs == nil || snapshot.EnvAllowlist == nil || len(snapshot.RenderCommand) == 0 {
		return invalid
	}
	if validateToolchainFonts(snapshot, files) != nil {
		return invalid
	}
	versionPresent := func(version *string, required bool) bool {
		return !required && version == nil || required && version != nil && *version != "" && validReportingText(*version, false)
	}
	if !versionPresent(snapshot.SlidevVersion, manifest.OutputKind == OutputKindSlidev) || !versionPresent(snapshot.MermaidVersion, mermaidUsed) || !versionPresent(snapshot.ChromiumVersion, browserUsed) {
		return invalid
	}
	if browserUsed {
		if snapshot.ChromiumExecutablePath == nil || !portableMemberPath(*snapshot.ChromiumExecutablePath) || snapshot.ChromiumExecutableSHA256 == nil || !portableDigestPattern.MatchString(*snapshot.ChromiumExecutableSHA256) {
			return invalid
		}
	} else if snapshot.ChromiumExecutablePath != nil || snapshot.ChromiumExecutableSHA256 != nil || len(snapshot.BrowserLaunchArgs) != 0 {
		return invalid
	}
	if snapshot.OSImageSHA256 != nil && !portableDigestPattern.MatchString(*snapshot.OSImageSHA256) {
		return invalid
	}
	previous := ""
	for _, entry := range snapshot.EnvAllowlist {
		if entry.Name <= previous || !validReportingIdentifier(entry.Name) || !portableDigestPattern.MatchString(entry.ValueSHA256) {
			return invalid
		}
		previous = entry.Name
	}
	for _, args := range [][]string{snapshot.BrowserLaunchArgs, snapshot.RenderCommand} {
		for _, arg := range args {
			if arg == "" || !validReportingText(arg, false) {
				return invalid
			}
		}
	}
	return nil
}
