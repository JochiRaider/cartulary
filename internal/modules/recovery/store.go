package recovery

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	VerificationUnverified VerificationState = "unverified"
	VerificationVerified   VerificationState = "verified"
	VerificationFailed     VerificationState = "failed"

	MinimumRetentionDuration     = 30 * 24 * time.Hour
	LatestSuccessfulBackupMaxAge = 24 * time.Hour
)

var (
	ErrBackupSetNotFound              = errors.New("recovery: backup set not found")
	ErrNoSuccessfulRetainedBackup     = errors.New("recovery: no successful retained backup")
	ErrAmbiguousBackupSelection       = errors.New("recovery: ambiguous latest successful retained backup selection")
	ErrLatestSuccessfulBackupStale    = errors.New("recovery: latest successful retained backup is older than 24 hours")
	ErrInvalidBackupMetadata          = errors.New("recovery: invalid backup metadata")
	ErrInvalidVerificationState       = errors.New("recovery: invalid verification state")
	ErrInvalidVerificationBasis       = errors.New("recovery: invalid verification basis")
	ErrVerificationTimestampRequired  = errors.New("recovery: verification timestamp required")
	ErrVerificationTimestampForbidden = errors.New("recovery: verification timestamp forbidden")
	ErrRetentionFloor                 = errors.New("recovery: retention floor not satisfied")
)

type VerificationState string

type Store struct {
	pool postgres.DB
}

type BackupSet struct {
	BackupSetID                           uuid.UUID
	ConsistencyPointAt                    time.Time
	PostgresRestoreAnchor                 string
	ObjectStoreRestoreAnchor              string
	PostgresArtifactKey                   string
	PostgresArtifactSHA256                string
	PostgresArtifactSizeBytes             int64
	ObjectStoreArtifactKey                string
	ObjectStoreArtifactSHA256             string
	ObjectStoreArtifactSizeBytes          int64
	IntegrityManifestKey                  string
	IntegrityManifestSHA256               string
	IntegrityManifestSizeBytes            int64
	CreatedAt                             time.Time
	RetainedUntil                         time.Time
	PostgresRestoreAnchorRetainedUntil    time.Time
	ObjectStoreRestoreAnchorRetainedUntil time.Time
	VerificationState                     VerificationState
	LastVerifiedRestoreAt                 *time.Time
	LastVerificationBasisSHA256           string
}

type RestoreVerificationRun struct {
	RestoreVerificationRunID uuid.UUID
	BackupSetID              uuid.UUID
	StartedAt                time.Time
	CompletedAt              time.Time
	VerificationState        VerificationState
	VerificationBasisSHA256  string
	FailureReason            string
	FailureMessage           string
	ConsistencyReport        RestoreConsistencyReport
}

type CreateRestoreVerificationRunParams struct {
	RestoreVerificationRunID uuid.UUID
	BackupSetID              uuid.UUID
	StartedAt                time.Time
	CompletedAt              time.Time
	VerificationState        VerificationState
	VerificationBasisSHA256  string
	FailureReason            string
	FailureMessage           string
	ConsistencyReport        RestoreConsistencyReport
}

type createBackupSetParams struct {
	BackupSetID                           uuid.UUID
	ConsistencyPointAt                    time.Time
	PostgresRestoreAnchor                 string
	ObjectStoreRestoreAnchor              string
	PostgresArtifactKey                   string
	PostgresArtifactSHA256                string
	PostgresArtifactSizeBytes             int64
	ObjectStoreArtifactKey                string
	ObjectStoreArtifactSHA256             string
	ObjectStoreArtifactSizeBytes          int64
	IntegrityManifestKey                  string
	IntegrityManifestSHA256               string
	IntegrityManifestSizeBytes            int64
	CreatedAt                             time.Time
	RetainedUntil                         time.Time
	PostgresRestoreAnchorRetainedUntil    time.Time
	ObjectStoreRestoreAnchorRetainedUntil time.Time
}

type LatestSuccessfulBackupStaleError struct {
	BackupSet BackupSet
	AsOf      time.Time
	MaxAge    time.Duration
}

