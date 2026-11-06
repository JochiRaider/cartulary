package links

import (
	"context"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/links/internal/valuecodec"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) publishProjectionEffectsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, mutations []Mutation) error {
	if s.projections == nil || len(mutations) == 0 {
		return nil
	}
	ids := make(map[uuid.UUID]struct{})
	for _, mutation := range mutations {
		if mutation.TargetKind() != "record_link" {
			continue
		}
		for _, value := range []any{mutation.BeforeValue(), mutation.AfterValue()} {
			if value == nil {
				continue
			}
			link, err := valuecodec.DecodeRecordLinkMutationValue(value.(map[string]any))
			if err != nil {
				return err
			}
			ids[link.SrcRecordID] = struct{}{}
			ids[link.DstRecordID] = struct{}{}
		}
	}
	ordered := make([]uuid.UUID, 0, len(ids))
	for id := range ids {
		ordered = append(ordered, id)
	}
	slices.SortFunc(ordered, func(a, b uuid.UUID) int { return strings.Compare(a.String(), b.String()) })
	if len(ordered) == 0 {
		return nil
	}
	return s.projections.RefreshLinkedRecordsTx(ctx, tx, incidentID, ordered)
}
