package recovery

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

const (
	ObjectStoreBackupManifestSchemaID = "cartulary.object_store_backup_manifest.v1"
	ObjectStoreBackupSummarySchemaID  = "cartulary.object_store_backup_summary.v1"
	ObjectStoreBackendSeaweedFSS3     = "seaweedfs_s3"

	restoreVerificationArtifactV1SchemaID = "cartulary.restore_verification.v1"
	historicalTimelineViewSchemaID        = "cartulary.view.timeline.v2"
)

type ObjectStoreBackupCaptureParams struct {
	BackupSetID               uuid.UUID
	ConsistencyPointAt        time.Time
	Bucket                    string
	Prefix                    string
	BlobObjectIDsByStorageRef map[string]uuid.UUID
}

type ObjectStoreBackupCaptureArtifacts struct {
	SnapshotBody []byte
	ManifestBody []byte
	SummaryBody  []byte
	Manifest     ObjectStoreBackupManifest
	Summary      ObjectStoreBackupSummary
}

type ObjectStoreBackupManifestParams struct {
	BackupSetID               uuid.UUID
	ConsistencyPointAt        time.Time
	Bucket                    string
	BlobObjectIDsByStorageRef map[string]uuid.UUID
}

type ObjectStoreBackupManifest struct {
	SchemaID           string                            `json:"schema_id"`
	BackupSetID        string                            `json:"backup_set_id"`
	ConsistencyPointAt time.Time                         `json:"consistency_point_at"`
	ObjectStoreBackend string                            `json:"object_store_backend"`
	Bucket             string                            `json:"bucket"`
	ObjectCount        int                               `json:"object_count"`
	TotalSizeBytes     int64                             `json:"total_size_bytes"`
	Objects            []ObjectStoreBackupManifestObject `json:"objects"`
	ManifestSHA256     string                            `json:"manifest_sha256"`
}

type ObjectStoreBackupManifestObject struct {
	ObjectBlobID       string `json:"object_blob_id,omitempty"`
	StorageRef         string `json:"storage_ref"`
	StorageRefSHA256   string `json:"storage_ref_sha256"`
	SizeBytes          int64  `json:"size_bytes"`
	SHA256             string `json:"sha256"`
	BackupMemberSHA256 string `json:"backup_member_sha256"`
}

type ObjectStoreBackupSummary struct {
	SchemaID           string                           `json:"schema_id"`
	BackupSetID        string                           `json:"backup_set_id"`
	ConsistencyPointAt time.Time                        `json:"consistency_point_at"`
	ObjectStoreBackend string                           `json:"object_store_backend"`
	BucketRef          RedactionRef                     `json:"bucket_ref"`
	ObjectCount        int                              `json:"object_count"`
	TotalSizeBytes     int64                            `json:"total_size_bytes"`
	ManifestSHA256     string                           `json:"manifest_sha256"`
	ObjectsSummary     []ObjectStoreBackupSummaryObject `json:"objects_summary,omitempty"`
}

type ObjectStoreBackupSummaryObject struct {
	StorageRefSHA256 string `json:"storage_ref_sha256"`
	SizeBytes        int64  `json:"size_bytes"`
	SHA256           string `json:"sha256"`
}

type RedactionRef struct {
	Redacted       bool   `json:"redacted"`
	RedactionClass string `json:"redaction_class"`
	SHA256         string `json:"sha256,omitempty"`
	Scheme         string `json:"scheme,omitempty"`
	HostSHA256     string `json:"host_sha256,omitempty"`
	PortPresent    *bool  `json:"port_present,omitempty"`
}

type RestoreVerificationArtifactV1 struct {
	SchemaID                string                               `json:"schema_id"`
	BackupSetID             string                               `json:"backup_set_id"`
	SelectedIncidentID      *string                              `json:"selected_incident_id"`
	IncidentOpenCheck       RestoreVerificationIncidentOpenCheck `json:"incident_open_check"`
	QueryViewSchemaID       string                               `json:"query_view_schema_id"`
	BlobCheckCounts         RestoreVerificationBlobCheckCounts   `json:"blob_check_counts"`
	ManifestCheckResult     string                               `json:"manifest_check_result"`
	ProjectionRebuildResult string                               `json:"projection_rebuild_result"`
	Result                  string                               `json:"result"`
	FailureReasons          []string                             `json:"failure_reasons"`
	ArtifactSHA256          string                               `json:"artifact_sha256"`
}

type RestoreVerificationIncidentOpenCheck struct {
	Status string `json:"status"`
}

type RestoreVerificationBlobCheckCounts struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

func CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx context.Context, store objectstore.Store, params ObjectStoreBackupCaptureParams) (ObjectStoreBackupCaptureArtifacts, error) {
	if store == nil {
		return ObjectStoreBackupCaptureArtifacts{}, fmt.Errorf("%w: object store is required", ErrInvalidBackupArtifact)
	}
	if params.BackupSetID == uuid.Nil {
		return ObjectStoreBackupCaptureArtifacts{}, fmt.Errorf("%w: backup_set_id is required for object-store backup manifest", ErrInvalidBackupArtifact)
	}
	if params.ConsistencyPointAt.IsZero() {
		return ObjectStoreBackupCaptureArtifacts{}, fmt.Errorf("%w: consistency_point_at is required for object-store backup manifest", ErrInvalidBackupArtifact)
	}
	if strings.TrimSpace(params.Bucket) == "" {
		return ObjectStoreBackupCaptureArtifacts{}, fmt.Errorf("%w: bucket is required for object-store backup manifest", ErrInvalidBackupArtifact)
	}
	objects, err := listBackupManifestObjects(ctx, store, params.Prefix)
	if err != nil {
		return ObjectStoreBackupCaptureArtifacts{}, err
	}
	items := make([]ObjectStoreSnapshotItem, 0, len(objects))
	for _, object := range objects {
		reader, info, err := getBackupManifestObject(ctx, store, object.Key)
		if err != nil {
			return ObjectStoreBackupCaptureArtifacts{}, err
		}
		body, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil {
			return ObjectStoreBackupCaptureArtifacts{}, fmt.Errorf("read object store backup object body %s: %w", object.Key, readErr)
		}
		if closeErr != nil {
			return ObjectStoreBackupCaptureArtifacts{}, fmt.Errorf("close object store backup object %s: %w", object.Key, closeErr)
		}
		if int64(len(body)) != info.Size {
			return ObjectStoreBackupCaptureArtifacts{}, fmt.Errorf("%w: object store backup size mismatch for %s", ErrInvalidBackupArtifact, object.Key)
		}
		items = append(items, ObjectStoreSnapshotItem{
			Key:         info.Key,
			SizeBytes:   info.Size,
			ContentType: info.ContentType,
			SHA256:      sha256Hex(body),
			BodyBase64:  base64.StdEncoding.EncodeToString(body),
		})
	}
	snapshot := ObjectStoreSnapshotArtifact{
		SchemaID: ObjectStoreSnapshotArtifactSchemaID,
		Objects:  items,
	}
	snapshotBody, err := json.Marshal(snapshot)
	if err != nil {
		return ObjectStoreBackupCaptureArtifacts{}, fmt.Errorf("encode object-store snapshot artifact: %w", err)
	}
	manifest, manifestBody, err := BuildSeaweedFSS3ObjectStoreBackupManifest(snapshot, ObjectStoreBackupManifestParams{
		BackupSetID:               params.BackupSetID,
		ConsistencyPointAt:        params.ConsistencyPointAt,
		Bucket:                    params.Bucket,
		BlobObjectIDsByStorageRef: params.BlobObjectIDsByStorageRef,
	})
	if err != nil {
		return ObjectStoreBackupCaptureArtifacts{}, err
	}
	summary, summaryBody, err := BuildObjectStoreBackupSummary(manifest)
	if err != nil {
		return ObjectStoreBackupCaptureArtifacts{}, err
	}
	return ObjectStoreBackupCaptureArtifacts{
		SnapshotBody: snapshotBody,
		ManifestBody: manifestBody,
		SummaryBody:  summaryBody,
		Manifest:     manifest,
		Summary:      summary,
	}, nil
}

