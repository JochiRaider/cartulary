package testsupport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type TargetFixture struct {
	Env         map[string]string
	Postgres    *pgxpool.Pool
	ObjectStore objectstore.Store
	cleanupOnce *sync.Once
}

func NewTargetFixture(env map[string]string, postgres *pgxpool.Pool, store objectstore.Store) TargetFixture {
	return TargetFixture{
		Env:         maps.Clone(env),
		Postgres:    postgres,
		ObjectStore: store,
		cleanupOnce: &sync.Once{},
	}
}

func (fixture TargetFixture) Cleanup() {
	if fixture.cleanupOnce == nil {
		return
	}
	fixture.cleanupOnce.Do(func() {
		clear(fixture.Env)
	})
}

type EvidenceLocation struct {
	ResultsRoot string
	RunID       string
	Target      string
	Group       string
}

func (location EvidenceLocation) Dir() (string, error) {
	parts := []struct {
		name  string
		value string
	}{
		{name: "results root", value: location.ResultsRoot},
		{name: "run ID", value: location.RunID},
		{name: "target", value: location.Target},
		{name: "group", value: location.Group},
	}
	for _, part := range parts {
		if strings.TrimSpace(part.value) == "" {
			return "", fmt.Errorf("recovery evidence %s is required", part.name)
		}
	}
	for _, part := range parts[1:] {
		if filepath.Base(filepath.Clean(part.value)) != part.value || part.value == "." || part.value == ".." {
			return "", fmt.Errorf("recovery evidence %s must be one normalized path segment", part.name)
		}
	}
	return filepath.Join(filepath.Clean(location.ResultsRoot), location.RunID, location.Target, location.Group), nil
}

func WriteEvidenceArtifact(t testing.TB, location EvidenceLocation, name string, body []byte) string {
	t.Helper()
	if filepath.Base(filepath.Clean(name)) != name || name == "." || name == ".." {
		t.Fatalf("Recovery evidence name must be one normalized path segment: %q", name)
	}
	dir, err := location.Dir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create Recovery evidence dir: %v", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatalf("secure Recovery evidence dir: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write Recovery evidence artifact: %v", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatalf("secure Recovery evidence artifact: %v", err)
	}
	return path
}

func RequireStoredArtifactProof(t testing.TB, storage recovery.BackupStorage, proof recovery.BackupArtifactProof) []byte {
	t.Helper()
	body, err := recovery.VerifyArtifactProof(context.Background(), storage, proof)
	if err != nil {
		t.Fatalf("verify stored artifact proof for %s: %v", proof.Key, err)
	}
	return body
}

func CaptureParams(params recovery.CaptureBackupSetParams) recovery.CaptureBackupSetParams {
	if params.PostgresRestoreAnchorRetainedUntil.IsZero() {
		params.PostgresRestoreAnchorRetainedUntil = params.RetainedUntil
	}
	if params.ObjectStoreRestoreAnchorRetainedUntil.IsZero() {
		params.ObjectStoreRestoreAnchorRetainedUntil = params.RetainedUntil
	}
	return params
}

type CaptureInput struct {
	Prefix                  string
	AsOf                    time.Time
	OlderBackupSetID        uuid.UUID
	OlderConsistencyPointAt time.Time
	BackupSetID             uuid.UUID
	ConsistencyPointAt      time.Time
	Postgres                *pgxpool.Pool
	ObjectStore             objectstore.Store
	ObjectStoreBucket       string
	Store                   *recovery.Store
	EvidenceObjects         recovery.EvidenceRecoveryProvider
	BackupStorage           recovery.BackupStorage
	BackupStorageRoot       string
	ExtensionCatalog        *recovery.ExtensionBackupCatalog
	EvidenceLocation        EvidenceLocation
	IncidentID              string
	TimelineRecordID        string
	TimelineRowVersion      int
	EvidenceRecordID        string
	ObjectBlobID            string
	BlobBody                []byte
}

