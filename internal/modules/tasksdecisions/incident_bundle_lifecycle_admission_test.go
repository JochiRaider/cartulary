package tasksdecisions_test

import (
	"bytes"
	"fmt"
	"maps"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestIncidentBundleTaskLifecycleAdmission_Unit(t *testing.T) {
	harness := tasksDecisionsPortabilityHarness{incidentID: uuid.New(), actor: authn.UserRecord{ID: uuid.New()}}
	port := newTaskDecisionIncidentBundleSourcePort(t)
	input := taskDecisionImportContext(harness, "lifecycle-admission")
	now := taskDecisionPortableTime()
	check := func(t *testing.T, row map[string]any, accepted bool) {
		t.Helper()
		bundle := taskDecisionBundle([]map[string]any{row}, nil)
		before := maps.Clone(bundle)
		for path, data := range before {
			before[path] = bytes.Clone(data)
		}
		_, err := port.PrepareImport(t.Context(), bundle, input)
		if accepted {
			if err != nil {
				t.Fatalf("legal Task source tuple rejected: %v", err)
			}
		} else {
			requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.dependent_fields_legal")
		}
		if !reflect.DeepEqual(bundle, before) {
			t.Fatal("preparation changed member bytes")
		}
	}
	// Each row specifies the source contract, independently of the policy helper.
	for _, state := range []struct {
		status                        string
		nullOwner, reason, completion bool
	}{
		{"open", false, false, false},
		{"in_progress", false, false, false},
		{"blocked", false, true, false},
		{"done", true, false, true},
		{"canceled", true, false, false},
	} {
		for _, owned := range []bool{false, true} {
			for reasonIndex, reason := range []any{nil, "", "Awaiting evidence"} {
				for clockIndex, clock := range []any{nil, now.Add(-time.Microsecond).Format(time.RFC3339Nano), now.Format(time.RFC3339Nano), now.Add(time.Microsecond).Format(time.RFC3339Nano)} {
					t.Run(fmt.Sprintf("%s owned=%t reason=%d completion=%d", state.status, owned, reasonIndex, clockIndex), func(t *testing.T) {
						row := validPortableTask(harness, uuid.New(), now)
						row["status"], row["blocked_reason"], row["completed_at"] = state.status, reason, clock
						if !owned {
							row["owner_user_id"] = nil
						}
						reasonOK := (state.reason && reasonIndex == 2) || (!state.reason && reasonIndex == 0)
						clockOK := (state.completion && clockIndex >= 2) || (!state.completion && clockIndex == 0)
						check(t, row, (owned || state.nullOwner) && reasonOK && clockOK)
					})
				}
			}
		}
	}
	t.Run("tuple_precedes_malformed_reference", func(t *testing.T) {
		row := validPortableTask(harness, uuid.New(), now)
		row["status"], row["requester_party_id"] = "done", "invalid-uuid"
		check(t, row, false)
	})
}
