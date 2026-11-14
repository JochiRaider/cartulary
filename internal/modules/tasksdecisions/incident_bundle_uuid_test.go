package tasksdecisions_test

import (
	"bytes"
	"fmt"
	"maps"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestIncidentBundleTasksDecisionsUUIDAdmission_Unit(t *testing.T) {
	canonical := uuid.MustParse("12345678-abcd-4abc-8abc-123456789abc")
	harness := tasksDecisionsPortabilityHarness{incidentID: canonical, actor: authn.UserRecord{ID: canonical}}
	port := newTaskDecisionIncidentBundleSourcePort(t)
	for _, field := range []struct {
		kind, name, invariant string
		nullable              bool
	}{
		{"task_request", "record_id", "tasks_decisions.envelope_type_scope", false},
		{"task_request", "incident_id", "tasks_decisions.envelope_type_scope", false},
		{"task_request", "owner_user_id", "tasks_decisions.dependent_fields_legal", true},
		{"task_request", "requester_party_id", "tasks_decisions.references_same_incident", true},
		{"task_request", "decision_record_id", "tasks_decisions.references_same_incident", true},
		{"decision", "record_id", "tasks_decisions.envelope_type_scope", false},
		{"decision", "incident_id", "tasks_decisions.envelope_type_scope", false},
		{"decision", "owner_user_id", "tasks_decisions.dependent_fields_legal", false},
	} {
		t.Run(field.kind+" "+field.name, func(t *testing.T) {
			for index, value := range []any{
				canonical.String(), nil, uuid.Nil.String(), "", "bad-uuid", true, 1,
				strings.ToUpper(canonical.String()), " " + canonical.String(), canonical.String() + " ",
				strings.ReplaceAll(canonical.String(), "-", ""), "{" + canonical.String() + "}", "urn:uuid:" + canonical.String(),
			} {
				t.Run(fmt.Sprintf("representation %d", index), func(t *testing.T) {
					input := taskDecisionImportContext(harness, "uuid-admission")
					// Membership must not make a zero actor a valid owner scalar.
					var err error
					input.Actors, err = sourceport.NewActorCatalog([]sourceport.ActorDescriptor{
						{SourceActorID: canonical.String(), DisplayName: "Owner"},
						{SourceActorID: uuid.Nil.String(), DisplayName: "Invalid zero actor"},
					})
					if err != nil {
						t.Fatal(err)
					}
					row := validPortableTask(harness, canonical, taskDecisionPortableTime())
					row["status"] = "canceled"
					if field.kind == "decision" {
						row = validPortableDecision(harness, canonical, taskDecisionPortableTime())
					}
					row[field.name] = value
					if field.name == "incident_id" && value == uuid.Nil.String() {
						input.IncidentID = uuid.Nil
					}
					bundle := taskDecisionBundle([]map[string]any{row}, nil)
					if field.kind == "decision" {
						bundle = taskDecisionBundle(nil, []map[string]any{row})
					}
					before := maps.Clone(bundle)
					for path, payload := range before {
						before[path] = bytes.Clone(payload)
					}
					_, err = port.PrepareImport(t.Context(), bundle, input)
					if index == 0 || (value == nil && field.nullable) {
						if err != nil {
							t.Fatalf("valid UUID representation rejected: %v", err)
						}
					} else {
						requireTasksDecisionsPortabilityFailure(t, err, field.invariant)
					}
					if !reflect.DeepEqual(bundle, before) {
						t.Fatal("UUID input was normalized or mutated")
					}
				})
			}
			t.Run("required member missing", func(t *testing.T) {
				row := validPortableTask(harness, canonical, taskDecisionPortableTime())
				if field.kind == "decision" {
					row = validPortableDecision(harness, canonical, taskDecisionPortableTime())
				}
				delete(row, field.name)
				bundle := taskDecisionBundle([]map[string]any{row}, nil)
				if field.kind == "decision" {
					bundle = taskDecisionBundle(nil, []map[string]any{row})
				}
				_, err := port.PrepareImport(t.Context(), bundle, taskDecisionImportContext(harness, "missing-uuid"))
				requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.envelope_type_scope")
			})
		})
	}
}
