package reporting

import (
	"encoding/json"
	"strings"
	"testing"
)

func testRetainedToolchainBinding(t *testing.T, bundle RenderBundle) {
	t.Helper()
	if err := validateBundleToolchain(bundle.Manifest, bundle.Files); err != nil {
		t.Fatal("valid retained toolchain rejected", err)
	}
	index := -1
	for i, file := range bundle.Files {
		if file.Role == renderBundleRoleToolchainSnapshot {
			index = i
		}
	}
	if index < 0 {
		t.Fatal("missing toolchain fixture")
	}
	for name, mutate := range map[string]func(map[string]any){
		"missing schema":         func(m map[string]any) { delete(m, "schema_id") },
		"unknown schema":         func(m map[string]any) { m["schema_id"] = "retired" },
		"unknown field":          func(m map[string]any) { m["private_content"] = "do-not-echo" },
		"missing nullable":       func(m map[string]any) { delete(m, "os_image_sha256") },
		"wrong admitted time":    func(m map[string]any) { m["created_at"] = "2026-01-01T00:00:00.000000Z" },
		"missing version":        func(m map[string]any) { m["node_version"] = "" },
		"bad digest":             func(m map[string]any) { m["font_manifest_sha256"] = "bad" },
		"retired font schema":    func(m map[string]any) { m["schema_id"] = "cartulary.reporting_toolchain_snapshot.v1" },
		"missing font inventory": func(m map[string]any) { delete(m, "font_files") },
		"empty font inventory":   func(m map[string]any) { m["font_files"] = []any{} },
		"altered font inventory": func(m map[string]any) {
			m["font_files"].([]any)[0].(map[string]any)["sha256"] = strings.Repeat("0", 64)
		},
		"wrong environment":   func(m map[string]any) { m["timezone"] = "local" },
		"wrong scale":         func(m map[string]any) { m["device_scale_factor"] = 2 },
		"null argv":           func(m map[string]any) { m["render_command"] = nil },
		"null environment":    func(m map[string]any) { m["env_allowlist"] = nil },
		"empty argv":          func(m map[string]any) { m["render_command"] = []string{} },
		"browser not bound":   func(m map[string]any) { m["chromium_executable_sha256"] = nil },
		"absolute browser":    func(m map[string]any) { m["chromium_executable_path"] = "/private/browser" },
		"unexecuted renderer": func(m map[string]any) { m["mermaid_version"] = "extra" },
	} {
		t.Run("toolchain "+name, func(t *testing.T) {
			var value map[string]any
			if err := json.Unmarshal(bundle.Files[index].Bytes, &value); err != nil {
				t.Fatal(err)
			}
			mutate(value)
			encoded, err := canonicalJSON(value)
			if err != nil {
				t.Fatal(err)
			}
			files := append([]RenderBundleFile{}, bundle.Files...)
			files[index].Bytes = encoded
			manifest := bundle.Manifest
			digest := reportingObjectDigest("cartulary.reporting_toolchain_snapshot.v2", encoded)
			manifest.ToolchainSnapshotSHA256 = &digest
			if err := validateBundleToolchain(manifest, files); err == nil || strings.Contains(err.Error(), "do-not-echo") {
				t.Fatal("rehashed invalid toolchain accepted or echoed", err)
			}
		})
	}
	for _, mutation := range []string{"missing", "duplicate", "path", "media", "optional", "root digest", "validation roster"} {
		candidate := bundle
		candidate.Files = append([]RenderBundleFile{}, bundle.Files...)
		switch mutation {
		case "missing":
			candidate.Files = append(candidate.Files[:index], candidate.Files[index+1:]...)
		case "duplicate":
			candidate.Files = append(candidate.Files, candidate.Files[index])
		case "path":
			candidate.Files[index].Path = "validation/other.json"
		case "media":
			candidate.Files[index].MediaType = "application/octet-stream"
		case "optional":
			candidate.Files[index].RequiredForRelease = false
		case "root digest":
			candidate.Manifest.ToolchainSnapshotSHA256 = nil
		case "validation roster":
			candidate.Manifest.ValidationArtifacts = []string{"validation/summary.json"}
		}
		if validateBundleToolchain(candidate.Manifest, candidate.Files) == nil {
			t.Fatal("invalid toolchain membership accepted", mutation)
		}
	}
}
