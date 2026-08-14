package performancefixture

import (
	"context"
	"errors"
	"fmt"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

const ContributionID = "auth.background_analysts.v1"

type CreateBackgroundAnalystRequest struct {
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
	CreateBackgroundAnalyst(context.Context, CreateBackgroundAnalystRequest) (Account, error)
}

type Provider struct {
	application Application
}

func New(application Application) (*Provider, error) {
	if application == nil {
		return nil, errors.New("auth performance fixture application is required")
	}
	return &Provider{application: application}, nil
}

func Descriptor() fixture.Descriptor {
	return fixture.Descriptor{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.auth",
		Dependencies:   []string{},
		ExpectedCounts: map[string]int{"accounts": 24, "sessions": 0},
	}
}

func (p *Provider) Descriptor() fixture.Descriptor { return Descriptor() }

func (p *Provider) Apply(ctx context.Context, state *fixture.BuildState) (fixture.Receipt, error) {
	if len(state.RuntimeBundle.BackgroundAccounts) != 24 {
		return fixture.Receipt{}, fmt.Errorf("auth performance fixture requires 24 background credentials, got %d", len(state.RuntimeBundle.BackgroundAccounts))
	}
	state.BackgroundUserIDs = make([]string, 0, len(state.RuntimeBundle.BackgroundAccounts))
	for index, credential := range state.RuntimeBundle.BackgroundAccounts {
		account, err := p.application.CreateBackgroundAnalyst(ctx, CreateBackgroundAnalystRequest{
			DisplayName:       fmt.Sprintf("AC-043 Background Analyst %02d", index+1),
			Email:             credential.Email,
			InitialPassword:   credential.Password,
			IsDeploymentAdmin: false,
			MFARequired:       false,
			Active:            true,
		})
		if err != nil {
			return fixture.Receipt{}, fmt.Errorf("create background analyst %d: %w", index+1, err)
		}
		if account.UserID == "" {
			return fixture.Receipt{}, fmt.Errorf("create background analyst %d returned an empty user identity", index+1)
		}
		state.BackgroundUserIDs = append(state.BackgroundUserIDs, account.UserID)
	}
	return fixture.Receipt{
		ContributionID: ContributionID,
		Version:        ContributionID,
		OwnerID:        "module.auth",
		Counts:         map[string]int{"accounts": 24, "sessions": 0},
	}, nil
}
