package blobref

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	StorageRefScheme      = "object://"
	MaxStorageKeyBytes    = 1024
	storageKeyIncidentSeg = "incidents"
	storageKeyBlobSeg     = "object-blobs"
)

type StorageKeyParts struct {
	IncidentID   uuid.UUID
	ObjectBlobID uuid.UUID
}

func ObjectBlobStorageKey(incidentID uuid.UUID, objectBlobID uuid.UUID) (string, error) {
	if incidentID == uuid.Nil {
		return "", fmt.Errorf("incident_id is required")
	}
	if objectBlobID == uuid.Nil {
		return "", fmt.Errorf("object_blob_id is required")
	}
	return storageKeyIncidentSeg + "/" + incidentID.String() + "/" + storageKeyBlobSeg + "/" + objectBlobID.String(), nil
}

func ParseObjectBlobStorageKey(key string) (StorageKeyParts, error) {
	if err := validateStorageKeySyntax(key); err != nil {
		return StorageKeyParts{}, err
	}
	parts := strings.Split(key, "/")
	if len(parts) != 4 || parts[0] != storageKeyIncidentSeg || parts[2] != storageKeyBlobSeg {
		return StorageKeyParts{}, fmt.Errorf("object blob storage key must use incidents/{incident_uuid}/object-blobs/{object_blob_uuid}")
	}
	incidentID, err := parseCanonicalUUID(parts[1], "incident_id")
	if err != nil {
		return StorageKeyParts{}, err
	}
	objectBlobID, err := parseCanonicalUUID(parts[3], "object_blob_id")
	if err != nil {
		return StorageKeyParts{}, err
	}
	return StorageKeyParts{IncidentID: incidentID, ObjectBlobID: objectBlobID}, nil
}

func ObjectBlobStorageRef(objectBlobID uuid.UUID) (string, error) {
	if objectBlobID == uuid.Nil {
		return "", fmt.Errorf("object_blob_id is required")
	}
	return StorageRefScheme + objectBlobID.String(), nil
}

func ParseObjectBlobStorageRef(ref string) (uuid.UUID, error) {
	if !strings.HasPrefix(ref, StorageRefScheme) {
		return uuid.Nil, fmt.Errorf("storage ref is not server-managed")
	}
	return parseCanonicalUUID(strings.TrimPrefix(ref, StorageRefScheme), "object_blob_id")
}

func validateStorageKeySyntax(key string) error {
	if key == "" {
		return fmt.Errorf("object blob storage key is required")
	}
	if len([]byte(key)) > MaxStorageKeyBytes {
		return fmt.Errorf("object blob storage key exceeds %d bytes", MaxStorageKeyBytes)
	}
	if strings.ContainsRune(key, '\x00') || strings.ContainsAny(key, "\r\n") {
		return fmt.Errorf("object blob storage key contains invalid control characters")
	}
	if strings.HasPrefix(key, "/") || filepath.IsAbs(key) {
		return fmt.Errorf("object blob storage key must be bucket-relative")
	}
	if strings.Contains(key, "//") {
		return fmt.Errorf("object blob storage key contains an empty segment")
	}
	cleaned := filepath.Clean(filepath.FromSlash(key))
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("object blob storage key escapes the object-store namespace")
	}
	return nil
}

func parseCanonicalUUID(value string, label string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s must be a UUID", label)
	}
	if parsed.String() != value {
		return uuid.Nil, fmt.Errorf("%s must be lowercase canonical UUID text", label)
	}
	if parsed == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%s must not be nil UUID", label)
	}
	return parsed, nil
}
