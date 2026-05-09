package collaboration_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/phase6test"
)

func TestSupportPhase6SharedHarness_SocketEventInventoryCoverage(t *testing.T) {
	inventory := phase6test.Phase6SocketEventInventory()
	phase6test.RequireSharedHarnessInventory(t, inventory)

	required := phase6test.RequiredHarnessIDs(inventory)
	for _, harness := range []phase6test.SharedHarnessID{
		phase6test.HarnessEnvelopeConsistency,
		phase6test.HarnessAuthorizationRederived,
		phase6test.HarnessFieldKeyConformance,
		phase6test.HarnessProjectionRebuild,
		phase6test.HarnessWebSocketLifecycle,
		phase6test.HarnessGridIdentity,
		phase6test.HarnessTopologyAuditSource,
	} {
		if !slices.Contains(required, harness) {
			t.Fatalf("socket event inventory must require %s, got %v", harness, required)
		}
	}
}

func TestSupportPhase6SharedHarness_SocketLifecycleEvidenceIndex(t *testing.T) {
	evidence := map[string][]string{
		"canonical_handshake": {
			"internal/modules/collaboration/phase6_socket_test.go::TestPhase6_IncidentSocketHandshakeResume_U_6_07/first_application_message_rejects_every_closed_token_except_hello_or_resume",
			"internal/modules/collaboration/phase6_integration_test.go::TestPhase6_TwoClientsPresenceReplay_I_6_01",
		},
		"first_message_closed_vocabulary": {
			"internal/modules/collaboration/phase6_socket_test.go::TestPhase6_IncidentSocketHandshakeResume_U_6_07/first_application_message_rejects_every_closed_token_except_hello_or_resume",
		},
		"heartbeat_idle_expiry": {
			"internal/modules/collaboration/phase6_socket_test.go::TestPhase6_IncidentSocketHeartbeatIdleExpiry_U_6_08",
		},
		"origin_rejection": {
			"internal/modules/collaboration/phase6_integration_test.go::TestPhase6_CookieSocketRejectsUntrustedOrigin_I_6_04",
		},
		"presence_scoping_and_ephemeral_state": {
			"internal/modules/collaboration/phase6_socket_test.go::TestPhase6_IncidentSocketPresenceScopeEphemeral_U_6_08",
			"internal/platform/ws/phase6_ws_test.go::TestSupportPhase6_PresenceReplayRevocationTransport/presence_snapshots_are_incident_scoped_sorted_and_expire",
		},
		"presence_update_real_socket": {
			"internal/modules/collaboration/phase6_socket_test.go::TestPhase6_IncidentSocketPresenceScopeEphemeral_U_6_08",
		},
		"record_changed_replay_filtering": {
			"internal/modules/collaboration/phase6_integration_test.go::TestPhase6_ResumeReplaysReplayableMessagesOnly_I_6_02",
			"internal/platform/ws/phase6_ws_test.go::TestSupportPhase6_PresenceReplayRevocationTransport/resume_replay_filters_to_replayable_incident_messages",
		},
		"job_progress_replay_filtering": {
			"internal/modules/collaboration/phase6_integration_test.go::TestPhase6_ResumeReplaysReplayableMessagesOnly_I_6_02",
		},
		"resume_reset_without_partial_replay": {
			"internal/modules/collaboration/phase6_socket_test.go::TestPhase6_IncidentSocketHandshakeResume_U_6_07/invalid_stale_or_mismatched_resume_resets_without_partial_replay",
			"internal/platform/ws/phase6_ws_test.go::TestSupportPhase6_PresenceReplayRevocationTransport/resume_reset_covers_expired_mismatched_and_too_old_tokens",
		},
		"revocation_reason_and_close": {
			"internal/modules/collaboration/phase6_socket_test.go::TestPhase6_IncidentSocketHeartbeatIdleExpiry_U_6_08",
			"internal/modules/collaboration/phase6_integration_test.go::TestSupportPhase6_IncidentSocketRevocationSources",
			"internal/platform/ws/phase6_ws_test.go::TestSupportPhase6_PresenceReplayRevocationTransport/revocation_subscribers_preserve_public_reason_codes",
		},
		"view_patch_canonicalization": {
			"internal/platform/ws/phase6_ws_test.go::TestSupportPhase6_PresenceReplayRevocationTransport/record_changes_emit_canonical_patch_cells_with_invalidate_fallback",
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
			if strings.Contains(reference, "phase6_shared_harness_test.go") {
				t.Fatalf("socket lifecycle obligation %s must not use inventory-only evidence %q", obligation, reference)
			}
		}
	}
}
