package imports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
)

type importScopedValue[T any] struct {
	Value      T
	IncidentID uuid.UUID
}

func (s *Service) applicationCreateSession(
	ctx context.Context,
	actorUserID uuid.UUID,
	envelope httpapi.UploadEnvelope,
	request CreateSessionRequest,
) (CreateAcceptedSessionResult, error, *httpapi.APIError) {
	sourceFileKind := detectSourceFileKind(envelope)
	units, apiErr := s.discoverImportUnits(envelope, sourceFileKind)
	if apiErr != nil {
		return CreateAcceptedSessionResult{}, nil, apiErr
	}
	normalized, err := json.Marshal(map[string]any{
		"incident_id":           request.IncidentID.String(),
		"client_txn_id":         request.ClientTxnID,
		"assistant_profile":     request.AssistantProfile,
		"source_content_sha256": envelope.FileSHA256Hex,
	})
	if err != nil {
		return CreateAcceptedSessionResult{}, err, nil
	}
	result, err := s.store.CreateAcceptedSession(ctx, CreateAcceptedSessionParams{
		ActorUserID:         actorUserID,
		Request:             request,
		SourceFileKind:      sourceFileKind,
		OriginalFilename:    envelope.FileName,
		SourceContentSHA256: envelope.FileSHA256Hex,
		SourceMediaType:     envelope.FileContentType,
		SourceByteSize:      int64(len(envelope.File)),
		SourceBytes:         envelope.File,
		Units:               units,
		NormalizedRequest:   normalized,
		Now:                 s.now(),
	})
	if err == nil && !result.Replayed {
		s.jobRunner.Notify(uuid.MustParse(result.Job.JobID))
	}
	return result, err, nil
}

func (s *Service) applicationGetSession(ctx context.Context, sessionID uuid.UUID) (importScopedValue[map[string]any], error) {
	value, incidentID, err := s.store.GetSession(ctx, sessionID)
	return importScopedValue[map[string]any]{Value: value, IncidentID: incidentID}, err
}

func (s *Service) applicationListUnits(ctx context.Context, sessionID uuid.UUID) (importScopedValue[[]map[string]any], error) {
	value, incidentID, err := s.store.ListUnits(ctx, sessionID)
	return importScopedValue[[]map[string]any]{Value: value, IncidentID: incidentID}, err
}

func (s *Service) applicationGetUnit(ctx context.Context, sessionID uuid.UUID, unitID uuid.UUID) (importScopedValue[map[string]any], error) {
	value, incidentID, err := s.store.GetUnit(ctx, sessionID, unitID)
	return importScopedValue[map[string]any]{Value: value, IncidentID: incidentID}, err
}

func (s *Service) applicationGetPreview(ctx context.Context, sessionID uuid.UUID, unitID uuid.UUID) (importScopedValue[map[string]any], error) {
	value, incidentID, err := s.store.GetPreview(ctx, sessionID, unitID)
	return importScopedValue[map[string]any]{Value: value, IncidentID: incidentID}, err
}

func (s *Service) applicationGetMappingContext(ctx context.Context, sessionID uuid.UUID, unitID uuid.UUID) (importScopedValue[[]map[string]any], error) {
	value, incidentID, err := s.store.GetUnitColumns(ctx, sessionID, unitID)
	return importScopedValue[[]map[string]any]{Value: value, IncidentID: incidentID}, err
}

func (s *Service) applicationPrepareMappingPreview(
	ctx context.Context,
	actorUserID uuid.UUID,
	incidentID uuid.UUID,
	route importSessionRoute,
	request MappingPreviewRequest,
) (ExtensionMappingPreviewResource, *httpapi.APIError) {
	mapping := ApprovedMapping{
		TargetKind:           request.TargetKind,
		ExtensionProfileID:   request.ExtensionProfileID,
		OwnerMappingSchemaID: request.OwnerMappingSchemaID,
		OwnerMapping:         append(json.RawMessage(nil), request.OwnerMapping...),
	}
	if apiErr := s.validateApprovedMapping(mapping); apiErr != nil {
		return ExtensionMappingPreviewResource{}, apiErr
	}
	target, ok := lookupApprovedImportTarget(mapping)
	if !ok {
		return ExtensionMappingPreviewResource{}, invalidImportRequest("target_kind", "target_kind_not_importable")
	}
	facade := s.extensionImportFacades[extensionImportFacadeKey(target)]
	if facade == nil {
		return ExtensionMappingPreviewResource{}, invalidImportRequest("target_kind", "owner_preview_contract_unavailable")
	}
	sourceCapability, err := s.store.SourceCapabilityForUnit(ctx, route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		return ExtensionMappingPreviewResource{}, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}}
	}
	if err != nil {
		return ExtensionMappingPreviewResource{}, internalAPIError(err)
	}
	result, err := facade.PrepareImportUnitMapping(ctx, ExtensionImportMappingRequest{
		IncidentID:           incidentID,
		ActorUserID:          actorUserID,
		TargetKind:           request.TargetKind,
		ExtensionProfileID:   request.ExtensionProfileID,
		ImportSessionID:      route.SessionID,
		ImportUnitID:         route.UnitID,
		SourceCapability:     sourceCapability,
		OwnerMappingSchemaID: request.OwnerMappingSchemaID,
		OwnerMapping:         append(json.RawMessage(nil), request.OwnerMapping...),
	})
	if err != nil {
		return ExtensionMappingPreviewResource{}, extensionFacadeAPIError(target, facade, err)
	}
	if err := facade.ValidateImportUnitMappingResult(result); err != nil {
		return ExtensionMappingPreviewResource{}, invalidImportRequest("owner_result", "owner_preview_validation_failed")
	}
	return ExtensionMappingPreviewResource{
		SchemaID:            ExtensionMappingPreviewResultSchemaID,
		ImportSessionID:     route.SessionID.String(),
		ImportUnitID:        route.UnitID.String(),
		TargetKind:          request.TargetKind,
		ExtensionProfileID:  request.ExtensionProfileID,
		OwnerResultSchemaID: result.OwnerResultSchemaID,
		OwnerResult:         result.OwnerResult,
	}, nil
}

