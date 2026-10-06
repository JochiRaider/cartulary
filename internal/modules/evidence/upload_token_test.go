package evidence

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestObjectUploadCapabilityVersion1IsRejected_Unit(t *testing.T) {
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatalf("load test master keys: %v", err)
	}
	versionOnePayload, err := json.Marshal(map[string]any{
		"v":                    1,
		"object_blob_id":       uuid.New(),
		"incident_id":          uuid.New(),
		"storage_key":          "evidence_lifecycle/version-1/object",
		"byte_size":            5,
		"expires_at_unix_nano": int64(2),
	})
	if err != nil {
		t.Fatalf("marshal version-1 claims: %v", err)
	}
	payloadSegment := base64.RawURLEncoding.EncodeToString(versionOnePayload)
	versionOneKey, err := authn.DerivePurposeKey(keys, "evidence-object-upload-v1")
	if err != nil {
		t.Fatal(err)
	}
	mac := hmac.New(sha256.New, versionOneKey[:])
	_, _ = mac.Write([]byte(payloadSegment))
	versionOneToken := objectUploadTokenPrefix + payloadSegment + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if _, err := decodeObjectUploadToken(keys, versionOneToken); !errors.Is(err, errInvalidObjectUploadToken) {
		t.Fatalf("decode version-1 upload target error = %v, want invalid token", err)
	}
}

func TestObjectUploadCapabilityOpaqueBindingAndBounds_Unit(t *testing.T) {
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatal(err)
	}
	storageKey := "incidents/private-storage-key/object-blobs/private-object"
	binding, err := objectUploadStorageBinding(keys, storageKey)
	if err != nil {
		t.Fatal(err)
	}
	claims := objectUploadTokenClaims{Version: objectUploadTokenVersion, LeaseID: uuid.New(), ObjectBlobID: uuid.New(), IncidentID: uuid.New(), IssuingUserID: uuid.New(), IssuingSessionID: uuid.New(), StorageKeyBinding: binding, ByteSize: 1, RequiredMethod: "PUT", RequiredHeadersSHA256: strings.Repeat("a", 64), AcceptedContractSHA256: strings.Repeat("b", 64), IssuedAtUnixNano: 1, ExpiresAtUnixNano: 2}
	token, err := encodeObjectUploadToken(keys, claims)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := decodeObjectUploadToken(keys, token)
	if err != nil || parsed != claims {
		t.Fatal("current capability round trip", err)
	}
	parts := strings.Split(strings.TrimPrefix(token, objectUploadTokenPrefix), ".")
	decoded, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(decoded), storageKey) || strings.Contains(string(decoded), `"storage_key":`) {
		t.Fatal("capability disclosed storage locator")
	}
	other, err := objectUploadStorageBinding(keys, storageKey+"-other")
	if err != nil || other == binding {
		t.Fatal("storage binding collision", err)
	}
	for _, bad := range []string{" " + token, token + "\n", token + "=", token[:len(token)-1], strings.Repeat("x", objectUploadTokenMaximumBytes+1)} {
		if _, err := decodeObjectUploadToken(keys, bad); err == nil {
			t.Fatal("invalid framing accepted")
		}
	}
	// Old payloads do not become current even under a current signing key.
	for _, version := range []int{1, 2} {
		claims.Version = version
		old, err := encodeObjectUploadToken(keys, claims)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := decodeObjectUploadToken(keys, old); err == nil {
			t.Fatal("old version accepted")
		}
	}
	claims.Version = objectUploadTokenVersion
	claims.StorageKeyBinding = strings.Repeat("x", objectUploadTokenMaximumBytes)
	if _, err := encodeObjectUploadToken(keys, claims); err == nil {
		t.Fatal("oversized token issued")
	}
}
