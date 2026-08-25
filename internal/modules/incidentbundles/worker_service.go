package incidentbundles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/importfinalizerport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const incidentBundleJobHandlerName = bundleWorkerKind

func importBundleRequestID(jobID uuid.UUID) string {
	return "incident_bundle_import:" + jobID.String()
}

type workerDiagnosticError struct {
	classification string
	jobID          uuid.UUID
}

func (e *workerDiagnosticError) Error() string {
	return fmt.Sprintf("incident bundle worker failure: class=%s job_id=%s", e.classification, e.jobID)
}

func newWorkerDiagnosticError(classification string, jobID uuid.UUID) error {
	switch classification {
	case "execution_observation_failed", "payload_load_failed":
	default:
		classification = "unclassified_failure"
	}
	return &workerDiagnosticError{classification: classification, jobID: jobID}
}

// JobRunner is the Jobs-owned handler registration and wake-up capability.
type JobRunner interface {
	RegisterHandler(string, jobs.HandlerFunc) error
	Notify(uuid.UUID)
}

// JobOperations is the Jobs-owned execution lifecycle capability.
type JobOperations interface {
	Get(context.Context, uuid.UUID) (jobs.Resource, error)
	ObserveExecution(context.Context, jobs.Execution) (jobs.Resource, error)
	UpdateProgress(context.Context, jobs.Execution, jobs.Progress, *string) (jobs.Resource, error)
	CompleteFailed(context.Context, jobs.Execution, jobs.FailureCompletion) (jobs.Resource, error)
	CompleteCanceled(context.Context, jobs.Execution, jobs.CancellationCompletion) (jobs.Resource, error)
}

// JobTransactions is the Jobs-owned transaction capability required for
// admission and import execution validation.
type JobTransactions interface {
	CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error)
	ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
}

// IncidentPublicationLock is the incident-owner serialization capability used
// by the final bundle publication transaction.
type IncidentPublicationLock interface {
	LockIncidentTx(context.Context, pgx.Tx, uuid.UUID) (bool, error)
}

type sourcePortCatalog interface {
	Ports() []sourceport.Port
}

type portabilityCoordinator interface {
	Export(context.Context, StatePresenceQuery, uuid.UUID) ([]ExtensionPayload, error)
	PrepareImport(context.Context, string, uuid.UUID, map[string][]byte) (PreparedPortability, error)
	ValidatePublication(context.Context, StatePresenceQuery, uuid.UUID) error
}

type transactionCoordinator interface {
	Execute(context.Context, crossownertransaction.Operation) (crossownertransaction.Result, error)
}

type incidentBundleWorker struct {
	store             *store
	pool              *pgxpool.Pool
	jobManager        JobOperations
	jobRunner         JobRunner
	results           incidentBundleJobResultSink
	storage           BundleStorage
	importFinalizer   importfinalizerport.Finalizer
	jobFinalizer      JobSuccessFinalizer
	portability       portabilityCoordinator
	publicationLock   IncidentPublicationLock
	transactions      transactionCoordinator
	projectionRebuild ImportProjectionRebuilder
	sourceCatalog     sourcePortCatalog
	blobPort          BlobPortability
	limits            Limits
	now               func() time.Time
}

func newIncidentBundleWorker(store *store, pool *pgxpool.Pool, jobManager JobOperations, jobRunner JobRunner, storage BundleStorage, importFinalizer importfinalizerport.Finalizer, jobFinalizer JobSuccessFinalizer, portability *PortabilityOrchestrator, publicationLock IncidentPublicationLock, projectionRebuild ImportProjectionRebuilder, sourceCatalog *sourceport.Catalog, blobPort BlobPortability, limits Limits, now func() time.Time) *incidentBundleWorker {
	return &incidentBundleWorker{
		store:             store,
		pool:              pool,
		jobManager:        jobManager,
		jobRunner:         jobRunner,
		results:           incidentBundleJobResultSink{manager: jobManager, store: store, finalizer: jobFinalizer, now: now},
		storage:           storage,
		importFinalizer:   importFinalizer,
		jobFinalizer:      jobFinalizer,
		portability:       portability,
		publicationLock:   publicationLock,
		projectionRebuild: projectionRebuild,
		sourceCatalog:     sourceCatalog,
		blobPort:          blobPort,
		limits:            limits,
		now:               now,
	}
}

func (w *incidentBundleWorker) registerJobHandler() error {
	if w == nil || w.jobRunner == nil {
		return jobs.ErrNotConfigured
	}
	return w.jobRunner.RegisterHandler(incidentBundleJobHandlerName, w.executeJobID)
}

