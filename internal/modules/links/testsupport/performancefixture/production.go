package performancefixture

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type ProductionApplication struct {
	pool postgres.DB
}

func NewProductionApplication(pool postgres.DB) (*ProductionApplication, error) {
	if pool == nil {
		return nil, fmt.Errorf("links performance fixture Postgres is required")
	}
	return &ProductionApplication{pool: pool}, nil
}

func (a *ProductionApplication) ValidateFixtureAssociations(ctx context.Context, incidentID string, expected Expectations) error {
	incidentUUID, err := uuid.Parse(incidentID)
	if err != nil {
		return err
	}
	queries := []struct {
		name  string
		query string
		want  int
	}{
		{
			name:  "links",
			query: `SELECT COUNT(*) FROM record_links WHERE incident_id = $1 AND link_type = 'observed_on_host' AND deleted_at IS NULL`,
			want:  expected.Links,
		},
		{
			name:  "mentions",
			query: `SELECT COUNT(*) FROM entity_mentions m JOIN records r ON r.record_id = m.source_record_id WHERE r.incident_id = $1 AND m.source_field_key = 'timeline.identity_refs'`,
			want:  expected.Mentions,
		},
		{
			name:  "tags",
			query: `SELECT COUNT(*) FROM record_tags WHERE incident_id = $1 AND normalized_tag_name LIKE 'perf-tag-%' AND deleted_at IS NULL`,
			want:  expected.Tags,
		},
	}
	for _, item := range queries {
		var got int
		if err := a.pool.QueryRow(ctx, item.query, incidentUUID).Scan(&got); err != nil {
			return fmt.Errorf("query fixture %s: %w", item.name, err)
		}
		if got != item.want {
			return fmt.Errorf("fixture %s count=%d want=%d", item.name, got, item.want)
		}
	}
	if expected.Stride != 20 {
		return fmt.Errorf("fixture association stride=%d want=20", expected.Stride)
	}
	return nil
}
