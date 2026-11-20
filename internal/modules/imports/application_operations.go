package imports

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
)

type importScopedValue[T any] struct {
	Value      T
	IncidentID uuid.UUID
}

func (s *service) applicationCreateSession(
	ctx context.Context,
	actorUserID uuid.UUID,
	envelope httpapi.UploadEnvelope,
	request createSessionRequest,
) (createAcceptedSessionResult, error, *httpapi.APIError) {
	sourceFileKind := detectSourceFileKind(envelope)
	units, apiErr := s.discoverImportUnits(envelope, sourceFileKind)
	if apiErr != nil {
		return createAcceptedSessionResult{}, nil, apiErr
	}
	normalized, err := json.Marshal(map[string]any{
		"incident_id":           request.IncidentID.String(),
		"client_txn_id":         request.ClientTxnID,
		"assistant_profile":     request.AssistantProfile,
		"source_content_sha256": envelope.FileSHA256Hex,
	})
	if err != nil {
		return createAcceptedSessionResult{}, err, nil
	}
	result, err := s.store.createAcceptedSession(ctx, createAcceptedSessionParams{
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

func (s *service) applicationSessionIncident(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	return s.store.getSessionIncident(ctx, sessionID)
}

func (s *service) applicationGetSession(ctx context.Context, sessionID uuid.UUID) (importScopedValue[map[string]any], error) {
	value, incidentID, err := s.store.getSession(ctx, sessionID)
	return importScopedValue[map[string]any]{Value: value, IncidentID: incidentID}, err
}

func (s *service) applicationListUnits(ctx context.Context, sessionID uuid.UUID) (importScopedValue[[]map[string]any], error) {
	value, incidentID, err := s.store.listUnits(ctx, sessionID)
	return importScopedValue[[]map[string]any]{Value: value, IncidentID: incidentID}, err
}

func (s *service) applicationGetUnit(ctx context.Context, sessionID uuid.UUID, unitID uuid.UUID) (importScopedValue[map[string]any], error) {
	value, incidentID, err := s.store.getUnit(ctx, sessionID, unitID)
	return importScopedValue[map[string]any]{Value: value, IncidentID: incidentID}, err
}

func (s *service) applicationPrepareMappingPreview(
	ctx context.Context,
	actorUserID uuid.UUID,
	incidentID uuid.UUID,
	route importSessionRoute,
	request mappingPreviewRequest,
) (extensionImportReadResource, *httpapi.APIError) {
	mapping := approvedMapping{
		TargetKind:           request.TargetKind,
		ExtensionProfileID:   request.ExtensionProfileID,
		OwnerMappingSchemaID: request.OwnerMappingSchemaID,
		OwnerMapping:         append(json.RawMessage(nil), request.OwnerMapping...),
	}
	if apiErr := s.validateApprovedMapping(mapping); apiErr != nil {
		return extensionImportReadResource{}, apiErr
	}
	target, ok := lookupApprovedImportTarget(mapping)
	if !ok {
		return extensionImportReadResource{}, invalidImportRequest("target_kind", "target_kind_not_importable")
	}
	facade := s.extensionImportFacades[analyticalImportTargetKey{
		TargetKind:         target.TargetKind,
		ExtensionProfileID: target.ExtensionProfileID,
	}]
	if facade == nil {
		return extensionImportReadResource{}, invalidImportRequest("target_kind", "owner_preview_contract_unavailable")
	}
	sourceCapability, err := s.store.sourceCapabilityForUnit(ctx, route.SessionID, route.UnitID)
	if err != nil {
		return extensionImportReadResource{}, unitSourceAPIError(err)
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
		return extensionImportReadResource{}, extensionFacadeAPIError(target, facade, err)
	}
	if err := facade.ValidateImportUnitMappingResult(result); err != nil {
		return extensionImportReadResource{}, invalidImportRequest("owner_result", "owner_preview_validation_failed")
	}
	return extensionImportReadResource{
		SchemaID:            extensionMappingPreviewResultSchemaID,
		ImportSessionID:     route.SessionID.String(),
		ImportUnitID:        route.UnitID.String(),
		TargetKind:          request.TargetKind,
		ExtensionProfileID:  request.ExtensionProfileID,
		OwnerResultSchemaID: result.OwnerResultSchemaID,
		OwnerResult:         result.OwnerResult,
	}, nil
}

func (s *service) applicationExecuteUnitAction(
	ctx context.Context,
	actorUserID uuid.UUID,
	route importSessionRoute,
	request actionRequest,
	routeKey string,
) (unitActionResult, error) {
	params := unitActionParams{
		ActorUserID:       actorUserID,
		SessionID:         route.SessionID,
		UnitID:            route.UnitID,
		RouteKey:          routeKey,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	}
	if routeKey == "imports.units.select" {
		return s.store.selectUnit(ctx, params)
	}
	return s.store.skipUnit(ctx, params)
}

func (s *service) applicationStartApply(
	ctx context.Context,
	actorUserID uuid.UUID,
	sessionID uuid.UUID,
	request applyRequest,
) (applyStartResult, error) {
	result, err := s.store.startApply(ctx, applyStartParams{
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

func (s *service) applicationCreateOperatorRegion(
	ctx context.Context,
	actorUserID uuid.UUID,
	route importSessionRoute,
	request regionRequest,
) (createOperatorRegionResult, error, *httpapi.APIError) {
	base, err := s.store.getRegionBase(ctx, route.SessionID, route.UnitID)
	if errors.Is(err, errSessionNotFound) {
		return createOperatorRegionResult{}, nil, httpapi.NotFound("import_session_not_found")
	}
	if errors.Is(err, errNotFound) {
		return createOperatorRegionResult{}, nil, httpapi.NotFound("import_unit_not_found")
	}
	if errors.Is(err, errInvalidSourceRect) {
		return createOperatorRegionResult{}, nil, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	if err != nil {
		// A retained base source whose bytes no longer match its recorded
		// identity or digest is no typed region-creation condition
		// (REQ-01-620e); like any other base-source failure it is
		// internal_error (REQ-01-706), never import_apply_blocked.
		return createOperatorRegionResult{}, nil, internalAPIError(err)
	}
	requested := request.SourceRect.sourceRectangle()
	baseRect, rectErr := parseSourceRectangle(base.Unit["source_rect_a1"].(string))
	if rectErr != nil || !sourceRectangleContains(baseRect, requested) ||
		!operatorRegionWithinLimits(requested, s.limits) {
		return createOperatorRegionResult{}, nil, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	locator, ok := base.Unit["locator"].(map[string]any)
	if !ok {
		return createOperatorRegionResult{}, nil, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	sheetName, ok := locator["sheet_name"].(string)
	if !ok || strings.TrimSpace(sheetName) == "" {
		return createOperatorRegionResult{}, nil, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	workbook, apiErr := indexXLSXWorkbook(base.SourceBytes, s.limits, s.archiveLimits)
	if apiErr != nil {
		return createOperatorRegionResult{}, nil, regionSourceAPIError(apiErr)
	}
	sheet := workbook.sheetsByName[sheetName]
	decoded, apiErr := workbook.decodeRectangle(sheet, requested, s.limits)
	if apiErr != nil {
		return createOperatorRegionResult{}, nil, regionSourceAPIError(apiErr)
	}
	locatorJSON, err := json.Marshal(map[string]any{
		"sheet_name": sheetName,
		"rect_a1":    sourceRectangleA1(requested),
	})
	if err != nil {
		return createOperatorRegionResult{}, err, nil
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
		return createOperatorRegionResult{}, nil, regionSourceAPIError(apiErr)
	}
	result, err := s.store.createOperatorRegion(ctx, createOperatorRegionParams{
		ActorUserID:                 actorUserID,
		SessionID:                   route.SessionID,
		BaseUnitID:                  route.UnitID,
		Request:                     request,
		Unit:                        unit,
		ExpectedSourceContentSHA256: base.SourceContentSHA256,
		Now:                         s.now(),
	})
	if errors.Is(err, errInvalidSourceRect) {
		return createOperatorRegionResult{}, nil, invalidImportRequest("source_rect", "invalid_source_rect")
	}
	if apiErr := regionAdmissionAPIError(err); apiErr != nil {
		return createOperatorRegionResult{}, nil, apiErr
	}
	return result, err, nil
}

// regionSourceAPIError classifies a failure to re-read the base unit's admitted
// source during region creation. The current archive and tabular limits still
// apply, so a limit failure keeps its 413 import_source_rejected (REQ-01-620e).
// Discovery already parsed the same bytes and decoded the whole base range, so
// any other failure, such as an unparseable workbook or cell, is an invariant
// break and internal_error (REQ-01-706).
func regionSourceAPIError(apiErr *httpapi.APIError) *httpapi.APIError {
	if apiErr != nil && apiErr.Code == "import_source_rejected" {
		return apiErr
	}
	return internalAPIError(errors.New("imports: region base source could not be re-read"))
}
