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

type CreateFixtureAnalystRequest struct {
	DisplayName       string
	Email             string
	InitialPassword   string
	IsDeploymentAdmin bool
	MFARequired       bool
	Active            bool
}

type Account struct {
	UserID string
}

type Application interface {
	CreateFixtureAnalyst(context.Context, CreateFixtureAnalystRequest) (Account, error)
}

type Provider struct {
	application       Application
	credentialSetID   string
	displayNamePrefix string
	descriptor        fixture.Descriptor
}

func New(application Application, descriptor fixture.Descriptor, credentialSetID string, displayNamePrefix string) (*Provider, error) {
	if application == nil {
		return nil, errors.New("auth performance fixture application is required")
	}
	if descriptor.ExpectedCounts["accounts"] < 1 || descriptor.ExpectedCounts["sessions"] != 0 {
		return nil, errors.New("auth performance fixture descriptor is incompatible")
	}
	if credentialSetID == "" || strings.TrimSpace(displayNamePrefix) == "" {
		return nil, errors.New("auth performance fixture credential set identity is required")
	}
	descriptor.Dependencies = slices.Clone(descriptor.Dependencies)
	descriptor.ExpectedCounts = maps.Clone(descriptor.ExpectedCounts)
	return &Provider{application: application, credentialSetID: credentialSetID, displayNamePrefix: displayNamePrefix, descriptor: descriptor}, nil
}

func (p *Provider) Descriptor() fixture.Descriptor {
	result := p.descriptor
	result.Dependencies = slices.Clone(result.Dependencies)
	result.ExpectedCounts = maps.Clone(result.ExpectedCounts)
	return result
}

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	wantAccounts := p.descriptor.ExpectedCounts["accounts"]
	credentials, ok := state.RuntimeBundle.Credentials(p.credentialSetID)
	if !ok || len(credentials) != wantAccounts {
		return fixture.Receipt{}, fmt.Errorf("auth performance fixture requires %d fixture credentials, got %d", wantAccounts, len(credentials))
	}
	state.PrincipalUserIDs = make([]string, 0, len(credentials))
	for index, credential := range credentials {
		account, err := p.application.CreateFixtureAnalyst(ctx, CreateFixtureAnalystRequest{
			DisplayName:       fmt.Sprintf("%s %02d", p.displayNamePrefix, index+1),
			Email:             credential.Principal,
			InitialPassword:   credential.Secret,
			IsDeploymentAdmin: false,
			MFARequired:       false,
			Active:            true,
		})
		if err != nil {
			return fixture.Receipt{}, fmt.Errorf("create fixture analyst %d: %w", index+1, err)
		}
		if account.UserID == "" {
			return fixture.Receipt{}, fmt.Errorf("create fixture analyst %d returned an empty user identity", index+1)
		}
		state.PrincipalUserIDs = append(state.PrincipalUserIDs, account.UserID)
	}
	return fixture.Receipt{
		ContributionID: p.descriptor.ContributionID,
		Version:        p.descriptor.Version,
		OwnerID:        p.descriptor.OwnerID,
		Counts:         maps.Clone(p.descriptor.ExpectedCounts),
	}, nil
}
