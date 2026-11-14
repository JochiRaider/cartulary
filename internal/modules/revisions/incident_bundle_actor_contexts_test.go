package revisions_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func exerciseRevisionActorContextAdmission(t *testing.T, h revisionsPortabilityHarness, probe *revisionsPreparationEnvelopeProbe) {
	step := func() map[string]any {
		return map[string]any{"import_id": uuid.NewString(), "actor_user_id": h.actor.ID.String()}
	}
	for _, test := range []struct {
		name, invariant string
		change          func(map[string]any)
	}{
		{"empty", "", func(map[string]any) {}},
		{"one", "", func(row map[string]any) { row["actor_import_contexts"] = []any{step()} }},
		{"missing", "revisions.references_complete", func(row map[string]any) { delete(row, "actor_import_contexts") }},
		{"null", "revisions.references_complete", func(row map[string]any) { row["actor_import_contexts"] = nil }},
		{"object", "revisions.references_complete", func(row map[string]any) { row["actor_import_contexts"] = step() }},
		{"null_step", "revisions.references_complete", func(row map[string]any) { row["actor_import_contexts"] = []any{nil} }},
		{"unknown_member", "revisions.references_complete", func(row map[string]any) {
			value := step()
			value["unknown"] = true
			row["actor_import_contexts"] = []any{value}
		}},
		{"missing_actor", "revisions.references_complete", func(row map[string]any) {
			value := step()
			delete(value, "actor_user_id")
			row["actor_import_contexts"] = []any{value}
		}},
		{"zero_import", "revisions.references_complete", func(row map[string]any) {
			value := step()
			value["import_id"] = uuid.Nil.String()
			row["actor_import_contexts"] = []any{value}
		}},
		{"noncanonical_import", "revisions.references_complete", func(row map[string]any) {
			value := step()
			value["import_id"] = "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA"
			row["actor_import_contexts"] = []any{value}
		}},
		{"unadmitted_actor", "revisions.actor_references_complete", func(row map[string]any) {
			value := step()
			value["actor_user_id"] = uuid.NewString()
			row["actor_import_contexts"] = []any{value}
		}},
		{"zero_actor", "revisions.actor_references_complete", func(row map[string]any) {
			value := step()
			value["actor_user_id"] = uuid.Nil.String()
			row["actor_import_contexts"] = []any{value}
		}},
		{"duplicate_import", "revisions.history_reconstruction", func(row map[string]any) { value := step(); row["actor_import_contexts"] = []any{value, value} }},
		{"current_import", "revisions.history_reconstruction", func(row map[string]any) {
			value := step()
			value["import_id"] = h.context.OperationID
			row["actor_import_contexts"] = []any{value}
		}},
		{"reference_before_actor", "revisions.references_complete", func(row map[string]any) {
			value := step()
			value["actor_user_id"] = uuid.NewString()
			row["actor_import_contexts"] = []any{value, nil}
		}},
	} {
		t.Run("context_"+test.name, func(t *testing.T) {
			bundle := h.validBundle(t)
			rows := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])
			test.change(rows[0])
			bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, rows)
			before := clonePortableBundle(bundle)
			_, err := h.port.PrepareImport(t.Context(), bundle, h.context)
			if test.invariant == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				requireRevisionsInvariant(t, err, test.invariant)
			}
			if !reflect.DeepEqual(before, bundle) || probe.calls != 0 {
				t.Fatal("path preparation mutated bytes or accessed the database")
			}
		})
	}
	for _, operation := range []string{"", uuid.Nil.String(), "invalid-operation", "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA"} {
		input := h.context
		input.OperationID = operation
		_, err := h.port.PrepareImport(t.Context(), h.validBundle(t), input)
		if operation == "" {
			if !errors.Is(err, sourceport.ErrInvalidCatalog) {
				t.Fatalf("missing operation binding: %v", err)
			}
		} else {
			requireRevisionsInvariant(t, err, "revisions.references_complete")
		}
	}
}

