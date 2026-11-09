package packformat

import (
	"context"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
)

func InventoryFindings(check string, m Manifest, inventory Inventory, operator bool, emit FindingSink) error {
	return formatError(envelope.InventoryFindings(check, m, inventory, operator, emit))
}
func inventoryProgram(m Manifest, inventory Inventory, operator bool, ids ...string) map[string]CheckFunc {
	program := map[string]CheckFunc{}
	for _, id := range ids {
		program[id] = func(_ context.Context, emit FindingSink) error {
			return InventoryFindings(id, m, inventory, operator, emit)
		}
	}
	return program
}

func TrustBindingFindings(m Manifest, manifestSHA, payloadSHA string, proposal TrustProposal, emit FindingSink) error {
	return formatError(envelope.TrustBindingFindings(m, manifestSHA, payloadSHA, proposal, emit))
}
