package reference_data

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
	"github.com/jackc/pgx/v5"
)

type Service struct {
	store               *Store
	authStore           *authn.Store
	jobManager          *jobs.Manager
	jobRunner           *jobs.Runner
	jobSuccessFinalizer JobSuccessFinalizer
	hub                 *platformws.Hub
	keys                authn.MasterKeys
	cursorCodec         *pagination.Codec
	storage             Storage
	limits              Limits
	deps                httpapi.DependencySet
	now                 func() time.Time
}

type RouteOption func(*routeOptions)

type routeOptions struct {
	jobSuccessFinalizer JobSuccessFinalizer
	storage             Storage
	limits              Limits
}

func WithJobSuccessFinalizer(finalizer JobSuccessFinalizer) RouteOption {
	return func(options *routeOptions) {
		options.jobSuccessFinalizer = finalizer
	}
}

func WithStorage(storage Storage) RouteOption {
	return func(options *routeOptions) {
		options.storage = storage
	}
}

func WithLimits(limits Limits) RouteOption {
	return func(options *routeOptions) {
		options.limits = limits
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
		if err := service.recoverReferencePackJobs(context.Background()); err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.reference_data", map[string]http.HandlerFunc{
			"activateReferencePackVersion": service.handleMember,
			"disableReferencePackVersion":  service.handleMember,
			"getReferencePackVersion":      service.handleMember,
			"importReferencePack":          service.handleMember,
			"listReferencePacks":           service.handleCollection,
			"refreshReferencePacks":        service.handleMember,
			"reverifyReferencePackVersion": service.handleMember,
		})
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
	if deps.Jobs != nil && options.jobSuccessFinalizer == nil {
		return nil, fmt.Errorf("Reference Pack admitted route requires a job success finalizer")
	}
	if deps.Jobs != nil && deps.JobRunner == nil {
		return nil, fmt.Errorf("Reference Pack admitted route requires the shared job runner")
	}
	if deps.Jobs != nil && options.storage == nil {
		return nil, fmt.Errorf("Reference Pack admitted route requires storage")
	}
	service := &Service{
		store:               NewStore(deps.Postgres),
		authStore:           authn.NewStore(deps.PostgresHandle()),
		jobManager:          deps.Jobs,
		jobRunner:           deps.JobRunner,
		jobSuccessFinalizer: options.jobSuccessFinalizer,
		hub:                 deps.WSHub,
		keys:                keys,
		cursorCodec:         cursorCodec,
		storage:             options.storage,
		limits:              options.limits,
		deps:                deps,
		now:                 now,
	}
	if err := service.registerJobHandler(); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *Service) registerJobHandler() error {
	if s == nil || s.jobRunner == nil {
		return nil
	}
	err := s.jobRunner.RegisterHandler(LifecycleWorkerKind, s.executeReferencePackJob)
	if errors.Is(err, jobs.ErrHandlerAlreadyRegistered) {
		return nil
	}
	return err
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
	stagingRef, err := s.storage.Stage(r.Context(), envelope.FileSHA256Hex, envelope.File)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := s.store.AcceptImport(r.Context(), ImportAcceptedParams{
		ActorUserID:       principal.User.ID,
		Request:           request,
		BundleSHA256:      envelope.FileSHA256Hex,
		BundleStagingRef:  stagingRef,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		_ = s.storage.RemoveStaged(stagingRef)
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		_ = s.storage.RemoveStaged(stagingRef)
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !result.Replayed {
		s.dispatchReferencePackJob(result.Job.JobID)
	} else {
		_ = s.storage.RemoveStaged(stagingRef)
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
		ActivationPreflight: func(record VersionRecord) error {
			if _, verificationErr := s.verifyStoredBundle(record); verificationErr != nil {
				return verificationErr
			}
			return nil
		},
		Now: s.now(),
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
	if s == nil || s.jobRunner == nil {
		return nil
	}
	return s.jobRunner.RecoverHandler(ctx, LifecycleWorkerKind)
}

func (s *Service) dispatchReferencePackJob(jobID string) {
	parsed, err := uuid.Parse(jobID)
	if err != nil {
		return
	}
	if s == nil || s.jobRunner == nil {
		return
	}
	_ = s.jobRunner.DispatchJobID(LifecycleWorkerKind, parsed)
}

func (s *Service) executeReferencePackJob(ctx context.Context, jobID uuid.UUID) error {
	payload, err := s.store.JobPayload(ctx, jobID)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(jobID, "internal_error", map[string]any{}))
		return fmt.Errorf("load reference pack job payload: %w", err)
	}
	runReferencePackWorkerStartHook(payload.JobKind)
	total := 1
	if payload.JobKind == "refresh" && len(payload.ResolvedPackKeys) > 0 {
		total = len(payload.ResolvedPackKeys)
	}
	if ok := s.markJobRunningOrResume(ctx, jobID, total); !ok {
		if payload.JobKind == "import" && payload.BundleStagingRef != nil {
			_ = s.storage.RemoveStaged(*payload.BundleStagingRef)
		}
		return nil
	}
	switch payload.JobKind {
	case "import":
		return s.executeImportJob(ctx, payload)
	case "reverify":
		return s.executeReverifyJob(ctx, payload)
	case "refresh":
		return s.executeRefreshJob(ctx, payload)
	default:
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(jobID, "internal_error", map[string]any{"job_kind": payload.JobKind}))
	}
	return nil
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

func (s *Service) executeImportJob(ctx context.Context, payload JobPayload) error {
	if payload.BundleStagingRef == nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return nil
	}
	removeStaging := true
	defer func() {
		if removeStaging {
			_ = s.storage.RemoveStaged(*payload.BundleStagingRef)
		}
	}()
	if s.jobCancelRequested(ctx, payload.JobID) {
		s.completeCanceled(ctx, payload.JobID, 0, 1)
		return nil
	}
	data, err := s.storage.ReadStaged(*payload.BundleStagingRef, referencePackStorageReadLimit(s.limits.ReferencePacks.MaxExtractedBytes))
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "reference_pack_verification_failed", map[string]any{"reason_code": "payload_missing"}))
		return nil
	}
	verification, verificationErr := s.verifyUpload(data, MediaTypeOctetStream)
	if verificationErr != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "reference_pack_verification_failed", map[string]any{"reason_code": verificationErr.ReasonCode}))
		return nil
	}
	if s.jobCancelRequested(ctx, payload.JobID) {
		s.completeCanceled(ctx, payload.JobID, 0, 1)
		return nil
	}
	bundleRef, err := s.storage.Publish(ctx, verification.BundleSHA256, data)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return nil
	}
	if s.jobCancelRequested(ctx, payload.JobID) {
		_ = s.storage.RemovePublished(bundleRef)
		s.completeCanceled(ctx, payload.JobID, 0, 1)
		return nil
	}
	err = s.finalizeReferencePackJobSuccess(ctx, payload, jobs.TransitionParams{
		JobID: payload.JobID, Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ResultSummary: &jobs.ResultSummary{
			Code: ResultReferencePackImported, Message: "Reference pack imported.",
			ResourceRefs: []jobs.ResourceRef{jobs.ResourceRef(referencePackResourceRef(verification.PackKey, verification.PackVersion))},
		},
	}, func(ctx context.Context, tx pgx.Tx) error {
		_, err := s.store.CompleteImportVerificationTx(ctx, tx, payload.JobID, payload.ActorUserID, *verification, bundleRef, s.now())
		return err
	})
	if err != nil {
		if !errors.Is(err, ErrJobFinalizationIndeterminate) {
			_ = s.storage.RemovePublished(bundleRef)
		}
		// Preserve staged input for retry. Indeterminate finalization also
		// preserves the complete published object because committed rows may
		// reference it.
		removeStaging = false
	}
	return err
}

