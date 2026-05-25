package incidentbundles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type Service struct {
	store         *Store
	authStore     *authn.Store
	incidentStore *incidents.Store
	jobManager    *jobs.Manager
	jobRunner     *jobs.Runner
	hub           *platformws.Hub
	keys          authn.MasterKeys
	deps          httpapi.DependencySet
	now           func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.ExtensionProfileClaimed(ProfileID) {
			return nil
		}
		service, err := newService(deps)
		if err != nil {
			return err
		}
		if err := service.recoverIncidentBundleJobs(context.Background()); err != nil {
			return err
		}
		mux.HandleFunc("/api/v1/incident-bundles/export", service.handleExport)
		mux.HandleFunc("/api/v1/incident-bundles/import", service.handleImport)
		mux.HandleFunc("/api/v1/incident-bundles/", service.handleBundleMember)
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
	return &Service{
		store:         NewStore(deps.Postgres),
		authStore:     authn.NewStore(deps.Postgres),
		incidentStore: incidents.NewStore(deps.Postgres),
		jobManager:    deps.Jobs,
		jobRunner:     deps.JobRunner,
		hub:           deps.WSHub,
		keys:          keys,
		deps:          deps,
		now:           now,
	}, nil
}

func (s *Service) handleBundleMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	bundleIDText := strings.TrimPrefix(r.URL.Path, "/api/v1/incident-bundles/")
	if bundleIDText == "" || strings.Contains(bundleIDText, "/") {
		http.NotFound(w, r)
		return
	}
	bundleID, err := uuid.Parse(bundleIDText)
	if err != nil {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	}
	if apiErr := auth.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.store.GetDescriptor(r.Context(), bundleID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, err := s.incidentStore.GetIncidentMembershipForUser(r.Context(), record.IncidentID, principal.User.ID); err != nil {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, record.Resource())
}

func (s *Service) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeExportRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, err := s.incidentStore.GetIncidentMembershipForUser(r.Context(), request.IncidentID, principal.User.ID); err != nil {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	}
	result, err := s.store.AcceptExport(r.Context(), ExportAcceptedParams{
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
		s.dispatchIncidentBundleJob(result.Job.JobID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *Service) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	envelope, envelopeErr := httpapi.ParseUploadEnvelope(r, httpapi.UploadEnvelopePolicy{FileContentTypes: IncidentBundleFileContentTypes})
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
		UploadedSHA256:    envelope.FileSHA256Hex,
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
		s.dispatchIncidentBundleJob(result.Job.JobID)
	} else {
		_ = os.Remove(stagingPath)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *Service) recoverIncidentBundleJobs(ctx context.Context) error {
	payloads, err := s.store.ListRecoverableJobPayloads(ctx)
	if err != nil {
		return err
	}
	for _, payload := range payloads {
		payload := payload
		if err := s.jobRunner.Dispatch(func(ctx context.Context) {
			s.executeIncidentBundlePayload(ctx, payload)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) dispatchIncidentBundleJob(jobIDText string) {
	jobID, err := uuid.Parse(jobIDText)
	if err != nil {
		return
	}
	_ = s.jobRunner.Dispatch(func(ctx context.Context) {
		payload, err := s.store.GetJobPayload(ctx, jobID)
		if err != nil {
			_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(jobID, "internal_error", map[string]any{}))
			return
		}
		s.executeIncidentBundlePayload(ctx, payload)
	})
}

func (s *Service) executeIncidentBundlePayload(ctx context.Context, payload JobPayload) {
	switch payload.JobKind {
	case "export":
		s.executeExportJob(ctx, payload)
	case "import":
		s.executeImportJob(ctx, payload)
	default:
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{"job_kind": payload.JobKind}))
	}
}

func (s *Service) executeExportJob(ctx context.Context, payload JobPayload) {
	if payload.IncidentID == nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	if !s.markJobRunningOrResume(ctx, payload.JobID, 1) {
		return
	}
	var normalized struct {
		ReferencePackMode    string   `json:"reference_pack_mode"`
		OptionalSections     []string `json:"optional_sections"`
		RequiredCapabilities []string `json:"required_capabilities"`
	}
	if err := json.Unmarshal(payload.RequestJSON, &normalized); err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	request := ExportRequest{
		IncidentID:           *payload.IncidentID,
		ReferencePackMode:    normalized.ReferencePackMode,
		OptionalSections:     normalized.OptionalSections,
		RequiredCapabilities: normalized.RequiredCapabilities,
		HistoryMode:          HistoryModeFull,
		BlobMode:             BlobModeFull,
	}
	bundleID := uuid.New()
	exportedAt := s.now().UTC()
	builder := BundleBuilder{pool: s.deps.Postgres, objectStore: s.deps.ObjectStore}
	built, err := builder.Build(ctx, *payload.IncidentID, request, bundleID, exportedAt)
	if err != nil {
		s.completeFailedFromError(ctx, payload.JobID, "incident_bundle_export_rejected", err)
		return
	}
	storagePath, err := s.persistBundle(bundleID.String(), built.Archive.Bytes)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	record, err := s.store.CompleteExportDescriptor(ctx, ExportCompleteParams{
		JobID:                payload.JobID,
		ActorUserID:          payload.ActorUserID,
		IncidentID:           *payload.IncidentID,
		BundleID:             bundleID,
		ExportedAt:           exportedAt,
		ManifestSHA256:       built.Archive.ManifestSHA256,
		ReferencePackMode:    built.Archive.Manifest.ReferencePackMode,
		OptionalSections:     built.Archive.Manifest.OptionalSections,
		RequiredCapabilities: built.Archive.Manifest.RequiredCapabilities,
		BundleSHA256:         built.BundleSHA256,
		BundleByteSize:       built.BundleByteSize,
		BundleStoragePath:    storagePath,
		ManifestFiles:        built.Archive.Manifest.Files,
	})
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	_, _ = s.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
		JobID:    payload.JobID,
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ResultSummary: &jobs.ResultSummary{
			Code:    ResultIncidentBundleExported,
			Message: "Incident bundle exported.",
			ResourceRefs: []jobs.ResourceRef{{
				Kind:  "incident_bundle",
				ID:    record.BundleID.String(),
				Route: "/api/v1/incident-bundles/" + record.BundleID.String(),
			}},
		},
	})
}

