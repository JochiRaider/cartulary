package incidenteffects

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type Application interface {
	TransitionIncidentLifecycle(
		context.Context,
		authn.UserRecord,
		uuid.UUID,
		string,
		incidents.IncidentLifecycleRequest,
		string,
		time.Time,
	) (incidents.IncidentLifecycleResult, error)
	DeleteMembership(
		context.Context,
		authn.UserRecord,
		uuid.UUID,
		uuid.UUID,
		incidents.MembershipDeleteRequest,
		string,
	) (incidents.MembershipDeleteResult, error)
}

type Notifier interface {
	NotifyIncidentClosed(context.Context, uuid.UUID, uuid.UUID)
	NotifyIncidentMembershipRevoked(context.Context, uuid.UUID, uuid.UUID, uuid.UUID)
}

// Coordinator owns the post-commit boundary between Incidents mutations and
// process-local Collaboration session effects. The application returns only
// after commit, and the notifier has no fallible delivery contract.
type Coordinator struct {
	application Application
	notifier    Notifier
}

func New(application Application, notifier Notifier) (*Coordinator, error) {
	if isNil(application) {
		return nil, errors.New("incident effects: application is required")
	}
	if isNil(notifier) {
		return nil, errors.New("incident effects: collaboration notifier is required")
	}
	return &Coordinator{application: application, notifier: notifier}, nil
}

func (c *Coordinator) CoordinateIncidentLifecycle(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	action string,
	request incidents.IncidentLifecycleRequest,
	requestID string,
	now time.Time,
) (incidents.IncidentLifecycleResult, error) {
	result, err := c.application.TransitionIncidentLifecycle(
		ctx,
		actor,
		incidentID,
		action,
		request,
		requestID,
		now,
	)
	if err != nil {
		return incidents.IncidentLifecycleResult{}, err
	}
	if err := result.Commit.Validate(); err != nil {
		return incidents.IncidentLifecycleResult{}, fmt.Errorf("incident effects: invalid lifecycle commit result: %w", err)
	}
	if action == "close" && result.Commit.Disposition == incidents.TerminalMutationNewCommit {
		c.notifier.NotifyIncidentClosed(ctx, result.Commit.EffectKey, incidentID)
	}
	return result, nil
}

func (c *Coordinator) CoordinateMembershipDeletion(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	userID uuid.UUID,
	request incidents.MembershipDeleteRequest,
	requestID string,
) (incidents.MembershipDeleteResult, error) {
	result, err := c.application.DeleteMembership(ctx, actor, incidentID, userID, request, requestID)
	if err != nil {
		return incidents.MembershipDeleteResult{}, err
	}
	if err := result.Commit.Validate(); err != nil {
		return incidents.MembershipDeleteResult{}, fmt.Errorf("incident effects: invalid membership delete commit result: %w", err)
	}
	if result.Commit.Disposition != incidents.TerminalMutationNewCommit {
		return incidents.MembershipDeleteResult{}, errors.New("incident effects: membership delete cannot replay")
	}
	c.notifier.NotifyIncidentMembershipRevoked(ctx, result.Commit.EffectKey, incidentID, userID)
	return result, nil
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	typed := reflect.ValueOf(value)
	switch typed.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return typed.IsNil()
	default:
		return false
	}
}
