package revisionassembly

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"
	"github.com/JochiRaider/cartulary/internal/testutil/historytest"
)

func assertCanonicalNonRowHistoryFixture(t *testing.T, target revisions.NonRowProviderContribution) {
	t.Helper()
	valid := historytest.NonRowMutation(target.TargetKind, historytest.RecordID, historytest.RecordID, historytest.RecordID, historytest.RecordID)
	if target.HistoryValidator == nil {
		t.Fatal("required owner validator absent")
	}
	if err := target.HistoryValidator.ValidateHistoryMutation(valid); err != nil {
		t.Fatalf("authored mutation invalid: %v", err)
	}
	assertHistoryProjection(t, target.HistoryProjector, historycontract.Facts{RecordID: historytest.RecordID, Operation: valid.OperationKind, Before: valid.BeforeValue, After: valid.AfterValue})
	baseline, err := target.HistoryProjector(historycontract.Facts{RecordID: historytest.RecordID, Operation: valid.OperationKind, Before: valid.BeforeValue, After: valid.AfterValue})
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []string{"before", "after"} {
		if side == "before" && valid.BeforeValue == nil {
			continue
		}
		for _, field := range []string{"upload_token", "credentials"} {
			invalid := historytest.NonRowMutation(target.TargetKind, historytest.RecordID, historytest.RecordID, historytest.RecordID, historytest.RecordID)
			value := invalid.AfterValue
			if side == "before" {
				value = invalid.BeforeValue
			}
			value[field] = "private-" + field
			if err := target.HistoryValidator.ValidateHistoryMutation(invalid); err == nil {
				t.Fatalf("private %s member %s admitted", side, field)
			}
			// The lower-level projector must also remain an explicit public
			// field projection, even when handed inadmissible private members.
			got, err := target.HistoryProjector(historycontract.Facts{RecordID: historytest.RecordID, Operation: invalid.OperationKind, Before: invalid.BeforeValue, After: invalid.AfterValue})
			if err != nil || !reflect.DeepEqual(got, baseline) {
				t.Fatalf("private %s member %s altered public facts: %v", side, field, err)
			}
		}
	}
}
