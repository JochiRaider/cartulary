package artifacts

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution("module.artifacts", recoverystate.AuthoritativeTables(
		"artifact_findings",
		"artifact_forensic_keywords",
		"artifact_investigative_queries",
		"artifacts",
		"handoff_risk_refs",
	))
}