func (input CaptureInput) Validate() error {
	if strings.TrimSpace(input.Prefix) == "" {
		return fmt.Errorf("recovery capture prefix is required")
	}
	if input.OlderBackupSetID == uuid.Nil || input.BackupSetID == uuid.Nil ||
		input.OlderBackupSetID == input.BackupSetID {
		return fmt.Errorf("recovery capture requires distinct non-zero backup identities")
	}
	if input.OlderConsistencyPointAt.IsZero() || input.ConsistencyPointAt.IsZero() {
		return fmt.Errorf("recovery capture consistency points are required")
	}
	if input.Postgres == nil || input.ObjectStore == nil || input.Store == nil || input.EvidenceObjects == nil ||
		input.BackupStorage == nil || input.ExtensionCatalog == nil {
		return fmt.Errorf("recovery capture requires borrowed database, object-store, storage, catalog, and repository dependencies")
	}
	if strings.TrimSpace(input.ObjectStoreBucket) == "" || strings.TrimSpace(input.BackupStorageRoot) == "" {
		return fmt.Errorf("recovery capture storage bindings are required")
	}
	if strings.TrimSpace(input.IncidentID) == "" || strings.TrimSpace(input.TimelineRecordID) == "" ||
		input.TimelineRowVersion < 1 || strings.TrimSpace(input.EvidenceRecordID) == "" ||
		strings.TrimSpace(input.ObjectBlobID) == "" || len(input.BlobBody) == 0 {
		return fmt.Errorf("recovery capture restored identities and expected facts are required")
	}
	if _, err := input.EvidenceLocation.Dir(); err != nil {
		return err
	}
	return nil
}

type CapturedSource struct {
	AsOf                    time.Time
	ConsistencyPointAt      time.Time
	BackupSetID             uuid.UUID
	IncidentID              string
	TimelineRecordID        string
	TimelineRowVersion      int
	EvidenceRecordID        string
	ObjectBlobID            string
	BlobSHA256              string
	AuthoritativeRowsSHA256 string
	ChangeSetsSHA256        string
	BlobHashesSHA256        string
	ChangeSetRowCount       int
	EvidenceCount           int
	HasEvidence             bool
	EvidenceBlob            EvidenceBlobConsistency
	ManifestEvidencePath    string
	SummaryEvidencePath     string
}

