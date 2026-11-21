package collectionfacts

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	evidencefacts "github.com/JochiRaider/cartulary/internal/modules/evidence/timelinefacts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func TestReaderMapsFactsLosslesslyInOrder(t *testing.T) {
	incidentID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	recordID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	replacementID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	evidenceIDs := []uuid.UUID{
		uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		uuid.MustParse("55555555-5555-4555-8555-555555555555"),
	}
	events := []string{}
	tx := &transactionMarker{}
	mentions := []workbookprojection.MentionFact{
		{MentionID: uuid.MustParse("66666666-6666-4666-8666-666666666666"), RawText: "host-a", SourceFieldKey: "timeline.host_refs"},
		{MentionID: uuid.MustParse("77777777-7777-4777-8777-777777777777"), RawText: "host-a", SourceFieldKey: "timeline.host_refs"},
	}
	linkFacts := LinkFacts{
		ResolvedLinks:       []LinkFact{{TargetRecordID: recordID, LinkType: "observed_on_host"}},
		Tags:                []TagFact{{RecordTagID: evidenceIDs[0], TagName: "first"}, {RecordTagID: evidenceIDs[1], TagName: "first"}},
		AttachedEvidenceIDs: append([]uuid.UUID(nil), evidenceIDs...),
		ReplacementRecordID: &replacementID,
	}
	evidenceFacts := []evidencefacts.TimelineFact{
		{RecordID: evidenceIDs[0], Title: "Disk image", LifecycleState: "available", UploadState: "available"},
		{RecordID: evidenceIDs[1], Title: "Memory capture", LifecycleState: "quarantined", UploadState: "quarantined"},
	}
	mentionReader := &mentionReaderStub{events: &events, facts: mentions}
	linkReader := &linkReaderStub{events: &events, facts: linkFacts}
	evidenceReader := &evidenceReaderStub{events: &events, facts: evidenceFacts}

	got, err := New(mentionReader, linkReader, evidenceReader).LoadTimelineCollectionFactsTx(
		context.Background(), tx, incidentID, recordID,
	)
	if err != nil {
		t.Fatalf("load collection facts: %v", err)
	}
	want := workbookprojection.CollectionFacts{
		Mentions:      mentions,
		ResolvedLinks: []workbookprojection.LinkFact{{TargetRecordID: recordID, LinkType: "observed_on_host"}},
		Tags:          []workbookprojection.TagFact{{RecordTagID: evidenceIDs[0], TagName: "first"}, {RecordTagID: evidenceIDs[1], TagName: "first"}},
		AttachedEvidence: []workbookprojection.EvidenceFact{
			{RecordID: evidenceIDs[0], Title: "Disk image", LifecycleState: "available", UploadState: "available"},
			{RecordID: evidenceIDs[1], Title: "Memory capture", LifecycleState: "quarantined", UploadState: "quarantined"},
		},
		ReplacementRecordID: &replacementID,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("facts = %#v, want %#v", got, want)
	}
	if !reflect.DeepEqual(events, []string{"mentions", "links", "evidence"}) {
		t.Fatalf("read order = %#v", events)
	}
	for name, observed := range map[string]pgx.Tx{
		"mentions": mentionReader.tx,
		"links":    linkReader.tx,
		"evidence": evidenceReader.tx,
	} {
		if observed != tx {
			t.Fatalf("%s reader received transaction %p, want %p", name, observed, tx)
		}
	}
	if !reflect.DeepEqual(evidenceReader.recordIDs, evidenceIDs) {
		t.Fatalf("evidence IDs = %#v, want %#v", evidenceReader.recordIDs, evidenceIDs)
	}

	empty, err := New(&mentionReaderStub{}, &linkReaderStub{}, &evidenceReaderStub{}).
		LoadTimelineCollectionFactsTx(context.Background(), tx, incidentID, recordID)
	if err != nil {
		t.Fatalf("load empty collection facts: %v", err)
	}
	if empty.AttachedEvidence == nil || len(empty.AttachedEvidence) != 0 {
		t.Fatalf("nil evidence mapped to %#v, want non-nil empty slice", empty.AttachedEvidence)
	}
}

