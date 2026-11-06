package subtypepresence

import (
	"bytes"
	"crypto/sha256"
	"hash"

	"github.com/google/uuid"
)

type inventoryType struct {
	count  uint64
	last   uuid.UUID
	digest hash.Hash
}
type inventoryDigest struct {
	count  uint64
	digest [32]byte
}

// Inventory retains one count and digest per closed record type. Owners emit
// bindings in increasing record-ID order within each type; duplicates fail.
type Inventory struct {
	incidentID uuid.UUID
	types      map[RecordType]*inventoryType
}
type Summary struct {
	incidentID uuid.UUID
	types      map[RecordType]inventoryDigest
}

func NewInventory(incidentID uuid.UUID) *Inventory {
	inventory := &Inventory{incidentID: incidentID, types: map[RecordType]*inventoryType{}}
	for kind := range expectedFamilyByType {
		digest := sha256.New()
		_, _ = digest.Write([]byte("cartulary.records.subtype_inventory.v1\x00" + string(kind) + "\x00"))
		_, _ = digest.Write(incidentID[:])
		inventory.types[kind] = &inventoryType{digest: digest}
	}
	return inventory
}
func (i *Inventory) Add(binding Binding) error {
	if i == nil || i.incidentID == uuid.Nil || binding.IncidentID != i.incidentID || binding.RecordID == uuid.Nil {
		return ErrIncompleteBindings
	}
	state := i.types[binding.RecordType]
	if state == nil || state.count > 0 && bytes.Compare(binding.RecordID[:], state.last[:]) <= 0 {
		return ErrIncompleteBindings
	}
	state.count++
	state.last = binding.RecordID
	_, _ = state.digest.Write(binding.RecordID[:])
	return nil
}
func (i *Inventory) Summary() Summary {
	result := Summary{incidentID: i.incidentID, types: map[RecordType]inventoryDigest{}}
	for kind, state := range i.types {
		var digest [32]byte
		copy(digest[:], state.digest.Sum(nil))
		result.types[kind] = inventoryDigest{count: state.count, digest: digest}
	}
	return result
}
func (s Summary) Equal(other Summary) bool {
	if s.incidentID == uuid.Nil || s.incidentID != other.incidentID || len(s.types) != len(expectedFamilyByType) || len(other.types) != len(expectedFamilyByType) {
		return false
	}
	for kind := range expectedFamilyByType {
		if s.types[kind] != other.types[kind] {
			return false
		}
	}
	return true
}
