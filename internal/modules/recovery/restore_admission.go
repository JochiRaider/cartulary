package recovery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/maintenancequery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

// FencedRestore owns one pending payload transaction and its independently
// committed admission record. Only the admitted runner can consume it.
type FencedRestore struct {
	operation                          sync.Mutex
	runner                             *RestoreRunner
	capability                         TargetControlCapability
	ctx                                context.Context
	cancel                             context.CancelFunc
	stop                               func() bool
	target                             RestoreTarget
	backupSet                          BackupSet
	tx                                 pgx.Tx
	admitted                           TargetControl
	service                            *VNextRestoreService
	proof                              BackupArtifactStreamProof
	evidence                           VNextRestoreVerificationEvidence
	used, committed, validated, closed bool
}

func (runner *RestoreRunner) prepareRestore(ctx context.Context, backupSet BackupSet) (*VNextRestoreService, BackupArtifactStreamProof, VNextRestoreVerificationEvidence, error) {
	if runner == nil || runner.storage == nil || runner.extensionBackups == nil || runner.stateCatalog == nil {
		return nil, BackupArtifactStreamProof{}, VNextRestoreVerificationEvidence{}, ErrInvalidBackupMetadata
	}
	if _, current := VNextLogicalRefFromMetadataKey(backupSet.IntegrityManifestKey); !current {
		return nil, BackupArtifactStreamProof{}, VNextRestoreVerificationEvidence{}, fmt.Errorf("%w: retired backup representation", ErrInvalidBackupArtifact)
	}
	streaming, err := RequireStreamingBackupStorage(runner.storage)
	if err != nil {
		return nil, BackupArtifactStreamProof{}, VNextRestoreVerificationEvidence{}, err
	}
	algorithms, err := NewVNextRestoreAlgorithmCatalog(runner.stateCatalog, RequiredVNextRestoreAlgorithmIDs(runner.stateCatalog)...)
	if err != nil {
		return nil, BackupArtifactStreamProof{}, VNextRestoreVerificationEvidence{}, err
	}
	service, err := NewVNextRestoreService(streaming, runner.stateCatalog, algorithms, runner.extensionBackups)
	if err != nil {
		return nil, BackupArtifactStreamProof{}, VNextRestoreVerificationEvidence{}, err
	}
	proof, err := VNextProofFromMetadata(ctx, streaming, backupSet.IntegrityManifestKey, vNextJSONContentType, backupSet.IntegrityManifestSizeBytes, backupSet.IntegrityManifestSHA256)
	if err != nil {
		return nil, BackupArtifactStreamProof{}, VNextRestoreVerificationEvidence{}, err
	}
	evidence, err := service.ReadVerificationEvidence(ctx, proof)
	return service, proof, evidence, err
}

