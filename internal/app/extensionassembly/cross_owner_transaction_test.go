package extensionassembly

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
)

func TestCrossOwnerDescriptorsUsesCurrentMachineContract(t *testing.T) {
	current := extensions.ParticipantContract{
		ParticipantID:         "current.participant",
		OwnerProfileID:        "current",
		ContractSHA256:        "current-sha256",
		ContractKind:          "cartulary.extension_transaction_participant_contract.v2",
		InputSchemaID:         "current.input.v1",
		PrepareAlgorithmID:    "current.prepare.v1",
		ValidationAlgorithmID: "current.validate.v1",
		WriteAlgorithmID:      "current.write.v1",
		SerializationKeyKinds: []string{"current.key"},
		OwnedStateFamilyIDs:   []string{"current.rows"},
	}
	legacy := current
	legacy.ParticipantID = "legacy.participant"
	legacy.ContractKind = "cartulary.extension_transaction_participant_contract.v1"

	descriptors := CrossOwnerDescriptors([]extensions.ParticipantContract{legacy, current})
	var foundCurrent bool
	for _, descriptor := range descriptors {
		switch descriptor.ParticipantID {
		case current.ParticipantID:
			foundCurrent = true
			if descriptor.OwnerProfileID != current.OwnerProfileID ||
				descriptor.ContractSHA256 != current.ContractSHA256 ||
				descriptor.InputSchemaID != current.InputSchemaID {
				t.Fatalf("current descriptor = %#v", descriptor)
			}
		case legacy.ParticipantID:
			t.Fatal("legacy documentation-bearing contract shape was admitted")
		}
	}
	if !foundCurrent {
		t.Fatal("current machine contract was not admitted")
	}
}
