package performancefixture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type ProductionApplication struct {
	actor authn.UserRecord
	store *authn.Store
	now   func() time.Time
}

func NewProductionApplication(pool postgres.DB, actor authn.UserRecord) (*ProductionApplication, error) {
	if pool == nil {
		return nil, fmt.Errorf("auth performance fixture Postgres is required")
	}
	if actor.ID.String() == "00000000-0000-0000-0000-000000000000" || !actor.IsActive || !actor.IsDeploymentAdmin {
		return nil, fmt.Errorf("auth performance fixture requires an active deployment-admin actor")
	}
	return &ProductionApplication{actor: actor, store: authn.NewStore(pool), now: time.Now}, nil
}

func (a *ProductionApplication) CreateBackgroundAnalyst(ctx context.Context, request CreateBackgroundAnalystRequest) (Account, error) {
	email, _, ok := authn.NormalizeEmailAddress(request.Email)
	if !ok {
		return Account{}, fmt.Errorf("normalize background analyst email")
	}
	displayName, ok := authn.NormalizeDisplayNameLine(request.DisplayName)
	if !ok {
		return Account{}, fmt.Errorf("normalize background analyst display name")
	}
	if _, err := authn.ValidatePasswordProvision(request.InitialPassword); err != nil {
		return Account{}, err
	}
	passwordHash, err := authn.HashPassword(request.InitialPassword)
	if err != nil {
		return Account{}, err
	}
	fingerprint := sha256.Sum256([]byte(email))
	clientTxnID := "performance-fixture-background-user-" + hex.EncodeToString(fingerprint[:8])
	payload, err := json.Marshal(map[string]any{
		"auth_kind":           "local",
		"display_name":        displayName,
		"email_digest":        hex.EncodeToString(fingerprint[:]),
		"is_deployment_admin": request.IsDeploymentAdmin,
		"mfa_required":        request.MFARequired,
		"password_hash":       passwordHash,
	})
	if err != nil {
		return Account{}, err
	}
	requestHash := sha256.Sum256(payload)
	if _, err := a.store.CreateUser(
		ctx,
		a.actor,
		email,
		displayName,
		passwordHash,
		request.MFARequired,
		request.IsDeploymentAdmin,
		clientTxnID,
		requestHash[:],
		"req-"+clientTxnID,
		a.now().UTC(),
	); err != nil {
		return Account{}, err
	}
	created, err := a.store.GetUserByNormalizedEmail(ctx, email)
	if err != nil {
		return Account{}, err
	}
	if created.IsActive != request.Active {
		return Account{}, fmt.Errorf("created background analyst active state mismatch")
	}
	return Account{UserID: created.ID.String()}, nil
}
