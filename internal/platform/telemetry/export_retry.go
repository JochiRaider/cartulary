package telemetry

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

type retrySettings struct {
	Enabled         bool
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxElapsedTime  time.Duration
	Multiplier      float64
}

type retryController struct {
	settings    retrySettings
	sampleDelay func(time.Duration) time.Duration
	sleep       func(context.Context, time.Duration) error
	now         func() time.Time
}

func newRetryController(cfg config.Config) retryController {
	return retryController{
		settings: retrySettings{
			Enabled:         cfg.Telemetry.Exporter.Retry.Enabled,
			InitialInterval: durationMS(cfg.Telemetry.Exporter.Retry.InitialIntervalMS),
			MaxInterval:     durationMS(cfg.Telemetry.Exporter.Retry.MaxIntervalMS),
			MaxElapsedTime:  durationMS(cfg.Telemetry.Exporter.Retry.MaxElapsedMS),
			Multiplier:      cfg.Telemetry.Exporter.Retry.Multiplier,
		},
		sampleDelay: randomFullJitterDelay,
		sleep:       sleepContext,
		now:         time.Now,
	}
}

func (c retryController) run(ctx context.Context, fn func(context.Context) error) error {
	if fn == nil {
		return nil
	}
	err := fn(ctx)
	if err == nil || !c.settings.Enabled || c.settings.MaxElapsedTime <= 0 {
		return err
	}
	now := c.now
	if now == nil {
		now = time.Now
	}
	started := now()
	for attempt := int64(1); ; attempt++ {
		classification, throttleDelay := classifyExportRetry(err)
		if classification != RetryTransient {
			return err
		}
		delay := throttleDelay
		if delay <= 0 {
			delay = c.retryDelay(attempt)
		}
		elapsed := now().Sub(started)
		if elapsed+delay > c.settings.MaxElapsedTime {
			return fmt.Errorf("max retry time would elapse: %w", err)
		}
		sleep := c.sleep
		if sleep == nil {
			sleep = sleepContext
		}
		if sleepErr := sleep(ctx, delay); sleepErr != nil {
			return sleepErr
		}
		err = fn(ctx)
		if err == nil {
			return nil
		}
		if now().Sub(started) >= c.settings.MaxElapsedTime {
			return fmt.Errorf("max retry time elapsed: %w", err)
		}
	}
}

func (c retryController) retryDelay(attempt int64) time.Duration {
	base := retryBaseDelay(c.settings, attempt)
	sample := c.sampleDelay
	if sample == nil {
		sample = randomFullJitterDelay
	}
	delay := sample(base)
	if delay < 0 {
		return 0
	}
	if delay > base {
		return base
	}
	return delay
}

func retryBaseDelay(settings retrySettings, attempt int64) time.Duration {
	if attempt < 1 || settings.InitialInterval <= 0 {
		return 0
	}
	multiplier := settings.Multiplier
	if multiplier < 1 {
		multiplier = 1
	}
	base := float64(settings.InitialInterval)
	if attempt > 1 {
		base *= math.Pow(multiplier, float64(attempt-1))
	}
	if max := float64(settings.MaxInterval); max > 0 && base > max {
		base = max
	}
	if base > float64(math.MaxInt64) {
		return time.Duration(math.MaxInt64)
	}
	return time.Duration(base)
}

func randomFullJitterDelay(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	limit := big.NewInt(max.Nanoseconds() + 1)
	sample, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return max
	}
	return time.Duration(sample.Int64())
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		if cause := context.Cause(ctx); cause != nil {
			return cause
		}
		return ctx.Err()
	}
}

func classifyExportRetry(err error) (RetryClassification, time.Duration) {
	if err == nil {
		return "", 0
	}
	if st, ok := status.FromError(err); ok && st.Code() != codes.OK {
		retryDelay := grpcRetryInfoDelay(st)
		return ClassifyOTLPGRPCStatus(st.Code().String(), retryDelay > 0), retryDelay
	}
	if strings.Contains(err.Error(), "retry-able request failure") {
		return RetryTransient, 0
	}
	return RetryPermanent, 0
}

func grpcRetryInfoDelay(st *status.Status) time.Duration {
	for _, detail := range st.Details() {
		if retryInfo, ok := detail.(*errdetails.RetryInfo); ok && retryInfo.RetryDelay != nil {
			return retryInfo.RetryDelay.AsDuration()
		}
	}
	return 0
}

type retryingSpanExporter struct {
	exporter   sdktrace.SpanExporter
	controller retryController
}

func (e *retryingSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if e == nil || e.exporter == nil {
		return nil
	}
	return e.controller.run(ctx, func(ctx context.Context) error {
		return e.exporter.ExportSpans(ctx, spans)
	})
}

func (e *retryingSpanExporter) Shutdown(ctx context.Context) error {
	if e == nil || e.exporter == nil {
		return nil
	}
	return e.exporter.Shutdown(ctx)
}

type retryingMetricExporter struct {
	exporter   sdkmetric.Exporter
	controller retryController
}

func (e *retryingMetricExporter) Temporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	if e == nil || e.exporter == nil {
		return metricdata.CumulativeTemporality
	}
	return e.exporter.Temporality(kind)
}

func (e *retryingMetricExporter) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	if e == nil || e.exporter == nil {
		return sdkmetric.DefaultAggregationSelector(kind)
	}
	return e.exporter.Aggregation(kind)
}

func (e *retryingMetricExporter) Export(ctx context.Context, metrics *metricdata.ResourceMetrics) error {
	if e == nil || e.exporter == nil {
		return nil
	}
	return e.controller.run(ctx, func(ctx context.Context) error {
		return e.exporter.Export(ctx, metrics)
	})
}

func (e *retryingMetricExporter) ForceFlush(ctx context.Context) error {
	if e == nil || e.exporter == nil {
		return nil
	}
	return e.exporter.ForceFlush(ctx)
}

func (e *retryingMetricExporter) Shutdown(ctx context.Context) error {
	if e == nil || e.exporter == nil {
		return nil
	}
	return e.exporter.Shutdown(ctx)
}

type retryingLogExporter struct {
	exporter   sdklog.Exporter
	controller retryController
}

func (e *retryingLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	if e == nil || e.exporter == nil {
		return nil
	}
	return e.controller.run(ctx, func(ctx context.Context) error {
		return e.exporter.Export(ctx, records)
	})
}

func (e *retryingLogExporter) Shutdown(ctx context.Context) error {
	if e == nil || e.exporter == nil {
		return nil
	}
	return e.exporter.Shutdown(ctx)
}

func (e *retryingLogExporter) ForceFlush(ctx context.Context) error {
	if e == nil || e.exporter == nil {
		return nil
	}
	return e.exporter.ForceFlush(ctx)
}
