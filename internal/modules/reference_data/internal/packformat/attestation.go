package packformat

import (
	"bytes"
	"errors"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
)

var attestationShape = compileProjection("attestation.v1.schema.json")

// ValidateAttestation admits the complete canonical historical object before
// publication or restore; an attestation is evidence, never a trust anchor.
func ValidateAttestation(data []byte) error {
	invalid := errors.New("reference pack: invalid historical attestation")
	if len(data) > 65536 {
		return invalid
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil || !validPackNumbers(value) || !attestationShape.matches(value) {
		return invalid
	}
	object := value.(map[string]any)
	canonical, err := canonicaljson.Marshal(value)
	if err != nil || !bytes.Equal(canonical, data) {
		return invalid
	}
	at := object["occurred_at"].(string)
	parsed, err := time.Parse(time.RFC3339Nano, at)
	if err != nil || parsed.UTC().Format(time.RFC3339Nano) != at {
		return invalid
	}
	for _, name := range []string{"actor_user_id", "operator_operation_id"} {
		if value := object[name]; value != nil {
			id, err := uuid.Parse(value.(string))
			if err != nil || id == uuid.Nil || id.String() != value {
				return invalid
			}
		}
	}
	if (object["actor_kind"] == "user") != (object["actor_user_id"] != nil) || (object["actor_kind"] == "local_operator") != (object["operator_operation_id"] != nil) {
		return invalid
	}
	if expiry := object["trust_valid_until"]; expiry != nil {
		value, err := time.Parse(time.RFC3339Nano, expiry.(string))
		if err != nil || value.UTC().Format(time.RFC3339Nano) != expiry {
			return invalid
		}
	}
	signers := object["verified_signer_key_ids"].([]any)
	previous := ""
	for _, id := range signers {
		if previous >= id.(string) {
			return invalid
		}
		previous = id.(string)
	}
	versions := object["trusted_metadata_versions"].(map[string]any)
	if object["distribution_kind"] == "packaged_builtin" {
		if object["verification_method"] != "packaged_release_manifest_v1" || object["container_sha256"] != nil || object["trust_repository_id"] != nil || object["trust_valid_until"] != nil || len(signers) != 0 {
			return invalid
		}
		for _, v := range versions {
			if v != nil {
				return invalid
			}
		}
	} else if object["verification_method"] != "tuf_1_0_35_offline_bundle_v1" {
		return invalid
	}
	if !slices.Contains([]string{"import_verification", "reverification", "refresh_verification"}, object["event_kind"].(string)) && object["container_sha256"] != nil {
		return invalid
	}
	id := object["attestation_id"].(string)
	delete(object, "attestation_id")
	preimage, err := canonicaljson.Marshal(object)
	if err != nil || id != "rpa_"+Digest(preimage) {
		return invalid
	}
	return nil
}
