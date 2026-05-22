package recovery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type RestoreStep string

const (
	RestoreStepPostgresRestore    RestoreStep = "postgres_restore"
	RestoreStepObjectStoreRestore RestoreStep = "object_store_restore"
	RestoreStepProjectionRebuild  RestoreStep = "projection_rebuild"
	RestoreStepConsistencyCheck   RestoreStep = "consistency_check"
	RestoreStepReadiness          RestoreStep = "readiness"
)

type RestoreRunner struct {
	store   *Store
	storage BackupStorage
	now     func() time.Time
}

type RestoreTarget struct {
	Postgres    postgres.DB
	ObjectStore objectstore.Store
	Projections RestoreProjectionRebuilder
	Readiness   RestoreReadinessGate
	Observer    RestoreStepObserver
}

type RestoreProjectionRebuilder interface {
	RebuildRestoreProjections(ctx context.Context) error
}

type RestoreReadinessGate interface {
	MarkRestoreReady(ctx context.Context, result RestoreResult) error
}

type RestoreStepObserver interface {
	RecordRestoreStep(step RestoreStep)
}

type RestoreResult struct {
	BackupSet         BackupSet
	ConsistencyReport RestoreConsistencyReport
}

type RestoreConsistencyReport struct {
	AuthoritativeRowsSHA256 string
	AuthoritativeRowCount   int
	ChangeSetsSHA256        string
	ChangeSetRowCount       int
	BlobHashesSHA256        string
	BlobCount               int
}

type selectedRestoreArtifacts struct {
	PostgresSnapshot    PostgresSnapshotArtifact
	ObjectStoreSnapshot ObjectStoreSnapshotArtifact
}

func NewRestoreRunner(store *Store, storage BackupStorage) *RestoreRunner {
	return &RestoreRunner{
		store:   store,
		storage: storage,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (runner *RestoreRunner) RestoreLatestSuccessfulRetained(ctx context.Context, target RestoreTarget, asOf time.Time) (RestoreResult, error) {
	if runner == nil || runner.store == nil || runner.storage == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore runner requires store and backup storage", ErrInvalidBackupMetadata)
	}
	if target.Postgres == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore target postgres is required", ErrInvalidBackupArtifact)
	}
	if target.ObjectStore == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore target object store is required", ErrInvalidBackupArtifact)
	}
	if target.Projections == nil {
		return RestoreResult{}, fmt.Errorf("%w: restore projection rebuilder is required", ErrInvalidBackupArtifact)
	}
	if asOf.IsZero() {
		asOf = runner.now()
	}
	backupSet, err := runner.store.RestoreCandidateBackup(ctx, asOf)
	if err != nil {
		return RestoreResult{}, err
	}
	artifacts, err := runner.loadSelectedRestoreArtifacts(ctx, backupSet)
	if err != nil {
		return RestoreResult{}, err
	}

	recordStep(target.Observer, RestoreStepPostgresRestore)
	if err := restorePostgresSnapshot(ctx, target.Postgres, artifacts.PostgresSnapshot); err != nil {
		return RestoreResult{}, err
	}

	recordStep(target.Observer, RestoreStepObjectStoreRestore)
	if err := restoreObjectStoreSnapshot(ctx, target.ObjectStore, artifacts.ObjectStoreSnapshot); err != nil {
		return RestoreResult{}, err
	}

	recordStep(target.Observer, RestoreStepProjectionRebuild)
	if err := target.Projections.RebuildRestoreProjections(ctx); err != nil {
		return RestoreResult{}, err
	}

	recordStep(target.Observer, RestoreStepConsistencyCheck)
	report, err := verifyRestoredConsistency(ctx, target, artifacts)
	if err != nil {
		return RestoreResult{}, err
	}
	result := RestoreResult{
		BackupSet:         backupSet,
		ConsistencyReport: report,
	}

	if target.Readiness != nil {
		recordStep(target.Observer, RestoreStepReadiness)
		if err := target.Readiness.MarkRestoreReady(ctx, result); err != nil {
			return RestoreResult{}, err
		}
	}
	return result, nil
}

