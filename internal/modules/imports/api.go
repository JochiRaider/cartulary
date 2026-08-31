package imports

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	ProfileID = "import"

	assistantProfileWorkbookImport = "phase2_workbook_import_v1"
	parserProfileWorkbookImport    = "cartulary.import.phase2_workbook_import.v1"
	parserVersionWorkbookImport    = "phase11_import_adapter_v1"

	sourceFileKindCSV  = "csv"
	sourceFileKindXLSX = "xlsx"

	mediaTypeCSV         = "text/csv"
	mediaTypeXLSX        = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	mediaTypeOctetStream = "application/octet-stream"

	importDiscoveryJobHandlerName = "import.discovery_worker_v1"
	importApplyJobHandlerName     = "import.apply_worker_v1"
	DiscoveryJobKind              = "import.discovery_v1"
	ApplyJobKind                  = "import.apply_v1"

	ImportTargetKindViewSchema       = "view_schema"
	importTargetKindNetworkFlowTable = "network_flow_table"
	networkFlowExtensionProfileID    = "network_flow_activity"

	extensionMappingPreviewResultSchemaID = "cartulary.imports.extension_mapping_preview_result.v1"
)

var importSessionFileContentTypes = []string{
	mediaTypeCSV,
	mediaTypeXLSX,
	mediaTypeOctetStream,
}

type createSessionRequest struct {
	IncidentID       uuid.UUID
	ClientTxnID      string
	AssistantProfile string
}

type sourceColumnMapping struct {
	SourceColumnOrdinal int            `json:"source_column_ordinal"`
	SourceHeaderText    any            `json:"source_header_text"`
	FieldKey            *string        `json:"field_key"`
	EntityBindingMode   *string        `json:"entity_binding_mode"`
	TransformID         *string        `json:"transform_id"`
	TransformOptions    map[string]any `json:"transform_options"`
	EmptyValuePolicy    string         `json:"empty_value_policy"`
}

type approvedMapping struct {
	TargetKind           string                `json:"target_kind,omitempty"`
	TargetViewSchemaID   string                `json:"target_view_schema_id,omitempty"`
	ExtensionProfileID   string                `json:"extension_profile_id,omitempty"`
	OwnerMappingSchemaID string                `json:"owner_mapping_schema_id,omitempty"`
	OwnerMapping         json.RawMessage       `json:"owner_mapping,omitempty"`
	UnknownColumnPolicy  string                `json:"unknown_column_policy,omitempty"`
	SourceColumns        []sourceColumnMapping `json:"source_columns"`
}

func (mapping approvedMapping) targetKindOrDefault() string {
	if mapping.TargetKind == "" {
		return ImportTargetKindViewSchema
	}
	return mapping.TargetKind
}

type mappingRequest struct {
	ClientTxnID     string
	HeaderRowRef    int
	DataStartRowRef int
	approvedMapping approvedMapping
	Fingerprint     string
	Normalized      []byte
}

type mappingPreviewRequest struct {
	TargetKind           string
	ExtensionProfileID   string
	OwnerMappingSchemaID string
	OwnerMapping         json.RawMessage
}

type extensionMappingPreviewResource struct {
	SchemaID            string         `json:"schema_id"`
	ImportSessionID     string         `json:"import_session_id"`
	ImportUnitID        string         `json:"import_unit_id"`
	TargetKind          string         `json:"target_kind"`
	ExtensionProfileID  string         `json:"extension_profile_id"`
	OwnerResultSchemaID string         `json:"owner_result_schema_id"`
	OwnerResult         map[string]any `json:"owner_result"`
}

type actionRequest struct {
	ClientTxnID string
	Reason      *string
	Normalized  []byte
}

type applyRequest struct {
	ClientTxnID     string
	SelectedUnitIDs *[]uuid.UUID
	Normalized      []byte
}

type regionSourceRect struct {
	StartRow    int `json:"start_row"`
	StartColumn int `json:"start_column"`
	EndRow      int `json:"end_row"`
	EndColumn   int `json:"end_column"`
}

