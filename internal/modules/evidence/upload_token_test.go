package evidence

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	versionOneKey := authn.DerivePurposeKey(keys, "evidence-object-upload-v1")
	mac := hmac.New(sha256.New, versionOneKey[:])
	_, _ = mac.Write([]byte(payloadSegment))
	versionOneToken := objectUploadTokenPrefix + payloadSegment + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if _, err := decodeObjectUploadToken(keys, versionOneToken); !errors.Is(err, errInvalidObjectUploadToken) {
		t.Fatalf("decode version-1 upload target error = %v, want invalid token", err)
	}
}
