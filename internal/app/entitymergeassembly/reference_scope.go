package entitymergeassembly

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/links"
)

type referenceDependencyReader interface {
	ProjectionDependentsTx(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) ([]uuid.UUID, error)
}

type referenceScope struct {
	links    referenceDependencyReader
	mentions referenceDependencyReader
}

func NewReferenceScope() merge.ReferenceScopePort {
	return referenceScope{links: links.NewStore(), mentions: entities.ProjectionDependencies{}}
}

func (scope referenceScope) DependentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, changed []uuid.UUID) ([]uuid.UUID, error) {
	var result []uuid.UUID
	// Both owners receive only the original authoritative scope. Returned peers
	// are never fed into another owner as a transitive traversal.
	for _, reader := range []referenceDependencyReader{scope.links, scope.mentions} {
		dependents, err := reader.ProjectionDependentsTx(ctx, tx, incidentID, append([]uuid.UUID(nil), changed...))
		if err != nil {
			return nil, err
		}
		result = append(result, dependents...)
	}
	return result, nil
}
