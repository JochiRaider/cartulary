package revisions

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestRollbackPlannerBuildsExplicitDeterministicPlan_Unit(t *testing.T) {
	left := uuid.MustParse("00000000-0000-4000-8000-000000000010")
	right := uuid.MustParse("00000000-0000-4000-8000-000000000002")
	changeSetID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	plan := rollbackPlan{
		Target:     rollbackMutationTarget{ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "host", TargetID: left.String(), BeforeValue: map[string]any{"cells": map[string]any{"host.name": "left-before"}}, AfterValue: map[string]any{"cells": map[string]any{"host.name": "left-after"}}},
		Targets:    []rollbackMutationTarget{{ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "host", TargetID: left.String(), BeforeValue: map[string]any{"cells": map[string]any{"host.name": "left-before"}}, AfterValue: map[string]any{"cells": map[string]any{"host.name": "left-after"}}}, {ChangeSetID: changeSetID, SequenceNo: 2, TargetKind: "host", TargetID: right.String(), BeforeValue: map[string]any{"cells": map[string]any{"host.name": "right-before"}}, AfterValue: map[string]any{"cells": map[string]any{"host.name": "right-after"}}}},
		Affected:   []uuid.UUID{left, right, left},
		Addressed:  left,
		RecordType: "host",
		WholeSet:   true,
	}
	targetSemantics := validTargetSemanticsCatalog(t, validProviderContributions())
	finalized, err := (rollbackPlanner{targetSemantics: targetSemantics}).finalize(plan, map[uuid.UUID]rollbackRecordEnvelope{
		left:  {RecordID: left, RecordType: "host", RowVersion: 7},
		right: {RecordID: right, RecordType: "host", RowVersion: 11},
	})
	if err != nil {
		t.Fatalf("finalize rollback plan: %v", err)
	}
	if !reflect.DeepEqual(finalized.Affected, []uuid.UUID{right, left}) {
		t.Fatalf("canonical affected records = %v", finalized.Affected)
	}
	if finalized.ExpectedVersions[left] != 7 || finalized.ExpectedVersions[right] != 11 {
		t.Fatalf("expected versions = %#v", finalized.ExpectedVersions)
	}
	if len(finalized.ApplyOrder) != 2 || finalized.ApplyOrder[0].Target.TargetID != right.String() || finalized.ApplyOrder[1].Target.TargetID != left.String() {
		t.Fatalf("reverse mutation apply order = %#v", finalized.ApplyOrder)
	}
	for index, step := range finalized.ApplyOrder {
		if step.Order != index+1 || step.ProviderID != "row/host" || step.TargetIdentity == "" || !reflect.DeepEqual(step.ChangedFieldKeys, []string{"host.name"}) {
			t.Fatalf("explicit plan step %d = %#v", index, step)
		}
	}
}

func TestRollbackRecordLockerCanonicalizesOrder_Unit(t *testing.T) {
	first := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	second := uuid.MustParse("00000000-0000-4000-8000-000000000002")
	got := (rollbackRecordLocker{}).orderedRecordIDs([]uuid.UUID{second, uuid.Nil, first, second})
	if !reflect.DeepEqual(got, []uuid.UUID{first, second}) {
		t.Fatalf("ordered rollback locks = %v", got)
	}
}
