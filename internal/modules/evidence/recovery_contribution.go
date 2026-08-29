package evidence

import (
	evidencerecovery "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/providers/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func NewRecoveryProvider(db postgres.DB) recovery.EvidenceRecoveryProvider {
	return evidencerecovery.New(db)
}