func (runner *RestoreRunner) loadSelectedRestoreArtifacts(ctx context.Context, backupSet BackupSet) (selectedRestoreArtifacts, error) {
	manifestProof := BackupArtifactProof{
		Key:       backupSet.IntegrityManifestKey,
		SHA256:    backupSet.IntegrityManifestSHA256,
		SizeBytes: backupSet.IntegrityManifestSizeBytes,
	}
	manifestBody, err := VerifyArtifactProof(ctx, runner.storage, manifestProof)
	if err != nil {
		return selectedRestoreArtifacts{}, fmt.Errorf("verify selected backup integrity manifest: %w", err)
	}
	manifest, err := DecodeIntegrityManifest(manifestBody)
	if err != nil {
		return selectedRestoreArtifacts{}, fmt.Errorf("%w: decode selected backup integrity manifest: %v", ErrInvalidBackupArtifact, err)
	}
	if err := validateSelectedRestoreManifest(backupSet, manifest); err != nil {
		return selectedRestoreArtifacts{}, err
	}
	postgresBody, err := VerifyArtifactProof(ctx, runner.storage, manifest.PostgresArtifact)
	if err != nil {
		return selectedRestoreArtifacts{}, fmt.Errorf("verify selected postgres artifact: %w", err)
	}
	objectBody, err := VerifyArtifactProof(ctx, runner.storage, manifest.ObjectStoreArtifact)
	if err != nil {
		return selectedRestoreArtifacts{}, fmt.Errorf("verify selected object-store artifact: %w", err)
	}
	postgresSnapshot, err := DecodePostgresSnapshotArtifact(postgresBody)
	if err != nil {
		return selectedRestoreArtifacts{}, err
	}
	objectSnapshot, err := DecodeObjectStoreSnapshotArtifact(objectBody)
	if err != nil {
		return selectedRestoreArtifacts{}, err
	}
	return selectedRestoreArtifacts{
		PostgresSnapshot:    postgresSnapshot,
		ObjectStoreSnapshot: objectSnapshot,
	}, nil
}

func validateSelectedRestoreManifest(backupSet BackupSet, manifest BackupIntegrityManifest) error {
	if manifest.SchemaID != BackupIntegrityManifestSchemaID {
		return fmt.Errorf("%w: unsupported integrity manifest schema %q", ErrInvalidBackupArtifact, manifest.SchemaID)
	}
	if manifest.BackupSetID != backupSet.BackupSetID.String() {
		return fmt.Errorf("%w: manifest backup_set_id does not match selected backup", ErrInvalidBackupArtifact)
	}
	if !manifest.ConsistencyPointAt.Equal(backupSet.ConsistencyPointAt) {
		return fmt.Errorf("%w: manifest consistency_point_at does not match selected backup", ErrInvalidBackupArtifact)
	}
	if manifest.PostgresRestoreAnchor != backupSet.PostgresRestoreAnchor ||
		manifest.ObjectStoreRestoreAnchor != backupSet.ObjectStoreRestoreAnchor {
		return fmt.Errorf("%w: manifest restore anchors do not match selected backup", ErrInvalidBackupArtifact)
	}
	if manifest.PostgresRestoreAnchor != backupStorageAnchorScheme+manifest.PostgresArtifact.Key ||
		manifest.ObjectStoreRestoreAnchor != backupStorageAnchorScheme+manifest.ObjectStoreArtifact.Key {
		return fmt.Errorf("%w: manifest restore anchors do not match artifact keys", ErrInvalidBackupArtifact)
	}
	if !backupProofMatches(manifest.PostgresArtifact, backupSet.PostgresArtifactKey, backupSet.PostgresArtifactSHA256, backupSet.PostgresArtifactSizeBytes) ||
		!backupProofMatches(manifest.ObjectStoreArtifact, backupSet.ObjectStoreArtifactKey, backupSet.ObjectStoreArtifactSHA256, backupSet.ObjectStoreArtifactSizeBytes) {
		return fmt.Errorf("%w: manifest artifact proofs do not match selected backup", ErrInvalidBackupArtifact)
	}
	return nil
}

