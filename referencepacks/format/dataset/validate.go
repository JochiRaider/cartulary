package dataset

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

// ValidateStructure completes the descriptor and row schema pass. It does not
// interpret source-specific meaning or write an index.
func ValidateStructure(ctx context.Context, source Source, expected Expected) (Result, error) {
	if err := StructureFindings(ctx, source, expected, firstFailure("content_schema_invalid")); err != nil {
		return Result{}, err
	}
	raw, err := descriptorBytes(ctx, source)
	if err != nil {
		return Result{}, err
	}
	d, err := decodeDescriptor(raw, false)
	return Result{Descriptor: d, DescriptorBytes: raw}, err
}

// Validate performs complete structural and semantic passes before writing a
// private index. Source must be immutable for every pass; callers discard any
// private writes on failure and own atomic publication after success.
func Validate(ctx context.Context, source Source, expected Expected, sink Sink) (Result, error) {
	result, err := ValidateStructure(ctx, source, expected)
	if err != nil {
		return Result{}, err
	}
	result, err = validateSemantics(ctx, source, expected, firstFailure("content_semantic_invalid"))
	if err != nil {
		return Result{}, err
	}
	descriptor := result.Descriptor
	if sink != nil {
		if err := rows(ctx, source, ObjectsPath, expected.Objects, MaxObjects, func(raw []byte) error {
			object, err := decodeObjectContext(ctx, raw)
			if err != nil {
				return err
			}
			postings, err := objectPostings(descriptor, object, nil)
			if err != nil {
				return err
			}
			return sink.Append(ctx, Row{ID: object.ID, Shape: "object", Canonical: slices.Clone(raw), Postings: postings})
		}); err != nil {
			return Result{}, err
		}
		if err := rows(ctx, source, RelationshipsPath, expected.Relationships, MaxRelationships, func(raw []byte) error {
			rel, err := decodeRelationshipContext(ctx, raw)
			if err != nil {
				return err
			}
			return sink.Append(ctx, Row{ID: rel.ID, Shape: "relationship", Canonical: slices.Clone(raw), Postings: []Posting{}})
		}); err != nil {
			return Result{}, err
		}
	}
	return result, ctx.Err()
}

// CanonicalInputs helps producers materialize postings in the required order.
// It rejects duplicates instead of deciding whether source facts may coalesce.
func CanonicalInputs(lookup Lookup, inputs []json.RawMessage) ([]json.RawMessage, error) {
	type keyed struct {
		key string
		raw json.RawMessage
	}
	if len(inputs) > MaxObjectPostings {
		return nil, excess("max_dataset_object_postings", MaxObjectPostings, uint64(len(inputs)), false)
	}
	values := make([]keyed, 0, len(inputs))
	for _, raw := range inputs {
		if len(raw) > MaxRowBytes {
			return nil, invalid()
		}
		v, err := canonicaljson.DecodeStrictBounded(raw, 2)
		if err != nil {
			return nil, invalid()
		}
		key, err := normalize(lookup, v)
		if err != nil {
			return nil, err
		}
		canonical, err := canonicaljson.Canonicalize(raw)
		if err != nil {
			return nil, invalid()
		}
		values = append(values, keyed{key, canonical})
	}
	slices.SortFunc(values, func(a, b keyed) int { return strings.Compare(a.key, b.key) })
	output := make([]json.RawMessage, 0, len(values))
	for i, value := range values {
		if i > 0 && values[i-1].key == value.key {
			return nil, semantic()
		}
		output = append(output, value.raw)
	}
	return output, nil
}
