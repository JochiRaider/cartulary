package reference_data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

// packsRouteFamily is the claimed Reference Pack family root (Core 01 §17.4).
const packsRouteFamily = "/api/v1/reference-packs"

// listOperationID is the collection list, the one family operation whose query
// is not the empty administrative query contract.
const listOperationID = "listReferencePacks"

type httpAdapter struct {
	coordinator AdministrativeApplication
	sessions    httpauth.SessionSlider
	jobRunner   jobNotifier
	cursorCodec *pagination.Codec
	now         func() time.Time
}

type jobNotifier interface{ Notify(uuid.UUID) }

// RegisterRoutes binds the Reference Pack family for central admission in its
// own order (Core 01 §17.4, REQ-01-707): the platform gate authenticates the
// session through the application's authenticator, then the family policy
// requires current deployment_admin; each selected operation then verifies its
// declared cookie CSRF, its query and its path parameters before its handler.
func RegisterRoutes(application AdministrativeApplication, notifier jobNotifier) httpapi.RouteRegistrar {
	return func(_ *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newHTTPAdapter(deps, application, notifier)
		if err != nil {
			return err
		}
		return service.bindFamily(deps)
	}
}

func (s *httpAdapter) bindFamily(deps httpapi.DependencySet) error {
	return httpapi.BindAdmittedOwnerFamilyRoutes(deps, httpapi.OwnerFamily{
		OwnerID:    "module.reference_data",
		ProfileID:  ProfileID,
		Root:       packsRouteFamily,
		Authorize:  authorizeFamily,
		AdmitQuery: admitOperationQuery,
		Handlers:   s.operationHandlers(),
	})
}

// operationHandlers maps each Reference Pack operation to its one handler.
func (s *httpAdapter) operationHandlers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"activateReferencePackVersion":      s.handleActivate,
		"disableReferencePackVersion":       s.handleDisable,
		"getReferencePackVersion":           s.handleRead,
		"getReferencePackValidationSummary": s.handleValidationSummary,
		"importReferencePack":               s.handleImport,
		listOperationID:                     s.handleCollection,
		"refreshReferencePacks":             s.handleRefresh,
		"reverifyReferencePackVersion":      s.handleReverify,
		"removeReferencePackVersion":        s.handleRemove,
	}
}

func newHTTPAdapter(deps httpapi.DependencySet, application AdministrativeApplication, notifier jobNotifier) (*httpAdapter, error) {
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
		cursorKey, err := authn.DerivePurposeKey(keys, "pagination-cursor-v2")
		if err != nil {
			return nil, err
		}
		cursorCodec = pagination.NewCodec(cursorKey)
	}
	if application == nil || notifier == nil {
		return nil, errors.New("reference pack: routes require a complete application and Jobs notifier")
	}
	return &httpAdapter{coordinator: application, sessions: authn.NewStore(deps.PostgresHandle()), jobRunner: notifier, cursorCodec: cursorCodec, now: now}, nil
}

func (s *httpAdapter) handleCollection(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.authorizedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	listScope, apiErr := parseReferencePackListScope(r.URL.RawQuery)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	binding, cursor, reason := s.cursorCodec.ResolveListRequest(listScope.Values, "reference_packs.list", principal.User.ID.String(), listScope.Scope)
	if reason != "" {
		writeAPIError(w, r, invalidPaginationRequest(reason))
		return
	}
	records, err := s.coordinator.ListVersions(r.Context())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	records = filterAdministrativeVersions(records, binding.Scope)
	resources := make([]map[string]any, 0, len(records))
	for _, record := range records {
		resources = append(resources, record.Resource())
	}
	rows, nextCursor, err := pagination.PageResources(binding, cursor, resources)
	if errors.Is(err, pagination.ErrInvalidCursorToken) {
		writeAPIError(w, r, invalidPaginationRequest(pagination.ReasonInvalidCursorToken))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	var nextCursorToken *string
	if nextCursor != nil {
		token, err := s.cursorCodec.Encode(*nextCursor)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		nextCursorToken = &token
	}
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"pack_versions": rows}, httpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextCursorToken != nil,
		NextCursor: nextCursorToken,
	})
}

func parseReferencePackListScope(rawQuery string) (listquery.Result, *httpapi.APIError) {
	result, queryErr := listquery.Parse(rawQuery, listquery.Config{
		Search: true,
		ExactFilters: map[string]listquery.ExactFilter{
			"pack_version_state": {Allowed: []string{ConditionStaged, ConditionVerifiedAvailable, ConditionDisabled, ConditionFailed, ConditionMissing}},
			"active":             {Allowed: []string{"true", "false"}},
		},
	})
	if queryErr == nil {
		return result, nil
	}
	if queryErr.Kind == listquery.ErrorKindPagination {
		return listquery.Result{}, invalidPaginationRequest(queryErr.ReasonCode)
	}
	return listquery.Result{}, invalidListQuery(queryErr.ReasonCode)
}

func searchableOptionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *httpAdapter) handleRead(w http.ResponseWriter, r *http.Request) {
	packKey, packVersion, apiErr := admittedPackVersion(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, apiErr := s.authorizedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.coordinator.GetVersion(r.Context(), packKey, packVersion)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, referencePackNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, record.Resource())
}

func (s *httpAdapter) handleValidationSummary(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.authorizedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	summaryID, ok := httpapi.PathString(r, "summary_id")
	if !ok {
		writeAPIError(w, r, internalAPIError(errors.New("reference pack: missing path admission")))
		return
	}
	summary, err := s.coordinator.GetValidationSummary(r.Context(), summaryID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, referencePackNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, summary)
}

func (s *httpAdapter) handleImport(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.authorizedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if s.coordinator == nil {
		writeAPIError(w, r, internalAPIError(errors.New("reference pack coordinator unavailable")))
		return
	}
	var pending *PendingImport
	defer func() {
		if pending != nil {
			_ = pending.Close()
		}
	}()
	envelope, envelopeErr, receiveErr := httpapi.ParseStreamingUploadEnvelope(r, httpapi.UploadEnvelopePolicy{FileContentTypes: ReferencePackFileContentTypes}, MaxAdministrativeRequestBytes, func(source io.Reader) (string, error) {
		var err error
		pending, err = s.coordinator.PrepareImport(r.Context(), source)
		if err != nil {
			return "", err
		}
		return pending.SHA256(), nil
	})
	if receiveErr != nil {
		writeAPIError(w, r, coordinatorAPIError(receiveErr))
		return
	}
	if envelopeErr != nil {
		writeAPIError(w, r, uploadEnvelopeAPIError(envelopeErr))
		return
	}
	request, apiErr := DecodeImportMetadata(envelope)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := pending.Accept(r.Context(), principal.User.ID, request.ClientTxnID)
	if errors.Is(err, ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, coordinatorAPIError(err))
		return
	}
	if !result.Replayed {
		s.dispatchReferencePackJob(result.Job.JobID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *httpAdapter) handleActivate(w http.ResponseWriter, r *http.Request) {
	packKey, packVersion, principal, request, ok := s.decodeAdminAction(w, r, false)
	if !ok {
		return
	}
	result, err := s.coordinator.Activate(r.Context(), ActionParams{
		ActorUserID: principal.User.ID,
		PackKey:     packKey,
		PackVersion: packVersion,
		Request:     request,
		Now:         s.now(),
	})
	s.writeActionResult(w, r, &principal, request.ClientTxnID, result, err)
}

func (s *httpAdapter) handleDisable(w http.ResponseWriter, r *http.Request) {
	packKey, packVersion, principal, request, ok := s.decodeAdminAction(w, r, false)
	if !ok {
		return
	}
	result, err := s.coordinator.Disable(r.Context(), ActionParams{
		ActorUserID: principal.User.ID,
		PackKey:     packKey,
		PackVersion: packVersion,
		Request:     request,
		Now:         s.now(),
	})
	s.writeActionResult(w, r, &principal, request.ClientTxnID, result, err)
}

func (s *httpAdapter) handleRemove(w http.ResponseWriter, r *http.Request) {
	key, version, principal, request, ok := s.decodeAdminAction(w, r, true)
	if !ok {
		return
	}
	result, err := s.coordinator.Remove(r.Context(), ActionParams{ActorUserID: principal.User.ID, PackKey: key, PackVersion: version, Request: request, Now: s.now()})
	s.writeActionResult(w, r, &principal, request.ClientTxnID, result, err)
}

func (s *httpAdapter) handleReverify(w http.ResponseWriter, r *http.Request) {
	packKey, packVersion, principal, request, ok := s.decodeAdminAction(w, r, false)
	if !ok {
		return
	}
	result, err := s.coordinator.VerifyRetained(r.Context(), VerificationRequest{Kind: "reverify", ActorUserID: principal.User.ID, ClientTxnID: request.ClientTxnID, PackKeys: []string{packKey}, KeysProvided: true, PackVersion: packVersion, Reason: request.Reason})
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, referencePackNotFound())
		return
	}
	if errors.Is(err, ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, coordinatorAPIError(err))
		return
	}
	if !result.Replayed {
		s.dispatchReferencePackJob(result.Job.JobID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, coordinatorAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *httpAdapter) handleRefresh(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.authorizedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeRefreshRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.coordinator.VerifyRetained(r.Context(), VerificationRequest{Kind: "refresh", ActorUserID: principal.User.ID, ClientTxnID: request.ClientTxnID, PackKeys: request.PackKeys, KeysProvided: request.PackKeysProvided})
	if errors.Is(err, ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, coordinatorAPIError(err))
		return
	}
	if !result.Replayed {
		s.dispatchReferencePackJob(result.Job.JobID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, coordinatorAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

// decodeAdminAction reads the admitted pack version and principal, then
// decodes the action body.
func (s *httpAdapter) decodeAdminAction(w http.ResponseWriter, r *http.Request, removal bool) (string, string, httpauth.Principal, ActionRequest, bool) {
	packKey, packVersion, apiErr := admittedPackVersion(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return "", "", httpauth.Principal{}, ActionRequest{}, false
	}
	principal, apiErr := s.authorizedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return "", "", httpauth.Principal{}, ActionRequest{}, false
	}
	request, apiErr := decodeActionRequest(r.Body, removal)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return "", "", httpauth.Principal{}, ActionRequest{}, false
	}
	return packKey, packVersion, principal, request, true
}

func (s *httpAdapter) writeActionResult(w http.ResponseWriter, r *http.Request, principal *httpauth.Principal, clientTxnID string, result ActionResult, err error) {
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, referencePackNotFound())
		return
	}
	if errors.Is(err, ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(clientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, coordinatorAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, actionReceipt{Version: result.Version})
}

