package revisions_test

import (
	"context"
	"encoding/json"
	"maps"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func exerciseCompleteObservationHistory(t *testing.T) {
	h := newRevisionsPortabilityHarness(t, "observation-history")
	id, indicatorID := uuid.New(), uuid.New()
	active := map[string]any{"indicator_observation_id": id.String(), "incident_id": h.incidentID.String(), "source_record_id": h.recordID.String(), "source_field_key": "host.hostname", "origin_kind": "manual_entry", "origin_locator": "record://source/span/0:16", "observed_text": "observed.example", "parsed_indicator_type": "domain_name", "normalized_candidate": "observed.example", "resolution_status": "unresolved", "resolved_indicator_record_id": nil, "row_version": 1, "created_by_user_id": h.actor.ID.String(), "created_at": h.createdAt.Format(time.RFC3339Nano), "resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil, "deleted_at": nil, "deleted_by_user_id": nil}
	resolved := maps.Clone(active)
	resolved["row_version"], resolved["resolution_status"], resolved["resolved_indicator_record_id"], resolved["resolved_by_user_id"], resolved["resolved_at"], resolved["resolution_method"] = 2, "resolved", indicatorID.String(), h.actor.ID.String(), h.createdAt.Add(time.Minute).Format(time.RFC3339Nano), "indicators.observations.resolve"
	dismissed := maps.Clone(resolved)
	dismissed["resolution_status"], dismissed["resolved_indicator_record_id"], dismissed["resolution_method"] = "dismissed", nil, "indicators.observations.dismiss"
	for _, test := range []struct{ name, operation, defect string }{
		{"create", "create", ""}, {"resolve", "resolve", ""}, {"dismiss", "dismiss", ""}, {"restore", "restore", ""}, {"rollback_create", "rollback", ""}, {"rollback_resolution", "rollback", "resolution"},
		{"missing_tombstone", "create", "partial"}, {"wrong_physical_target", "create", "target"}, {"legacy_target", "create", "legacy"}, {"unknown_origin", "create", "origin"}, {"resolution_tuple", "resolve", "tuple"}, {"immutable_provenance", "resolve", "immutable"}, {"version_jump", "resolve", "version"}, {"partial_before", "resolve", "partial_before"}, {"resolution_before_creation", "resolve", "clock"},
	} {
		t.Run(test.name, func(t *testing.T) {
			base := h.validBundle(t)
			var before map[string]any
			after := maps.Clone(active)
			switch test.operation {
			case "resolve":
				before, after = maps.Clone(active), maps.Clone(resolved)
			case "dismiss":
				before, after = maps.Clone(active), maps.Clone(dismissed)
			case "restore":
				before = maps.Clone(dismissed)
				after["row_version"] = 3
			case "rollback":
				if test.defect == "resolution" {
					before = maps.Clone(resolved)
					after["row_version"] = 3
				} else {
					before = maps.Clone(active)
					after["row_version"], after["deleted_at"], after["deleted_by_user_id"] = 2, h.createdAt.Add(time.Minute).Format(time.RFC3339Nano), h.actor.ID.String()
				}
			}
			rows := decodePortableRows(t, base[revisionsTestMutationsPath])
			rows = append(rows, map[string]any{"change_set_id": h.changeSet.String(), "sequence_no": 2, "target_kind": "indicator_observation", "target_id": id.String(), "operation_kind": test.operation, "before_version_id": nil, "after_version_id": nil, "before_value": before, "after_value": after})
			base[revisionsTestMutationsPath] = encodePortableRows(t, rows)
			appendPortableCreationFixture(t, base, h, indicatorID, "indicator", "cartulary.revisions.snapshot.indicator.v1", portableIndicatorFixture(h, indicatorID, "observed.example", "domain_name:observed.example"))
			candidate := portabilitysupport.MapBundle{}
			for path, payload := range base {
				candidate[path] = append([]byte(nil), payload...)
			}
			rows = decodePortableRows(t, base[revisionsTestMutationsPath])
			mutation := rows[1]
			switch test.defect {
			case "partial":
				delete(mutation["after_value"].(map[string]any), "deleted_at")
			case "target":
				mutation["target_id"] = uuid.NewString()
			case "legacy":
				mutation["target_id"] = "indicator_observation:" + id.String()
			case "origin":
				mutation["after_value"].(map[string]any)["origin_kind"] = "interactive_cell"
			case "tuple":
				mutation["after_value"].(map[string]any)["resolution_method"] = nil
			case "immutable":
				mutation["after_value"].(map[string]any)["observed_text"] = "changed"
			case "version":
				mutation["after_value"].(map[string]any)["row_version"] = 4
			case "partial_before":
				delete(mutation["before_value"].(map[string]any), "created_at")
			case "clock":
				mutation["after_value"].(map[string]any)["resolved_at"] = h.createdAt.Add(-time.Minute).Format(time.RFC3339Nano)
			}
			candidate[revisionsTestMutationsPath] = encodePortableRows(t, rows)
			input := h.freshImportContext()
			prepared, err := h.port.PrepareImport(t.Context(), candidate, input)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()
			if _, err := tx.Exec(t.Context(), `INSERT INTO records(record_id,incident_id,record_type,row_version,created_by_user_id,updated_by_user_id,created_at,updated_at) VALUES($1,$2,'indicator',1,$3,$3,$4,$4)`, indicatorID, h.incidentID, h.actor.ID, h.createdAt); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(t.Context(), `INSERT INTO indicators(record_id,incident_id,indicator_type,value_kind,display_value,normalized_value,dedupe_key) VALUES($1,$2,'domain_name','atomic','observed.example','observed.example','domain_name:observed.example')`, indicatorID, h.incidentID); err != nil {
				t.Fatal(err)
			}
			clockArg := func(raw any) any {
				if raw == nil {
					return nil
				}
				at, err := time.Parse(time.RFC3339Nano, raw.(string))
				if err != nil {
					t.Fatal(err)
				}
				return at
			}
			if _, err := tx.Exec(t.Context(), `INSERT INTO indicator_observations(indicator_observation_id,incident_id,source_record_id,source_field_key,origin_kind,origin_locator,observed_text,parsed_indicator_type,normalized_candidate,resolution_status,resolved_indicator_record_id,row_version,created_by_user_id,created_at,resolved_by_user_id,resolved_at,resolution_method,deleted_at,deleted_by_user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`, id, h.incidentID, h.recordID, after["source_field_key"], after["origin_kind"], after["origin_locator"], after["observed_text"], after["parsed_indicator_type"], after["normalized_candidate"], after["resolution_status"], after["resolved_indicator_record_id"], after["row_version"], h.actor.ID, h.createdAt, after["resolved_by_user_id"], clockArg(after["resolved_at"]), after["resolution_method"], clockArg(after["deleted_at"]), after["deleted_by_user_id"]); err != nil {
				t.Fatal(err)
			}
			counts := func(query interface {
				QueryRow(context.Context, string, ...any) pgx.Row
			}, want int) {
				t.Helper()
				for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
					expected := want
					if table == "change_set_mutations" {
						expected *= 3
					} else if table == "record_revisions" {
						expected *= 2
					}
					var got int
					if err := query.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&got); err != nil || got != expected {
						t.Fatalf("%s count=%d want=%d error=%v", table, got, expected, err)
					}
				}
			}
			err = h.port.ApplyImportTx(t.Context(), tx, prepared, input)
			if test.defect == "" || test.defect == "resolution" {
				if err != nil {
					t.Fatalf("apply complete Observation history: %v", err)
				}
				counts(tx, 1)
				if err := h.port.ValidateImportTx(t.Context(), tx, prepared, input); err != nil {
					t.Fatalf("validate complete Observation history: %v", err)
				}
			} else {
				requireRevisionsInvariant(t, err, "revisions.history_reconstruction")
				counts(tx, 0)
				valid, err := h.port.PrepareImport(t.Context(), base, input)
				if err != nil {
					t.Fatal(err)
				}
				if err := h.port.ApplyImportTx(t.Context(), tx, valid, input); err != nil {
					t.Fatalf("apply valid control: %v", err)
				}
				encode := func(v any) any {
					if v == nil {
						return nil
					}
					b, err := json.Marshal(v)
					if err != nil {
						t.Fatal(err)
					}
					return b
				}
				if _, err := tx.Exec(t.Context(), `UPDATE change_set_mutations SET target_id=$2,before_value=$3,after_value=$4 WHERE change_set_id=$1 AND sequence_no=2`, h.changeSet, mutation["target_id"], encode(mutation["before_value"]), encode(mutation["after_value"])); err != nil {
					t.Fatal(err)
				}
				counts(tx, 1)
				requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.history_reconstruction")
			}
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
			counts(h.db, 0)
			var remaining int
			if err := h.db.QueryRow(t.Context(), `SELECT count(*) FROM indicator_observations WHERE indicator_observation_id=$1`, id).Scan(&remaining); err != nil || remaining != 0 {
				t.Fatalf("Observation rollback=%d error=%v", remaining, err)
			}
			for _, table := range []string{"indicators", "records"} {
				if err := h.db.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE record_id=$1", indicatorID).Scan(&remaining); err != nil || remaining != 0 {
					t.Fatalf("auxiliary %s rollback=%d error=%v", table, remaining, err)
				}
			}
		})
	}
}
