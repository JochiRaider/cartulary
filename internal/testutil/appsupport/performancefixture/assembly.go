package performancefixture

import (
	"context"
	"errors"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	authfixture "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/performancefixture"
	entitiesfixture "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport/performancefixture"
	incidentsfixture "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/performancefixture"
	linksfixture "github.com/JochiRaider/cartulary/internal/modules/links/testsupport/performancefixture"
	projectionsfixture "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport/performancefixture"
	timelinefixture "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/performancefixture"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

type Dependencies struct {
	Auth        authfixture.Application
	Incidents   incidentsfixture.Application
	Entities    entitiesfixture.Application
	Timeline    timelinefixture.Application
	Links       linksfixture.Application
	Projections projectionsfixture.Application
	Validation  SemanticValidationApplication
}

// SemanticExpectations is the AC-043 adapter projection consumed by its
// production semantic query. Generic fixture mechanics retain the structural
// generated maps and do not know these expectation identities.
type SemanticExpectations struct {
	TimelineRows             int
	HostRows                 int
	IdentityRows             int
	TagAssignments           int
	MentionAssignments       int
	LinkAssignments          int
	BackgroundAnalysts       int
	ActiveSessions           int
	DefaultView              bool
	ProjectionReady          bool
	RelationshipDistribution bool
	Authorization            bool
	NoActiveSessions         bool
}

type SemanticValidationApplication interface {
	ValidateFixtureSemantics(context.Context, string, SemanticExpectations) (fixture.SemanticValidation, error)
}

const ac043BackgroundCredentialSetID = "background_analysts"

func New(profile performancefixtureprofile.Profile, dependencies Dependencies) (*fixture.Assembler, error) {
	descriptors, err := fixture.Descriptors(profile)
	if err != nil {
		return nil, err
	}
	descriptor := func(ownerID string) (fixture.Descriptor, error) {
		for _, candidate := range descriptors {
			if candidate.OwnerID == ownerID {
				return candidate, nil
			}
		}
		return fixture.Descriptor{}, fmt.Errorf("AC-043 generated profile is missing owner contribution %s", ownerID)
	}
	authDescriptor, err := descriptor("module.auth")
	if err != nil {
		return nil, err
	}
	auth, err := authfixture.New(dependencies.Auth, authDescriptor, ac043BackgroundCredentialSetID)
	if err != nil {
		return nil, err
	}
	incidentsDescriptor, err := descriptor("module.incidents")
	if err != nil {
		return nil, err
	}
	incidents, err := incidentsfixture.New(dependencies.Incidents, incidentsDescriptor)
	if err != nil {
		return nil, err
	}
	entitiesDescriptor, err := descriptor("module.entities")
	if err != nil {
		return nil, err
	}
	entities, err := entitiesfixture.New(dependencies.Entities, entitiesDescriptor)
	if err != nil {
		return nil, err
	}
	timelineDescriptor, err := descriptor("module.timeline")
	if err != nil {
		return nil, err
	}
	timeline, err := timelinefixture.New(dependencies.Timeline, timelineDescriptor)
	if err != nil {
		return nil, err
	}
	expectations, err := semanticExpectations(profile)
	if err != nil {
		return nil, err
	}
	if expectations.LinkAssignments < 1 || expectations.TimelineRows%expectations.LinkAssignments != 0 {
		return nil, errors.New("AC-043 generated relationship distribution is incompatible")
	}
	linksDescriptor, err := descriptor("module.links")
	if err != nil {
		return nil, err
	}
	links, err := linksfixture.New(dependencies.Links, linksDescriptor, expectations.TimelineRows/expectations.LinkAssignments)
	if err != nil {
		return nil, err
	}
	projectionRows := map[string]int{
		"cartulary.view.hosts.v1":      expectations.HostRows,
		"cartulary.view.identities.v1": expectations.IdentityRows,
		"cartulary.view.timeline.v2":   expectations.TimelineRows,
	}
	projectionExpectations := make([]projectionsfixture.SetExpectation, 0)
	for _, ref := range profile.SourceContractRefs {
		if ref.SchemaID == "cartulary.view_schema_source.v1" {
			exactRows, ok := projectionRows[ref.ContractID]
			if !ok {
				return nil, fmt.Errorf("AC-043 generated profile has no projection count for %s", ref.ContractID)
			}
			projectionExpectations = append(projectionExpectations, projectionsfixture.SetExpectation{
				ViewSchemaID: ref.ContractID,
				ExactRows:    exactRows,
			})
		}
	}
	projectionsDescriptor, err := descriptor("module.projections")
	if err != nil {
		return nil, err
	}
	projections, err := projectionsfixture.New(dependencies.Projections, projectionsDescriptor, projectionExpectations)
	if err != nil {
		return nil, err
	}
	if dependencies.Validation == nil {
		return nil, errors.New("performance fixture semantic validation application is required")
	}
	return fixture.NewAssembler(
		profile,
		semanticValidator{application: dependencies.Validation, expectations: expectations},
		auth,
		incidents,
		entities,
		timeline,
		links,
		projections,
	)
}

func semanticExpectations(profile performancefixtureprofile.Profile) (SemanticExpectations, error) {
	counts := make(map[string]int, len(profile.SemanticExpectations.Counts))
	for _, expectation := range profile.SemanticExpectations.Counts {
		if _, duplicate := counts[expectation.ExpectationID]; duplicate {
			return SemanticExpectations{}, fmt.Errorf("duplicate AC-043 semantic count %s", expectation.ExpectationID)
		}
		counts[expectation.ExpectationID] = expectation.Exact
	}
	conditions := make(map[string]bool, len(profile.SemanticExpectations.Conditions))
	for _, expectation := range profile.SemanticExpectations.Conditions {
		if _, duplicate := conditions[expectation.ExpectationID]; duplicate {
			return SemanticExpectations{}, fmt.Errorf("duplicate AC-043 semantic condition %s", expectation.ExpectationID)
		}
		conditions[expectation.ExpectationID] = expectation.Required
	}
	requiredCounts := []string{
		"active_sessions", "background_analysts", "host_rows", "identity_rows",
		"link_assignments", "mention_assignments", "tag_assignments", "timeline_rows",
	}
	for _, key := range requiredCounts {
		if _, ok := counts[key]; !ok {
			return SemanticExpectations{}, fmt.Errorf("AC-043 generated profile is missing semantic count %s", key)
		}
	}
	requiredConditions := []string{
		"authorization", "default_view", "no_active_sessions", "projection_ready", "relationship_distribution",
	}
	for _, key := range requiredConditions {
		if _, ok := conditions[key]; !ok {
			return SemanticExpectations{}, fmt.Errorf("AC-043 generated profile is missing semantic condition %s", key)
		}
	}
	return SemanticExpectations{
		TimelineRows:             counts["timeline_rows"],
		HostRows:                 counts["host_rows"],
		IdentityRows:             counts["identity_rows"],
		TagAssignments:           counts["tag_assignments"],
		MentionAssignments:       counts["mention_assignments"],
		LinkAssignments:          counts["link_assignments"],
		BackgroundAnalysts:       counts["background_analysts"],
		ActiveSessions:           counts["active_sessions"],
		DefaultView:              conditions["default_view"],
		ProjectionReady:          conditions["projection_ready"],
		RelationshipDistribution: conditions["relationship_distribution"],
		Authorization:            conditions["authorization"],
		NoActiveSessions:         conditions["no_active_sessions"],
	}, nil
}

type semanticValidator struct {
	application  SemanticValidationApplication
	expectations SemanticExpectations
}

func (v semanticValidator) Validate(ctx context.Context, state *fixture.BuildState) (fixture.SemanticValidation, error) {
	if state.IncidentID == "" {
		return fixture.SemanticValidation{}, errors.New("performance fixture semantic validation requires an incident identity")
	}
	return v.application.ValidateFixtureSemantics(ctx, state.IncidentID, v.expectations)
}
