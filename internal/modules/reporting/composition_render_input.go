package reporting

import (
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
)

func decodeCompositionRenderInput(raw json.RawMessage) (reportcomposition.RenderSource, error) {
	source, err := reportcomposition.DecodeRenderSource(raw)
	if err == nil {
		return source, nil
	}
	var invalid *reportcomposition.RenderSourceError
	reason := "invalid_scalar_value"
	if errors.As(err, &invalid) {
		switch invalid.ReasonCode {
		case "composition_digest_mismatch", "composition_template_mismatch", "authored_title_limit_exceeded", "authored_text_limit_exceeded", "diagram_label_override_invalid", "diagram_layout_invalid", "diagram_layout_duplicate_target", "diagram_layout_unknown_target":
			reason = invalid.ReasonCode
		case "composition_schema_invalid", "composition_source_invalid":
			reason = "invalid_schema_id"
		}
	}
	return reportcomposition.RenderSource{}, newRenderValidationError("composition_invalid", reason)
}

func validateCompositionRenderBinding(raw json.RawMessage, contract TemplateContract, model RedactedExportModel, scope string, profile RedactionProfileView) error {
	if len(raw) == 0 {
		return nil
	}
	source, err := decodeCompositionRenderInput(raw)
	if err != nil {
		return err
	}
	if source.TemplateID != contract.TemplateID || source.TemplateVersion != contract.TemplateVersion {
		return newRenderValidationError("composition_invalid", "composition_template_mismatch")
	}
	if source.IncidentID != model.IncidentID {
		return newRenderValidationError("composition_invalid", "composition_digest_mismatch")
	}
	preview := model.PreviewAttemptID != nil
	if (source.SchemaID == reportcomposition.PreviewSourceSchemaID) != preview || (preview && scope != ReleaseScopeInternalDraft) {
		return newRenderValidationError("composition_invalid", "invalid_schema_id")
	}
	if scope == ReleaseScopeExternal {
		var ops []compositionDeckOp
		if err := json.Unmarshal(source.DeckOps, &ops); err != nil {
			return compositionFailure("invalid_schema_id")
		}
		for _, op := range ops {
			if op.OnUnresolved == "drop" {
				return compositionFailure("composition_drop_invalid_for_external_release")
			}
		}
	}
	if scope == ReleaseScopeExternal && !profile.AllowAuthoredPresentationText {
		var authored []json.RawMessage
		if err := json.Unmarshal(source.AuthoredTexts, &authored); err != nil {
			return newRenderValidationError("composition_invalid", "invalid_schema_id")
		}
		if len(authored) != 0 {
			return newRenderValidationError("composition_invalid", "authored_text_not_permitted")
		}
	}
	return nil
}
