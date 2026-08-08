package testsupport

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

type Capture struct {
	oldMeterProvider  metric.MeterProvider
	oldTracerProvider trace.TracerProvider
	reader            *sdkmetric.ManualReader
	meterProvider     *sdkmetric.MeterProvider
	spanRecorder      *tracetest.SpanRecorder
	tracerProvider    *sdktrace.TracerProvider
}

type Span struct {
	Name       string
	StartedAt  time.Time
	FinishedAt time.Time
	Attributes map[string]string
}

type MetricPoint struct {
	Name       string
	Value      int64
	Attributes map[string]string
}

func StartCapture() *Capture {
	reader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	capture := &Capture{
		oldMeterProvider: otel.GetMeterProvider(), oldTracerProvider: otel.GetTracerProvider(),
		reader: reader, meterProvider: meterProvider, spanRecorder: spanRecorder, tracerProvider: tracerProvider,
	}
	otel.SetMeterProvider(meterProvider)
	otel.SetTracerProvider(tracerProvider)
	return capture
}

func (capture *Capture) Close(ctx context.Context) {
	if capture == nil {
		return
	}
	otel.SetMeterProvider(capture.oldMeterProvider)
	otel.SetTracerProvider(capture.oldTracerProvider)
	_ = capture.meterProvider.Shutdown(ctx)
	_ = capture.tracerProvider.Shutdown(ctx)
}

func (capture *Capture) EndedSpans() []Span {
	if capture == nil {
		return nil
	}
	ended := capture.spanRecorder.Ended()
	result := make([]Span, 0, len(ended))
	for _, span := range ended {
		result = append(result, Span{
			Name: span.Name(), StartedAt: span.StartTime(), FinishedAt: span.EndTime(),
			Attributes: stringAttributes(span.Attributes()),
		})
	}
	return result
}

func (capture *Capture) MetricPoints(ctx context.Context) ([]MetricPoint, error) {
	if capture == nil {
		return nil, nil
	}
	var resource metricdata.ResourceMetrics
	if err := capture.reader.Collect(ctx, &resource); err != nil {
		return nil, err
	}
	var result []MetricPoint
	for _, scope := range resource.ScopeMetrics {
		for _, measurement := range scope.Metrics {
			switch data := measurement.Data.(type) {
			case metricdata.Gauge[int64]:
				for _, point := range data.DataPoints {
					result = append(result, MetricPoint{Name: measurement.Name, Value: point.Value, Attributes: stringAttributes(point.Attributes.ToSlice())})
				}
			case metricdata.Sum[int64]:
				for _, point := range data.DataPoints {
					result = append(result, MetricPoint{Name: measurement.Name, Value: point.Value, Attributes: stringAttributes(point.Attributes.ToSlice())})
				}
			}
		}
	}
	return result, nil
}

func stringAttributes(attrs []attribute.KeyValue) map[string]string {
	result := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		result[string(attr.Key)] = attr.Value.AsString()
	}
	return result
}
