package imports

import (
	"context"
	"io"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
)

type sourceInspectionRequest struct {
	TargetKind         string
	ExtensionProfileID string
}

func decodeSourceInspectionRequest(reader io.Reader) (sourceInspectionRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return sourceInspectionRequest{}, apiErr
	}
	for key := range raw {
		if key != "target_kind" && key != "extension_profile_id" {
			return sourceInspectionRequest{}, invalidImportRequest(key, "unknown_field")
		}
	}
	var request sourceInspectionRequest
	if apiErr := decodeRequiredImportString(raw, "target_kind", &request.TargetKind); apiErr != nil {
		return sourceInspectionRequest{}, apiErr
	}
	if apiErr := decodeRequiredImportString(raw, "extension_profile_id", &request.ExtensionProfileID); apiErr != nil {
		return sourceInspectionRequest{}, apiErr
	}
	return request, nil
}

func (s *service) applicationInspectSource(ctx context.Context, actorID, incidentID uuid.UUID, route importSessionRoute, request sourceInspectionRequest) (extensionImportReadResource, *httpapi.APIError) {
	mapping := approvedMapping{TargetKind: request.TargetKind, ExtensionProfileID: request.ExtensionProfileID}
	if apiErr := s.validateApprovedMapping(mapping); apiErr != nil {
		return extensionImportReadResource{}, apiErr
	}
	target, ok := lookupApprovedImportTarget(mapping)
	if !ok {
		return extensionImportReadResource{}, invalidImportRequest("target_kind", "target_kind_not_importable")
	}
	facade := s.extensionImportFacades[analyticalImportTargetKey(request)]
	if facade == nil {
		return extensionImportReadResource{}, invalidImportRequest("target_kind", "owner_preview_contract_unavailable")
	}
	source, err := s.store.sourceCapabilityForUnit(ctx, route.SessionID, route.UnitID)
	if err != nil {
		return extensionImportReadResource{}, internalAPIError(err)
	}
	result, err := facade.InspectImportUnitSource(ctx, ExtensionImportSourceRequest{
		IncidentID: incidentID, ActorUserID: actorID, TargetKind: request.TargetKind, ExtensionProfileID: request.ExtensionProfileID,
		ImportSessionID: route.SessionID, ImportUnitID: route.UnitID, SourceCapability: source,
	})
	if err != nil {
		return extensionImportReadResource{}, extensionFacadeAPIError(target, facade, err)
	}
	if facade.ValidateImportUnitSourceResult(result) != nil {
		return extensionImportReadResource{}, invalidImportRequest("owner_result", "owner_preview_validation_failed")
	}
	return extensionImportReadResource{
		SchemaID: "cartulary.imports.extension_source_inspection_result.v1", ImportSessionID: route.SessionID.String(), ImportUnitID: route.UnitID.String(),
		TargetKind: request.TargetKind, ExtensionProfileID: request.ExtensionProfileID, OwnerResultSchemaID: result.OwnerResultSchemaID, OwnerResult: result.OwnerResult,
	}, nil
}
