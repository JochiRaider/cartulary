package testsupport

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// MustBuild delegates to the sole production Projections assembly path.
func MustBuild(t testing.TB, db postgres.DB) *projectionassembly.Runtime {
	t.Helper()
	runtime, err := projectionassembly.Build(db)
	if err != nil {
		t.Fatalf("compose production Projections runtime for test: %v", err)
	}
	return runtime
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
