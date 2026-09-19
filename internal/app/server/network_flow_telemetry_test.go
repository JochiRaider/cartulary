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

func TestNetworkFlowTelemetryUsesClosedSignalsAndCleanupFreshness_Unit(t *testing.T) {
	ctx := context.Background()
	capture := telemetrytest.StartCapture()
	t.Cleanup(func() { capture.Close(context.Background()) })
	observer := newNetworkFlowTelemetryObserver("1.2.3")
	now := time.Now()
	observer.now = func() time.Time { return now }
	initial, err := capture.MetricPoints(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, point := range initial {
		if point.Name == telemetry.NetworkFlowCleanupLastSuccessAgeMetricName {
			t.Fatal("freshness emitted before success")
		}
	}
	observer.ObserveGraphPhase(ctx, networkflow.GraphPhaseTelemetryObservation{
		Operation: "SENTINEL/private/operation", Phase: "source_scan", GraphMode: "default_flow_edge_v1",
		Result: "SENTINEL raw result", ErrorClass: "SENTINEL SQL select secret", Duration: 1250 * time.Millisecond,
	})
	observer.ObserveGraphResult(ctx, networkflow.GraphResultTelemetryObservation{
		GraphMode: "default_flow_edge_v1", Result: "success", ContributingRows: 4, Vertices: 3, Edges: 2,
	})
	observer.ObserveGraphCleanup(ctx, networkflow.GraphCleanupTelemetryObservation{
		Operation: "SENTINEL/source-owner", Result: "success", Duration: 2 * time.Second,
		DeletedLeases: 5, DeletedResults: 1, Examined: 3, Continuation: true,
	})

	now = now.Add(27 * time.Minute)
	points, err := capture.MetricPoints(ctx)
	if err != nil {
		t.Fatal(err)
	}
	wantNames := map[string]bool{
		telemetry.NetworkFlowGraphPhaseDurationMetricName:    false,
		telemetry.NetworkFlowGraphRowsMetricName:             false,
		telemetry.NetworkFlowGraphObjectsMetricName:          false,
		telemetry.NetworkFlowCleanupOperationsMetricName:     false,
		telemetry.NetworkFlowCleanupDurationMetricName:       false,
		telemetry.NetworkFlowCleanupDeletedMetricName:        false,
		telemetry.NetworkFlowCleanupExaminedMetricName:       false,
		telemetry.NetworkFlowCleanupContinuationMetricName:   false,
		telemetry.NetworkFlowCleanupLastSuccessAgeMetricName: false,
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
		case telemetry.NetworkFlowCleanupExaminedMetricName:
			if point.Value != 3 || len(point.Attributes) != 0 {
				t.Fatalf("examined progress: %#v", point)
			}
		case telemetry.NetworkFlowCleanupContinuationMetricName:
			if point.Value != 1 || len(point.Attributes) != 0 {
				t.Fatalf("continuation point = %#v", point)
			}
		case telemetry.NetworkFlowCleanupLastSuccessAgeMetricName:
			if !point.IsFloat || point.FloatValue != (27*time.Minute).Seconds() || len(point.Attributes) != 0 {
				t.Fatalf("freshness point = %#v", point)
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
	// Later failure preserves confirmed progress and the successful pacing decision;
	// freshness continues aging without a new observation or database query.
	observer.ObserveGraphCleanup(ctx, networkflow.GraphCleanupTelemetryObservation{Result: "failed", Examined: 2, DeletedResults: 1})
	now = now.Add(time.Minute)
	points, err = capture.MetricPoints(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, point := range points {
		switch point.Name {
		case telemetry.NetworkFlowCleanupExaminedMetricName:
			if point.Value != 5 {
				t.Fatalf("lost committed progress on failure: %#v", point)
			}
		case telemetry.NetworkFlowCleanupContinuationMetricName:
			if point.Value != 1 {
				t.Fatalf("failure replaced continuation: %#v", point)
			}
		case telemetry.NetworkFlowCleanupLastSuccessAgeMetricName:
			if point.FloatValue != (28 * time.Minute).Seconds() {
				t.Fatalf("failure reset freshness: %#v", point)
			}
		}
	}
	observer.ObserveGraphCleanup(ctx, networkflow.GraphCleanupTelemetryObservation{Result: "success"})
	points, err = capture.MetricPoints(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, point := range points {
		if point.Name == telemetry.NetworkFlowCleanupContinuationMetricName && point.Value != 0 {
			t.Fatalf("empty sweep retained continuation: %#v", point)
		}
		if point.Name == telemetry.NetworkFlowCleanupLastSuccessAgeMetricName && point.FloatValue != 0 {
			t.Fatalf("success did not reset freshness: %#v", point)
		}
		if point.Name == "cartulary.network_flow.cleanup.eligible" || point.Name == "cartulary.network_flow.cleanup.oldest_eligible_result.age" {
			t.Fatalf("retired metric emitted: %s", point.Name)
		}
	}

}

func TestNetworkFlowTelemetryNoSDKAndInvalidProgressAreNonDisruptive_Unit(t *testing.T) {
	observer := newNetworkFlowTelemetryObserver(telemetry.VersionUnknown)
	observer.ObserveGraphPhase(context.Background(), networkflow.GraphPhaseTelemetryObservation{
		Phase: "SENTINEL invalid phase", GraphMode: "SENTINEL mode", Result: "success",
	})
	observer.ObserveGraphCleanup(context.Background(), networkflow.GraphCleanupTelemetryObservation{
		Result: "success", Examined: -1,
	})
	observer.mu.RLock()
	defer observer.mu.RUnlock()
	if !observer.lastCleanupSuccess.IsZero() {
		t.Fatal("invalid telemetry progress was admitted")
	}
}

func networkFlowTelemetryMetricAttribute(metricName string, attributeName string) bool {
	allowed := map[string]map[string]bool{
		telemetry.NetworkFlowGraphPhaseDurationMetricName:    {"cartulary.phase": true, "cartulary.graph_mode": true, "cartulary.result": true, "cartulary.error_class": true},
		telemetry.NetworkFlowGraphRowsMetricName:             {"cartulary.graph_mode": true, "cartulary.result": true},
		telemetry.NetworkFlowGraphObjectsMetricName:          {"cartulary.graph_mode": true, "cartulary.graph_object_kind": true, "cartulary.result": true},
		telemetry.NetworkFlowCleanupOperationsMetricName:     {"cartulary.operation": true, "cartulary.result": true, "cartulary.error_class": true},
		telemetry.NetworkFlowCleanupDurationMetricName:       {"cartulary.operation": true, "cartulary.result": true, "cartulary.error_class": true},
		telemetry.NetworkFlowCleanupDeletedMetricName:        {"cartulary.graph_object_kind": true, "cartulary.result": true},
		telemetry.NetworkFlowCleanupExaminedMetricName:       {},
		telemetry.NetworkFlowCleanupContinuationMetricName:   {},
		telemetry.NetworkFlowCleanupLastSuccessAgeMetricName: {},
	}
	return allowed[metricName][attributeName]
}
