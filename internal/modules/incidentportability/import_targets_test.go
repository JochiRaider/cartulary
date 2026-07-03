package incidentportability

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestImportNDJSONRejectsUnregisteredTargetDescriptor(t *testing.T) {
	err := ImportNDJSON(context.Background(), nil, ImportTargetDescriptor{
		LogicalBundlePath: "data/records.ndjson",
		TargetRelation:    "records",
	}, []byte(`{"record_id":"11111111-1111-4111-8111-111111111111","incident_id":"22222222-2222-4222-8222-222222222222","record_type":"timeline_event","row_version":1}`+"\n"), uuid.New(), nil)
	var verification *VerificationFailure
	if !errors.As(err, &verification) || verification.ReasonCode != "malformed_manifest" {
		t.Fatalf("unregistered target error got %T %v", err, err)
	}
}

func TestSourceRowIDUsesRegisteredTargetIdentity(t *testing.T) {
	row := map[string]any{
		"saved_view_id": "33333333-3333-4333-8333-333333333333",
		"record_id":     "44444444-4444-4444-8444-444444444444",
	}
	if got := SourceRowID(TargetSavedViews, row); got != "33333333-3333-4333-8333-333333333333" {
		t.Fatalf("saved view source row id got %q", got)
	}
	composite := map[string]any{"change_set_id": "55555555-5555-4555-8555-555555555555", "sequence_no": jsonNumber("2")}
	if got := SourceRowID(TargetChangeSetMutations, composite); got != "55555555-5555-4555-8555-555555555555:2" {
		t.Fatalf("composite source row id got %q", got)
	}
}

type jsonNumber string

func (n jsonNumber) String() string {
	return string(n)
}
