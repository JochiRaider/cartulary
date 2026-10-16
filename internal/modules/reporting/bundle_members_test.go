package reporting

import "testing"

// Reachable through the existing artifact-boundary catalog entrypoint.
func assertShareableBundleMembers(t *testing.T) {
	t.Helper()
	members := []RenderBundleManifestFile{
		{Path: "validation/summary.json", Role: renderBundleRoleValidationSummary, MediaType: "application/vnd.cartulary.reporting-validation+json"},
		{Path: "validation/toolchain.json", Role: renderBundleRoleToolchainSnapshot, MediaType: "application/vnd.cartulary.reporting-toolchain+json"},
		{Path: "validation/redaction-manifest.json", Role: renderBundleRoleRedactionManifest, MediaType: "application/vnd.cartulary.redaction-manifest+json"},
		{Path: "validation/redaction-profile-view.json", Role: renderBundleRoleRedactionProfileView, MediaType: "application/vnd.cartulary.redaction-profile-view+json"},
		{Path: "validation/token-manifest.json", Role: renderBundleRoleTokenManifest, MediaType: "application/vnd.cartulary.reporting-token-manifest+json"},
		{Path: "slides.md", Role: renderBundleRoleSourceSlidev, MediaType: "text/markdown; charset=utf-8"},
		{Path: "diagrams/d-1.mmd", Role: renderBundleRoleSourceMermaid, MediaType: "text/vnd.cartulary.mermaid; charset=utf-8"},
		{Path: "deck.pdf", Role: renderBundleRoleRenderedPDF, MediaType: "application/pdf"},
		{Path: "diagrams/d-1.svg", Role: renderBundleRoleRenderedSVG, MediaType: "image/svg+xml"},
		{Path: "intermediate/deck-model.json", Role: renderBundleRoleDeckModel, MediaType: "application/vnd.cartulary.reporting-deck-model+json"},
		{Path: "intermediate/diagrams/d-1.json", Role: renderBundleRoleDiagramModel, MediaType: "application/vnd.cartulary.reporting-diagram+json"},
	}
	for _, member := range members {
		member.RequiredForRelease = true
		t.Run("bundle contract/"+member.Role, func(t *testing.T) {
			if !shareableBundleMember(member) {
				t.Fatal("declared member rejected")
			}
			for _, other := range members {
				if other.Role == member.Role {
					continue
				}
				changed := member
				changed.Path = other.Path
				if shareableBundleMember(changed) {
					t.Fatalf("role accepted another role's path: %s", other.Role)
				}
				changed = member
				changed.MediaType = other.MediaType
				if other.MediaType != member.MediaType && shareableBundleMember(changed) {
					t.Fatalf("role accepted another media type: %s", other.MediaType)
				}
			}
			for _, candidate := range []string{"manifest.json", "internal/reveal-map.json", "unowned/" + member.Path, "../" + member.Path, member.Path + "/extra", member.Path + ".html"} {
				changed := member
				changed.Path = candidate
				if shareableBundleMember(changed) {
					t.Fatalf("invalid path admitted: %s", candidate)
				}
			}
			member.RequiredForRelease = false
			if shareableBundleMember(member) {
				t.Fatal("required member marked optional")
			}
		})
	}
	for _, invalid := range []string{"assets/a/cafe\u0301.png", "assets/a/../b.png", "assets//b.png", "assets/a/./b.png", "/assets/b.png", "C:/assets/b.png", "assets\\b.png", "assets/a/\x00.png", "assets/a/\xff.png"} {
		if portableMemberPath(invalid) {
			t.Fatalf("noncanonical path admitted: %q", invalid)
		}
	}
	if !portableMemberPath("assets/a/caf\u00e9.png") {
		t.Fatal("valid NFC path rejected")
	}
}
