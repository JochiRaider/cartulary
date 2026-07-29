package recovery

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// backupRepository is the private persistence boundary for backup metadata.
// SQLC and concrete database types remain confined to Store.
type backupRepository interface {
	createCapturedBackupSet(context.Context, createBackupSetParams) (BackupSet, error)
	GetBackupSet(context.Context, uuid.UUID) (BackupSet, error)
	ListRetainedBackupSetMetadata(context.Context, time.Time) ([]BackupSet, error)
	ListBackupsDueForRestoreVerification(context.Context, time.Time, string) ([]BackupSet, error)
}

// verificationRepository is the private persistence boundary for verification
// attempts and backup attestation state.
type verificationRepository interface {
	RecordRestoreVerificationCompletion(context.Context, CreateRestoreVerificationRunParams) (BackupSet, RestoreVerificationRun, error)
}

type recoveryRepository interface {
	backupRepository
	verificationRepository
}
