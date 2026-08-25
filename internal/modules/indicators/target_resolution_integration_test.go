package indicators_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestIndicatorTargetRoleFailuresAreAtomic_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "indicator-target-role-resolution")
	application := newIndicatorTestApplication(t, harness.DB, revisionsupport.MustAppender(t))
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "indicator-target-role@example.test", "Indicator Target Role", "IndicatorTargetRolePass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-target-role-incident", "IR-IND-TARGET-ROLE", "Indicator target roles")
	foreignIncident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-target-role-foreign-incident", "IR-IND-TARGET-FOREIGN", "Foreign Indicator target roles")
	now := time.Date(2026, 8, 3, 22, 0, 0, 0, time.UTC)

	mainIndicator := createTargetRoleIndicator(t, application, actor.ID, incident.ID, "main-target.example", "main")
	deletedIndicator := createTargetRoleIndicator(t, application, actor.ID, incident.ID, "deleted-target.example", "deleted")
	priorIndicator := createTargetRoleIndicator(t, application, actor.ID, incident.ID, "prior-target.example", "prior")
	foreignIndicator := createTargetRoleIndicator(t, application, actor.ID, foreignIncident.ID, "foreign-target.example", "foreign")

	validSource := uuid.New()
	wrongTypeTarget := uuid.New()
	deletedSource := uuid.New()
	foreignSource := uuid.New()
	transitionSource := uuid.New()
	priorSourceDependency := uuid.New()
	priorTargetDependency := uuid.New()
	for _, recordID := range []uuid.UUID{validSource, wrongTypeTarget, deletedSource, transitionSource, priorSourceDependency, priorTargetDependency} {
		timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, recordID)
	}
	timelinetest.SeedTimelineRecord(t, harness.DB, foreignIncident.ID, actor.ID, foreignSource)
	deleteTargetRoleEnvelope(t, harness.DB, deletedIndicator, actor.ID, now.Add(4*time.Second))
	deleteTargetRoleEnvelope(t, harness.DB, deletedSource, actor.ID, now.Add(5*time.Second))

	unresolved, err := application.CreateIndicatorObservation(ctx, actor.ID, manualObservationParams(incident.ID, transitionSource, timelinetest.FieldSourceText, nil, "txn-target-role-unresolved"))
	if err != nil {
		t.Fatalf("create unresolved observation: %v", err)
	}
	priorSourceObservation, err := application.CreateIndicatorObservation(ctx, actor.ID, manualObservationParams(incident.ID, priorSourceDependency, timelinetest.FieldSourceText, nil, "txn-target-role-prior-source"))
	if err != nil {
		t.Fatalf("create prior-source observation: %v", err)
	}
	priorTargetObservation, err := application.CreateIndicatorObservation(ctx, actor.ID, manualObservationParams(incident.ID, priorTargetDependency, timelinetest.FieldSourceText, &priorIndicator, "txn-target-role-prior-target"))
	if err != nil {
		t.Fatalf("create prior-target observation: %v", err)
	}
	deleteTargetRoleEnvelope(t, harness.DB, priorSourceDependency, actor.ID, now.Add(6*time.Second))
	deleteTargetRoleEnvelope(t, harness.DB, priorIndicator, actor.ID, now.Add(7*time.Second))

	assertFailure := func(name string, want error, operation func() error) {
		t.Helper()
		before := loadIndicatorMutationFootprint(t, harness.DB)
		err := operation()
		if !errors.Is(err, want) {
			t.Fatalf("%s error = %v, want %v", name, err, want)
		}
		after := loadIndicatorMutationFootprint(t, harness.DB)
		if after != before {
			t.Fatalf("%s changed durable footprint: before=%+v after=%+v", name, before, after)
		}
	}

	for name, sourceID := range map[string]uuid.UUID{
		"missing source": uuid.New(),
		"deleted source": deletedSource,
		"foreign source": foreignSource,
	} {
		name, sourceID := name, sourceID
		suffix := strings.ReplaceAll(name, " ", "-")
		assertFailure(name, indicators.ErrIndicatorSourceNotFound, func() error {
			_, callErr := application.CreateIndicatorObservation(ctx, actor.ID, manualObservationParams(incident.ID, sourceID, timelinetest.FieldSourceText, nil, "txn-target-role-"+suffix))
			return callErr
		})
	}

	for name, targetID := range map[string]uuid.UUID{
		"missing resolved target":    uuid.New(),
		"wrong-type resolved target": wrongTypeTarget,
		"deleted resolved target":    deletedIndicator,
		"foreign resolved target":    foreignIndicator,
	} {
		name, targetID := name, targetID
		suffix := strings.ReplaceAll(name, " ", "-")
		assertFailure(name, indicators.ErrResolvedIndicatorNotFound, func() error {
			_, callErr := application.CreateIndicatorObservation(ctx, actor.ID, manualObservationParams(incident.ID, validSource, timelinetest.FieldSourceText, &targetID, "txn-target-role-"+suffix))
			return callErr
		})
		assertFailure("transition "+name, indicators.ErrResolvedIndicatorNotFound, func() error {
			_, callErr := application.ResolveIndicatorObservation(ctx, actor.ID, indicators.IndicatorObservationResolveParams{
				ObservationID: unresolved.Observation.ObservationID, ResolvedIndicatorRecordID: targetID,
				BaseRowVersion: 1, ClientTxnID: "txn-target-role-transition-" + suffix,
				RequestID: "req-target-role-transition-" + suffix,
			})
			return callErr
		})
	}

	for name, targetID := range map[string]uuid.UUID{
		"missing addressed Indicator":    uuid.New(),
		"wrong-type addressed Indicator": wrongTypeTarget,
		"deleted addressed Indicator":    deletedIndicator,
		"foreign addressed Indicator":    foreignIndicator,
	} {
		name, targetID := name, targetID
		suffix := strings.ReplaceAll(name, " ", "-")
		assertFailure(name, indicators.ErrIndicatorNotFound, func() error {
			params := lifecycleAppendParams(incident.ID, targetID, 1, now.Add(time.Hour), "txn-target-role-"+suffix)
			_, callErr := application.AppendIndicatorLifecycleInterval(ctx, actor.ID, params)
			return callErr
		})
	}

	for name, supportID := range map[string]uuid.UUID{
		"missing support": uuid.New(),
		"deleted support": deletedSource,
		"foreign support": foreignSource,
	} {
		name, supportID := name, supportID
		suffix := strings.ReplaceAll(name, " ", "-")
		assertFailure(name, indicators.ErrInvalidCreateRequest, func() error {
			params := lifecycleAppendParams(incident.ID, mainIndicator, 1, now.Add(2*time.Hour), "txn-target-role-"+suffix)
			params.SupportRefs = []uuid.UUID{supportID}
			_, callErr := application.AppendIndicatorLifecycleInterval(ctx, actor.ID, params)
			return callErr
		})
	}

	assertFailure("unavailable prior source", indicators.ErrIndicatorObservationNotFound, func() error {
		_, callErr := application.DismissIndicatorObservation(ctx, actor.ID, indicators.IndicatorObservationActionParams{
			ObservationID: priorSourceObservation.Observation.ObservationID, BaseRowVersion: 1,
			ClientTxnID: "txn-target-role-prior-source-dismiss", RequestID: "req-target-role-prior-source-dismiss",
		})
		return callErr
	})
	assertFailure("unavailable prior target", indicators.ErrIndicatorObservationNotFound, func() error {
		_, callErr := application.DismissIndicatorObservation(ctx, actor.ID, indicators.IndicatorObservationActionParams{
			ObservationID: priorTargetObservation.Observation.ObservationID, BaseRowVersion: 1,
			ClientTxnID: "txn-target-role-prior-target-dismiss", RequestID: "req-target-role-prior-target-dismiss",
		})
		return callErr
	})
}

