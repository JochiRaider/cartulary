package evidence

import (
	projectionprovider "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/providers/projection"
	projectioncontract "github.com/JochiRaider/cartulary/internal/modules/evidence/projectioncontract"
)

// NewProjectionContribution constructs Evidence's complete source-owned
// projection contribution without exposing its concrete source provider.
func NewProjectionContribution() (projectioncontract.Contribution, error) {
	return projectioncontract.NewContribution(projectionprovider.NewSource())
}
