package projections

type ImportPolicy struct {
	ApprovedRootImporters    []string
	ApprovedAdapterPackages  []string
	ApprovedContractPackages []string
}

func ProductionImportPolicy() ImportPolicy {
	return ImportPolicy{
		ApprovedRootImporters:    approvedProductionProjectionRootImporterPaths(),
		ApprovedAdapterPackages:  approvedProductionProjectionAdapterPackages(),
		ApprovedContractPackages: approvedProductionProjectionContractPackages(),
	}
}

func approvedProductionProjectionRootImporterPaths() []string {
	return []string{
		"internal/app/projectionassembly/catalog.go",
		"internal/app/recoveryassembly/state_catalog.go",
		"internal/app/timelineassembly/assembly.go",
		"internal/app/workbookassembly/catalog.go",
		"internal/modules/artifacts/import_projection.go",
		"internal/modules/artifacts/workbook_facade.go",
		"internal/modules/assessments/store.go",
		"internal/modules/entities/hostidentity/ports.go",
		"internal/modules/entities/mentions/ports.go",
		"internal/modules/entities/merge/ports.go",
		"internal/modules/evidence/import_projection.go",
		"internal/modules/evidence/store.go",
		"internal/modules/evidence/workbook_facade.go",
		"internal/modules/parties/store.go",
		"internal/modules/parties/workbook_facade.go",
		"internal/modules/tasksdecisions/import_projection.go",
		"internal/modules/tasksdecisions/supersede_facade.go",
		"internal/modules/tasksdecisions/workbook_facade.go",
		"internal/modules/workbook/contribution_catalog.go",
		"internal/modules/workbook/routes.go",
		"internal/modules/workbook/store.go",
	}
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
	return []string{}
}

func approvedProductionProjectionContractPackages() []string {
	return []string{
		"internal/modules/projections/providercontract",
	}
}