func BuildSeaweedFSS3ObjectStoreBackupManifest(snapshot ObjectStoreSnapshotArtifact, params ObjectStoreBackupManifestParams) (ObjectStoreBackupManifest, []byte, error) {
	if params.BackupSetID == uuid.Nil {
		return ObjectStoreBackupManifest{}, nil, fmt.Errorf("%w: backup_set_id is required for object-store backup manifest", ErrInvalidBackupArtifact)
	}
	if params.ConsistencyPointAt.IsZero() {
		return ObjectStoreBackupManifest{}, nil, fmt.Errorf("%w: consistency_point_at is required for object-store backup manifest", ErrInvalidBackupArtifact)
	}
	bucket := strings.TrimSpace(params.Bucket)
	if bucket == "" {
		return ObjectStoreBackupManifest{}, nil, fmt.Errorf("%w: bucket is required for object-store backup manifest", ErrInvalidBackupArtifact)
	}
	if snapshot.SchemaID != ObjectStoreSnapshotArtifactSchemaID {
		return ObjectStoreBackupManifest{}, nil, fmt.Errorf("%w: unsupported object-store snapshot schema %q", ErrInvalidBackupArtifact, snapshot.SchemaID)
	}
	objects := append([]ObjectStoreSnapshotItem(nil), snapshot.Objects...)
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Key < objects[j].Key
	})
	manifestObjects := make([]ObjectStoreBackupManifestObject, 0, len(objects))
	totalSize := int64(0)
	for _, object := range objects {
		if strings.TrimSpace(object.Key) == "" {
			return ObjectStoreBackupManifest{}, nil, fmt.Errorf("%w: object-store backup manifest storage_ref is required", ErrInvalidBackupArtifact)
		}
		if object.SizeBytes < 0 {
			return ObjectStoreBackupManifest{}, nil, fmt.Errorf("%w: object-store backup manifest size_bytes is negative for %s", ErrInvalidBackupArtifact, object.Key)
		}
		if !validSHA256Hex(object.SHA256) {
			return ObjectStoreBackupManifest{}, nil, fmt.Errorf("%w: object-store backup manifest object sha256 is required for %s", ErrInvalidBackupArtifact, object.Key)
		}
		entry := ObjectStoreBackupManifestObject{
			StorageRef:         object.Key,
			StorageRefSHA256:   sha256Hex([]byte(object.Key)),
			SizeBytes:          object.SizeBytes,
			SHA256:             object.SHA256,
			BackupMemberSHA256: object.SHA256,
		}
		if objectBlobID, ok := params.BlobObjectIDsByStorageRef[object.Key]; ok && objectBlobID != uuid.Nil {
			entry.ObjectBlobID = objectBlobID.String()
		}
		manifestObjects = append(manifestObjects, entry)
		totalSize += object.SizeBytes
	}
	manifest := ObjectStoreBackupManifest{
		SchemaID:           ObjectStoreBackupManifestSchemaID,
		BackupSetID:        params.BackupSetID.String(),
		ConsistencyPointAt: backupTimestamp(params.ConsistencyPointAt),
		ObjectStoreBackend: ObjectStoreBackendSeaweedFSS3,
		Bucket:             bucket,
		ObjectCount:        len(manifestObjects),
		TotalSizeBytes:     totalSize,
		Objects:            manifestObjects,
	}
	body, err := encodeObjectStoreBackupManifest(manifest)
	if err != nil {
		return ObjectStoreBackupManifest{}, nil, err
	}
	manifest.ManifestSHA256 = sha256Hex(canonicalObjectStoreBackupManifestBytes(manifest, false))
	return manifest, body, nil
}

func BuildObjectStoreBackupSummary(manifest ObjectStoreBackupManifest) (ObjectStoreBackupSummary, []byte, error) {
	if err := ValidateObjectStoreBackupManifest(manifest); err != nil {
		return ObjectStoreBackupSummary{}, nil, err
	}
	objects := make([]ObjectStoreBackupSummaryObject, 0, len(manifest.Objects))
	for _, object := range manifest.Objects {
		objects = append(objects, ObjectStoreBackupSummaryObject{
			StorageRefSHA256: object.StorageRefSHA256,
			SizeBytes:        object.SizeBytes,
			SHA256:           object.SHA256,
		})
	}
	summary := ObjectStoreBackupSummary{
		SchemaID:           ObjectStoreBackupSummarySchemaID,
		BackupSetID:        manifest.BackupSetID,
		ConsistencyPointAt: manifest.ConsistencyPointAt,
		ObjectStoreBackend: manifest.ObjectStoreBackend,
		BucketRef:          hashRedactionRef("bucket", manifest.Bucket),
		ObjectCount:        manifest.ObjectCount,
		TotalSizeBytes:     manifest.TotalSizeBytes,
		ManifestSHA256:     manifest.ManifestSHA256,
		ObjectsSummary:     objects,
	}
	body, err := encodeObjectStoreBackupSummary(summary)
	if err != nil {
		return ObjectStoreBackupSummary{}, nil, err
	}
	return summary, body, nil
}

func encodeObjectStoreBackupManifest(manifest ObjectStoreBackupManifest) ([]byte, error) {
	if err := ValidateObjectStoreBackupManifestWithoutDigest(manifest); err != nil {
		return nil, err
	}
	manifest.ManifestSHA256 = sha256Hex(canonicalObjectStoreBackupManifestBytes(manifest, false))
	if err := ValidateObjectStoreBackupManifest(manifest); err != nil {
		return nil, err
	}
	return canonicalObjectStoreBackupManifestBytes(manifest, true), nil
}

