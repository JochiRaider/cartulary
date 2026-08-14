package performancefixtureassembly

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	authfixture "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/performancefixture"
	entitiesfixture "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport/performancefixture"
	linksfixture "github.com/JochiRaider/cartulary/internal/modules/links/testsupport/performancefixture"
	timelinefixture "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/performancefixture"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

func TestClosedAssemblerProducesExactDeterministicRedactedSemantics(t *testing.T) {
	t.Parallel()
	assemble := func(marker string) (fixture.Result, *fixture.BuildState, *fakeApplications) {
		t.Helper()
		applications := &fakeApplications{}
		assembler, err := New(Dependencies{
			Auth: applications, Incidents: applications, Entities: applications,
			Timeline: applications, Links: applications, Projections: applications,
			Validation: applications,
		})
		if err != nil {
			t.Fatal(err)
		}
		state := &fixture.BuildState{
			SnapshotKey:   fixtureKey,
			Seed:          20260405,
			RuntimeBundle: runtimeBundle(marker),
		}
		result, err := assembler.Assemble(context.Background(), state)
		if err != nil {
			t.Fatal(err)
		}
		return result, state, applications
	}
	first, firstState, firstApplications := assemble("first")
	second, _, _ := assemble("second")
	if first.SemanticValidationDigest != second.SemanticValidationDigest {
		t.Fatalf("suite-random credentials changed semantic digest: %s != %s", first.SemanticValidationDigest, second.SemanticValidationDigest)
	}
	if !reflect.DeepEqual(first.Validation.Counts, ExpectedSemanticCounts()) {
		t.Fatalf("semantic counts=%#v want=%#v", first.Validation.Counts, ExpectedSemanticCounts())
	}
	if got := len(first.Receipts); got != 6 {
		t.Fatalf("receipt count=%d want=6", got)
	}
	if firstApplications.timelineRows != 20000 || firstApplications.hosts != 1000 || firstApplications.identities != 1000 {
		t.Fatalf("unexpected fixture dimensions: %#v", firstApplications)
	}
	if firstApplications.links != 1000 || firstApplications.mentions != 1000 || firstApplications.tags != 1000 {
		t.Fatalf("unexpected association dimensions: %#v", firstApplications)
	}
	if firstApplications.sessions != 0 || firstApplications.memberships != 24 || firstApplications.accounts != 24 {
		t.Fatalf("unexpected Auth/Incidents state: %#v", firstApplications)
	}
	if err := fixture.ValidateReceiptRedaction(first, firstState); err != nil {
		t.Fatal(err)
	}
}

func TestClosedAssemblerPropagatesOwnerFailure(t *testing.T) {
	t.Parallel()
	applications := &fakeApplications{failureAt: "timeline"}
	assembler, err := New(Dependencies{
		Auth: applications, Incidents: applications, Entities: applications,
		Timeline: applications, Links: applications, Projections: applications,
		Validation: applications,
	})
	if err != nil {
		t.Fatal(err)
	}
	state := &fixture.BuildState{SnapshotKey: fixtureKey, Seed: 20260405, RuntimeBundle: runtimeBundle("failure")}
	if _, err := assembler.Assemble(context.Background(), state); err == nil {
		t.Fatal("expected injected owner failure")
	}
	if applications.links != 0 {
		t.Fatal("later contribution ran after owner failure")
	}
}

