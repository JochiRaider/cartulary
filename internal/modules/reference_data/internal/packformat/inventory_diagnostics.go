package packformat

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
)

// InventoryFindings uses only admitted manifest ordinals and virtual sorted
// archive ordinals. An undeclared hostile name can never become a diagnostic.
// The coordinator combines this stream with retained-object findings at the
// same check rank, before proceeding to any later comparison.
func InventoryFindings(check string, m Manifest, inventory Inventory, operator bool, emit FindingSink) error {
	if emit == nil {
		return errors.New("reference pack: inventory finding sink required")
	}
	declared := map[string]bool{"manifest.json": true}
	fixed := []string{"manifest.json"}
	if operator {
		fixed = append(fixed, "bundle.json", "metadata/timestamp.json", "metadata/snapshot.json", "metadata/targets.json")
	}
	for _, path := range fixed {
		declared[path] = true
	}
	for _, file := range m.Files {
		declared[file.Path] = true
	}
	switch check {
	case "member_missing":
		for _, path := range fixed {
			if _, exists := inventory[path]; !exists {
				if err := emit(Finding{Path: path}); err != nil {
					return err
				}
			}
		}
		for i, file := range m.Files {
			if _, exists := inventory[file.Path]; !exists {
				if err := emit(Finding{Path: "$.files[" + strconv.Itoa(i) + "]"}); err != nil {
					return err
				}
			}
		}
	case "member_extra":
		paths := make([]string, 0, len(inventory))
		for path := range inventory {
			paths = append(paths, path)
		}
		slices.Sort(paths)
		for i, path := range paths {
			if !declared[path] && !(operator && strings.HasPrefix(path, "metadata/") && strings.HasSuffix(path, ".root.json")) {
				if err := emit(Finding{Path: "$.archive_members[" + strconv.Itoa(i) + "]"}); err != nil {
					return err
				}
			}
		}
	case "member_length", "member_hash":
		for i, file := range m.Files {
			actual, exists := inventory[file.Path]
			if !exists {
				return errors.New("reference pack: inventory comparison before presence admission")
			}
			field := "size_bytes"
			bad := actual.Size != file.Size
			if check == "member_hash" {
				field = "sha256"
				bad = !equalHex(actual.SHA256, file.SHA256)
			}
			if bad {
				if err := emit(Finding{Path: "$.files[" + strconv.Itoa(i) + "]." + field}); err != nil {
					return err
				}
			}
		}
	default:
		return errors.New("reference pack: unknown inventory check")
	}
	return nil
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

// TrustBindingFindings compares each independent signed identity component.
// Values remain private; the declared manifest field identifies each mismatch.
func TrustBindingFindings(m Manifest, manifestSHA, payloadSHA string, proposal TrustProposal, emit FindingSink) error {
	if emit == nil {
		return errors.New("reference pack: binding finding sink required")
	}
	b := proposal.Binding
	sequence, valid := integer(b["pack_release_sequence"], 1, 9007199254740991)
	for _, test := range []struct {
		field   string
		matches bool
	}{
		{"trust_repository_id", m.Repository != nil && b["trust_repository_id"] == *m.Repository},
		{"pack_key", b["pack_key"] == m.Key},
		{"pack_version", b["pack_version"] == m.Version},
		{"pack_release_sequence", valid && sequence == m.Sequence},
		{"manifest_sha256", b["manifest_sha256"] == manifestSHA},
		{"payload_sha256", b["payload_sha256"] == payloadSHA},
	} {
		if !test.matches {
			if err := emit(Finding{Path: "$.binding." + test.field}); err != nil {
				return err
			}
		}
	}
	return nil
}
