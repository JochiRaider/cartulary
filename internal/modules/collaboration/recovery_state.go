package collaboration

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.collaboration",
		recoverystate.SecurityStateTables(
			"collaboration.invalidate_restore_generation.v1",
			"collaboration_event_intents",
			"collaboration_incident_stream_cursors",
			"collaboration_replay_events",
			"collaboration_resume_tokens",
		),
	)
}
