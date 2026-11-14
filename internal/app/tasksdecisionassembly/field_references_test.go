package tasksdecisionassembly

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	linkfacts "github.com/JochiRaider/cartulary/internal/modules/links/facts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
)

func exerciseFieldReferenceAdapter(t *testing.T) {
	incident, source, target := uuid.New(), uuid.New(), uuid.New()
	reader := &fakeFieldReferenceReader{facts: []linkfacts.FieldReferenceFact{{DestinationRecordID: target, LinkType: links.LinkTypeReferencesRecord}}}
	adapter := fieldReferenceAdapter{reader: reader}
	facts, err := adapter.LoadFieldReferencesTx(context.Background(), nil, incident, source, "task.decision_record_id")
	if err != nil || len(facts) != 1 || facts[0].DestinationRecordID != target || facts[0].LinkType != "references_record" {
		t.Fatalf("field binding projection=%#v error=%v", facts, err)
	}
	if reader.incident != incident || reader.source != source || reader.field != "task.decision_record_id" {
		t.Fatal("field binding scope was not forwarded exactly")
	}
	reader.facts[0].DestinationRecordID = uuid.New()
	if facts[0].DestinationRecordID != target {
		t.Fatal("field bindings alias source storage")
	}
	reader.facts = nil
	empty, err := adapter.LoadFieldReferencesTx(context.Background(), nil, incident, source, "task.decision_record_id")
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("empty field bindings=%#v error=%v", empty, err)
	}
	reader.facts = []linkfacts.FieldReferenceFact{{DestinationRecordID: target, LinkType: links.LinkTypeReferencesRecord}}
	reader.err = errors.New("field binding failure")
	partial, err := adapter.LoadFieldReferencesTx(context.Background(), nil, incident, source, "task.decision_record_id")
	if !errors.Is(err, reader.err) || partial != nil {
		t.Fatalf("partial field bindings=%#v error=%v", partial, err)
	}
	if contribution, err := tasksdecisions.NewRevisionContribution(NewLinkFactsCapability(), nil, NewSupersessionReader(), NewReferenceTargetCapability()); err == nil || len(contribution.Records) != 0 {
		t.Fatalf("missing field reader produced contribution=%#v error=%v", contribution, err)
	}
	if contribution, err := tasksdecisions.NewRevisionContribution(nil, NewFieldReferenceReader(), NewSupersessionReader(), NewReferenceTargetCapability()); err == nil || len(contribution.Records) != 0 {
		t.Fatalf("missing active Links facts produced contribution=%#v error=%v", contribution, err)
	}
}

type fakeFieldReferenceReader struct {
	facts            []linkfacts.FieldReferenceFact
	err              error
	incident, source uuid.UUID
	field            string
}

func (r *fakeFieldReferenceReader) LoadFieldReferencesTx(_ context.Context, _ pgx.Tx, incident, source uuid.UUID, field string) ([]linkfacts.FieldReferenceFact, error) {
	r.incident, r.source, r.field = incident, source, field
	return r.facts, r.err
}