func exerciseRevisionActorContextChains(t *testing.T) {
	for _, scenario := range []string{"boundary", "reversed_history_order", "gapped_history_order", "repeated_import", "same_context_exact", "same_context_actor_drift", "native_actor_drift", "nonactor_drift", "reverse_boundary", "divergent_path", "inconsistent_actor_binding", "wrong_boundary_actor"} {
		t.Run(scenario, func(t *testing.T) {
			h := newRevisionsPortabilityHarness(t, "actor-context")
			for _, table := range []string{"records", "hosts"} {
				if _, err := h.db.Exec(t.Context(), "UPDATE "+table+" SET row_version=2 WHERE record_id=$1", h.recordID); err != nil {
					t.Fatal(err)
				}
			}
			actorA, actorB := uuid.NewString(), uuid.NewString()
			actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: h.actor.ID.String()}, {SourceActorID: actorA}, {SourceActorID: actorB}})
			if err != nil {
				t.Fatal(err)
			}
			h.context.Actors = actors
			a := map[string]any{"import_id": uuid.NewString(), "actor_user_id": actorA}
			b := map[string]any{"import_id": uuid.NewString(), "actor_user_id": actorB}
			firstPath, nextPath := []any{a, b}, []any{b}
			before := clonePortableRow(t, h.snapshot)
			replacePortableActorIDs(before, h.actor.ID.String(), actorA)
			switch scenario {
			case "repeated_import":
				c := map[string]any{"import_id": uuid.NewString(), "actor_user_id": h.actor.ID.String()}
				firstPath, nextPath = append(firstPath, c), append(nextPath, c)
			case "same_context_exact":
				firstPath = nextPath
				before = clonePortableRow(t, h.snapshot)
			case "same_context_actor_drift":
				firstPath = nextPath
			case "native_actor_drift":
				firstPath, nextPath = []any{}, []any{}
			case "nonactor_drift":
				before["source"].(map[string]any)["display_name"] = "Contradiction"
			case "reverse_boundary":
				firstPath, nextPath = nextPath, firstPath
			case "divergent_path":
				nextPath = []any{map[string]any{"import_id": uuid.NewString(), "actor_user_id": actorB}}
			case "inconsistent_actor_binding":
				nextPath = []any{map[string]any{"import_id": b["import_id"], "actor_user_id": actorA}}
			case "wrong_boundary_actor":
				replacePortableActorIDs(before, actorA, actorB)
			}
			base := h.validBundle(t)
			first := decodePortableRows(t, base[revisionsTestChangeSetsPath])[0]
			next := clonePortableRow(t, first)
			next["history_order"] = 2
			if scenario == "reversed_history_order" {
				first["history_order"], next["history_order"] = 2, 1
			}
			if scenario == "gapped_history_order" {
				next["history_order"] = 17
			}
			first["change_set_id"], next["change_set_id"] = "00000000-0000-4000-8000-000000000001", "00000000-0000-4000-8000-000000000002"
			first["actor_import_contexts"], next["actor_import_contexts"] = firstPath, nextPath
			base[revisionsTestChangeSetsPath] = encodePortableRows(t, []map[string]any{first, next})
			after := clonePortableRow(t, h.snapshot)
			if scenario != "same_context_exact" {
				replacePortableActorIDs(after, h.actor.ID.String(), actorA)
			}
			for _, member := range []string{"record", "source"} {
				after[member].(map[string]any)["row_version"] = 2
			}
			history := decodePortableRows(t, base[revisionsTestHistoryPath])[0]
			history["change_set_id"] = first["change_set_id"]
			historyNext := clonePortableRow(t, history)
			historyNext["change_set_id"], historyNext["revision_id"], historyNext["row_version"], historyNext["before_json"], historyNext["after_json"] = next["change_set_id"], 2, 2, before, after
			base[revisionsTestHistoryPath] = encodePortableRows(t, []map[string]any{history, historyNext})
			mutation := decodePortableRows(t, base[revisionsTestMutationsPath])[0]
			mutation["change_set_id"] = first["change_set_id"]
			mutationNext := clonePortableRow(t, mutation)
			mutationNext["change_set_id"], mutationNext["operation_kind"], mutationNext["before_value"], mutationNext["after_value"] = next["change_set_id"], "patch", before, after
			mutationNext["before_version_id"], mutationNext["after_version_id"] = "host:"+h.recordID.String()+":1", "host:"+h.recordID.String()+":2"
			base[revisionsTestMutationsPath] = encodePortableRows(t, []map[string]any{mutation, mutationNext})
			err = h.applyAndValidate(t, base)
			if scenario == "boundary" || scenario == "gapped_history_order" || scenario == "repeated_import" || scenario == "same_context_exact" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				requireRevisionsInvariant(t, err, "revisions.history_reconstruction")
			}
		})
	}
}