type AmbiguousBackupSelectionError struct {
	ConsistencyPointAt time.Time
	BackupSetIDs       []uuid.UUID
}

func (err *AmbiguousBackupSelectionError) Error() string {
	return ErrAmbiguousBackupSelection.Error()
}

func (err *AmbiguousBackupSelectionError) Unwrap() error {
	return ErrAmbiguousBackupSelection
}

func (err *LatestSuccessfulBackupStaleError) Error() string {
	return ErrLatestSuccessfulBackupStale.Error()
}

func (err *LatestSuccessfulBackupStaleError) Unwrap() error {
	return ErrLatestSuccessfulBackupStale
}

func NewStore(pool postgres.DB) *Store {
	return &Store{pool: pool}
}

func (s *Store) createCapturedBackupSet(ctx context.Context, params createBackupSetParams) (BackupSet, error) {
	normalized, err := normalizeCreateBackupSetParams(params)
	if err != nil {
		return BackupSet{}, err
	}

	row, err := sqlc.New(s.pool).CreateBackupSet(ctx, sqlc.CreateBackupSetParams{
		BackupSetID:                           pgUUID(normalized.BackupSetID),
		ConsistencyPointAt:                    pgTimestamptz(normalized.ConsistencyPointAt),
		PostgresRestoreAnchor:                 normalized.PostgresRestoreAnchor,
		ObjectStoreRestoreAnchor:              normalized.ObjectStoreRestoreAnchor,
		PostgresArtifactKey:                   normalized.PostgresArtifactKey,
		PostgresArtifactSha256:                normalized.PostgresArtifactSHA256,
		PostgresArtifactSizeBytes:             normalized.PostgresArtifactSizeBytes,
		ObjectStoreArtifactKey:                normalized.ObjectStoreArtifactKey,
		ObjectStoreArtifactSha256:             normalized.ObjectStoreArtifactSHA256,
		ObjectStoreArtifactSizeBytes:          normalized.ObjectStoreArtifactSizeBytes,
		IntegrityManifestKey:                  normalized.IntegrityManifestKey,
		IntegrityManifestSha256:               normalized.IntegrityManifestSHA256,
		IntegrityManifestSizeBytes:            normalized.IntegrityManifestSizeBytes,
		CreatedAt:                             pgTimestamptz(normalized.CreatedAt),
		RetainedUntil:                         pgTimestamptz(normalized.RetainedUntil),
		PostgresRestoreAnchorRetainedUntil:    pgTimestamptz(normalized.PostgresRestoreAnchorRetainedUntil),
		ObjectStoreRestoreAnchorRetainedUntil: pgTimestamptz(normalized.ObjectStoreRestoreAnchorRetainedUntil),
	})
	if err != nil {
		return BackupSet{}, fmt.Errorf("create backup set: %w", err)
	}
	return backupSetFromSQL(row)
}

func (s *Store) GetBackupSet(ctx context.Context, backupSetID uuid.UUID) (BackupSet, error) {
	row, err := sqlc.New(s.pool).GetBackupSetByID(ctx, pgUUID(backupSetID))
	if errors.Is(err, pgx.ErrNoRows) {
		return BackupSet{}, ErrBackupSetNotFound
	}
	if err != nil {
		return BackupSet{}, fmt.Errorf("get backup set: %w", err)
	}
	return backupSetFromSQL(row)
}

func (s *Store) LatestSuccessfulRetainedBackup(ctx context.Context, asOf time.Time) (BackupSet, error) {
	asOf = normalizeAsOf(asOf)
	row, err := sqlc.New(s.pool).GetLatestSuccessfulRetainedBackupSet(ctx, pgTimestamptz(asOf))
	if errors.Is(err, pgx.ErrNoRows) {
		return BackupSet{}, ErrNoSuccessfulRetainedBackup
	}
	if err != nil {
		return BackupSet{}, fmt.Errorf("get latest successful retained backup: %w", err)
	}
	backupSet, err := backupSetFromSQL(row)
	if err != nil {
		return BackupSet{}, err
	}
	if backupSet.ConsistencyPointAt.Before(asOf.Add(-LatestSuccessfulBackupMaxAge)) {
		return BackupSet{}, &LatestSuccessfulBackupStaleError{
			BackupSet: backupSet,
			AsOf:      asOf,
			MaxAge:    LatestSuccessfulBackupMaxAge,
		}
	}
	return backupSet, nil
}

