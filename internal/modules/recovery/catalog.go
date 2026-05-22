package recovery

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/google/uuid"
)

type BackupCatalog struct {
	store   *Store
	storage BackupStorage
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

func NewBackupCatalog(store *Store, storage BackupStorage) *BackupCatalog {
	return &BackupCatalog{store: store, storage: storage}
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
	if catalog == nil || catalog.store == nil || catalog.storage == nil {
		return BackupCatalogSelection{}, fmt.Errorf("%w: backup catalog requires store and backup storage", ErrInvalidBackupMetadata)
	}
	asOf = normalizeAsOf(asOf)
	backupSets, err := catalog.store.ListSuccessfulRetainedBackups(ctx, asOf)
	if err != nil {
		return BackupCatalogSelection{}, err
	}
	sort.Slice(backupSets, func(i, j int) bool {
		if backupSets[i].ConsistencyPointAt.Equal(backupSets[j].ConsistencyPointAt) {
			return backupSets[i].BackupSetID.String() < backupSets[j].BackupSetID.String()
		}
		return backupSets[i].ConsistencyPointAt.After(backupSets[j].ConsistencyPointAt)
	})

	diagnostics := make([]BackupDurabilityDiagnostic, 0)
	for index := 0; index < len(backupSets); {
		point := backupSets[index].ConsistencyPointAt
		group := make([]BackupSet, 0, 1)
		for index < len(backupSets) && backupSets[index].ConsistencyPointAt.Equal(point) {
			candidate := backupSets[index]
			if err := catalog.VerifyBackupSetDurability(ctx, candidate); err == nil {
				group = append(group, candidate)
			} else {
				diagnostics = append(diagnostics, BackupDurabilityDiagnostic{
					BackupSetID:        candidate.BackupSetID,
					ConsistencyPointAt: candidate.ConsistencyPointAt,
					Code:               backupDurabilityDiagnosticCode(err),
				})
			}
			index++
		}
		if len(group) == 0 {
			continue
		}
		if len(group) != 1 {
			ids := make([]uuid.UUID, 0, len(group))
			for _, candidate := range group {
				ids = append(ids, candidate.BackupSetID)
			}
			return BackupCatalogSelection{}, &AmbiguousBackupSelectionError{
				ConsistencyPointAt: point,
				BackupSetIDs:       ids,
			}
		}
		backupSet := group[0]
		if backupSet.ConsistencyPointAt.Before(asOf.Add(-LatestSuccessfulBackupMaxAge)) {
			return BackupCatalogSelection{}, &LatestSuccessfulBackupStaleError{
				BackupSet: backupSet,
				AsOf:      asOf,
				MaxAge:    LatestSuccessfulBackupMaxAge,
			}
		}
		return BackupCatalogSelection{
			BackupSet:             backupSet,
			DurabilityDiagnostics: diagnostics,
		}, nil
	}
	return BackupCatalogSelection{}, ErrNoSuccessfulRetainedBackup
}

func (catalog *BackupCatalog) VerifyBackupSetDurability(ctx context.Context, backupSet BackupSet) error {
	if catalog == nil || catalog.storage == nil {
		return fmt.Errorf("%w: backup catalog requires backup storage", ErrInvalidBackupMetadata)
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
	if _, err := VerifyArtifactProof(ctx, catalog.storage, manifest.PostgresArtifact); err != nil {
		return fmt.Errorf("verify postgres backup artifact: %w", err)
	}
	if _, err := VerifyArtifactProof(ctx, catalog.storage, manifest.ObjectStoreArtifact); err != nil {
		return fmt.Errorf("verify object-store backup artifact: %w", err)
	}
	return nil
}

func (catalog *BackupCatalog) ListBackupsDueForRestoreVerification(ctx context.Context, asOf time.Time, verificationBasisSHA256 string) ([]BackupSet, error) {
	if catalog == nil || catalog.store == nil || catalog.storage == nil {
		return nil, fmt.Errorf("%w: backup catalog requires store and backup storage", ErrInvalidBackupMetadata)
	}
	backupSets, err := catalog.store.ListBackupsDueForRestoreVerification(ctx, asOf, verificationBasisSHA256)
	if err != nil {
		return nil, err
	}
	sort.Slice(backupSets, func(i, j int) bool {
		if backupSets[i].ConsistencyPointAt.Equal(backupSets[j].ConsistencyPointAt) {
			return backupSets[i].BackupSetID.String() < backupSets[j].BackupSetID.String()
		}
		return backupSets[i].ConsistencyPointAt.After(backupSets[j].ConsistencyPointAt)
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