func CaptureSourceBackup(t testing.TB, ctx context.Context, input CaptureInput) CapturedSource {
	t.Helper()
	if err := input.Validate(); err != nil {
		t.Fatal(err)
	}
	asOf := input.AsOf.UTC().Truncate(time.Second)
	if input.AsOf.IsZero() {
		asOf = time.Now().UTC().Truncate(time.Second)
	}
	if !input.OlderConsistencyPointAt.Before(input.ConsistencyPointAt) ||
		!input.ConsistencyPointAt.Before(asOf) {
		t.Fatalf("Recovery capture consistency points must be ordered before as-of: older=%s selected=%s as_of=%s", input.OlderConsistencyPointAt, input.ConsistencyPointAt, asOf)
	}

	evidenceBlob := RequireEvidenceBlobConsistency(t, input.Postgres, input.EvidenceRecordID)
	if evidenceBlob.ObjectBlobID != input.ObjectBlobID {
		t.Fatalf("source Evidence object_blob_id got %s want %s", evidenceBlob.ObjectBlobID, input.ObjectBlobID)
	}
	blobSum := sha256.Sum256(input.BlobBody)
	blobSHA256 := hex.EncodeToString(blobSum[:])
	if evidenceBlob.BlobSHA256 != blobSHA256 || evidenceBlob.ObservedSHA256 != blobSHA256 {
		t.Fatalf("source Evidence/blob sha256 got %#v want %s", evidenceBlob, blobSHA256)
	}
	changeSetRowCount := RequireChangeSetRowCount(t, input.Postgres)

	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, input.Postgres)
	if err != nil {
		t.Fatalf("capture source postgres restore artifact: %v", err)
	}
	postgresSnapshot, err := recovery.DecodePostgresSnapshotArtifact(postgresArtifact)
	if err != nil {
		t.Fatalf("decode source postgres restore artifact: %v", err)
	}
	authoritativeRowsSHA256, _, err := SnapshotDigest(postgresSnapshot, nil)
	if err != nil {
		t.Fatalf("digest source authoritative rows: %v", err)
	}
	changeSetsSHA256, digestChangeSetRowCount, err := SnapshotDigest(postgresSnapshot, IsChangeSetTable)
	if err != nil {
		t.Fatalf("digest source change-set rows: %v", err)
	}
	if digestChangeSetRowCount != changeSetRowCount {
		t.Fatalf("source change-set digest count got %d want %d", digestChangeSetRowCount, changeSetRowCount)
	}
	blobHashesSHA256 := BlobConsistencyDigest(evidenceBlob.StorageKey, blobSHA256)
	blobIndex, err := recovery.AvailableBlobObjectIDsByStorageRef(ctx, input.EvidenceObjects)
	if err != nil {
		t.Fatalf("index source blob storage refs: %v", err)
	}

	olderArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, input.ObjectStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               input.OlderBackupSetID,
		ConsistencyPointAt:        input.OlderConsistencyPointAt,
		Bucket:                    input.ObjectStoreBucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		t.Fatalf("capture older object-store backup artifacts: %v", err)
	}
	selectedArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, input.ObjectStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               input.BackupSetID,
		ConsistencyPointAt:        input.ConsistencyPointAt,
		Bucket:                    input.ObjectStoreBucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		t.Fatalf("capture selected object-store backup artifacts: %v", err)
	}
	if !bytes.Contains(selectedArtifacts.SnapshotBody, []byte("body_base64")) {
		t.Fatal("object-store restore artifact does not include restorable object bodies")
	}

	capture := recovery.NewCaptureService(input.Store, input.BackupStorage, input.ExtensionCatalog)
	olderCreatedAt := asOf.Add(-10 * time.Minute)
	selectedCreatedAt := asOf.Add(-2 * time.Minute)
	if _, err := capture.CaptureBackupSet(ctx, CaptureParams(recovery.CaptureBackupSetParams{
		BackupSetID:                       input.OlderBackupSetID,
		ConsistencyPointAt:                input.OlderConsistencyPointAt,
		CreatedAt:                         olderCreatedAt,
		RetainedUntil:                     olderCreatedAt.Add(31 * 24 * time.Hour),
		PostgresArtifact:                  recovery.BackupArtifact{Body: postgresArtifact, ContentType: "application/json"},
		ObjectStoreArtifact:               recovery.BackupArtifact{Body: olderArtifacts.SnapshotBody, ContentType: "application/json"},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: olderArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: olderArtifacts.SummaryBody, ContentType: "application/json"},
	})); err != nil {
		t.Fatalf("capture older retained backup set: %v", err)
	}
	if _, err := capture.CaptureBackupSet(ctx, CaptureParams(recovery.CaptureBackupSetParams{
		BackupSetID:                       input.BackupSetID,
		ConsistencyPointAt:                input.ConsistencyPointAt,
		CreatedAt:                         selectedCreatedAt,
		RetainedUntil:                     selectedCreatedAt.Add(31 * 24 * time.Hour),
		PostgresArtifact:                  recovery.BackupArtifact{Body: postgresArtifact, ContentType: "application/json"},
		ObjectStoreArtifact:               recovery.BackupArtifact{Body: selectedArtifacts.SnapshotBody, ContentType: "application/json"},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: selectedArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: selectedArtifacts.SummaryBody, ContentType: "application/json"},
	})); err != nil {
		t.Fatalf("capture selected retained backup set: %v", err)
	}
	rawPostgresArtifact, err := os.ReadFile(filepath.Join(input.BackupStorageRoot, filepath.FromSlash("backup_sets/"+input.BackupSetID.String()+"/postgres-artifact.json")))
	if err != nil {
		t.Fatalf("read raw encrypted process backup artifact: %v", err)
	}
	if bytes.Contains(rawPostgresArtifact, []byte(input.IncidentID)) ||
		bytes.Contains(rawPostgresArtifact, input.BlobBody) {
		t.Fatal("raw process backup artifact contains plaintext source data")
	}

	manifestPath := WriteEvidenceArtifact(t, input.EvidenceLocation, "object-store-backup-manifest.json", selectedArtifacts.ManifestBody)
	summaryPath := WriteEvidenceArtifact(t, input.EvidenceLocation, "object-store-backup-summary.json", selectedArtifacts.SummaryBody)
	manifest, err := recovery.DecodeObjectStoreBackupManifestArtifact(selectedArtifacts.ManifestBody)
	if err != nil {
		t.Fatalf("decode retained object-store backup manifest: %v", err)
	}
	if manifest.BackupSetID != input.BackupSetID.String() ||
		!manifest.ConsistencyPointAt.Equal(input.ConsistencyPointAt) ||
		manifest.ObjectStoreBackend != recovery.ObjectStoreBackendSeaweedFSS3 || manifest.ObjectCount == 0 {
		t.Fatalf("object-store backup manifest does not satisfy backup predicate at %s: %#v", manifestPath, manifest)
	}
	for _, object := range manifest.Objects {
		if object.SHA256 == "" || object.BackupMemberSHA256 == "" {
			t.Fatalf("manifest object has missing sha256 proof at %s: %#v", manifestPath, object)
		}
	}
	manifestObject := RequireManifestObject(t, manifest, input.ObjectBlobID)
	if manifestObject.SHA256 != blobSHA256 || manifestObject.StorageRef != evidenceBlob.StorageKey {
		t.Fatalf("source manifest object got %#v want sha=%s storage_key=%s", manifestObject, blobSHA256, evidenceBlob.StorageKey)
	}
	if _, err := recovery.DecodeObjectStoreBackupSummaryArtifact(selectedArtifacts.SummaryBody); err != nil {
		t.Fatalf("decode retained object-store backup summary: %v", err)
	}
	if bytes.Contains(selectedArtifacts.SummaryBody, []byte(input.ObjectStoreBucket)) {
		t.Fatalf("shareable object-store backup summary leaked raw bucket name at %s", summaryPath)
	}
	for _, object := range manifest.Objects {
		if bytes.Contains(selectedArtifacts.SummaryBody, []byte(object.StorageRef)) {
			t.Fatalf("shareable object-store backup summary leaked raw storage ref at %s", summaryPath)
		}
	}

	return CapturedSource{
		AsOf:                    asOf,
		ConsistencyPointAt:      input.ConsistencyPointAt.UTC(),
		BackupSetID:             input.BackupSetID,
		IncidentID:              input.IncidentID,
		TimelineRecordID:        input.TimelineRecordID,
		TimelineRowVersion:      input.TimelineRowVersion,
		EvidenceRecordID:        input.EvidenceRecordID,
		ObjectBlobID:            input.ObjectBlobID,
		BlobSHA256:              blobSHA256,
		AuthoritativeRowsSHA256: authoritativeRowsSHA256,
		ChangeSetsSHA256:        changeSetsSHA256,
		BlobHashesSHA256:        blobHashesSHA256,
		ChangeSetRowCount:       changeSetRowCount,
		EvidenceCount:           1,
		HasEvidence:             true,
		EvidenceBlob:            evidenceBlob,
		ManifestEvidencePath:    manifestPath,
		SummaryEvidencePath:     summaryPath,
	}
}

