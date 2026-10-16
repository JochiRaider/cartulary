package reporting

import (
	"encoding/json"
	"reflect"
	"testing"
)

func testTemplateManifestContract(t *testing.T) {
	t.Run("declared local asset pack", testTemplatePackAssets)
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok || validateTemplateContract(contract, OutputKindSlidev, ReleaseScopeInternalDraft) != nil {
		t.Fatal("built-in manifest rejected")
	}
	raw, err := canonicalJSON(templateManifestView(contract))
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	keys := []string{"schema_id", "template_id", "template_version", "manifest_version", "deck_title", "author", "theme", "supported_output_kinds", "supported_release_scopes", "allowed_export_model_bindings", "sections", "layouts", "narrative_slots", "assets", "render_profiles", "supported_output_options", "declared_limits", "extension_features"}
	if len(object) != len(keys) {
		t.Fatal("manifest contains a private field or omits a public member")
	}
	for _, key := range keys {
		if object[key] == nil || string(object[key]) == "null" {
			t.Fatal("missing materialized manifest member", key)
		}
	}
	for _, version := range []string{"1", "2", "3", "latest"} {
		if _, ok := ResolveTemplateContract(DefaultTemplateID, version); ok {
			t.Fatal("retired template admitted", version)
		}
	}
	for _, value := range []string{"null", "[]", `[{"schema_id":"aggregate_category_allowlist.v1","allowlist_id":"unsafe","field_key":"category","labels":["public"]}]`} {
		var candidate map[string]json.RawMessage
		if err := json.Unmarshal(raw, &candidate); err != nil {
			t.Fatal(err)
		}
		candidate["aggregate_category_allowlists"] = json.RawMessage(value)
		encoded, err := canonicalJSON(candidate)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := decodePortableObject[TemplateContract](encoded, reportingObjectDigest(templateManifestSchema, encoded)); err == nil {
			t.Fatal("retired aggregate allowlist accepted", value)
		}
	}
	for name, mutate := range map[string]func(*TemplateContract){
		"schema":                 func(c *TemplateContract) { c.SchemaID = "retired" },
		"bindings":               func(c *TemplateContract) { c.AllowedExportModelBindings = nil },
		"numeric address":        func(c *TemplateContract) { c.AllowedExportModelBindings = []string{"$.sections[0]"} },
		"two wildcards":          func(c *TemplateContract) { c.AllowedExportModelBindings = []string{"$.sections[*].blocks[*]"} },
		"undeclared binding":     func(c *TemplateContract) { c.Sections[0].Bindings = []string{"$.records[*]"} },
		"duplicate section":      func(c *TemplateContract) { c.Sections = append(c.Sections, c.Sections[0]) },
		"missing layout":         func(c *TemplateContract) { c.Sections[0].SlideLayoutID = fixtureNote("missing") },
		"duplicate layout":       func(c *TemplateContract) { c.Layouts = append(c.Layouts, c.Layouts[0]) },
		"unknown block":          func(c *TemplateContract) { c.Layouts[0].AllowedBlockKinds = []string{"future"} },
		"unowned block spelling": func(c *TemplateContract) { c.Layouts[0].AllowedBlockKinds = []string{"overflow_notice"} },
		"unbound asset":          func(c *TemplateContract) { c.Assets = []TemplateAsset{{AssetID: "unbound"}} },
		"missing options":        func(c *TemplateContract) { c.SupportedOutputOptions = c.SupportedOutputOptions[:1] },
		"wrong options":          func(c *TemplateContract) { c.SupportedOutputOptions[0].PDF = true },
		"unsupported PNG":        func(c *TemplateContract) { c.SupportedOutputOptions[0].PNG = true },
		"unsupported PPTX":       func(c *TemplateContract) { c.SupportedOutputOptions[1].PPTX = true },
		"unknown expansion":      func(c *TemplateContract) { c.Sections[0].Expansion = "future" },
		"unknown clicks":         func(c *TemplateContract) { c.Sections[0].ClickProfile = "future" },
		"wrong geometry":         func(c *TemplateContract) { c.RenderProfiles[0].ViewportCSSPX = "1920x1080" },
		"required extension": func(c *TemplateContract) {
			c.ExtensionFeatures = []TemplateExtensionFeature{{FeatureID: "future.feature", Required: true}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			var candidate TemplateContract
			if json.Unmarshal(raw, &candidate) != nil {
				t.Fatal("fixture")
			}
			mutate(&candidate)
			if validateTemplateContract(candidate, OutputKindSlidev, ReleaseScopeInternalDraft) == nil {
				t.Fatal("invalid template admitted")
			}
		})
	}
	for _, test := range []struct {
		allowed, requested string
		want               bool
	}{
		{"$.sections[*]", "$.sections['sec-0001'].blocks['b'].fields['value']", true},
		{"$.sections['sec-0001']", "$.sections[*]", false},
		{"$.sections[*].blocks['only']", "$.sections['one'].blocks['other']", false},
		{"$.sections[*]", "$.records['sec-0001']", false},
		{"$.sections[*]", "$.sections[0]", false},
		{"$.sections[*]", "$.sections[*].blocks[*]", false},
		{`$.sections['a\'b']`, `$.sections['a\'b'].title`, true},
	} {
		if templateBindingContains(test.allowed, test.requested) != test.want {
			t.Fatal("binding subset differs", test)
		}
	}
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/incident/a", ContentClass: ContentClassSourceEvidence, Value: "removed"},
		{Path: "/incident/b", ContentClass: ContentClassSourceEvidence, Value: "selected"},
		{Path: "/incident/c", ContentClass: ContentClassSourceEvidence, Value: "unselected"},
	})
	paths := fixtureModelPaths(t, model.SnapshotContent, "/incident/a", "/incident/b", "/incident/c")
	profile, _, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	drop := paths["/incident/a"]
	profile.Rules = []RedactionRule{{RuleID: "drop-first", Path: &drop, Action: RedactionActionSpec{Type: ActionDrop}}}
	digest, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, "", ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	contract.Sections[0].Bindings = []string{paths["/incident/b"]}
	contract.Sections[0].Title = "Declared title"
	contract.Sections[0].SlideLayoutID = fixtureNote("section")
	before, _ := canonicalJSON(result.Model.Sections)
	deck, err := fixtureDerivedDeck(contract, result.Model, renderOutputOptions{}, nil)
	if err != nil || len(deck.Slides) != 1 || deck.Slides[0].Title != "Declared title" || deck.Slides[0].Layout != "section" || len(deck.Slides[0].Blocks) != 1 || deck.Slides[0].Blocks[0].Fields[0].Value != "selected" {
		t.Fatal("template selection changed target after redaction renumbering", err)
	}
	after, _ := canonicalJSON(result.Model.Sections)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("template derivation mutated retained input")
	}
	contract.Sections[0].Bindings = []string{drop}
	if _, err := fixtureDerivedDeck(contract, result.Model, renderOutputOptions{}, nil); err == nil {
		t.Fatal("required dropped binding was refilled from source")
	}
}
