package reporting

import (
	"context"
	"fmt"
	"strings"
)

type templateSectionInstance struct {
	Declaration  TemplateSection
	Section      ReportingSection
	ExpansionKey *string
	TitleRefs    reportingContainerRefs
}

// Selection traverses the typed retained tree. It never evaluates JSONPath or
// opens a source path, and it never fills a removed field from snapshot data.
func deriveTemplateSections(ctx context.Context, contract TemplateContract, model RedactedExportModel) ([]templateSectionInstance, error) {
	source, admittedPaths := model.Sections, model.admittedFieldPaths
	bindings, err := templateFieldBindings(model.SnapshotContent, admittedPaths)
	if err != nil {
		return nil, err
	}
	instances := []templateSectionInstance{}
	limits, err := resolveRenderLimits(contract.DeclaredLimits)
	if err != nil {
		return nil, err
	}
	budget := templateExpansionBudget{limits: limits}
	for index, decl := range contract.Sections {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		dimensions, err := templateSectionDimensions(decl, model)
		if err != nil {
			return nil, err
		}
		if len(dimensions) == 0 && decl.Required {
			return nil, newRenderValidationError("export_model_invalid", "required_section_empty")
		}
		for _, binding := range decl.Bindings {
			if !supportedTemplateFieldBinding(binding) {
				return nil, fmt.Errorf("%w: unsupported source binding", ErrUndeclaredTemplateBinding)
			}
		}
		for dimensionIndex, dimension := range dimensions {
			if !strings.Contains(decl.Title, "{expansion_label}") {
				dimension.labelRefs = emptyContainerRefs()
			}
			selected := map[string]ReportingField{}
			var visit func(string, string, []ReportingBlock) error
			visit = func(parent, member string, blocks []ReportingBlock) error {
				for _, block := range blocks {
					if err := ctx.Err(); err != nil {
						return err
					}
					path, err := keyedModelPath(parent, member, block.BlockID)
					if err != nil {
						return err
					}
					for _, field := range block.Fields {
						if !dimension.includes(field) {
							continue
						}
						if len(field.SourceRefs) == 0 {
							return newRenderValidationError("export_model_invalid", "dangling_source_ref")
						}
						if len(bindings[field.SourceRefs[0]]) == 0 {
							return newRenderValidationError("export_model_invalid", "dangling_source_ref")
						}
						if model.ReleaseScope == ReleaseScopeExternal && !decl.IncludeSuperseded && model.lifecycleRestrictions[field.SourceRefs[0]] == lifecycleSuperseded {
							continue
						}
						if templateBindingsSelect(decl.Bindings, bindings[field.SourceRefs[0]]) {
							selected[field.SourceRefs[0]] = field
						}
					}
					if err := visit(path, "children", block.Children); err != nil {
						return err
					}
				}
				return nil
			}
			for _, section := range source {
				path, err := keyedModelPath("$", "sections", section.SectionID)
				if err != nil {
					return nil, err
				}
				if err := visit(path, "blocks", section.Blocks); err != nil {
					return nil, err
				}
			}
			projected, err := projectRetainedSections(ctx, source, selected)
			if err != nil {
				return nil, err
			}
			ordering := fmt.Sprintf("%04d", index+1)
			if dimension.key != nil {
				ordering += fmt.Sprintf(".%04d", dimensionIndex+1)
			}
			section := ReportingSection{SchemaID: ReportingSectionSchemaID, SectionID: "sec-" + ordering, OrderingKey: ordering, SectionKind: decl.SectionKind, Title: strings.ReplaceAll(decl.Title, "{expansion_label}", dimension.label), Blocks: []ReportingBlock{}}
			for _, p := range projected {
				section.Blocks = append(section.Blocks, p.Blocks...)
			}
			narratives, err := deriveNarrativeBlocks(contract, decl, dimension, model)
			if err != nil {
				return nil, err
			}
			section.Blocks = append(section.Blocks, narratives...)
			for _, block := range narratives {
				for _, field := range block.Fields {
					selected[field.SourceRefs[0]] = field
				}
			}
			assetBlocks := templateAssetBlocks(decl, model)
			if len(section.Blocks) == 0 && len(assetBlocks) == 0 && decl.SectionKind != "validation" && len(decl.DiagramDecls) == 0 {
				if decl.Required {
					return nil, newRenderValidationError("export_model_invalid", "required_section_empty")
				}
				continue
			}
			// Rebuild selected case fields before appending typed presentation
			// references, which have no invented case-source identity.
			if len(section.Blocks) > 0 {
				projected, err = projectRetainedSections(ctx, []ReportingSection{section}, selected)
				if err != nil {
					return nil, err
				}
				if len(projected) != 1 {
					return nil, newRenderValidationError("export_model_invalid", "required_section_empty")
				}
				section = projected[0]
			}
			section.Blocks = append(section.Blocks, assetBlocks...)
			if err := budget.include(section); err != nil {
				return nil, err
			}
			instances = append(instances, templateSectionInstance{Declaration: decl, Section: section, ExpansionKey: dimension.key, TitleRefs: dimension.labelRefs.normalized()})
		}
	}
	return instances, nil
}
