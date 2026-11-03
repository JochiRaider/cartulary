package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
)

const (
	maxRenderModelBytes = 67_108_864
	maxRenderModelItems = 1_048_576
)

// Admission proves the original frozen identity before returning a detached
// model. Rendering cannot substitute a newly hashed model for the retained input.
func admitFrozenExportModel(ctx context.Context, model ExportModel, digest string) (ExportModel, error) {
	if err := ctx.Err(); err != nil {
		return ExportModel{}, err
	}
	if renderModelItemCount(model) > maxRenderModelItems {
		return ExportModel{}, errors.New("reporting: frozen export model item limit")
	}
	encoded, err := canonicalJSON(model)
	if err != nil {
		return ExportModel{}, err
	}
	if len(encoded) > maxRenderModelBytes {
		return ExportModel{}, errors.New("reporting: frozen export model byte limit")
	}
	if reportingObjectDigest(ExportModelSchemaID, encoded) != digest {
		return ExportModel{}, errors.New("reporting: frozen export model digest invalid")
	}
	if err := ctx.Err(); err != nil {
		return ExportModel{}, err
	}
	if validateRenderModelIdentity(model) != nil {
		return ExportModel{}, errors.New("reporting: invalid frozen export model")
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	var admitted ExportModel
	if err := decoder.Decode(&admitted); err != nil {
		return ExportModel{}, err
	}
	canonical, err := canonicalJSON(admitted)
	if err != nil || !bytes.Equal(encoded, canonical) {
		return ExportModel{}, errors.New("reporting: frozen model changed during admission")
	}
	if _, err := admitted.RedactionFields(); err != nil {
		return ExportModel{}, err
	}
	if err := ctx.Err(); err != nil {
		return ExportModel{}, err
	}
	return admitted, nil
}

func renderModelItemCount(model ExportModel) int {
	count := len(model.Sections) + len(model.Records) + len(model.Relationships) +
		len(model.TimelineEvents) + len(model.Subjects) + len(model.Diagrams) +
		len(model.Assets) + len(model.SupportIndex) + len(model.SourceIndex) + len(model.PresentationTexts)
	var visit func([]ReportingBlock)
	visit = func(blocks []ReportingBlock) {
		count += len(blocks)
		for _, block := range blocks {
			visit(block.Children)
		}
	}
	for _, section := range model.Sections {
		visit(section.Blocks)
	}
	return count
}
