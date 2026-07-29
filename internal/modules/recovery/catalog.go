package recovery

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/google/uuid"

	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

type BackupCatalog struct {
	store            backupRepository
	storage          BackupStorage
	extensionBackups *ExtensionBackupCatalog
	stateCatalog     *recoverystate.Catalog
}

type BackupCatalogSelection struct {
	BackupSet             BackupSet
	DurabilityDiagnostics []BackupDurabilityDiagnostic
}

type BackupDurabilityDiagnostic struct {
	BackupSetID        uuid.UUID
	ConsistencyPointAt time.Time
	Code               string
}

func NewBackupCatalog(
	store backupRepository,
	storage BackupStorage,
	extensionBackups *ExtensionBackupCatalog,
	stateCatalog ...*recoverystate.Catalog,
) *BackupCatalog {
	var state *recoverystate.Catalog
	if len(stateCatalog) == 1 {
		state = stateCatalog[0]
	}
	return &BackupCatalog{
		store: store, storage: storage, extensionBackups: extensionBackups, stateCatalog: state,
	}
}

func (catalog *BackupCatalog) LatestSuccessfulRetainedBackup(ctx context.Context, asOf time.Time) (BackupSet, error) {
	return catalog.RestoreCandidateBackup(ctx, asOf)
}

func (catalog *BackupCatalog) RestoreCandidateBackup(ctx context.Context, asOf time.Time) (BackupSet, error) {
	selection, err := catalog.RestoreCandidateBackupSelection(ctx, asOf)
	if err != nil {
		return BackupSet{}, err
	}
	return selection.BackupSet, nil
}

func (catalog *BackupCatalog) RestoreCandidateBackupSelection(ctx context.Context, asOf time.Time) (BackupCatalogSelection, error) {
	if catalog == nil || catalog.store == nil || catalog.storage == nil || catalog.extensionBackups == nil {
		return BackupCatalogSelection{}, fmt.Errorf("%w: backup catalog requires store and backup storage", ErrInvalidBackupMetadata)
	}
	asOf = normalizeAsOf(asOf)
	backupSets, err := catalog.store.ListRetainedBackupSetMetadata(ctx, asOf)
	if err != nil {
		return BackupCatalogSelection{}, err
	}
	sort.Slice(backupSets, func(i, j int) bool {
		if !backupSets[i].ConsistencyPointAt.Equal(backupSets[j].ConsistencyPointAt) {
			return backupSets[i].ConsistencyPointAt.After(backupSets[j].ConsistencyPointAt)
		}
		if !backupSets[i].CreatedAt.Equal(backupSets[j].CreatedAt) {
			return backupSets[i].CreatedAt.After(backupSets[j].CreatedAt)
		}
		return backupSets[i].BackupSetID.String() > backupSets[j].BackupSetID.String()
	})

	diagnostics := make([]BackupDurabilityDiagnostic, 0)
	for _, candidate := range backupSets {
		if err := catalog.VerifyBackupSetDurability(ctx, candidate); err != nil {
			diagnostics = append(diagnostics, BackupDurabilityDiagnostic{
				BackupSetID:        candidate.BackupSetID,
				ConsistencyPointAt: candidate.ConsistencyPointAt,
				Code:               backupDurabilityDiagnosticCode(err),
			})
			continue
		}
		if candidate.ConsistencyPointAt.Before(asOf.Add(-LatestSuccessfulBackupMaxAge)) {
			return BackupCatalogSelection{}, &LatestSuccessfulBackupStaleError{
				BackupSet: candidate,
				AsOf:      asOf,
				MaxAge:    LatestSuccessfulBackupMaxAge,
			}
		}
		return BackupCatalogSelection{
			BackupSet:             candidate,
			DurabilityDiagnostics: diagnostics,
		}, nil
	}
	return BackupCatalogSelection{}, ErrNoSuccessfulRetainedBackup
}

