package administrativeaudit

import (
	"encoding/json"
	"errors"
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

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
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
	if result.Scope["actor_user_id"] != actorID.String() ||
		result.Scope["occurred_at_gte"] != "2035-07-25T19:30:00Z" ||
		result.Scope["occurred_at_lt"] != "2035-07-25T19:31:00Z" {
		t.Fatalf("administrative-audit scope is not canonically normalized: %#v", result.Scope)
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
	for _, artifact := range gencontracts.AuditArtifacts {
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

	cases := []struct {
		name    string
		changes []Change
	}{
		{name: "duplicate", changes: []Change{Visible("role", nil, "admin"), Visible("role", "admin", "viewer")}},
		{name: "secret field", changes: []Change{Visible("password_hash", nil, "hash")}},
		{name: "secret nested key", changes: []Change{Visible("configuration", nil, map[string]any{"client_secret": "secret"})}},
		{name: "typed secret map", changes: []Change{Visible("configuration", nil, map[string]string{"session_token": "secret"})}},
		{name: "redacted value", changes: []Change{{FieldPath: "password", ValueState: ValueRedacted, After: "secret"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			invalid := event
			invalid.Changes = tc.changes
			if _, err := validateAndNormalizeEvent(invalid); !errors.Is(err, ErrUnsafeChanges) {
				t.Fatalf("expected ErrUnsafeChanges, got %v", err)
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
	otherIncidentID := uuid.New()
	raw.IncidentID = &otherIncidentID
	if err := validateRawProjectionPair(raw, event); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected mismatched scope rejection, got %v", err)
	}
}

func TestProductionAuditWritesUseAdministrativeAuditBoundary(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	internalRoot := filepath.Join(root, "internal")
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
		if strings.Contains(string(payload), "INSERT INTO deployment_admin_audit_events") {
			violations = append(violations, filepath.ToSlash(relative))
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
