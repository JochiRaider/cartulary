package conflicttest

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

func NewCodec(scope string) conflicttokens.ConflictTokenCodec {
	key := sha256.Sum256([]byte("cartulary-test-conflict-token-v3:" + scope))
	secretRef := "test-" + scope
	manifest := fmt.Sprintf(`{"schema_id":"cartulary.revisions_conflict_token_key_ring.v1","algorithm":"aes_256_gcm_v1","keys":[{"conflict_token_key_id":"%s","state":"active","secret_ref":{"kind":"env","name":"%s"}}]}`, secretRef, secretRef)
	ring, err := conflicttokens.ParseConflictTokenKeyRing([]byte(manifest), map[string]string{
		"CARTULARY_SECRET_" + normalizedSecretSuffix(secretRef): base64.RawURLEncoding.EncodeToString(key[:]),
	}, time.Now().UTC(), conflicttokens.KeyRingParseOptions{AllowFixtureKey: true})
	if err != nil {
		panic(err)
	}
	codec, err := conflicttokens.NewConflictTokenCodec(ring)
	if err != nil {
		panic(err)
	}
	return codec
}

func normalizedSecretSuffix(value string) string {
	result := make([]byte, 0, len(value))
	for _, char := range []byte(value) {
		switch {
		case char >= 'a' && char <= 'z':
			result = append(result, char-'a'+'A')
		case char >= 'A' && char <= 'Z', char >= '0' && char <= '9':
			result = append(result, char)
		default:
			result = append(result, '_')
		}
	}
	return string(result)
}
