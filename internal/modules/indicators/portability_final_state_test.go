package indicators_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestIndicatorPortableFinalStateInvariants_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "indicator-portability-final-state")
	actor := authstoretest.SeedLocalUserRecord(
		t, harness.DB, "indicator-portability-invariants@example.test",
		"Indicator Portability Invariants", "IndicatorPortabilityPass1!", false, false, true,
	)
	incident := appsupport.CreateIncidentInStore(
		t, harness.DB, actor, "txn-indicator-portability-invariants", "IR-IND-PORT-INV",
		"Indicator portability invariants",
	)
	foreignIncident := appsupport.CreateIncidentInStore(
		t, harness.DB, actor, "txn-indicator-portability-foreign", "IR-IND-PORT-FOR",
		"Indicator portability foreign incident",
	)

	testCases := []struct {
		name      string
		invariant string
		arrange   func(testing.TB, pgx.Tx, *portableIntegrationScenario)
		mutate    func(testing.TB, pgx.Tx, portableIntegrationScenario)
	}{
		{name: "valid"},
		{
			name: "observation_same_incident", invariant: "indicators.observation_same_incident",
			arrange: func(_ testing.TB, _ pgx.Tx, scenario *portableIntegrationScenario) {
				scenario.sourceIncidentID = foreignIncident.ID
			},
		},
		{
			name: "interval_same_incident", invariant: "indicators.interval_same_incident",
			arrange: func(t testing.TB, tx pgx.Tx, scenario *portableIntegrationScenario) {
				scenario.intervalIndicatorID = uuid.MustParse("00000000-0000-4000-8000-000000000030")
				seedPortableIndicatorEnvelopeAndRow(
					t, tx, foreignIncident.ID, scenario.intervalIndicatorID, actor.ID, scenario.timestamp,
				)
			},
		},
		{
			name: "repeated_observations_preserved", invariant: "indicators.repeated_observations_preserved",
			mutate: func(t testing.TB, tx pgx.Tx, scenario portableIntegrationScenario) {
				if _, err := tx.Exec(
					ctx, `DELETE FROM indicator_observations WHERE indicator_observation_id = $1`,
					scenario.observationID,
				); err != nil {
					t.Fatalf("delete admitted observation: %v", err)
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin portability transaction: %v", err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			scenario := newPortableIntegrationScenario(incident.ID, actor.ID)
			if testCase.arrange != nil {
				testCase.arrange(t, tx, &scenario)
			}
			seedPortableRecordEnvelope(
				t, tx, scenario.incidentID, scenario.indicatorID, "indicator",
				scenario.actorID, scenario.timestamp,
			)
			seedPortableRecordEnvelope(
				t, tx, scenario.sourceIncidentID, scenario.sourceRecordID, "timeline_event",
				scenario.actorID, scenario.timestamp,
			)
			recorder := &indicatorPortableAttributionRecorder{}
			for _, column := range []string{"created_by_user_id", "updated_by_user_id"} {
				if err := recorder.RecordImportedAttribution(
					"records", scenario.indicatorID.String(), column, scenario.actorID.String(),
				); err != nil {
					t.Fatalf("record envelope attribution: %v", err)
				}
			}
			importContext := scenario.importContext(t, recorder, "indicator-final-state-"+testCase.name)
			port := indicators.NewIncidentBundleContribution().SourcePort
			prepared, err := port.PrepareImport(ctx, scenario.bundle(t), importContext)
			if err != nil {
				t.Fatalf("prepare portable scenario: %v", err)
			}
			if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
				if testCase.invariant == "indicators.observation_same_incident" ||
					testCase.invariant == "indicators.interval_same_incident" {
					assertPortableIndicatorInvariant(t, err, testCase.invariant)
					return
				}
				t.Fatalf("apply portable scenario: %v", err)
			}
			if testCase.mutate != nil {
				testCase.mutate(t, tx, scenario)
			}
			err = port.ValidateImportTx(ctx, tx, prepared, importContext)
			if testCase.invariant == "" {
				if err != nil {
					t.Fatalf("validate portable scenario: %v", err)
				}
				return
			}
			assertPortableIndicatorInvariant(t, err, testCase.invariant)
		})
	}
}