func DecodeObjectStoreBackupManifestArtifact(body []byte) (ObjectStoreBackupManifest, error) {
	if err := rejectDuplicateJSONKeys(body); err != nil {
		return ObjectStoreBackupManifest{}, fmt.Errorf("%w: object-store backup manifest JSON object keys must be unique: %v", ErrInvalidBackupArtifact, err)
	}
	var manifest ObjectStoreBackupManifest
	if err := decodeStrictJSON(body, &manifest); err != nil {
		return ObjectStoreBackupManifest{}, fmt.Errorf("%w: decode object-store backup manifest: %v", ErrInvalidBackupArtifact, err)
	}
	if err := ValidateObjectStoreBackupManifest(manifest); err != nil {
		return ObjectStoreBackupManifest{}, err
	}
	canonical := canonicalObjectStoreBackupManifestBytes(manifest, true)
	if !bytes.Equal(body, canonical) {
		return ObjectStoreBackupManifest{}, fmt.Errorf("%w: object-store backup manifest is not canonical JSON", ErrInvalidBackupArtifact)
	}
	return manifest, nil
}

func ValidateObjectStoreBackupManifest(manifest ObjectStoreBackupManifest) error {
	if err := ValidateObjectStoreBackupManifestWithoutDigest(manifest); err != nil {
		return err
	}
	if !validSHA256Hex(manifest.ManifestSHA256) {
		return fmt.Errorf("%w: object-store backup manifest_sha256 is required", ErrInvalidBackupArtifact)
	}
	if got := sha256Hex(canonicalObjectStoreBackupManifestBytes(manifest, false)); got != manifest.ManifestSHA256 {
		return fmt.Errorf("%w: object-store backup manifest_sha256 mismatch", ErrInvalidBackupArtifact)
	}
	return nil
}

func ValidateObjectStoreBackupManifestWithoutDigest(manifest ObjectStoreBackupManifest) error {
	if manifest.SchemaID != ObjectStoreBackupManifestSchemaID {
		return fmt.Errorf("%w: unsupported object-store backup manifest schema %q", ErrInvalidBackupArtifact, manifest.SchemaID)
	}
	if _, err := uuid.Parse(manifest.BackupSetID); err != nil {
		return fmt.Errorf("%w: object-store backup manifest backup_set_id must be a UUID", ErrInvalidBackupArtifact)
	}
	if manifest.ConsistencyPointAt.IsZero() {
		return fmt.Errorf("%w: object-store backup manifest consistency_point_at is required", ErrInvalidBackupArtifact)
	}
	if manifest.ObjectStoreBackend != ObjectStoreBackendSeaweedFSS3 {
		return fmt.Errorf("%w: object-store backup manifest backend must be %s", ErrInvalidBackupArtifact, ObjectStoreBackendSeaweedFSS3)
	}
	if strings.TrimSpace(manifest.Bucket) == "" {
		return fmt.Errorf("%w: object-store backup manifest bucket is required", ErrInvalidBackupArtifact)
	}
	if manifest.ObjectCount != len(manifest.Objects) {
		return fmt.Errorf("%w: object-store backup manifest object_count mismatch", ErrInvalidBackupArtifact)
	}
	totalSize := int64(0)
	previousRef := ""
	seen := make(map[string]struct{}, len(manifest.Objects))
	for index, object := range manifest.Objects {
		if strings.TrimSpace(object.StorageRef) == "" {
			return fmt.Errorf("%w: object-store backup manifest storage_ref is required", ErrInvalidBackupArtifact)
		}
		if index > 0 && previousRef >= object.StorageRef {
			return fmt.Errorf("%w: object-store backup manifest objects are not sorted by storage_ref", ErrInvalidBackupArtifact)
		}
		previousRef = object.StorageRef
		if _, ok := seen[object.StorageRef]; ok {
			return fmt.Errorf("%w: object-store backup manifest contains duplicate storage_ref", ErrInvalidBackupArtifact)
		}
		seen[object.StorageRef] = struct{}{}
		if object.StorageRefSHA256 != sha256Hex([]byte(object.StorageRef)) {
			return fmt.Errorf("%w: object-store backup manifest storage_ref_sha256 mismatch", ErrInvalidBackupArtifact)
		}
		if object.SizeBytes < 0 {
			return fmt.Errorf("%w: object-store backup manifest size_bytes is negative", ErrInvalidBackupArtifact)
		}
		if !validSHA256Hex(object.SHA256) {
			return fmt.Errorf("%w: object-store backup manifest object sha256 is required", ErrInvalidBackupArtifact)
		}
		if !validSHA256Hex(object.BackupMemberSHA256) {
			return fmt.Errorf("%w: object-store backup manifest backup_member_sha256 is required", ErrInvalidBackupArtifact)
		}
		if strings.TrimSpace(object.ObjectBlobID) != "" {
			if _, err := uuid.Parse(object.ObjectBlobID); err != nil {
				return fmt.Errorf("%w: object-store backup manifest object_blob_id must be a UUID", ErrInvalidBackupArtifact)
			}
		}
		totalSize += object.SizeBytes
	}
	if totalSize != manifest.TotalSizeBytes {
		return fmt.Errorf("%w: object-store backup manifest total_size_bytes mismatch", ErrInvalidBackupArtifact)
	}
	return nil
}

