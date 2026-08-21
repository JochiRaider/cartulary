package workbookroutetest

import (
	"sort"
	"testing"
)

type SharedHarnessID string

const (
	HarnessEnvelopeConsistency     SharedHarnessID = "envelope_consistency"
	HarnessAuthorizationRederived  SharedHarnessID = "authorization_rederivation"
	HarnessDivergentReplay         SharedHarnessID = "divergent_replay"
	HarnessClosedVocabulary        SharedHarnessID = "closed_vocabulary"
	HarnessWritableStringNormalize SharedHarnessID = "writable_string_normalization"
	HarnessFieldKeyConformance     SharedHarnessID = "field_key_conformance"
	HarnessProjectionRebuild       SharedHarnessID = "projection_rebuild"
	HarnessWebSocketLifecycle      SharedHarnessID = "websocket_lifecycle"
	HarnessGridIdentity            SharedHarnessID = "grid_identity"
	HarnessBrowserCommands         SharedHarnessID = "browser_commands"
	HarnessTopologyAuditSource     SharedHarnessID = "topology_audit_source"
)

type HarnessRequirement string

const (
	HarnessRequired      HarnessRequirement = "required"
	HarnessNotApplicable HarnessRequirement = "n/a"
)

var WorkbookHarnessIDs = []SharedHarnessID{
	HarnessEnvelopeConsistency,
	HarnessAuthorizationRederived,
	HarnessDivergentReplay,
	HarnessClosedVocabulary,
	HarnessWritableStringNormalize,
	HarnessFieldKeyConformance,
	HarnessProjectionRebuild,
	HarnessWebSocketLifecycle,
	HarnessGridIdentity,
	HarnessBrowserCommands,
	HarnessTopologyAuditSource,
}

type HarnessInventoryEntry struct {
	Key          string
	Surface      string
	Requirements map[SharedHarnessID]HarnessRequirement
}

func WorkbookRouteInventory() []HarnessInventoryEntry {
	return []HarnessInventoryEntry{
		{
			Key:     "workbook_rows_create",
			Surface: "POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows",
			Requirements: requiredHarnesses(
				HarnessEnvelopeConsistency,
				HarnessAuthorizationRederived,
				HarnessDivergentReplay,
				HarnessWritableStringNormalize,
				HarnessFieldKeyConformance,
				HarnessProjectionRebuild,
				HarnessTopologyAuditSource,
			),
		},
		{
			Key:     "workbook_records_patch",
			Surface: "PATCH /api/v1/records/{record_id}",
			Requirements: requiredHarnesses(
				HarnessEnvelopeConsistency,
				HarnessAuthorizationRederived,
				HarnessDivergentReplay,
				HarnessWritableStringNormalize,
				HarnessFieldKeyConformance,
				HarnessProjectionRebuild,
				HarnessTopologyAuditSource,
			),
		},
		{
			Key:     "workbook_conflicts_resolve",
			Surface: "POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve",
			Requirements: requiredHarnesses(
				HarnessEnvelopeConsistency,
				HarnessAuthorizationRederived,
				HarnessDivergentReplay,
				HarnessClosedVocabulary,
				HarnessWritableStringNormalize,
				HarnessFieldKeyConformance,
				HarnessProjectionRebuild,
				HarnessWebSocketLifecycle,
				HarnessTopologyAuditSource,
			),
		},
	}
}

func RequireSharedHarnessInventory(t testing.TB, inventory []HarnessInventoryEntry) {
	t.Helper()
	if len(inventory) == 0 {
		t.Fatal("shared harness inventory must not be empty")
	}
	seenKeys := make(map[string]struct{}, len(inventory))
	for _, entry := range inventory {
		if entry.Key == "" {
			t.Fatalf("shared harness inventory entry missing key: %+v", entry)
		}
		if entry.Surface == "" {
			t.Fatalf("shared harness inventory %s missing surface", entry.Key)
		}
		if _, ok := seenKeys[entry.Key]; ok {
			t.Fatalf("duplicate shared harness inventory key %s", entry.Key)
		}
		seenKeys[entry.Key] = struct{}{}
		for _, harness := range WorkbookHarnessIDs {
			requirement, ok := entry.Requirements[harness]
			if !ok {
				t.Fatalf("shared harness inventory %s missing requirement for %s", entry.Key, harness)
			}
			if requirement != HarnessRequired && requirement != HarnessNotApplicable {
				t.Fatalf("shared harness inventory %s has invalid requirement %q for %s", entry.Key, requirement, harness)
			}
		}
	}
}

func RequiredHarnessIDs(inventory []HarnessInventoryEntry) []SharedHarnessID {
	required := map[SharedHarnessID]struct{}{}
	for _, entry := range inventory {
		for harness, requirement := range entry.Requirements {
			if requirement == HarnessRequired {
				required[harness] = struct{}{}
			}
		}
	}
	ids := make([]SharedHarnessID, 0, len(required))
	for harness := range required {
		ids = append(ids, harness)
	}
	sort.Slice(ids, func(left, right int) bool {
		return ids[left] < ids[right]
	})
	return ids
}

func requiredHarnesses(required ...SharedHarnessID) map[SharedHarnessID]HarnessRequirement {
	requirements := make(map[SharedHarnessID]HarnessRequirement, len(WorkbookHarnessIDs))
	for _, harness := range WorkbookHarnessIDs {
		requirements[harness] = HarnessNotApplicable
	}
	for _, harness := range required {
		requirements[harness] = HarnessRequired
	}
	return requirements
}
