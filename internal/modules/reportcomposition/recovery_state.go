package reportcomposition

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution("module.reportcomposition", recoverystate.AuthoritativeTables(
		"report_composition_preview_attempts",
		"report_composition_release_bindings",
		"report_composition_versions",
		"report_compositions",
	))
}