func ValidateObjectStoreBackupManifestForBackup(backupSet BackupSet, manifest ObjectStoreBackupManifest) error {
	if err := ValidateObjectStoreBackupManifest(manifest); err != nil {
		return err
	}
	if manifest.BackupSetID != backupSet.BackupSetID.String() {
		return fmt.Errorf("%w: object-store backup manifest backup_set_id does not match selected backup", ErrInvalidBackupArtifact)
	}
	if !manifest.ConsistencyPointAt.Equal(backupSet.ConsistencyPointAt) {
		return fmt.Errorf("%w: object-store backup manifest consistency_point_at does not match selected backup", ErrInvalidBackupArtifact)
	}
	return nil
}

func ValidateObjectStoreManifestAgainstSnapshot(manifest ObjectStoreBackupManifest, snapshot ObjectStoreSnapshotArtifact) error {
	if err := ValidateObjectStoreBackupManifest(manifest); err != nil {
		return err
	}
	if snapshot.SchemaID != ObjectStoreSnapshotArtifactSchemaID {
		return fmt.Errorf("%w: unsupported object-store snapshot schema %q", ErrInvalidBackupArtifact, snapshot.SchemaID)
	}
	objects := append([]ObjectStoreSnapshotItem(nil), snapshot.Objects...)
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Key < objects[j].Key
	})
	if len(objects) != len(manifest.Objects) {
		return fmt.Errorf("%w: object-store backup manifest object count does not match snapshot", ErrInvalidBackupArtifact)
	}
	totalSize := int64(0)
	for index, item := range objects {
		entry := manifest.Objects[index]
		if entry.StorageRef != item.Key {
			return fmt.Errorf("%w: object-store backup manifest storage_ref does not match snapshot", ErrInvalidBackupArtifact)
		}
		if entry.SizeBytes != item.SizeBytes || entry.SHA256 != item.SHA256 || entry.BackupMemberSHA256 != item.SHA256 {
			return fmt.Errorf("%w: object-store backup manifest object proof does not match snapshot", ErrInvalidBackupArtifact)
		}
		totalSize += item.SizeBytes
	}
	if totalSize != manifest.TotalSizeBytes {
		return fmt.Errorf("%w: object-store backup manifest total size does not match snapshot", ErrInvalidBackupArtifact)
	}
	return nil
}

func encodeObjectStoreBackupSummary(summary ObjectStoreBackupSummary) ([]byte, error) {
	if err := ValidateObjectStoreBackupSummary(summary); err != nil {
		return nil, err
	}
	return canonicalObjectStoreBackupSummaryBytes(summary), nil
}

func DecodeObjectStoreBackupSummaryArtifact(body []byte) (ObjectStoreBackupSummary, error) {
	if err := rejectDuplicateJSONKeys(body); err != nil {
		return ObjectStoreBackupSummary{}, fmt.Errorf("%w: object-store backup summary JSON object keys must be unique: %v", ErrInvalidBackupArtifact, err)
	}
	var summary ObjectStoreBackupSummary
	if err := decodeStrictJSON(body, &summary); err != nil {
		return ObjectStoreBackupSummary{}, fmt.Errorf("%w: decode object-store backup summary: %v", ErrInvalidBackupArtifact, err)
	}
	if err := ValidateObjectStoreBackupSummary(summary); err != nil {
		return ObjectStoreBackupSummary{}, err
	}
	if !bytes.Equal(body, canonicalObjectStoreBackupSummaryBytes(summary)) {
		return ObjectStoreBackupSummary{}, fmt.Errorf("%w: object-store backup summary is not canonical JSON", ErrInvalidBackupArtifact)
	}
	return summary, nil
}

