package server

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
	telemetrytest "github.com/JochiRaider/cartulary/internal/platform/telemetry/testsupport"
)

func TestEvidenceCleanupTelemetryUsesClosedMetricsAndLogFields(t *testing.T) {
	ctx := context.Background()
	capture := telemetrytest.StartCapture()
	t.Cleanup(func() { capture.Close(context.Background()) })
	observer, err := newEvidenceCleanupTelemetryObserver("1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	observer.ObserveCleanupSweep(ctx, evidence.CleanupSweepObservation{
		Operation:             "SENTINEL_OPERATION_WITH_OBJECT_KEY/incidents/private/blob",
		Result:                "SENTINEL_RAW_RESULT",
		ErrorClass:            "SENTINEL raw error: capability=secret",
		Duration:              1250 * time.Millisecond,
		HealthSnapshotValid:   true,
		OverdueBlobCount:      3,
		OldestEligibleBlobAge: 27 * time.Minute,
	})

	points, err := capture.MetricPoints(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		telemetry.EvidenceCleanupOperationsMetricName:    false,
		telemetry.EvidenceCleanupSweepDurationMetricName: false,
		telemetry.EvidenceCleanupOverdueMetricName:       false,
		telemetry.EvidenceCleanupOldestAgeMetricName:     false,
	}
	for _, point := range points {
		if _, relevant := want[point.Name]; !relevant {
			continue
		}
		want[point.Name] = true
		for key, value := range point.Attributes {
			if key != "cartulary.operation" && key != "cartulary.result" && key != "cartulary.error_class" {
				t.Fatalf("cleanup metric %s emitted unregistered attribute %q", point.Name, key)
			}
			if strings.Contains(value, "SENTINEL") || strings.Contains(value, "incidents/") || strings.Contains(value, "secret") {
				t.Fatalf("cleanup metric %s leaked forbidden value %q", point.Name, value)
			}
		}
		switch point.Name {
		case telemetry.EvidenceCleanupOperationsMetricName, telemetry.EvidenceCleanupSweepDurationMetricName:
			if point.Attributes["cartulary.operation"] != "cleanup_sweep" ||
				point.Attributes["cartulary.result"] != "failed" ||
				point.Attributes["cartulary.error_class"] != "internal_error" {
				t.Fatalf("cleanup operation metric attributes = %#v", point.Attributes)
			}
		case telemetry.EvidenceCleanupOverdueMetricName, telemetry.EvidenceCleanupOldestAgeMetricName:
			if len(point.Attributes) != 0 {
				t.Fatalf("cleanup health gauge has attributes: %#v", point.Attributes)
			}
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("cleanup telemetry omitted registered metric %s: %#v", name, points)
		}
	}

	operation, result, errorClass := closedEvidenceCleanupTelemetry(evidence.CleanupSweepObservation{
		Operation: "raw-object-key", Result: "raw-result", ErrorClass: "raw error",
	})
	logAttrs := cleanupLogAttributes(operation, result, errorClass)
	if len(logAttrs) != 4 || logAttrs["cartulary.module"] != "evidence" ||
		logAttrs["cartulary.operation"] != "cleanup_sweep" ||
		logAttrs["cartulary.result"] != "failed" ||
		logAttrs["cartulary.error_class"] != "internal_error" {
		t.Fatalf("cleanup log attributes are not closed: %#v", logAttrs)
	}
	if body := "Evidence cleanup sweep completed."; strings.Contains(body, "SENTINEL") || strings.Contains(body, "error") {
		t.Fatalf("cleanup log body is not static and safe: %q", body)
	}
}

func TestEvidenceCleanupDispatcherStartsOnlyAfterServingReadiness(t *testing.T) {
	controller, _, lifecycle := preparedPublicationController(t)
	if err := controller.commit(); err != nil {
		t.Fatal(err)
	}
	acknowledgeAllPublicationComponents(t, controller)
	dispatcher := &serverCleanupLifecycle{started: make(chan bool, 1), readiness: lifecycle.AdmissionOpen}
	runtime := &Runtime{
		publication:               controller,
		lifecycle:                 lifecycle,
		evidenceCleanupDispatcher: dispatcher,
	}
	runtime.own(func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = dispatcher.Close(closeCtx)
	})
	t.Cleanup(runtime.Close)
	select {
	case <-dispatcher.started:
		t.Fatal("cleanup dispatcher started before serving readiness")
	default:
	}
	if err := runtime.ActivatePublication(); err != nil {
		t.Fatal(err)
	}
	select {
	case ready := <-dispatcher.started:
		if !ready {
			t.Fatal("cleanup first sweep began before serving readiness opened")
		}
	case <-time.After(time.Second):
		t.Fatal("cleanup first sweep did not begin after serving readiness")
	}
}

type serverCleanupLifecycle struct {
	started   chan bool
	readiness func() bool
}

func (lifecycle *serverCleanupLifecycle) Start(context.Context) error {
	lifecycle.started <- lifecycle.readiness()
	return nil
}

func (*serverCleanupLifecycle) Close(context.Context) error { return nil }