func (w *incidentBundleWorker) dispatch(jobIDText string) error {
	if w == nil || w.jobRunner == nil {
		return jobs.ErrNotConfigured
	}
	jobID, err := uuid.Parse(jobIDText)
	if err != nil {
		return err
	}
	w.jobRunner.Notify(jobID)
	return nil
}

func (w *incidentBundleWorker) executeJobID(ctx context.Context, execution jobs.Execution) error {
	jobID := execution.JobID()
	if _, err := w.jobManager.ObserveExecution(ctx, execution); err != nil {
		return newWorkerDiagnosticError("execution_observation_failed", jobID)
	}
	payload, err := w.store.getJobPayload(ctx, jobID)
	if err != nil {
		w.results.completeInternalFailure(ctx, execution)
		return newWorkerDiagnosticError("payload_load_failed", jobID)
	}
	w.executePayload(ctx, execution, payload)
	return nil
}

func (w *incidentBundleWorker) executePayload(ctx context.Context, execution jobs.Execution, payload jobPayload) {
	switch payload.JobKind {
	case "export":
		w.executeExportJob(ctx, execution, payload)
	case "import":
		w.executeImportJob(ctx, execution, payload)
	default:
		w.results.completeInternalFailure(ctx, execution)
	}
}

func (w *incidentBundleWorker) executeExportJob(ctx context.Context, execution jobs.Execution, payload jobPayload) {
	if payload.IncidentID == nil {
		w.results.completeInternalFailure(ctx, execution)
		return
	}
	if !w.prepareClaimedJob(ctx, execution, 1) {
		return
	}
	var normalized struct {
		ReferencePackMode    string   `json:"reference_pack_mode"`
		OptionalSections     []string `json:"optional_sections"`
		RequiredCapabilities []string `json:"required_capabilities"`
	}
	if err := json.Unmarshal(payload.RequestJSON, &normalized); err != nil {
		w.results.completeInternalFailure(ctx, execution)
		return
	}
	request := exportRequest{
		IncidentID:           *payload.IncidentID,
		ReferencePackMode:    normalized.ReferencePackMode,
		OptionalSections:     normalized.OptionalSections,
		RequiredCapabilities: normalized.RequiredCapabilities,
		HistoryMode:          historyModeFull,
		BlobMode:             blobModeFull,
	}
	bundleID := uuid.New()
	exportedAt := w.now().UTC()
	builder := bundleBuilder{pool: w.pool, blobPort: w.blobPort, portability: w.portability, sourceCatalog: w.sourceCatalog}
	built, err := builder.build(ctx, *payload.IncidentID, request, bundleID, exportedAt)
	if err != nil {
		w.results.completeFailedFromError(ctx, execution, "incident_bundle_export_rejected", err)
		return
	}
	storageReference, err := w.storage.Publish(ctx, bundleID.String(), built.Archive.Bytes)
	if err != nil {
		w.results.completeInternalFailure(ctx, execution)
		return
	}
	completeParams := exportCompleteParams{
		JobID: payload.JobID, ActorUserID: payload.ActorUserID,
		IncidentID: *payload.IncidentID, BundleID: bundleID, ExportedAt: exportedAt,
		ManifestSHA256:       built.Archive.ManifestSHA256,
		ReferencePackMode:    built.Archive.Manifest.ReferencePackMode,
		OptionalSections:     built.Archive.Manifest.OptionalSections,
		RequiredCapabilities: built.Archive.Manifest.RequiredCapabilities,
		BundleSHA256:         built.BundleSHA256, BundleByteSize: built.BundleByteSize,
		BundleStorageRef: storageReference, ManifestFiles: built.Archive.Manifest.Files,
	}
	var record descriptorRecord
	_, err = w.jobFinalizer.FinalizeIncidentBundleJobSuccess(ctx, JobSuccessFinalization{
		Execution:  execution,
		Completion: exportSuccessCompletion(bundleID),
		FinalCommitID: "incident_portability.export:" +
			bundleID.String() + ":" + built.Archive.ManifestSHA256,
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			found, lockErr := w.publicationLock.LockIncidentTx(ctx, tx, *payload.IncidentID)
			if lockErr != nil {
				return lockErr
			}
			if !found {
				return errNotFound
			}
			if guardErr := w.portability.ValidatePublication(ctx, tx, *payload.IncidentID); guardErr != nil {
				return guardErr
			}
			var mutationErr error
			record, mutationErr = w.store.completeExportDescriptorTx(ctx, tx, completeParams)
			return mutationErr
		},
	})
	if err != nil {
		w.handleExportFinalizationError(ctx, execution, storageReference, err)
		return
	}
	if record.BundleID != bundleID {
		return
	}
}

