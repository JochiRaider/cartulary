package reference_data

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/google/uuid"
)

type Service struct {
	store       *Store
	authStore   *authn.Store
	jobManager  *jobs.Manager
	jobRunner   *jobs.Runner
	hub         *platformws.Hub
	keys        authn.MasterKeys
	cursorCodec *pagination.Codec
	deps        httpapi.DependencySet
	now         func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.ExtensionProfileClaimedIn(deps.ExtensionProfiles, ProfileID) {
			return nil
		}
		service, err := newService(deps)
		if err != nil {
			return err
		}
		if err := service.recoverReferencePackJobs(context.Background()); err != nil {
			return err
		}
		mux.HandleFunc("/api/v1/reference-packs", service.handleCollection)
		mux.HandleFunc("/api/v1/reference-packs/", service.handleMember)
		return nil
	}
}

func newService(deps httpapi.DependencySet) (*Service, error) {
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
	return &Service{
		store:       NewStore(deps.Postgres),
		authStore:   authn.NewStore(deps.PostgresHandle()),
		jobManager:  deps.Jobs,
		jobRunner:   deps.JobRunner,
		hub:         deps.WSHub,
		keys:        keys,
		cursorCodec: cursorCodec,
		deps:        deps,
		now:         now,
	}, nil
}

