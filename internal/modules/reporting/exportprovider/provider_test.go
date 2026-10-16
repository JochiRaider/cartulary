package exportprovider

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestProviderOutputValidateAndBindings(t *testing.T) {
	assertExplicitOwnerConstruction(t)
	assertOwnerQueryBinding(t)
	output := ProviderOutput{
		SourceFacts: EmptySourceFacts(),
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
			StableSubjectRef:        "subject:host:record:host-1",
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
	fields, err := output.BoundFields()
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 || fields[0].Path != "/hosts/host-1" {
		t.Fatalf("BoundFields() = %#v", fields)
	}
	fields[0].Path = "/mutated"
	if output.FieldFacts[0].Path != "/hosts/host-1" {
		t.Fatalf("BoundFields() returned storage alias")
	}
}

func TestProviderOutputValidateRejectsFlatOrIncompleteFacts(t *testing.T) {
	assertProviderBindings(t)
	output := ProviderOutput{
		SourceFacts: EmptySourceFacts(),
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
		SourceFacts: EmptySourceFacts(),
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

func assertProviderBindings(t *testing.T) {
	t.Helper()
	fresh := func() ProviderOutput {
		t.Helper()
		out, err := NewProviderOutput("test", []BoundField{{RecordID: "source-id", RecordType: "host", FieldFact: FieldFact{SchemaID: FieldFactSchemaID, Path: "/hosts/source-id", SourceFamily: "host", ContentClass: "derived_analytic", Value: "value", DisclosurePartitionRefs: []string{"party:a"}, SupportRefs: []string{"/evidence/e"}}}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	out := fresh()
	// All declarations name the same selector; the record ID is independent.
	out.FieldFacts[0].Path = "/selector/label"
	out.RecordFacts[0].FieldPaths[0] = "/selector/label"
	out.SubjectFacts[0].FieldPaths[0] = "/selector/label"
	out.SupportRefFacts[0].SourcePath = "/selector/label"
	out.DisclosurePartitionFacts[0].FieldPath = "/selector/label"
	bound, err := out.BoundFields()
	if err != nil || len(bound) != 1 || bound[0].RecordID != "source-id" || bound[0].RecordType != "host" {
		t.Fatal("explicit owner identity lost", bound, err)
	}
	for name, change := range map[string]func(*ProviderOutput){
		"missing record":       func(o *ProviderOutput) { o.RecordFacts = nil },
		"duplicate record":     func(o *ProviderOutput) { o.RecordFacts = append(o.RecordFacts, o.RecordFacts[0]) },
		"dangling field":       func(o *ProviderOutput) { o.RecordFacts[0].FieldPaths = []string{"/absent"} },
		"foreign field family": func(o *ProviderOutput) { o.RecordFacts[0].SourceFamily = "identity" },
		"overlapping record": func(o *ProviderOutput) {
			other := o.RecordFacts[0]
			other.RecordID = "other"
			o.RecordFacts = append(o.RecordFacts, other)
		},
		"lost record partition":   func(o *ProviderOutput) { o.RecordFacts[0].DisclosurePartitionRefs = nil },
		"lost record support":     func(o *ProviderOutput) { o.RecordFacts[0].SupportRefs = nil },
		"duplicate subject":       func(o *ProviderOutput) { o.SubjectFacts = append(o.SubjectFacts, o.SubjectFacts[0]) },
		"missing subject":         func(o *ProviderOutput) { o.SubjectFacts = nil },
		"legacy subject identity": func(o *ProviderOutput) { o.SubjectFacts[0].StableSubjectRef = "host:source-id" },
		"foreign subject family":  func(o *ProviderOutput) { o.SubjectFacts[0].SubjectKind = "party" },
		"lost subject partition":  func(o *ProviderOutput) { o.SubjectFacts[0].DisclosurePartitionRefs = nil },
		"foreign subject":         func(o *ProviderOutput) { o.SubjectFacts[0].SourceRecordID = "other" },
		"dangling subject field":  func(o *ProviderOutput) { o.SubjectFacts[0].FieldPaths = []string{"/absent"} },
		"foreign timeline": func(o *ProviderOutput) {
			o.TimelineEventFacts = []TimelineEventFact{{SchemaID: TimelineEventFactSchemaID, TimelineEventID: "source-id", SourceFamily: "timeline_event", FieldPaths: []string{"/hosts/source-id"}}}
		},
		"missing support":         func(o *ProviderOutput) { o.SupportRefFacts = nil },
		"dangling support":        func(o *ProviderOutput) { o.SupportRefFacts[0].SourcePath = "/absent" },
		"duplicate support":       func(o *ProviderOutput) { o.SupportRefFacts = append(o.SupportRefFacts, o.SupportRefFacts[0]) },
		"foreign support":         func(o *ProviderOutput) { o.SupportRefFacts[0].TargetPath = "/evidence/other" },
		"foreign support binding": func(o *ProviderOutput) { o.SupportRefFacts[0].SupportTargetRef = "/evidence/other" },
		"missing partition":       func(o *ProviderOutput) { o.DisclosurePartitionFacts = nil },
		"foreign partition owner": func(o *ProviderOutput) { o.DisclosurePartitionFacts[0].SourceRecordID = "other" },
		"unclaimed partition":     func(o *ProviderOutput) { o.DisclosurePartitionFacts[0].PartitionRef = "party:b" },
		"duplicate partition": func(o *ProviderOutput) {
			o.DisclosurePartitionFacts = append(o.DisclosurePartitionFacts, o.DisclosurePartitionFacts[0])
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := fresh()
			change(&candidate)
			fields, err := candidate.BoundFields()
			if err == nil || fields != nil {
				t.Fatal("malformed owner facts published partial output", fields, err)
			}
		})
	}
}

func assertExplicitOwnerConstruction(t *testing.T) {
	t.Helper()
	fields := []BoundField{
		{RecordID: "timeline-id", RecordType: "timeline_event", FieldFact: FieldFact{SchemaID: FieldFactSchemaID, Path: "/arbitrary/two", SourceFamily: "timeline_event", ContentClass: "source_evidence", DisclosurePartitionRefs: []string{"party:b"}, SupportRefs: []string{"/evidence/e", "/evidence/e"}}},
		{RecordID: "timeline-id", RecordType: "timeline_event", FieldFact: FieldFact{SchemaID: FieldFactSchemaID, Path: "/arbitrary/one", SourceFamily: "timeline_event", ContentClass: "source_evidence", DisclosurePartitionRefs: []string{"party:a"}}},
	}
	timeline := []TimelineSortFact{{RecordID: "timeline-id", RecordCreatedAt: "2026-05-01T00:00:00.000000Z", ActivityParseState: "missing", TimeBucket: "unresolved_time"}}
	output, err := NewProviderOutput("owner", fields, timeline)
	if err != nil {
		t.Fatal(err)
	}
	if len(output.RecordFacts) != 1 || output.RecordFacts[0].RecordID != "timeline-id" || len(output.TimelineEventFacts) != 1 || output.TimelineEventFacts[0].TimelineEventID != "timeline-id" || len(output.TimelineEventFacts[0].FieldPaths) != 2 {
		t.Fatal("source identity was replaced by selector spelling", output)
	}
	if fields[0].Path != "/arbitrary/two" || len(fields[0].SupportRefs) != 2 || len(output.FieldFacts[1].SupportRefs) != 2 || len(output.SupportRefFacts) != 1 {
		t.Fatal("construction mutated source order or repeated support targets")
	}
	reversed, err := NewProviderOutput("owner", []BoundField{fields[1], fields[0]}, timeline)
	if err != nil || !reflect.DeepEqual(output, reversed) {
		t.Fatal("source order changed canonical facts", err)
	}
	for _, facts := range [][]TimelineSortFact{nil, {timeline[0], timeline[0]}, {{RecordID: "other", RecordCreatedAt: "2026-05-01T00:00:00.000000Z", ActivityParseState: "missing", TimeBucket: "unresolved_time"}}} {
		if got, err := NewProviderOutput("owner", fields, facts); err == nil || !reflect.DeepEqual(got, ProviderOutput{}) {
			t.Fatal("missing, duplicate or foreign sort facts admitted", err)
		}
	}
	for name, change := range map[string]func(*ProviderOutput){
		"retired output": func(o *ProviderOutput) { o.SchemaID = "cartulary.reporting_owner_provider_output.v1" },
		"retired event": func(o *ProviderOutput) {
			o.TimelineEventFacts[0].SchemaID = "cartulary.reporting_owner_timeline_event_fact.v1"
		},
		"missing sort":          func(o *ProviderOutput) { o.TimelineEventFacts[0].Sort = TimelineSortFact{} },
		"foreign sort identity": func(o *ProviderOutput) { o.TimelineEventFacts[0].Sort.RecordID = "other" },
		"missing event":         func(o *ProviderOutput) { o.TimelineEventFacts = nil },
		"lost event field":      func(o *ProviderOutput) { o.TimelineEventFacts[0].FieldPaths = o.TimelineEventFacts[0].FieldPaths[:1] },
		"lost event partition":  func(o *ProviderOutput) { o.TimelineEventFacts[0].DisclosurePartitionRefs = nil },
		"lost event support":    func(o *ProviderOutput) { o.TimelineEventFacts[0].SupportRefs = nil },
	} {
		t.Run(name, func(t *testing.T) {
			candidate, err := NewProviderOutput("owner", fields, timeline)
			if err != nil {
				t.Fatal(err)
			}
			change(&candidate)
			if bound, err := candidate.BoundFields(); err == nil || bound != nil {
				t.Fatal("invalid event contribution admitted")
			}
		})
	}
	for name, change := range map[string]func([]BoundField){
		"missing identity": func(f []BoundField) { f[0].RecordID = "" },
		"missing type":     func(f []BoundField) { f[0].RecordType = "" },
		"conflicting type": func(f []BoundField) { f[0].RecordType = "host" },
		"duplicate path":   func(f []BoundField) { f[0].Path = f[1].Path },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := append([]BoundField(nil), fields...)
			change(candidate)
			got, err := NewProviderOutput("owner", candidate, timeline)
			if err == nil || !reflect.DeepEqual(got, ProviderOutput{}) {
				t.Fatal("failed construction returned partial output", got, err)
			}
		})
	}
	mention := BoundField{RecordID: "mention-id", RecordType: "entity_mention", FieldFact: FieldFact{SchemaID: FieldFactSchemaID, Path: "/opaque/selector", SourceFamily: "entity_mention", ContentClass: "source_evidence", Value: map[string]any{"entity_type": "identity"}}}
	mentionOutput, err := NewProviderOutput("mentions", []BoundField{mention}, nil)
	if err != nil || len(mentionOutput.SubjectFacts) != 1 || mentionOutput.SubjectFacts[0].StableSubjectRef != "subject:identity:mention:mention-id" {
		t.Fatal("mention identity lost", mentionOutput, err)
	}
	mention.Value = map[string]any{"entity_type": "unknown"}
	if got, err := NewProviderOutput("mentions", []BoundField{mention}, nil); err == nil || !reflect.DeepEqual(got, ProviderOutput{}) {
		t.Fatal("unclassified mention admitted")
	}
}

func assertOwnerQueryBinding(t *testing.T) {
	t.Helper()
	query := []FieldQuery{{Prefix: "nested/selector", SQL: "owner query"}}
	for name, raw := range map[string]string{
		"exact integer":    `{"counter":9223372036854775807}`,
		"duplicate member": `{"counter":1,"counter":2}`,
		"malformed":        `{`,
	} {
		t.Run(name, func(t *testing.T) {
			rows := &ownerQueryRows{raw: []byte(raw)}
			tx := &ownerQueryTx{rows: rows}
			output, err := CollectQueryProviderOutputTx(context.Background(), tx, uuid.New(), "owner", nil, query)
			if !rows.closed {
				t.Fatal("query rows were not closed")
			}
			if name != "exact integer" {
				if err == nil || !reflect.DeepEqual(output, ProviderOutput{}) {
					t.Fatal("malformed query published partial facts")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if output.RecordFacts[0].RecordID != "source-id" || output.RecordFacts[0].RecordType != "owner-type" || output.FieldFacts[0].Path != "/nested/selector/source-id" {
				t.Fatal("query identity was inferred", output)
			}
			value := output.FieldFacts[0].Value.(map[string]any)["counter"]
			if n, ok := value.(json.Number); !ok || n.String() != "9223372036854775807" {
				t.Fatal("query rounded source integer", value)
			}
		})
	}
	sentinel := errors.New("source failure")
	for _, stage := range []string{"query", "scan", "rows"} {
		t.Run(stage+" failure", func(t *testing.T) {
			rows := &ownerQueryRows{raw: []byte(`{}`)}
			tx := &ownerQueryTx{rows: rows}
			switch stage {
			case "query":
				tx.err = sentinel
			case "scan":
				rows.scanErr = sentinel
			case "rows":
				rows.err = sentinel
			}
			output, err := CollectQueryProviderOutputTx(context.Background(), tx, uuid.New(), "owner", nil, query)
			if !errors.Is(err, sentinel) || !reflect.DeepEqual(output, ProviderOutput{}) {
				t.Fatal("failed query published partial output", output, err)
			}
			if stage != "query" && !rows.closed {
				t.Fatal("failed query left rows open")
			}
		})
	}
}

type ownerQueryTx struct {
	pgx.Tx
	rows *ownerQueryRows
	err  error
}

func (tx *ownerQueryTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return tx.rows, tx.err
}

type ownerQueryRows struct {
	pgx.Rows
	raw          []byte
	read, closed bool
	scanErr, err error
}

func (r *ownerQueryRows) Next() bool {
	if r.read {
		return false
	}
	r.read = true
	return true
}
func (r *ownerQueryRows) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}
	if len(dest) != 5 {
		return errors.New("unexpected source projection")
	}
	for i, value := range []string{"source-id", "owner-family", "owner-type", "source_evidence"} {
		*dest[i].(*string) = value
	}
	*dest[4].(*[]byte) = r.raw
	return nil
}
func (r *ownerQueryRows) Close()     { r.closed = true }
func (r *ownerQueryRows) Err() error { return r.err }
