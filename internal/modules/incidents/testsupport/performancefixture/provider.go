package performancefixture

import (
	"context"
	"errors"
	"fmt"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

const ContributionID = "incidents.workspace.v1"

type Application interface {
	CreateFixtureWorkspaceIncident(context.Context, int) (string, error)
	AddFixtureMembership(context.Context, string, string, string) error
}

type Provider struct {
	application Application
}

func New(application Application) (*Provider, error) {
	if application == nil {
		return nil, errors.New("incidents performance fixture application is required")
	}
	return &Provider{application: application}, nil
}

func Descriptor() fixture.Descriptor {
	return fixture.Descriptor{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.incidents",
		Dependencies:   []string{"auth.background_analysts.v1"},
		ExpectedCounts: map[string]int{"incidents": 1, "memberships": 24, "workspaces": 1},
	}
}

func (p *Provider) Descriptor() fixture.Descriptor { return Descriptor() }

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if len(state.BackgroundUserIDs) != 24 {
		return fixture.Receipt{}, fmt.Errorf("incidents performance fixture requires 24 Auth identities, got %d", len(state.BackgroundUserIDs))
	}
	incidentID, err := p.application.CreateFixtureWorkspaceIncident(ctx, state.Seed)
	if err != nil {
		return fixture.Receipt{}, fmt.Errorf("create fixture workspace incident: %w", err)
	}
	if incidentID == "" {
		return fixture.Receipt{}, errors.New("create fixture workspace incident returned an empty identity")
	}
	for _, userID := range state.BackgroundUserIDs {
		if err := p.application.AddFixtureMembership(ctx, incidentID, userID, "editor"); err != nil {
			return fixture.Receipt{}, fmt.Errorf("add fixture membership: %w", err)
		}
	}
	state.IncidentID = incidentID
	return fixture.Receipt{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.incidents",
		Counts:         map[string]int{"incidents": 1, "memberships": 24, "workspaces": 1},
	}, nil
}