func (s *Service) applicationExecuteUnitAction(
	ctx context.Context,
	actorUserID uuid.UUID,
	route importSessionRoute,
	request ActionRequest,
	routeKey string,
) (UnitActionResult, error) {
	params := UnitActionParams{
		ActorUserID:       actorUserID,
		SessionID:         route.SessionID,
		UnitID:            route.UnitID,
		RouteKey:          routeKey,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	}
	if routeKey == "imports.units.select" {
		return s.store.SelectUnit(ctx, params)
	}
	return s.store.SkipUnit(ctx, params)
}

func (s *Service) applicationStartApply(
	ctx context.Context,
	actorUserID uuid.UUID,
	sessionID uuid.UUID,
	request ApplyRequest,
) (ApplyStartResult, error) {
	result, err := s.store.StartApply(ctx, ApplyStartParams{
		ActorUserID:       actorUserID,
		SessionID:         sessionID,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if err == nil && !result.Replayed {
		s.jobRunner.Notify(uuid.MustParse(result.Job.JobID))
	}
	return result, err
}

func (s *Service) applicationCreateOperatorRegion(
	ctx context.Context,
	actorUserID uuid.UUID,
	route importSessionRoute,
	request RegionRequest,
) (CreateOperatorRegionResult, error, *httpapi.APIError) {
	base, err := s.store.GetRegionBase(ctx, route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		return CreateOperatorRegionResult{}, nil, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}}
	}
	if errors.Is(err, ErrInvalidSourceRect) {
		return CreateOperatorRegionResult{}, nil, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	if err != nil {
		return CreateOperatorRegionResult{}, err, nil
	}
	requested := request.SourceRect.sourceRectangle()
	baseRect, rectErr := parseSourceRectangle(base.Unit["source_rect_a1"].(string))
	if rectErr != nil || !sourceRectangleContains(baseRect, requested) ||
		!operatorRegionWithinLimits(requested, s.limits) {
		return CreateOperatorRegionResult{}, nil, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	locator, ok := base.Unit["locator"].(map[string]any)
	if !ok {
		return CreateOperatorRegionResult{}, nil, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	sheetName, ok := locator["sheet_name"].(string)
	if !ok || strings.TrimSpace(sheetName) == "" {
		return CreateOperatorRegionResult{}, nil, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	workbook, apiErr := indexXLSXWorkbook(base.SourceBytes, s.limits, s.archiveLimits)
	if apiErr != nil {
		return CreateOperatorRegionResult{}, nil, apiErr
	}
	sheet := workbook.sheetsByName[sheetName]
	decoded, apiErr := workbook.decodeRectangle(sheet, requested, s.limits)
	if apiErr != nil {
		return CreateOperatorRegionResult{}, nil, apiErr
	}
	locatorJSON, err := json.Marshal(map[string]any{
		"sheet_name": sheetName,
		"rect_a1":    sourceRectangleA1(requested),
	})
	if err != nil {
		return CreateOperatorRegionResult{}, err, nil
	}
	unit, apiErr := s.discoveredImportUnitAt(
		decoded.rows,
		"operator_region",
		string(locatorJSON),
		requested,
		decoded.warningCodes,
		decoded.blockingColumnOrdinals,
	)
	if apiErr != nil {
		return CreateOperatorRegionResult{}, nil, apiErr
	}
	result, err := s.store.CreateOperatorRegion(ctx, CreateOperatorRegionParams{
		ActorUserID:                 actorUserID,
		SessionID:                   route.SessionID,
		BaseUnitID:                  route.UnitID,
		Request:                     request,
		Unit:                        unit,
		ExpectedSourceContentSHA256: base.SourceContentSHA256,
		Now:                         s.now(),
	})
	if errors.Is(err, ErrInvalidSourceRect) {
		return CreateOperatorRegionResult{}, nil, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	return result, err, nil
}
