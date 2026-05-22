package recovery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	BackupIntegrityManifestSchemaID = "cartulary.backup_integrity_manifest.v1"
	backupStorageAnchorScheme       = "backup-storage://"
)

var (
	ErrInvalidBackupArtifact = errors.New("recovery: invalid backup artifact")
	ErrUnsupportedBackupRoot = errors.New("recovery: unsupported backup storage root")
)

type CaptureService struct {
	store   *Store
	storage BackupStorage
	now     func() time.Time
}

type BackupStorage interface {
	WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (BackupArtifactProof, error)
	ReadArtifact(ctx context.Context, key string) ([]byte, error)
}

type FilesystemBackupStorage struct {
	root string
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
}

type BackupArtifactProof struct {
	Key         string `json:"key"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentType string `json:"content_type"`
}

type BackupIntegrityManifest struct {
	SchemaID                              string              `json:"schema_id"`
	BackupSetID                           string              `json:"backup_set_id"`
	ConsistencyPointAt                    time.Time           `json:"consistency_point_at"`
	CreatedAt                             time.Time           `json:"created_at"`
	RetainedUntil                         time.Time           `json:"retained_until"`
	PostgresRestoreAnchor                 string              `json:"postgres_restore_anchor"`
	ObjectStoreRestoreAnchor              string              `json:"object_store_restore_anchor"`
	PostgresRestoreAnchorRetainedUntil    time.Time           `json:"postgres_restore_anchor_retained_until"`
	ObjectStoreRestoreAnchorRetainedUntil time.Time           `json:"object_store_restore_anchor_retained_until"`
	PostgresArtifact                      BackupArtifactProof `json:"postgres_artifact"`
	ObjectStoreArtifact                   BackupArtifactProof `json:"object_store_artifact"`
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

func NewCaptureService(store *Store, storage BackupStorage) *CaptureService {
	return &CaptureService{
		store:   store,
		storage: storage,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func NewBackupStorageFromConfig(cfg config.Config) (BackupStorage, error) {
	switch cfg.Roots.BackupStorage.BindingKind {
	case "filesystem_root":
		return NewFilesystemBackupStorage(cfg.Roots.BackupStorage.Path)
	case "managed_service":
		return nil, fmt.Errorf("%w: managed_service backup storage capture is not implemented", ErrUnsupportedBackupRoot)
	default:
		return nil, fmt.Errorf("%w: roots.backup_storage.binding_kind must be configured", ErrUnsupportedBackupRoot)
	}
}

func NewFilesystemBackupStorage(root string) (*FilesystemBackupStorage, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("%w: backup storage root path is required", ErrUnsupportedBackupRoot)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create backup storage root: %w", err)
	}
	return &FilesystemBackupStorage{root: filepath.Clean(root)}, nil
}

func (storage *FilesystemBackupStorage) WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (BackupArtifactProof, error) {
	if err := ctx.Err(); err != nil {
		return BackupArtifactProof{}, err
	}
	normalizedKey, err := normalizeArtifactKey(key)
	if err != nil {
		return BackupArtifactProof{}, err
	}
	if len(body) == 0 {
		return BackupArtifactProof{}, fmt.Errorf("%w: artifact body is empty", ErrInvalidBackupArtifact)
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	target := filepath.Join(storage.root, filepath.FromSlash(normalizedKey))
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return BackupArtifactProof{}, fmt.Errorf("create backup artifact directory: %w", err)
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- key is normalized beneath the configured backup storage root.
	if err != nil {
		return BackupArtifactProof{}, fmt.Errorf("create backup artifact %s: %w", normalizedKey, err)
	}
	_, writeErr := file.Write(body)
	closeErr := file.Close()
	if writeErr != nil {
		return BackupArtifactProof{}, fmt.Errorf("write backup artifact %s: %w", normalizedKey, writeErr)
	}
	if closeErr != nil {
		return BackupArtifactProof{}, fmt.Errorf("close backup artifact %s: %w", normalizedKey, closeErr)
	}
	stat, err := os.Stat(target) // #nosec G304 -- target is produced from a normalized backup artifact key.
	if err != nil {
		return BackupArtifactProof{}, fmt.Errorf("stat backup artifact %s: %w", normalizedKey, err)
	}
	if stat.Size() != int64(len(body)) {
		return BackupArtifactProof{}, fmt.Errorf("%w: artifact size mismatch for %s", ErrInvalidBackupArtifact, normalizedKey)
	}
	return BackupArtifactProof{
		Key:         normalizedKey,
		SHA256:      sha256Hex(body),
		SizeBytes:   stat.Size(),
		ContentType: contentType,
	}, nil
}

func (storage *FilesystemBackupStorage) ReadArtifact(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	normalizedKey, err := normalizeArtifactKey(key)
	if err != nil {
		return nil, err
	}
	target := filepath.Join(storage.root, filepath.FromSlash(normalizedKey))
	body, err := os.ReadFile(target) // #nosec G304 -- key is normalized beneath the configured backup storage root.
	if err != nil {
		return nil, fmt.Errorf("read backup artifact %s: %w", normalizedKey, err)
	}
	return body, nil
}

func VerifyArtifactProof(ctx context.Context, storage BackupStorage, proof BackupArtifactProof) ([]byte, error) {
	if storage == nil {
		return nil, fmt.Errorf("%w: backup storage is required", ErrInvalidBackupArtifact)
	}
	body, err := storage.ReadArtifact(ctx, proof.Key)
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
	if service == nil || service.store == nil || service.storage == nil {
		return BackupSet{}, fmt.Errorf("%w: capture service requires store and backup storage", ErrInvalidBackupMetadata)
	}
	if params.BackupSetID == uuid.Nil {
		params.BackupSetID = uuid.New()
	}
	if params.CreatedAt.IsZero() {
		params.CreatedAt = service.now().UTC()
	} else {
		params.CreatedAt = params.CreatedAt.UTC()
	}
	params.ConsistencyPointAt = params.ConsistencyPointAt.UTC()
	if params.RetainedUntil.IsZero() {
		params.RetainedUntil = params.CreatedAt.Add(MinimumRetentionDuration)
	} else {
		params.RetainedUntil = params.RetainedUntil.UTC()
	}
	if params.PostgresRestoreAnchorRetainedUntil.IsZero() {
		params.PostgresRestoreAnchorRetainedUntil = params.RetainedUntil
	} else {
		params.PostgresRestoreAnchorRetainedUntil = params.PostgresRestoreAnchorRetainedUntil.UTC()
	}
	if params.ObjectStoreRestoreAnchorRetainedUntil.IsZero() {
		params.ObjectStoreRestoreAnchorRetainedUntil = params.RetainedUntil
	} else {
		params.ObjectStoreRestoreAnchorRetainedUntil = params.ObjectStoreRestoreAnchorRetainedUntil.UTC()
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

	postgresAnchor := backupStorageAnchorScheme + postgresProof.Key
	objectAnchor := backupStorageAnchorScheme + objectProof.Key
	manifest := BackupIntegrityManifest{
		SchemaID:                              BackupIntegrityManifestSchemaID,
		BackupSetID:                           params.BackupSetID.String(),
		ConsistencyPointAt:                    params.ConsistencyPointAt,
		CreatedAt:                             params.CreatedAt,
		RetainedUntil:                         params.RetainedUntil,
		PostgresRestoreAnchor:                 postgresAnchor,
		ObjectStoreRestoreAnchor:              objectAnchor,
		PostgresRestoreAnchorRetainedUntil:    params.PostgresRestoreAnchorRetainedUntil,
		ObjectStoreRestoreAnchorRetainedUntil: params.ObjectStoreRestoreAnchorRetainedUntil,
		PostgresArtifact:                      postgresProof,
		ObjectStoreArtifact:                   objectProof,
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

func CaptureObjectStoreSnapshotArtifact(ctx context.Context, store objectstore.Store, prefix string) ([]byte, error) {
	if store == nil {
		return nil, fmt.Errorf("%w: object store is required", ErrInvalidBackupArtifact)
	}
	objects, err := store.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("list object store snapshot: %w", err)
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Key < objects[j].Key
	})
	items := make([]ObjectStoreSnapshotItem, 0, len(objects))
	for _, object := range objects {
		reader, info, err := store.ReadObject(ctx, object.Key, objectstore.ReadOptions{})
		if err != nil {
			return nil, fmt.Errorf("read object store snapshot object %s: %w", object.Key, err)
		}
		digest, err := hashReader(reader)
		closeErr := reader.Close()
		if err != nil {
			return nil, fmt.Errorf("hash object store snapshot object %s: %w", object.Key, err)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close object store snapshot object %s: %w", object.Key, closeErr)
		}
		items = append(items, ObjectStoreSnapshotItem{
			Key:         info.Key,
			SizeBytes:   info.Size,
			ContentType: info.ContentType,
			SHA256:      digest,
		})
	}
	return json.Marshal(ObjectStoreSnapshotArtifact{
		SchemaID: "cartulary.object_store_snapshot_artifact.v1",
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
		SchemaID: "cartulary.postgres_snapshot_artifact.v1",
		Tables:   tables,
	})
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

func hashReader(reader io.Reader) (string, error) {
	digest := sha256.New()
	if _, err := io.Copy(digest, reader); err != nil {
		return "", err
	}
	return hashHex(digest), nil
}

func hashHex(digest hash.Hash) string {
	return hex.EncodeToString(digest.Sum(nil))
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
