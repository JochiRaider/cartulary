package indicators_test

import (
	"encoding/json"
	"maps"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/observationhistory"
	referencefixture "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/google/uuid"
)

func requireCompleteObservationHistory(t *testing.T) {
	const id = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	const incident = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	const source = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	const actor = "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
	const indicator = "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee"
	const at = "2026-11-11T12:00:00.123456Z"
	const later = "2026-11-11T12:01:00.123456Z"
	base := map[string]any{"indicator_observation_id": id, "incident_id": incident, "source_record_id": source, "source_field_key": "timeline.raw_activity_text", "origin_kind": "manual_entry", "origin_locator": "record://source/span/0:14", "observed_text": "before.example", "parsed_indicator_type": "domain_name", "normalized_candidate": "before.example", "resolution_status": "unresolved", "resolved_indicator_record_id": nil, "row_version": int64(1), "created_by_user_id": actor, "created_at": at, "resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil, "deleted_at": nil, "deleted_by_user_id": nil}
	clock, _ := time.Parse(time.RFC3339Nano, at)
	ptr := func(s string) *string { return &s }
	want := observationhistory.Value{ObservationID: uuid.MustParse(id), IncidentID: uuid.MustParse(incident), SourceRecordID: uuid.MustParse(source), SourceFieldKey: "timeline.raw_activity_text", OriginKind: "manual_entry", OriginLocator: "record://source/span/0:14", ObservedText: "before.example", ParsedIndicatorType: ptr("domain_name"), NormalizedCandidate: ptr("before.example"), ResolutionStatus: "unresolved", RowVersion: 1, CreatedByUserID: uuid.MustParse(actor), CreatedAt: clock}
	got, err := observationhistory.Decode(base)
	if err != nil || !reflect.DeepEqual(got, want) || !reflect.DeepEqual(want.Map(), base) {
		t.Fatalf("complete value=%#v expected=%#v error=%v", got, want, err)
	}
	var validator revisions.HistoryValidator
	for _, target := range indicators.NewRevisionContribution(referencefixture.IndicatorRegistry{}).NonRowTargets {
		if target.TargetKind == "indicator_observation" {
			validator = target.HistoryValidator
		}
	}
	if validator == nil {
		t.Fatal("required Observation validator missing")
	}
	check := func(name, op, target string, before, after map[string]any, valid bool) {
		t.Run(name, func(t *testing.T) {
			originalBefore, originalAfter := maps.Clone(before), maps.Clone(after)
			err := validator.ValidateHistoryMutation(revisions.StoredMutation{TargetKind: "indicator_observation", TargetID: target, OperationKind: op, BeforeValue: before, AfterValue: after})
			if (err == nil) != valid {
				t.Fatalf("admission valid=%v error=%v", valid, err)
			}
			if !reflect.DeepEqual(before, originalBefore) || !reflect.DeepEqual(after, originalAfter) {
				t.Fatal("validator changed retained values")
			}
		})
	}
	resolved := maps.Clone(base)
	resolved["resolution_status"], resolved["resolved_indicator_record_id"], resolved["resolved_by_user_id"], resolved["resolved_at"], resolved["resolution_method"], resolved["row_version"] = "resolved", indicator, actor, later, "indicators.observations.resolve", int64(2)
	dismissed := maps.Clone(resolved)
	dismissed["resolution_status"], dismissed["resolved_indicator_record_id"], dismissed["resolution_method"] = "dismissed", nil, "indicators.observations.dismiss"
	restored := maps.Clone(base)
	restored["row_version"] = int64(3)
	dead := maps.Clone(base)
	dead["deleted_at"], dead["deleted_by_user_id"], dead["row_version"] = later, actor, int64(2)
	check("create", "create", id, nil, base, true)
	createdResolved := maps.Clone(resolved)
	createdResolved["row_version"] = int64(1)
	check("resolved_create", "create", id, nil, createdResolved, true)
	check("resolve", "resolve", id, base, resolved, true)
	check("dismiss_unresolved", "dismiss", id, base, dismissed, true)
	dismissedAfterResolved := maps.Clone(dismissed)
	dismissedAfterResolved["row_version"] = int64(3)
	check("dismiss_resolved", "dismiss", id, resolved, dismissedAfterResolved, true)
	check("restore", "restore", id, dismissed, restored, true)
	check("rollback_create", "rollback", id, base, dead, true)
	check("rollback_resolution", "rollback", id, resolved, restored, true)
	reresolved := maps.Clone(resolved)
	reresolved["resolved_indicator_record_id"], reresolved["row_version"] = source, int64(3)
	check("reresolve", "resolve", id, resolved, reresolved, true)
	back := maps.Clone(resolved)
	back["row_version"] = int64(4)
	check("rollback_reresolve", "rollback", id, reresolved, back, true)
	for _, origin := range []string{"manual_entry", "clipboard_paste", "csv_import", "xlsx_import", "api_import", "extraction", "system"} {
		v := maps.Clone(base)
		v["origin_kind"] = origin
		check("origin_"+origin, "create", id, nil, v, true)
	}
	raw := maps.Clone(base)
	raw["observed_text"], raw["origin_locator"], raw["source_field_key"] = " \t ", " locator ", " field "
	check("raw_text_preserved", "create", id, nil, raw, true)
	noParse := maps.Clone(base)
	noParse["parsed_indicator_type"], noParse["normalized_candidate"] = nil, nil
	check("no_parse", "create", id, nil, noParse, true)
	noCandidate := maps.Clone(base)
	noCandidate["normalized_candidate"] = nil
	check("type_guess_only", "create", id, nil, noCandidate, true)
	for key := range base {
		v := maps.Clone(base)
		delete(v, key)
		check("missing_"+key, "create", id, nil, v, false)
		v = maps.Clone(base)
		v[key] = false
		check("wrong_type_"+key, "create", id, nil, v, false)
	}
	for _, bad := range []struct {
		name, key string
		value     any
	}{
		{"extra", "extra", nil}, {"nil_identity", "indicator_observation_id", uuid.Nil.String()}, {"upper_identity", "indicator_observation_id", strings.ToUpper(id)}, {"compact_identity", "incident_id", strings.ReplaceAll(incident, "-", "")}, {"wrong_actor", "created_by_user_id", "invalid"},
		{"blank_clock", "created_at", ""}, {"offset_clock", "created_at", "2026-11-11T12:00:00.123456+00:00"}, {"year_zero", "created_at", "0000-01-01T00:00:00Z"}, {"nano_clock", "created_at", "2026-11-11T12:00:00.1234567Z"}, {"padded_fraction", "created_at", "2026-11-11T12:00:00.1234560Z"},
		{"empty_source_field", "source_field_key", ""}, {"empty_locator", "origin_locator", ""}, {"empty_observation", "observed_text", ""}, {"nul_text", "observed_text", "x\x00y"}, {"invalid_utf8", "origin_locator", string([]byte{255})},
		{"origin_alias", "origin_kind", "interactive_cell"}, {"origin_pad", "origin_kind", " manual_entry"}, {"type_alias", "parsed_indicator_type", "domain"}, {"type_pad", "parsed_indicator_type", "domain_name "}, {"candidate_no_type", "parsed_indicator_type", nil}, {"empty_candidate", "normalized_candidate", ""},
		{"unknown_status", "resolution_status", "unknown"}, {"incomplete_resolved", "resolution_status", "resolved"}, {"incomplete_dismissed", "resolution_status", "dismissed"}, {"unresolved_target", "resolved_indicator_record_id", indicator}, {"unresolved_actor", "resolved_by_user_id", actor}, {"unresolved_clock", "resolved_at", later}, {"unresolved_method", "resolution_method", "method"}, {"empty_optional_uuid", "resolved_indicator_record_id", ""}, {"empty_optional_time", "deleted_at", ""}, {"unpaired_tombstone", "deleted_at", later},
		{"zero_version", "row_version", int64(0)}, {"float_version", "row_version", float64(1)}, {"fraction_version", "row_version", json.Number("1.0")}, {"padded_version", "row_version", json.Number("01")}, {"overflow_version", "row_version", json.Number("9223372036854775808")},
	} {
		v := maps.Clone(base)
		v[bad.key] = bad.value
		check(bad.name, "create", id, nil, v, false)
	}
	check("create_before", "create", id, base, base, false)
	createdDead := maps.Clone(dead)
	createdDead["row_version"] = int64(1)
	check("created_dead", "create", id, nil, createdDead, false)
	dismissedCreate := maps.Clone(dismissed)
	dismissedCreate["row_version"] = int64(1)
	check("created_dismissed", "create", id, nil, dismissedCreate, false)
	check("alternate_target", "create", "indicator_observation:"+id, nil, base, false)
	check("wrong_target", "create", source, nil, base, false)
	check("unknown_operation", "update", id, base, resolved, false)
	check("resolve_dismissed", "resolve", id, dismissed, reresolved, false)
	unresolvedNext := maps.Clone(base)
	unresolvedNext["row_version"] = int64(2)
	check("restore_unresolved", "restore", id, base, unresolvedNext, false)
	check("dismiss_dismissed", "dismiss", id, dismissed, dismissedAfterResolved, false)
	sameTarget := maps.Clone(resolved)
	sameTarget["row_version"] = int64(3)
	check("same_target", "resolve", id, resolved, sameTarget, false)
	check("noop_rollback", "rollback", id, resolved, sameTarget, false)
	for key, value := range map[string]any{"incident_id": source, "source_record_id": incident, "source_field_key": "other.field", "origin_kind": "csv_import", "origin_locator": "other", "observed_text": "other.example", "parsed_indicator_type": "text", "normalized_candidate": "other.example", "created_by_user_id": source, "created_at": "2026-11-11T11:00:00Z"} {
		v := maps.Clone(resolved)
		v[key] = value
		check("immutable_"+key, "resolve", id, base, v, false)
	}
	jump := maps.Clone(resolved)
	jump["row_version"] = int64(4)
	check("version_jump", "resolve", id, base, jump, false)
	for _, test := range []struct {
		input any
		want  int64
	}{{int64(9007199254740993), 9007199254740993}, {json.Number("9007199254740993"), 9007199254740993}, {int64(math.MaxInt64), math.MaxInt64}} {
		v := maps.Clone(base)
		v["row_version"] = test.input
		got, err := observationhistory.Decode(v)
		if err != nil || got.RowVersion != test.want {
			t.Fatalf("exact large version=%d error=%v", got.RowVersion, err)
		}
	}
	overflow := maps.Clone(base)
	overflow["row_version"] = int64(math.MaxInt64)
	check("overflow_pair", "resolve", id, overflow, resolved, false)
	for _, defect := range []string{"missing_actor", "empty_method", "resolved_before_create", "deleted_before_resolve"} {
		v := maps.Clone(resolved)
		switch defect {
		case "missing_actor":
			v["resolved_by_user_id"] = nil
		case "empty_method":
			v["resolution_method"] = ""
		case "resolved_before_create":
			v["resolved_at"] = "2026-11-11T11:00:00Z"
		case "deleted_before_resolve":
			v["deleted_at"], v["deleted_by_user_id"] = at, actor
		}
		if _, err := observationhistory.Decode(v); err == nil {
			t.Fatalf("illegal tuple %s admitted", defect)
		}
	}
	tombstoneDrift := maps.Clone(dead)
	tombstoneDrift["observed_text"] = "changed"
	check("tombstone_source_drift", "rollback", id, base, tombstoneDrift, false)
}
