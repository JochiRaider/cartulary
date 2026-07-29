package recovery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	BackupIntegrityManifestSchemaID     = "cartulary.backup_integrity_manifest.v2"
	PostgresSnapshotArtifactSchemaID    = "cartulary.postgres_snapshot_artifact.v1"
	ObjectStoreSnapshotArtifactSchemaID = "cartulary.object_store_snapshot_artifact.v2"
	backupStorageAnchorScheme           = "backup-storage://"
)

var (
	ErrInvalidBackupArtifact = errors.New("recovery: invalid backup artifact")
)

type CaptureService struct {
	store            backupRepository
	storage          BackupStorage
	extensionBackups *ExtensionBackupCatalog
	now              func() time.Time
}

type BackupStorage interface {
	WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (BackupArtifactProof, error)
	ReadArtifact(ctx context.Context, key string, maxBytes int64) ([]byte, error)
}

func CloseBackupStorage(storage BackupStorage) error {
	if closer, ok := storage.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

type BackupArtifact struct {
	Body        []byte
	ContentType string
}

type CaptureBackupSetParams struct {
	BackupSetID                           uuid.UUID
	ConsistencyPointAt                    time.Time
	CreatedAt                             time.Time
	RetainedUntil                         time.Time
	PostgresRestoreAnchorRetainedUntil    time.Time
	ObjectStoreRestoreAnchorRetainedUntil time.Time
	PostgresArtifact                      BackupArtifact
	ObjectStoreArtifact                   BackupArtifact
	ObjectStoreBackupManifestArtifact     BackupArtifact
	ObjectStoreBackupSummaryArtifact      BackupArtifact
}

type BackupArtifactProof struct {
	Key         string `json:"key"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentType string `json:"content_type"`
}

type BackupIntegrityManifest struct {
	SchemaID                              string                       `json:"schema_id"`
	BackupSetID                           string                       `json:"backup_set_id"`
	ConsistencyPointAt                    time.Time                    `json:"consistency_point_at"`
	CreatedAt                             time.Time                    `json:"created_at"`
	RetainedUntil                         time.Time                    `json:"retained_until"`
	StorageEncryption                     BackupStorageEncryptionProof `json:"storage_encryption"`
	PostgresRestoreAnchor                 string                       `json:"postgres_restore_anchor"`
	ObjectStoreRestoreAnchor              string                       `json:"object_store_restore_anchor"`
	PostgresRestoreAnchorRetainedUntil    time.Time                    `json:"postgres_restore_anchor_retained_until"`
	ObjectStoreRestoreAnchorRetainedUntil time.Time                    `json:"object_store_restore_anchor_retained_until"`
	PostgresArtifact                      BackupArtifactProof          `json:"postgres_artifact"`
	ObjectStoreArtifact                   BackupArtifactProof          `json:"object_store_artifact"`
	ObjectStoreBackupManifestArtifact     *BackupArtifactProof         `json:"object_store_backup_manifest_artifact,omitempty"`
	ObjectStoreBackupSummaryArtifact      *BackupArtifactProof         `json:"object_store_backup_summary_artifact,omitempty"`
	ExtensionBindings                     []ExtensionBindingProof      `json:"extension_bindings"`
}

type ObjectStoreSnapshotArtifact struct {
	SchemaID string                    `json:"schema_id"`
	Objects  []ObjectStoreSnapshotItem `json:"objects"`
}

type ObjectStoreSnapshotItem struct {
	Key         string `json:"key"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentType string `json:"content_type"`
	SHA256      string `json:"sha256"`
	BodyBase64  string `json:"body_base64"`
}

type PostgresSnapshotArtifact struct {
	SchemaID string                  `json:"schema_id"`
	Tables   []PostgresSnapshotTable `json:"tables"`
}

type PostgresSnapshotTable struct {
	TableName string            `json:"table_name"`
	RowCount  int64             `json:"row_count"`
	Rows      []json.RawMessage `json:"rows"`
}

func NewCaptureService(store backupRepository, storage BackupStorage, extensionBackups *ExtensionBackupCatalog) *CaptureService {
	return &CaptureService{
		store:            store,
		storage:          storage,
		extensionBackups: extensionBackups,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func VerifyArtifactProof(ctx context.Context, storage BackupStorage, proof BackupArtifactProof) ([]byte, error) {
	if storage == nil {
		return nil, fmt.Errorf("%w: backup storage is required", ErrInvalidBackupArtifact)
	}
	body, err := storage.ReadArtifact(ctx, proof.Key, proof.SizeBytes)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) != proof.SizeBytes {
		return nil, fmt.Errorf("%w: artifact size mismatch for %s", ErrInvalidBackupArtifact, proof.Key)
	}
	if sha256Hex(body) != proof.SHA256 {
		return nil, fmt.Errorf("%w: artifact sha256 mismatch for %s", ErrInvalidBackupArtifact, proof.Key)
	}
	return body, nil
}

func (service *CaptureService) CaptureBackupSet(ctx context.Context, params CaptureBackupSetParams) (BackupSet, error) {
	if service == nil || service.store == nil || service.storage == nil || service.extensionBackups == nil {
		return BackupSet{}, fmt.Errorf("%w: capture service requires store and backup storage", ErrInvalidBackupMetadata)
	}
	storageEncryption, err := backupStorageEncryptionProof(service.storage)
	if err != nil {
		return BackupSet{}, err
	}
	if params.BackupSetID == uuid.Nil {
		params.BackupSetID = uuid.New()
	}
	if params.CreatedAt.IsZero() {
		params.CreatedAt = service.now().UTC()
	} else {
		params.CreatedAt = params.CreatedAt.UTC()
	}
	params.CreatedAt = backupTimestamp(params.CreatedAt)
	params.ConsistencyPointAt = backupTimestamp(params.ConsistencyPointAt.UTC())
	if params.RetainedUntil.IsZero() {
		params.RetainedUntil = params.CreatedAt.Add(MinimumRetentionDuration)
	} else {
		params.RetainedUntil = params.RetainedUntil.UTC()
	}
	params.RetainedUntil = backupTimestamp(params.RetainedUntil)
	if params.PostgresRestoreAnchorRetainedUntil.IsZero() {
		params.PostgresRestoreAnchorRetainedUntil = params.RetainedUntil
	} else {
		params.PostgresRestoreAnchorRetainedUntil = params.PostgresRestoreAnchorRetainedUntil.UTC()
	}
	params.PostgresRestoreAnchorRetainedUntil = backupTimestamp(params.PostgresRestoreAnchorRetainedUntil)
	if params.ObjectStoreRestoreAnchorRetainedUntil.IsZero() {
		params.ObjectStoreRestoreAnchorRetainedUntil = params.RetainedUntil
	} else {
		params.ObjectStoreRestoreAnchorRetainedUntil = params.ObjectStoreRestoreAnchorRetainedUntil.UTC()
	}
	params.ObjectStoreRestoreAnchorRetainedUntil = backupTimestamp(params.ObjectStoreRestoreAnchorRetainedUntil)

	postgresSnapshot, err := DecodePostgresSnapshotArtifact(params.PostgresArtifact.Body)
	if err != nil {
		return BackupSet{}, err
	}
	extensionBindings, err := captureExtensionBindingProofs(service.extensionBackups, postgresSnapshot)
	if err != nil {
		return BackupSet{}, err
	}

	prefix := "backup_sets/" + params.BackupSetID.String()
	postgresProof, err := service.storage.WriteArtifact(ctx, prefix+"/postgres-artifact.json", params.PostgresArtifact.Body, artifactContentType(params.PostgresArtifact))
	if err != nil {
		return BackupSet{}, fmt.Errorf("capture postgres backup artifact: %w", err)
	}
	objectProof, err := service.storage.WriteArtifact(ctx, prefix+"/object-store-artifact.json", params.ObjectStoreArtifact.Body, artifactContentType(params.ObjectStoreArtifact))
	if err != nil {
		return BackupSet{}, fmt.Errorf("capture object-store backup artifact: %w", err)
	}
	var objectManifestProof *BackupArtifactProof
	if len(params.ObjectStoreBackupManifestArtifact.Body) > 0 {
		proof, err := service.storage.WriteArtifact(ctx, prefix+"/object-store-backup-manifest.json", params.ObjectStoreBackupManifestArtifact.Body, artifactContentType(params.ObjectStoreBackupManifestArtifact))
		if err != nil {
			return BackupSet{}, fmt.Errorf("capture object-store backup manifest artifact: %w", err)
		}
		objectManifestProof = &proof
	}
	var objectSummaryProof *BackupArtifactProof
	if len(params.ObjectStoreBackupSummaryArtifact.Body) > 0 {
		proof, err := service.storage.WriteArtifact(ctx, prefix+"/object-store-backup-summary.json", params.ObjectStoreBackupSummaryArtifact.Body, artifactContentType(params.ObjectStoreBackupSummaryArtifact))
		if err != nil {
			return BackupSet{}, fmt.Errorf("capture object-store backup summary artifact: %w", err)
		}
		objectSummaryProof = &proof
	}

	postgresAnchor := backupStorageAnchorScheme + postgresProof.Key
	objectAnchor := backupStorageAnchorScheme + objectProof.Key
	manifest := BackupIntegrityManifest{
		SchemaID:                              BackupIntegrityManifestSchemaID,
		BackupSetID:                           params.BackupSetID.String(),
		ConsistencyPointAt:                    params.ConsistencyPointAt,
		CreatedAt:                             params.CreatedAt,
		RetainedUntil:                         params.RetainedUntil,
		StorageEncryption:                     storageEncryption,
		PostgresRestoreAnchor:                 postgresAnchor,
		ObjectStoreRestoreAnchor:              objectAnchor,
		PostgresRestoreAnchorRetainedUntil:    params.PostgresRestoreAnchorRetainedUntil,
		ObjectStoreRestoreAnchorRetainedUntil: params.ObjectStoreRestoreAnchorRetainedUntil,
		PostgresArtifact:                      postgresProof,
		ObjectStoreArtifact:                   objectProof,
		ObjectStoreBackupManifestArtifact:     objectManifestProof,
		ObjectStoreBackupSummaryArtifact:      objectSummaryProof,
		ExtensionBindings:                     extensionBindings,
	}
	manifestBody, err := json.Marshal(manifest)
	if err != nil {
		return BackupSet{}, fmt.Errorf("encode backup integrity manifest: %w", err)
	}
	manifestProof, err := service.storage.WriteArtifact(ctx, prefix+"/integrity-manifest.json", manifestBody, "application/json")
	if err != nil {
		return BackupSet{}, fmt.Errorf("capture backup integrity manifest: %w", err)
	}

	return service.store.createCapturedBackupSet(ctx, createBackupSetParams{
		BackupSetID:                           params.BackupSetID,
		ConsistencyPointAt:                    params.ConsistencyPointAt,
		PostgresRestoreAnchor:                 postgresAnchor,
		ObjectStoreRestoreAnchor:              objectAnchor,
		PostgresArtifactKey:                   postgresProof.Key,
		PostgresArtifactSHA256:                postgresProof.SHA256,
		PostgresArtifactSizeBytes:             postgresProof.SizeBytes,
		ObjectStoreArtifactKey:                objectProof.Key,
		ObjectStoreArtifactSHA256:             objectProof.SHA256,
		ObjectStoreArtifactSizeBytes:          objectProof.SizeBytes,
		IntegrityManifestKey:                  manifestProof.Key,
		IntegrityManifestSHA256:               manifestProof.SHA256,
		IntegrityManifestSizeBytes:            manifestProof.SizeBytes,
		CreatedAt:                             params.CreatedAt,
		RetainedUntil:                         params.RetainedUntil,
		PostgresRestoreAnchorRetainedUntil:    params.PostgresRestoreAnchorRetainedUntil,
		ObjectStoreRestoreAnchorRetainedUntil: params.ObjectStoreRestoreAnchorRetainedUntil,
	})
}

func backupTimestamp(value time.Time) time.Time {
	return value.UTC().Truncate(time.Microsecond)
}

func CaptureObjectStoreSnapshotArtifact(ctx context.Context, store objectstore.Store, prefix string) ([]byte, error) {
	if store == nil {
		return nil, fmt.Errorf("%w: object store is required", ErrInvalidBackupArtifact)
	}
	objects, err := listBackupManifestObjects(ctx, store, prefix)
	if err != nil {
		return nil, fmt.Errorf("list object store snapshot: %w", err)
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Key < objects[j].Key
	})
	items := make([]ObjectStoreSnapshotItem, 0, len(objects))
	for _, object := range objects {
		reader, info, err := getBackupManifestObject(ctx, store, object.Key)
		if err != nil {
			return nil, fmt.Errorf("read object store snapshot object %s: %w", object.Key, err)
		}
		body, err := io.ReadAll(reader)
		closeErr := reader.Close()
		if err != nil {
			return nil, fmt.Errorf("read object store snapshot object body %s: %w", object.Key, err)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close object store snapshot object %s: %w", object.Key, closeErr)
		}
		if int64(len(body)) != info.Size {
			return nil, fmt.Errorf("%w: object store snapshot size mismatch for %s", ErrInvalidBackupArtifact, object.Key)
		}
		items = append(items, ObjectStoreSnapshotItem{
			Key:         info.Key,
			SizeBytes:   info.Size,
			ContentType: info.ContentType,
			SHA256:      sha256Hex(body),
			BodyBase64:  base64.StdEncoding.EncodeToString(body),
		})
	}
	return json.Marshal(ObjectStoreSnapshotArtifact{
		SchemaID: ObjectStoreSnapshotArtifactSchemaID,
		Objects:  items,
	})
}

func CapturePostgresSnapshotArtifact(ctx context.Context, db postgres.DB) ([]byte, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: postgres DB is required", ErrInvalidBackupArtifact)
	}
	rows, err := db.Query(ctx, `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
ORDER BY table_name ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list postgres catalog tables: %w", err)
	}
	defer rows.Close()

	tables := make([]PostgresSnapshotTable, 0)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("scan postgres catalog table: %w", err)
		}
		if !IsAuthoritativePostgresSnapshotTable(tableName) {
			continue
		}
		identifier := pgx.Identifier{tableName}.Sanitize()
		query := fmt.Sprintf(`
SELECT COALESCE(jsonb_agg(row_json ORDER BY row_json::text), '[]'::jsonb)
FROM (
    SELECT to_jsonb(snapshot_row) AS row_json
    FROM %s AS snapshot_row
) AS snapshot_rows
`, identifier)
		var rawRows []byte
		if err := db.QueryRow(ctx, query).Scan(&rawRows); err != nil {
			return nil, fmt.Errorf("snapshot postgres table %s: %w", tableName, err)
		}
		var tableRows []json.RawMessage
		if err := json.Unmarshal(rawRows, &tableRows); err != nil {
			return nil, fmt.Errorf("decode postgres table snapshot %s: %w", tableName, err)
		}
		tables = append(tables, PostgresSnapshotTable{
			TableName: tableName,
			RowCount:  int64(len(tableRows)),
			Rows:      tableRows,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate postgres catalog tables: %w", err)
	}
	return json.Marshal(PostgresSnapshotArtifact{
		SchemaID: PostgresSnapshotArtifactSchemaID,
		Tables:   tables,
	})
}

func IsAuthoritativePostgresSnapshotTable(tableName string) bool {
	normalized := strings.TrimSpace(strings.ToLower(tableName))
	if normalized == "" || strings.HasSuffix(normalized, "_grid_projection") {
		return false
	}
	switch normalized {
	case "backup_sets",
		"evidence_access_handles",
		"goose_db_version",
		"pending_totp_enrollments",
		"restore_verification_runs",
		"route_idempotency",
		"schema_migration_lineage",
		"user_sessions":
		return false
	default:
		return true
	}
}

func artifactContentType(artifact BackupArtifact) string {
	if strings.TrimSpace(artifact.ContentType) == "" {
		return "application/octet-stream"
	}
	return artifact.ContentType
}

func normalizeArtifactKey(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", fmt.Errorf("%w: artifact key is required", ErrInvalidBackupArtifact)
	}
	if path.IsAbs(trimmed) {
		return "", fmt.Errorf("%w: artifact key must be relative", ErrInvalidBackupArtifact)
	}
	normalized := path.Clean(trimmed)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") || strings.Contains(normalized, "/../") {
		return "", fmt.Errorf("%w: artifact key escapes backup storage root", ErrInvalidBackupArtifact)
	}
	return normalized, nil
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func DecodeIntegrityManifest(body []byte) (BackupIntegrityManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var manifest BackupIntegrityManifest
	if err := decoder.Decode(&manifest); err != nil {
		return BackupIntegrityManifest{}, err
	}
	return manifest, nil
}