func (s *Service) executeReverifyJob(ctx context.Context, payload JobPayload) error {
	if payload.PackKey == nil || payload.PackVersion == nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return nil
	}
	if s.jobCancelRequested(ctx, payload.JobID) {
		s.completeCanceled(ctx, payload.JobID, 0, 1)
		return nil
	}
	record, err := s.store.GetVersion(ctx, *payload.PackKey, *payload.PackVersion)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "reference_pack_not_found", map[string]any{}))
		return nil
	}
	verification, verificationErr := s.verifyStoredBundle(record)
	if s.jobCancelRequested(ctx, payload.JobID) {
		s.completeCanceled(ctx, payload.JobID, 0, 1)
		return nil
	}
	if verificationErr != nil {
		if _, err := s.store.ApplyVerificationResult(ctx, record, verification, verificationErr, "reverify", payload.ActorUserID, payload.JobID, s.now()); err != nil {
			_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
			return nil
		}
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "reference_pack_verification_failed", map[string]any{"reason_code": verificationErr.ReasonCode}))
		return nil
	}
	if s.jobCancelRequested(ctx, payload.JobID) {
		s.completeCanceled(ctx, payload.JobID, 0, 1)
		return nil
	}
	return s.finalizeReferencePackJobSuccess(ctx, payload, jobs.TransitionParams{
		JobID: payload.JobID, Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ResultSummary: &jobs.ResultSummary{
			Code: ResultReferencePackReverified, Message: "Reference pack reverified.",
			ResourceRefs: []jobs.ResourceRef{jobs.ResourceRef(referencePackResourceRef(record.PackKey, record.PackVersion))},
		},
	}, func(ctx context.Context, tx pgx.Tx) error {
		_, err := s.store.ApplyVerificationResultTx(ctx, tx, record, verification, nil, "reverify", payload.ActorUserID, payload.JobID, s.now())
		return err
	})
}

