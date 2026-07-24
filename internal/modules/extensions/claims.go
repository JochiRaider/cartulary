package extensions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
)

var extensionProfileIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// ResolvedClaimSet is the immutable identity shared by every consumer of the
// process-lifetime extension claim decision. It intentionally contains no
// configuration values or profile implementation details.
type ResolvedClaimSet struct {
	profileIDs []string
	sha256     string
}

// NewResolvedClaimSet canonicalizes the admitted extension profile IDs and
// binds the result to the canonical resolved-claim-set digest.
func NewResolvedClaimSet(profileIDs []string) (ResolvedClaimSet, error) {
	unique := make(map[string]struct{}, len(profileIDs))
	for _, profileID := range profileIDs {
		if profileID == "base" || !extensionProfileIDPattern.MatchString(profileID) {
			return ResolvedClaimSet{}, fmt.Errorf("invalid resolved extension profile id")
		}
		unique[profileID] = struct{}{}
	}

	canonicalIDs := make([]string, 0, len(unique))
	for profileID := range unique {
		canonicalIDs = append(canonicalIDs, profileID)
	}
	sort.Strings(canonicalIDs)
	canonicalBytes, err := json.Marshal(struct {
		ProfileIDs []string `json:"profile_ids"`
	}{ProfileIDs: canonicalIDs})
	if err != nil {
		return ResolvedClaimSet{}, fmt.Errorf("encode resolved extension claim set: %w", err)
	}
	digest := sha256.Sum256(canonicalBytes)
	return ResolvedClaimSet{
		profileIDs: canonicalIDs,
		sha256:     hex.EncodeToString(digest[:]),
	}, nil
}

func (claims ResolvedClaimSet) ProfileIDs() []string {
	return append([]string(nil), claims.profileIDs...)
}

func (claims ResolvedClaimSet) SHA256() string {
	return claims.sha256
}

func (claims ResolvedClaimSet) Contains(profileID string) bool {
	index := sort.SearchStrings(claims.profileIDs, profileID)
	return index < len(claims.profileIDs) && claims.profileIDs[index] == profileID
}