func TestReaderStopsAtFirstFailure(t *testing.T) {
	stages := []struct {
		name        string
		mentionErr  error
		linkErr     error
		evidenceErr error
		wantEvents  []string
		wantPrefix  string
	}{
		{name: "mentions", mentionErr: errors.New("mention failure"), wantEvents: []string{"mentions"}, wantPrefix: "load Timeline entity facts: mention failure"},
		{name: "links", linkErr: errors.New("link failure"), wantEvents: []string{"mentions", "links"}, wantPrefix: "load Timeline link facts: link failure"},
		{name: "evidence", evidenceErr: errors.New("evidence failure"), wantEvents: []string{"mentions", "links", "evidence"}, wantPrefix: "load Timeline evidence facts: evidence failure"},
		{name: "cancellation", mentionErr: context.Canceled, wantEvents: []string{"mentions"}, wantPrefix: "load Timeline entity facts: context canceled"},
	}
	for _, stage := range stages {
		t.Run(stage.name, func(t *testing.T) {
			events := []string{}
			reader := New(
				&mentionReaderStub{events: &events, err: stage.mentionErr},
				&linkReaderStub{events: &events, err: stage.linkErr},
				&evidenceReaderStub{events: &events, err: stage.evidenceErr},
			)
			facts, err := reader.LoadTimelineCollectionFactsTx(context.Background(), &transactionMarker{}, uuid.New(), uuid.New())
			if err == nil || err.Error() != stage.wantPrefix {
				t.Fatalf("error = %v, want %q", err, stage.wantPrefix)
			}
			if !reflect.DeepEqual(facts, workbookprojection.CollectionFacts{}) {
				t.Fatalf("failure returned partial facts: %#v", facts)
			}
			if !reflect.DeepEqual(events, stage.wantEvents) {
				t.Fatalf("events = %#v, want %#v", events, stage.wantEvents)
			}
		})
	}
}

func TestReaderRequiresAllDependenciesBeforeReading(t *testing.T) {
	events := []string{}
	mentions := &mentionReaderStub{events: &events}
	linkReader := &linkReaderStub{events: &events}
	evidenceReader := &evidenceReaderStub{events: &events}
	readers := []Reader{
		New(nil, linkReader, evidenceReader),
		New(mentions, nil, evidenceReader),
		New(mentions, linkReader, nil),
	}
	for index, reader := range readers {
		if _, err := reader.LoadTimelineCollectionFactsTx(context.Background(), &transactionMarker{}, uuid.New(), uuid.New()); err == nil {
			t.Fatalf("missing dependency case %d succeeded", index)
		}
	}
	if len(events) != 0 {
		t.Fatalf("missing dependency issued reads: %#v", events)
	}
}

type transactionMarker struct{ pgx.Tx }

type mentionReaderStub struct {
	events *[]string
	tx     pgx.Tx
	facts  []workbookprojection.MentionFact
	err    error
}

func (reader *mentionReaderStub) LoadMentionsTx(_ context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID][]workbookprojection.MentionFact, error) {
	if reader.events != nil {
		*reader.events = append(*reader.events, "mentions")
	}
	reader.tx = tx
	if reader.err != nil {
		return nil, reader.err
	}
	result := map[uuid.UUID][]workbookprojection.MentionFact{}
	for _, recordID := range recordIDs {
		if reader.facts != nil {
			result[recordID] = reader.facts
		}
	}
	return result, nil
}

type linkReaderStub struct {
	events *[]string
	tx     pgx.Tx
	facts  LinkFacts
	err    error
}

func (reader *linkReaderStub) LoadTx(_ context.Context, tx pgx.Tx, _ uuid.UUID, recordIDs []uuid.UUID) (map[uuid.UUID]LinkFacts, error) {
	if reader.events != nil {
		*reader.events = append(*reader.events, "links")
	}
	reader.tx = tx
	if reader.err != nil {
		return nil, reader.err
	}
	result := map[uuid.UUID]LinkFacts{}
	for _, recordID := range recordIDs {
		result[recordID] = reader.facts
	}
	return result, nil
}

type evidenceReaderStub struct {
	events    *[]string
	tx        pgx.Tx
	recordIDs []uuid.UUID
	facts     []evidencefacts.TimelineFact
	err       error
}

func (reader *evidenceReaderStub) LoadTx(_ context.Context, tx pgx.Tx, _ uuid.UUID, recordIDs []uuid.UUID) ([]evidencefacts.TimelineFact, error) {
	if reader.events != nil {
		*reader.events = append(*reader.events, "evidence")
	}
	reader.tx = tx
	reader.recordIDs = append([]uuid.UUID(nil), recordIDs...)
	return reader.facts, reader.err
}

