package reporting

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func assertPassingRenderSummary(t *testing.T) {
	t.Helper()
	for _, preview := range []bool{false, true} {
		name := "release validation summary"
		if preview {
			name = "preview validation summary"
		}
		t.Run(name, func(t *testing.T) {
			contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
			identity := fixtureRenderIdentity()
			if preview {
				identity.PreviewAttemptID, identity.ReleaseID = identity.ReleaseID, nil
			}
			model := RedactedExportModel{ExportModel: ExportModel{RenderIdentity: identity, SchemaID: ExportModelSchemaID, SnapshotContent: SnapshotContent{SnapshotID: "40000000-0000-0000-0000-000000000002", DerivationVersion: DerivationVersion}}, Fields: []RedactedField{
				{Path: "/incident/title", ContentClass: ContentClassCuratedNarrative, Value: "Synthetic summary"},
				{Path: "/incident/status", ContentClass: ContentClassDerivedAnalytic, Value: "active"},
			}}
			bindFixtureRetainedSections(&model)
			bundle, err := fixtureRenderer().fixtureRenderReportBundle(context.Background(), contract, OutputKindSlidev, model, RedactionProfileView{}, strings.Repeat("a", 64), ReleaseScopeInternalDraft, fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalDraft), nil, RedactionBundleArtifacts{})
			if err != nil {
				t.Fatal(err)
			}
			testRetainedToolchainBinding(t, bundle)
			var raw []byte
			for _, file := range bundle.Files {
				if file.Role == renderBundleRoleValidationSummary {
					raw = file.Bytes
				}
			}
			if err := validatePassedRenderSummary(raw, bundle.Manifest); err != nil {
				t.Fatal("valid summary rejected", err)
			}
			mutations := map[string]func(map[string]any){
				"missing release member": func(m map[string]any) { delete(m, "release_id") },
				"wrong release identity": func(m map[string]any) { m["release_id"] = model.SnapshotID },
				"missing snapshot":       func(m map[string]any) { delete(m, "snapshot_id") },
				"null snapshot":          func(m map[string]any) { m["snapshot_id"] = nil },
				"wrong snapshot":         func(m map[string]any) { m["snapshot_id"] = "40000000-0000-0000-0000-000000000003" },
				"changed admission time": func(m map[string]any) { m["created_at"] = "2026-01-01T00:00:00.000000Z" },
				"failed summary":         func(m map[string]any) { m["result"] = "failed" },
				"early stage":            func(m map[string]any) { m["terminal_stage"] = "slidev_source" },
				"undeclared top level":   func(m map[string]any) { m["private_content"] = "do-not-echo" },
				"undeclared diagnostic":  func(m map[string]any) { m["safe_details"].(map[string]any)["source_only"] = true },
				"forbidden diagnostic":   func(m map[string]any) { m["safe_details"].(map[string]any)["path"] = "do-not-echo" },
				"null issues":            func(m map[string]any) { m["issues"] = nil },
				"nonzero issues":         func(m map[string]any) { m["issue_count"] = 1 },
				"failure code":           func(m map[string]any) { m["failure_code"] = "render_timeout" },
			}
			for name, mutate := range mutations {
				t.Run(name, func(t *testing.T) {
					var value map[string]any
					if err := json.Unmarshal(raw, &value); err != nil {
						t.Fatal(err)
					}
					mutate(value)
					changed, err := canonicalJSON(value)
					if err != nil {
						t.Fatal(err)
					}
					manifest := bundle.Manifest
					manifest.ValidationSummarySHA256 = reportingObjectDigest("cartulary.reporting_render_validation_summary.v1", changed)
					if err := validatePassedRenderSummary(changed, manifest); err == nil || strings.Contains(err.Error(), "do-not-echo") {
						t.Fatal("rehashed invalid summary admitted or diagnostic echoed", err)
					}
				})
			}
		})
	}
}
