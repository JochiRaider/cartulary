package postgres

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.database_migrations",
		recoverystate.SchemaMetadataTables("schema_migration_lineage"),
	)
}