func (s *Service) executeRefreshJob(ctx context.Context, payload JobPayload) error {
	total := len(payload.ResolvedPackKeys)
	if total == 0 {
		return s.finalizeReferencePackJobSuccess(ctx, payload, jobs.TransitionParams{
			JobID: payload.JobID, Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
			ResultSummary: &jobs.ResultSummary{
				Code: ResultReferencePacksRefreshed, Message: "Reference packs refreshed.",
			},
		})
	}
	records, err := s.store.ListVersionsForPackKeys(ctx, payload.ResolvedPackKeys)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return nil
	}
	type verifiedRecord struct {
		record       VersionRecord
		verification *VerificationResult
	}
	verified := make([]verifiedRecord, 0, len(records))
	for i, record := range records {
		if s.jobCancelRequested(ctx, payload.JobID) {
			s.completeCanceled(ctx, payload.JobID, i, total)
			return nil
		}
		verification, verificationErr := s.verifyStoredBundle(record)
		if s.jobCancelRequested(ctx, payload.JobID) {
			s.completeCanceled(ctx, payload.JobID, i, total)
			return nil
		}
		if verificationErr != nil {
			if _, err := s.store.ApplyVerificationResult(ctx, record, verification, verificationErr, "refresh", payload.ActorUserID, payload.JobID, s.now()); err != nil {
				_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
				return nil
			}
			_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "reference_pack_verification_failed", map[string]any{"reason_code": verificationErr.ReasonCode}))
			return nil
		}
		verified = append(verified, verifiedRecord{record: record, verification: verification})
	}
	if s.jobCancelRequested(ctx, payload.JobID) {
		s.completeCanceled(ctx, payload.JobID, 0, total)
		return nil
	}
	return s.finalizeReferencePackJobSuccess(ctx, payload, jobs.TransitionParams{
		JobID: payload.JobID, Progress: jobs.Progress{Completed: total, Total: &total},
		ResultSummary: &jobs.ResultSummary{
			Code: ResultReferencePacksRefreshed, Message: "Reference packs refreshed.",
			ResourceRefs: sortedRefs(records),
		},
	}, func(ctx context.Context, tx pgx.Tx) error {
		for _, item := range verified {
			if _, err := s.store.ApplyVerificationResultTx(ctx, tx, item.record, item.verification, nil, "refresh", payload.ActorUserID, payload.JobID, s.now()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) finalizeReferencePackJobSuccess(
	ctx context.Context,
	payload JobPayload,
	transition jobs.TransitionParams,
	mutations ...JobSuccessMutation,
) error {
	if s.jobSuccessFinalizer == nil {
		return fmt.Errorf("Reference Pack job success finalizer is unavailable")
	}
	var mutation JobSuccessMutation
	if len(mutations) > 0 {
		mutation = mutations[0]
	}
	_, err := s.jobSuccessFinalizer.FinalizeReferencePackJobSuccess(ctx, JobSuccessFinalization{
		Transition:    transition,
		FinalCommitID: ProfileID + "." + payload.JobKind + ":" + payload.JobID.String(),
		Mutate:        mutation,
	})
	return err
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
		ArchiveLimits:   s.limits.Archives,
		ReferenceLimits: s.limits.ReferencePacks,
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
	data, err := s.storage.ReadPublished(record.BundleStorageRef, referencePackStorageReadLimit(s.limits.ReferencePacks.MaxExtractedBytes))
	if err != nil {
		return nil, &VerificationError{ReasonCode: "payload_missing"}
	}
	return s.verifyUpload(data, MediaTypeOctetStream)
}

func referencePackStorageReadLimit(maxExtractedBytes int64) int64 {
	const (
		archiveOverheadBytes int64 = 64 << 20
		maxInt64                   = int64(1<<63 - 1)
	)
	if maxExtractedBytes <= 0 {
		return archiveOverheadBytes
	}
	if maxExtractedBytes > maxInt64-archiveOverheadBytes {
		return maxInt64
	}
	return maxExtractedBytes + archiveOverheadBytes
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
