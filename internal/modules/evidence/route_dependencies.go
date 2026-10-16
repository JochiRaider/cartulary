package evidence

import (
	"context"
	"io"
	"net/url"

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
	storageKey string,
) (createdUploadTarget, error) {
	binding, err := objectUploadStorageBinding(service.keys, storageKey)
	if err != nil {
		return createdUploadTarget{}, err
	}
	claims.StorageKeyBinding = binding
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
type routeObjectStoreAdapter struct{ store objectstore.TypedStore }

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
