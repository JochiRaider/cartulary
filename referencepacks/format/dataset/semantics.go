package dataset

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strconv"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
)

type semanticChecks struct {
	path    string
	id      *string
	emit    schema.FindingSink
	err     error
	invalid bool
}

func (c *semanticChecks) require(ok bool, field string) {
	if ok || c.err != nil {
		return
	}
	c.invalid = true
	c.err = c.emit(schema.Finding{Path: c.path + field, EntryID: c.id})
}
func (c *semanticChecks) refs(refs []SourceRef, count int) {
	var previous SourceRef
	for i, ref := range refs {
		path := ".source_refs[" + strconv.Itoa(i) + "]"
		c.require(ref.ArtifactIndex < count, path+".artifact_index")
		c.require(line(ref.Locator), path+".locator")
		c.require(i == 0 || ref.ArtifactIndex > previous.ArtifactIndex || ref.ArtifactIndex == previous.ArtifactIndex && ref.Locator > previous.Locator, path)
		previous = ref
	}
}

type admission struct {
	ctx        context.Context
	descriptor Descriptor
	expected   Expected
	result     *Result
	objects    map[string]string
	keys       map[string]map[string]string
	coverage   map[string]Coverage
	previous   string
	emit       schema.FindingSink
}

func (s *admission) sink() schema.FindingSink {
	if s.emit == nil {
		return firstFailure("content_semantic_invalid")
	}
	return s.emit
}
func (s *admission) object(raw []byte) error {
	ctx := s.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	o, err := decodeObjectContext(ctx, raw)
	if err != nil {
		return err
	}
	c := semanticChecks{path: "$.objects[" + strconv.FormatInt(s.result.Objects, 10) + "]", emit: s.sink(), id: attributable(o.ID)}
	c.require(o.ID > s.previous && line(o.ID), ".object_id")
	c.require(line(o.DisplayLabel), ".display_label")
	c.require(description(o.Description), ".description")
	c.refs(o.SourceRefs, s.expected.SourceArtifacts)
	replacement := ""
	if o.Replacement != nil {
		c.require(o.Deprecated && line(*o.Replacement), ".replacement_object_id")
		replacement = *o.Replacement
	}
	if c.err != nil {
		return c.err
	}
	postings, err := postingFindings(s.descriptor, o, s.result, c.path, func(f schema.Finding) error { f.EntryID = c.id; return s.sink()(f) })
	if err != nil {
		return err
	}
	for _, lookup := range s.descriptor.Lookups {
		coverage := s.coverage[lookup.Name]
		found := false
		for _, posting := range postings {
			if posting.Lookup != lookup.Name {
				continue
			}
			found = true
			_, exists := s.keys[lookup.Name][posting.Key]
			c.require(!exists || !lookup.Unique, ".lookup_values")
			if c.err != nil {
				return c.err
			}
			if !exists {
				coverage.DistinctKeys++
			}
			s.keys[lookup.Name][posting.Key] = o.ID
			coverage.Postings++
		}
		if found {
			coverage.Objects++
		}
		s.coverage[lookup.Name] = coverage
	}
	if _, exists := s.objects[o.ID]; !exists {
		s.objects[o.ID] = replacement
	}
	s.previous = o.ID
	s.result.Objects++
	return nil
}

func objectPostings(d Descriptor, o Object, budget *Result) ([]Posting, error) {
	return postingFindings(d, o, budget, "$", firstFailure("content_semantic_invalid"))
}
func postingFindings(d Descriptor, o Object, budget *Result, path string, emit schema.FindingSink) ([]Posting, error) {
	definitions := make(map[string]Lookup, len(d.Lookups))
	for _, lookup := range d.Lookups {
		definitions[lookup.Name] = lookup
	}
	count := 0
	for _, values := range o.LookupValues {
		count += len(values)
	}
	count = max(count, o.inputPostings)
	if count > MaxObjectPostings {
		return nil, reportFailure(path+".lookup_values", excess("max_dataset_object_postings", MaxObjectPostings, uint64(count), false), emit)
	}
	if budget != nil && int64(count) > MaxPostings-budget.Postings {
		return nil, reportFailure(path+".lookup_values", excess("max_dataset_postings", MaxPostings, uint64(budget.Postings)+uint64(count), false), emit)
	}
	postings := make([]Posting, 0, count)
	for i, name := range slices.Sorted(maps.Keys(o.LookupValues)) {
		p := path + ".lookup_values[" + strconv.Itoa(i) + "]"
		definition, exists := definitions[name]
		if !exists {
			if err := emit(schema.Finding{Path: p}); err != nil {
				return nil, err
			}
			continue
		}
		previous := ""
		for j, raw := range o.LookupValues[name] {
			p := p + "[" + strconv.Itoa(j) + "]"
			value, err := canonicaljson.DecodeStrictBounded(raw, 2)
			if err != nil {
				return nil, invalid()
			}
			remaining := int64(MaxTotalKeyBytes)
			if budget != nil {
				remaining -= budget.EncodedKeyBytes
			}
			key, err := normalizeBudget(definition, value, remaining)
			if err != nil {
				if err = reportFailure(p, err, emit); err != nil {
					return nil, err
				}
				continue
			}
			if key <= previous {
				if err := emit(schema.Finding{Path: p}); err != nil {
					return nil, err
				}
			}
			previous = key
			if budget != nil {
				budget.Postings++
				budget.EncodedKeyBytes += int64(len(key))
			}
			postings = append(postings, Posting{Lookup: name, Key: key})
		}
	}
	return postings, nil
}

