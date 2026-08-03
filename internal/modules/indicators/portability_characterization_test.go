package indicators_test

import (
	"bytes"
	"context"
	"slices"
	"testing"
	"time"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestIndicatorPortableRowsCharacterization_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "indicator-portable-rows-characterization")
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "indicator-portability@example.test", "Indicator Portability", "IndicatorPortabilityPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-indicator-portability-incident", "IR-IND-PORT", "Indicator portability characterization")
	store := newIndicatorTestStore(t, harness.DB, revisionsupport.MustAppender(t))

	created, err := store.CreateIndicatorRow(ctx, actor, incident.ID, indicators.CreateCommand{
		ClientTxnID:   "txn-indicator-portability-create",
		IndicatorType: "domain_name",
		ValueKind:     "atomic",
		DisplayValue:  "PORTABLE[.]EXAMPLE.TEST",
	}, []byte("indicator-portability-create"), "req-indicator-portability-create", indicatortest.BaseTime)
	if err != nil {
		t.Fatalf("create indicator: %v", err)
	}
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.RecordID)
	if _, _, err := store.CreateIndicatorObservation(ctx, actor, indicators.IndicatorObservationCreateParams{
		IncidentID:                incident.ID,
		SourceRecordID:            timelinetest.RecordID,
		SourceFieldKey:            timelinetest.FieldSourceText,
		Producer:                  indicators.ManualEntryObservationProducer(),
		OriginLocator:             "indicator-portability-characterization",
		ObservedText:              "PORTABLE[.]EXAMPLE.TEST",
		ResolvedIndicatorRecordID: &created.RecordID,
		CreatedAt:                 indicatortest.PastTime,
	}); err != nil {
		t.Fatalf("create observation: %v", err)
	}
	if _, _, err := store.AppendIndicatorLifecycleInterval(ctx, actor, indicators.IndicatorLifecycleAppendParams{
		IncidentID:        incident.ID,
		IndicatorRecordID: created.RecordID,
		LifecycleState:    "active",
		ValidFrom:         indicatortest.PastTime,
		CreatedAt:         indicatortest.PastTime,
	}); err != nil {
		t.Fatalf("append interval: %v", err)
	}

	contribution := indicators.NewIncidentBundleContribution()
	first, err := contribution.SourcePort.Export(ctx, sourceport.ExportContext{Query: harness.DB, IncidentID: incident.ID})
	if err != nil {
		t.Fatalf("first export: %v", err)
	}
	second, err := contribution.SourcePort.Export(ctx, sourceport.ExportContext{Query: harness.DB, IncidentID: incident.ID})
	if err != nil {
		t.Fatalf("second export: %v", err)
	}
	if len(first) != 3 || len(second) != len(first) {
		t.Fatalf("portable files = %d/%d, want 3", len(first), len(second))
	}
	for index := range first {
		if first[index].Path != second[index].Path || !bytes.Equal(first[index].Payload, second[index].Payload) {
			t.Fatalf("export %q is not deterministic", first[index].Path)
		}
	}

	wantKeys := map[string][]string{
		"data/indicators.ndjson": {
			"created_at", "created_by_user_id", "dedupe_key", "defanged_value", "deleted_at", "deleted_by_user_id",
			"display_value", "hash_algorithm", "hash_value", "incident_id", "indicator_type", "normalized_value",
			"record_id", "row_version", "stix_pattern", "updated_at", "updated_by_user_id", "value_kind",
		},
		"data/indicator_observations.ndjson": {
			"created_at", "created_by_user_id", "deleted_at", "deleted_by_user_id", "incident_id", "indicator_observation_id",
			"normalized_candidate", "observed_text", "origin_kind", "origin_locator", "parsed_indicator_type", "resolution_method",
			"resolution_status", "resolved_at", "resolved_by_user_id", "resolved_indicator_record_id", "row_version", "source_field_key", "source_record_id",
		},
		"data/indicator_state_intervals.ndjson": {
			"assessed_at", "assessor", "confidence", "created_at", "created_by_user_id", "deleted_at", "deleted_by_user_id",
			"incident_id", "indicator_record_id", "indicator_state_interval_id", "lifecycle_state", "rationale", "row_version",
			"support_refs", "valid_from", "valid_to",
		},
	}
	for _, file := range first {
		rows, err := incidentportability.DecodeNDJSON(file.Payload)
		if err != nil || len(rows) != 1 {
			t.Fatalf("decode %s: rows=%d err=%v payload=%s", file.Path, len(rows), err, file.Payload)
		}
		keys := make([]string, 0, len(rows[0]))
		for key := range rows[0] {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		expected := append([]string(nil), wantKeys[file.Path]...)
		slices.Sort(expected)
		if !slices.Equal(keys, expected) {
			t.Fatalf("%s keys = %v, want %v", file.Path, keys, expected)
		}
		if rows[0]["deleted_at"] != nil || rows[0]["deleted_by_user_id"] != nil {
			t.Fatalf("%s nullable tombstone fields were not explicit nulls: %#v", file.Path, rows[0])
		}
		if file.Path == "data/indicator_observations.ndjson" && rows[0]["origin_kind"] != "manual_entry" {
			t.Fatalf("%s origin_kind = %#v, want manual_entry", file.Path, rows[0]["origin_kind"])
		}
		createdAt, ok := rows[0]["created_at"].(string)
		if !ok {
			t.Fatalf("%s created_at = %#v", file.Path, rows[0]["created_at"])
		}
		if _, err := time.Parse(time.RFC3339Nano, createdAt); err != nil {
			t.Fatalf("%s created_at is not RFC3339: %q: %v", file.Path, createdAt, err)
		}
	}

	descriptor := contribution.SourcePort.Descriptor()
	if len(descriptor.Paths) != 3 {
		t.Fatalf("source-port paths = %#v", descriptor.Paths)
	}
	for _, path := range descriptor.Paths {
		if !slices.Equal(path.Versions, []int{1, 2}) {
			t.Fatalf("%s versions = %v, want [1 2]", path.LogicalPath, path.Versions)
		}
	}
}
