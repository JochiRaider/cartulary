package incidentbundles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

const incidentBundleJobHandlerName = BundleWorkerKind

type incidentBundleJobRunner interface {
	RegisterHandler(string, jobs.HandlerFunc) error
	RecoverHandler(context.Context, string) error
	DispatchJobID(string, uuid.UUID) error
}

type incidentBundleWorker struct {
	store             *Store
	jobManager        *jobs.Manager
	jobRunner         incidentBundleJobRunner
	results           incidentBundleJobResultSink
	storage           BundleStorage
	importFinalizer   incidents.IncidentBundleImportFinalizer
	jobFinalizer      JobSuccessFinalizer
	portability       *PortabilityOrchestrator
	transactions      *crossownertransaction.Coordinator
	projectionRebuild importProjectionRebuilder
	sourceCatalog     *sourceport.Catalog
	historicalIntents historicalIntentPolicy
	limits            Limits
	deps              httpapi.DependencySet
	now               func() time.Time
}

func newIncidentBundleWorker(store *Store, deps httpapi.DependencySet, storage BundleStorage, importFinalizer incidents.IncidentBundleImportFinalizer, jobFinalizer JobSuccessFinalizer, portability *PortabilityOrchestrator, transactions *crossownertransaction.Coordinator, projectionRebuild importProjectionRebuilder, sourceCatalog *sourceport.Catalog, historicalIntents historicalIntentPolicy, limits Limits, now func() time.Time) *incidentBundleWorker {
	return &incidentBundleWorker{
		store:             store,
		jobManager:        deps.Jobs,
		jobRunner:         deps.JobRunner,
		results:           incidentBundleJobResultSink{manager: deps.Jobs, store: store, now: now},
		storage:           storage,
		importFinalizer:   importFinalizer,
		jobFinalizer:      jobFinalizer,
		portability:       portability,
		transactions:      transactions,
		projectionRebuild: projectionRebuild,
		sourceCatalog:     sourceCatalog,
		historicalIntents: historicalIntents,
		limits:            limits,
		deps:              deps,
		now:               now,
	}
}

func (w *incidentBundleWorker) registerJobHandler() error {
	if w == nil || w.jobRunner == nil {
		return jobs.ErrNotConfigured
	}
	return w.jobRunner.RegisterHandler(incidentBundleJobHandlerName, w.executeJobID)
}

func (w *incidentBundleWorker) recoverJobs(ctx context.Context) error {
	if w == nil || w.jobRunner == nil {
		return jobs.ErrNotConfigured
	}
	return w.jobRunner.RecoverHandler(ctx, incidentBundleJobHandlerName)
}

func (w *incidentBundleWorker) dispatch(jobIDText string) error {
	if w == nil || w.jobRunner == nil {
		return jobs.ErrNotConfigured
	}
	jobID, err := uuid.Parse(jobIDText)
	if err != nil {
		return err
	}
	return w.jobRunner.DispatchJobID(incidentBundleJobHandlerName, jobID)
}

func (w *incidentBundleWorker) executeJobID(ctx context.Context, jobID uuid.UUID) error {
	payload, err := w.store.GetJobPayload(ctx, jobID)
	if err != nil {
		w.results.completeFailed(ctx, failedTransition(jobID, "internal_error", map[string]any{}))
		return fmt.Errorf("load incident bundle job payload: %w", err)
	}
	w.executePayload(ctx, payload)
	return nil
}

func (w *incidentBundleWorker) executePayload(ctx context.Context, payload JobPayload) {
	switch payload.JobKind {
	case "export":
		w.executeExportJob(ctx, payload)
	case "import":
		w.executeImportJob(ctx, payload)
	default:
		w.results.completeFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{"job_kind": payload.JobKind}))
	}
}