func (catalog *BackupCatalog) VerifyBackupSetDurability(ctx context.Context, backupSet BackupSet) error {
	if catalog == nil || catalog.storage == nil || catalog.extensionBackups == nil {
		return fmt.Errorf("%w: backup catalog requires backup storage", ErrInvalidBackupMetadata)
	}
	if _, vNext := VNextLogicalRefFromMetadataKey(backupSet.IntegrityManifestKey); vNext {
		return verifyVNextBackupSetDurability(
			ctx,
			catalog.storage,
			catalog.stateCatalog,
			backupSet,
		)
	}
	manifestProof := BackupArtifactProof{
		Key:       backupSet.IntegrityManifestKey,
		SHA256:    backupSet.IntegrityManifestSHA256,
		SizeBytes: backupSet.IntegrityManifestSizeBytes,
	}
	manifestBody, err := VerifyArtifactProof(ctx, catalog.storage, manifestProof)
	if err != nil {
		return fmt.Errorf("verify backup integrity manifest: %w", err)
	}
	manifest, err := DecodeIntegrityManifest(manifestBody)
	if err != nil {
		return fmt.Errorf("%w: decode backup integrity manifest: %v", ErrInvalidBackupArtifact, err)
	}
	if err := validateSelectedRestoreManifest(backupSet, manifest); err != nil {
		return err
	}
	postgresBody, err := VerifyArtifactProof(ctx, catalog.storage, manifest.PostgresArtifact)
	if err != nil {
		return fmt.Errorf("verify postgres backup artifact: %w", err)
	}
	postgresSnapshot, err := DecodePostgresSnapshotArtifact(postgresBody)
	if err != nil {
		return err
	}
	if err := validateExtensionBindingProofs(catalog.extensionBackups, manifest.ExtensionBindings, postgresSnapshot); err != nil {
		return err
	}
	objectBody, err := VerifyArtifactProof(ctx, catalog.storage, manifest.ObjectStoreArtifact)
	if err != nil {
		return fmt.Errorf("verify object-store backup artifact: %w", err)
	}
	if manifest.ObjectStoreBackupManifestArtifact == nil {
		return fmt.Errorf("%w: object-store backup manifest artifact is required", ErrInvalidBackupArtifact)
	}
	objectManifestBody, err := VerifyArtifactProof(ctx, catalog.storage, *manifest.ObjectStoreBackupManifestArtifact)
	if err != nil {
		return fmt.Errorf("verify object-store backup manifest artifact: %w", err)
	}
	objectManifest, err := DecodeObjectStoreBackupManifestArtifact(objectManifestBody)
	if err != nil {
		return err
	}
	if err := ValidateObjectStoreBackupManifestForBackup(backupSet, objectManifest); err != nil {
		return err
	}
	objectSnapshot, err := DecodeObjectStoreSnapshotArtifact(objectBody)
	if err != nil {
		return err
	}
	if err := ValidateObjectStoreManifestAgainstSnapshot(objectManifest, objectSnapshot); err != nil {
		return err
	}
	if manifest.ObjectStoreBackupSummaryArtifact != nil {
		summaryBody, err := VerifyArtifactProof(ctx, catalog.storage, *manifest.ObjectStoreBackupSummaryArtifact)
		if err != nil {
			return fmt.Errorf("verify object-store backup summary artifact: %w", err)
		}
		if _, err := DecodeObjectStoreBackupSummaryArtifact(summaryBody); err != nil {
			return err
		}
	}
	return nil
}

func (catalog *BackupCatalog) ListBackupsDueForRestoreVerification(ctx context.Context, asOf time.Time, verificationBasisSHA256 string) ([]BackupSet, error) {
	if catalog == nil || catalog.store == nil || catalog.storage == nil || catalog.extensionBackups == nil {
		return nil, fmt.Errorf("%w: backup catalog requires store and backup storage", ErrInvalidBackupMetadata)
	}
	backupSets, err := catalog.store.ListBackupsDueForRestoreVerification(ctx, asOf, verificationBasisSHA256)
	if err != nil {
		return nil, err
	}
	sort.Slice(backupSets, func(i, j int) bool {
		if !backupSets[i].ConsistencyPointAt.Equal(backupSets[j].ConsistencyPointAt) {
			return backupSets[i].ConsistencyPointAt.Before(backupSets[j].ConsistencyPointAt)
		}
		return backupSets[i].BackupSetID.String() < backupSets[j].BackupSetID.String()
	})
	return backupSets, nil
}

func backupDurabilityDiagnosticCode(err error) string {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "artifact_missing"
	case errors.Is(err, ErrInvalidBackupArtifact):
		return "invalid_backup_artifact"
	case errors.Is(err, ErrEncryptedBackupStorage):
		return "encrypted_backup_storage_required"
	case errors.Is(err, ErrRecoveryMasterKeyRequired):
		return "recovery_master_key_required"
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return "interrupted"
	default:
		return "durability_check_failed"
	}
}
