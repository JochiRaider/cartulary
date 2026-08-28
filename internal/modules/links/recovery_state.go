package links

import (
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/sourcestate"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func RecoveryStateContribution() (recoverystate.Contribution, error) {
	return sourcestate.RecoveryStateContribution()
}
