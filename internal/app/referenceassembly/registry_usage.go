package referenceassembly

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/jackc/pgx/v5"
)

// RegistryUsage composes source-owned projections without exposing their
// persistence to Reference Data.
type RegistryUsage struct{}

func (RegistryUsage) ReferencedRegistryEntriesTx(ctx context.Context, tx pgx.Tx, key string) ([]string, error) {
	switch key {
	case "type_registry.host":
		return entities.ReferencedHostRegistryEntriesTx(ctx, tx)
	case "type_registry.evidence":
		return evidence.ReferencedEvidenceRegistryEntriesTx(ctx, tx)
	case "type_registry.indicator":
		return indicators.ReferencedIndicatorRegistryEntriesTx(ctx, tx)
	default:
		return nil, errors.New("reference pack: unknown registry usage owner")
	}
}
