package reporting

import "context"

// A reference carries the restrictions of its target, including restrictions
// inherited through other references. This graph contains only frozen identities
// and disclosure labels: neither source text nor renderer labels are inspected.
// Edges point from a dependency to its consumers. Monotone propagation over a
// finite label set terminates even when records cite one another cyclically.
type disclosureNode struct {
	partitions map[string]bool
	lifecycle  lifecycleRestriction
	consumers  map[*disclosureNode]bool
	queued     bool
}

func assignReferencedFieldPartitions(ctx context.Context, model SnapshotContent, fields []ExportField, scope string, states map[string]lifecycleRestriction) ([]ExportField, error) {
	groups, err := snapshotSourceGroups(model)
	if err != nil {
		return nil, err
	}
	invalid := func() error { return newRenderValidationError("export_model_invalid", "dangling_source_ref") }
	newNode := func() *disclosureNode {
		return &disclosureNode{partitions: map[string]bool{}, consumers: map[*disclosureNode]bool{}}
	}
	queue := []*disclosureNode{}
	enqueue := func(node *disclosureNode) {
		if !node.queued {
			node.queued = true
			queue = append(queue, node)
		}
	}
	seed := func(node *disclosureNode, refs []string) error {
		if validateDisclosurePartitionSet(refs) != nil {
			return invalid()
		}
		for _, ref := range refs {
			node.partitions[ref] = true
		}
		enqueue(node)
		return nil
	}
	depend := func(consumer, dependency *disclosureNode) error {
		if consumer == nil || dependency == nil {
			return invalid()
		}
		dependency.consumers[consumer] = true
		return nil
	}
	fieldNodes := map[string]*disclosureNode{}
	for _, field := range fields {
		if field.Path == "" || fieldNodes[field.Path] != nil {
			return nil, invalid()
		}
		node := newNode()
		node.lifecycle = field.lifecycle
		fieldNodes[field.Path] = node
		if err := seed(node, field.DisclosurePartitionRefs); err != nil {
			return nil, err
		}
	}
	// Record envelopes and domain records represent one Core identity. Their
	// aggregate restriction is the union, while each field retains its own label.
	records := map[string]*disclosureNode{}
	for id, state := range states {
		node := newNode()
		node.lifecycle = state
		records[id] = node
		enqueue(node)
	}
	groupNodes := map[sourceOwner]*disclosureNode{}
	targets := map[supportTarget]*disclosureNode{}
	incident := newNode()
	targets[supportTarget{"source_record", "record:incident_metadata:" + model.IncidentID}] = incident
	for _, field := range fields {
		if field.SourceFamily == "incident_metadata" {
			if err := depend(incident, fieldNodes[field.Path]); err != nil {
				return nil, err
			}
		}
	}
	for _, record := range groups {
		family := record.SourceRecordRef.SourceFamily
		core := family == "record_envelope" || coreTypeForSourceFamily(family) != ""
		var node *disclosureNode
		if core {
			node = records[record.RecordID]
		}
		if node == nil {
			node = newNode()
		}
		if core {
			records[record.RecordID] = node
		}
		groupNodes[sourceOwner{family, record.RecordID}] = node
		targets[supportTarget{"source_record", "record:" + record.SourceRecordRef.SourceFamily + ":" + record.RecordID}] = node
		if record.RecordType == "evidence" && record.SourceRecordRef.SourceFamily == "evidence" {
			targets[supportTarget{"evidence_item", "evidence:" + record.RecordID}] = node
		}
		if record.RecordType == "artifact" {
			targets[supportTarget{"artifact_record", "artifact:" + record.RecordID}] = node
		}
		for _, path := range record.SourceRefs {
			if err := depend(node, fieldNodes[path]); err != nil {
				return nil, err
			}
		}
	}
	for _, event := range model.TimelineEvents {
		targets[supportTarget{"timeline_event", "timeline:" + event.RecordID}] = records[event.RecordID]
	}
	for _, relationship := range model.Relationships {
		node := groupNodes[sourceOwner{"record_link", relationship.RelationshipID}]
		targets[supportTarget{"relationship", "relationship:" + relationship.RelationshipID}] = node
		for _, endpoint := range []RelationshipEndpointRef{relationship.SrcRecordRef, relationship.DstRecordRef} {
			ref := endpoint.SourceRecordRef
			target := targets[supportTarget{"source_record", "record:" + ref.SourceFamily + ":" + ref.SourceRecordID}]
			if ref.SourceSnapshotID != model.SnapshotID || target == nil {
				return nil, invalid()
			}
			for _, path := range relationship.SourceRefs {
				if err := depend(fieldNodes[path], target); err != nil {
					return nil, err
				}
			}
		}
	}
	// A rough-capture mention is a subject even though it has no canonical
	// Host/Identity record or Party assignment. It is never implicitly public.
	for _, subject := range model.Subjects {
		if subject.EntityMentionID == nil {
			continue
		}
		partition := "internal_only"
		if scope == ReleaseScopeExternal {
			partition = "blocked"
		}
		for _, path := range subject.SourceRefs {
			node := fieldNodes[path]
			if node == nil {
				return nil, invalid()
			}
			if err := seed(node, []string{partition}); err != nil {
				return nil, err
			}
		}
	}
	// A source field key names a semantic input, which may be a mention list
	// rather than a scalar field in this projection. Until a finer structural
	// split is derived, the entire origin record carries the mention's union.
	// Resolution changes the dependency to an exact canonical subject; it never
	// erases source restrictions or creates a synthetic unresolved identity.
	originFields := map[string][]string{}
	for _, record := range groups {
		if family := record.SourceRecordRef.SourceFamily; family != "record_envelope" && coreTypeForSourceFamily(family) == "" {
			continue
		}
		originFields[record.RecordID] = append(originFields[record.RecordID], record.SourceRefs...)
	}
	for _, record := range groups {
		if record.SourceRecordRef.SourceFamily == "record_tag" {
			for _, field := range record.Fields {
				if field.FieldKey != "record_id" {
					continue
				}
				id, ok := field.Value.(string)
				if !ok {
					return nil, invalid()
				}
				for _, path := range record.SourceRefs {
					if err := depend(fieldNodes[path], records[id]); err != nil {
						return nil, err
					}
				}
			}
		}
		if record.SourceRecordRef.SourceFamily != "entity_mention" {
			continue
		}
		binding, err := reportingMentionSource(record)
		if err != nil {
			return nil, err
		}
		for _, path := range record.SourceRefs {
			if err := depend(fieldNodes[path], records[binding.OriginRecordID]); err != nil {
				return nil, err
			}
			if binding.ResolvedRecordID != nil {
				target := targets[supportTarget{"source_record", "record:" + binding.Family + ":" + *binding.ResolvedRecordID}]
				if err := depend(fieldNodes[path], target); err != nil {
					return nil, err
				}
			}
		}
		for _, path := range originFields[binding.OriginRecordID] {
			if err := depend(fieldNodes[path], groupNodes[sourceOwner{"entity_mention", record.RecordID}]); err != nil {
				return nil, err
			}
		}
	}
	support := map[string]*disclosureNode{}
	for _, ref := range model.SupportIndex {
		node := newNode()
		if support[ref.SupportRefID] != nil || seed(node, ref.DisclosurePartitionRefs) != nil {
			return nil, invalidSupportRef()
		}
		support[ref.SupportRefID] = node
		if err := depend(node, targets[supportTarget{ref.SupportKind, ref.SupportTargetRef}]); err != nil {
			return nil, invalidSupportRef()
		}
	}
	for _, field := range fields {
		for _, ref := range field.SupportRefs {
			if err := depend(fieldNodes[field.Path], support[ref]); err != nil {
				return nil, invalidSupportRef()
			}
		}
	}
	// No graph mutation occurs during propagation. New labels alone enqueue a
	// consumer; revisiting an edge with an already-known label is a no-op.
	for len(queue) != 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		node := queue[0]
		queue[0] = nil
		queue = queue[1:]
		node.queued = false
		for consumer := range node.consumers {
			changed := false
			if node.lifecycle > consumer.lifecycle {
				consumer.lifecycle, changed = node.lifecycle, true
			}
			for ref := range node.partitions {
				if !consumer.partitions[ref] {
					consumer.partitions[ref], changed = true, true
				}
			}
			if changed {
				enqueue(consumer)
			}
		}
	}
	result := make([]ExportField, len(fields))
	for i, field := range fields {
		result[i] = field
		refs := []string{}
		for ref := range fieldNodes[field.Path].partitions {
			refs = append(refs, ref)
		}
		result[i].DisclosurePartitionRefs = orderedDisclosurePartitions(refs)
		result[i].lifecycle = fieldNodes[field.Path].lifecycle
	}
	return result, nil
}
