package indicators_test

import (
	"encoding/json"
	"maps"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/intervalhistory"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/google/uuid"
)

func requireCompleteIntervalHistory(t *testing.T, target revisions.NonRowProviderContribution) {
	const id = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	const incident = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	const record = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	const actor = "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
	const support = "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee"
	const at = "2026-11-11T12:00:00.123456Z"
	const deleted = "2026-11-11T12:01:00.123456Z"
	base := map[string]any{"indicator_state_interval_id": id, "incident_id": incident, "indicator_record_id": record, "lifecycle_state": "active",
		"valid_from": at, "valid_to": nil, "confidence": nil, "rationale": nil, "support_refs": []string{}, "assessor": nil, "assessed_at": at,
		"row_version": int64(1), "created_by_user_id": actor, "created_at": at, "deleted_at": nil, "deleted_by_user_id": nil}
	dead := maps.Clone(base)
	dead["row_version"] = int64(2)
	dead["deleted_at"], dead["deleted_by_user_id"] = deleted, actor
	validator := target.HistoryValidator
	if validator == nil {
		t.Fatal("required interval validator missing")
	}
	check := func(name, operation, targetID string, before, after map[string]any, valid bool) {
		t.Run(name, func(t *testing.T) {
			mutation := revisions.StoredMutation{TargetKind: "indicator_state_interval", TargetID: targetID, OperationKind: operation, BeforeValue: before, AfterValue: after}
			original, err := json.Marshal(mutation)
			if err != nil {
				t.Fatal(err)
			}
			err = validator.ValidateHistoryMutation(mutation)
			if (err == nil) != valid {
				t.Fatalf("admission valid=%v error=%v", valid, err)
			}
			unchanged, err := json.Marshal(mutation)
			if err != nil || string(original) != string(unchanged) {
				t.Fatal("validator mutated retained values")
			}
		})
	}
	check("create", "create", id, nil, base, true)
	check("inverse", "rollback", id, base, dead, true)
	clock, err := time.Parse(time.RFC3339Nano, at)
	if err != nil {
		t.Fatal(err)
	}
	value := intervalhistory.Value{IntervalID: uuid.MustParse(id), IncidentID: uuid.MustParse(incident), IndicatorRecordID: uuid.MustParse(record), LifecycleState: "active", ValidFrom: clock,
		SupportRefs: []uuid.UUID{}, AssessedAt: clock, RowVersion: 1, CreatedByUserID: uuid.MustParse(actor), CreatedAt: clock}
	if got := value.Map(); !reflect.DeepEqual(got, base) {
		t.Fatalf("encoded interval=%#v want=%#v", got, base)
	}
	death, _ := time.Parse(time.RFC3339Nano, deleted)
	value.DeletedAt = &death
	value.DeletedByUserID = &value.CreatedByUserID
	value.RowVersion = 2
	if got := value.Map(); !reflect.DeepEqual(got, dead) {
		t.Fatalf("encoded tombstone=%#v want=%#v", got, dead)
	}
	for _, state := range []string{"active", "benign", "false_positive", "retired"} {
		v := maps.Clone(base)
		v["lifecycle_state"] = state
		check(state, "create", id, nil, v, true)
	}
	for _, confidence := range []any{0, 100, json.Number("50")} {
		v := maps.Clone(base)
		v["confidence"] = confidence
		v["rationale"], v["assessor"] = "", ""
		v["support_refs"] = []any{support}
		v["valid_to"] = at
		check("optional_values", "create", id, nil, v, true)
	}
	for _, key := range []string{"indicator_state_interval_id", "incident_id", "indicator_record_id", "lifecycle_state", "valid_from", "valid_to", "confidence", "rationale", "support_refs", "assessor", "assessed_at", "row_version", "created_by_user_id", "created_at", "deleted_at", "deleted_by_user_id"} {
		for _, side := range []string{"before", "after"} {
			for _, defect := range []string{"missing", "wrong_type"} {
				before, after := maps.Clone(base), maps.Clone(dead)
				v := before
				if side == "after" {
					v = after
				}
				if defect == "missing" {
					delete(v, key)
				} else {
					v[key] = true
				}
				check(side+"_"+key+"_"+defect, "rollback", id, before, after, false)
			}
		}
	}
	for _, bad := range []struct {
		name, key string
		value     any
	}{
		{"extra", "extra", nil}, {"nil_id", "indicator_state_interval_id", nil}, {"zero_id", "incident_id", uuid.Nil.String()}, {"uppercase_uuid", "indicator_record_id", strings.ToUpper(record)},
		{"bare_uuid", "created_by_user_id", strings.ReplaceAll(actor, "-", "")}, {"invalid_uuid", "deleted_by_user_id", "no"},
		{"unknown_state", "lifecycle_state", "inactive"}, {"padded_state", "lifecycle_state", " active"}, {"uppercase_state", "lifecycle_state", "ACTIVE"},
		{"offset_clock", "created_at", "2026-11-11T12:00:00.123456+00:00"}, {"zero_year", "created_at", "0000-01-01T00:00:00Z"}, {"submicro_clock", "assessed_at", "2026-11-11T12:00:00.1234567Z"},
		{"noncanonical_clock", "valid_from", "2026-11-11T12:00:00.000000Z"}, {"reversed_validity", "valid_to", "2026-11-11T11:59:59Z"}, {"reversed_deletion", "deleted_at", "2026-11-11T11:59:59Z"},
		{"missing_deletion_actor", "deleted_by_user_id", nil}, {"missing_deletion_clock", "deleted_at", nil}, {"negative_confidence", "confidence", -1}, {"high_confidence", "confidence", 101}, {"float_confidence", "confidence", float64(1)},
		{"nul_text", "rationale", "bad\x00text"}, {"invalid_utf8", "assessor", string([]byte{255})}, {"null_support", "support_refs", nil}, {"nil_array", "support_refs", []string(nil)},
		{"duplicate_support", "support_refs", []any{support, support}}, {"invalid_support", "support_refs", []any{"no"}}, {"wrong_support_scalar", "support_refs", []any{1}},
		{"zero_version", "row_version", 0}, {"float_version", "row_version", float64(2)}, {"fraction_version", "row_version", json.Number("2.0")}, {"leading_zero_version", "row_version", json.Number("02")}, {"overflow_version", "row_version", json.Number("9223372036854775808")},
	} {
		before, after := maps.Clone(base), maps.Clone(dead)
		after[bad.key] = bad.value
		// Some malformed JSON-number/UTF8 cases cannot be marshalled; call the pure
		// admission boundary directly with a valid control already established above.
		if bad.name == "leading_zero_version" || bad.name == "invalid_utf8" {
			if err := validator.ValidateHistoryMutation(revisions.StoredMutation{TargetKind: "indicator_state_interval", TargetID: id, OperationKind: "rollback", BeforeValue: before, AfterValue: after}); err == nil {
				t.Fatalf("%s admitted", bad.name)
			}
			continue
		}
		check(bad.name, "rollback", id, before, after, false)
	}
	for _, change := range []struct {
		key   string
		value any
	}{{"indicator_state_interval_id", uuid.NewString()}, {"incident_id", uuid.NewString()}, {"indicator_record_id", uuid.NewString()}, {"lifecycle_state", "benign"}, {"valid_from", "2026-11-11T11:59:00Z"}, {"valid_to", at}, {"confidence", 50}, {"rationale", "changed"}, {"support_refs", []string{support}}, {"assessor", "other"}, {"assessed_at", "2026-11-11T11:59:00Z"}, {"created_by_user_id", uuid.NewString()}, {"created_at", "2026-11-11T11:59:00Z"}} {
		after := maps.Clone(dead)
		after[change.key] = change.value
		check("immutable_"+change.key, "rollback", id, base, after, false)
	}
	before, after := maps.Clone(base), maps.Clone(dead)
	before["row_version"] = json.Number("9223372036854775806")
	after["row_version"] = json.Number("9223372036854775807")
	check("exact_int64_boundary", "rollback", id, before, after, true)
	before["row_version"] = int64(math.MaxInt64)
	check("version_overflow", "rollback", id, before, after, false)
	check("noop", "rollback", id, base, base, false)
	check("reverse_tombstone", "rollback", id, dead, base, false)
	check("create_before", "create", id, base, base, false)
	check("create_deleted", "create", id, nil, dead, false)
	after = maps.Clone(base)
	after["row_version"] = 2
	check("create_version", "create", id, nil, after, false)
	after = maps.Clone(dead)
	after["row_version"] = 3
	check("skipped_version", "rollback", id, base, after, false)
	for _, targetID := range []string{"indicator_state_interval:" + id, strings.ToUpper(id), uuid.NewString(), ""} {
		check("target_grammar", "create", targetID, nil, base, false)
	}
	for _, op := range []string{"patch", "resolve", "delete", ""} {
		check("operation_grammar", op, id, base, dead, false)
	}
	if err := validator.ValidateHistoryMutation(revisions.StoredMutation{TargetKind: "indicator_observation", TargetID: id, OperationKind: "create", AfterValue: base}); err == nil {
		t.Fatal("wrong target kind admitted")
	}
}