func (s *Service) executeImportJob(ctx context.Context, payload JobPayload) {
	if payload.BundleStagingPath == nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	defer func() { _ = os.Remove(*payload.BundleStagingPath) }()
	if !s.markJobRunningOrResume(ctx, payload.JobID, 1) {
		return
	}
	data, err := os.ReadFile(*payload.BundleStagingPath)
	if err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "incident_bundle_import_rejected", map[string]any{"reason_code": "missing_required_file"}))
		return
	}
	verified, err := VerifyBundle(VerificationInput{Bundle: data, Limits: s.deps.Config.Limits})
	if err != nil {
		s.completeFailedFromError(ctx, payload.JobID, "incident_bundle_import_rejected", err)
		return
	}
	importer := Importer{pool: s.deps.Postgres, objectStore: s.deps.ObjectStore}
	incidentID, err := importer.Import(ctx, verified, payload.ActorUserID)
	if err != nil {
		s.completeFailedFromError(ctx, payload.JobID, "incident_bundle_import_rejected", err)
		return
	}
	if err := s.store.MarkImportComplete(ctx, payload.JobID, incidentID, verified.ManifestSHA256, s.now()); err != nil {
		_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	_, _ = s.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
		JobID:    payload.JobID,
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ResultSummary: &jobs.ResultSummary{
			Code:    ResultIncidentBundleImported,
			Message: "Incident bundle imported.",
			ResourceRefs: []jobs.ResourceRef{{
				Kind:  "incident",
				ID:    incidentID.String(),
				Route: "/api/v1/incidents/" + incidentID.String(),
			}},
		},
	})
}

func (s *Service) completeFailedFromError(ctx context.Context, jobID uuid.UUID, code string, err error) {
	reason := "missing_required_file"
	var verificationErr *VerificationError
	if errors.As(err, &verificationErr) {
		reason = verificationErr.ReasonCode
	}
	s.store.MarkJobFailure(ctx, jobID, reason, s.now())
	_, _ = s.jobManager.CompleteFailed(ctx, failedTransition(jobID, code, map[string]any{"reason_code": reason}))
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
		_, _ = s.jobManager.CompleteCanceled(ctx, jobs.TransitionParams{JobID: jobID, Progress: jobs.Progress{Completed: 0, Total: &total}})
		return false
	default:
		return false
	}
}

func (s *Service) stageBundle(fileSHA string, data []byte) (string, error) {
	root := s.deps.Config.Roots.TemporaryWork.Path
	if strings.TrimSpace(root) == "" {
		root = os.TempDir()
	}
	bundleDir := filepath.Join(root, "incident-bundles", "imports")
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

func (s *Service) persistBundle(bundleID string, data []byte) (string, error) {
	root := s.deps.Config.Roots.ExportOutputs.Path
	if strings.TrimSpace(root) == "" {
		root = os.TempDir()
	}
	bundleDir := filepath.Join(root, "incident-bundles")
	if err := os.MkdirAll(bundleDir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(bundleDir, bundleID+".zip")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Service) requireDeploymentAdmin(r *http.Request, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	principal, apiErr := s.authenticateSessionRequest(r, stateChanging)
	if apiErr != nil {
		return auth.SessionPrincipal{}, apiErr
	}
	if apiErr := auth.RequireDeploymentAdmin(principal.User); apiErr != nil {
		return auth.SessionPrincipal{}, apiErr
	}
	return principal, nil
}

func (s *Service) authenticateSessionRequest(r *http.Request, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	return auth.AuthenticateSessionRequest(r, auth.SessionAuthOptions{
		Store:         s.authStore,
		Keys:          s.keys,
		Hub:           s.hub,
		Now:           s.now,
		StateChanging: stateChanging,
	})
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *auth.SessionPrincipal, method string, path string) error {
	if !auth.ShouldSlideIdleExpiry(method, path) {
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
