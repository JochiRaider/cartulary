package timeline_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/internal/sourcevalue"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

const timelineEventPath = "data/timeline_records.ndjson"
const timelineTextFields = "date_entered_text analyst_text mitre_stage_text device_object_text ip_address_text activity_utc_text activity_local_text raw_activity_text activity_synopsis_text data_source_text"

func timelineEventRow(incident uuid.UUID, index int) map[string]any {
	row := map[string]any{"record_id": fmt.Sprintf("10000000-0000-4000-8000-%012d", index), "incident_id": incident.String(), "capture_state": "rough", "reviewed_by_user_id": nil, "reviewed_at": nil, "superseded_by_user_id": nil, "superseded_at": nil, "activity_utc_generated": false, "activity_local_generated": false, "activity_time_pair_state": "disabled"}
	for _, field := range strings.Fields(timelineTextFields) {
		row[field] = " " + field + " e\u0301\r\n\tkept "
	}
	return row
}
func timelineEventBundle(t testing.TB, rows ...map[string]any) portabilitysupport.MapBundle {
	t.Helper()
	bundle := timelineProfileBundle(t)
	var raw bytes.Buffer
	for _, row := range rows {
		if err := json.NewEncoder(&raw).Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	bundle[timelineEventPath] = raw.Bytes()
	return bundle
}

func TestTimelinePortableEventAdmission_Unit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	incident, actor := uuid.New(), uuid.New()
	input := timelineProfileInput(t, incident, actor)
	port := timeline.NewIncidentBundleSourcePort()
	base := timelineEventRow(incident, 1)
	prepare := func(rows ...map[string]any) error {
		bundle := timelineEventBundle(t, rows...)
		before := maps.Clone(bundle)
		for path, raw := range before {
			before[path] = bytes.Clone(raw)
		}
		_, err := port.PrepareImport(ctx, bundle, input)
		if !reflect.DeepEqual(before, bundle) {
			t.Fatal("event admission mutated bytes")
		}
		return err
	}
	if err := prepare(base); err != nil {
		t.Fatal(err)
	}
	if err := prepare(); err != nil {
		t.Fatal(err)
	}
	required := map[string]bool{"record_id": true, "incident_id": true, "capture_state": true, "activity_utc_generated": true, "activity_local_generated": true, "activity_time_pair_state": true}
	for field := range base {
		for _, variant := range []string{"missing", "type", "null"} {
			t.Run(field+"/"+variant, func(t *testing.T) {
				row := maps.Clone(base)
				switch variant {
				case "missing":
					delete(row, field)
				case "type":
					row[field] = []any{}
				case "null":
					row[field] = nil
				}
				err := prepare(row)
				if variant == "null" && !required[field] {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				invariant := "version_shape_exact"
				if field == "record_id" {
					invariant = "source_identity_admitted"
				}
				requireTimelineProfileFailure(t, err, invariant)
			})
		}
	}
	for _, field := range []string{"record_id", "incident_id", "reviewed_by_user_id", "superseded_by_user_id"} {
		for _, value := range []any{uuid.Nil.String(), "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", "aaaaaaaaaaaa4aaa8aaaaaaaaaaaaaaa", "bad", true} {
			row := maps.Clone(base)
			row[field] = value
			invariant := "version_shape_exact"
			if field == "record_id" {
				invariant = "source_identity_admitted"
			}
			requireTimelineProfileFailure(t, prepare(row), invariant)
		}
	}
	for _, field := range strings.Fields(timelineTextFields) {
		invalid := maps.Clone(base)
		invalid[field] = string([]byte{0xff})
		if _, ok := sourcevalue.Parse(invalid); ok {
			t.Fatal("invalid source UTF-8 admitted", field)
		}
		for _, value := range []string{"", strings.Repeat("界", 32768), "e\u0301\r\n\t<raw>\u200d"} {
			row := maps.Clone(base)
			row[field] = value
			if err := prepare(row); err != nil {
				t.Fatal(field, err)
			}
		}
		for _, value := range []string{strings.Repeat("界", 32769), "bad\x00text", "bad\x01text", "bad\x7ftext"} {
			row := maps.Clone(base)
			row[field] = value
			requireTimelineProfileFailure(t, prepare(row), "version_shape_exact")
		}
	}
	for _, field := range []string{"reviewed_at", "superseded_at"} {
		for _, stamp := range []string{"0000-01-01T00:00:00Z", "2026-01-01T00:00:00.1234567Z", "2026-01-01T00:00:00.100000Z", "2026-01-01T01:00:00+01:00"} {
			row := maps.Clone(base)
			row[field] = stamp
			requireTimelineProfileFailure(t, prepare(row), "version_shape_exact")
		}
	}
	for _, state := range []string{"rough", "enriched", "reviewed", "superseded", "unknown"} {
		for mask := 0; mask < 16; mask++ {
			t.Run(fmt.Sprintf("lifecycle/%s/%d", state, mask), func(t *testing.T) {
				row := maps.Clone(base)
				row["capture_state"] = state
				if mask&1 != 0 {
					row["reviewed_by_user_id"] = timelineSourceActor.String()
				}
				if mask&2 != 0 {
					row["reviewed_at"] = "2026-01-01T00:00:00Z"
				}
				if mask&4 != 0 {
					row["superseded_by_user_id"] = timelineSourceActor.String()
				}
				if mask&8 != 0 {
					row["superseded_at"] = "2026-01-02T00:00:00.123456+00:00"
				}
				allowed := ((state == "rough" || state == "enriched") && mask == 0) || (state == "reviewed" && mask == 3) || (state == "superseded" && (mask == 12 || mask == 15))
				err := prepare(row)
				if allowed {
					if err != nil {
						t.Fatal(err)
					}
				} else {
					requireTimelineProfileFailure(t, err, "lifecycle_coherent")
				}
			})
		}
	}
	for _, tc := range []struct {
		name, utc, local, pair       string
		utcGenerated, localGenerated bool
		invariant                    string
	}{
		{"disabled retained UTC", "2026-01-01T00:00:00Z", "", "disabled", true, false, ""},
		{"unavailable retained local", "", "2025-12-31T19:00:00-05:00", "conversion_unavailable", false, true, ""},
		{"both generated", "2026-01-01T00:00:00Z", "2025-12-31T19:00:00-05:00", "paired_generated", true, true, "generated_time_coherent"},
		{"bad UTC", "not a date", "local", "paired_generated", true, false, "generated_time_coherent"},
		{"noncanonical UTC", "2026-01-01 00:00:00", "local", "paired_generated", true, false, "generated_time_coherent"},
		{"bad local", "utc", "not a date", "paired_generated", false, true, "generated_time_coherent"},
		{"empty", "", "", "empty", false, false, ""},
		{"nonempty empty", "utc", "", "empty", false, false, "paired_time_coherent"},
		{"paired UTC", "2026-01-01T00:00:00Z", "user local", "paired_generated", true, false, ""},
		{"paired local", "user utc", "2025-12-31T19:00:00-05:00", "paired_generated", false, true, ""},
		{"paired without generated", "utc", "local", "paired_generated", false, false, "paired_time_coherent"},
		{"preserved", "utc", "local", "paired_user_preserved", false, false, ""},
		{"preserved generated", "2026-01-01T00:00:00Z", "local", "paired_user_preserved", true, false, "paired_time_coherent"},
		{"mismatch", "utc", "local", "paired_mismatch", false, false, ""},
		{"mismatch generated", "2026-01-01T00:00:00Z", "local", "paired_mismatch", true, false, ""},
		{"missing pair member", "utc", "", "paired_mismatch", false, false, "paired_time_coherent"},
		{"unknown pair", "utc", "local", "unknown", false, false, "paired_time_coherent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row := maps.Clone(base)
			row["activity_utc_text"], row["activity_local_text"], row["activity_time_pair_state"], row["activity_utc_generated"], row["activity_local_generated"] = tc.utc, tc.local, tc.pair, tc.utcGenerated, tc.localGenerated
			err := prepare(row)
			if tc.invariant == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				requireTimelineProfileFailure(t, err, tc.invariant)
			}
		})
	}
	for _, field := range []string{"activity_utc_generated", "activity_local_generated"} {
		row := maps.Clone(base)
		row[field] = "false"
		requireTimelineProfileFailure(t, prepare(row), "version_shape_exact")
	}
	foreign := maps.Clone(base)
	foreign["incident_id"] = uuid.NewString()
	requireTimelineProfileFailure(t, prepare(foreign), "envelope_type_scope")
	actorless := maps.Clone(base)
	actorless["capture_state"], actorless["reviewed_at"], actorless["reviewed_by_user_id"] = "reviewed", "2026-01-01T00:00:00Z", uuid.NewString()
	requireTimelineProfileFailure(t, prepare(actorless), "event_attribution_exact")
	second := timelineEventRow(incident, 2)
	requireTimelineProfileFailure(t, prepare(base, base), "source_identity_admitted")
	requireTimelineProfileFailure(t, prepare(second, base), "source_identity_admitted")
	badLifecycle := maps.Clone(base)
	badLifecycle["capture_state"] = "bad"
	second["analyst_text"] = 42
	requireTimelineProfileFailure(t, prepare(badLifecycle, second), "version_shape_exact")
	for _, field := range []string{"row_version", "recorded_at", "edited_at", "created_by_user_id", "updated_by_user_id", "unknown"} {
		row := maps.Clone(base)
		row[field] = true
		requireTimelineProfileFailure(t, prepare(row), "version_shape_exact")
	}
	for _, raw := range []string{"null\n", "[]\n", "{} {}\n", "{\"record_id\":\"a\",\"record_id\":\"b\"}\n", "{\"raw_activity_text\":\"\\ud800\"}\n"} {
		bundle := timelineEventBundle(t)
		bundle[timelineEventPath] = []byte(raw)
		_, err := port.PrepareImport(ctx, bundle, input)
		requireTimelineProfileFailure(t, err, "version_shape_exact")
	}
}
