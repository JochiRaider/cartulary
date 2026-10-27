package revisions

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type contributedTargetWitness struct {
	tx        pgx.Tx
	appliedTx pgx.Tx
	requests  []rollbackcontract.NonRowTarget
	applied   rollbackcontract.ApplyInverseRequest
	recordID  uuid.UUID
}

func (w *contributedTargetWitness) DescribeTx(_ context.Context, tx pgx.Tx, request rollbackcontract.DescribeRequest) (rollbackcontract.TargetDescriptor, error) {
	w.tx = tx
	w.requests = append(w.requests, request.Target)
	return rollbackcontract.TargetDescriptor{AffectedRecordIDs: []uuid.UUID{w.recordID}}, nil
}
func (w *contributedTargetWitness) ApplyInverseTx(_ context.Context, tx pgx.Tx, request rollbackcontract.ApplyInverseRequest) (rollbackcontract.ApplyInverseResult, error) {
	w.appliedTx = tx
	w.applied = request
	return rollbackcontract.ApplyInverseResult{AffectedRecordIDs: []uuid.UUID{w.recordID}, BeforeValue: request.Target.AfterValue, AfterValue: request.Target.BeforeValue}, nil
}
func TestGenericRollbackDispatchUsesContributedTarget_Unit(t *testing.T) {
	const kind = "future_relationship"
	id, incident, actor := uuid.New(), uuid.New(), uuid.New()
	witness := &contributedTargetWitness{recordID: id}
	catalog, err := compileTargetSemanticsCatalog([]targetSemanticsRequirement{{TargetKind: kind, SourceOwnerID: "links", DispatchClass: rollbackcontract.DispatchNonRow, HistoryRecordIDFields: []string{"record_id"}, Addressability: HistorySingleEntry}}, []ProviderContribution{{SourceOwnerModule: SourceOwnerLinks, NonRowTargets: []NonRowProviderContribution{{SourceOwnerModule: SourceOwnerLinks, TargetKind: kind, HistoryProjector: catalogTestHistoryProjector, HistoryFacet: NewFieldAssociationHistoryFacet([]string{"record_id"}, HistorySingleEntry), RollbackProvider: witness}}}})
	if err != nil {
		t.Fatal(err)
	}
	target := rollbackMutationTarget{TargetKind: kind, TargetID: "future-relationship:1", ChangeSetID: uuid.New(), SequenceNo: 7, OperationKind: "patch", BeforeValue: map[string]any{"record_id": id.String(), "value": "before"}, AfterValue: map[string]any{"record_id": id.String(), "value": "after"}}
	tx := &commandObservedTx{}
	now := time.Now().UTC()
	result, err := (rollbackTransactionalApplier{commandStore: &commandStore{targetSemantics: catalog}}).executeNonRowInverseTx(context.Background(), tx, NewActorID(actor), incident, target, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(witness.requests) != 1 || witness.tx != tx || witness.appliedTx != tx || witness.applied.ActorUserID != actor || witness.applied.Now != now || witness.applied.Target.IncidentID != incident || witness.applied.Target.TargetKind != kind || !reflect.DeepEqual(witness.applied.Target.BeforeValue, target.BeforeValue) {
		t.Fatalf("provider did not receive generic command: %#v", witness)
	}
	if !reflect.DeepEqual(result.AfterValue, target.BeforeValue) || !reflect.DeepEqual(result.AffectedRecordIDs, []uuid.UUID{id}) {
		t.Fatalf("provider result lost: %#v", result)
	}
	if _, err := (rollbackTransactionalApplier{commandStore: &commandStore{targetSemantics: catalog}}).executeNonRowInverseTx(context.Background(), tx, NewActorID(actor), incident, rollbackMutationTarget{TargetKind: "unregistered"}, now); err == nil {
		t.Fatal("unregistered kind accepted")
	}
}