type portableIntegrationScenario struct {
	incidentID          uuid.UUID
	actorID             uuid.UUID
	indicatorID         uuid.UUID
	observationID       uuid.UUID
	intervalID          uuid.UUID
	sourceRecordID      uuid.UUID
	sourceIncidentID    uuid.UUID
	intervalIndicatorID uuid.UUID
	timestamp           time.Time
}

func newPortableIntegrationScenario(incidentID, actorID uuid.UUID) portableIntegrationScenario {
	indicatorID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	return portableIntegrationScenario{
		incidentID: incidentID, actorID: actorID, indicatorID: indicatorID,
		observationID:    uuid.MustParse("00000000-0000-4000-8000-000000000010"),
		intervalID:       uuid.MustParse("00000000-0000-4000-8000-000000000020"),
		sourceRecordID:   uuid.MustParse("00000000-0000-4000-8000-000000000102"),
		sourceIncidentID: incidentID, intervalIndicatorID: indicatorID,
		timestamp: time.Date(2025, 1, 2, 3, 4, 5, 1000, time.UTC),
	}
}

func (scenario portableIntegrationScenario) importContext(
	t testing.TB,
	recorder incidentportability.AttributionRecorder,
	operationID string,
) sourceport.ImportContext {
	t.Helper()
	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{
		SourceActorID: scenario.actorID.String(),
	}})
	if err != nil {
		t.Fatalf("actor catalog: %v", err)
	}
	return sourceport.ImportContext{
		IncidentID: scenario.incidentID, ActorUserID: scenario.actorID,
		BundleVersion: 2, OperationID: operationID, Attributions: recorder, Actors: actors,
	}
}

func (scenario portableIntegrationScenario) bundle(t testing.TB) sourceport.MapBundle {
	t.Helper()
	canonical, err := identity.Canonicalize(identity.Input{
		IndicatorType: "domain_name", ValueKind: "atomic", DisplayValue: "example.test",
	})
	if err != nil {
		t.Fatalf("canonicalize integration fixture: %v", err)
	}
	createdAt := formatPortableScenarioTimestamp(scenario.timestamp)
	indicatorRow := map[string]any{
		"record_id": scenario.indicatorID.String(), "incident_id": scenario.incidentID.String(),
		"indicator_type": canonical.IndicatorType, "value_kind": canonical.ValueKind,
		"display_value": canonical.DisplayValue, "normalized_value": canonical.NormalizedValue,
		"dedupe_key": canonical.DedupeKey, "defanged_value": nil,
		"hash_algorithm": nil, "hash_value": nil, "stix_pattern": nil, "row_version": 1,
		"created_at": createdAt, "updated_at": createdAt,
		"created_by_user_id": scenario.actorID.String(), "updated_by_user_id": scenario.actorID.String(),
		"deleted_at": nil, "deleted_by_user_id": nil,
	}
	observationRow := map[string]any{
		"indicator_observation_id": scenario.observationID.String(),
		"incident_id":              scenario.incidentID.String(), "source_record_id": scenario.sourceRecordID.String(),
		"source_field_key": "source_text", "origin_kind": "manual_entry",
		"origin_locator": "portable-fixture", "observed_text": "not parsed",
		"parsed_indicator_type": nil, "normalized_candidate": nil,
		"resolution_status": "unresolved", "resolved_indicator_record_id": nil,
		"row_version": 1, "created_by_user_id": scenario.actorID.String(), "created_at": createdAt,
		"resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil,
		"deleted_at": nil, "deleted_by_user_id": nil,
	}
	intervalRow := map[string]any{
		"indicator_state_interval_id": scenario.intervalID.String(),
		"incident_id":                 scenario.incidentID.String(),
		"indicator_record_id":         scenario.intervalIndicatorID.String(),
		"lifecycle_state":             "active", "valid_from": createdAt, "valid_to": nil,
		"confidence": nil, "rationale": nil, "support_refs": []any{}, "assessor": nil,
		"assessed_at": createdAt, "row_version": 1,
		"created_by_user_id": scenario.actorID.String(), "created_at": createdAt,
		"deleted_at": nil, "deleted_by_user_id": nil,
	}
	return sourceport.MapBundle{
		"data/indicators.ndjson":                marshalPortableRows(t, indicatorRow),
		"data/indicator_observations.ndjson":    marshalPortableRows(t, observationRow),
		"data/indicator_state_intervals.ndjson": marshalPortableRows(t, intervalRow),
	}
}

