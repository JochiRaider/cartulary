package collaboration_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
)

func TestEventIntentExportsNoMutableFields_Unit(t *testing.T) {
	typeOfIntent := reflect.TypeOf(privatestream.EventIntent{})
	for index := 0; index < typeOfIntent.NumField(); index++ {
		field := typeOfIntent.Field(index)
		if field.IsExported() {
			t.Fatalf("EventIntent field %q remains exported", field.Name)
		}
	}
}

func TestEventIntentValidatesEveryEventFamily_Unit(t *testing.T) {
	incidentID := uuid.New()
	now := time.Date(2026, 7, 28, 5, 0, 0, 0, time.UTC)

	t.Run("job progress", func(t *testing.T) {
		_, err := privatestream.NewJobProgressIntent(
			"job_progress:invalid",
			incidentID,
			map[string]any{"job_id": "job-invalid"},
			"job:job-invalid",
			now,
		)
		if err == nil {
			t.Fatal("invalid job progress payload was admitted")
		}
		_, err = privatestream.NewJobProgressIntent(
			"job_progress:valid",
			incidentID,
			map[string]any{
				"job_id": "job-valid",
				"scope": map[string]any{
					"kind":        "incident",
					"incident_id": incidentID.String(),
				},
				"status":     "queued",
				"progress":   map[string]any{"completed": 0, "total": nil},
				"updated_at": now,
				"future":     map[string]any{"accepted": true},
			},
			"job:job-valid",
			now,
		)
		if err != nil {
			t.Fatalf("valid additive job progress payload: %v", err)
		}
	})

	t.Run("extension resource change", func(t *testing.T) {
		_, err := privatestream.NewExtensionResourceChangedIntent(
			"extension_resource_changed:invalid",
			incidentID,
			map[string]any{
				"extension_profile_id": "network_flow_activity",
				"resource_kind":        "network_flow_table",
				"resource_id":          "nft_invalid",
				"change_kind":          "remove",
				"reason_code":          "renamed",
			},
			"network_flow_table:nft_invalid",
			now,
		)
		if err == nil {
			t.Fatal("invalid extension resource payload was admitted")
		}
		_, err = privatestream.NewExtensionResourceChangedIntent(
			"extension_resource_changed:valid",
			incidentID,
			map[string]any{
				"extension_profile_id": "network_flow_activity",
				"resource_kind":        "network_flow_table",
				"resource_id":          "nft_valid",
				"change_kind":          "invalidate",
				"reason_code":          "renamed",
				"workspace_refs": []map[string]any{{
					"kind":                 "extension_workspace",
					"extension_profile_id": "network_flow_activity",
					"workspace_key":        "network_analysis",
				}},
				"future": true,
			},
			"network_flow_table:nft_valid",
			now,
		)
		if err != nil {
			t.Fatalf("valid additive extension resource payload: %v", err)
		}
	})

	t.Run("record change", func(t *testing.T) {
		_, err := privatestream.NewRecordChangedIntent(
			"record_changed:invalid",
			incidentID,
			map[string]any{"not": "a record change"},
			uuid.New(),
			uuid.New(),
			1,
			"record:invalid",
			0,
			now,
		)
		if err == nil {
			t.Fatal("invalid record change payload was admitted")
		}
	})
}
