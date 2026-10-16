package reporting

import (
	"bytes"
	"encoding/json"
	"testing"
)

func testSnapshotSourceIndex(t *testing.T) {
	path := "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['display_name']"
	id, err := sourceReferenceID(fixtureSnapshotID, sourceOwner{"host", "host-1"}, &path)
	if err != nil || id != "src_f1437768faa0dd83324de53430dcb4b965d0ea0a9f8c86b37aec2d13dfd255cf" {
		t.Fatal("independent canonical source identity vector differs", id, err)
	}
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/private-selector/a", SourceFamily: "host", SourceRecordID: "host-1", SourceRecordType: "host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Host"}, DisclosurePartitionRefs: []string{"public"}},
		{Path: "/private-selector/b", SourceFamily: "evidence", SourceRecordID: "evidence-1", SourceRecordType: "evidence", ContentClass: ContentClassSourceEvidence, Value: "Citation", SupportRefs: []string{"/hosts/host-1"}},
		{Path: "/not-an-incident-prefix/title", SourceFamily: "incident_metadata", ContentClass: ContentClassDerivedAnalytic, Value: "Incident"},
	})
	encoded, err := canonicalJSON(model)
	if err != nil || bytes.Contains(encoded, []byte("/private-selector/")) || bytes.Contains(encoded, []byte("/not-an-incident-prefix/")) || len(model.SourceIndex) != 6 {
		t.Fatal("private owner selectors survived canonical materialization", err)
	}
	index := snapshotSourceIndex(model.SnapshotContent)
	support := model.SupportIndex[0]
	if support.SourceRefID == nil || index[*support.SourceRefID].SourceRecordID != "host-1" || index[*support.SourceRefID].SourcePath != nil || len(support.DisclosurePartitionRefs) != 1 || support.DisclosurePartitionRefs[0] != "public" {
		t.Fatal("support lost its explicit source target or disclosure")
	}
	fields, err := model.RedactionFields()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, field := range fields {
		if field.Value == "Incident" {
			found = field.SourceFamily == "incident_metadata"
		}
	}
	if !found {
		t.Fatal("incident provenance depended on selector spelling")
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	object["source_index"].([]any)[0].(map[string]any)["private_selector"] = "/unadmitted"
	unknown, err := canonicalJSON(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := decodePortableObject[ExportModel](unknown, reportingObjectDigest(ExportModelSchemaID, unknown)); err == nil {
		t.Fatal("portable reader accepted unknown source-reference member")
	}
	for name, change := range map[string]func(*SnapshotContent){
		"absent index":          func(m *SnapshotContent) { m.SourceIndex = nil },
		"duplicate index":       func(m *SnapshotContent) { m.SourceIndex = append(m.SourceIndex, m.SourceIndex[0]) },
		"retired identity":      func(m *SnapshotContent) { m.SourceIndex[0].SourceRefID = "/hosts/host-1" },
		"foreign snapshot":      func(m *SnapshotContent) { m.SourceIndex[0].SourceSnapshotID = "other" },
		"foreign owner":         func(m *SnapshotContent) { m.SourceIndex[0].SourceRecordID = "other" },
		"unsafe summary":        func(m *SnapshotContent) { value := "raw value"; m.SourceIndex[0].SourceSummary = &value },
		"altered partition":     func(m *SnapshotContent) { m.SourceIndex[0].DisclosurePartitionRefs = []string{"blocked"} },
		"missing source anchor": func(m *SnapshotContent) { m.Records[0].SourceRecordRef.SourceRefID = nil },
		"wrong source anchor": func(m *SnapshotContent) {
			id := m.Records[1].SourceRefs[0]
			m.Records[0].SourceRecordRef.SourceRefID = &id
		},
		"missing support source":  func(m *SnapshotContent) { m.SupportIndex[0].SourceRefID = nil },
		"wrong support partition": func(m *SnapshotContent) { m.SupportIndex[0].DisclosurePartitionRefs = []string{"blocked"} },
		"missing field address": func(m *SnapshotContent) {
			for i := range m.SourceIndex {
				if m.SourceIndex[i].SourcePath != nil {
					m.SourceIndex[i].SourcePath = nil
					return
				}
			}
		},
		"moved field address": func(m *SnapshotContent) { m.Sections[0].Blocks[0].BlockID = "changed" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate, err := copyReportingModel(model.SnapshotContent)
			if err != nil {
				t.Fatal(err)
			}
			change(&candidate)
			if validateSnapshotContent(candidate) == nil {
				t.Fatal("altered provenance admitted")
			}
			if values, err := candidate.RedactionFields(); err == nil || values != nil {
				t.Fatal("altered provenance produced redaction inputs", err)
			}
			if result, _, err := bindRenderModel(SnapshotModel{SchemaID: SnapshotModelSchemaID, SnapshotModelID: model.SnapshotModelID, SourceFacts: fixtureSourceFactsForContent(candidate), SnapshotContent: candidate}, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil); err == nil || result.SchemaID != "" {
				t.Fatal("altered provenance admitted to render", err)
			}
		})
	}
}