func backupProofMatches(proof BackupArtifactProof, key string, sha256 string, sizeBytes int64) bool {
	return proof.Key == key && proof.SHA256 == sha256 && proof.SizeBytes == sizeBytes
}

func DecodePostgresSnapshotArtifact(body []byte) (PostgresSnapshotArtifact, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var artifact PostgresSnapshotArtifact
	if err := decoder.Decode(&artifact); err != nil {
		return PostgresSnapshotArtifact{}, fmt.Errorf("%w: decode postgres snapshot artifact: %v", ErrInvalidBackupArtifact, err)
	}
	if artifact.SchemaID != PostgresSnapshotArtifactSchemaID {
		return PostgresSnapshotArtifact{}, fmt.Errorf("%w: unsupported postgres snapshot schema %q", ErrInvalidBackupArtifact, artifact.SchemaID)
	}
	seen := make(map[string]struct{}, len(artifact.Tables))
	for _, table := range artifact.Tables {
		if !IsAuthoritativePostgresSnapshotTable(table.TableName) {
			return PostgresSnapshotArtifact{}, fmt.Errorf("%w: postgres snapshot contains non-authoritative table %q", ErrInvalidBackupArtifact, table.TableName)
		}
		if _, exists := seen[table.TableName]; exists {
			return PostgresSnapshotArtifact{}, fmt.Errorf("%w: postgres snapshot contains duplicate table %q", ErrInvalidBackupArtifact, table.TableName)
		}
		seen[table.TableName] = struct{}{}
		if table.RowCount != int64(len(table.Rows)) {
			return PostgresSnapshotArtifact{}, fmt.Errorf("%w: postgres snapshot row_count mismatch for %q", ErrInvalidBackupArtifact, table.TableName)
		}
	}
	sort.Slice(artifact.Tables, func(i, j int) bool {
		return artifact.Tables[i].TableName < artifact.Tables[j].TableName
	})
	return artifact, nil
}

func DecodeObjectStoreSnapshotArtifact(body []byte) (ObjectStoreSnapshotArtifact, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var artifact ObjectStoreSnapshotArtifact
	if err := decoder.Decode(&artifact); err != nil {
		return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: decode object-store snapshot artifact: %v", ErrInvalidBackupArtifact, err)
	}
	if artifact.SchemaID != ObjectStoreSnapshotArtifactSchemaID {
		return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: unsupported object-store snapshot schema %q", ErrInvalidBackupArtifact, artifact.SchemaID)
	}
	seen := make(map[string]struct{}, len(artifact.Objects))
	for index, item := range artifact.Objects {
		if strings.TrimSpace(item.Key) == "" {
			return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: object-store snapshot object key is required", ErrInvalidBackupArtifact)
		}
		if _, exists := seen[item.Key]; exists {
			return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: object-store snapshot contains duplicate key %q", ErrInvalidBackupArtifact, item.Key)
		}
		seen[item.Key] = struct{}{}
		body, err := base64.StdEncoding.DecodeString(item.BodyBase64)
		if err != nil {
			return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: decode object-store body for %s: %v", ErrInvalidBackupArtifact, item.Key, err)
		}
		if int64(len(body)) != item.SizeBytes {
			return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: object-store snapshot size mismatch for %s", ErrInvalidBackupArtifact, item.Key)
		}
		if sha256Hex(body) != item.SHA256 {
			return ObjectStoreSnapshotArtifact{}, fmt.Errorf("%w: object-store snapshot sha256 mismatch for %s", ErrInvalidBackupArtifact, item.Key)
		}
		artifact.Objects[index].ContentType = artifactContentType(BackupArtifact{ContentType: item.ContentType})
	}
	sort.Slice(artifact.Objects, func(i, j int) bool {
		return artifact.Objects[i].Key < artifact.Objects[j].Key
	})
	return artifact, nil
}

