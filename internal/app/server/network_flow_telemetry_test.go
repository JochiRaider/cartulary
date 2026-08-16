package server

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
	telemetrytest "github.com/JochiRaider/cartulary/internal/platform/telemetry/testsupport"
)

func TestNetworkFlowTelemetryUsesClosedSignalsAndPublishedAtAge_Unit(t *testing.T) {
	ctx := context.Background()
	capture := telemetrytest.StartCapture()
	t.Cleanup(func() { capture.Close(context.Background()) })
	observer := newNetworkFlowTelemetryObserver("1.2.3")
	observer.ObserveGraphPhase(ctx, networkflow.GraphPhaseTelemetryObservation{
		Operation: "SENTINEL/private/operation", Phase: "source_scan", GraphMode: "default_flow_edge_v1",
		Result: "SENTINEL raw result", ErrorClass: "SENTINEL SQL select secret", Duration: 1250 * time.Millisecond,
	})
	observer.ObserveGraphResult(ctx, networkflow.GraphResultTelemetryObservation{
		GraphMode: "default_flow_edge_v1", Result: "success", ContributingRows: 4, Vertices: 3, Edges: 2,
	})
	oldestAge := 27 * time.Minute
	observer.ObserveGraphCleanup(ctx, networkflow.GraphCleanupTelemetryObservation{
		Operation: "SENTINEL/source-owner", Result: "success", Duration: 2 * time.Second,
		DeletedLeases: 5, DeletedResults: 1, HealthSnapshotValid: true,
		EligibleResultBacklog: 2, OldestEligibleResultAge: &oldestAge,
	})

	points, err := capture.MetricPoints(ctx)
	if err != nil {
		t.Fatal(err)
	}
	wantNames := map[string]bool{
		telemetry.NetworkFlowGraphPhaseDurationMetricName: false,
		telemetry.NetworkFlowGraphRowsMetricName:          false,
		telemetry.NetworkFlowGraphObjectsMetricName:       false,
		telemetry.NetworkFlowCleanupOperationsMetricName:  false,
		telemetry.NetworkFlowCleanupDurationMetricName:    false,
		telemetry.NetworkFlowCleanupDeletedMetricName:     false,
		telemetry.NetworkFlowCleanupEligibleMetricName:    false,
		telemetry.NetworkFlowCleanupOldestAgeMetricName:   false,
	}
	for _, point := range points {
		if _, relevant := wantNames[point.Name]; !relevant {
			continue
		}
		wantNames[point.Name] = true
		for key, value := range point.Attributes {
			if !networkFlowTelemetryMetricAttribute(point.Name, key) {
				t.Fatalf("metric %s emitted unregistered attribute %q", point.Name, key)
			}
			if strings.Contains(value, "SENTINEL") || strings.Contains(value, "select") || strings.Contains(value, "secret") {
				t.Fatalf("metric %s leaked forbidden value %q", point.Name, value)
			}
		}
		switch point.Name {
		case telemetry.NetworkFlowGraphPhaseDurationMetricName:
			if point.Attributes["cartulary.phase"] != "source_scan" || point.Attributes["cartulary.result"] != "failed" ||
				point.Attributes["cartulary.error_class"] != "internal_error" {
				t.Fatalf("phase attributes = %#v", point.Attributes)
			}
		case telemetry.NetworkFlowCleanupEligibleMetricName:
			if point.Value != 2 || len(point.Attributes) != 0 {
				t.Fatalf("eligible backlog point = %#v", point)
			}
		case telemetry.NetworkFlowCleanupOldestAgeMetricName:
			if !point.IsFloat || point.FloatValue != oldestAge.Seconds() || len(point.Attributes) != 0 {
				t.Fatalf("oldest published-at age point = %#v", point)
			}
		}
	}
	for name, seen := range wantNames {
		if !seen {
			t.Fatalf("Network Flow telemetry omitted registered metric %s: %#v", name, points)
		}
	}

	spans := capture.EndedSpans()
	if len(spans) != 2 {
		t.Fatalf("Network Flow spans = %d want 2: %#v", len(spans), spans)
	}
	for _, span := range spans {
		if span.Name != "cartulary.network_flow.graph.phase" && span.Name != "cartulary.network_flow.cleanup" {
			t.Fatalf("unexpected Network Flow span %q", span.Name)
		}
		for key, value := range span.Attributes {
			if strings.Contains(key, "id") || strings.Contains(value, "SENTINEL") || strings.Contains(value, "secret") {
				t.Fatalf("Network Flow span leaked forbidden field %q=%q", key, value)
			}
		}
	}
}

func TestNetworkFlowTelemetryNoSDKAndInvalidHealthAreNonDisruptive_Unit(t *testing.T) {
	observer := newNetworkFlowTelemetryObserver(telemetry.VersionUnknown)
	negative := -time.Second
	observer.ObserveGraphPhase(context.Background(), networkflow.GraphPhaseTelemetryObservation{
		Phase: "SENTINEL invalid phase", GraphMode: "SENTINEL mode", Result: "success",
	})
	observer.ObserveGraphCleanup(context.Background(), networkflow.GraphCleanupTelemetryObservation{
		Result: "success", HealthSnapshotValid: true, EligibleResultBacklog: -1,
		OldestEligibleResultAge: &negative,
	})
	observer.mu.RLock()
	defer observer.mu.RUnlock()
	if observer.cleanupHealthObserved {
		t.Fatal("invalid telemetry health snapshot was admitted")
	}
}

func networkFlowTelemetryMetricAttribute(metricName string, attributeName string) bool {
	allowed := map[string]map[string]bool{
		telemetry.NetworkFlowGraphPhaseDurationMetricName: {"cartulary.phase": true, "cartulary.graph_mode": true, "cartulary.result": true, "cartulary.error_class": true},
		telemetry.NetworkFlowGraphRowsMetricName:          {"cartulary.graph_mode": true, "cartulary.result": true},
		telemetry.NetworkFlowGraphObjectsMetricName:       {"cartulary.graph_mode": true, "cartulary.graph_object_kind": true, "cartulary.result": true},
		telemetry.NetworkFlowCleanupOperationsMetricName:  {"cartulary.operation": true, "cartulary.result": true, "cartulary.error_class": true},
		telemetry.NetworkFlowCleanupDurationMetricName:    {"cartulary.operation": true, "cartulary.result": true, "cartulary.error_class": true},
		telemetry.NetworkFlowCleanupDeletedMetricName:     {"cartulary.graph_object_kind": true, "cartulary.result": true},
		telemetry.NetworkFlowCleanupEligibleMetricName:    {},
		telemetry.NetworkFlowCleanupOldestAgeMetricName:   {},
	}
	return allowed[metricName][attributeName]
}
