package conflicts

import (
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const ConflictTokenVersion = conflicttokens.ConflictTokenVersion

type ConflictTokenClaims = conflicttokens.ConflictTokenClaims
type ConflictTokenCodec = conflicttokens.ConflictTokenCodec

func NewConflictTokenCodec(keys authn.MasterKeys) ConflictTokenCodec {
	return conflicttokens.NewConflictTokenCodec(keys)
}

func NewConflictTokenCodecForTesting(scope string) ConflictTokenCodec {
	return conflicttokens.NewConflictTokenCodecForTesting(scope)
}

func RequestHashTokenValue(requestHash []byte) string {
	return conflicttokens.RequestHashTokenValue(requestHash)
}
