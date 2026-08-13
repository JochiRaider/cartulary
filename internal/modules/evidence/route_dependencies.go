package evidence

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/url"
	"strings"

	evidencepolicy "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/policy"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type uploadCapabilityService struct {
	keys authn.MasterKeys
}

type createdUploadTarget struct {
	Target objectstore.UploadTarget
	Token  string
}

func (service uploadCapabilityService) createTarget(
	claims objectUploadTokenClaims,
) (createdUploadTarget, error) {
	token, err := encodeObjectUploadToken(service.keys, claims)
	if err != nil {
		return createdUploadTarget{}, err
	}
	return createdUploadTarget{
		Target: objectstore.UploadTarget{
			Href:    "/api/v1/object-uploads/" + url.PathEscape(token),
			Method:  "PUT",
			Headers: map[string]string{},
		},
		Token: token,
	}, nil
}

// routeObjectStoreAdapter contains the transport-to-platform object-store
// adaptation. It owns no Evidence database state or authorization decisions.
type routeObjectStoreAdapter struct {
	store objectstore.TypedStore
}

func (adapter routeObjectStoreAdapter) observeUploadedObject(
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
	hash := sha256.New()
	if _, err := io.Copy(hash, object); err != nil {
		return nil, err
	}
	contentType := strings.TrimSpace(stat.ContentType)
	if contentType == "" {
		contentType = firstNonEmptyPtr(blob.ContentTypeHint, nil, "application/octet-stream")
	}
	return &observedObject{Size: stat.Size, ContentType: contentType, SHA256Hex: fmt.Sprintf("%x", hash.Sum(nil))}, nil
}

func (adapter routeObjectStoreAdapter) verifyEvidenceObjectAvailable(
	ctx context.Context,
	access evidenceAccessRecord,
) (string, *httpapi.APIError) {
	if access.StorageKey == nil || access.ObjectBlobID == nil {
		return "evidence_inconsistent", nil
	}
	if err := evidencepolicy.ValidatePersistedObjectBlobStorageKey(*access.StorageKey, access.IncidentID, *access.ObjectBlobID); err != nil {
		return "", objectStoreDependencyAPIError(err)
	}
	if _, err := adapter.store.Head(ctx, objectstore.HeadObjectRequest{
		Key:     *access.StorageKey,
		Purpose: objectstore.PurposeProductRead,
	}); err != nil {
		if apiErr := objectStoreDependencyAPIError(err); apiErr != nil {
			return "", apiErr
		}
		return "blob_missing", nil
	}
	return "", nil
}

func (adapter routeObjectStoreAdapter) put(
	ctx context.Context,
	key string,
	body io.Reader,
	size int64,
	contentType string,
	purpose objectstore.Purpose,
) error {
	_, err := adapter.store.Put(ctx, objectstore.PutObjectRequest{
		Key:         key,
		Body:        body,
		Size:        size,
		ContentType: contentType,
		Purpose:     purpose,
	})
	return err
}

func (adapter routeObjectStoreAdapter) get(
	ctx context.Context,
	key string,
	options objectstore.ReadOptions,
	purpose objectstore.Purpose,
) (io.ReadCloser, objectstore.ObjectInfo, error) {
	return adapter.store.Get(ctx, objectstore.GetObjectRequest{
		Key:        key,
		RangeStart: options.RangeStart,
		RangeEnd:   options.RangeEnd,
		Purpose:    purpose,
	})
}