func createTargetRoleIndicator(t testing.TB, application *indicators.Application, actorUserID uuid.UUID, incidentID uuid.UUID, value string, suffix string) uuid.UUID {
	t.Helper()
	result, err := application.CreateIndicatorRow(context.Background(), actorUserID, incidentID, indicators.CreateCommand{
		ClientTxnID:   "txn-target-role-indicator-" + suffix,
		IndicatorType: "domain_name",
		ValueKind:     "atomic",
		DisplayValue:  value,
	}, "req-target-role-indicator-"+suffix)
	if err != nil {
		t.Fatalf("create target-role Indicator %s: %v", suffix, err)
	}
	return result.RecordID
}

func deleteTargetRoleEnvelope(t testing.TB, db postgres.DB, recordID uuid.UUID, actorID uuid.UUID, deletedAt time.Time) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `
UPDATE records
   SET deleted_at = $2,
       deleted_by_user_id = $3,
       updated_at = $2,
       updated_by_user_id = $3,
       row_version = row_version + 1
 WHERE record_id = $1
`, recordID, deletedAt, actorID); err != nil {
		t.Fatalf("delete target-role envelope %s: %v", recordID, err)
	}
}

type indicatorMutationFootprint struct {
	RecordsVersionSum int64
	Observations      int64
	Intervals         int64
	ChangeSets        int64
	Collaboration     int64
	Idempotency       int64
}

func loadIndicatorMutationFootprint(t testing.TB, db postgres.DB) indicatorMutationFootprint {
	t.Helper()
	var result indicatorMutationFootprint
	if err := db.QueryRow(context.Background(), `
SELECT
    COALESCE((SELECT SUM(row_version) FROM records), 0),
    (SELECT COUNT(*) FROM indicator_observations),
    (SELECT COUNT(*) FROM indicator_state_intervals),
    (SELECT COUNT(*) FROM change_sets),
    (SELECT COUNT(*) FROM route_idempotency)
`).Scan(&result.RecordsVersionSum, &result.Observations, &result.Intervals, &result.ChangeSets, &result.Idempotency); err != nil {
		t.Fatalf("load Indicator mutation footprint: %v", err)
	}
	result.Collaboration = int64(collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{}))
	return result
}
