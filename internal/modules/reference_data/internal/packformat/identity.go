package packformat

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
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

// These two boundaries let the coordinator check both the admitted container
// and retained published members before advancing to the next registry rank.
func ValidateInventoryLengths(manifest Manifest, inventory Inventory, operator bool) error {
	return runCheckRange(context.Background(), "member_missing", "member_length", nil, inventoryProgram(manifest, inventory, operator, "member_missing", "member_extra", "member_length"), false)
}

func ValidateInventoryHashes(manifest Manifest, inventory Inventory) error {
	return runCheckRange(context.Background(), "member_hash", "member_hash", nil, inventoryProgram(manifest, inventory, false, "member_hash"), false)
}

type SetMember struct {
	Key            string `json:"pack_key"`
	Version        string `json:"pack_version"`
	ManifestSHA256 string `json:"manifest_sha256"`
	PayloadSHA256  string `json:"payload_sha256"`
	Contract       string `json:"pack_contract_version"`
	ProfileID      string `json:"content_profile_id"`
	ProfileVersion string `json:"content_profile_version"`
}
type Set struct {
	SchemaID string      `json:"schema_id"`
	ID       string      `json:"pack_set_id"`
	SHA256   string      `json:"pack_set_sha256"`
	Members  []SetMember `json:"members"`
}

func admittedSetCardinality(count int) bool { return count >= 3 && count <= 64 }

func BuildSet(members []SetMember) (Set, error) {
	if !admittedSetCardinality(len(members)) {
		return Set{}, fail("contract_incompatible")
	}
	members = slices.Clone(members)
	slices.SortFunc(members, func(a, b SetMember) int { return strings.Compare(a.Key, b.Key) })
	seen := map[string]bool{}
	for _, member := range members {
		p, ok := profiles[member.Key]
		_, manifestOK := hexBytes(member.ManifestSHA256, 32)
		_, payloadOK := hexBytes(member.PayloadSHA256, 32)
		if !ok || seen[member.Key] || !versionPattern.MatchString(member.Version) || !manifestOK || !payloadOK || member.Contract != ContractID || member.ProfileID != p.ID || member.ProfileVersion != p.Version {
			return Set{}, fail("contract_incompatible")
		}
		seen[member.Key] = true
	}
	for key, p := range profiles {
		if p.Required && !seen[key] {
			return Set{}, fail("required_registry_gap")
		}
	}
	preimage, err := canonicaljson.Marshal(struct {
		Schema  string      `json:"schema_id"`
		Members []SetMember `json:"members"`
	}{"cartulary.reference_pack_set.v1", members})
	if err != nil {
		return Set{}, err
	}
	digest := Digest(preimage)
	return Set{SchemaID: "cartulary.reference_pack_set.v1", ID: "rpset_" + digest, SHA256: digest, Members: members}, nil
}

// SameContent never compares mutable verification or administrative fields.
func SameContent(a, b Set) bool {
	left, e1 := canonicaljson.Marshal(a.Members)
	right, e2 := canonicaljson.Marshal(b.Members)
	return e1 == nil && e2 == nil && bytes.Equal(left, right)
}
