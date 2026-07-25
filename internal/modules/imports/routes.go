package imports

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/google/uuid"
)

type Service struct {
	store                    *Store
	incidentAccess           incidents.Access
	authStore                *authn.Store
	jobManager               *jobs.Manager
	jobRunner                *jobs.Runner
	timelineStore            *timeline.Facade
	hub                      *platformws.Hub
	keys                     authn.MasterKeys
	cursorCodec              *pagination.Codec
	limits                   Limits
	archiveLimits            ArchiveLimits
	extensionImportFacades   map[string]ExtensionImportFacade
	extensionProfileAdmitted func(string) bool
	jobSuccessFinalizer      JobSuccessFinalizer
	now                      func() time.Time
}

type RouteOption func(*routeOptions)

type routeOptions struct {
	extensionProfileAdmitted func(string) bool
	jobSuccessFinalizer      JobSuccessFinalizer
	limits                   Limits
	archiveLimits            ArchiveLimits
}

func WithLimits(limits Limits, archiveLimits ArchiveLimits) RouteOption {
	return func(options *routeOptions) {
		options.limits = limits
		options.archiveLimits = archiveLimits
	}
}

func WithExtensionProfileAdmission(admitted func(string) bool) RouteOption {
	return func(options *routeOptions) {
		options.extensionProfileAdmitted = admitted
	}
}

func WithJobSuccessFinalizer(finalizer JobSuccessFinalizer) RouteOption {
	return func(options *routeOptions) {
		options.jobSuccessFinalizer = finalizer
	}
}

func RegisterRoutes(options ...RouteOption) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		settings := routeOptions{}
		for _, option := range options {
			if option != nil {
				option(&settings)
			}
		}
		service, err := newService(deps, settings)
		if err != nil {
			return err
		}
		mux.HandleFunc("/api/v1/import-sessions", service.handleImportSessionsCollection)
		mux.HandleFunc("/api/v1/import-sessions/", service.handleImportSessionsMember)
		return nil
	}
}

func newService(deps httpapi.DependencySet, options routeOptions) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	cursorCodec := deps.CursorCodec
	if cursorCodec == nil {
		cursorKey := authn.DerivePurposeKey(keys, "pagination-cursor-v1")
		cursorCodec = pagination.NewCodec(cursorKey[:])
	}
	timelineStore := timeline.FacadeFromDependencies(deps)
	extensionImportFacades, err := extensionImportFacadesFromDependencies(deps)
	if err != nil {
		return nil, err
	}
	if deps.Jobs != nil && options.jobSuccessFinalizer == nil {
		return nil, fmt.Errorf("Import admitted route requires a job success finalizer")
	}
	extensionProfileAdmitted := options.extensionProfileAdmitted
	if extensionProfileAdmitted == nil {
		extensionProfileAdmitted = func(string) bool { return false }
	}
	service := &Service{
		store:                    NewStore(deps.Postgres),
		incidentAccess:           incidents.NewAccess(deps.PostgresHandle()),
		authStore:                authn.NewStore(deps.PostgresHandle()),
		jobManager:               deps.Jobs,
		jobRunner:                deps.JobRunner,
		timelineStore:            timelineStore,
		hub:                      deps.WSHub,
		keys:                     keys,
		cursorCodec:              cursorCodec,
		limits:                   options.limits,
		archiveLimits:            options.archiveLimits,
		extensionImportFacades:   extensionImportFacades,
		extensionProfileAdmitted: extensionProfileAdmitted,
		jobSuccessFinalizer:      options.jobSuccessFinalizer,
		now:                      now,
	}
	if err := service.registerJobHandlers(); err != nil {
		return nil, err
	}
	if err := service.recoverImportJobs(context.Background()); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *Service) registerJobHandlers() error {
	if s == nil || s.jobRunner == nil {
		return nil
	}
	if err := s.jobRunner.RegisterHandler(importDiscoveryJobHandlerName, s.executeDiscoveryJob); err != nil && !errors.Is(err, jobs.ErrHandlerAlreadyRegistered) {
		return err
	}
	if err := s.jobRunner.RegisterHandler(importApplyJobHandlerName, s.executeApplyJob); err != nil && !errors.Is(err, jobs.ErrHandlerAlreadyRegistered) {
		return err
	}
	return nil
}

func (s *Service) recoverImportJobs(ctx context.Context) error {
	if s == nil || s.jobRunner == nil {
		return nil
	}
	if err := s.jobRunner.RecoverHandler(ctx, importDiscoveryJobHandlerName); err != nil {
		return err
	}
	return s.jobRunner.RecoverHandler(ctx, importApplyJobHandlerName)
}

