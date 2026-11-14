package revisions_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"maps"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func exerciseCompleteAliasHistory(t *testing.T) {
	h := newRevisionsPortabilityHarness(t, "alias-history")
	id := uuid.New()
	active := map[string]any{
		"entity_alias_id": id.String(), "incident_id": h.incidentID.String(), "record_id": h.recordID.String(),
		"entity_type": "host", "raw_text": "Portable Alias", "normalized_text": "Portable Alias",
		"classification": "suggestion_only", "created_by_user_id": h.actor.ID.String(), "created_at": h.createdAt.Format(time.RFC3339Nano), "deleted_at": nil,
	}
	for _, test := range []struct{ name, operation, defect string }{
		{"create", "create", ""}, {"delete", "delete", ""}, {"rollback_delete", "rollback", ""}, {"rollback_restore", "rollback", ""},
		{"partial_create", "create", "partial"}, {"legacy_target", "create", "legacy"}, {"wrong_physical_id", "create", "target"},
		{"classification", "create", "classification"}, {"clock", "delete", "clock"}, {"immutable_rollback", "rollback", "immutable"}, {"partial_rollback", "rollback", "partial_before"},
	} {
		t.Run(test.name, func(t *testing.T) {
			base := h.validBundle(t)
			rows := decodePortableRows(t, base[revisionsTestMutationsPath])
			var before map[string]any
			after := maps.Clone(active)
			var deletedAt any
			if test.operation != "create" {
				before = maps.Clone(active)
				deletedAt = h.createdAt.Add(time.Minute)
				after["deleted_at"] = h.createdAt.Add(time.Minute).Format(time.RFC3339Nano)
			}
			if test.name == "rollback_restore" {
				before, after = after, before
				deletedAt = nil
			}
			rows = append(rows, map[string]any{
				"change_set_id": h.changeSet.String(), "sequence_no": 2, "target_kind": "entity_alias", "target_id": "entity_alias:" + id.String(),
				"operation_kind": test.operation, "before_version_id": nil, "after_version_id": nil, "before_value": before, "after_value": after,
			})
			base[revisionsTestMutationsPath] = encodePortableRows(t, rows)
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
				delete(mutation["before_value"].(map[string]any), "entity_alias_id")
			case "legacy":
				mutation["target_id"] = "entity_alias:" + base64.RawURLEncoding.EncodeToString([]byte(h.recordID.String())) + ":aG9zdA:UG9ydGFibGUgQWxpYXM"
			case "target":
				mutation["target_id"] = "entity_alias:" + uuid.NewString()
			case "classification":
				mutation["after_value"].(map[string]any)["classification"] = "canonical"
			case "immutable":
				mutation["after_value"].(map[string]any)["raw_text"] = "Changed Alias"
				mutation["after_value"].(map[string]any)["normalized_text"] = "Changed Alias"
			case "clock":
				mutation["after_value"].(map[string]any)["deleted_at"] = h.createdAt.Add(-time.Microsecond).Format(time.RFC3339Nano)
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
			if _, err := tx.Exec(t.Context(), `INSERT INTO entity_aliases
(entity_alias_id,incident_id,record_id,entity_type,raw_text,normalized_text,classification,created_by_user_id,created_at,deleted_at)
VALUES($1,$2,$3,'host','Portable Alias','Portable Alias','suggestion_only',$4,$5,$6)`, id, h.incidentID, h.recordID, h.actor.ID, h.createdAt, deletedAt); err != nil {
				t.Fatal(err)
			}
			counts := func(query interface {
				QueryRow(context.Context, string, ...any) pgx.Row
			}, want int) {
				t.Helper()
				for _, table := range []string{"change_sets", "change_set_mutations", "record_revisions"} {
					expected := want
					if table == "change_set_mutations" {
						expected *= 2
					}
					var got int
					if err := query.QueryRow(t.Context(), "SELECT count(*) FROM "+table+" WHERE change_set_id=$1", h.changeSet).Scan(&got); err != nil || got != expected {
						t.Fatalf("%s count=%d want=%d error=%v", table, got, expected, err)
					}
				}
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
				requireRevisionsInvariant(t, err, "revisions.history_reconstruction")
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
				requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.history_reconstruction")
			}
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
			counts(h.db, 0)
			var remaining int
			if err := h.db.QueryRow(t.Context(), `SELECT count(*) FROM entity_aliases WHERE entity_alias_id=$1`, id).Scan(&remaining); err != nil || remaining != 0 {
				t.Fatalf("source rollback count=%d error=%v", remaining, err)
			}
		})
	}
}
