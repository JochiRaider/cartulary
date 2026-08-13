package policy

import (
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
)

const (
	ObjectBlobStorageKeyMalformedReason        = "object_blob_storage_key_malformed"
	ObjectBlobStorageKeyIdentityMismatchReason = "object_blob_storage_key_identity_mismatch"
)

type PersistedObjectBlobStorageKeyError struct {
	reasonCode string
}

func (e *PersistedObjectBlobStorageKeyError) Error() string {
	return "persisted object blob storage_key violates object_blob_storage_key_v1"
}

func PersistedObjectBlobStorageKeyErrorReason(err error) (string, bool) {
	var storageKeyError *PersistedObjectBlobStorageKeyError
	if !errors.As(err, &storageKeyError) {
		return "", false
	}
	return storageKeyError.reasonCode, true
}

func ValidatePersistedObjectBlobStorageKey(key string, incidentID uuid.UUID, objectBlobID uuid.UUID) error {
	parts, err := blobref.ParseObjectBlobStorageKey(key)
	if err != nil {
		return &PersistedObjectBlobStorageKeyError{reasonCode: ObjectBlobStorageKeyMalformedReason}
	}
	if parts.IncidentID != incidentID || parts.ObjectBlobID != objectBlobID {
		return &PersistedObjectBlobStorageKeyError{reasonCode: ObjectBlobStorageKeyIdentityMismatchReason}
	}
	return nil
}

func IsServerManagedStorageRef(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), blobref.StorageRefScheme)
}

// ServerManagedStorageRefMatchesAssociation rejects orphaned or mismatched
// logical object references while permitting external reference schemes.
func ServerManagedStorageRefMatchesAssociation(value string, objectBlobID *uuid.UUID) bool {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, blobref.StorageRefScheme) {
		return true
	}
	parsed, err := blobref.ParseObjectBlobStorageRef(trimmed)
	if err != nil {
		return false
	}
	return objectBlobID != nil && parsed == *objectBlobID
}
