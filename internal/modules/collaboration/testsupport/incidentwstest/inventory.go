package incidentwstest

import (
	"sort"
	"testing"
)

type HarnessID string

const (
	HarnessEnvelopeConsistency    HarnessID = "envelope_consistency"
	HarnessAuthorizationRederived HarnessID = "authorization_rederivation"
	HarnessFieldKeyConformance    HarnessID = "field_key_conformance"
	HarnessProjectionRebuild      HarnessID = "projection_rebuild"
	HarnessWebSocketLifecycle     HarnessID = "websocket_lifecycle"
	HarnessGridIdentity           HarnessID = "grid_identity"
	HarnessTopologyAuditSource    HarnessID = "topology_audit_source"
)

type HarnessRequirement string

const (
	HarnessRequired      HarnessRequirement = "required"
	HarnessNotApplicable HarnessRequirement = "n/a"
)

var socketHarnessIDs = []HarnessID{
	HarnessEnvelopeConsistency,
	HarnessAuthorizationRederived,
	HarnessFieldKeyConformance,
	HarnessProjectionRebuild,
	HarnessWebSocketLifecycle,
	HarnessGridIdentity,
	HarnessTopologyAuditSource,
}

type HarnessInventoryEntry struct {
	Key          string
	Surface      string
	Requirements map[HarnessID]HarnessRequirement
}

func SocketEventInventory() []HarnessInventoryEntry {
	return []HarnessInventoryEntry{
		{
			Key:     "incident_socket_upgrade",
			Surface: "GET /ws/incidents/{incident_id}",
			Requirements: requiredHarnesses(
				HarnessEnvelopeConsistency,
				HarnessAuthorizationRederived,
				HarnessWebSocketLifecycle,
				HarnessTopologyAuditSource,
			),
		},
		{
			Key:     "incident_socket_hello_resume",
			Surface: "hello|resume",
			Requirements: requiredHarnesses(
				HarnessAuthorizationRederived,
				HarnessFieldKeyConformance,
				HarnessWebSocketLifecycle,
			),
		},
		{
			Key:     "incident_socket_presence",
			Surface: "presence_snapshot|presence_delta|presence_update",
			Requirements: requiredHarnesses(
				HarnessAuthorizationRederived,
				HarnessFieldKeyConformance,
				HarnessWebSocketLifecycle,
				HarnessGridIdentity,
			),
		},
		{
			Key:     "incident_socket_replayable_events",
			Surface: "record_changed|job_progress|session_revoked",
			Requirements: requiredHarnesses(
				HarnessAuthorizationRederived,
				HarnessFieldKeyConformance,
				HarnessWebSocketLifecycle,
				HarnessProjectionRebuild,
				HarnessTopologyAuditSource,
			),
		},
	}
}

func RequireHarnessInventory(t testing.TB, inventory []HarnessInventoryEntry) {
	t.Helper()
	if len(inventory) == 0 {
		t.Fatal("incident websocket harness inventory must not be empty")
	}
	seenKeys := make(map[string]struct{}, len(inventory))
	for _, entry := range inventory {
		if entry.Key == "" {
			t.Fatalf("incident websocket harness inventory entry missing key: %+v", entry)
		}
		if entry.Surface == "" {
			t.Fatalf("incident websocket harness inventory %s missing surface", entry.Key)
		}
		if _, ok := seenKeys[entry.Key]; ok {
			t.Fatalf("duplicate incident websocket harness inventory key %s", entry.Key)
		}
		seenKeys[entry.Key] = struct{}{}
		for _, harness := range socketHarnessIDs {
			requirement, ok := entry.Requirements[harness]
			if !ok {
				t.Fatalf("incident websocket harness inventory %s missing requirement for %s", entry.Key, harness)
			}
			if requirement != HarnessRequired && requirement != HarnessNotApplicable {
				t.Fatalf("incident websocket harness inventory %s has invalid requirement %q for %s", entry.Key, requirement, harness)
			}
		}
	}
}

func RequiredHarnessIDs(inventory []HarnessInventoryEntry) []HarnessID {
	required := map[HarnessID]struct{}{}
	for _, entry := range inventory {
		for harness, requirement := range entry.Requirements {
			if requirement == HarnessRequired {
				required[harness] = struct{}{}
			}
		}
	}
	ids := make([]HarnessID, 0, len(required))
	for harness := range required {
		ids = append(ids, harness)
	}
	sort.Slice(ids, func(left, right int) bool {
		return ids[left] < ids[right]
	})
	return ids
}

func requiredHarnesses(required ...HarnessID) map[HarnessID]HarnessRequirement {
	requirements := make(map[HarnessID]HarnessRequirement, len(socketHarnessIDs))
	for _, harness := range socketHarnessIDs {
		requirements[harness] = HarnessNotApplicable
	}
	for _, harness := range required {
		requirements[harness] = HarnessRequired
	}
	return requirements
}