func TestClosedDescriptorsMatchAuthoredRegistry(t *testing.T) {
	t.Parallel()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source path")
	}
	registryPath := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../tools/performance_fixture_snapshot_owner.json"))
	raw, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	var registry struct {
		Profiles []struct {
			FixtureProfileID string `json:"fixture_profile_id"`
			Contributions    []struct {
				ContributionID  string         `json:"contribution_id"`
				OwnerID         string         `json:"owner_id"`
				Version         string         `json:"version"`
				Dependencies    []string       `json:"dependencies"`
				ExpectedReceipt map[string]int `json:"expected_receipt"`
			} `json:"contributions"`
			ValidationRules map[string]any `json:"validation_rules"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal(raw, &registry); err != nil {
		t.Fatal(err)
	}
	if len(registry.Profiles) != 1 || registry.Profiles[0].FixtureProfileID != fixture.LargeGridProfileID {
		t.Fatalf("unexpected authored fixture profiles: %#v", registry.Profiles)
	}
	got := make([]fixture.Descriptor, len(registry.Profiles[0].Contributions))
	for index, contribution := range registry.Profiles[0].Contributions {
		got[index] = fixture.Descriptor{
			ContributionID: contribution.ContributionID,
			Version:        contribution.Version,
			OwnerID:        contribution.OwnerID,
			Dependencies:   contribution.Dependencies,
			ExpectedCounts: contribution.ExpectedReceipt,
		}
	}
	if !reflect.DeepEqual(got, ExpectedDescriptors()) {
		t.Fatalf("authored registry contribution drift:\ngot  %#v\nwant %#v", got, ExpectedDescriptors())
	}
}

const fixtureKey = "85a9ceb4cc34f66356baa07b68bf7f3636844beef90aa51ad8b1751d4b046c72"

func runtimeBundle(marker string) fixture.RuntimeBundle {
	accounts := make([]fixture.BackgroundAccount, 24)
	for index := range accounts {
		accounts[index] = fixture.BackgroundAccount{
			Email:    fmt.Sprintf("%s-%02d@example.test", marker, index),
			Password: fmt.Sprintf("Ac043-%s-%02d-password-material", marker, index),
		}
	}
	return fixture.RuntimeBundle{
		SchemaID:           fixture.RuntimeSchemaID,
		FixtureProfileID:   fixture.LargeGridProfileID,
		SnapshotKey:        fixtureKey,
		BackgroundAccounts: accounts,
	}
}

type fakeApplications struct {
	failureAt      string
	accounts       int
	sessions       int
	memberships    int
	hosts          int
	identities     int
	timelineRows   int
	links          int
	mentions       int
	tags           int
	projectionSets int
}

func (f *fakeApplications) CreateBackgroundAnalyst(_ context.Context, request authfixture.CreateBackgroundAnalystRequest) (authfixture.Account, error) {
	if f.failureAt == "auth" {
		return authfixture.Account{}, errors.New("injected Auth failure")
	}
	if request.Email == "" || request.InitialPassword == "" || request.MFARequired || request.IsDeploymentAdmin || !request.Active {
		return authfixture.Account{}, errors.New("invalid Auth fixture request")
	}
	f.accounts++
	return authfixture.Account{UserID: fmt.Sprintf("private-user-%02d", f.accounts)}, nil
}

func (f *fakeApplications) CreateFixtureWorkspaceIncident(context.Context, int) (string, error) {
	if f.failureAt == "incidents" {
		return "", errors.New("injected Incidents failure")
	}
	return "private-incident-id", nil
}

func (f *fakeApplications) AddFixtureMembership(_ context.Context, incidentID string, userID string, role string) error {
	if incidentID == "" || userID == "" || role != "editor" {
		return errors.New("invalid Incidents fixture membership")
	}
	f.memberships++
	return nil
}

func (f *fakeApplications) CreateFixtureHosts(_ context.Context, incidentID string, hosts []entitiesfixture.Host) error {
	if f.failureAt == "entities" {
		return errors.New("injected Entities failure")
	}
	if incidentID == "" || len(hosts) != 500 {
		return errors.New("invalid Hosts fixture batch")
	}
	f.hosts += len(hosts)
	return nil
}

func (f *fakeApplications) CreateFixtureIdentities(_ context.Context, incidentID string, identities []entitiesfixture.Identity) error {
	if incidentID == "" || len(identities) != 500 {
		return errors.New("invalid Identities fixture batch")
	}
	f.identities += len(identities)
	return nil
}

func (f *fakeApplications) CreateFixtureTimelineRows(_ context.Context, incidentID string, rows []timelinefixture.Row) error {
	if f.failureAt == "timeline" {
		return errors.New("injected Timeline failure")
	}
	if incidentID == "" || len(rows) != 500 {
		return errors.New("invalid Timeline fixture batch")
	}
	for _, row := range rows {
		if row.Tag != "" {
			f.tags++
		}
		if row.IdentityRef != "" {
			f.mentions++
		}
		if row.DataSource != "" {
			f.links++
		}
	}
	f.timelineRows += len(rows)
	return nil
}

func (f *fakeApplications) ValidateFixtureAssociations(_ context.Context, incidentID string, expected linksfixture.Expectations) error {
	if f.failureAt == "links" {
		return errors.New("injected Links failure")
	}
	if incidentID == "" || expected.Links != f.links || expected.Mentions != f.mentions || expected.Tags != f.tags || expected.Stride != 20 {
		return errors.New("fixture association validation failed")
	}
	return nil
}

func (f *fakeApplications) ValidateFixtureProjectionSets(_ context.Context, incidentID string, viewSchemaIDs []string) error {
	if f.failureAt == "projections" {
		return errors.New("injected Projections failure")
	}
	if incidentID == "" || len(viewSchemaIDs) != 3 {
		return errors.New("fixture projection validation failed")
	}
	f.projectionSets = len(viewSchemaIDs)
	return nil
}

func (f *fakeApplications) ValidateFixtureSemantics(_ context.Context, incidentID string, expected SemanticExpectations) (fixture.SemanticValidation, error) {
	if f.failureAt == "validation" {
		return fixture.SemanticValidation{}, errors.New("injected semantic validation failure")
	}
	if incidentID == "" || f.accounts != expected.BackgroundAnalysts || f.memberships != expected.BackgroundAnalysts || f.timelineRows != expected.TimelineRows || f.hosts != expected.HostRows || f.identities != expected.IdentityRows || f.links != expected.LinkAssignments || f.mentions != expected.MentionAssignments || f.tags != expected.TagAssignments || f.sessions != expected.ActiveSessions || f.projectionSets != 3 {
		return fixture.SemanticValidation{}, errors.New("fixture semantic validation failed")
	}
	return fixture.SemanticValidation{
		Counts:                   ExpectedSemanticCounts(),
		RelationshipDistribution: true,
		DefaultView:              expected.DefaultView,
		Authorization:            true,
		ProjectionReadiness:      expected.ProjectionReady,
		NoActiveSessions:         expected.ActiveSessions == 0,
	}, nil
}