func ValidateObjectStoreBackupSummary(summary ObjectStoreBackupSummary) error {
	if summary.SchemaID != ObjectStoreBackupSummarySchemaID {
		return fmt.Errorf("%w: unsupported object-store backup summary schema %q", ErrInvalidBackupArtifact, summary.SchemaID)
	}
	if _, err := uuid.Parse(summary.BackupSetID); err != nil {
		return fmt.Errorf("%w: object-store backup summary backup_set_id must be a UUID", ErrInvalidBackupArtifact)
	}
	if summary.ConsistencyPointAt.IsZero() {
		return fmt.Errorf("%w: object-store backup summary consistency_point_at is required", ErrInvalidBackupArtifact)
	}
	if summary.ObjectStoreBackend != ObjectStoreBackendSeaweedFSS3 {
		return fmt.Errorf("%w: object-store backup summary backend must be %s", ErrInvalidBackupArtifact, ObjectStoreBackendSeaweedFSS3)
	}
	if err := validateRedactionRef(summary.BucketRef, "bucket"); err != nil {
		return err
	}
	if summary.ObjectCount != len(summary.ObjectsSummary) {
		return fmt.Errorf("%w: object-store backup summary object_count mismatch", ErrInvalidBackupArtifact)
	}
	if !validSHA256Hex(summary.ManifestSHA256) {
		return fmt.Errorf("%w: object-store backup summary manifest_sha256 is required", ErrInvalidBackupArtifact)
	}
	totalSize := int64(0)
	for _, object := range summary.ObjectsSummary {
		if !validSHA256Hex(object.StorageRefSHA256) || !validSHA256Hex(object.SHA256) {
			return fmt.Errorf("%w: object-store backup summary object hashes are required", ErrInvalidBackupArtifact)
		}
		if object.SizeBytes < 0 {
			return fmt.Errorf("%w: object-store backup summary size_bytes is negative", ErrInvalidBackupArtifact)
		}
		totalSize += object.SizeBytes
	}
	if totalSize != summary.TotalSizeBytes {
		return fmt.Errorf("%w: object-store backup summary total_size_bytes mismatch", ErrInvalidBackupArtifact)
	}
	return nil
}

func DecodeRestoreVerificationArtifactV1(body []byte) (RestoreVerificationArtifactV1, error) {
	if err := rejectDuplicateJSONKeys(body); err != nil {
		return RestoreVerificationArtifactV1{}, fmt.Errorf("%w: restore verification artifact JSON object keys must be unique: %v", ErrInvalidBackupArtifact, err)
	}
	var artifact RestoreVerificationArtifactV1
	if err := decodeStrictJSON(body, &artifact); err != nil {
		return RestoreVerificationArtifactV1{}, fmt.Errorf("%w: decode restore verification artifact: %v", ErrInvalidBackupArtifact, err)
	}
	if err := validateRestoreVerificationArtifactV1(artifact); err != nil {
		return RestoreVerificationArtifactV1{}, err
	}
	if !bytes.Equal(body, canonicalRestoreVerificationArtifactBytes(artifact, true)) {
		return RestoreVerificationArtifactV1{}, fmt.Errorf("%w: restore verification artifact is not canonical JSON", ErrInvalidBackupArtifact)
	}
	return artifact, nil
}

func validateRestoreVerificationArtifactV1(artifact RestoreVerificationArtifactV1) error {
	if err := validateRestoreVerificationArtifactV1WithoutDigest(artifact); err != nil {
		return err
	}
	if !validSHA256Hex(artifact.ArtifactSHA256) {
		return fmt.Errorf("%w: restore verification artifact_sha256 is required", ErrInvalidBackupArtifact)
	}
	if got := sha256Hex(canonicalRestoreVerificationArtifactBytes(artifact, false)); got != artifact.ArtifactSHA256 {
		return fmt.Errorf("%w: restore verification artifact_sha256 mismatch", ErrInvalidBackupArtifact)
	}
	return nil
}

