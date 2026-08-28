package incidents

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/internal/sourcestate"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func RecoveryStateContribution() (recoverystate.Contribution, error) {
	recovery, err := sourcestate.Recovery()
	if err != nil {
		return recoverystate.Contribution{}, fmt.Errorf("incidents Recovery state: %w", err)
	}
	return recoverystate.NewContribution(
		recovery.OwnerID,
		recoverystate.AuthoritativeTables(recovery.Relations...),
	), nil
}
