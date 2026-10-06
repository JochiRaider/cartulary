package application

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func (service Service) BackupExportLatest(ctx context.Context, request BackupExportLatestRequest, progress ProgressSink) (outcome Result, operationErr error) {
	ReportProgress(progress, "preflight", 0, nil)
	_, pool, storage, closeSource, err := service.openSourceRuntime(ctx, request.SourceConfigPath)
	if err != nil {
		return Result{}, err
	}
	defer closeSource()
	unlock, err := service.acquireOperationLock(ctx, pool)
	if err != nil {
		return Result{}, err
	}
	defer unlock()
	parsed := operationRequest{OperationID: request.OperationID, Operation: OperationBackupExportLatest, StartedAt: service.now(), SourceConfigPath: request.SourceConfigPath, ArtifactKinds: []string{"recovery_transfer"}}
	if err := service.recordRecoveryStart(ctx, pool, parsed); err != nil {
		return Result{}, err
	}
	defer func() {
		ReportProgress(progress, "journal_write", 0, nil)
		service.finishJournalAndAudit(ctx, pool, parsed, outcome, &operationErr)
		if operationErr == nil {
			ReportProgress(progress, "finalize", 1, IntPtr(1))
		}
	}()
	if service.NewTransferDirectory == nil || service.ReleaseIdentity == nil {
		return Result{}, NewFailure(FailureTransferPublication, errors.New("transfer composition unavailable"))
	}
	release, err := service.ReleaseIdentity()
	if err != nil {
		return Result{}, NewFailure(FailureTransferReleaseMismatch, err)
	}
	ReportProgress(progress, "catalog_select", 0, nil)
	catalog := recovery.NewBackupCatalog(recovery.NewStore(pool), storage, service.ExtensionBackups, service.RecoveryStateCatalog)
	selection, err := catalog.IntactRetainedBackupSelection(ctx, service.now())
	if err != nil {
		return Result{}, classifyAdmissionFailure(err)
	}
	outcome = ResultForStoredBackupSet(selection.BackupSet)
	destination, err := service.NewTransferDirectory(request.OutputDirectory)
	if err != nil {
		return outcome, NewFailure(FailureTransferPublication, err)
	}
	defer func() {
		if err := destination.Close(); err != nil && operationErr == nil {
			operationErr = NewFailure(FailureTransferPublication, err)
		}
	}()
	ReportProgress(progress, "artifact_validate", 0, nil)
	ReportProgress(progress, "transfer_copy", 0, nil)
	if _, err := catalog.ExportTransfer(ctx, selection.BackupSet, release, service.now(), destination); err != nil {
		return outcome, NewFailure(FailureTransferCopy, err)
	}
	if err := destination.Publish(ctx); err != nil {
		return outcome, NewFailure(FailureTransferPublication, err)
	}
	outcome = ResultForBackupSet(selection.BackupSet, "recovery_transfer", recovery.TransferManifestSchemaID)
	return outcome, nil
}