func (w *incidentBundleWorker) executeExportJob(ctx context.Context, payload JobPayload) {
	if payload.IncidentID == nil {
		w.results.completeFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
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
		w.results.completeFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
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
	builder := BundleBuilder{pool: w.deps.Postgres, objectStore: w.deps.ObjectStore, portability: w.portability, sourceCatalog: w.sourceCatalog}
	built, err := builder.Build(ctx, *payload.IncidentID, request, bundleID, exportedAt)
	if err != nil {
		w.results.completeFailedFromError(ctx, payload.JobID, "incident_bundle_export_rejected", err)
		return
	}
	storageReference, err := w.storage.Publish(ctx, bundleID.String(), built.Archive.Bytes)
	if err != nil {
		w.results.completeFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	completeParams := ExportCompleteParams{
		JobID: payload.JobID, ActorUserID: payload.ActorUserID,
		IncidentID: *payload.IncidentID, BundleID: bundleID, ExportedAt: exportedAt,
		ManifestSHA256:       built.Archive.ManifestSHA256,
		ReferencePackMode:    built.Archive.Manifest.ReferencePackMode,
		OptionalSections:     built.Archive.Manifest.OptionalSections,
		RequiredCapabilities: built.Archive.Manifest.RequiredCapabilities,
		BundleSHA256:         built.BundleSHA256, BundleByteSize: built.BundleByteSize,
		BundleStorageRef: storageReference, ManifestFiles: built.Archive.Manifest.Files,
	}
	var record DescriptorRecord
	_, err = w.jobFinalizer.FinalizeIncidentBundleJobSuccess(ctx, JobSuccessFinalization{
		Transition: exportSuccessTransition(payload.JobID, bundleID),
		FinalCommitID: "incident_portability.export:" +
			bundleID.String() + ":" + built.Archive.ManifestSHA256,
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			var mutationErr error
			record, mutationErr = w.store.CompleteExportDescriptorTx(ctx, tx, completeParams)
			return mutationErr
		},
	})
	if err != nil {
		if !errors.Is(err, ErrJobFinalizationIndeterminate) {
			_ = w.storage.RemovePublished(storageReference)
			w.results.completeFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		}
		return
	}
	if record.BundleID != bundleID {
		return
	}
}

func (w *incidentBundleWorker) executeImportJob(ctx context.Context, payload JobPayload) {
	if payload.BundleStagingRef == nil {
		w.results.completeFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	defer func() {
		_ = w.storage.RemoveStaged(*payload.BundleStagingRef)
	}()
	if !w.markJobRunningOrResume(ctx, payload.JobID, 1) {
		return
	}
	data, err := w.storage.ReadStaged(*payload.BundleStagingRef, incidentBundleStagingReadLimit(w.limits.IncidentBundles.MaxExtractedBytes))
	if err != nil {
		w.results.completeFailed(ctx, failedTransition(payload.JobID, "incident_bundle_import_rejected", map[string]any{"reason_code": "missing_required_file"}))
		return
	}
	verified, err := VerifyBundle(VerificationInput{Bundle: data, Limits: w.limits})
	if err != nil {
		w.results.completeFailedFromError(ctx, payload.JobID, "incident_bundle_import_rejected", err)
		return
	}
	requestID := incidents.ImportBundleRequestID(payload.JobID)
	importer := Importer{
		pool:              w.deps.Postgres,
		objectStore:       w.deps.ObjectStore,
		finalizer:         w.importFinalizer,
		projectionRebuild: w.projectionRebuild,
		sourceCatalog:     w.sourceCatalog,
		historicalIntents: w.historicalIntents,
	}
	importParams := ImportParams{
		ActorUserID: payload.ActorUserID,
		PublishedAt: w.now().UTC(),
		RequestID:   &requestID,
		OperationID: payload.JobID.String(),
	}
	prepared, err := importer.PrepareImport(ctx, verified, importParams)
	if err != nil {
		var verificationErr *VerificationError
		if errors.As(err, &verificationErr) {
			w.results.completeFailedFromError(ctx, payload.JobID, "incident_bundle_import_rejected", err)
			return
		}
		w.results.completeFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	committed := false
	finalityUnknown := false
	defer func() {
		if !committed && !finalityUnknown {
			prepared.Cleanup(context.WithoutCancel(ctx))
		}
	}()
	portability, err := w.portability.PrepareImport(ctx, payload.JobID.String(), prepared.IncidentID, verified.Files)
	if err != nil {
		w.results.completeFailedFromError(ctx, payload.JobID, "incident_bundle_import_rejected", err)
		return
	}
	defer func() {
		if !committed && !finalityUnknown {
			_ = portability.Abandon(context.WithoutCancel(ctx))
		}
	}()
	coreParticipant, err := NewImportTransactionParticipant(prepared, importParams, payload.JobID, verified.ManifestSHA256)
	if err != nil {
		w.results.completeFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	participants := append([]crossownertransaction.Participant{coreParticipant}, portability.Participants...)
	result, err := w.transactions.Execute(ctx, crossownertransaction.Operation{
		OperationID: payload.JobID.String(), NormalizedRequestSHA256: verified.ManifestSHA256,
		Participants: participants,
		Finalizer: importJobTransactionFinalizer{
			finalizer: w.jobFinalizer, jobID: payload.JobID,
			manifestSHA256: verified.ManifestSHA256,
		},
	})
	if err != nil {
		if crossownertransaction.IsFatalIntegrity(err) {
			finalityUnknown = true
			return
		}
		var verificationErr *VerificationError
		if errors.As(err, &verificationErr) {
			w.results.completeFailedFromError(ctx, payload.JobID, "incident_bundle_import_rejected", verificationErr)
			return
		}
		w.results.completeFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	value, ok := result.ParticipantValues[ImportTransactionParticipantID].(ImportTransactionResult)
	if !ok || value.IncidentID != prepared.IncidentID {
		w.results.completeFailed(ctx, failedTransition(payload.JobID, "internal_error", map[string]any{}))
		return
	}
	committed = true
	prepared.stagedObjectKeys = nil
	portability.Committed()
	if verified.Manifest.BundleVersion == 1 {
		telemetry.RecordIncidentBundleV1Import(context.WithoutCancel(ctx), "")
	}
}

func incidentBundleStagingReadLimit(maxExtractedBytes int64) int64 {
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

type incidentBundleJobResultSink struct {
	manager *jobs.Manager
	store   *Store
	now     func() time.Time
}

func (s incidentBundleJobResultSink) completeFailed(ctx context.Context, params jobs.TransitionParams) {
	_, _ = s.manager.CompleteFailed(ctx, params)
}

func (s incidentBundleJobResultSink) completeFailedFromError(ctx context.Context, jobID uuid.UUID, code string, err error) {
	reason, details := incidentBundleFailureDetails(code, err)
	s.store.MarkJobFailure(ctx, jobID, reason, s.now())
	_, _ = s.manager.CompleteFailed(ctx, failedTransition(jobID, code, details))
}

func incidentBundleFailureDetails(code string, err error) (string, map[string]any) {
	reason := "missing_required_file"
	details := map[string]any{}
	var verificationErr *VerificationError
	if errors.As(err, &verificationErr) {
		reason = verificationErr.ReasonCode
		if reason == "source_family_invalid" &&
			verificationErr.SourceFamilyID != "" && verificationErr.InvariantID != "" {
			details["source_family_id"] = verificationErr.SourceFamilyID
			details["invariant_id"] = verificationErr.InvariantID
		}
	} else if code == "incident_bundle_import_rejected" {
		invariantID := ""
		switch {
		case errors.Is(err, ErrPortabilityUnavailable), errors.Is(err, ErrPortabilityBlocked):
			invariantID = "extension_payload.participant_admitted"
		case errors.Is(err, ErrPortabilityLimit):
			invariantID = "extension_payload.resource_bounded"
		case errors.Is(err, ErrPortabilityPayload):
			invariantID = "extension_payload.schema_digest_valid"
		case errors.Is(err, ErrPortabilityResult):
			invariantID = "extension_payload.contract_compatible"
		}
		if invariantID != "" {
			reason = "source_family_invalid"
			details["source_family_id"] = "extension_payload"
			details["invariant_id"] = invariantID
		}
	}
	details["reason_code"] = reason
	return reason, details
}

func exportSuccessTransition(jobID uuid.UUID, bundleID uuid.UUID) jobs.TransitionParams {
	return jobs.TransitionParams{
		JobID:    jobID,
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ResultSummary: &jobs.ResultSummary{
			Code:    ResultIncidentBundleExported,
			Message: "Incident bundle exported.",
			ResourceRefs: []jobs.ResourceRef{{
				Kind:  "incident_bundle",
				ID:    bundleID.String(),
				Route: "/api/v1/incident-bundles/" + bundleID.String(),
			}},
		},
	}
}

func importSuccessTransition(jobID uuid.UUID, incidentID uuid.UUID) jobs.TransitionParams {
	return jobs.TransitionParams{
		JobID:    jobID,
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