func (s *Service) handleImportSessionsCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	envelope, envelopeErr := httpapi.ParseUploadEnvelope(r, httpapi.UploadEnvelopePolicy{FileContentTypes: ImportSessionFileContentTypes})
	if envelopeErr != nil {
		writeAPIError(w, r, uploadEnvelopeAPIError(envelopeErr))
		return
	}
	request, apiErr := DecodeCreateSessionMetadata(envelope)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), request.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	sourceFileKind := detectSourceFileKind(envelope)
	units, apiErr := s.discoverImportUnits(envelope, sourceFileKind)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	normalized, err := json.Marshal(map[string]any{
		"incident_id":           request.IncidentID.String(),
		"client_txn_id":         request.ClientTxnID,
		"assistant_profile":     request.AssistantProfile,
		"source_content_sha256": envelope.FileSHA256Hex,
	})
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := s.store.CreateAcceptedSession(r.Context(), CreateAcceptedSessionParams{
		ActorUserID:         principal.User.ID,
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
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !result.Replayed {
		if err := s.executeDiscoveryJob(r.Context(), uuid.MustParse(result.Job.JobID)); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *Service) handleImportSessionsMember(w http.ResponseWriter, r *http.Request) {
	route, ok := parseImportSessionPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch route.Kind {
	case "session":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		resource, incidentID, err := s.store.GetSession(r.Context(), route.SessionID)
		if errors.Is(err, ErrNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
	case "units":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		binding, cursor, reasonCode := s.cursorCodec.ResolveRequest(
			r.URL.Query(),
			"imports.units.list",
			principal.User.ID.String(),
			map[string]string{"import_session_id": route.SessionID.String()},
		)
		if reasonCode != "" {
			writeAPIError(w, r, invalidPaginationRequest(reasonCode))
			return
		}
		units, incidentID, err := s.store.ListUnits(r.Context(), route.SessionID)
		if errors.Is(err, ErrNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		rows, nextCursor, err := pagination.PageResources(binding, cursor, units)
		if errors.Is(err, pagination.ErrInvalidCursorToken) {
			writeAPIError(w, r, invalidPaginationRequest(pagination.ReasonInvalidCursorToken))
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		var nextToken *string
		if nextCursor != nil {
			token, err := s.cursorCodec.Encode(*nextCursor)
			if err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			nextToken = &token
		}
		_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"import_units": rows}, httpapi.PagingMeta{
			Limit:      binding.Limit,
			HasMore:    nextToken != nil,
			NextCursor: nextToken,
		})
	case "unit":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		unit, incidentID, err := s.store.GetUnit(r.Context(), route.SessionID, route.UnitID)
		if errors.Is(err, ErrNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, unit)
	case "preview":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		preview, incidentID, err := s.store.GetPreview(r.Context(), route.SessionID, route.UnitID)
		if errors.Is(err, ErrNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, preview)
	case "mapping_preview":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleMappingPreview(w, r, principal, route)
	case "mapping":
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleMapping(w, r, principal, route)
	case "select":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleSelect(w, r, principal, route)
	case "skip":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleSkip(w, r, principal, route)
	case "apply":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleApply(w, r, principal, route)
	default:
		http.NotFound(w, r)
	}
}

func (s *Service) handleMappingPreview(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, incidentID, err := s.store.GetUnit(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeMappingPreviewRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	mapping := ApprovedMapping{
		TargetKind:           request.TargetKind,
		ExtensionProfileID:   request.ExtensionProfileID,
		OwnerMappingSchemaID: request.OwnerMappingSchemaID,
		OwnerMapping:         append(json.RawMessage(nil), request.OwnerMapping...),
	}
	if apiErr := s.validateApprovedMapping(mapping); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	target, ok := lookupApprovedImportTarget(mapping)
	if !ok {
		writeAPIError(w, r, invalidImportRequest("target_kind", "target_kind_not_importable"))
		return
	}
	facade := s.extensionImportFacades[extensionImportFacadeKey(target)]
	if facade == nil {
		writeAPIError(w, r, invalidImportRequest("target_kind", "owner_apply_contract_unavailable"))
		return
	}
	sourceCapability, err := s.store.SourceCapabilityForUnit(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := facade.PrepareImportUnitMapping(r.Context(), ExtensionImportMappingRequest{
		IncidentID:           incidentID,
		ActorUserID:          principal.User.ID,
		TargetKind:           request.TargetKind,
		ExtensionProfileID:   request.ExtensionProfileID,
		ImportSessionID:      route.SessionID,
		ImportUnitID:         route.UnitID,
		SourceCapability:     sourceCapability,
		OwnerMappingSchemaID: request.OwnerMappingSchemaID,
		OwnerMapping:         append(json.RawMessage(nil), request.OwnerMapping...),
	})
	if err != nil {
		writeAPIError(w, r, extensionFacadeAPIError(err))
		return
	}
	if err := facade.ValidateImportUnitMappingResult(result); err != nil {
		writeAPIError(w, r, internalAPIError(fmt.Errorf("extension mapping preview result validation failed: %w", err)))
		return
	}
	resource := ExtensionMappingPreviewResource{
		SchemaID:            ExtensionMappingPreviewResultSchemaID,
		ImportSessionID:     route.SessionID.String(),
		ImportUnitID:        route.UnitID.String(),
		TargetKind:          request.TargetKind,
		ExtensionProfileID:  request.ExtensionProfileID,
		OwnerResultSchemaID: result.OwnerResultSchemaID,
		OwnerResult:         result.OwnerResult,
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleMapping(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	columns, incidentID, err := s.store.GetUnitColumns(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeMappingRequest(r.Body, columns)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	materialized, apiErr := s.prepareApprovedMapping(r.Context(), principal.User.ID, incidentID, route, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request = materialized
	unit, _, err := s.store.SaveMapping(r.Context(), MappingParams{
		ActorUserID:       principal.User.ID,
		SessionID:         route.SessionID,
		UnitID:            route.UnitID,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, unit)
}

func (s *Service) handleSelect(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	s.handleUnitAction(w, r, principal, route, "imports.units.select", false, s.store.SelectUnit)
}

func (s *Service) handleSkip(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	s.handleUnitAction(w, r, principal, route, "imports.units.skip", true, s.store.SkipUnit)
}

func (s *Service) handleUnitAction(
	w http.ResponseWriter,
	r *http.Request,
	principal httpauth.Principal,
	route importSessionRoute,
	routeKey string,
	allowReason bool,
	action func(context.Context, UnitActionParams) (UnitActionResult, error),
) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, incidentID, err := s.store.GetUnit(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeActionRequest(r.Body, allowReason)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := action(r.Context(), UnitActionParams{
		ActorUserID:       principal.User.ID,
		SessionID:         route.SessionID,
		UnitID:            route.UnitID,
		RouteKey:          routeKey,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) handleApply(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	session, incidentID, err := s.store.GetSession(r.Context(), route.SessionID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeApplyRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_ = session
	result, err := s.store.StartApply(r.Context(), ApplyStartParams{
		ActorUserID:       principal.User.ID,
		SessionID:         route.SessionID,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if !result.Replayed {
		if err := s.executeApplyJob(r.Context(), uuid.MustParse(result.Job.JobID)); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

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

func (s *Service) executeApplyJob(ctx context.Context, jobID uuid.UUID) error {
	var payload applyJobHandlerPayload
	if err := s.decodeJobPayload(ctx, jobID, &payload); err != nil {
		return err
	}
	incidentID, err := uuid.Parse(payload.IncidentID)
	if err != nil {
		return err
	}
	sessionID, err := uuid.Parse(payload.ImportSessionID)
	if err != nil {
		return err
	}
	actorID, err := uuid.Parse(payload.ActorUserID)
	if err != nil {
		return err
	}
	selected, err := parseUUIDStrings(payload.SelectedUnitIDs)
	if err != nil {
		return err
	}
	actor, err := s.authStore.GetUserByID(ctx, actorID)
	if err != nil {
		return err
	}
	job, err := s.jobManager.Get(ctx, jobID)
	if err != nil {
		return err
	}
	return s.completeApplyJob(ctx, actor, ApplyStartResult{
		Job:             job,
		IncidentID:      incidentID,
		ImportSessionID: sessionID,
		ClientTxnID:     payload.ClientTxnID,
		SelectedUnitIDs: selected,
	})
}

func (s *Service) completeApplyJob(ctx context.Context, actor authn.UserRecord, start ApplyStartResult) error {
	jobID, err := uuid.Parse(start.Job.JobID)
	if err != nil {
		return err
	}
	total := len(start.SelectedUnitIDs)
	if !s.markJobRunningOrResume(ctx, jobID, total) {
		s.cancelApplySessionForTerminalJob(ctx, jobID, start)
		return nil
	}
	units, err := s.store.GetApplyUnits(ctx, start.ImportSessionID, start.SelectedUnitIDs)
	if err != nil {
		s.failApplyJob(ctx, jobID, start, "import_apply_failed", err)
		return nil
	}
	extensionRefs := make([]jobs.ResourceRef, 0)
	completed := 0
	failed := 0
	seenUnits := make(map[uuid.UUID]struct{}, len(units))
	for _, unit := range units {
		seenUnits[unit.UnitID] = struct{}{}
		if s.jobCancelRequested(ctx, jobID) {
			if err := s.store.CancelApply(ctx, start.ImportSessionID, start.SelectedUnitIDs, s.now()); err != nil {
				return err
			}
			_, err := s.jobManager.CompleteCanceled(ctx, jobs.TransitionParams{
				JobID:    jobID,
				Progress: jobs.Progress{Completed: completed, Total: &total},
			})
			return err
		}
		refs, err := s.applyUnit(ctx, actor, start, unit)
		if err != nil {
			if statusErr := s.store.MarkApplyUnitStatus(ctx, start.ImportSessionID, unit.UnitID, "failed", s.now()); statusErr != nil {
				s.failApplyJob(ctx, jobID, start, "import_apply_failed", statusErr)
				return nil
			}
			failed++
			completed++
			continue
		}
		if err := s.store.MarkApplyUnitStatus(ctx, start.ImportSessionID, unit.UnitID, "applied", s.now()); err != nil {
			s.failApplyJob(ctx, jobID, start, "import_apply_failed", err)
			return nil
		}
		extensionRefs = append(extensionRefs, refs...)
		completed++
	}
	for _, unitID := range start.SelectedUnitIDs {
		if _, ok := seenUnits[unitID]; ok {
			continue
		}
		if err := s.store.MarkApplyUnitStatus(ctx, start.ImportSessionID, unitID, "failed", s.now()); err != nil {
			s.failApplyJob(ctx, jobID, start, "import_apply_failed", err)
			return nil
		}
		failed++
		completed++
	}
	if completed-failed == 0 {
		s.failApplyJob(ctx, jobID, start, "import_apply_failed", fmt.Errorf("all selected import units failed"))
		return nil
	}
	status := "applied"
	code := "import_session_applied"
	if failed > 0 {
		status = "partially_applied"
		code = "import_session_partially_applied"
	}
	_, err = s.jobSuccessFinalizer.FinalizeImportJobSuccess(ctx, JobSuccessFinalization{
		FinalCommitID: "import.apply:" + jobID.String(),
		Transition: jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: total, Total: &total},
			ResultSummary: &jobs.ResultSummary{
				Code:         code,
				Message:      "Import session applied.",
				ResourceRefs: importApplyResourceRefs(start.ImportSessionID, extensionRefs),
			},
		},
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			return s.store.CompleteApplyTx(ctx, tx, start.ImportSessionID, start.SelectedUnitIDs, status, s.now())
		},
	})
	return err
}

func (s *Service) failApplyJob(ctx context.Context, jobID uuid.UUID, start ApplyStartResult, code string, err error) {
	_ = s.store.FailApply(ctx, start.ImportSessionID, start.SelectedUnitIDs, s.now())
	_, _ = s.jobManager.CompleteFailed(ctx, jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 0, Total: intPtr(len(start.SelectedUnitIDs))},
		ErrorSummary: &jobs.ErrorSummary{
			Code:      code,
			Message:   "Import apply failed.",
			Retryable: false,
			Details:   map[string]any{},
		},
	})
	_ = err
}

func (s *Service) cancelApplySessionForTerminalJob(ctx context.Context, jobID uuid.UUID, start ApplyStartResult) {
	job, err := s.jobManager.Get(ctx, jobID)
	if err == nil && job.Status == jobs.StatusCanceled {
		_ = s.store.CancelApply(ctx, start.ImportSessionID, start.SelectedUnitIDs, s.now())
	}
}

func (s *Service) jobCancelRequested(ctx context.Context, jobID uuid.UUID) bool {
	job, err := s.jobManager.Get(ctx, jobID)
	return err == nil && job.Status == jobs.StatusCancelRequested
}

func (s *Service) applyUnit(ctx context.Context, actor authn.UserRecord, start ApplyStartResult, unit ApplyUnitData) ([]jobs.ResourceRef, error) {
	target, ok := lookupApprovedImportTarget(unit.ApprovedMapping)
	if !ok || !target.importable(s.extensionProfileClaimed) {
		if unit.ApprovedMapping.targetKindOrDefault() == ImportTargetKindViewSchema {
			return nil, importApplyBlockedError("target_view_schema_not_importable")
		}
		return nil, importApplyBlockedError("target_kind_not_importable")
	}
	if unit.ApprovedMapping.targetKindOrDefault() != ImportTargetKindViewSchema {
		if !target.ownerApplyFacadeAvailable() {
			return nil, importApplyBlockedError("owner_apply_contract_unavailable")
		}
		return s.applyExtensionOwnerUnit(ctx, actor, start, unit, target)
	}
	if !target.ownerCreateFacadeAvailable() {
		return nil, importApplyBlockedError("owner_create_contract_unavailable")
	}
	switch unit.ApprovedMapping.TargetViewSchemaID {
	case timeline.TimelineViewSchemaID:
	default:
		return nil, s.applyGenericOwnerUnit(ctx, actor, start, unit, target)
	}
	for _, sourceRow := range unit.SourceRows {
		rowRef, _ := intFromAny(sourceRow["source_row_ref"])
		clientTxnID := fmt.Sprintf("import:%s:%s:%d:%s", start.ImportSessionID, unit.UnitID, rowRef, start.ClientTxnID)
		payload, err := importRowPayload(unit.ApprovedMapping, sourceRow, clientTxnID)
		if err != nil {
			return nil, err
		}
		request, apiErr := timeline.DecodeTimelineCreateRequest(bytes.NewReader(payload))
		if apiErr != nil {
			return nil, fmt.Errorf("decode imported timeline row: %s", apiErr.Code)
		}
		request.RawCaptureColumns = importRawCaptureColumns(start, unit, sourceRow, rowRef)
		if _, err := s.timelineStore.CreateImportedRow(ctx, timeline.CreateRowCommand{
			Actor:      actor,
			IncidentID: start.IncidentID,
			Request:    request,
			RequestID:  "req-" + clientTxnID,
			Now:        s.now(),
		}); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func importApplyResourceRefs(sessionID uuid.UUID, extensionRefs []jobs.ResourceRef) []jobs.ResourceRef {
	refs := make([]jobs.ResourceRef, 0, 1+len(extensionRefs))
	refs = append(refs, jobs.ResourceRef{
		Kind:  "import_session",
		ID:    sessionID.String(),
		Route: "/api/v1/import-sessions/" + sessionID.String(),
	})
	if len(extensionRefs) == 0 {
		return refs
	}
	sorted := append([]jobs.ResourceRef(nil), extensionRefs...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Kind != sorted[j].Kind {
			return sorted[i].Kind < sorted[j].Kind
		}
		if sorted[i].Route != sorted[j].Route {
			return sorted[i].Route < sorted[j].Route
		}
		return sorted[i].ID < sorted[j].ID
	})
	return append(refs, sorted...)
}

func importRawCaptureColumns(start ApplyStartResult, unit ApplyUnitData, sourceRow map[string]any, rowRef int) []timeline.ClipboardRawImportColumn {
	if unit.ApprovedMapping.TargetViewSchemaID != timeline.TimelineViewSchemaID || unit.ApprovedMapping.UnknownColumnPolicy != "preserve_raw_capture" {
		return nil
	}
	cells := sourceRowCellsByOrdinal(sourceRow)
	columns := make([]timeline.ClipboardRawImportColumn, 0)
	for _, column := range unit.ApprovedMapping.SourceColumns {
		if column.FieldKey != nil {
			continue
		}
		cell := cells[column.SourceColumnOrdinal]
		rawValue, _ := cell["display_text"].(string)
		cellKind, _ := cell["cell_kind"].(string)
		columns = append(columns, timeline.ClipboardRawImportColumn{
			SourceKind:          "file_import",
			ImportSessionID:     start.ImportSessionID.String(),
			ImportUnitID:        unit.UnitID.String(),
			MappingFingerprint:  unit.MappingFingerprint,
			SourceFileKind:      unit.SourceFileKind,
			SourceContentSHA256: unit.SourceContentSHA256,
			ParserProfileID:     unit.ParserProfileID,
			ParserVersion:       unit.ParserVersion,
			LocatorKind:         unit.LocatorKind,
			Locator:             unit.Locator,
			SourceRectA1:        unit.SourceRectA1,
			SourceRowOrdinal:    rowRef,
			SourceColumnOrdinal: column.SourceColumnOrdinal,
			SourceHeaderText:    column.SourceHeaderText,
			RawValue:            rawValue,
			CellKind:            cellKind,
		})
	}
	return columns
}

func importRowPayload(mapping ApprovedMapping, sourceRow map[string]any, clientTxnID string) ([]byte, error) {
	values := map[string]any{"client_txn_id": clientTxnID}
	cellsByOrdinal := map[int]string{}
	for ordinal, cell := range sourceRowCellsByOrdinal(sourceRow) {
		if text, ok := cell["display_text"].(string); ok {
			cellsByOrdinal[ordinal] = text
		}
	}
	for _, column := range mapping.SourceColumns {
		if column.FieldKey == nil {
			continue
		}
		text := cellsByOrdinal[column.SourceColumnOrdinal]
		transformed, err := transformImportValue(text, column)
		if err != nil {
			return nil, err
		}
		if transformed == "" {
			if column.EmptyValuePolicy == "omit_field" {
				continue
			}
			values[*column.FieldKey] = nil
			continue
		}
		switch {
		case mapping.TargetViewSchemaID == timeline.TimelineViewSchemaID && *column.FieldKey == "timeline.host_refs":
			values[*column.FieldKey] = collectionPayload([]map[string]any{{"op": "add_token", "raw_text": transformed}})
		case mapping.TargetViewSchemaID == timeline.TimelineViewSchemaID && *column.FieldKey == "timeline.identity_refs":
			values[*column.FieldKey] = collectionPayload([]map[string]any{{"op": "add_token", "raw_text": transformed}})
		case mapping.TargetViewSchemaID == timeline.TimelineViewSchemaID && *column.FieldKey == "timeline.tags":
			values[*column.FieldKey] = collectionPayload([]map[string]any{{"op": "add_tag", "tag_name": transformed}})
		default:
			values[*column.FieldKey] = transformed
		}
	}
	return json.Marshal(values)
}

func sourceRowCellsByOrdinal(sourceRow map[string]any) map[int]map[string]any {
	cellsByOrdinal := map[int]map[string]any{}
	if cells, ok := sourceRow["cells"].([]any); ok {
		for _, rawCell := range cells {
			cell, ok := rawCell.(map[string]any)
			if !ok {
				continue
			}
			ordinal, ok := intFromAny(cell["source_column_ordinal"])
			if ok {
				cellsByOrdinal[ordinal] = cell
			}
		}
		return cellsByOrdinal
	}
	if cells, ok := sourceRow["cells"].([]map[string]any); ok {
		for _, cell := range cells {
			ordinal, ok := intFromAny(cell["source_column_ordinal"])
			if ok {
				cellsByOrdinal[ordinal] = cell
			}
		}
	}
	return cellsByOrdinal
}

func transformImportValue(value string, column SourceColumnMapping) (string, error) {
	result := value
	if column.TransformID == nil {
		return result, nil
	}
	switch *column.TransformID {
	case "trim_v1":
		return strings.TrimSpace(result), nil
	case "collapse_whitespace_v1":
		return strings.Join(strings.Fields(result), " "), nil
	case "lowercase_v1":
		return strings.ToLower(result), nil
	case "split_delimited_v1":
		delimiter, _ := column.TransformOptions["delimiter"].(string)
		trimItems, _ := column.TransformOptions["trim_items"].(bool)
		dropEmpty, _ := column.TransformOptions["drop_empty_items"].(bool)
		parts := strings.Split(result, delimiter)
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimItems {
				part = strings.TrimSpace(part)
			}
			if dropEmpty && part == "" {
				continue
			}
			out = append(out, part)
		}
		return strings.Join(out, delimiter), nil
	default:
		return "", fmt.Errorf("unsupported transform %q", *column.TransformID)
	}
}

func collectionPayload(actions []map[string]any) map[string]any {
	return map[string]any{
		"kind":    "collection_actions_v1",
		"actions": actions,
	}
}

func writeImportStoreError(w http.ResponseWriter, r *http.Request, err error, clientTxnID string) bool {
	if err == nil {
		return true
	}
	var stateConflict *StateConflictError
	var applyBlocked *ApplyBlockedError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, clientTxnConflict(clientTxnID))
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}})
	case errors.Is(err, ErrNotFound):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
	case errors.As(err, &stateConflict):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "import_state_conflict", Details: map[string]any{"reason_code": stateConflict.ReasonCode}})
	case errors.As(err, &applyBlocked):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "import_apply_blocked", Details: map[string]any{"reason_code": applyBlocked.ReasonCode}})
	default:
		writeAPIError(w, r, internalAPIError(err))
	}
	return false
}

func (s *Service) discoverImportUnits(envelope httpapi.UploadEnvelope, sourceFileKind string) ([]DiscoveredUnit, *httpapi.APIError) {
	switch sourceFileKind {
	case SourceFileKindCSV:
		if int64(len(envelope.File)) > s.limits.MaxCSVSourceBytes {
			return nil, importSourceRejected("csv_source_too_large", int64(len(envelope.File)), s.limits.MaxCSVSourceBytes)
		}
		maxColumns := int(s.limits.MaxColumns)
		parsedRows, err := tabularingest.ParseTableWithMaxColumns(string(envelope.File), "csv", maxColumns)
		if err != nil {
			if strings.Contains(err.Error(), "column count exceeded") {
				return nil, importSourceRejected("import_columns_exceeded", int64(maxColumns+1), s.limits.MaxColumns)
			}
			return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
		}
		unit, apiErr := s.discoveredImportUnit(tabularCellsFromStrings(parsedRows), "csv_file", "file")
		if apiErr != nil {
			return nil, apiErr
		}
		return []DiscoveredUnit{unit}, nil
	case SourceFileKindXLSX:
		if int64(len(envelope.File)) > s.limits.MaxXLSXSourceBytes {
			return nil, importSourceRejected("xlsx_source_too_large", int64(len(envelope.File)), s.limits.MaxXLSXSourceBytes)
		}
		tables, apiErr := parseXLSXTables(envelope.File, s.limits, s.archiveLimits)
		if apiErr != nil {
			return nil, apiErr
		}
		units := make([]DiscoveredUnit, 0, len(tables))
		for _, table := range tables {
			rect := SourceRect(len(table.Rows), len(table.Rows[0]))
			locatorBytes, err := json.Marshal(map[string]any{
				"sheet_name": table.SheetName,
				"rect_a1":    rect,
			})
			if err != nil {
				return nil, internalAPIError(err)
			}
			unit, unitErr := s.discoveredImportUnit(table.Rows, "xlsx_used_range", string(locatorBytes))
			if unitErr != nil {
				return nil, unitErr
			}
			units = append(units, unit)
		}
		return units, nil
	default:
		return nil, importSourceUnsupported("encrypted_or_unparseable_workbook")
	}
}

func (s *Service) discoveredImportUnit(rows [][]tabularCell, locatorKind string, locator string) (DiscoveredUnit, *httpapi.APIError) {
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	dataRows := len(rows) - 1
	if dataRows < 0 {
		dataRows = 0
	}
	if int64(dataRows) > s.limits.MaxRows {
		return DiscoveredUnit{}, importSourceRejected("import_rows_exceeded", int64(dataRows), s.limits.MaxRows)
	}
	if int64(dataRows*maxCols) > s.limits.MaxCells {
		return DiscoveredUnit{}, importSourceRejected("import_cells_exceeded", int64(dataRows*maxCols), s.limits.MaxCells)
	}
	columns := make([]map[string]any, 0, maxCols)
	if len(rows) > 0 {
		for index := 0; index < maxCols; index++ {
			var header any
			if index < len(rows[0]) && rows[0][index].DisplayText != "" {
				header = rows[0][index].DisplayText
			}
			columns = append(columns, map[string]any{
				"source_column_ordinal": index + 1,
				"source_header_text":    header,
			})
		}
	}
	previewRows := make([]map[string]any, 0)
	sourceRows := make([]map[string]any, 0)
	for rowIndex := 1; rowIndex < len(rows) && len(previewRows) < 50; rowIndex++ {
		sourceRowRef := rowIndex + 1
		cells := make([]map[string]any, 0, maxCols)
		for columnIndex := 0; columnIndex < maxCols; columnIndex++ {
			cell := tabularCell{CellKind: "blank"}
			if columnIndex < len(rows[rowIndex]) {
				cell = rows[rowIndex][columnIndex]
			}
			cells = append(cells, map[string]any{
				"source_column_ordinal": columnIndex + 1,
				"display_text":          cell.DisplayText,
				"cell_kind":             cell.CellKind,
			})
		}
		previewRows = append(previewRows, map[string]any{
			"source_row_ref": sourceRowRef,
			"cells":          cells,
		})
	}
	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		sourceRowRef := rowIndex + 1
		cells := make([]map[string]any, 0, maxCols)
		for columnIndex := 0; columnIndex < maxCols; columnIndex++ {
			cell := tabularCell{CellKind: "blank"}
			if columnIndex < len(rows[rowIndex]) {
				cell = rows[rowIndex][columnIndex]
			}
			cells = append(cells, map[string]any{
				"source_column_ordinal": columnIndex + 1,
				"display_text":          cell.DisplayText,
				"cell_kind":             cell.CellKind,
			})
		}
		sourceRows = append(sourceRows, map[string]any{
			"source_row_ref": sourceRowRef,
			"cells":          cells,
		})
	}
	return DiscoveredUnit{
		LocatorKind:         locatorKind,
		Locator:             locator,
		SourceRectA1:        SourceRect(len(rows), maxCols),
		HeaderRowRef:        1,
		DataStartRowRef:     2,
		InferredRowCount:    dataRows,
		InferredColumnCount: maxCols,
		WarningCodes:        []string{},
		Columns:             columns,
		SourceRows:          sourceRows,
		PreviewRows:         previewRows,
	}, nil
}

func detectSourceFileKind(envelope httpapi.UploadEnvelope) string {
	if looksLikeXLSX(envelope.File) {
		return SourceFileKindXLSX
	}
	return SourceFileKindCSV
}

func looksLikeXLSX(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return (data[0] == 'P' && data[1] == 'K' && data[2] == 0x03 && data[3] == 0x04) ||
		(data[0] == 'P' && data[1] == 'K' && data[2] == 0x05 && data[3] == 0x06) ||
		(data[0] == 'P' && data[1] == 'K' && data[2] == 0x07 && data[3] == 0x08)
}

func (s *Service) executeDiscoveryJob(ctx context.Context, jobID uuid.UUID) error {
	var payload discoveryJobHandlerPayload
	if err := s.decodeJobPayload(ctx, jobID, &payload); err != nil {
		return err
	}
	sessionID, err := uuid.Parse(payload.ImportSessionID)
	if err != nil {
		return err
	}
	return s.completeDiscoveryJob(ctx, jobID, sessionID)
}

func (s *Service) completeDiscoveryJob(ctx context.Context, jobID uuid.UUID, importSessionID uuid.UUID) error {
	total := 1
	if !s.markJobRunningOrResume(ctx, jobID, total) {
		job, err := s.jobManager.Get(ctx, jobID)
		if err == nil && job.Status == jobs.StatusCanceled {
			return s.store.CancelDiscovery(ctx, importSessionID, s.now())
		}
		return nil
	}
	_, err := s.jobSuccessFinalizer.FinalizeImportJobSuccess(ctx, JobSuccessFinalization{
		FinalCommitID: "import.discovery:" + jobID.String(),
		Transition: jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: 1, Total: &total},
			ResultSummary: &jobs.ResultSummary{
				Code:    "import_session_discovered",
				Message: "Import session discovered.",
				ResourceRefs: []jobs.ResourceRef{{
					Kind:  "import_session",
					ID:    importSessionID.String(),
					Route: "/api/v1/import-sessions/" + importSessionID.String(),
				}},
			},
		},
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			return s.store.MarkDiscoveredTx(ctx, tx, importSessionID, s.now())
		},
	})
	return err
}