func validateRestoreVerificationArtifactV1WithoutDigest(artifact RestoreVerificationArtifactV1) error {
	if artifact.SchemaID != restoreVerificationArtifactV1SchemaID {
		return fmt.Errorf("%w: unsupported restore verification artifact schema %q", ErrInvalidBackupArtifact, artifact.SchemaID)
	}
	if _, err := uuid.Parse(artifact.BackupSetID); err != nil {
		return fmt.Errorf("%w: restore verification backup_set_id must be a UUID", ErrInvalidBackupArtifact)
	}
	if artifact.SelectedIncidentID != nil {
		if _, err := uuid.Parse(*artifact.SelectedIncidentID); err != nil {
			return fmt.Errorf("%w: restore verification selected_incident_id must be a UUID or null", ErrInvalidBackupArtifact)
		}
	}
	switch artifact.IncidentOpenCheck.Status {
	case "pass", "fail", "skipped_no_incidents":
	default:
		return fmt.Errorf("%w: restore verification incident_open_check.status is outside the closed vocabulary", ErrInvalidBackupArtifact)
	}
	if artifact.SelectedIncidentID == nil && artifact.IncidentOpenCheck.Status != "skipped_no_incidents" {
		return fmt.Errorf("%w: zero-incident restore verification must use skipped_no_incidents", ErrInvalidBackupArtifact)
	}
	if artifact.SelectedIncidentID != nil && artifact.IncidentOpenCheck.Status == "skipped_no_incidents" {
		return fmt.Errorf("%w: non-zero incident restore verification cannot skip incident check", ErrInvalidBackupArtifact)
	}
	if artifact.QueryViewSchemaID != "" && artifact.QueryViewSchemaID != historicalTimelineViewSchemaID {
		return fmt.Errorf("%w: restore verification query_view_schema_id is unsupported", ErrInvalidBackupArtifact)
	}
	if artifact.SelectedIncidentID != nil && artifact.QueryViewSchemaID != historicalTimelineViewSchemaID {
		return fmt.Errorf("%w: restore verification query_view_schema_id is required when an incident is checked", ErrInvalidBackupArtifact)
	}
	if artifact.BlobCheckCounts.Total < 0 || artifact.BlobCheckCounts.Passed < 0 || artifact.BlobCheckCounts.Failed < 0 ||
		artifact.BlobCheckCounts.Passed+artifact.BlobCheckCounts.Failed != artifact.BlobCheckCounts.Total {
		return fmt.Errorf("%w: restore verification blob_check_counts are inconsistent", ErrInvalidBackupArtifact)
	}
	if artifact.ManifestCheckResult != "pass" && artifact.ManifestCheckResult != "fail" {
		return fmt.Errorf("%w: restore verification manifest_check_result is outside the closed vocabulary", ErrInvalidBackupArtifact)
	}
	if artifact.ProjectionRebuildResult != "pass" && artifact.ProjectionRebuildResult != "fail" {
		return fmt.Errorf("%w: restore verification projection_rebuild_result is outside the closed vocabulary", ErrInvalidBackupArtifact)
	}
	switch artifact.Result {
	case "pass":
		if len(artifact.FailureReasons) != 0 {
			return fmt.Errorf("%w: passing restore verification artifact cannot include failure_reasons", ErrInvalidBackupArtifact)
		}
		if artifact.ManifestCheckResult != "pass" || artifact.ProjectionRebuildResult != "pass" ||
			artifact.BlobCheckCounts.Failed != 0 || artifact.IncidentOpenCheck.Status == "fail" {
			return fmt.Errorf("%w: passing restore verification artifact has failed checks", ErrInvalidBackupArtifact)
		}
	case "fail":
		if len(artifact.FailureReasons) == 0 {
			return fmt.Errorf("%w: failed restore verification artifact requires failure_reasons", ErrInvalidBackupArtifact)
		}
	default:
		return fmt.Errorf("%w: restore verification result is outside the closed vocabulary", ErrInvalidBackupArtifact)
	}
	for _, reason := range artifact.FailureReasons {
		if strings.TrimSpace(reason) == "" || strings.ContainsAny(reason, "\r\n ") {
			return fmt.Errorf("%w: restore verification failure_reasons must be closed tokens", ErrInvalidBackupArtifact)
		}
	}
	return nil
}

func canonicalObjectStoreBackupManifestBytes(manifest ObjectStoreBackupManifest, includeDigest bool) []byte {
	objects := make([]any, 0, len(manifest.Objects))
	for _, object := range manifest.Objects {
		item := map[string]any{
			"backup_member_sha256": object.BackupMemberSHA256,
			"sha256":               object.SHA256,
			"size_bytes":           object.SizeBytes,
			"storage_ref":          object.StorageRef,
			"storage_ref_sha256":   object.StorageRefSHA256,
		}
		if strings.TrimSpace(object.ObjectBlobID) != "" {
			item["object_blob_id"] = object.ObjectBlobID
		}
		objects = append(objects, item)
	}
	value := map[string]any{
		"backup_set_id":        manifest.BackupSetID,
		"bucket":               manifest.Bucket,
		"consistency_point_at": manifest.ConsistencyPointAt,
		"object_count":         manifest.ObjectCount,
		"object_store_backend": manifest.ObjectStoreBackend,
		"objects":              objects,
		"schema_id":            manifest.SchemaID,
		"total_size_bytes":     manifest.TotalSizeBytes,
	}
	if includeDigest {
		value["manifest_sha256"] = manifest.ManifestSHA256
	}
	return marshalCanonical(value)
}

func canonicalObjectStoreBackupSummaryBytes(summary ObjectStoreBackupSummary) []byte {
	objects := make([]any, 0, len(summary.ObjectsSummary))
	for _, object := range summary.ObjectsSummary {
		objects = append(objects, map[string]any{
			"sha256":             object.SHA256,
			"size_bytes":         object.SizeBytes,
			"storage_ref_sha256": object.StorageRefSHA256,
		})
	}
	return marshalCanonical(map[string]any{
		"backup_set_id":        summary.BackupSetID,
		"bucket_ref":           redactionRefCanonicalMap(summary.BucketRef),
		"consistency_point_at": summary.ConsistencyPointAt,
		"manifest_sha256":      summary.ManifestSHA256,
		"object_count":         summary.ObjectCount,
		"object_store_backend": summary.ObjectStoreBackend,
		"objects_summary":      objects,
		"schema_id":            summary.SchemaID,
		"total_size_bytes":     summary.TotalSizeBytes,
	})
}

