package projections

func approvedProductionProjectionImporterPaths() []string {
	return []string{
		"internal/app/operator.go",
		"internal/modules/artifacts/import_projection.go",
		"internal/modules/artifacts/linkednotes/facade.go",
		"internal/modules/assessments/store.go",
		"internal/modules/entities/store.go",
		"internal/modules/evidence/import_projection.go",
		"internal/modules/evidence/store.go",
		"internal/modules/incidentbundles/source.go",
		"internal/modules/parties/store.go",
		"internal/modules/revisions/delete_restore_store.go",
		"internal/modules/revisions/rollback_store.go",
		"internal/modules/tasksdecisions/import_projection.go",
		"internal/modules/tasksdecisions/supersede_facade.go",
		"internal/modules/timeline/ports.go",
		"internal/modules/workbook/mutation_store.go",
		"internal/modules/workbook/store.go",
	}
}

func approvedProductionProjectionImporterSet() map[string]struct{} {
	paths := approvedProductionProjectionImporterPaths()
	approved := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		approved[path] = struct{}{}
	}
	return approved
}
