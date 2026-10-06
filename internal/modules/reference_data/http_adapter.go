package reference_data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

type httpAdapter struct {
	coordinator AdministrativeApplication
	authStore   *authn.Store
	jobRunner   jobNotifier
	keys        authn.MasterKeys
	cursorCodec *pagination.Codec
	now         func() time.Time
}

type jobNotifier interface{ Notify(uuid.UUID) }

func RegisterRoutes(application AdministrativeApplication, notifier jobNotifier) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newHTTPAdapter(deps, application, notifier)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.reference_data", map[string]http.HandlerFunc{
			"activateReferencePackVersion":      service.handleMember,
			"disableReferencePackVersion":       service.handleMember,
			"getReferencePackVersion":           service.handleMember,
			"getReferencePackValidationSummary": service.handleValidationSummary,
			"importReferencePack":               service.handleMember,
			"listReferencePacks":                service.handleCollection,
			"refreshReferencePacks":             service.handleMember,
			"reverifyReferencePackVersion":      service.handleMember,
			"removeReferencePackVersion":        service.handleMember,
		})
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
	return &httpAdapter{coordinator: application, authStore: authn.NewStore(deps.PostgresHandle()), jobRunner: notifier, keys: keys, cursorCodec: cursorCodec, now: now}, nil
}

func (s *httpAdapter) handleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
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
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
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
	case http.MethodPost:
		http.NotFound(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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

func (s *httpAdapter) handleMember(w http.ResponseWriter, r *http.Request) {
	route, ok := parseReferencePackPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if route.Kind == "import" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleImport(w, r)
		return
	}
	if route.Kind == "refresh" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleRefresh(w, r)
		return
	}
	if route.Kind == "missing_version" {
		writeAPIError(w, r, invalidReferencePackRequest("pack_version", "pack_version_required"))
		return
	}

	switch {
	case route.Kind == "read" && r.Method == http.MethodGet:
		s.handleRead(w, r, route.PackKey, route.PackVersion)
	case route.Kind == "activate" && r.Method == http.MethodPost:
		s.handleActivate(w, r, route.PackKey, route.PackVersion)
	case route.Kind == "disable" && r.Method == http.MethodPost:
		s.handleDisable(w, r, route.PackKey, route.PackVersion)
	case route.Kind == "remove" && r.Method == http.MethodPost:
		s.handleRemove(w, r, route.PackKey, route.PackVersion)
	case route.Kind == "reverify" && r.Method == http.MethodPost:
		s.handleReverify(w, r, route.PackKey, route.PackVersion)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *httpAdapter) handleRead(w http.ResponseWriter, r *http.Request, packKey string, packVersion string) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, record.Resource())
}

func (s *httpAdapter) handleValidationSummary(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.requireDeploymentAdmin(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	summary, err := s.coordinator.GetValidationSummary(r.Context(), r.PathValue("summary_id"))
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, referencePackNotFound())
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
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, summary)
}

func (s *httpAdapter) handleImport(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.requireDeploymentAdmin(r, true)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *httpAdapter) handleActivate(w http.ResponseWriter, r *http.Request, packKey string, packVersion string) {
	principal, request, ok := s.decodeAdminAction(w, r)
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

func (s *httpAdapter) handleDisable(w http.ResponseWriter, r *http.Request, packKey string, packVersion string) {
	principal, request, ok := s.decodeAdminAction(w, r)
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

func (s *httpAdapter) handleRemove(w http.ResponseWriter, r *http.Request, key, version string) {
	principal, request, ok := s.decodeAdminAction(w, r)
	if !ok {
		return
	}
	result, err := s.coordinator.Remove(r.Context(), ActionParams{ActorUserID: principal.User.ID, PackKey: key, PackVersion: version, Request: request, Now: s.now()})
	s.writeActionResult(w, r, &principal, request.ClientTxnID, result, err)
}

func (s *httpAdapter) handleReverify(w http.ResponseWriter, r *http.Request, packKey string, packVersion string) {
	principal, request, ok := s.decodeAdminAction(w, r)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, coordinatorAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *httpAdapter) handleRefresh(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.requireDeploymentAdmin(r, true)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, coordinatorAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *httpAdapter) decodeAdminAction(w http.ResponseWriter, r *http.Request) (httpauth.Principal, ActionRequest, bool) {
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return httpauth.Principal{}, ActionRequest{}, false
	}
	request, apiErr := decodeActionRequest(r.Body, strings.HasSuffix(r.URL.Path, "/remove"))
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return httpauth.Principal{}, ActionRequest{}, false
	}
	return principal, request, true
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
	if err := s.slideSessionIfNeeded(r.Context(), principal, r.Method, r.URL.Path); err != nil {
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

func (s *httpAdapter) requireDeploymentAdmin(r *http.Request, stateChanging bool) (httpauth.Principal, *httpapi.APIError) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: stateChanging})
	if apiErr != nil {
		return httpauth.Principal{}, apiErr
	}
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
		return httpauth.Principal{}, apiErr
	}
	return principal, nil
}

func (s *httpAdapter) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

type parsedReferencePackRoute struct {
	Kind        string
	PackKey     string
	PackVersion string
}

func parseReferencePackPath(requestPath string) (parsedReferencePackRoute, bool) {
	rest := strings.TrimPrefix(requestPath, "/api/v1/reference-packs/")
	if rest == requestPath || rest == "" {
		return parsedReferencePackRoute{}, false
	}
	parts := strings.Split(rest, "/")
	if len(parts) == 1 {
		switch parts[0] {
		case "import":
			return parsedReferencePackRoute{Kind: "import"}, true
		case "refresh":
			return parsedReferencePackRoute{Kind: "refresh"}, true
		default:
			return parsedReferencePackRoute{}, false
		}
	}
	if len(parts) == 2 {
		if parts[1] == "activate" || parts[1] == "disable" || parts[1] == "reverify" || parts[1] == "remove" {
			return parsedReferencePackRoute{Kind: "missing_version"}, true
		}
		packKey, ok1 := unescapePathSegment(parts[0])
		packVersion, ok2 := unescapePathSegment(parts[1])
		if !ok1 || !ok2 {
			return parsedReferencePackRoute{}, false
		}
		return parsedReferencePackRoute{Kind: "read", PackKey: packKey, PackVersion: packVersion}, true
	}
	if len(parts) == 3 {
		switch parts[2] {
		case "activate", "disable", "reverify", "remove":
			packKey, ok1 := unescapePathSegment(parts[0])
			packVersion, ok2 := unescapePathSegment(parts[1])
			if !ok1 || !ok2 {
				return parsedReferencePackRoute{}, false
			}
			return parsedReferencePackRoute{Kind: parts[2], PackKey: packKey, PackVersion: packVersion}, true
		default:
			return parsedReferencePackRoute{}, false
		}
	}
	return parsedReferencePackRoute{}, false
}

func unescapePathSegment(raw string) (string, bool) {
	value, err := url.PathUnescape(raw)
	if err != nil || strings.TrimSpace(value) == "" || strings.Contains(value, "/") {
		return "", false
	}
	return value, true
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
