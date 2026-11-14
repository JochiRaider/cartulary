package rollback

import (
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
	"github.com/google/uuid"
	"maps"
	"testing"
)

func TestParseObservationTargetRequiresCompleteTombstones(t *testing.T) {
	t.Parallel()
	value := validRollbackObservationValue()
	target := rollbackcontract.NonRowTarget{TargetKind: "indicator_observation", TargetID: value["indicator_observation_id"].(string), OperationKind: "create", AfterValue: value}
	identity, _, _, err := parseChildTarget(target)
	if err != nil || identity.deletedAt != nil || identity.deletedBy != nil {
		t.Fatalf("complete active observation=%#v error=%v", identity, err)
	}
	for _, field := range []string{"deleted_at", "deleted_by_user_id"} {
		partial := maps.Clone(value)
		delete(partial, field)
		target.AfterValue = partial
		if _, _, _, err := parseChildTarget(target); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("missing %s admitted: %v", field, err)
		}
	}
}
func TestParseObservationResolveRejectsIdentityDrift(t *testing.T) {
	t.Parallel()
	before := validRollbackObservationValue()
	after := maps.Clone(before)
	after["resolution_status"], after["resolved_indicator_record_id"], after["resolved_by_user_id"], after["resolved_at"], after["resolution_method"], after["row_version"] = "resolved", uuid.NewString(), before["created_by_user_id"], before["created_at"], "indicators.observations.resolve", int64(2)
	target := rollbackcontract.NonRowTarget{TargetKind: "indicator_observation", TargetID: before["indicator_observation_id"].(string), OperationKind: "resolve", BeforeValue: before, AfterValue: after}
	if _, _, _, err := parseChildTarget(target); err != nil {
		t.Fatalf("valid resolve control: %v", err)
	}
	after["source_record_id"] = uuid.NewString()
	if _, _, _, err := parseChildTarget(target); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("identity drift error=%v", err)
	}
}
