package telemetry

import (
	"context"
	"slices"
	"strings"
	"testing"
)

func TestSpanRegistryClosed(t *testing.T) {
	rows := SpanRegistry()
	wantFamilies := []string{
		"http_server",
		"workbook_query",
		"workbook_mutation",
		"workbook_projection",
		"websocket_lifecycle",
		"websocket_event_send",
		"job_enqueue",
		"job_run",
		"postgres_dependency",
		"objectstore_dependency",
	}
	if len(rows) != len(wantFamilies) {
		t.Fatalf("span registry row count = %d want %d", len(rows), len(wantFamilies))
	}
	seenFamilies := map[string]bool{}
	seenNames := map[string]bool{}
	for i, row := range rows {
		if row.Family != wantFamilies[i] {
			t.Fatalf("span family %d = %q want %q", i, row.Family, wantFamilies[i])
		}
		if seenFamilies[row.Family] || row.Name == "" || row.Kind == "" || !RegisteredScope(row.Scope) {
			t.Fatalf("invalid span registry row: %#v", row)
		}
		seenFamilies[row.Family] = true
		if row.Name != "<HTTP_METHOD> <route_template>" {
			if seenNames[row.Name] {
				t.Fatalf("duplicate span name %q", row.Name)
			}
			seenNames[row.Name] = true
		}
		if row.ParentRule == "" || row.LinkRule != "none" || row.StatusRule == "" || row.LifecycleBoundary == "" {
			t.Fatalf("span lifecycle/status/causality metadata incomplete: %#v", row)
		}
		for _, attr := range row.RequiredAttributes {
			if strings.Contains(attr, "id") && attr != "cartulary.view_schema_id" {
				t.Fatalf("span required attribute contains forbidden identity shape: %#v", row)
			}
		}
	}
}

func TestMetricRegistryClosed(t *testing.T) {
	rows := MetricRegistry()
	wantNames := []string{
		"cartulary.http.server.request.duration",
		"cartulary.workbook.query.duration",
		"cartulary.workbook.mutation.duration",
		"cartulary.workbook.rows.returned",
		"cartulary.collaboration.connections.active",
		"cartulary.collaboration.events.sent",
		"cartulary.jobs.active",
		"cartulary.jobs.duration",
		"cartulary.postgres.operation.duration",
		"cartulary.objectstore.operation.duration",
		"cartulary.objectstore.transfer.bytes",
		IncidentBundleV1ImportMetricName,
		TelemetryExportFailureMetricName,
		TelemetryItemDroppedMetricName,
		TelemetryQueueDepthMetricName,
	}
	if len(rows) != len(wantNames) {
		t.Fatalf("metric registry row count = %d want %d", len(rows), len(wantNames))
	}
	seen := map[string]MetricRegistryRow{}
	for i, row := range rows {
		if row.Name != wantNames[i] {
			t.Fatalf("metric %d = %q want %q", i, row.Name, wantNames[i])
		}
		folded := strings.ToLower(row.Name)
		if folded != row.Name {
			t.Fatalf("metric name must be lowercase ASCII: %q", row.Name)
		}
		if previous, exists := seen[folded]; exists {
			t.Fatalf("case-insensitive duplicate metric %q conflicts with %#v", row.Name, previous)
		}
		seen[folded] = row
		if row.InstrumentKind == "" || row.Unit == "" || row.Description == "" || row.Aggregation == "" || row.Temporality == "" {
			t.Fatalf("metric identity incomplete: %#v", row)
		}
		if row.OverflowBehavior != "drop_metric_overflow" {
			t.Fatalf("metric overflow behavior mismatch: %#v", row)
		}
		for _, attr := range append(row.AllowedAttributes, row.OptionalAttributes...) {
			if strings.Contains(attr, "id") && attr != "cartulary.view_schema_id" {
				t.Fatalf("metric attribute contains forbidden identity shape: %#v", row)
			}
		}
	}
	if !slices.Equal(MetricRegistry()[3].Buckets, []float64{0, 1, 5, 10, 25, 50, 100, 250, 500}) {
		t.Fatalf("workbook rows histogram buckets mismatch: %#v", MetricRegistry()[3].Buckets)
	}
	if row := seen["cartulary.jobs.active"]; row.InstrumentKind != "ObservableGauge" || row.Unit != "{job}" || !slices.Equal(row.AllowedAttributes, []string{"cartulary.job_kind"}) {
		t.Fatalf("jobs active metric row mismatch: %#v", row)
	}
}

func TestIncidentBundleV1ImportMetricIsAttributeFree_Unit(t *testing.T) {
	var found *MetricRegistryRow
	for _, row := range MetricRegistry() {
		if row.Name == IncidentBundleV1ImportMetricName {
			current := row
			found = &current
			break
		}
	}
	if found == nil {
		t.Fatal("v1 import compatibility metric is not registered")
	}
	if len(found.AllowedAttributes) != 0 || len(found.OptionalAttributes) != 0 {
		t.Fatalf("v1 import metric admits public identifiers: %#v", *found)
	}
	RecordIncidentBundleV1Import(context.Background(), "test")
}
