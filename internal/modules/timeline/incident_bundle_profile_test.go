package timeline_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

const timelineProfilePath = "data/timeline_time_profiles.ndjson"

var timelineSourceActor = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")

func timelineProfileRow(incident uuid.UUID) map[string]any {
	return map[string]any{"incident_id": incident.String(), "enabled": true, "local_offset_minutes": -300, "local_label": "UTC−05", "profile_version": json.Number("9007199254740993"), "updated_at": "2026-01-02T03:04:05.123456+00:00", "updated_by_user_id": timelineSourceActor.String()}
}
func timelineProfileBundle(t testing.TB, rows ...map[string]any) portabilitysupport.MapBundle {
	t.Helper()
	var data bytes.Buffer
	for _, row := range rows {
		if err := json.NewEncoder(&data).Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	return portabilitysupport.MapBundle{timelineProfilePath: data.Bytes(), "data/timeline_records.ndjson": nil, "data/timeline_source_provenance.ndjson": nil}
}
func timelineProfileInput(t testing.TB, incident, actor uuid.UUID) sourceport.ImportContext {
	t.Helper()
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: timelineSourceActor.String(), DisplayName: "Historical profile author"}})
	if err != nil {
		t.Fatal(err)
	}
	return sourceport.ImportContext{IncidentID: incident, ActorUserID: actor, OperationID: "timeline-profile", BundleVersion: 8, Actors: actors}
}
func requireTimelineProfileFailure(t testing.TB, err error, invariant string) {
	t.Helper()
	var failure *sourceport.Failure
	if !errors.As(err, &failure) || failure.FamilyID() != "timeline" || failure.InvariantID() != "timeline."+invariant {
		t.Fatalf("failure=%v want timeline.%s", err, invariant)
	}
}

type timelineProfileAttributions struct {
	local uuid.UUID
	rows  []incidentportability.ImportedAttribution
}

func (a *timelineProfileAttributions) RecordImportedAttribution(table, row, column, actor string) error {
	a.rows = append(a.rows, incidentportability.ImportedAttribution{SourceTable: table, SourceRowID: row, SourceColumn: column, SourceActorID: actor, LocalUserID: a.local})
	return nil
}
func (a *timelineProfileAttributions) VisitImportedAttributions(table, row string, visit func(incidentportability.ImportedAttribution) error) error {
	for _, value := range a.rows {
		if value.SourceTable == table && (row == "" || value.SourceRowID == row) {
			if err := visit(value); err != nil {
				return err
			}
		}
	}
	return nil
}

type forbiddenTimelineProfileTx struct{ pgx.Tx }

