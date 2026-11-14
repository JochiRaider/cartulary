package tasksdecisions_test

import (
	"bytes"
	"fmt"
	"maps"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func TestIncidentBundleTasksDecisionsPreparationPrecedence_Unit(t *testing.T) {
	harness := tasksDecisionsPortabilityHarness{incidentID: uuid.New(), actor: authn.UserRecord{ID: uuid.New()}}
	port := newTaskDecisionIncidentBundleSourcePort(t)
	input := taskDecisionImportContext(harness, "prepare-precedence")
	ids := []uuid.UUID{uuid.MustParse("00000000-0000-4000-8000-000000000001"), uuid.MustParse("00000000-0000-4000-8000-000000000002"), uuid.MustParse("00000000-0000-4000-8000-000000000003")}
	row := func(kind string, index int) map[string]any {
		if kind == "decision" {
			return validPortableDecision(harness, ids[index], taskDecisionPortableTime())
		}
		return validPortableTask(harness, ids[index], taskDecisionPortableTime())
	}
	check := func(t testing.TB, bundle portabilitysupport.MapBundle, invariant string) {
		t.Helper()
		before := maps.Clone(bundle)
		for path, value := range before {
			before[path] = bytes.Clone(value)
		}
		_, err := port.PrepareImport(t.Context(), bundle, input)
		if invariant == "" {
			if err != nil {
				t.Fatal(err)
			}
		} else {
			requireTasksDecisionsPortabilityFailure(t, err, invariant)
		}
		if !reflect.DeepEqual(bundle, before) {
			t.Fatal("preparation changed member bytes")
		}
	}
	defects := []struct {
		name, invariant string
		taskOnly        bool
		apply           func(map[string]any)
	}{
		{"shape", "tasks_decisions.envelope_type_scope", false, func(r map[string]any) { r["unknown"] = true }},
		{"lifecycle", "tasks_decisions.lifecycle_legal", false, func(r map[string]any) { r["status"] = "invalid-state" }},
		{"dependent", "tasks_decisions.dependent_fields_legal", false, func(r map[string]any) { r["created_at"] = "invalid-clock" }},
		{"reference", "tasks_decisions.references_same_incident", true, func(r map[string]any) { r["requester_party_id"] = "invalid-uuid" }},
	}
	for _, kinds := range [][2]string{{"task_request", "task_request"}, {"decision", "decision"}, {"task_request", "decision"}, {"decision", "task_request"}} {
		for high := 0; high < len(defects)-1; high++ {
			for low := high + 1; low < len(defects); low++ {
				for _, highFirst := range []bool{false, true} {
					first, second := low, high
					if highFirst {
						first, second = high, low
					}
					if (defects[first].taskOnly && kinds[0] != "task_request") || (defects[second].taskOnly && kinds[1] != "task_request") {
						continue
					}
					t.Run(fmt.Sprintf("%s %s %s %s first=%t", kinds[0], kinds[1], defects[high].name, defects[low].name, highFirst), func(t *testing.T) {
						rows := []map[string]any{row(kinds[0], 0), row(kinds[1], 1)}
						defects[first].apply(rows[0])
						defects[second].apply(rows[1])
						var tasks, decisions []map[string]any
						for index, kind := range kinds {
							if kind == "decision" {
								decisions = append(decisions, rows[index])
							} else {
								tasks = append(tasks, rows[index])
							}
						}
						check(t, taskDecisionBundle(tasks, decisions), defects[high].invariant)
					})
				}
			}
		}
	}
	for _, kind := range []string{"task_request", "decision"} {
		t.Run(kind+" lifecycle before stored text in one row", func(t *testing.T) {
			r := row(kind, 0)
			r["status"] = "invalid-state"
			field := "title"
			if kind == "decision" {
				field = "summary"
			}
			r[field] = " noncanonical text"
			bundle := taskDecisionBundle([]map[string]any{r}, nil)
			if kind == "decision" {
				bundle = taskDecisionBundle(nil, []map[string]any{r})
			}
			check(t, bundle, "tasks_decisions.lifecycle_legal")
		})
	}
	t.Run("later dependent scalar precedes requester UUID in one row", func(t *testing.T) {
		r := row("task_request", 0)
		r["requester_party_id"], r["completed_at"] = "bad-uuid", "bad-clock"
		check(t, taskDecisionBundle([]map[string]any{r}, nil), "tasks_decisions.dependent_fields_legal")
	})
	for _, variant := range []string{"duplicate after rejected row", "decreasing after rejected row", "late malformed task", "late malformed decision"} {
		t.Run(variant, func(t *testing.T) {
			first, second := row("task_request", 0), row("task_request", 1)
			first["created_at"] = "bad-clock"
			bundle := taskDecisionBundle([]map[string]any{first, second}, nil)
			switch variant {
			case "duplicate after rejected row":
				second["record_id"] = first["record_id"]
				bundle = taskDecisionBundle([]map[string]any{first, second}, nil)
			case "decreasing after rejected row":
				first["record_id"], second["record_id"] = second["record_id"], first["record_id"]
				bundle = taskDecisionBundle([]map[string]any{first, second}, nil)
			case "late malformed task":
				bundle[tasksBundlePath] = append(bundle[tasksBundlePath], []byte("{broken\n")...)
			case "late malformed decision":
				bundle[decisionsBundlePath] = []byte("{broken\n")
			}
			check(t, bundle, "tasks_decisions.envelope_type_scope")
		})
	}
	t.Run("valid canonical family", func(t *testing.T) {
		check(t, taskDecisionBundle([]map[string]any{row("task_request", 0), row("task_request", 1)}, []map[string]any{row("decision", 2)}), "")
	})
}