// TestReaderPartitionsAPageWithOneReadPerSource: a page of records costs one
// read per fact source, and each record receives exactly its own facts with
// attached Evidence in the Evidence reader's display order.
func TestReaderPartitionsAPageWithOneReadPerSource(t *testing.T) {
	incidentID := uuid.New()
	first, second, bare := uuid.New(), uuid.New(), uuid.New()
	disk, memory := uuid.New(), uuid.New()
	events := []string{}
	mentionFact := workbookprojection.MentionFact{MentionID: uuid.New(), RawText: "host-a", SourceFieldKey: "timeline.host_refs"}
	mentions := &pageMentionReader{events: &events, facts: map[uuid.UUID][]workbookprojection.MentionFact{first: {mentionFact}}}
	links := &pageLinkReader{events: &events, facts: map[uuid.UUID]LinkFacts{
		first:  {AttachedEvidenceIDs: []uuid.UUID{memory, disk}, Tags: []TagFact{{RecordTagID: uuid.New(), TagName: "first"}}},
		second: {AttachedEvidenceIDs: []uuid.UUID{memory}},
		bare:   {},
	}}
	evidence := &evidenceReaderStub{events: &events, facts: []evidencefacts.TimelineFact{
		{RecordID: disk, Title: "Disk image", LifecycleState: "received", UploadState: "complete"},
		{RecordID: memory, Title: "Memory capture", LifecycleState: "received", UploadState: "complete"},
	}}
	got, err := New(mentions, links, evidence).LoadTimelineCollectionFactsForRecordsTx(context.Background(), &transactionMarker{}, incidentID, []uuid.UUID{first, second, bare})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(events, []string{"mentions", "links", "evidence"}) {
		t.Fatalf("reads = %#v, want one per source", events)
	}
	if !reflect.DeepEqual(evidence.recordIDs, []uuid.UUID{memory, disk}) {
		t.Fatalf("evidence read IDs = %#v, want the distinct union in first-seen order", evidence.recordIDs)
	}
	diskFact := workbookprojection.EvidenceFact{RecordID: disk, Title: "Disk image", LifecycleState: "received", UploadState: "complete"}
	memoryFact := workbookprojection.EvidenceFact{RecordID: memory, Title: "Memory capture", LifecycleState: "received", UploadState: "complete"}
	want := map[uuid.UUID]workbookprojection.CollectionFacts{
		first: {
			Mentions: []workbookprojection.MentionFact{mentionFact}, ResolvedLinks: []workbookprojection.LinkFact{},
			Tags: []workbookprojection.TagFact{{RecordTagID: links.facts[first].Tags[0].RecordTagID, TagName: "first"}}, AttachedEvidence: []workbookprojection.EvidenceFact{diskFact, memoryFact},
		},
		second: {
			Mentions: []workbookprojection.MentionFact{}, ResolvedLinks: []workbookprojection.LinkFact{},
			Tags: []workbookprojection.TagFact{}, AttachedEvidence: []workbookprojection.EvidenceFact{memoryFact},
		},
		bare: {
			Mentions: []workbookprojection.MentionFact{}, ResolvedLinks: []workbookprojection.LinkFact{},
			Tags: []workbookprojection.TagFact{}, AttachedEvidence: []workbookprojection.EvidenceFact{},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("facts = %#v, want %#v", got, want)
	}
}

type pageMentionReader struct {
	events *[]string
	facts  map[uuid.UUID][]workbookprojection.MentionFact
}

func (reader *pageMentionReader) LoadMentionsTx(_ context.Context, _ pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID][]workbookprojection.MentionFact, error) {
	*reader.events = append(*reader.events, "mentions")
	result := map[uuid.UUID][]workbookprojection.MentionFact{}
	for _, recordID := range recordIDs {
		if facts, ok := reader.facts[recordID]; ok {
			result[recordID] = facts
		}
	}
	return result, nil
}

type pageLinkReader struct {
	events *[]string
	facts  map[uuid.UUID]LinkFacts
}

func (reader *pageLinkReader) LoadTx(_ context.Context, _ pgx.Tx, _ uuid.UUID, recordIDs []uuid.UUID) (map[uuid.UUID]LinkFacts, error) {
	*reader.events = append(*reader.events, "links")
	result := map[uuid.UUID]LinkFacts{}
	for _, recordID := range recordIDs {
		result[recordID] = reader.facts[recordID]
	}
	return result, nil
}
