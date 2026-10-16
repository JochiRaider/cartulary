package reporting

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"time"
)

func fixtureTemplateAssetPack(t *testing.T) TemplatePack {
	t.Helper()
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	contract.TemplateID, contract.TemplateVersion = "test.local-images", "1"
	contract.SupportedOutputKinds = []string{OutputKindSlidev}
	contract.SupportedOutputOptions = contract.SupportedOutputOptions[1:]
	contract.AllowedExportModelBindings = []string{"$.assets[*]", "$.sections[*]"}
	contract.Sections[0].Bindings = []string{"$.assets[*]", "$.sections[*]"}
	pixel := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	pixel.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, pixel); err != nil {
		t.Fatal(err)
	}
	files := []TemplatePackFile{
		{SourcePath: "images/pixel.png", Bytes: encoded.Bytes()},
		{SourcePath: "images/sign.svg", Bytes: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" viewBox="0 0 40 40"><rect x="0" y="0" width="40" height="40" fill="#238853"/></svg>`)},
	}
	contract.Assets = []TemplateAsset{
		{SchemaID: "template_asset_item.v1", AssetID: "image", AssetRole: "local_asset", BundlePath: "assets/image/pixel.png", SourcePath: files[0].SourcePath, MediaType: "image/png", ByteSize: int64(len(files[0].Bytes)), SHA256: hashHex(files[0].Bytes), RequiredForRelease: true},
		{SchemaID: "template_asset_item.v1", AssetID: "logo", AssetRole: "local_asset", BundlePath: "assets/logo/café.svg", SourcePath: files[1].SourcePath, MediaType: "image/svg+xml", ByteSize: int64(len(files[1].Bytes)), SHA256: hashHex(files[1].Bytes), RequiredForRelease: true},
	}
	return TemplatePack{Manifest: contract, Files: files}
}

func fixtureTemplateAssetInput(t *testing.T) renderInput {
	return fixtureTemplateAssetInputForScope(t, ReleaseScopeInternalReview)
}

func fixtureTemplateAssetInputForScope(t *testing.T, scope string) renderInput {
	t.Helper()
	pack := fixtureTemplateAssetPack(t)
	model, digest, err := buildStructuredTestExportModel("00000000-0000-0000-0000-000000000001", fixtureSnapshotID,
		time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC), SourceBoundaryTokenPrefix+strings.Repeat("0", 64), scope, nil,
		[]ExportField{{Path: "/incident/title", ContentClass: ContentClassDerivedAnalytic, Value: "Asset report", DisclosurePartitionRefs: []string{"public"}}})
	if err != nil {
		t.Fatal(err)
	}
	profile := InternalRedactionProfileID
	if scope == ReleaseScopeExternal {
		profile = ExternalRedactionProfileID
	}
	request := CreateReleaseRequest{OutputKind: OutputKindSlidev, OutputOptions: fixtureOutputOptions(OutputKindSlidev, scope), ReleaseScope: scope, RedactionProfileID: profile, RedactionProfileVersion: DefaultRedactionProfileVersion}
	input := renderCandidateFixture(request, pack.Manifest, model, digest)
	input.TemplateFiles = pack.Files
	return input
}