func restorePostgresSnapshot(ctx context.Context, db postgres.DB, artifact PostgresSnapshotArtifact) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin postgres restore: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
		return fmt.Errorf("disable postgres restore referential triggers: %w", err)
	}
	tableNames := make([]string, 0, len(artifact.Tables))
	for _, table := range artifact.Tables {
		tableNames = append(tableNames, table.TableName)
	}
	if len(tableNames) > 0 {
		truncateSQL := "TRUNCATE " + sanitizedTableList(tableNames) + " RESTART IDENTITY CASCADE"
		if _, err := tx.Exec(ctx, truncateSQL); err != nil {
			return fmt.Errorf("truncate postgres restore target tables: %w", err)
		}
	}
	for _, table := range artifact.Tables {
		identifier := pgx.Identifier{table.TableName}.Sanitize()
		insertSQL := fmt.Sprintf("INSERT INTO %s SELECT * FROM jsonb_populate_record(NULL::%s, $1::jsonb)", identifier, identifier)
		for _, rawRow := range table.Rows {
			if _, err := tx.Exec(ctx, insertSQL, string(rawRow)); err != nil {
				return fmt.Errorf("restore postgres table %s: %w", table.TableName, err)
			}
		}
		if err := resetOwnedSequences(ctx, tx, table.TableName); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit postgres restore: %w", err)
	}
	return nil
}

func sanitizedTableList(tableNames []string) string {
	parts := make([]string, 0, len(tableNames))
	for _, tableName := range tableNames {
		parts = append(parts, pgx.Identifier{tableName}.Sanitize())
	}
	return strings.Join(parts, ", ")
}

func resetOwnedSequences(ctx context.Context, tx pgx.Tx, tableName string) error {
	rows, err := tx.Query(ctx, `
SELECT column_name
FROM information_schema.columns
WHERE table_schema = 'public'
  AND table_name = $1
  AND column_default LIKE 'nextval(%'
ORDER BY ordinal_position ASC
`, tableName)
	if err != nil {
		return fmt.Errorf("list postgres restore sequences for %s: %w", tableName, err)
	}

	columnNames := make([]string, 0)
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			rows.Close()
			return fmt.Errorf("scan postgres restore sequence for %s: %w", tableName, err)
		}
		columnNames = append(columnNames, columnName)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate postgres restore sequences for %s: %w", tableName, err)
	}
	rows.Close()

	for _, columnName := range columnNames {
		var sequenceName pgtype.Text
		if err := tx.QueryRow(ctx, `SELECT pg_get_serial_sequence($1, $2)`, "public."+tableName, columnName).Scan(&sequenceName); err != nil {
			return fmt.Errorf("resolve postgres restore sequence for %s.%s: %w", tableName, columnName, err)
		}
		if !sequenceName.Valid || sequenceName.String == "" {
			continue
		}
		identifier := pgx.Identifier{tableName}.Sanitize()
		columnIdentifier := pgx.Identifier{columnName}.Sanitize()
		var nextValue int64
		query := fmt.Sprintf("SELECT COALESCE(MAX(%s), 0)::bigint + 1 FROM %s", columnIdentifier, identifier)
		if err := tx.QueryRow(ctx, query).Scan(&nextValue); err != nil {
			return fmt.Errorf("compute postgres restore sequence value for %s.%s: %w", tableName, columnName, err)
		}
		if _, err := tx.Exec(ctx, `SELECT setval($1, $2, false)`, sequenceName.String, nextValue); err != nil {
			return fmt.Errorf("reset postgres restore sequence for %s.%s: %w", tableName, columnName, err)
		}
	}
	return nil
}

func restoreObjectStoreSnapshot(ctx context.Context, store objectstore.Store, artifact ObjectStoreSnapshotArtifact) error {
	existing, err := store.ListObjects(ctx, "")
	if err != nil {
		return fmt.Errorf("list object-store restore target: %w", err)
	}
	sort.Slice(existing, func(i, j int) bool {
		return existing[i].Key < existing[j].Key
	})
	for _, object := range existing {
		if err := store.DeleteObject(ctx, object.Key); err != nil {
			return fmt.Errorf("delete object-store restore target object %s: %w", object.Key, err)
		}
	}
	for _, item := range artifact.Objects {
		body, err := base64.StdEncoding.DecodeString(item.BodyBase64)
		if err != nil {
			return fmt.Errorf("%w: decode object-store restore body for %s: %v", ErrInvalidBackupArtifact, item.Key, err)
		}
		if err := store.PutObject(ctx, item.Key, bytes.NewReader(body), int64(len(body)), item.ContentType); err != nil {
			return fmt.Errorf("restore object-store object %s: %w", item.Key, err)
		}
	}
	return nil
}

