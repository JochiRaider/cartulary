package performancefixture

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	authfixture "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/performancefixture"
	entitiesfixture "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport/performancefixture"
	incidentsfixture "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/performancefixture"
	linksfixture "github.com/JochiRaider/cartulary/internal/modules/links/testsupport/performancefixture"
	projectionsfixture "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport/performancefixture"
	timelinefixture "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/performancefixture"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

func NewProduction(
	pool postgres.DB,
	actor authn.UserRecord,
	conflictTokenSeed string,
	profile performancefixtureprofile.Profile,
) (*fixture.Assembler, error) {
	if len(conflictTokenSeed) < 16 {
		return nil, errors.New("performance fixture conflict token seed is incomplete")
	}
	owners, err := NewOwners(pool, conflicttest.NewCodec(conflictTokenSeed[:16]))
	if err != nil {
		return nil, err
	}
	authApplication, err := authfixture.NewProductionApplication(pool, actor)
	if err != nil {
		return nil, err
	}
	incidentApplication, err := incidentsfixture.NewProductionApplication(pool, actor)
	if err != nil {
		return nil, err
	}
	entityApplication, err := entitiesfixture.NewProductionApplication(owners.Entities, actor)
	if err != nil {
		return nil, err
	}
	timelineApplication, err := timelinefixture.NewProductionApplication(owners.Timeline, actor)
	if err != nil {
		return nil, err
	}
	linkApplication, err := linksfixture.NewProductionApplication(pool)
	if err != nil {
		return nil, err
	}
	projectionApplication, err := projectionsfixture.NewProductionApplication(owners.Projections)
	if err != nil {
		return nil, err
	}
	return New(profile, Dependencies{
		Auth:        authApplication,
		Incidents:   incidentApplication,
		Entities:    entityApplication,
		Timeline:    timelineApplication,
		Links:       linkApplication,
		Projections: projectionApplication,
		Validation:  &productionSemanticValidator{pool: pool, profile: profile},
	})
}

type productionSemanticValidator struct {
	pool    postgres.DB
	profile performancefixtureprofile.Profile
}

func (v *productionSemanticValidator) ValidateFixtureSemantics(ctx context.Context, incidentID string, expected SemanticExpectations) (fixture.SemanticValidation, error) {
	incidentUUID, err := uuid.Parse(incidentID)
	if err != nil {
		return fixture.SemanticValidation{}, err
	}
	counts := map[string]int{}
	queries := []struct {
		key   string
		query string
		args  []any
	}{
		{key: "timeline_rows", query: `SELECT COUNT(*) FROM timeline_events WHERE incident_id = $1`, args: []any{incidentUUID}},
		{key: "host_rows", query: `SELECT COUNT(*) FROM hosts WHERE incident_id = $1 AND merged_into_record_id IS NULL`, args: []any{incidentUUID}},
		{key: "identity_rows", query: `SELECT COUNT(*) FROM identities WHERE incident_id = $1 AND merged_into_record_id IS NULL`, args: []any{incidentUUID}},
		{key: "tag_assignments", query: `SELECT COUNT(*) FROM record_tags WHERE incident_id = $1 AND normalized_tag_name LIKE 'perf-tag-%' AND deleted_at IS NULL`, args: []any{incidentUUID}},
		{key: "mention_assignments", query: `SELECT COUNT(*) FROM entity_mentions m JOIN records r ON r.record_id = m.source_record_id WHERE r.incident_id = $1 AND m.source_field_key = 'timeline.identity_refs'`, args: []any{incidentUUID}},
		{key: "link_assignments", query: `SELECT COUNT(*) FROM record_links WHERE incident_id = $1 AND link_type = 'observed_on_host' AND deleted_at IS NULL`, args: []any{incidentUUID}},
		{key: "background_analysts", query: `SELECT COUNT(*) FROM incident_memberships m JOIN users u ON u.id = m.user_id WHERE m.incident_id = $1 AND m.role = 'editor' AND u.is_active AND NOT u.is_deployment_admin AND NOT u.mfa_required`, args: []any{incidentUUID}},
		{key: "active_sessions", query: `SELECT COUNT(*) FROM user_sessions WHERE revoked_at IS NULL`},
	}
	for _, query := range queries {
		var count int
		if err := v.pool.QueryRow(ctx, query.query, query.args...).Scan(&count); err != nil {
			return fixture.SemanticValidation{}, fmt.Errorf("query performance fixture %s: %w", query.key, err)
		}
		counts[query.key] = count
	}
	var defaultViewCount int
	if err := v.pool.QueryRow(ctx, `SELECT COUNT(*) FROM incident_workbook_preferences WHERE incident_id = $1`, incidentUUID).Scan(&defaultViewCount); err != nil {
		return fixture.SemanticValidation{}, fmt.Errorf("query performance fixture default view: %w", err)
	}
	var distributedRows int
	if err := v.pool.QueryRow(ctx, `
SELECT COUNT(*)
  FROM timeline_events t
 WHERE t.incident_id = $1
   AND t.activity_synopsis_text LIKE 'Performance Timeline %'
   AND t.data_source_text LIKE 'https://fixture-%'`, incidentUUID).Scan(&distributedRows); err != nil {
		return fixture.SemanticValidation{}, fmt.Errorf("query performance fixture relationship distribution: %w", err)
	}
	generated, err := semanticExpectations(v.profile)
	if err != nil {
		return fixture.SemanticValidation{}, err
	}
	return fixture.SemanticValidation{
		Counts: counts,
		Conditions: map[string]bool{
			"authorization":             counts["background_analysts"] == generated.BackgroundAnalysts,
			"default_view":              expected.DefaultView && defaultViewCount == 1,
			"no_active_sessions":        counts["active_sessions"] == expected.ActiveSessions,
			"projection_ready":          expected.ProjectionReady,
			"relationship_distribution": distributedRows == expected.LinkAssignments,
		},
	}, nil
}