func (s *Service) handleCollection(w http.ResponseWriter, r *http.Request) {
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
		records, err := s.store.ListVersions(r.Context())
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		records = filterReferencePackVersions(records, binding.Scope)
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
			"pack_version_state":  {Allowed: []string{ConditionStaged, ConditionVerifiedAvailable, ConditionDisabled, ConditionFailed, ConditionMissing}},
			"verification_result": {Allowed: []string{VerificationPending, VerificationPassed, VerificationFailed}},
			"active":              {Allowed: []string{"true", "false"}},
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

func filterReferencePackVersions(records []VersionRecord, scope map[string]string) []VersionRecord {
	tokens := strings.Fields(scope["search"])
	packVersionState := scope["pack_version_state"]
	verificationResult := scope["verification_result"]
	active := scope["active"]
	filtered := records[:0]
	for _, record := range records {
		if packVersionState != "" && publicCondition(record.StoredStatus, record.VerificationResult) != packVersionState {
			continue
		}
		if verificationResult != "" && record.VerificationResult != verificationResult {
			continue
		}
		if active != "" && (record.Active != (active == "true")) {
			continue
		}
		if !listquery.MatchSearchTokens(tokens,
			record.PackKey,
			record.PackKind,
			record.PackVersion,
			searchableOptionalString(record.SourceIdentifier),
			record.ManifestSHA256,
			record.PayloadSHA256,
			searchableOptionalString(record.SignerKeyID),
		) {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered
}

func searchableOptionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *Service) handleMember(w http.ResponseWriter, r *http.Request) {
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
	case route.Kind == "reverify" && r.Method == http.MethodPost:
		s.handleReverify(w, r, route.PackKey, route.PackVersion)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleRead(w http.ResponseWriter, r *http.Request, packKey string, packVersion string) {
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
	record, err := s.store.GetVersion(r.Context(), packKey, packVersion)
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

func (s *Service) handleImport(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	envelope, envelopeErr := httpapi.ParseUploadEnvelope(r, httpapi.UploadEnvelopePolicy{FileContentTypes: ReferencePackFileContentTypes})
	if envelopeErr != nil {
		writeAPIError(w, r, uploadEnvelopeAPIError(envelopeErr))
		return
	}
	request, apiErr := DecodeImportMetadata(envelope)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	stagingPath, err := s.stageBundle(envelope.FileSHA256Hex, envelope.File)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := s.store.AcceptImport(r.Context(), ImportAcceptedParams{
		ActorUserID:       principal.User.ID,
		Request:           request,
		BundleSHA256:      envelope.FileSHA256Hex,
		BundleStagingPath: stagingPath,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		_ = os.Remove(stagingPath)
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		_ = os.Remove(stagingPath)
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !result.Replayed {
		s.dispatchReferencePackJob(result.Job.JobID)
	} else {
		_ = os.Remove(stagingPath)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *Service) handleActivate(w http.ResponseWriter, r *http.Request, packKey string, packVersion string) {
	principal, request, ok := s.decodeAdminAction(w, r)
	if !ok {
		return
	}
	result, err := s.store.Activate(r.Context(), ActionParams{
		ActorUserID: principal.User.ID,
		PackKey:     packKey,
		PackVersion: packVersion,
		Request:     request,
		Now:         s.now(),
	})
	s.writeActionResult(w, r, &principal, request.ClientTxnID, result, err)
}

func (s *Service) handleDisable(w http.ResponseWriter, r *http.Request, packKey string, packVersion string) {
	principal, request, ok := s.decodeAdminAction(w, r)
	if !ok {
		return
	}
	result, err := s.store.Disable(r.Context(), ActionParams{
		ActorUserID: principal.User.ID,
		PackKey:     packKey,
		PackVersion: packVersion,
		Request:     request,
		Now:         s.now(),
	})
	s.writeActionResult(w, r, &principal, request.ClientTxnID, result, err)
}

func (s *Service) handleReverify(w http.ResponseWriter, r *http.Request, packKey string, packVersion string) {
	principal, request, ok := s.decodeAdminAction(w, r)
	if !ok {
		return
	}
	result, err := s.store.AcceptReverify(r.Context(), ActionParams{
		ActorUserID: principal.User.ID,
		PackKey:     packKey,
		PackVersion: packVersion,
		Request:     request,
		Now:         s.now(),
	})
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, referencePackNotFound())
		return
	}
	var wrapped apiError
	if errors.As(err, &wrapped) {
		writeAPIError(w, r, wrapped.apiErr)
		return
	}
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
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

func (s *Service) handleRefresh(w http.ResponseWriter, r *http.Request) {
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
	if replay, payload, ok, err := s.store.LookupRefreshReplay(r.Context(), principal.User.ID, request.ClientTxnID); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	} else if ok {
		if request.PackKeysProvided {
			resolved := append([]string(nil), request.PackKeys...)
			sort.Strings(resolved)
			if !sameStringSet(resolved, payload.ResolvedPackKeys) {
				writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
				return
			}
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, replay.Job)
		return
	}
	visiblePackKeys, err := s.store.ListPackKeys(r.Context())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	resolved, apiErr := ValidateRefreshPackKeys(request, visiblePackKeys)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, err = NormalizeRefreshRequest(request, resolved)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := s.store.AcceptRefresh(r.Context(), RefreshAcceptedParams{
		ActorUserID:       principal.User.ID,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
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
		s.dispatchReferencePackJob(result.Job.JobID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *Service) decodeAdminAction(w http.ResponseWriter, r *http.Request) (httpauth.Principal, ActionRequest, bool) {
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return httpauth.Principal{}, ActionRequest{}, false
	}
	request, apiErr := DecodeActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return httpauth.Principal{}, ActionRequest{}, false
	}
	return principal, request, true
}

func (s *Service) writeActionResult(w http.ResponseWriter, r *http.Request, principal *httpauth.Principal, clientTxnID string, result ActionResult, err error) {
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, referencePackNotFound())
		return
	}
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(clientTxnID))
		return
	}
	var wrapped apiError
	if errors.As(err, &wrapped) {
		writeAPIError(w, r, wrapped.apiErr)
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) recoverReferencePackJobs(ctx context.Context) error {
	jobIDs, err := s.store.PendingJobIDs(ctx)
	if err != nil {
		return err
	}
	for _, jobID := range jobIDs {
		s.dispatchReferencePackJob(jobID.String())
	}
	return nil
}

func (s *Service) dispatchReferencePackJob(jobID string) {
	parsed, err := uuid.Parse(jobID)
	if err != nil {
		return
	}
	work := func(ctx context.Context) {
		s.executeReferencePackJob(ctx, parsed)
	}
	if s.jobRunner != nil {
		if err := s.jobRunner.Dispatch(work); err == nil {
			return
		}
	}
	go work(context.Background())
}

func (s *Service) executeReferencePackJob(ctx context.Context, jobID uuid.UUID) {
	payload, err := s.store.JobPayload(ctx, jobID)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(jobID, "internal_error", map[string]any{}))
		return
	}
	runReferencePackWorkerStartHook(payload.JobKind)
	total := 1
	if payload.JobKind == "refresh" && len(payload.ResolvedPackKeys) > 0 {
		total = len(payload.ResolvedPackKeys)
	}
	if ok := s.markJobRunningOrResume(ctx, jobID, total); !ok {
		return
	}
	switch payload.JobKind {
	case "import":
		s.executeImportJob(ctx, payload)
	case "reverify":
		s.executeReverifyJob(ctx, payload)
	case "refresh":
		s.executeRefreshJob(ctx, payload)
	default:
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(jobID, "internal_error", map[string]any{"job_kind": payload.JobKind}))
	}
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

func (s *Service) executeImportJob(ctx context.Context, payload JobPayload) {
	if payload.BundleStagingPath == nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	defer func() { _ = os.Remove(*payload.BundleStagingPath) }()
	if s.jobCancelRequested(ctx, payload.JobID) {
		s.completeCanceled(ctx, payload.JobID, 0, 1)
		return
	}
	data, err := os.ReadFile(*payload.BundleStagingPath)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "reference_pack_verification_failed", map[string]any{"reason_code": "payload_missing"}))
		return
	}
	verification, verificationErr := s.verifyUpload(data, MediaTypeOctetStream)
	if verificationErr != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "reference_pack_verification_failed", map[string]any{"reason_code": verificationErr.ReasonCode}))
		return
	}
	if s.jobCancelRequested(ctx, payload.JobID) {
		s.completeCanceled(ctx, payload.JobID, 0, 1)
		return
	}
	bundlePath, err := s.persistBundle(verification.BundleSHA256, data)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	record, err := s.store.CompleteImportVerification(ctx, payload.JobID, payload.ActorUserID, *verification, bundlePath, s.now())
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	_, _ = s.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
		JobID:    payload.JobID,
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ResultSummary: &jobs.ResultSummary{
			Code:         ResultReferencePackImported,
			Message:      "Reference pack imported.",
			ResourceRefs: []jobs.ResourceRef{jobs.ResourceRef(referencePackResourceRef(record.PackKey, record.PackVersion))},
		},
	})
}

