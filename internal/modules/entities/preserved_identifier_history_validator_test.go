package entities_test

import (
	"encoding/json"
	"maps"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/preservedhistory"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func requireCanonicalPreservedIdentifierHistory(t *testing.T, contribution revisions.NonRowProviderContribution) {
	t.Helper()
	const id = "11111111-1111-4111-8111-111111111111"
	const incident = "22222222-2222-4222-8222-222222222222"
	const record = "33333333-3333-4333-8333-333333333333"
	const actor = "44444444-4444-4444-8444-444444444444"
	base := map[string]any{
		"entity_preserved_identifier_id": id, "incident_id": incident, "record_id": record,
		"entity_type": "host", "identifier_type": "fqdn", "raw_value": "HOST.EXAMPLE.TEST", "normalized_value": "host.example.test",
		"classification": "exact_match_reuse", "created_by_user_id": actor, "created_at": "2026-11-11T12:00:00.123456Z", "deleted_at": nil,
	}
	deleted := maps.Clone(base)
	deleted["deleted_at"] = "2026-11-12T12:00:00Z"
	target := "entity_preserved_identifier:" + id
	validator := contribution.HistoryValidator
	if validator == nil {
		t.Fatal("missing required preserved-identifier history validator")
	}
	check := func(name, operation, targetID string, before, after map[string]any, valid bool) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			mutation := revisions.StoredMutation{TargetKind: "entity_preserved_identifier", TargetID: targetID, OperationKind: operation, BeforeValue: before, AfterValue: after}
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
				t.Fatal("retained input changed")
			}
		})
	}
	check("create", "create", target, nil, base, true)
	check("rollback", "rollback", target, base, deleted, true)
	equalDeletion := maps.Clone(deleted)
	equalDeletion["deleted_at"] = base["created_at"]
	check("rollback_equal_creation", "rollback", target, base, equalDeletion, true)
	earlyDeletion := maps.Clone(deleted)
	earlyDeletion["deleted_at"] = "2026-11-11T12:00:00.123455Z"
	check("rollback_before_creation", "rollback", target, base, earlyDeletion, false)

	for entity, classes := range map[string][]string{"host": {"aad_device_id", "fqdn", "hostname"}, "identity": {"aad_object_id", "sid", "upn", "email", "sam_account_name"}} {
		for _, class := range classes {
			for _, classification := range []string{"exact_match_reuse", "suggestion_only", "provenance_only"} {
				value := maps.Clone(base)
				value["entity_type"], value["identifier_type"], value["classification"] = entity, class, classification
				value["raw_value"], value["normalized_value"] = "ABCD", "abcd"
				if class == "sid" {
					value["normalized_value"] = "ABCD"
				}
				check(entity+"/"+class+"/"+classification, "create", target, nil, value, true)
			}
		}
	}
	createdAt := time.Date(2026, 11, 11, 12, 0, 0, 123456000, time.UTC)
	deletedAt := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	value := preservedhistory.Value{ID: uuid.MustParse(id), IncidentID: uuid.MustParse(incident), RecordID: uuid.MustParse(record), CreatedByUserID: uuid.MustParse(actor), EntityType: "host", IdentifierType: "fqdn", RawValue: "HOST.EXAMPLE.TEST", NormalizedValue: "host.example.test", Classification: "exact_match_reuse", CreatedAt: createdAt}
	if got := value.Map(); !reflect.DeepEqual(got, base) || value.TargetID() != target {
		t.Fatalf("active encoding=%#v target=%s", got, value.TargetID())
	}
	value.DeletedAt = &deletedAt
	if got := value.Map(); !reflect.DeepEqual(got, deleted) {
		t.Fatalf("deleted encoding=%#v", got)
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
				check(side+"/"+field+"/"+defect, "rollback", target, left, right, false)
			}
		}
	}
	for _, test := range []struct {
		name, field string
		value       any
	}{
		{"unknown_field", "projection_label", "extra"}, {"zero_id", "entity_preserved_identifier_id", uuid.Nil.String()},
		{"uppercase_incident", "incident_id", "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA"}, {"wrong_entity", "entity_type", "party"},
		{"wrong_class", "identifier_type", "email"}, {"wrong_classification", "classification", "canonical"},
		{"blank_raw", "raw_value", " \t"}, {"NUL_raw", "raw_value", "bad\x00"}, {"bad_normalization", "normalized_value", "HOST.EXAMPLE.TEST"},
		{"year_zero", "created_at", "0000-01-01T00:00:00Z"}, {"submicrosecond", "created_at", "2026-11-11T12:00:00.1234567Z"},
		{"offset", "created_at", "2026-11-11T12:00:00+00:00"}, {"trailing_zero", "created_at", "2026-11-11T12:00:00.1234560Z"},
	} {
		for _, side := range []string{"before", "after"} {
			left, right := maps.Clone(base), maps.Clone(deleted)
			selected := left
			if side == "after" {
				selected = right
			}
			selected[test.field] = test.value
			check(side+"/"+test.name, "rollback", target, left, right, false)
		}
	}
	for _, test := range []struct {
		field string
		value any
	}{
		{"incident_id", uuid.NewString()}, {"record_id", uuid.NewString()}, {"created_by_user_id", uuid.NewString()},
		{"raw_value", "Host.Example.Test"}, {"classification", "suggestion_only"}, {"identifier_type", "hostname"}, {"created_at", "2026-11-10T12:00:00Z"},
	} {
		right := maps.Clone(deleted)
		right[test.field] = test.value
		check("immutable/"+test.field, "rollback", target, base, right, false)
	}
	for _, badTarget := range []string{id, "entity_alias:" + id, "entity_preserved_identifier:" + uuid.NewString(), "entity_preserved_identifier:cmVjb3Jk:aG9zdA:ZnFkbg:dmFsdWU:Y2xhc3M", strings.ToUpper(target)} {
		check("target/"+badTarget, "create", badTarget, nil, base, false)
	}
	check("create_before", "create", target, base, base, false)
	check("create_deleted", "create", target, nil, deleted, false)
	check("rollback_missing_before", "rollback", target, nil, deleted, false)
	check("rollback_noop", "rollback", target, base, base, false)
	check("rollback_restore", "rollback", target, deleted, base, false)
	for _, operation := range []string{"patch", "delete", "restore", ""} {
		check("operation/"+operation, operation, target, base, deleted, false)
	}
	check("missing_after", "create", target, nil, nil, false)
	if err := validator.ValidateHistoryMutation(revisions.StoredMutation{TargetKind: "entity_alias", TargetID: target, OperationKind: "create", AfterValue: base}); err == nil {
		t.Fatal("wrong target kind admitted")
	}
}