func (s *admission) finishObjects(ctx context.Context) error {
	for _, id := range slices.Sorted(maps.Keys(s.objects)) {
		if err := ctx.Err(); err != nil {
			return err
		}
		current := id
		visited := map[string]bool{}
		for count := 0; current != ""; count++ {
			next, exists := s.objects[current]
			if count == 32 || visited[current] || !exists {
				if err := s.sink()(schema.Finding{Path: "$.objects", EntryID: attributable(id)}); err != nil {
					return err
				}
				break
			}
			visited[current] = true
			current = next
		}
	}
	for i, lookup := range s.descriptor.Lookups {
		if s.coverage[lookup.Name] != lookup.Coverage {
			if err := s.sink()(schema.Finding{Path: "$.descriptor.lookups[" + strconv.Itoa(i) + "].coverage"}); err != nil {
				return err
			}
		}
	}
	return nil
}

// SemanticsFindings requires a successful immutable structural pass. It visits
// every applicable row and reference and emits no source-specific assertions.
func SemanticsFindings(ctx context.Context, source Source, expected Expected, emit schema.FindingSink) error {
	if source == nil || emit == nil {
		return errors.New("reference dataset: incomplete semantic check")
	}
	_, err := validateSemantics(ctx, source, expected, emit)
	return guardResult(err)
}
func validateSemantics(ctx context.Context, source Source, expected Expected, emit schema.FindingSink) (Result, error) {
	raw, err := descriptorBytes(ctx, source)
	if err != nil {
		return Result{}, err
	}
	d, err := decodeDescriptor(raw, false)
	if err != nil {
		return Result{}, err
	}
	valid := true
	if err := descriptorRules(d, true, func(f schema.Finding) error { valid = false; return emit(f) }); err != nil {
		return Result{}, err
	}
	// Duplicate definitions have no unambiguous normalization or coverage.
	if !valid {
		return Result{}, nil
	}
	result := Result{Descriptor: d, DescriptorBytes: raw}
	s := admission{ctx: ctx, descriptor: d, expected: expected, result: &result, objects: map[string]string{}, keys: map[string]map[string]string{}, coverage: map[string]Coverage{}, emit: emit}
	for _, lookup := range d.Lookups {
		s.keys[lookup.Name] = map[string]string{}
	}
	if err := rows(ctx, source, ObjectsPath, expected.Objects, MaxObjects, s.object); err != nil {
		return Result{}, err
	}
	if err := s.finishObjects(ctx); err != nil {
		return Result{}, err
	}
	seen := map[string]bool{}
	var previous [4]string
	err = rows(ctx, source, RelationshipsPath, expected.Relationships, MaxRelationships, func(raw []byte) error {
		rel, err := decodeRelationshipContext(ctx, raw)
		if err != nil {
			return err
		}
		c := semanticChecks{path: "$.relationships[" + strconv.FormatInt(result.Relationships, 10) + "]", emit: emit, id: attributable(rel.ID)}
		order := [4]string{rel.Type, rel.Source, rel.Target, rel.ID}
		c.require(!seen[rel.ID] && line(rel.ID), ".relationship_id")
		_, sourceExists := s.objects[rel.Source]
		_, targetExists := s.objects[rel.Target]
		c.require(sourceExists && line(rel.Source), ".source_object_id")
		c.require(targetExists && line(rel.Target), ".target_object_id")
		c.refs(rel.SourceRefs, expected.SourceArtifacts)
		c.require(result.Relationships == 0 || slices.Compare(order[:], previous[:]) > 0, "")
		seen[rel.ID] = true
		previous = order
		result.Relationships++
		return c.err
	})
	return result, err
}

func attributable(id string) *string {
	if len(id) <= 512 && line(id) {
		return &id
	}
	return nil
}
