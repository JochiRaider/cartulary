package exportprovider

import "testing"

func TestProviderOutputValidateAndFields(t *testing.T) {
	output := ProviderOutput{
		SchemaID:    ProviderOutputSchemaID,
		ProviderKey: "entities.hostidentity",
		FieldFacts: []FieldFact{{
			SchemaID:                FieldFactSchemaID,
			Path:                    "/hosts/host-1",
			ContentClass:            "derived_analytic",
			SourceFamily:            "host",
			Value:                   map[string]any{"hostname": "host-1"},
			DisclosurePartitionRefs: []string{"party:alpha"},
			SupportRefs:             []string{"/record_envelopes/evidence-1"},
		}},
		RecordFacts: []RecordFact{{
			SchemaID:                RecordFactSchemaID,
			SourceFamily:            "host",
			RecordID:                "host-1",
			RecordType:              "host",
			FieldPaths:              []string{"/hosts/host-1"},
			DisclosurePartitionRefs: []string{"party:alpha"},
			SupportRefs:             []string{"/record_envelopes/evidence-1"},
		}},
		TimelineEventFacts: []TimelineEventFact{},
		SubjectFacts: []SubjectFact{{
			SchemaID:                SubjectFactSchemaID,
			StableSubjectRef:        "host:host-1",
			SubjectKind:             "host",
			SourceFamily:            "host",
			SourceRecordID:          "host-1",
			FieldPaths:              []string{"/hosts/host-1"},
			DisclosurePartitionRefs: []string{"party:alpha"},
		}},
		SupportRefFacts: []SupportRefFact{{
			SchemaID:         SupportRefFactSchemaID,
			SupportRefID:     "support-record_envelopes-evidence-1",
			SourcePath:       "/hosts/host-1",
			TargetPath:       "/record_envelopes/evidence-1",
			SupportTargetRef: "/record_envelopes/evidence-1",
		}},
		DisclosurePartitionFacts: []DisclosurePartitionFact{{
			SchemaID:       DisclosurePartitionFactSchemaID,
			PartitionRef:   "party:alpha",
			SourceFamily:   "host",
			SourceRecordID: "host-1",
			FieldPath:      "/hosts/host-1",
		}},
	}
	if err := output.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	fields := output.Fields()
	if len(fields) != 1 || fields[0].Path != "/hosts/host-1" {
		t.Fatalf("Fields() = %#v", fields)
	}
	fields[0].Path = "/mutated"
	if output.FieldFacts[0].Path != "/hosts/host-1" {
		t.Fatalf("Fields() returned storage alias")
	}
}

func TestProviderOutputValidateRejectsFlatOrIncompleteFacts(t *testing.T) {
	output := ProviderOutput{
		SchemaID:    ProviderOutputSchemaID,
		ProviderKey: "records",
		FieldFacts: []FieldFact{{
			SchemaID:     FieldFactSchemaID,
			Path:         "/record_envelopes/record-1",
			ContentClass: "derived_analytic",
			SourceFamily: "record_envelope",
		}, {
			SchemaID:     FieldFactSchemaID,
			Path:         "/record_envelopes/record-1",
			ContentClass: "derived_analytic",
			SourceFamily: "record_envelope",
		}},
	}
	if err := output.Validate(); err == nil {
		t.Fatal("Validate() succeeded for duplicate field facts")
	}

	output = ProviderOutput{
		SchemaID:    ProviderOutputSchemaID,
		ProviderKey: "records",
		RecordFacts: []RecordFact{{
			SchemaID: RecordFactSchemaID,
			RecordID: "record-1",
		}},
	}
	if err := output.Validate(); err == nil {
		t.Fatal("Validate() succeeded for incomplete record fact")
	}
}
