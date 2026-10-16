package reporting

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

func templateBindingContains(allowed, requested string) bool {
	a, ok := parseModelPath(allowed, true)
	if !ok {
		return false
	}
	b, ok := parseModelPath(requested, true)
	if !ok || len(a) > len(b) {
		return false
	}
	for i, step := range a {
		if step.Key != b[i].Key || (!step.Wildcard && (b[i].Wildcard || step.Value != b[i].Value)) {
			return false
		}
	}
	return true
}

func validateTemplateContract(c TemplateContract, kind, scope string) error {
	invalid := func() error { return newRenderValidationError("bundle_manifest_invalid", "invalid_scalar_value") }
	if _, err := resolveRenderLimits(c.DeclaredLimits); err != nil {
		return err
	}
	if c.SchemaID != templateManifestSchema || c.ManifestVersion != "3" || !validReportingIdentifier(c.TemplateID) || !validReportingIdentifier(c.TemplateVersion) || c.TemplateVersion == "latest" ||
		!validReportingText(c.DeckTitle, false) || utf8.RuneCountInString(c.DeckTitle) > 120 || !validReportingText(c.Author, false) || utf8.RuneCountInString(c.Author) > 120 ||
		strings.ContainsAny(c.DeckTitle+c.Author, "{}") || c.Theme != "default" || !c.SupportsOutputKind(kind) || !c.SupportsReleaseScope(scope) {
		return invalid()
	}
	uniqueTokens := func(items, allowed []string, required bool) bool {
		if items == nil || required && len(items) == 0 {
			return false
		}
		seen := map[string]bool{}
		for _, item := range items {
			if seen[item] || !slices.Contains(allowed, item) {
				return false
			}
			seen[item] = true
		}
		return true
	}
	if !uniqueTokens(c.SupportedOutputKinds, supportedOutputKinds(), true) || !uniqueTokens(c.SupportedReleaseScopes, supportedReleaseScopes(), true) ||
		c.AllowedExportModelBindings == nil || c.Sections == nil || len(c.Sections) == 0 || c.Layouts == nil || len(c.Layouts) == 0 ||
		c.NarrativeSlots == nil || c.Assets == nil || c.RenderProfiles == nil || len(c.RenderProfiles) == 0 || c.SupportedOutputOptions == nil || c.ExtensionFeatures == nil {
		return invalid()
	}
	for i, binding := range c.AllowedExportModelBindings {
		if _, ok := parseModelPath(binding, true); !ok || i > 0 && c.AllowedExportModelBindings[i-1] >= binding {
			return fmt.Errorf("%w: invalid allowed model address", ErrUndeclaredTemplateBinding)
		}
	}
	allowedBinding := func(binding string) bool {
		for _, allowed := range c.AllowedExportModelBindings {
			if templateBindingContains(allowed, binding) {
				return true
			}
		}
		return false
	}
	blockKinds := supportedReportingBlockKinds()
	layouts := map[string]bool{}
	for _, layout := range c.Layouts {
		if !validReportingIdentifier(layout.LayoutID) || layouts[layout.LayoutID] || !slices.Contains([]string{"cover", "default", "section"}, layout.SlidevLayout) || !uniqueTokens(layout.AllowedBlockKinds, blockKinds, true) {
			return invalid()
		}
		layouts[layout.LayoutID] = true
	}
	sections, diagrams := map[string]bool{}, map[string]bool{}
	for _, section := range c.Sections {
		if !reportingMemberName.MatchString(section.DeclID) || sections[section.DeclID] || !slices.Contains([]string{"timeline", "entity_summary", "evidence_summary", "relationship_summary", "diagram", "narrative", "appendix", "validation"}, section.SectionKind) ||
			!slices.Contains([]string{"none", "per_party", "per_subject"}, section.Expansion) || !validReportingText(section.Title, false) ||
			!slices.Contains([]string{"none", "reveal_blocks", "reveal_list_items"}, section.ClickProfile) || section.Bindings == nil || section.DiagramDecls == nil {
			return invalid()
		}
		literal := strings.ReplaceAll(section.Title, "{expansion_label}", "")
		if strings.ContainsAny(literal, "{}") || section.Expansion == "none" && literal != section.Title {
			return invalid()
		}
		if section.SlideLayoutID != nil && !layouts[*section.SlideLayoutID] || section.SlideLayoutID == nil && (!layouts["cover"] || !layouts["default"]) {
			return invalid()
		}
		if section.SelectionRuleID != nil && (section.SectionKind != "timeline" || !validReportingIdentifier(*section.SelectionRuleID)) {
			return invalid()
		}
		sections[section.DeclID] = true
		seenBindings := map[string]bool{}
		for _, binding := range section.Bindings {
			if !allowedBinding(binding) || seenBindings[binding] || !supportedTemplateFieldBinding(binding) {
				return fmt.Errorf("%w: undeclared model address", ErrUndeclaredTemplateBinding)
			}
			seenBindings[binding] = true
			steps, _ := parseModelPath(binding, true)
			if len(steps) == 2 && steps[0].Value == "assets" && !steps[1].Wildcard {
				found := false
				for _, asset := range c.Assets {
					if asset.AssetID == steps[1].Value && asset.AssetRole == "local_theme" {
						return invalid()
					}
					found = found || asset.AssetID == steps[1].Value
				}
				if !found {
					return newRenderValidationError("undeclared_asset", "undeclared_asset_ref")
				}
			}
		}
		for _, d := range section.DiagramDecls {
			if d.SchemaID != "template_diagram_decl.v1" || !validReportingIdentifier(d.DeclID) || diagrams[d.DeclID] || !slices.Contains([]string{"flowchart", "sequence"}, d.DiagramKind) || !slices.Contains([]string{"graph", "timeline"}, d.DiagramSourceKind) || (d.DiagramSourceKind == "graph") != (d.SourceGraphViewID != nil) || d.SelectionRule.SchemaID != "diagram_selection_rule.v1" {
				return invalid()
			}
			if d.SourceGraphViewID != nil && !validReportingIdentifier(*d.SourceGraphViewID) {
				return invalid()
			}
			diagrams[d.DeclID] = true
		}
	}
	slots := map[string]bool{}
	for _, slot := range c.NarrativeSlots {
		if !validReportingIdentifier(slot.SlotID) || slots[slot.SlotID] || !sections[slot.SectionDeclID] || slot.MaxChars <= 0 || !allowedBinding(slot.Binding) {
			return invalid()
		}
		steps, ok := parseModelPath(slot.Binding, true)
		if !ok || len(steps) != 4 || steps[0].Key || steps[0].Value != "records" || !steps[1].Key || steps[2].Key || steps[2].Value != "fields" || !steps[3].Key {
			return invalid()
		}
		slots[slot.SlotID] = true
	}
	if err := validateTemplateAssetDeclarations(c); err != nil {
		return err
	}
	profiles := map[string]bool{}
	for _, p := range c.RenderProfiles {
		if !validReportingIdentifier(p.RenderEnvironmentProfileID) || profiles[p.RenderEnvironmentProfileID] || p.ViewportCSSPX != "1280x720" || p.DeviceScaleFactor != 1 || p.Locale != "C.UTF-8" || !slices.Contains([]string{"light", "dark"}, p.ColorScheme) {
			return invalid()
		}
		profiles[p.RenderEnvironmentProfileID] = true
	}
	if len(c.SupportedOutputOptions) != len(c.SupportedOutputKinds) {
		return invalid()
	}
	for i, o := range c.SupportedOutputOptions {
		if o.SchemaID != "template_output_option_support.v1" || !c.SupportsOutputKind(o.OutputKind) || i > 0 && c.SupportedOutputOptions[i-1].OutputKind >= o.OutputKind || o.PNG || o.PPTX || o.OutputKind == OutputKindMermaid && o.PDF {
			return invalid()
		}
	}
	features := map[string]bool{}
	for _, f := range c.ExtensionFeatures {
		if !validReportingIdentifier(f.FeatureID) || features[f.FeatureID] || (f.PayloadPath == nil) != (f.PayloadSHA256 == nil) || f.PayloadSHA256 != nil && !sha256HexPattern.MatchString(*f.PayloadSHA256) || f.Required {
			return invalid()
		}
		features[f.FeatureID] = true
	}
	return nil
}

func validateTemplateOutputOptions(c TemplateContract, options renderOutputOptions, kind string) error {
	for _, support := range c.SupportedOutputOptions {
		if support.OutputKind != kind {
			continue
		}
		if options.PDF && !support.PDF || options.SVG && !support.SVG || options.PNG && !support.PNG || options.PPTX && !support.PPTX || options.RenderedDiagrams && !support.RenderedDiagrams {
			return newRenderValidationError("bundle_manifest_invalid", "invalid_scalar_value")
		}
		return nil
	}
	return newRenderValidationError("bundle_manifest_invalid", "required_member_missing")
}
