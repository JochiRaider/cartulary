package packformat

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var provenanceShape = compileProjection("provenance.v1.schema.json")

// ValidateProvenance admits the exact immutable anchor for every member of an
// already admitted set. It does not authenticate source-supplied provenance.
func ValidateProvenance(data []byte, set Set) error {
	invalid := errors.New("reference pack: invalid retained provenance")
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return invalid
	}
	anchors, ok := value.([]any)
	if !ok || len(anchors) != len(set.Members) {
		return invalid
	}
	canonical, err := canonicaljson.Marshal(value)
	if err != nil || !bytes.Equal(canonical, data) {
		return invalid
	}
	for i, value := range anchors {
		if !provenanceShape.matches(value) {
			return invalid
		}
		a := value.(map[string]any)
		m := set.Members[i]
		p, known := profiles[m.Key]
		if !known || a["pack_set_id"] != set.ID || a["pack_key"] != m.Key || a["pack_version"] != m.Version ||
			a["manifest_sha256"] != m.ManifestSHA256 || a["payload_sha256"] != m.PayloadSHA256 ||
			a["pack_contract_version"] != m.Contract || a["content_profile_id"] != m.ProfileID ||
			a["content_profile_version"] != m.ProfileVersion || a["authority_class"] != p.Authority {
			return invalid
		}
		verified, err := time.Parse(time.RFC3339Nano, a["last_verified_at"].(string))
		if err != nil || verified.IsZero() || verified.UTC().Format(time.RFC3339Nano) != a["last_verified_at"] {
			return invalid
		}
		if sourceTime, ok := a["source_as_of"].(string); ok {
			if _, err := time.Parse(time.RFC3339, sourceTime); err != nil {
				return invalid
			}
		}
		signers := a["verified_signer_key_ids"].([]any)
		if a["verification_method"] == "packaged_release_manifest_v1" {
			if a["trust_valid_until"] != nil || len(signers) != 0 {
				return invalid
			}
		} else {
			expiry, ok := a["trust_valid_until"].(string)
			if !ok || len(signers) == 0 {
				return invalid
			}
			until, err := time.Parse(time.RFC3339Nano, expiry)
			if err != nil || !verified.Before(until) || until.UTC().Format(time.RFC3339Nano) != expiry {
				return invalid
			}
			previous := ""
			for _, value := range signers {
				id := value.(string)
				if id <= previous {
					return invalid
				}
				previous = id
			}
		}
		encoded, err := canonicaljson.Marshal(a["license"])
		if err != nil {
			return invalid
		}
		var license License
		if json.Unmarshal(encoded, &license) != nil || !validLicense(license) {
			return invalid
		}
	}
	return nil
}
