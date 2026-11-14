package policy

import (
	"encoding/hex"
	"strings"
)

// Recovery requires every finalized object's exact observations, including
// quarantined objects. Missing observations cannot make a required row vanish.
func ValidRecoveryBlobMetadata(declared int64, observed *int64, digest, expected *string) bool {
	if declared < 0 || observed == nil || *observed != declared || digest == nil || len(*digest) != 64 || strings.ToLower(*digest) != *digest || (expected != nil && *expected != *digest) {
		return false
	}
	_, err := hex.DecodeString(*digest)
	return err == nil
}
