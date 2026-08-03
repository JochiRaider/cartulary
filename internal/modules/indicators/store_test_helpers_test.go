package indicators_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func newIndicatorTestStore(t testing.TB, db postgres.DB, appender *revisions.Appender) *indicators.Store {
	t.Helper()
	coordinator := newIndicatorTestProjectionCoordinator(t, db)
	store, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:    db,
		Revisions:   appender,
		Projections: coordinator,
		SourceText:  indicatorassembly.NewSourceTextPort(coordinator),
	})
	if err != nil {
		t.Fatalf("compose Indicator test store: %v", err)
	}
	return store
}

func newIndicatorTestProjectionCoordinator(t testing.TB, db postgres.DB) *projections.Coordinator {
	t.Helper()
	return timelineassembly.NewProjectionBundle(db).Coordinator
}

func manualObservationParams(incidentID uuid.UUID, sourceID uuid.UUID, fieldKey string, resolvedID *uuid.UUID, clientTxnID string) indicators.IndicatorObservationCreateParams {
	return indicators.IndicatorObservationCreateParams{
		IncidentID: incidentID, SourceRecordID: sourceID, BaseRowVersion: 1,
		SourceFieldKey: fieldKey, SpanStartByte: 0, SpanEndByte: len("record-support-source-row"),
		ResolvedIndicatorRecordID: resolvedID, ClientTxnID: clientTxnID,
		RequestID: "req-" + clientTxnID, RequestHash: []byte("hash-" + clientTxnID),
	}
}

func lifecycleAppendParams(incidentID uuid.UUID, indicatorID uuid.UUID, baseRowVersion int64, validFrom time.Time, clientTxnID string) indicators.IndicatorLifecycleAppendParams {
	return indicators.IndicatorLifecycleAppendParams{
		IncidentID: incidentID, IndicatorRecordID: indicatorID, BaseRowVersion: baseRowVersion,
		LifecycleState: "active", ValidFrom: validFrom, SupportRefs: []uuid.UUID{},
		ClientTxnID: clientTxnID, RequestID: "req-" + clientTxnID, RequestHash: []byte("hash-" + clientTxnID),
	}
}

type indicatorProjectionRow struct {
	RecordID            uuid.UUID
	RowVersion          int64
	IndicatorType       string
	ValueKind           string
	DisplayValue        string
	NormalizedValue     *string
	DefangedValue       *string
	HashAlgorithm       *string
	HashValue           *string
	STIXPattern         *string
	FirstObservedAt     *time.Time
	LastObservedAt      *time.Time
	ObservationCount    int
	LifecycleSummary    *string
	SupportingLinkCount int
}

func lookupIndicatorProjection(t testing.TB, db postgres.DB, recordID uuid.UUID) indicatorProjectionRow {
	t.Helper()
	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin projection lookup: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	projected, err := newIndicatorTestProjectionCoordinator(t, db).LoadRowTx(
		context.Background(),
		tx,
		indicators.ViewSchemaID,
		recordID,
	)
	if err != nil {
		t.Fatalf("load Indicator projection through coordinator: %v", err)
	}
	cells, ok := projected["cells"].(map[string]any)
	if !ok {
		t.Fatalf("Indicator projection cells = %#v", projected["cells"])
	}
	return indicatorProjectionRow{
		RecordID:            uuid.MustParse(projected["record_id"].(string)),
		RowVersion:          indicatorProjectionInteger(t, projected["row_version"]),
		IndicatorType:       indicatorProjectionText(t, cells, "indicator.indicator_type"),
		ValueKind:           indicatorProjectionText(t, cells, "indicator.value_kind"),
		DisplayValue:        indicatorProjectionText(t, cells, "indicator.display_value"),
		NormalizedValue:     indicatorProjectionOptionalText(t, cells, "indicator.normalized_value"),
		DefangedValue:       indicatorProjectionOptionalText(t, cells, "indicator.defanged_value"),
		HashAlgorithm:       indicatorProjectionOptionalText(t, cells, "indicator.hash_algorithm"),
		HashValue:           indicatorProjectionOptionalText(t, cells, "indicator.hash_value"),
		STIXPattern:         indicatorProjectionOptionalText(t, cells, "indicator.stix_pattern"),
		FirstObservedAt:     indicatorProjectionOptionalTime(t, cells, "indicator.first_observed_at"),
		LastObservedAt:      indicatorProjectionOptionalTime(t, cells, "indicator.last_observed_at"),
		ObservationCount:    int(indicatorProjectionInteger(t, indicatorProjectionCellValue(t, cells, "indicator.observation_count"))),
		LifecycleSummary:    indicatorProjectionOptionalText(t, cells, "indicator.lifecycle_summary"),
		SupportingLinkCount: int(indicatorProjectionInteger(t, indicatorProjectionCellValue(t, cells, "indicator.supporting_link_count"))),
	}
}

func indicatorProjectionCellValue(t testing.TB, cells map[string]any, fieldKey string) any {
	t.Helper()
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		t.Fatalf("projection cell %s = %#v", fieldKey, cells[fieldKey])
	}
	return cell["value"]
}

func indicatorProjectionText(t testing.TB, cells map[string]any, fieldKey string) string {
	t.Helper()
	value, ok := indicatorProjectionCellValue(t, cells, fieldKey).(string)
	if !ok {
		t.Fatalf("projection text %s has wrong type", fieldKey)
	}
	return value
}

func indicatorProjectionOptionalText(t testing.TB, cells map[string]any, fieldKey string) *string {
	t.Helper()
	value := indicatorProjectionCellValue(t, cells, fieldKey)
	if value == nil {
		return nil
	}
	text, ok := value.(string)
	if !ok {
		t.Fatalf("projection optional text %s = %#v", fieldKey, value)
	}
	return &text
}

func indicatorProjectionOptionalTime(t testing.TB, cells map[string]any, fieldKey string) *time.Time {
	t.Helper()
	value := indicatorProjectionCellValue(t, cells, fieldKey)
	if value == nil {
		return nil
	}
	text, ok := value.(string)
	if !ok {
		t.Fatalf("projection timestamp %s = %#v", fieldKey, value)
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		t.Fatalf("parse projection timestamp %s: %v", fieldKey, err)
	}
	parsed = parsed.UTC()
	return &parsed
}

func indicatorProjectionInteger(t testing.TB, value any) int64 {
	t.Helper()
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	default:
		t.Fatalf("projection integer = %#v", value)
		return 0
	}
}
