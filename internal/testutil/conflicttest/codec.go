package conflicttest

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func NewCodec(scope string) conflicttokens.ConflictTokenCodec {
	key := sha256.Sum256([]byte("cartulary-test-conflict-token-v2:" + scope))
	keys, err := authn.LoadMasterKeys(map[string]string{
		authn.AuthMasterKeyEnv: base64.RawStdEncoding.EncodeToString(key[:]),
	})
	if err != nil {
		panic(err)
	}
	return conflicttokens.NewConflictTokenCodec(keys)
}
