package collaboration_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/incidentwstest"
)

func TestSocketEventInventoryCoverage(t *testing.T) {
	t.Run("record change intent builds a sorted compact patch", testRecordChangeIntentBuildsSortedCompactPatch)

	inventory := incidentwstest.SocketEventInventory()
	incidentwstest.RequireHarnessInventory(t, inventory)

	required := incidentwstest.RequiredHarnessIDs(inventory)
	for _, harness := range []incidentwstest.HarnessID{
		incidentwstest.HarnessEnvelopeConsistency,
		incidentwstest.HarnessAuthorizationRederived,
		incidentwstest.HarnessFieldKeyConformance,
		incidentwstest.HarnessProjectionRebuild,
		incidentwstest.HarnessWebSocketLifecycle,
		incidentwstest.HarnessGridIdentity,
		incidentwstest.HarnessTopologyAuditSource,
	} {
		if !slices.Contains(required, harness) {
			t.Fatalf("socket event inventory must require %s, got %v", harness, required)
		}
	}
}

func TestSocketLifecycleEvidenceIndex(t *testing.T) {
	evidence := map[string][]string{
		"canonical_handshake": {
			"internal/modules/collaboration/socket_test.go::TestIncidentSocketHandshakeResume_Unit/first_application_message_rejects_every_closed_token_except_hello_or_resume",
			"internal/modules/collaboration/integration_test.go::TestTwoClientsPresenceReplay_Integration",
		},
		"first_message_closed_vocabulary": {
			"internal/modules/collaboration/socket_test.go::TestIncidentSocketHandshakeResume_Unit/first_application_message_rejects_every_closed_token_except_hello_or_resume",
		},
		"heartbeat_idle_expiry": {
			"internal/modules/collaboration/socket_test.go::TestIncidentSocketHeartbeatIdleExpiry_Unit",
		},
		"origin_rejection": {
			"internal/modules/collaboration/integration_test.go::TestCookieSocketRejectsUntrustedOrigin_Integration",
		},
		"presence_scoping_and_ephemeral_state": {
			"internal/modules/collaboration/socket_test.go::TestIncidentSocketPresenceScopeEphemeral_Unit",
			"internal/modules/collaboration/hub_test.go::TestPresenceReplayRevocationTransport/presence_snapshots_are_incident_scoped_sorted_and_expire",
		},
		"presence_update_real_socket": {
			"internal/modules/collaboration/socket_test.go::TestIncidentSocketPresenceScopeEphemeral_Unit",
		},
		"record_changed_replay_filtering": {
			"internal/modules/collaboration/integration_test.go::TestResumeReplaysReplayableMessagesOnly_Integration",
			"internal/modules/collaboration/hub_test.go::TestPresenceReplayRevocationTransport/resume_replay_filters_to_replayable_incident_messages",
		},
		"job_progress_replay_filtering": {
			"internal/modules/collaboration/integration_test.go::TestResumeReplaysReplayableMessagesOnly_Integration",
		},
		"resume_reset_without_partial_replay": {
			"internal/modules/collaboration/socket_test.go::TestIncidentSocketHandshakeResume_Unit/invalid_stale_or_mismatched_resume_resets_without_partial_replay",
			"internal/modules/collaboration/hub_test.go::TestPresenceReplayRevocationTransport/resume_reset_covers_expired_mismatched_and_too_old_tokens",
		},
		"revocation_reason_and_close": {
			"internal/modules/collaboration/socket_test.go::TestIncidentSocketHeartbeatIdleExpiry_Unit",
			"internal/modules/collaboration/integration_test.go::TestIncidentSocketRevocationSources",
			"internal/modules/collaboration/hub_test.go::TestPresenceReplayRevocationTransport/revocation_subscribers_preserve_public_reason_codes",
		},
		"view_patch_canonicalization": {
			"internal/modules/collaboration/hub_test.go::TestPresenceReplayRevocationTransport/record_changes_emit_canonical_patch_cells_with_invalidate_fallback",
		},
	}

	requiredObligations := []string{
		"canonical_handshake",
		"first_message_closed_vocabulary",
		"heartbeat_idle_expiry",
		"origin_rejection",
		"presence_scoping_and_ephemeral_state",
		"presence_update_real_socket",
		"record_changed_replay_filtering",
		"job_progress_replay_filtering",
		"resume_reset_without_partial_replay",
		"revocation_reason_and_close",
		"view_patch_canonicalization",
	}

	for _, obligation := range requiredObligations {
		references := evidence[obligation]
		if len(references) == 0 {
			t.Fatalf("socket lifecycle obligation %s has no behavior evidence", obligation)
		}
		for _, reference := range references {
			if !strings.Contains(reference, "::Test") {
				t.Fatalf("socket lifecycle obligation %s references non-test evidence %q", obligation, reference)
			}
			if strings.Contains(reference, "collaboration_shared_harness_test.go") {
				t.Fatalf("socket lifecycle obligation %s must not use inventory-only evidence %q", obligation, reference)
			}
		}
	}
}
