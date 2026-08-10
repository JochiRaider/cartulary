package artifacts

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() (recoverystate.Contribution, error) {
	manifest, err := loadSourceStateManifest()
	if err != nil {
		return recoverystate.Contribution{}, err
	}
	return recoverystate.NewContribution("module.artifacts", manifest.recoveryTables()), nil
}
