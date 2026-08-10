package revisions

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution("module.revisions", recoverystate.AuthoritativeTables(
		"change_set_mutations",
		"change_sets",
		"record_history_entry_refs",
		"record_revision_conflict_facts",
		"record_revisions",
	))
}