func testTemplatePackAssets(t *testing.T) {
	testTemplateThemes(t)
	pack := fixtureTemplateAssetPack(t)
	registry, err := NewTemplatePackRegistry(pack)
	if err != nil {
		t.Fatal(err)
	}
	pack.Files[0].Bytes[0] = 0
	pack.Manifest.Sections[0].Title = "changed"
	got, ok := registry.resolve("test.local-images", "1")
	if !ok || validateTemplatePackFiles(got.Manifest, got.Files) != nil || got.Manifest.Sections[0].Title == "changed" {
		t.Fatal("registry retained caller aliases")
	}
	got.Files[0].Bytes[0] = 0
	got, ok = registry.resolve("test.local-images", "1")
	if !ok || validateTemplatePackFiles(got.Manifest, got.Files) != nil {
		t.Fatal("resolved copy mutated registry")
	}
	if _, err := NewTemplatePackRegistry(got, got); err == nil {
		t.Fatal("duplicate template identity admitted")
	}
	for name, mutate := range map[string]func(*TemplatePack){
		"missing bytes": func(p *TemplatePack) { p.Files = p.Files[:1] },
		"extra bytes": func(p *TemplatePack) {
			p.Files = append(p.Files, TemplatePackFile{SourcePath: "undeclared", Bytes: []byte("secret")})
		},
		"wrong bytes":              func(p *TemplatePack) { p.Files[0].Bytes[0] = 0 },
		"media mismatch":           func(p *TemplatePack) { p.Manifest.Assets[0].MediaType = "image/jpeg" },
		"html":                     func(p *TemplatePack) { p.Manifest.Assets[0].MediaType = "text/html" },
		"size mismatch":            func(p *TemplatePack) { p.Manifest.Assets[0].ByteSize++ },
		"path escape":              func(p *TemplatePack) { p.Manifest.Assets[0].BundlePath = "assets/image/../x.png" },
		"source escape":            func(p *TemplatePack) { p.Manifest.Assets[0].SourcePath = "../image.png" },
		"non NFC":                  func(p *TemplatePack) { p.Manifest.Assets[1].BundlePath = "assets/logo/cafe\u0301.svg" },
		"duplicate id":             func(p *TemplatePack) { p.Manifest.Assets[1].AssetID = "image" },
		"duplicate source":         func(p *TemplatePack) { p.Manifest.Assets[1].SourcePath = p.Manifest.Assets[0].SourcePath },
		"optional referenced file": func(p *TemplatePack) { p.Manifest.Assets[0].RequiredForRelease = false },
		"operative limit":          func(p *TemplatePack) { p.Manifest.DeclaredLimits["local_assets.count"] = []byte("1") },
		"active SVG": func(p *TemplatePack) {
			p.Files[1].Bytes = []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)
			p.Manifest.Assets[1].SHA256 = hashHex(p.Files[1].Bytes)
			p.Manifest.Assets[1].ByteSize = int64(len(p.Files[1].Bytes))
		},
	} {
		t.Run("template pack/"+name, func(t *testing.T) {
			p := fixtureTemplateAssetPack(t)
			mutate(&p)
			if _, err := NewTemplatePackRegistry(p); err == nil {
				t.Fatal("invalid pack admitted")
			}
		})
	}
	input := fixtureTemplateAssetInput(t)
	digest, _ := templateManifestDigest(input.TemplateContract)
	if err := validateAdmittedTemplate(input, digest); err != nil {
		t.Fatal(err)
	}
	input.TemplateContract.DeckTitle = "altered"
	if validateAdmittedTemplate(input, digest) == nil {
		t.Fatal("self-consistent altered manifest admitted")
	}
	input = fixtureTemplateAssetInput(t)
	rendered, reason, err := fixtureRenderer().Render(context.Background(), input)
	if err != nil {
		t.Fatalf("local assets: %s: %v", reason, err)
	}
	assertTemplateAssetOutput(t, input, rendered)
	for _, action := range []string{ActionDrop, ActionMask, ActionStub, ActionTruncate} {
		t.Run("asset action/"+action, func(t *testing.T) {
			input := fixtureTemplateAssetInput(t)
			path := "$.assets['logo']"
			rule := RedactionActionSpec{Type: action}
			if action == ActionTruncate {
				n := 2
				rule.MaxChars = &n
			}
			input.RedactionProfile.Rules = append(input.RedactionProfile.Rules, RedactionRule{RuleID: "asset-rule", Path: &path, Action: rule})
			input.RedactionProfileSHA256, err = ValidateRedactionProfile(input.RedactionProfile)
			if err != nil {
				t.Fatal(err)
			}
			result, err := redactRenderInput(context.Background(), input)
			if action != ActionDrop {
				if err == nil {
					t.Fatal("scalar transformation admitted for binary asset")
				}
				return
			}
			if err != nil || len(result.Model.Assets) != 1 || result.Model.Assets[0].AssetID != "image" {
				t.Fatal("asset drop did not remove declaration", err)
			}
			files, err := retainedTemplateAssetFiles(input.TemplateContract, input.TemplateFiles, result.Model.Assets)
			if err != nil || len(files) != 1 {
				t.Fatal("asset drop did not remove bytes", err)
			}
		})
	}
}

func assertTemplateAssetOutput(t *testing.T, input renderInput, rendered RenderedRelease) {
	t.Helper()
	if err := validateRenderCandidateBinding(input, rendered); err != nil {
		t.Fatal(err)
	}
	if err := validateShareableBundle(rendered.RenderBundle); err != nil {
		t.Fatal(err)
	}
	if len(rendered.Redaction.Model.Assets) != 2 {
		t.Fatal("missing typed asset inventory")
	}
	assets, source := 0, ""
	for _, file := range rendered.RenderBundle.Files {
		if file.Role == renderBundleRoleLocalAsset {
			assets++
			if err := validateLocalImage(file.MediaType, file.Bytes); err != nil {
				t.Fatal(err)
			}
		}
		if file.Path == "slides.md" {
			source = string(file.Bytes)
		}
	}
	if assets != 2 || !strings.Contains(source, "![image](assets/image/pixel.png)") || !strings.Contains(source, "![logo](assets/logo/caf%C3%A9.svg)") {
		t.Fatal("declared local images not serialized and copied")
	}
	decisions := 0
	for _, entry := range rendered.Redaction.Manifest.Entries {
		if strings.HasPrefix(entry.Path, "$.assets[") {
			decisions++
			if entry.ContentClass != ContentClassTemplateBoilerplate || entry.Action != ActionAllow {
				t.Fatal("asset rule evidence differs")
			}
		}
	}
	if decisions != 2 {
		t.Fatal("asset decisions missing")
	}
}
