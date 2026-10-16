package reporting

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
	"github.com/JochiRaider/cartulary/internal/testutil/reportingassets"
)

const fixtureThemeCSS = `@font-face {
font-family: "rpt-body";
src: url("fonts/body.woff2") format("woff2");
font-style: normal;
font-weight: 400;
font-display: block;
}
:root {
--cartulary-report-text: #183650;
--cartulary-report-accent: #238853;
--cartulary-report-font-family: "rpt-body";
--cartulary-report-font-size: 24px;
--cartulary-report-logo: url("images/brand.svg");
}
`

func fixtureTemplateThemeInput(t *testing.T, scope string) renderInput {
	t.Helper()
	input := fixtureTemplateAssetInputForScope(t, scope)
	input.TemplateID = "test.local-theme"
	input.TemplateContract.TemplateID = input.TemplateID
	for _, file := range []struct {
		id, path, media string
		bytes           []byte
	}{
		{"body", "theme/fonts/body.woff2", "font/woff2", reportingassets.ThemeFont()},
		{"brand", "theme/images/brand.svg", "image/svg+xml", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20"><rect width="20" height="20" fill="#238853"/></svg>`)},
		{"theme", "theme/style.css", themeCSSMedia, []byte(fixtureThemeCSS)},
	} {
		input.TemplateContract.Assets = append(input.TemplateContract.Assets, TemplateAsset{SchemaID: "template_asset_item.v1", AssetID: file.id, AssetRole: "local_theme", BundlePath: file.path, SourcePath: file.path, MediaType: file.media, SHA256: hashHex(file.bytes), ByteSize: int64(len(file.bytes)), RequiredForRelease: true})
		input.TemplateFiles = append(input.TemplateFiles, TemplatePackFile{SourcePath: file.path, Bytes: file.bytes})
	}
	sort.Slice(input.TemplateContract.Assets, func(i, j int) bool {
		return input.TemplateContract.Assets[i].BundlePath < input.TemplateContract.Assets[j].BundlePath
	})
	sort.Slice(input.TemplateFiles, func(i, j int) bool { return input.TemplateFiles[i].SourcePath < input.TemplateFiles[j].SourcePath })
	return input
}

func replaceThemeFixtureFile(input *renderInput, path string, value []byte) {
	for i := range input.TemplateFiles {
		if input.TemplateFiles[i].SourcePath == path {
			input.TemplateFiles[i].Bytes = value
		}
	}
	for i := range input.TemplateContract.Assets {
		if input.TemplateContract.Assets[i].SourcePath == path {
			input.TemplateContract.Assets[i].SHA256 = hashHex(value)
			input.TemplateContract.Assets[i].ByteSize = int64(len(value))
		}
	}
}

func testTemplateThemes(t *testing.T) {
	installed := []renderprocess.InstallationFile{
		{Path: "runtime/renderer/package/fonts/test.woff2", SHA256: strings.Repeat("a", 64)},
		{Path: "runtime/renderer/package/theme.css", SHA256: strings.Repeat("b", 64)},
		{Path: "runtime/fonts/fonts.conf", SHA256: strings.Repeat("c", 64)},
		{Path: "runtime/fonts/system.ttf", SHA256: strings.Repeat("d", 64)},
		{Path: "runtime/renderer/package/main.js", SHA256: strings.Repeat("e", 64)},
	}
	fonts := installedToolchainFontFiles(installed)
	if len(fonts) != 4 || fonts[0].Path != "installed/runtime/fonts/fonts.conf" || fonts[3].Path != "installed/runtime/renderer/package/theme.css" {
		t.Fatal("installed inventory omitted package webfonts or configuration")
	}
	for i, j := 0, len(installed)-1; i < j; i, j = i+1, j-1 {
		installed[i], installed[j] = installed[j], installed[i]
	}
	if !sameReportingValue(fonts, installedToolchainFontFiles(installed)) {
		t.Fatal("font inventory depends on enumeration order")
	}
	input := fixtureTemplateThemeInput(t, ReleaseScopeInternalReview)
	block := ReportingBlock{Fields: []ReportingField{{FieldKey: "asset_id", Value: "body"}}}
	if _, err := serializeSlidevAssetRef(block, input.TemplateContract); err == nil {
		t.Fatal("theme font accepted as image content")
	}
	for _, id := range []string{"body", "brand", "theme"} {
		candidate := fixtureTemplateThemeInput(t, ReleaseScopeInternalReview)
		candidate.TemplateContract.Sections[0].Bindings = []string{"$.assets['" + id + "']"}
		if validateTemplateContract(candidate.TemplateContract, OutputKindSlidev, ReleaseScopeInternalReview) == nil {
			t.Fatal("exact theme content binding accepted", id)
		}
	}
	if _, err := NewTemplatePackRegistry(TemplatePack{Manifest: input.TemplateContract, Files: input.TemplateFiles}); err != nil {
		t.Fatal("valid local theme", err)
	}
	for name, css := range map[string]string{
		"remote":             strings.ReplaceAll(fixtureThemeCSS, `url("fonts/body.woff2")`, `url("https://example.invalid/font")`),
		"traversal":          strings.ReplaceAll(fixtureThemeCSS, `fonts/body.woff2`, `../fonts/body.woff2`),
		"encoded":            strings.ReplaceAll(fixtureThemeCSS, `fonts/body.woff2`, `fonts/%62ody.woff2`),
		"import":             "@import \"other.css\";\n" + fixtureThemeCSS,
		"selector":           strings.ReplaceAll(fixtureThemeCSS, ":root", ".slidev-layout"),
		"hidden content":     strings.ReplaceAll(fixtureThemeCSS, "--cartulary-report-text: #183650;", "display: none;"),
		"expression":         strings.ReplaceAll(fixtureThemeCSS, "24px", "calc(1px + 23px)"),
		"escape":             strings.ReplaceAll(fixtureThemeCSS, "#238853", `\23 238853`),
		"comment":            fixtureThemeCSS + "/* hidden */",
		"duplicate root":     fixtureThemeCSS + ":root { --cartulary-report-text: #ffffff; }",
		"duplicate property": strings.ReplaceAll(fixtureThemeCSS, "24px;", "24px; --cartulary-report-font-size: 24px;"),
		"unselected font":    strings.ReplaceAll(fixtureThemeCSS, `--cartulary-report-font-family: "rpt-body";`, ""),
		"font fallback":      strings.ReplaceAll(fixtureThemeCSS, `--cartulary-report-font-family: "rpt-body";`, `--cartulary-report-font-family: "rpt-body", serif;`),
		"unselected logo":    strings.ReplaceAll(fixtureThemeCSS, `--cartulary-report-logo: url("images/brand.svg");`, ""),
		"font collection":    strings.ReplaceAll(fixtureThemeCSS, "font-weight: 400;", "font-weight: 100 900;"),
	} {
		t.Run("local theme/"+name, func(t *testing.T) {
			candidate := fixtureTemplateThemeInput(t, ReleaseScopeInternalReview)
			replaceThemeFixtureFile(&candidate, "theme/style.css", []byte(css))
			if _, err := NewTemplatePackRegistry(TemplatePack{Manifest: candidate.TemplateContract, Files: candidate.TemplateFiles}); err == nil {
				t.Fatal("invalid theme admitted")
			}
		})
	}
	for _, at := range []int{0, 4, 8, 14, 16, 20, 28, 32, 36, 40, 44} {
		font := reportingassets.ThemeFont()
		font[at] = 255
		if validateThemeFont(font) == nil {
			t.Fatal("invalid WOFF2 header admitted", at)
		}
	}
	rendered, reason, err := fixtureRenderer().Render(context.Background(), input)
	if err != nil {
		t.Fatal("theme preparation", reason, err)
	}
	assertTemplateThemeOutput(t, input, rendered)
	path := "$.assets['theme']"
	input.RedactionProfile.Rules = append(input.RedactionProfile.Rules, RedactionRule{RuleID: "deny-theme", Path: &path, Action: RedactionActionSpec{Type: ActionDrop}})
	input.RedactionProfileSHA256, err = ValidateRedactionProfile(input.RedactionProfile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := redactRenderInput(context.Background(), input); err == nil {
		t.Fatal("required theme drop silently selected default styling")
	}
}

func assertTemplateThemeOutput(t *testing.T, input renderInput, rendered RenderedRelease) {
	t.Helper()
	if err := validateRenderCandidateBinding(input, rendered); err != nil {
		t.Fatal(err)
	}
	if err := validateShareableBundle(rendered.RenderBundle); err != nil {
		t.Fatal(err)
	}
	themes := 0
	for _, file := range rendered.RenderBundle.Files {
		if file.Role == renderBundleRoleLocalTheme {
			themes++
		}
		if file.Path == "slides.md" && strings.Contains(string(file.Bytes), "](theme/") {
			t.Fatal("theme member emitted as image content")
		}
		if file.Role == renderBundleRoleToolchainSnapshot {
			var snapshot ToolchainSnapshot
			if err := json.Unmarshal(file.Bytes, &snapshot); err != nil {
				t.Fatal(err)
			}
			if len(templateFontFiles(rendered.RenderBundle.Files)) != 2 || validateToolchainFonts(snapshot, rendered.RenderBundle.Files) != nil {
				t.Fatal("font inventory omitted template configuration or font")
			}
			for i, font := range snapshot.FontFiles {
				if font.Path != "template/theme/style.css" {
					continue
				}
				altered := snapshot
				altered.FontFiles = append(append([]ToolchainFontFile{}, snapshot.FontFiles[:i]...), snapshot.FontFiles[i+1:]...)
				raw, err := canonicalJSON(altered.FontFiles)
				if err != nil {
					t.Fatal(err)
				}
				altered.FontManifestSHA256 = hashHex(raw)
				if validateToolchainFonts(altered, rendered.RenderBundle.Files) == nil {
					t.Fatal("rehashed incomplete usable font inventory accepted")
				}
			}
		}
	}
	if themes != 3 || len(rendered.Redaction.Model.Assets) != 5 {
		t.Fatal("incomplete retained theme inventory")
	}
}

func assertNativeTemplateThemes(t *testing.T, directory, temporaryRoot string) {
	t.Helper()
	renderer, err := NewInstalledRenderer(directory, temporaryRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{ReleaseScopeInternalReview, ReleaseScopeExternal} {
		input := fixtureTemplateThemeInput(t, scope)
		rendered, reason, err := renderer.Render(context.Background(), input)
		if err != nil {
			t.Fatal("native theme", scope, reason, err)
		}
		assertTemplateThemeOutput(t, input, rendered)
		if scope == ReleaseScopeExternal && rendered.RenderBundle.Manifest.ExternalDeterminismSHA256 == nil {
			t.Fatal("theme omitted external determinism evidence")
		}
	}
	input := fixtureTemplateThemeInput(t, ReleaseScopeInternalReview)
	font := reportingassets.ThemeFont()
	clear(font[48:]) // Header is admissible; the confined decoder must reject it.
	if validateThemeFont(font) != nil {
		t.Fatal("invalid-font fixture did not reach the decoder boundary")
	}
	replaceThemeFixtureFile(&input, "theme/fonts/body.woff2", font)
	rendered, _, err := renderer.Render(context.Background(), input)
	if err == nil || len(rendered.RenderBundle.Files) != 0 {
		t.Fatal("failed declared font silently used fallback or published artifacts")
	}
}
