package indicators

import (
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/sourcestate"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func RecoveryStateContribution() (recoverystate.Contribution, error) {
	catalog, err := sourcestate.Load()
	if err != nil {
		return recoverystate.Contribution{}, err
	}
	authoritative := catalog.AuthoritativeRelations()
	authoritativeNames := make([]string, 0, len(authoritative))
	for _, relation := range authoritative {
		authoritativeNames = append(authoritativeNames, relation.TableName)
	}
	tables := recoverystate.AuthoritativeTables(authoritativeNames...)
	for _, relation := range catalog.RebuildableRelations() {
		tables = append(tables, recoverystate.RebuildableTables(
			relation.RebuildInvariantID,
			relation.TableName,
		)...)
	}
	return recoverystate.NewContribution("module.indicators", tables), nil
}
