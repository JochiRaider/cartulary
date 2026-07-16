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

	AssistantProfilePhase2WorkbookImport = "phase2_workbook_import_v1"
	ParserProfilePhase2WorkbookImport    = "cartulary.import.phase2_workbook_import.v1"
	ParserVersionPhase11                 = "phase11_import_adapter_v1"

	SourceFileKindCSV  = "csv"
	SourceFileKindXLSX = "xlsx"

	MediaTypeCSV         = "text/csv"
	MediaTypeXLSX        = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	MediaTypeOctetStream = "application/octet-stream"

	importDiscoveryJobHandlerName = "imports.discovery"
	importApplyJobHandlerName     = "imports.apply"

	ImportTargetKindViewSchema       = "view_schema"
	ImportTargetKindNetworkFlowTable = "network_flow_table"
	NetworkFlowExtensionProfileID    = "network_flow_activity"

	ExtensionMappingPreviewResultSchemaID = "cartulary.imports.extension_mapping_preview_result.v1"
)

var ImportSessionFileContentTypes = []string{
	MediaTypeCSV,
	MediaTypeXLSX,
	MediaTypeOctetStream,
}

type CreateSessionRequest struct {
	IncidentID       uuid.UUID
	ClientTxnID      string
	AssistantProfile string
}

type SourceColumnMapping struct {
	SourceColumnOrdinal int            `json:"source_column_ordinal"`
	SourceHeaderText    any            `json:"source_header_text"`
	FieldKey            *string        `json:"field_key"`
	EntityBindingMode   *string        `json:"entity_binding_mode"`
	TransformID         *string        `json:"transform_id"`
	TransformOptions    map[string]any `json:"transform_options"`
	EmptyValuePolicy    string         `json:"empty_value_policy"`
}

type ApprovedMapping struct {
	TargetKind           string                `json:"target_kind,omitempty"`
	TargetViewSchemaID   string                `json:"target_view_schema_id,omitempty"`
	ExtensionProfileID   string                `json:"extension_profile_id,omitempty"`
	OwnerMappingSchemaID string                `json:"owner_mapping_schema_id,omitempty"`
	OwnerMapping         json.RawMessage       `json:"owner_mapping,omitempty"`
	UnknownColumnPolicy  string                `json:"unknown_column_policy,omitempty"`
	SourceColumns        []SourceColumnMapping `json:"source_columns"`
}

func (mapping ApprovedMapping) targetKindOrDefault() string {
	if mapping.TargetKind == "" {
		return ImportTargetKindViewSchema
	}
	return mapping.TargetKind
}

type MappingRequest struct {
	ClientTxnID     string
	HeaderRowRef    int
	DataStartRowRef int
	ApprovedMapping ApprovedMapping
	Fingerprint     string
	Normalized      []byte
}

type MappingPreviewRequest struct {
	TargetKind           string
	ExtensionProfileID   string
	OwnerMappingSchemaID string
	OwnerMapping         json.RawMessage
}

type ExtensionMappingPreviewResource struct {
	SchemaID            string         `json:"schema_id"`
	ImportSessionID     string         `json:"import_session_id"`
	ImportUnitID        string         `json:"import_unit_id"`
	TargetKind          string         `json:"target_kind"`
	ExtensionProfileID  string         `json:"extension_profile_id"`
	OwnerResultSchemaID string         `json:"owner_result_schema_id"`
	OwnerResult         map[string]any `json:"owner_result"`
}

type ActionRequest struct {
	ClientTxnID string
	Reason      *string
	Normalized  []byte
}

type ApplyRequest struct {
	ClientTxnID     string
	SelectedUnitIDs *[]uuid.UUID
	Normalized      []byte
}

