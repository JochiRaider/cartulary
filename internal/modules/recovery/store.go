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
	ErrLatestSuccessfulBackupStale    = errors.New("recovery: latest successful retained backup is older than 24 hours")
	ErrInvalidBackupMetadata          = errors.New("recovery: invalid backup metadata")
	ErrInvalidVerificationState       = errors.New("recovery: invalid verification state")
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

func (s *Store) UpdateVerificationState(ctx context.Context, backupSetID uuid.UUID, state VerificationState, lastVerifiedRestoreAt *time.Time) (BackupSet, error) {
	if err := validateVerificationTransition(state, lastVerifiedRestoreAt); err != nil {
		return BackupSet{}, err
	}
	row, err := sqlc.New(s.pool).UpdateBackupSetVerificationState(ctx, sqlc.UpdateBackupSetVerificationStateParams{
		BackupSetID:           pgUUID(backupSetID),
		VerificationState:     string(state),
		LastVerifiedRestoreAt: pgOptionalTimestamptz(lastVerifiedRestoreAt),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return BackupSet{}, ErrBackupSetNotFound
	}
	if err != nil {
		return BackupSet{}, fmt.Errorf("update backup verification state: %w", err)
	}
	return backupSetFromSQL(row)
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

func validateVerificationTransition(state VerificationState, lastVerifiedRestoreAt *time.Time) error {
	if !validVerificationState(state) {
		return ErrInvalidVerificationState
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