func (rect regionSourceRect) sourceRectangle() sourceRectangle {
	return sourceRectangle{
		left: rect.StartColumn, top: rect.StartRow, right: rect.EndColumn, bottom: rect.EndRow,
	}
}

type regionRequest struct {
	ClientTxnID string           `json:"client_txn_id"`
	SourceRect  regionSourceRect `json:"source_rect"`
	Normalized  []byte           `json:"-"`
}

func decodeCreateSessionMetadata(envelope httpapi.UploadEnvelope) (createSessionRequest, *httpapi.APIError) {
	allowed := map[string]struct{}{
		"incident_id":       {},
		"client_txn_id":     {},
		"assistant_profile": {},
	}
	for key := range envelope.Metadata {
		if _, ok := allowed[key]; !ok {
			return createSessionRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	var request createSessionRequest
	if value, ok := envelope.Metadata["incident_id"]; !ok {
		return createSessionRequest{}, invalidImportRequest("incident_id", "missing_required_field")
	} else {
		var raw string
		if err := json.Unmarshal(value, &raw); err != nil {
			return createSessionRequest{}, invalidImportRequest("incident_id", "invalid_value")
		}
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return createSessionRequest{}, invalidImportRequest("incident_id", "invalid_value")
		}
		request.IncidentID = parsed
	}
	if value, ok := envelope.Metadata["client_txn_id"]; !ok {
		return createSessionRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return createSessionRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	}
	request.AssistantProfile = assistantProfileWorkbookImport
	if value, ok := envelope.Metadata["assistant_profile"]; ok {
		if bytesEqualJSONNull(value) {
			return createSessionRequest{}, invalidImportRequest("assistant_profile", "field_not_nullable")
		}
		var profile string
		if err := json.Unmarshal(value, &profile); err != nil {
			return createSessionRequest{}, invalidImportRequest("assistant_profile", "invalid_value")
		}
		if profile != assistantProfileWorkbookImport {
			return createSessionRequest{}, invalidImportRequest("assistant_profile", "unsupported_assistant_profile")
		}
		request.AssistantProfile = profile
	}
	return request, nil
}

func decodeMappingRequest(reader io.Reader, discoveredColumns []map[string]any) (mappingRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return mappingRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id":           {},
		"target_kind":             {},
		"target_view_schema_id":   {},
		"extension_profile_id":    {},
		"owner_mapping_schema_id": {},
		"owner_mapping":           {},
		"header_row_ref":          {},
		"data_start_row_ref":      {},
		"unknown_column_policy":   {},
		"source_columns":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return mappingRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	request := mappingRequest{}
	if value, ok := raw["client_txn_id"]; !ok {
		return mappingRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return mappingRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	}
	mapping := approvedMapping{}
	if value, ok := raw["target_kind"]; ok {
		if bytesEqualJSONNull(value) {
			return mappingRequest{}, invalidImportRequest("target_kind", "field_not_nullable")
		}
		if err := json.Unmarshal(value, &mapping.TargetKind); err != nil || strings.TrimSpace(mapping.TargetKind) == "" {
			return mappingRequest{}, invalidImportRequest("target_kind", "invalid_value")
		}
	}
	extensionVariant := mapping.TargetKind != ""
	if extensionVariant {
		if _, ok := raw["target_view_schema_id"]; ok {
			return mappingRequest{}, invalidImportRequest("target_view_schema_id", "invalid_target_variant")
		}
		if _, ok := raw["unknown_column_policy"]; ok {
			return mappingRequest{}, invalidImportRequest("unknown_column_policy", "invalid_target_variant")
		}
		if value, ok := raw["extension_profile_id"]; !ok {
			return mappingRequest{}, invalidImportRequest("extension_profile_id", "missing_required_field")
		} else if err := json.Unmarshal(value, &mapping.ExtensionProfileID); err != nil || strings.TrimSpace(mapping.ExtensionProfileID) == "" {
			return mappingRequest{}, invalidImportRequest("extension_profile_id", "invalid_value")
		}
		if value, ok := raw["owner_mapping_schema_id"]; !ok {
			return mappingRequest{}, invalidImportRequest("owner_mapping_schema_id", "missing_required_field")
		} else if err := json.Unmarshal(value, &mapping.OwnerMappingSchemaID); err != nil || strings.TrimSpace(mapping.OwnerMappingSchemaID) == "" {
			return mappingRequest{}, invalidImportRequest("owner_mapping_schema_id", "invalid_value")
		}
		if value, ok := raw["owner_mapping"]; !ok {
			return mappingRequest{}, invalidImportRequest("owner_mapping", "missing_required_field")
		} else if bytesEqualJSONNull(value) || !json.Valid(value) {
			return mappingRequest{}, invalidImportRequest("owner_mapping", "invalid_value")
		} else {
			mapping.OwnerMapping = append(json.RawMessage(nil), value...)
		}
	} else {
		if _, ok := raw["extension_profile_id"]; ok {
			return mappingRequest{}, invalidImportRequest("extension_profile_id", "invalid_target_variant")
		}
		if _, ok := raw["owner_mapping_schema_id"]; ok {
			return mappingRequest{}, invalidImportRequest("owner_mapping_schema_id", "invalid_target_variant")
		}
		if _, ok := raw["owner_mapping"]; ok {
			return mappingRequest{}, invalidImportRequest("owner_mapping", "invalid_target_variant")
		}
		if value, ok := raw["target_view_schema_id"]; !ok {
			return mappingRequest{}, invalidImportRequest("target_view_schema_id", "missing_required_field")
		} else if err := json.Unmarshal(value, &mapping.TargetViewSchemaID); err != nil || strings.TrimSpace(mapping.TargetViewSchemaID) == "" {
			return mappingRequest{}, invalidImportRequest("target_view_schema_id", "invalid_value")
		}
		if value, ok := raw["unknown_column_policy"]; !ok {
			return mappingRequest{}, invalidImportRequest("unknown_column_policy", "missing_required_field")
		} else if err := json.Unmarshal(value, &mapping.UnknownColumnPolicy); err != nil || !validUnknownColumnPolicy(mapping.UnknownColumnPolicy) {
			return mappingRequest{}, invalidImportRequest("unknown_column_policy", "invalid_unknown_column_policy")
		}
	}
	if value, ok := raw["header_row_ref"]; !ok {
		return mappingRequest{}, invalidImportRequest("header_row_ref", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.HeaderRowRef); err != nil || request.HeaderRowRef < 1 {
		return mappingRequest{}, invalidImportRequest("header_row_ref", "invalid_row_reference")
	}
	if value, ok := raw["data_start_row_ref"]; !ok {
		return mappingRequest{}, invalidImportRequest("data_start_row_ref", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.DataStartRowRef); err != nil || request.DataStartRowRef < request.HeaderRowRef+1 {
		return mappingRequest{}, invalidImportRequest("data_start_row_ref", "invalid_row_reference")
	}
	var rawColumns []json.RawMessage
	if value, ok := raw["source_columns"]; !ok {
		return mappingRequest{}, invalidImportRequest("source_columns", "missing_required_field")
	} else if err := json.Unmarshal(value, &rawColumns); err != nil {
		return mappingRequest{}, invalidImportRequest("source_columns", "invalid_source_columns")
	}
	if len(rawColumns) != len(discoveredColumns) {
		return mappingRequest{}, invalidImportRequest("source_columns", "invalid_source_columns")
	}
	mapping.SourceColumns = make([]sourceColumnMapping, 0, len(rawColumns))
	seenTargets := map[string]struct{}{}
	for index, rawColumn := range rawColumns {
		column, apiErr := decodeSourceColumnMapping(rawColumn, extensionVariant)
		if apiErr != nil {
			return mappingRequest{}, apiErr
		}
		wantOrdinal := index + 1
		if column.SourceColumnOrdinal != wantOrdinal {
			return mappingRequest{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
		discovered := discoveredColumns[index]
		if !sourceHeaderEqual(column.SourceHeaderText, discovered["source_header_text"]) {
			return mappingRequest{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
		if column.FieldKey != nil {
			if _, ok := seenTargets[*column.FieldKey]; ok {
				return mappingRequest{}, invalidImportRequest("source_columns", "duplicate_target_field")
			}
			seenTargets[*column.FieldKey] = struct{}{}
		}
		mapping.SourceColumns = append(mapping.SourceColumns, column)
	}
	request.approvedMapping = mapping
	normalizedMapping := normalizedMappingPayload(request)
	normalized, err := json.Marshal(normalizedMapping)
	if err != nil {
		return mappingRequest{}, internalAPIError(err)
	}
	sum := sha256.Sum256(normalized)
	request.Fingerprint = hex.EncodeToString(sum[:])
	if err := rebuildMappingRequestNormalized(&request); err != nil {
		return mappingRequest{}, internalAPIError(err)
	}
	return request, nil
}

func decodeMappingPreviewRequest(reader io.Reader) (mappingPreviewRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return mappingPreviewRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"target_kind":             {},
		"extension_profile_id":    {},
		"owner_mapping_schema_id": {},
		"owner_mapping":           {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return mappingPreviewRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	request := mappingPreviewRequest{}
	if apiErr := decodeRequiredImportString(raw, "target_kind", &request.TargetKind); apiErr != nil {
		return mappingPreviewRequest{}, apiErr
	}
	if apiErr := decodeRequiredImportString(raw, "extension_profile_id", &request.ExtensionProfileID); apiErr != nil {
		return mappingPreviewRequest{}, apiErr
	}
	if apiErr := decodeRequiredImportString(raw, "owner_mapping_schema_id", &request.OwnerMappingSchemaID); apiErr != nil {
		return mappingPreviewRequest{}, apiErr
	}
	ownerMapping, ok := raw["owner_mapping"]
	if !ok {
		return mappingPreviewRequest{}, invalidImportRequest("owner_mapping", "missing_required_field")
	}
	if bytesEqualJSONNull(ownerMapping) {
		return mappingPreviewRequest{}, invalidImportRequest("owner_mapping", "field_not_nullable")
	}
	if _, err := httpapi.DecodeStrictJSONObject(bytes.NewReader(ownerMapping)); err != nil {
		return mappingPreviewRequest{}, invalidImportRequest("owner_mapping", "invalid_value")
	}
	request.OwnerMapping = append(json.RawMessage(nil), ownerMapping...)
	return request, nil
}

func decodeRequiredImportString(raw map[string]json.RawMessage, field string, destination *string) *httpapi.APIError {
	value, ok := raw[field]
	if !ok {
		return invalidImportRequest(field, "missing_required_field")
	}
	if bytesEqualJSONNull(value) {
		return invalidImportRequest(field, "field_not_nullable")
	}
	if err := json.Unmarshal(value, destination); err != nil || strings.TrimSpace(*destination) == "" {
		return invalidImportRequest(field, "invalid_value")
	}
	return nil
}

func normalizedMappingPayload(request mappingRequest) map[string]any {
	mapping := request.approvedMapping
	extensionVariant := mapping.TargetKind != ""
	normalizedMapping := map[string]any{
		"header_row_ref":     request.HeaderRowRef,
		"data_start_row_ref": request.DataStartRowRef,
		"source_columns":     mapping.SourceColumns,
	}
	if extensionVariant {
		normalizedMapping["target_kind"] = mapping.TargetKind
		normalizedMapping["extension_profile_id"] = mapping.ExtensionProfileID
		normalizedMapping["owner_mapping_schema_id"] = mapping.OwnerMappingSchemaID
		normalizedMapping["owner_mapping"] = json.RawMessage(mapping.OwnerMapping)
	} else {
		normalizedMapping["target_view_schema_id"] = mapping.TargetViewSchemaID
		normalizedMapping["unknown_column_policy"] = mapping.UnknownColumnPolicy
	}
	return normalizedMapping
}

func rebuildMappingRequestNormalized(request *mappingRequest) error {
	normalized, err := json.Marshal(normalizedMappingPayload(*request))
	if err != nil {
		return err
	}
	request.Normalized, err = json.Marshal(map[string]any{
		"client_txn_id": request.ClientTxnID,
		"mapping":       json.RawMessage(normalized),
	})
	return err
}

func decodeActionRequest(reader io.Reader, allowReason bool) (actionRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return actionRequest{}, apiErr
	}
	allowed := map[string]struct{}{"client_txn_id": {}}
	if allowReason {
		allowed["reason"] = struct{}{}
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return actionRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	var request actionRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return actionRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return actionRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["reason"]; ok && !bytesEqualJSONNull(value) {
		var reason string
		if err := json.Unmarshal(value, &reason); err != nil {
			return actionRequest{}, invalidImportRequest("reason", "invalid_value")
		}
		request.Reason = &reason
	}
	normalized := map[string]any{"client_txn_id": request.ClientTxnID}
	if allowReason {
		if request.Reason != nil {
			normalized["reason"] = *request.Reason
		} else {
			normalized["reason"] = nil
		}
	}
	var err error
	request.Normalized, err = json.Marshal(normalized)
	if err != nil {
		return actionRequest{}, internalAPIError(err)
	}
	return request, nil
}

func decodeApplyRequest(reader io.Reader) (applyRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return applyRequest{}, apiErr
	}
	allowed := map[string]struct{}{"client_txn_id": {}, "selected_unit_ids": {}}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return applyRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	var request applyRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return applyRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return applyRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	}
	normalized := map[string]any{"client_txn_id": request.ClientTxnID}
	if value, ok := raw["selected_unit_ids"]; ok {
		if bytesEqualJSONNull(value) {
			return applyRequest{}, invalidImportRequest("selected_unit_ids", "field_not_nullable")
		}
		var ids []string
		if err := json.Unmarshal(value, &ids); err != nil {
			return applyRequest{}, invalidImportRequest("selected_unit_ids", "invalid_selected_unit_ids")
		}
		parsed := make([]uuid.UUID, 0, len(ids))
		seen := map[uuid.UUID]struct{}{}
		for _, rawID := range ids {
			id, err := uuid.Parse(rawID)
			if err != nil || id.String() != rawID {
				return applyRequest{}, invalidImportRequest("selected_unit_ids", "invalid_selected_unit_ids")
			}
			if _, ok := seen[id]; ok {
				return applyRequest{}, invalidImportRequest("selected_unit_ids", "invalid_selected_unit_ids")
			}
			seen[id] = struct{}{}
			parsed = append(parsed, id)
		}
		request.SelectedUnitIDs = &parsed
		normalized["selected_unit_ids"] = uuidStrings(parsed)
	} else {
		normalized["selected_unit_ids"] = nil
	}
	var err error
	request.Normalized, err = json.Marshal(normalized)
	if err != nil {
		return applyRequest{}, internalAPIError(err)
	}
	return request, nil
}

func decodeRegionRequest(reader io.Reader) (regionRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return regionRequest{}, apiErr
	}
	allowed := map[string]struct{}{"client_txn_id": {}, "source_rect": {}}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return regionRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	var request regionRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return regionRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return regionRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	}
	value, ok := raw["source_rect"]
	if !ok {
		return regionRequest{}, invalidImportRequest("source_rect", "missing_required_field")
	}
	var sourceRectRaw map[string]json.RawMessage
	if err := json.Unmarshal(value, &sourceRectRaw); err != nil || sourceRectRaw == nil {
		return regionRequest{}, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	rectAllowed := map[string]struct{}{
		"start_row": {}, "start_column": {}, "end_row": {}, "end_column": {},
	}
	for key := range sourceRectRaw {
		if _, ok := rectAllowed[key]; !ok {
			return regionRequest{}, invalidImportRequest("source_rect", "invalid_source_rect")
		}
	}
	members := []struct {
		name   string
		target *int
	}{
		{name: "start_row", target: &request.SourceRect.StartRow},
		{name: "start_column", target: &request.SourceRect.StartColumn},
		{name: "end_row", target: &request.SourceRect.EndRow},
		{name: "end_column", target: &request.SourceRect.EndColumn},
	}
	for _, member := range members {
		rawValue, exists := sourceRectRaw[member.name]
		if !exists || json.Unmarshal(rawValue, member.target) != nil || *member.target <= 0 {
			return regionRequest{}, invalidImportRequest("source_rect", "invalid_source_rect")
		}
	}
	if request.SourceRect.StartRow > request.SourceRect.EndRow ||
		request.SourceRect.StartColumn > request.SourceRect.EndColumn {
		return regionRequest{}, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	var err error
	request.Normalized, err = json.Marshal(map[string]any{
		"client_txn_id": request.ClientTxnID,
		"source_rect":   request.SourceRect,
	})
	if err != nil {
		return regionRequest{}, internalAPIError(err)
	}
	return request, nil
}

func decodeJSONObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return nil, invalidImportRequest("", "request_not_object")
	}
	return raw, nil
}

func decodeSourceColumnMapping(raw json.RawMessage, extensionVariant bool) (sourceColumnMapping, *httpapi.APIError) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	}
	allowed := map[string]struct{}{
		"source_column_ordinal": {},
		"source_header_text":    {},
		"field_key":             {},
		"entity_binding_mode":   {},
		"transform_id":          {},
		"transform_options":     {},
		"empty_value_policy":    {},
	}
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
	}
	var column sourceColumnMapping
	if value, ok := object["source_column_ordinal"]; !ok {
		return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	} else if err := json.Unmarshal(value, &column.SourceColumnOrdinal); err != nil || column.SourceColumnOrdinal < 1 {
		return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	}
	if value, ok := object["source_header_text"]; ok && !bytesEqualJSONNull(value) {
		var header string
		if err := json.Unmarshal(value, &header); err != nil {
			return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
		column.SourceHeaderText = header
	}
	if value, ok := object["field_key"]; !ok {
		return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	} else if !bytesEqualJSONNull(value) {
		var fieldKey string
		if err := json.Unmarshal(value, &fieldKey); err != nil || strings.TrimSpace(fieldKey) == "" {
			return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
		column.FieldKey = &fieldKey
	}
	if value, ok := object["entity_binding_mode"]; !ok {
		return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	} else if !bytesEqualJSONNull(value) {
		var mode string
		if err := json.Unmarshal(value, &mode); err != nil || strings.TrimSpace(mode) == "" {
			return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
		column.EntityBindingMode = &mode
	}
	if value, ok := object["transform_id"]; !ok {
		return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	} else if !bytesEqualJSONNull(value) {
		var transformID string
		if err := json.Unmarshal(value, &transformID); err != nil || (!extensionVariant && !validTransformID(transformID)) || strings.TrimSpace(transformID) == "" {
			return sourceColumnMapping{}, invalidImportRequest("transform_id", "invalid_transform")
		}
		column.TransformID = &transformID
	}
	column.TransformOptions = map[string]any{}
	if value, ok := object["transform_options"]; !ok {
		return sourceColumnMapping{}, invalidImportRequest("transform_options", "missing_required_field")
	} else if err := json.Unmarshal(value, &column.TransformOptions); err != nil || column.TransformOptions == nil {
		return sourceColumnMapping{}, invalidImportRequest("transform_options", "invalid_transform")
	}
	if !extensionVariant {
		if apiErr := validateTransformOptions(column.TransformID, column.TransformOptions); apiErr != nil {
			return sourceColumnMapping{}, apiErr
		}
	} else if column.TransformOptions == nil {
		column.TransformOptions = map[string]any{}
	}
	if value, ok := object["empty_value_policy"]; !ok {
		return sourceColumnMapping{}, invalidImportRequest("empty_value_policy", "missing_required_field")
	} else if err := json.Unmarshal(value, &column.EmptyValuePolicy); err != nil || strings.TrimSpace(column.EmptyValuePolicy) == "" || (!extensionVariant && !validEmptyValuePolicy(column.EmptyValuePolicy)) {
		return sourceColumnMapping{}, invalidImportRequest("empty_value_policy", "invalid_empty_value_policy")
	}
	if column.FieldKey == nil {
		if column.EntityBindingMode != nil || column.EmptyValuePolicy != "omit_field" {
			return sourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
	}
	return column, nil
}

func validUnknownColumnPolicy(value string) bool {
	switch value {
	case "preserve_raw_capture", "preserve_custom_attrs", "reject_if_unmapped":
		return true
	default:
		return false
	}
}

func validTransformID(value string) bool {
	switch value {
	case "trim_v1", "collapse_whitespace_v1", "lowercase_v1", "split_delimited_v1":
		return true
	default:
		return false
	}
}

func validEmptyValuePolicy(value string) bool {
	switch value {
	case "omit_field", "write_null":
		return true
	default:
		return false
	}
}

func validateTransformOptions(transformID *string, options map[string]any) *httpapi.APIError {
	if transformID == nil || *transformID != "split_delimited_v1" {
		if len(options) != 0 {
			return invalidImportRequest("transform_options", "invalid_transform")
		}
		return nil
	}
	for key := range options {
		switch key {
		case "delimiter", "trim_items", "drop_empty_items":
		default:
			return invalidImportRequest("transform_options", "invalid_transform")
		}
	}
	delimiter, ok := options["delimiter"].(string)
	if !ok || !slices.Contains([]string{",", ";", "|", "\n", "\t"}, delimiter) {
		return invalidImportRequest("transform_options", "invalid_transform")
	}
	if value, ok := options["trim_items"]; ok {
		if _, ok := value.(bool); !ok {
			return invalidImportRequest("transform_options", "invalid_transform")
		}
	}
	if value, ok := options["drop_empty_items"]; ok {
		if _, ok := value.(bool); !ok {
			return invalidImportRequest("transform_options", "invalid_transform")
		}
	}
	return nil
}

func sourceHeaderEqual(left any, right any) bool {
	leftBytes, _ := json.Marshal(left)
	rightBytes, _ := json.Marshal(right)
	return string(leftBytes) == string(rightBytes)
}

func uuidStrings(ids []uuid.UUID) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, id.String())
	}
	return values
}

func parseUUIDStrings(values []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		parsed, err := uuid.Parse(value)
		if err != nil {
			return nil, err
		}
		ids = append(ids, parsed)
	}
	return ids, nil
}

func invalidImportRequest(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{"reason_code": reasonCode}
	if field != "" {
		details["field"] = field
	}
	return &httpapi.APIError{Status: 400, Code: "invalid_import_request", Details: details}
}

func invalidImportRequestWithOwner(
	field string,
	reasonCode string,
	ownerError map[string]any,
) *httpapi.APIError {
	apiErr := invalidImportRequest(field, reasonCode)
	if len(ownerError) != 0 {
		apiErr.Details["owner_error"] = ownerError
	}
	return apiErr
}

func uploadEnvelopeAPIError(apiErr *httpapi.UploadEnvelopeError) *httpapi.APIError {
	if apiErr == nil {
		return nil
	}
	return &httpapi.APIError{
		Status:  400,
		Code:    "invalid_import_request",
		Message: fmt.Sprintf("invalid import request: %s", apiErr.ReasonCode),
		Details: apiErr.Details(),
	}
}

func importSourceRejected(reasonCode string, requestedBytes int64, limitBytes int64) *httpapi.APIError {
	return &httpapi.APIError{
		Status: 413,
		Code:   "import_source_rejected",
		Details: map[string]any{
			"reason_code":            reasonCode,
			"requested_byte_size":    requestedBytes,
			"configured_limit_bytes": limitBytes,
		},
	}
}

func importSourceUnsupported(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status: 409,
		Code:   "import_source_unsupported",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func clientTxnConflict(clientTxnID string) *httpapi.APIError {
	return &httpapi.APIError{Status: 409, Code: "client_txn_conflict", Details: map[string]any{"client_txn_id": clientTxnID}}
}

func invalidPaginationRequest(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func bytesEqualJSONNull(value json.RawMessage) bool {
	return strings.TrimSpace(string(value)) == "null"
}
