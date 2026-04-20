package entities_test

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	entities "github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

// Support-only projection rebuild determinism for Phase 4 entity projections.
func TestProjectionRebuildSupport_Phase4(t *testing.T) {
	t.Run("host and identity projection rows rebuild deterministically after entity upserts", func(t *testing.T) {
		harness := phase4test.StartStore(t, "phase4-projection-support-entity-projections")
		store := entities.NewStore(harness.Pool)
		projectionStore := projections.NewStore(harness.Pool)
		actor := phase4test.SeedLocalUserFlags(t, harness.DB, "phase4-projection-support@example.test", "Phase4 Projection Support", "Phase4ProjectionSupportPass1!", false, false, true)
		incident := phase4test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase4-projection-support-incident", "IR-P4SUP-01", "Phase 4 projection support")

		hostResult, err := store.CreateHostRow(context.Background(), actor, incident.ID, entities.CreateRequest{
			ClientTxnID: "txn-phase4-i-4-04-host",
			Values: map[string]string{
				"host.display_name": "Gateway 04",
				"host.hostname":     "GATEWAY-04",
				"host.fqdn":         "gateway-04.corp.example",
			},
		}, []byte("txn-phase4-i-4-04-host"), "req-phase4-i-4-04-host", golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("create host row: %v", err)
		}
		identityResult, err := store.CreateIdentityRow(context.Background(), actor, incident.ID, entities.CreateRequest{
			ClientTxnID: "txn-phase4-i-4-04-identity",
			Values: map[string]string{
				"identity.display_name":     "Alex Analyst 04",
				"identity.email":            "alex.analyst.04@example.test",
				"identity.sam_account_name": "ALEX04",
			},
		}, []byte("txn-phase4-i-4-04-identity"), "req-phase4-i-4-04-identity", golden.Phase4BaseTime.Add(time.Minute))
		if err != nil {
			t.Fatalf("create identity row: %v", err)
		}

		beforeHost := lookupHostProjectionSnapshot(t, harness.DB, hostResult.RecordID)
		beforeIdentity := lookupIdentityProjectionSnapshot(t, harness.DB, identityResult.RecordID)

		if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM host_grid_projection WHERE incident_id = $1`, incident.ID); err != nil {
			t.Fatalf("clear host projection rows: %v", err)
		}
		if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM identity_grid_projection WHERE incident_id = $1`, incident.ID); err != nil {
			t.Fatalf("clear identity projection rows: %v", err)
		}

		if err := projectionStore.RebuildIncidentHosts(context.Background(), incident.ID); err != nil {
			t.Fatalf("rebuild host projections: %v", err)
		}
		if err := projectionStore.RebuildIncidentIdentities(context.Background(), incident.ID); err != nil {
			t.Fatalf("rebuild identity projections: %v", err)
		}

		afterHost := lookupHostProjectionSnapshot(t, harness.DB, hostResult.RecordID)
		afterIdentity := lookupIdentityProjectionSnapshot(t, harness.DB, identityResult.RecordID)
		if !reflect.DeepEqual(beforeHost, afterHost) {
			t.Fatalf("expected deterministic host projection rebuild, before=%#v after=%#v", beforeHost, afterHost)
		}
		if !reflect.DeepEqual(beforeIdentity, afterIdentity) {
			t.Fatalf("expected deterministic identity projection rebuild, before=%#v after=%#v", beforeIdentity, afterIdentity)
		}
	})

	t.Run("indicator projection rows rebuild deterministically after observation and lifecycle fan-in", func(t *testing.T) {
		harness := phase4test.StartStore(t, "phase4-projection-support-indicator-projections")
		store := entities.NewStore(harness.Pool)
		projectionStore := projections.NewStore(harness.Pool)
		actor := phase4test.SeedLocalUserFlags(t, harness.DB, "phase4-projection-support-indicator@example.test", "Phase4 Projection Support Indicator", "Phase4ProjectionSupportIndicatorPass1!", false, false, true)
		incident := phase4test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase4-projection-support-indicator-incident", "IR-P4SUP-02", "Phase 4 projection support indicator")

		created, err := store.CreateIndicatorRow(context.Background(), actor, incident.ID, entities.CreateRequest{
			ClientTxnID: "txn-phase4-projection-support-indicator-create",
			Values: map[string]string{
				"indicator.indicator_type":   golden.Phase4IndicatorExamples[0].IndicatorType,
				"indicator.value_kind":       golden.Phase4IndicatorExamples[0].ValueKind,
				"indicator.display_value":    golden.Phase4IndicatorExamples[0].DisplayValue,
				"indicator.normalized_value": golden.Phase4IndicatorExamples[0].NormalizedValue,
			},
		}, []byte("txn-phase4-projection-support-indicator-create"), "req-phase4-projection-support-indicator-create", golden.Phase4BaseTime)
		if err != nil {
			t.Fatalf("create indicator row: %v", err)
		}
		phase4test.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineRecordID)
		phase4test.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.Phase4TimelineSiblingRecordID)

		if _, _, err := store.CreateIndicatorObservation(context.Background(), actor, entities.IndicatorObservationCreateParams{
			IncidentID:                incident.ID,
			SourceRecordID:            golden.Phase4TimelineRecordID,
			SourceFieldKey:            golden.Phase4FieldTimelineSourceText,
			OriginKind:                "interactive_cell",
			OriginLocator:             "view:timeline/record:1/cell:timeline.source_text/span:1-9",
			ObservedText:              "203[.]0[.]113[.]24",
			ResolvedIndicatorRecordID: &created.RecordID,
			CreatedAt:                 golden.Phase4PastTime,
		}); err != nil {
			t.Fatalf("create indicator observation one: %v", err)
		}
		if _, _, err := store.CreateIndicatorObservation(context.Background(), actor, entities.IndicatorObservationCreateParams{
			IncidentID:                incident.ID,
			SourceRecordID:            golden.Phase4TimelineSiblingRecordID,
			SourceFieldKey:            golden.Phase4FieldTimelineSummary,
			OriginKind:                "interactive_cell",
			OriginLocator:             "view:timeline/record:2/cell:timeline.summary/span:1-9",
			ObservedText:              "203[.]0[.]113[.]24",
			ResolvedIndicatorRecordID: &created.RecordID,
			CreatedAt:                 golden.Phase4BaseTime,
		}); err != nil {
			t.Fatalf("create indicator observation two: %v", err)
		}
		if _, _, err := store.AppendIndicatorLifecycleInterval(context.Background(), actor, entities.IndicatorLifecycleAppendParams{
			IncidentID:        incident.ID,
			IndicatorRecordID: created.RecordID,
			LifecycleState:    "active",
			ValidFrom:         golden.Phase4PastTime,
			CreatedAt:         golden.Phase4PastTime,
		}); err != nil {
			t.Fatalf("append indicator lifecycle interval: %v", err)
		}

		before := lookupIndicatorProjection(t, harness.DB, created.RecordID)
		if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM indicator_grid_projection WHERE incident_id = $1`, incident.ID); err != nil {
			t.Fatalf("clear indicator projection rows: %v", err)
		}
		if err := projectionStore.RebuildIncidentIndicators(context.Background(), incident.ID); err != nil {
			t.Fatalf("rebuild indicator projections: %v", err)
		}

		after := lookupIndicatorProjection(t, harness.DB, created.RecordID)
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("expected deterministic indicator projection rebuild, before=%#v after=%#v", before, after)
		}
	})
}

type hostProjectionSnapshot struct {
	RecordID    uuid.UUID
	IncidentID  uuid.UUID
	RowVersion  int64
	DisplayName string
	Hostname    string
	HostState   string
	EditedAt    time.Time
}

type identityProjectionSnapshot struct {
	RecordID       uuid.UUID
	IncidentID     uuid.UUID
	RowVersion     int64
	DisplayName    string
	Email          *string
	SamAccountName *string
	IdentityState  string
	EditedAt       time.Time
}

func lookupHostProjectionSnapshot(t *testing.T, db queryRower, recordID uuid.UUID) hostProjectionSnapshot {
	t.Helper()

	var (
		snapshot      hostProjectionSnapshot
		recordIDRaw   string
		incidentIDRaw string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT record_id::text, incident_id::text, row_version, display_name, hostname, host_state, edited_at
  FROM host_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&recordIDRaw, &incidentIDRaw, &snapshot.RowVersion, &snapshot.DisplayName, &snapshot.Hostname, &snapshot.HostState, &snapshot.EditedAt); err != nil {
		t.Fatalf("lookup host projection snapshot: %v", err)
	}
	snapshot.RecordID = phase4test.MustUUID(t, recordIDRaw)
	snapshot.IncidentID = phase4test.MustUUID(t, incidentIDRaw)
	snapshot.EditedAt = snapshot.EditedAt.UTC()
	return snapshot
}

func lookupIdentityProjectionSnapshot(t *testing.T, db queryRower, recordID uuid.UUID) identityProjectionSnapshot {
	t.Helper()

	var (
		snapshot       identityProjectionSnapshot
		recordIDRaw    string
		incidentIDRaw  string
		email          *string
		samAccountName *string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT record_id::text, incident_id::text, row_version, display_name, email::text, sam_account_name, identity_state, edited_at
  FROM identity_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&recordIDRaw, &incidentIDRaw, &snapshot.RowVersion, &snapshot.DisplayName, &email, &samAccountName, &snapshot.IdentityState, &snapshot.EditedAt); err != nil {
		t.Fatalf("lookup identity projection snapshot: %v", err)
	}
	snapshot.RecordID = phase4test.MustUUID(t, recordIDRaw)
	snapshot.IncidentID = phase4test.MustUUID(t, incidentIDRaw)
	snapshot.Email = email
	snapshot.SamAccountName = samAccountName
	snapshot.EditedAt = snapshot.EditedAt.UTC()
	return snapshot
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
