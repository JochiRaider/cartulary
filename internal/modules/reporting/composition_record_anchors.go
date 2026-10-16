package reporting

// This is a private lookup over admitted provenance, not an alternative record
// identity. Auxiliary sources never become Core records through a shared ID.
type deckRecordAnchors struct {
	counts  map[string]int
	sources map[string]string
}

func newDeckRecordAnchors(records []ReportingRecordSummary, sources []ReportingSourceRef) *deckRecordAnchors {
	index := &deckRecordAnchors{counts: map[string]int{}, sources: map[string]string{}}
	for _, record := range records {
		index.counts[record.RecordID]++
	}
	for _, source := range sources {
		if source.SourcePath != nil && index.counts[source.SourceRecordID] != 0 &&
			(source.SourceFamily == "record_envelope" || coreTypeForSourceFamily(source.SourceFamily) != "") {
			index.sources[source.SourceRefID] = source.SourceRecordID
		}
	}
	return index
}

func (index *deckRecordAnchors) resolve(anchor *compositionRecordAnchor) error {
	if anchor.AnchorKind != "record_anchor" || anchor.RecordID == "" {
		return compositionFailure("invalid_scalar_value")
	}
	if index == nil || index.counts[anchor.RecordID] == 0 {
		return compositionFailure("composition_anchor_unresolved")
	}
	if index.counts[anchor.RecordID] != 1 {
		return compositionFailure("composition_anchor_ambiguous")
	}
	return nil
}

// Use the current derived subtree, not cached aggregate references: an earlier
// operation may remove its only matching child. Authored insertion cannot
// manufacture a derived target, including through a retained parent.
func (index *deckRecordAnchors) blockContains(block ReportingBlock, recordID string, derived map[string]bool) bool {
	if !derived[block.BlockID] {
		return false
	}
	for _, field := range block.Fields {
		for _, ref := range field.SourceRefs {
			if index.sources[ref] == recordID {
				return true
			}
		}
	}
	for _, child := range block.Children {
		if index.blockContains(child, recordID, derived) {
			return true
		}
	}
	return false
}