type EvidenceBlobConsistency struct {
	ObjectBlobID   string
	LifecycleState string
	UploadState    string
	BlobSHA256     string
	ObservedSHA256 string
	StorageKey     string
}

func RequireChangeSetRowCount(t testing.TB, pool *pgxpool.Pool) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), `
SELECT (SELECT COUNT(*) FROM change_sets)
     + (SELECT COUNT(*) FROM change_set_mutations)
     + (SELECT COUNT(*) FROM record_history_entry_refs)
     + (SELECT COUNT(*) FROM record_revision_conflict_facts)
     + (SELECT COUNT(*) FROM record_revisions)
`).Scan(&count); err != nil {
		t.Fatalf("count change-set rows: %v", err)
	}
	return count
}

func RequireEvidenceBlobConsistency(t testing.TB, pool *pgxpool.Pool, evidenceRecordID string) EvidenceBlobConsistency {
	t.Helper()
	var facts EvidenceBlobConsistency
	if err := pool.QueryRow(context.Background(), `
SELECT e.object_blob_id::text,
       e.lifecycle_state,
       b.upload_state,
       COALESCE(e.blob_hash, ''),
       COALESCE(b.observed_sha256_hex, ''),
       b.storage_key
  FROM evidence e
  JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
`, evidenceRecordID).Scan(
		&facts.ObjectBlobID,
		&facts.LifecycleState,
		&facts.UploadState,
		&facts.BlobSHA256,
		&facts.ObservedSHA256,
		&facts.StorageKey,
	); err != nil {
		t.Fatalf("read Evidence/blob consistency: %v", err)
	}
	return facts
}

