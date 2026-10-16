package reporting

import "sort"

func templateAssetDeclaration(asset TemplateAsset) ReportingAssetDeclaration {
	return ReportingAssetDeclaration{
		SchemaID: "cartulary.reporting_asset_declaration.v1", AssetID: asset.AssetID,
		AssetKind: asset.AssetRole, BundlePath: asset.BundlePath, MediaType: asset.MediaType,
		SHA256: asset.SHA256, ByteSize: asset.ByteSize, DeclaredBy: "template_pack",
		RequiredForRelease: asset.RequiredForRelease, DisclosurePartitionRefs: []string{"public"},
	}
}

func addTemplateAssetDecisionPaths(contract TemplateContract, paths map[string]string) error {
	for _, asset := range contract.Assets {
		path, err := keyedModelPath("$", "assets", asset.AssetID)
		if err != nil {
			return err
		}
		if _, exists := paths[path]; exists {
			return newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
		}
		paths[path] = path
	}
	return nil
}

// Template pack bytes are static public presentation, never case-derived
// images or evidence downloads. Rules select the whole typed asset; scalar
// transformations cannot safely preserve a binary artifact's declaration.
func selectTemplateAssets(contract *TemplateContract, paths map[string]string, profile RedactionProfileView, scope string, effective []string) (map[string]ReportingAssetDeclaration, []selectedRedactionField, error) {
	declarations := map[string]ReportingAssetDeclaration{}
	if contract == nil {
		return declarations, nil, nil
	}
	if err := addTemplateAssetDecisionPaths(*contract, paths); err != nil {
		return nil, nil, err
	}
	candidates := []ExportField{}
	for _, asset := range contract.Assets {
		path, _ := keyedModelPath("$", "assets", asset.AssetID)
		selected := asset.AssetRole == "local_theme"
		for _, section := range contract.Sections {
			selected = selected || templateBindingsSelect(section.Bindings, []string{path})
		}
		if !selected {
			continue
		}
		field := ExportField{Path: path, ContentClass: ContentClassTemplateBoilerplate, Value: asset.AssetID, DisclosurePartitionRefs: []string{"public"}}
		_, action, _ := selectRedactionRule(field, path, profile)
		if action.Type != ActionAllow && (action.Type != ActionDrop || asset.AssetRole == "local_theme") || action.TokenBacked {
			return nil, nil, newRenderValidationError("redaction_manifest_invalid", "redaction_action_unresolved")
		}
		declarations[path] = templateAssetDeclaration(asset)
		candidates = append(candidates, field)
	}
	selected, err := selectRedactionFields(candidates, paths, profile, scope, effective)
	return declarations, selected, err
}

func templateAssetBlocks(decl TemplateSection, model RedactedExportModel) []ReportingBlock {
	blocks := []ReportingBlock{}
	for _, asset := range model.Assets {
		if asset.AssetKind == "local_theme" {
			continue
		}
		path, _ := keyedModelPath("$", "assets", asset.AssetID)
		if !templateBindingsSelect(decl.Bindings, []string{path}) {
			continue
		}
		refs := []string{}
		blocks = append(blocks, ReportingBlock{SchemaID: ReportingBlockSchemaID, BlockKind: "asset_ref", ContentClass: ContentClassTemplateBoilerplate,
			Fields:   []ReportingField{{SchemaID: "cartulary.reporting_field.v1", FieldKey: "asset_id", FieldOrdinal: 1, SourceValueState: "derived", RedactedValueState: "unchanged", Value: asset.AssetID, SourceRefs: refs, SupportRefs: refs, DisclosurePartitionRefs: []string{"public"}}},
			Children: []ReportingBlock{}, SourceRefs: refs, SupportRefs: refs, DisclosurePartitionRefs: []string{"public"}})
	}
	return blocks
}

func sortAssetDeclarations(assets []ReportingAssetDeclaration) {
	sort.Slice(assets, func(i, j int) bool {
		return assets[i].BundlePath < assets[j].BundlePath || assets[i].BundlePath == assets[j].BundlePath && assets[i].AssetID < assets[j].AssetID
	})
}
