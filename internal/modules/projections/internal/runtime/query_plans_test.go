package runtime

import (
	"slices"
	"testing"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectioncontract"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/internal/queryengine"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectioncontract"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func contractPlansForTest() map[string]queryengine.Surface {
	plans := make([]queryengine.Surface, 0)
	plans = append(plans, queryengine.TimelinePlans()...)
	plans = append(plans, queryengine.IndicatorPlans()...)
	plans = append(plans, queryengine.AssessmentPlans()...)
	plans = append(plans, queryengine.ArtifactPlans()...)
	plans = append(plans, queryengine.EvidencePlans()...)
	plans = append(plans, queryengine.PartyPlans()...)
	plans = append(plans, queryengine.TaskRequestPlans()...)
	plans = append(plans, queryengine.DecisionPlans()...)
	surfaces := make(map[string]queryengine.Surface, len(plans))
	for _, plan := range plans {
		surface, err := queryengine.CompileSurface(plan)
		if err != nil {
			panic(err)
		}
		if _, exists := surfaces[surface.ViewSchemaID]; exists {
			panic("duplicate query surface " + surface.ViewSchemaID)
		}
		surfaces[surface.ViewSchemaID] = surface
	}
	return surfaces
}

func TestPrivateCompiledPlansExactlyMatchSemanticIntentsAndViewSchemas(t *testing.T) {
	evidenceIntent, err := evidenceprojection.SurfaceIntent()
	if err != nil {
		t.Fatalf("Evidence semantic intent: %v", err)
	}
	partyIntent, err := partyprojection.SurfaceIntent()
	if err != nil {
		t.Fatalf("Party semantic intent: %v", err)
	}
	artifactContribution, err := artifactprojection.NewContribution(artifactProjectionIntentSourceStub{})
	if err != nil {
		t.Fatalf("construct Artifact projection contribution: %v", err)
	}
	artifactIntents := artifactContribution.ProjectionContribution().SurfaceIntents()
	indicatorContribution, err := indicatorprojection.NewContribution(indicatorProjectionIntentSourceStub{})
	if err != nil {
		t.Fatalf("construct Indicator projection contribution: %v", err)
	}
	indicatorIntents := indicatorContribution.ProjectionContribution().SurfaceIntents()
	assessmentContribution, err := assessmentprojection.NewContribution(assessmentProjectionIntentSourceStub{})
	if err != nil {
		t.Fatalf("construct Assessment projection contribution: %v", err)
	}
	assessmentIntents := assessmentContribution.ProjectionContribution().SurfaceIntents()
	taskDecisionContribution, err := taskdecisionprojection.NewContribution(
		&catalogTaskRequestSource{},
		&catalogDecisionSource{},
	)
	if err != nil {
		t.Fatalf("construct Tasks/Decisions projection contribution: %v", err)
	}
	taskDecisionIntents := taskDecisionContribution.ProjectionContribution().SurfaceIntents()
	intents := []providercontract.SurfaceIntent{
		timelineprojection.SurfaceIntent(),
		evidenceIntent,
		partyIntent,
	}
	intents = append(intents, assessmentIntents...)
	intents = append(intents, artifactIntents...)
	intents = append(intents, indicatorIntents...)
	intents = append(intents, taskDecisionIntents...)
	intentByView := make(map[string]providercontract.SurfaceIntent, len(intents))
	for _, intent := range intents {
		if _, exists := intentByView[intent.ViewSchemaID]; exists {
			t.Fatalf("duplicate semantic intent %s", intent.ViewSchemaID)
		}
		intentByView[intent.ViewSchemaID] = intent
	}

	plans := contractPlansForTest()
	if len(plans) != len(intentByView) {
		t.Fatalf("compiled plan count = %d, semantic intent count = %d", len(plans), len(intentByView))
	}
	for viewSchemaID, intent := range intentByView {
		plan, exists := plans[viewSchemaID]
		if !exists {
			t.Errorf("semantic intent %s has no private compiled plan", viewSchemaID)
			continue
		}
		planFields := make([]string, 0, len(plan.Fields))
		for _, field := range plan.Fields {
			planFields = append(planFields, field.Key)
		}
		slices.Sort(planFields)
		intentFields := slices.Clone(intent.FieldKeys)
		slices.Sort(intentFields)
		if !slices.Equal(planFields, intentFields) {
			t.Errorf("%s compiled fields = %v, semantic fields = %v", viewSchemaID, planFields, intentFields)
		}
		schema, ok := viewschema.Lookup(viewSchemaID)
		if !ok {
			t.Errorf("%s has no view schema", viewSchemaID)
			continue
		}
		schemaFields := make([]string, 0, len(schema.Fields()))
		for fieldKey := range schema.Fields() {
			schemaFields = append(schemaFields, fieldKey)
		}
		slices.Sort(schemaFields)
		if !slices.Equal(planFields, schemaFields) {
			t.Errorf("%s compiled fields = %v, schema fields = %v", viewSchemaID, planFields, schemaFields)
		}
	}
}

type artifactProjectionIntentSourceStub struct {
	artifactprojection.SourceReader
}

type assessmentProjectionIntentSourceStub struct {
	assessmentprojection.SourceReader
}

type indicatorProjectionIntentSourceStub struct {
	indicatorprojection.SourceReader
}
