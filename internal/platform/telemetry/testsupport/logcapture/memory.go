package logcapture

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// Record is an observation at the instrumentation boundary, before export.
type Record struct {
	Body           string
	Scope          string
	Version        string
	SeverityNumber int
	SeverityText   string
	Attributes     map[string]string
}

type Capture struct {
	previous  otellog.LoggerProvider
	provider  *sdklog.LoggerProvider
	processor *memoryProcessor
}

// Start installs a synchronous recorder. Callers must serialize global-provider
// changes and close the capture before another test uses the global logger.
func Start() *Capture {
	processor := &memoryProcessor{}
	capture := &Capture{
		previous: logglobal.GetLoggerProvider(), processor: processor,
		provider: sdklog.NewLoggerProvider(sdklog.WithProcessor(processor)),
	}
	logglobal.SetLoggerProvider(capture.provider)
	return capture
}

func (capture *Capture) Close(ctx context.Context) error {
	logglobal.SetLoggerProvider(capture.previous)
	return capture.provider.Shutdown(ctx)
}

func (capture *Capture) Records() []Record {
	capture.processor.mu.Lock()
	defer capture.processor.mu.Unlock()
	result := make([]Record, 0, len(capture.processor.records))
	for _, record := range capture.processor.records {
		attrs := make(map[string]string, record.AttributesLen())
		record.WalkAttributes(func(attr attribute.KeyValue) bool {
			attrs[string(attr.Key)] = attr.Value.AsString()
			return true
		})
		result = append(result, Record{
			Body: record.Body().AsString(), Scope: record.InstrumentationScope().Name,
			Version:        record.InstrumentationScope().Version,
			SeverityNumber: int(record.Severity()), SeverityText: record.SeverityText(),
			Attributes: attrs,
		})
	}
	return result
}

type memoryProcessor struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (*memoryProcessor) Enabled(context.Context, sdklog.EnabledParameters) bool { return true }

func (processor *memoryProcessor) OnEmit(_ context.Context, record *sdklog.Record) error {
	processor.mu.Lock()
	defer processor.mu.Unlock()
	processor.records = append(processor.records, record.Clone())
	return nil
}

func (*memoryProcessor) Shutdown(context.Context) error   { return nil }
func (*memoryProcessor) ForceFlush(context.Context) error { return nil }