func (s *Store) RestoreCandidateBackup(ctx context.Context, asOf time.Time) (BackupSet, error) {
	asOf = normalizeAsOf(asOf)
	backupSets, err := s.ListSuccessfulRetainedBackups(ctx, asOf)
	if err != nil {
		return BackupSet{}, err
	}
	if len(backupSets) == 0 {
		return BackupSet{}, ErrNoSuccessfulRetainedBackup
	}
	latestPoint := backupSets[0].ConsistencyPointAt
	for _, backupSet := range backupSets[1:] {
		if backupSet.ConsistencyPointAt.After(latestPoint) {
			latestPoint = backupSet.ConsistencyPointAt
		}
	}
	candidates := make([]BackupSet, 0, 1)
	for _, backupSet := range backupSets {
		if backupSet.ConsistencyPointAt.Equal(latestPoint) {
			candidates = append(candidates, backupSet)
		}
	}
	if len(candidates) != 1 {
		ids := make([]uuid.UUID, 0, len(candidates))
		for _, candidate := range candidates {
			ids = append(ids, candidate.BackupSetID)
		}
		return BackupSet{}, &AmbiguousBackupSelectionError{
			ConsistencyPointAt: latestPoint,
			BackupSetIDs:       ids,
		}
	}
	backupSet := candidates[0]
	if backupSet.ConsistencyPointAt.Before(asOf.Add(-LatestSuccessfulBackupMaxAge)) {
		return BackupSet{}, &LatestSuccessfulBackupStaleError{
			BackupSet: backupSet,
			AsOf:      asOf,
			MaxAge:    LatestSuccessfulBackupMaxAge,
		}
	}
	return backupSet, nil
}

func (s *Store) ListSuccessfulRetainedBackups(ctx context.Context, asOf time.Time) ([]BackupSet, error) {
	rows, err := sqlc.New(s.pool).ListSuccessfulRetainedBackupSets(ctx, pgTimestamptz(normalizeAsOf(asOf)))
	if err != nil {
		return nil, fmt.Errorf("list successful retained backups: %w", err)
	}
	backupSets := make([]BackupSet, 0, len(rows))
	for _, row := range rows {
		backupSet, err := backupSetFromSQL(row)
		if err != nil {
			return nil, err
		}
		backupSets = append(backupSets, backupSet)
	}
	return backupSets, nil
}

func (s *Store) ListBackupsDueForRestoreVerification(ctx context.Context, asOf time.Time, verificationBasisSHA256 string) ([]BackupSet, error) {
	if !validOptionalSHA256Hex(verificationBasisSHA256) {
		return nil, ErrInvalidVerificationBasis
	}
	rows, err := sqlc.New(s.pool).ListBackupSetsDueForRestoreVerification(ctx, sqlc.ListBackupSetsDueForRestoreVerificationParams{
		RetainedUntil:               pgTimestamptz(normalizeAsOf(asOf)),
		LastVerificationBasisSha256: pgOptionalText(verificationBasisSHA256),
	})
	if err != nil {
		return nil, fmt.Errorf("list backup sets due for restore verification: %w", err)
	}
	backupSets := make([]BackupSet, 0, len(rows))
	for _, row := range rows {
		backupSet, err := backupSetFromSQL(row)
		if err != nil {
			return nil, err
		}
		backupSets = append(backupSets, backupSet)
	}
	return backupSets, nil
}

