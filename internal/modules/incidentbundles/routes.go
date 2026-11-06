package incidentbundles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

type service struct {
	store          *store
	authStore      *authn.Store
	incidentAccess *admission.Checker
	storage        BundleStorage
	worker         *incidentBundleWorker
	keys           authn.MasterKeys
	now            func() time.Time
}

// RegisterRoutes returns the Incident Bundles registrar only after the module
// facade has completed its coordinator and worker lifecycle.
func (m *Module) RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if err := m.routeDependenciesReady(); err != nil {
			return err
		}
		service, err := newService(deps, m)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.incidentbundles", map[string]http.HandlerFunc{
			"exportIncidentBundle":        service.handleExport,
			"getIncidentBundleDescriptor": service.handleBundleMember,
			"getIncidentBundleContent":    service.handleBundleContent,
			"importIncidentBundle":        service.handleImport,
		})
	}
}

func newService(deps httpapi.DependencySet, module *Module) (*service, error) {
	if module == nil || module.pool == nil || module.store == nil || module.worker == nil || module.storage == nil || module.now == nil {
		return nil, fmt.Errorf("incident bundle module is incomplete")
	}
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	return &service{
		store:          module.store,
		authStore:      authn.NewStore(module.pool),
		incidentAccess: admission.NewChecker(module.pool),
		storage:        module.storage,
		worker:         module.worker,
		keys:           keys,
		now:            module.now,
	}, nil
}

func (s *service) handleBundleMember(w http.ResponseWriter, r *http.Request) {
	s.serveBundle(w, r, false)
}
func (s *service) handleBundleContent(w http.ResponseWriter, r *http.Request) {
	s.serveBundle(w, r, true)
}
func (s *service) serveBundle(w http.ResponseWriter, r *http.Request, content bool) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	bundleIDText := strings.TrimPrefix(r.URL.Path, "/api/v1/incident-bundles/")
	if content {
		bundleIDText = strings.TrimSuffix(bundleIDText, "/content")
	}
	if bundleIDText == "" || strings.Contains(bundleIDText, "/") {
		http.NotFound(w, r)
		return
	}
	bundleID, err := uuid.Parse(bundleIDText)
	if err != nil {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.store.getDescriptor(r.Context(), bundleID)
	if errors.Is(err, errNotFound) {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	if _, err := s.incidentAccess.Check(r.Context(), record.IncidentID, principal.User.ID, admission.Requirement{AllowedRoles: admission.RolesMember, Lifecycle: admission.LifecycleAny}); admission.IsDenied(err, admission.DenialNotVisible) {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	} else if err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	if content {
		if record.BundleStorageRef.String() != "incident-bundles/"+record.BundleID.String()+".zip" {
			writeAPIError(w, r, internalAPIError())
			return
		}
		reader, err := s.storage.OpenPublished(r.Context(), record.BundleStorageRef, ArchiveContent{ByteSize: record.BundleByteSize, SHA256: record.BundleSHA256})
		if err != nil {
			writeAPIError(w, r, internalAPIError())
			return
		}
		defer reader.Close()
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", `attachment; filename="incident-`+record.BundleID.String()+`.zip"`)
		w.Header().Set("Content-Length", strconv.FormatInt(record.BundleByteSize, 10))
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("ETag", `"sha256:`+record.BundleSHA256+`"`)
		if _, err := io.CopyBuffer(w, reader, make([]byte, 32<<10)); err != nil {
			panic(http.ErrAbortHandler)
		}
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, record.resource())
}

func (s *service) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeExportRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, err := s.incidentAccess.Check(r.Context(), request.IncidentID, principal.User.ID, admission.Requirement{AllowedRoles: admission.RolesMember, Lifecycle: admission.LifecycleAny}); admission.IsDenied(err, admission.DenialNotVisible) {
		writeAPIError(w, r, incidentBundleNotFound())
		return
	} else if err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	if request.CapabilityActivationRequested {
		writeAPIError(w, r, extensionCapabilityNotSupported())
		return
	}
	result, err := s.store.acceptExport(r.Context(), exportAcceptedParams{
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
		writeAPIError(w, r, internalAPIError())
		return
	}
	if !result.Replayed {
		_ = s.worker.dispatch(result.Job.JobID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *service) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := s.requireDeploymentAdmin(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	var staged StagedBundle
	accepted := false
	defer func() {
		if !accepted && staged.BundleStagingRef.String() != "" {
			_ = s.storage.RemoveStaged(staged.BundleStagingRef)
		}
	}()
	envelope, envelopeErr, receiveErr := httpapi.ParseStreamingUploadEnvelope(r,
		httpapi.UploadEnvelopePolicy{FileContentTypes: acceptedIncidentBundleFileContentTypes()}, PortabilityParticipantByteLimit,
		func(source io.Reader) (string, error) {
			var err error
			staged, err = s.storage.Stage(r.Context(), source, incidentBundleArchiveByteLimit(s.worker.limits.IncidentBundles.MaxExtractedBytes))
			return staged.SHA256, err
		})
	if receiveErr != nil {
		if errors.Is(receiveErr, ErrArchiveBytesExceeded) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "incident_bundle_import_rejected", Details: map[string]any{"reason_code": "archive_bytes_exceeded"}})
		} else {
			writeAPIError(w, r, internalAPIError())
		}
		return
	}
	if envelopeErr != nil {
		writeAPIError(w, r, uploadEnvelopeAPIError(envelopeErr))
		return
	}
	request, apiErr := decodeImportMetadata(envelope)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	stagingReference := staged.BundleStagingRef
	result, err := s.store.acceptImport(r.Context(), importAcceptedParams{
		ActorUserID:       principal.User.ID,
		Request:           request,
		UploadedSHA256:    envelope.FileSHA256Hex,
		BundleStagingRef:  stagingReference,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		if errors.Is(err, errJobAdmissionIndeterminate) {
			accepted = true
		}
		writeAPIError(w, r, internalAPIError())
		return
	}
	if !result.Replayed {
		accepted = true
		_ = s.worker.dispatch(result.Job.JobID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *service) requireDeploymentAdmin(r *http.Request, stateChanging bool) (httpauth.Principal, *httpapi.APIError) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: stateChanging})
	if apiErr != nil {
		return httpauth.Principal{}, apiErr
	}
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
		return httpauth.Principal{}, apiErr
	}
	return principal, nil
}

func (s *service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}
