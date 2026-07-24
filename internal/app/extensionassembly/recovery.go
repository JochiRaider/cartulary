package extensionassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
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
			tables, err := recoveryPostgresTables(plan.ProfileID, binding.BindingID)
			if err != nil {
				return nil, err
			}
			historical := make([]recovery.ExtensionBackupCodec, 0, len(codec.HistoricalCodecs))
			for _, identity := range codec.HistoricalCodecs {
				packaged, present := plan.Codecs[identity.CodecID]
				if !present || packaged.SHA256 != identity.SHA256 || packaged.BindingID != binding.BindingID {
					return nil, fmt.Errorf("historical recovery codec is not exactly packaged for %s", binding.BindingID)
				}
				historical = append(historical, recoveryCodec(packaged))
			}
			bindings = append(bindings, recovery.ExtensionBackupBinding{
				ProfileID: plan.ProfileID, ImplementationBindingSHA256: plan.ImplementationBindingSHA256,
				PhysicalStateBindingSHA256: plan.PhysicalStateBindingSHA256,
				BindingID:                  binding.BindingID, LogicalFamilyID: binding.LogicalFamilyID,
				StorageKind: binding.StorageKind, RestoreOrderGroup: binding.RestoreOrderGroup,
				PostRestoreValidationAlgorithm: binding.PostRestoreValidationAlgorithmID,
				CurrentCodec:                   recoveryCodec(codec),
				HistoricalCodecs:               historical,
				PostgresTables:                 tables,
			})
		}
	}
	return recovery.NewExtensionBackupCatalog(bindings, []recovery.ExtensionPristineMetadata{{
		ProfileID:          networkflow.ProfileID,
		MigrationLineageID: "network_flow_activity.state_v1",
		StateVersion:       1,
		MetadataVersion:    1,
	}})
}

func recoveryCodec(codec extensions.BackupCodec) recovery.ExtensionBackupCodec {
	return recovery.ExtensionBackupCodec{
		CodecID: codec.CodecID, CodecSHA256: codec.SHA256, MaxItems: codec.MaxItems,
		MaxEntryBytes: codec.MaxEntryBytes, MaxTotalBytes: codec.MaxBindingBytes,
	}
}

func recoveryPostgresTables(profileID, bindingID string) ([]string, error) {
	switch profileID {
	case networkflow.ProfileID:
		return networkflow.RecoveryPostgresTables(bindingID)
	default:
		return nil, fmt.Errorf("no recovery physical binding provider for profile %s", profileID)
	}
}