func DecodeCreateSessionMetadata(envelope httpapi.UploadEnvelope) (CreateSessionRequest, *httpapi.APIError) {
	allowed := map[string]struct{}{
		"incident_id":       {},
		"client_txn_id":     {},
		"assistant_profile": {},
	}
	for key := range envelope.Metadata {
		if _, ok := allowed[key]; !ok {
			return CreateSessionRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	var request CreateSessionRequest
	if value, ok := envelope.Metadata["incident_id"]; !ok {
		return CreateSessionRequest{}, invalidImportRequest("incident_id", "missing_required_field")
	} else {
		var raw string
		if err := json.Unmarshal(value, &raw); err != nil {
			return CreateSessionRequest{}, invalidImportRequest("incident_id", "invalid_value")
		}
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return CreateSessionRequest{}, invalidImportRequest("incident_id", "invalid_value")
		}
		request.IncidentID = parsed
	}
	if value, ok := envelope.Metadata["client_txn_id"]; !ok {
		return CreateSessionRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateSessionRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	}
	request.AssistantProfile = AssistantProfilePhase2WorkbookImport
	if value, ok := envelope.Metadata["assistant_profile"]; ok {
		if bytesEqualJSONNull(value) {
			return CreateSessionRequest{}, invalidImportRequest("assistant_profile", "field_not_nullable")
		}
		var profile string
		if err := json.Unmarshal(value, &profile); err != nil {
			return CreateSessionRequest{}, invalidImportRequest("assistant_profile", "invalid_value")
		}
		if profile != AssistantProfilePhase2WorkbookImport {
			return CreateSessionRequest{}, invalidImportRequest("assistant_profile", "unsupported_assistant_profile")
		}
		request.AssistantProfile = profile
	}
	return request, nil
}

func DecodeMappingRequest(reader io.Reader, discoveredColumns []map[string]any) (MappingRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return MappingRequest{}, apiErr
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
			return MappingRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	request := MappingRequest{}
	if value, ok := raw["client_txn_id"]; !ok {
		return MappingRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return MappingRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	}
	mapping := ApprovedMapping{}
	if value, ok := raw["target_kind"]; ok {
		if bytesEqualJSONNull(value) {
			return MappingRequest{}, invalidImportRequest("target_kind", "field_not_nullable")
		}
		if err := json.Unmarshal(value, &mapping.TargetKind); err != nil || strings.TrimSpace(mapping.TargetKind) == "" {
			return MappingRequest{}, invalidImportRequest("target_kind", "invalid_value")
		}
	}
	extensionVariant := mapping.TargetKind != ""
	if extensionVariant {
		if _, ok := raw["target_view_schema_id"]; ok {
			return MappingRequest{}, invalidImportRequest("target_view_schema_id", "invalid_mapping_variant")
		}
		if _, ok := raw["unknown_column_policy"]; ok {
			return MappingRequest{}, invalidImportRequest("unknown_column_policy", "invalid_mapping_variant")
		}
		if value, ok := raw["extension_profile_id"]; !ok {
			return MappingRequest{}, invalidImportRequest("extension_profile_id", "missing_required_field")
		} else if err := json.Unmarshal(value, &mapping.ExtensionProfileID); err != nil || strings.TrimSpace(mapping.ExtensionProfileID) == "" {
			return MappingRequest{}, invalidImportRequest("extension_profile_id", "invalid_value")
		}
		if value, ok := raw["owner_mapping_schema_id"]; !ok {
			return MappingRequest{}, invalidImportRequest("owner_mapping_schema_id", "missing_required_field")
		} else if err := json.Unmarshal(value, &mapping.OwnerMappingSchemaID); err != nil || strings.TrimSpace(mapping.OwnerMappingSchemaID) == "" {
			return MappingRequest{}, invalidImportRequest("owner_mapping_schema_id", "invalid_value")
		}
		if value, ok := raw["owner_mapping"]; !ok {
			return MappingRequest{}, invalidImportRequest("owner_mapping", "missing_required_field")
		} else if bytesEqualJSONNull(value) || !json.Valid(value) {
			return MappingRequest{}, invalidImportRequest("owner_mapping", "invalid_value")
		} else {
			mapping.OwnerMapping = append(json.RawMessage(nil), value...)
		}
	} else {
		if _, ok := raw["extension_profile_id"]; ok {
			return MappingRequest{}, invalidImportRequest("extension_profile_id", "invalid_mapping_variant")
		}
		if _, ok := raw["owner_mapping_schema_id"]; ok {
			return MappingRequest{}, invalidImportRequest("owner_mapping_schema_id", "invalid_mapping_variant")
		}
		if _, ok := raw["owner_mapping"]; ok {
			return MappingRequest{}, invalidImportRequest("owner_mapping", "invalid_mapping_variant")
		}
		if value, ok := raw["target_view_schema_id"]; !ok {
			return MappingRequest{}, invalidImportRequest("target_view_schema_id", "missing_required_field")
		} else if err := json.Unmarshal(value, &mapping.TargetViewSchemaID); err != nil || strings.TrimSpace(mapping.TargetViewSchemaID) == "" {
			return MappingRequest{}, invalidImportRequest("target_view_schema_id", "invalid_value")
		}
		if value, ok := raw["unknown_column_policy"]; !ok {
			return MappingRequest{}, invalidImportRequest("unknown_column_policy", "missing_required_field")
		} else if err := json.Unmarshal(value, &mapping.UnknownColumnPolicy); err != nil || !validUnknownColumnPolicy(mapping.UnknownColumnPolicy) {
			return MappingRequest{}, invalidImportRequest("unknown_column_policy", "invalid_unknown_column_policy")
		}
	}
	if value, ok := raw["header_row_ref"]; !ok {
		return MappingRequest{}, invalidImportRequest("header_row_ref", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.HeaderRowRef); err != nil || request.HeaderRowRef < 1 {
		return MappingRequest{}, invalidImportRequest("header_row_ref", "invalid_row_reference")
	}
	if value, ok := raw["data_start_row_ref"]; !ok {
		return MappingRequest{}, invalidImportRequest("data_start_row_ref", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.DataStartRowRef); err != nil || request.DataStartRowRef < request.HeaderRowRef+1 {
		return MappingRequest{}, invalidImportRequest("data_start_row_ref", "invalid_row_reference")
	}
	var rawColumns []json.RawMessage
	if value, ok := raw["source_columns"]; !ok {
		return MappingRequest{}, invalidImportRequest("source_columns", "missing_required_field")
	} else if err := json.Unmarshal(value, &rawColumns); err != nil {
		return MappingRequest{}, invalidImportRequest("source_columns", "invalid_source_columns")
	}
	if len(rawColumns) != len(discoveredColumns) {
		return MappingRequest{}, invalidImportRequest("source_columns", "invalid_source_columns")
	}
	mapping.SourceColumns = make([]SourceColumnMapping, 0, len(rawColumns))
	seenTargets := map[string]struct{}{}
	for index, rawColumn := range rawColumns {
		column, apiErr := decodeSourceColumnMapping(rawColumn, extensionVariant)
		if apiErr != nil {
			return MappingRequest{}, apiErr
		}
		wantOrdinal := index + 1
		if column.SourceColumnOrdinal != wantOrdinal {
			return MappingRequest{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
		discovered := discoveredColumns[index]
		if !sourceHeaderEqual(column.SourceHeaderText, discovered["source_header_text"]) {
			return MappingRequest{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
		if column.FieldKey != nil {
			if _, ok := seenTargets[*column.FieldKey]; ok {
				return MappingRequest{}, invalidImportRequest("source_columns", "duplicate_target_field")
			}
			seenTargets[*column.FieldKey] = struct{}{}
		}
		mapping.SourceColumns = append(mapping.SourceColumns, column)
	}
	request.ApprovedMapping = mapping
	normalizedMapping := normalizedMappingPayload(request)
	normalized, err := json.Marshal(normalizedMapping)
	if err != nil {
		return MappingRequest{}, internalAPIError(err)
	}
	sum := sha256.Sum256(normalized)
	request.Fingerprint = hex.EncodeToString(sum[:])
	if err := RebuildMappingRequestNormalized(&request); err != nil {
		return MappingRequest{}, internalAPIError(err)
	}
	return request, nil
}

func DecodeMappingPreviewRequest(reader io.Reader) (MappingPreviewRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return MappingPreviewRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"target_kind":             {},
		"extension_profile_id":    {},
		"owner_mapping_schema_id": {},
		"owner_mapping":           {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return MappingPreviewRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	request := MappingPreviewRequest{}
	if apiErr := decodeRequiredImportString(raw, "target_kind", &request.TargetKind); apiErr != nil {
		return MappingPreviewRequest{}, apiErr
	}
	if apiErr := decodeRequiredImportString(raw, "extension_profile_id", &request.ExtensionProfileID); apiErr != nil {
		return MappingPreviewRequest{}, apiErr
	}
	if apiErr := decodeRequiredImportString(raw, "owner_mapping_schema_id", &request.OwnerMappingSchemaID); apiErr != nil {
		return MappingPreviewRequest{}, apiErr
	}
	ownerMapping, ok := raw["owner_mapping"]
	if !ok {
		return MappingPreviewRequest{}, invalidImportRequest("owner_mapping", "missing_required_field")
	}
	if bytesEqualJSONNull(ownerMapping) {
		return MappingPreviewRequest{}, invalidImportRequest("owner_mapping", "field_not_nullable")
	}
	if _, err := httpapi.DecodeStrictJSONObject(bytes.NewReader(ownerMapping)); err != nil {
		return MappingPreviewRequest{}, invalidImportRequest("owner_mapping", "invalid_value")
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

func normalizedMappingPayload(request MappingRequest) map[string]any {
	mapping := request.ApprovedMapping
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

func RebuildMappingRequestNormalized(request *MappingRequest) error {
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

func DecodeActionRequest(reader io.Reader, allowReason bool) (ActionRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return ActionRequest{}, apiErr
	}
	allowed := map[string]struct{}{"client_txn_id": {}}
	if allowReason {
		allowed["reason"] = struct{}{}
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ActionRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	var request ActionRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return ActionRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ActionRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["reason"]; ok && !bytesEqualJSONNull(value) {
		var reason string
		if err := json.Unmarshal(value, &reason); err != nil {
			return ActionRequest{}, invalidImportRequest("reason", "invalid_value")
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
		return ActionRequest{}, internalAPIError(err)
	}
	return request, nil
}

func DecodeApplyRequest(reader io.Reader) (ApplyRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return ApplyRequest{}, apiErr
	}
	allowed := map[string]struct{}{"client_txn_id": {}, "selected_unit_ids": {}}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ApplyRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	var request ApplyRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return ApplyRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ApplyRequest{}, invalidImportRequest("client_txn_id", "missing_required_field")
	}
	normalized := map[string]any{"client_txn_id": request.ClientTxnID}
	if value, ok := raw["selected_unit_ids"]; ok {
		if bytesEqualJSONNull(value) {
			return ApplyRequest{}, invalidImportRequest("selected_unit_ids", "field_not_nullable")
		}
		var ids []string
		if err := json.Unmarshal(value, &ids); err != nil {
			return ApplyRequest{}, invalidImportRequest("selected_unit_ids", "invalid_selected_unit_ids")
		}
		parsed := make([]uuid.UUID, 0, len(ids))
		seen := map[uuid.UUID]struct{}{}
		for _, rawID := range ids {
			id, err := uuid.Parse(rawID)
			if err != nil || id.String() != rawID {
				return ApplyRequest{}, invalidImportRequest("selected_unit_ids", "invalid_selected_unit_ids")
			}
			if _, ok := seen[id]; ok {
				return ApplyRequest{}, invalidImportRequest("selected_unit_ids", "invalid_selected_unit_ids")
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
		return ApplyRequest{}, internalAPIError(err)
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

func decodeSourceColumnMapping(raw json.RawMessage, extensionVariant bool) (SourceColumnMapping, *httpapi.APIError) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
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
			return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
	}
	var column SourceColumnMapping
	if value, ok := object["source_column_ordinal"]; !ok {
		return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	} else if err := json.Unmarshal(value, &column.SourceColumnOrdinal); err != nil || column.SourceColumnOrdinal < 1 {
		return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	}
	if value, ok := object["source_header_text"]; ok && !bytesEqualJSONNull(value) {
		var header string
		if err := json.Unmarshal(value, &header); err != nil {
			return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
		column.SourceHeaderText = header
	}
	if value, ok := object["field_key"]; !ok {
		return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	} else if !bytesEqualJSONNull(value) {
		var fieldKey string
		if err := json.Unmarshal(value, &fieldKey); err != nil || strings.TrimSpace(fieldKey) == "" {
			return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
		column.FieldKey = &fieldKey
	}
	if value, ok := object["entity_binding_mode"]; !ok {
		return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	} else if !bytesEqualJSONNull(value) {
		var mode string
		if err := json.Unmarshal(value, &mode); err != nil || strings.TrimSpace(mode) == "" {
			return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
		}
		column.EntityBindingMode = &mode
	}
	if value, ok := object["transform_id"]; !ok {
		return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
	} else if !bytesEqualJSONNull(value) {
		var transformID string
		if err := json.Unmarshal(value, &transformID); err != nil || (!extensionVariant && !validTransformID(transformID)) || strings.TrimSpace(transformID) == "" {
			return SourceColumnMapping{}, invalidImportRequest("transform_id", "invalid_transform")
		}
		column.TransformID = &transformID
	}
	column.TransformOptions = map[string]any{}
	if value, ok := object["transform_options"]; !ok {
		return SourceColumnMapping{}, invalidImportRequest("transform_options", "missing_required_field")
	} else if err := json.Unmarshal(value, &column.TransformOptions); err != nil || column.TransformOptions == nil {
		return SourceColumnMapping{}, invalidImportRequest("transform_options", "invalid_transform")
	}
	if !extensionVariant {
		if apiErr := validateTransformOptions(column.TransformID, column.TransformOptions); apiErr != nil {
			return SourceColumnMapping{}, apiErr
		}
	} else if column.TransformOptions == nil {
		column.TransformOptions = map[string]any{}
	}
	if value, ok := object["empty_value_policy"]; !ok {
		return SourceColumnMapping{}, invalidImportRequest("empty_value_policy", "missing_required_field")
	} else if err := json.Unmarshal(value, &column.EmptyValuePolicy); err != nil || strings.TrimSpace(column.EmptyValuePolicy) == "" || (!extensionVariant && !validEmptyValuePolicy(column.EmptyValuePolicy)) {
		return SourceColumnMapping{}, invalidImportRequest("empty_value_policy", "invalid_empty_value_policy")
	}
	if column.FieldKey == nil {
		if column.EntityBindingMode != nil || column.EmptyValuePolicy != "omit_field" {
			return SourceColumnMapping{}, invalidImportRequest("source_columns", "invalid_source_columns")
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
