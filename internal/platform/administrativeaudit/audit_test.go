package administrativeaudit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contractaudit"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func TestAdministrativeAuditHTTPListContractIsScopeCorrectAndNormalized(t *testing.T) {
	actorID := uuid.MustParse("AAAAAAAA-BBBB-4CCC-8DDD-EEEEEEEEEEEE")
	result, queryErr := ParseListScope(
		"actor_user_id="+actorID.String()+"&action_code=membership_created&target_kind=incident_membership&target_id=target-1&occurred_at_gte=2035-07-25T15:30:00-04:00&occurred_at_lt=2035-07-25T19:31:00Z",
		ScopeIncident,
	)
	if queryErr != nil {
		t.Fatalf("parse valid incident administrative-audit scope: %#v", queryErr)
	}
	wantScope := map[string]string{
		"action_code":     "membership_created",
		"actor_user_id":   "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee",
		"occurred_at_gte": "2035-07-25T19:30:00Z",
		"occurred_at_lt":  "2035-07-25T19:31:00Z",
		"target_id":       "target-1",
		"target_kind":     "incident_membership",
	}
	if len(result.Scope) != len(wantScope) {
		t.Fatalf("administrative-audit scope has unexpected members: %#v", result.Scope)
	}
	for key, want := range wantScope {
		if got := fmt.Sprint(result.Scope[key]); got != want {
			t.Fatalf("administrative-audit scope %s = %q, want %q (scope %#v)", key, got, want, result.Scope)
		}
	}
	incidentID := uuid.MustParse("11111111-2222-4333-8444-555555555555")
	filter, reasonCode := PageRequest(pagination.Binding{
		Route:       "incident.membership_audit_events.list",
		ActorUserID: actorID.String(),
		Limit:       100,
		Scope:       result.Scope,
	}, nil, ScopeIncident, &incidentID)
	if reasonCode != "" {
		t.Fatalf("build incident administrative-audit page request: %s", reasonCode)
	}
	if filter.ScopeKind != ScopeIncident || filter.ScopeID == nil || *filter.ScopeID != incidentID ||
		filter.Limit != 101 || !reflect.DeepEqual(filter.AllowedActionCodes, ActionCodes(ScopeIncident)) {
		t.Fatalf("incident administrative-audit filter is not scope constrained: %#v", filter)
	}

	_, queryErr = ParseListScope("action_code=user_created", ScopeIncident)
	if queryErr == nil || queryErr.Kind != listquery.ErrorKindList || queryErr.ReasonCode != listquery.ReasonInvalidFilterValue {
		t.Fatalf("deployment action must be invalid on incident route: %#v", queryErr)
	}
	_, queryErr = ParseListScope("target_id=target-1", ScopeIncident)
	if queryErr == nil || queryErr.ReasonCode != listquery.ReasonInvalidFilterValue {
		t.Fatalf("target_id without target_kind must be invalid: %#v", queryErr)
	}
}

func TestAdministrativeAuditContractMatchesImplementation(t *testing.T) {
	var contract struct {
		ScopeKinds                  []string `json:"scope_kinds"`
		ActorKinds                  []string `json:"actor_kinds"`
		Sources                     []string `json:"sources"`
		ValueStates                 []string `json:"value_states"`
		ForbiddenVisibleFieldTokens []string `json:"forbidden_visible_field_tokens"`
		Actions                     []struct {
			ActionCode string `json:"action_code"`
			ScopeKind  string `json:"scope_kind"`
			TargetKind string `json:"target_kind"`
		} `json:"actions"`
	}
	found := false
	for _, artifact := range gencontracts.Artifacts {
		if artifact.Path != "contracts/audit/index.json" {
			continue
		}
		if err := json.Unmarshal([]byte(artifact.JSON), &contract); err != nil {
			t.Fatalf("decode generated audit contract: %v", err)
		}
		found = true
	}
	if !found {
		t.Fatal("generated audit contract artifact is missing")
	}

	assertSameStrings(t, contract.ScopeKinds, []string{ScopeDeployment, ScopeIncident}, "scope kinds")
	assertSameStrings(t, contract.ActorKinds, []string{ActorOperator, ActorSystem, ActorUser}, "actor kinds")
	assertSameStrings(t, contract.Sources, []string{SourceAPI, SourceOperator, SourceStartup, SourceSystem, SourceUI}, "sources")
	assertSameStrings(t, contract.ValueStates, []string{ValueRedacted, ValueVisible}, "value states")
	assertSameStrings(t, contract.ForbiddenVisibleFieldTokens, forbiddenVisibleFieldTokens, "forbidden visible field tokens")

	gotBindings := make([]string, 0, len(contract.Actions))
	for _, action := range contract.Actions {
		gotBindings = append(gotBindings, action.ActionCode+"|"+action.ScopeKind+"|"+action.TargetKind)
	}
	wantBindings := make([]string, 0, len(contract.Actions))
	for actionCode, bindings := range actionBindings {
		for _, binding := range bindings {
			wantBindings = append(wantBindings, actionCode+"|"+binding.scopeKind+"|"+binding.targetKind)
		}
	}
	assertSameStrings(t, gotBindings, wantBindings, "action bindings")
}

