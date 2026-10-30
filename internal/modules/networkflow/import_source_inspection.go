package networkflow

import (
	"context"
	"encoding/json"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
)

func (f *importFacade) InspectImportUnitSource(ctx context.Context, request imports.ExtensionImportSourceRequest) (imports.ExtensionImportSourceResult, error) {
	if f == nil || f.sourceStore == nil {
		return imports.ExtensionImportSourceResult{}, importOwnerError("network_flow_target_unavailable", map[string]any{"reason_code": "owner_apply_contract_unavailable"})
	}
	if request.TargetKind != targetKindNetworkFlowTable || request.ExtensionProfileID != ProfileID {
		return imports.ExtensionImportSourceResult{}, importOwnerError("network_flow_mapping_invalid", map[string]any{"reason_code": "variant_member_conflict"})
	}
	stream, err := f.sourceStore.OpenSourceStream(ctx, request.SourceCapability.SourceStreamRef)
	if err != nil {
		return imports.ExtensionImportSourceResult{}, err
	}
	defer func() { _ = stream.Reader.Close() }()
	decoder, err := newCSVDecoder(stream.Reader, request.SourceCapability.SourceContentSHA256, f.limits, parseModePreview)
	if err != nil {
		return imports.ExtensionImportSourceResult{}, facadeError(err)
	}
	// Header inspection never calls Next: even the first data record is outside
	// its semantic scope. Successful finish still certifies the complete bytes.
	if err := decoder.finish(); err != io.EOF {
		return imports.ExtensionImportSourceResult{}, facadeError(err)
	}
	return imports.ExtensionImportSourceResult{
		OwnerResultSchemaID: importSourceInspectionResultSchemaID,
		OwnerResult: map[string]any{
			"schema_id":             importSourceInspectionResultSchemaID,
			"source_content_sha256": decoder.SourceSHA256,
			"source_columns":        decoder.Columns,
		},
	}, nil
}

func (f *importFacade) ValidateImportUnitSourceResult(result imports.ExtensionImportSourceResult) error {
	if result.OwnerResultSchemaID != importSourceInspectionResultSchemaID || result.OwnerResult == nil {
		return errInvalidPreview
	}
	raw, err := json.Marshal(result.OwnerResult)
	if err != nil {
		return errInvalidPreview
	}
	var payload struct {
		SchemaID            string          `json:"schema_id"`
		SourceContentSHA256 string          `json:"source_content_sha256"`
		SourceColumns       json.RawMessage `json:"source_columns"`
	}
	if _, err := decodePreviewObject(raw, "schema_id source_content_sha256 source_columns", "", "", &payload); err != nil {
		return err
	}
	if payload.SchemaID != importSourceInspectionResultSchemaID || !validLowerSHA256(payload.SourceContentSHA256) {
		return errInvalidPreview
	}
	_, err = decodePreviewColumns(payload.SourceColumns, f.limits, 0)
	return err
}
