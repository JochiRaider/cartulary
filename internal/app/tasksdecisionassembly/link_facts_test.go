package tasksdecisionassembly

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
)

func TestLinkFactsAdapterCopiesScalarFactsAndFailsWithoutPartialResults_Unit(t *testing.T) {
	incidentID := uuid.New()
	recordID := uuid.New()
	targetID := uuid.New()
	fieldKey := "task.decision_record_id"
	capability := newLinkFactsCapability(fakeLinkRecordFactReader{facts: links.ActiveFacts{
		RecordLinks: []links.RecordLinkFact{{
			SrcRecordID: recordID, DstRecordID: targetID,
			LinkType: links.LinkType(links.LinkTypeReferencesRecord), FieldKey: &fieldKey,
		}},
	}})
	facts, err := capability.LoadRecordLinkFactsTx(context.Background(), nil, incidentID, recordID)
	if err != nil || len(facts) != 1 {
		t.Fatalf("mapped facts = %#v, err=%v", facts, err)
	}
	if facts[0].SourceRecordID != recordID || facts[0].DestinationRecordID != targetID ||
		facts[0].LinkType != links.LinkTypeReferencesRecord.String() || !facts[0].HasFieldKey || facts[0].FieldKey != fieldKey {
		t.Fatalf("mapped fact = %#v", facts[0])
	}
	fieldKey = "caller mutation"
	if facts[0].FieldKey != "task.decision_record_id" {
		t.Fatalf("mapped fact retained source pointer: %#v", facts[0])
	}

	empty, err := newLinkFactsCapability(fakeLinkRecordFactReader{facts: links.ActiveFacts{RecordLinks: []links.RecordLinkFact{}}}).LoadRecordLinkFactsTx(
		context.Background(), nil, incidentID, recordID,
	)
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("empty facts = %#v, err=%v, want non-nil empty", empty, err)
	}

	wantErr := errors.New("fact read failed")
	partial, err := newLinkFactsCapability(fakeLinkRecordFactReader{
		facts: links.ActiveFacts{RecordLinks: []links.RecordLinkFact{{SrcRecordID: recordID}}},
		err:   wantErr,
	}).LoadRecordLinkFactsTx(context.Background(), nil, incidentID, recordID)
	if !errors.Is(err, wantErr) || partial != nil {
		t.Fatalf("failed facts = %#v, err=%v, want nil and wrapped source error", partial, err)
	}
}

type fakeLinkRecordFactReader struct {
	facts links.ActiveFacts
	err   error
}

func (reader fakeLinkRecordFactReader) LoadRecordTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (links.ActiveFacts, error) {
	return reader.facts, reader.err
}
