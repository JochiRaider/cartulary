package entities_test

import (
	"encoding/json"
	"maps"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions/historyvalue"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func requireCanonicalMentionHistory(t *testing.T, target revisions.NonRowProviderContribution) {
	t.Helper()
	if target.HistoryValidator == nil {
		t.Fatal("Mention history validator absent")
	}
	const mention = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	const source = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	const actor = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	const resolved = "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
	const created = "2026-11-13T12:00:00.123456Z"
	before := map[string]any{
		"entity_mention_id": mention, "source_record_id": source,
		"entity_type": "host", "source_field_key": "timeline.host_refs", "origin_kind": "manual_entry",
		"origin_locator": "cell", "raw_text": "Host A", "normalized_text": "host a",
		"resolution_status": "unresolved", "row_version": json.Number("1"), "ordinal": json.Number("1"),
		"created_by_user_id": actor, "created_at": created,
		"resolved_record_id": nil, "resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil,
	}
	after := maps.Clone(before)
	after["row_version"], after["resolution_status"] = json.Number("2"), "resolved"
	after["resolved_record_id"], after["resolved_by_user_id"] = resolved, actor
	after["resolved_at"], after["resolution_method"] = "2026-11-13T12:01:00Z", "explicit_resolve_route"
	check := func(t *testing.T, operation string, before, after map[string]any, valid bool) {
		t.Helper()
		mutation := revisions.StoredMutation{TargetKind: "entity_mention", TargetID: mention, OperationKind: operation, BeforeValue: before, AfterValue: after}
		bytes, err := json.Marshal(mutation)
		if err != nil {
			t.Fatal(err)
		}
		err = target.HistoryValidator.ValidateHistoryMutation(mutation)
		if (err == nil) != valid {
			t.Fatalf("admission valid=%t error=%v", valid, err)
		}
		unchanged, err := json.Marshal(mutation)
		if err != nil || string(bytes) != string(unchanged) {
			t.Fatal("retained facts were changed")
		}
	}
	check(t, "patch", before, after, true)
	restored := maps.Clone(before)
	restored["row_version"] = json.Number("3")
	check(t, "rollback", after, restored, true)
	for _, version := range []int64{9007199254740993, 9223372036854775806} {
		left, right := maps.Clone(before), maps.Clone(after)
		left["row_version"], right["row_version"] = json.Number(strconv.FormatInt(version, 10)), json.Number(strconv.FormatInt(version+1, 10))
		check(t, "patch", left, right, true)
	}
	at, err := time.Parse(time.RFC3339Nano, created)
	if err != nil {
		t.Fatal(err)
	}
	row := historyvalue.Value{MentionID: uuid.MustParse(mention), SourceRecordID: uuid.MustParse(source), EntityType: "host", SourceFieldKey: "timeline.host_refs", OriginKind: "manual_entry", OriginLocator: "cell", RawText: "Host A", NormalizedText: "host a", ResolutionStatus: "unresolved", RowVersion: 1, Ordinal: 1, CreatedByUserID: uuid.MustParse(actor), CreatedAt: at}
	encoded := row.Map()
	wantBytes, _ := json.Marshal(before)
	gotBytes, err := json.Marshal(encoded)
	if err != nil || string(gotBytes) != string(wantBytes) {
		t.Fatalf("complete owner encoding=%s want=%s error=%v", gotBytes, wantBytes, err)
	}
	resolvedID, resolver := uuid.MustParse(resolved), uuid.MustParse(actor)
	resolvedAt := time.Date(2026, 11, 13, 12, 1, 0, 0, time.UTC)
	method := "explicit_resolve_route"
	row.RowVersion, row.ResolutionStatus = 2, "resolved"
	row.ResolvedRecordID, row.ResolvedByUserID = &resolvedID, &resolver
	row.ResolvedAt, row.ResolutionMethod = &resolvedAt, &method
	wantBytes, err = json.Marshal(after)
	if err != nil {
		t.Fatal(err)
	}
	gotBytes, err = json.Marshal(row.Map())
	if err != nil || string(gotBytes) != string(wantBytes) {
		t.Fatalf("resolved owner encoding=%s want=%s error=%v", gotBytes, wantBytes, err)
	}
	for _, side := range []string{"before", "after"} {
		for field := range before {
			for _, defect := range []string{"missing", "type"} {
				t.Run(side+"/"+field+"/"+defect, func(t *testing.T) {
					left, right := maps.Clone(before), maps.Clone(after)
					value := left
					if side == "after" {
						value = right
					}
					if defect == "missing" {
						delete(value, field)
					} else {
						value[field] = []any{}
					}
					check(t, "patch", left, right, false)
				})
			}
		}
	}
	for _, test := range []struct {
		name, field string
		value       any
	}{
		{"derived incident", "incident_id", source},
		{"zero ID", "entity_mention_id", uuid.Nil.String()},
		{"upper ID", "source_record_id", "BBBBBBBB-BBBB-4BBB-8BBB-BBBBBBBBBBBB"},
		{"type vocabulary", "entity_type", "evidence"},
		{"origin vocabulary", "origin_kind", "unknown"},
		{"status vocabulary", "resolution_status", "active"},
		{"blank field", "source_field_key", " \t"},
		{"NUL observed text", "raw_text", "bad\x00"},
		{"blank method", "resolution_method", " "},
		{"wrong tuple", "resolved_record_id", nil},
		{"zero ordinal", "ordinal", json.Number("0")},
		{"large ordinal", "ordinal", json.Number("2147483648")},
		{"version gap", "row_version", json.Number("3")},
		{"float version", "row_version", float64(2)},
		{"fraction version", "row_version", json.Number("2.0")},
		{"overflow version", "row_version", json.Number("9223372036854775808")},
		{"year zero", "resolved_at", "0000-01-01T00:00:00Z"},
		{"noncanonical UTC", "resolved_at", "2026-11-13T12:01:00+00:00"},
		{"submicro time", "resolved_at", "2026-11-13T12:01:00.000000001Z"},
	} {
		t.Run(test.name, func(t *testing.T) {
			right := maps.Clone(after)
			right[test.field] = test.value
			check(t, "patch", before, right, false)
		})
	}
	for _, field := range []string{"source_record_id", "entity_type", "source_field_key", "origin_kind", "origin_locator", "raw_text", "normalized_text", "ordinal", "created_by_user_id", "created_at"} {
		t.Run("immutable/"+field, func(t *testing.T) {
			right := maps.Clone(after)
			switch field {
			case "source_record_id", "created_by_user_id":
				right[field] = resolved
			case "entity_type":
				right[field] = "identity"
			case "origin_kind":
				right[field] = "system"
			case "ordinal":
				right[field] = json.Number("2")
			case "created_at":
				right[field] = "2026-11-13T12:00:01Z"
			default:
				right[field] = "changed"
			}
			check(t, "patch", before, right, false)
		})
	}
	noChange := maps.Clone(before)
	noChange["row_version"] = json.Number("2")
	check(t, "patch", before, noChange, false)
	check(t, "patch", nil, after, false)
	check(t, "patch", before, nil, false)
	check(t, "create", nil, after, false)
	check(t, "delete", before, after, false)
	wrongTarget := revisions.StoredMutation{TargetKind: "entity_mention", TargetID: "entity_mention:" + mention, OperationKind: "patch", BeforeValue: before, AfterValue: after}
	if err := target.HistoryValidator.ValidateHistoryMutation(wrongTarget); err == nil {
		t.Fatal("alternate target grammar accepted")
	}
}
