package evidence

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"io"
	"strings"
)

// Blob observation belongs to the background verification owner; route handlers
// have no capability to compute a successful stored-byte observation.
var errBlobVerificationMismatch = errors.New("evidence: stored object violates its accepted upload contract")

type blobVerificationObjectReader struct {
	store objectstore.TypedStore
}

func (adapter blobVerificationObjectReader) observeUploadedObject(
	ctx context.Context,
	blob blobRecord,
) (*observedObject, error) {
	if err := evidencepolicy.ValidatePersistedObjectBlobStorageKey(blob.StorageKey, blob.IncidentID, blob.ObjectBlobID); err != nil {
		return nil, err
	}
	stat, err := adapter.store.Head(ctx, objectstore.HeadObjectRequest{
		Key:     blob.StorageKey,
		Purpose: objectstore.PurposeProductUpload,
	})
	if err != nil {
		return nil, err
	}
	object, _, err := adapter.store.Get(ctx, objectstore.GetObjectRequest{
		Key:     blob.StorageKey,
		Purpose: objectstore.PurposeProductRead,
	})
	if err != nil {
		return nil, err
	}
	defer object.Close()
	stopClose := context.AfterFunc(ctx, func() { _ = object.Close() })
	defer stopClose()
	hash := sha256.New()
	readSize, err := io.Copy(hash, io.LimitReader(object, blob.ByteSize+1))
	if err != nil {
		return nil, err
	}
	if readSize != stat.Size || readSize != blob.ByteSize {
		return nil, errBlobVerificationMismatch
	}
	contentType := strings.TrimSpace(stat.ContentType)
	if contentType == "" {
		contentType = firstNonEmptyPtr(blob.ContentTypeHint, nil, "application/octet-stream")
	}
	return &observedObject{Size: stat.Size, ContentType: contentType, SHA256Hex: fmt.Sprintf("%x", hash.Sum(nil))}, nil
}
