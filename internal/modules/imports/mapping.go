package imports

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/google/uuid"
)

func (s *Service) extensionProfileClaimed(profileID string) bool {
	return s.extensionProfileAdmitted(profileID)
}

func (s *Service) prepareApprovedMapping(ctx context.Context, actorUserID uuid.UUID, incidentID uuid.UUID, route importSessionRoute, request MappingRequest) (MappingRequest, *httpapi.APIError) {
	if apiErr := s.validateApprovedMapping(request.ApprovedMapping); apiErr != nil {
		return MappingRequest{}, apiErr
	}
	if request.ApprovedMapping.targetKindOrDefault() == ImportTargetKindViewSchema {
		return request, nil
	}
	target, ok := lookupApprovedImportTarget(request.ApprovedMapping)
	if !ok {
		return MappingRequest{}, invalidImportRequest("target_kind", "target_kind_not_importable")
	}
	facade := s.extensionImportFacades[extensionImportFacadeKey(target)]
	if facade == nil {
		return MappingRequest{}, invalidImportRequest("target_kind", "owner_apply_contract_unavailable")
	}
	sourceCapability, err := s.store.SourceCapabilityForUnit(ctx, route.SessionID, route.UnitID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return MappingRequest{}, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}}
		}
		return MappingRequest{}, internalAPIError(err)
	}
	result, err := facade.PrepareImportUnitMapping(ctx, ExtensionImportMappingRequest{
		IncidentID:           incidentID,
		ActorUserID:          actorUserID,
		TargetKind:           request.ApprovedMapping.TargetKind,
		ExtensionProfileID:   request.ApprovedMapping.ExtensionProfileID,
		ImportSessionID:      route.SessionID,
		ImportUnitID:         route.UnitID,
		SourceCapability:     sourceCapability,
		OwnerMappingSchemaID: request.ApprovedMapping.OwnerMappingSchemaID,
		OwnerMapping:         append(json.RawMessage(nil), request.ApprovedMapping.OwnerMapping...),
		ClientTxnID:          request.ClientTxnID,
	})
	if err != nil {
		return MappingRequest{}, extensionFacadeAPIError(err)
	}
	if err := facade.ValidateImportUnitMappingResult(result); err != nil {
		return MappingRequest{}, internalAPIError(fmt.Errorf("extension mapping preview result validation failed: %w", err))
	}
	if len(result.OwnerMapping) == 0 || result.MappingFingerprint == "" || result.OwnerResultSchemaID == "" || result.OwnerResult == nil {
		return MappingRequest{}, internalAPIError(fmt.Errorf("extension mapping facade returned incomplete mapping result"))
	}
	request.ApprovedMapping.OwnerMapping = append(json.RawMessage(nil), result.OwnerMapping...)
	request.Fingerprint = result.MappingFingerprint
	if err := RebuildMappingRequestNormalized(&request); err != nil {
		return MappingRequest{}, internalAPIError(err)
	}
	return request, nil
}

func (s *Service) validateApprovedMapping(mapping ApprovedMapping) *httpapi.APIError {
	target, ok := lookupApprovedImportTarget(mapping)
	if !ok || !target.importable(s.extensionProfileClaimed) {
		if mapping.targetKindOrDefault() == ImportTargetKindViewSchema {
			return invalidImportRequest("target_view_schema_id", "target_view_schema_not_importable")
		}
		return invalidImportRequest("target_kind", "target_kind_not_importable")
	}
	if mapping.targetKindOrDefault() != ImportTargetKindViewSchema {
		if !target.ownerApplyFacadeAvailable() {
			return invalidImportRequest("target_kind", "owner_apply_contract_unavailable")
		}
		return nil
	}
	schema, ok := viewschema.Lookup(mapping.TargetViewSchemaID)
	if !ok {
		return invalidImportRequest("target_view_schema_id", "target_view_schema_not_importable")
	}
	switch mapping.UnknownColumnPolicy {
	case "preserve_raw_capture":
		if !target.AllowRawCapture {
			return invalidImportRequest("unknown_column_policy", "unknown_column_policy_not_supported_for_target")
		}
	case "preserve_custom_attrs":
		if !target.AllowCustomAttrs {
			return invalidImportRequest("unknown_column_policy", "unknown_column_policy_not_supported_for_target")
		}
	}
	fields := schema.Fields()
	for _, column := range mapping.SourceColumns {
		if column.FieldKey == nil {
			if mapping.UnknownColumnPolicy == "reject_if_unmapped" {
				return invalidImportRequest("source_columns", "invalid_source_columns")
			}
			continue
		}
		field, ok := fields[*column.FieldKey]
		if !ok || (!field.Writable && !field.CreateWritable) {
			return invalidImportRequest("source_columns", "field_not_import_writable")
		}
		if field.EntityBindingMode == nil {
			if column.EntityBindingMode != nil {
				return invalidImportRequest("source_columns", "invalid_source_columns")
			}
		} else if column.EntityBindingMode == nil || *column.EntityBindingMode != *field.EntityBindingMode {
			return invalidImportRequest("source_columns", "invalid_source_columns")
		}
	}
	return nil
}

func extensionFacadeAPIError(err error) *httpapi.APIError {
	var applyBlocked *ApplyBlockedError
	if errors.As(err, &applyBlocked) && applyBlocked.ReasonCode != "" {
		field := "owner_mapping"
		if applyBlocked.Field != "" {
			field = applyBlocked.Field
		}
		return invalidImportRequest(field, applyBlocked.ReasonCode)
	}
	return invalidImportRequest("owner_mapping", "owner_mapping_invalid")
}
