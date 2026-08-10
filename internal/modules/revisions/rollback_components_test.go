package revisions

import (
	"errors"
	"os"
	"reflect"
	"strings"
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

func TestRollbackDecompositionBoundaries_Unit(t *testing.T) {
	for _, path := range []string{
		"rollback_apply.go",
		"rollback_coordinator.go",
		"rollback_locker.go",
		"rollback_model.go",
		"rollback_planner.go",
		"rollback_publication.go",
		"rollback_query_repository.go",
		"rollback_result.go",
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("required rollback component %s: %v", path, err)
		}
	}
	if _, err := os.Stat("rollback_store.go"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("monolithic rollback store must remain removed, stat error = %v", err)
	}
	for _, path := range []string{"rollback_apply.go", "rollback_coordinator.go", "rollback_planner.go", "rollback_result.go"} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, token := range []string{"SELECT ", "INSERT INTO ", "UPDATE ", "DELETE FROM "} {
			if strings.Contains(string(contents), token) {
				t.Fatalf("%s contains repository token %q", path, token)
			}
		}
	}
	apply, err := os.ReadFile("rollback_apply.go")
	if err != nil {
		t.Fatalf("read rollback applier: %v", err)
	}
	for _, token := range []string{".appender.", "a.rebuildProjectionsTx("} {
		if strings.Contains(string(apply), token) {
			t.Fatalf("rollback applier bypasses publication seam with %q", token)
		}
	}
	for _, path := range []string{"rollback_apply.go", "rollback_planner.go", "rollback_query_repository.go"} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, token := range []string{
			"RowProviderCatalog",
			"NonRowProviderCatalog",
			"rowRollbackProviders",
			"nonRowRollbackProviders",
			`case "record"`,
			`case "record_link"`,
			`case "record_tag"`,
			`case "entity_mention"`,
			`case "artifact_evidence_link"`,
			`case "evidence_locator"`,
		} {
			if strings.Contains(string(contents), token) {
				t.Fatalf("generic rollback component %s contains prohibited source dispatch %q", path, token)
			}
		}
	}
}
