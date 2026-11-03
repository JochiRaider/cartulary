package extensionassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/extensioncodec"
)

// GeneratedRecoveryCatalog is the application composition edge from the
// admitted Extensions package and owner-provided physical bindings to
// Recovery's immutable catalog.
func GeneratedRecoveryCatalog() (*recovery.ExtensionBackupCatalog, error) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		return nil, err
	}
	return RecoveryCatalog(coordinator.BackupPlans())
}

func RecoveryCatalog(plans []extensions.BackupPlan) (*recovery.ExtensionBackupCatalog, error) {
	bindings := make([]recovery.ExtensionBackupBinding, 0)
	for _, plan := range plans {
		for _, binding := range plan.Bindings {
			codec, present := plan.Codecs[binding.BackupCodecID]
			if !present || codec.SHA256 != binding.BackupCodecSHA256 ||
				codec.BindingID != binding.BindingID || codec.StorageKind != binding.StorageKind {
				return nil, fmt.Errorf("recovery codec binding mismatch for %s", binding.BindingID)
			}
			if codec.Postgres == nil || codec.Postgres.Binding().BindingID != binding.BindingID {
				return nil, fmt.Errorf("unresolved logical recovery codec for %s", binding.BindingID)
			}
			mapping := codec.Postgres.Mapping()
			if mapping.SchemaName != "public" {
				return nil, fmt.Errorf("unsupported recovery schema for %s", binding.BindingID)
			}
			if binding.PostRestoreValidationAlgorithmID != extensioncodec.PostgresRestoreAlgorithmID {
				return nil, fmt.Errorf("unbound extension restore validator for %s", binding.BindingID)
			}
			bindings = append(bindings, recovery.ExtensionBackupBinding{
				ProfileID: plan.ProfileID, ImplementationBindingSHA256: plan.ImplementationBindingSHA256,
				PhysicalStateBindingSHA256: plan.PhysicalStateBindingSHA256,
				BindingID:                  binding.BindingID, LogicalFamilyID: binding.LogicalFamilyID,
				StorageKind: binding.StorageKind, RestoreOrderGroup: binding.RestoreOrderGroup,
				PostRestoreValidationAlgorithm: binding.PostRestoreValidationAlgorithmID,
				CurrentCodec:                   recoveryCodec(codec),
			})
		}
	}
	return recovery.NewExtensionBackupCatalog(bindings, nil)
}

func recoveryCodec(codec extensions.BackupCodec) recovery.ExtensionBackupCodec {
	return recovery.ExtensionBackupCodec{
		CodecID: codec.CodecID, CodecSHA256: codec.SHA256, MaxItems: codec.MaxItems,
		MaxEntryBytes: codec.MaxEntryBytes, MaxTotalBytes: codec.MaxBindingBytes,
		Postgres: codec.Postgres,
	}
}
