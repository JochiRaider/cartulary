package reporting

import "testing"

func testSnapshotSectionStructure(t *testing.T) {
	for name, change := range map[string]func(*ExportModel){
		"section identity":       func(m *ExportModel) { m.Sections[0].SectionID = "other" },
		"section order":          func(m *ExportModel) { m.Sections[0].OrderingKey = "0000" },
		"missing section order":  func(m *ExportModel) { m.Sections[0].OrderingKey = "" },
		"section kind":           func(m *ExportModel) { m.Sections[0].SectionKind = "future" },
		"section title controls": func(m *ExportModel) { m.Sections[0].Title = "bad\nline" },
		"section references":     func(m *ExportModel) { m.Sections[0].SourceRefs = []string{} },
		"section restrictions":   func(m *ExportModel) { m.Sections[0].DisclosurePartitionRefs = []string{"blocked"} },
		"section validation":     func(m *ExportModel) { m.Sections[0].SectionValidation.IssueCount = 1 },
		"missing blocks":         func(m *ExportModel) { m.Sections[0].Blocks = nil },
		"missing children":       func(m *ExportModel) { m.Sections[0].Blocks[0].Children = nil },
		"block identity":         func(m *ExportModel) { m.Sections[0].Blocks[0].BlockID = "other" },
		"ordinal gap":            func(m *ExportModel) { m.Sections[0].Blocks[0].BlockOrdinal = 2 },
		"wrong parent":           func(m *ExportModel) { id := "parent"; m.Sections[0].Blocks[0].ParentBlockID = &id },
		"duplicate container":    func(m *ExportModel) { m.Sections[0].Blocks = append(m.Sections[0].Blocks, m.Sections[0].Blocks[0]) },
		"block kind":             func(m *ExportModel) { m.Sections[0].Blocks[0].BlockKind = "future" },
		"block source refs":      func(m *ExportModel) { m.Sections[0].Blocks[0].SourceRefs = []string{} },
		"block restrictions":     func(m *ExportModel) { m.Sections[0].Blocks[0].DisclosurePartitionRefs = []string{"blocked"} },
	} {
		t.Run(name, func(t *testing.T) {
			model := fixtureStructuredRenderModel(t, []ExportField{{Path: "/incident/a", ContentClass: ContentClassSourceEvidence, Value: "source", DisclosurePartitionRefs: []string{"public"}}})
			change(&model)
			if err := validateSnapshotSectionStructure(model.SnapshotContent); err == nil {
				t.Fatal("invalid frozen hierarchy admitted")
			}
			if fields, err := model.RedactionFields(); err == nil || fields != nil {
				t.Fatal("invalid hierarchy produced a partial field view")
			}
			raw, err := canonicalJSON(model)
			if err != nil {
				t.Fatal(err)
			}
			decoded, _, err := decodePortableObject[ExportModel](raw, reportingObjectDigest(model.SchemaID, raw))
			if err == nil && validateRenderModelIdentity(decoded) == nil {
				t.Fatal("rehashed invalid hierarchy admitted")
			}
		})
	}
}