func (w *incidentBundleWorker) handleExportFinalizationError(ctx context.Context, execution jobs.Execution, storageReference BundleStorageRef, err error) {
	if errors.Is(err, ErrJobFinalizationIndeterminate) {
		return
	}
	_ = w.storage.RemovePublished(storageReference)
	if errors.Is(err, ErrPortabilityBlocked) {
		w.results.completeFailedFromError(ctx, execution, "incident_bundle_export_rejected", err)
		return
	}
	w.results.completeInternalFailure(ctx, execution)
}

func (w *incidentBundleWorker) executeImportJob(ctx context.Context, execution jobs.Execution, payload jobPayload) {
	if payload.BundleStagingRef == nil {
		w.results.completeInternalFailure(ctx, execution)
		return
	}
	defer func() {
		_ = w.storage.RemoveStaged(*payload.BundleStagingRef)
	}()
	if !w.prepareClaimedJob(ctx, execution, 1) {
		return
	}
	data, err := w.storage.ReadStaged(*payload.BundleStagingRef, incidentBundleStagingReadLimit(w.limits.IncidentBundles.MaxExtractedBytes))
	if err != nil {
		w.results.completeFailed(ctx, execution, failedCompletion("incident_bundle_import_rejected", map[string]any{"reason_code": "missing_required_file"}))
		return
	}
	verified, err := verifyBundle(verificationInput{Bundle: data, Limits: w.limits})
	if err != nil {
		if errors.Is(err, errExtensionCapabilityNotSupported) {
			w.results.completeFailedFromError(ctx, execution, "extension_capability_not_supported", err)
			return
		}
		w.results.completeFailedFromError(ctx, execution, "incident_bundle_import_rejected", err)
		return
	}
	requestID := importBundleRequestID(payload.JobID)
	importer := importer{
		pool:              w.pool,
		blobPort:          w.blobPort,
		finalizer:         w.importFinalizer,
		projectionRebuild: w.projectionRebuild,
		sourceCatalog:     w.sourceCatalog,
	}
	importParams := importParams{
		ActorUserID: payload.ActorUserID,
		PublishedAt: w.now().UTC(),
		RequestID:   &requestID,
		OperationID: payload.JobID.String(),
	}
	prepared, err := importer.prepareImport(ctx, verified, importParams)
	if err != nil {
		var verificationErr *verificationError
		if errors.As(err, &verificationErr) {
			w.results.completeFailedFromError(ctx, execution, "incident_bundle_import_rejected", err)
			return
		}
		w.results.completeInternalFailure(ctx, execution)
		return
	}
	committed := false
	finalityUnknown := false
	defer func() {
		if !committed && !finalityUnknown {
			prepared.cleanup(context.WithoutCancel(ctx))
		}
	}()
	portability, err := w.portability.PrepareImport(ctx, payload.JobID.String(), prepared.IncidentID, verified.Files)
	if err != nil {
		w.results.completeFailedFromError(ctx, execution, "incident_bundle_import_rejected", err)
		return
	}
	defer func() {
		if !committed && !finalityUnknown {
			_ = portability.Abandon(context.WithoutCancel(ctx))
		}
	}()
	coreParticipant, err := newImportTransactionParticipant(prepared, importParams, execution, verified.ManifestSHA256)
	if err != nil {
		w.results.completeInternalFailure(ctx, execution)
		return
	}
	participants := append([]crossownertransaction.Participant{coreParticipant}, portability.Participants...)
	result, err := w.transactions.Execute(ctx, crossownertransaction.Operation{
		OperationID: payload.JobID.String(), NormalizedRequestSHA256: verified.ManifestSHA256,
		Participants: participants,
		Finalizer: importJobTransactionFinalizer{
			finalizer: w.jobFinalizer, execution: execution,
			manifestSHA256: verified.ManifestSHA256,
		},
	})
	if err != nil {
		if crossownertransaction.IsFatalIntegrity(err) {
			finalityUnknown = true
			return
		}
		var verificationErr *verificationError
		if errors.As(err, &verificationErr) {
			w.results.completeFailedFromError(ctx, execution, "incident_bundle_import_rejected", verificationErr)
			return
		}
		w.results.completeInternalFailure(ctx, execution)
		return
	}
	value, ok := result.ParticipantValues[ImportTransactionParticipantID].(importTransactionResult)
	if !ok || value.IncidentID != prepared.IncidentID {
		w.results.completeInternalFailure(ctx, execution)
		return
	}
	committed = true
	prepared.stagedObjectKeys = nil
	portability.Committed()
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

func (w *incidentBundleWorker) prepareClaimedJob(ctx context.Context, execution jobs.Execution, total int) bool {
	if total <= 0 {
		total = 1
	}
	job, err := w.jobManager.ObserveExecution(ctx, execution)
	if err != nil {
		return false
	}
	switch job.Status {
	case jobs.StatusRunning:
		_, err := w.jobManager.UpdateProgress(ctx, execution, jobs.Progress{Completed: 0, Total: &total}, nil)
		return err == nil
	case jobs.StatusCancelRequested:
		_, _ = w.jobManager.CompleteCanceled(ctx, execution, jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 0, Total: &total}})
		return false
	default:
		return false
	}
}

