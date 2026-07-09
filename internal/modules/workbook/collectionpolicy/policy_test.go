package collectionpolicy

import (
	"reflect"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestRegistryCoversWritableCollectionReviewFields(t *testing.T) {
	want := make(map[string]struct{})
	for _, resource := range viewschema.ListPublicResources() {
		schema, ok := viewschema.Lookup(resource.ViewSchemaID)
		if !ok {
			t.Fatalf("schema %s missing from registry", resource.ViewSchemaID)
		}
		for fieldKey, field := range schema.Fields() {
			if field.Writable && field.ConflictResolutionClass == "collection_review" {
				want[fieldKey] = struct{}{}
			}
		}
	}

	got := make(map[string]struct{})
	for _, policy := range All() {
		got[policy.FieldKey] = struct{}{}
		if len(policy.AllowedOps) == 0 {
			t.Fatalf("policy %s has no allowed operations", policy.FieldKey)
		}
		if policy.Owner == OwnerLinks && policy.ItemFamily != ItemFamilyRecordTag && policy.LinkType == "" {
			t.Fatalf("links policy %s has no link type", policy.FieldKey)
		}
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collection policy registry drift\ngot:  %v\nwant: %v", FieldKeys(), sortedKeys(want))
	}
}

func TestRegistryReturnsDefensiveCopies(t *testing.T) {
	policy := MustLookup("timeline.attached_evidence_ids")
	policy.ChangedFieldKeys[0] = "mutated"
	policy.AllowedOps[0] = "mutated"

	next := MustLookup("timeline.attached_evidence_ids")
	if next.ChangedFieldKeys[0] != "timeline.attached_evidence_ids" {
		t.Fatalf("changed field keys were not defensively copied: %v", next.ChangedFieldKeys)
	}
	if next.AllowedOps[0] != "add_record_ref" {
		t.Fatalf("allowed ops were not defensively copied: %v", next.AllowedOps)
	}
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
