package envelope

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

func Digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func equalHex(left, right string) bool {
	a, ok := hexBytes(left, 32)
	b, other := hexBytes(right, 32)
	return ok && other && subtle.ConstantTimeCompare(a, b) == 1
}
func ValidateInventory(manifest Manifest, inventory Inventory, operator bool) error {
	if err := ValidateInventoryLengths(manifest, inventory, operator); err != nil {
		return err
	}
	return ValidateInventoryHashes(manifest, inventory)
}
func ValidateInventoryLengths(manifest Manifest, inventory Inventory, operator bool) error {
	return runCheckRange(context.Background(), "member_missing", "member_length", inventoryProgram(manifest, inventory, operator, "member_missing", "member_extra", "member_length"))
}
func ValidateInventoryHashes(manifest Manifest, inventory Inventory) error {
	return runCheckRange(context.Background(), "member_hash", "member_hash", inventoryProgram(manifest, inventory, false, "member_hash"))
}