func TestTimelinePortableProfileAdmission_Unit(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	incident, local := uuid.New(), uuid.New()
	port := timeline.NewIncidentBundleSourcePort()
	input := timelineProfileInput(t, incident, local)
	base := timelineProfileRow(incident)
	prepare := func(rows ...map[string]any) error {
		bundle := timelineProfileBundle(t, rows...)
		before := maps.Clone(bundle)
		for path, raw := range before {
			before[path] = bytes.Clone(raw)
		}
		_, err := port.PrepareImport(ctx, bundle, input)
		if !reflect.DeepEqual(bundle, before) {
			t.Fatal("preparation mutated source bytes")
		}
		return err
	}
	if err := prepare(base); err != nil {
		t.Fatal(err)
	}
	if err := prepare(); err != nil {
		t.Fatal("absent profile", err)
	}
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
				if variant == "null" && (field == "local_label" || field == "updated_by_user_id") {
					if err := prepare(row); err != nil {
						t.Fatal(err)
					}
					return
				}
				invariant := "version_shape_exact"
				if field == "incident_id" {
					invariant = "source_identity_admitted"
				}
				if variant == "null" && field == "local_offset_minutes" {
					invariant = "profile_coherent"
				}
				requireTimelineProfileFailure(t, prepare(row), invariant)
			})
		}
	}
	for field, values := range map[string][]any{
		"incident_id":          {uuid.Nil.String(), "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", "aaaaaaaaaaaa4aaa8aaaaaaaaaaaaaaa", "bad", true},
		"updated_by_user_id":   {uuid.Nil.String(), "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", "aaaaaaaaaaaa4aaa8aaaaaaaaaaaaaaa", "bad", true},
		"enabled":              {"true", "false", 0, 1},
		"profile_version":      {0, -1, json.Number("9223372036854775808"), json.Number("1e2"), json.Number("1.0"), "2", true},
		"local_offset_minutes": {-841, 841, json.Number("-0"), json.Number("0.0"), json.Number("1e2"), "60", true},
		"local_label":          {"", " ", " padded", "padded ", "e\u0301", "bad\nline", "\x00bad", strings.Repeat("界", 513)},
		"updated_at":           {"0000-01-01T00:00:00Z", "2026-01-01T00:00:00.1234567Z", "2026-01-01T00:00:00.100000Z", "2026-01-01T01:00:00+01:00", "2026-01-01", "2026-01-01T00:00:00z"},
		"unknown":              {true},
	} {
		for _, value := range values {
			row := maps.Clone(base)
			row[field] = value
			invariant := "version_shape_exact"
			if field == "incident_id" {
				invariant = "source_identity_admitted"
			}
			requireTimelineProfileFailure(t, prepare(row), invariant)
		}
	}
	for field, values := range map[string][]any{
		"enabled": {false}, "profile_version": {1, json.Number("9223372036854775807")},
		"local_offset_minutes": {-840, 0, 840}, "local_label": {"Café\u200d", strings.Repeat("界", 512)},
		"updated_at": {"0001-01-01T00:00:00Z", "9999-12-31T23:59:59.999999+00:00", "2026-01-01T00:00:00.1Z"},
	} {
		for _, value := range values {
			row := maps.Clone(base)
			row[field] = value
			if err := prepare(row); err != nil {
				t.Fatal(field, value, err)
			}
		}
	}
	disabled := maps.Clone(base)
	disabled["enabled"], disabled["local_offset_minutes"] = false, nil
	if err := prepare(disabled); err != nil {
		t.Fatal(err)
	}
	foreign := maps.Clone(base)
	foreign["incident_id"] = uuid.NewString()
	requireTimelineProfileFailure(t, prepare(foreign), "profile_coherent")
	unknownActor := maps.Clone(base)
	unknownActor["updated_by_user_id"] = uuid.NewString()
	requireTimelineProfileFailure(t, prepare(unknownActor), "profile_attribution_exact")
	requireTimelineProfileFailure(t, prepare(base, base), "source_identity_admitted")
	for _, raw := range []string{"null\n", "[]\n", "{} {}\n", "\n", "{\"incident_id\":\"a\",\"incident_id\":\"b\"}\n", "{\"local_label\":\"\\ud800\"}\n", "{\"local_label\":\"" + string([]byte{0xff}) + "\"}\n"} {
		bundle := timelineProfileBundle(t)
		bundle[timelineProfilePath] = []byte(raw)
		_, err := port.PrepareImport(ctx, bundle, input)
		requireTimelineProfileFailure(t, err, "version_shape_exact")
	}
	bundle := timelineProfileBundle(t, base)
	prepared, err := port.PrepareImport(ctx, bundle, input)
	if err != nil {
		t.Fatal(err)
	}
	input.Attributions = &timelineProfileAttributions{local: local}
	for _, variant := range []string{"port", "incident", "actor", "operation", "version", "recorder", "nil tx"} {
		t.Run("binding/"+variant, func(t *testing.T) {
			other, target := input, port
			var tx pgx.Tx = forbiddenTimelineProfileTx{}
			switch variant {
			case "port":
				target = timeline.NewIncidentBundleSourcePort()
			case "incident":
				other.IncidentID = uuid.New()
			case "actor":
				other.ActorUserID = uuid.New()
			case "operation":
				other.OperationID = "different"
			case "version":
				other.BundleVersion = 7
			case "recorder":
				other.Attributions = nil
			case "nil tx":
				tx = nil
			}
			for _, call := range []func(context.Context, pgx.Tx, sourceport.Prepared, sourceport.ImportContext) error{target.ApplyImportTx, target.ValidateImportTx} {
				if err := call(ctx, tx, prepared, other); !errors.Is(err, sourceport.ErrPreparedBinding) {
					t.Fatalf("binding error=%v", err)
				}
			}
		})
	}
	bundle[timelineProfilePath] = append(bundle[timelineProfilePath], '\n')
	if err := port.ApplyImportTx(ctx, forbiddenTimelineProfileTx{}, prepared, input); err == nil {
		t.Fatal("changed sealed profile admitted")
	}
}
