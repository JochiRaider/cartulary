package performancefixture

import (
	"context"
	"testing"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestTimelineFixtureObservesDefaultView_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "timeline_fixture_default")
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "fixture-default@example.test", DisplayName: "Fixture default"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, uuid.NewString(), "IR-FIXTURE-DEFAULT", "Fixture default")
	appsupport.CreateIncidentInStore(t, harness.DB, actor, uuid.NewString(), "IR-FIXTURE-OTHER", "Unrelated default")
	profile := profileForTest(t)
	expected, err := semanticExpectations(profile)
	if err != nil {
		t.Fatal(err)
	}
	validator := productionSemanticValidator{pool: harness.DB, profile: profile}
	// This small fixture isolates the default-view predicate. It does not claim
	// the other population conditions or materialize a measurement workload.
	assertDefault := func(want bool) {
		t.Helper()
		actual, err := validator.ValidateFixtureSemantics(ctx, incident.ID.String(), expected)
		if err != nil {
			t.Fatal(err)
		}
		if value, present := actual.Conditions["default_view"]; !present || value != want {
			t.Fatalf("default view condition = %v (present=%v), want %v", value, present, want)
		}
	}
	assertDefault(true)
	for _, state := range []struct {
		statement string
		want      bool
	}{
		{`UPDATE incident_workbook_preferences SET default_sheet_ref='{}'::jsonb WHERE incident_id=$1`, false},
		{`UPDATE incident_workbook_preferences SET default_sheet_ref=NULL WHERE incident_id=$1`, true},
		{`DELETE FROM incident_workbook_preferences WHERE incident_id=$1`, false},
	} {
		result, err := harness.DB.Exec(ctx, state.statement, incident.ID)
		if err != nil || result.RowsAffected() != 1 {
			t.Fatalf("change exact target preference: err=%v rows=%d", err, result.RowsAffected())
		}
		assertDefault(state.want)
	}
}
