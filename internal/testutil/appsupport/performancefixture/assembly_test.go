package performancefixture

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	authfixture "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/performancefixture"
	entitiesfixture "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport/performancefixture"
	linksfixture "github.com/JochiRaider/cartulary/internal/modules/links/testsupport/performancefixture"
	projectionsfixture "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport/performancefixture"
	timelinefixture "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/performancefixture"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

func TestClosedAssemblerProducesExactDeterministicRedactedSemantics(t *testing.T) {
	t.Parallel()
	assemble := func(marker string) (fixture.Result, *fixture.BuildState, *fakeApplications) {
		t.Helper()
		profile := profileForTest(t)
		applications := &fakeApplications{}
		assembler, err := New(profile, Dependencies{
			Auth: applications, Incidents: applications, Entities: applications,
			Timeline: applications, Links: applications, Projections: applications,
			Validation: applications,
		})
		if err != nil {
			t.Fatal(err)
		}
		state := &fixture.BuildState{
			FixtureProfileID: profile.FixtureProfileID,
			SnapshotKey:      fixtureKey,
			Seed:             profile.Seed,
			RuntimeBundle:    runtimeBundle(marker, profile),
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
	wantCounts := semanticCountsForTest(profileForTest(t))
	if !reflect.DeepEqual(first.Validation.Counts, wantCounts) {
		t.Fatalf("semantic counts=%#v want=%#v", first.Validation.Counts, wantCounts)
	}
	if got := len(first.Receipts); got != 6 {
		t.Fatalf("receipt count=%d want=6", got)
	}
	if got := len(first.ContributionDiagnostics); got != 6 {
		t.Fatalf("contribution diagnostics=%d want=6", got)
	}
	timelineDiagnostic := first.ContributionDiagnostics[3]
	if timelineDiagnostic.ContributionID != "timeline.large_grid.v1" ||
		timelineDiagnostic.OwnerID != "module.timeline" || timelineDiagnostic.Batch == nil ||
		timelineDiagnostic.Batch.Strategy != "owner_set_oriented" ||
		timelineDiagnostic.Batch.BatchCount != 1 ||
		timelineDiagnostic.Batch.ConfiguredBatchSize != 20000 ||
		timelineDiagnostic.Batch.ItemCount != 20000 {
		t.Fatalf("unexpected Timeline build diagnostics: %#v", timelineDiagnostic)
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
	profile := profileForTest(t)
	applications := &fakeApplications{failureAt: "timeline"}
	assembler, err := New(profile, Dependencies{
		Auth: applications, Incidents: applications, Entities: applications,
		Timeline: applications, Links: applications, Projections: applications,
		Validation: applications,
	})
	if err != nil {
		t.Fatal(err)
	}
	state := &fixture.BuildState{FixtureProfileID: profile.FixtureProfileID, SnapshotKey: fixtureKey, Seed: profile.Seed, RuntimeBundle: runtimeBundle("failure", profile)}
	if _, err := assembler.Assemble(context.Background(), state); err == nil {
		t.Fatal("expected injected owner failure")
	}
	if applications.links != 0 {
		t.Fatal("later contribution ran after owner failure")
	}
}

func TestClosedAssemblerConsumesGeneratedDescriptor(t *testing.T) {
	t.Parallel()
	profile := profileForTest(t)
	applications := &fakeApplications{}
	if _, err := New(profile, Dependencies{
		Auth: applications, Incidents: applications, Entities: applications,
		Timeline: applications, Links: applications, Projections: applications,
		Validation: applications,
	}); err != nil {
		t.Fatalf("generated descriptor did not admit the owner-local providers: %v", err)
	}
}

const fixtureKey = "85a9ceb4cc34f66356baa07b68bf7f3636844beef90aa51ad8b1751d4b046c72"

func runtimeBundle(marker string, profile performancefixtureprofile.Profile) fixture.RuntimeBundle {
	credentials := make([]fixture.RuntimeCredential, profile.RuntimeCredentialSets[0].AccountCount)
	for index := range credentials {
		credentials[index] = fixture.RuntimeCredential{
			Principal: fmt.Sprintf("%s-%02d@example.test", marker, index),
			Secret:    fmt.Sprintf("Ac043-%s-%02d-password-material", marker, index),
		}
	}
	return fixture.RuntimeBundle{
		SchemaID:         profile.ArtifactPolicy.RuntimeSchemaID,
		FixtureProfileID: profile.FixtureProfileID,
		SnapshotKey:      fixtureKey,
		CredentialSets: []fixture.RuntimeCredentialSet{{
			SetID:          profile.RuntimeCredentialSets[0].SetID,
			CredentialKind: profile.RuntimeCredentialSets[0].CredentialKind,
			Credentials:    credentials,
		}},
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
	if incidentID == "" || len(rows) != 20000 {
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

func (f *fakeApplications) ValidateFixtureProjectionSets(_ context.Context, incidentID string, expectations []projectionsfixture.SetExpectation) error {
	if f.failureAt == "projections" {
		return errors.New("injected Projections failure")
	}
	if incidentID == "" || len(expectations) != 3 {
		return errors.New("fixture projection validation failed")
	}
	for _, expectation := range expectations {
		if expectation.ViewSchemaID == "" || expectation.ExactRows < 1 {
			return errors.New("fixture projection expectation is incomplete")
		}
	}
	f.projectionSets = len(expectations)
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
		Counts: map[string]int{
			"active_sessions":     f.sessions,
			"background_analysts": f.accounts,
			"host_rows":           f.hosts,
			"identity_rows":       f.identities,
			"link_assignments":    f.links,
			"mention_assignments": f.mentions,
			"tag_assignments":     f.tags,
			"timeline_rows":       f.timelineRows,
		},
		Conditions: map[string]bool{
			"authorization":             expected.Authorization,
			"default_view":              expected.DefaultView,
			"no_active_sessions":        expected.NoActiveSessions,
			"projection_ready":          expected.ProjectionReady,
			"relationship_distribution": expected.RelationshipDistribution,
		},
	}, nil
}

func profileForTest(t *testing.T) performancefixtureprofile.Profile {
	t.Helper()
	for _, profile := range performancefixtureprofile.Profiles() {
		if profile.Status == "active" {
			return profile
		}
	}
	t.Fatal("generated active fixture profile is missing")
	return performancefixtureprofile.Profile{}
}

func semanticCountsForTest(profile performancefixtureprofile.Profile) map[string]int {
	result := make(map[string]int, len(profile.SemanticExpectations.Counts))
	for _, expectation := range profile.SemanticExpectations.Counts {
		result[expectation.ExpectationID] = expectation.Exact
	}
	return result
}
