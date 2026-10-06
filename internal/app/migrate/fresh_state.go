package migrate

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

// Resolve bindings before connecting. The migration owner invokes the read-only
// inspection under its initialization exclusion, both before and after migration.
func freshStorageInspector(cfg configassembly.Deployment) (func(context.Context) error, error) {
	objects, err := objectstore.ResolveSettings(configassembly.ObjectStoreBinding(cfg), nil)
	if err != nil {
		return nil, err
	}
	roots := []string{cfg.Roots.BackupStorage.Path, cfg.Roots.ReferencePackStorage.Path, cfg.Roots.TemporaryWork.Path, cfg.Roots.ExportOutputs.Path}
	return func(ctx context.Context) error {
		for _, root := range roots {
			if err := ctx.Err(); err != nil {
				return err
			}
			empty, err := rootedfs.InspectEmpty(root)
			if err != nil || !empty {
				return database_migrations.ErrIncompatibleCryptoState
			}
		}
		empty, err := objectstore.InspectEmpty(ctx, objects)
		if err != nil || !empty {
			return database_migrations.ErrIncompatibleCryptoState
		}
		return nil
	}, nil
}
