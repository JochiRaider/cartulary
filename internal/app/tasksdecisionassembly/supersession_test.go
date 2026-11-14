package tasksdecisionassembly

import (
	"context"
	"errors"
	linkfacts "github.com/JochiRaider/cartulary/internal/modules/links/facts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"testing"
)

func exerciseSupersessionAdapter(t *testing.T) {
	incident, source, target := uuid.New(), uuid.New(), uuid.New()
	ctx := t.Context()
	tx := &struct{ pgx.Tx }{}
	reader := &fakeSupersessionReader{facts: []linkfacts.SupersessionFact{{SourceRecordID: source, DestinationRecordID: target}}}
	adapter := supersessionAdapter{reader: reader}
	facts, err := adapter.LoadRetainedSupersessionTx(ctx, tx, incident, source)
	if err != nil || len(facts) != 1 || facts[0].SourceRecordID != source || facts[0].DestinationRecordID != target {
		t.Fatalf("supersession projection=%#v error=%v", facts, err)
	}
	if reader.ctx != ctx || reader.tx != tx || reader.incident != incident || reader.record != source {
		t.Fatal("supersession read did not preserve transaction/scope")
	}
	reader.facts[0].SourceRecordID = uuid.New()
	reader.facts[0].DestinationRecordID = uuid.New()
	if facts[0].SourceRecordID != source || facts[0].DestinationRecordID != target {
		t.Fatal("supersession facts alias source storage")
	}
	reader.facts = nil
	empty, err := adapter.LoadRetainedSupersessionTx(ctx, tx, incident, source)
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("empty facts=%#v error=%v", empty, err)
	}
	reader.facts = []linkfacts.SupersessionFact{{SourceRecordID: source, DestinationRecordID: target}}
	reader.err = errors.New("supersession read failed")
	partial, err := adapter.LoadRetainedSupersessionTx(ctx, tx, incident, source)
	if !errors.Is(err, reader.err) || partial != nil {
		t.Fatalf("partial facts=%#v error=%v", partial, err)
	}
	if contribution, err := tasksdecisions.NewRevisionContribution(NewLinkFactsCapability(), NewFieldReferenceReader(), nil, NewReferenceTargetCapability()); err == nil || len(contribution.Records) != 0 {
		t.Fatalf("missing supersession reader produced contribution=%#v error=%v", contribution, err)
	}
}

type fakeSupersessionReader struct {
	facts            []linkfacts.SupersessionFact
	err              error
	ctx              context.Context
	tx               pgx.Tx
	incident, record uuid.UUID
}

func (r *fakeSupersessionReader) LoadRetainedSupersessionTx(ctx context.Context, tx pgx.Tx, incident, record uuid.UUID) ([]linkfacts.SupersessionFact, error) {
	r.ctx, r.tx, r.incident, r.record = ctx, tx, incident, record
	return r.facts, r.err
}
