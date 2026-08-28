package incidents

import (
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
)

func TestIncidentAuditKindAndSourceMappingsAreClosedAndExhaustive_Unit(t *testing.T) {
	t.Parallel()
	kinds := []struct {
		kind      auditEventKind
		raw       string
		action    string
		projected bool
	}{
		{kind: auditIncidentCreated, raw: "incident_created"},
		{kind: auditIncidentUpdated, raw: "incident_updated"},
		{kind: auditIncidentClosed, raw: "incident_close"},
		{kind: auditIncidentReopened, raw: "incident_reopen"},
		{kind: auditMembershipCreated, raw: "incident_membership_created", action: administrativeaudit.ActionMembershipCreated, projected: true},
		{kind: auditMembershipUpdated, raw: "incident_membership_updated", action: administrativeaudit.ActionMembershipRoleChanged, projected: true},
		{kind: auditMembershipDeleted, raw: "incident_membership_deleted", action: administrativeaudit.ActionMembershipDeleted, projected: true},
	}
	for _, test := range kinds {
		if got := test.kind.rawValue(); got != test.raw {
			t.Fatalf("audit kind %d raw value = %q, want %q", test.kind, got, test.raw)
		}
		action, projected := test.kind.membershipAction()
		if action != test.action || projected != test.projected {
			t.Fatalf("audit kind %d projection = (%q, %t), want (%q, %t)", test.kind, action, projected, test.action, test.projected)
		}
	}
	if got := auditEventKind(0).rawValue(); got != "" {
		t.Fatalf("unknown audit kind mapped to %q", got)
	}
	if action, projected := auditEventKind(0).membershipAction(); action != "" || projected {
		t.Fatalf("unknown audit kind projection = (%q, %t)", action, projected)
	}

	if got := auditSourceAPI.publicValue(); got != administrativeaudit.SourceAPI {
		t.Fatalf("API audit source = %q", got)
	}
	if got := auditSourceSystem.publicValue(); got != administrativeaudit.SourceSystem {
		t.Fatalf("system audit source = %q", got)
	}
	if got := auditEventSource(0).publicValue(); got != "" {
		t.Fatalf("unknown audit source mapped to %q", got)
	}
}

func TestMembershipAuditRoleFactsAreExplicitAndValidated_Unit(t *testing.T) {
	t.Parallel()
	viewer := auditRole("viewer")
	reviewer := auditRole("reviewer")
	tests := []struct {
		name       string
		kind       auditEventKind
		roles      auditRoleFacts
		wantBefore any
		wantAfter  any
	}{
		{name: "create", kind: auditMembershipCreated, roles: auditRoleFacts{after: viewer}, wantAfter: "viewer"},
		{name: "update", kind: auditMembershipUpdated, roles: auditRoleFacts{before: viewer, after: reviewer}, wantBefore: "viewer", wantAfter: "reviewer"},
		{name: "delete", kind: auditMembershipDeleted, roles: auditRoleFacts{before: reviewer}, wantBefore: "reviewer"},
	}
	for _, test := range tests {
		changes, err := membershipAuditChanges(test.kind, test.roles)
		if err != nil {
			t.Fatalf("%s role facts: %v", test.name, err)
		}
		if len(changes) != 1 || changes[0].FieldPath != "role" || changes[0].ValueState != administrativeaudit.ValueVisible ||
			changes[0].Before != test.wantBefore || changes[0].After != test.wantAfter {
			t.Fatalf("%s role changes = %#v", test.name, changes)
		}
	}

	invalid := auditRole("owner")
	invalidFacts := []struct {
		kind  auditEventKind
		roles auditRoleFacts
	}{
		{kind: auditMembershipCreated},
		{kind: auditMembershipCreated, roles: auditRoleFacts{before: viewer, after: reviewer}},
		{kind: auditMembershipUpdated, roles: auditRoleFacts{before: viewer}},
		{kind: auditMembershipDeleted, roles: auditRoleFacts{before: reviewer, after: viewer}},
		{kind: auditMembershipDeleted, roles: auditRoleFacts{before: invalid}},
		{kind: auditIncidentCreated, roles: auditRoleFacts{before: viewer}},
	}
	for _, test := range invalidFacts {
		if _, err := membershipAuditChanges(test.kind, test.roles); err == nil {
			t.Fatalf("invalid role facts admitted: kind=%d roles=%#v", test.kind, test.roles)
		}
	}
}

func TestTerminalMutationCommitConstructorsAndQueriesFailClosed_Unit(t *testing.T) {
	t.Parallel()
	if commit, err := NewTerminalMutationCommit(uuid.Nil); err == nil || commit.IsNewCommit() || commit.IsReplay() {
		t.Fatalf("zero effect key constructed a terminal commit: commit=%#v err=%v", commit, err)
	}

	effectKey := uuid.New()
	commit, err := NewTerminalMutationCommit(effectKey)
	if err != nil {
		t.Fatalf("construct new terminal commit: %v", err)
	}
	gotEffectKey, present := commit.EffectKey()
	if !commit.IsNewCommit() || commit.IsReplay() || !present || gotEffectKey != effectKey {
		t.Fatalf("new terminal commit queries = commit=%#v key=%s present=%t", commit, gotEffectKey, present)
	}

	replay := ReplayTerminalMutationCommit()
	if replay.IsNewCommit() || !replay.IsReplay() {
		t.Fatalf("replay terminal commit queries = %#v", replay)
	}
	if key, present := replay.EffectKey(); present || key != uuid.Nil {
		t.Fatalf("replay exposed effect key %s present=%t", key, present)
	}

	var zero TerminalMutationCommit
	if zero.IsNewCommit() || zero.IsReplay() {
		t.Fatalf("zero terminal commit was classified: %#v", zero)
	}
	if key, present := zero.EffectKey(); present || key != uuid.Nil {
		t.Fatalf("zero terminal commit exposed effect key %s present=%t", key, present)
	}
}
