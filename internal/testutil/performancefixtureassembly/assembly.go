package performancefixtureassembly

import (
	"context"
	"fmt"

	authfixture "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/performancefixture"
	entitiesfixture "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport/performancefixture"
	incidentsfixture "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/performancefixture"
	linksfixture "github.com/JochiRaider/cartulary/internal/modules/links/testsupport/performancefixture"
	projectionsfixture "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport/performancefixture"
	timelinefixture "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/performancefixture"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

type Dependencies struct {
	Auth        authfixture.Application
	Incidents   incidentsfixture.Application
	Entities    entitiesfixture.Application
	Timeline    timelinefixture.Application
	Links       linksfixture.Application
	Projections projectionsfixture.Application
	Validation  SemanticValidationApplication
}

type SemanticExpectations struct {
	TimelineRows       int
	HostRows           int
	IdentityRows       int
	TagAssignments     int
	MentionAssignments int
	LinkAssignments    int
	BackgroundAnalysts int
	ActiveSessions     int
	DefaultView        bool
	ProjectionReady    bool
}

type SemanticValidationApplication interface {
	ValidateFixtureSemantics(context.Context, string, SemanticExpectations) (fixture.SemanticValidation, error)
}

func New(dependencies Dependencies) (*fixture.Assembler, error) {
	auth, err := authfixture.New(dependencies.Auth)
	if err != nil {
		return nil, err
	}
	incidents, err := incidentsfixture.New(dependencies.Incidents)
	if err != nil {
		return nil, err
	}
	entities, err := entitiesfixture.New(dependencies.Entities)
	if err != nil {
		return nil, err
	}
	timeline, err := timelinefixture.New(dependencies.Timeline)
	if err != nil {
		return nil, err
	}
	links, err := linksfixture.New(dependencies.Links)
	if err != nil {
		return nil, err
	}
	projections, err := projectionsfixture.New(dependencies.Projections)
	if err != nil {
		return nil, err
	}
	if dependencies.Validation == nil {
		return nil, fmt.Errorf("performance fixture semantic validation application is required")
	}
	return fixture.NewAssembler(
		ExpectedDescriptors(),
		ExpectedSemanticCounts(),
		semanticValidator{application: dependencies.Validation},
		auth,
		incidents,
		entities,
		timeline,
		links,
		projections,
	)
}

func ExpectedDescriptors() []fixture.Descriptor {
	return []fixture.Descriptor{
		authfixture.Descriptor(),
		incidentsfixture.Descriptor(),
		entitiesfixture.Descriptor(),
		timelinefixture.Descriptor(),
		linksfixture.Descriptor(),
		projectionsfixture.Descriptor(),
	}
}

func ExpectedSemanticCounts() map[string]int {
	return map[string]int{
		"active_sessions":     0,
		"background_analysts": 24,
		"host_rows":           1000,
		"identity_rows":       1000,
		"link_assignments":    1000,
		"mention_assignments": 1000,
		"tag_assignments":     1000,
		"timeline_rows":       20000,
	}
}

func ExpectedSemantics() SemanticExpectations {
	return SemanticExpectations{
		TimelineRows:       20000,
		HostRows:           1000,
		IdentityRows:       1000,
		TagAssignments:     1000,
		MentionAssignments: 1000,
		LinkAssignments:    1000,
		BackgroundAnalysts: 24,
		ActiveSessions:     0,
		DefaultView:        true,
		ProjectionReady:    true,
	}
}

type semanticValidator struct {
	application SemanticValidationApplication
}

func (v semanticValidator) Validate(ctx context.Context, state *fixture.BuildState) (fixture.SemanticValidation, error) {
	if state.IncidentID == "" {
		return fixture.SemanticValidation{}, fmt.Errorf("performance fixture semantic validation requires an incident identity")
	}
	return v.application.ValidateFixtureSemantics(ctx, state.IncidentID, ExpectedSemantics())
}