func (s *httpAdapter) dispatchReferencePackJob(jobID string) {
	parsed, err := uuid.Parse(jobID)
	if err != nil {
		return
	}
	if s == nil || s.jobRunner == nil {
		return
	}
	s.jobRunner.Notify(parsed)
}

// authorizeFamily is the family policy (Core 01 §17.4): after the gate's
// session authentication, every family route requires current
// deployment_admin, before canonicality, method handling and every
// route-specific check. A denial never slides the session.
func authorizeFamily(r *http.Request) *httpapi.APIError {
	principal, ok := httpauth.AdmittedPrincipal(r)
	if !ok {
		return internalAPIError(errors.New("reference pack: missing family authentication"))
	}
	return httpauth.RequireDeploymentAdmin(principal.User)
}

// admitOperationQuery applies the empty administrative query contract to every
// operation except the collection list, after the operation's cookie CSRF and
// before its path parameters (Core 01 §17.4).
func admitOperationQuery(r *http.Request, operationID string) *httpapi.APIError {
	if operationID == listOperationID {
		return nil
	}
	return validateEmptyAdministrativeQuery(r.URL.RawQuery)
}

func (s *httpAdapter) authorizedPrincipal(r *http.Request) (httpauth.Principal, *httpapi.APIError) {
	principal, ok := httpauth.AdmittedPrincipal(r)
	if !ok || !principal.User.IsDeploymentAdmin {
		return httpauth.Principal{}, internalAPIError(errors.New("reference pack: missing family admission"))
	}
	return principal, nil
}

// admittedPackVersion returns the centrally admitted pack_key and
// pack_version path parameters.
func admittedPackVersion(r *http.Request) (string, string, *httpapi.APIError) {
	packKey, keyAdmitted := httpapi.PathString(r, "pack_key")
	packVersion, versionAdmitted := httpapi.PathString(r, "pack_version")
	if !keyAdmitted || !versionAdmitted {
		return "", "", internalAPIError(errors.New("reference pack: missing path admission"))
	}
	return packKey, packVersion, nil
}

func (s *httpAdapter) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.sessions, principal, s.now)
}

func (v AdministrativeVersion) Resource() map[string]any {
	// Paging consumes JSON-shaped values, including explicit nulls. This is a
	// projection of the closed typed resource, not a second schema definition.
	raw, _ := json.Marshal(v)
	var resource map[string]any
	_ = json.Unmarshal(raw, &resource)
	return resource
}

func filterAdministrativeVersions(versions []AdministrativeVersion, scope map[string]string) []AdministrativeVersion {
	result := versions[:0]
	for _, v := range versions {
		if state := scope["pack_version_state"]; state != "" && state != v.Condition {
			continue
		}
		if active := scope["active"]; active != "" && v.Active != (active == "true") {
			continue
		}
		if !listquery.MatchSearchTokens(strings.Fields(scope["search"]), v.PackKey, v.PackKind, v.PackVersion, searchableOptionalString(v.SourceIdentifier), searchableOptionalString(v.ManifestSHA256), searchableOptionalString(v.PayloadSHA256)) {
			continue
		}
		result = append(result, v)
	}
	return result
}
