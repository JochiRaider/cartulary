package imports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/google/uuid"
)

func (s *service) extensionProfileClaimed(profileID string) bool {
	return s.extensionProfileAdmitted(profileID)
}

func (s *service) prepareApprovedMapping(ctx context.Context, actorUserID uuid.UUID, incidentID uuid.UUID, route importSessionRoute, request mappingRequest) (mappingRequest, *httpapi.APIError) {
	if apiErr := s.validateApprovedMapping(request.approvedMapping); apiErr != nil {
		return mappingRequest{}, apiErr
	}
	if request.approvedMapping.targetKindOrDefault() == ImportTargetKindViewSchema {
		return request, nil
	}
	target, ok := lookupApprovedImportTarget(request.approvedMapping)
	if !ok {
		return mappingRequest{}, invalidImportRequest("target_kind", "target_kind_not_importable")
	}
	facade := s.extensionImportFacades[analyticalImportTargetKey{
		TargetKind:         target.TargetKind,
		ExtensionProfileID: target.ExtensionProfileID,
	}]
	if facade == nil {
		return mappingRequest{}, invalidImportRequest(
			"target_kind",
			"owner_preview_contract_unavailable",
		)
	}
	sourceCapability, err := s.store.sourceCapabilityForUnit(ctx, route.SessionID, route.UnitID)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return mappingRequest{}, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}}
		}
		return mappingRequest{}, internalAPIError(err)
	}
	result, err := facade.PrepareImportUnitMapping(ctx, ExtensionImportMappingRequest{
		IncidentID:           incidentID,
		ActorUserID:          actorUserID,
		TargetKind:           request.approvedMapping.TargetKind,
		ExtensionProfileID:   request.approvedMapping.ExtensionProfileID,
		ImportSessionID:      route.SessionID,
		ImportUnitID:         route.UnitID,
		SourceCapability:     sourceCapability,
		OwnerMappingSchemaID: request.approvedMapping.OwnerMappingSchemaID,
		OwnerMapping:         append(json.RawMessage(nil), request.approvedMapping.OwnerMapping...),
		ClientTxnID:          request.ClientTxnID,
	})
	if err != nil {
		return mappingRequest{}, extensionFacadeAPIError(target, facade, err)
	}
	if err := facade.ValidateImportUnitMappingResult(result); err != nil {
		return mappingRequest{}, invalidImportRequest(
			"owner_result",
			"owner_preview_validation_failed",
		)
	}
	if len(result.OwnerMapping) == 0 || result.MappingFingerprint == "" || result.OwnerResultSchemaID == "" || result.OwnerResult == nil {
		return mappingRequest{}, invalidImportRequest(
			"owner_result",
			"owner_preview_validation_failed",
		)
	}
	request.approvedMapping.OwnerMapping = append(json.RawMessage(nil), result.OwnerMapping...)
	request.Fingerprint = result.MappingFingerprint
	if err := rebuildMappingRequestNormalized(&request); err != nil {
		return mappingRequest{}, internalAPIError(err)
	}
	return request, nil
}

func (s *service) applicationApproveMapping(
	ctx context.Context,
	actorUserID uuid.UUID,
	incidentID uuid.UUID,
	route importSessionRoute,
	request mappingRequest,
) (map[string]any, error, *httpapi.APIError) {
	materialized, apiErr := s.prepareApprovedMapping(
		ctx,
		actorUserID,
		incidentID,
		route,
		request,
	)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	unit, _, err := s.store.saveMapping(ctx, mappingParams{
		ActorUserID:       actorUserID,
		SessionID:         route.SessionID,
		UnitID:            route.UnitID,
		Request:           materialized,
		NormalizedRequest: materialized.Normalized,
		Now:               s.now(),
	})
	return unit, err, nil
}

func (s *service) validateApprovedMapping(mapping approvedMapping) *httpapi.APIError {
	target, ok := lookupApprovedImportTarget(mapping)
	if !ok || !target.importable(s.extensionProfileClaimed) {
		if mapping.targetKindOrDefault() == ImportTargetKindViewSchema {
			return invalidImportRequest("target_view_schema_id", "target_view_schema_not_importable")
		}
		return invalidImportRequest("target_kind", "target_kind_not_importable")
	}
	if mapping.targetKindOrDefault() != ImportTargetKindViewSchema {
		if !target.ownerApplyFacadeAvailable() {
			return invalidImportRequest("target_kind", "owner_preview_contract_unavailable")
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
		if column.EmptyValuePolicy == "write_null" && !field.Clearable {
			return invalidImportRequest("empty_value_policy", "invalid_empty_value_policy")
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

func extensionFacadeAPIError(
	target importTarget,
	facade ExtensionImportFacade,
	err error,
) *httpapi.APIError {
	failure := translateExtensionOwnerFailure(target, facade, err)
	field := "owner_mapping"
	if ownerError, ok := failure.Details["owner_error"].(map[string]any); ok {
		return invalidImportRequestWithOwner(
			field,
			"owner_preview_validation_failed",
			ownerError,
		)
	}
	if failure.ReasonCode == "owner_apply_contract_unavailable" {
		return invalidImportRequest(field, "owner_preview_contract_unavailable")
	}
	return invalidImportRequest(field, "owner_preview_validation_failed")
}
