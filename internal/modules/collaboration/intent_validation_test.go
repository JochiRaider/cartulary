package collaboration_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
)

func TestEventIntentValidatesEveryEventFamily_Unit(t *testing.T) {
	incidentID := uuid.New()
	now := time.Date(2026, 7, 28, 5, 0, 0, 0, time.UTC)

	t.Run("job progress", func(t *testing.T) {
		_, err := collaboration.NewEventIntent(
			"job_progress:invalid",
			incidentID,
			collaboration.EventFamilyJobProgress,
			map[string]any{"job_id": "job-invalid"},
			"job:job-invalid",
			0,
			now,
		)
		if err == nil {
			t.Fatal("invalid job progress payload was admitted")
		}
		_, err = collaboration.NewEventIntent(
			"job_progress:valid",
			incidentID,
			collaboration.EventFamilyJobProgress,
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
			0,
			now,
		)
		if err != nil {
			t.Fatalf("valid additive job progress payload: %v", err)
		}
	})

	t.Run("extension resource change", func(t *testing.T) {
		_, err := collaboration.NewEventIntent(
			"extension_resource_changed:invalid",
			incidentID,
			collaboration.EventFamilyExtensionResourceChange,
			map[string]any{
				"extension_profile_id": "network_flow_activity",
				"resource_kind":        "network_flow_table",
				"resource_id":          "nft_invalid",
				"change_kind":          "remove",
				"reason_code":          "renamed",
			},
			"network_flow_table:nft_invalid",
			0,
			now,
		)
		if err == nil {
			t.Fatal("invalid extension resource payload was admitted")
		}
		_, err = collaboration.NewEventIntent(
			"extension_resource_changed:valid",
			incidentID,
			collaboration.EventFamilyExtensionResourceChange,
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
			0,
			now,
		)
		if err != nil {
			t.Fatalf("valid additive extension resource payload: %v", err)
		}
	})

	t.Run("record change", func(t *testing.T) {
		_, err := collaboration.NewEventIntent(
			"record_changed:invalid",
			incidentID,
			collaboration.EventFamilyRecordChanged,
			map[string]any{"not": "a record change"},
			"record:invalid",
			0,
			now,
		)
		if err == nil {
			t.Fatal("invalid record change payload was admitted")
		}
	})
}
