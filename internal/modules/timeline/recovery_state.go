package timeline

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution("module.timeline", recoverystate.AuthoritativeTables(
		"timeline_events",
		"timeline_source_provenance",
		"timeline_time_conversion_profiles",
	))
}
