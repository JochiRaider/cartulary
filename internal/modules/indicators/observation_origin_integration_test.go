package indicators_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestIndicatorObservationOriginConstraint_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "indicator-observation-origin-constraint")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "indicator-origin@example.test", "Indicator Origin", "IndicatorOriginPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-indicator-origin-incident", "IR-IND-ORIGIN", "Indicator observation origin")
	store := newIndicatorTestStore(t, harness.DB, revisionsupport.MustAppender(t))

	sourceID := uuid.New()
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, sourceID)
	observation, _, err := store.CreateIndicatorObservation(ctx, actor, indicators.IndicatorObservationCreateParams{
		IncidentID:     incident.ID,
		SourceRecordID: sourceID,
		SourceFieldKey: timelinetest.FieldSourceText,
		Producer:       indicators.ManualEntryObservationProducer(),
		OriginLocator:  "origin-constraint-test",
		ObservedText:   "constraint.example.test",
		CreatedAt:      indicatortest.BaseTime,
	})
	if err != nil {
		t.Fatalf("create valid observation: %v", err)
	}
	if observation.OriginKind.String() != "manual_entry" {
		t.Fatalf("stored origin = %q", observation.OriginKind)
	}

	if _, err := harness.DB.Exec(ctx, `SAVEPOINT invalid_origin_attempt`); err != nil {
		t.Fatalf("create invalid-origin savepoint: %v", err)
	}
	if _, err := harness.DB.Exec(ctx, `UPDATE indicator_observations SET origin_kind = ' manual_entry' WHERE indicator_observation_id = $1`, observation.ObservationID); err == nil {
		t.Fatal("database accepted non-exact observation origin")
	}
	if _, err := harness.DB.Exec(ctx, `ROLLBACK TO SAVEPOINT invalid_origin_attempt`); err != nil {
		t.Fatalf("restore after invalid-origin attempt: %v", err)
	}
	if _, err := harness.DB.Exec(ctx, `UPDATE indicator_observations SET origin_kind = 'system' WHERE indicator_observation_id = $1`, observation.ObservationID); err != nil {
		t.Fatalf("database rejected stored system origin: %v", err)
	}

	var validated bool
	if err := harness.DB.QueryRow(ctx, `
SELECT convalidated
  FROM pg_constraint
 WHERE conname = 'indicator_observations_origin_kind_ck'
`).Scan(&validated); err != nil || !validated {
		t.Fatalf("origin constraint validation = %t, %v", validated, err)
	}
}
