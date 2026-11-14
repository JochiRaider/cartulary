package revisions_test

import (
	"context"
	"encoding/json"
	"maps"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func exerciseCompleteMentionHistory(t *testing.T) {
	h := newRevisionsPortabilityHarness(t, "mention-history")
	sourceID, mentionID := uuid.New(), uuid.New()
	if _, err := h.db.Exec(t.Context(), `INSERT INTO records
(record_id,incident_id,record_type,created_by_user_id,created_at,updated_by_user_id,updated_at,row_version)
VALUES($1,$2,'timeline_event',$3,$4,$3,$4,1)`, sourceID, h.incidentID, h.actor.ID, h.createdAt); err != nil {
		t.Fatal(err)
	}
	if _, err := h.db.Exec(t.Context(), `INSERT INTO timeline_events
(record_id,incident_id,activity_synopsis_text,capture_state,created_by_user_id,updated_by_user_id,recorded_at,edited_at)
VALUES($1,$2,'Mention source','rough',$3,$3,$4,$4)`, sourceID, h.incidentID, h.actor.ID, h.createdAt); err != nil {
		t.Fatal(err)
	}
	if _, err := h.db.Exec(t.Context(), `INSERT INTO entity_mentions
(entity_mention_id,source_record_id,entity_type,source_field_key,origin_kind,origin_locator,raw_text,normalized_text,resolution_status,row_version,ordinal,created_by_user_id,created_at,resolved_record_id,resolved_by_user_id,resolved_at,resolution_method)
VALUES($1,$2,'host','timeline.host_refs','manual_entry','portable-mention','Portable host','portable host','resolved',2,7,$3,$4,$5,$3,$4,'explicit_resolve_route')`, mentionID, sourceID, h.actor.ID, h.createdAt, h.recordID); err != nil {
		t.Fatal(err)
	}
	before := map[string]any{
		"entity_mention_id": mentionID.String(), "source_record_id": sourceID.String(),
		"entity_type": "host", "source_field_key": "timeline.host_refs", "origin_kind": "manual_entry",
		"origin_locator": "portable-mention", "raw_text": "Portable host", "normalized_text": "portable host",
		"resolution_status": "unresolved", "row_version": 1, "ordinal": 7,
		"created_by_user_id": h.actor.ID.String(), "created_at": h.createdAt.Format(time.RFC3339Nano),
		"resolved_record_id": nil, "resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil,
	}
	after := maps.Clone(before)
	after["resolution_status"], after["row_version"] = "resolved", 2
	after["resolved_record_id"], after["resolved_by_user_id"] = h.recordID.String(), h.actor.ID.String()
	after["resolved_at"], after["resolution_method"] = h.createdAt.Format(time.RFC3339Nano), "explicit_resolve_route"
	base := h.validBundle(t)
	mutations := decodePortableRows(t, base[revisionsTestMutationsPath])
	mutations = append(mutations, map[string]any{
		"change_set_id": h.changeSet.String(), "sequence_no": 2, "target_kind": "entity_mention", "target_id": mentionID.String(),
		"operation_kind": "patch", "before_version_id": nil, "after_version_id": nil, "before_value": before, "after_value": after,
	})
	base[revisionsTestMutationsPath] = encodePortableRows(t, mutations)
	source := canonicalRowSnapshot(sourceID, h.incidentID, h.actor.ID, "timeline_event", "cartulary.revisions.snapshot.timeline_event.v1", 1, map[string]any{"activity_synopsis_text": "Mention source", "capture_state": "rough"})["source"].(map[string]any)
	source["recorded_at"], source["edited_at"] = h.snapshot["record"].(map[string]any)["created_at"], h.snapshot["record"].(map[string]any)["updated_at"]
	appendPortableCreationFixture(t, base, h, sourceID, "timeline_event", "cartulary.revisions.snapshot.timeline_event.v1", source)
	for _, defect := range []string{"none", "before_missing_creation", "after_missing_ordinal", "derived_incident", "target_disagrees", "immutable_changed", "version_gap", "resolution_tuple", "version_only", "unsupported_operation"} {
		t.Run(defect, func(t *testing.T) {
			candidate := portabilitysupport.MapBundle{}
			for path, payload := range base {
				candidate[path] = append([]byte(nil), payload...)
			}
			rows := decodePortableRows(t, candidate[revisionsTestMutationsPath])
			mutation := rows[1]
			left, right := mutation["before_value"].(map[string]any), mutation["after_value"].(map[string]any)
			switch defect {
			case "before_missing_creation":
				delete(left, "created_at")
			case "after_missing_ordinal":
				delete(right, "ordinal")
			case "derived_incident":
				right["incident_id"] = h.incidentID.String()
			case "target_disagrees":
				mutation["target_id"] = uuid.NewString()
			case "immutable_changed":
				right["raw_text"] = "Changed input"
			case "version_gap":
				right["row_version"] = 3
			case "resolution_tuple":
				right["resolved_at"] = nil
			case "version_only":
				mutation["after_value"] = maps.Clone(left)
				mutation["after_value"].(map[string]any)["row_version"] = 2
			case "unsupported_operation":
				mutation["operation_kind"] = "create"
			}
			candidate[revisionsTestMutationsPath] = encodePortableRows(t, rows)
			input := h.freshImportContext()
			prepared, err := h.port.PrepareImport(t.Context(), candidate, input)
			if err != nil {
				t.Fatalf("prepare scalar-valid Mention history: %v", err)
			}
			tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()
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
			if defect == "none" {
				if err != nil {
					t.Fatalf("apply complete Mention history: %v", err)
				}
				counts(tx, 1)
				if err := h.port.ValidateImportTx(t.Context(), tx, prepared, input); err != nil {
					t.Fatalf("validate complete Mention history: %v", err)
				}
			} else {
				requireRevisionsInvariant(t, err, "revisions.history_reconstruction")
				counts(tx, 0)
				valid, err := h.port.PrepareImport(t.Context(), base, input)
				if err != nil {
					t.Fatal(err)
				}
				if err := h.port.ApplyImportTx(t.Context(), tx, valid, input); err != nil {
					t.Fatalf("apply valid baseline after rejected Mention: %v", err)
				}
				encode := func(value any) []byte {
					payload, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					return payload
				}
				// Exact storage agreement isolates the independent post-apply history gate.
				if _, err := tx.Exec(t.Context(), `UPDATE change_set_mutations SET target_id=$2,operation_kind=$3,before_value=$4,after_value=$5 WHERE change_set_id=$1 AND sequence_no=2`, h.changeSet, mutation["target_id"], mutation["operation_kind"], encode(mutation["before_value"]), encode(mutation["after_value"])); err != nil {
					t.Fatal(err)
				}
				counts(tx, 1)
				requireRevisionsInvariant(t, h.port.ValidateImportTx(t.Context(), tx, prepared, input), "revisions.history_reconstruction")
			}
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
			counts(h.db, 0)
		})
	}
}