func (s *Store) UpdateVerificationState(ctx context.Context, backupSetID uuid.UUID, state VerificationState, lastVerifiedRestoreAt *time.Time, verificationBasisSHA256 ...string) (BackupSet, error) {
	basis := ""
	if len(verificationBasisSHA256) > 0 {
		basis = verificationBasisSHA256[0]
	}
	if err := validateVerificationTransition(state, lastVerifiedRestoreAt, basis); err != nil {
		return BackupSet{}, err
	}
	row, err := sqlc.New(s.pool).UpdateBackupSetVerificationState(ctx, sqlc.UpdateBackupSetVerificationStateParams{
		BackupSetID:                 pgUUID(backupSetID),
		VerificationState:           string(state),
		LastVerifiedRestoreAt:       pgOptionalTimestamptz(lastVerifiedRestoreAt),
		LastVerificationBasisSha256: pgOptionalText(basis),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return BackupSet{}, ErrBackupSetNotFound
	}
	if err != nil {
		return BackupSet{}, fmt.Errorf("update backup verification state: %w", err)
	}
	return backupSetFromSQL(row)
}

func (s *Store) CreateRestoreVerificationRun(ctx context.Context, params CreateRestoreVerificationRunParams) (RestoreVerificationRun, error) {
	normalized, err := normalizeRestoreVerificationRunParams(params)
	if err != nil {
		return RestoreVerificationRun{}, err
	}
	row, err := sqlc.New(s.pool).CreateRestoreVerificationRun(ctx, sqlc.CreateRestoreVerificationRunParams{
		RestoreVerificationRunID: pgUUID(normalized.RestoreVerificationRunID),
		BackupSetID:              pgUUID(normalized.BackupSetID),
		StartedAt:                pgTimestamptz(normalized.StartedAt),
		CompletedAt:              pgTimestamptz(normalized.CompletedAt),
		VerificationState:        string(normalized.VerificationState),
		VerificationBasisSha256:  normalized.VerificationBasisSHA256,
		FailureReason:            pgOptionalText(normalized.FailureReason),
		FailureMessage:           pgOptionalText(normalized.FailureMessage),
		AuthoritativeRowsSha256:  pgOptionalText(normalized.ConsistencyReport.AuthoritativeRowsSHA256),
		AuthoritativeRowCount:    pgOptionalInt4(normalized.ConsistencyReport.AuthoritativeRowCount),
		ChangeSetsSha256:         pgOptionalText(normalized.ConsistencyReport.ChangeSetsSHA256),
		ChangeSetRowCount:        pgOptionalInt4(normalized.ConsistencyReport.ChangeSetRowCount),
		BlobHashesSha256:         pgOptionalText(normalized.ConsistencyReport.BlobHashesSHA256),
		BlobCount:                pgOptionalInt4(normalized.ConsistencyReport.BlobCount),
	})
	if err != nil {
		return RestoreVerificationRun{}, fmt.Errorf("create restore verification run: %w", err)
	}
	return restoreVerificationRunFromSQL(row)
}

func (s *Store) RecordRestoreVerificationCompletion(ctx context.Context, params CreateRestoreVerificationRunParams) (BackupSet, RestoreVerificationRun, error) {
	normalized, err := normalizeRestoreVerificationRunParams(params)
	if err != nil {
		return BackupSet{}, RestoreVerificationRun{}, err
	}
	lastVerifiedAt := normalized.CompletedAt
	if err := validateVerificationTransition(normalized.VerificationState, &lastVerifiedAt, normalized.VerificationBasisSHA256); err != nil {
		return BackupSet{}, RestoreVerificationRun{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return BackupSet{}, RestoreVerificationRun{}, fmt.Errorf("begin restore verification completion: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queries := sqlc.New(tx)
	backupRow, err := queries.UpdateBackupSetVerificationState(ctx, sqlc.UpdateBackupSetVerificationStateParams{
		BackupSetID:                 pgUUID(normalized.BackupSetID),
		VerificationState:           string(normalized.VerificationState),
		LastVerifiedRestoreAt:       pgTimestamptz(lastVerifiedAt),
		LastVerificationBasisSha256: pgOptionalText(normalized.VerificationBasisSHA256),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return BackupSet{}, RestoreVerificationRun{}, ErrBackupSetNotFound
	}
	if err != nil {
		return BackupSet{}, RestoreVerificationRun{}, fmt.Errorf("update backup verification state: %w", err)
	}
	runRow, err := queries.CreateRestoreVerificationRun(ctx, sqlc.CreateRestoreVerificationRunParams{
		RestoreVerificationRunID: pgUUID(normalized.RestoreVerificationRunID),
		BackupSetID:              pgUUID(normalized.BackupSetID),
		StartedAt:                pgTimestamptz(normalized.StartedAt),
		CompletedAt:              pgTimestamptz(normalized.CompletedAt),
		VerificationState:        string(normalized.VerificationState),
		VerificationBasisSha256:  normalized.VerificationBasisSHA256,
		FailureReason:            pgOptionalText(normalized.FailureReason),
		FailureMessage:           pgOptionalText(normalized.FailureMessage),
		AuthoritativeRowsSha256:  pgOptionalText(normalized.ConsistencyReport.AuthoritativeRowsSHA256),
		AuthoritativeRowCount:    pgOptionalInt4(normalized.ConsistencyReport.AuthoritativeRowCount),
		ChangeSetsSha256:         pgOptionalText(normalized.ConsistencyReport.ChangeSetsSHA256),
		ChangeSetRowCount:        pgOptionalInt4(normalized.ConsistencyReport.ChangeSetRowCount),
		BlobHashesSha256:         pgOptionalText(normalized.ConsistencyReport.BlobHashesSHA256),
		BlobCount:                pgOptionalInt4(normalized.ConsistencyReport.BlobCount),
	})
	if err != nil {
		return BackupSet{}, RestoreVerificationRun{}, fmt.Errorf("create restore verification run: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return BackupSet{}, RestoreVerificationRun{}, fmt.Errorf("commit restore verification completion: %w", err)
	}
	backupSet, err := backupSetFromSQL(backupRow)
	if err != nil {
		return BackupSet{}, RestoreVerificationRun{}, err
	}
	run, err := restoreVerificationRunFromSQL(runRow)
	if err != nil {
		return BackupSet{}, RestoreVerificationRun{}, err
	}
	return backupSet, run, nil
}

func normalizeCreateBackupSetParams(params createBackupSetParams) (createBackupSetParams, error) {
	if params.BackupSetID == uuid.Nil {
		params.BackupSetID = uuid.New()
	}
	if params.ConsistencyPointAt.IsZero() {
		return createBackupSetParams{}, fmt.Errorf("%w: consistency_point_at is required", ErrInvalidBackupMetadata)
	}
	if strings.TrimSpace(params.PostgresRestoreAnchor) == "" {
		return createBackupSetParams{}, fmt.Errorf("%w: postgres_restore_anchor is required", ErrInvalidBackupMetadata)
	}
	if strings.TrimSpace(params.ObjectStoreRestoreAnchor) == "" {
		return createBackupSetParams{}, fmt.Errorf("%w: object_store_restore_anchor is required", ErrInvalidBackupMetadata)
	}
	if strings.TrimSpace(params.PostgresArtifactKey) == "" {
		return createBackupSetParams{}, fmt.Errorf("%w: postgres_artifact_key is required", ErrInvalidBackupMetadata)
	}
	if strings.TrimSpace(params.ObjectStoreArtifactKey) == "" {
		return createBackupSetParams{}, fmt.Errorf("%w: object_store_artifact_key is required", ErrInvalidBackupMetadata)
	}
	if strings.TrimSpace(params.IntegrityManifestKey) == "" {
		return createBackupSetParams{}, fmt.Errorf("%w: integrity_manifest_key is required", ErrInvalidBackupMetadata)
	}
	if !validSHA256Hex(params.PostgresArtifactSHA256) {
		return createBackupSetParams{}, fmt.Errorf("%w: postgres_artifact_sha256 is required", ErrInvalidBackupMetadata)
	}
	if !validSHA256Hex(params.ObjectStoreArtifactSHA256) {
		return createBackupSetParams{}, fmt.Errorf("%w: object_store_artifact_sha256 is required", ErrInvalidBackupMetadata)
	}
	if !validSHA256Hex(params.IntegrityManifestSHA256) {
		return createBackupSetParams{}, fmt.Errorf("%w: integrity_manifest_sha256 is required", ErrInvalidBackupMetadata)
	}
	if params.PostgresArtifactSizeBytes <= 0 {
		return createBackupSetParams{}, fmt.Errorf("%w: postgres_artifact_size_bytes is required", ErrInvalidBackupMetadata)
	}
	if params.ObjectStoreArtifactSizeBytes <= 0 {
		return createBackupSetParams{}, fmt.Errorf("%w: object_store_artifact_size_bytes is required", ErrInvalidBackupMetadata)
	}
	if params.IntegrityManifestSizeBytes <= 0 {
		return createBackupSetParams{}, fmt.Errorf("%w: integrity_manifest_size_bytes is required", ErrInvalidBackupMetadata)
	}
	if params.CreatedAt.IsZero() {
		params.CreatedAt = time.Now().UTC()
	}
	params.CreatedAt = params.CreatedAt.UTC()
	params.ConsistencyPointAt = params.ConsistencyPointAt.UTC()

	retentionFloor := params.CreatedAt.Add(MinimumRetentionDuration)
	if params.RetainedUntil.IsZero() {
		params.RetainedUntil = retentionFloor
	}
	params.RetainedUntil = params.RetainedUntil.UTC()
	if params.RetainedUntil.Before(retentionFloor) {
		return createBackupSetParams{}, fmt.Errorf("%w: retained_until before created_at plus 30 days", ErrRetentionFloor)
	}

	if params.PostgresRestoreAnchorRetainedUntil.IsZero() {
		params.PostgresRestoreAnchorRetainedUntil = params.RetainedUntil
	}
	params.PostgresRestoreAnchorRetainedUntil = params.PostgresRestoreAnchorRetainedUntil.UTC()
	if params.PostgresRestoreAnchorRetainedUntil.Before(retentionFloor) {
		return createBackupSetParams{}, fmt.Errorf("%w: postgres_restore_anchor_retained_until before created_at plus 30 days", ErrRetentionFloor)
	}

	if params.ObjectStoreRestoreAnchorRetainedUntil.IsZero() {
		params.ObjectStoreRestoreAnchorRetainedUntil = params.RetainedUntil
	}
	params.ObjectStoreRestoreAnchorRetainedUntil = params.ObjectStoreRestoreAnchorRetainedUntil.UTC()
	if params.ObjectStoreRestoreAnchorRetainedUntil.Before(retentionFloor) {
		return createBackupSetParams{}, fmt.Errorf("%w: object_store_restore_anchor_retained_until before created_at plus 30 days", ErrRetentionFloor)
	}

	return params, nil
}

func normalizeRestoreVerificationRunParams(params CreateRestoreVerificationRunParams) (CreateRestoreVerificationRunParams, error) {
	if params.RestoreVerificationRunID == uuid.Nil {
		params.RestoreVerificationRunID = uuid.New()
	}
	if params.BackupSetID == uuid.Nil {
		return CreateRestoreVerificationRunParams{}, fmt.Errorf("%w: backup_set_id is required", ErrInvalidBackupMetadata)
	}
	if params.StartedAt.IsZero() {
		return CreateRestoreVerificationRunParams{}, fmt.Errorf("%w: started_at is required", ErrInvalidBackupMetadata)
	}
	if params.CompletedAt.IsZero() {
		return CreateRestoreVerificationRunParams{}, fmt.Errorf("%w: completed_at is required", ErrInvalidBackupMetadata)
	}
	params.StartedAt = params.StartedAt.UTC()
	params.CompletedAt = params.CompletedAt.UTC()
	if params.CompletedAt.Before(params.StartedAt) {
		return CreateRestoreVerificationRunParams{}, fmt.Errorf("%w: completed_at before started_at", ErrInvalidBackupMetadata)
	}
	if !validVerificationState(params.VerificationState) || params.VerificationState == VerificationUnverified {
		return CreateRestoreVerificationRunParams{}, ErrInvalidVerificationState
	}
	if !validSHA256Hex(params.VerificationBasisSHA256) {
		return CreateRestoreVerificationRunParams{}, ErrInvalidVerificationBasis
	}
	if params.VerificationState == VerificationFailed {
		if strings.TrimSpace(params.FailureReason) == "" || strings.TrimSpace(params.FailureMessage) == "" {
			return CreateRestoreVerificationRunParams{}, fmt.Errorf("%w: failed restore verification requires redacted failure details", ErrInvalidBackupMetadata)
		}
		params.FailureReason = strings.TrimSpace(params.FailureReason)
		params.FailureMessage = strings.TrimSpace(params.FailureMessage)
		return params, nil
	}
	if strings.TrimSpace(params.FailureReason) != "" || strings.TrimSpace(params.FailureMessage) != "" {
		return CreateRestoreVerificationRunParams{}, fmt.Errorf("%w: successful restore verification cannot include failure details", ErrInvalidBackupMetadata)
	}
	if params.ConsistencyReport.AuthoritativeRowsSHA256 == "" ||
		params.ConsistencyReport.ChangeSetsSHA256 == "" ||
		params.ConsistencyReport.BlobHashesSHA256 == "" {
		return CreateRestoreVerificationRunParams{}, fmt.Errorf("%w: successful restore verification requires consistency hashes", ErrInvalidBackupMetadata)
	}
	return params, nil
}

func validateVerificationTransition(state VerificationState, lastVerifiedRestoreAt *time.Time, verificationBasisSHA256 string) error {
	if !validVerificationState(state) {
		return ErrInvalidVerificationState
	}
	if !validOptionalSHA256Hex(verificationBasisSHA256) {
		return ErrInvalidVerificationBasis
	}
	if state == VerificationUnverified {
		if lastVerifiedRestoreAt != nil {
			return ErrVerificationTimestampForbidden
		}
		return nil
	}
	if lastVerifiedRestoreAt == nil || lastVerifiedRestoreAt.IsZero() {
		return ErrVerificationTimestampRequired
	}
	normalized := lastVerifiedRestoreAt.UTC()
	*lastVerifiedRestoreAt = normalized
	return nil
}

func validVerificationState(state VerificationState) bool {
	switch state {
	case VerificationUnverified, VerificationVerified, VerificationFailed:
		return true
	default:
		return false
	}
}

func validSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func validOptionalSHA256Hex(value string) bool {
	return strings.TrimSpace(value) == "" || validSHA256Hex(value)
}

func normalizeAsOf(asOf time.Time) time.Time {
	if asOf.IsZero() {
		return time.Now().UTC()
	}
	return asOf.UTC()
}

func backupSetFromSQL(row sqlc.BackupSet) (BackupSet, error) {
	backupSetID, err := uuidFromPG(row.BackupSetID)
	if err != nil {
		return BackupSet{}, fmt.Errorf("backup_set_id: %w", err)
	}
	consistencyPointAt, err := timeFromPG(row.ConsistencyPointAt)
	if err != nil {
		return BackupSet{}, fmt.Errorf("consistency_point_at: %w", err)
	}
	createdAt, err := timeFromPG(row.CreatedAt)
	if err != nil {
		return BackupSet{}, fmt.Errorf("created_at: %w", err)
	}
	retainedUntil, err := timeFromPG(row.RetainedUntil)
	if err != nil {
		return BackupSet{}, fmt.Errorf("retained_until: %w", err)
	}
	postgresAnchorRetainedUntil, err := timeFromPG(row.PostgresRestoreAnchorRetainedUntil)
	if err != nil {
		return BackupSet{}, fmt.Errorf("postgres_restore_anchor_retained_until: %w", err)
	}
	objectStoreAnchorRetainedUntil, err := timeFromPG(row.ObjectStoreRestoreAnchorRetainedUntil)
	if err != nil {
		return BackupSet{}, fmt.Errorf("object_store_restore_anchor_retained_until: %w", err)
	}
	state := VerificationState(row.VerificationState)
	if !validVerificationState(state) {
		return BackupSet{}, fmt.Errorf("%w: %s", ErrInvalidVerificationState, row.VerificationState)
	}
	return BackupSet{
		BackupSetID:                           backupSetID,
		ConsistencyPointAt:                    consistencyPointAt,
		PostgresRestoreAnchor:                 row.PostgresRestoreAnchor,
		ObjectStoreRestoreAnchor:              row.ObjectStoreRestoreAnchor,
		PostgresArtifactKey:                   row.PostgresArtifactKey,
		PostgresArtifactSHA256:                row.PostgresArtifactSha256,
		PostgresArtifactSizeBytes:             row.PostgresArtifactSizeBytes,
		ObjectStoreArtifactKey:                row.ObjectStoreArtifactKey,
		ObjectStoreArtifactSHA256:             row.ObjectStoreArtifactSha256,
		ObjectStoreArtifactSizeBytes:          row.ObjectStoreArtifactSizeBytes,
		IntegrityManifestKey:                  row.IntegrityManifestKey,
		IntegrityManifestSHA256:               row.IntegrityManifestSha256,
		IntegrityManifestSizeBytes:            row.IntegrityManifestSizeBytes,
		CreatedAt:                             createdAt,
		RetainedUntil:                         retainedUntil,
		PostgresRestoreAnchorRetainedUntil:    postgresAnchorRetainedUntil,
		ObjectStoreRestoreAnchorRetainedUntil: objectStoreAnchorRetainedUntil,
		VerificationState:                     state,
		LastVerifiedRestoreAt:                 optionalTimeFromPG(row.LastVerifiedRestoreAt),
		LastVerificationBasisSHA256:           optionalTextFromPG(row.LastVerificationBasisSha256),
	}, nil
}

func restoreVerificationRunFromSQL(row sqlc.RestoreVerificationRun) (RestoreVerificationRun, error) {
	runID, err := uuidFromPG(row.RestoreVerificationRunID)
	if err != nil {
		return RestoreVerificationRun{}, fmt.Errorf("restore_verification_run_id: %w", err)
	}
	backupSetID, err := uuidFromPG(row.BackupSetID)
	if err != nil {
		return RestoreVerificationRun{}, fmt.Errorf("backup_set_id: %w", err)
	}
	startedAt, err := timeFromPG(row.StartedAt)
	if err != nil {
		return RestoreVerificationRun{}, fmt.Errorf("started_at: %w", err)
	}
	completedAt, err := timeFromPG(row.CompletedAt)
	if err != nil {
		return RestoreVerificationRun{}, fmt.Errorf("completed_at: %w", err)
	}
	state := VerificationState(row.VerificationState)
	if !validVerificationState(state) || state == VerificationUnverified {
		return RestoreVerificationRun{}, fmt.Errorf("%w: %s", ErrInvalidVerificationState, row.VerificationState)
	}
	return RestoreVerificationRun{
		RestoreVerificationRunID: runID,
		BackupSetID:              backupSetID,
		StartedAt:                startedAt,
		CompletedAt:              completedAt,
		VerificationState:        state,
		VerificationBasisSHA256:  row.VerificationBasisSha256,
		FailureReason:            optionalTextFromPG(row.FailureReason),
		FailureMessage:           optionalTextFromPG(row.FailureMessage),
		ConsistencyReport: RestoreConsistencyReport{
			AuthoritativeRowsSHA256: optionalTextFromPG(row.AuthoritativeRowsSha256),
			AuthoritativeRowCount:   optionalInt4FromPG(row.AuthoritativeRowCount),
			ChangeSetsSHA256:        optionalTextFromPG(row.ChangeSetsSha256),
			ChangeSetRowCount:       optionalInt4FromPG(row.ChangeSetRowCount),
			BlobHashesSHA256:        optionalTextFromPG(row.BlobHashesSha256),
			BlobCount:               optionalInt4FromPG(row.BlobCount),
		},
	}, nil
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.FromBytes(value.Bytes[:])
}

func pgTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func pgOptionalTimestamptz(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgTimestamptz(*value)
}

func pgOptionalText(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func pgOptionalInt4(value int) pgtype.Int4 {
	if value == 0 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(value), Valid: true}
}

func timeFromPG(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, errors.New("missing timestamp")
	}
	return value.Time.UTC(), nil
}

func optionalTimeFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	parsed := value.Time.UTC()
	return &parsed
}

func optionalTextFromPG(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func optionalInt4FromPG(value pgtype.Int4) int {
	if !value.Valid {
		return 0
	}
	return int(value.Int32)
}
