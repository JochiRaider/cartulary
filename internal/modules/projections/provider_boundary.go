package projections

func approvedProductionProjectionRootImporterPaths() []string {
	return []string{}
}

func approvedProductionProjectionRootImporterSet() map[string]struct{} {
	paths := approvedProductionProjectionRootImporterPaths()
	approved := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		approved[path] = struct{}{}
	}
	return approved
}

func approvedProductionProjectionAdapterPackages() []string {
	return []string{
		"internal/modules/projections/adapters",
	}
}

func approvedProductionProjectionContractPackages() []string {
	return []string{
		"internal/modules/projections/providercontract",
	}
}