func verifyRestoredConsistency(ctx context.Context, target RestoreTarget, artifacts selectedRestoreArtifacts) (RestoreConsistencyReport, error) {
	authoritativeDigest, authoritativeCount, err := postgresSnapshotDigest(artifacts.PostgresSnapshot, nil)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	targetSnapshotBody, err := CapturePostgresSnapshotArtifact(ctx, target.Postgres)
	if err != nil {
		return RestoreConsistencyReport{}, fmt.Errorf("capture restored postgres consistency snapshot: %w", err)
	}
	targetSnapshot, err := DecodePostgresSnapshotArtifact(targetSnapshotBody)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	targetDigest, _, err := postgresSnapshotDigest(targetSnapshot, nil)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	if targetDigest != authoritativeDigest {
		return RestoreConsistencyReport{}, fmt.Errorf("%w: restored authoritative row digest mismatch", ErrInvalidBackupArtifact)
	}

	changeSetDigest, changeSetCount, err := postgresSnapshotDigest(artifacts.PostgresSnapshot, isChangeSetTable)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	targetChangeSetDigest, _, err := postgresSnapshotDigest(targetSnapshot, isChangeSetTable)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	if targetChangeSetDigest != changeSetDigest {
		return RestoreConsistencyReport{}, fmt.Errorf("%w: restored change-set digest mismatch", ErrInvalidBackupArtifact)
	}

	blobDigest, blobCount, err := verifyRestoredBlobHashes(ctx, target, artifacts.ObjectStoreSnapshot)
	if err != nil {
		return RestoreConsistencyReport{}, err
	}
	return RestoreConsistencyReport{
		AuthoritativeRowsSHA256: authoritativeDigest,
		AuthoritativeRowCount:   authoritativeCount,
		ChangeSetsSHA256:        changeSetDigest,
		ChangeSetRowCount:       changeSetCount,
		BlobHashesSHA256:        blobDigest,
		BlobCount:               blobCount,
	}, nil
}

func postgresSnapshotDigest(artifact PostgresSnapshotArtifact, include func(string) bool) (string, int, error) {
	digest := sha256.New()
	rowCount := 0
	tables := append([]PostgresSnapshotTable(nil), artifact.Tables...)
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].TableName < tables[j].TableName
	})
	for _, table := range tables {
		if include != nil && !include(table.TableName) {
			continue
		}
		_, _ = digest.Write([]byte("table:" + table.TableName + "\n"))
		rows := make([]string, 0, len(table.Rows))
		for _, rawRow := range table.Rows {
			normalized, err := normalizeJSONForDigest(rawRow)
			if err != nil {
				return "", 0, fmt.Errorf("normalize postgres snapshot row for %s: %w", table.TableName, err)
			}
			rows = append(rows, normalized)
		}
		sort.Strings(rows)
		for _, row := range rows {
			_, _ = digest.Write([]byte(row))
			_, _ = digest.Write([]byte("\n"))
			rowCount++
		}
	}
	return hex.EncodeToString(digest.Sum(nil)), rowCount, nil
}

func normalizeJSONForDigest(raw json.RawMessage) (string, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	normalized, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(normalized), nil
}

func isChangeSetTable(tableName string) bool {
	switch tableName {
	case "change_sets", "change_set_mutations", "record_history_entry_refs", "record_revisions":
		return true
	default:
		return false
	}
}

