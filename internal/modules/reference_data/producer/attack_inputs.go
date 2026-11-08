package producer

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"

	attackinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/attack"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

// These inputs are producer provenance, never a runtime profile registry.
type attackInputs struct {
	lock                                 SourceLock
	policy                               Policy
	profileID                            string
	lockBytes, profileBytes, policyBytes []byte
}

func admitATTACKInputs(lockBytes, profileBytes, policyBytes []byte) (attackInputs, error) {
	current, err := attackinputs.Files.ReadFile("source-profile.native.v2.r2.json")
	if err != nil {
		return attackInputs{}, err
	}
	if !bytes.Equal(profileBytes, current) {
		return attackInputs{}, errors.New("exact frozen ATT&CK source profile required")
	}
	var profile struct {
		SchemaID     string            `json:"schema_id"`
		ID           string            `json:"source_profile_id"`
		Status       string            `json:"status"`
		Locks        []string          `json:"accepted_source_lock_sha256"`
		PolicySHA256 string            `json:"field_policy_sha256"`
		Schemas      map[string]string `json:"adopted_schema_sha256"`
	}
	// The entire document has already been bound to immutable compiled bytes.
	if err := json.Unmarshal(profileBytes, &profile); err != nil {
		return attackInputs{}, err
	}
	if profile.SchemaID != "cartulary.attack_source_profile.v1" || profile.ID != "cartulary.attack.all-domains.native.v2.r2" || profile.Status != "adopted_frozen" {
		return attackInputs{}, errors.New("frozen adopted source profile required")
	}
	lock, err := ReadLock(lockBytes)
	if err != nil {
		return attackInputs{}, err
	}
	if !slices.Contains(profile.Locks, Digest(lockBytes)) {
		return attackInputs{}, errors.New("source lock outside frozen profile")
	}
	if profile.PolicySHA256 != Digest(policyBytes) {
		return attackInputs{}, errors.New("field policy identity mismatch")
	}
	policy, err := ReadPolicy(policyBytes)
	if err != nil {
		return attackInputs{}, err
	}
	if len(profile.Schemas) != 3 {
		return attackInputs{}, errors.New("native schema bindings required")
	}
	for _, name := range []string{"profiles/framework.attack.native_source.v2.schema.json", "profiles/framework.attack.object.v2.schema.json", "profiles/framework.attack.relationship.v2.schema.json"} {
		if profile.Schemas[name] != packformat.ProjectionDigest(name) {
			return attackInputs{}, errors.New("frozen native schema identity mismatch")
		}
	}
	return attackInputs{lock: lock, policy: policy, profileID: profile.ID, lockBytes: bytes.Clone(lockBytes), profileBytes: bytes.Clone(profileBytes), policyBytes: bytes.Clone(policyBytes)}, nil
}
