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

func TestLegacyObjectUploadCapabilitiesFailClosed_Unit(t *testing.T) {
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatalf("load test master keys: %v", err)
	}
	legacyPayload, err := json.Marshal(map[string]any{
		"v":                    1,
		"object_blob_id":       uuid.New(),
		"incident_id":          uuid.New(),
		"storage_key":          "evidence_lifecycle/legacy/object",
		"byte_size":            5,
		"expires_at_unix_nano": int64(2),
	})
	if err != nil {
		t.Fatalf("marshal legacy claims: %v", err)
	}
	payloadSegment := base64.RawURLEncoding.EncodeToString(legacyPayload)
	legacyKey := authn.DerivePurposeKey(keys, "evidence-object-upload-v1")
	mac := hmac.New(sha256.New, legacyKey[:])
	_, _ = mac.Write([]byte(payloadSegment))
	legacyToken := objectUploadTokenPrefix + payloadSegment + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if _, err := decodeObjectUploadToken(keys, legacyToken); !errors.Is(err, errInvalidObjectUploadToken) {
		t.Fatalf("decode legacy upload target error = %v, want invalid token", err)
	}
}
