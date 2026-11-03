package crossownertransaction

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// DescriptorDigest binds the complete closed participant contract. Composition
// compares owner implementations against the independently admitted catalog.
func DescriptorDigest(d Descriptor) string {
	var owner any = d.OwnerProfileID
	if d.OwnerProfileID == "" {
		owner = nil
	}
	encoded, err := canonicaljson.MarshalExtension(map[string]any{
		"schema_id":      "cartulary.extension_transaction_participant_contract.v3",
		"participant_id": d.ParticipantID, "owner_profile_id": owner,
		"participant_input_schema_id": d.InputSchemaID, "prepare_algorithm_id": d.PrepareAlgorithmID,
		"validation_algorithm_id": d.ValidationAlgorithmID, "write_algorithm_id": d.WriteAlgorithmID,
		"serialization_key_kinds": d.SerializationKeyKinds, "owned_state_family_ids": d.OwnedStateFamilyIDs,
	})
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func SameDescriptor(a, b Descriptor) bool {
	return a.ParticipantID == b.ParticipantID && a.OwnerProfileID == b.OwnerProfileID && a.ContractSHA256 == b.ContractSHA256 &&
		a.InputSchemaID == b.InputSchemaID && a.PrepareAlgorithmID == b.PrepareAlgorithmID && a.ValidationAlgorithmID == b.ValidationAlgorithmID && a.WriteAlgorithmID == b.WriteAlgorithmID &&
		slices.Equal(a.SerializationKeyKinds, b.SerializationKeyKinds) && slices.Equal(a.OwnedStateFamilyIDs, b.OwnedStateFamilyIDs)
}