func canonicalRestoreVerificationArtifactBytes(artifact RestoreVerificationArtifactV1, includeDigest bool) []byte {
	value := map[string]any{
		"backup_set_id":             artifact.BackupSetID,
		"blob_check_counts":         map[string]any{"failed": artifact.BlobCheckCounts.Failed, "passed": artifact.BlobCheckCounts.Passed, "total": artifact.BlobCheckCounts.Total},
		"failure_reasons":           artifact.FailureReasons,
		"incident_open_check":       map[string]any{"status": artifact.IncidentOpenCheck.Status},
		"manifest_check_result":     artifact.ManifestCheckResult,
		"projection_rebuild_result": artifact.ProjectionRebuildResult,
		"query_view_schema_id":      artifact.QueryViewSchemaID,
		"result":                    artifact.Result,
		"schema_id":                 artifact.SchemaID,
		"selected_incident_id":      artifact.SelectedIncidentID,
	}
	if includeDigest {
		value["artifact_sha256"] = artifact.ArtifactSHA256
	}
	return marshalCanonical(value)
}

func marshalCanonical(value any) []byte {
	body, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("canonical JSON encode failed: %v", err))
	}
	return append(body, '\n')
}

func listBackupManifestObjects(ctx context.Context, store objectstore.Store, prefix string) ([]objectstore.ObjectInfo, error) {
	if typed, ok := store.(objectstore.TypedStore); ok {
		result, err := typed.ListPrefix(ctx, objectstore.ListPrefixRequest{
			Prefix:  prefix,
			Purpose: objectstore.PurposeBackupManifest,
		})
		if err != nil {
			return nil, fmt.Errorf("list object store backup manifest objects: %w", err)
		}
		return result.Objects, nil
	}
	objects, err := store.ListObjects(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("list object store backup manifest objects: %w", err)
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Key < objects[j].Key
	})
	return objects, nil
}

func getBackupManifestObject(ctx context.Context, store objectstore.Store, key string) (io.ReadCloser, objectstore.ObjectInfo, error) {
	if typed, ok := store.(objectstore.TypedStore); ok {
		reader, info, err := typed.Get(ctx, objectstore.GetObjectRequest{
			Key:     key,
			Purpose: objectstore.PurposeBackupManifest,
		})
		if err != nil {
			return nil, objectstore.ObjectInfo{}, fmt.Errorf("read object store backup object %s: %w", key, err)
		}
		return reader, info, nil
	}
	reader, info, err := store.ReadObject(ctx, key, objectstore.ReadOptions{})
	if err != nil {
		return nil, objectstore.ObjectInfo{}, fmt.Errorf("read object store backup object %s: %w", key, err)
	}
	return reader, info, nil
}

func hashRedactionRef(redactionClass string, raw string) RedactionRef {
	return RedactionRef{
		Redacted:       true,
		RedactionClass: redactionClass,
		SHA256:         sha256Hex([]byte(raw)),
	}
}

func redactionRefCanonicalMap(ref RedactionRef) map[string]any {
	value := map[string]any{
		"redacted":        ref.Redacted,
		"redaction_class": ref.RedactionClass,
	}
	if ref.SHA256 != "" {
		value["sha256"] = ref.SHA256
	}
	if ref.Scheme != "" {
		value["scheme"] = ref.Scheme
	}
	if ref.HostSHA256 != "" {
		value["host_sha256"] = ref.HostSHA256
	}
	if ref.PortPresent != nil {
		value["port_present"] = *ref.PortPresent
	}
	return value
}

func validateRedactionRef(ref RedactionRef, redactionClass string) error {
	if !ref.Redacted {
		return fmt.Errorf("%w: %s redaction ref must be redacted", ErrInvalidBackupArtifact, redactionClass)
	}
	if ref.RedactionClass != redactionClass {
		return fmt.Errorf("%w: redaction_class must be %s", ErrInvalidBackupArtifact, redactionClass)
	}
	if ref.SHA256 != "" && !validSHA256Hex(ref.SHA256) {
		return fmt.Errorf("%w: redaction ref sha256 must be lowercase hex", ErrInvalidBackupArtifact)
	}
	return nil
}

func decodeStrictJSON(body []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return fmt.Errorf("trailing JSON content")
	}
	return nil
}

func rejectDuplicateJSONKeys(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := scanUniqueJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON content")
		}
		return err
	}
	return nil
}

func scanUniqueJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("object key is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate object key %q", key)
			}
			seen[key] = struct{}{}
			if err := scanUniqueJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim('}') {
			return fmt.Errorf("object not closed")
		}
	case '[':
		for decoder.More() {
			if err := scanUniqueJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return err
		}
		if end != json.Delim(']') {
			return fmt.Errorf("array not closed")
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	return nil
}
