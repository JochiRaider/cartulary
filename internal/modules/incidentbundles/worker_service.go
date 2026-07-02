package incidentbundles

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type incidentBundleWorker struct {
	store           *Store
	jobManager      *jobs.Manager
	jobRunner       *jobs.Runner
	files           *bundleFileStore
	importFinalizer incidents.IncidentBundleImportFinalizer
	deps            httpapi.DependencySet
	now             func() time.Time
}

func newIncidentBundleWorker(store *Store, deps httpapi.DependencySet, files *bundleFileStore, importFinalizer incidents.IncidentBundleImportFinalizer, now func() time.Time) *incidentBundleWorker {
	return &incidentBundleWorker{
		store:           store,
		jobManager:      deps.Jobs,
		jobRunner:       deps.JobRunner,
		files:           files,
		importFinalizer: importFinalizer,
		deps:            deps,
		now:             now,
	}
}

func (w *incidentBundleWorker) recoverJobs(ctx context.Context) error {
	payloads, err := w.store.ListRecoverableJobPayloads(ctx)
	if err != nil {
		return err
	}
	for _, payload := range payloads {
		payload := payload
		if err := w.jobRunner.Dispatch(func(ctx context.Context) {
			w.executePayload(ctx, payload)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (w *incidentBundleWorker) dispatch(jobIDText string) {
	jobID, err := uuid.Parse(jobIDText)
	if err != nil {
		return
	}
	_ = w.jobRunner.Dispatch(func(ctx context.Context) {
		payload, err := w.store.GetJobPayload(ctx, jobID)
		if err != nil {
			_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(jobID, "internal_error", map[string]any{}))
			return
		}
		w.executePayload(ctx, payload)
	})
}

func (w *incidentBundleWorker) executePayload(ctx context.Context, payload JobPayload) {
	runIncidentBundleWorkerStartHook(payload.JobKind)
	switch payload.JobKind {
	case "export":
		w.executeExportJob(ctx, payload)
	case "import":
		w.executeImportJob(ctx, payload)
	default:
		_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{"job_kind": payload.JobKind}))
	}
}

func (w *incidentBundleWorker) executeExportJob(ctx context.Context, payload JobPayload) {
	if payload.IncidentID == nil {
		_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	if !w.markJobRunningOrResume(ctx, payload.JobID, 1) {
		return
	}
	var normalized struct {
		ReferencePackMode    string   `json:"reference_pack_mode"`
		OptionalSections     []string `json:"optional_sections"`
		RequiredCapabilities []string `json:"required_capabilities"`
	}
	if err := json.Unmarshal(payload.RequestJSON, &normalized); err != nil {
		_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
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
	exportedAt := w.now().UTC()
	builder := BundleBuilder{pool: w.deps.Postgres, objectStore: w.deps.ObjectStore}
	built, err := builder.Build(ctx, *payload.IncidentID, request, bundleID, exportedAt)
	if err != nil {
		w.completeFailedFromError(ctx, payload.JobID, "incident_bundle_export_rejected", err)
		return
	}
	storagePath, err := w.files.persistBundle(bundleID.String(), built.Archive.Bytes)
	if err != nil {
		_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	record, err := w.store.CompleteExportDescriptor(ctx, ExportCompleteParams{
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
		_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	_, _ = w.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
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

func (w *incidentBundleWorker) executeImportJob(ctx context.Context, payload JobPayload) {
	if payload.BundleStagingPath == nil {
		_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	defer w.files.remove(*payload.BundleStagingPath)
	if !w.markJobRunningOrResume(ctx, payload.JobID, 1) {
		return
	}
	data, err := os.ReadFile(*payload.BundleStagingPath)
	if err != nil {
		_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "incident_bundle_import_rejected", map[string]any{"reason_code": "missing_required_file"}))
		return
	}
	verified, err := VerifyBundle(VerificationInput{Bundle: data, Limits: w.deps.Config.Limits})
	if err != nil {
		w.completeFailedFromError(ctx, payload.JobID, "incident_bundle_import_rejected", err)
		return
	}
	requestID := incidents.ImportBundleRequestID(payload.JobID)
	importer := Importer{
		pool:              w.deps.Postgres,
		objectStore:       w.deps.ObjectStore,
		finalizer:         w.importFinalizer,
		projectionRebuild: projectionadapters.NewIncidentImportRebuilder(w.deps.Postgres),
	}
	incidentID, err := importer.Import(ctx, verified, ImportParams{
		ActorUserID: payload.ActorUserID,
		PublishedAt: w.now().UTC(),
		RequestID:   &requestID,
	})
	if err != nil {
		var verificationErr *VerificationError
		if errors.As(err, &verificationErr) {
			w.completeFailedFromError(ctx, payload.JobID, "incident_bundle_import_rejected", err)
			return
		}
		_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	if err := w.store.MarkImportComplete(ctx, payload.JobID, incidentID, verified.ManifestSHA256, w.now()); err != nil {
		_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	_, _ = w.jobManager.CompleteSucceeded(ctx, jobs.TransitionParams{
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

func (w *incidentBundleWorker) completeFailedFromError(ctx context.Context, jobID uuid.UUID, code string, err error) {
	reason := "missing_required_file"
	var verificationErr *VerificationError
	if errors.As(err, &verificationErr) {
		reason = verificationErr.ReasonCode
	}
	w.store.MarkJobFailure(ctx, jobID, reason, w.now())
	_, _ = w.jobManager.CompleteFailed(ctx, failedTransition(jobID, code, map[string]any{"reason_code": reason}))
}

func (w *incidentBundleWorker) markJobRunningOrResume(ctx context.Context, jobID uuid.UUID, total int) bool {
	if total <= 0 {
		total = 1
	}
	if _, err := w.jobManager.MarkRunning(ctx, jobID, jobs.Progress{Completed: 0, Total: &total}, nil); err == nil {
		return true
	}
	job, err := w.jobManager.Get(ctx, jobID)
	if err != nil {
		return false
	}
	switch job.Status {
	case jobs.StatusRunning:
		return true
	case jobs.StatusCancelRequested:
		_, _ = w.jobManager.CompleteCanceled(ctx, jobs.TransitionParams{JobID: jobID, Progress: jobs.Progress{Completed: 0, Total: &total}})
		return false
	default:
		return false
	}
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