func seedPortableRecordEnvelope(
	t testing.TB,
	tx pgx.Tx,
	incidentID, recordID uuid.UUID,
	recordType string,
	actorID uuid.UUID,
	timestamp time.Time,
) {
	t.Helper()
	if _, err := tx.Exec(context.Background(), `
INSERT INTO records (
    record_id, incident_id, record_type, created_at, created_by_user_id,
    updated_at, updated_by_user_id, row_version
)
VALUES ($1, $2, $3, $4, $5, $4, $5, 1)
`, recordID, incidentID, recordType, timestamp, actorID); err != nil {
		t.Fatalf("seed portable record envelope: %v", err)
	}
}

func seedPortableIndicatorEnvelopeAndRow(
	t testing.TB,
	tx pgx.Tx,
	incidentID, recordID, actorID uuid.UUID,
	timestamp time.Time,
) {
	t.Helper()
	seedPortableRecordEnvelope(t, tx, incidentID, recordID, "indicator", actorID, timestamp)
	canonical, err := identity.Canonicalize(identity.Input{
		IndicatorType: "domain_name", ValueKind: "atomic", DisplayValue: "foreign.example.test",
	})
	if err != nil {
		t.Fatalf("canonicalize foreign indicator: %v", err)
	}
	if _, err := tx.Exec(context.Background(), `
INSERT INTO indicators (
    record_id, incident_id, indicator_type, value_kind, display_value,
    normalized_value, dedupe_key
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`, recordID, incidentID, canonical.IndicatorType, canonical.ValueKind,
		canonical.DisplayValue, canonical.NormalizedValue, canonical.DedupeKey); err != nil {
		t.Fatalf("seed foreign indicator: %v", err)
	}
}

func marshalPortableRows(t testing.TB, rows ...map[string]any) []byte {
	t.Helper()
	payload := make([]byte, 0, len(rows)*256)
	for _, row := range rows {
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatalf("marshal portable row: %v", err)
		}
		payload = append(payload, encoded...)
		payload = append(payload, '\n')
	}
	return payload
}

func formatPortableScenarioTimestamp(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05.999999+00:00")
}

func assertPortableIndicatorInvariant(t testing.TB, err error, invariantID string) {
	t.Helper()
	want := "incident bundle source family indicators failed invariant " + invariantID
	if err == nil || err.Error() != want {
		t.Fatalf("failure = %#v, want indicators/%s", err, invariantID)
	}
}

type indicatorPortableAttributionRecorder struct {
	rows []incidentportability.ImportedAttribution
}

func (recorder *indicatorPortableAttributionRecorder) RecordImportedAttribution(
	table, sourceRowID, column, sourceActorID string,
) error {
	if recorder == nil || table == "" || sourceRowID == "" || column == "" || sourceActorID == "" {
		return errors.New("invalid portable attribution")
	}
	for _, row := range recorder.rows {
		if row.SourceTable == table && row.SourceRowID == sourceRowID && row.SourceColumn == column {
			if row.SourceActorID == sourceActorID {
				return nil
			}
			return errors.New("conflicting portable attribution")
		}
	}
	recorder.rows = append(recorder.rows, incidentportability.ImportedAttribution{
		SourceTable: table, SourceRowID: sourceRowID,
		SourceColumn: column, SourceActorID: sourceActorID,
	})
	return nil
}

func (recorder *indicatorPortableAttributionRecorder) ImportedAttributions() []incidentportability.ImportedAttribution {
	if recorder == nil {
		return nil
	}
	return append([]incidentportability.ImportedAttribution(nil), recorder.rows...)
}