// AdmitRestore authenticates the selected current artifacts, fences the actual
// target tables, and rechecks emptiness and mappings before publishing restoring.
// The mandatory journal callback runs after that durable publication and before
// this method permits any payload mutation. Its failure leaves quarantine intact.
func (runner *RestoreRunner) AdmitRestore(ctx context.Context, capability TargetControlCapability, target RestoreTarget, backupSet BackupSet, identity RestoreTargetIdentity, writeAdmission func(context.Context) error) (_ *FencedRestore, resultErr error) {
	if capability == nil || capability.TargetDatabase() == nil || writeAdmission == nil || target.fence != nil {
		return nil, ErrTargetIncomplete
	}
	if err := capability.AssertHeld(); err != nil {
		return nil, err
	}
	if err := identity.validate(); err != nil {
		return nil, err
	}
	if identity.BackupSetID != backupSet.BackupSetID || target.ObjectStore == nil || target.Projections == nil || target.EvidenceObjects == nil {
		return nil, ErrInvalidBackupArtifact
	}
	work, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(capability.Context(), cancel)
	fence := &FencedRestore{runner: runner, capability: capability, ctx: work, cancel: cancel, stop: stop}
	returned := false
	defer func() {
		if !returned {
			_ = fence.Close()
		}
	}()
	target.Postgres = capability.TargetDatabase()
	target.RestoreOperationID = identity.OperationID
	target.TargetGenerationID = identity.TargetGenerationID
	target.fence = fence
	fence.target = target
	fence.backupSet = backupSet
	if backupSet.LastVerifiedRestoreAt != nil {
		copied := *backupSet.LastVerifiedRestoreAt
		fence.backupSet.LastVerifiedRestoreAt = &copied
	}
	var err error
	fence.service, fence.proof, fence.evidence, err = runner.prepareRestore(work, backupSet)
	if err != nil {
		return nil, err
	}
	if fence.evidence.GraphRestoreArtifacts.RecoveryStateCatalogSHA256 != identity.RecoveryStateCatalogSHA256 || fence.evidence.codecRegistrySHA256 != identity.CodecRegistrySHA256 {
		return nil, ErrInvalidBackupArtifact
	}
	// This first inspection is read only. The same checks run again under the
	// pending transaction's locks; neither source history nor a marker replaces it.
	if err := requireEmptyRestoreTarget(work, target, runner.extensionBackups, runner.stateCatalog); err != nil {
		return nil, err
	}
	fence.tx, err = beginRestoreTargetTransaction(work, target.Postgres, runner.stateCatalog)
	if err != nil {
		return nil, err
	}
	if err := requireEmptyRestoreTargetWithReader(work, target, fence.tx, runner.extensionBackups, runner.stateCatalog); err != nil {
		return nil, err
	}
	current, err := ReadTargetControl(work, fence.tx)
	if err != nil {
		return nil, err
	}
	next := identity.control()
	if current.State != TargetInitialized && (current.State != TargetVerificationReady || !sameTargetBinding(current, next)) {
		return nil, ErrTargetIncomplete
	}
	fence.admitted, err = advanceTargetControl(work, capability, current, next)
	if err != nil {
		return nil, err
	}
	if err := writeAdmission(work); err != nil {
		return nil, err
	}
	if err := capability.AssertHeld(); err != nil {
		return nil, err
	}
	returned = true
	return fence, nil
}

// Target returns the immutable resource selection retained by admission. The
// runner uses its sealed copy, so replacing fields on this copy grants no access.
func (fence *FencedRestore) Target() RestoreTarget { return fence.target }

func (fence *FencedRestore) Close() error {
	if fence == nil {
		return nil
	}
	fence.cancel()
	fence.stop()
	fence.operation.Lock()
	defer fence.operation.Unlock()
	if fence.closed {
		return nil
	}
	fence.closed = true
	if fence.tx == nil || fence.committed {
		return nil
	}
	return rollbackRestoreTransaction(fence.tx)
}

// Complete requires both committed payload and all independent Core checks.
// The caller supplies the digest of its durably successful terminal record only
// after any verification attestation has also completed.
func (fence *FencedRestore) Complete(ctx context.Context, proof TargetCompletionProof) error {
	if fence == nil {
		return ErrTargetIncomplete
	}
	fence.operation.Lock()
	defer fence.operation.Unlock()
	if fence.closed || !fence.committed || !fence.validated || !validSHA256Hex(proof.EvidenceSHA256) || !sameTargetOperation(fence.admitted, proof.Identity.control()) {
		return ErrTargetIncomplete
	}
	if err := fence.ctx.Err(); err != nil {
		return err
	}
	next := fence.admitted
	next.State = TargetComplete
	next.CompletionEvidenceSHA256 = targetControlText(proof.EvidenceSHA256)
	_, err := advanceTargetControl(ctx, fence.capability, fence.admitted, next)
	return err
}

// The protected control singleton remains outside this transaction so its
// quarantine commit survives rollback of all payload work.
func beginRestoreTargetTransaction(ctx context.Context, db dbport.DB, catalog *recoverystate.Catalog) (_ pgx.Tx, resultErr error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		if resultErr != nil {
			_ = rollbackRestoreTransaction(tx)
		}
	}()
	tables, err := mutableRestoreTargetTables(ctx, tx, catalog)
	if err != nil {
		return nil, err
	}
	command, err := maintenancequery.LockMutable(tables)
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, sqlshape.CommandText(command)); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `SET LOCAL row_security = off`); err != nil {
		return nil, err
	}
	return tx, nil
}

func rollbackRestoreTransaction(tx pgx.Tx) error {
	cleanup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := tx.Rollback(cleanup)
	if errors.Is(err, pgx.ErrTxClosed) {
		return nil
	}
	return err
}