func TestAdministrativeAuditValidationSortsAndRejectsUnsafeChanges(t *testing.T) {
	targetID := uuid.NewString()
	now := time.Date(2026, time.July, 25, 19, 30, 0, 0, time.UTC)
	event := Event{
		ScopeKind:  ScopeDeployment,
		OccurredAt: now,
		ActorKind:  ActorSystem,
		Source:     SourceSystem,
		ActionCode: ActionUserCreated,
		TargetKind: TargetUser,
		TargetID:   &targetID,
		Changes: []Change{
			Visible("mfa_required", nil, true),
			Visible("display_name", nil, "Analyst"),
		},
	}
	changes, err := validateAndNormalizeEvent(event)
	if err != nil {
		t.Fatalf("validate safe event: %v", err)
	}
	if got := []string{changes[0].FieldPath, changes[1].FieldPath}; !reflect.DeepEqual(got, []string{"display_name", "mfa_required"}) {
		t.Fatalf("changes were not sorted: %v", got)
	}

	// A redacted secret-bearing field is the legal way to record a secret.
	redacted := event
	redacted.Changes = []Change{Redacted("password")}
	if _, err := validateAndNormalizeEvent(redacted); err != nil {
		t.Fatalf("redacted secret field must be accepted: %v", err)
	}

	cases := []struct {
		name    string
		changes []Change
		detail  string
	}{
		{name: "duplicate", changes: []Change{Visible("role", nil, "admin"), Visible("role", "admin", "viewer")}, detail: `duplicate field_path "role"`},
		{name: "secret field", changes: []Change{Visible("password_hash", nil, "hash")}, detail: `visible field "password_hash" is secret-bearing`},
		{name: "secret nested key", changes: []Change{Visible("configuration", nil, map[string]any{"client_secret": "secret"})}, detail: `visible field "configuration" is secret-bearing`},
		{name: "typed secret map", changes: []Change{Visible("configuration", nil, map[string]string{"session_token": "secret"})}, detail: `visible field "configuration" is secret-bearing`},
		// The redacted cases use a non-secret name, so only the non-null value
		// is wrong.
		{name: "redacted after value", changes: []Change{{FieldPath: "display_name", ValueState: ValueRedacted, After: "Analyst"}}, detail: `redacted field "display_name" must contain null values`},
		{name: "redacted before value", changes: []Change{{FieldPath: "display_name", ValueState: ValueRedacted, Before: "Analyst"}}, detail: `redacted field "display_name" must contain null values`},
		{name: "blank field path", changes: []Change{Visible(" ", nil, "Analyst")}, detail: "field_path is required"},
		{name: "unknown value state", changes: []Change{{FieldPath: "display_name", ValueState: "hidden"}}, detail: `unknown value_state "hidden"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			invalid := event
			invalid.Changes = tc.changes
			_, err := validateAndNormalizeEvent(invalid)
			if !errors.Is(err, ErrUnsafeChanges) || !strings.Contains(err.Error(), tc.detail) {
				t.Fatalf("expected ErrUnsafeChanges with %q, got %v", tc.detail, err)
			}
		})
	}
}

func TestAdministrativeAuditRawProjectionPairMustMatch(t *testing.T) {
	actorID := uuid.New()
	incidentID := uuid.New()
	now := time.Now().UTC()
	event := Event{
		ScopeKind:   ScopeIncident,
		ScopeID:     &incidentID,
		OccurredAt:  now,
		ActorKind:   ActorUser,
		ActorUserID: &actorID,
	}
	raw := RawEvent{ActorUserID: &actorID, IncidentID: &incidentID, OccurredAt: now}
	if err := validateRawProjectionPair(raw, event); err != nil {
		t.Fatalf("expected matching pair: %v", err)
	}
	deployment := event
	deployment.ScopeKind = ScopeDeployment
	deployment.ScopeID = nil
	deploymentRaw := raw
	deploymentRaw.IncidentID = nil
	if err := validateRawProjectionPair(deploymentRaw, deployment); err != nil {
		t.Fatalf("expected matching deployment pair: %v", err)
	}

	otherActorID := uuid.New()
	otherIncidentID := uuid.New()
	cases := []struct {
		name   string
		raw    RawEvent
		event  Event
		detail string
	}{
		{name: "zero occurrence", raw: RawEvent{ActorUserID: &actorID, IncidentID: &incidentID}, event: event, detail: "occurrence times must match"},
		{name: "different occurrence", raw: RawEvent{ActorUserID: &actorID, IncidentID: &incidentID, OccurredAt: now.Add(time.Second)}, event: event, detail: "occurrence times must match"},
		{name: "different actor", raw: RawEvent{ActorUserID: &otherActorID, IncidentID: &incidentID, OccurredAt: now}, event: event, detail: "actor_user_id must match"},
		{name: "missing raw actor", raw: RawEvent{IncidentID: &incidentID, OccurredAt: now}, event: event, detail: "actor_user_id must match"},
		{name: "deployment with incident raw", raw: raw, event: deployment, detail: "deployment projection cannot pair with incident raw event"},
		{name: "different incident", raw: RawEvent{ActorUserID: &actorID, IncidentID: &otherIncidentID, OccurredAt: now}, event: event, detail: "raw incident_id must match projected scope_id"},
		{name: "missing raw incident", raw: RawEvent{ActorUserID: &actorID, OccurredAt: now}, event: event, detail: "raw incident_id must match projected scope_id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRawProjectionPair(tc.raw, tc.event)
			if !errors.Is(err, ErrInvalidEvent) || !strings.Contains(err.Error(), tc.detail) {
				t.Fatalf("expected ErrInvalidEvent with %q, got %v", tc.detail, err)
			}
		})
	}
}

func TestProductionAuditWritesUseAdministrativeAuditBoundary(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	internalRoot := filepath.Join(root, "internal")
	writes := []string{"INSERT INTO deployment_admin_audit_events", "INSERT INTO administrative_audit_projections"}
	owner, err := os.ReadFile(filepath.Join(root, "internal", "platform", "administrativeaudit", "audit.go"))
	if err != nil {
		t.Fatalf("read audit owner: %v", err)
	}
	for _, write := range writes {
		if !strings.Contains(string(owner), write) {
			t.Fatalf("audit owner no longer contains %q; the boundary scan would be vacuous", write)
		}
	}
	var violations []string
	if err := filepath.WalkDir(internalRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if relative == "internal/gen" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") || relative == "internal/platform/administrativeaudit/audit.go" {
			return nil
		}
		payload, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, write := range writes {
			if strings.Contains(string(payload), write) {
				violations = append(violations, filepath.ToSlash(relative)+": "+write)
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("scan production audit writes: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("production audit writes bypass administrativeaudit: %v", violations)
	}
}

func assertSameStrings(t testing.TB, got []string, want []string, label string) {
	t.Helper()
	got = append([]string(nil), got...)
	want = append([]string(nil), want...)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s mismatch:\ngot  %v\nwant %v", label, got, want)
	}
}

func TestAdministrativeAuditRecoveryContribution(t *testing.T) {
	contribution := RecoveryStateContribution()
	for _, table := range contribution.Tables {
		if table.TableName != "administrative_audit_projections" {
			continue
		}
		if contribution.OwnerID != "module.audit" || table.StateClass != recoverystate.StateAuthoritative ||
			table.BackupInclusion != recoverystate.InclusionRequired || table.RestoreAction != recoverystate.RestoreState ||
			table.CodecID == nil || *table.CodecID != recoverystate.PostgresUnitCodecID {
			t.Fatal("administrative audit projections must be authoritative, required and restored with the current PostgreSQL unit codec")
		}
		return
	}
	t.Fatal("administrative audit projections are absent from the owner's Recovery contribution")
}
