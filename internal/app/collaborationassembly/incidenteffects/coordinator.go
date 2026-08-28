package incidenteffects

import (
	"context"
	"errors"
	"reflect"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type Application interface {
	TransitionIncidentLifecycle(
		context.Context,
		authn.UserRecord,
		uuid.UUID,
		incidents.IncidentLifecycleAdmission,
		string,
	) (incidents.IncidentLifecycleResult, error)
	DeleteMembership(
		context.Context,
		authn.UserRecord,
		uuid.UUID,
		uuid.UUID,
		incidents.MembershipDeleteAdmission,
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
	request incidents.IncidentLifecycleAdmission,
	requestID string,
) (incidents.IncidentLifecycleResult, error) {
	result, err := c.application.TransitionIncidentLifecycle(
		ctx,
		actor,
		incidentID,
		request,
		requestID,
	)
	if err != nil {
		return incidents.IncidentLifecycleResult{}, err
	}
	if result.Commit.IsReplay() {
		return result, nil
	}
	effectKey, isNewCommit := result.Commit.EffectKey()
	if !isNewCommit {
		return incidents.IncidentLifecycleResult{}, errors.New("incident effects: invalid lifecycle commit result")
	}
	if request.Action() == incidents.LifecycleActionClose {
		c.notifier.NotifyIncidentClosed(ctx, effectKey, incidentID)
	}
	return result, nil
}

func (c *Coordinator) CoordinateMembershipDeletion(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	userID uuid.UUID,
	request incidents.MembershipDeleteAdmission,
	requestID string,
) (incidents.MembershipDeleteResult, error) {
	result, err := c.application.DeleteMembership(ctx, actor, incidentID, userID, request, requestID)
	if err != nil {
		return incidents.MembershipDeleteResult{}, err
	}
	effectKey, isNewCommit := result.Commit.EffectKey()
	if !isNewCommit {
		return incidents.MembershipDeleteResult{}, errors.New("incident effects: membership delete cannot replay")
	}
	c.notifier.NotifyIncidentMembershipRevoked(ctx, effectKey, incidentID, userID)
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
