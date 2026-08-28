package incidenteffects

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/admissiontest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type applicationStub struct {
	lifecycleResult incidents.IncidentLifecycleResult
	lifecycleErr    error
	deleteResult    incidents.MembershipDeleteResult
	deleteErr       error
	sequence        *[]string
}

func (stub *applicationStub) TransitionIncidentLifecycle(
	context.Context,
	authn.UserRecord,
	uuid.UUID,
	incidents.IncidentLifecycleAdmission,
	string,
) (incidents.IncidentLifecycleResult, error) {
	if stub.sequence != nil {
		*stub.sequence = append(*stub.sequence, "commit")
	}
	return stub.lifecycleResult, stub.lifecycleErr
}

func (stub *applicationStub) DeleteMembership(
	context.Context,
	authn.UserRecord,
	uuid.UUID,
	uuid.UUID,
	incidents.MembershipDeleteAdmission,
	string,
) (incidents.MembershipDeleteResult, error) {
	if stub.sequence != nil {
		*stub.sequence = append(*stub.sequence, "commit")
	}
	return stub.deleteResult, stub.deleteErr
}

type notification struct {
	kind       string
	effectKey  uuid.UUID
	incidentID uuid.UUID
	userID     uuid.UUID
}

type notifierStub struct {
	notifications []notification
	sequence      *[]string
}

func (stub *notifierStub) NotifyIncidentClosed(_ context.Context, effectKey uuid.UUID, incidentID uuid.UUID) {
	if stub.sequence != nil {
		*stub.sequence = append(*stub.sequence, "effect")
	}
	stub.notifications = append(stub.notifications, notification{kind: "closed", effectKey: effectKey, incidentID: incidentID})
}

func (stub *notifierStub) NotifyIncidentMembershipRevoked(_ context.Context, effectKey uuid.UUID, incidentID uuid.UUID, userID uuid.UUID) {
	if stub.sequence != nil {
		*stub.sequence = append(*stub.sequence, "effect")
	}
	stub.notifications = append(stub.notifications, notification{kind: "membership_revoked", effectKey: effectKey, incidentID: incidentID, userID: userID})
}

func TestCoordinatorRequiresCompleteComposition(t *testing.T) {
	notifier := &notifierStub{}
	if _, err := New(nil, notifier); err == nil {
		t.Fatal("missing application must fail composition")
	}
	var typedNilApplication *applicationStub
	if _, err := New(typedNilApplication, notifier); err == nil {
		t.Fatal("typed nil application must fail composition")
	}
	if _, err := New(&applicationStub{}, nil); err == nil {
		t.Fatal("missing notifier must fail composition")
	}
}

func TestCoordinatorEmitsOnlyAfterFreshTerminalCommit(t *testing.T) {
	incidentID := uuid.New()
	effectKey := uuid.New()
	sequence := make([]string, 0, 2)
	notifier := &notifierStub{sequence: &sequence}
	application := &applicationStub{
		lifecycleResult: incidents.IncidentLifecycleResult{Commit: requireNewCommit(t, effectKey)},
		sequence:        &sequence,
	}
	coordinator, err := New(application, notifier)
	if err != nil {
		t.Fatalf("compose coordinator: %v", err)
	}
	closeRequest := admissiontest.IncidentLifecycle(t, incidents.LifecycleActionClose, admissiontest.IncidentLifecycleInput{
		BaseIncidentVersion: 1, ClientTxnID: "txn-close", Reason: "closed",
	})
	reopenRequest := admissiontest.IncidentLifecycle(t, incidents.LifecycleActionReopen, admissiontest.IncidentLifecycleInput{
		BaseIncidentVersion: 2, ClientTxnID: "txn-reopen", Reason: "reopened",
	})
	if _, err := coordinator.CoordinateIncidentLifecycle(
		context.Background(), authn.UserRecord{}, incidentID,
		closeRequest, "request",
	); err != nil {
		t.Fatalf("coordinate close: %v", err)
	}
	if !reflect.DeepEqual(sequence, []string{"commit", "effect"}) {
		t.Fatalf("effect ordering = %#v, want commit then effect", sequence)
	}
	want := []notification{{kind: "closed", effectKey: effectKey, incidentID: incidentID}}
	if !reflect.DeepEqual(notifier.notifications, want) {
		t.Fatalf("close notifications = %#v, want %#v", notifier.notifications, want)
	}

	application.lifecycleResult.Commit = incidents.ReplayTerminalMutationCommit()
	if _, err := coordinator.CoordinateIncidentLifecycle(
		context.Background(), authn.UserRecord{}, incidentID,
		closeRequest, "request",
	); err != nil {
		t.Fatalf("coordinate close replay: %v", err)
	}
	application.lifecycleResult.Commit = requireNewCommit(t, uuid.New())
	if _, err := coordinator.CoordinateIncidentLifecycle(
		context.Background(), authn.UserRecord{}, incidentID,
		reopenRequest, "request",
	); err != nil {
		t.Fatalf("coordinate reopen: %v", err)
	}
	if len(notifier.notifications) != 1 {
		t.Fatalf("replay or reopen emitted terminal effects: %#v", notifier.notifications)
	}
}

