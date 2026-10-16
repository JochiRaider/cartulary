package reporting

import (
	"errors"
	"sort"
	"testing"
	"time"
)

func testSnapshotSupportClosure(t *testing.T, model ExportModel, source []ExportField) {
	// Independently fixed canonical tuple/domain vector; do not derive the
	// expected identity through the production serializer or hashing helper.
	const want = "sup_e0c27e396a64af9f61ceca39b9c1cf98f416a9c7a9b02af1d686ceed4da65b89"
	if len(model.SupportIndex) != 1 || model.SupportIndex[0].SupportRefID != want {
		t.Fatal("support identity does not match the registered canonical vector")
	}
	if got := model.Records[1].Fields[0].SupportRefs; len(got) != 1 || got[0] != want {
		t.Fatal("retained field did not bind its indexed support identity", got)
	}
	id := "00000000-0000-0000-0000-000000000499"
	snapshot := SnapshotModel{SourceFacts: fixtureSourceFactsForContent(model.SnapshotContent), SchemaID: SnapshotModelSchemaID, SnapshotModelID: model.SnapshotModelID, SnapshotContent: model.SnapshotContent}
	preview, _, err := bindRenderModel(snapshot, RenderIdentity{PreviewAttemptID: &id, RenderAdmittedAt: model.RenderAdmittedAt}, ReleaseScopeInternalDraft, nil)
	if err != nil || preview.SupportIndex[0].SupportRefID != want {
		t.Fatal("preview changed snapshot support identity", err)
	}
	otherSnapshotID, err := supportReferenceID("00000000-0000-0000-0000-000000000498", supportTarget{"evidence_item", "evidence:metadata"})
	if err != nil || otherSnapshotID == want {
		t.Fatal("support identity lost its snapshot binding", err)
	}
	for _, selector := range []string{"/evidence/missing", "/timeline/metadata", "/relationships/metadata", "/hosts/metadata"} {
		candidate := cloneExportFields(source)
		candidate[0].SupportRefs = []string{selector}
		value, digest, err := buildStructuredSnapshotModel(model.IncidentID, fixtureSnapshotID, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), "boundary", fixtureSnapshotSource(fixtureOwnerFields(candidate)))
		var invalid *renderValidationError
		if !errors.As(err, &invalid) || invalid.ReasonCode != "invalid_support_ref" || value.SnapshotID != "" || digest != "" {
			t.Fatalf("absent/wrong-family target %q published a snapshot: %v", selector, err)
		}
	}
	mutations := map[string]func(*SnapshotContent){
		"missing index":              func(m *SnapshotContent) { m.SupportIndex = nil },
		"duplicate index identity":   func(m *SnapshotContent) { m.SupportIndex = append(m.SupportIndex, m.SupportIndex[0]) },
		"retired truncated identity": func(m *SnapshotContent) { m.SupportIndex[0].SupportRefID = "sup_0123456789abcdef01234567" },
		"foreign snapshot":           func(m *SnapshotContent) { m.SupportIndex[0].SourceSnapshotID = "another-snapshot" },
		"altered target":             func(m *SnapshotContent) { m.SupportIndex[0].SupportTargetRef = "evidence:source" },
		"missing target record":      func(m *SnapshotContent) { m.Records = m.Records[1:] },
		"ambiguous target record":    func(m *SnapshotContent) { m.Records = append(m.Records, m.Records[0]) },
		"same id different family":   func(m *SnapshotContent) { m.Records[0].SourceRecordRef.SourceFamily = "host" },
		"source selector in field": func(m *SnapshotContent) {
			m.Sections[0].Blocks[1].Fields[0].SupportRefs = []string{"/evidence/metadata"}
		},
		"unknown section reference": func(m *SnapshotContent) { m.Sections[0].SupportRefs = []string{"sup_unknown"} },
		"unknown nested block reference": func(m *SnapshotContent) {
			m.Sections[0].Blocks[0].Children = []ReportingBlock{{SupportRefs: []string{"sup_unknown"}}}
		},
		"unknown record reference": func(m *SnapshotContent) { m.Records[0].SupportRefs = []string{"sup_unknown"} },
		"unknown event reference": func(m *SnapshotContent) {
			m.TimelineEvents = []ReportingTimelineEvent{{SupportRefs: []string{"sup_unknown"}}}
		},
		"unknown relationship reference": func(m *SnapshotContent) {
			m.Relationships = []ReportingRelationshipSummary{{SupportRefs: []string{"sup_unknown"}}}
		},
		"orphaned index entry": func(m *SnapshotContent) {
			target := supportTarget{"source_record", "record:incident_metadata:" + m.IncidentID}
			id, err := supportReferenceID(m.SnapshotID, target)
			if err != nil {
				t.Fatal(err)
			}
			m.SupportIndex = append(m.SupportIndex, ReportingSupportRef{SchemaID: "cartulary.reporting_support_ref.v1", SupportRefID: id, SupportKind: target.kind, SupportTargetRef: target.ref, SourceSnapshotID: m.SnapshotID})
			sort.Slice(m.SupportIndex, func(i, j int) bool { return m.SupportIndex[i].SupportRefID < m.SupportIndex[j].SupportRefID })
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			candidate, err := copyReportingModel(model)
			if err != nil {
				t.Fatal(err)
			}
			mutate(&candidate.SnapshotContent)
			var invalid *renderValidationError
			if err := validateSnapshotSupportReferences(candidate.SnapshotContent); !errors.As(err, &invalid) || invalid.FailureCode != "export_model_invalid" || invalid.ReasonCode != "invalid_support_ref" {
				t.Fatal("support owner failed to detect the altered graph", err)
			}
			// Some hostile fixtures also violate independent field/identity
			// invariants. Their canonical stage ordering, not validator call
			// order, determines the first failure; the support defect must
			// remain in the complete retained set.
			failure := renderFailureDetails(validateRenderModelIdentity(candidate))
			found := false
			if failure.issues != nil {
				for _, issue := range failure.issues.retained {
					if issue.FailureCode != nil && *issue.FailureCode == "export_model_invalid" && issue.ReasonCode != nil && *issue.ReasonCode == "invalid_support_ref" {
						found = true
					}
				}
			}
			if !found || validateRenderFailure(failure) != nil {
				t.Fatal("model admission discarded or misclassified the support defect")
			}
			if fields, err := candidate.RedactionFields(); err == nil || fields != nil {
				t.Fatal("redaction used an invalid retained support graph")
			}
		})
	}
}