func (s *Service) executeReverifyJob(ctx context.Context, payload JobPayload) {
	if payload.PackKey == nil || payload.PackVersion == nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	if s.jobCancelRequested(ctx, payload.JobID) {
		s.completeCanceled(ctx, payload.JobID, 0, 1)
		return
	}
	record, err := s.store.GetVersion(ctx, *payload.PackKey, *payload.PackVersion)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "reference_pack_not_found", map[string]any{}))
		return
	}
	verification, verificationErr := s.verifyStoredBundle(record)
	updated, applyErr := s.store.ApplyVerificationResult(ctx, record, verification, verificationErr, "reverify", payload.ActorUserID, payload.JobID, s.now())
	if applyErr != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	if verificationErr != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "reference_pack_verification_failed", map[string]any{"reason_code": verificationErr.ReasonCode}))
		return
	}
	_, _ = s.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
		JobID:    payload.JobID,
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ResultSummary: &jobs.ResultSummary{
			Code:         ResultReferencePackReverified,
			Message:      "Reference pack reverified.",
			ResourceRefs: []jobs.ResourceRef{jobs.ResourceRef(referencePackResourceRef(updated.PackKey, updated.PackVersion))},
		},
	})
}

func (s *Service) executeRefreshJob(ctx context.Context, payload JobPayload) {
	total := len(payload.ResolvedPackKeys)
	if total == 0 {
		_, _ = s.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
			JobID:    payload.JobID,
			Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
			ResultSummary: &jobs.ResultSummary{
				Code:    ResultReferencePacksRefreshed,
				Message: "Reference packs refreshed.",
			},
		})
		return
	}
	records, err := s.store.ListVersionsForPackKeys(ctx, payload.ResolvedPackKeys)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	var updatedRecords []VersionRecord
	for i, record := range records {
		if s.jobCancelRequested(ctx, payload.JobID) {
			s.completeCanceled(ctx, payload.JobID, i, total)
			return
		}
		verification, verificationErr := s.verifyStoredBundle(record)
		updated, err := s.store.ApplyVerificationResult(ctx, record, verification, verificationErr, "refresh", payload.ActorUserID, payload.JobID, s.now())
		if err != nil {
			_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
			return
		}
		if verificationErr != nil {
			_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "reference_pack_verification_failed", map[string]any{"reason_code": verificationErr.ReasonCode}))
			return
		}
		updatedRecords = append(updatedRecords, updated)
	}
	_, _ = s.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
		JobID:    payload.JobID,
		Progress: jobs.Progress{Completed: total, Total: &total},
		ResultSummary: &jobs.ResultSummary{
			Code:         ResultReferencePacksRefreshed,
			Message:      "Reference packs refreshed.",
			ResourceRefs: sortedRefs(updatedRecords),
		},
	})
}

