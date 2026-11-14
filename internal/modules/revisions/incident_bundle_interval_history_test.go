package revisions_test

import (
	"context"
	"encoding/json"
	"maps"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func exerciseCompleteIntervalHistory(t *testing.T) {
	h := newRevisionsPortabilityHarness(t, "interval-history")
	id := uuid.New()
	indicatorID := uuid.New()
	active := map[string]any{"indicator_state_interval_id": id.String(), "incident_id": h.incidentID.String(), "indicator_record_id": indicatorID.String(),
		"lifecycle_state": "active", "valid_from": h.createdAt.Format(time.RFC3339Nano), "valid_to": nil, "confidence": nil, "rationale": nil, "support_refs": []string{h.recordID.String()}, "assessor": nil,
		"assessed_at": h.createdAt.Format(time.RFC3339Nano), "row_version": 1, "created_by_user_id": h.actor.ID.String(), "created_at": h.createdAt.Format(time.RFC3339Nano), "deleted_at": nil, "deleted_by_user_id": nil}

	for _, test := range []struct{ name, operation, defect string }{
		{"create", "create", ""}, {"rollback", "rollback", ""},
		{"partial_create", "create", "partial"}, {"legacy_target", "create", "legacy"}, {"wrong_physical_id", "create", "target"},
		{"lifecycle", "create", "lifecycle"}, {"immutable_rollback", "rollback", "immutable"}, {"partial_rollback", "rollback", "partial_before"}, {"paired_deletion", "rollback", "actor"}, {"version_jump", "rollback", "version"}, {"duplicate_support", "create", "support"},
	} {
		t.Run(test.name, func(t *testing.T) {
			base := h.validBundle(t)
			rows := decodePortableRows(t, base[revisionsTestMutationsPath])
			var before map[string]any
			after := maps.Clone(active)
			var deletedAt any
			if test.operation == "rollback" {
				before = maps.Clone(active)
				deletedAt = h.createdAt.Add(time.Minute)
				after["deleted_at"] = h.createdAt.Add(time.Minute).Format(time.RFC3339Nano)
				after["deleted_by_user_id"] = h.actor.ID.String()
				after["row_version"] = 2
			}
			rows = append(rows, map[string]any{
				"change_set_id": h.changeSet.String(), "sequence_no": 2, "target_kind": "indicator_state_interval", "target_id": id.String(),
				"operation_kind": test.operation, "before_version_id": nil, "after_version_id": nil, "before_value": before, "after_value": after,
			})
			base[revisionsTestMutationsPath] = encodePortableRows(t, rows)
			appendPortableCreationFixture(t, base, h, indicatorID, "indicator", "cartulary.revisions.snapshot.indicator.v1", portableIndicatorFixture(h, indicatorID, "interval.example.test", indicatortest.CanonicalDedupeKey(t, "domain_name", "atomic", "interval.example.test")))
			candidate := portabilitysupport.MapBundle{}
			for path, payload := range base {
				candidate[path] = append([]byte(nil), payload...)
			}
			rows = decodePortableRows(t, base[revisionsTestMutationsPath])
			mutation := rows[1]
			switch test.defect {
			case "partial":
				delete(mutation["after_value"].(map[string]any), "created_at")
			case "partial_before":
				delete(mutation["before_value"].(map[string]any), "indicator_state_interval_id")
			case "legacy":
				mutation["target_id"] = "indicator_state_interval:" + id.String()
			case "target":
				mutation["target_id"] = uuid.NewString()
			case "lifecycle":
				mutation["after_value"].(map[string]any)["lifecycle_state"] = "inactive"
			case "immutable":
				mutation["after_value"].(map[string]any)["lifecycle_state"] = "benign"
			case "actor":
				mutation["after_value"].(map[string]any)["deleted_by_user_id"] = nil
			case "version":
				mutation["after_value"].(map[string]any)["row_version"] = 3
			case "support":
				mutation["after_value"].(map[string]any)["support_refs"] = []string{h.recordID.String(), h.recordID.String()}
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
			if _, err := tx.Exec(t.Context(), `INSERT INTO records(record_id,incident_id,record_type,row_version,created_by_user_id,updated_by_user_id,created_at,updated_at)
VALUES($1,$2,'indicator',1,$3,$3,$4,$4)`, indicatorID, h.incidentID, h.actor.ID, h.createdAt); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(t.Context(), `INSERT INTO indicators(record_id,incident_id,indicator_type,value_kind,display_value,normalized_value,dedupe_key)
 VALUES($1,$2,'domain_name','atomic','interval.example.test','interval.example.test',$3)`, indicatorID, h.incidentID, indicatortest.CanonicalDedupeKey(t, "domain_name", "atomic", "interval.example.test")); err != nil {
				t.Fatal(err)
			}
			var deletedBy any
			version := 1
			if deletedAt != nil {
				deletedBy = h.actor.ID
				version = 2
			}
			if _, err := tx.Exec(t.Context(), `INSERT INTO indicator_state_intervals
 (indicator_state_interval_id,incident_id,indicator_record_id,lifecycle_state,valid_from,valid_to,confidence,rationale,support_refs,assessor,assessed_at,row_version,created_by_user_id,created_at,deleted_at,deleted_by_user_id)
 VALUES($1,$2,$3,'active',$4,NULL,NULL,NULL,jsonb_build_array($5::text),NULL,$4,$6,$7,$4,$8,$9)`, id, h.incidentID, indicatorID, h.createdAt, h.recordID, version, h.actor.ID, deletedAt, deletedBy); err != nil {
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
			wantInvariant := "revisions.history_reconstruction"
			if test.defect == "support" {
				wantInvariant = "revisions.references_complete"
			}
			err = h.port.ApplyImportTx(t.Context(), tx, prepared, input)
			if test.defect == "" {
				if err != nil {
					t.Fatalf("apply complete physical history: %v", err)
				}
				counts(tx, 1)
				if err := h.port.ValidateImportTx(t.Context(), tx, prepared, input); err != nil {
					t.Fatalf("validate complete physical history: %v", err)
				}
			} else {
				requireRevisionsInvariant(t, err, wantInvariant)
				counts(tx, 0)
				valid, err := h.port.PrepareImport(t.Context(), base, input)
				if err != nil {
					t.Fatal(err)
				}
				if err := h.port.ApplyImportTx(t.Context(), tx, valid, input); err != nil {
					t.Fatalf("apply valid control: %v", err)
				}
				encode := func(value any) any {
					if value == nil {
						return nil
					}
					payload, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					return payload
				}
				if _, err := tx.Exec(t.Context(), `UPDATE change_set_mutations SET target_id=$2,before_value=$3,after_value=$4 WHERE change_set_id=$1 AND sequence_no=2`, h.changeSet, mutation["target_id"], encode(mutation["before_value"]), encode(mutation["after_value"])); err != nil {
					t.Fatal(err)
				}
				counts(tx, 1)
				requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), wantInvariant)
			}
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
			counts(h.db, 0)
			var remaining int
			if err := h.db.QueryRow(t.Context(), `SELECT count(*) FROM indicator_state_intervals WHERE indicator_state_interval_id=$1`, id).Scan(&remaining); err != nil || remaining != 0 {
				t.Fatalf("source rollback count=%d error=%v", remaining, err)
			}
			for _, table := range []string{"indicators", "records"} {
				if err := h.db.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE record_id=$1", indicatorID).Scan(&remaining); err != nil || remaining != 0 {
					t.Fatalf("auxiliary source rollback %s count=%d error=%v", table, remaining, err)
				}
			}

		})
	}
}
