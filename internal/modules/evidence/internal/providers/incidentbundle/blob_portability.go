package incidentbundle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type blobPortability struct {
	store objectstore.TypedStore
}

func NewBlobPortability(store objectstore.TypedStore) *blobPortability {
	return &blobPortability{store: store}
}

func (portability *blobPortability) ExportBlobFiles(
	ctx context.Context,
	queryer incidentportability.Queryer,
	incidentID uuid.UUID,
	files map[string][]byte,
) error {
	rows, err := queryer.Query(ctx, `
SELECT storage_key, observed_sha256_hex
  FROM object_blobs
 WHERE incident_id = $1
   AND upload_state = 'available'
   AND observed_sha256_hex IS NOT NULL
 ORDER BY observed_sha256_hex
`, incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var storageKey string
		var sha string
		if err := rows.Scan(&storageKey, &sha); err != nil {
			return err
		}
		reader, _, err := portability.store.Get(ctx, objectstore.GetObjectRequest{
			Key:     storageKey,
			Purpose: objectstore.PurposeMigrationCopy,
		})
		if err != nil {
			return &incidentportability.VerificationFailure{ReasonCode: "missing_required_blob"}
		}
		data, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			return &incidentportability.VerificationFailure{ReasonCode: "missing_required_blob"}
		}
		if hashHex(data) != sha {
			return &incidentportability.VerificationFailure{ReasonCode: "missing_required_blob"}
		}
		files["blobs/sha256/"+sha] = data
	}
	return rows.Err()
}

func (portability *blobPortability) RewriteAndStageObjectBlobs(
	ctx context.Context,
	files map[string][]byte,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	attributions incidentportability.AttributionRecorder,
) ([]byte, []string, error) {
	rows, err := incidentportability.DecodeNDJSON(files["data/object_blobs.ndjson"])
	if err != nil {
		return nil, nil, err
	}
	writtenKeys := make([]string, 0, len(rows))
	var buffer bytes.Buffer
	for _, row := range rows {
		if err := incidentportability.RemapTopLevelUserFields(row, "object_blobs", []string{"object_blob_id"}, actorUserID, attributions); err != nil {
			return nil, nil, err
		}
		objectBlobText, ok := row["object_blob_id"].(string)
		if !ok || objectBlobText == "" {
			return nil, writtenKeys, &incidentportability.VerificationFailure{ReasonCode: "malformed_manifest"}
		}
		objectBlobID, err := uuid.Parse(objectBlobText)
		if err != nil {
			return nil, writtenKeys, &incidentportability.VerificationFailure{ReasonCode: "malformed_manifest"}
		}
		storageKey, err := blobref.ObjectBlobStorageKey(incidentID, objectBlobID)
		if err != nil {
			return nil, writtenKeys, &incidentportability.VerificationFailure{ReasonCode: "malformed_manifest"}
		}
		row["storage_key"] = storageKey
		state, _ := row["upload_state"].(string)
		if state != "available" {
			line, err := incidentportability.CanonicalJSONString(row)
			if err != nil {
				return nil, writtenKeys, err
			}
			buffer.Write(line)
			continue
		}
		sha, _ := row["observed_sha256_hex"].(string)
		if sha == "" {
			sha, _ = row["expected_sha256_hex"].(string)
		}
		data, ok := files["blobs/sha256/"+sha]
		if !ok {
			return nil, writtenKeys, &incidentportability.VerificationFailure{ReasonCode: "missing_required_blob"}
		}
		if hashHex(data) != sha {
			return nil, writtenKeys, &incidentportability.VerificationFailure{ReasonCode: "blob_hash_mismatch"}
		}
		contentType, _ := row["observed_content_type"].(string)
		if contentType == "" {
			contentType, _ = row["content_type_hint"].(string)
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		if _, err := portability.store.Put(ctx, objectstore.PutObjectRequest{
			Key:         storageKey,
			Body:        bytes.NewReader(data),
			Size:        int64(len(data)),
			ContentType: contentType,
			Purpose:     objectstore.PurposeMigrationCopy,
		}); err != nil {
			return nil, writtenKeys, err
		}
		writtenKeys = append(writtenKeys, storageKey)
		line, err := incidentportability.CanonicalJSONString(row)
		if err != nil {
			return nil, writtenKeys, err
		}
		buffer.Write(line)
	}
	return buffer.Bytes(), writtenKeys, nil
}

func (portability *blobPortability) CleanupStagedObjects(ctx context.Context, keys []string) {
	for _, key := range keys {
		_ = portability.store.Delete(ctx, objectstore.DeleteObjectRequest{
			Key:     key,
			Purpose: objectstore.PurposeStagedCleanup,
		})
	}
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
