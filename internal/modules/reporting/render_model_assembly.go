package reporting

import "context"

// Only semantic identities are kept outside the canonical value. Case content
// lives in ExportModel; neither serializers nor replay reselect snapshot data.
type renderSectionOrigin struct {
	declarationIndex int
	expansionKey     *string
	titleRefs        reportingContainerRefs
}

type renderModelAssembly struct {
	templateSHA256 string
	templateFiles  []TemplatePackFile
	modelSHA256    string
	sections       map[string]renderSectionOrigin
}

func assembleRenderModel(ctx context.Context, input renderInput, result RedactionResult, graphs map[string]selectedGraphDiagram) (RedactionResult, error) {
	if err := ctx.Err(); err != nil {
		return RedactionResult{}, err
	}
	contract, model := input.TemplateContract, result.Model
	if err := validateTemplatePackFiles(contract, input.TemplateFiles); err != nil {
		return RedactionResult{}, err
	}
	limits, err := resolveRenderLimits(contract.DeclaredLimits)
	if err != nil {
		return RedactionResult{}, err
	}
	instances, err := deriveTemplateSections(ctx, contract, model)
	if err != nil {
		return RedactionResult{}, err
	}
	diagrams, err := deriveDiagramModels(contract, model, input.CompositionJSON, input.ReleaseScope, input.OutputKind, graphs)
	if err != nil {
		return RedactionResult{}, err
	}
	for _, bound := range []struct {
		key   string
		count int
	}{
		{"diagrams.count", len(diagrams)}, {"records.count", len(model.Records)},
		{"relationships.count", len(model.Relationships)}, {"subjects.count", len(model.Subjects)},
		{"support_refs.count", len(model.SupportIndex)},
	} {
		if err := limits.check(bound.key, int64(bound.count)); err != nil {
			return RedactionResult{}, err
		}
	}
	templateBytes, err := canonicalJSON(templateManifestView(contract))
	if err != nil {
		return RedactionResult{}, err
	}
	pack, err := cloneTemplatePack(TemplatePack{Manifest: contract, Files: input.TemplateFiles})
	if err != nil {
		return RedactionResult{}, err
	}
	assembly := &renderModelAssembly{templateSHA256: reportingObjectDigest(templateManifestSchema, templateBytes), templateFiles: pack.Files, sections: map[string]renderSectionOrigin{}}
	model.Sections = []ReportingSection{}
	model.Diagrams = diagrams
	byID := map[string]reportingDiagramModel{}
	for _, diagram := range diagrams {
		byID[diagram.DiagramID] = diagram
	}
	declarations := map[string]int{}
	for index, decl := range contract.Sections {
		declarations[decl.DeclID] = index
	}
	budget := templateExpansionBudget{limits: limits}
	titles := map[string]reportingContainerRefs{}
	for _, instance := range instances {
		if err := ctx.Err(); err != nil {
			return RedactionResult{}, err
		}
		section, decl := instance.Section, instance.Declaration
		for _, declared := range decl.DiagramDecls {
			diagram, exists := byID[declared.DeclID]
			if exists {
				section.Blocks = append(section.Blocks, diagramReferenceBlock(diagram))
			} else if declared.Required {
				return RedactionResult{}, compositionFailure("composition_anchor_unresolved")
			}
		}
		if len(section.Blocks) == 0 && section.SectionKind != "validation" {
			if decl.Required {
				return RedactionResult{}, newRenderValidationError("export_model_invalid", "required_section_empty")
			}
			continue
		}
		rebuildCompositionBlocks(section.Blocks, section.SectionID, false)
		refs := emptyContainerRefs()
		refs.include(instance.TitleRefs.source, instance.TitleRefs.support, instance.TitleRefs.partitions)
		for _, block := range section.Blocks {
			refs.include(block.SourceRefs, block.SupportRefs, block.DisclosurePartitionRefs)
		}
		refs = refs.normalized()
		section.SourceRefs, section.SupportRefs, section.DisclosurePartitionRefs = refs.source, refs.support, refs.partitions
		section.ContentClassSummary, err = summarizeReportingBlocks(section.Blocks)
		if err != nil {
			return RedactionResult{}, err
		}
		section.SectionValidation = ReportingSectionValidation{Result: "passed"}
		if err := budget.include(section); err != nil {
			return RedactionResult{}, err
		}
		model.Sections = append(model.Sections, section)
		assembly.sections[section.SectionID] = renderSectionOrigin{declarations[decl.DeclID], cloneStringPtr(instance.ExpansionKey), instance.TitleRefs}
		titles[section.SectionID] = instance.TitleRefs
	}
	if err := validateSectionStructure(model.SnapshotContent, titles); err != nil {
		return RedactionResult{}, err
	}
	if err := validateSnapshotFieldValues(model.SnapshotContent); err != nil {
		return RedactionResult{}, err
	}
	encoded, err := canonicalJSON(model)
	if err != nil {
		return RedactionResult{}, err
	}
	if err := limits.check("export_model.canonical_bytes", int64(len(encoded))); err != nil {
		return RedactionResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return RedactionResult{}, err
	}
	assembly.modelSHA256 = reportingObjectDigest(model.SchemaID, encoded)
	model.assembly = assembly
	result.Model, result.ModelJSON, result.ModelSHA256 = model, encoded, assembly.modelSHA256
	return result, nil
}

// Each consumer receives its own copy so composition/chunking cannot mutate
// canonical content. The digest also rejects an accidentally altered result.
func readAssembledRenderModel(contract TemplateContract, model RedactedExportModel) (RedactedExportModel, error) {
	invalid := newRenderValidationError("export_model_invalid", "invalid_scalar_value")
	if model.assembly == nil {
		return RedactedExportModel{}, invalid
	}
	if err := validateTemplatePackFiles(contract, model.assembly.templateFiles); err != nil {
		return RedactedExportModel{}, err
	}
	templateBytes, err := canonicalJSON(templateManifestView(contract))
	if err != nil || reportingObjectDigest("cartulary.reporting_template_pack_manifest.v3", templateBytes) != model.assembly.templateSHA256 {
		return RedactedExportModel{}, invalid
	}
	encoded, err := canonicalJSON(model)
	if err != nil || reportingObjectDigest(model.SchemaID, encoded) != model.assembly.modelSHA256 {
		return RedactedExportModel{}, invalid
	}
	copy, err := copyReportingModel(model.ExportModel)
	if err != nil {
		return RedactedExportModel{}, err
	}
	for index := range copy.Diagrams {
		copy.Diagrams[index].compositionOwned = model.Diagrams[index].compositionOwned
	}
	model.ExportModel = copy
	return model, nil
}
