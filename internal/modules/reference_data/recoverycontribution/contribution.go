package recoverycontribution

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.reference_data",
		recoverystate.AuthoritativeTables(
			"reference_pack_repositories",
			"reference_pack_roots",
			"reference_pack_key_state",
			"reference_pack_candidates",
			"reference_pack_versions",
			"reference_pack_envelopes",
			"reference_pack_metadata_versions",
			"reference_pack_operations",
			"reference_pack_operation_dependency_keys",
			"reference_pack_operation_keys",
			"reference_pack_operation_members",
			"reference_pack_operation_repositories",
			"reference_pack_attempts",
			"reference_pack_attempt_members",
			"reference_pack_objects",
			"reference_pack_object_refs",
			"reference_pack_index_generations",
			"reference_pack_indexes",
			"reference_pack_lookup_keys",
			"reference_pack_sets",
			"reference_pack_set_members",
			"reference_pack_current_set",
			"reference_pack_pins",
			"reference_pack_version_pins",
			"reference_pack_portable_catalogs",
			"reference_pack_portable_preparations",
			"reference_pack_portable_selections",
			"reference_pack_registry_usage",
			"reference_pack_events",
			"reference_pack_release_bindings",
		),
		recoverystate.AuthoritativeObjectFamily(
			"reference_packs.members",
			"reference_data.snapshot_member_inventory.v2",
			"reference_data.validate_member_inventory.v2",
			"reference_data.restore_member_inventory.v2",
		),
	)
}
