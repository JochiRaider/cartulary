package dataset

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"io"
	"slices"
	"strings"
)

type encodedMembers map[string][]byte

func (m encodedMembers) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	b, ok := m[name]
	if !ok {
		return nil, ErrStorage
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

// Encode is a producer-side serializer for already projected records. It sorts
// identities and lookup values, calculates actual normalized coverage, and runs
// admission on the completed bytes. It never derives facts or drops duplicates.
func Encode(ctx context.Context, lookups []Lookup, objects []Object, relationships []Relationship, sourceArtifacts int) (map[string][]byte, Result, error) {
	var empty Result
	if len(lookups) > MaxLookups || len(objects) > MaxObjects || len(relationships) > MaxRelationships || sourceArtifacts < 1 || sourceArtifacts > 64 {
		return nil, empty, invalid()
	}
	descriptor := Descriptor{SchemaID: DescriptorSchema, Lookups: slices.Clone(lookups)}
	if descriptor.Lookups == nil {
		descriptor.Lookups = []Lookup{}
	}
	slices.SortFunc(descriptor.Lookups, func(a, b Lookup) int { return strings.Compare(a.Name, b.Name) })
	for i := range descriptor.Lookups {
		descriptor.Lookups[i].Coverage = Coverage{}
	}
	raw, err := canonicaljson.Marshal(descriptor)
	if err != nil {
		return nil, empty, err
	}
	if _, err = DecodeDescriptor(raw); err != nil {
		return nil, empty, err
	}
	ordered := slices.Clone(objects)
	edges := slices.Clone(relationships)
	slices.SortFunc(ordered, func(a, b Object) int { return strings.Compare(a.ID, b.ID) })
	slices.SortFunc(edges, func(a, b Relationship) int {
		return slices.Compare([]string{a.Type, a.Source, a.Target, a.ID}, []string{b.Type, b.Source, b.Target, b.ID})
	})
	expected := Expected{Objects: int64(len(objects)), Relationships: int64(len(relationships)), SourceArtifacts: sourceArtifacts}
	result := Result{}
	state := admission{ctx: ctx, descriptor: descriptor, expected: expected, result: &result, objects: map[string]string{}, keys: map[string]map[string]string{}, coverage: map[string]Coverage{}}
	for _, lookup := range descriptor.Lookups {
		state.keys[lookup.Name] = map[string]string{}
	}
	members := encodedMembers{ObjectsPath: {}, RelationshipsPath: {}}
	var total int64
	appendRow := func(path string, raw []byte) error {
		if len(raw)+1 > MaxRowBytes {
			return excess("max_line_bytes", MaxRowBytes, uint64(len(raw)+1), true)
		}
		if int64(len(raw)+1) > 536870912-total {
			return errors.New("reference dataset: encoded payload size limit")
		}
		total += int64(len(raw) + 1)
		members[path] = append(members[path], raw...)
		members[path] = append(members[path], '\n')
		return nil
	}
	for _, object := range ordered {
		if err := ctx.Err(); err != nil {
			return nil, empty, err
		}
		values := map[string][]json.RawMessage{}
		count := 0
		for _, inputs := range object.LookupValues {
			count += len(inputs)
			if count > MaxObjectPostings {
				return nil, empty, excess("max_dataset_object_postings", MaxObjectPostings, uint64(count), false)
			}
		}
		for name, input := range object.LookupValues {
			values[name] = slices.Clone(input)
		}
		for _, lookup := range descriptor.Lookups {
			inputs, exists := values[lookup.Name]
			if !exists {
				continue
			}
			inputs, err = CanonicalInputs(lookup, inputs)
			if err != nil {
				return nil, empty, err
			}
			values[lookup.Name] = inputs
		}
		object.LookupValues = values
		raw, err := canonicaljson.Marshal(object)
		if err != nil {
			return nil, empty, err
		}
		if len(raw)+1 > MaxRowBytes {
			return nil, empty, excess("max_line_bytes", MaxRowBytes, uint64(len(raw)+1), true)
		}
		if err = state.object(raw); err != nil {
			return nil, empty, err
		}
		if err = appendRow(ObjectsPath, raw); err != nil {
			return nil, empty, err
		}
	}
	for i, lookup := range descriptor.Lookups {
		descriptor.Lookups[i].Coverage = state.coverage[lookup.Name]
	}
	raw, err = canonicaljson.Marshal(descriptor)
	if err != nil {
		return nil, empty, err
	}
	members[DescriptorPath] = raw
	for _, edge := range edges {
		if err := ctx.Err(); err != nil {
			return nil, empty, err
		}
		raw, err := canonicaljson.Marshal(edge)
		if err != nil {
			return nil, empty, err
		}
		if err = appendRow(RelationshipsPath, raw); err != nil {
			return nil, empty, err
		}
	}
	result, err = Validate(ctx, members, expected, nil)
	if err != nil {
		return nil, empty, err
	}
	return members, result, nil
}