func (s *Service) jobCancelRequested(ctx context.Context, jobID uuid.UUID) bool {
	job, err := s.jobManager.Get(ctx, jobID)
	return err == nil && job.Status == jobs.StatusCancelRequested
}

func (s *Service) completeCanceled(ctx context.Context, jobID uuid.UUID, completed int, total int) {
	_, _ = s.jobManager.CompleteCanceled(ctx, jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: completed, Total: &total},
	})
}

func sameStringSet(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	sortedLeft := append([]string(nil), left...)
	sortedRight := append([]string(nil), right...)
	sort.Strings(sortedLeft)
	sort.Strings(sortedRight)
	for i := range sortedLeft {
		if sortedLeft[i] != sortedRight[i] {
			return false
		}
	}
	return true
}

func failedTransition(jobID uuid.UUID, code string, details map[string]any) jobs.TransitionParams {
	return jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ErrorSummary: &jobs.ErrorSummary{
			Code:      code,
			Message:   code,
			Retryable: false,
			Details:   details,
		},
	}
}

func (s *Service) verifyUpload(bundle []byte, contentType string) (*VerificationResult, *VerificationError) {
	result, err := VerifyBundle(VerificationInput{
		Bundle:          bundle,
		ContentType:     contentType,
		ArchiveLimits:   s.deps.Config.Limits.Archives,
		ReferenceLimits: s.deps.Config.Limits.ReferencePacks,
	})
	if err != nil {
		var verificationErr *VerificationError
		if errors.As(err, &verificationErr) {
			return nil, verificationErr
		}
		return nil, &VerificationError{ReasonCode: "payload_missing"}
	}
	return &result, nil
}

func (s *Service) verifyStoredBundle(record VersionRecord) (*VerificationResult, *VerificationError) {
	data, err := os.ReadFile(record.BundleStoragePath)
	if err != nil {
		return nil, &VerificationError{ReasonCode: "payload_missing"}
	}
	return s.verifyUpload(data, MediaTypeOctetStream)
}

func (s *Service) stageBundle(fileSHA string, data []byte) (string, error) {
	root := s.deps.Config.Roots.TemporaryWork.Path
	if strings.TrimSpace(root) == "" {
		root = os.TempDir()
	}
	bundleDir := filepath.Join(root, "reference-packs", "imports")
	if err := os.MkdirAll(bundleDir, 0o700); err != nil {
		return "", err
	}
	name := fileSHA
	if strings.TrimSpace(name) == "" {
		name = uuid.NewString()
	}
	path := filepath.Join(bundleDir, name+"-"+uuid.NewString()+".bundle")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Service) requireDeploymentAdmin(r *http.Request, stateChanging bool) (httpauth.Principal, *httpapi.APIError) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: stateChanging})
	if apiErr != nil {
		return httpauth.Principal{}, apiErr
	}
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
		return httpauth.Principal{}, apiErr
	}
	return principal, nil
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	if !httpauth.ShouldSlideIdleExpiry(method, path) {
		return nil
	}
	sliding := authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}
	persisted, err := s.authStore.SlideSession(ctx, principal.Session.ID, sliding)
	if err != nil {
		return err
	}
	principal.Session.LastQualifyingActivityAt = persisted.LastQualifyingActivityAt
	principal.Session.IdleExpiresAt = persisted.IdleExpiresAt
	principal.Session.SessionExpiresAt = persisted.SessionExpiresAt
	return nil
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
		if parts[1] == "activate" || parts[1] == "disable" || parts[1] == "reverify" {
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
		case "activate", "disable", "reverify":
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