func verifyRestoredBlobHashes(ctx context.Context, target RestoreTarget, artifact ObjectStoreSnapshotArtifact) (string, int, error) {
	digest := sha256.New()
	objects := append([]ObjectStoreSnapshotItem(nil), artifact.Objects...)
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Key < objects[j].Key
	})
	for _, item := range objects {
		reader, info, err := target.ObjectStore.ReadObject(ctx, item.Key, objectstore.ReadOptions{})
		if err != nil {
			return "", 0, fmt.Errorf("read restored object %s: %w", item.Key, err)
		}
		body, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return "", 0, fmt.Errorf("read restored object body %s: %w", item.Key, readErr)
		}
		if closeErr != nil {
			return "", 0, fmt.Errorf("close restored object %s: %w", item.Key, closeErr)
		}
		if info.Size != item.SizeBytes || int64(len(body)) != item.SizeBytes || sha256Hex(body) != item.SHA256 {
			return "", 0, fmt.Errorf("%w: restored object proof mismatch for %s", ErrInvalidBackupArtifact, item.Key)
		}
		_, _ = digest.Write([]byte("object:" + item.Key + ":" + item.SHA256 + "\n"))
	}

	rowDigest, err := verifyRestoredBlobRows(ctx, target.Postgres, target.ObjectStore)
	if err != nil {
		return "", 0, err
	}
	_, _ = digest.Write([]byte("blob_rows:" + rowDigest + "\n"))
	return hex.EncodeToString(digest.Sum(nil)), len(objects), nil
}

func verifyRestoredBlobRows(ctx context.Context, db postgres.DB, store objectstore.Store) (string, error) {
	rows, err := db.Query(ctx, `
SELECT b.storage_key,
       b.byte_size,
       b.observed_size,
       b.expected_sha256_hex,
       b.observed_sha256_hex,
       e.blob_hash
  FROM object_blobs b
  LEFT JOIN evidence e
    ON e.object_blob_id = b.object_blob_id
 WHERE b.upload_state = 'available'
 ORDER BY b.storage_key ASC, b.object_blob_id ASC
`)
	if err != nil {
		return "", fmt.Errorf("list restored blob rows: %w", err)
	}
	defer rows.Close()

	digest := sha256.New()
	for rows.Next() {
		var storageKey string
		var byteSize int64
		var observedSize pgtype.Int8
		var expectedSHA pgtype.Text
		var observedSHA pgtype.Text
		var blobHash pgtype.Text
		if err := rows.Scan(&storageKey, &byteSize, &observedSize, &expectedSHA, &observedSHA, &blobHash); err != nil {
			return "", fmt.Errorf("scan restored blob row: %w", err)
		}
		reader, _, err := store.ReadObject(ctx, storageKey, objectstore.ReadOptions{})
		if err != nil {
			return "", fmt.Errorf("read restored blob row object %s: %w", storageKey, err)
		}
		body, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return "", fmt.Errorf("read restored blob row body %s: %w", storageKey, readErr)
		}
		if closeErr != nil {
			return "", fmt.Errorf("close restored blob row object %s: %w", storageKey, closeErr)
		}
		sha := sha256Hex(body)
		if int64(len(body)) != byteSize {
			return "", fmt.Errorf("%w: restored blob row byte_size mismatch for %s", ErrInvalidBackupArtifact, storageKey)
		}
		if observedSize.Valid && observedSize.Int64 != int64(len(body)) {
			return "", fmt.Errorf("%w: restored blob row observed_size mismatch for %s", ErrInvalidBackupArtifact, storageKey)
		}
		for label, value := range map[string]pgtype.Text{
			"expected_sha256_hex": expectedSHA,
			"observed_sha256_hex": observedSHA,
			"blob_hash":           blobHash,
		} {
			if !value.Valid || strings.TrimSpace(value.String) == "" {
				continue
			}
			if !blobHashMatches(value.String, sha) {
				return "", fmt.Errorf("%w: restored blob row %s mismatch for %s", ErrInvalidBackupArtifact, label, storageKey)
			}
		}
		_, _ = digest.Write([]byte(storageKey + ":" + sha + "\n"))
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate restored blob rows: %w", err)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func blobHashMatches(value string, sha string) bool {
	normalized := strings.TrimSpace(strings.ToLower(value))
	return normalized == sha || normalized == "sha256:"+sha
}

func recordStep(observer RestoreStepObserver, step RestoreStep) {
	if observer != nil {
		observer.RecordRestoreStep(step)
	}
}
