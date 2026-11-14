package entities_test

import (
	"encoding/json"
	"maps"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/aliashistory"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func requireCanonicalAliasHistory(t *testing.T, contribution revisions.NonRowProviderContribution) {
	t.Helper()
	const id = "11111111-1111-4111-8111-111111111111"
	const incident = "22222222-2222-4222-8222-222222222222"
	const record = "33333333-3333-4333-8333-333333333333"
	const actor = "44444444-4444-4444-8444-444444444444"
	base := map[string]any{"entity_alias_id": id, "incident_id": incident, "record_id": record, "entity_type": "host",
		"raw_text": "Alias α", "normalized_text": "Alias α", "classification": "suggestion_only",
		"created_by_user_id": actor, "created_at": "2026-11-11T12:00:00.123456Z", "deleted_at": nil}
	deleted := maps.Clone(base)
	deleted["deleted_at"] = "2026-11-12T12:00:00Z"
	target := "entity_alias:" + id
	validator := contribution.HistoryValidator
	if validator == nil {
		t.Fatal("missing Alias history validator")
	}
	check := func(name, operation, targetID string, before, after map[string]any, valid bool) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			mutation := revisions.StoredMutation{TargetKind: "entity_alias", TargetID: targetID, OperationKind: operation, BeforeValue: before, AfterValue: after}
			original, err := json.Marshal(mutation)
			if err != nil {
				t.Fatal(err)
			}
			err = validator.ValidateHistoryMutation(mutation)
			if (err == nil) != valid {
				t.Fatalf("valid=%v error=%v", valid, err)
			}
			current, err := json.Marshal(mutation)
			if err != nil || string(current) != string(original) {
				t.Fatal("retained Alias changed")
			}
		})
	}
	check("create", "create", target, nil, base, true)
	check("delete", "delete", target, base, deleted, true)
	check("rollback_create", "rollback", target, base, deleted, true)
	check("rollback_delete", "rollback", target, deleted, base, true)
	for _, entity := range []string{"host", "identity"} {
		value := maps.Clone(base)
		value["entity_type"] = entity
		check(entity, "create", target, nil, value, true)
	}
	max := maps.Clone(base)
	max["raw_text"], max["normalized_text"] = strings.Repeat("𐐀", 256), strings.Repeat("𐐀", 256)
	check("256_scalars", "create", target, nil, max, true)
	equalClock := maps.Clone(deleted)
	equalClock["deleted_at"] = base["created_at"]
	check("equal_lifetime_clocks", "delete", target, base, equalClock, true)
	createdAt := time.Date(2026, 11, 11, 12, 0, 0, 123456000, time.UTC)
	deletedAt := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	value := aliashistory.Value{ID: uuid.MustParse(id), IncidentID: uuid.MustParse(incident), RecordID: uuid.MustParse(record), CreatedByUserID: uuid.MustParse(actor), EntityType: "host", RawText: "Alias α", NormalizedText: "Alias α", Classification: "suggestion_only", CreatedAt: createdAt}
	if got := value.Map(); !reflect.DeepEqual(got, base) || value.TargetID() != target {
		t.Fatalf("active Alias=%#v target=%s", got, value.TargetID())
	}
	value.DeletedAt = &deletedAt
	if got := value.Map(); !reflect.DeepEqual(got, deleted) {
		t.Fatalf("deleted Alias=%#v", got)
	}
	for _, side := range []string{"before", "after"} {
		for field := range base {
			for _, defect := range []string{"missing", "wrong_type"} {
				left, right := maps.Clone(base), maps.Clone(deleted)
				selected := left
				if side == "after" {
					selected = right
				}
				if defect == "missing" {
					delete(selected, field)
				} else {
					selected[field] = 42
				}
				check(side+"/"+field+"/"+defect, "delete", target, left, right, false)
			}
		}
	}
	for _, test := range []struct {
		name, field string
		value       any
	}{
		{"extra", "projection_label", "label"}, {"zero_id", "entity_alias_id", uuid.Nil.String()},
		{"uppercase", "incident_id", "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA"}, {"wrong_entity", "entity_type", "party"},
		{"wrong_classification", "classification", "exact_match_reuse"}, {"blank_raw", "raw_text", " \t"}, {"NUL_raw", "raw_text", "bad\x00"},
		{"noncanonical_raw", "raw_text", " Alias α "}, {"decomposed", "raw_text", "A\u0301lias"}, {"case_folded", "normalized_text", "alias α"},
		{"too_long", "raw_text", strings.Repeat("𐐀", 257)}, {"year_zero", "created_at", "0000-01-01T00:00:00Z"},
		{"submicrosecond", "created_at", "2026-11-11T12:00:00.1234567Z"}, {"offset", "created_at", "2026-11-11T12:00:00+00:00"},
		{"trailing_zero", "created_at", "2026-11-11T12:00:00.1234560Z"},
	} {
		for _, side := range []string{"before", "after"} {
			left, right := maps.Clone(base), maps.Clone(deleted)
			selected := left
			if side == "after" {
				selected = right
			}
			selected[test.field] = test.value
			check(side+"/"+test.name, "delete", target, left, right, false)
		}
	}
	for _, clock := range []string{"2026-11-11T12:00:00.123455Z", "2026-11-12T12:00:00.0000001Z", "2026-11-12T12:00:00+00:00", "bad"} {
		after := maps.Clone(deleted)
		after["deleted_at"] = clock
		check("deleted_clock/"+clock, "delete", target, base, after, false)
	}
	for _, field := range []string{"incident_id", "record_id", "created_by_user_id"} {
		after := maps.Clone(deleted)
		after[field] = uuid.NewString()
		check("immutable/"+field, "delete", target, base, after, false)
	}
	after := maps.Clone(deleted)
	after["entity_type"] = "identity"
	check("immutable/entity_type", "delete", target, base, after, false)
	after = maps.Clone(deleted)
	after["raw_text"], after["normalized_text"] = "Alias β", "Alias β"
	check("immutable/text", "delete", target, base, after, false)
	after = maps.Clone(deleted)
	after["created_at"] = "2026-11-10T12:00:00Z"
	check("immutable/created_at", "delete", target, base, after, false)
	for _, badTarget := range []string{id, "entity_mention:" + id, "entity_alias:" + uuid.NewString(), "entity_alias:cmVjb3Jk:aG9zdA:YWxpYXM", strings.ToUpper(target)} {
		check("target/"+badTarget, "create", badTarget, nil, base, false)
	}
	check("create_before", "create", target, base, base, false)
	check("create_deleted", "create", target, nil, deleted, false)
	check("delete_noop", "delete", target, base, base, false)
	check("delete_restore", "delete", target, deleted, base, false)
	check("rollback_noop", "rollback", target, deleted, deleted, false)
	check("missing_before", "rollback", target, nil, deleted, false)
	check("missing_after", "create", target, nil, nil, false)
	for _, operation := range []string{"patch", "restore", ""} {
		check("operation/"+operation, operation, target, base, deleted, false)
	}
	if err := validator.ValidateHistoryMutation(revisions.StoredMutation{TargetKind: "entity_preserved_identifier", TargetID: target, OperationKind: "create", AfterValue: base}); err == nil {
		t.Fatal("wrong kind admitted")
	}
}