func TestCoordinatorSuppressesEffectsForFailureAndInvalidResults(t *testing.T) {
	application := &applicationStub{lifecycleErr: errors.New("rollback")}
	notifier := &notifierStub{}
	coordinator, err := New(application, notifier)
	if err != nil {
		t.Fatalf("compose coordinator: %v", err)
	}
	closeRequest := admissiontest.IncidentLifecycle(t, incidents.LifecycleActionClose, admissiontest.IncidentLifecycleInput{
		BaseIncidentVersion: 1, ClientTxnID: "txn-failure", Reason: "closed",
	})
	if _, err := coordinator.CoordinateIncidentLifecycle(
		context.Background(), authn.UserRecord{}, uuid.New(),
		closeRequest, "request",
	); err == nil {
		t.Fatal("application failure must be returned")
	}
	if len(notifier.notifications) != 0 {
		t.Fatalf("failed mutation emitted effects: %#v", notifier.notifications)
	}

	application.lifecycleErr = nil
	application.lifecycleResult = incidents.IncidentLifecycleResult{}
	if _, err := coordinator.CoordinateIncidentLifecycle(
		context.Background(), authn.UserRecord{}, uuid.New(),
		closeRequest, "request",
	); err == nil {
		t.Fatal("unknown disposition must fail closed")
	}
	if len(notifier.notifications) != 0 {
		t.Fatalf("invalid result emitted effects: %#v", notifier.notifications)
	}
}

func TestCoordinatorTargetsOnlyDeletedMembership(t *testing.T) {
	incidentID := uuid.New()
	userID := uuid.New()
	effectKey := uuid.New()
	application := &applicationStub{
		deleteResult: incidents.MembershipDeleteResult{Commit: requireNewCommit(t, effectKey)},
	}
	notifier := &notifierStub{}
	coordinator, err := New(application, notifier)
	if err != nil {
		t.Fatalf("compose coordinator: %v", err)
	}
	deleteRequest := admissiontest.MembershipDelete(t, 1)
	if _, err := coordinator.CoordinateMembershipDeletion(
		context.Background(), authn.UserRecord{}, incidentID, userID,
		deleteRequest, "request",
	); err != nil {
		t.Fatalf("coordinate membership delete: %v", err)
	}
	want := []notification{{kind: "membership_revoked", effectKey: effectKey, incidentID: incidentID, userID: userID}}
	if !reflect.DeepEqual(notifier.notifications, want) {
		t.Fatalf("membership notifications = %#v, want %#v", notifier.notifications, want)
	}

	application.deleteResult.Commit = incidents.ReplayTerminalMutationCommit()
	if _, err := coordinator.CoordinateMembershipDeletion(
		context.Background(), authn.UserRecord{}, incidentID, userID,
		deleteRequest, "request",
	); err == nil {
		t.Fatal("membership delete replay must fail closed")
	}
	if len(notifier.notifications) != 1 {
		t.Fatalf("invalid delete result emitted effects: %#v", notifier.notifications)
	}
}

func requireNewCommit(t testing.TB, effectKey uuid.UUID) incidents.TerminalMutationCommit {
	t.Helper()
	commit, err := incidents.NewTerminalMutationCommit(effectKey)
	if err != nil {
		t.Fatalf("construct terminal mutation commit: %v", err)
	}
	return commit
}
