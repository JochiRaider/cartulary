// packagerecovery creates one historical encrypted input fixture for Make-owned
// qualification. It is never shipped and supplies no restore implementation.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/incidentportabilityassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "historical package fixture failed:", err)
		os.Exit(1)
	}
}
func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	loaded, err := configassembly.Load(configassembly.LoadOptions{Path: "/etc/cartulary/config.toml"})
	if err != nil {
		return err
	}
	cfg := loaded.Deployment()
	settings, err := postgres.ResolveSettings(configassembly.PostgresBinding(cfg), postgres.PurposeRecovery, nil)
	if err != nil {
		return err
	}
	pool, err := postgres.Setup(ctx, settings)
	if err != nil {
		return err
	}
	defer pool.Close()
	objectsSettings, err := objectstore.ResolveSettings(configassembly.ObjectStoreBinding(cfg), nil)
	if err != nil {
		return err
	}
	objects, err := objectstore.Setup(ctx, objectsSettings, configassembly.ObjectStoreInstrumentation(cfg))
	if err != nil {
		return err
	}
	defer objects.Close()
	packs, err := referenceassembly.NewRecoveryStorage(cfg.Roots.TemporaryWork.Path, cfg.Roots.ReferencePackStorage.Path, reference_data.DefaultLimits())
	if err != nil {
		return err
	}
	defer packs.Close()
	exports, err := incidentportabilityassembly.NewRecoveryStorage(cfg.Roots.ExportOutputs.Path)
	if err != nil {
		return err
	}
	defer exports.Close()
	storage, err := recoveryassembly.NewBackupStorage(cfg.Roots.BackupStorage.BindingKind, cfg.Roots.BackupStorage.Path, nil)
	if err != nil {
		return err
	}
	defer recovery.CloseBackupStorage(storage)
	state, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		return err
	}
	inventories, err := recoveryassembly.CurrentVNextObjectInventoryCatalog(recoveryassembly.NewVNextObjectSource(objects), packs, exports)
	if err != nil {
		return err
	}
	streaming, err := recovery.RequireStreamingBackupStorage(storage)
	if err != nil {
		return err
	}
	capture, err := recovery.NewVNextCaptureService(recoveryassembly.NewVNextSnapshotRepository(pool), streaming, state, inventories)
	if err != nil {
		return err
	}
	point := time.Now().UTC().Add(-40 * 24 * time.Hour)
	captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{BackupSetID: uuid.New(), ConsistencyPointAt: point, CreatedAt: point.Add(-time.Minute), RetainedUntil: point.Add(31 * 24 * time.Hour)})
	if err != nil {
		return err
	}
	backup, err := recovery.NewStore(pool).PublishVNextCapturedBackup(ctx, captured)
	if err != nil {
		return err
	}
	extension, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		return err
	}
	manifest, err := os.ReadFile("/etc/cartulary/release-manifest.json")
	if err != nil {
		return err
	}
	digest := sha256.Sum256(manifest)
	destination, err := recoveryassembly.NewTransferDirectory("/transfer-output/stale")
	if err != nil {
		return err
	}
	defer destination.Close()
	catalog := recovery.NewBackupCatalog(recovery.NewStore(pool), storage, extension, state)
	if _, err = catalog.ExportTransfer(ctx, backup, hex.EncodeToString(digest[:]), point.Add(time.Second), destination); err != nil {
		return err
	}
	if err = destination.Publish(ctx); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"backup_set_id": backup.BackupSetID, "consistency_point_at": point, "fixture": "completed_export_after_source_retention_expiry"})
}