type incidentBundleJobResultSink struct {
	manager   JobOperations
	store     *store
	finalizer JobSuccessFinalizer
	now       func() time.Time
}

func (s incidentBundleJobResultSink) completeFailed(ctx context.Context, execution jobs.Execution, completion jobs.FailureCompletion) {
	if resource, err := s.manager.ObserveExecution(ctx, execution); err == nil && resource.Status == jobs.StatusCancelRequested {
		_, _ = s.manager.CompleteCanceled(ctx, execution, jobs.CancellationCompletion{Progress: completion.Progress})
		return
	}
	_, _ = s.manager.CompleteFailed(ctx, execution, completion)
}

func (s incidentBundleJobResultSink) completeInternalFailure(ctx context.Context, execution jobs.Execution) {
	s.completeFailed(ctx, execution, internalFailureCompletion())
}

func (s incidentBundleJobResultSink) completeFailedFromError(ctx context.Context, execution jobs.Execution, code string, err error) {
	if code == "internal_error" {
		s.completeInternalFailure(ctx, execution)
		return
	}
	reason, details := incidentBundleFailureDetails(code, err)
	if resource, observeErr := s.manager.ObserveExecution(ctx, execution); observeErr == nil && resource.Status == jobs.StatusCancelRequested {
		_, _ = s.manager.CompleteCanceled(ctx, execution, jobs.CancellationCompletion{Progress: jobs.Progress{Completed: 1, Total: intPtr(1)}})
		return
	}
	if s.finalizer == nil {
		s.completeFailed(ctx, execution, failedCompletion(code, details))
		return
	}
	_, _ = s.finalizer.FinalizeIncidentBundleJobFailure(ctx, JobFailureFinalization{
		Execution:  execution,
		Completion: failedCompletion(code, details),
		Mutate: func(ctx context.Context, tx pgx.Tx) error {
			return s.store.markJobFailureTx(ctx, tx, execution.JobID(), reason, s.now())
		},
	})
}

func incidentBundleFailureDetails(code string, err error) (string, map[string]any) {
	if errors.Is(err, errExtensionCapabilityNotSupported) {
		return "extension_capability_not_supported", map[string]any{"profile_id": ProfileID}
	}
	var portabilityFailure *PortabilityFailure
	if code == "incident_bundle_export_rejected" && errors.As(err, &portabilityFailure) && errors.Is(err, ErrPortabilityBlocked) {
		return "extension_state_not_portable", map[string]any{
			"reason_code": "extension_state_not_portable",
			"profile_id":  portabilityFailure.ProfileID,
		}
	}
	reason := "missing_required_file"
	details := map[string]any{}
	var verificationErr *verificationError
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

func exportSuccessCompletion(bundleID uuid.UUID) jobs.SuccessCompletion {
	return jobs.SuccessCompletion{
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ResultSummary: jobs.ResultSummary{
			Code:    resultIncidentBundleExported,
			Message: "Incident bundle exported.",
			ResourceRefs: []jobs.ResourceRef{{
				Kind:  "incident_bundle",
				ID:    bundleID.String(),
				Route: "/api/v1/incident-bundles/" + bundleID.String(),
			}},
		},
	}
}

func importSuccessCompletion(incidentID uuid.UUID) jobs.SuccessCompletion {
	return jobs.SuccessCompletion{
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ResultSummary: jobs.ResultSummary{
			Code:    resultIncidentBundleImported,
			Message: "Incident bundle imported.",
			ResourceRefs: []jobs.ResourceRef{{
				Kind:  "incident",
				ID:    incidentID.String(),
				Route: "/api/v1/incidents/" + incidentID.String(),
			}},
		},
	}
}

func failedCompletion(code string, details map[string]any) jobs.FailureCompletion {
	if code == "internal_error" {
		details = map[string]any{}
	}
	return jobs.FailureCompletion{
		Progress: jobs.Progress{Completed: 1, Total: intPtr(1)},
		ErrorSummary: jobs.ErrorSummary{
			Code:      code,
			Message:   code,
			Retryable: false,
			Details:   details,
		},
	}
}

func internalFailureCompletion() jobs.FailureCompletion {
	return failedCompletion("internal_error", map[string]any{})
}
