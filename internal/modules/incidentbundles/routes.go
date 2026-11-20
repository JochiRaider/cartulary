package incidentbundles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

type service struct {
	store          *store
	authStore      *authn.Store
	incidentAccess *admission.Checker
	storage        BundleStorage
	worker         *incidentBundleWorker
	now            func() time.Time
}

// RegisterRoutes returns the Incident Bundles registrar only after the module
// facade has completed its coordinator and worker lifecycle. Each operation is
// bound for central admission (Core 01 REQ-01-707): the registry resolves the
// route, authenticates by the declared security and admits bundle_id before a
// handler runs.
func (m *Module) RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if err := m.routeDependenciesReady(); err != nil {
			return err
		}
		service, err := newService(m)
		if err != nil {
			return err
		}
		return httpapi.BindAdmittedOwnerRoutes(mux, deps, "module.incidentbundles", map[string]http.HandlerFunc{
			"exportIncidentBundle":        service.handleExport,
			"getIncidentBundleDescriptor": service.handleGetDescriptor,
			"getIncidentBundleContent":    service.handleGetContent,
			"importIncidentBundle":        service.handleImport,
		})
	}
}

func newService(module *Module) (*service, error) {
	if module == nil || module.pool == nil || module.store == nil || module.worker == nil || module.storage == nil || module.now == nil {
		return nil, fmt.Errorf("incident bundle module is incomplete")
	}
	return &service{
		store:          module.store,
		authStore:      authn.NewStore(module.pool),
		incidentAccess: admission.NewChecker(module.pool),
		storage:        module.storage,
		worker:         module.worker,
		now:            module.now,
	}, nil
}

func (s *service) handleGetDescriptor(w http.ResponseWriter, r *http.Request) {
	s.serveBundle(w, r, false)
}

func (s *service) handleGetContent(w http.ResponseWriter, r *http.Request) {
	s.serveBundle(w, r, true)
}

func (s *service) serveBundle(w http.ResponseWriter, r *http.Request, content bool) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	bundleID, identified := httpapi.PathUUID(r, "bundle_id")
	if !identified {
		writeAPIError(w, r, internalAPIError())
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
	// REQ-01-704: the deployment role is decided only for a caller who can
	// see the bundle's incident; anyone else gets the missing bundle's response.
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-707 step 5: the singleton-read query check is route handling,
	// after the bundle's visibility and role gate.
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
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
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// The body's incident_id addresses the exported incident (REQ-01-484), so
	// the body is decoded before that incident's visibility can be decided.
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
	// Export is an incident-context operation that needs both current
	// membership and deployment_admin (REQ-01-608); the role is decided only
	// after visibility (REQ-01-704).
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *service) handleImport(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError())
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

// admittedPrincipal returns the principal central admission established; its
// absence is a binding error.
func admittedPrincipal(r *http.Request) (httpauth.Principal, *httpapi.APIError) {
	principal, authenticated := httpauth.AdmittedPrincipal(r)
	if !authenticated {
		return httpauth.Principal{}, internalAPIError()
	}
	return principal, nil
}

func (s *service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, s.now)
}
