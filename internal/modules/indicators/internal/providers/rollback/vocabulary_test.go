package rollback

import (
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestRollbackVocabularyAndLifecycleSupportReferencesAreExact(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"DOMAIN_NAME", " domain_name", "domain_name ", "domain", "unknown"} {
		if err := (Provider{}).ValidateRollbackValue(map[string]any{"source": map[string]any{"indicator_type": value}}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("rollback Indicator type %q = %v", value, err)
		}
	}
	for _, value := range []string{"ATOMIC", " atomic", "atomic ", "literal", "unknown"} {
		if err := (Provider{}).ValidateRollbackValue(map[string]any{"source": map[string]any{"value_kind": value}}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("rollback value kind %q = %v", value, err)
		}
	}
	for _, value := range []string{"UNRESOLVED", " unresolved", "unresolved ", "open", "unknown"} {
		observation := validRollbackObservationValue()
		observation["resolution_status"] = value
		if _, err := parseChildValue("indicator_observation", observation); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("rollback observation status %q = %v", value, err)
		}
	}
	for _, value := range []string{"ACTIVE", " active", "active ", "inactive", "unknown"} {
		interval := validRollbackIntervalValue()
		interval["lifecycle_state"] = value
		if _, err := parseChildValue("indicator_state_interval", interval); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("rollback lifecycle state %q = %v", value, err)
		}
	}
	interval := validRollbackIntervalValue()
	const supportID = "00000000-0000-4000-8000-000000000005"
	interval["support_refs"] = []any{supportID, supportID}
	if _, err := parseChildValue("indicator_state_interval", interval); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("rollback duplicate support refs = %v", err)
	}
}

func validRollbackObservationValue() map[string]any {
	return map[string]any{
		"incident_id": "00000000-0000-4000-8000-000000000001", "row_version": float64(1),
		"indicator_observation_id": "00000000-0000-4000-8000-000000000002",
		"source_record_id":         "00000000-0000-4000-8000-000000000003", "source_field_key": "source_text",
		"resolution_status": "unresolved", "origin_kind": "manual_entry", "origin_locator": "rollback",
		"observed_text": "example.test", "created_by_user_id": "00000000-0000-4000-8000-000000000004",
		"created_at": "2026-08-22T12:00:00Z", "resolved_indicator_record_id": nil,
		"resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil,
		"deleted_at": nil, "deleted_by_user_id": nil,
	}
}

func validRollbackIntervalValue() map[string]any {
	return map[string]any{
		"incident_id": "00000000-0000-4000-8000-000000000001", "row_version": float64(1),
		"indicator_state_interval_id": "00000000-0000-4000-8000-000000000002",
		"indicator_record_id":         "00000000-0000-4000-8000-000000000003", "lifecycle_state": "active",
		"valid_from": "2026-08-22T12:00:00Z", "support_refs": []any{},
		"created_by_user_id": "00000000-0000-4000-8000-000000000004",
		"created_at":         "2026-08-22T12:00:00Z", "deleted_at": nil, "deleted_by_user_id": nil,
	}
}