func RequireStoredObjectSHA256(t testing.TB, store objectstore.Store, storageKey string) string {
	t.Helper()
	reader, _, err := store.ReadObject(context.Background(), storageKey, objectstore.ReadOptions{})
	if err != nil {
		t.Fatalf("read restored object: %v", err)
	}
	defer reader.Close()
	digest := sha256.New()
	if _, err := io.Copy(digest, reader); err != nil {
		t.Fatalf("hash restored object: %v", err)
	}
	return hex.EncodeToString(digest.Sum(nil))
}

func RequireManifestObject(t testing.TB, manifest recovery.ObjectStoreBackupManifest, objectBlobID string) recovery.ObjectStoreBackupManifestObject {
	t.Helper()
	for _, object := range manifest.Objects {
		if object.ObjectBlobID == objectBlobID {
			return object
		}
	}
	t.Fatalf("manifest does not contain object_blob_id %s: %#v", objectBlobID, manifest.Objects)
	return recovery.ObjectStoreBackupManifestObject{}
}

func SnapshotDigest(artifact recovery.PostgresSnapshotArtifact, include func(string) bool) (string, int, error) {
	digest := sha256.New()
	rowCount := 0
	tables := append([]recovery.PostgresSnapshotTable(nil), artifact.Tables...)
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
			var value any
			if err := json.Unmarshal(rawRow, &value); err != nil {
				return "", 0, err
			}
			normalized, err := json.Marshal(value)
			if err != nil {
				return "", 0, err
			}
			rows = append(rows, string(normalized))
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

func IsChangeSetTable(tableName string) bool {
	switch tableName {
	case "change_sets", "change_set_mutations", "record_history_entry_refs", "record_revision_conflict_facts", "record_revisions":
		return true
	default:
		return false
	}
}

func BlobConsistencyDigest(storageKey string, objectSHA256 string) string {
	rowDigest := sha256.Sum256([]byte(storageKey + ":" + objectSHA256 + "\n"))
	digest := sha256.New()
	_, _ = digest.Write([]byte("object:" + storageKey + ":" + objectSHA256 + "\n"))
	_, _ = digest.Write([]byte("blob_rows:" + hex.EncodeToString(rowDigest[:]) + "\n"))
	return hex.EncodeToString(digest.Sum(nil))
}

type RestoreExpectation struct {
	BackupSetID             uuid.UUID
	ConsistencyPointAt      time.Time
	AuthoritativeRowsSHA256 string
	ChangeSetsSHA256        string
	BlobHashesSHA256        string
	ChangeSetRowCount       int
	BlobCount               int
	EvidenceRecordID        string
	EvidenceBlob            EvidenceBlobConsistency
	BlobSHA256              string
}

type restoreGate struct {
	marked bool
	result recovery.RestoreResult
	steps  []recovery.RestoreStep
}

func (gate *restoreGate) MarkRestoreReady(_ context.Context, result recovery.RestoreResult) error {
	gate.marked = true
	gate.result = result
	return nil
}

func (gate *restoreGate) RecordRestoreStep(step recovery.RestoreStep) {
	gate.steps = append(gate.steps, step)
}

func RestoreLatest(t testing.TB, ctx context.Context, runner *recovery.RestoreRunner, postgres *pgxpool.Pool, target recovery.RestoreTarget, asOf time.Time, expected RestoreExpectation) recovery.RestoreResult {
	t.Helper()
	gate := &restoreGate{}
	target.Readiness = gate
	target.Observer = gate
	result, err := runner.RestoreLatestSuccessfulRetained(ctx, target, asOf)
	if err != nil {
		t.Fatalf("restore latest retained backup into fresh environment: %v", err)
	}
	wantSteps := []recovery.RestoreStep{
		recovery.RestoreStepPostgresRestore,
		recovery.RestoreStepObjectStoreRestore,
		recovery.RestoreStepExtensionBindings,
		recovery.RestoreStepProjectionRebuild,
		recovery.RestoreStepConsistencyCheck,
		recovery.RestoreStepReadiness,
	}
	if !gate.marked || !slices.Equal(gate.steps, wantSteps) ||
		gate.result.BackupSet.BackupSetID != expected.BackupSetID ||
		!gate.result.ProjectionRebuildResult.ReadinessSatisfied() {
		t.Fatalf("restore readiness contract got marked=%v steps=%v result=%#v", gate.marked, gate.steps, gate.result)
	}
	if result.BackupSet.BackupSetID != expected.BackupSetID ||
		!result.BackupSet.ConsistencyPointAt.Equal(expected.ConsistencyPointAt) {
		t.Fatalf("restore selected backup got id=%s point=%s want id=%s point=%s", result.BackupSet.BackupSetID, result.BackupSet.ConsistencyPointAt, expected.BackupSetID, expected.ConsistencyPointAt)
	}
	if result.ConsistencyReport.AuthoritativeRowsSHA256 != expected.AuthoritativeRowsSHA256 ||
		result.ConsistencyReport.ChangeSetsSHA256 != expected.ChangeSetsSHA256 ||
		result.ConsistencyReport.BlobHashesSHA256 != expected.BlobHashesSHA256 ||
		result.ConsistencyReport.ChangeSetRowCount != expected.ChangeSetRowCount ||
		result.ConsistencyReport.BlobCount != expected.BlobCount {
		t.Fatalf("restored consistency got %#v want %#v", result.ConsistencyReport, expected)
	}
	if postgres == nil {
		t.Fatal("restored Postgres observation pool is required")
	}
	if got := RequireChangeSetRowCount(t, postgres); got != expected.ChangeSetRowCount {
		t.Fatalf("restored change-set row count got %d want %d", got, expected.ChangeSetRowCount)
	}
	restoredEvidence := RequireEvidenceBlobConsistency(t, postgres, expected.EvidenceRecordID)
	if restoredEvidence != expected.EvidenceBlob {
		t.Fatalf("restored Evidence/blob facts got %#v want %#v", restoredEvidence, expected.EvidenceBlob)
	}
	if got := RequireStoredObjectSHA256(t, target.ObjectStore, restoredEvidence.StorageKey); got != expected.BlobSHA256 {
		t.Fatalf("restored object sha256 got %s want %s", got, expected.BlobSHA256)
	}
	manifestObject := RequireManifestObject(t, result.ObjectStoreBackupManifest, expected.EvidenceBlob.ObjectBlobID)
	if manifestObject.SHA256 != expected.BlobSHA256 || manifestObject.StorageRef != expected.EvidenceBlob.StorageKey {
		t.Fatalf("restored manifest object got %#v want sha=%s storage_key=%s", manifestObject, expected.BlobSHA256, expected.EvidenceBlob.StorageKey)
	}
	return result
}

type VerificationExpectation struct {
	BackupSetID        uuid.UUID
	ConsistencyPointAt time.Time
	IncidentID         string
	ObjectCount        int64
	RegistrationID     string
	ViewSchemaID       string
}

func RequireVerificationArtifact(t testing.TB, storage recovery.BackupStorage, result recovery.RestoreVerificationResult, location EvidenceLocation, expected VerificationExpectation) string {
	t.Helper()
	body := RequireStoredArtifactProof(t, storage, result.ArtifactProof)
	artifact, err := recovery.DecodeRestoreVerificationArtifact(body)
	if err != nil {
		t.Fatalf("decode restore verification artifact: %v", err)
	}
	if artifact.BackupSetID != expected.BackupSetID.String() ||
		!artifact.ConsistencyPointAt.Equal(expected.ConsistencyPointAt) ||
		artifact.SelectedIncidentID == nil || *artifact.SelectedIncidentID != expected.IncidentID ||
		artifact.WorkbookProbe.Status != "executed" ||
		artifact.WorkbookProbe.RegistrationID != expected.RegistrationID ||
		artifact.WorkbookProbe.ViewSchemaID != expected.ViewSchemaID ||
		artifact.RestoredObjectCount != expected.ObjectCount || artifact.Result != "pass" {
		t.Fatalf("restore verification artifact got %#v want %#v", artifact, expected)
	}
	return WriteEvidenceArtifact(t, location, "restore-verification.json", body)
}
