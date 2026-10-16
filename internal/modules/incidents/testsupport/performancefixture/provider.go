package performancefixture

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

type WorkspaceIdentity struct {
	KeyPrefix string
	Title     string
}

type WorkspaceRequest struct {
	Seed     int
	Identity WorkspaceIdentity
}

type Application interface {
	CreateFixtureWorkspaceIncident(context.Context, WorkspaceRequest) (string, error)
	AddFixtureMembership(context.Context, string, string, string) error
}

type Provider struct {
	application Application
	descriptor  fixture.Descriptor
	identity    WorkspaceIdentity
}

func New(application Application, descriptor fixture.Descriptor, identity WorkspaceIdentity) (*Provider, error) {
	if application == nil || strings.TrimSpace(identity.KeyPrefix) == "" || strings.TrimSpace(identity.Title) == "" {
		return nil, errors.New("incidents performance fixture application is required")
	}
	if descriptor.ExpectedCounts["incidents"] != 1 || descriptor.ExpectedCounts["workspaces"] != 1 || descriptor.ExpectedCounts["memberships"] < 1 {
		return nil, errors.New("incidents performance fixture descriptor is incompatible")
	}
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return &Provider{application: application, descriptor: descriptor, identity: identity}, nil
}

func (p *Provider) Descriptor() fixture.Descriptor {
	result := p.descriptor
	result.Dependencies = slices.Clone(result.Dependencies)
	result.ExpectedCounts = maps.Clone(result.ExpectedCounts)
	return result
}

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	wantMemberships := p.descriptor.ExpectedCounts["memberships"]
	if len(state.PrincipalUserIDs) != wantMemberships {
		return fixture.Receipt{}, fmt.Errorf("incidents performance fixture requires %d Auth identities, got %d", wantMemberships, len(state.PrincipalUserIDs))
	}
	incidentID, err := p.application.CreateFixtureWorkspaceIncident(ctx, WorkspaceRequest{Seed: state.Seed, Identity: p.identity})
	if err != nil {
		return fixture.Receipt{}, fmt.Errorf("create fixture workspace incident: %w", err)
	}
	if incidentID == "" {
		return fixture.Receipt{}, errors.New("create fixture workspace incident returned an empty identity")
	}
	for _, userID := range state.PrincipalUserIDs {
		if err := p.application.AddFixtureMembership(ctx, incidentID, userID, "editor"); err != nil {
			return fixture.Receipt{}, fmt.Errorf("add fixture membership: %w", err)
		}
	}
	state.IncidentID = incidentID
	return fixture.Receipt{
		ContributionID: p.descriptor.ContributionID,
		Version:        p.descriptor.Version,
		OwnerID:        p.descriptor.OwnerID,
		Counts:         maps.Clone(p.descriptor.ExpectedCounts),
	}, nil
}