func (s *Service) decodeJobPayload(ctx context.Context, jobID uuid.UUID, target any) error {
	payload, err := s.jobManager.HandlerPayload(ctx, jobID)
	if err != nil {
		return err
	}
	if len(payload) == 0 {
		return fmt.Errorf("missing import job payload")
	}
	return json.Unmarshal(payload, target)
}

func (s *Service) markJobRunningOrResume(ctx context.Context, jobID uuid.UUID, total int) bool {
	if total <= 0 {
		total = 1
	}
	if _, err := s.jobManager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 0, Total: &total}, nil); err == nil {
		return true
	}
	job, err := s.jobManager.Get(ctx, jobID)
	if err != nil {
		return false
	}
	switch job.Status {
	case jobs.StatusRunning:
		return true
	case jobs.StatusCancelRequested:
		_, _ = s.jobManager.CompleteCanceled(ctx, jobs.TransitionParams{
			JobID:    jobID,
			Progress: jobs.Progress{Completed: 0, Total: &total},
		})
		return false
	default:
		return false
	}
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, s.incidentAccess, incidentID, userID, roles...)
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

type importSessionRoute struct {
	Kind      string
	SessionID uuid.UUID
	UnitID    uuid.UUID
}

func parseImportSessionPath(path string) (importSessionRoute, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/import-sessions/")
	if rest == path || rest == "" {
		return importSessionRoute{}, false
	}
	parts := strings.Split(rest, "/")
	sessionID, err := uuid.Parse(parts[0])
	if err != nil {
		return importSessionRoute{}, false
	}
	if len(parts) == 1 {
		return importSessionRoute{Kind: "session", SessionID: sessionID}, true
	}
	if len(parts) == 2 && parts[1] == "units" {
		return importSessionRoute{Kind: "units", SessionID: sessionID}, true
	}
	if len(parts) == 3 && parts[1] == "units" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "unit", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "preview" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "preview", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "mapping-preview" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "mapping_preview", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "mapping" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "mapping", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "select" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "select", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "skip" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "skip", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 2 && parts[1] == "apply" {
		return importSessionRoute{Kind: "apply", SessionID: sessionID}, true
	}
	return importSessionRoute{}, false
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
