package telemetry

import (
	"context"
	"sync"
	"testing"
	"time"

	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func TestDropNewLogProcessorRetainsQueuedRecord(t *testing.T) {
	exporter := &capturingLogExporter{}
	var droppedSignal string
	var droppedReason string
	processor := newDropNewLogProcessor(exporter, dropNewLogProcessorConfig{
		MaxQueueSize:       1,
		MaxExportBatchSize: 1,
		ExportInterval:     time.Hour,
		ExportTimeout:      time.Second,
		SignalKind:         "logs",
		OnDrop: func(_ context.Context, signalKind string, reason string) {
			droppedSignal = signalKind
			droppedReason = reason
		},
	})
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = processor.Shutdown(ctx)
	}()

	first := sdklog.Record{}
	first.SetBody(otellog.StringValue("first"))
	second := sdklog.Record{}
	second.SetBody(otellog.StringValue("second"))

	if err := processor.OnEmit(context.Background(), &first); err != nil {
		t.Fatalf("emit first record: %v", err)
	}
	if err := processor.OnEmit(context.Background(), &second); err != nil {
		t.Fatalf("emit second record: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := processor.ForceFlush(ctx); err != nil {
		t.Fatalf("force flush log processor: %v", err)
	}

	if got := exporter.bodies(); len(got) != 1 || got[0] != "first" {
		t.Fatalf("drop_new must retain queued record and drop the newly offered one, got %#v", got)
	}
	if droppedSignal != "logs" || droppedReason != QueueFullDropReason {
		t.Fatalf("drop_new should report queue-full self metric inputs, got signal=%q reason=%q", droppedSignal, droppedReason)
	}
}

type capturingLogExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (e *capturingLogExporter) Export(_ context.Context, records []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, record := range records {
		e.records = append(e.records, record.Clone())
	}
	return nil
}

func (e *capturingLogExporter) Shutdown(context.Context) error {
	return nil
}

func (e *capturingLogExporter) ForceFlush(context.Context) error {
	return nil
}

func (e *capturingLogExporter) bodies() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	bodies := make([]string, 0, len(e.records))
	for i := range e.records {
		bodies = append(bodies, e.records[i].Body().AsString())
	}
	return bodies
}
