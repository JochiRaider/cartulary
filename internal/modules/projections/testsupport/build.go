package testsupport

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelinefactassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entityprovider "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/projectionprovider"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	evidenceprovider "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionprovider"
	indicatorprovider "github.com/JochiRaider/cartulary/internal/modules/indicators/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	taskdecisionprovider "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
	timelineprovider "github.com/JochiRaider/cartulary/internal/modules/timeline/projectionprovider"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// MustBuild constructs the production-shaped Projections ports for tests. It
// requires every canonical source-owner contribution and exposes no private
// runtime, catalog, or storage values.
func MustBuild(t testing.TB, db postgres.DB) projectionadapters.Ports {
	t.Helper()

	timelineContribution, err := timelineprojection.NewContribution(timelineprovider.NewSource(timelinefactassembly.NewLinkReader()))
	if err != nil {
		t.Fatalf("compose Timeline projection contribution: %v", err)
	}
	entitiesContribution, err := entityprojection.NewContribution(entityprovider.NewSource())
	if err != nil {
		t.Fatalf("compose Entities projection contribution: %v", err)
	}
	indicatorsContribution, err := indicatorprovider.NewContribution()
	if err != nil {
		t.Fatalf("compose Indicators projection contribution: %v", err)
	}
	assessmentsContribution, err := assessmentassembly.NewProjectionContribution()
	if err != nil {
		t.Fatalf("compose Assessments projection contribution: %v", err)
	}
	artifactsContribution, err := artifacts.NewProjectionContribution()
	if err != nil {
		t.Fatalf("compose Artifacts projection contribution: %v", err)
	}
	evidenceContribution, err := evidenceprovider.NewContribution()
	if err != nil {
		t.Fatalf("compose Evidence projection contribution: %v", err)
	}
	partiesContribution, err := parties.NewProjectionContribution()
	if err != nil {
		t.Fatalf("compose Parties projection contribution: %v", err)
	}
	taskDecisionContribution, err := taskdecisionprovider.NewContribution()
	if err != nil {
		t.Fatalf("compose Tasks/Decisions projection contribution: %v", err)
	}

	ports, err := projectionadapters.New(projectionadapters.Dependencies{
		Postgres:       db,
		Timeline:       timelineContribution,
		Entities:       entitiesContribution,
		Indicators:     indicatorsContribution,
		Assessments:    assessmentsContribution,
		Artifacts:      artifactsContribution,
		Evidence:       evidenceContribution,
		Parties:        partiesContribution,
		TasksDecisions: taskDecisionContribution,
	})
	if err != nil {
		t.Fatalf("compose Projections test ports: %v", err)
	}
	return ports
}

// MustAssessmentSource exposes the typed canonical source for focused source
// enumeration tests without importing application assembly from private
// Projections tests.
func MustAssessmentSource(t testing.TB) assessmentprojection.SourceReader {
	t.Helper()
	contribution, err := assessmentassembly.NewProjectionContribution()
	if err != nil {
		t.Fatalf("compose Assessments projection contribution: %v", err)
	}
	return contribution.Source()
}
